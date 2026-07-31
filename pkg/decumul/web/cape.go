package web

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
)

// capeHistory renders the Shiller CAPE series since 1881 as a single-axis line
// (no dual axis: the implied return is a readout, not a second scale), with the
// historical median marked and today's value emphasised at the right end. The
// accent-coloured line darkens to amber in the terminal theme. Static, so it is
// served with the page metadata.
func capeHistory() string {
	s := capeSeries()
	if len(s) == 0 {
		return ""
	}
	xs := make([]float64, len(s))
	ys := make([]float64, len(s))
	for i, p := range s {
		xs[i] = capeYear(p.date)
		ys[i] = p.cape
	}
	snap := capeSnapshot()
	return chart.MultiLine(
		chart.Options{Width: 760, Height: 300},
		"", "CAPE (PE10)",
		[]chart.XYSeries{{Name: "CAPE", Xs: xs, Ys: ys, Color: "#0B7285"}},
		chart.Marker{Axis: 'y', Value: snap.Median, Label: "median"},
		chart.Marker{Axis: 'x', Value: xs[len(xs)-1], Label: "today"},
	)
}

// capeYear turns a "YYYY-MM-DD" date into a fractional year for the x-axis.
func capeYear(date string) float64 {
	y, m := 0, 1
	if len(date) >= 7 {
		y, _ = strconv.Atoi(date[:4])
		m, _ = strconv.Atoi(date[5:7])
	}
	return float64(y) + (float64(m)-0.5)/12
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
	Stale       bool    `json:"stale"`       // observation older than a year: warn, never present as "now"
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
		Stale:       capeIsStale(last.date, time.Now()),
	}
}

// capeIsStale reports whether the latest bundled observation is more than a
// year old. A three-year-old reading once shipped as "Valuation now"; the UI
// must be told so it can say "as of <month>" with a warning instead.
func capeIsStale(date string, now time.Time) bool {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return true
	}
	return now.Sub(t) > 366*24*time.Hour
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
