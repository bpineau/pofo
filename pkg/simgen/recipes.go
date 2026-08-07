package simgen

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// ComponentsFrom is how far back component histories are requested; actual
// frames start at the youngest component's first quote.
var ComponentsFrom = time.Date(1962, 1, 1, 0, 0, 0, 0, time.UTC)

// minBackcastR2 is the floor under which a regression-based reconstruction
// is considered too unfaithful to be written at all.
const minBackcastR2 = 0.35

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
		wintonRecipe(),
		zrozRecipe(),
		dbxgRecipe(),
		mthRecipe(),
		dbmfRecipe(),
		dbmfpaRecipe(),
		dbmfeRecipe(),
		kmlmRecipe(),
		aqrmfRecipe(),
		aqrmfHedgedRecipe(),
		indepEuropeRecipe(),
		ctaRecipe(),
		amundiVolRecipe(),
		bhmgRecipe(),
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
// Treasuries 20+yr, duration ~17) and currency (USD, the fund's own quote
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
func dtlaRecipe() Recipe {
	return Recipe{
		ID:              "DTLA",
		Name:            "iShares $ Treasury 20+ Acc: VUSTX long Treasury",
		Method:          "1.13×VUSTX (Vanguard Long-Term Treasury, 1986→, extended TREASURY-LONG daily from 1962, geared to the 20+ duration), real DTLA grafted from 2018",
		Build:           composite("DTLA (long Treasury, accumulating)", []Leg{{ID: "VUSTX", Weight: longTreasuryGearing}}, "", 0),
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
// (LU0290358497, EUR, real from 2007) from the same EUR money-market index as
// ERNX, less the fund's 0.10%/yr TER. EURCASH-EUR compounds the 3-month
// interbank rate, which runs above the overnight ESTR the fund tracks by a
// term premium (~0.2%/yr on average over the overlap); the TER deduction only
// partly offsets it, so the proxy sits ~0.2%/yr above the real fund. That
// residual is immaterial for a cash sleeve and, since real XEON is grafted from
// 2007, only touches the pre-2007 tail. The real quotes take over from 2007.
func xeonRecipe() Recipe {
	return Recipe{
		ID:     "LU0290358497",
		Name:   "Xtrackers EUR Overnight Rate Swap: EUR overnight cash",
		Method: "EUR money-market index EURCASH-EUR (3-month interbank compounded, 1994->, interpolated to business days) less 0.10%/yr TER; sits ~0.2%/yr above ESTR (3M-vs-overnight term premium only partly offset); real XEON grafted from 2007",
		Build: func(f Fetcher, from time.Time) (*marketdata.Series, error) {
			s, err := eurCashDaily(f, from)
			if err != nil {
				return nil, err
			}
			return afterFee(s, 0.0010), nil
		},
		ValidateAgainst: "LU0290358497",
		SpliceReal:      "LU0290358497",
	}
}

// eurCashDaily fetches the bundled EUR money-market index (EURCASH-EUR, the
// FRED 3-month interbank rate compounded, monthly from 1994) and expands it to
// business-day granularity by geometric interpolation between the monthly
// anchors. A money-market accrual index grows by a near-constant daily rate
// within each month, so unlike a risk asset it carries no real intramonth
// variation to lose: the interpolation is faithful and yields a genuinely
// daily series (rather than feeding month-sized steps to daily statistics),
// with no external data. The last anchor is appended as-is so the series ends
// exactly on a real EURCASH-EUR level.
func eurCashDaily(f Fetcher, from time.Time) (*marketdata.Series, error) {
	m, err := f.Fetch("EURCASH-EUR", from)
	if err != nil {
		return nil, err
	}
	if m == nil || len(m.Points) < 2 {
		return nil, fmt.Errorf("EURCASH-EUR: empty history")
	}
	s := &marketdata.Series{Name: "EUR cash (money-market, daily)", Source: "simdata"}
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
	return s, nil
}

// dpgtRecipe rebuilds the Dimensional Global Targeted Value UCITS ETF
// (IE000S67ID55, launched 2025) from Dimensional's own long-running US and
// international small-cap value mutual funds, the same shop and factor design,
// blended 60/40 US / developed-ex-US, net of the 0.44% TER. The only market
// quote is the LSE line in GBP, so the USD blend is re-expressed in GBP at the
// GBP/USD spot rate (GBPUSD=X extended to 1971 by the daily FRED refdata, so
// the start is set by DISVX ~1994) to match the real series, which is grafted
// from inception.
func dpgtRecipe() Recipe {
	return Recipe{
		ID:              "IE000S67ID55",
		Name:            "Dimensional Global Targeted Value: DFA small-cap value blend (GBP)",
		Method:          "0.60×DFSVX (US small value) + 0.40×DISVX (intl developed small value), 0.44%/yr fees, converted USD→GBP at GBPUSD spot (FRED daily refdata back to 1971), real DPGT grafted from 2025",
		Build:           dpgtBuild,
		ValidateAgainst: "IE000S67ID55",
		SpliceReal:      "IE000S67ID55",
	}
}

// dpgtBuild builds the 60/40 DFA small-cap value blend in USD, then converts
// each daily return into GBP via the GBP/USD spot rate (a GBP-denominated NAV
// equals the USD NAV divided by the USD-per-GBP rate), so the simulated
// history matches the GBP quote the real DPGT trades in. The cross is
// forward-filled onto the blend's own trading calendar (see fxOnDates)
// rather than joined into the frame, which would pollute the calendar with
// the FX feed's weekend prints.
func dpgtBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	legs := []Leg{{ID: "DFSVX", Weight: 0.60}, {ID: "DISVX", Weight: 0.40}}
	fr, err := BuildFrame(extend(f), []string{"DFSVX", "DISVX"}, from)
	if err != nil {
		return nil, err
	}
	usd, err := Composite(fr, legs, "", 0.0044)
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
// (IE0003R87OG3, AVWS, EUR-quoted on the FT line, launched 2024-09-25) from
// Dimensional's long-running small-cap value mutual funds, the same size+value
// design, blended 40/60 US / developed-ex-US to match the fund's ~40 % North
// America geography (versus DPGT's 60/40 US tilt), net of the 0.39 % TER. The
// blend is built in USD and re-expressed in EUR at the EURUSD spot (Yahoo daily,
// ~2003, as for the DBMFE and NTSZ euro legs), which sets the start, so the real
// series it splices onto (also EUR) lines up in currency. Real Avantis quotes
// are grafted from inception. This makes the file-recommended global buy option
// backtestable; ZPRV (US small-cap value) still reaches deeper, to 1963.
func avantisRecipe() Recipe {
	return Recipe{
		ID:              "IE0003R87OG3",
		Name:            "Avantis Global Small Cap Value: DFA small-cap value blend (EUR)",
		Method:          "0.40×DFSVX (US small value) + 0.60×DISVX (intl developed small value), 0.39%/yr fees, converted USD→EUR at EURUSD spot (Yahoo daily ~2003), real Avantis grafted from 2024",
		Build:           avantisBuild,
		ValidateAgainst: "IE0003R87OG3",
		SpliceReal:      "IE0003R87OG3",
	}
}

// avantisBuild builds the 40/60 DFA small-cap value blend in USD, then converts
// each daily return into EUR via the EURUSD spot rate, so the simulated history
// matches the EUR quote the real Avantis Global Small Cap Value trades in. The
// construction mirrors dpgtBuild (its GBP-quoted Dimensional sibling), with the
// US/international split shifted to the Avantis fund's regional weights.
func avantisBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	legs := []Leg{{ID: "DFSVX", Weight: 0.40}, {ID: "DISVX", Weight: 0.60}}
	fr, err := BuildFrame(extend(f), []string{"DFSVX", "DISVX"}, from)
	if err != nil {
		return nil, err
	}
	usd, err := Composite(fr, legs, "", 0.0039)
	if err != nil {
		return nil, err
	}
	return convertDaily("Avantis Global SCV (USD small-value blend expressed in EUR)",
		extend(f), "EURUSD=X", from, fr.Dates, usd)
}

// scvwRecipe rebuilds US small-cap value from DFA US Small Cap Value
// (DFSVX, 1993→, total return), itself extended before 1993 by the Ken French
// small-value factor (refdata USSCV-USD, daily from 1963-07), with the real
// SPDR ZPRV grafted on top. Cross-checked once against the MSCI USA Small Cap
// Value Weighted index (weekly corr 0.90, CAGR 11.4% vs 10.4% over 1997-2015)
// to confirm faithfulness.
func scvwRecipe() Recipe {
	return Recipe{
		ID:              "IE00BSPLC413",
		Name:            "SPDR MSCI USA Small Cap Value Weighted",
		Method:          "Ken French small-value factor (USSCV-USD, daily 1963-07→) spliced before DFSVX (DFA US Small Cap Value, 1993→); real ZPRV grafted from 2015",
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

// chsnRecipe backcasts the UBS Core Euro Inflation Linked 1-10 ETF
// (LU1645380442, a brand-new 2025 EUR-acc share class) from the longer-running
// iShares Euro Inflation Linked Govt Bond ETF (IBCI, daily FT NAVs from its
// 2005 inception). IBCI tracks the all-maturity euro-area linker index
// (duration ~8) versus ~4.8 for the 1-10 segment, so it is held at 0.60x with
// the rest in EUR cash (bundled EURCASH-EUR money-market index) to match
// durations; same asset class and currency (EUR), no FX leg needed. The
// scaling brought the validation beta from 0.53 to 0.85 and the tracking
// error from 2.3 to 1.5%/yr, and the FT source moved the start from 2009
// back to 2005. The real CHSN quotes are grafted on top from 2025.
func chsnRecipe() Recipe {
	return Recipe{
		ID:     "LU1645380442",
		Name:   "UBS Core Euro Inflation Linked 1-10: euro govt linker proxy",
		Method: "0.60×IBCI (iShares Euro Inflation Linked Govt Bond, all-maturity euro-linker, FT daily from 2005, duration-matched to 1-10) + 0.40×EUR cash (EURCASH-EUR); real CHSN grafted from 2025",
		Build: composite("CHSN (euro inflation-linked proxy)", []Leg{
			{ID: "IBCI", Weight: 0.60},
			{ID: "EURCASH-EUR", Weight: 0.40},
		}, "", 0),
		ValidateAgainst: "LU1645380442",
		SpliceReal:      "LU1645380442",
	}
}

// tip1eRecipe backcasts the UBS Core Bloomberg TIPS 1-10 EUR-hedged ETF
// (LU1459801780, real from 2016) as US TIPS hedged to EUR: Vanguard's
// Inflation-Protected Securities fund (VIPSX, US TIPS total return, 2000->)
// financed at USD cash (^IRX) and re-earning EUR cash (bundled EURCASH-EUR money-
// market index). This is the standard FX-hedge identity: a hedged foreign return
// equals the local return plus the domestic (EUR) cash rate minus the foreign
// (USD) cash rate, so the EUR investor pays the (usually negative, post-2015)
// EUR-minus-USD carry on top of the TIPS return. VIPSX is all-maturity (duration
// ~7), so it is held at 0.64x to match the 1-10 segment's shorter duration (~4.5,
// the rest implicitly in the hedged EUR cash leg); this brought the validation
// beta from 0.33 to ~0.5 and the overlap CAGR within ~0.1%/yr of the real fund.
// The real fund is grafted on top from its 2016 inception. Daily correlation is
// modest (~0.37) because VIPSX is a US-close mutual fund and the fund trades in
// Zurich; the weekly correlation (~0.85) is the meaningful figure.
func tip1eRecipe() Recipe {
	return Recipe{
		ID:     "LU1459801780",
		Name:   "UBS Core BBG TIPS 1-10 (EUR-hedged): US TIPS hedged to EUR",
		Method: "0.64×VIPSX (Vanguard Inflation-Protected, US TIPS TR, 2000->, duration-matched to 1-10) financed at USD cash ^IRX and re-earning EUR cash (EURCASH-EUR) = EUR-hedged TIPS; real 42C0 grafted from 2016",
		Build: composite("42C0 (EUR-hedged US TIPS)", []Leg{
			{ID: "VIPSX", Weight: 0.64, Excess: true},
			{ID: "EURCASH-EUR", Weight: 0.35},
		}, "^IRX", 0),
		ValidateAgainst: "LU1459801780",
		SpliceReal:      "LU1459801780",
	}
}

func shyRecipe() Recipe {
	return Recipe{
		ID:              "SHY",
		Name:            "iShares 1-3 Year Treasury Bond ETF",
		Method:          "VFISX (Vanguard Short-Term Treasury, 1991→), real SHY grafted from 2002",
		Build:           composite("SHY (short Treasury)", []Leg{{ID: "VFISX", Weight: 1}}, "", 0),
		ValidateAgainst: "SHY",
		SpliceReal:      "SHY",
	}
}

func rssbRecipe() Recipe {
	return Recipe{
		ID:     "RSSB",
		Name:   "Return Stacked Global Stocks & Bonds",
		Method: "100% world equity + 100% (VFITX − cash) Treasury stack (1999→), real RSSB grafted from 2023",
		Build: composite("RSSB (100/100 stocks+bonds replication)", []Leg{
			{ID: "VFINX", Weight: 0.60},
			{ID: "VTMGX", Weight: 0.30},
			{ID: "VEIEX", Weight: 0.10},
			{ID: "VFITX", Weight: 1.00, Excess: true},
		}, "^IRX", 0),
		ValidateAgainst: "RSSB",
		SpliceReal:      "RSSB",
	}
}

// gdeRecipe backcasts the WisdomTree Efficient Gold Plus Equity Strategy Fund
// (GDE, US-listed, real from 2022): for every dollar, ~90 % large-cap US equity
// held as the funded core plus a ~90 % gold overlay run through futures, with
// the residual cash as collateral. It mirrors the NTSX efficient-core method,
// swapping the Treasury overlay for gold: the equity leg earns its full return,
// the gold leg earns the spot excess over cash (gold futures ~= spot financed at
// the T-bill rate, no carry yield), and 0.10 of NAV sits in cash. Every leg is
// deep (VFINX -> S&P 500 TR, XAUUSD gold ~1968, ^IRX T-bill), so the composite
// reaches back to the gold leg's floor. Real GDE quotes are grafted from
// inception; same currency (USD), no FX leg.
func gdeRecipe() Recipe {
	return Recipe{
		ID:     "GDE",
		Name:   "WisdomTree Efficient Gold Plus Equity: 90/90 replication",
		Method: "0.90×VFINX + 0.90×(GC=F gold − cash ^IRX) overlay + 0.10×cash, daily rebalancing, 0.20%/yr fees, real GDE grafted from 2022",
		Build: composite("GDE (90/90 gold+equity replication)", []Leg{
			{ID: "VFINX", Weight: 0.90},
			{ID: "GC=F", Weight: 0.90, Excess: true},
			{ID: "^IRX", Weight: 0.10},
		}, "^IRX", 0.0020),
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
		Method:          "1.00×VFINX + 1.00×(TSMOM trend − cash ^IRX) overlay, the overlay's months and its level from the bundled net pure-trend reference (2000→, where that reference begins), 0.96%/yr fees, real RSST grafted from 2023",
		Build:           stackedTrend("RSST (100% stocks + TSMOM overlay)", "VFINX", mfConfig(0.10, 0), 0.0096),
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
		Method:          "1.00×VFITX + 1.00×(TSMOM trend − cash ^IRX) overlay, the overlay's months and its level from the bundled net pure-trend reference (2000→, where that reference begins), 0.97%/yr fees, real RSBT grafted from 2023",
		Build:           stackedTrend("RSBT (100% bonds + TSMOM overlay)", "VFITX", mfConfig(0.10, 0), 0.0097),
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

// eresMondeFeeGap is the extra yearly charge the FCPE wrapper adds on top of
// the 0.20%/yr tracker TER already embedded in the wpeaBuild path: the fund's
// all-in charge is 0.50%/yr, so 0.30 points remain to be deducted.
const eresMondeFeeGap = 0.0050 - 0.0020

// eresMondeRecipe builds ERES Xtrackers Actions Monde M (QS0009135623), the
// world-equity FCPE offered inside Eres employee-savings plans. The fund is a
// feeder into two Xtrackers MSCI World trackers (75% the swap-replicated 1D
// class, 25% the physical 1D class), so its exposure is plainly MSCI World net
// total return in EUR; what distinguishes it from a directly-held tracker is
// the wrapper's all-in 0.50%/yr charge.
//
// It is therefore built as the WPEA path (MSCI World net TR expressed in EUR,
// already net of a 0.20%/yr tracker TER, riding real iShares Core MSCI World
// over 2009-2024) with the remaining eresMondeFeeGap deducted continuously.
// Two consequences are worth stating rather than discovering later. First,
// there is nothing to validate against and no real series to graft: an FCPE
// has no public quotation, and this share class only launched in 2024-03, so
// the whole series is a reconstruction by construction, honest only insofar as
// "MSCI World minus a stated fee" describes the fund. Second, the TER-only
// convention of every tracker recipe applies: the swap leg's substitute-basket
// economics (which in practice recover part of the US dividend withholding a
// net index assumes lost) are not modelled, so the series is, if anything, a
// touch conservative for the 75% swap-replicated share.
func eresMondeRecipe() Recipe {
	return Recipe{
		ID:     "ERESMONDEM",
		Name:   "ERES Xtrackers Actions Monde M (FCPE): MSCI World net TR in EUR, 0.50%/yr all-in",
		Method: "MSCI World net TR in EUR (the WPEA path: real IWDA from 2009, MSCIWORLD-USD refdata + daily shape before, converted USD→EUR at EURUSD spot back to 1971) less the FCPE's 0.50%/yr all-in charge; no real quotes exist for an FCPE, nothing is grafted",
		Build:  eresMondeBuild,
	}
}

// eresMondeBuild is wpeaBuild net of the extra wrapper charge.
func eresMondeBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	base, err := wpeaBuild(f, from)
	if err != nil {
		return nil, err
	}
	return afterAnnualFee("ERESMONDEM (MSCI World net TR in EUR, FCPE charges)", base, eresMondeFeeGap), nil
}

// afterAnnualFee returns a copy of s with a continuous annual fee deducted:
// each daily step keeps its own return less the fee accrued over the days
// elapsed, so the drag compounds on the calendar rather than on quote count
// (holidays and weekends carry fees too). A NEGATIVE annual is an uplift, for
// a donor share class measured to carry a cost the target class does not (see
// aqrBEURFeeWedge).
func afterAnnualFee(name string, s *marketdata.Series, annual float64) *marketdata.Series {
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
		level *= step * math.Pow(1-annual, days/365.25)
		out.Points[i] = marketdata.Point{Date: s.Points[i].Date, Close: level}
	}
	return &out
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
// TSMOM trend overlay, net of 0.80%/yr fees.
func wintonRecipe() Recipe {
	return Recipe{
		ID:              "IE000O1VI174",
		Name:            "Winton Trend-Enhanced Global Equity: equities + TSMOM overlay",
		Method:          "0.60×VFINX + 0.40×VTMGX + 0.50×(TSMOM trend, its months and its level from the bundled net pure-trend reference), 0.80%/yr fees (2000→, where that reference begins)",
		Build:           wintonBuild,
		ValidateAgainst: "IE000O1VI174",
	}
}

// wintonBuild blends a 60/40 equity core with a half-weighted TSMOM trend
// overlay (the trend run as a pure excess strategy, no collateral).
func wintonBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	ids := append([]string{"^IRX", "VFINX", "VTMGX"}, mfMarkets...)
	fr, err := BuildFrame(extend(f), ids, from)
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
	const feeDaily = 0.0080 / 252
	s := &marketdata.Series{Name: "Winton Trend-Enhanced Global Equity (replication)", Source: "simdata"}
	val := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[start], Close: val})
	for i := 1; i < len(trend); i++ {
		k := start + i
		rEq := 0.6*vfinx[k] + 0.4*vtmgx[k]
		rTrend := trend[i]/trend[i-1] - 1
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
func stackedTrend(name, coreID string, cfg TSMOMConfig, annualFee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		ids := append([]string{coreID, cfg.CashID}, cfg.Markets...)
		fr, err := BuildFrame(extend(f), ids, from)
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
		core, cash := fr.Returns[coreID], fr.Returns[cfg.CashID]

		// Daily overlay excess-over-cash of the anchored reconstruction.
		overlay := make([]float64, len(trend))
		for i := 1; i < len(trend); i++ {
			overlay[i] = (trend[i]/trend[i-1] - 1) - cash[start+i]
		}

		feeDaily := annualFee / 252
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
			out = shapedSeries(alignMonthEnd(anchor, shape), shape)
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

// composite is the shared Build for constant-weight linear recipes.
func composite(name string, legs []Leg, cashID string, fee float64) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		var ids []string
		if cashID != "" {
			ids = append(ids, cashID)
		}
		for _, l := range legs {
			ids = append(ids, l.ID)
		}
		fr, err := BuildFrame(extend(f), ids, from)
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

// ntsxRecipe rebuilds the WisdomTree US Efficient Core (90 % US equities +
// 60 % treasury futures ladder) from Vanguard index funds and the T-bill
// rate. The simdata file extends the UCITS share class (IE000KF370H3); the
// validation runs against the US-listed twin NTSX, which has traded since
// 2018 with the exact same strategy.
func ntsxRecipe() Recipe {
	return Recipe{
		ID:     "IE000KF370H3",
		Name:   "WisdomTree US Efficient Core: 90/60 replication",
		Method: "0.90×VFINX + 0.60×(VFITX − cash ^IRX) + 0.10×cash, daily rebalancing, 0.20%/yr fees",
		Build: composite("NTSX (90/60 replication)", []Leg{
			{ID: "VFINX", Weight: 0.90},
			{ID: "VFITX", Weight: 0.60, Excess: true},
			{ID: "^IRX", Weight: 0.10},
		}, "^IRX", 0.0020),
		ValidateAgainst: "NTSX-US",
		SpliceReal:      "NTSX-US",
	}
}

// ntsgRecipe is the global variant (NTSG UCITS): 90 % global developed
// equities approximated as 60/40 US/international, plus the same 60 %
// treasury overlay.
func ntsgRecipe() Recipe {
	return Recipe{
		ID:     "IE00077IIPQ8",
		Name:   "WisdomTree Global Efficient Core: global 90/60 replication",
		Method: "0.54×VFINX (extended with S&P 500 TR ~1871) + 0.36×VTMGX (dev-ex-US, DEVEXUS-USD ~1969) + 0.60×(VFITX − cash ^IRX, both extended: CMT Treasury TR ~1953, T-bill ~1934) + 0.10×cash, 0.25%/yr fees; start now set by the dev-ex-US leg (~1969)",
		Build: composite("NTSG (global 90/60 replication)", []Leg{
			{ID: "VFINX", Weight: 0.54},
			{ID: "VTMGX", Weight: 0.36},
			{ID: "VFITX", Weight: 0.60, Excess: true},
			{ID: "^IRX", Weight: 0.10},
		}, "^IRX", 0.0025),
		ValidateAgainst: "IE00077IIPQ8",
	}
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
	values, err := Composite(fr, legs, "NTSZ-CASH", 0.0020)
	if err != nil {
		return nil, err
	}
	return SeriesFromFrame("NTSZ (eurozone 90/60 replication)", fr, values), nil
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
// bond total return (EUROGOV-LONG-EUR, ~1970, with the ECB daily 30-year shape
// EUROGOV-LONG-DAILY from 2004), less the fund's TER. That reference is a
// constant-maturity ~30-year par bond, modified duration ~19, matching the iBoxx
// EUR Eurozone Sovereigns 25+ index DBXG tracks; it is a genuine long-bond
// reconstruction (yield path through TreasuryTR) rather than a levered shorter
// bond, so it neither overstates the return in a sustained bond bull nor needs a
// financing leg. Euro-native throughout, no FX leg; real DBXG is grafted from
// 2007. The pre-2004 tail is monthly and its long yield is synthesized from the
// 10-year (the ECB curve only reaches 2004), a modest approximation next to the
// two decades of real quotes the fund itself carries; see cmd/gen-euro-refdata.
func dbxgRecipe() Recipe {
	return Recipe{
		ID:              "DBXG",
		Name:            "Xtrackers Eurozone Govt 25+: long euro-gov bond",
		Method:          "long euro-area government bond TR (EUROGOV-LONG-EUR, ~30y constant maturity, modified duration ~19, ~1970 with the ECB daily 30y shape EUROGOV-LONG-DAILY from 2004) less 0.15%/yr TER; euro-native, real DBXG grafted from 2007",
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
		Method:          "long euro-area government bond TR (EUROGOV-LONG-EUR, ~30y constant maturity, ~1970, with the ECB daily 30y shape EUROGOV-LONG-DAILY from 2004) less 0.07%/yr TER; euro-native, real MTH grafted from 2017",
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
// classes carry ~2.15%/yr of fees inside the NAV: no fee adjustment.
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

// euroGovLongDaily builds the long (25+ segment, ~30-year, duration ~19) euro-area
// government bond total return at daily granularity: the bundled monthly
// EUROGOV-LONG-EUR anchors (~1970) carrying the ECB daily 30-year shape
// (EUROGOV-LONG-DAILY, 2004->) where it overlaps, the long-maturity twin of the
// EUROGOV-EUR / EUROGOV-DAILY pair the NTSZ bond leg uses.
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
func iefRecipe() Recipe {
	return Recipe{
		ID:              "IEF",
		Name:            "iShares 7-10Y Treasury: VFITX intermediate Treasury",
		Method:          "VFITX (Vanguard Intermediate-Term Treasury, 1991→), real IEF grafted from 2002",
		Build:           composite("IEF (intermediate Treasury)", []Leg{{ID: "VFITX", Weight: 1}}, "", 0),
		ValidateAgainst: "IEF",
		SpliceReal:      "IEF",
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
// hedge they are bought for. The gearing is what the overlap measures support
// (it halves the CAGR gap against the real DTLA, from 0.57 to 0.32 points a
// year over 2018-2026); the reported beta cannot arbitrate it, being a
// real-on-sim slope depressed by the US-close donor against a European listing.
// The older idtlRecipe holds the same bonds ungeared and is left alone rather
// than silently moved, so the distributing twin runs ~12% short on duration.
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
			"financed at USD cash ^IRX and re-earning EUR cash (EURCASH-EUR) = EUR-hedged US long Treasuries, less 0.10%/yr TER; total return, nothing grafted (the real DTLE series is a distributing NAV, price-only)",
		Build:           dtleBuild,
		ValidateAgainst: "DTLE",
	}
}

// dtleBuild hedges the USD 20+ Treasury segment into EUR: the local leg earns
// its excess over USD cash, the hedge gives back EUR cash, and the TER comes
// off. From 2018 the local leg is the REAL accumulating USD twin (DTLA), which
// holds the same bonds and prices on the same London close as DTLE itself, so
// the timing mismatch that a US mutual-fund donor carries disappears over the
// whole window the comparison is made on; before it, the geared VUSTX
// reconstruction stands, as it does inside DTLA's own history.
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

	const feeDaily = 0.0010 / 252
	s := &marketdata.Series{
		Name: "DTLE (EUR-hedged US long Treasury, total return)", Source: "simdata", Currency: "EUR",
	}
	v := 100.0
	s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[0], Close: v})
	for k := 1; k < len(fr.Dates); k++ {
		excess := longTreasuryGearing * (vustx[k] - irx[k])
		if r, ok := twin[fr.Dates[k]]; ok {
			excess = r - irx[k]
		}
		v *= 1 + excess + eur[k] - feeDaily
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

// dbmf/kmlm/cta reconstruct managed-futures trend from a generic 12-month
// TSMOM on a cross-asset futures basket. Measured daily correlation vs the
// real funds (self-generated, no official index): DBMF 0.52, KMLM 0.35, CTA
// 0.20; these funds run faster/idiosyncratic strategies a generic trend
// model only partly captures, but it is a faithful "diversified trend"
// proxy, and the real fund is grafted on top from its inception.
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

// dbiDonors are the real managed-futures NAVs spliced behind the DBi family,
// nearest trade first. Measured against DBMF on their common window (monthly
// correlation, the honest yardstick for a sleeve held for years): the Virtus
// AlphaSimplex fund 0.81 from 2010-08, the Guggenheim fund 0.72 from 2007-02,
// Man AHL Diversified 0.77 from 1996-03. The futures-price reconstruction
// manages 0.60 over the same window, so real quotes of another manager's
// programme beat it and are used wherever they exist; the reconstruction only
// covers what they cannot reach, which since AHL joined means before 1996-03.
//
// ahlDiversified is the deepest donor of the family and the one that made the
// sparse-cadence handling necessary: the fund dealt WEEKLY until 2016, so the
// whole segment it contributes (1996-03 to 2007-02) arrives one NAV a week and
// is projected onto the reconstruction's daily calendar (see DonorChain). Its
// record reproduces the manager's published one (+1649.5 % from 1996-03-26 to
// 2026-02-27 against a published +1649.47 %) and its crisis years are the real
// thing: 1998 +41 %, 2008 +33 %, 2022 +12 %.
const ahlDiversified = "IE0000360275"

var dbiDonors = []string{"ASFYX", "RYMFX", ahlDiversified}

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
//	               supplement effective 2025-08-25)
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
var trendFeeLoad = map[string]float64{
	"ASFYX":        0.0145,
	"RYMFX":        0.0199,
	"AHLPX":        0.0191,
	"AQMIX":        0.0129,
	ahlDiversified: 0.0274,
	"DBMF":         0.0085,
	"LU2951555585": 0.0075,
	"DBMFE":        0.0075,
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

func dbmfRecipe() Recipe {
	return Recipe{
		ID:   "DBMF",
		Name: "iMGP DBi Managed Futures: real managed-futures NAVs, then a TSMOM reconstruction",
		Method: "real NAVs of other managed-futures programmes spliced behind the fund, volatility-matched to it and lifted to its own 0.85%/yr fee load (Virtus AlphaSimplex 2010-08→ +0.60%/yr, Guggenheim 2007-02→ +1.14%/yr, Man AHL Diversified 1996-03→ +1.89%/yr, its weekly NAVs projected onto the reconstruction's daily calendar), " +
			"the file starting at that deepest donor's own first NAV, real DBMF grafted from 2019",
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
			"real NAVs of other managed-futures programmes behind it, fee-aligned the same way (2010-08 +0.70%/yr, 2007-02 +1.24%/yr, Man AHL Diversified 1996-03 +1.99%/yr), the file starting at that deepest donor's own first NAV, real DBMF.PA grafted from 2025",
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
		Method: "the US-listed DBMF itself from 2019, real NAVs of other managed-futures programmes behind it, each lifted to the UCITS class's 0.75%/yr fee load (2010-08, 2007-02, Man AHL Diversified 1996-03), the file starting at that deepest donor's own first NAV, " +
			"the whole converted USD→EUR at EURUSD spot (bundled ECU/DM/EUR proxy back to 1971), real DBMFE grafted from 2025",
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

// kmlmRecipe reconstructs KMLM from the same TSMOM engine at a higher vol
// target, real KMLM quotes grafted on top (see dbmf/kmlm/cta note above).
func kmlmRecipe() Recipe {
	return Recipe{
		ID:              "KMLM",
		Name:            "KraneShares KMLM: TSMOM replication",
		Method:          "real NAVs of other managed-futures programmes spliced behind the fund, each lifted to its 0.90%/yr fee load (Virtus AlphaSimplex 2010-08→ +0.55%/yr, Guggenheim 2007-02→ +1.09%/yr, Man AHL Diversified 1996-03→ +1.84%/yr), the file starting at that deepest donor's own first NAV; the 12-month TSMOM engine (14% vol target ~ the fund's realized 14.7%) supplies the daily texture the weekly-dealing donor is projected onto; real KMLM grafted from 2020",
		Build:           chainedTrend("KMLM (donor chain)", "KMLM", feeAligned("KMLM", []string{"ASFYX", "RYMFX", ahlDiversified}), mfConfig(0.14, 0.0090)),
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
		Build:           chainedTrend("AQR Managed Futures (donor chain)", "LU1103257975", feeAligned("LU1103257975", []string{"AQMIX", "RYMFX", ahlDiversified}), mfConfig(0.09, 0.0079)),
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
		ValidateAgainst: "LU1662501532",
		SpliceReal:      "LU1662501532",
	}
}

// aqrHedgedBuild builds the EUR-hedged AQR series: the TSMOM reconstruction
// hedged to EUR (r_eur = r_usd − usd_cash + eur_cash, accruals from the
// strategy's own frame), then the real B EUR sister class spliced over it,
// lifted by aqrBEURFeeWedge and truncated at its first long NAV gap (see
// aqrmfHedgedRecipe for the measured rationale of both). When the donor is
// unavailable (offline builds) the reconstruction stands alone.
//
// The truncation is load bearing beyond ending the segment where the class was
// emptied. The donor's own NAV history resumes on 2023-10-19 with one stale
// print at its pre-gap level (113.91) before the re-seeded class prints 99.70
// the next day: read as a return, that single artefact is a -12.5% day, and it
// alone drags the post-gap daily correlation with RAEF from 0.999 to 0.777.
// Nothing downstream needs those NAVs anyway, since real RAEF quotes cover
// 2021-04 onward, so the head of the series is the only part kept.
func aqrHedgedBuild(f Fetcher, from time.Time) (*marketdata.Series, error) {
	cfg := mfConfig(0.09, 0.0079)
	// The USD leg is the same donor chain the unhedged class uses: the
	// manager's own US fund from 2010, another manager's programme behind it,
	// and the reconstruction only for what neither reaches.
	// Fee-aligned as class A, because it IS the class A chain: only the pre-2015
	// leg of it survives here, and the B EUR donor spliced over that carries its
	// own measured wedge (aqrBEURFeeWedge) instead.
	usd, err := chainedTrend("AQR MF USD (donor chain)", "LU1103257975",
		feeAligned("LU1103257975", []string{"AQMIX", "RYMFX", ahlDiversified}), cfg)(f, from)
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
	donor = afterAnnualFee("AQR Managed Futures B EUR (real donor, performance fee added back)",
		donor, -aqrBEURFeeWedge)
	return Splice(donor, hedged), nil
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

// ctaRecipe reconstructs Simplify CTA from the same TSMOM engine, real CTA
// quotes grafted on top (see dbmf/kmlm/cta note above).
func ctaRecipe() Recipe {
	return Recipe{
		ID:              "CTA",
		Name:            "Simplify CTA: TSMOM replication",
		Method:          "real NAVs of other managed-futures programmes spliced behind the fund, each lifted to its 0.75%/yr fee load (Man AHL US 2014-08→ +1.16%/yr, Virtus AlphaSimplex 2010-08→ +0.70%/yr, Guggenheim 2007-02→ +1.24%/yr, Man AHL Diversified 1996-03→ +1.99%/yr), the file starting at that deepest donor's own first NAV; the 12-month TSMOM engine (16% vol target ~ the fund's realized 16.9%) supplies the daily texture the weekly-dealing donor is projected onto; real CTA grafted from 2022",
		Build:           chainedTrend("CTA (donor chain)", "CTA", feeAligned("CTA", []string{"AHLPX", "ASFYX", "RYMFX", ahlDiversified}), mfConfig(0.16, 0.0075)),
		ValidateAgainst: "CTA",
		SpliceReal:      "CTA",
	}
}

// backcastBuild wraps FitBackcast: the model is fitted on the asset's real
// history, then projected over the whole frame. Honest but limited: only
// the systematic exposures survive.
func backcastBuild(name, realID string, ids []string) func(Fetcher, time.Time) (*marketdata.Series, error) {
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		fr, err := BuildFrame(extend(f), ids, from)
		if err != nil {
			return nil, err
		}
		real, err := f.Fetch(realID, from)
		if err != nil {
			return nil, err
		}
		values, r2, _, err := FitBackcast(fr, real, ids)
		if err != nil {
			return nil, err
		}
		if r2 < minBackcastR2 {
			return nil, fmt.Errorf("%w: R² in-sample %.2f < %.2f", ErrUnfaithful, r2, minBackcastR2)
		}
		return SeriesFromFrame(fmt.Sprintf("%s (backcast R²=%.2f)", name, r2), fr, values), nil
	}
}

// amundiVolRecipe attempts a regression backcast of the Amundi Volatility
// World fund on VIX variations; volatility-trading funds are idiosyncratic,
// so this only ships when the in-sample fit clears minBackcastR2.
func amundiVolRecipe() Recipe {
	return Recipe{
		ID:              "LU0319687124",
		Name:            "Amundi Volatility World: backcast on ^VIX",
		Method:          "regression of the fund's daily returns on Δ^VIX and VFISX (2007→), replayed before 2007; residuals dropped",
		Build:           backcastBuild("Amundi Volatility World", "LU0319687124", []string{"^VIX", "VFISX"}),
		ValidateAgainst: "LU0319687124",
	}
}

// bhmgRecipe attempts the same exercise for BH Macro; discretionary macro
// rarely regresses well on asset-class factors, in which case nothing is
// written.
func bhmgRecipe() Recipe {
	return Recipe{
		ID:              "GG00BQBFY362",
		Name:            "BH Macro: factor backcast",
		Method:          "regression of daily returns on VUSTX, VFINX, GC=F (2007→), replayed before; residuals dropped",
		Build:           backcastBuild("BH Macro", "GG00BQBFY362", []string{"VUSTX", "VFINX", "GC=F"}),
		ValidateAgainst: "GG00BQBFY362",
	}
}
