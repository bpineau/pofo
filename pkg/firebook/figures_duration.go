package firebook

import (
	"fmt"
	"math"
	"strings"
)

// This file holds the duration plate of allocation-actions-obligations: the
// remaining duration of the article's three bond vehicles as the years pass.
// Everything is closed-form arithmetic on a par-market hypothesis, computed at
// render time from the constants below, so there is nothing frozen that could
// drift; figures_duration_test.go recomputes the same durations from the
// textbook Macaulay formula and checks the three end conditions.

const (
	// durParYield is the flat yield the plate prices every vehicle at, and
	// durParCoupon the annual coupon of the coupon-bearing ones. Taking them
	// equal prices the bonds at par, which is the cleanest hypothesis for
	// comparing vehicles: only the cash-flow calendar then differs.
	durParYield  = 0.03
	durParCoupon = 0.03
	// The three vehicles, with the maturities the article's prose names.
	durRollLo    = 7.0  // the rolling ETF's tranche: "7-10 ans"
	durRollHi    = 10.0 // ... which it never leaves
	durDatedLife = 8.0  // the dated-maturity fund liquidates after 8 years
	durZeroLife  = 12.0 // the zero-coupon funds a spending date 12 years out
)

// macaulayDuration returns the Macaulay duration, in years, of a bond with n
// years left, an annual coupon c (a fraction of the face) and yield y: each
// remaining cash flow weighted by its share of the present value. Coupons fall
// on the anniversaries counted back from redemption, so the function is
// continuous in n, decays to 0 as n does, and returns exactly n for a zero
// coupon (c = 0), whose only flow is the redemption.
func macaulayDuration(n, c, y float64) float64 {
	if n <= 0 {
		return 0
	}
	var pv, weighted float64
	for i := 0; ; i++ {
		t := n - float64(i)
		if t <= 1e-9 {
			break
		}
		cf := c
		if i == 0 {
			cf += 1 // the face, redeemed with the last coupon
		}
		d := cf / math.Pow(1+y, t)
		pv += d
		weighted += t * d
	}
	return weighted / pv
}

// durRollingDuration is the constant duration of the rolling ETF: the average
// duration of its maturity tranche, which it holds forever by selling the
// bonds that fall out of the low end and buying at the high end.
func durRollingDuration() float64 {
	var sum float64
	var n int
	for m := durRollLo; m <= durRollHi+1e-9; m++ {
		sum += macaulayDuration(m, durParCoupon, durParYield)
		n++
	}
	return sum / float64(n)
}

// --- Bond vehicles: what each one's duration does with the years ---
func figDurationVehicules() string {
	m := mapper(0, 12, 0, 12, 76, 500, 316, 78)
	var b strings.Builder
	b.WriteString(plateHead("le véhicule change la donne", "Trois véhicules, trois destins de duration"))
	// horizontal grid + mono ticks
	for _, g := range []float64{0, 3, 6, 9, 12} {
		gy := m(0, g)[1]
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(76, gy, 500, gy, col, 1))
		b.WriteString(mTxt(66, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(76, 64, 10.5, figMuted, "start", "400", "duration restante (années)"))
	for _, t := range []float64{0, 2, 4, 6, 8, 10, 12} {
		p := m(t, 0)
		b.WriteString(mTxt(p[0], 332, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt(288, 352, 11, figMuted, "middle", "400", "années de détention"))

	// The two vehicles that do reach zero, each ending on a dot. Duration is
	// read once a year, on the anniversary just after the coupon: sampling it
	// continuously would add the intra-year sawtooth (duration ticks back up
	// the day a coupon leaves the bond), which is real but unreadable here.
	series := func(life, coupon float64) [][2]float64 {
		var pts [][2]float64
		for t := 0.0; t <= life+1e-9; t++ {
			pts = append(pts, m(t, macaulayDuration(life-t, coupon, durParYield)))
		}
		return pts
	}
	end := func(pts [][2]float64, fill string) {
		p := pts[len(pts)-1]
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, p[0], p[1], fill)
	}
	zero := series(durZeroLife, 0)
	dated := series(durDatedLife, durParCoupon)
	b.WriteString(poly(zero, figBad, 2.2, ""))
	b.WriteString(poly(dated, figBlue, 2.2, ""))
	end(zero, figBad)
	end(dated, figBlue)

	// the rolling ETF, flat, and running off the plate: it has no maturity
	roll := durRollingDuration()
	flat := m(0, roll)[1]
	b.WriteString(line(76, flat, 500, flat, figAccent, 2.8))
	b.WriteString(mTxt(66, flat+3.5, 10, figDeep, "end", "600",
		strings.Replace(fmt.Sprintf("%.1f", roll), ".", ",", 1)))
	b.WriteString(dashLine(500, flat, 598, flat, figAccent, 2.2, "6 5"))
	fmt.Fprintf(&b, `<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		600.0, flat-5.5, 614.0, flat, 600.0, flat+5.5, figAccent)

	// direct labels, no legend
	b.WriteString(sTxt(598, flat-14, 11.5, figDeep, "end", "600", "ETF roulant « 7-10 ans »"))
	b.WriteString(sTxt(598, flat+23, 10.5, figMuted, "end", "400", "n'arrive jamais à échéance"))
	b.WriteString(sTxt(282, 293, 11.5, figBlue, "end", "600", "Fonds à échéance 8 ans"))
	b.WriteString(sTxt(508, 320, 11.5, figBad, "start", "600", "Zéro-coupon 12 ans"))
	return svg(640, 372, b.String())
}
