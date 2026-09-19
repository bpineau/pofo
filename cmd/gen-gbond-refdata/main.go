// Command gen-gbond-refdata builds the bundled NON-US government bond
// reference series behind the global bond overlay of the WisdomTree Global
// Efficient Core reconstruction (NTSG, pkg/simgen). It runs at
// data-generation time only (network); the pofo binary embeds the CSVs and
// never fetches OECD, the Bundesbank or the Japanese Ministry of Finance.
//
// The fund's bond sleeve is a basket of government bond futures in four
// currencies, not a US-only ladder: the audited holdings put roughly 80 % of
// the notional on US Treasuries, 11 % on German, 6 % on Japanese and 3 % on
// British paper. The US leg is covered by real fund donors that already reach
// 1953 (VFITX, VUSTX and the CMT reconstructions behind them); the other three
// had no euro-, yen- or sterling-denominated building block in the repository,
// which is what this generator supplies. Six series are written into
// pkg/datasets/refdata/:
//
//   - BUND-EUR.csv     German government bond total return (10-year benchmark,
//     monthly, 1956-05→): the month-ends of the Bundesbank's
//     own daily curve from the day it starts (1997-08), and
//     before it the OECD long-term government bond yield for
//     Germany rebased onto them at the junction; both run
//     through the constant-maturity reconstruction
//     simgen.TreasuryTR. The euro-area aggregate EUROGOV-EUR is
//     deliberately NOT reused: it smears the periphery spreads
//     of 2011-2012 that a Bund basket never carried. Real data
//     sets the level wherever it exists, because an OECD monthly
//     yield is the month's AVERAGE of the daily quotes: driven
//     by it alone this reconstruction read 0.83 of the real
//     curve's monthly volatility and correlated with it as
//     strongly one month late (0.55) as contemporaneously
//     (0.67), an average's signature rather than a close's.
//   - BUND-DAILY.csv   the same at daily granularity (1997-08→): the
//     Bundesbank's own daily 10-year point of the listed
//     federal securities term structure (Svensson), through
//     the same TreasuryTR. Daily shape for BUND-EUR, and the
//     series its month-ends are taken from.
//   - JGB-JPY.csv      Japanese government bond total return (10-year
//     benchmark, DAILY, 1986-07→): the Ministry of Finance's
//     historical JGB interest-rate table, 10Y column, through
//     TreasuryTR.
//   - GILT-GBP.csv     British government bond total return (10-year
//     benchmark, monthly, 1960-01→): the OECD long-term
//     government bond yield for the United Kingdom through
//     TreasuryTR. No daily shape in this pass: the sleeve is
//     ~1.8 % of the fund's net assets and the Bank of England
//     daily curve ships as a workbook, so a monthly texture is
//     accepted there and documented in the recipe. This is the
//     one leg with no real daily or month-end curve to splice
//     onto, so it carries the month-AVERAGE cadence end to end
//     and says so in its header (monthAverageNote): measured
//     against IGLT.L over 2008-2026 it reads 0.78 of the fund's
//     monthly volatility and correlates 0.62 contemporaneously
//     against 0.53 one month late.
//   - JPCASH-JPY.csv   Japanese overnight money-market accrual (monthly,
//     1985-07→): the OECD immediate (call money) rate for
//     Japan, compounded. What the yen futures leg finances at.
//   - GBCASH-GBP.csv   British overnight money-market accrual (monthly,
//     1978-01→): the OECD immediate (interbank) rate for the
//     United Kingdom, compounded. What the sterling futures
//     leg finances at.
//
// The OECD series come from DBnomics (free, key-less), the same mirror the
// macro panel and the euro reference series use, but from the CURRENT
// short-term-statistics dataflow (OECD/DSD_STES@DF_FINMARK) rather than the
// legacy MEI dataset, which stopped being updated in 2024-01. The Bundesbank
// series is on DBnomics too; the Japanese one is a plain CSV served by the
// Ministry of Finance.
//
// Every series is checked before it is written (-check, on by default): each
// one's CAGR and annualized volatility over a window with a known answer, and
// the German reconstruction against the euro-area one over the years when
// their spreads were thin. Nothing here is trusted on the strength of having
// downloaded cleanly.
//
// Usage: gen-gbond-refdata [-base URL] [-jgb URL] [-dir path] [-dry] [-check=false]
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/simgen"
)

const (
	defaultBase = "https://api.db.nomics.world/v22"
	// defaultJGB is the Ministry of Finance's historical JGB interest-rate
	// table. Only the /english/ path serves it; the Japanese one answers with
	// a portal page.
	defaultJGB = "https://www.mof.go.jp/english/policy/jgbs/reference/interest_rate/historical/jgbcme_all.csv"
)

// bondMaturity is the constant maturity (years) every leg of the basket is
// priced at. The fund holds a per-country ladder rather than a single tenor,
// but each ladder is built around its market's benchmark bond and lands within
// a year or so of a ten-year duration once the short and the long rung are
// averaged, so a single 10-year par bond per country is the honest summary of
// what the sleeve owns and of the only yield history that exists that deep.
const bondMaturity = 10.0

// jgbTenor is the column of the Ministry of Finance table that carries the
// 10-year benchmark. The table also publishes 1Y to 40Y; the shorter rungs
// only start quoting at various later dates and the longer ones are thin.
const jgbTenor = "10Y"

func main() {
	base := flag.String("base", defaultBase, "DBnomics API base URL")
	jgbURL := flag.String("jgb", defaultJGB, "Ministry of Finance historical JGB rates CSV")
	dir := flag.String("dir", "pkg/datasets/refdata", "output refdata directory")
	dry := flag.Bool("dry", false, "print coverage and checks without writing")
	check := flag.Bool("check", true, "run the sanity checks before writing")
	flag.Parse()

	// German 10-year, monthly (1956-05→) and daily (1997-08→). The monthly file
	// is the OECD-driven tail up to the day the Bundesbank curve starts and that
	// curve's own month-ends after it: see spliceCurve and atMonthEnd.
	bundYield := fetch(*base, "OECD/DSD_STES@DF_FINMARK/DEU.M.IRLT.PA._Z._Z._Z._Z.N")
	bundSynth := simgen.TreasuryTR("German government bond total return (10y benchmark, OECD monthly yield)", asSeries(bundYield), bondMaturity, 0)
	bundSynth.Points = atMonthEnd(bundSynth.Points)
	report("BUND-SYN", bundSynth.Points)

	bundDailyYield := fetch(*base, "BUBA/BBSIS/D.I.ZST.ZI.EUR.S1311.B.A604.R10XX.R.A.A._Z._Z.A")
	bundDaily := simgen.TreasuryTR("German government bond total return (10y benchmark, daily)", asSeries(bundDailyYield), bondMaturity, 0)
	report("BUND-DAILY", bundDaily.Points)

	bund, bundSplice := spliceCurve("BUND-EUR", bundSynth, bundDaily)
	log.Printf("BUND-EUR splice: OECD tail to %s, real Bundesbank curve from %s (real era rebased x%.4f onto the tail, junction month return %+.2f%%)",
		bundSplice.lastSynth.Format("2006-01"), bundSplice.at.Format("2006-01-02"), bundSplice.factor, bundSplice.seam*100)
	report("BUND-EUR", bund.Points)

	// Japanese 10-year, daily (1986-07→), and the yen call-money accrual.
	jgbYield := fetchJGB(*jgbURL)
	jgb := simgen.TreasuryTR("Japanese government bond total return (10y benchmark, daily)", jgbYield, bondMaturity, 0)
	report("JGB-JPY", jgb.Points)

	jpRate := fetch(*base, "OECD/DSD_STES@DF_FINMARK/JPN.M.IRSTCI.PA._Z._Z._Z._Z.N")
	jpCash := atAccrualEnd(accrue(jpRate))
	report("JPCASH-JPY", jpCash)

	// British 10-year, monthly (1960-01→), and the sterling interbank accrual.
	giltYield := fetch(*base, "OECD/DSD_STES@DF_FINMARK/GBR.M.IRLT.PA._Z._Z._Z._Z.N")
	gilt := simgen.TreasuryTR("British government bond total return (10y benchmark, monthly)", asSeries(giltYield), bondMaturity, 0)
	gilt.Points = atMonthEnd(gilt.Points)
	report("GILT-GBP", gilt.Points)

	gbRate := fetch(*base, "OECD/DSD_STES@DF_FINMARK/GBR.M.IRSTCI.PA._Z._Z._Z._Z.N")
	gbCash := atAccrualEnd(accrue(gbRate))
	report("GBCASH-GBP", gbCash)

	if *check {
		runChecks(*dir, bund, bundSynth, bundDaily, jgb, gilt, jpCash, gbCash, bundSplice)
	}
	if *dry {
		return
	}
	write(*dir, "BUND-EUR", "German government bond total return (10-year benchmark, EUR, monthly)",
		fmt.Sprintf("month-ends of the Bundesbank daily term structure of listed federal securities (Svensson), 10-year residual maturity, from %s; before it the OECD long-term government bond yield DEU.M.IRLT (dataflow DSD_STES@DF_FINMARK, ~1956) rebased onto them at the junction; both run through the constant-maturity reconstruction (TreasuryTR, %.0fy par); via DBnomics. German leg of the NTSG global bond overlay; the euro-area aggregate EUROGOV-EUR is not reused, it carries periphery spreads a Bund basket never had. %s",
			bundSplice.at.Format("2006-01"), bondMaturity, monthAverageNote), bund.Points)
	write(*dir, "BUND-DAILY", "German government bond total return (10-year benchmark, EUR, daily)",
		fmt.Sprintf("Bundesbank daily term structure of listed federal securities (Svensson), 10-year residual maturity, BBSIS D.I.ZST.ZI.EUR.S1311.B.A604.R10XX.R.A.A (~1997-08) run through TreasuryTR (%.0fy par); via DBnomics. Daily shape for BUND-EUR.", bondMaturity), bundDaily.Points)
	write(*dir, "JGB-JPY", "Japanese government bond total return (10-year benchmark, JPY, daily)",
		fmt.Sprintf("Japanese Ministry of Finance historical JGB interest rates (jgbcme_all.csv), %s column (~1986-07), run through TreasuryTR (%.0fy par). Japanese leg of the NTSG global bond overlay; the yield is negative on 453 days between 2016-02 and 2020-05 and the reconstruction prices those days rather than flat-lining them.", jgbTenor, bondMaturity), jgb.Points)
	write(*dir, "GILT-GBP", "British government bond total return (10-year benchmark, GBP, monthly)",
		fmt.Sprintf("OECD long-term government bond yield GBR.M.IRLT (dataflow DSD_STES@DF_FINMARK, ~1960) run through TreasuryTR (%.0fy par); via DBnomics. British leg of the NTSG global bond overlay; monthly texture accepted, the sleeve is ~1.8%% of net assets. No real daily or month-end gilt curve is spliced in front of it, unlike the German leg: the Bank of England publishes its daily curve as a workbook rather than a series, so this file carries the month-average cadence end to end. %s", bondMaturity, monthAverageNote), gilt.Points)
	write(*dir, "JPCASH-JPY", "Japanese overnight money-market accrual (JPY, monthly)",
		"OECD immediate interest rate, call money JPN.M.IRSTCI (dataflow DSD_STES@DF_FINMARK, ~1985-07) compounded into a money-market level, each row dated the month-END it closes (atAccrualEnd); via DBnomics. What the yen leg of the NTSG bond overlay finances at.", jpCash)
	write(*dir, "GBCASH-GBP", "British overnight money-market accrual (GBP, monthly)",
		"OECD immediate interest rate, interbank GBR.M.IRSTCI (dataflow DSD_STES@DF_FINMARK, ~1978-01) compounded into a money-market level, each row dated the month-END it closes (atAccrualEnd); via DBnomics. What the sterling leg of the NTSG bond overlay finances at.", gbCash)
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
	cl := &http.Client{Timeout: 120 * time.Second}
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

// fetchJGB downloads the Ministry of Finance's historical JGB rate table and
// returns the benchmark tenor's yields as a series. The file carries two
// header lines (a title row, then the tenor row), dates as YYYY/M/D with no
// zero padding, and a bare "-" wherever a tenor did not quote: the 10-year
// column only opens in 1986-07, fifteen years after the table starts.
func fetchJGB(url string) *marketdata.Series {
	cl := &http.Client{Timeout: 120 * time.Second}
	resp, err := cl.Get(url)
	if err != nil {
		log.Fatalf("JGB: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("JGB: HTTP %d", resp.StatusCode)
	}
	rows, err := csv.NewReader(io.LimitReader(resp.Body, 32<<20)).ReadAll()
	if err != nil {
		log.Fatalf("JGB: parse: %v", err)
	}
	if len(rows) < 3 {
		log.Fatalf("JGB: only %d rows", len(rows))
	}
	col := -1
	for i, h := range rows[1] {
		if strings.TrimSpace(h) == jgbTenor {
			col = i
		}
	}
	if col < 0 {
		log.Fatalf("JGB: no %s column in %q", jgbTenor, rows[1])
	}
	s := &marketdata.Series{Name: "JGB " + jgbTenor + " yield", Source: "simdata"}
	for _, r := range rows[2:] {
		if len(r) <= col {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(r[col]), 64)
		if err != nil {
			continue // "-": the tenor did not quote that day
		}
		d, err := time.Parse("2006/1/2", strings.TrimSpace(r[0]))
		if err != nil {
			log.Fatalf("JGB: bad date %q: %v", r[0], err)
		}
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: v})
	}
	if len(s.Points) < 2 {
		log.Fatalf("JGB: only %d usable observations", len(s.Points))
	}
	return s
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

// monthAverageNote is the sentence every OECD-driven monthly file carries, so
// that a reader of the CSV alone knows what one of its rows IS. The same
// sentence is in cmd/gen-euro-refdata, the other reader of this dataflow.
const monthAverageNote = "CADENCE: the OECD observation for a month is that month's AVERAGE of the daily quotes, so the level is reached mid-month and its monthly returns are smoothed. The label is the month's last day, the convention every other monthly series here carries, so that the file's steps and its junction with a real month-end segment are one month each; it is a name for the month, not a claim that the level is a month-end close."

// atMonthEnd re-dates monthly points onto the last day of the month they
// belong to, never past today. See cmd/gen-euro-refdata's copy for the full
// reasoning: an OECD monthly observation is the month's AVERAGE, so no label is
// exactly right, and the one this bundle uses everywhere else names the month.
// The relabelling leaves every monthly return, calendar year and CAGR
// bit-identical, and buys a clean junction with the real curve segment.
func atMonthEnd(pts []marketdata.Point) []marketdata.Point {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	out := make([]marketdata.Point, len(pts))
	for i, p := range pts {
		end := time.Date(p.Date.Year(), p.Date.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1)
		if end.After(today) {
			end = today
		}
		out[i] = marketdata.Point{Date: end, Close: p.Close}
	}
	return out
}

// atAccrualEnd re-dates a money-market accrual onto the month-end each of its
// levels actually belongs to, which is the day BEFORE the date accrue gives it:
// the level written at "M-01" is the cash a holder had once month M-1 had been
// earned in full, i.e. the close of M-1. See cmd/gen-euro-refdata's copy for the
// full reasoning; nothing but the label moves, every interest payment having
// been computed from the source's own dates before this runs.
func atAccrualEnd(pts []marketdata.Point) []marketdata.Point {
	out := make([]marketdata.Point, len(pts))
	for i, p := range pts {
		out[i] = marketdata.Point{Date: p.Date.AddDate(0, 0, -1), Close: p.Close}
	}
	return out
}

// curveSplice records where a deep OECD-driven tail hands over to the real
// daily curve, so the junction can be reported and checked rather than assumed.
type curveSplice struct {
	at        time.Time // first real month-end kept, and the start of the real era
	lastSynth time.Time // last OECD-driven month kept in front of it
	factor    float64   // level rebasing applied to the real era, to meet the tail
	seam      float64   // the (OECD-driven) return carried across the junction
}

// spliceCurve joins the deep OECD-driven monthly tail to the reconstruction
// built on the real Bundesbank daily curve, at the first month that curve
// covers. The doctrine is the donor chains': real data sets the LEVEL wherever
// real data exists, and the month-average tail only fills the years in front of
// it. Every month the curve reaches is taken from the curve, on the day its
// month-end fell, multiplied by one constant factor chosen so the two meet with
// no level jump, exactly as marketdata.ExtendBack rebases one series onto
// another; rebasing the REAL part keeps the file based at 100 on its first date,
// and a constant factor is not a return either way. The month straddling the
// junction keeps the OECD-driven return, the only one available for it, which
// also absorbs the curve's partial first month.
func spliceCurve(id string, synth, curve *marketdata.Series) (*marketdata.Series, curveSplice) {
	ends := monthEnds(curve.Points)
	if len(ends) == 0 {
		log.Fatalf("%s: the daily curve is empty, nothing to splice onto", id)
	}
	key := ends[0].Date.Format("2006-01")
	var tail []marketdata.Point
	var at float64
	for _, p := range synth.Points {
		switch k := p.Date.Format("2006-01"); {
		case k < key:
			tail = append(tail, p)
		case k == key:
			at = p.Close
		}
	}
	if at == 0 || len(tail) == 0 {
		log.Fatalf("%s: the deep tail does not reach the curve's first month (%s)", id, key)
	}
	sp := curveSplice{at: ends[0].Date, lastSynth: tail[len(tail)-1].Date, factor: at / ends[0].Close}
	out := &marketdata.Series{Name: id + " (monthly)", Source: synth.Source}
	out.Points = append(out.Points, tail...)
	for _, p := range ends {
		out.Points = append(out.Points, marketdata.Point{Date: p.Date, Close: p.Close * sp.factor})
	}
	sp.seam = out.Points[len(tail)].Close/tail[len(tail)-1].Close - 1
	return out, sp
}

// monthEnds keeps the last observation of each calendar month, on its own date.
func monthEnds(pts []marketdata.Point) []marketdata.Point {
	var out []marketdata.Point
	for _, p := range pts {
		if n := len(out); n > 0 && out[n-1].Date.Format("2006-01") == p.Date.Format("2006-01") {
			out[n-1] = p
			continue
		}
		out = append(out, p)
	}
	return out
}

// longestStep returns the longest interval between consecutive points, in days,
// and the date it ends on. On a monthly file anything past 45 days means a row
// carries more than one month of return, which is what a dating-convention
// junction produces.
func longestStep(pts []marketdata.Point) (float64, time.Time) {
	worst, at := 0.0, time.Time{}
	for i := 1; i < len(pts); i++ {
		if d := pts[i].Date.Sub(pts[i-1].Date).Hours() / 24; d > worst {
			worst, at = d, pts[i].Date
		}
	}
	return worst, at
}

// accrue compounds a short-rate series (annualized percent) into a
// money-market level (base 100), pro rata temporis. Each step accrues at the
// rate that PREVAILED over it, i.e. the previous observation's level, so the
// index never earns a rate before it was published.
//
// A cash accrual has no smoothing defect to repair, unlike the bond
// reconstructions above: the OECD's month-average rate is exactly what a
// money-market roll held through that month earns, and these two indices carry
// 0.6 and 1.3 %/yr of annualized monthly-return dispersion against a bond's 5 to
// 6, so there is nothing for a half-month of phase to damage. What they do get
// is the bundle's month-end LABEL, from atAccrualEnd, so that a composite
// reading a cash leg beside a bond leg sees one point a month and not two.
func accrue(rate []obs) []marketdata.Point {
	out := make([]marketdata.Point, 0, len(rate))
	val := 100.0
	out = append(out, marketdata.Point{Date: rate[0].date, Close: val})
	for i := 1; i < len(rate); i++ {
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
	log.Printf("%-11s %5d points  %s..%s  CAGR %.2f%%/yr",
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
// The five checks, and why each number is the one to expect:
//
//   - German vs euro-area. Over the euro's first decade, before the sovereign
//     crisis reopened them, the spreads between the euro aggregate and the Bund
//     were a few tens of basis points. The two reconstructions must therefore
//     agree closely: less than a point a year between their 1999-2010 CAGRs, and
//     a monthly correlation above 0.90 from 2005, the year from which both files
//     are on their real month-end curve. It is the one check that grades the
//     German series against something already bundled.
//   - German daily curve vs the OECD tail (1998-2026). The daily Bundesbank
//     curve and the monthly OECD yield describe the same bond, so their monthly
//     volatilities must agree to within about a third. The daily one is expected
//     to be the higher of the two: the OECD publishes a monthly AVERAGE yield,
//     which damps whatever happened inside the month (measured: ratio 0.83, and
//     the monthly returns correlate 0.67 at lag 0 against 0.55 at lag -1, an
//     average's signature). That is why the shipped file takes the real curve's
//     month-ends from 1997-08 and leaves the OECD tail only the years nothing
//     else reaches; the check grades that tail, which is the part still in the
//     file, and what it is really for is a wrong curve point, since a 2-year or
//     a 30-year residual maturity would be off by a factor rather than a fifth.
//   - The German splice. After the junction the shipped monthly series IS the
//     daily one, sampled: every month-end must match the Bundesbank
//     reconstruction to the last bit, or the sampling has slipped a month. The
//     rebasing factor and the single month of return carried across the junction
//     are checked for sanity, and so is the junction's own STEP: a tail dated
//     the first of the month handing over to a month-end segment spans 60 days
//     and carries two months of return in one row, which is what atMonthEnd
//     removes. No monthly step may exceed 45 days, the same bar
//     pkg/datasets/golden/gaps_test.go holds the whole bundle to.
//   - Japanese, yield-curve-control era (2016-2021). A 10-year JGB pinned at
//     zero by the Bank of Japan is the calmest government bond of the modern
//     record: its total return must be quiet (volatility under 3 %/yr) and
//     essentially flat (|CAGR| under 3 %/yr). If the reconstruction flat-lines
//     the negative-yield days instead of pricing them, the volatility collapses
//     towards zero and this catches it from below.
//   - British, the 1970s (1970-1980). The gilt market's worst decade: double
//     digit yields, a nominal total return positive but far behind inflation,
//     and volatility well above a calm bond's. Anything outside 0-15 %/yr with
//     a volatility under 4 % would mean the yield series was misread.
//   - The two cash accruals. A money-market index is monotone by construction,
//     and both countries' post-war short rates sit inside 0-20 %/yr, so their
//     CAGR over the whole span must land inside 0-12 %/yr with no drawdown.
func runChecks(dir string, bund, bundSynth, bundDaily, jgb, gilt *marketdata.Series, jpCash, gbCash []marketdata.Point, splice curveSplice) {
	failed := 0
	fail := func(format string, a ...any) {
		failed++
		log.Printf("CHECK FAILED: "+format, a...)
	}

	if euro, err := readRefdata(dir, "EUROGOV-EUR"); err != nil {
		log.Printf("check: EUROGOV-EUR unavailable (%v), skipping the euro-area cross-check", err)
	} else {
		// Mirror of the check gen-euro-refdata runs on the euro-area side, and
		// windowed the same way and for the same reason: the level over the whole
		// thin-spread decade, the monthly correlation only from 2005, the year
		// from which both files are on their real month-end curve (this one's
		// from 1997-08, the euro-area one's from 2004-09). The bar is 0.90
		// because that is what two unsmoothed reconstructions of two different
		// sovereign markets actually share (measured 0.94, sovereign crisis
		// included); two month-average ones used to read higher only because they
		// shared their averaging.
		from, to := date(1999, 1), date(2010, 1)
		cb, cg := cagr(bund, from, to), cagr(euro, from, to)
		corr := monthlyCorr(bund, euro, date(2005, 1), to)
		log.Printf("check BUND-EUR vs EUROGOV-EUR: CAGR 1999-2010 %.2f%% vs %.2f%% (gap %+.2f), monthly corr 2005-2010 %.3f",
			cb*100, cg*100, (cb-cg)*100, corr)
		if corr < 0.90 || math.Abs(cb-cg) > 0.01 {
			fail("the German and euro-area reconstructions diverge over the years their spreads were thin")
		}
	}

	from, to := date(1998, 1), bundDaily.Last().Date
	vd, vm := vol(bundDaily, from, to), vol(bundSynth, from, to)
	log.Printf("check BUND daily curve vs OECD tail 1998-2026: vol %.2f%% vs %.2f%%/yr (ratio %.2f), monthly corr %.3f (not gated, monthly-average yield)",
		vd*100, vm*100, vd/vm, monthlyCorr(bundSynth, bundDaily, from, to))
	if vd/vm < 0.8 || vd/vm > 1.35 {
		fail("the daily Bundesbank curve and the monthly OECD yield do not describe the same bond")
	}

	worst, at := 0.0, time.Time{}
	curveEnds := make(map[string]float64, len(bundDaily.Points)/20)
	for _, p := range monthEnds(bundDaily.Points) {
		curveEnds[p.Date.Format("2006-01-02")] = p.Close
	}
	months := 0
	for _, p := range bund.Points {
		if p.Date.Before(splice.at) {
			continue
		}
		months++
		r, ok := curveEnds[p.Date.Format("2006-01-02")]
		if !ok {
			fail("BUND-EUR carries %s, which is not a month-end of the Bundesbank curve", p.Date.Format("2006-01-02"))
			continue
		}
		if d := math.Abs(p.Close/(r*splice.factor) - 1); d > worst {
			worst, at = d, p.Date
		}
	}
	step, when := longestStep(bund.Points)
	log.Printf("check BUND-EUR splice: %d real month-ends from %s (worst deviation from the rebased curve %.1e, on %s), real era rebased x%.4f, junction return %+.2f%%, longest monthly step %.0f days ending %s",
		months, splice.at.Format("2006-01-02"), worst, at.Format("2006-01-02"), splice.factor, splice.seam*100, step, when.Format("2006-01-02"))
	if months < 12 || worst > 1e-9 {
		fail("BUND-EUR does not reproduce the Bundesbank curve it is sampled from")
	}
	if splice.factor <= 0 || math.Abs(splice.seam) > 0.15 {
		fail("the BUND-EUR splice at %s is not continuous (factor %.4f, junction return %+.2f%%)", splice.at.Format("2006-01"), splice.factor, splice.seam*100)
	}
	if step > 45 {
		fail("BUND-EUR takes a %.0f-day step ending %s: one row carries two months of return", step, when.Format("2006-01-02"))
	}
	if gstep, gwhen := longestStep(gilt.Points); gstep > 45 {
		fail("GILT-GBP takes a %.0f-day step ending %s: one row carries two months of return", gstep, gwhen.Format("2006-01-02"))
	}

	from, to = date(2016, 1), date(2021, 1)
	cj, vj := cagr(jgb, from, to), vol(jgb, from, to)
	log.Printf("check JGB-JPY 2016-2021 (yield-curve control): CAGR %+.2f%%/yr, vol %.2f%%/yr", cj*100, vj*100)
	if vj > 0.03 || vj < 0.005 || math.Abs(cj) > 0.03 {
		fail("the pinned-yield era of the JGB is not quiet and flat")
	}

	from, to = date(1970, 1), date(1980, 1)
	cg, vg := cagr(gilt, from, to), vol(gilt, from, to)
	log.Printf("check GILT-GBP 1970-1980: CAGR %+.2f%%/yr, vol %.2f%%/yr", cg*100, vg*100)
	if cg < 0 || cg > 0.15 || vg < 0.04 {
		fail("the 1970s gilt market does not look like the 1970s gilt market")
	}

	for _, c := range []struct {
		id  string
		pts []marketdata.Point
	}{{"JPCASH-JPY", jpCash}, {"GBCASH-GBP", gbCash}} {
		s := &marketdata.Series{Points: c.pts}
		rate, dd := cagr(s, s.First().Date, s.Last().Date), drawdown(c.pts)
		log.Printf("check %s: %.2f%%/yr over %s..%s, deepest drawdown %.2f%%", c.id, rate*100,
			s.First().Date.Format("2006-01"), s.Last().Date.Format("2006-01"), dd*100)
		if rate < 0 || rate > 0.12 || dd < -0.02 {
			fail("%s is not a plausible money-market accrual", c.id)
		}
	}

	if failed > 0 {
		log.Fatalf("%d sanity check(s) failed, nothing written", failed)
	}
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
// fraction. A money-market accrual is nearly monotone, but not exactly: a
// central bank that sets a negative policy rate makes cash shrink, as the Bank
// of Japan did from 2016 to 2024, so the check is a shallow floor rather than
// strict monotonicity.
func drawdown(pts []marketdata.Point) float64 {
	peak, worst := math.Inf(-1), 0.0
	for _, p := range pts {
		peak = max(peak, p.Close)
		worst = min(worst, p.Close/peak-1)
	}
	return worst
}
