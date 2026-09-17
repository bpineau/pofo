package simgen

import (
	"math"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// bondPrice returns the clean price (per 100 face) of an annual-coupon bond
// with coupon rate c and n years to maturity, discounted at yield y. y, c are
// decimals and n is non-negative; y may be negative, as government bond yields
// have been for years at a time in Germany and in Japan, and only has to stay
// above -100 % for the discount factor to exist. A bond whose coupon equals its
// yield prices at exactly 100 for any maturity (a par bond), at a negative
// yield as at a positive one.
//
// Zero is the annuity factor's one singular point, and its limit there is the
// undiscounted sum of the coupons plus the face: an epsilon guard rather than a
// yield floor, so a day at exactly 0.000 % prices instead of being skipped.
func bondPrice(y, c, n float64) float64 {
	if math.Abs(y) < 1e-12 {
		return 100 * (c*n + 1)
	}
	d := math.Pow(1+y, -n)
	return 100 * (c*(1-d)/y + d)
}

// TreasuryTR reconstructs the total-return index of a constant-maturity
// Treasury bond from a yield series, so a treasury sleeve reaches back as far
// as the yield history (FRED constant-maturity series run to 1953) rather than
// its fund's inception.
//
// Each step it holds a fresh maturityYears par bond (coupon set to the
// prevailing yield, so it starts at par), lets it age by the step's length,
// accrues the coupon, and reprices the now slightly shorter bond at the next
// yield: total return = (repriced price + accrued coupon)/100 − 1, minus a
// continuous annualFee. Repricing is exact (it captures convexity without a
// Taylor term). Yields are read as annualised percent (FRED's convention) and
// may be sampled at any cadence; each step uses its own year fraction. The
// index starts at 100 on the first yield date.
//
// It models annual coupons (Treasuries pay semiannual); the small difference is
// second-order and is absorbed when the series is rescaled at its splice point
// and validated against the real fund on their overlap.
//
// A NEGATIVE yield is priced, not skipped. Japan's 10-year benchmark closed
// below zero on 453 days between 2016-02 and 2020-05 and the German and
// euro-area curves spent years there, all of them years in which those bonds
// returned a great deal: treating a sub-zero yield as missing data would
// flat-line the reconstruction through the largest rally in its record. Only a
// yield at or below -100 %/yr is refused, since the discount factor stops
// existing there.
func TreasuryTR(name string, yields *marketdata.Series, maturityYears, annualFee float64) *marketdata.Series {
	return constantMaturityTR(name, yields, annualFee, func(y0, y1, dt float64) float64 {
		if y0 <= -1 || y1 <= -1 {
			return math.NaN()
		}
		return (bondPrice(y1, y0, maturityYears-dt)+100*y0*dt)/100 - 1
	})
}

// stripPrice returns the price (per 1 of face) of a zero-coupon bond with n
// years to maturity at the SEMIANNUAL bond-equivalent yield y, the convention
// every Treasury yield is quoted in: y/2 is compounded twice a year. n is
// non-negative and y only has to stay above -200 %/yr for the discount factor
// to exist.
//
// The semiannual convention is not cosmetic at this maturity. A 27-year zero
// discounted annually instead has a modified duration of n/(1+y), not
// n/(1+y/2): 24.1 rather than 25.5 years at a 12 % yield, a 5 % error in the
// one quantity the series exists to carry.
func stripPrice(y, n float64) float64 {
	return math.Pow(1+y/2, -2*n)
}

// TreasuryZeroTR reconstructs the total-return index of a constant-maturity
// ZERO-COUPON Treasury position (a STRIPS ladder held at a fixed point of the
// curve) from a yield series, the way TreasuryTR does for a par coupon bond.
//
// A STRIP pays nothing until it matures, so its whole return is the repricing
// of one discount factor: each step holds a fresh maturityYears zero, lets it
// age by the step's length and reprices it at the next yield, total return =
// P(y1, T − dt)/P(y0, T) − 1, minus a continuous annualFee. Carry comes out as
// the pull to par, so there is no coupon term and nothing to reinvest. Yields
// are read as annualised percent at any cadence and the index starts at 100 on
// the first yield date.
//
// This is a SEPARATE engine from TreasuryTR rather than a maturity setting on
// it, because the two answer a rate move differently and increasingly so as
// rates rise. A coupon bond's duration SHRINKS when its yield rises (the early
// coupons weigh more), a zero's does not: at a 3 % yield a 22-year par bond and
// a 27-year zero have modified durations of 16.0 and 26.6, a ratio of 1.66, and
// at 12 % of 7.7 and 25.5, a ratio of 3.31. Any fixed multiple of a coupon fund
// is therefore right in one rate regime and wrong by a factor of two in the
// other, which is why a long-STRIPS fund is reconstructed here from the yield
// itself.
//
// Two approximations are worth naming, both of them level rather than path
// effects. The yield fed in is a PAR yield, while the strip is discounted at a
// zero rate: on an upward-sloping curve the zero rate sits ABOVE the par yield
// of the same maturity, so the reconstruction understates the carry a little
// (see docs/long-treasury-zero-coupon-design.md for the measured sign and
// size). And the constant-maturity convention reprices the aged strip at its
// ORIGINAL curve point, not at the slightly shorter one it has rolled down to,
// so the roll-down gain of a positively-sloped curve is left out too. Both push
// the same way, and both are small next to a 20 %/yr volatility.
func TreasuryZeroTR(name string, yields *marketdata.Series, maturityYears, annualFee float64) *marketdata.Series {
	return constantMaturityTR(name, yields, annualFee, func(y0, y1, dt float64) float64 {
		if y0 <= -2 || y1 <= -2 {
			return math.NaN()
		}
		return stripPrice(y1, maturityYears-dt)/stripPrice(y0, maturityYears) - 1
	})
}

// constantMaturityTR is the loop both reconstructions share: it walks the yield
// series, asks step for the total return of each period from the yields at its
// ends and its length in years, and compounds that net of a continuous fee. A
// period that step cannot price (an unpriceable yield, reported as NaN) carries
// the index forward unchanged, fee and all, rather than dropping the date: a
// period nobody can price is not one to charge for.
func constantMaturityTR(name string, yields *marketdata.Series, annualFee float64, step func(y0, y1, dt float64) float64) *marketdata.Series {
	s := &marketdata.Series{Name: name, Source: "simdata"}
	pts := yields.Points
	if len(pts) < 2 {
		return s
	}
	val := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: pts[0].Date, Close: val})
	for i := 1; i < len(pts); i++ {
		dt := pts[i].Date.Sub(pts[i-1].Date).Hours() / 24 / 365.25
		if r := step(pts[i-1].Close/100, pts[i].Close/100, dt); !math.IsNaN(r) {
			val *= 1 + r - annualFee*dt
		}
		s.Points = append(s.Points, marketdata.Point{Date: pts[i].Date, Close: val})
	}
	return s
}
