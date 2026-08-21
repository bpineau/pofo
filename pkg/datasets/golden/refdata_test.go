package golden

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// These golden tests validate the bundled long backcast series
// (pkg/datasets/refdata/) themselves, complementing the SPY/URTH fixtures that
// validate the daily computations. The references were cross-checked against
// public sources on 2026-07-01 (refreshed 2026-07-18); each series is asserted
// over several year-to-year windows AND, for the two headline equity anchors,
// against published single-year calendar returns (see the *Yearly tests).
//
// The refdata series are MONTHLY and each point is dated on its month-END: the
// last trading day of the month for SP500-USD/MSCIWORLD-USD, so the level is the
// real month-end close and the daily shape (^GSPC / ^990100) can be pinned to it
// exactly (see simgen.alignMonthEnd). A published calendar-decade or calendar-
// year return is then a December-to-December compounding, and refWindow selects
// Dec(y0)..Dec(y1) inclusive.
//
// Only CAGR and MaxDrawdown are asserted on the monthly windows: metrics.Compute
// annualizes volatility with sqrt(252) (daily), which is meaningless on monthly
// points, so the volatility/Sharpe/Sortino conventions stay validated by the
// daily fixtures. Where a vol sanity band is useful it is computed as
// sigma_m*sqrt(12) directly.

// refFinding documents two things this validation surfaced.
//
// (1) MSCIWORLD-USD is the MSCI World NET total-return index, not gross as its
// CSV header once claimed: its Dec2012->Dec2024 CAGR is 10.82%/yr and
// Dec2014->Dec2024 is 9.95%/yr, matching MSCI World NET USD exactly (gross was
// 11.41%/yr and 10.52%/yr, the withholding drag). That is the right proxy for an
// Irish-domiciled UCITS World ETF (IWDA/URTH track the net index), and the recipe
// only deducts the TER on top.
//
// (2) Both anchors were month-END values but dated the FIRST of the month, which
// anchorShape (pinning to the first shape date on or after the anchor date) then
// slid ~1 month, understating e.g. the 2022 drawdown by ten points in the shaped
// reconstruction. Fixed 2026-07-18: MSCIWORLD-USD was relabeled to month-end
// dates (values unchanged, still the Curvo net-TR export) and SP500-USD was
// rebuilt month-end from point-in-time sources (^SP500TR 1988->, ^GSPC + Shiller
// dividend 1928-1988, Shiller 1871-1928) by cmd/gen-sp500-refdata, replacing the
// Shiller monthly AVERAGE that smeared every turning point (1987 came out -0.1%
// against the real +5.3%). simgen.alignMonthEnd now snaps each anchor onto the
// shape's own last trading day so the fix cannot silently regress.
//
// (3) The two remaining Curvo exports, DEVEXUS-USD and EM-USD, carried the same
// first-of-month labels and were relabeled the same way on 2026-08-20; the snap
// was extended to the splice path in simgen/extend.go, which is where those two
// are consumed (as the long proxies behind VTMGX and VEIEX) and which shaped
// without aligning. Unfixed, the developed-ex-US leg rebuilt 2022 at -4.3%
// against the index's -14.3%. Only anchors listed in simgen.monthEndAnchor are
// snapped: the Treasury, euro-govt and WTI references run on monthly AVERAGE
// observations dated the first of the month, and moving those would slide them
// the other way.

func loadRefdata(t *testing.T, id string) *marketdata.Series {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil || !ok {
		t.Fatalf("refdata %s: ok=%v err=%v", id, ok, err)
	}
	return s
}

// calYearDaily is the December-to-December total return (%) of calendar year y
// from a DAILY series: the last close on or before 31 Dec of y over that of y-1.
func calYearDaily(s *marketdata.Series, y int) float64 {
	dec := func(yr int) float64 {
		cut := time.Date(yr, 12, 31, 23, 0, 0, 0, time.UTC)
		var v float64
		for _, p := range s.Points {
			if p.Date.After(cut) {
				break
			}
			v = p.Close
		}
		return v
	}
	a, b := dec(y-1), dec(y)
	if a == 0 {
		return 0
	}
	return (b/a - 1) * 100
}

// refWindow returns the monthly points from the December of y0 (inclusive) to
// the December of y1 (inclusive), i.e. a (y1-y0)-year December-to-December span.
func refWindow(t *testing.T, s *marketdata.Series, y0, y1 int) ([]time.Time, []float64) {
	t.Helper()
	var dates []time.Time
	var values []float64
	for _, p := range s.Points {
		afterStart := p.Date.Year() > y0 || (p.Date.Year() == y0 && p.Date.Month() == 12)
		beforeEnd := p.Date.Year() <= y1
		if afterStart && beforeEnd {
			dates = append(dates, p.Date)
			values = append(values, p.Close)
		}
	}
	if len(values) < 12 {
		t.Fatalf("refWindow %d..%d: only %d points", y0, y1, len(values))
	}
	return dates, values
}

// monthlyVol is the annualized volatility of monthly returns (sigma_m*sqrt(12)),
// the meaningful figure for these monthly series.
func monthlyVol(values []float64) float64 {
	r := metrics.Returns(values)
	if len(r) < 2 {
		return math.NaN()
	}
	m := metrics.Mean(r)
	var v float64
	for _, x := range r {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v/float64(len(r)-1)) * math.Sqrt(12)
}

type refCase struct {
	name       string
	y0, y1     int
	cagr, ctol float64 // reference CAGR (%) and tolerance
	minDD      float64 // MaxDrawdown must be at least this deep (more negative), 0 to skip
	volLo      float64 // monthly-annualized vol lower/upper sanity band (%), 0 to skip
	volHi      float64
}

func runRefCases(t *testing.T, s *marketdata.Series, cases []refCase) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dates, values := refWindow(t, s, c.y0, c.y1)
			st, err := metrics.Compute(dates, values)
			if err != nil {
				t.Fatal(err)
			}
			within(t, "CAGR", st.CAGR*100, c.cagr, c.ctol)
			if c.minDD != 0 && st.MaxDrawdown*100 > c.minDD {
				t.Errorf("MaxDrawdown = %.1f %%, expected at least %.1f %% deep", st.MaxDrawdown*100, c.minDD)
			}
			if c.volHi != 0 {
				if v := monthlyVol(values) * 100; v < c.volLo || v > c.volHi {
					t.Errorf("monthly vol = %.1f %%, expected within [%.1f, %.1f]", v, c.volLo, c.volHi)
				}
			}
		})
	}
}

// TestGoldenSP500 validates SP500-USD (month-end S&P 500 total return; see
// cmd/gen-sp500-refdata) against the widely published S&P 500 TR decade returns
// (nominal, dividends reinvested), e.g. S&P Dow Jones / dqydj: 1970s 5.9 %,
// 1980s 17.6 %, 1990s 18.2 %, 2000s -0.9 %, 2010s 13.6 %; the reconstruction
// sits within ~0.35 point of these. The month-end rebuild carries a realistic
// monthly volatility (~15 %/yr annualized), where the earlier Shiller monthly
// AVERAGE smoothed it below 15.
func TestGoldenSP500(t *testing.T) {
	s := loadRefdata(t, "SP500-USD")
	runRefCases(t, s, []refCase{
		{name: "1970s", y0: 1969, y1: 1979, cagr: 5.9, ctol: 0.5, volLo: 11, volHi: 17},
		{name: "1980s", y0: 1979, y1: 1989, cagr: 17.6, ctol: 0.6},
		{name: "1990s", y0: 1989, y1: 1999, cagr: 18.2, ctol: 0.5},
		{name: "2000s (lost decade)", y0: 1999, y1: 2009, cagr: -0.9, ctol: 0.5, minDD: -46},
		{name: "2010s", y0: 2009, y1: 2019, cagr: 13.6, ctol: 0.5},
		// Long run: since 1928 the S&P TR is ~10 %/yr nominal; this shorter,
		// stronger 1971-2024 span runs a touch above at ~11 %.
		{name: "1971-2024", y0: 1971, y1: 2024, cagr: 11.0, ctol: 0.5, volLo: 13, volHi: 17},
	})
}

// TestGoldenSP500Yearly pins the SHAPED S&P 500 reconstruction (the embedded
// SP500 simdata: SP500-USD levels carried on the ^GSPC daily shape) to published
// S&P 500 total-return CALENDAR-year returns, the check that actually exercises
// the month-end fix end to end. Since 1988 the anchor is the ^SP500TR index
// itself, so the fit is exact; before it the ^GSPC+dividend rebuild holds to a
// few tenths of a point (1933, a chaotic ~54 % year, is the one looser point).
// Published S&P 500 TR (slickcharts / S&P Dow Jones).
func TestGoldenSP500Yearly(t *testing.T) {
	s := loadSimdata(t, "SP500")
	for _, c := range []struct {
		year     int
		ref, tol float64
	}{
		{1954, 52.6, 1.0}, {1958, 43.4, 0.6}, {1974, -26.5, 0.6}, {1980, 32.4, 0.6},
		{1987, 5.3, 0.6}, {1995, 37.6, 0.4}, {2000, -9.1, 0.4}, {2002, -22.1, 0.4},
		{2008, -37.0, 0.4}, {2013, 32.4, 0.4}, {2018, -4.4, 0.4}, {2019, 31.5, 0.4},
		{2020, 18.4, 0.4}, {2021, 28.7, 0.4}, {2022, -18.1, 0.4}, {2023, 26.3, 0.4},
		{2024, 25.0, 0.4},
	} {
		within(t, "SP500 "+strconv.Itoa(c.year), calYearDaily(s, c.year), c.ref, c.tol)
	}
}

// TestGoldenMSCIWorldYearly pins the SHAPED MSCI World reconstruction (the
// embedded MSCIWORLD simdata: MSCIWORLD-USD net-TR levels on the ^990100 daily
// shape) to published MSCI World NET USD calendar-year returns (MSCI index
// factsheets). The net index is what the Curvo anchor holds, so the Dec-to-Dec
// figures reproduce it to a fraction of a point; this is the test that would
// have caught the pre-fix one-month slip (2022 came out near -8 % instead of
// -18 %).
func TestGoldenMSCIWorldYearly(t *testing.T) {
	s := loadSimdata(t, "MSCIWORLD")
	for _, c := range []struct {
		year     int
		ref, tol float64
	}{
		{2008, -40.7, 0.6}, {2013, 26.7, 0.5}, {2015, -0.9, 0.5}, {2017, 22.4, 0.5},
		{2018, -8.7, 0.5}, {2019, 27.7, 0.5}, {2020, 15.9, 0.5}, {2021, 21.8, 0.5},
		{2022, -18.1, 0.5}, {2023, 23.8, 0.5}, {2024, 18.7, 0.5},
	} {
		within(t, "MSCIWORLD "+strconv.Itoa(c.year), calYearDaily(s, c.year), c.ref, c.tol)
	}
}

// TestGoldenMSCIWorld validates MSCIWORLD-USD as the MSCI World NET total-return
// index (see refFinding): 2013-2024 net 10.82 %/yr and 2015-2024 net 9.95 %/yr
// (gross was 11.41 % and 10.52 %); the 2000s net was ~-0.2 %/yr.
func TestGoldenMSCIWorld(t *testing.T) {
	s := loadRefdata(t, "MSCIWORLD-USD")
	runRefCases(t, s, []refCase{
		{name: "2013-2024 (net)", y0: 2012, y1: 2024, cagr: 10.82, ctol: 0.4, volLo: 12, volHi: 17},
		{name: "2015-2024 (net)", y0: 2014, y1: 2024, cagr: 9.95, ctol: 0.4},
		{name: "2000s (net)", y0: 1999, y1: 2009, cagr: -0.2, ctol: 0.6, minDD: -50},
		{name: "1971-2024 (net)", y0: 1971, y1: 2024, cagr: 9.1, ctol: 0.5, volLo: 13, volHi: 17},
	})
}

// TestGoldenDevExUSA validates DEVEXUS-USD as the MSCI World ex USA NET
// total-return index, the developed-ex-US leg behind VT/VWRA and the intl half
// of every World reconstruction. The reference CAGRs are compounded from the
// published net calendar years below, so the two checks share one source.
func TestGoldenDevExUSA(t *testing.T) {
	s := loadRefdata(t, "DEVEXUS-USD")
	runRefCases(t, s, []refCase{
		{name: "2013-2024 (net)", y0: 2012, y1: 2024, cagr: 5.65, ctol: 0.3, volLo: 12, volHi: 18},
		{name: "2013-2025 (net)", y0: 2012, y1: 2025, cagr: 7.46, ctol: 0.3},
		// No published figure spans the whole file, so the long window is a
		// plausibility band rather than a pin: a developed-ex-US equity index
		// compounding near 9 %/yr at ~17 % volatility, and a global financial
		// crisis at least 50 points deep (MSCI publishes 60.1 % for the gross
		// index on daily data, 2007-10-31 to 2009-03-09).
		{name: "since 1970 (band)", y0: 1969, y1: 2025, cagr: 8.8, ctol: 1.0, minDD: -50, volLo: 13, volHi: 19},
	})
}

// TestGoldenDevExUSAYearly pins DEVEXUS-USD to published MSCI World ex USA NET
// USD calendar-year returns (MSCI index factsheet msci-world-ex-usa-index-net.pdf,
// Jul 31 2026 edition, msci.com/documents/10199/255599/). The anchor reproduces
// every one of them to the basis point, which is what identifies it as that
// index and, more to the point here, what proves its levels are month-END
// values: a one-month slip would move 2022 by ten points (it did, before the
// 2026-08-20 relabeling).
//
// The anchor is monthly, so this test validates the LEVELS and their labels.
// That the daily SHAPING preserves them is checked where the shaping code lives,
// by simgen.TestAlignMonthEndPreservesCalendarYears, which recomputes the blend
// from these same bundled files.
func TestGoldenDevExUSAYearly(t *testing.T) {
	s := loadRefdata(t, "DEVEXUS-USD")
	for _, c := range []struct {
		year int
		ref  float64
	}{
		{2012, 16.41}, {2013, 21.02}, {2014, -4.32}, {2015, -3.04}, {2016, 2.75},
		{2017, 24.21}, {2018, -14.09}, {2019, 22.49}, {2020, 7.59}, {2021, 12.62},
		{2022, -14.29}, {2023, 17.94}, {2024, 4.70}, {2025, 31.85},
	} {
		within(t, "DEVEXUS-USD "+strconv.Itoa(c.year), calYearDaily(s, c.year), c.ref, 0.05)
	}
}

// TestGoldenEM validates EM-USD, the emerging-market leg behind VEIEX/EIMI/VWRA,
// against the published MSCI Emerging Markets NET calendar years compounded into
// CAGRs. It is a Curvo export rather than the index itself and runs +0.31
// point/yr rich on average over 2012-2025 (see TestGoldenEMYearly), so the
// tolerances are a shade wider than the developed-ex-US ones.
func TestGoldenEM(t *testing.T) {
	s := loadRefdata(t, "EM-USD")
	runRefCases(t, s, []refCase{
		{name: "2013-2024 (net)", y0: 2012, y1: 2024, cagr: 2.61, ctol: 0.5, volLo: 14, volHi: 20},
		{name: "2013-2025 (net)", y0: 2012, y1: 2025, cagr: 4.71, ctol: 0.5},
		// Plausibility band over the whole file, as for DEVEXUS-USD: emerging
		// equity near 10 %/yr since 1987 at ~22 % volatility, with a 2008 drop
		// of at least 55 points.
		{name: "since 1988 (band)", y0: 1987, y1: 2025, cagr: 10.6, ctol: 1.0, minDD: -55, volLo: 18, volHi: 26},
	})
}

// TestGoldenEMYearly pins EM-USD to published MSCI Emerging Markets NET USD
// calendar-year returns (MSCI index factsheet msci-emerging-markets-index-usd-net.pdf,
// Jul 31 2026 edition). The fit is close but not exact: the export runs +0.31
// point/yr rich on average (max +0.47 in 2017, -0.38 in 2025, whose months are
// chained from a tracking series rather than exported), a LEVEL divergence that
// has been there since the file was adopted and is not what this test is for.
// What it catches is a date slip, which moves a calendar year by whole points,
// so the tolerance is set at 0.6 point: loose enough for the known bias, an
// order of magnitude tighter than a one-month slide.
func TestGoldenEMYearly(t *testing.T) {
	s := loadRefdata(t, "EM-USD")
	for _, c := range []struct {
		year int
		ref  float64
	}{
		{2012, 18.22}, {2013, -2.60}, {2014, -2.19}, {2015, -14.92}, {2016, 11.19},
		{2017, 37.28}, {2018, -14.57}, {2019, 18.42}, {2020, 18.31}, {2021, -2.54},
		{2022, -20.09}, {2023, 9.83}, {2024, 7.50}, {2025, 33.57},
	} {
		within(t, "EM-USD "+strconv.Itoa(c.year), calYearDaily(s, c.year), c.ref, 0.6)
	}
}

// TestGoldenGold validates XAUUSD-LBMA (London/LBMA PM fix, daily since
// 1968-04) against published nominal gold returns. The volatile 1979/1980
// boundary is skipped (year-end fixes there are extreme prints); the modern
// and long windows are validated. Gold nominal: 2000s ~14-15 %/yr, since
// 1971 ~8 %/yr.
func TestGoldenGold(t *testing.T) {
	s := loadRefdata(t, "XAUUSD-LBMA")
	runRefCases(t, s, []refCase{
		{name: "2000s", y0: 1999, y1: 2009, cagr: 14.9, ctol: 1.0},
		{name: "2000-2020", y0: 1999, y1: 2020, cagr: 9.4, ctol: 1.0},
		{name: "1971-2024", y0: 1971, y1: 2024, cagr: 8.0, ctol: 1.0, minDD: -55},
	})
}

// yearRet is the December-to-December total return (%) of calendar year y,
// for the monthly first-of-month refdata convention (see refWindow).
func yearRet(t *testing.T, s *marketdata.Series, y int) float64 {
	t.Helper()
	_, values := refWindow(t, s, y-1, y)
	return (values[len(values)-1]/values[0] - 1) * 100
}

// TestGoldenTreasuries validates the constant-maturity Treasury total-return
// reconstructions (TREASURY-INT-USD from GS5, TREASURY-LONG-USD from GS20,
// both via simgen.TreasuryTR) against the published Ibbotson SBBI yearly
// returns for intermediate- and long-term government bonds: 1969 (IT -0.7 %,
// LT -5.1 %), 1982 (IT +29.1 %, LT +40.4 %), 1994 (IT -5.1 %, LT -7.8 %) and
// 1995 (IT +16.8 %, LT +31.7 %). A 5-year (resp. 20-year) constant-maturity
// par-bond reconstruction is not the SBBI portfolio, so a couple of points of
// tolerance is expected, but the fit is tight enough to catch any unit,
// day-count or repricing regression.
func TestGoldenTreasuries(t *testing.T) {
	ti := loadRefdata(t, "TREASURY-INT-USD")
	tl := loadRefdata(t, "TREASURY-LONG-USD")
	for _, c := range []struct {
		year         int
		intRef, ltol float64
		longRef      float64
		itol         float64
	}{
		{year: 1969, intRef: -0.7, itol: 1.5, longRef: -5.1, ltol: 2.0},
		{year: 1982, intRef: 29.1, itol: 2.5, longRef: 40.4, ltol: 4.0},
		{year: 1994, intRef: -5.1, itol: 1.5, longRef: -7.8, ltol: 2.0},
		{year: 1995, intRef: 16.8, itol: 1.5, longRef: 31.7, ltol: 2.5},
	} {
		within(t, "INT "+strconv.Itoa(c.year), yearRet(t, ti, c.year), c.intRef, c.itol)
		within(t, "LONG "+strconv.Itoa(c.year), yearRet(t, tl, c.year), c.longRef, c.ltol)
	}
	// Long-run sanity: intermediate treasuries ~6 %/yr and long treasuries
	// ~7 %/yr over 1972-2021 (SBBI-era figures).
	runRefCases(t, ti, []refCase{{name: "1972-2021", y0: 1971, y1: 2021, cagr: 6.4, ctol: 0.8, volLo: 3, volHi: 7}})
	runRefCases(t, tl, []refCase{{name: "1972-2021", y0: 1971, y1: 2021, cagr: 7.6, ctol: 1.0, volLo: 9, volHi: 14}})
}

// TestGoldenTreasuryDailyShapes validates the DAILY shape series behind the
// monthly Treasury refdata (TREASURY-INT-DAILY from DGS5, TREASURY-LONG-DAILY
// from DGS20). Their levels are never authoritative (anchorShape re-anchors
// them monthly), so the checks are looser: calendar-year returns near the
// SBBI anchors, daily density, and an annualized daily volatility in the
// historically documented band for the 1980-1985 rate shock (long treasuries
// realized ~13-15 % then).
func TestGoldenTreasuryDailyShapes(t *testing.T) {
	for _, c := range []struct {
		id           string
		y1969, y1982 float64
		tol          float64
		volLo, volHi float64
	}{
		{id: "TREASURY-INT-DAILY", y1969: -0.7, y1982: 29.1, tol: 4.0, volLo: 5, volHi: 10},
		{id: "TREASURY-LONG-DAILY", y1969: -5.1, y1982: 40.4, tol: 6.0, volLo: 11, volHi: 17},
	} {
		s := loadRefdata(t, c.id)
		last := func(cut time.Time) float64 {
			v := math.NaN()
			for _, p := range s.Points {
				if p.Date.After(cut) {
					break
				}
				v = p.Close
			}
			return v
		}
		dyear := func(y int) float64 {
			a := last(time.Date(y-1, 12, 31, 0, 0, 0, 0, time.UTC))
			b := last(time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC))
			return (b/a - 1) * 100
		}
		within(t, c.id+" 1969", dyear(1969), c.y1969, c.tol)
		within(t, c.id+" 1982", dyear(1982), c.y1982, c.tol)

		// Daily density: ~250 points per year, every year covered.
		days := 0
		for _, p := range s.Points {
			if p.Date.Year() == 1975 {
				days++
			}
		}
		if days < 230 {
			t.Errorf("%s: 1975 carries %d points, want daily density", c.id, days)
		}
		// Annualized daily vol over the 1980-1985 rate shock.
		var lr []float64
		var prev float64
		for _, p := range s.Points {
			if p.Date.Year() >= 1980 && p.Date.Year() <= 1985 {
				if prev > 0 {
					lr = append(lr, math.Log(p.Close/prev))
				}
				prev = p.Close
			}
		}
		m := 0.0
		for _, x := range lr {
			m += x
		}
		m /= float64(len(lr))
		v := 0.0
		for _, x := range lr {
			v += (x - m) * (x - m)
		}
		vol := math.Sqrt(v/float64(len(lr)-1)) * math.Sqrt(252) * 100
		if vol < c.volLo || vol > c.volHi {
			t.Errorf("%s: 1980-1985 daily vol = %.1f %%, expected within [%.1f, %.1f]", c.id, vol, c.volLo, c.volHi)
		}
	}
}

// TestGoldenWTIRolled validates WTI-ER-USD, the rolled-futures EXCESS return of
// WTI crude (cmd/gen-wti-refdata), against the published S&P GSCI Crude Oil
// TOTAL RETURN index, whose year-end levels are a matter of SEC record (Barclays
// Bank PLC, iPath S&P GSCI Crude Oil ETN pricing supplement, form 424B2 filed
// 2016-11-17, accession 0001104659-16-157831).
//
// The published series is FUNDED and the bundled one is not, so the test
// compounds WTI-ER-USD with the bundled 3-month T-bill before comparing. It then
// checks the thing that made the series worth bundling at all: the roll yield,
// the gap between a rolled position and the spot price the repo already had,
// which is large and changes sign by era.
//
// The reconstruction is not the index and is not claimed to be: 27 of the 29
// comparable years land within five points, and the tolerances below are the
// measured divergences. 1990 is the one wide year (+16.5 points) and is left out
// of the year list on purpose: the Gulf War pushed the front-to-second spread
// into double digits a month, which levers any roll-timing difference beyond
// what a spot check can arbitrate.
func TestGoldenWTIRolled(t *testing.T) {
	s := loadRefdata(t, "WTI-ER-USD")
	if got := s.Points[0].Date.Format("2006-01-02"); got != "1985-01-02" {
		t.Errorf("history starts %s, want 1985-01-02", got)
	}
	// EIA discontinued its NYMEX futures price series after this day; the
	// file stops where its evidence stops.
	if got := s.Points[len(s.Points)-1].Date.Format("2006-01-02"); got != "2024-04-05" {
		t.Errorf("history ends %s, want 2024-04-05", got)
	}

	// The tolerances are absolute points of calendar-year return, so a big
	// year buys a wide one for the same relative agreement: 1999 diverges by
	// 2.2 % relative, which on a +123 % year is five points.
	funded := fundWithTBill(t, s)
	for _, c := range []struct {
		year     int
		ref, tol float64
	}{
		{1989, 94.27, 1.0}, {1992, 3.65, 0.8}, {1996, 108.31, 2.5}, {1999, 122.69, 6.5},
		{2001, -25.44, 0.8}, {2003, 27.47, 0.8}, {2007, 47.45, 0.8}, {2008, -55.47, 2.5},
		{2009, 7.15, 3.0}, {2010, -0.11, 0.8}, {2012, -11.52, 0.8}, {2014, -42.56, 1.0},
		{2015, -45.34, 0.8},
	} {
		within(t, "WTI rolled, funded, "+strconv.Itoa(c.year), calYearDaily(funded, c.year), c.ref, c.tol)
	}

	// The roll yield itself. Crude was backwardated through the 1990s and in
	// deep contango over 2006-2016, so a rolled position must beat spot by a
	// wide margin in the first era and lose to it by a wider one in the
	// second. Getting this sign wrong is the whole error the series exists to
	// prevent, so the bands are loose and the signs are not negotiable.
	spot := loadRefdata(t, "WTI-USD")
	for _, c := range []struct {
		name         string
		y0, y1       int
		gapLo, gapHi float64 // rolled CAGR minus spot CAGR, in points per year
	}{
		{name: "backwardation era 1986-2000", y0: 1986, y1: 2000, gapLo: 5, gapHi: 15},
		{name: "contango era 2006-2016", y0: 2006, y1: 2016, gapLo: -18, gapHi: -8},
	} {
		t.Run(c.name, func(t *testing.T) {
			gap := (annualizedOver(t, s, c.y0, c.y1) - annualizedOver(t, spot, c.y0, c.y1)) * 100
			if gap < c.gapLo || gap > c.gapHi {
				t.Errorf("roll yield = %+.1f points/yr, expected within [%+.1f, %+.1f]", gap, c.gapLo, c.gapHi)
			}
		})
	}
}

// fundWithTBill compounds an excess-return series with the bundled 3-month
// T-bill, the way the published S&P GSCI total-return indices are funded: the
// rate is a DISCOUNT rate in percent on a 91-day bill and it accrues on every
// calendar day, weekends included.
func fundWithTBill(t *testing.T, s *marketdata.Series) *marketdata.Series {
	t.Helper()
	bill := loadRefdata(t, "TBILL-3M")
	byMonth := make(map[string]float64, len(bill.Points))
	var latest float64
	for _, p := range bill.Points {
		byMonth[p.Date.Format("2006-01")] = p.Close
		latest = p.Close
	}
	daily := func(d time.Time) float64 {
		r, ok := byMonth[d.Format("2006-01")]
		if !ok {
			r = latest
		}
		return math.Pow(1/(1-91.0/360.0*r/100), 1.0/91.0) - 1
	}
	out := &marketdata.Series{Points: make([]marketdata.Point, len(s.Points))}
	level := s.Points[0].Close
	out.Points[0] = marketdata.Point{Date: s.Points[0].Date, Close: level}
	for i := 1; i < len(s.Points); i++ {
		level *= s.Points[i].Close / s.Points[i-1].Close
		for d := s.Points[i-1].Date.AddDate(0, 0, 1); !d.After(s.Points[i].Date); d = d.AddDate(0, 0, 1) {
			level *= 1 + daily(d)
		}
		out.Points[i] = marketdata.Point{Date: s.Points[i].Date, Close: level}
	}
	return out
}

// annualizedOver is the compound annual rate between the last observation of
// year y0-1 and the last of y1, so it works on daily and monthly series alike.
func annualizedOver(t *testing.T, s *marketdata.Series, y0, y1 int) float64 {
	t.Helper()
	pick := func(year int) (time.Time, float64) {
		cut := time.Date(year, 12, 31, 23, 0, 0, 0, time.UTC)
		var d time.Time
		var v float64
		for _, p := range s.Points {
			if p.Date.After(cut) {
				break
			}
			d, v = p.Date, p.Close
		}
		return d, v
	}
	d0, v0 := pick(y0 - 1)
	d1, v1 := pick(y1)
	if v0 <= 0 || v1 <= 0 || !d1.After(d0) {
		t.Fatalf("no %d..%d window in the series", y0, y1)
	}
	years := d1.Sub(d0).Hours() / 24 / 365.25
	return math.Pow(v1/v0, 1/years) - 1
}
