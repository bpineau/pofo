package simgen

import (
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// The long zero-coupon Treasury reconstruction.
//
// A STRIPS fund is not a leveraged coupon fund, and the difference is the whole
// reason this file exists. The recipes that reconstruct a 20+ year COUPON ETF
// gear their donor fund over cash (longTreasuryGearing), which works because
// the two hold the same kind of bond and differ only by a little duration. A
// 25+ year STRIPS fund differs by KIND: it owns discount paper with no coupon,
// whose duration is its maturity whatever the rate level, while the coupon
// fund's duration collapses as yields rise. The gearing that matches them is
// therefore a function of the yield, not a constant:
//
//	long yield     22y par bond    27y strip    gearing
//	3 %            16.0            26.6         1.66
//	6 %            12.1            26.2         2.16
//	9 %             9.5            25.8         2.72
//	12 %            7.7            25.5         3.31
//
// (modified durations in years, semiannual convention). A single multiple
// fitted on the 2 to 4 % yields of the 2010s is right there and roughly HALF of
// what the 1980-1985 average of 12 % demands, which is how a backcast ends up
// reporting a 17 %/yr volatility and a 47 % worst drawdown for an instrument
// whose own arithmetic says 25 % and far deeper. The fix is not a better
// multiple, since no constant can span the two regimes: it is to price the
// strip off the long yield itself, which is what TreasuryZeroTR does and what
// the bundled TREASURY-LONG-YIELD series (cmd/gen-tyield-refdata) supplies back
// to 1953. See docs/long-treasury-zero-coupon-design.md.

// longTreasuryYieldID is the bundled long Treasury par-yield history every
// zero-coupon reconstruction is priced off: the Fed's H.15 30-year constant
// maturity, business daily from 1977 and carried back to 1953-04 by the
// 20-year point mapped onto it (monthly before 1962). Annualized percent
// levels, a rate and not a price.
const longTreasuryYieldID = "TREASURY-LONG-YIELD"

// zrozMaturity is the constant maturity (years) the ZROZ reconstruction holds.
// The fund tracks the 25+ year segment of the Treasury principal STRIPS market,
// some two dozen lines spread over the 25 to 30 year range, so the average
// maturity of what it owns sits near 27 years; at the yields of the live window
// that is a modified duration of about 26.5, against the 26.0 the fund itself
// publishes. Nothing here is fitted to the fund's returns: it is the maturity
// of the paper, and the validation reads the vol ratio it produces rather than
// setting it.
const zrozMaturity = 27.0

// zrozTER is ZROZ's ongoing charge (0.15 %/yr, FRACTION per year as everything
// in this package). The whole load is due: it is charged on a fee-free yield
// reconstruction, not on a donor fund whose NAV already carries a manager's
// charges.
const zrozTER = 0.0015

// zrozBuild reconstructs ZROZ as what it is, a constant-maturity long
// zero-coupon Treasury position priced off the long yield, from 1953-04.
func zrozBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	y, err := refdata(f, longTreasuryYieldID, from)
	if err != nil {
		return nil, err
	}
	s := TreasuryZeroTR("ZROZ (25+ year Treasury STRIPS, 27y constant maturity)", y, zrozMaturity, zrozTER)
	s.Currency = "USD"
	return s, nil
}
