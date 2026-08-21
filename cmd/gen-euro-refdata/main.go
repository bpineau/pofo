// Command gen-euro-refdata builds the bundled euro-area reference series that
// let the eurozone recipes (NTSZ, the WisdomTree Eurozone Efficient Core) reach
// back before their fetchable components. It runs at data-generation time only
// (network); the pofo binary embeds the CSVs and never fetches OECD or the ECB.
//
// Everything is sourced from DBnomics (free, key-less), the same mirror the
// macro panel uses. The OECD series come from the CURRENT short-term-statistics
// dataflow (OECD/DSD_STES@DF_FINMARK) rather than the legacy MEI dataset, which
// stopped being updated in 2024-01, and from its EA20 (euro area, 20 countries
// since Croatia joined in 2023) aggregate rather than EA19: over their whole
// common history the two aggregates agree to within 1.5e-4 percentage point on
// the yield and to the last published digit on the share-price index, and EA20
// is the one that keeps being revised and extended. Six series are written into
// pkg/datasets/refdata/:
//
//   - EMU-EUR.csv       eurozone equity net total return (monthly, ~1986):
//     the OECD euro-area share-price index (EA20.M.SHARE, a
//     price index) grossed to a net total return by a constant
//     net dividend yield (netDivYield), calibrated on the
//     overlap with the real MSCI Eurozone (EZU) in EUR. Proxy
//     behind EZU in the equity leg.
//   - EUROGOV-EUR.csv   euro-area government bond total return (monthly, ~1970):
//     the OECD euro-area long-term government bond yield
//     (EA20.M.IRLT) run through the constant-maturity
//     reconstruction simgen.TreasuryTR. Proxy behind the real
//     euro-govt bond ETF in the bond leg.
//   - EUROGOV-DAILY.csv euro-area government bond TR at daily granularity
//     (~2004): the ECB daily 10-year euro-area yield curve
//     point run through the same TreasuryTR. Daily shape for
//     EUROGOV-EUR, so the reconstruction stops feeding
//     month-sized moves to daily statistics after 2004.
//   - EUROGOV-LONG-EUR.csv   long euro-area government bond TR (25+ segment,
//     modified duration ~17; monthly, ~1970): the OECD 10-year
//     yield mapped to a 25-year yield (calibrated on the
//     2004-2026 ECB curve) run through TreasuryTR at a 24-year
//     par maturity (vol-matched to DBXG). Proxy behind the euro
//     25+ govt ETF (DBXG).
//   - EUROGOV-LONG-DAILY.csv long euro-area government bond TR at daily
//     granularity (~2004): the ECB daily 25-year euro-area yield
//     curve point run through the same TreasuryTR. Daily shape
//     for EUROGOV-LONG-EUR.
//   - DECASH-EUR.csv    German 3-month money-market accrual (monthly, ~1960):
//     the pre-euro cash proxy (Germany was the anchor economy
//     and the DM the reference currency), spliced under the
//     euro money-market index EURCASH-EUR at 1994 to carry the
//     cash leg back before the euro.
//
// Every series is checked before it is written (-check, on by default): its
// shape where an outside answer is known, its freshness (a dataflow that quietly
// freezes, as MEI did, must fail loudly rather than ship a stale tail), and the
// two reconstructions that can be graded against something the repository
// already trusts. Nothing here is trusted on the strength of having downloaded
// cleanly.
//
// Usage: gen-euro-refdata [-base URL] [-dir path] [-dry] [-check=false]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/simgen"
)

const defaultBase = "https://api.db.nomics.world/v22"

// netDivYield is the constant net dividend (and universe-drift) yield added to
// the euro-area share-price index to turn its price return into a net total
// return. It is calibrated so the proxy's total return matches the real MSCI
// Eurozone (EZU, net TR) expressed in EUR over their 2001-2023 overlap: EZU's
// EUR CAGR there is ~3.05%/yr against ~0.84%/yr for the price index, a ~2.2%/yr
// gap. A constant modestly understates the richer pre-2000 dividend yield, so
// the deep tail is, if anything, conservative on the equity leg. The proxy's
// level is rescaled where EZU takes over, so only its pre-2000 return drift
// matters; this pins that drift to the real overlap the way TreasuryTR's small
// biases are absorbed at its splice.
const netDivYield = 0.022

// euroBondMaturity is the constant maturity (years) of the reconstructed euro
// government bond, matching the ~10-year benchmark the OECD and ECB yields
// track and the intermediate-to-long duration of the fund's bond-futures ladder.
const euroBondMaturity = 10.0

// euroLongMaturity is the constant maturity (years) of the reconstructed LONG
// euro government bond (EUROGOV-LONG-*), a proxy for the 25+ segment of the euro
// sovereign curve behind DBXG. It is driven by the ECB 25-year yield-curve point
// and priced at a slightly shorter 24-year par maturity (modified duration ~17),
// calibrated so the daily reconstruction matches DBXG's ~14.4%/yr realized
// volatility over 2007-2026: the fitted long curve is a touch more volatile than
// the traded fund per year of maturity, so pricing at exactly the 25-year point
// would overstate the risk. The deep monthly tail feeds the same maturity a
// 25-year yield synthesized from the 10-year (below).
const euroLongMaturity = 24.0

// The deep monthly tail (~1970-2004) predates the ECB yield curve, so its long
// yield is synthesized from the OECD long-term (10-year) benchmark as
// euroLongIntercept + euroLongSlope*y10. Both are calibrated on the 2004-2026
// ECB curve overlap, where the 25-year point regresses on the 10-year as
// 25y = 0.571 + 0.962*10y: a ~0.5%/yr term premium and a slightly damped
// (~0.96x) sensitivity, so the deep long bond carries the long end's own level
// and volatility rather than the raw 10-year path. The real ECB 25-year yield
// takes over from 2004 through the daily series.
const euroLongIntercept = 0.571
const euroLongSlope = 0.9615

func main() {
	base := flag.String("base", defaultBase, "DBnomics API base URL")
	dir := flag.String("dir", "pkg/datasets/refdata", "output refdata directory")
	dry := flag.Bool("dry", false, "print coverage and checks without writing")
	check := flag.Bool("check", true, "run the sanity checks before writing")
	flag.Parse()

	// Euro-area government bond TR, monthly (~1970) and daily (~2004).
	govYield := fetch(*base, "OECD/DSD_STES@DF_FINMARK/EA20.M.IRLT.PA._Z._Z._Z._Z.N")
	govMonthly := simgen.TreasuryTR("Euro-area government bond total return (10y benchmark, monthly)", asSeries(govYield), euroBondMaturity, 0)
	report("EUROGOV-EUR", govMonthly.Points)

	govDailyYield := fetch(*base, "ECB/YC/B.U2.EUR.4F.G_N_A.SV_C_YM.SR_10Y")
	govDaily := simgen.TreasuryTR("Euro-area government bond total return (10y benchmark, daily)", asSeries(govDailyYield), euroBondMaturity, 0)
	report("EUROGOV-DAILY", govDaily.Points)

	// Long euro-area government bond TR (25+ segment), monthly (~1970) and daily
	// (~2004). The monthly tail synthesizes a 25-year yield from the OECD
	// long-term (10-year) benchmark; the daily series uses the real ECB 25-year
	// yield-curve point. Both
	// are priced at euroLongMaturity (vol-matched to DBXG, see above).
	longMonthlyYield := affine(asSeries(govYield), euroLongSlope, euroLongIntercept)
	govLongMonthly := simgen.TreasuryTR("Long euro-area government bond total return (25+, monthly)", longMonthlyYield, euroLongMaturity, 0)
	report("EUROGOV-LONG-EUR", govLongMonthly.Points)

	govLongDailyYield := fetch(*base, "ECB/YC/B.U2.EUR.4F.G_N_A.SV_C_YM.SR_25Y")
	govLongDaily := simgen.TreasuryTR("Long euro-area government bond total return (25+, daily)", asSeries(govLongDailyYield), euroLongMaturity, 0)
	report("EUROGOV-LONG-DAILY", govLongDaily.Points)

	// Eurozone equity net TR, monthly (~1986).
	price := fetch(*base, "OECD/DSD_STES@DF_FINMARK/EA20.M.SHARE.IX._Z._Z._Z._Z.N")
	equity := grossUp(price, netDivYield)
	report("EMU-EUR", equity)

	// German 3-month money-market accrual, monthly (~1960), for the pre-euro
	// cash tail. Trimmed at 1995 so it only ever feeds the splice under
	// EURCASH-EUR (which starts 1994).
	shortRate := fetch(*base, "OECD/DSD_STES@DF_FINMARK/DEU.M.IR3TIB.PA._Z._Z._Z._Z.N")
	cash := accrue(shortRate, time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))
	report("DECASH-EUR", cash)

	if *check {
		runChecks(*dir, govMonthly, govDaily, govLongMonthly, govLongDaily, equity, cash)
	}
	if *dry {
		return
	}
	write(*dir, "EMU-EUR", "Eurozone equity total return (OECD euro-area share prices grossed to net TR, EUR, monthly)",
		fmt.Sprintf("OECD euro-area share-price index EA20.M.SHARE (dataflow DSD_STES@DF_FINMARK, price only, ~1986-12) grossed to a net total return by a constant %.1f%%/yr net dividend yield calibrated on the EZU (MSCI Eurozone net TR) EUR overlap; via DBnomics. Proxy behind EZU.", netDivYield*100), equity)
	write(*dir, "EUROGOV-EUR", "Euro-area government bond total return (10-year benchmark, EUR, monthly)",
		"OECD euro-area long-term government bond yield EA20.M.IRLT (dataflow DSD_STES@DF_FINMARK, ~1970-01) run through the constant-maturity reconstruction (TreasuryTR, 10y); via DBnomics. Proxy behind the euro-govt bond ETF.", govMonthly.Points)
	write(*dir, "EUROGOV-DAILY", "Euro-area government bond total return (10-year benchmark, EUR, daily)",
		"ECB daily euro-area 10y yield-curve point B.U2.EUR.4F.G_N_A.SV_C_YM.SR_10Y (~2004) run through TreasuryTR (10y); via DBnomics. Daily shape for EUROGOV-EUR.", govDaily.Points)
	write(*dir, "EUROGOV-LONG-EUR", "Long euro-area government bond total return (25+ segment, EUR, monthly)",
		fmt.Sprintf("OECD euro-area long-term government bond yield EA20.M.IRLT (dataflow DSD_STES@DF_FINMARK, ~1970-01) mapped to a 25y yield (%.3f+%.4f*10y, calibrated on the 2004-2026 ECB curve) run through TreasuryTR (%.0fy par, modified duration ~17, vol-matched to DBXG); via DBnomics. Proxy behind the euro 25+ govt ETF (DBXG).", euroLongIntercept, euroLongSlope, euroLongMaturity), govLongMonthly.Points)
	write(*dir, "EUROGOV-LONG-DAILY", "Long euro-area government bond total return (25+ segment, EUR, daily)",
		"ECB daily euro-area 25y yield-curve point B.U2.EUR.4F.G_N_A.SV_C_YM.SR_25Y (~2004) run through TreasuryTR (24y par, modified duration ~17, vol-matched to DBXG); via DBnomics. Daily shape for EUROGOV-LONG-EUR.", govLongDaily.Points)
	write(*dir, "DECASH-EUR", "German 3-month money-market accrual (EUR/DM, monthly)",
		"OECD German 3-month interbank rate DEU.M.IR3TIB (dataflow DSD_STES@DF_FINMARK, ~1960-01) compounded into a money-market level; via DBnomics. Pre-euro cash tail spliced under EURCASH-EUR at 1994.", cash)
}

// obs is one dated observation.
type obs struct {
	date time.Time
	val  float64
}

// fetch downloads one DBnomics series and returns its non-null observations in
// date order. Monthly ("YYYY-MM") and daily ("YYYY-MM-DD") periods are both
// accepted; a monthly period is anchored on the first of the month.
func fetch(base, path string) []obs {
	url := fmt.Sprintf("%s/series/%s?observations=1", base, path)
	cl := &http.Client{Timeout: 60 * time.Second}
	resp, err := cl.Get(url)
	if err != nil {
		log.Fatalf("%s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: HTTP %d", path, resp.StatusCode)
	}
	var body struct {
		Series struct {
			Docs []struct {
				Period []string `json:"period"`
				Value  []any    `json:"value"`
			} `json:"docs"`
		} `json:"series"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Fatalf("%s: decode: %v", path, err)
	}
	if len(body.Series.Docs) == 0 {
		log.Fatalf("%s: no series returned", path)
	}
	doc := body.Series.Docs[0]
	out := make([]obs, 0, len(doc.Period))
	for i, per := range doc.Period {
		v, ok := doc.Value[i].(float64)
		if !ok {
			continue // DBnomics encodes gaps as the JSON string "NA" or null
		}
		t, err := parsePeriod(per)
		if err != nil {
			log.Fatalf("%s: bad period %q: %v", path, per, err)
		}
		out = append(out, obs{date: t, val: v})
	}
	if len(out) < 2 {
		log.Fatalf("%s: only %d usable observations", path, len(out))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].date.Before(out[j].date) })
	return out
}

func parsePeriod(p string) (time.Time, error) {
	if len(p) == 7 { // YYYY-MM
		return time.Parse("2006-01", p)
	}
	return time.Parse("2006-01-02", p)
}

// asSeries wraps yield observations as a marketdata series (levels are the
// annualized percent yields) for TreasuryTR.
func asSeries(o []obs) *marketdata.Series {
	s := &marketdata.Series{Name: "yield", Source: "simdata"}
	for _, p := range o {
		s.Points = append(s.Points, marketdata.Point{Date: p.date, Close: p.val})
	}
	return s
}

// affine returns a copy of a yield series with every level mapped to
// slope*level + intercept (percent in, percent out), used to synthesize a
// longer-maturity yield from a shorter benchmark over the deep tail.
func affine(y *marketdata.Series, slope, intercept float64) *marketdata.Series {
	out := &marketdata.Series{Name: y.Name, Source: y.Source}
	out.Points = make([]marketdata.Point, len(y.Points))
	for i, p := range y.Points {
		out.Points[i] = marketdata.Point{Date: p.Date, Close: slope*p.Close + intercept}
	}
	return out
}

// grossUp turns a price index into a net-total-return level (base 100) by
// compounding each period's price return together with the pro-rata-temporis
// constant net dividend yield.
func grossUp(price []obs, annualDiv float64) []marketdata.Point {
	out := make([]marketdata.Point, 0, len(price))
	val := 100.0
	out = append(out, marketdata.Point{Date: price[0].date, Close: val})
	for i := 1; i < len(price); i++ {
		yrs := price[i].date.Sub(price[i-1].date).Hours() / 24 / 365.25
		val *= (price[i].val / price[i-1].val) * math.Pow(1+annualDiv, yrs)
		out = append(out, marketdata.Point{Date: price[i].date, Close: val})
	}
	return out
}

// accrue compounds a short-rate series (annualized percent) into a
// money-market level (base 100), pro rata temporis, up to (but excluding) end.
func accrue(rate []obs, end time.Time) []marketdata.Point {
	out := make([]marketdata.Point, 0, len(rate))
	val := 100.0
	out = append(out, marketdata.Point{Date: rate[0].date, Close: val})
	for i := 1; i < len(rate); i++ {
		if !rate[i].date.Before(end) {
			break
		}
		yrs := rate[i].date.Sub(rate[i-1].date).Hours() / 24 / 365.25
		val *= math.Pow(1+rate[i-1].val/100, yrs)
		out = append(out, marketdata.Point{Date: rate[i].date, Close: val})
	}
	return out
}

func report(id string, pts []marketdata.Point) {
	if len(pts) == 0 {
		log.Fatalf("%s: empty", id)
	}
	first, last := pts[0], pts[len(pts)-1]
	yrs := last.Date.Sub(first.Date).Hours() / 24 / 365.25
	cagr := math.Pow(last.Close/first.Close, 1/yrs) - 1
	log.Printf("%-13s %4d points  %s..%s  CAGR %.2f%%/yr",
		id, len(pts), first.Date.Format("2006-01"), last.Date.Format("2006-01"), cagr*100)
}

func write(dir, id, name, source string, pts []marketdata.Point) {
	var b strings.Builder
	b.WriteString("# pofo simdata v1\n")
	fmt.Fprintf(&b, "# id: %s\n", id)
	fmt.Fprintf(&b, "# name: %s\n", name)
	fmt.Fprintf(&b, "# source: %s\n", source)
	b.WriteString("date,close\n")
	for _, p := range pts {
		fmt.Fprintf(&b, "%s,%.6f\n", p.Date.Format("2006-01-02"), p.Close)
	}
	path := filepath.Join(dir, id+".csv")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
	log.Printf("wrote %s", path)
}

// runChecks measures each new series where an outside answer is known, and
// stops the generator when one of them is not met. A reconstruction that
// downloaded cleanly has proved nothing: the house rule is that a series is
// validated against a reference BEFORE anything is allowed to trust it.
//
// The checks, and why each number is the one to expect:
//
//   - Freshness. The MEI dataflow this generator used to read went on answering
//     HTTP 200 for two and a half years after it stopped being updated, so a
//     stale tail is the failure mode to catch first: every OECD-sourced series
//     must reach within a year of today, and the two ECB curve series within a
//     quarter (the ECB publishes daily, the OECD with a lag and revisions).
//   - Flat runs. The daily series shipped before this check carried a 698-day
//     stretch (2019-04 to 2022-01, the covid drawdown included) at one constant
//     level, from a degraded fetch that reported nothing and passed review: a
//     total return that repeats to the last decimal for weeks is not a bond, so
//     a daily series may not repeat a level more than 5 observations running,
//     and a monthly one more than 2.
//   - Euro-area vs German (1999-2010). Over the euro's first decade, before the
//     sovereign crisis reopened them, the spreads between the euro aggregate and
//     the Bund were a few tens of basis points, so the euro-area reconstruction
//     must track the bundled BUND-EUR closely: monthly correlation above 0.95
//     and less than a point a year between their CAGRs. It is the mirror image
//     of the check gen-gbond-refdata runs on the German side.
//   - Euro-area daily vs monthly (2005->). The ECB daily curve and the OECD
//     monthly yield describe the same 10-year bond, so their monthly
//     volatilities must agree to within about a third, the daily one being the
//     higher: the OECD publishes a monthly AVERAGE yield, which damps whatever
//     happened inside the month. What this really catches is a wrong curve
//     point, which would be off by a factor rather than by a fifth.
//   - The long bond's volatility (2007->). EUROGOV-LONG-DAILY is vol-matched to
//     DBXG's ~14.4%/yr realized volatility over 2007-2026, which is what
//     euroLongMaturity was trimmed to 24 years for; it must still land inside
//     11-18%/yr, and above the 10-year's, or the maturity is no longer the one
//     the calibration assumed.
//   - The long tail's synthesis (2004->). Where the synthesized 25-year yield
//     and the real ECB one overlap, the two reconstructions must agree on level
//     and on risk: less than a point a year of CAGR gap, and volatilities within
//     a third of each other. This grades euroLongIntercept/euroLongSlope against
//     the curve they were fitted on, so a refreshed long-term yield that no
//     longer maps that way is caught. Their monthly CORRELATION is logged but
//     not gated: the OECD publishes a monthly AVERAGE yield while the ECB curve
//     is read month-end, which offsets the two by half a month and drags the
//     correlation of their monthly returns down whatever the mapping does.
//   - The equity gross-up (2001-2023). netDivYield is calibrated so the proxy
//     matches EZU (MSCI Eurozone, net TR, in EUR) over that window, ~3.05%/yr.
//     The check recomputes the proxy's CAGR there and requires it inside
//     2.5-3.6%/yr: a revised share-price index that moved the calibration shows
//     up here rather than silently in the pre-2000 drift.
//   - The cash accrual. A money-market index is monotone by construction and
//     post-war German 3-month rates sit inside 0-15%/yr, so DECASH-EUR's CAGR
//     over its whole span must land inside 0-12%/yr with no drawdown.
func runChecks(dir string, gov, govDaily, govLong, govLongDaily *marketdata.Series, equity, cash []marketdata.Point) {
	failed := 0
	fail := func(format string, a ...any) {
		failed++
		log.Printf("CHECK FAILED: "+format, a...)
	}

	now := time.Now().UTC()
	for _, f := range []struct {
		id  string
		pts []marketdata.Point
		max time.Duration
	}{
		{"EUROGOV-EUR", gov.Points, 365 * 24 * time.Hour},
		{"EUROGOV-LONG-EUR", govLong.Points, 365 * 24 * time.Hour},
		{"EMU-EUR", equity, 365 * 24 * time.Hour},
		{"EUROGOV-DAILY", govDaily.Points, 92 * 24 * time.Hour},
		{"EUROGOV-LONG-DAILY", govLongDaily.Points, 92 * 24 * time.Hour},
	} {
		last := f.pts[len(f.pts)-1].Date
		log.Printf("check %s freshness: last %s (%.0f days old)", f.id, last.Format("2006-01-02"), now.Sub(last).Hours()/24)
		if now.Sub(last) > f.max {
			fail("%s stops at %s: its source dataflow looks frozen", f.id, last.Format("2006-01"))
		}
	}

	for _, f := range []struct {
		id  string
		pts []marketdata.Point
		max int
	}{
		{"EUROGOV-EUR", gov.Points, 2},
		{"EUROGOV-LONG-EUR", govLong.Points, 2},
		{"EMU-EUR", equity, 2},
		{"DECASH-EUR", cash, 2},
		{"EUROGOV-DAILY", govDaily.Points, 5},
		{"EUROGOV-LONG-DAILY", govLongDaily.Points, 5},
	} {
		n, at := longestFlatRun(f.pts)
		where := "none"
		if n > 0 {
			where = "ending " + at.Format("2006-01-02")
		}
		log.Printf("check %s flat runs: longest %d repeated observations (%s)", f.id, n, where)
		if n > f.max {
			fail("%s repeats one level for %d observations ending %s: the fetch was degraded", f.id, n, at.Format("2006-01-02"))
		}
	}

	if bund, err := readRefdata(dir, "BUND-EUR"); err != nil {
		log.Printf("check: BUND-EUR unavailable (%v), skipping the German cross-check", err)
	} else {
		from, to := date(1999, 1), date(2010, 1)
		ce, cb := cagr(gov, from, to), cagr(bund, from, to)
		corr := monthlyCorr(gov, bund, from, to)
		log.Printf("check EUROGOV-EUR vs BUND-EUR 1999-2010: CAGR %.2f%% vs %.2f%% (gap %+.2f), monthly corr %.3f",
			ce*100, cb*100, (ce-cb)*100, corr)
		if corr < 0.95 || math.Abs(ce-cb) > 0.01 {
			fail("the euro-area and German reconstructions diverge over the years their spreads were thin")
		}
	}

	from, to := date(2005, 1), govDaily.Last().Date.AddDate(0, 0, 1)
	vd, vm := vol(govDaily, from, to), vol(gov, from, to)
	log.Printf("check EUROGOV daily vs monthly 2005->: vol %.2f%% vs %.2f%%/yr (ratio %.2f)", vd*100, vm*100, vd/vm)
	if vd/vm < 0.8 || vd/vm > 1.35 {
		fail("the daily ECB curve and the monthly OECD yield do not describe the same bond")
	}

	from, to = date(2007, 1), govLongDaily.Last().Date.AddDate(0, 0, 1)
	vl, v10 := vol(govLongDaily, from, to), vol(govDaily, from, to)
	log.Printf("check EUROGOV-LONG-DAILY 2007->: vol %.2f%%/yr (10y %.2f%%/yr, ratio %.2f)", vl*100, v10*100, vl/v10)
	if vl < 0.11 || vl > 0.18 || vl < v10 {
		fail("the long reconstruction is no longer vol-matched to DBXG (~14.4%%/yr)")
	}

	from, to = date(2004, 9), govLong.Last().Date.AddDate(0, 0, 1)
	cs, cr := cagr(govLong, from, to), cagr(govLongDaily, from, to)
	vs, vr := vol(govLong, from, to), vol(govLongDaily, from, to)
	log.Printf("check EUROGOV-LONG synthesized vs ECB curve 2004->: CAGR %.2f%% vs %.2f%% (gap %+.2f), vol %.2f%% vs %.2f%%/yr (ratio %.2f), monthly corr %.3f (not gated, monthly-average yield)",
		cs*100, cr*100, (cs-cr)*100, vs*100, vr*100, vs/vr, monthlyCorr(govLong, govLongDaily, from, to))
	if math.Abs(cs-cr) > 0.01 || vs/vr < 0.7 || vs/vr > 1.35 {
		fail("the synthesized 25y yield (%.3f+%.4f*10y) no longer matches the ECB curve it was fitted on", euroLongIntercept, euroLongSlope)
	}

	eq := &marketdata.Series{Points: equity}
	from, to = date(2001, 1), date(2024, 1)
	ceq := cagr(eq, from, to)
	log.Printf("check EMU-EUR 2001-2023 (netDivYield %.1f%%/yr calibration vs EZU ~3.05%%/yr): CAGR %.2f%%/yr", netDivYield*100, ceq*100)
	if ceq < 0.025 || ceq > 0.036 {
		fail("the gross-up no longer reproduces the EZU net-TR overlap netDivYield was calibrated on")
	}

	c := &marketdata.Series{Points: cash}
	rate, dd := cagr(c, c.First().Date, c.Last().Date.AddDate(0, 0, 1)), drawdown(cash)
	log.Printf("check DECASH-EUR: %.2f%%/yr over %s..%s, deepest drawdown %.2f%%", rate*100,
		c.First().Date.Format("2006-01"), c.Last().Date.Format("2006-01"), dd*100)
	if rate < 0 || rate > 0.12 || dd < 0 {
		fail("DECASH-EUR is not a plausible money-market accrual")
	}

	if failed > 0 {
		log.Fatalf("%d sanity check(s) failed, nothing written", failed)
	}
}

// longestFlatRun returns the longest run of consecutive observations carrying
// the exact same level, and the date it ends on. A reconstruction is a product
// of compounding, so an exactly repeated level means the input stopped moving,
// which for a bond total return means the input stopped arriving.
func longestFlatRun(pts []marketdata.Point) (int, time.Time) {
	best, at, run := 0, time.Time{}, 0
	for i := 1; i < len(pts); i++ {
		if pts[i].Close != pts[i-1].Close {
			run = 0
			continue
		}
		run++
		if run > best {
			best, at = run, pts[i].Date
		}
	}
	return best, at
}

// readRefdata loads an already-bundled reference series, so a new series can be
// graded against one the repository already trusts.
func readRefdata(dir, id string) (*marketdata.Series, error) {
	s, ok, err := marketdata.ReadSimdataFS(os.DirFS(dir), id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s: not found in %s", id, dir)
	}
	return s, nil
}

func date(y, m int) time.Time { return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC) }

// window returns the levels of s over [from, to).
func window(s *marketdata.Series, from, to time.Time) []marketdata.Point {
	var out []marketdata.Point
	for _, p := range s.Points {
		if !p.Date.Before(from) && p.Date.Before(to) {
			out = append(out, p)
		}
	}
	return out
}

func cagr(s *marketdata.Series, from, to time.Time) float64 {
	pts := window(s, from, to)
	if len(pts) < 2 {
		return math.NaN()
	}
	yrs := pts[len(pts)-1].Date.Sub(pts[0].Date).Hours() / 24 / 365.25
	return math.Pow(pts[len(pts)-1].Close/pts[0].Close, 1/yrs) - 1
}

// vol is the annualized volatility of the MONTHLY returns over the window, so
// a daily and a monthly series of the same bond are compared on one cadence.
func vol(s *marketdata.Series, from, to time.Time) float64 {
	r := monthlyReturns(monthly(s, from, to))
	if len(r) < 12 {
		return math.NaN()
	}
	return stdev(r) * math.Sqrt(12)
}

// monthEnd is one calendar month's closing level, keyed "YYYY-MM", in order.
type monthEnd struct {
	key   string
	close float64
}

// monthly returns the last level of each calendar month in the window.
func monthly(s *marketdata.Series, from, to time.Time) []monthEnd {
	var out []monthEnd
	for _, p := range window(s, from, to) {
		k := p.Date.Format("2006-01")
		if n := len(out); n > 0 && out[n-1].key == k {
			out[n-1].close = p.Close
			continue
		}
		out = append(out, monthEnd{key: k, close: p.Close})
	}
	return out
}

func monthlyReturns(m []monthEnd) []float64 {
	r := make([]float64, 0, len(m))
	for i := 1; i < len(m); i++ {
		r = append(r, m[i].close/m[i-1].close-1)
	}
	return r
}

// monthlyCorr correlates two series' monthly returns over the window, matching
// them by calendar month rather than by position, so two series that skip a
// different month are never compared off by one.
func monthlyCorr(a, b *marketdata.Series, from, to time.Time) float64 {
	byKey := make(map[string]float64)
	for _, m := range monthly(b, from, to) {
		byKey[m.key] = m.close
	}
	var ma, mb []monthEnd
	for _, m := range monthly(a, from, to) {
		if c, ok := byKey[m.key]; ok {
			ma = append(ma, m)
			mb = append(mb, monthEnd{key: m.key, close: c})
		}
	}
	ra, rb := monthlyReturns(ma), monthlyReturns(mb)
	if len(ra) < 12 {
		return math.NaN()
	}
	return pearson(ra, rb)
}

func stdev(xs []float64) float64 {
	m := metrics.Mean(xs)
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)))
}

func pearson(a, b []float64) float64 {
	ma, mb := metrics.Mean(a), metrics.Mean(b)
	var cov, va, vb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		cov += da * db
		va += da * da
		vb += db * db
	}
	if va <= 0 || vb <= 0 {
		return math.NaN()
	}
	return cov / math.Sqrt(va*vb)
}

// drawdown is the deepest peak-to-trough fall of a level series, as a negative
// fraction. The German 3-month rate never went negative over DECASH-EUR's span
// (it is trimmed at 1995), so this accrual is strictly monotone.
func drawdown(pts []marketdata.Point) float64 {
	peak, worst := math.Inf(-1), 0.0
	for _, p := range pts {
		peak = max(peak, p.Close)
		worst = min(worst, p.Close/peak-1)
	}
	return worst
}
