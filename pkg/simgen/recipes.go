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
		wintonRecipe(),
		zrozRecipe(),
		dbmfRecipe(),
		dbmfpaRecipe(),
		dbmfeRecipe(),
		kmlmRecipe(),
		ctaRecipe(),
		amundiVolRecipe(),
		bhmgRecipe(),
		rssbRecipe(),
		gdeRecipe(),
		rsstRecipe(),
		rsbtRecipe(),
		vtRecipe(),
		xauusdRecipe(),
		shyRecipe(),
		scvwRecipe(),
		spxRecipe(),
		dpgtRecipe(),
		avantisRecipe(),
		chsnRecipe(),
		tip1eRecipe(),
		idtlRecipe(),
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
// full return, and the trend leg (KMLM, the deepest pure-trend backcast here,
// no equity of its own) is stacked as excess over cash, since a managed-futures
// program is run on collateral that already earns the T-bill rate. The floor is
// set by the trend leg. Real RSST quotes are grafted from inception; same
// currency (USD), no FX leg.
func rsstRecipe() Recipe {
	return Recipe{
		ID:              "RSST",
		Name:            "Return Stacked US Stocks & Managed Futures: 100/100 replication",
		Method:          "1.00×VFINX + 1.00×(deep TSMOM trend − cash ^IRX) overlay (~1989→), 0.96%/yr fees, real RSST grafted from 2023",
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
		Method:          "1.00×VFITX + 1.00×(deep TSMOM trend − cash ^IRX) overlay (~1989→), 0.97%/yr fees, real RSBT grafted from 2023",
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

// wintonRecipe rebuilds the Winton Trend-Enhanced Global Equity fund as
// global equities (60/40 US/international) plus a half-sized self-generated
// TSMOM trend overlay, net of 0.80%/yr fees.
func wintonRecipe() Recipe {
	return Recipe{
		ID:              "IE000O1VI174",
		Name:            "Winton Trend-Enhanced Global Equity: equities + TSMOM overlay",
		Method:          "0.60×VFINX + 0.40×VTMGX + 0.50×(TSMOM trend), 0.80%/yr fees (~2001→)",
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
	return s, nil
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
// a per-fund volatility target and fee.
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

// tsmom is the shared Build for trend-following reconstructions: it builds a
// frame on the markets and runs the TSMOM engine, returning the strategy
// index aligned to the dates after the signal warm-up.
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
		s := &marketdata.Series{Name: name, Source: "simdata"}
		for i, v := range values {
			s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[start+i], Close: v})
		}
		return s, nil
	}
}

// stackedTrend backcasts a Return Stacked fund: a funded core (coreID, an equity
// or bond index deep via refdata) plus a 100 % managed-futures overlay stacked
// on top. The overlay is the same deep TSMOM replication used for KMLM/DBMF/CTA
// (cfg), reconstructed from the underlying futures back to ~1989, which a plain
// composite leg (limited to the real fund's 2020s inception) cannot reach. The
// TSMOM index earns cash on its collateral (cfg.EarnCash), so its excess over
// cash is the pure trend overlay: r = coreReturn + (trendReturn − cash) − fee.
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
		core, cash := fr.Returns[coreID], fr.Returns[cfg.CashID]
		feeDaily := annualFee / 252
		values := make([]float64, len(trend))
		values[0] = 100
		for i := 1; i < len(trend); i++ {
			k := start + i
			trendRet := trend[i]/trend[i-1] - 1
			r := core[k] + (trendRet - cash[k]) - feeDaily
			values[i] = values[i-1] * (1 + r)
		}
		s := &marketdata.Series{Name: name, Source: "simdata"}
		for i, v := range values {
			s.Points = append(s.Points, marketdata.Point{Date: fr.Dates[start+i], Close: v})
		}
		return s, nil
	}
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
	return func(f Fetcher, from time.Time) (*marketdata.Series, error) {
		net, err := f.Fetch("MSCIWORLD-USD", from)
		if err != nil || net == nil || len(net.Points) <= 300 {
			return fallback(f, from)
		}
		out := net
		if shape, serr := f.Fetch(msciWorldShapeID, from); serr == nil && shape != nil && len(shape.Points) > 300 {
			out = shapedSeries(net, shape)
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

func dbmfRecipe() Recipe {
	return Recipe{
		ID:              "DBMF",
		Name:            "iMGP DBi Managed Futures: TSMOM replication",
		Method:          "12-month TSMOM on a cross-asset futures basket (gold→LBMA fix ~1968, crude→WTI spot ~1946, dev-ex-US→DEVEXUS-USD ~1969, EM→EM-USD ~1989, treasuries→CMT TR ~1953; start now set by the EM leg ~1989), real DBMF grafted from 2019",
		Build:           tsmom("DBMF (TSMOM replication)", mfConfig(0.10, 0.0085)),
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
		ID:              "LU2951555585",
		Name:            "iMGP DBi Managed Futures UCITS USD: TSMOM replication",
		Method:          "12-month TSMOM on a cross-asset futures basket (~2001→), real DBMF.PA grafted from 2025",
		Build:           tsmom("DBMF.PA (TSMOM replication)", mfConfig(0.10, 0.0075)),
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
		ID:              "DBMFE",
		Name:            "iMGP DBi Managed Futures EUR unhedged: TSMOM replication in EUR",
		Method:          "12-month TSMOM on a cross-asset futures basket, converted USD→EUR at EURUSD spot (bundled ECU/DM/EUR proxy back to 1971), real DBMFE grafted from 2025",
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
	cfg := mfConfig(0.10, 0.0085) // identical USD strategy to dbmfRecipe
	fr, err := BuildFrame(extend(f), append([]string{cfg.CashID}, cfg.Markets...), from)
	if err != nil {
		return nil, err
	}
	usd, start, err := TSMOM(fr, cfg)
	if err != nil {
		return nil, err
	}
	return convertDaily("DBMFE (USD TSMOM converted to unhedged EUR)",
		extend(f), "EURUSD=X", from, fr.Dates[start:], usd)
}

// kmlmRecipe reconstructs KMLM from the same TSMOM engine at a higher vol
// target, real KMLM quotes grafted on top (see dbmf/kmlm/cta note above).
func kmlmRecipe() Recipe {
	return Recipe{
		ID:              "KMLM",
		Name:            "KraneShares KMLM: TSMOM replication",
		Method:          "12-month TSMOM on a cross-asset futures basket (~2001→, higher vol target), real KMLM grafted from 2020",
		Build:           tsmom("KMLM (TSMOM replication)", mfConfig(0.13, 0.0090)),
		ValidateAgainst: "KMLM",
		SpliceReal:      "KMLM",
	}
}

// ctaRecipe reconstructs Simplify CTA from the same TSMOM engine, real CTA
// quotes grafted on top (see dbmf/kmlm/cta note above).
func ctaRecipe() Recipe {
	return Recipe{
		ID:              "CTA",
		Name:            "Simplify CTA: TSMOM replication",
		Method:          "12-month TSMOM on a cross-asset futures basket (~2001→, high vol target ~ the fund's ~18%), real CTA grafted from 2022",
		Build:           tsmom("CTA (TSMOM replication)", mfConfig(0.15, 0.0075)),
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
