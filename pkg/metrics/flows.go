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
	Date   time.Time
	Amount float64
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
// flows: daily returns r_t = (V_t - F_t) / V_{t-1} are chain-linked, so
// contributions and withdrawals are neutralized and the result measures the
// strategy, not the saver. Days with a non-positive base are skipped. ok is
// false when the series has fewer than two points or mismatched lengths.
func TWR(dates []time.Time, values []float64, flows []Flow) (float64, bool) {
	if len(dates) != len(values) || len(values) < 2 {
		return 0, false
	}
	byDay := flowsByDay(flows)
	total := 1.0
	for i := 1; i < len(values); i++ {
		if values[i-1] <= 0 {
			continue
		}
		total *= (values[i] - byDay[dates[i]]) / values[i-1]
	}
	return total - 1, true
}

// FlowReturns yields the flow-adjusted daily returns of a value series:
// (V_t - F_t)/V_{t-1} - 1. Saturday and Sunday points are dropped, so a
// calendar-daily series (weekends forward-filled flat) does not dilute its
// volatility; a trading-day series is unaffected. Days with a non-positive
// base are skipped. Feed the result to Volatility, Sharpe or Sortino.
func FlowReturns(dates []time.Time, values []float64, flows []Flow) []float64 {
	if len(dates) != len(values) {
		return nil
	}
	byDay := flowsByDay(flows)
	var out []float64
	for i := 1; i < len(values); i++ {
		if values[i-1] <= 0 {
			continue
		}
		if wd := dates[i].Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		out = append(out, (values[i]-byDay[dates[i]])/values[i-1]-1)
	}
	return out
}

// Volatility is the annualized sample standard deviation of daily returns
// (252 trading days), the same figure Compute reports. NaN for fewer than
// two returns.
func Volatility(returns []float64) float64 {
	if len(returns) < 2 {
		return math.NaN()
	}
	m := Mean(returns)
	ss := 0.0
	for _, r := range returns {
		ss += (r - m) * (r - m)
	}
	return math.Sqrt(ss/float64(len(returns)-1)) * math.Sqrt(tradingDaysPerYear)
}

// Sharpe is the annualized mean daily excess return over rfAnnual divided by
// Volatility: the same arithmetic-annualization convention as Compute, which
// fixes rfAnnual at zero. NaN when the volatility is zero or undefined.
func Sharpe(returns []float64, rfAnnual float64) float64 {
	v := Volatility(returns)
	if !(v > 0) {
		return math.NaN()
	}
	return (Mean(returns)*tradingDaysPerYear - rfAnnual) / v
}

// Sortino replaces Sharpe's denominator with the downside deviation against
// the daily risk-free target rfAnnual/252. NaN when there is no downside or
// no return at all.
func Sortino(returns []float64, rfAnnual float64) float64 {
	if len(returns) == 0 {
		return math.NaN()
	}
	target := rfAnnual / tradingDaysPerYear
	ss := 0.0
	for _, r := range returns {
		if r < target {
			ss += (r - target) * (r - target)
		}
	}
	down := math.Sqrt(ss/float64(len(returns))) * math.Sqrt(tradingDaysPerYear)
	if !(down > 0) {
		return math.NaN()
	}
	return (Mean(returns)*tradingDaysPerYear - rfAnnual) / down
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
