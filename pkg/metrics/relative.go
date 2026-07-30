package metrics

import (
	"math"
	"slices"
	"time"
)

// Drawdowns returns the running drawdown of a value series: for each point,
// the fraction lost from the highest value seen so far (0 at new highs,
// -0.25 when 25 % below the peak).
func Drawdowns(values []float64) []float64 {
	out := make([]float64, len(values))
	peak := math.Inf(-1)
	for i, v := range values {
		if v > peak {
			peak = v
		}
		out[i] = v/peak - 1
	}
	return out
}

// RollingCAGR computes the annualized return of every rolling window of the
// given calendar length (one window per starting point) and summarizes the
// distribution. ok is false when the series is shorter than the window.
func RollingCAGR(dates []time.Time, values []float64, years float64) (worst, median, best float64, windows int, ok bool) {
	if len(dates) != len(values) || years <= 0 {
		return 0, 0, 0, 0, false
	}
	span := time.Duration(years * 365.25 * 24 * float64(time.Hour))
	var cagrs []float64
	j := 0
	for i := range dates {
		target := dates[i].Add(span)
		for j < len(dates) && dates[j].Before(target) {
			j++
		}
		if j >= len(dates) {
			break
		}
		actualYears := dates[j].Sub(dates[i]).Hours() / 24 / 365.25
		if values[i] > 0 && values[j] > 0 && actualYears > 0 {
			cagrs = append(cagrs, math.Pow(values[j]/values[i], 1/actualYears)-1)
		}
		j = max(j-1, 0) // windows overlap: restart search near previous end
	}
	if len(cagrs) < 2 {
		return 0, 0, 0, 0, false
	}
	sorted := append([]float64(nil), cagrs...)
	slices.Sort(sorted)
	return sorted[0], sorted[len(sorted)/2], sorted[len(sorted)-1], len(cagrs), true
}

// Relative compares a series with a benchmark on their common dates and
// derives the classic relative-performance statistics. ok is false when
// fewer than 30 dates overlap. Conventions: daily returns, 252-day
// annualization, risk-free rate 0 (consistent with Compute).
type Relative struct {
	Beta        float64
	Alpha       float64 // Jensen's alpha, annualized (0.02 = +2 %/yr)
	InfoRatio   float64 // mean active return / tracking error, annualized
	UpCapture   float64 // geometric capture on benchmark up days (1.10 = 110 %)
	DownCapture float64 // geometric capture on benchmark down days
}

// VsBenchmark computes Relative statistics for values against the benchmark.
func VsBenchmark(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64) (Relative, bool) {
	if len(dates) != len(values) || len(benchDates) != len(benchValues) || len(dates) < 2 || len(benchDates) < 2 {
		return Relative{}, false
	}
	bench := make(map[time.Time]float64, len(benchDates)-1)
	for i := 1; i < len(benchDates); i++ {
		bench[benchDates[i]] = benchValues[i]/benchValues[i-1] - 1
	}
	var rp, rb []float64
	for i := 1; i < len(dates); i++ {
		if br, found := bench[dates[i]]; found {
			rp = append(rp, values[i]/values[i-1]-1)
			rb = append(rb, br)
		}
	}
	if len(rp) < minBetaOverlap {
		return Relative{}, false
	}
	mp, mb := Mean(rp), Mean(rb)
	var cov, varb, varActive float64
	upP, upB, downP, downB := 1.0, 1.0, 1.0, 1.0
	nUp, nDown := 0, 0
	meanActive := mp - mb
	for i := range rp {
		cov += (rp[i] - mp) * (rb[i] - mb)
		varb += (rb[i] - mb) * (rb[i] - mb)
		active := rp[i] - rb[i]
		varActive += (active - meanActive) * (active - meanActive)
		switch {
		case rb[i] > 0:
			upP *= 1 + rp[i]
			upB *= 1 + rb[i]
			nUp++
		case rb[i] < 0:
			downP *= 1 + rp[i]
			downB *= 1 + rb[i]
			nDown++
		}
	}
	if varb == 0 {
		return Relative{}, false
	}
	var r Relative
	r.Beta = cov / varb
	r.Alpha = (mp - r.Beta*mb) * tradingDaysPerYear
	if te := math.Sqrt(varActive/float64(len(rp))) * math.Sqrt(tradingDaysPerYear); te > 0 {
		r.InfoRatio = meanActive * tradingDaysPerYear / te
	}
	capture := func(p, b float64, n int) float64 {
		if n == 0 {
			return math.NaN()
		}
		gp := math.Pow(p, 1/float64(n)) - 1
		gb := math.Pow(b, 1/float64(n)) - 1
		if gb == 0 {
			return math.NaN()
		}
		return gp / gb
	}
	r.UpCapture = capture(upP, upB, nUp)
	r.DownCapture = capture(downP, downB, nDown)
	return r, true
}
