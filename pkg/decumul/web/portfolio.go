package web

import (
	"fmt"
	"math"
	"sort"
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
// monthly real returns into a scenario.Panel (indexed [asset][month]). Rows are
// aligned on shared calendar months (intersection of month keys), not by
// trailing position, so holdings with different start/end months or internal
// gaps stay column-aligned: every column is the same month across all assets.
// Only genuine one-month returns count; a return spanning a gap is dropped.
// Monthly sampling gives the historical models ~12x more data points than
// annual, so the bootstrap captures intra-year regimes and the cohorts model
// has many more windows.
func BuildMonthlyPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, error) {
	if len(assets) == 0 {
		return scenario.Panel{}, fmt.Errorf("no assets")
	}
	// Per asset, the real return of each calendar month keyed by month index,
	// keeping only months whose previous month is also present (true one-month
	// returns). counts tracks how many assets cover each month.
	perAsset := make([]map[int]float64, len(assets))
	weights := make([]float64, len(assets))
	counts := make(map[int]int)
	for i, a := range assets {
		pts := lastPerMonth(a.Points)
		rets := scenario.Deflate(pts, hicp)
		m := make(map[int]float64, len(rets))
		for j, r := range rets {
			key := monthKey(pts[j+1].Date)
			if key != monthKey(pts[j].Date)+1 {
				continue // not calendar-consecutive: a spanning return, drop it
			}
			m[key] = r
			counts[key]++
		}
		perAsset[i] = m
		weights[i] = a.Weight
	}
	// Months covered by every asset, in ascending order.
	var common []int
	for key, c := range counts {
		if c == len(assets) {
			common = append(common, key)
		}
	}
	sort.Ints(common)
	if len(common) == 0 {
		return scenario.Panel{}, fmt.Errorf("not enough overlapping monthly history")
	}
	rows := make([][]float64, len(assets))
	for i := range assets {
		row := make([]float64, len(common))
		for k, key := range common {
			row[k] = perAsset[i][key]
		}
		rows[i] = row
	}
	return scenario.Panel{Returns: rows, Weights: normalize(weights)}, nil
}

// monthKey maps a date to a dense month index (year*12 + month), so consecutive
// calendar months differ by exactly one.
func monthKey(t time.Time) int { return t.Year()*12 + int(t.Month()) - 1 }

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

// FitParametric returns the mean and standard deviation of the weighted
// real ANNUAL returns of a monthly panel, to seed the parametric mu/sigma
// sliders. Both are the directly relevant quantities for an i.i.d. annual
// model: mu is the arithmetic mean of the realised annual real returns and
// sigma is their dispersion.
//
// This sigma is typically BELOW the volatility shown on the main report,
// which annualises daily returns (×√252). The two measure different things:
// daily-annualised vol overstates the realised dispersion of annual returns
// whenever returns mean-revert or the strategy trends (vol drag). The annual
// dispersion is the honest input for the annual kernel; the slider lets the
// user raise sigma toward the headline figure for a more conservative test.
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
