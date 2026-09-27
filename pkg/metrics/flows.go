package metrics

import (
	"math"
	"time"
)

// Flow is an external cash movement into (positive) or out of (negative) the
// measured scope, booked at the start of its day. Dates must carry the same
// normalization as the value series they accompany (pofo series use
// 00:00 UTC): flows are matched to days by exact time.Time equality.
type Flow struct {
	Date   time.Time // the day the flow is booked, at its start (00:00 UTC for pofo series)
	Amount float64   // in the value series' unit: positive in, negative out
}

// flowsByDay sums flow amounts per exact date.
func flowsByDay(flows []Flow) map[time.Time]float64 {
	if len(flows) == 0 {
		return nil
	}
	byDay := make(map[time.Time]float64, len(flows))
	for _, f := range flows {
		byDay[f.Date] += f.Amount
	}
	return byDay
}

// TWR is the time-weighted total return of a value series with external
// flows: daily returns r_t = V_t / (V_{t-1} + F_t) - 1 are chain-linked, so
// contributions and withdrawals are neutralized and the result measures the
// strategy, not the saver. Flows are booked at the start of their day, so a
// same-day contribution earns that day and belongs in the return's base;
// dividing by V_{t-1} alone would charge a large flow's first-day P/L against
// the tiny pre-flow value and detonate the chain. Days with a non-positive
// base (V_{t-1} + F_t) are skipped. ok is false when the series has fewer than
// two points or mismatched lengths.
func TWR(dates []time.Time, values []float64, flows []Flow) (float64, bool) {
	if len(dates) != len(values) || len(values) < 2 {
		return 0, false
	}
	byDay := flowsByDay(flows)
	total := 1.0
	for i := 1; i < len(values); i++ {
		base := values[i-1] + byDay[dates[i]]
		if base <= 0 {
			continue
		}
		total *= values[i] / base
	}
	return total - 1, true
}

// FlowReturns yields the flow-adjusted returns of a value series,
// V_t/(V_{t-1} + F_t) - 1 (the same start-of-day flow convention as TWR),
// and their cadence, ready for Volatility, Sharpe or Sortino.
// Saturday and Sunday points are dropped, so a calendar-daily series
// (weekends forward-filled flat) does not dilute its volatility and reads as
// the trading-day series it is; a trading-day series is unaffected. Days with
// a non-positive base (V_{t-1} + F_t) are skipped. periodsPerYear is NaN when
// no return survives.
func FlowReturns(dates []time.Time, values []float64, flows []Flow) (returns []float64, periodsPerYear float64) {
	if len(dates) != len(values) {
		return nil, math.NaN()
	}
	byDay := flowsByDay(flows)
	var spans []float64
	for i, last := 1, 0; i < len(values); i++ {
		base := values[i-1] + byDay[dates[i]]
		if wd := dates[i].Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		// A return runs from the last weekday point, so a weekend folded
		// into Monday counts as the one trading period it is.
		span := dates[i].Sub(dates[last]).Hours() / 24
		last = i
		if base <= 0 {
			continue
		}
		returns = append(returns, values[i]/base-1)
		spans = append(spans, span)
	}
	return returns, cadence(spans)
}

// Volatility is the sample standard deviation of per-period returns,
// annualized at periodsPerYear (TradingDaysPerYear for daily returns,
// PeriodsPerYear of their dates in general): the figure Compute reports as
// Stats.Volatility. NaN for fewer than two returns.
func Volatility(returns []float64, periodsPerYear float64) float64 {
	if len(returns) < 2 {
		return math.NaN()
	}
	m := Mean(returns)
	ss := 0.0
	for _, r := range returns {
		ss += (r - m) * (r - m)
	}
	return math.Sqrt(ss/float64(len(returns)-1)) * math.Sqrt(periodsPerYear)
}

// Sharpe is the mean per-period excess return over rfAnnual, annualized at
// periodsPerYear, divided by Volatility: the same arithmetic-annualization
// convention as Compute, which fixes rfAnnual at zero. NaN when the
// volatility is zero or undefined.
func Sharpe(returns []float64, rfAnnual, periodsPerYear float64) float64 {
	v := Volatility(returns, periodsPerYear)
	if !(v > 0) {
		return math.NaN()
	}
	return (Mean(returns)*periodsPerYear - rfAnnual) / v
}

// Sortino replaces Sharpe's denominator with the downside deviation against
// the per-period risk-free target rfAnnual/periodsPerYear, annualized at
// periodsPerYear. NaN when there is no downside or no return at all.
func Sortino(returns []float64, rfAnnual, periodsPerYear float64) float64 {
	if len(returns) == 0 {
		return math.NaN()
	}
	target := rfAnnual / periodsPerYear
	ss := 0.0
	for _, r := range returns {
		if r < target {
			ss += (r - target) * (r - target)
		}
	}
	down := math.Sqrt(ss/float64(len(returns))) * math.Sqrt(periodsPerYear)
	if !(down > 0) {
		return math.NaN()
	}
	return (Mean(returns)*periodsPerYear - rfAnnual) / down
}

// Annualize converts a cumulative return earned over a calendar-day span
// into a compound annual rate ((1+total)^(365.25/days) - 1). It returns 0
// when days is not positive or the capital was wiped out (total <= -1);
// annualizing sub-year spans amplifies noise, so gate short windows before
// calling (see the FIRE and finador reports for the customary thresholds).
func Annualize(totalReturn float64, days int) float64 {
	if days <= 0 || totalReturn <= -1 {
		return 0
	}
	return math.Pow(1+totalReturn, daysPerYear/float64(days)) - 1
}
