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

// Fit seeds the parametric sliders from a portfolio's history: the i.i.d.
// annual Student-t the kernel draws from.
type Fit struct {
	Mu    float64 // mean realised annual real return
	Sigma float64 // annualised volatility, from the monthly dispersion (σ_m·√12)
	Df    float64 // Student-t dof seeded from the monthly excess kurtosis
}

// Valid reports whether the fit carries usable estimates. A panel shorter
// than about two years yields the zero Fit; seeding the UI sliders with
// µ=0/σ=0 turns the central case into a certain-doom model, so callers must
// keep their defaults when a fit is not Valid.
func (f Fit) Valid() bool { return f.Sigma > 0 }

// FitParametric estimates the parametric annual model from a monthly panel.
//
//   - Mu is the arithmetic mean of the realised annual real returns.
//   - Sigma is the monthly real-return standard deviation scaled by √12, far
//     more stable than the std of the ~20 annual points and the right i.i.d.
//     annual sigma for the model. It is typically BELOW the volatility shown on
//     the main report (daily returns ×√252): daily-annualised vol overstates
//     the dispersion realised at an annual horizon when returns mean-revert or
//     trend (vol drag). The slider lets the user raise it for a more
//     conservative test.
//   - Df is seeded from the monthly excess kurtosis (Student-t excess kurtosis
//     6/(df−4)). It is a rough hint: monthly returns are fatter-tailed than the
//     annual aggregates the model actually draws (returns sum toward normal), so
//     this errs toward heavier tails. A user-adjustable seed, not a precise fit.
func FitParametric(panel scenario.Panel, weights []float64) Fit {
	monthly := panel.Combine(weights)
	annual := scenario.Annualize(monthly, 12)
	if len(annual) == 0 || len(monthly) < 2 {
		return Fit{}
	}
	return Fit{
		Mu:    metrics.Mean(annual),
		Sigma: stdev(monthly) * math.Sqrt(12),
		Df:    dofFromKurtosis(metrics.ExcessKurtosis(monthly)),
	}
}

// stdev is the sample (n−1) standard deviation; 0 for fewer than two points.
func stdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := metrics.Mean(xs)
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)-1))
}

// dofFromKurtosis maps a sample excess kurtosis to a Student-t dof seed,
// inverting the t excess kurtosis 6/(df−4). Thin-tailed or undefined samples
// map to the near-normal end (30); the result is clamped to the slider range.
func dofFromKurtosis(excess float64) float64 {
	if math.IsNaN(excess) || excess <= 0 {
		return 30
	}
	return math.Max(3, math.Min(4+6/excess, 30))
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
