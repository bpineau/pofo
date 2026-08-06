package web

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/scenario"
)

// AssetSeries is one holding's weight and its (nominal) price points,
// already converted to the report currency.
type AssetSeries struct {
	Weight float64
	Points []marketdata.Point
}

// BuildMonthlyPanel deflates each asset by hicp and aligns the resulting
// monthly real returns into a scenario.Panel (indexed [asset][month]). Assets
// are truncated to their common number of monthly returns so every row has
// the same length. Monthly sampling gives the historical models ~12x more
// data points than annual, so the bootstrap captures intra-year regimes and
// the cohorts model has many more windows.
func BuildMonthlyPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, error) {
	if len(assets) == 0 {
		return scenario.Panel{}, fmt.Errorf("no assets")
	}
	rows := make([][]float64, len(assets))
	weights := make([]float64, len(assets))
	min := -1
	for i, a := range assets {
		rows[i] = scenario.Deflate(lastPerMonth(a.Points), hicp)
		weights[i] = a.Weight
		if min < 0 || len(rows[i]) < min {
			min = len(rows[i])
		}
	}
	if min <= 0 {
		return scenario.Panel{}, fmt.Errorf("not enough history")
	}
	for i := range rows {
		rows[i] = rows[i][len(rows[i])-min:] // keep the last min months (common window)
	}
	return scenario.Panel{Returns: rows, Weights: normalize(weights)}, nil
}

// lastPerMonth keeps the last point of each calendar month, ascending.
func lastPerMonth(points []marketdata.Point) []marketdata.Point {
	var out []marketdata.Point
	sameMonth := func(a, b time.Time) bool {
		return a.Year() == b.Year() && a.Month() == b.Month()
	}
	for _, p := range points {
		if n := len(out); n > 0 && sameMonth(out[n-1].Date, p.Date) {
			out[n-1] = p
		} else {
			out = append(out, p)
		}
	}
	return out
}

// FitParametric returns the annualised mean and standard deviation of the
// weighted real returns of a monthly panel, to seed the parametric mu/sigma
// sliders. The monthly returns are compounded into annual returns first, so
// the figures are directly comparable to the (annual) parametric model.
func FitParametric(panel scenario.Panel, weights []float64) (mu, sigma float64) {
	annual := scenario.Annualize(panel.Combine(weights), 12)
	if len(annual) == 0 {
		return 0, 0
	}
	mu = metrics.Mean(annual)
	for _, r := range annual {
		sigma += (r - mu) * (r - mu)
	}
	if len(annual) > 1 {
		sigma = math.Sqrt(sigma / float64(len(annual)-1))
	}
	return mu, sigma
}

// normalize scales weights to sum to 1 (returned unchanged if they sum to 0).
func normalize(w []float64) []float64 {
	sum := 0.0
	for _, x := range w {
		sum += x
	}
	if sum == 0 {
		return w
	}
	out := make([]float64, len(w))
	for i, x := range w {
		out[i] = x / sum
	}
	return out
}
