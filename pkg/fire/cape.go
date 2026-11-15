package fire

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// CAPESnapshot summarises where equity valuations stand today, the single best
// predictor of the next decade's real return and therefore of a retirement's
// make-or-break first years.
type CAPESnapshot struct {
	AsOf        string  `json:"asOf"`        // month of the latest observation
	Value       float64 `json:"value"`       // latest CAPE (PE10)
	Percentile  float64 `json:"percentile"`  // 0-100 rank in the full history
	Median      float64 `json:"median"`      // historical median CAPE
	ImpliedReal float64 `json:"impliedReal"` // 1/CAPE, geometric real, as a fraction
	Stale       bool    `json:"stale"`       // observation older than a year: warn, never present as "now"
}

// capeHistory is the bundled Shiller CAPE history, ascending by date, parsed once.
var capeHistory = sync.OnceValue(func() []CAPEPoint {
	return parseCape(datasets.CAPE())
})

// CAPEPoint is one dated observation of the cyclically adjusted price/earnings ratio.
type CAPEPoint struct {
	Date  string  // YYYY-MM-DD observation date
	Value float64 // positive dimensionless price/earnings ratio
}

// CurrentCAPE returns the current valuation reading: the latest CAPE, its
// percentile in the full history, the historical median, and the implied real
// return (the earnings yield 1/CAPE, a first-order estimate of the next decade).
func CurrentCAPE() CAPESnapshot {
	s := capeHistory()
	if len(s) == 0 {
		return CAPESnapshot{}
	}
	last := s[len(s)-1]
	vals := make([]float64, len(s))
	below := 0
	for i, p := range s {
		vals[i] = p.Value
		if p.Value < last.Value {
			below++
		}
	}
	sort.Float64s(vals)
	median := vals[len(vals)/2]
	return CAPESnapshot{
		AsOf:        last.Date,
		Value:       last.Value,
		Percentile:  100 * float64(below) / float64(len(s)),
		Median:      median,
		ImpliedReal: 1 / last.Value,
		Stale:       CAPEIsStale(last.Date, time.Now()),
	}
}

// CAPEIsStale reports whether the latest bundled observation is more than a
// year old. A three-year-old reading once shipped as "Valuation now"; the UI
// must be told so it can say "as of <month>" with a warning instead.
func CAPEIsStale(date string, now time.Time) bool {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return true
	}
	return now.Sub(t) > 366*24*time.Hour
}

func parseCape(csv []byte) []CAPEPoint {
	var out []CAPEPoint
	for _, line := range strings.Split(string(csv), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "date,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) < 2 {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(f[1]), 64)
		if err != nil || !finite(v) || v <= 0 {
			continue
		}
		out = append(out, CAPEPoint{Date: f[0], Value: v})
	}
	return out
}

// CAPEHistory returns the bundled chronological Shiller observations. The
// returned slice is shared read-only data and must not be modified.
func CAPEHistory() []CAPEPoint { return capeHistory() }

// Valid reports a finite positive latest observation with a valid date.
func (c CAPESnapshot) Valid() bool {
	_, err := time.Parse("2006-01-02", c.AsOf)
	return err == nil && c.Value > 0 && !math.IsNaN(c.Value) && !math.IsInf(c.Value, 0)
}

// ArithmeticMean converts the CAPE-implied geometric yield to an arithmetic
// real annual mean using sigma squared / 2. Sigma is an annual fraction.
// An invalid or unavailable snapshot returns NaN and an error; stale but valid
// observations remain usable with the snapshot's explicit Stale indication.
func (c CAPESnapshot) ArithmeticMean(sigma float64) (float64, error) {
	if !c.Valid() || math.IsNaN(sigma) || math.IsInf(sigma, 0) || sigma < 0 {
		return math.NaN(), fmt.Errorf("fire: invalid CAPE observation or volatility")
	}
	return c.ImpliedReal + sigma*sigma/2, nil
}
