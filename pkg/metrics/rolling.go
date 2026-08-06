package metrics

import (
	"math"
	"time"
)

// Rolling applies fn to every trailing calendar window of the given length
// (in years) and returns one (date, value) pair per window, the date being
// the window's end. A window at index j spans the points whose dates fall in
// (dates[j]-years, dates[j]]; only windows that reach back the full length
// are emitted. fn receives the window's values (levels, in order). ok is
// false when no full-length window fits in the series.
//
// It is the shared engine behind RollingVol, RollingSharpe, RollingSortino
// and RollingUlcer; pass any summarizing fn for a custom rolling statistic.
func Rolling(dates []time.Time, values []float64, years float64, fn func(window []float64) float64) (points []time.Time, out []float64, ok bool) {
	if len(dates) != len(values) || years <= 0 {
		return nil, nil, false
	}
	span := time.Duration(years * 365.25 * 24 * float64(time.Hour))
	i := 0
	for j := range dates {
		lo := dates[j].Add(-span)
		if dates[0].After(lo) {
			continue // series does not yet cover a full window ending at j
		}
		for dates[i].Before(lo) || dates[i].Equal(lo) {
			i++
		}
		points = append(points, dates[j])
		out = append(out, fn(values[i:j+1]))
	}
	if len(out) == 0 {
		return nil, nil, false
	}
	return points, out, true
}

// RollingVol is the annualized volatility of each trailing window of the
// given length in years.
func RollingVol(dates []time.Time, values []float64, years float64) ([]time.Time, []float64, bool) {
	return Rolling(dates, values, years, windowVol)
}

// RollingSharpe is the annualized Sharpe ratio (risk-free 0) of each trailing
// window of the given length in years.
func RollingSharpe(dates []time.Time, values []float64, years float64) ([]time.Time, []float64, bool) {
	return Rolling(dates, values, years, func(w []float64) float64 {
		r := Returns(w)
		vol := windowVol(w)
		if vol == 0 {
			return math.NaN()
		}
		return Mean(r) * tradingDaysPerYear / vol
	})
}

// RollingSortino is the annualized Sortino ratio (downside deviation, target
// 0) of each trailing window of the given length in years.
func RollingSortino(dates []time.Time, values []float64, years float64) ([]time.Time, []float64, bool) {
	return Rolling(dates, values, years, func(w []float64) float64 {
		r := Returns(w)
		if len(r) == 0 {
			return math.NaN()
		}
		var downSq float64
		for _, x := range r {
			if x < 0 {
				downSq += x * x
			}
		}
		dd := math.Sqrt(downSq/float64(len(r))) * math.Sqrt(tradingDaysPerYear)
		if dd == 0 {
			return math.NaN()
		}
		return Mean(r) * tradingDaysPerYear / dd
	})
}

// RollingUlcer is the Ulcer Index (in percent points) of each trailing
// window of the given length in years.
func RollingUlcer(dates []time.Time, values []float64, years float64) ([]time.Time, []float64, bool) {
	return Rolling(dates, values, years, func(w []float64) float64 {
		dd := Drawdowns(w)
		var sumSq float64
		for _, d := range dd {
			sumSq += d * d * 10000
		}
		return math.Sqrt(sumSq / float64(len(dd)))
	})
}

// windowVol is the annualized standard deviation of a value window's daily
// returns, the building block of the rolling Sharpe and volatility wrappers.
func windowVol(window []float64) float64 {
	r := Returns(window)
	if len(r) < 2 {
		return 0
	}
	m := Mean(r)
	var variance float64
	for _, x := range r {
		variance += (x - m) * (x - m)
	}
	return math.Sqrt(variance/float64(len(r)-1)) * math.Sqrt(tradingDaysPerYear)
}
