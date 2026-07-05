package web

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
)

// capeGauge renders the current valuation reading as a half-circle gauge: the
// needle at today's CAPE percentile, the value and the implied real return at
// the centre. Static (it does not depend on the sliders), so it is served with
// the page metadata.
func capeGauge() string {
	s := capeSnapshot()
	if s.Value == 0 {
		return ""
	}
	caption := fmt.Sprintf("CAPE %s · %.1f%% implied real", s.AsOf, s.ImpliedReal*100)
	return chart.Gauge(chart.Options{Width: 380, Height: 214},
		strconv.FormatFloat(s.Value, 'f', 1, 64), caption, "cheap", "rich", s.Percentile/100)
}

// CapeSnapshot summarises where equity valuations stand today, the single best
// predictor of the next decade's real return and therefore of a retirement's
// make-or-break first years.
type CapeSnapshot struct {
	AsOf        string  `json:"asOf"`        // month of the latest observation
	Value       float64 `json:"value"`       // latest CAPE (PE10)
	Percentile  float64 `json:"percentile"`  // 0-100 rank in the full history
	Median      float64 `json:"median"`      // historical median CAPE
	ImpliedReal float64 `json:"impliedReal"` // 1/CAPE, geometric real, as a fraction
}

// capeSeries is the bundled Shiller CAPE history, ascending by date, parsed once.
var capeSeries = sync.OnceValue(func() []capePoint {
	return parseCape(datasets.CAPE())
})

type capePoint struct {
	date string
	cape float64
}

// capeSnapshot returns the current valuation reading: the latest CAPE, its
// percentile in the full history, the historical median, and the implied real
// return (the earnings yield 1/CAPE, a first-order estimate of the next decade).
func capeSnapshot() CapeSnapshot {
	s := capeSeries()
	if len(s) == 0 {
		return CapeSnapshot{}
	}
	last := s[len(s)-1]
	vals := make([]float64, len(s))
	below := 0
	for i, p := range s {
		vals[i] = p.cape
		if p.cape < last.cape {
			below++
		}
	}
	sort.Float64s(vals)
	median := vals[len(vals)/2]
	return CapeSnapshot{
		AsOf:        last.date,
		Value:       last.cape,
		Percentile:  100 * float64(below) / float64(len(s)),
		Median:      median,
		ImpliedReal: 1 / last.cape,
	}
}

// capeAdjustedMu returns the central arithmetic mean to plan on when the user
// opts to anchor returns to today's valuation. It converts the CAPE-implied real
// geometric return (1/CAPE) to arithmetic (adding sigma^2/2, the volatility
// drag) so it feeds ParametricSource.Mu consistently. Rich valuations pull the
// central case down; cheap ones lift it.
func capeAdjustedMu(sigma float64) float64 {
	return capeSnapshot().ImpliedReal + sigma*sigma/2
}

func parseCape(csv []byte) []capePoint {
	var out []capePoint
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
		if err != nil || v <= 0 {
			continue
		}
		out = append(out, capePoint{date: f[0], cape: v})
	}
	return out
}
