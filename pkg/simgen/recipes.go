package simgen

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// ComponentsFrom is how far back component histories are requested; actual
// frames start at the youngest component's first quote.
var ComponentsFrom = time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)

// All returns every bundled reconstruction recipe.
func All() []Recipe {
	return []Recipe{
		ntsxRecipe(),
		iefRecipe(),
		tltRecipe(),
		ntsgRecipe(),
		ntszRecipe(),
		urthRecipe(),
		iwdaRecipe(),
		wpeaRecipe(),
		msciworldIndexRecipe(),
		sp500IndexRecipe(),
		btop50IndexRecipe(),
		ilsIndexRecipe(),
		ilsHedgedIndexRecipe(),
		gamCatBondRecipe(),
		solidumCatBondRecipe(),
		plenumCatBondRecipe(),
		btop50HedgedIndexRecipe(),
		wintonRecipe(),
		zrozRecipe(),
		dbxgRecipe(),
		mthRecipe(),
		dbmfRecipe(),
		dbmfpaRecipe(),
		dbmfeRecipe(),
		mfehRecipe(),
		kmlmRecipe(),
		aqrmfRecipe(),
		aqrmfHedgedRecipe(),
		aqrIAETRecipe(),
		aqrIAE1FTRecipe(),
		indepEuropeRecipe(),
		ctaRecipe(),
		rssbRecipe(),
		gdeRecipe(),
		rsstRecipe(),
		rsbtRecipe(),
		vtRecipe(),
		xauusdRecipe(),
		iglnRecipe(),
		shyRecipe(),
		scvwRecipe(),
		spxRecipe(),
		dpgtRecipe(),
		avantisRecipe(),
		chsnRecipe(),
		tip1eRecipe(),
		idtlRecipe(),
		dtlaRecipe(),
		dtleRecipe(),
		eresMondeRecipe(),
		eresDatadogRecipe(),
		ernaRecipe(),
		ernxRecipe(),
		xeonRecipe(),
		eimiRecipe(),
		vwraRecipe(),
		vtiRecipe(),
		icomRecipe(),
	}
}

// icomRecipe backcasts the iShares Diversified Commodity Swap UCITS ETF
// (IE00BDFL4P12, USD, real from 2009), which tracks the Bloomberg Commodity
// Total Return index, from the Bloomberg Commodity excess-return index (^BCOM,
// Yahoo daily from 1991: spot plus roll yield, no collateral) plus the T-bill
// rate (^IRX) as fully invested collateral: a total-return commodity index is
// the excess-return index earning cash on its notional, so ER + cash = TR. The
// real ICOM quotes are grafted from inception; same currency (USD), no FX leg.
// ^BCOM only needs to cover the pre-2009 tail, which it does cleanly.
func icomRecipe() Recipe {
	return Recipe{
		ID:     "IE00BDFL4P12",
		Name:   "iShares Diversified Commodity: Bloomberg Commodity TR",
		Method: "^BCOM (Bloomberg Commodity excess-return index, Yahoo daily from 1991) + ^IRX T-bill collateral = total return, real ICOM grafted from 2009",
		Build: composite("ICOM (Bloomberg Commodity TR)", []Leg{
			{ID: "^BCOM", Weight: 1},
			{ID: "^IRX", Weight: 1},
		}, "^IRX", 0),
		ValidateAgainst: "IE00BDFL4P12",
		SpliceReal:      "IE00BDFL4P12",
	}
}

// eimiRecipe backcasts the iShares Core MSCI EM IMI UCITS ETF (IE00BKM4GZ66,
// USD, real from 2014) from Vanguard Emerging Markets (VEIEX, 1994->, itself
// carried back to the MSCI EM net total-return reconstruction EM-USD, ~1988).
// MSCI EM IMI only differs from standard MSCI EM by a small-cap tail; VEIEX is
// the same broad EM equity exposure, so it is the faithful long proxy. Real
// EIMI grafted from inception, same currency (USD), no FX leg.
func eimiRecipe() Recipe {
	return Recipe{
		ID:              "IE00BKM4GZ66",
		Name:            "iShares Core MSCI EM IMI: emerging-market equity (VEIEX)",
		Method:          "VEIEX (Vanguard Emerging Markets, 1994->, extended EM-USD MSCI EM net TR to ~1988), real EIMI grafted from 2014",
		Build:           composite("EIMI (emerging markets)", []Leg{{ID: "VEIEX", Weight: 1}}, "", 0),
		ValidateAgainst: "IE00BKM4GZ66",
		SpliceReal:      "IE00BKM4GZ66",
	}
}

// vwraRecipe backcasts the Vanguard FTSE All-World UCITS ETF (IE00BK5BQT80,
// USD, real from 2019) with the same 60/30/10 US / developed-ex-US / emerging
// blend as VT: FTSE All-World (large+mid, developed+emerging) is Vanguard's
// Total World universe minus the small-cap tail, so the blend is the faithful
// long proxy. The youngest leg (VEIEX/EM-USD) sets the ~1988 start; real VWRA
// grafted from inception, same currency (USD), no FX leg.
func vwraRecipe() Recipe {
	return Recipe{
		ID:     "IE00BK5BQT80",
		Name:   "Vanguard FTSE All-World: world equity (US/dev-ex-US/EM blend)",
		Method: "0.60×VFINX + 0.30×VTMGX + 0.10×VEIEX (US/developed/EM, ~1988), real VWRA grafted from 2019",
		Build: composite("VWRA (FTSE All-World replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.30},
			{ID: "VEIEX", Weight: 0.10},
		}, "", 0),
		ValidateAgainst: "IE00BK5BQT80",
		SpliceReal:      "IE00BK5BQT80",
	}
}

// vtiRecipe backcasts the Vanguard Total Stock Market ETF (VTI, USD, real from
// 2001) from Vanguard 500 (VFINX, 1976->, carried back to the S&P 500 total
// return SP500-USD ~1871). The total US market and the S&P 500 differ only by
// a mid/small-cap tail and track at ~0.99 daily correlation, so VFINX is the
// faithful deep-history proxy; real VTI grafted from inception.
func vtiRecipe() Recipe {
	return Recipe{
		ID:              "VTI",
		Name:            "Vanguard Total US Market: S&P 500 proxy (VFINX)",
		Method:          "VFINX (Vanguard 500, 1976->, extended SP500-USD total return to ~1871; total US market ≈ S&P 500), real VTI grafted from 2001",
		Build:           composite("VTI (total US market)", []Leg{{ID: "VFINX", Weight: 1}}, "", 0),
		ValidateAgainst: "VTI",
		SpliceReal:      "VTI",
	}
}

// idtlRecipe extends the iShares $ Treasury Bond 20+yr UCITS ETF
// (IE00BSKRJZ44, USD, real from 2015) with the same long-Treasury proxy as its
// US-listed twin TLT: Vanguard Long-Term Treasury (VUSTX, 1986->), carried
// further back by the constant-maturity Treasury total-return reconstruction
// (TREASURY-LONG, daily from 1962, monthly from 1953). Same asset (US
// Treasuries 20+yr, effective duration ~15 as of mid-2026, and shortening as
// yields rise) and currency (USD, the fund's own quote
// line), so no FX leg is needed; the real IDTL quotes are grafted from
// inception.
func idtlRecipe() Recipe {
	return Recipe{
		ID:              "IE00BSKRJZ44",
		Name:            "iShares $ Treasury 20+yr: VUSTX long Treasury",
		Method:          "VUSTX (Vanguard Long-Term Treasury, 1986->, extended TREASURY-LONG daily from 1962), real IDTL grafted from 2015",
		Build:           composite("IDTL (long Treasury)", []Leg{{ID: "VUSTX", Weight: 1}}, "", 0),
		ValidateAgainst: "IE00BSKRJZ44",
		SpliceReal:      "IE00BSKRJZ44",
	}
}

// dtlaRecipe is idtlRecipe for the ACCUMULATING share class of the same fund
// (IE00BFM6TC58, DTLA, real from 2018): same bonds, same donor, and the class a
// taxable euro holder actually wants, since the distributing twin turns the
// coupon stream into yearly taxable income. Its quotes come from Yahoo, whose
// closes are dividend-adjusted, so unlike the EUR-hedged DTLE the real series
// is a genuine total return and can be grafted.
//
// The donor is geared in the same EXCESS form as zrozRecipe, iefRecipe and
// shyRecipe, cash + g × (VUSTX − cash), and it did not use to be: until 2026-08
// it multiplied VUSTX's whole total return by g, which gears the coupon along
// with the duration. A longer bond does not earn 1.13 times a shorter one's
// yield, so the plain form paid the file (g − 1) × the bill rate every year,
// and the bill rate is what the deep past is made of: 0.92 pt/yr over
// 1962-1985, 0.75 over 1986-2002, 0.57 over the whole file. It also read
// ABOVE the ungeared TLT reconstruction over 1962-1985 (5.94 against 5.28 %/yr),
// which a duration gearing cannot do in an era the long bond spent losing
// money. In the excess form the ladder is monotone the right way (TLT 5.28,
// DTLA 5.02, ZROZ 3.91 %/yr over 1962-1985) and the live-window CAGR gap falls
// from +0.38 to +0.03 pt with every path statistic unchanged to the digit.
func dtlaRecipe() Recipe {
	return Recipe{
		ID:     "DTLA",
		Name:   "iShares $ Treasury 20+ Acc: VUSTX long Treasury",
		Method: "cash + 1.13×(VUSTX − cash) (Vanguard Long-Term Treasury, 1986→, extended TREASURY-LONG daily from 1962, geared to the 20+ duration), real DTLA grafted from 2018",
		Build: composite("DTLA (long Treasury, accumulating)", []Leg{
			{ID: "VUSTX", Weight: longTreasuryGearing, Excess: true},
			{ID: "^IRX", Weight: 1},
		}, "^IRX", 0),
		ValidateAgainst: "DTLA",
		SpliceReal:      "DTLA",
	}
}

// ernaRecipe backcasts the iShares USD Ultrashort Bond UCITS ETF
// (IE00BGCSB447, USD money-market, real from 2018) as USD cash: the 13-week
// T-bill rate (^IRX, daily; extended by the FRED 3-month T-bill TBILL-3M back
// to the 1930s) compounded into a money-market index. The fund earns a small
// investment-grade credit spread over bills that the pre-inception proxy omits
// (the grafted real quotes carry it); with duration ~0 the bill rate is the
// faithful cash-equivalent backcast.
//
// That omission is what the audit's path verdict reads, and it is not a defect
// to repair. The engine's monthly correlation with the fund is 0.29 over 96
// months and 0.81 once TWO months are dropped, March 2020 (-3.46 %) and its
// April reversal (+3.54 %); dropping any further month leaves it at 0.79-0.81.
// Those two months are the investment-grade ultrashort dislocation, real to the
// day in both US-listed siblings (ICSH -1.5 %, NEAR -6.2 % on 2020-03-19
// against this fund's -3.2 %), so they are neither a bad print nor anything a
// hygiene rule should touch. They also net out: the level gap is -0.42 %/yr,
// and real quotes are grafted from 2018, so the reconstruction only ever
// governs a tail where no credit spread is observable anyway.
func ernaRecipe() Recipe {
	return Recipe{
		ID:              "IE00BGCSB447",
		Name:            "iShares USD Ultrashort Bond: USD cash (T-bill)",
		Method:          "USD cash ^IRX (13-week T-bill, daily; extended TBILL-3M to the 1930s) compounded, IG credit spread omitted pre-inception; real ERNA grafted from 2018",
		Build:           composite("ERNA (USD cash)", []Leg{{ID: "^IRX", Weight: 1}}, "^IRX", 0),
		ValidateAgainst: "IE00BGCSB447",
		SpliceReal:      "IE00BGCSB447",
	}
}

// ernxRecipe backcasts the iShares EUR Ultrashort Bond UCITS ETF
// (IE000RHYOR04, EUR money-market) as EUR cash: the bundled EUR money-market
// index (EURCASH-EUR) rendered at business-day granularity by eurCashDaily.
// Like ERNA it omits the fund's small IG credit spread pre-inception (the
// grafted real quotes carry it); same currency (EUR), so no FX leg.
func ernxRecipe() Recipe {
	return Recipe{
		ID:     "IE000RHYOR04",
		Name:   "iShares EUR Ultrashort Bond: EUR cash (money-market)",
		Method: "EUR money-market index EURCASH-EUR (3-month interbank compounded, 1994->) interpolated to business days, IG credit spread omitted pre-inception; real ERNX grafted from inception",
		Build: func(f Fetcher, from time.Time) (*marketdata.Series, error) {
			return eurCashDaily(f, from)
		},
		ValidateAgainst: "IE000RHYOR04",
		SpliceReal:      "IE000RHYOR04",
	}
}

// xeonRecipe backcasts the Xtrackers II EUR Overnight Rate Swap UCITS ETF
// (LU0290358497, EUR, real from 2007), which tracks a euro OVERNIGHT accrual
// index, from the euro overnight rates themselves (eurOvernightDeep) less the
// fund's 0.10%/yr TER.
//
// It takes the DEEP chain (the German money-market accrual before the euro
// existed, ~1960) rather than stopping at the 1994 start of the euro-area
// money-market index, for the same reason the German bond sleeve finances on
// that chain: Germany was the anchor economy and the mark the reference
// currency, so its own short rate is what euro-area cash was. The depth is
// what lets a euro floor line stand in a backtest that reaches the 1980s
// instead of becoming its binding constraint.
//
// It used to be rebuilt from the 3-month interbank index the ERNX recipe uses,
// and a 3-month rate is not an overnight rate: measured over 1999-2025 on the
// two compounded paths, the 3-month index earns 1.585 %/yr against the
// overnight's 1.426, a term premium of 0.16 points a year that a 0.10 % TER
// deduction cannot absorb. On the fund's own overlap (2007-06 to 2025-12,
// 4695 days) that showed as a reconstruction sitting 0.19 points a year above
// the fund; the overnight path takes the gap to -0.01, with no correlation
// lost (daily 0.000 -> 0.013, weekly 0.538 -> 0.544).
//
// One caveat survives either construction, and it should. The fund's series is
// an exchange-listed NAV that carries microstructure noise, while any accrual
// index is smooth by construction: the daily correlation between them is ~0.01
// and would be near zero however well the level were rebuilt. It is not a
// grading failure, it is two different objects, and no reconstruction of a cash
// sleeve should try to reproduce a bid-ask bounce. Read this line on its level.
func xeonRecipe() Recipe {
	return Recipe{
		ID:     "LU0290358497",
		Name:   "Xtrackers EUR Overnight Rate Swap: EUR overnight cash",
		Method: "euro overnight rate compounded ACT/360 (ESTR from 2019-10, EONIA 1999-01 before it) less 0.10%/yr TER, extended by the 3-month EURCASH-EUR money-market index before the euro (1994->) and by the German money-market accrual DECASH-EUR before that (~1960); real XEON grafted from 2007",
		Build: func(f Fetcher, from time.Time) (*marketdata.Series, error) {
			s, err := eurOvernightDeep(f, from)
			if err != nil {
				return nil, err
			}
			return afterFee(s, 0.0010), nil
		},
		ValidateAgainst: "LU0290358497",
		SpliceReal:      "LU0290358497",
	}
}

// eurOvernightDaily compounds the euro overnight rate into a total-return
// accrual index, ACT/360 over the calendar days between two publications, as
// the published compounded indices do (so a Friday rate accrues the weekend).
//
// The euro area has TWO overnight rates and they must be taken in this order:
// ESTR from its 2019-10-01 start, EONIA before. Over their overlap EONIA was
// DEFINED as ESTR plus 8.5 basis points, and the two series bundled here
// reproduce that constant to the fourth decimal on all 579 common days, which
// is the check that the splice joins the right pair.
//
// Before the euro, the 3-month money-market index carries the tail
// (eurCashDaily): it is the wrong tenor, but a 1994-1999 stub of a cash sleeve
// is not worth a second data source. If neither overnight rate answers, that
// index stands alone and the build says so rather than failing.
func eurOvernightDaily(f Fetcher, from time.Time) (*marketdata.Series, error) {
	deep, derr := eurCashDaily(f, from)
	rate := overnightEUR(f, from)
	if rate == nil || len(rate.Points) < 2 {
		fmt.Fprintf(os.Stderr, "eurOvernight: no euro overnight rate available, the 3-month index stands alone\n")
		return deep, derr
	}
	idx := &marketdata.Series{Name: "EUR overnight accrual (ESTR, EONIA before)", Source: "simdata", Currency: "EUR"}
	level := 100.0
	idx.Points = append(idx.Points, marketdata.Point{Date: rate.Points[0].Date, Close: level})
	for i := 1; i < len(rate.Points); i++ {
		days := rate.Points[i].Date.Sub(rate.Points[i-1].Date).Hours() / 24
		level *= 1 + rate.Points[i-1].Close/100*days/360
		idx.Points = append(idx.Points, marketdata.Point{Date: rate.Points[i].Date, Close: level})
	}
	if derr == nil && deep != nil {
		marketdata.ExtendBack(idx, deep)
	}
	return idx, nil
}

// overnightEUR splices ^ESTR over ^EONIA. Both are annualized percent levels,
// so the splice is a plain concatenation at ESTR's first date: a rate is a
// rate. A source that does not answer is skipped, leaving the other to cover
// what it can.
func overnightEUR(f Fetcher, from time.Time) *marketdata.Series {
	out := &marketdata.Series{Symbol: "^ESTR"}
	estr, serr := f.Fetch("^ESTR", from)
	eonia, oerr := f.Fetch("^EONIA", from)
	if oerr == nil && eonia != nil {
		out.Points = append(out.Points, eonia.Points...)
	}
	if serr != nil || estr == nil || len(estr.Points) < 2 {
		return out
	}
	cut := estr.First().Date
	kept := out.Points[:0]
	for _, p := range out.Points {
		if p.Date.Before(cut) {
			kept = append(kept, p)
		}
	}
	out.Points = append(kept, estr.Points...)
	return out
}

// eurCashDaily fetches the bundled EUR money-market index (EURCASH-EUR, the
// FRED 3-month interbank rate compounded, monthly from 1994) and expands it to
// business-day granularity (cashDaily).
func eurCashDaily(f Fetcher, from time.Time) (*marketdata.Series, error) {
	m, err := f.Fetch("EURCASH-EUR", from)
	if err != nil {
		return nil, err
	}
	if m == nil || len(m.Points) < 2 {
		return nil, fmt.Errorf("EURCASH-EUR: empty history")
	}
	return cashDaily("EUR cash (money-market, daily)", m), nil
}

// cashDaily expands a monthly money-market accrual index to business-day
// granularity by geometric interpolation between the monthly anchors. Such an
// index grows by a near-constant daily rate within each month, so unlike a risk
// asset it carries no real intramonth variation to lose: the interpolation is
// faithful and yields a genuinely daily series (rather than feeding month-sized
// steps to daily statistics), with no external data. The last anchor is appended
// as-is so the series ends exactly on a real published level.
func cashDaily(name string, m *marketdata.Series) *marketdata.Series {
	s := &marketdata.Series{Name: name, Source: "simdata"}
	for i := 0; i+1 < len(m.Points); i++ {
		t0, t1 := m.Points[i].Date, m.Points[i+1].Date
		l0, l1 := m.Points[i].Close, m.Points[i+1].Close
		span := t1.Sub(t0).Hours() / 24
		if span <= 0 || l0 <= 0 || l1 <= 0 {
			continue
		}
		for d := t0; d.Before(t1); d = d.AddDate(0, 0, 1) {
			if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
				continue
			}
			frac := d.Sub(t0).Hours() / 24 / span
			s.Points = append(s.Points, marketdata.Point{Date: d, Close: l0 * math.Pow(l1/l0, frac)})
		}
	}
	s.Points = append(s.Points, m.Points[len(m.Points)-1])
	return s
}

// dpgtRecipe rebuilds the Dimensional Global Targeted Value UCITS ETF
// (IE000S67ID55, launched 2025) from Dimensional's own long-running US and
// international small-cap value mutual funds, the same shop and factor design,
// blended 60/40 US / developed-ex-US, fee-aligned to the 0.44% TER. The only
// market quote is the LSE line in GBP, so the USD blend is re-expressed in GBP
// at the GBP/USD spot rate (GBPUSD=X extended to 1971 by the daily FRED
// refdata, so the start is set by DISVX ~1994) to match the real series, which
// is grafted from inception.
//
// # The donor pair is fee-aligned, from price lists only
//
// The recipe deducted the fund's WHOLE 0.44 %/yr until 2026-08, on top of two
// mutual-fund NAVs that already arrive net of their own managers' charges: the
// blend was billed roughly twice, and the deep history it lends every consumer
// ran about a third of a point a year cold. It now charges the difference and
// never more (dpgtFee), exactly as the Avantis sibling does: DFSVX 0.31 %,
// DISVX 0.43 % (0.358 % at 60/40) against the UCITS fund's 0.44, so 0.082 %/yr
// remains due. Both loads are read off the funds' pages (2026), never off an
// observed return gap.
//
// The whole file sits in one era, so the charge is a constant rather than a
// schedule: the frame starts at DISVX's own first NAV (1994-12), both donors
// quoting throughout. The Dimensional funds charged more in the 1990s than
// their current list says, so the early years carry a touch too much fee, in
// the conservative direction.
func dpgtRecipe() Recipe {
	return Recipe{
		ID:              "IE000S67ID55",
		Name:            "Dimensional Global Targeted Value: DFA small-cap value blend (GBP)",
		Method:          "0.60×DFSVX (US small value) + 0.40×DISVX (intl developed small value), fee-aligned to the fund's 0.44%/yr load (0.082%/yr over the pair's own 0.358%), converted USD→GBP at GBPUSD spot (FRED daily refdata back to 1971), real DPGT grafted from 2025",
		Build:           dpgtBuild,
		ValidateAgainst: "IE000S67ID55",
		SpliceReal:      "IE000S67ID55",
	}
}

// The DPGT geography and what the fund's ongoing charge exceeds its donors'
// (see dpgtRecipe). Floored at zero, as everywhere: a donor dearer than its
// target keeps its cost rather than being credited the difference.
const (
	dpgtUS   = 0.60
	dpgtIntl = 1 - dpgtUS

	dpgtTER = 0.0044
	dpgtFee = max(0, dpgtTER-(dpgtUS*dfsvxTER+dpgtIntl*disvxTER))
)

// dpgtBuild builds the 60/40 DFA small-cap value blend in USD, then converts
// each daily return into GBP via the GBP/USD spot rate (a GBP-denominated NAV
// equals the USD NAV divided by the USD-per-GBP rate), so the simulated
// history matches the GBP quote the real DPGT trades in. The cross is
// forward-filled onto the blend's own trading calendar (see fxOnDates)
// rather than joined into the frame, which would pollute the calendar with
// the FX feed's weekend prints.
func dpgtBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	legs := []Leg{{ID: "DFSVX", Weight: dpgtUS}, {ID: "DISVX", Weight: dpgtIntl}}
	fr, err := BuildFrame(extend(f), []string{"DFSVX", "DISVX"}, from)
	if err != nil {
		return nil, err
	}
	usd, err := Composite(fr, legs, "", dpgtFee)
	if err != nil {
		return nil, err
	}
	return convertDaily("DPGT (USD small-value blend expressed in GBP)",
		extend(f), "GBPUSD=X", from, fr.Dates, usd)
}

// fxOnDates fetches a currency cross and forward-fills its level onto the
// given trading calendar, so a conversion never adds the FX feed's own dates
// (weekend prints, foreign holidays) to a strategy's frame. Dates before the
// cross's history are dropped from the front: ok[i] reports coverage.
func fxOnDates(f Fetcher, cross string, from time.Time, dates []time.Time) (levels []float64, covered []bool, err error) {
	fx, err := f.Fetch(cross, from)
	if err != nil {
		return nil, nil, fmt.Errorf("FX cross %s: %w", cross, err)
	}
	if fx == nil || len(fx.Points) == 0 {
		return nil, nil, fmt.Errorf("FX cross %s: empty history", cross)
	}
	levels = make([]float64, len(dates))
	covered = make([]bool, len(dates))
	for i, d := range dates {
		if v, _, ok := fx.At(d); ok {
			levels[i], covered[i] = v, true
		}
	}
	return levels, covered, nil
}

// convertDaily re-expresses a USD strategy index in another currency at the
// given cross (quoted as USD per unit of the target currency): a converted
// NAV equals the USD NAV divided by the rate, so r = (1+rUSD)/(1+rFX) − 1
// per step. The output starts at the first date the cross covers.
func convertDaily(name string, f Fetcher, cross string, from time.Time, dates []time.Time, usd []float64) (*marketdata.Series, error) {
	fx, covered, err := fxOnDates(f, cross, from, dates)
	if err != nil {
		return nil, err
	}
	s := &marketdata.Series{Name: name, Source: "simdata"}
	val := 100.0
	for i := 1; i < len(usd); i++ {
		if !covered[i-1] || !covered[i] {
			continue
		}
		if len(s.Points) == 0 {
			s.Points = append(s.Points, marketdata.Point{Date: dates[i-1], Close: val})
		}
		rUSD := usd[i]/usd[i-1] - 1
		rFX := fx[i]/fx[i-1] - 1
		val *= (1 + rUSD) / (1 + rFX)
		s.Points = append(s.Points, marketdata.Point{Date: dates[i], Close: val})
	}
	if len(s.Points) < 2 {
		return nil, fmt.Errorf("%s: no overlap between the strategy and %s", name, cross)
	}
	return s, nil
}

// avantisRecipe rebuilds the Avantis Global Small Cap Value UCITS ETF
// (IE0003R87OG3, AVWS, EUR-quoted on the FT line, launched 2024-09-25) as a
// US / developed-ex-US small-cap value pair, taken from the SAME MANAGER's own
// sleeves for as long as they exist and from Dimensional's long-running mutual
// funds before that. The blend is built in USD and re-expressed in EUR at the
// EURUSD spot (Yahoo daily, ~1994, which sets the start), so the real series it
// splices onto (also EUR) lines up in currency, and the real Avantis quotes are
// grafted from inception. This makes the file-recommended global buy option
// backtestable; ZPRV (US small-cap value) still reaches deeper, to 1963.
//
// # The geography is 70 % US, and used to be shipped inverted
//
// The recipe shipped 0.40 US / 0.60 international until 2026-08, on the belief
// that the fund held ~40 % North America. It holds the opposite, and three
// independent readings agree on it:
//
//   - the fund's own factsheet (31/07/2026) publishes United States 69.2 %,
//     Japan 10.2 %, United Kingdom 3.6 %, Canada 3.3 %, Australia 2.7 %;
//   - regressing the fund's monthly returns, converted to USD at EURUSD spot,
//     on the two donor pairs over their 21 common months gives 0.73 DFSVX /
//     0.30 DISVX (R2 0.973) and 0.68 AVUV / 0.33 AVDV (R2 0.967);
//   - a global developed small-cap value benchmark is about two thirds US by
//     construction, so 0.40 was never plausible a priori either.
//
// The adopted split is ROUNDED to 0.70 / 0.30: twenty-one months cannot justify
// two decimals, and the rounded pair sits between the two regressions and on
// the published geography. Measured monthly over those 21 months, in the EUR the
// file ships (correlation with the fund, CAGR gap, annualized volatility against
// the fund's own 16.1 %):
//
//	0.40 DFSVX + 0.60 DISVX (shipped until 2026-08)  0.934  +3.3 pts  12.6 %
//	0.70 DFSVX + 0.30 DISVX                          0.988  -1.3 pts  15.5 %
//	0.40 AVUV  + 0.60 AVDV                           0.936  +4.2 pts  13.3 %
//	0.70 AVUV  + 0.30 AVDV  (adopted)                0.984  -0.3 pt   16.2 %
//
// # The nearest donor is the same manager
//
// AVUV and AVDV are Avantis' own US and international small-cap value ETFs,
// real from 2019-09-26, running the same process the UCITS fund runs. They are
// therefore the donor for 2019-09 onward, exactly as the managed-futures chains
// prefer a same-manager NAV to a reconstruction, and the Dimensional pair stays
// behind them for 1994-2019. The two are spliced with marketdata.ExtendBack
// rather than through DonorChain: the donor here is a BLEND rather than a
// fetchable identifier, and with only 21 months of fund quotes to calibrate on,
// DonorChain's volatility match would be noise (the table above shows the pair
// already lands within 1 % of the fund's volatility unscaled).
//
// # Both donor segments are fee-aligned, from price lists only
//
// A donor's returns are net of ITS ongoing charge, not the target's, and here
// the donors are CHEAPER than the UCITS wrapper they stand in for, so each
// segment is charged the difference and never more (avwsFee). Loads are read off
// the fund pages (2026) and never off an observed return gap: AVUV 0.25 %, AVDV
// 0.36 % (0.283 % blended), DFSVX 0.31 %, DISVX 0.43 % (0.346 % blended), against
// the UCITS fund's 0.39 %. That is 0.107 %/yr charged to the Avantis segment and
// 0.044 %/yr to the Dimensional one, both small enough to round to nothing over a
// year and worth applying anyway since they cost one multiplication. The
// Dimensional funds charged more in the 1990s than their current list says, so
// the deep segment carries a touch too much fee, in the conservative direction.
func avantisRecipe() Recipe {
	return Recipe{
		ID:              "IE0003R87OG3",
		Name:            "Avantis Global Small Cap Value: same-manager sleeves over a DFA small-cap value blend (EUR)",
		Method:          "0.70×AVUV + 0.30×AVDV (Avantis' own US and intl small-value sleeves, real from 2019-09) spliced over 0.70×DFSVX + 0.30×DISVX (~1994), each fee-aligned to the fund's 0.39%/yr load, converted USD→EUR at EURUSD spot; real Avantis grafted from 2024-10",
		Build:           avantisBuild,
		ValidateAgainst: "IE0003R87OG3",
		SpliceReal:      "IE0003R87OG3",
	}
}

// The Avantis fund's own geography, rounded (see avantisRecipe), and the
// published ongoing charges of the fund and of its four donors.
const (
	avwsUS   = 0.70
	avwsIntl = 1 - avwsUS

	avwsTER  = 0.0039
	avuvTER  = 0.0025
	avdvTER  = 0.0036
	dfsvxTER = 0.0031
	disvxTER = 0.0043
)

// avantisBuild splices the same-manager small-value pair over the Dimensional
// one in USD, then converts each daily return into EUR via the EURUSD spot rate,
// so the simulated history matches the EUR quote the real Avantis Global Small
// Cap Value trades in. dpgtBuild is its GBP-quoted Dimensional-only sibling.
func avantisBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	usd, err := avwsBlend(f, from, "AVUV", avuvTER, "AVDV", avdvTER)
	if err != nil {
		return nil, err
	}
	dfa, err := avwsBlend(f, from, "DFSVX", dfsvxTER, "DISVX", disvxTER)
	if err != nil {
		return nil, err
	}
	marketdata.ExtendBack(usd, dfa)
	dates := make([]time.Time, len(usd.Points))
	values := make([]float64, len(usd.Points))
	for i, p := range usd.Points {
		dates[i], values[i] = p.Date, p.Close
	}
	return convertDaily("Avantis Global SCV (USD small-value blend expressed in EUR)",
		extend(f), "EURUSD=X", from, dates, values)
}

// avwsBlend builds one US / developed-ex-US small-cap value pair at the fund's
// own geography, in USD, charging only what the fund's ongoing charge exceeds
// the pair's own. A pair dearer than the fund is left as its managers published
// it rather than credited the difference: a donor may lose its wrapper's cost
// advantage, never gain a return it did not earn.
func avwsBlend(f Fetcher, from time.Time, usID string, usTER float64, intlID string, intlTER float64) (*marketdata.Series, error) {
	fr, err := BuildFrame(extend(f), []string{usID, intlID}, from)
	if err != nil {
		return nil, err
	}
	legs := []Leg{{ID: usID, Weight: avwsUS}, {ID: intlID, Weight: avwsIntl}}
	values, err := Composite(fr, legs, "", max(0, avwsTER-(avwsUS*usTER+avwsIntl*intlTER)))
	if err != nil {
		return nil, err
	}
	return SeriesFromFrame(usID+"/"+intlID+" small-cap value blend (USD)", fr, values), nil
}

// scvwRecipe rebuilds US small-cap value from DFA US Small Cap Value
// (DFSVX, 1993→, total return), itself extended before 1993 by the Ken French
// small-value factor (refdata USSCV-USD, daily from 1963-07), with the real
// SPDR ZPRV grafted on top. Cross-checked once against the MSCI USA Small Cap
// Value Weighted index (weekly corr 0.90, CAGR 11.4% vs 10.4% over 1997-2015)
// to confirm faithfulness.
//
// # The three segments, and what each is worth
//
// The file is three records end to end, and only the deepest is not a fund:
//
//	1963-07 .. 1993-02  Ken French SMALL HiBM, a gross academic factor
//	1993-02 .. 2015-02  DFSVX, a real fund's NAV, net of its own charges
//	2015-02 ..          ZPRV itself
//
// The middle segment is the one that earns the file its length. Measured
// monthly against the real fund over their 137 common months, DFSVX tracks ZPRV
// at a correlation of 0.972 for a CAGR gap of -0.30 pt/yr and a volatility ratio
// of 1.001: a real US small-value fund is very nearly the fund this file is
// about. DFA's Targeted Value fund (DFFVX) measures marginally better still
// (0.976, gap -0.00) but starts in 2000-02, seven years later, so it would add
// no history and displace a donor that is already good; it is left out and
// stays here as the measurement that says why.
//
// # The deep segment is gross, and pays a measured haircut for it
//
// A Ken French portfolio charges nothing and trades for free. Since 2026-08 the
// factor therefore gives back 1.0 %/yr before it is spliced (longBackFee, which
// carries the full measurement table and the two caveats): that is the wedge it
// shows against DFSVX over their thirty-three common years, +1.02 pts/yr with a
// standard error of 0.69, at a monthly correlation of 0.980. Truncating the file
// at 1993 instead was considered and rejected: this tail is not a simulation of
// a path the way a reconstructed trend book is, it is the realized return of the
// actual stocks, and its one defect is a level a long overlap can measure. Over
// 1963-07 to 1993-02 the segment now compounds at 16.96 %/yr where it read
// 18.14 %; nothing after 1993-02 moves, since the haircut touches the proxy
// alone and ExtendBack pins the level at the junction.
func scvwRecipe() Recipe {
	return Recipe{
		ID:              "IE00BSPLC413",
		Name:            "SPDR MSCI USA Small Cap Value Weighted",
		Method:          "Ken French small-value factor (USSCV-USD, daily 1963-07→, gross, so less a measured 1.0%/yr) spliced before DFSVX (DFA US Small Cap Value, 1993→); real ZPRV grafted from 2015",
		Build:           composite("US small-cap value (DFSVX)", []Leg{{ID: "DFSVX", Weight: 1}}, "", 0),
		ValidateAgainst: "IE00BSPLC413",
		SpliceReal:      "IE00BSPLC413",
	}
}

// spxRecipe backcasts the Vanguard S&P 500 UCITS ETF (IE00BFMXXD54, USD Acc,
// real from 2019) as the S&P 500 total return, from Vanguard 500 (VFINX, 1976→,
// itself extended to the 1871 Shiller S&P 500 total return via the SP500-USD
// refdata in the longBack table). USD throughout, so the EUR (or any) view
// converts with real FX across the whole history; the real VUAA quotes
// auto-splice from 2019. Reached by the "SP500" catalog alias, hence SP500SIM.
func spxRecipe() Recipe {
	return Recipe{
		ID:              "IE00BFMXXD54",
		Name:            "Vanguard S&P 500 UCITS ETF: S&P 500 total return (1871 with -refdata)",
		Method:          "VFINX (Vanguard 500, extended with the S&P 500 total return via SP500-USD refdata ~1871); real VUAA auto-spliced from 2019",
		Build:           composite("S&P 500 total return (VFINX)", []Leg{{ID: "VFINX", Weight: 1}}, "", 0),
		ValidateAgainst: "IE00BFMXXD54",
	}
}

// chsnDonor is the distributing twin of the target share class: the same UBS
// Core Euro Inflation Linked 1-10 portfolio, listed on Xetra since 2017-10-31,
// eight years older than the EUR-acc class the file is about.
//
// It is spliced from the YAHOO ADJUSTED closes and from nothing else. Those
// closes add the semi-annual distributions back, a clean ex-date step function
// worth +17.90 % cumulated over the 8.63 years to 2026-08, and the result
// agrees with the live accumulating class to -0.047 %/yr over their 222 common
// days: two share classes of one portfolio, as they should be. The FT series of
// the very same listing (xid 437262179) is a PRICE return and misses about
// 1.93 %/yr of income, the DTLE trap; it is forbidden here.
const chsnDonor = "LU1645380368"

// chsnDonorFeeSteps lifts the donor class onto the target class's price list.
// Both are 0.08 %/yr today, but the distributing class carried more for most of
// the spliced era: 0.20 %/yr from its 2017 launch to the end of 2023, 0.10 %/yr
// through 2025-07-14, 0.08 %/yr since (the day the accumulating class itself
// launched). The uplift is the difference, and nothing else: it is read off the
// two published schedules and never derived from the return gap the two
// records show, which would hand the backcast a performance nobody earned.
//
// The schedule is stepped rather than averaged into the single 0.10 %/yr the
// era's time weights would give, because the steps are documented and free: a
// flat uplift would over-credit the recent years and under-credit the deep
// ones, which are exactly the years the file exists to cover.
var chsnDonorFeeSteps = []feeStep{
	{Annual: -0.0012}, // 0.20 % donor vs 0.08 % target
	{From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Annual: -0.0002}, // 0.10 % vs 0.08 %
	{From: time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC), Annual: 0},      // both at 0.08 %
}

// chsnIBCIBeta is the share of the all-maturity euro linker ETF (IBCI) held
// against EUR cash to reproduce the 1-10 segment, and it is MEASURED, not
// modelled. Regressing the real 1-10 total return on IBCI, both in excess of
// EUR cash, over 2017-2026 gives a beta of 0.622 (R2 0.857, sub-period betas
// 0.598 to 0.632, alpha +0.48 %/yr with a standard error of 0.53, so no level
// correction is permitted alongside it). 0.60 sits inside that range and is
// what the file already shipped, so it stays.
//
// The paragraph this replaces reasoned from a duration RATIO and expected it to
// drift toward 0.64 as the two funds' factsheet durations moved. That model was
// wrong: the 1-10 leg's duration is stable near 4.9 years, plus or minus 0.3,
// while the all-maturity leg's fell from 8.7 to 7.4, so the ratio moved for a
// reason that has nothing to do with how the two co-move. The regression does
// not move with it, and it is the justification now.
const chsnIBCIBeta = 0.60

// chsnRecipe backcasts the UBS Core Euro Inflation Linked 1-10 ETF
// (LU1645380442, a EUR-acc share class launched 2025-07) in three segments,
// real quotes for as long as real quotes of this exact portfolio exist:
//
//	2005-11 .. 2017-10  0.60×IBCI + 0.40×EUR cash, a proxy
//	2017-10 .. 2025-08  LU1645380368, the fund's own distributing class
//	2025-08 ..          LU1645380442 itself (SpliceReal)
//
// The middle segment is new and is the point. The proxy alone lagged the real
// 1-10 total return by about 0.46 %/yr over 2017-2026 (+11.76 % against
// +16.44 % cumulated), and the miss is not spread evenly: it is 2.23 points in
// 2021 and 2.47 in 2022, the two years short linkers earned their inflation
// carry and a duration-matched blend of an all-maturity fund with cash did not.
// Eight years of the fund's own history retire that question entirely.
//
// # The texture trap
//
// The donor's Yahoo prints are STALE before roughly 2020-11: 79 % of the days
// in 2018 and 80 % in 2019 print the previous close, with flat runs reaching 24
// trading days, against 3 % or fewer from 2021. The LEVEL path is sound and the
// daily texture is fiction, so no rescaling of any kind is applied to that era:
// a scale factor can only stretch a shape, and this one has to be replaced.
//
// The fiction is measurable, and it runs the way a fortnight's move landing on
// a single print runs, not the way a flat quote intuitively suggests. Over
// 2018-2020 the raw donor annualizes 4.32 %/yr of DAILY volatility against
// 3.56 % monthly, where the reshaped segment reads 3.52 % and 3.47 %: the
// staleness inflates the daily figure by four fifths of a point and leaves the
// monthly one nearly intact, which is exactly the signature of a level that is
// right and a calendar that is not. Volatility-matching the era would have
// entrenched that artefact rather than removed it.
//
// chsnBuild takes the honest reading instead. A repeated print carries no
// information, so only the days the donor's level actually MOVED are kept, and
// they are used as anchors carrying the proxy's daily shape in between
// (shapedSeries). The result passes exactly through every real NAV the donor
// published and takes its day-to-day texture from the linker-plus-cash blend
// over the stale era. From 2021 the donor prints properly (3 % zero-return days
// or fewer), consecutive anchors are one trading day apart, and the mechanism
// is a pass-through by construction: no cutoff date is written down anywhere,
// the data says where the staleness ends. Measurement confirms it, the 2021-2026
// stretch reading 4.31 %/yr of daily volatility in the file against the donor's
// own 4.34 %.
//
// What survives the treatment is the level, exactly: on every day the donor
// published a move, the file's ratio to it drifts by the fee uplift and by
// nothing else (+0.118 %/yr through 2023, +0.020 % in 2024, flat after).
//
// # Considered and rejected
//
// Dimensional Euro Inflation Linked Intermediate Duration (IE00B3N38C44, FT
// daily from 2011-06) as a 2011-2017 donor, to shorten the proxy era further.
// It is a real fund and would feel like the better record, but it tracks the
// 5-10 bucket actively, and it measures worse: monthly correlation 0.869
// against the real 1-10, where the scaled-IBCI proxy reaches 0.925. More real,
// less faithful; the proxy keeps the era.
func chsnRecipe() Recipe {
	return Recipe{
		ID:   "LU1645380442",
		Name: "UBS Core Euro Inflation Linked 1-10: the fund's own distributing class, then a euro linker proxy",
		Method: "real NAVs of the distributing twin LU1645380368 (Yahoo adjusted = total return, from 2017-10, lifted +0.12%/yr then +0.02%/yr onto the target's 0.08% price list, its stale pre-2021 prints reshaped on the proxy's texture), " +
			"behind it 0.60×IBCI (iShares Euro Inflation Linked Govt Bond, all-maturity euro-linker, FT daily from 2005, measured beta 0.622) + 0.40×EUR cash (EURCASH-EUR); real CHSN grafted from 2025",
		Donors:          []string{chsnDonor, "IE00B0M62X26"},
		Build:           chsnBuild,
		ValidateAgainst: "LU1645380442",
		SpliceReal:      "LU1645380442",
	}
}

// chsnBuild assembles the two reconstructed segments: the distributing class,
// fee-aligned and reshaped where its prints are stale, with the linker-plus-cash
// proxy spliced behind it at its 2017 inception. If the donor is unreachable
// (an offline build, a dead symbol) the proxy stands alone, which is exactly
// the file that shipped before.
func chsnBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	proxy, err := composite("CHSN (euro inflation-linked proxy)", []Leg{
		{ID: "IBCI", Weight: chsnIBCIBeta},
		{ID: "EURCASH-EUR", Weight: 1 - chsnIBCIBeta},
	}, "", 0)(f, from)
	if err != nil {
		return nil, err
	}
	donor, derr := f.Fetch(chsnDonor, from)
	if derr != nil || donor == nil || len(donor.Points) < 30 {
		fmt.Fprintf(os.Stderr, "chsn: donor %s unavailable (%v), the IBCI proxy stands alone\n", chsnDonor, derr)
		return proxy, nil
	}
	aligned := afterFeeSteps("CHSN (distributing class, fee-aligned)", movesOnly(donor), chsnDonorFeeSteps)
	out := shapedSeries(aligned, proxy)
	out.SimulatedBefore = time.Time{} // let ExtendBack put the proxy era in front
	marketdata.ExtendBack(out, proxy)
	out.Name = "CHSN (distributing class, then euro inflation-linked proxy)"
	out.Source = "simdata"
	out.SimulatedBefore = time.Time{}
	return out, nil
}

// movesOnly drops the points that merely repeat the previous close. On a feed
// that prints every session it is a no-op; on one that goes stale it turns a
// flat run into the gap it really is, which is what lets shapedSeries fill the
// run with a live texture instead of pinning it flat. The first and last points
// are always kept, so the segment keeps both of its ends.
func movesOnly(s *marketdata.Series) *marketdata.Series {
	out := *s
	out.Points = make([]marketdata.Point, 0, len(s.Points))
	for i, p := range s.Points {
		if i == 0 || i == len(s.Points)-1 || p.Close != out.Points[len(out.Points)-1].Close {
			out.Points = append(out.Points, p)
		}
	}
	return &out
}

// tip1eRecipe backcasts the UBS Core Bloomberg TIPS 1-10 EUR-hedged ETF
// (LU1459801780, real from 2016) as US TIPS hedged to EUR, through the standard
// FX-hedge identity: a hedged foreign return equals the local return less the
// foreign (USD) cash rate plus the domestic (EUR) cash rate, so the EUR
// investor collects the EUR-minus-USD carry (deeply negative over 2015-2022, so
// most of the live overlap) on top of the TIPS return.
//
// The local leg is two REAL TIPS ETFs rather than the all-maturity mutual fund
// that used to carry it alone: 0.60 TIP (all maturities) + 0.40 STIP (0-5
// years), both with market prices, blended so the pair's own risk lands on the
// fund's. Before STIP's 2010-12 start the geared VIPSX path takes over
// (tipsTailGearing), spliced.
//
// Measured on the fund's own live overlap (2016-09 to 2026-07, 2403 days), the
// swap earns its slot on all three criteria of its adoption bar, monthly
// correlation, volatility and level:
//
//	shipped 0.64×VIPSX     monthly 0.926, monthly vol 82 % of the fund's, CAGR gap -0.62 pt
//	0.60 TIP + 0.40 STIP   monthly 0.941, monthly vol 101 %,               CAGR gap +0.23 pt
//
// Two thirds of that level move belong to a plain bug rather than to the donor.
// A fully hedged position earns the domestic cash rate on its WHOLE capital,
// and the recipe was earning it on 0.35 of it, the same fraction as the leg it
// was not holding in TIPS. Fixing that alone takes the old construction's gap
// from -0.62 to -0.13; the hedged sibling dtleBuild always had it right.
//
// Volatility is read MONTHLY here, and deliberately. The fund's daily quotes
// carry exchange noise a reconstruction cannot and should not reproduce: its
// daily volatility is 4.93 % against 3.84 % monthly, a 1.28 ratio where the
// reconstruction's is 1.11, and matching the daily figure would buy a longer
// duration than the fund holds. The blend's implied duration (~5.0 against the
// 1-10 index's ~4.4) is already at the long end, which is the side to err on.
//
// Two variants were measured and rejected. TIP alone, rescaled to 0.66, matches
// the level best of all (gap -0.03) and does NOT improve the months (0.927), so
// it fails the bar's first clause: a single all-maturity fund cannot reproduce
// what the short end does on its own. And serving the EUR cash leg at
// business-day granularity (eurCashDaily, as ERNX and XEON do) costs weekly
// correlation, 0.809 against 0.846, because it adds to the frame the days the
// US bond market does not trade; the monthly index stays.
//
// Daily correlation stays modest (~0.38) for the reason it always did: the
// legs are US-close series and the fund is struck in Europe. The weekly (0.85)
// and monthly (0.94) figures are the meaningful ones.
//
// The validation line's beta and tracking error move the "wrong" way (0.52 to
// 0.42 and 4.9 to 5.2 %/yr) and that is arithmetic, not damage. Beta is
// corr × sigma_real / sigma_sim, so with a daily correlation pinned near 0.38
// by the clock, beta only reaches 1 for a reconstruction carrying two fifths
// of the fund's risk. Both statistics reward under-risking here; the file
// refuses to buy them that way.
func tip1eRecipe() Recipe {
	return Recipe{
		ID:              "LU1459801780",
		Name:            "UBS Core BBG TIPS 1-10 (EUR-hedged): US TIPS hedged to EUR",
		Method:          "0.60×TIP + 0.40×STIP (real US TIPS ETFs, 2010-12->, risk-matched to the fund) financed at USD cash ^IRX and earning EUR cash (EURCASH-EUR) on the whole capital = EUR-hedged TIPS; 0.76×VIPSX on the same identity before STIP exists (2000->); real 42C0 grafted from 2016",
		Build:           tip1eBuild,
		ValidateAgainst: "LU1459801780",
		SpliceReal:      "LU1459801780",
	}
}

// tipsTailGearing is the weight the VIPSX donor carries before the ETF blend
// starts. It is the ratio of the blend's monthly volatility to hedged VIPSX's
// over their whole common window (2010-12 to 2026-07, 188 months): 3.77 %
// against 4.93 %. Splicing an ungeared VIPSX, or the 0.64 the recipe used to
// carry, would leave the pre-2010 decade running a fifth colder in risk than
// the era after it, which is the one thing a spliced file must not do.
const tipsTailGearing = 0.76

// tip1eBuild assembles the hedged TIPS series: the ETF blend where both ETFs
// quote, the geared VIPSX path on the identical hedge identity behind it. If
// the ETFs are unreachable the VIPSX path stands alone, which is also what an
// offline build gets.
func tip1eBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	tail, err := composite("42C0 (EUR-hedged US TIPS, VIPSX era)", []Leg{
		{ID: "VIPSX", Weight: tipsTailGearing, Excess: true},
		{ID: "EURCASH-EUR", Weight: 1.00},
	}, "^IRX", 0)(f, from)
	if err != nil {
		return nil, err
	}
	blend, berr := composite("42C0 (EUR-hedged US TIPS)", []Leg{
		{ID: "TIP", Weight: 0.60, Excess: true},
		{ID: "STIP", Weight: 0.40, Excess: true},
		{ID: "EURCASH-EUR", Weight: 1.00},
	}, "^IRX", 0)(f, from)
	if berr != nil {
		fmt.Fprintf(os.Stderr, "tip1e: TIP/STIP blend unavailable (%v), the VIPSX path stands alone\n", berr)
		return tail, nil
	}
	marketdata.ExtendBack(blend, tail)
	return blend, nil
}

func rssbRecipe() Recipe {
	return Recipe{
		ID:     "RSSB",
		Name:   "Return Stacked Global Stocks & Bonds",
		Method: "100% world equity + 100% (VFITX − overnight financing: SOFR 2018→, effective fed funds 1954→, T-bill before) Treasury stack (1999→), real RSSB grafted from 2023",
		Build: composite("RSSB (100/100 stocks+bonds replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.30},
			{ID: "VEIEX", Weight: 0.10},
			{ID: "VFITX", Weight: 1.00, Excess: true},
		}, usdOvernight, 0),
		ValidateAgainst: "RSSB",
		SpliceReal:      "RSSB",
	}
}

// gdeRecipe backcasts the WisdomTree Efficient Gold Plus Equity Strategy Fund
// (GDE, US-listed, real from 2022): for every dollar, ~90 % large-cap US equity
// held as the funded core plus a ~90 % gold overlay run through futures, with
// the residual cash as collateral. It mirrors the NTSX efficient-core method,
// swapping the Treasury overlay for gold: the equity leg earns its full return,
// the gold leg earns the spot excess over its FINANCING (gold futures ~= spot
// financed at the overnight rate, no carry yield), and 0.10 of NAV sits in
// bills. Every leg is deep (VFINX -> S&P 500 TR, XAUUSD gold ~1968, the
// financing rate ~1934 through the T-bill), so the composite reaches back to
// the gold leg's floor. Real GDE quotes are grafted from inception; same
// currency (USD), no FX leg.
//
// Only one leg is a fund here: the gold overlay is a futures price and the
// collateral a bill rate, and neither charges anything, so the equity donor
// alone carries a load (0.90×0.14 = 0.126 %/yr). The fund's 0.20 is deducted
// in full before VFINX's own quotes and the remaining 0.074 after (feeGap).
func gdeRecipe() Recipe {
	return Recipe{
		ID:     "GDE",
		Name:   "WisdomTree Efficient Gold Plus Equity: 90/90 replication",
		Method: "0.90×VFINX + 0.90×(GC=F gold − overnight financing: SOFR 2018→, effective fed funds 1954→, T-bill before) overlay + 0.10×cash ^IRX, daily rebalancing, the 0.20%/yr TER less the equity donor's own 0.126%/yr where VFINX carries the leg, real GDE grafted from 2022",
		Build: alignedComposite("GDE (90/90 gold+equity replication)", []pricedLeg{
			{ID: "VFINX", Weight: 0.90, Load: vfinxTER},
			{ID: "GC=F", Weight: 0.90, Excess: true},
			{ID: "^IRX", Weight: 0.10},
		}, usdOvernight, 0.0020),
		ValidateAgainst: "GDE",
		SpliceReal:      "GDE",
	}
}

// rsstRecipe backcasts the Return Stacked US Stocks & Managed Futures ETF (RSST,
// US-listed, real from 2023): for every dollar, ~100 % large-cap US equity held
// as the funded core plus a ~100 % managed-futures overlay. It mirrors the GDE
// method, swapping the gold overlay for trend: the equity leg (VFINX) earns its
// full return, and the trend leg is stacked as excess over cash, since a
// managed-futures programme is run on collateral that already earns the T-bill
// rate. The floor is set by the trend leg's reference, which begins in 2000
// (see stackedTrend). Real RSST quotes are grafted from inception; same
// currency (USD), no FX leg.
func rsstRecipe() Recipe {
	return Recipe{
		ID:              "RSST",
		Name:            "Return Stacked US Stocks & Managed Futures: 100/100 replication",
		Method:          "1.00×VFINX + 1.00×(TSMOM trend − overnight financing: SOFR 2018→, effective fed funds 1954→, T-bill before) overlay, the overlay's months and its level from the bundled net pure-trend reference (2000→, where that reference begins), 0.96%/yr fees less the equity donor's own 0.14%, real RSST grafted from 2023",
		Build:           stackedTrend("RSST (100% stocks + TSMOM overlay)", "VFINX", vfinxTER, mfConfig(0.10, 0), 0.0096),
		ValidateAgainst: "RSST",
		SpliceReal:      "RSST",
	}
}

// rsbtRecipe backcasts the Return Stacked Bonds & Managed Futures ETF (RSBT,
// US-listed, real from 2023): ~100 % core US bonds as the funded core plus a
// ~100 % managed-futures overlay. Same construction as rsstRecipe with the
// intermediate Treasury proxy (VFITX, as in rssbRecipe) standing in for the
// fund's core bond sleeve, and KMLM stacked as the trend excess over cash.
func rsbtRecipe() Recipe {
	return Recipe{
		ID:              "RSBT",
		Name:            "Return Stacked Bonds & Managed Futures: 100/100 replication",
		Method:          "1.00×VFITX + 1.00×(TSMOM trend − overnight financing: SOFR 2018→, effective fed funds 1954→, T-bill before) overlay, the overlay's months and its level from the bundled net pure-trend reference (2000→, where that reference begins), 0.97%/yr fees less the bond donor's own 0.20%, real RSBT grafted from 2023",
		Build:           stackedTrend("RSBT (100% bonds + TSMOM overlay)", "VFITX", vfitxTER, mfConfig(0.10, 0), 0.0097),
		ValidateAgainst: "RSBT",
		SpliceReal:      "RSBT",
	}
}

func vtRecipe() Recipe {
	return Recipe{
		ID:     "VT",
		Name:   "Vanguard Total World Stock",
		Method: "0.60×VFINX + 0.30×VTMGX + 0.10×VEIEX (US/developed/EM world, 1999→), real VT grafted from 2008",
		Build: composite("VT (total world replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.30},
			{ID: "VEIEX", Weight: 0.10},
		}, "", 0),
		ValidateAgainst: "VT",
		SpliceReal:      "VT",
	}
}

// iwdaRecipe gives the iShares Core MSCI World (2009) the same 60/40
// US/international reconstruction as URTH, so MSCI-World portfolios reach
// back to 1999.
func iwdaRecipe() Recipe {
	return Recipe{
		ID:     "IE00B4L5Y983",
		Name:   "iShares Core MSCI World: MSCI World total return (1969 with -refdata)",
		Method: "real MSCI World net TR (MSCIWORLD-USD refdata, monthly 1969→) with the daily shape of the MSCI World price index (^990100-USD-STRD, 1972→), less 0.20%/yr TER; without the refdata falls back to 0.60×VFINX+0.40×VTMGX (1999)",
		Build: msciWorld(0.0020, composite("IWDA (MSCI World replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.40},
		}, "", 0.0020)),
		ValidateAgainst: "IE00B4L5Y983",
	}
}

// msciworldIndexRecipe is the pure MSCI World Net TR benchmark: the same
// reconstruction as the IWDA/URTH trackers but with NO fund fee, served as the
// non-investable MSCIWORLD index (fees 0, no ISIN). The CAGR runs about a
// tracker's TER above IE00B4L5Y983 by design; correlation stays ~1.0.
func msciworldIndexRecipe() Recipe {
	return Recipe{
		ID:     "MSCIWORLD",
		Name:   "MSCI World Net TR (index, fee-free)",
		Method: "MSCI World net total return: MSCIWORLD-USD refdata (monthly 1969→) with the daily shape of the MSCI World price index (^990100-USD-STRD, 1972→), no fund fee; fallback 0.60×VFINX+0.40×VTMGX (1999)",
		Build: msciWorld(0, composite("MSCI World (replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.40},
		}, "", 0)),
		ValidateAgainst: "IE00B4L5Y983",
	}
}

// wpeaRecipe backcasts the iShares MSCI World Swap PEA UCITS ETF
// (IE0002XZSHO1, WPEA, EUR Acc, real from 2024-03-26): the PEA-eligible
// synthetic MSCI World, expressed in EUR. WPEA carries the same 0.20%/yr TER as
// the iShares Core MSCI World (IWDA, IE00B4L5Y983) and tracks the same index, so
// its history is that fund's net total-return path re-expressed in EUR: real
// IWDA quotes from 2009, the IWDA/URTH monthly reconstruction before that
// (MSCIWORLD-USD refdata from 1969, ^990100-USD-STRD daily shape from 1972),
// all converted USD->EUR at the EUR/USD spot (extended to 1971 by the bundled
// ECU/DM/EUR proxy, as for the DBMFE/Avantis euro legs), with real WPEA quotes
// grafted from 2024. A EUR investor's MSCI World carries the unhedged USD/EUR
// currency move, so this series differs from the USD IWDA view by exactly that
// FX path. Leaning on real IWDA for 2009-2024, rather than the coarse monthly
// anchor, keeps the recent decade faithful; the 0.20% TER is the headline charge
// only, the swap's substitute-basket cost is not modelled separately (the
// TER-only convention of every tracker recipe).
func wpeaRecipe() Recipe {
	return Recipe{
		ID:              "WPEA",
		Name:            "iShares MSCI World Swap PEA: MSCI World net TR in EUR (1971 with -refdata)",
		Method:          "iShares Core MSCI World net TR in USD (real IWDA from 2009, MSCIWORLD-USD refdata + ^990100-USD-STRD daily shape before, 0.20%/yr TER) converted USD->EUR at EURUSD spot (bundled ECU/DM/EUR proxy back to 1971); real WPEA grafted from 2024",
		Build:           wpeaBuild,
		ValidateAgainst: "WPEA",
		SpliceReal:      "WPEA",
	}
}

// wpeaBuild builds the USD net-TR MSCI World that WPEA tracks, then converts each
// daily return into an unhedged EUR return so the simulated history matches the
// EUR quote WPEA trades in. The USD leg is real iShares Core MSCI World
// (IE00B4L5Y983, USD, 2009->, already net of the same 0.20% TER) with the
// IWDA/URTH monthly reconstruction (also carrying 0.20%) extended behind it, so
// the mid-period rides real MSCI World rather than the coarse anchor while the
// equal TERs keep the blend uniformly net-of-fee. The USD->EUR step mirrors
// avantisBuild/dbmfeBuild: a EUR-denominated NAV equals the USD NAV divided by
// the USD-per-EUR rate, so r_eur = (1+r_usd)/(1+r_fx)-1.
func wpeaBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	recon, err := msciWorld(0.0020, composite("WPEA (MSCI World replication)", []Leg{
		{ID: "VFINX", Weight: 0.60},
		{ID: "VTMGX", Weight: 0.40},
	}, "", 0.0020))(f, from)
	if err != nil {
		return nil, err
	}
	// Ride real iShares Core MSCI World over the mid-period when reachable; if
	// it is not, the reconstruction alone still yields a valid EUR MSCI World.
	usd := recon
	if real, rerr := f.Fetch("IE00B4L5Y983", from); rerr == nil && real != nil && len(real.Points) > 300 {
		grafted := *real
		marketdata.ExtendBack(&grafted, recon)
		usd = &grafted
	}
	dates := make([]time.Time, len(usd.Points))
	lvl := make([]float64, len(usd.Points))
	for i, p := range usd.Points {
		dates[i], lvl[i] = p.Date, p.Close
	}
	return convertDaily("WPEA (MSCI World net TR expressed in EUR)",
		extend(f), "EURUSD=X", from, dates, lvl)
}

// eresMondeCharge is what the FCPE wrapper costs on top of the two ETFs it
// holds, a fraction per year: 0.35 % management (the Part M maximum, charged
// in full) plus 0.06 % of transaction costs per the FY2025 report, both of
// which the fund's NAV bears and neither of which its ETFs do. The ETFs' own
// TERs (0.21 % induced for the 75/25 mix) are NOT here: each leg below is a
// real ETF NAV, already net of them. Measured over 2024-03 to 2026-08 the NAV
// lagged the two ETFs' true total return by about 0.75 %/yr, the sum of these
// charges and the small cash line the fund keeps; the a-priori figure is kept,
// the residual being within the measurement's noise on two and a half years.
const eresMondeCharge = 0.0035 + 0.0006

// eresMondeLegs are the FCPE's two holdings, each with the chain of real
// series standing behind it (nearest first) and its published TER. The fund's
// own class comes first, its accumulating sibling behind (same portfolio,
// same swap or basket, another fee), and the MSCI World EUR path behind both.
// Every donor is lifted to the class's TER (feeUplift), never to close a
// measured gap: DBXW carries 0.45 %/yr against XWD1's 0.19, so its years are
// credited +0.26 %/yr, which is the fee difference and nothing else.
var eresMondeLegs = []eresLeg{
	{Weight: 0.75, TER: 0.0019, Donors: []eresDonor{{"XWD1.DE", 0.0019}, {"DBXW.DE", 0.0045}}},
	{Weight: 0.25, TER: 0.0012, Donors: []eresDonor{{"XDWL.DE", 0.0012}, {"XDWD.DE", 0.0012}}},
}

// eresLeg is one holding of the FCPE: its weight, its own TER and the real
// series that stand behind it.
type eresLeg struct {
	Weight float64
	TER    float64
	Donors []eresDonor
}

// eresDonor is a real ETF class and its published TER (fraction per year).
type eresDonor struct {
	ID  string
	TER float64
}

// eresMondeRecipe builds ERES Xtrackers Actions Monde M, the world-equity FCPE
// of Eres employee-savings plans (no ISIN; the company's share code
// 990000135629 is its identifier). The fund holds permanently 75 % Xtrackers
// MSCI World Swap UCITS ETF 1D (LU2263803533) and 25 % Xtrackers MSCI World
// UCITS ETF 1D (IE00BK1PV551), reinvests their distributions and charges its
// own wrapper fee on top; its NAV values each ETF at the ETF's official NAV,
// i.e. MSCI World with every market at its own close of the day.
//
// The reconstruction is therefore the daily-rebalanced 75/25 blend of the two
// ETFs' total-return paths less eresMondeCharge, each ETF read from its own
// class for as long as it quotes (2021-03 for the swap 1D class, 2015-04 for
// the physical), from its accumulating sibling before (DBXW.DE from 2008-01,
// lifted by the 0.26 %/yr fee difference; XDWD.DE from 2014-08, same TER), and
// from the MSCI World net-TR-in-EUR path (wpeaBuild: real IWDA from 2009,
// MSCIWORLD-USD refdata and the daily index shape before, EURUSD spot back to
// 1971) behind both, lifted from IWDA's 0.20 % to each class's TER. The real
// NAVs, served live by the airfund source and bundled as ERESMONDEM-NAV, are
// grafted on top from 2024-03-05, so the level is the fund's own.
//
// Two limits are stated rather than discovered. Xetra closes at 17:30 CET while
// the NAV is struck after New York closes, so the daily texture of the donor
// years carries a half-session timing smear (daily correlation ~0.6 against the
// real NAV, whole-window level within the charge's noise); it is invisible at
// the monthly cadence a backtest reads. And before 2008 the swap leg's edge
// over the net index (the swap ETF outran its physical sibling by ~0.35 %/yr
// over 2024-2026) is not modelled, so the deep tail is, if anything, a touch
// conservative.
func eresMondeRecipe() Recipe {
	return Recipe{
		ID:              "ERESMONDEM",
		Name:            "ERES Xtrackers Actions Monde M (FCPE): 75 % Xtrackers MSCI World Swap 1D + 25 % Xtrackers MSCI World 1D, wrapper charge 0.41 %/yr",
		Method:          "daily-rebalanced 75/25 blend of the two Xtrackers MSCI World ETFs the fund holds (each read from its own 1D class, its 1C sibling before it lifted to the class's TER, the MSCI World net TR in EUR path before both, back to 1971) less the FCPE's 0.41 %/yr wrapper charge; real NAVs grafted from 2024-03-05",
		Build:           eresMondeBuild,
		ValidateAgainst: "ERESMONDEM",
		SpliceReal:      "ERESMONDEM",
	}
}

// eresMondeBuild assembles the two legs and blends them.
func eresMondeBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	world, err := wpeaBuild(f, from)
	if err != nil {
		return nil, err
	}
	const worldTER = 0.0020 // IWDA's, carried by the whole wpeaBuild path
	fetched := map[string]*marketdata.Series{}
	var legs []Leg
	for _, l := range eresMondeLegs {
		chain, err := eresLegChain(f, from, l, world, worldTER)
		if err != nil {
			return nil, err
		}
		id := "ERESMONDEM-LEG-" + l.Donors[0].ID
		chain.Symbol = id
		fetched[id] = chain
		legs = append(legs, Leg{ID: id, Weight: l.Weight})
	}
	blend, err := composite("ERESMONDEM (75/25 blend of the two ETF legs)", legs, "", 0)(seriesFetcher{fetched, f}, from)
	if err != nil {
		return nil, err
	}
	blend.Currency = "EUR"
	return afterAnnualFee("ERESMONDEM (blend less the FCPE's wrapper charge)", blend, eresMondeCharge), nil
}

// eresLegChain builds one leg: the nearest donor that quotes, each further
// donor spliced behind it, the world path behind them all, every donor's
// returns lifted from its own TER to the leg's. A donor that cannot be
// fetched, quotes too briefly to matter, or quotes in another currency is
// skipped with a note, so the recipe still builds (offline, or should a
// listing die) on the deeper series behind it.
func eresLegChain(f Fetcher, from time.Time, l eresLeg, world *marketdata.Series, worldTER float64) (*marketdata.Series, error) {
	var out *marketdata.Series
	for _, d := range l.Donors {
		s, err := f.Fetch(d.ID, from)
		if err != nil || s == nil || len(s.Points) < 300 {
			fmt.Fprintf(os.Stderr, "ERESMONDEM: donor %s unavailable, the leg reads the next series behind it\n", d.ID)
			continue
		}
		if s.Currency != "" && s.Currency != "EUR" {
			return nil, fmt.Errorf("ERESMONDEM: donor %s quotes in %s, want EUR", d.ID, s.Currency)
		}
		lifted := afterAnnualFee(d.ID, s, feeUplift(l.TER, d.TER))
		if out == nil {
			out = lifted
			continue
		}
		marketdata.ExtendBack(out, lifted)
		out.SimulatedBefore = time.Time{} // let ExtendBack splice one more segment
	}
	back := afterAnnualFee("MSCI World net TR in EUR", world, feeUplift(l.TER, worldTER))
	back.SimulatedBefore = time.Time{}
	if out == nil {
		return back, nil
	}
	marketdata.ExtendBack(out, back)
	out.SimulatedBefore = time.Time{}
	return out, nil
}

// feeUplift is the (negative) annual charge that puts a donor carrying
// donorTER on a class carrying targetTER: the donor is credited the fee it
// paid in excess of the target's, and owes nothing when it was the cheaper
// (a donor may lose a wrapper's cost advantage, never gain a return it did
// not earn, as the trend price list states).
func feeUplift(targetTER, donorTER float64) float64 {
	if donorTER <= targetTER {
		return 0
	}
	return targetTER - donorTER
}

// seriesFetcher answers a few ids from memory and everything else from the
// fetcher behind it, so a series assembled in a recipe can enter a Frame.
type seriesFetcher struct {
	own      map[string]*marketdata.Series
	fallback Fetcher
}

func (sf seriesFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	if s, ok := sf.own[id]; ok {
		return s, nil
	}
	return sf.fallback.Fetch(id, from)
}

// eresDatadogCharge is the FCPE's all-in charge, a fraction per year: 0.61 %
// charged in FY2025 (management only; no induced cost, no transaction cost
// billed), against a 1.50 % maximum. Kept a priori, as every wrapper charge.
const eresDatadogCharge = 0.0061

// eresDatadogRecipe builds "Actions Datadog C", the single-stock FCPE of the
// Datadog France employee-savings plan (no ISIN; share code 990000124099):
// 90-100 % Datadog class A shares (DDOG, NASDAQ), the rest cash, NAV in EUR,
// weekly. The reconstruction is the DDOG price converted USD->EUR at the
// EURUSD spot (DDOG pays no dividend, so price is total return) less the
// wrapper charge, from the IPO of 2019-09-19, with the real weekly NAVs
// grafted on top from 2021-07-22. Nothing stands before the IPO: a single
// stock has no donor, and the file stops where its evidence stops.
//
// The fund values the share at the NASDAQ OPENING price of the valuation day
// (measured on the 293 NAVs: the returns fit the open at 2.0 % rmse against
// 3.6 % for the close, and on the 5-minute history the best-fitting instant is
// 09:30 New York, worsening monotonically to the close; -24.6 % on 2026-08-06
// against a -20.8 % close), so the engine, built on closes, disagrees with the
// real series by a session on every valuation day. The NAV was weekly until
// 2026-07-13 and is daily since, so the per-observation statistics of the
// weekly years read ~sqrt(5) off anyway: judge it on the monthly cadence.
func eresDatadogRecipe() Recipe {
	return Recipe{
		ID:              "ERES_DATADOG",
		Name:            "Actions Datadog C (FCPE): the DDOG share in EUR, wrapper charge 0.61 %/yr",
		Method:          "DDOG (NASDAQ, no dividend) converted USD->EUR at the EURUSD spot from the 2019-09-19 IPO, less the FCPE's 0.61 %/yr charge; real weekly NAVs grafted from 2021-07-22; nothing before the IPO",
		Build:           eresDatadogBuild,
		ValidateAgainst: "ERES_DATADOG",
		SpliceReal:      "ERES_DATADOG",
	}
}

// eresDatadogBuild converts the share price and deducts the charge.
func eresDatadogBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	ddog, err := f.Fetch("DDOG", from)
	if err != nil {
		return nil, err
	}
	if ddog == nil || len(ddog.Points) < 2 {
		return nil, fmt.Errorf("ERES_DATADOG: no DDOG history")
	}
	dates := make([]time.Time, len(ddog.Points))
	usd := make([]float64, len(ddog.Points))
	for i, p := range ddog.Points {
		dates[i], usd[i] = p.Date, p.Close
	}
	eur, err := convertDaily("ERES_DATADOG (DDOG expressed in EUR)", extend(f), "EURUSD=X", from, dates, usd)
	if err != nil {
		return nil, err
	}
	return afterAnnualFee("ERES_DATADOG (DDOG in EUR less the FCPE's charge)", eur, eresDatadogCharge), nil
}

// afterAnnualFee returns a copy of s with a continuous annual fee deducted:
// each daily step keeps its own return less the fee accrued over the days
// elapsed, so the drag compounds on the calendar rather than on quote count
// (holidays and weekends carry fees too). A NEGATIVE annual is an uplift, for
// a donor share class measured to carry a cost the target class does not (see
// aqrBEURFeeWedge).
func afterAnnualFee(name string, s *marketdata.Series, annual float64) *marketdata.Series {
	return afterFeeSteps(name, s, []feeStep{{Annual: annual}})
}

// feeStep is one segment of a piecewise-constant charge: Annual is in force
// from From (inclusive) until the next step. A reconstruction needs the
// schedule rather than a constant because its donors change nature partway
// through, a fee-free index covering what comes before a fund's own NAVs (see
// feeGap).
type feeStep struct {
	From   time.Time
	Annual float64
}

// afterFeeSteps is afterAnnualFee with a schedule: the same calendar-day
// compounding, with the charge read off steps at each date. Steps must be
// sorted by From; a date before the first step pays that first step.
func afterFeeSteps(name string, s *marketdata.Series, steps []feeStep) *marketdata.Series {
	out := *s
	out.Name = name
	out.Points = make([]marketdata.Point, len(s.Points))
	if len(s.Points) == 0 {
		return &out
	}
	out.Points[0] = s.Points[0]
	level := s.Points[0].Close
	for i := 1; i < len(s.Points); i++ {
		days := s.Points[i].Date.Sub(s.Points[i-1].Date).Hours() / 24
		step := s.Points[i].Close / s.Points[i-1].Close
		level *= step * math.Pow(1-feeAt(steps, s.Points[i].Date), days/365.25)
		out.Points[i] = marketdata.Point{Date: s.Points[i].Date, Close: level}
	}
	return &out
}

// feeAt returns the charge in force on d, 0 when the schedule is empty.
func feeAt(steps []feeStep, d time.Time) float64 {
	annual := 0.0
	for i, s := range steps {
		if i == 0 || !d.Before(s.From) {
			annual = s.Annual
		}
	}
	return annual
}

// sp500IndexRecipe is the pure S&P 500 Total Return benchmark: SP500-USD
// refdata with the ^GSPC daily shape, no fund fee, served as the
// non-investable SP500 index (fees 0, no ISIN). Its CAGR runs about a
// tracker's TER above IE00BFMXXD54 by design; correlation stays ~1.0.
func sp500IndexRecipe() Recipe {
	return Recipe{
		ID:              "SP500",
		Name:            "S&P 500 Total Return (index, fee-free)",
		Method:          "S&P 500 total return: SP500-USD refdata (monthly ~1871→) with the daily shape of the S&P 500 price index (^GSPC, 1927→), no fund fee; fallback VFINX (1976→)",
		Build:           sp500Index(0, composite("S&P 500 (VFINX)", []Leg{{ID: "VFINX", Weight: 1}}, "", 0)),
		ValidateAgainst: "IE00BFMXXD54",
	}
}

// wintonRecipe rebuilds the Winton Trend-Enhanced Global Equity fund as
// global equities (60/40 US/international) plus a half-sized self-generated
// TSMOM trend overlay, net of what its 0.80%/yr load exceeds its equity
// donors' own (wintonFee): the two Vanguard funds charge 0.60×0.14 +
// 0.40×0.05 = 0.112 %/yr inside their NAVs, and the file starts at the trend
// reference's 2000-01, where both already quote, so the charge is a constant.
// The overlay leg is left as its net reference gives it, for the reason
// stackedTrend states.
func wintonRecipe() Recipe {
	return Recipe{
		ID:              "IE000O1VI174",
		Name:            "Winton Trend-Enhanced Global Equity: equities + TSMOM overlay",
		Method:          "0.60×VFINX + 0.40×VTMGX + 0.50×(TSMOM trend − overnight financing, its months and its level from the bundled net pure-trend reference), 0.80%/yr fees less the equity donors' own 0.112%/yr (2000→, where that reference begins)",
		Build:           wintonBuild,
		ValidateAgainst: "IE000O1VI174",
	}
}

// wintonBuild blends a 60/40 equity core with a half-weighted TSMOM trend
// overlay (the trend run as a pure excess strategy, no collateral).
func wintonBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	ids := append([]string{"^IRX", usdOvernight, "VFINX", "VTMGX"}, mfMarkets...)
	fr, err := BuildFrame(financed(extend(f)), ids, from)
	if err != nil {
		return nil, err
	}
	cfg := mfConfig(0.10, 0)
	cfg.EarnCash = false
	trend, start, err := TSMOM(fr, cfg)
	if err != nil {
		return nil, err
	}
	// The overlay runs as a pure excess strategy (EarnCash=false), so its own
	// return is the excess. Anchor its monthly path AND its level on the net
	// pure-trend reference, which is the record of the trade this sleeve
	// replicates: the reference is funded, so its own cash leg is stripped
	// (TrendAnchor.Funded) against the frame's real cash accruals before it is
	// rescaled, and not added back, since the output is an overlay.
	cash := fr.Returns[cfg.CashID][start:]
	trend, err = AnchorTrend(f, PureTrendAnchor, fr.Dates[start:], trend, cash, cfg.TargetVol, cfg.EarnCash)
	if err != nil {
		return nil, err
	}
	vfinx, vtmgx := fr.Returns["VFINX"], fr.Returns["VTMGX"]
	fin := fr.Returns[usdOvernight]
	const (
		wintonUS  = 0.60
		wintonTER = 0.0080
		feeDaily  = max(0, wintonTER-(wintonUS*vfinxTER+(1-wintonUS)*vtmgxTER)) / 252
	)
	s := &marketdata.Series{Name: "Winton Trend-Enhanced Global Equity (replication)", Source: "simdata"}
	val := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[start], Close: val})
	for i := 1; i < len(trend); i++ {
		k := start + i
		rEq := wintonUS*vfinx[k] + (1-wintonUS)*vtmgx[k]
		// The overlay is already an excess over the bill rate (EarnCash=false
		// above), and a stack pays the overnight financing rate rather than
		// that bill rate, so it gives back the difference. Same convention as
		// stackedTrend, written on an excess rather than a funded record.
		rTrend := trend[i]/trend[i-1] - 1 - (fin[k] - cash[i])
		val *= 1 + rEq + 0.5*rTrend - feeDaily
		s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[k], Close: val})
	}
	return trimToAnchor(f, PureTrendAnchor, s)
}

// Find returns the recipe whose ID or validation target matches id.
func Find(id string) (Recipe, bool) {
	canonical := marketdata.CanonicalID(id)
	for _, r := range All() {
		if r.ID == canonical || r.ID == id {
			return r, true
		}
	}
	return Recipe{}, false
}

// mfMarkets is the cross-asset futures basket traded by the managed-futures
// trend reconstructions (equities, bonds, commodities; currencies omitted;
// no fetchable series). The youngest component (gold/oil futures, ~2000)
// sets the start date.
var mfMarkets = []string{"VFINX", "VTMGX", "VEIEX", "VFITX", "VUSTX", "GC=F", "CL=F"}

// mfConfig is the standard 12-month time-series-momentum configuration, with
// a per-fund volatility target and fee. The engine realizes the target it is
// given (its risk model rescales daily, see TSMOMConfig), so each target is
// simply the volatility the fund itself realized over the window where both
// exist: DBMF 11.5 %, KMLM 14 %, CTA 16 %, AQR 9 %. Read them as measured
// anchors, not as the funds' stated targets, which are rounder and mostly
// unpublished.
func mfConfig(targetVol, annualFee float64) TSMOMConfig {
	return TSMOMConfig{
		Markets:     mfMarkets,
		CashID:      "^IRX",
		Lookback:    252,
		VolWindow:   63,
		Rebalance:   21,
		TargetVol:   targetVol,
		MaxLeverage: 2,
		AnnualFee:   annualFee,
		EarnCash:    true,
	}
}

// tsmom is the shared Build for the deep tail of the managed-futures
// reconstructions: it builds a frame on the markets, runs the TSMOM engine for
// the daily texture, and anchors every month of it on the bundled NET trend
// reference (NetTrendAnchor), rescaled to the fund's volatility target. The
// result is aligned to the dates after the signal warm-up.
//
// There is no information-ratio pin here, and that is the point. The engine
// earns an in-sample IR of ~0.5-0.85 that no real managed-futures programme has
// sustained, and a pin existed to take it away; but the anchor carries a
// composite of real programmes already NET of their managers' fees, so the
// level it imposes is an investable one. Dragging it further would charge the
// fee twice, and a constant drag is in any case the wrong instrument for the
// job: it reproduces an index's information ratio while deepening every
// drought into a bleed. The overlay builds (stackedTrend, wintonBuild) reason
// the same way on their own narrower reference (PureTrendAnchor).
func tsmom(name string, cfg TSMOMConfig) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		fr, err := BuildFrame(extend(f), append([]string{cfg.CashID}, cfg.Markets...), from)
		if err != nil {
			return nil, err
		}
		values, start, err := TSMOM(fr, cfg)
		if err != nil {
			return nil, err
		}
		dates, cash := fr.Dates[start:], fr.Returns[cfg.CashID][start:]
		values, err = AnchorTrend(f, NetTrendAnchor, dates, values, cash, cfg.TargetVol, cfg.EarnCash)
		if err != nil {
			return nil, err
		}
		s := &marketdata.Series{Name: name, Source: "simdata"}
		for i, v := range values {
			s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[start+i], Close: v})
		}
		return s, nil
	}
}

// No reconstruction in this file is levelled by hand any more. Every one of
// them takes its level from an anchor that is already investable: the
// diversified funds from a net managed-futures composite (NetTrendAnchor), the
// overlays from a net pure-trend one (PureTrendAnchor). A drag on top of a net
// record would charge the constituent managers' fees twice.
//
// The information ratios that used to do that levelling, and the measurement
// behind each of them, are recorded in docs/trend-reconstruction-design.md:
// they are what any future claim about these funds' level has to be argued
// against, and they are a measurement rather than a computation, so that is
// where they live.

// stackedTrend backcasts a Return Stacked fund: a funded core (coreID, an equity
// or bond index deep via refdata) plus a 100 % managed-futures overlay stacked
// on top. The overlay is the same TSMOM engine used for KMLM/DBMF/CTA (cfg),
// which a plain composite leg (limited to the real fund's 2020s inception)
// cannot reach. The engine earns cash on its collateral (cfg.EarnCash), so its
// excess over cash is the trend overlay, and the stacked return is
// r = coreReturn + trendExcess − fee.
//
// Both the overlay's months and its LEVEL come from the net pure-trend
// reference (PureTrendAnchor), which is a published record of the trade this
// sleeve replicates, already net of the constituent managers' fees. Nothing
// levels it afterwards: it is already an investable level, and the engine's own
// in-sample information ratio never reaches the output. The price of that is
// length, and it is the price this family accepts: the file starts where the
// reference does rather than at the engine's own 1989 floor (trimToAnchor).
//
// # What the fee is charged on, and what it is not
//
// coreLoad is the ongoing charge already inside the CORE donor's own NAV, and
// the fund's fee is charged less that much: a Vanguard index fund standing in
// for a sleeve the fund holds directly has been billed once already, and
// annualFee on top of it bills it twice (the doctrine is spelled out at
// feeGap). Both cores quote over the whole file, which starts at the
// reference's 2000-01, so the charge is a constant and needs no schedule.
//
// The OVERLAY leg is deliberately left alone, and the number says why. Its
// reference is a composite of real programmes whose returns arrive net of their
// managers' fees, estimated at 2 %/yr in trendFeeLoad because those managers
// publish no schedule. Reading that estimate as a donor load would hand the
// overlay back about two points a year, more than every other correction in
// this file put together, on a figure that is not a price list; and the two
// funds carrying this build disagree on the sign of the residual over their own
// live windows (the reconstruction runs cold against RSST and hot against
// RSBT). An estimate that large, arbitrated by nothing, stays out.
func stackedTrend(name, coreID string, coreLoad float64, cfg TSMOMConfig, annualFee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		ids := append([]string{coreID, cfg.CashID, usdOvernight}, cfg.Markets...)
		fr, err := BuildFrame(financed(extend(f)), ids, from)
		if err != nil {
			return nil, err
		}
		trend, start, err := TSMOM(fr, cfg)
		if err != nil {
			return nil, err
		}
		trend, err = AnchorTrend(f, PureTrendAnchor, fr.Dates[start:], trend, fr.Returns[cfg.CashID][start:], cfg.TargetVol, cfg.EarnCash)
		if err != nil {
			return nil, err
		}
		core, fin := fr.Returns[coreID], fr.Returns[usdOvernight]

		// Daily overlay of the anchored reconstruction: a FUNDED trend record
		// (cfg.EarnCash) less what the stack pays to carry that notional, which
		// is the overnight financing rate and not the bill rate the programme's
		// own collateral earns. The two rates part company by 0.01 to 0.15
		// points a year on these funds' live windows (see usdOvernight); the
		// engine and the anchor keep reading cfg.CashID, because the reference
		// they work on is a funded record of a book that holds bills.
		overlay := make([]float64, len(trend))
		for i := 1; i < len(trend); i++ {
			overlay[i] = (trend[i]/trend[i-1] - 1) - fin[start+i]
		}

		feeDaily := max(0, annualFee-coreLoad) / 252
		values := make([]float64, len(trend))
		values[0] = 100
		for i := 1; i < len(trend); i++ {
			k := start + i
			r := core[k] + overlay[i] - feeDaily
			values[i] = values[i-1] * (1 + r)
		}
		s := &marketdata.Series{Name: name, Source: "simdata"}
		for i, v := range values {
			s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[start+i], Close: v})
		}
		return trimToAnchor(f, PureTrendAnchor, s)
	}
}

// trimToAnchor cuts a reconstruction back to the first date its reference
// covers. Reliability bounds length here: what stands in front of the reference
// is the engine alone, unanchored and unlevelled, and a shorter history that is
// worth reading beats a longer one that is not.
func trimToAnchor(f Fetcher, ref TrendAnchor, s *marketdata.Series) (*marketdata.Series, error) {
	start, err := AnchorStart(f, ref)
	if err != nil {
		return nil, err
	}
	out := marketdata.Trim(s, start, time.Time{})
	if len(out.Points) < 250 {
		return nil, fmt.Errorf("%s: only %d points from %s", ref.ID, len(out.Points), start.Format("2006-01-02"))
	}
	return out, nil
}

// msciWorldShapeID is the Yahoo daily MSCI World PRICE index (1972→). Its
// levels lag total return by the dividend yield (it carries no income), so
// it never sets levels: it only supplies the intra-month daily shape behind
// the monthly net-TR anchors (see anchorShape).
const msciWorldShapeID = "^990100-USD-STRD"

// msciWorld returns the Build shared by the MSCI World trackers (IWDA,
// URTH): the monthly net total-return index served as MSCIWORLD-USD
// refdata (1969→) sets the levels, the daily price index supplies the
// intra-month shape from 1972, and the tracker's TER is deducted last.
// The refdata file stays embedded/local, so without it everything falls
// back to the given fetchable proxy Build; without the daily shape the
// backcast simply stays monthly, and a shape that stops short (a
// truncated fetch) blends what it covers while the later anchors keep
// their monthly cadence (shapedSeries never drops them).
// The >300-point guards distinguish the real long series from an
// accidental short fetch of the same symbol.
func msciWorld(annualFee float64, fallback func(Fetcher, time.Time) (*marketdata.Series, error)) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return shapedIndex("MSCIWORLD-USD", msciWorldShapeID, annualFee, fallback)
}

// btop50IndexRecipe serves the managed-futures reference AS ITSELF, the way
// MSCIWORLD serves the equity one: the monthly net composite of real
// programmes (TREND-NET-USD, 1986-12→, each constituent already net of its own
// manager's fees) with the daily texture of the net pure-trend composite
// (TREND-PURE-NET-USD, 2000-01→) inside each month.
//
// It exists because every managed-futures RECONSTRUCTION in this package stops
// at 1996-03, the first NAV of the deepest real donor, which puts the 1987
// crash, 1990 and the 1994 bond rout out of reach of any book carrying a trend
// sleeve. This id reaches 1986-12 instead, and it is a different object from
// those reconstructions, deliberately: nothing here is rescaled to a fund's
// volatility target. The index runs at its own ~9.3% volatility against the
// ~15% a UCITS trend fund targets, so a sleeve held through this line carries
// roughly 60% of the risk the real sleeve would; that understatement is the
// price of the extra decade, and it is the honest direction to err in (see
// "The tail that was removed" in docs/trend-reconstruction-design.md, where
// rescaling an index to a fund's target is exactly what discredited the older
// 1988 tail). Non-investable: no ISIN, no fund fee added, and the index's
// annual-rebalanced equal weighting of the 50% largest programmes is not
// something a household can buy.
func btop50IndexRecipe() Recipe {
	return Recipe{
		ID:     "BTOP50",
		Name:   "Barclay BTOP50 managed futures (index, net of manager fees)",
		Method: "monthly BTOP50 net composite (TREND-NET-USD refdata, 1986-12→) with the daily shape of the net pure-trend composite (TREND-PURE-NET-USD, 2000-01→); no fund fee added and no volatility rescaling, the index is served at its own level and its own ~9.3% volatility",
		Build:  btop50Local,
	}
}

// btop50Local is the index in its own currency, shared by the raw and the
// hedged recipes. Monthly before 2000 (the daily shape donor's start), daily
// after, exactly as MSCIWORLD is monthly before 1972.
var btop50Local = shapedIndex(NetTrendAnchor.ID, PureTrendAnchor.ID, 0,
	func(Fetcher, time.Time) (*marketdata.Series, error) {
		return nil, fmt.Errorf("%s: refdata missing, no fallback (an index of hedge-fund programmes cannot be replicated from prices)", NetTrendAnchor.ID)
	})

// btop50HedgedIndexRecipe is the same index expressed in EUR with the currency
// risk hedged away, which is the form a euro book actually compares against:
// every investable trend line a European household can buy is either a EUR
// share class or EUR-hedged (DBMFE, RAEF). Holding the raw USD index instead
// would put ~10%/yr of EURUSD volatility and four decades of dollar drift into
// what is supposed to be a test of the trend sleeve.
//
// The arithmetic is the standard hedged-return identity, the same one dtleBuild
// applies to the long-Treasury segment: the index is a FUNDED total return
// (its collateral earns USD cash), so hedging strips that cash leg and gives
// back the euro one, day by day, which is what a rolling forward does.
func btop50HedgedIndexRecipe() Recipe {
	return Recipe{
		ID:     "BTOP50E",
		Name:   "Barclay BTOP50 managed futures, hedged to EUR (index)",
		Method: "the BTOP50 net composite (see BTOP50) financed at USD cash (^IRX, extended by TBILL-3M) and re-earning euro cash (the deep euro overnight chain, German money market before the euro) = the EUR-hedged view of the index; no fund fee added, no volatility rescaling",
		Build:  btop50HedgedBuild,
	}
}

// btop50HedgedBuild turns the USD index into its EUR-hedged view through the
// shared hedged-return identity (hedgeToEUR).
func btop50HedgedBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	local, err := btop50Local(f, from)
	if err != nil {
		return nil, err
	}
	return hedgeToEUR("BTOP50 hedged to EUR (index)", local, f, from)
}

// sp500ShapeID is the Yahoo daily S&P 500 PRICE index (1927→): like the MSCI
// World price index it carries no dividends, so it only supplies the daily
// shape behind the monthly SP500-USD total-return anchors, never the levels.
const sp500ShapeID = "^GSPC"

// sp500Index is the S&P 500 counterpart of msciWorld: the SP500-USD total
// return refdata (~1871) sets the levels, ^GSPC supplies the daily shape from
// 1927, and any fee is deducted last (0 for the pure index benchmark).
func sp500Index(annualFee float64, fallback func(Fetcher, time.Time) (*marketdata.Series, error)) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return shapedIndex("SP500-USD", sp500ShapeID, annualFee, fallback)
}

// shapedIndex builds a daily total-return index from a coarse (monthly)
// total-return anchor refdata series plus the daily intra-period shape of a
// matching price index, applying an optional annual fee last so a pre-fee
// index level can become an after-cost investable one. It falls back to the
// given fetchable proxy Build when the anchor refdata is absent (it stays
// local), and stays at the anchor's own cadence when the shape is missing;
// shapedSeries never drops anchors a truncated shape fails to cover. The
// >300-point guards distinguish the real long series from a short fetch of
// the same symbol.
//
// The anchor levels are month-END total returns (a point for month M is that
// month's closing level); alignMonthEnd snaps each onto the shape's own last
// trading day of the month first, so the monthly level is pinned to the real
// month-end close and the daily shape fills the days between. Skip that step and
// a calendar-month-end anchor whose end lands on a weekend pins one trading day
// into the next month, sliding the whole reconstruction (this is exactly how a
// naive blend understated the 2022 drawdown by ten points).
func shapedIndex(anchorID, shapeID string, annualFee float64, fallback func(Fetcher, time.Time) (*marketdata.Series, error)) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		anchor, err := f.Fetch(anchorID, from)
		if err != nil || anchor == nil || len(anchor.Points) <= 300 {
			return fallback(f, from)
		}
		out := anchor
		if shape, serr := f.Fetch(shapeID, from); serr == nil && shape != nil && len(shape.Points) > 300 {
			shape.Points = despike(shape.Points)
			out = shapedSeries(alignMonthEnd(anchorID, anchor, shape), shape)
		}
		return afterFee(out, annualFee), nil
	}
}

// afterFee returns a copy of s with a continuous annual fee applied, so a
// pre-fee index level becomes an after-cost investable one.
func afterFee(s *marketdata.Series, annual float64) *marketdata.Series {
	if annual <= 0 || len(s.Points) == 0 {
		return s
	}
	out := *s
	out.Points = make([]marketdata.Point, len(s.Points))
	t0 := s.Points[0].Date
	for i, p := range s.Points {
		yrs := p.Date.Sub(t0).Hours() / 24 / 365.25
		out.Points[i] = marketdata.Point{Date: p.Date, Close: p.Close * math.Pow(1-annual, yrs)}
	}
	return &out
}

// composite is the shared Build for constant-weight linear recipes. cashID is
// the rate the Excess legs FINANCE at, not the rate a cash leg earns: a
// collateral sleeve is an ordinary leg of its own (see usdOvernight for why the
// two are different rates).
func composite(name string, legs []Leg, cashID string, fee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		var ids []string
		if cashID != "" {
			ids = append(ids, cashID)
		}
		for _, l := range legs {
			ids = append(ids, l.ID)
		}
		fr, err := BuildFrame(financed(extend(f)), ids, from)
		if err != nil {
			return nil, err
		}
		values, err := Composite(fr, legs, cashID, fee)
		if err != nil {
			return nil, err
		}
		return SeriesFromFrame(name, fr, values), nil
	}
}

// A donor's NAV is ALREADY NET of its own manager's charges, so a recipe that
// deducts its target's whole ongoing charge on top of fund donors bills the
// reconstruction twice. What a target still owes is the DIFFERENCE between its
// own load and its donors', floored at zero: a donor may lose its wrapper's
// cost advantage, never gain a return it did not earn. avantisRecipe states the
// doctrine at length; these constants are its price list.
//
// Every figure is the vehicle's current published ongoing charge (2026), read
// off the fund page and NEVER off an observed return gap, which would grant the
// donor's manager whatever skill separates the two records:
//
//	IWDA   0.20 %  iShares Core MSCI World UCITS ETF (IE00B4L5Y983)
//	VFINX  0.14 %  Vanguard 500 Index Fund Investor Shares
//	VFITX  0.20 %  Vanguard Intermediate-Term Treasury Fund Investor Shares
//	VUSTX  0.20 %  Vanguard Long-Term Treasury Fund Investor Shares
//	VTMGX  0.05 %  Vanguard Developed Markets Index Fund Admiral Shares
//	VEIEX  0.29 %  Vanguard Emerging Markets Stock Index Fund Investor Shares
//	EZU    0.50 %  iShares MSCI Eurozone ETF
//	EUNH   0.07 %  iShares Core € Govt Bond UCITS ETF (IE00B4WXJJ64)
//	DTLA   0.07 %  iShares $ Treasury Bond 20+yr UCITS ETF USD Acc (IE00BFM6TC58)
//
// The Dimensional pair keeps its own constants (dfsvxTER, disvxTER) next to the
// recipe that first priced it. Two loads are deliberately absent: a rate (^IRX,
// EURCASH-EUR), a futures price (GC=F) and an index (^BCOM, the refdata
// reconstructions) charge nothing, so a target's whole load is due on them.
const (
	iwdaTER  = 0.0020
	vfinxTER = 0.0014
	vfitxTER = 0.0020
	vustxTER = 0.0020
	vtmgxTER = 0.0005
	ezuTER   = 0.0050
	eunhTER  = 0.0007
	dtlaTER  = 0.0007
)

// pricedLeg is a composite leg plus the ongoing charge that already sits inside
// its donor's own NAV, on the leg's own notional (a 0.60 leg of a fund charging
// 0.20 %/yr costs the composite 0.12). Load is 0 for a rate, a futures price or
// an index, which charge nothing.
type pricedLeg struct {
	ID     string
	Weight float64
	Excess bool
	Load   float64
}

func (p pricedLeg) leg() Leg { return Leg{ID: p.ID, Weight: p.Weight, Excess: p.Excess} }

// feeGap is what a reconstruction still owes its target's price list: the
// target's load less the loads its donors already carry, floored at zero.
//
// It is a SCHEDULE and not a constant because a deep composite's donors change
// nature partway through. A leg extended by longBack rides a fee-free index or
// CMT reconstruction before the fund's own quotes begin (1980-01 for VFINX,
// 1991-10 for VFITX, 1986-05 for VUSTX, 1999-08 for VTMGX, 2000-07 for EZU,
// 2009-04 for EUNH.DE), and over that era the target's whole charge IS due,
// since nothing has been deducted yet. Each donor's step therefore falls on its
// own first quote, read from the RAW fetcher (the one Build receives, before
// composite wraps it with extend), so the boundary is the fund's real inception
// rather than a date written down here.
func feeGap(f Fetcher, from time.Time, target float64, legs []pricedLeg) ([]feeStep, error) {
	steps := []feeStep{{Annual: target}}
	carried := 0.0
	type entry struct {
		date time.Time
		load float64
	}
	var starts []entry
	for _, l := range legs {
		if l.Load == 0 {
			continue
		}
		s, err := f.Fetch(l.ID, from)
		if err != nil {
			return nil, fmt.Errorf("donor %s: %w", l.ID, err)
		}
		if s == nil || len(s.Points) == 0 {
			return nil, fmt.Errorf("donor %s: empty history", l.ID)
		}
		starts = append(starts, entry{s.Points[0].Date, l.Weight * l.Load})
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i].date.Before(starts[j].date) })
	for _, st := range starts {
		carried += st.load
		steps = append(steps, feeStep{From: st.date, Annual: max(0, target-carried)})
	}
	return steps, nil
}

// alignedComposite is composite() charging feeGap instead of the target's whole
// ongoing charge: the same constant-weight linear build, then the schedule
// deducted on the calendar (afterFeeSteps) rather than per trading day, as
// every other fee in this file is.
func alignedComposite(name string, legs []pricedLeg, cashID string, target float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	plain := make([]Leg, len(legs))
	for i, l := range legs {
		plain[i] = l.leg()
	}
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		s, err := composite(name, plain, cashID, 0)(f, from)
		if err != nil {
			return nil, err
		}
		steps, err := feeGap(f, from, target, legs)
		if err != nil {
			return nil, err
		}
		return afterFeeSteps(name, s, steps), nil
	}
}

// ntsxRecipe rebuilds the WisdomTree US Efficient Core (90 % US equities +
// 60 % treasury futures ladder) from Vanguard index funds and two rates: the
// Treasury overlay FINANCES at the USD overnight rate (usdOvernight, what a
// future's implied repo costs) while the 10 % collateral sleeve EARNS the
// T-bill rate (^IRX, what the fund's own bills pay). The simdata file extends
// the UCITS share class (IE000KF370H3); the validation runs against the
// US-listed twin NTSX, which has traded since 2018 with the exact same
// strategy.
//
// The fund's 0.20 %/yr is charged through feeGap, so the two Vanguard donors
// pay it only where their own NAVs do not already carry a charge. Their
// wrappers alone cost 0.90×0.14 + 0.60×0.20 = 0.246 %/yr, more than the fund
// itself, so from VFITX's 1991 start the reconstruction owes nothing further;
// over 1980-1991 it owes 0.20 − 0.126 = 0.074, and before VFINX's own quotes
// the whole 0.20, the legs being fee-free reconstructions there.
func ntsxRecipe() Recipe {
	return Recipe{
		ID:     "IE000KF370H3",
		Name:   "WisdomTree US Efficient Core: 90/60 replication",
		Method: "0.90×VFINX + 0.60×(VFITX − overnight financing: SOFR 2018→, effective fed funds 1954→, T-bill before) + 0.10×cash ^IRX, daily rebalancing, the 0.20%/yr TER charged only where the donors' own charges (0.246%/yr blended) do not already cover it",
		Build: alignedComposite("NTSX (90/60 replication)", []pricedLeg{
			{ID: "VFINX", Weight: 0.90, Load: vfinxTER},
			{ID: "VFITX", Weight: 0.60, Excess: true, Load: vfitxTER},
			{ID: "^IRX", Weight: 0.10},
		}, usdOvernight, 0.0020),
		ValidateAgainst: "NTSX-US",
		SpliceReal:      "NTSX-US",
	}
}

// ntsgRecipe is the global variant (NTSG UCITS, WisdomTree Global Efficient
// Core, real from 2024-11): 90 % global developed equities + a 60 % government
// bond futures overlay + 10 % T-bill collateral, all in USD.
//
// It used to rebuild both of those from US mutual funds: a frozen 60/40 US /
// developed-ex-US equity split and a bond overlay entirely on the intermediate
// Treasury fund. Both were wrong in a way the fund's own disclosures settle, and
// the reconstruction paid for it on its whole validation window (daily
// correlation 0.40, tracking error 1.11 times the fund's volatility):
//
//   - the equity book is the MSCI World universe, ~72 % US, not 60. Twelve
//     points of US underweight is a large factor bet to run inside a
//     replication, and it is unnecessary: this repository already carries the
//     MSCI World net total return to 1969 (MSCIWORLD-USD plus the daily price
//     shape ^990100-USD-STRD), and the real iShares Core MSCI World (IWDA,
//     2009-09 on) to ride the recent decade.
//   - the bond basket is a FOUR-CURRENCY government futures basket, 80 % US /
//     11 % German / 6 % Japanese / 3 % British, every leg near a ten-year
//     duration. The whole of globalbond.go answers for that, including why the
//     foreign legs are local excess returns carrying no currency at all.
//
// The equity leg also fixes the other half of the daily-correlation problem.
// The fund prices at a European valuation point, and its old donors were US
// mutual fund NAVs struck hours later; IWDA closes in London, so from 2009 the
// two are struck within the same session.
//
// # The approximation that remains, and its sign
//
// The fund tracks the WisdomTree Global Efficient Core Index (Bloomberg
// WTNTSGN, net total return), whose equity universe is a proprietary top-1500
// developed-markets cap-weighted selection over the same 21 countries as MSCI
// World, with ESG exclusions applied. Plain MSCI World is therefore a proxy, not
// the index. It is an honest one, and the direction is known: MSCI's own
// comparison of its screened world index against the plain one runs +0.36 %/yr
// with 0.81 % tracking error over ten years, so the screen has if anything
// helped, and using the unscreened parent is the conservative choice rather than
// a flattering one. The fund's published book is 3 points lighter in the US than
// the plain index for the same reason.
//
// # Fees
//
// The 0.25 %/yr charge goes through the feeGap discipline, on the loads its
// donors already carry on their own notionals: IWDA 0.90×0.20 = 0.180 from
// 2009-09, VFITX 0.60×0.80×0.78×0.20 = 0.075 from 1991-10, VUSTX
// 0.60×0.80×0.22×0.20 = 0.021 from 1986-05, everything else (the reference
// reconstructions and the two rates) fee-free. So the schedule is 0.250 %/yr
// before 1986-05, 0.229 to 1991-10, 0.154 to 2009-09, and nothing after: past
// that date the donors alone charge 0.276 %/yr, more than the fund does.
//
// # Depth
//
// The floor is the deepest date every leg covers, which is the equity one:
// MSCIWORLD-USD opens in 1969-12 and the bond and collateral legs both reach the
// 1950s or before. Two sleeves of the bond basket open later than that (see
// globalbond.go), and the overlay renormalizes over what quotes rather than
// shortening the file.
func ntsgRecipe() Recipe {
	return Recipe{
		ID:     "IE00077IIPQ8",
		Name:   "WisdomTree Global Efficient Core: global 90/60 replication",
		Method: "0.90×MSCI World net TR (real IWDA from 2009-09, MSCIWORLD-USD refdata ~1969 with the ^990100 daily shape before) + 0.60×a four-currency government bond futures overlay (80% US VFITX/VUSTX duration blend − overnight financing, 11% BUND-EUR, 6% JGB-JPY, 3% GILT-GBP, each in local excess return over its own money-market rate, weights renormalized before a sleeve opens) + 0.10×cash ^IRX, the 0.25%/yr TER charged only where the donors' own charges do not already cover it; start set by the equity leg (~1969-12)",
		Build:  ntsgBuild,
		// The real fund opened in 2024-11: the overlap clears the 60-point floor
		// but not by much, so the card's CAGR comparison is indicative and the
		// correlations are what to read.
		ValidateAgainst: "IE00077IIPQ8",
		// The record standing behind the equity leg for fifteen of the file's
		// years, graded on its own overlap by the audit's chain panel.
		Donors: []string{"IE00B4L5Y983"},
	}
}

// The synthetic identifiers under which ntsgBuild serves its two pre-built legs
// to the frame. They never leave this build.
const (
	ntsgEquityID = "NTSG-EQ"
	ntsgBondID   = "NTSG-BOND"
)

// ntsgBuild assembles NTSG from a pre-built global equity leg and a pre-built
// global bond overlay, served to the standard frame/Composite machinery under
// synthetic ids (the same shape as ntszBuild). The collateral leg is the ordinary
// T-bill rate: it is cash the fund really holds, not financing, so it earns ^IRX
// and not the overnight rate the futures pay (see usdOvernight).
//
// The overlay already carries its own financing, sleeve by sleeve and currency by
// currency, so it enters the composite as a PLAIN leg at the fund's 0.60
// notional rather than as an Excess one: netting it a second time against the
// USD rate would charge the German, Japanese and British sleeves a financing
// cost in the wrong currency on top of the one they have already paid.
//
// The fee schedule is built on the REAL donors rather than on the synthetic
// ids, since it is IWDA's, VFITX's and VUSTX's own inceptions that end the
// fee-free reference era behind each of them (see ntsgRecipe for the
// arithmetic).
func ntsgBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	eq, err := ntsgEquityUSD(f, from)
	if err != nil {
		return nil, err
	}
	bonds, err := globalBondOverlay(f, from)
	if err != nil {
		return nil, err
	}
	inj := injected{inner: extend(f), have: map[string]*marketdata.Series{
		ntsgEquityID: eq,
		ntsgBondID:   bonds,
	}}
	legs := []Leg{
		{ID: ntsgEquityID, Weight: 0.90},
		{ID: ntsgBondID, Weight: 0.60},
		{ID: "^IRX", Weight: 0.10},
	}
	fr, err := BuildFrame(inj, []string{"^IRX", ntsgEquityID, ntsgBondID}, from)
	if err != nil {
		return nil, err
	}
	values, err := Composite(fr, legs, "^IRX", 0)
	if err != nil {
		return nil, err
	}
	steps, err := feeGap(f, from, 0.0025, []pricedLeg{
		{ID: "IE00B4L5Y983", Weight: 0.90, Load: iwdaTER},
		{ID: "VFITX", Weight: 0.60 * usdBondShare * usdShortShare, Load: vfitxTER},
		{ID: "VUSTX", Weight: 0.60 * usdBondShare * (1 - usdShortShare), Load: vustxTER},
	})
	if err != nil {
		return nil, err
	}
	name := "NTSG (global 90/60 replication)"
	return afterFeeSteps(name, SeriesFromFrame(name, fr, values), steps), nil
}

// ntsgEquityUSD builds the fund's equity book: the MSCI World net total return
// in USD, real iShares Core MSCI World (IWDA, 2009-09 on, London close and
// already net of its own 0.20 %/yr) grafted on top of the monthly MSCIWORLD-USD
// reconstruction with its daily price shape (1969-12 on, fee-free, the feeGap
// schedule charging the fund's own load over that era).
//
// Riding the real tracker over the recent decade is what makes the daily
// statistics mean anything: the fund is struck at a European valuation point,
// and a London close is the same session while a US mutual fund's NAV is not.
// If IWDA cannot be fetched the reconstruction alone still yields a valid MSCI
// World, one strictly coarser.
func ntsgEquityUSD(f Fetcher, from time.Time) (*marketdata.Series, error) {
	recon, err := msciWorld(0, composite("NTSG (MSCI World replication)", []Leg{
		{ID: "VFINX", Weight: 0.60},
		{ID: "VTMGX", Weight: 0.40},
	}, "", 0))(f, from)
	if err != nil {
		return nil, err
	}
	real, rerr := f.Fetch("IE00B4L5Y983", from)
	if rerr != nil || real == nil || len(real.Points) <= 300 {
		fmt.Fprintf(os.Stderr, "NTSG: real IWDA unavailable (%v), the MSCI World reconstruction stands alone\n", rerr)
		return recon, nil
	}
	grafted := *real
	marketdata.ExtendBack(&grafted, recon)
	return &grafted, nil
}

// ntszRecipe is the eurozone variant (NTSZ UCITS, WisdomTree Eurozone Efficient
// Core, launched 2025-09): 90% eurozone equities + 60% euro government bond
// futures, all EUR-denominated. Unlike NTSX/NTSG, whose deep past leans on
// long-running US index funds and USD refdata, every leg here is euro-native, so
// each is extended by a bundled euro-area reference series (gen-euro-refdata):
//
//   - equity: the real MSCI Eurozone (EZU, USD, 2000->) expressed in EUR, then
//     extended back with EMU-EUR (euro-area equity net TR, ~1986). This is the
//     deep floor of the composite, so it starts ~1986.
//   - bond: the real iShares Core Euro Govt Bond (EUNH.DE, EUR, 2009->) extended
//     by EUROGOV-EUR (euro-area 10y government bond TR, ~1970, with the ECB daily
//     shape EUROGOV-DAILY from 2004), run as a futures overlay financed at EUR
//     cash.
//   - cash: the euro money-market index (EURCASH-EUR, 1994->) carried to daily
//     granularity and extended before the euro by DECASH-EUR (the German 3-month
//     money-market accrual, ~1960).
//
// The real NTSZ quotes are grafted on top from inception; the overlap is short
// (the fund is months old), so the validation is thin and the value is the deep
// reconstruction, not a tight tracking claim.
func ntszRecipe() Recipe {
	return Recipe{
		ID:     "IE000OV4XWA3",
		Name:   "WisdomTree Eurozone Efficient Core: eurozone 90/60 replication",
		Method: "0.90×EZU (MSCI Eurozone, USD→EUR, extended EMU-EUR euro-area equity net TR ~1986) + 0.60×(EUNH.DE euro govt bond − EUR cash, extended EUROGOV-EUR ~1970 with ECB daily shape) + 0.10×EUR cash (EURCASH-EUR extended DECASH-EUR German money-market ~1960), 0.20%/yr fees; start now set by the equity leg (~1986)",
		Build:  ntszBuild,
		// The real fund is months old (2025-09): the overlap barely clears the
		// 60-point floor, so treat the validation as indicative only.
		ValidateAgainst: "IE000OV4XWA3",
		SpliceReal:      "IE000OV4XWA3",
	}
}

// ntszBuild assembles NTSZ from euro-native legs. The equity leg (EZU in EUR,
// extended by EMU-EUR) and the deep EUR cash leg (EURCASH-EUR extended by
// DECASH-EUR) are pre-built here and served to the frame under synthetic ids;
// the bond leg (EUNH.DE) reaches back through the standard extend() splice.
//
// The fee schedule is built on the REAL donors rather than on those synthetic
// ids, since it is EZU's and EUNH.DE's own inceptions (2000-07 and 2009-04)
// that end the fee-free euro-area reference series behind them. Both are
// wrapped funds, and the equity one is an expensive US-listed regional ETF:
// 0.90×0.50 + 0.60×0.07 = 0.492 %/yr against the fund's own 0.20, so from 2000
// the reconstruction owes nothing more and the whole 0.20 is charged only over
// the reference era before it.
func ntszBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	eq, err := ntszEquityEUR(f, from)
	if err != nil {
		return nil, err
	}
	cash, err := eurCashDeep(f, from)
	if err != nil {
		return nil, err
	}
	inj := injected{inner: extend(f), have: map[string]*marketdata.Series{
		"NTSZ-EQ":   eq,
		"NTSZ-CASH": cash,
	}}
	legs := []Leg{
		{ID: "NTSZ-EQ", Weight: 0.90},
		{ID: "EUNH.DE", Weight: 0.60, Excess: true},
		{ID: "NTSZ-CASH", Weight: 0.10},
	}
	fr, err := BuildFrame(inj, []string{"NTSZ-CASH", "NTSZ-EQ", "EUNH.DE"}, from)
	if err != nil {
		return nil, err
	}
	values, err := Composite(fr, legs, "NTSZ-CASH", 0)
	if err != nil {
		return nil, err
	}
	steps, err := feeGap(f, from, 0.0020, []pricedLeg{
		{ID: "EZU", Weight: 0.90, Load: ezuTER},
		{ID: "EUNH.DE", Weight: 0.60, Load: eunhTER},
	})
	if err != nil {
		return nil, err
	}
	name := "NTSZ (eurozone 90/60 replication)"
	return afterFeeSteps(name, SeriesFromFrame(name, fr, values), steps), nil
}

// ntszEquityEUR builds the eurozone equity leg: the real MSCI Eurozone ETF (EZU,
// USD) re-expressed in EUR at the EURUSD spot (like the unhedged DBMFE leg),
// then extended before EZU's history with the euro-area equity net-TR proxy
// (EMU-EUR, ~1986). EZU is US-listed, so the fetcher returns it in USD; the
// proxy is already EUR, hence the conversion happens before the splice.
func ntszEquityEUR(f Fetcher, from time.Time) (*marketdata.Series, error) {
	ezu, err := f.Fetch("EZU", from)
	if err != nil {
		return nil, err
	}
	if ezu == nil || len(ezu.Points) < 2 {
		return nil, fmt.Errorf("EZU: empty history")
	}
	dates := make([]time.Time, len(ezu.Points))
	usd := make([]float64, len(ezu.Points))
	for i, p := range ezu.Points {
		dates[i], usd[i] = p.Date, p.Close
	}
	eur, err := convertDaily("EZU (MSCI Eurozone) in EUR", extend(f), "EURUSD=X", from, dates, usd)
	if err != nil {
		return nil, err
	}
	if proxy, perr := f.Fetch("EMU-EUR", from); perr == nil && proxy != nil {
		marketdata.ExtendBack(eur, proxy)
	}
	return eur, nil
}

// eurCashDeep extends the daily EUR money-market series (eurCashDaily over
// EURCASH-EUR, 1994->) before the euro with the German 3-month money-market
// accrual (DECASH-EUR, ~1960), the natural pre-euro cash proxy. Germany was the
// anchor economy and the DM the reference currency, so its short rate stands in
// for euro cash before EURCASH-EUR begins.
func eurCashDeep(f Fetcher, from time.Time) (*marketdata.Series, error) {
	cash, err := eurCashDaily(f, from)
	if err != nil {
		return nil, err
	}
	if deep, derr := f.Fetch("DECASH-EUR", from); derr == nil && deep != nil {
		marketdata.ExtendBack(cash, deep)
	}
	return cash, nil
}

// injected serves pre-built component series by id and delegates every other
// fetch to the wrapped fetcher, so a composite can mix synthetic legs (built in
// code) with the standard extend()-spliced components.
type injected struct {
	inner Fetcher
	have  map[string]*marketdata.Series
}

func (j injected) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	if s, ok := j.have[id]; ok {
		return s, nil
	}
	return j.inner.Fetch(id, from)
}

// dbxgFee is DBXG's 0.15 %/yr TER, deducted from the pre-inception proxy (the
// real quotes grafted from 2007 already carry it).
const dbxgFee = 0.0015

// dbxgRecipe rebuilds the Xtrackers II Eurozone Government Bond 25+ UCITS ETF
// (LU0290357846, EUR, real from 2007) from the bundled long euro-area government
// bond total return (EUROGOV-LONG-EUR, ~1970, with the ECB daily 25-year shape
// EUROGOV-LONG-DAILY from 2004), less the fund's TER. That reference is a
// constant-maturity 24-year par bond, modified duration ~17, vol-matched to the
// iBoxx EUR Eurozone Sovereigns 25+ index DBXG tracks; it is a genuine long-bond
// reconstruction (yield path through TreasuryTR) rather than a levered shorter
// bond, so it neither overstates the return in a sustained bond bull nor needs a
// financing leg. Euro-native throughout, no FX leg; real DBXG is grafted from
// 2007. From 2004-09 both the level and the shape come from the real ECB
// 25-year curve point; only the deeper monthly tail synthesizes its long yield
// from the 10-year, the years no real long curve exists for; see
// cmd/gen-euro-refdata.
func dbxgRecipe() Recipe {
	return Recipe{
		ID:              "DBXG",
		Name:            "Xtrackers Eurozone Govt 25+: long euro-gov bond",
		Method:          "long euro-area government bond TR (EUROGOV-LONG-EUR, 24y constant maturity, modified duration ~17, ~1970; the real ECB 25y curve point sets level and shape from 2004 via EUROGOV-LONG-DAILY) less 0.15%/yr TER; euro-native, real DBXG grafted from 2007",
		Build:           dbxgBuild,
		ValidateAgainst: "DBXG",
		SpliceReal:      "DBXG",
	}
}

// dbxgBuild is the long euro-area government bond reference (euroGovLongDaily)
// net of the fund's TER: DBXG holds the 25+ bonds physically, so its return is
// the long-bond total return less fees.
func dbxgBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	gov, err := euroGovLongDaily(f, from)
	if err != nil {
		return nil, err
	}
	return afterFee(gov, dbxgFee), nil
}

// mthFee is MTH's 0.07 %/yr TER, deducted from the pre-inception proxy (the
// real quotes grafted from 2017 already carry it).
const mthFee = 0.0007

// mthRecipe rebuilds the Amundi Euro Government Bond 25+Y UCITS ETF Acc
// (MTH, LU1686832194, EUR, real from 2017) from the same long euro-area
// government bond total return as dbxgRecipe: the two funds track
// near-identical eurozone 25+ sovereign indexes, so they share one
// reconstruction and differ only by TER and inception date.
func mthRecipe() Recipe {
	return Recipe{
		ID:              "MTH",
		Name:            "Amundi Euro Govt Bond 25+Y: long euro-gov bond",
		Method:          "long euro-area government bond TR (EUROGOV-LONG-EUR, 24y constant maturity, modified duration ~17, ~1970; the real ECB 25y curve point sets level and shape from 2004 via EUROGOV-LONG-DAILY) less 0.07%/yr TER; euro-native, real MTH grafted from 2017",
		Build:           mthBuild,
		ValidateAgainst: "MTH",
		SpliceReal:      "MTH",
	}
}

// mthBuild is the long euro-gov reference net of MTH's TER, exactly like
// dbxgBuild with the cheaper fee.
func mthBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	gov, err := euroGovLongDaily(f, from)
	if err != nil {
		return nil, err
	}
	return afterFee(gov, mthFee), nil
}

// indepEuropeRecipe backcasts Independance AM Europe Small (LU1832174962,
// real from 2018) with its older sibling, Independance AM France Small & Mid
// (LU0131510165, same manager and process since the 1990s): on their
// 2018-2026 overlap the two funds' monthly returns correlate at ~0.95, which
// makes the sibling the best available proxy for the strategy's BEHAVIOR.
// The LEVEL is another matter: the graft carries the France fund's record,
// and France smalls both outperformed Europe before 2018 (the manager's
// golden era) and underperformed after, so read the backcast as regime
// shape, not as what a Europe fund would have earned. Both A(C) share
// classes carry the same 2.11%/yr of recurring charges inside the NAV (both
// KIDs of 17/02/2026), plus the same 10% performance fee: no fee adjustment.
func indepEuropeRecipe() Recipe {
	return Recipe{
		ID:              "LU1832174962",
		Name:            "Independance AM Europe Small: France sibling graft",
		Method:          "Independance AM France Small & Mid NAV (LU0131510165, same manager and process, monthly correlation ~0.95 on the 2018-2026 overlap; fees already in NAV, no adjustment) as the pre-inception proxy; real LU1832174962 grafted from 2018",
		Build:           indepEuropeBuild,
		ValidateAgainst: "LU1832174962",
		SpliceReal:      "LU1832174962",
	}
}

// indepEuropeBuild is the sibling fund's NAV, unchanged: same manager, same
// process, fees baked into both NAVs.
func indepEuropeBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	return f.Fetch("LU0131510165", from)
}

// euroGovLongDaily builds the long (25+ segment, 24-year par, duration ~17)
// euro-area government bond total return at daily granularity: the bundled
// monthly EUROGOV-LONG-EUR anchors (~1970) carrying the ECB daily 25-year shape
// (EUROGOV-LONG-DAILY, 2004->) where it overlaps, the long-maturity twin of the
// EUROGOV-EUR / EUROGOV-DAILY pair the NTSZ bond leg uses.
//
// The anchors govern the level, so what they carry after 2004 matters: they are
// themselves the month-ends of that same ECB curve (cmd/gen-euro-refdata splices
// them there), which is why the reconstruction reproduces the real long curve
// over those years instead of a 25-year yield synthesized from the 10-year one.
func euroGovLongDaily(f Fetcher, from time.Time) (*marketdata.Series, error) {
	anchors, err := f.Fetch("EUROGOV-LONG-EUR", from)
	if err != nil {
		return nil, err
	}
	if anchors == nil || len(anchors.Points) < 2 {
		return nil, fmt.Errorf("EUROGOV-LONG-EUR: empty history")
	}
	if shape, serr := f.Fetch("EUROGOV-LONG-DAILY", from); serr == nil && shape != nil {
		anchors = shapedSeries(anchors, shape)
	}
	return anchors, nil
}

// urthRecipe approximates the MSCI World as a fixed 60/40 US/developed-
// ex-US blend of Vanguard index funds.
func urthRecipe() Recipe {
	return Recipe{
		ID:     "URTH",
		Name:   "iShares MSCI World: MSCI World total return (1969 with -refdata)",
		Method: "real MSCI World net TR (MSCIWORLD-USD refdata, monthly 1969→) with the daily shape of the MSCI World price index (^990100-USD-STRD, 1972→), less 0.24%/yr TER; without the refdata falls back to 0.60×VFINX+0.40×VTMGX (1999)",
		Build: msciWorld(0.0024, composite("URTH (MSCI World replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.40},
		}, "", 0.0024)),
		ValidateAgainst: "URTH",
	}
}

// iefRecipe and tltRecipe extend the treasury ETFs (2002) with their
// long-running Vanguard equivalents (VFITX 1991→, VUSTX 1986→).
//
// VFITX is a 5-10 year fund and IEF holds the 7-10 segment, so the donor is
// short of the target's duration and the file used to say so loudly: 4.81 %
// annualized volatility against the fund's 6.80 over their whole 2002-2026
// overlap, a fifth of the rate sensitivity the sleeve is bought for simply
// missing. It is geared by intTreasuryGearing, over collateral earning the
// bill rate as zrozRecipe does, since IEF owns its bonds outright and the cash
// term is only the arithmetic residue of writing g × bond as
// cash + g × (bond − cash).
//
// One caveat travels with the deep tail: before VFITX's own 1991 start the leg
// is the 5-year constant-maturity reconstruction (TREASURY-INT-USD), shorter
// still than VFITX, so 1.41 leaves the 1953-1991 segment under-durated rather
// than over. The error keeps its sign, which is the sign worth keeping.
func iefRecipe() Recipe {
	return Recipe{
		ID:     "IEF",
		Name:   "iShares 7-10Y Treasury: VFITX intermediate Treasury",
		Method: "cash + 1.41×(VFITX − cash) (Vanguard Intermediate-Term Treasury, 1991→, geared to the 7-10 segment's duration), real IEF grafted from 2002",
		Build: composite("IEF (intermediate Treasury)", []Leg{
			{ID: "VFITX", Weight: intTreasuryGearing, Excess: true},
			{ID: "^IRX", Weight: 1},
		}, "^IRX", 0),
		ValidateAgainst: "IEF",
		SpliceReal:      "IEF",
	}
}

// intTreasuryGearing and shortTreasuryGearing put the two Vanguard Treasury
// donors on the segment their target ETF actually holds. Both are the donor's
// own volatility ratio against the fund, measured on the whole 2002-2026
// overlap (6030 common days) and nothing else:
//
//	IEF on VFITX   6.80 % against 4.81 %   ratio 1.412, beta 1.333, corr 0.945
//	SHY on VFISX   1.52 % against 2.03 %   ratio 0.751, beta 0.609, corr 0.812
//
// A volatility ratio is the right instrument here because these are two
// Treasury funds on the same curve: what separates them is duration, and for
// the same yield factor volatility is proportional to it. The same measurement
// reproduces the long-bond gearing this file already carries from a duration
// table, which is the check that the two instruments agree: TLT on VUSTX gives
// 1.131 over the same window against longTreasuryGearing's 17/15 = 1.133.
//
// The TE-minimizing gearing is lower in both cases (1.35 for IEF, and the beta
// is lower still at 1.33), and it is deliberately not used: with correlation
// under one, minimizing tracking error under-risks the sleeve by that factor,
// and understating rate sensitivity is the expensive error for a line held as
// a deflation hedge.
//
// What each gearing bought, on the fund's own overlap (correlation is
// invariant under a scalar, so it is the beta, the tracking error and the
// volatility that move):
//
//	IEF  beta 1.33 -> 0.95, TE 2.75 -> 2.26 %/yr, vol 4.81 -> 6.79 % against 6.80, CAGR gap -0.29 -> +0.28 pt
//	SHY  beta 0.61 -> 0.81, TE 1.19 -> 0.93 %/yr, vol 2.03 -> 1.52 % against 1.52, CAGR gap +0.19 -> +0.08 pt
const (
	intTreasuryGearing   = 1.41
	shortTreasuryGearing = 0.75
)

// shyRecipe extends the 1-3 year Treasury ETF with Vanguard's short-Treasury
// fund, geared DOWN: VFISX runs longer than the 1-3 segment (2.03 % annualized
// volatility against SHY's 1.52 over their 2002-2026 overlap), so the donor is
// held at shortTreasuryGearing over collateral, the mirror image of what IEF
// needs. See that constant for the measurements and for why the ratio rather
// than the beta sets it.
func shyRecipe() Recipe {
	return Recipe{
		ID:     "SHY",
		Name:   "iShares 1-3 Year Treasury Bond ETF",
		Method: "cash + 0.75×(VFISX − cash) (Vanguard Short-Term Treasury, 1991→, geared down to the 1-3 segment's duration), real SHY grafted from 2002",
		Build: composite("SHY (short Treasury)", []Leg{
			{ID: "VFISX", Weight: shortTreasuryGearing, Excess: true},
			{ID: "^IRX", Weight: 1},
		}, "^IRX", 0),
		ValidateAgainst: "SHY",
		SpliceReal:      "SHY",
	}
}

func tltRecipe() Recipe {
	return Recipe{
		ID:              "TLT",
		Name:            "iShares 20+Y Treasury: VUSTX long Treasury",
		Method:          "VUSTX (Vanguard Long-Term Treasury, 1986→), real TLT grafted from 2002",
		Build:           composite("TLT (long Treasury)", []Leg{{ID: "VUSTX", Weight: 1}}, "", 0),
		ValidateAgainst: "TLT",
		SpliceReal:      "TLT",
	}
}

// longTreasuryGearing is the weight the VUSTX donor carries in the recipes that
// target the 20+ Treasury segment (dtlaRecipe, dtleRecipe). VUSTX is a 10-25
// year fund, effective duration ~15, against a ~17 target, so the donor is
// geared to the duration ratio: the point of these lines is their rate
// sensitivity, and a 12% shortfall in it would quietly understate the deflation
// hedge they are bought for. The reported beta cannot arbitrate it, being a
// real-on-sim slope depressed by the US-close donor against a European listing.
// The older idtlRecipe holds the same bonds ungeared and is left alone rather
// than silently moved, so the distributing twin runs ~12% short on duration.
//
// The duration table is corroborated by the only artefact-free measurement of
// the same ratio, the US-listed twin TLT against VUSTX (both struck at the US
// close), whose whole 2002-2026 overlap reads 1.131. That ratio is not
// constant, though, because duration itself shrinks as yields rise: it reads
// 1.243 over 2002-2010, 1.110 over 2011-2018 and 1.06 over 2018-2026. 17/15 is
// therefore the whole-window value, and it lands inside the era the constant is
// actually used in, since the real quotes cover everything from 2018.
//
// Measuring the ratio on DTLA's OWN quotes gives 1.02, and that reading is
// twice unusable. It is the live window, where the gearing is never applied;
// and it is depressed by the listing, DTLA reading 0.963 of TLT's monthly
// volatility and the distributing twin IDTL 0.970, the same deficit for the
// same reason (a 16:30 London close against a 16:00 New York one), while DTLA
// and IDTL agree with each other at 0.993. Undone, the two corrections put
// DTLA's own window back on the 1.06 its twin measures there. Gearing at 1.06
// was tested end to end anyway and the level refuses it: the live-window CAGR
// gap goes from +0.03 to +0.36 pt.
const longTreasuryGearing = 17.0 / 15.0

// dtleRecipe backcasts the iShares $ Treasury Bond 20+yr UCITS ETF EUR Hedged
// (IE00BD8PGZ49, DTLE, real from 2017-09) as US long Treasuries hedged to EUR.
// It applies the standard FX-hedge identity, exactly as tip1eRecipe does for
// hedged TIPS: a hedged foreign return equals the local return financed at the
// foreign (USD) cash rate plus the domestic (EUR) cash rate earned on the
// hedged capital, so the EUR investor collects the EUR-minus-USD carry (deeply
// negative in 2018-2024, which is most of the real overlap) on top of the bond
// return. The local leg is VUSTX, the same donor the unhedged USD sibling
// (idtlRecipe) uses, itself extended to 1962 by the TREASURY-LONG-USD 20-year
// CMT reconstruction and its daily shape, geared by dtleDuration to reach the
// 20+ segment's duration. Currency-wise the output is EUR-native by
// construction (no FX path in it), which is the point of the hedged class: the
// US curve without the dollar.
//
// It lives under its OWN identifier, DTLETR, rather than extending DTLE, and
// nothing is grafted on top: DTLE is a DISTRIBUTING share class, so its only
// public series (FT NAV) is a PRICE return that omits the coupons it pays out,
// which for a long-Treasury fund is most of the return. Extending DTLE would
// have spliced that price series over the recent decade and silently corrupted
// it, and every SIM consumer would have inherited the mix. The overlap check
// against the real fund is kept for the shape, and its CAGR gap is the finding
// rather than an error: the reconstruction beats the quotes by the coupons the
// price series drops, ~1.7%/yr over 2017-2026, which is what a EUR-hedged US
// long-bond sleeve distributes once the hedge has paid away the dollar-euro
// rate gap. Taking the local leg from the accumulating twin rather than from a
// US mutual fund also removed the timing mismatch that used to cost most of
// the daily agreement: 0.91 daily and 0.93 weekly against 0.70 and 0.89
// before, beta 0.95 against 0.79, tracking error 6.5%/yr against 12.2%.
//
// One caveat on the reference itself: the FT NAV prints -13.1% on 2020-03-12
// where the accumulating twin, holding the same bonds, moves -5.6%. Nothing
// is grafted from DTLE, so the print only costs a little of the measured
// agreement.
//
// Read it as the total-return view of the EUR-hedged US 20+ segment, which is
// what a backtest needs; the tradable line stays DTLE, and a French investor
// should weigh its annual coupon taxation against an accumulating alternative
// before holding it in a taxable account.
func dtleRecipe() Recipe {
	return Recipe{
		ID:   "DTLETR",
		Name: "US long Treasuries hedged to EUR (total return, the DTLE segment)",
		Method: "the accumulating USD twin's own total return (real DTLA from 2018, 1.13×VUSTX geared to the 20+ duration before, extended TREASURY-LONG daily from 1962) " +
			"financed at USD cash ^IRX and re-earning EUR cash (EURCASH-EUR) = EUR-hedged US long Treasuries, less what the class's 0.10%/yr TER exceeds its donor's own charge (0.03%/yr over the twin, nothing over the dearer geared mutual fund); total return, nothing grafted (the real DTLE series is a distributing NAV, price-only)",
		Build:           dtleBuild,
		ValidateAgainst: "DTLE",
	}
}

// dtleBuild hedges the USD 20+ Treasury segment into EUR: the local leg earns
// its excess over USD cash, the hedge gives back EUR cash, and what the hedged
// class costs OVER ITS DONOR comes off. From 2018 the local leg is the REAL
// accumulating USD twin (DTLA), which holds the same bonds and prices on the
// same London close as DTLE itself, so the timing mismatch that a US mutual-fund
// donor carries disappears over the whole window the comparison is made on;
// before it, the geared VUSTX reconstruction stands, as it does inside DTLA's
// own history.
//
// Both donors are wrapped funds, so neither owes the hedged class's whole
// 0.10 %/yr, which the build deducted on top of them until 2026-08. The twin
// charges 0.07 % (iShares fund page, 2026), leaving 0.03 to deduct over its
// era; the geared mutual fund charges 1.13×0.20 = 0.227 %, more than the target
// itself, so the VUSTX era owes nothing and is floored at zero. The charge
// therefore follows the donor date by date, in the same branch that picks the
// leg, rather than averaging two eras into one wrong constant.
func dtleBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	fr, err := BuildFrame(extend(f), []string{"^IRX", "VUSTX", "EURCASH-EUR"}, from)
	if err != nil {
		return nil, err
	}
	irx, vustx, eur := fr.Returns["^IRX"], fr.Returns["VUSTX"], fr.Returns["EURCASH-EUR"]

	twin := map[time.Time]float64{}
	if real, err := f.Fetch("DTLA", from); err == nil && real != nil {
		for i := 1; i < len(real.Points); i++ {
			if real.Points[i-1].Close > 0 {
				twin[real.Points[i].Date] = real.Points[i].Close/real.Points[i-1].Close - 1
			}
		}
	}

	const (
		dtleTER  = 0.0010
		twinFee  = max(0, dtleTER-dtlaTER) / 252 // the real accumulating twin's era
		vustxFee = max(0, dtleTER-longTreasuryGearing*vustxTER) / 252
	)
	s := &marketdata.Series{
		Name: "DTLE (EUR-hedged US long Treasury, total return)", Source: "simdata", Currency: "EUR",
	}
	v := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[0], Close: v})
	for k := 1; k < len(fr.Dates); k++ {
		excess, fee := longTreasuryGearing*(vustx[k]-irx[k]), vustxFee
		if r, ok := twin[fr.Dates[k]]; ok {
			excess, fee = r-irx[k], twinFee
		}
		v *= 1 + excess + eur[k] - fee
		s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[k], Close: v})
	}
	return s, nil
}

// zrozRecipe approximates 25+ year zero-coupon STRIPS by leveraging the long
// Treasury fund VUSTX to 1.65× over cash (its ~25-year duration matches
// ZROZ's) ON TOP of the fully invested collateral earning cash: a STRIPS
// fund owns its bonds outright, so the backcast must credit the cash rate
// the excess formulation strips out. Without the collateral leg the sim
// lagged the real fund by the T-bill average (~1.2%/yr over 2009-2026) and
// collapsed in the high-rate 1960s-1980s (-6%/yr in the 60s, +1.9%/yr
// full-period vs ~+6% for long Treasuries themselves). Real ZROZ quotes are
// grafted on top.
func zrozRecipe() Recipe {
	return Recipe{
		ID:     "ZROZ",
		Name:   "PIMCO 25+Y zero-coupon: 1.65× long Treasury",
		Method: "cash + 1.65×(VUSTX − cash) (leveraged long Treasury ≈ 25+ STRIPS duration, 1986→), real ZROZ grafted from 2009",
		Build: composite("ZROZ (cash + 1.65x long Treasury excess)", []Leg{
			{ID: "VUSTX", Weight: 1.65, Excess: true},
			{ID: "^IRX", Weight: 1},
		}, "^IRX", 0),
		ValidateAgainst: "ZROZ",
		SpliceReal:      "ZROZ",
	}
}

// dbmf/kmlm/cta reconstruct managed-futures trend out of REAL records of the
// same trade, spliced behind each fund nearest first and each fee-aligned to
// it: the published composite the fund replicates where there is one, the
// closest single funds otherwise, and Man AHL Diversified at the bottom, where
// every file starts (1996-03). The 12-month TSMOM engine on a cross-asset
// futures basket is still built and still matters, but only as the daily
// texture the weekly-dealing deepest donor is projected onto; it is no longer
// shipped in front of the chain. See docs/trend-reconstruction-design.md for
// what each layer is worth, and DonorChain for how they are joined.

// xauusdRecipe snapshots gold: XAU/USD spot has decades of real history (~1968),
// so the "reconstruction" is simply the real spot series, embedded so the long
// history is available offline and as the gold proxy for other builds.
func xauusdRecipe() Recipe {
	return Recipe{
		ID:              "XAUUSD",
		Name:            "Gold (XAU/USD spot)",
		Method:          "real gold spot (XAU/USD daily, ~2000→) extended back with the daily London/LBMA PM gold fix (bundled refdata XAUUSD-LBMA, 1968→)",
		Build:           xauusdBuild,
		ValidateAgainst: "XAUUSD",
	}
}

// xauusdBuild returns the gold spot series: the fetchable daily XAU/USD quote
// (~2000→) with the bundled daily London/LBMA PM gold fix (XAUUSD-LBMA, 1968→)
// spliced behind it, so a gold sleeve covers the whole post-Bretton-Woods
// floating era. If the daily quote is unavailable the monthly fix stands alone.
func xauusdBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	long, _ := f.Fetch("XAUUSD-LBMA", from)
	s, err := f.Fetch("XAUUSD", from)
	if err != nil || s == nil || len(s.Points) == 0 {
		if long != nil && len(long.Points) > 0 {
			return long, nil
		}
		return s, err
	}
	if long != nil {
		marketdata.ExtendBack(s, long)
	}
	return s, nil
}

// iglnRecipe backcasts the iShares Physical Gold ETC (IE00B4ND3602, USD, real
// from 2011) as spot gold minus its 0.12%/yr storage fee. The single leg is
// GC=F (the XAU/USD gold spot the XAUUSD sleeve and the managed-futures recipes
// already trade, real ~2000), extended back to 1968 by the bundled daily
// London/LBMA PM gold fix (refdata XAUUSD-LBMA). The fund is USD-quoted, like
// gold spot, so no FX leg is needed; the real IGLN quotes are grafted from
// inception. A physically-backed ETC tracks the LBMA price minus storage, which
// the fee term reproduces.
func iglnRecipe() Recipe {
	return Recipe{
		ID:              "IE00B4ND3602",
		Name:            "iShares Physical Gold: spot gold minus storage",
		Method:          "GC=F (XAU/USD gold spot, extended to the daily London/LBMA PM fix XAUUSD-LBMA ~1968) minus 0.12%/yr storage, real IGLN grafted from 2011",
		Build:           composite("IGLN (physical gold)", []Leg{{ID: "GC=F", Weight: 1}}, "", 0.0012),
		ValidateAgainst: "IE00B4ND3602",
		SpliceReal:      "IE00B4ND3602",
	}
}

// The two bundled daily net composites, offered to a chain as donors rather
// than as anchors. Both are published records of what a book of REAL programmes
// paid its investors, each constituent net of its own manager's fees, daily
// from 2000-01-03 (cmd/gen-sgtrend-refdata holds their provenance).
//
// A composite is a strange thing to splice behind a single fund, and it earns
// the slot only where it beats every single fund available: it is diversified
// where the target is one manager, so its months are smoother, and it must be
// levered to the target's volatility, which levers its drawdowns with it. Two
// funds of this family nonetheless track their own index better than they track
// any other manager's NAV, and for the same reason in both cases: they are
// built to reproduce it. See docs/trend-reconstruction-design.md for the
// arbitration.
const (
	allStylesIndex = "TREND-ALLSTYLES-NET-USD" // the index the DBi family replicates
	pureTrendIndex = "TREND-PURE-NET-USD"      // the published index closest to Simplify CTA's own risk
)

// dbiDonors are the records spliced behind the DBi family, nearest trade first.
// Measured against DBMF over its own live window (monthly correlation, the
// honest yardstick for a sleeve held for years): the composite half-projected
// on the fund's own ten futures 0.89 from 2000-03 (DBiDonorID, see
// dbireplica.go), the raw all-styles composite 0.85 for the quarter before
// that, Man AHL Diversified 0.77 from 1996-03. The single funds that held the
// 2000-2019 slot until 2026-08 (the Virtus AlphaSimplex fund 0.81 from 2010-08
// and the Guggenheim fund 0.72 from 2007-02) are gone from this chain: the
// index covers their whole era and tracks the fund better than either, which is
// unsurprising given that reproducing that index is what DBi's programme does.
//
// The raw composite is kept behind the projected one rather than dropped: the
// projection costs sixty trading days of warm-up, and the index itself is a
// better donor for that quarter than a weekly-dealing fund of another manager.
//
// ahlDiversified is the deepest donor of the family and the one that made the
// sparse-cadence handling necessary: the fund dealt WEEKLY until 2016, so the
// whole segment it contributes (1996-03 to 2000-01) arrives one NAV a week and
// is projected onto the reconstruction's daily calendar (see DonorChain). Its
// record reproduces the manager's published one (+1649.5 % from 1996-03-26 to
// 2026-02-27 against a published +1649.47 %) and its crisis years are the real
// thing: 1998 +41 %, 2008 +33 %, 2022 +12 %.
const ahlDiversified = "IE0000360275"

var dbiDonors = []string{DBiDonorID, allStylesIndex, ahlDiversified}

// The other chains' donors, nearest trade first. Each recipe declares them
// once and passes the same slice to feeAligned and to Recipe.Donors, so the
// audit report (see audit.go) grades the chain the file is actually built on.
var (
	kmlmDonors  = []string{"ASFYX", "RYMFX", ahlDiversified}
	aqrmfDonors = []string{"AQMIX", "RYMFX", ahlDiversified}
	ctaDonors   = []string{pureTrendIndex, ahlDiversified}
)

// trendFeeLoad is the documented MANAGEMENT-AND-EXPENSE load of every vehicle
// in the managed-futures family, donors and targets alike, as a fraction per
// year. Performance fees are NOT in it (trendPerfFee holds the one that is
// documented), and neither is anything a fund pays out of the trade itself:
// this is the price list, not the record.
//
// The donors of this family are old-school 1940-Act funds and an offshore
// vehicle charging 1.3 to 2.7 %/yr; the funds they stand in for are modern
// ETFs and UCITS classes at 0.75 to 0.90 %. A donor segment therefore runs
// roughly a point a year colder than the fund it covers for, and that cold
// belongs to the wrapper, not to the trade.
//
// Every figure is a net expense ratio (after contractual waivers) from the
// vehicle's own current fee table unless said otherwise:
//
//	ASFYX  1.45 %  Virtus AlphaSimplex Managed Futures Strategy, class I:
//	               1.59 % total, 1.45 % after the contractual cap running to
//	               2027 (summary prospectus, 2026-04-28)
//	RYMFX  1.99 %  Guggenheim Managed Futures Strategy, class P: 2.18 %
//	               total, 1.99 % after waiver (summary prospectus, 2026);
//	               its swap counterparties' own management and performance
//	               fees are embedded in the swap returns and excluded from
//	               that table, so 1.99 % understates what the shareholder pays
//	AHLPX  1.91 %  American Beacon AHL Managed Futures Strategy, Investor
//	               class: 1.91 % total, no waiver in force (prospectus
//	               supplement effective 2025-08-25). It is priced here and
//	               spliced nowhere: it held the Simplify CTA chain's 2014-2022
//	               slot until the index donor took the whole era, and stays in
//	               the table so restoring it needs no fresh research
//	AQMIX  1.29 %  AQR Managed Futures Strategy, class I: 1.29 % total and
//	               after reimbursement (summary prospectus dated 2024-05-01)
//	AHL    2.74 %  Man AHL Diversified plc, the USD accumulating class: the
//	               ongoing charge a fund database reports for this ISIN
//	               (2026). The company's own audited accounts are harsher
//	               still: a 3.00 %/yr investment management fee on class A, a
//	               1.00 %/yr introducing broker fee and a 20 % performance fee
//	               on net new profits (semi-annual financial statements to
//	               2016-12-26, note 8), so 2.74 % is the conservative reading
//	DBMF   0.85 %  iMGP DBi Managed Futures Strategy ETF (fund page, 2026);
//	               it is a donor to the UCITS classes and a target itself
//	UCITS  0.75 %  the DBi UCITS share classes LU2951555585 and DBMFE
//	KMLM   0.90 %  KraneShares Mount Lucas Managed Futures Index Strategy ETF
//	CTA    0.75 %  Simplify Managed Futures Strategy ETF
//	AQR A  0.79 %  AQR Managed Futures UCITS class A base cost: 0.60 %
//	               management + 0.18 % expense cap + 0.01 % subscription tax
//	               (prospectus supplement); see trendPerfFee for the rest
//
// The DBi family's nearest donor (DBiDonorID) is an average of two things, so
// its load is the average of theirs: half the composite's estimated 2 %, half a
// futures portfolio that pays no manager at all, hence 1.00 %.
//
// One entry is ESTIMATED rather than read off a price list, and it is the only
// one: the two INDEX donors. An index of funds levies nothing itself, but every
// return in it arrives net of a constituent manager's own fees, so what it
// carries as a load is its constituents'. Those are private programmes and
// publish no schedule; 2 % management is the industry standard for the era and
// the vehicles they run, and it is what the donors of this very file charge
// (1.3 to 2.7 %, mean 1.85 %). The house rule keeps it conservative in the same
// direction as everywhere else: the constituents' PERFORMANCE fees, worth
// another 1 to 3 points a year in a good decade, are ignored, so the uplift
// gives a fund back its wrapper's difference and never the manager's cut. The
// residual is deliberately left in the reconstruction, where it reads as the
// replicator's edge over the index it replicates, which is what it is.
var trendFeeLoad = map[string]float64{
	"ASFYX":        0.0145,
	"RYMFX":        0.0199,
	"AHLPX":        0.0191, // priced but no longer spliced, see above
	"AQMIX":        0.0129,
	ahlDiversified: 0.0274,
	allStylesIndex: 0.0200, // estimated, see above
	pureTrendIndex: 0.0200, // estimated, see above
	DBiDonorID:     0.0100, // half the composite's estimated load, half a fee-free futures book
	"DBMF":         0.0085,
	"LU2951555585": 0.0075,
	"DBMFE":        0.0075,
	"MFEH":         0.0075,
	"KMLM":         0.0090,
	"CTA":          0.0075,
	"LU1103257975": 0.0079,
}

// trendPerfFee is the documented performance fee a TARGET pays and its donors
// do not, as a fraction per year. Only one fund of the family has one that is
// both charged and measured: AQR Managed Futures UCITS class A pays 10 % of
// the excess over the ML 3-Month T-Bill hurdle, which the audited accounts put
// at 1.58 points of average class NAV in the year to 31 March 2026. The
// flat-fee classes and every US ETF here pay none.
var trendPerfFee = map[string]float64{"LU1103257975": 0.0158}

// feeAligned pairs each donor with the uplift that makes it carry the target's
// fee load instead of its own: the donor's management-and-expense load minus
// the target's, floored at zero.
//
// The uplift is read off published fee schedules and NEVER off an observed
// return gap. A gap between two managers contains their skill, and this family
// has a lot of it: DBi beats every peer and index over its own live window by
// about six points a year. Closing that with a "fee" constant would grant the
// manager's alpha to the backcast, which is curve fitting. Only the price list
// is transferable.
//
// Performance fees enter asymmetrically, and both directions are deliberately
// conservative, meaning both can only make the uplift too small:
//
//   - the DONOR's performance fee is ignored. Man AHL's 20 % of net new
//     profits and Guggenheim's embedded swap fees are real money the donor's
//     shareholders never saw, and leaving them out keeps the corrected segment
//     colder than the truth.
//   - the TARGET's performance fee is subtracted, because the target's own
//     record is already net of it and the donor owes only the DIFFERENCE. This
//     is what keeps the AQR chain honest: class A's base list is 0.50 points
//     under AQMIX's, but class A also pays a performance fee AQMIX does not,
//     so the aligned uplift is zero rather than half a point. Measurement
//     agrees, the AQMIX chain already tracking class A's live CAGR to within
//     0.1 point.
//
// A vehicle with no entry in trendFeeLoad is a programming error, not a free
// donor: the whole point is that no segment is spliced without a documented
// price.
func feeAligned(target string, donors []string) []Donor {
	out := make([]Donor, len(donors))
	for i, id := range donors {
		uplift := feeLoad(id) - feeLoad(target) - trendPerfFee[target]
		if uplift < 0 {
			uplift = 0
		}
		out[i] = Donor{ID: id, Uplift: uplift}
	}
	return out
}

// feeLoad is trendFeeLoad's lookup, and it panics on a miss: a donor or target
// whose price list nobody looked up would otherwise be silently aligned
// against zero.
func feeLoad(id string) float64 {
	fee, ok := trendFeeLoad[id]
	if !ok {
		panic("simgen: no documented fee load for " + id)
	}
	return fee
}

// chainedTrend builds a managed-futures history out of real NAVs of the
// closest programmes, for as far back as they go and NO FURTHER. calibrate
// names the fund the donors are volatility-matched to (the fund itself, or the
// twin it is a share class of), and donors is the chain nearest first, each
// carrying the fee uplift that puts it on the target's price list (feeAligned).
// The file therefore starts at the deepest donor's own first NAV, 1996-03 for
// this family.
//
// The reconstruction is still built, and it still matters: it is the daily
// texture a sparsely-dealing donor is projected onto (see DonorChain). The
// oldest donors are weekly-dealing funds, and only the engine can say what a
// trend book did between two of their NAVs. What it no longer does is stand in
// front of the chain as shipped history. A reconstruction anchored on a monthly
// composite is a decent account of a decade in aggregate and a poor one of any
// month a reader might look up, and a history that is worth reading throughout
// beats a longer one that is not (see docs/trend-reconstruction-design.md).
func chainedTrend(name, calibrate string, donors []Donor, cfg TSMOMConfig) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		texture, err := tsmom(name+" (daily texture)", cfg)(f, from)
		if err != nil {
			return nil, err
		}
		chain, err := DonorChain(f, cfg.CashID, calibrate, donors, from, texture)
		if err != nil {
			return nil, err
		}
		chain.Name = name
		chain.Source = "simdata"
		chain.SimulatedBefore = time.Time{}
		return chain, nil
	}
}

// dbiChain is chainedTrend for the DBi family, all volatility-matched to the
// US-listed ETF.
func dbiChain(name string, donors []Donor, cfg TSMOMConfig) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return chainedTrend(name, "DBMF", donors, cfg)
}

// dbmfRecipe reconstructs the iMGP DBi Managed Futures ETF behind the very
// index its programme sets out to replicate, and behind that index READ THE WAY
// THE FUND READS IT: the published net ALL-STYLES composite, half of it
// projected onto the ten futures contracts the fund actually holds
// (DBiDonorID, see dbireplica.go), daily from 2000-03.
//
// Each step of that was measured on the fund's own live window (2019-05 to
// 2026-07). Swapping the two single funds of other managers for the index they
// replicate took the monthly correlation from 0.81 to 0.85 in 2026-08;
// projecting half the index onto the fund's own instrument set takes it to
// 0.89, cuts the tracking error from 10.0 % to 8.2 % of a 12.4 % volatility,
// and, the criterion that actually decides a donor, cuts the swing of the CAGR
// gap between two disjoint halves of the window from 5.2 points to 1.2.
//
// The residual level gap has a name, and it survives: DBi copies the index
// constituents' positions at 0.85 % flat where the index arrives net of those
// constituents' 2-and-20, and the uplift only claims back the 2, so the
// reconstruction is expected to sit under the fund by roughly what a
// performance fee costs. It does, by about 1.3 points a year. Closing the rest
// would mean granting DBi's own edge to the backcast.
func dbmfRecipe() Recipe {
	return Recipe{
		ID:   "DBMF",
		Name: "iMGP DBi Managed Futures: real managed-futures NAVs, then a TSMOM reconstruction",
		Method: "the published net all-styles managed-futures composite the fund replicates, half of it projected onto the ten futures contracts the fund holds (the fund's own published process, run on the index), spliced behind it from 2000-03, volatility-matched to it and lifted +0.15%/yr for the fee load it carries; " +
			"the raw composite for the quarter before that (+1.15%/yr), " +
			"then real NAVs of Man AHL Diversified (1996-03, +1.89%/yr, its weekly NAVs projected onto the reconstruction's daily calendar), the file starting at that deepest donor's own first NAV, real DBMF grafted from 2019",
		Donors:          dbiDonors,
		Build:           dbiChain("DBMF (donor chain)", feeAligned("DBMF", dbiDonors), mfConfig(0.115, 0.0085)),
		ValidateAgainst: "DBMF",
		SpliceReal:      "DBMF",
	}
}

// dbmfpaRecipe reconstructs the UCITS USD share class (DBMF.PA,
// LU2951555585, Paris-listed, launched 2025-04-22) of the iMGP DBi
// managed-futures fund: the same USD TSMOM replication as the US-listed DBMF,
// with the real UCITS quotes grafted from inception. Same strategy and
// currency (USD, unhedged) as DBMF, only a different (UCITS) wrapper, at the
// UCITS 0.75% TER.
func dbmfpaRecipe() Recipe {
	return Recipe{
		ID:   "LU2951555585",
		Name: "iMGP DBi Managed Futures UCITS USD: the US ETF, then the donor chain",
		Method: "the US-listed DBMF itself (same manager, same strategy, same currency: monthly correlation 0.97 on their overlap) from 2019, lifted 0.10%/yr for the cheaper UCITS fee load, " +
			"then the net all-styles composite half-projected onto the fund's own ten futures (2000-03, +0.25%/yr), the raw composite for the quarter before it (+1.25%/yr) and Man AHL Diversified (1996-03, +1.99%/yr), the file starting at that deepest donor's own first NAV, real DBMF.PA grafted from 2025",
		Donors:          append([]string{"DBMF"}, dbiDonors...),
		Build:           dbiChain("DBMF.PA (donor chain)", feeAligned("LU2951555585", append([]string{"DBMF"}, dbiDonors...)), mfConfig(0.115, 0.0075)),
		ValidateAgainst: "LU2951555585",
		SpliceReal:      "LU2951555585",
	}
}

// dbmfeRecipe reconstructs the *unhedged* EUR share class (DBMFE,
// LU2951555403, Paris-listed, launched 2025-03-24) of the iMGP DBi
// managed-futures fund. It runs the same USD TSMOM replication as DBMF, then
// re-expresses it in EUR at the EUR/USD spot rate (unhedged), so the EUR
// investor also carries the USD/EUR currency move on top of the strategy. The
// real DBMFE quotes are grafted from inception. EURUSD=X (Yahoo, ~2003→) is
// extended back to 1971 by the bundled ECU/DM/EUR proxy, so the start date is now
// set by the strategy's own youngest leg, not the FX cross.
func dbmfeRecipe() Recipe {
	return Recipe{
		ID:   "DBMFE",
		Name: "iMGP DBi Managed Futures EUR unhedged: the US ETF and its donor chain, in EUR",
		Method: "the US-listed DBMF itself from 2019, then the net all-styles composite half-projected onto the fund's own ten futures (2000-03), the raw composite for the quarter before it and Man AHL Diversified (1996-03), each lifted to the UCITS class's 0.75%/yr fee load, the file starting at that deepest donor's own first NAV, " +
			"the whole converted USD→EUR at EURUSD spot (bundled ECU/DM/EUR proxy back to 1971), real DBMFE grafted from 2025",
		Donors:          append([]string{"DBMF"}, dbiDonors...),
		Build:           dbmfeBuild,
		ValidateAgainst: "DBMFE",
		SpliceReal:      "DBMFE",
	}
}

// dbmfeBuild runs the USD DBMF strategy and converts each daily return into an
// unhedged EUR return via the EUR/USD spot rate: a EUR-denominated NAV equals
// the USD NAV divided by the USD-per-EUR rate, so r_eur = (1+r_usd)/(1+r_fx)−1
// where r_fx is the EURUSD (USD per EUR) daily change. The cross is
// forward-filled onto the strategy's own trading calendar (see fxOnDates)
// rather than joined into the frame, which would pollute the calendar with
// the FX feed's weekend prints.
func dbmfeBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	cfg := mfConfig(0.115, 0.0085) // identical USD strategy to dbmfRecipe
	usd, err := dbiChain("DBMFE (USD donor chain)", feeAligned("DBMFE", append([]string{"DBMF"}, dbiDonors...)), cfg)(f, from)
	if err != nil {
		return nil, err
	}
	dates := make([]time.Time, len(usd.Points))
	values := make([]float64, len(usd.Points))
	for i, p := range usd.Points {
		dates[i], values[i] = p.Date, p.Close
	}
	return convertDaily("DBMFE (USD donor chain converted to unhedged EUR)",
		extend(f), "EURUSD=X", from, dates, values)
}

// mfehRecipe reconstructs the EUR-HEDGED share class (MFEH, LU3359622902,
// "R EUR HP", Xetra-listed, launched 2026-05-13) of the same iMGP DBi
// managed-futures sub-fund as DBMFE and LU2951555585: one portfolio, three
// share classes. Where DBMFE re-expresses the USD strategy at EURUSD spot
// (unhedged, so the holder carries the dollar on the whole position), this
// class hedges the NAV back to euro, which the standard identity prices
// exactly: the USD return, financed at USD cash, re-earning EUR cash. The TER
// is the same 0.75%/yr as the two sister classes; the hedge's only cost is
// the carry gap between the two money markets (plus forward roll frictions
// the identity ignores, small for EURUSD).
//
// Measured against the unhedged sister on the shared backcast (monthly,
// 1996-03 to 2026-01): CAGR 8.96% vs 9.52%, volatility 12.7% vs 15.6%, worst
// drawdown -21.6% vs -25.2%. The hedge removes about three points of
// volatility (all of it EURUSD noise, uncorrelated with the strategy) and
// the drawdowns the dollar added; the 0.9%/yr of return it gave up over that
// window is dollar drift plus carry, path-dependent, not a fee. For the
// crisis-alpha role the difference is the point: the unhedged class is a
// bundled bet (trend plus long dollar) that pays double when the dollar
// rallies into the crisis (2022: +29.3% vs +19.9% hedged) and gives most of
// the alpha back when the crisis comes with a falling dollar (a 1990); the
// hedged class pays the strategy's return in the holder's currency either
// way.
func mfehRecipe() Recipe {
	return Recipe{
		ID:   "MFEH",
		Name: "iMGP DBi Managed Futures EUR-hedged: the USD chain, hedged to EUR",
		Method: "the same USD donor chain as the UCITS USD class (the US-listed DBMF from 2019, the net all-styles composite half-projected onto the fund's ten futures 2000-03, the raw composite before it, Man AHL Diversified 1996-03, fee-aligned to the 0.75%/yr UCITS load) " +
			"hedged to EUR via the FX-hedge identity (− USD cash ^IRX + EUR cash EURCASH-EUR); real MFEH grafted from its 2026-05 inception",
		Donors: append([]string{"DBMF"}, dbiDonors...),
		Build:  mfehBuild,
		// ValidateAgainst deliberately empty at 2026-08: the class is weeks
		// old and its real NAVs fall 3 points short of the 60-point overlap
		// the grader requires. Re-add `ValidateAgainst: "MFEH"` at the next
		// regeneration; the real quotes are already grafted meanwhile.
		SpliceReal: "MFEH",
	}
}

// mfehBuild runs the USD chain and applies the hedge day by day, reading the
// carry off the two cash series themselves (as aqrHedgedRecon does) so the
// chain's own trading calendar is kept and no date is left unhedged.
func mfehBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	cfg := mfConfig(0.115, 0.0075)
	usd, err := dbiChain("MFEH (USD donor chain)", feeAligned("MFEH", append([]string{"DBMF"}, dbiDonors...)), cfg)(f, from)
	if err != nil {
		return nil, err
	}
	irxLvl, err := extend(f).Fetch(cfg.CashID, from)
	if err != nil {
		return nil, err
	}
	eurIdx, err := f.Fetch("EURCASH-EUR", from)
	if err != nil {
		return nil, err
	}
	hedged := &marketdata.Series{Name: "iMGP DBi Managed Futures (EUR-hedged)", Source: "simdata", Currency: "EUR"}
	val := 100.0
	hedged.Points = append(hedged.Points, marketdata.Point{Date: usd.Points[0].Date, Close: val})
	for i := 1; i < len(usd.Points); i++ {
		prev, cur := usd.Points[i-1].Date, usd.Points[i].Date
		rUSD := usd.Points[i].Close/usd.Points[i-1].Close - 1
		val *= 1 + rUSD + eurCashReturn(eurIdx, prev, cur) - cashAccrual(irxLvl, prev, cur)
		hedged.Points = append(hedged.Points, marketdata.Point{Date: cur, Close: val})
	}
	return hedged, nil
}

// kmlmRecipe reconstructs KMLM from the same TSMOM engine at a higher vol
// target, real KMLM quotes grafted on top (see dbmf/kmlm/cta note above).
func kmlmRecipe() Recipe {
	return Recipe{
		ID:              "KMLM",
		Name:            "KraneShares KMLM: TSMOM replication",
		Method:          "real NAVs of other managed-futures programmes spliced behind the fund, each lifted to its 0.90%/yr fee load (Virtus AlphaSimplex 2010-08→ +0.55%/yr, Guggenheim 2007-02→ +1.09%/yr, Man AHL Diversified 1996-03→ +1.84%/yr), the file starting at that deepest donor's own first NAV; the 12-month TSMOM engine (14% vol target ~ the fund's realized 14.7%) supplies the daily texture the weekly-dealing donor is projected onto; real KMLM grafted from 2020",
		Donors:          kmlmDonors,
		Build:           chainedTrend("KMLM (donor chain)", "KMLM", feeAligned("KMLM", kmlmDonors), mfConfig(0.14, 0.0090)),
		ValidateAgainst: "KMLM",
		SpliceReal:      "KMLM",
	}
}

// aqrmfRecipe reconstructs the AQR Managed Futures UCITS Fund (LU1103257975,
// USD, real from 2015-05) from the same TSMOM engine, real AQR quotes grafted
// on top. AQR's multi-horizon trend model is genuinely distinct from the
// SG/DBi replication that RSBT and DBMF track, and its live record (2015→) has
// materially outrun the MLM index (KMLM), so it is the better real-world pick
// for the "second, independent trend method" slot. The proxy vol target (9%)
// matches the volatility the class itself realized over its live window
// (9.3%), below DBMF (11.5%) and KMLM (14%). USD native, so no FX leg (the
// fetch layer converts to view).
//
// The 0.79 %/yr fee charged to the reconstruction is class A's BASE cost only
// (mgmt 0.60% + 0.18% expense cap + 0.01% subscription tax, per the prospectus
// supplement). Class A also pays a 10% performance fee over the ML 3-Month
// T-Bill hurdle, which the audited accounts put at 1.58 points of average class
// NAV for the year to 31 March 2026. That fee is deliberately NOT modelled: it
// only bites in strong trend years, and the reconstruction's monthly path comes
// from a composite of real programmes already net of their own managers' fees,
// performance fees included, so charging it again would double-count. The
// consequence is that the reconstructed pre-2015 leg carries an industry fee
// load rather than this class's own; prefer the flat-fee EUR class
// (LU1662501532, aqrmfHedgedRecipe) when the cost of a strong decade matters.
//
// That same performance fee is why the AQMIX donor takes no fee uplift although
// its own list price sits 0.50 points above this class's base: see feeAligned.
func aqrmfRecipe() Recipe {
	return Recipe{
		ID:   "LU1103257975",
		Name: "AQR Managed Futures UCITS: TSMOM replication",
		Method: "the manager's own US fund (AQMIX, same programme: monthly correlation 0.93 over their 11-year overlap, no fee uplift since the class's own performance fee already covers the difference) from 2010, real NAVs of other managed-futures programmes behind it (Guggenheim 2007-02, Man AHL Diversified 1996-03 +0.37%/yr), " +
			"the file starting at that deepest donor's own first NAV, at a 9% vol target ~ the class's realized 9.3%, real AQR grafted from 2015",
		Donors:          aqrmfDonors,
		Build:           chainedTrend("AQR Managed Futures (donor chain)", "LU1103257975", feeAligned("LU1103257975", aqrmfDonors), mfConfig(0.09, 0.0079)),
		ValidateAgainst: "LU1103257975",
		SpliceReal:      "LU1103257975",
	}
}

// aqrBEURFeeWedge is the annualized return the B EUR donor class
// (LU1103258197) gives up to the flat-fee RAEF class (LU1662501532) in a trend
// DROUGHT, added back to the donor segment of the backcast. It is the low end
// of the wedge measured on the four disjoint windows where both classes have
// real NAVs (RAEF CAGR minus B EUR CAGR, common dates only):
//
//	2021-04-01..2021-12-21  176 d  RAEF -10.74%/yr  +0.45 pts/yr
//	2023-10-20..2024-12-31  288 d  RAEF  +1.75%/yr  +0.61 pts/yr
//	2025-01-01..2025-12-31  237 d  RAEF +13.83%/yr  +1.58 pts/yr
//	2026-01-01..2026-07-15  128 d  RAEF +13.75%/yr  +1.91 pts/yr
//
// The sign is the same on all four and the magnitude tracks the fund's own
// return, which is the signature of the 10% performance fee class B pays over
// an EUR STR hurdle and RAEF does not: folding the post-gap window by month,
// class B lags RAEF by 1.92 pts/yr in the months RAEF gains and leads it by
// 0.11 pts/yr in the months RAEF loses. A constant is therefore the wrong
// shape for the fee, and the only defensible constant is the smallest one.
//
// 0.45 pts/yr is also the regime-matched one: it comes from the single window
// in which the fund was losing money, and the donor segment it corrects
// (2015-03 to 2021-12) is a drought throughout, the class compounding at
// -4.45%/yr over it. In any better regime the correction is too small rather
// than too large, so the backcast stays conservative where trend earns its
// keep.
const aqrBEURFeeWedge = 0.0045

// aqrmfHedgedRecipe backcasts the EUR-hedged retail share class of the AQR
// Managed Futures UCITS fund (LU1662501532 "RAEF", flat fee / no performance
// fee, real from 2021-04). Before RAEF's inception the fund itself already
// existed, so the backcast prefers a REAL sister share class over any
// reconstruction: the legacy B EUR class (LU1103258197, EUR-hedged) carries
// continuous real NAVs from 2015-03 until its class was emptied in 2021-12.
//
// It is the same trade in the same wrapper, and it measures like it: on the
// 2021 overlap the two classes agree at a daily correlation of 1.000, and on
// the post-gap overlap (2023-10-20 to 2026-07-15, 655 common days) at 0.999
// daily and 0.999 monthly. Nothing else comes close. The manager's own US fund
// hedged into EUR, which is what the pre-2015 leg falls back on, reaches only
// 0.957 and 0.963 monthly on those same two windows, so the sister class keeps
// the 2015-2021 slot.
//
// What separates the two classes is cost, not strategy: class B carries a 10%
// PERFORMANCE fee over the EUR STR hurdle, high-on-high, crystallised each
// 31 March, that RAEF does not (2.03 points of average class NAV in the year
// to 31 March 2026, per the audited accounts). The donor segment is therefore
// lifted by a constant aqrBEURFeeWedge (0.45 %/yr), the low end of the wedge
// measured on four disjoint windows and the one measured in a drought like the
// donor's own; see that constant for the table and for why a constant cannot
// be the whole answer. Modelling the fee properly (10% of the excess over EUR
// STR above a running high-water mark) is the remaining refinement; it is not
// done here because the crystallisation path per share class is not observable
// from daily NAVs alone.
//
// Only before 2015-03 does the series fall back to the TSMOM reconstruction (same engine
// and IR pin as aqrmfRecipe), re-expressed EUR-hedged via the standard
// FX-hedge identity used by the EUR-hedged TIPS recipe (tip1eRecipe): a
// hedged foreign return equals the local (USD) return minus the USD cash
// rate (^IRX) plus the EUR cash rate (EURCASH-EUR). Real LU1662501532 quotes
// are grafted on top from the class's 2021-04 inception at fetch time.
func aqrmfHedgedRecipe() Recipe {
	return Recipe{
		ID:              "LU1662501532",
		Name:            "AQR Managed Futures UCITS (EUR-hedged): real B EUR sister class over a hedged TSMOM backcast",
		Method:          "real B EUR sister class (LU1103258197, same fund, daily correlation 1.000 on the 2021 overlap and 0.999 on the 2023-2026 one) from 2015-03 to its 2021-12 NAV gap, lifted by a constant 0.45 %/yr, the low end of the measured 0.45-1.91 pts/yr the donor's 10% performance fee over EUR STR costs it against the flat-fee class; before that the same USD donor chain as the unhedged class (AQMIX 2010, Guggenheim 2007-02, Man AHL Diversified 1996-03, where the file starts) hedged to EUR via the FX-hedge identity (− USD cash ^IRX + EUR cash EURCASH-EUR); real LU1662501532 grafted from its 2021-04 inception",
		Build:           aqrHedgedBuild,
		Donors:          append([]string{"LU1103258197"}, aqrmfDonors...),
		ValidateAgainst: "LU1662501532",
		SpliceReal:      "LU1662501532",
	}
}

// aqrHedgedBuild builds the EUR-hedged AQR series for the FLAT-fee class
// (RAEF): the shared reconstruction with the B EUR donor lifted by
// aqrBEURFeeWedge, the performance fee the donor pays and RAEF does not.
func aqrHedgedBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	return aqrHedgedRecon(f, from, -aqrBEURFeeWedge,
		"AQR Managed Futures B EUR (real donor, performance fee added back)")
}

// aqrHedgedRecon builds the EUR-hedged AQR series every EUR class of the
// sub-fund shares: the TSMOM reconstruction hedged to EUR (r_eur = r_usd −
// usd_cash + eur_cash, accruals from the strategy's own frame), then the real
// B EUR sister class spliced over it, charged donorAdj and truncated at its
// first long NAV gap. When the donor is unavailable (offline builds) the
// reconstruction stands alone.
//
// donorAdj is what carries the donor from ITS price list to the target's, in
// the sign afterAnnualFee takes: positive charges the donor (the target is
// dearer), negative lifts it (the target is cheaper). Every EUR class holds
// the same portfolio, so this difference is the whole of what separates their
// NAVs, and the caller reads it off published ongoing charges rather than off
// an observed return gap. It is one number and not a schedule because the
// classes' price lists have not moved over the donor's era.
//
// The truncation is load bearing beyond ending the segment where the class was
// emptied. The donor's own NAV history resumes on 2023-10-19 with one stale
// print at its pre-gap level (113.91) before the re-seeded class prints 99.70
// the next day: read as a return, that single artefact is a -12.5% day, and it
// alone drags the post-gap daily correlation with RAEF from 0.999 to 0.777.
// Nothing downstream needs those NAVs anyway, since real RAEF quotes cover
// 2021-04 onward, so the head of the series is the only part kept.
func aqrHedgedRecon(f Fetcher, from time.Time, donorAdj float64, donorName string) (*marketdata.Series, error) {
	cfg := mfConfig(0.09, 0.0079)
	// The USD leg is the same donor chain the unhedged class uses: the
	// manager's own US fund from 2010, another manager's programme behind it,
	// and the reconstruction only for what neither reaches.
	// Fee-aligned as class A, because it IS the class A chain: only the pre-2015
	// leg of it survives here, and the B EUR donor spliced over that carries its
	// own measured wedge (aqrBEURFeeWedge) instead.
	usd, err := chainedTrend("AQR MF USD (donor chain)", "LU1103257975",
		feeAligned("LU1103257975", aqrmfDonors), cfg)(f, from)
	if err != nil {
		return nil, err
	}

	// EUR-minus-USD carry, read off the two cash series themselves rather than
	// off a frame: the donor chain keeps the donors' own trading calendar, and
	// a date the frame happens not to hold would otherwise be left unhedged.
	irxLvl, err := extend(f).Fetch(cfg.CashID, from)
	if err != nil {
		return nil, err
	}
	eurIdx, err := f.Fetch("EURCASH-EUR", from)
	if err != nil {
		return nil, err
	}
	hedged := &marketdata.Series{Name: "AQR Managed Futures (EUR-hedged)", Source: "simdata", Currency: "EUR"}
	val := 100.0
	hedged.Points = append(hedged.Points, marketdata.Point{Date: usd.Points[0].Date, Close: val})
	for i := 1; i < len(usd.Points); i++ {
		prev, cur := usd.Points[i-1].Date, usd.Points[i].Date
		rUSD := usd.Points[i].Close/usd.Points[i-1].Close - 1
		val *= 1 + rUSD + eurCashReturn(eurIdx, prev, cur) - cashAccrual(irxLvl, prev, cur)
		hedged.Points = append(hedged.Points, marketdata.Point{Date: cur, Close: val})
	}

	// Prefer the real B EUR sister class over the reconstruction wherever it
	// has continuous NAVs: truncate it at its first long gap (the class was
	// emptied in 2021-12).
	donor, err := f.Fetch("LU1103258197", from)
	if err != nil || donor == nil || len(donor.Points) == 0 {
		return hedged, nil
	}
	donor = truncateAtGap(donor, 30)
	donor = afterAnnualFee(donorName, donor, donorAdj)
	return Splice(donor, hedged), nil
}

// Ongoing charges of the AQR Managed Futures UCITS EUR classes: the part of
// each class's price list that is levied INSIDE its net asset value, as a
// fraction per year. Sources agree to a basis point: the Financial Times
// tearsheet of each class, and the audited Total Net Expense Ratio the Swiss
// annex of the report publishes per class before any performance fee (RAEF
// 0.24 %, IAET 0.78 % as at 30 September 2025).
//
// These are what separates the classes' NAVs, and NOT the headline management
// charge: RAEF's prospectus list is 1.29 % (1.00 management, 0.24 expense cap,
// 0.05 taxe d'abonnement), yet only 0.23 % ever reaches its NAV, because the
// prospectus lets the manager waive the management fee at its sole discretion
// and for that class it does. Its published record is therefore that of a share
// class carrying a quarter point of costs, and any comparison of these classes
// on NAV alone reads a fee plumbing difference as manager skill.
//
// The waiver is revocable, which makes aqrRAEFOngoing the least durable number
// in this file: reinstating the fee would take RAEF to 1.29 %/yr and would move
// the IAE1FT conversion below from 1.03 to nearly zero. Re-read the FT ongoing
// charge and the audited Swiss TNER before trusting it after a refresh. The
// other three are prospectus arithmetic and do not drift: IAE1FT is exactly
// 1.00 + 0.25 + 0.01, and IAET's cap sits 0.05 above B EUR's, which is the
// whole of the alignment aqrIAETRecipe applies. See docs/aqr-mf.txt.
const (
	aqrRAEFOngoing   = 0.0023 // LU1662501532, flat, no performance fee
	aqrBEUROngoing   = 0.0073 // LU1103258197, plus a 10 % performance fee
	aqrIAETOngoing   = 0.0078 // LU1662498721, plus a 10 % performance fee
	aqrIAE1FTOngoing = 0.0126 // LU2622190622, flat, no performance fee
)

// aqrIAETRecipe backcasts the IAET EUR class of the AQR Managed Futures UCITS
// fund (LU1662498721, real from 2022-03), an institutional-series class that
// retail platforms list and that is, for some buyers, the only reachable way
// into this programme.
//
// Its donor is the legacy B EUR class (LU1103258197) rather than the flat-fee
// RAEF, and the reason is that the two share a FEE STRUCTURE and not merely a
// portfolio: 0.60 % management inside a ~0.75 % ongoing charge, plus the same
// 10 % performance fee over the same EUR STR hurdle, high-on-high, crystallised
// each 31 March. A performance fee cannot be removed from a NAV after the fact
// (the crystallisation path is not observable from daily prices, which is what
// forces the flat-fee class to correct its donor with a constant, see
// aqrBEURFeeWedge); it does not need to be removed when the target pays the
// same one. The donor carries it, already correctly shaped, for free.
//
// What is left to align is the ongoing charge, 0.78 % against the donor's
// 0.73 %, so the donor segment is charged 0.05 %/yr. Measurement agrees and was
// taken before the recipe was written: over the 672 common days the two classes
// both quote (2023-10-20 to 2026-08-07, the donor's stale re-seeding print
// dropped), their daily returns correlate at 0.9991 and IAET trails B EUR by
// 0.096 points a year, against the 0.05 the price lists predict.
//
// Before the donor's own 2015-03 start the series falls back on the same
// EUR-hedged reconstruction the other classes use, fee-aligned as class A.
// That alignment fits this class better than any other: class A's list (0.79 %
// plus 10 % of the excess over a cash hurdle) is IAET's list to a basis point.
func aqrIAETRecipe() Recipe {
	return Recipe{
		ID:   "LU1662498721",
		Name: "AQR Managed Futures UCITS (IAET EUR): the sister class that pays the same fees",
		Method: "real B EUR sister class (LU1103258197, same 0.60% management and same 10% performance fee over EUR STR, daily correlation 0.9991 on their 672-day overlap) from 2015-03 to its 2021-12 NAV gap, charged 0.05 %/yr for the difference between the two published ongoing charges (0.78 against 0.73); " +
			"before that the same USD donor chain as the unhedged class (AQMIX 2010, Guggenheim 2007-02, Man AHL Diversified 1996-03, where the file starts) hedged to EUR via the FX-hedge identity (− USD cash ^IRX + EUR cash EURCASH-EUR); real LU1662498721 grafted from its 2022-03 inception",
		Build:           aqrIAETBuild,
		Donors:          append([]string{"LU1103258197"}, aqrmfDonors...),
		ValidateAgainst: "LU1662498721",
		SpliceReal:      "LU1662498721",
	}
}

// aqrIAETBuild is the shared EUR-hedged reconstruction with the B EUR donor
// charged the difference between the two classes' ongoing charges (positive:
// IAET is the dearer of the two), then RAEF laid over whatever the donor's
// 2021-12 emptying leaves uncovered.
//
// That bridge exists because the donor stops two and a half months before IAET
// starts quoting, which would leave the file with a hole AND with no window on
// which to validate it at all. RAEF is the only real EUR class covering the
// stretch, and it is flat-fee, so it is charged the base difference of the two
// ongoing charges and NOTHING for the performance fee IAET pays and it does
// not: that fee is not removable from a NAV after the fact, and it is not
// modelled here.
//
// Two consequences, both deliberate. The file's consumed reconstruction carries
// no performance fee over 2021-12 to 2022-03, eleven weeks out of thirty years.
// And the validation window, which lies entirely inside the bridge, reads that
// missing fee as a level gap of roughly three quarters of a point a year: it
// grades the bridge, not the donor. The donor's own grade is the measurement in
// aqrIAETRecipe, 0.9991 daily correlation and 0.096 points a year against the
// price lists' 0.05.
func aqrIAETBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	donorEra, err := aqrHedgedRecon(f, from, aqrIAETOngoing-aqrBEUROngoing,
		"AQR Managed Futures B EUR (real donor, charged IAET's ongoing charge)")
	if err != nil {
		return nil, err
	}
	bridge, err := f.Fetch("LU1662501532", from)
	if err != nil || bridge == nil || len(bridge.Points) == 0 {
		return donorEra, nil
	}
	bridge = afterAnnualFee("AQR Managed Futures RAEF EUR (bridge, charged IAET's ongoing charge)",
		trimBefore(bridge, donorEra.Last().Date), aqrIAETOngoing-aqrRAEFOngoing)
	if len(bridge.Points) == 0 {
		return donorEra, nil
	}
	return Splice(bridge, donorEra), nil
}

// trimBefore drops the points of s dated before d, keeping the tail. It returns
// an empty series when nothing is left, which a caller must handle.
func trimBefore(s *marketdata.Series, d time.Time) *marketdata.Series {
	out := *s
	out.Points = nil
	for _, p := range s.Points {
		if !p.Date.Before(d) {
			out.Points = append(out.Points, p)
		}
	}
	return &out
}

// aqrIAE1FTRecipe backcasts the IAE1FT EUR class (LU2622190622, launched
// 2026-05), the flat-fee institutional-series class: 1.26 % of ongoing charge
// inside the NAV and NO performance fee.
//
// Its donor is RAEF (LU1662501532), the other flat-fee EUR class, taken whole:
// the real quotes from 2021-04 and, behind them, the reconstruction RAEF's own
// recipe assembles. Same portfolio, same absence of a performance fee, so the
// conversion is one constant: 1.03 %/yr, the difference between the two
// published ongoing charges (1.26 against 0.23). Inheriting RAEF's series also
// inherits the one judgement call in it, the constant that corrects its own B
// EUR donor, which is the right outcome: two flat-fee classes of one fund
// should not disagree about their shared past.
//
// Measurement agrees and was taken before the recipe was written: over the 59
// common days the two classes both quote (2026-05-11 to 2026-08-07) their daily
// returns correlate at 1.0000 and IAE1FT trails RAEF by 0.995 points a year,
// against the 1.03 the price lists predict.
//
// The economics are worth stating because they are not what the two headline
// numbers suggest: this class and RAEF carry the same 1.00 % management charge,
// but here it is levied inside the NAV and there it is not. A holder of this
// class sees its whole cost in the price it is quoted at.
func aqrIAE1FTRecipe() Recipe {
	return Recipe{
		ID:   "LU2622190622",
		Name: "AQR Managed Futures UCITS (IAE1FT EUR): the flat-fee sister class, charged in the NAV",
		Method: "the flat-fee RAEF class (LU1662501532) whole, its real quotes from 2021-04 over its own reconstruction (real B EUR 2015-03, then AQMIX 2010, Guggenheim 2007-02, Man AHL Diversified 1996-03 hedged to EUR, where the file starts), " +
			"charged 1.03 %/yr for the difference between the two published ongoing charges (1.26 against 0.23), the two classes being flat-fee and otherwise identical; real LU2622190622 grafted from its 2026-05 inception",
		Build:           aqrIAE1FTBuild,
		Donors:          append([]string{"LU1662501532", "LU1103258197"}, aqrmfDonors...),
		ValidateAgainst: "LU2622190622",
		SpliceReal:      "LU2622190622",
	}
}

// aqrIAE1FTBuild is RAEF's whole published history, real quotes included,
// charged the difference between the two flat-fee classes' ongoing charges.
// The real RAEF quotes are spliced here rather than left to SpliceReal, which
// serves the TARGET's own quotes and knows nothing of a sister class.
//
// The charge is a SCHEDULE and not a constant because RAEF's own series is not
// at RAEF's price list throughout. From its B EUR donor's first NAV it is: the
// donor is lifted to the flat-fee class's economics and the real quotes after
// it ARE that class. Before that date the leg is the TSMOM reconstruction
// fee-aligned as class A, whose list (0.79 % plus a performance fee worth 1.58
// points of average class NAV in the year to 31 March 2026) is already heavier
// than this class's 1.26 % flat. Charging the difference there as well would
// correct one fee twice, which is the quiet way to lose a point a year, so
// that era is charged nothing and stays where RAEF's own recipe leaves it.
func aqrIAE1FTBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	recon, err := aqrHedgedBuild(f, from)
	if err != nil {
		return nil, err
	}
	whole := recon
	if real, rerr := f.Fetch("LU1662501532", from); rerr == nil && real != nil && len(real.Points) > 0 {
		whole = Splice(real, recon)
	}
	donor, derr := f.Fetch("LU1103258197", from)
	if derr != nil || donor == nil || len(donor.Points) == 0 {
		return nil, fmt.Errorf("IAE1FT: the B EUR donor dates the fee schedule: %w", derr)
	}
	steps := []feeStep{
		{Annual: 0},
		{From: donor.Points[0].Date, Annual: aqrIAE1FTOngoing - aqrRAEFOngoing},
	}
	return afterFeeSteps("AQR Managed Futures RAEF EUR (real class, charged IAE1FT's ongoing charge)",
		whole, steps), nil
}

// truncateAtGap cuts a series at its first quote gap longer than maxDays
// calendar days, keeping the continuous head.
func truncateAtGap(s *marketdata.Series, maxDays int) *marketdata.Series {
	for i := 1; i < len(s.Points); i++ {
		if s.Points[i].Date.Sub(s.Points[i-1].Date).Hours() > float64(maxDays)*24 {
			t := *s
			t.Points = s.Points[:i]
			return &t
		}
	}
	return s
}

// ctaRecipe reconstructs Simplify CTA behind the published pure-trend
// composite, real CTA quotes grafted on top (see dbmf/kmlm/cta note above).
//
// The index is the donor here for the reason it is one for the DBi family: a
// fund tracks a published trend index better than it tracks another manager's
// fund. The four single funds that held these two decades until 2026-08 reached
// 0.39 monthly against CTA's live window between them; the pure-trend composite
// reaches 0.58, and closes the level gap from -3.4 to about -3.2 points a year.
//
// This file is the loosest fit of the family and the reason is the fund, not
// the engine: the engine's own monthly correlation with the real quotes (0.574)
// is the donor index's correlation with them (0.573), so the TSMOM texture adds
// no path error of its own, and no published index does better. Simplify CTA is
// not a pure trend programme. Its 2024 shareholder report attributes a 14.5 %
// year against 3.5 % for its named benchmark to "short interest rate and
// related positions" and to carry earned on curve shape, which is a book no
// trend index holds. Three months of 2024 (April, August, October: +7.9, +5.5
// and +7.2 points against the donor) carry the whole level gap; drop that
// calendar year and the fund's excess over the donor falls from +5.6 to
// +1.6 %/yr on a 14 %/yr tracking error, one fifth of a standard error from
// zero. There is no systematic coldness to correct, and correcting it would
// write one year of a manager's rates bet into three decades of history.
//
// The broader all-styles composite is the index the fund's own report names,
// and it correlates better on this short window (0.63 against 0.57), yet it is
// NOT used: it runs at half the fund's volatility so the chain would have to
// lever it 1.9 times (against 1.5 for the pure-trend one), close to the point
// where volMatch stops believing two series are the same trade, and measured
// end to end it leaves the reconstruction 5.5 points a year cold instead of
// 3.2. A correlation bought by levering an index nearly twofold is not the
// trade this file wants.
func ctaRecipe() Recipe {
	return Recipe{
		ID:              "CTA",
		Name:            "Simplify CTA: TSMOM replication",
		Method:          "the published net pure-trend composite, the index closest to the fund's own volatility, spliced behind it from 2000-01 and lifted +1.25%/yr to put its constituents' 2%/yr management load on the fund's own 0.75%, then real NAVs of Man AHL Diversified (1996-03, +1.99%/yr), the file starting at that deepest donor's own first NAV; the 12-month TSMOM engine (16% vol target ~ the fund's realized 16.9%) supplies the daily texture the weekly-dealing donor is projected onto; real CTA grafted from 2022",
		Donors:          ctaDonors,
		Build:           chainedTrend("CTA (donor chain)", "CTA", feeAligned("CTA", ctaDonors), mfConfig(0.16, 0.0075)),
		ValidateAgainst: "CTA",
		SpliceReal:      "CTA",
	}
}
