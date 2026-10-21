package metrics

import (
	"math"
	"time"
)

// Flow is an external cash movement into (positive) or out of (negative) the
// measured scope. The flows of a day are netted, and the net is booked by its
// direction: money coming in at the start of the day (it earns that day),
// money going out at its close (what leaves is valued at the day's close, so
// it earned the day before leaving); see FlowIndex. Dates must carry the same
// normalization as the value series they accompany (pofo series use
// 00:00 UTC): flows are matched to days by exact time.Time equality.
type Flow struct {
	Date   time.Time // the day the flow is booked (00:00 UTC for pofo series)
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

// FlowIndex chain-links the flow-neutral daily growth of a value series with
// external flows into an index that starts at 1: what one unit held from the
// first point became, contributions and withdrawals neutralized. It is the
// series every statistic of such a scope reads (TWR is its last point less
// one, Report's drawdown is measured on it), because the value itself falls
// by the whole amount of a withdrawal or a sale, which is money leaving the
// scope, not money lost. nil when the lengths differ or the series is empty.
//
// The flows of a day are netted and the net is booked by its direction.
// Money coming in is booked at the start of the day, r_t = V_t/(V_{t-1}+F_t)
// - 1: it earns that day, and a large contribution never has its first-day
// move charged to a tiny pre-flow value. Money going out is booked at the
// close, r_t = (V_t - F_t)/V_{t-1} - 1 with F_t negative: a sale is valued at
// the day's close, so the shares earned the day before they left. Booked at
// the start instead, the outflow would be subtracted from a previous close it
// never faced, and selling a whole line on a down day would read as a -100 %
// day that zeroes everything after it.
//
// A day with nothing to earn on, a non-positive base or a scope that holds
// nothing before nor after (a fee posted once the account is empty), keeps
// the index flat and is no return at all for FlowReturns. Between two flows
// the index is the value times a constant scale rather than a running
// product, so a flowless stretch reads exactly the value's ratios and an
// exact return to a peak stays exact.
func FlowIndex(dates []time.Time, values []float64, flows []Flow) []float64 {
	index, _ := flowGrowth(dates, values, flows)
	return index
}

// flowGrowth is FlowIndex plus, per point, whether that day's growth was
// measured (live) or held flat for want of a base. live[0] is false.
func flowGrowth(dates []time.Time, values []float64, flows []Flow) (index []float64, live []bool) {
	if len(dates) != len(values) || len(values) == 0 {
		return nil, nil
	}
	byDay := flowsByDay(flows)
	index = make([]float64, len(values))
	live = make([]bool, len(values))
	index[0] = 1
	scale := 0.0 // index per unit of value; 0 until a positive value sets it
	if values[0] > 0 {
		scale = 1 / values[0]
	}
	for i := 1; i < len(values); i++ {
		v, prev, flow := values[i], values[i-1], byDay[dates[i]]
		if flow == 0 && scale > 0 && v > 0 && prev > 0 {
			index[i], live[i] = scale*v, true
			continue // the scale holds until the next flow
		}
		index[i] = index[i-1]
		num, base := v, prev+flow // money in: at the start of the day
		if flow < 0 {
			num, base = v-flow, prev // money out: at the close
		}
		if base > 0 && (prev > 0 || v > 0) {
			index[i] *= num / base
			live[i] = true
		}
		scale = 0
		if v > 0 {
			scale = index[i] / v
		}
	}
	return index, live
}

// TWR is the time-weighted total return of a value series with external
// flows: the last point of FlowIndex less one, so contributions and
// withdrawals are neutralized and the result measures the strategy, not the
// saver (inflows booked at the start of their day, outflows at its close, a
// day without a base skipped; see FlowIndex). ok is false when the series has
// fewer than two points or mismatched lengths.
func TWR(dates []time.Time, values []float64, flows []Flow) (float64, bool) {
	if len(dates) != len(values) || len(values) < 2 {
		return 0, false
	}
	index := FlowIndex(dates, values, flows)
	return index[len(index)-1] - 1, true
}

// FlowReturns yields the flow-adjusted returns of a value series, read on
// its FlowIndex (inflows at the start of the day, outflows at the close),
// and their cadence, ready for Volatility, Sharpe or Sortino. Saturday and
// Sunday points are folded into the next weekday's return, so a
// calendar-daily series (weekends forward-filled flat) does not dilute its
// volatility and reads as the trading-day series it is, while a series that
// really moves at the weekend keeps that move; a trading-day series is
// unaffected. A day without a base (see FlowIndex) yields no return.
// periodsPerYear is NaN when no return survives.
func FlowReturns(dates []time.Time, values []float64, flows []Flow) (returns []float64, periodsPerYear float64) {
	index, live := flowGrowth(dates, values, flows)
	if index == nil {
		return nil, math.NaN()
	}
	var spans []float64
	for i, last := 1, 0; i < len(index); i++ {
		if wd := dates[i].Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		// A return runs from the last weekday point, so a weekend folded
		// into Monday counts as the one trading period it is.
		span := dates[i].Sub(dates[last]).Hours() / 24
		from := last
		last = i
		if !live[i] || !(index[from] > 0) {
			continue
		}
		returns = append(returns, index[i]/index[from]-1)
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
