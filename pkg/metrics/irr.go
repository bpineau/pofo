package metrics

import (
	"math"
	"time"
)

// IRR computes the money-weighted annual rate of return of an investment.
//
// flows are the investor's cash movements, dated: negative when money goes
// into the portfolio (initial capital, contributions), positive when it
// comes out (withdrawals). finalValue is the portfolio's worth at finalDate
// and counts as a terminal inflow. IRR returns the annual rate r solving
// NPV(r) = 0, found by bisection; ok is false when no rate in
// (-99 %, +1000 %) balances the flows (e.g. all flows the same sign).
//
// Units: the rate is a fraction per year (0.07 = 7 %/year). Years are
// counted as 365.25 days.
func IRR(dates []time.Time, flows []float64, finalDate time.Time, finalValue float64) (rate float64, ok bool) {
	if len(dates) != len(flows) || len(dates) == 0 {
		return 0, false
	}
	t0 := dates[0]
	years := func(t time.Time) float64 {
		return t.Sub(t0).Hours() / 24 / 365.25
	}
	npv := func(r float64) float64 {
		sum := 0.0
		for i, f := range flows {
			sum += f / math.Pow(1+r, years(dates[i]))
		}
		return sum + finalValue/math.Pow(1+r, years(finalDate))
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
