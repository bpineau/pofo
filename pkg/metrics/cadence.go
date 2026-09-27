package metrics

import (
	"math"
	"time"
)

// TradingDaysPerYear is the cadence of a series of daily market closes, the
// count PeriodsPerYear returns for any calendar of trading days (exchange
// holidays and the union calendar of several markets included). Pass it where
// returns are daily by construction and no dates travel with them.
const TradingDaysPerYear = 252

// cadences are the canonical observation counts per year a measured cadence
// snaps to: yearly, semi-annual, quarterly, monthly, semi-monthly, fortnightly,
// weekly, trading days, and every calendar day (a union calendar that quotes
// weekends, crypto, a forward-filled daily ledger).
var cadences = [...]float64{1, 2, 4, 12, 24, 26, 52, TradingDaysPerYear, 365}

// snapTolerance is how far, as a ratio, a measured cadence may sit from a
// canonical one and still snap to it: wide enough to absorb holidays, a
// missing quote or a short month, narrow enough that a genuinely unusual
// cadence (a six-day trading week, a bimonthly NAV) keeps its own count.
const snapTolerance = 1.10

// spacings are the upper bounds, in calendar days, of the regimes a gap
// between two observations can belong to: daily (up to four days, so a
// weekend or a holiday stays daily), weekly, fortnightly or semi-monthly,
// monthly, quarterly, semi-annual and yearly.
var spacings = [...]float64{4.5, 10.5, 20.5, 45.5, 135.5, 270.5, math.Inf(1)}

// PeriodsPerYear is the cadence of a dated series: how many observations it
// holds per year, the factor that annualizes its per-period statistics
// (volatility scales by its square root, a mean return by the count itself).
//
// It measures the density of the series' PREVAILING spacing. Each gap between
// consecutive dates falls in a regime (daily, up to four calendar days so a
// weekend or a holiday stays daily; weekly; fortnightly or semi-monthly;
// monthly; quarterly; semi-annual; yearly), the regime most returns belong to
// wins, and its observations are counted against the calendar days they span.
// A long hole in the series (a suspended fund, a donor chain's seam) therefore
// cannot dilute the count, and neither can a sparse deep history in front of
// a dense recent one. The measurement then snaps to the nearest canonical
// cadence (1, 2, 4, 12, 24, 26, 52, 252 or 365 a year) when it lies within
// 10 % of it, and is returned as measured otherwise. Every trading-day
// calendar reads exactly TradingDaysPerYear, so a daily statistic is the one a
// fixed 252 would give; a monthly series reads 12, a weekly NAV 52, a series
// quoting every calendar day 365.
//
// A series that CHANGES cadence (a fund that published weekly, then daily)
// has no single right answer: its per-period returns measure different
// horizons, and whichever count annualizes them misreads the other part. The
// prevailing regime wins here; resample such a series to one cadence first
// (marketdata.Series.Resample) when its statistics matter.
//
// dates must be ascending. It returns NaN for fewer than two dates or no
// positive spacing between them.
func PeriodsPerYear(dates []time.Time) float64 {
	if len(dates) < 2 {
		return math.NaN()
	}
	gaps := make([]float64, len(dates)-1)
	for i := 1; i < len(dates); i++ {
		gaps[i-1] = dates[i].Sub(dates[i-1]).Hours() / 24
	}
	return cadence(gaps)
}

// cadence is PeriodsPerYear over the calendar-day spans of a set of returns,
// consecutive or not: the pairing behind Beta and VsBenchmark measures each
// return over its own interval, and those intervals are what it annualizes.
func cadence(gaps []float64) float64 {
	var n [len(spacings)]int
	var days [len(spacings)]float64
	for _, g := range gaps {
		if !(g > 0) || math.IsInf(g, 1) {
			continue // a non-positive or endless span measures nothing
		}
		k := 0
		for g > spacings[k] {
			k++
		}
		n[k]++
		days[k] += g
	}
	mode := 0
	for k := range n {
		if n[k] > n[mode] {
			mode = k
		}
	}
	if n[mode] == 0 {
		return math.NaN()
	}
	return snap(float64(n[mode]) / days[mode] * daysPerYear)
}

// snap returns the canonical cadence nearest to a measured one when it lies
// within snapTolerance of it, and the measurement itself otherwise.
func snap(measured float64) float64 {
	best, dist := measured, math.Log(snapTolerance)
	for _, c := range cadences {
		if d := math.Abs(math.Log(measured / c)); d <= dist {
			best, dist = c, d
		}
	}
	return best
}

// cadence is the cadence of date-paired returns, each measured over its own
// interval, from start[k] to end[k].
func (p paired) cadence() float64 {
	gaps := make([]float64, len(p.end))
	for k := range p.end {
		gaps[k] = p.end[k].Sub(p.start[k]).Hours() / 24
	}
	return cadence(gaps)
}
