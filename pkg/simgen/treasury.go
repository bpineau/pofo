package simgen

import (
	"math"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// bondPrice returns the clean price (per 100 face) of an annual-coupon bond
// with coupon rate c and n years to maturity, discounted at yield y. y, c are
// decimals; y must be positive and n non-negative. A bond whose coupon equals
// its yield prices at exactly 100 for any maturity (a par bond).
func bondPrice(y, c, n float64) float64 {
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
func TreasuryTR(name string, yields *marketdata.Series, maturityYears, annualFee float64) *marketdata.Series {
	s := &marketdata.Series{Name: name, Source: "simdata"}
	pts := yields.Points
	if len(pts) < 2 {
		return s
	}
	val := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: pts[0].Date, Close: val})
	for i := 1; i < len(pts); i++ {
		y0, y1 := pts[i-1].Close/100, pts[i].Close/100
		if y0 <= 0 || y1 <= 0 {
			s.Points = append(s.Points, marketdata.Point{Date: pts[i].Date, Close: val})
			continue
		}
		dt := pts[i].Date.Sub(pts[i-1].Date).Hours() / 24 / 365.25
		r := (bondPrice(y1, y0, maturityYears-dt)+100*y0*dt)/100 - 1
		val *= 1 + r - annualFee*dt
		s.Points = append(s.Points, marketdata.Point{Date: pts[i].Date, Close: val})
	}
	return s
}
