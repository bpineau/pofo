package metrics

import (
	"math"
	"slices"
	"time"
)

// IRR computes the money-weighted annual rate of return of an investment.
//
// flows are the investor's cash movements, dated: negative when money goes
// into the portfolio (initial capital, contributions), positive when it
// comes out (withdrawals). finalValue is the portfolio's worth at finalDate
// and counts as a terminal inflow. IRR returns the annual rate r solving
// NPV(r) = 0, found by bisection. It reports ok=false when the interval
// [-99 %, +1000 %] does not bracket a solution or fewer than two dates
// have a nonzero net cash movement.
//
// Units: the rate is a fraction per year (0.07 = 7 %/year). Years are
// counted as 365.25 days.
func IRR(dates []time.Time, flows []float64, finalDate time.Time, finalValue float64) (rate float64, ok bool) {
	if len(dates) != len(flows) || len(dates) == 0 {
		return 0, false
	}
	cash := make([]Flow, 0, len(flows)+1)
	for i, amount := range flows {
		cash = append(cash, Flow{Date: dates[i], Amount: amount})
	}
	cash = append(cash, Flow{Date: finalDate, Amount: finalValue})
	slices.SortStableFunc(cash, func(a, b Flow) int { return a.Date.Compare(b.Date) })
	net := cash[:0]
	for _, f := range cash {
		if n := len(net); n != 0 && net[n-1].Date.Equal(f.Date) {
			net[n-1].Amount += f.Amount
		} else {
			net = append(net, f)
		}
	}
	net = slices.DeleteFunc(net, func(f Flow) bool { return f.Amount == 0 })
	if len(net) < 2 {
		return 0, false
	}
	npv := func(r float64) float64 {
		// Moving the valuation date rescales every cash flow by the same
		// positive factor. Choose the end that only discounts amounts so
		// the low-rate bracket remains finite over a long history.
		at := net[0].Date
		if r < 0 {
			at = net[len(net)-1].Date
		}
		sum := 0.0
		for _, f := range net {
			years := at.Sub(f.Date).Hours() / 24 / 365.25
			sum += f.Amount * math.Pow(1+r, years)
		}
		return sum
	}
	lo, hi := -0.99, 10.0
	flo, fhi := npv(lo), npv(hi)
	if flo == 0 {
		return lo, true
	}
	if fhi == 0 {
		return hi, true
	}
	if (flo > 0) == (fhi > 0) {
		return 0, false
	}
	for i := 0; i < 200; i++ {
		mid := (lo + hi) / 2
		fm := npv(mid)
		if fm == 0 || hi-lo < 1e-10 {
			return mid, true
		}
		if (fm > 0) == (flo > 0) {
			lo, flo = mid, fm
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2, true
}
