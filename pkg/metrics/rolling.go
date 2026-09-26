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

// RollingBeta is Beta over every trailing calendar window of the given length
// in years: the slope of the series' simple returns on the benchmark's, both
// as fractions, paired by date as Beta pairs them (each return from its own
// series' previous point, on the dates both quote). The windows are
// RollingVol's: one per paired date, ending there and holding the returns
// measured from a point inside (end-years, end], emitted only once the pairs
// reach back the full length, so on a shared calendar the dates are exactly
// RollingVol's.
//
// Unlike Beta, no window needs 30 pairs (the window length is the caller's
// choice, and a one-year window of monthly data holds 12): a window with
// fewer than two pairs or a flat benchmark reads NaN. ok is false when the
// slices are mismatched, years is not positive, or no full window fits.
func RollingBeta(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64, years float64) ([]time.Time, []float64, bool) {
	return rollingPaired(pairReturns(dates, values, benchDates, benchValues), years, func(own, bench []float64) float64 {
		return slope(bench, own)
	})
}

// RollingCorr is Corr over every trailing calendar window of the given length
// in years, on the date-paired returns and the windows RollingBeta uses: a
// number in [-1, 1] per window, 0 where either side is flat in the window,
// NaN where the window holds fewer than two pairs. ok is false when the
// slices are mismatched, years is not positive, or no full window fits.
func RollingCorr(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64, years float64) ([]time.Time, []float64, bool) {
	return rollingPaired(pairReturns(dates, values, benchDates, benchValues), years, func(own, bench []float64) float64 {
		if len(own) < 2 {
			return math.NaN()
		}
		return Corr(own, bench)
	})
}

// rollingPaired is Rolling for date-paired returns. The window ending at the
// j-th pair's date holds the pairs measured from a point after end-years, the
// returns Rolling's level window (end-years, end] yields, and is emitted only
// when the first pair is measured from a point at or before end-years.
func rollingPaired(p paired, years float64, fn func(own, bench []float64) float64) (points []time.Time, out []float64, ok bool) {
	if years <= 0 || len(p.own) == 0 {
		return nil, nil, false
	}
	span := time.Duration(years * 365.25 * 24 * float64(time.Hour))
	i := 0
	for j, end := range p.end {
		lo := end.Add(-span)
		if p.start[0].After(lo) {
			continue // the pairs do not yet cover a full window ending at j
		}
		for i <= j && !p.start[i].After(lo) {
			i++
		}
		points = append(points, end)
		out = append(out, fn(p.own[i:j+1], p.bench[i:j+1]))
	}
	if len(out) == 0 {
		return nil, nil, false
	}
	return points, out, true
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
