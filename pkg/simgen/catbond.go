package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// This file builds the insurance-linked-securities (ILS, "cat bond") family:
// the reference index itself, its EUR-hedged view, and the backcast of the one
// UCITS share class a euro household can actually hold.
//
// THE REFERENCE. ilsAnchorID is the monthly ILS Advisers Fund Index served as
// refdata (2006-01→, cmd/gen-catbond-refdata): an equally weighted composite
// of the REAL funds investing in non-life insurance-linked securities, each
// already net of its own manager's fees. It is the same kind of object
// TREND-NET-USD is for managed futures, and it is used the same way: as the
// only anchor, never rescaled when it is served as a benchmark, rescaled to
// the target's own volatility when it stands in for a specific fund.
//
// CADENCE. The index is monthly and nothing here invents a daily texture for
// it. There is no ILS series with daily granularity to borrow one from, and a
// single day of cat bond return is not a path a reader could look up anyway:
// the asset is priced weekly at best, and its losses arrive as one step when an
// event happens. So the index return always lands on a month end, and the only
// reason the hedged file carries a daily calendar at all is that its two cash
// legs accrue every day (hedgeToEUR, as in BTOP50E). Annualizing such a file
// per observation happens to be right, since twelve monthly steps spread over
// 252 days annualize to the same figure the monthly returns do; annualizing the
// funds' own WEEKLY quotes per observation does not, and their statistics must
// be read monthly.
//
// RELIABILITY BOUNDS LENGTH. Nothing reaches before 2006-01, the index's first
// month, even though cat bonds have traded since 1997 and the Swiss Re market
// index starts in 2002: no public monthly record of what an ILS FUND paid its
// investors reaches further, and the market index is a different object (gross
// of every fund fee, and 2002-2005 would have to be spliced on an estimate).
// Hurricane Katrina therefore sits outside every file here, which is stated in
// the catalog notes and in docs/catbond-sleeve-design.md rather than papered
// over.

// ilsAnchorID is the monthly NET ILS fund composite (refdata, 2006-01→).
const ilsAnchorID = "ILS-NET-USD"

// ilsIndexLocal is the index in its own currency: the anchor, served as it
// stands. No fee is added (the constituents' fees are already in it) and no
// volatility rescaling is applied, exactly as btop50Local serves BTOP50.
func ilsIndexLocal(f Fetcher, from time.Time) (*marketdata.Series, error) {
	s, err := f.Fetch(ilsAnchorID, from)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ilsAnchorID, err)
	}
	if s == nil || len(s.Points) < 120 {
		return nil, fmt.Errorf("%s: refdata missing or too short, no fallback (an index of ILS funds cannot be replicated from prices)", ilsAnchorID)
	}
	out := *s
	out.Name, out.Source, out.Currency = "ILS funds, monthly net composite (index)", "simdata", "USD"
	return &out, nil
}

// ilsIndexRecipe serves the ILS reference AS ITSELF, the way BTOP50 serves the
// managed-futures one: a non-investable benchmark a book can be tested through,
// at its own level and its own ~3.4 %/yr volatility, with no fund fee added on
// top of the constituents' own.
func ilsIndexRecipe() Recipe {
	return Recipe{
		ID:     "ILSFUND",
		Name:   "ILS Advisers Fund Index (index, net of manager fees)",
		Method: "the monthly ILS fund composite (ILS-NET-USD refdata, 2006-01→) served as it stands: no fund fee added, no volatility rescaling, month-end cadence throughout",
		Build:  ilsIndexLocal,
	}
}

// ilsHedgedIndexRecipe is the same index expressed in EUR with the currency
// risk hedged away, which is the form a euro book actually compares against:
// every cat bond share class a European household can buy is EUR-hedged, and
// holding the raw USD index instead would put ~9 %/yr of EURUSD volatility
// into a sleeve whose whole volatility is 3.4 %.
//
// The arithmetic is the standard hedged-return identity BTOP50E uses. It is
// unusually exact here: a catastrophe bond is a floating-rate note, its
// collateral earns the money-market rate and its coupon is that rate plus a
// spread, so the index IS a funded total return and hedging really does swap
// the dollar cash leg for the euro one, which is what the fund's rolling
// forwards do.
func ilsHedgedIndexRecipe() Recipe {
	return Recipe{
		ID:     "ILSFUNDE",
		Name:   "ILS Advisers Fund Index, hedged to EUR (index)",
		Method: "the ILS net composite (see ILSFUND) financed at USD cash (^IRX, extended by TBILL-3M) and re-earning euro cash (the deep euro overnight chain) = the EUR-hedged view of the index; no fund fee added, no volatility rescaling",
		Build:  ilsHedgedLocal,
	}
}

func ilsHedgedLocal(f Fetcher, from time.Time) (*marketdata.Series, error) {
	local, err := ilsIndexLocal(f, from)
	if err != nil {
		return nil, err
	}
	return hedgeToEUR("ILS funds hedged to EUR (index)", local, f, from)
}

// hedgeToEUR turns a FUNDED total return quoted in USD into its EUR-hedged
// view: over every step, the local return less USD cash plus EUR cash. The
// union of dates is daily even where the local series is monthly, so the hedge
// accrues every day while the index return lands on its month end, which is
// what the two legs really do.
func hedgeToEUR(name string, local *marketdata.Series, f Fetcher, from time.Time) (*marketdata.Series, error) {
	fr, err := BuildFrame(extend(f), []string{"^IRX"}, from)
	if err != nil {
		return nil, err
	}
	usd := &marketdata.Series{Name: "USD cash (accrual)", Source: "simdata"}
	v := 100.0
	for k, d := range fr.Dates {
		v *= 1 + fr.Returns["^IRX"][k]
		usd.Points = append(usd.Points, marketdata.Point{Date: d, Close: v})
	}
	eur, err := eurOvernightDeep(f, from)
	if err != nil {
		return nil, err
	}
	dates, lv := marketdata.Align([]*marketdata.Series{local, usd, eur}, local.First().Date, time.Time{})
	if len(dates) < 2 {
		return nil, fmt.Errorf("%s: %d aligned dates", name, len(dates))
	}
	out := &marketdata.Series{Name: name, Source: "simdata", Currency: "EUR"}
	v = 100.0
	out.Points = append(out.Points, marketdata.Point{Date: dates[0], Close: v})
	for k := 1; k < len(dates); k++ {
		r := 0.0
		for i, leg := range [3]float64{1, -1, 1} {
			if lv[i][k-1] > 0 && lv[i][k] > 0 {
				r += leg * (lv[i][k]/lv[i][k-1] - 1)
			}
		}
		v *= 1 + r
		out.Points = append(out.Points, marketdata.Point{Date: dates[k], Close: v})
	}
	return out, nil
}

// ilsFeeLoad is the management-and-expense load of each vehicle in this family,
// in fraction per year, read off published fee tables and NEVER off an observed
// return gap (see feeAligned in recipes.go for why that rule exists).
//
// The index entry is an ESTIMATE, and the only one: an index of funds levies
// nothing itself, but every return in it arrives net of a constituent's own
// charge, and those constituents are private ILS vehicles that publish no
// schedule. 1.50 %/yr is the middle of what the observable end of the same
// market charges (the retail UCITS classes below run 1.32 to 1.75 %, and
// institutional ILS mandates run lower), and it is within a quarter-point of
// every target here, so the alignment it produces is small by construction.
var ilsFeeLoad = map[string]float64{
	"ILSFUNDE":     0.0150, // ESTIMATED, see above
	"IE00B3Q8M574": 0.0141, // GAM Star Cat Bond, EUR hedged ordinary accumulation: ongoing charge
	"LI0049587301": 0.0120, // Solidum Cat Bond Fund R EUR hedged accumulation: ongoing charge
	"LI0115208543": 0.0175, // Plenum CAT Bond Defensive R EUR: ongoing charge
	"LU0951570927": 0.0132, // Schroder GAIA Cat Bond IF EUR hedged: ongoing charge
}

// ilsUplift is the constant a donor segment is lifted by so it carries the
// TARGET's fee load rather than its own, floored at zero (a donor cheaper than
// its target is left alone rather than made worse, which keeps the correction
// one-directional and conservative).
func ilsUplift(target, donor string) float64 {
	up, ok := ilsFeeLoad[donor]
	if !ok {
		panic("simgen: no documented fee load for " + donor)
	}
	t, ok := ilsFeeLoad[target]
	if !ok {
		panic("simgen: no documented fee load for " + target)
	}
	return math.Max(0, up-t)
}

// catBondBuild is the Build shared by the cat bond share classes: the
// EUR-hedged index, lifted to the target's fee load and rescaled to the
// target's own monthly volatility, spliced behind the fund's real NAVs.
//
// The volatility match is what separates a fund file from the index benchmark.
// These classes are not all the same risk: a "defensive" mandate holding the
// low-expected-loss end of the market realizes barely two thirds of the whole
// market's volatility, and splicing the market's history in front of it
// unchanged would credit it with losses it was built not to take. The ratio is
// measured on MONTH-END returns over the common window, because the index only
// has months and the funds deal weekly, and it is applied to the index's
// excess over euro cash so the cash leg is not stretched with it.
func catBondBuild(target string) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		donor, err := ilsHedgedLocal(f, from)
		if err != nil {
			return nil, err
		}
		if up := ilsUplift(target, "ILSFUNDE"); up > 0 {
			donor = afterAnnualFee(donor.Name+" (fee-aligned)", donor, -up)
		}
		fund, err := f.Fetch(target, from)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", target, err)
		}
		if fund == nil || len(fund.Points) < 60 {
			return nil, fmt.Errorf("%s: no usable quotes to calibrate on", target)
		}
		cash, err := eurOvernightDeep(f, from)
		if err != nil {
			return nil, err
		}
		scaled, err := monthlyVolMatch(fund, donor, cash)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", target, err)
		}
		return scaled, nil
	}
}

// monthlyVolMatch rescales donor so its MONTH-END returns realize the same
// standard deviation the reference does over their common months, on returns in
// excess of cash so the cash leg keeps its own size. It is volMatch's sibling
// for the case volMatch cannot serve: a monthly donor and a weekly reference
// share almost no observation dates, so a per-observation match would find
// nothing to measure and silently skip the donor.
//
// cashIdx is a money-market INDEX LEVEL, not an annualized rate: the euro cash
// leg this family finances at is eurOvernightDeep, and reading it as a rate the
// way cashAccrual does would inflate it a hundredfold.
func monthlyVolMatch(ref, donor, cashIdx *marketdata.Series) (*marketdata.Series, error) {
	lo, hi := time.Time{}, time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	a, b := monthlyReturns(ref, lo, hi), monthlyReturns(donor, lo, hi)
	var ra, rb []float64
	for m, r := range b {
		if v, ok := a[m]; ok {
			ra = append(ra, v)
			rb = append(rb, r)
		}
	}
	if len(ra) < 36 {
		return nil, fmt.Errorf("%d common months, want at least 36", len(ra))
	}
	sa, sb := stdev(ra), stdev(rb)
	if sa <= 0 || sb <= 0 {
		return nil, fmt.Errorf("degenerate volatility (%.4f, %.4f)", sa, sb)
	}
	k := sa / sb
	if k < 0.5 || k > 2 {
		return nil, fmt.Errorf("volatility ratio %.2f outside [0.5, 2]: not the same trade", k)
	}
	out := &marketdata.Series{Name: donor.Name, Source: donor.Source, Currency: donor.Currency}
	v := 100.0
	out.Points = append(out.Points, marketdata.Point{Date: donor.First().Date, Close: v})
	for i := 1; i < len(donor.Points); i++ {
		if donor.Points[i-1].Close <= 0 {
			continue
		}
		r := donor.Points[i].Close/donor.Points[i-1].Close - 1
		c := eurCashReturn(cashIdx, donor.Points[i-1].Date, donor.Points[i].Date)
		v *= 1 + k*(r-c) + c
		out.Points = append(out.Points, marketdata.Point{Date: donor.Points[i].Date, Close: v})
	}
	return out, nil
}

// gamCatBondRecipe backcasts the GAM Star Cat Bond EUR-hedged accumulation
// class (IE00B3Q8M574, real NAVs from 2011-10) with the EUR-hedged ILS fund
// index in front of it, from 2006-01. The extra five and a half years are what
// this file exists for: they hold the 2008 collateral shock, the only stretch
// on record where cat bonds fell WITH equities instead of beside them, and no
// investable cat bond record a European household could have held reaches it.
func gamCatBondRecipe() Recipe {
	return Recipe{
		ID:     "IE00B3Q8M574",
		Name:   "GAM Star Cat Bond EUR hedged (backcast)",
		Method: "the EUR-hedged ILS fund composite (see ILSFUNDE), lifted to the class's ongoing charge and rescaled to its own monthly volatility, spliced behind the fund's real NAVs; monthly before 2011-10, the fund's own weekly cadence after",
		Build:  catBondBuild("IE00B3Q8M574"),

		ValidateAgainst: "IE00B3Q8M574",
		SpliceReal:      "IE00B3Q8M574",
		Donors:          []string{"ILSFUNDE"},
	}
}

// plenumCatBondRecipe backcasts the Plenum CAT Bond Defensive R EUR class
// (LI0115208543, real NAVs from 2010-09) the same way. "Defensive" is the whole
// point of the volatility match here: the mandate holds the low-expected-loss
// end of the market and realizes about four fifths of the index's volatility,
// so the index is scaled down before it is spliced.
func plenumCatBondRecipe() Recipe {
	return Recipe{
		ID:     "LI0115208543",
		Name:   "Plenum CAT Bond Defensive R EUR (backcast)",
		Method: "the EUR-hedged ILS fund composite (see ILSFUNDE), lifted to the class's ongoing charge and rescaled to its own monthly volatility, spliced behind the fund's real NAVs; monthly before 2010-09, the fund's own weekly cadence after",
		Build:  catBondBuild("LI0115208543"),

		ValidateAgainst: "LI0115208543",
		SpliceReal:      "LI0115208543",
		Donors:          []string{"ILSFUNDE"},
	}
}

// solidumCatBondRecipe backcasts the Solidum Cat Bond Fund R EUR hedged class
// (LI0049587301, real NAVs from 2009-09), the deepest publicly quoted cat bond
// record a euro household could have held, and the cheapest of the retail
// classes here. It deals semi-monthly, which the monthly match and the monthly
// prefix both accommodate.
func solidumCatBondRecipe() Recipe {
	return Recipe{
		ID:     "LI0049587301",
		Name:   "Solidum Cat Bond Fund R EUR hedged (backcast)",
		Method: "the EUR-hedged ILS fund composite (see ILSFUNDE), lifted to the class's ongoing charge and rescaled to its own monthly volatility, spliced behind the fund's real NAVs; monthly before 2009-09, the fund's own semi-monthly cadence after",
		Build:  catBondBuild("LI0049587301"),

		ValidateAgainst: "LI0049587301",
		SpliceReal:      "LI0049587301",
		Donors:          []string{"ILSFUNDE"},
	}
}
