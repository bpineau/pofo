package web

import (
	"fmt"
	"math"

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

// BuildPanel deflates each asset by hicp and aligns the resulting annual
// real returns into a scenario.Panel. Assets are truncated to their common
// number of yearly returns so every row has the same length.
func BuildPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, error) {
	if len(assets) == 0 {
		return scenario.Panel{}, fmt.Errorf("no assets")
	}
	rows := make([][]float64, len(assets))
	weights := make([]float64, len(assets))
	min := -1
	for i, a := range assets {
		rows[i] = annualReal(a.Points, hicp)
		weights[i] = a.Weight
		if min < 0 || len(rows[i]) < min {
			min = len(rows[i])
		}
	}
	if min <= 0 {
		return scenario.Panel{}, fmt.Errorf("not enough history")
	}
	for i := range rows {
		rows[i] = rows[i][len(rows[i])-min:] // keep the last min years (common window)
	}
	return scenario.Panel{Returns: rows, Weights: normalize(weights)}, nil
}

// annualReal samples one real return per calendar year from points using the
// last quote of each year, deflated by hicp.
func annualReal(points, hicp []marketdata.Point) []float64 {
	yearly := lastPerYear(points)
	return scenario.Deflate(yearly, hicp)
}

// lastPerYear keeps the last point of each calendar year, ascending.
func lastPerYear(points []marketdata.Point) []marketdata.Point {
	var out []marketdata.Point
	for _, p := range points {
		if n := len(out); n > 0 && out[n-1].Date.Year() == p.Date.Year() {
			out[n-1] = p
		} else {
			out = append(out, p)
		}
	}
	return out
}

// FitParametric returns the sample mean and standard deviation of the
// weighted annual real returns, to seed the parametric sliders.
func FitParametric(panel scenario.Panel, weights []float64) (mu, sigma float64) {
	seq := panel.Combine(weights)
	if len(seq) == 0 {
		return 0, 0
	}
	mu = metrics.Mean(seq)
	for _, r := range seq {
		sigma += (r - mu) * (r - mu)
	}
	if len(seq) > 1 {
		sigma = math.Sqrt(sigma / float64(len(seq)-1))
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
