// Command gen-euro-refdata builds the bundled euro-area reference series that
// let the eurozone recipes (NTSZ, the WisdomTree Eurozone Efficient Core) reach
// back before their fetchable components. It runs at data-generation time only
// (network); the pofo binary embeds the CSVs and never fetches OECD or the ECB.
//
// Everything is sourced from DBnomics (free, key-less), the same mirror the
// macro panel uses. Four series are written into pkg/datasets/refdata/:
//
//   - EMU-EUR.csv       eurozone equity net total return (monthly, ~1986):
//     the OECD euro-area share-price index (EA19.SPASTT01, a
//     price index) grossed to a net total return by a constant
//     net dividend yield (netDivYield), calibrated on the
//     overlap with the real MSCI Eurozone (EZU) in EUR. Proxy
//     behind EZU in the equity leg.
//   - EUROGOV-EUR.csv   euro-area government bond total return (monthly, ~1970):
//     the OECD euro-area 10-year benchmark yield
//     (EA19.IRLTLT01) run through the constant-maturity
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
// Usage: gen-euro-refdata [-base URL] [-dir path] [-dry]
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
// yield is synthesized from the OECD 10-year benchmark as
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
	dry := flag.Bool("dry", false, "print coverage without writing")
	flag.Parse()

	// Euro-area government bond TR, monthly (~1970) and daily (~2004).
	govYield := fetch(*base, "OECD/MEI/EA19.IRLTLT01.ST.M")
	govMonthly := simgen.TreasuryTR("Euro-area government bond total return (10y benchmark, monthly)", asSeries(govYield), euroBondMaturity, 0)
	report("EUROGOV-EUR", govMonthly.Points)

	govDailyYield := fetch(*base, "ECB/YC/B.U2.EUR.4F.G_N_A.SV_C_YM.SR_10Y")
	govDaily := simgen.TreasuryTR("Euro-area government bond total return (10y benchmark, daily)", asSeries(govDailyYield), euroBondMaturity, 0)
	report("EUROGOV-DAILY", govDaily.Points)

	// Long euro-area government bond TR (25+ segment), monthly (~1970) and daily
	// (~2004). The monthly tail synthesizes a 25-year yield from the OECD
	// 10-year; the daily series uses the real ECB 25-year yield-curve point. Both
	// are priced at euroLongMaturity (vol-matched to DBXG, see above).
	longMonthlyYield := affine(asSeries(govYield), euroLongSlope, euroLongIntercept)
	govLongMonthly := simgen.TreasuryTR("Long euro-area government bond total return (25+, monthly)", longMonthlyYield, euroLongMaturity, 0)
	report("EUROGOV-LONG-EUR", govLongMonthly.Points)

	govLongDailyYield := fetch(*base, "ECB/YC/B.U2.EUR.4F.G_N_A.SV_C_YM.SR_25Y")
	govLongDaily := simgen.TreasuryTR("Long euro-area government bond total return (25+, daily)", asSeries(govLongDailyYield), euroLongMaturity, 0)
	report("EUROGOV-LONG-DAILY", govLongDaily.Points)

	// Eurozone equity net TR, monthly (~1986).
	price := fetch(*base, "OECD/MEI/EA19.SPASTT01.IXOB.M")
	equity := grossUp(price, netDivYield)
	report("EMU-EUR", equity)

	// German 3-month money-market accrual, monthly (~1960), for the pre-euro
	// cash tail. Trimmed at 1995 so it only ever feeds the splice under
	// EURCASH-EUR (which starts 1994).
	shortRate := fetch(*base, "OECD/MEI/DEU.IR3TIB01.ST.M")
	cash := accrue(shortRate, time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC))
	report("DECASH-EUR", cash)

	if *dry {
		return
	}
	write(*dir, "EMU-EUR", "Eurozone equity total return (OECD euro-area share prices grossed to net TR, EUR, monthly)",
		fmt.Sprintf("OECD MEI euro-area share-price index EA19.SPASTT01 (price only, ~1986) grossed to a net total return by a constant %.1f%%/yr net dividend yield calibrated on the EZU (MSCI Eurozone net TR) EUR overlap; via DBnomics. Proxy behind EZU.", netDivYield*100), equity)
	write(*dir, "EUROGOV-EUR", "Euro-area government bond total return (10-year benchmark, EUR, monthly)",
		"OECD MEI euro-area 10y benchmark yield EA19.IRLTLT01 (~1970) run through the constant-maturity reconstruction (TreasuryTR, 10y); via DBnomics. Proxy behind the euro-govt bond ETF.", govMonthly.Points)
	write(*dir, "EUROGOV-DAILY", "Euro-area government bond total return (10-year benchmark, EUR, daily)",
		"ECB daily euro-area 10y yield-curve point B.U2.EUR.4F.G_N_A.SV_C_YM.SR_10Y (~2004) run through TreasuryTR (10y); via DBnomics. Daily shape for EUROGOV-EUR.", govDaily.Points)
	write(*dir, "EUROGOV-LONG-EUR", "Long euro-area government bond total return (25+ segment, EUR, monthly)",
		fmt.Sprintf("OECD MEI euro-area 10y benchmark yield EA19.IRLTLT01 (~1970) mapped to a 25y yield (%.3f+%.4f*10y, calibrated on the 2004-2026 ECB curve) run through TreasuryTR (%.0fy par, modified duration ~17, vol-matched to DBXG); via DBnomics. Proxy behind the euro 25+ govt ETF (DBXG).", euroLongIntercept, euroLongSlope, euroLongMaturity), govLongMonthly.Points)
	write(*dir, "EUROGOV-LONG-DAILY", "Long euro-area government bond total return (25+ segment, EUR, daily)",
		"ECB daily euro-area 25y yield-curve point B.U2.EUR.4F.G_N_A.SV_C_YM.SR_25Y (~2004) run through TreasuryTR (24y par, modified duration ~17, vol-matched to DBXG); via DBnomics. Daily shape for EUROGOV-LONG-EUR.", govLongDaily.Points)
	write(*dir, "DECASH-EUR", "German 3-month money-market accrual (EUR/DM, monthly)",
		"OECD MEI German 3-month interbank rate DEU.IR3TIB01 (~1960) compounded into a money-market level; via DBnomics. Pre-euro cash tail spliced under EURCASH-EUR at 1994.", cash)
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
