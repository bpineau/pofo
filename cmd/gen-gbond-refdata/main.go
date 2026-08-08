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
//     monthly, 1956-05→): the OECD long-term government bond
//     yield for Germany run through the constant-maturity
//     reconstruction simgen.TreasuryTR. The euro-area
//     aggregate EUROGOV-EUR is deliberately NOT reused: it
//     smears the periphery spreads of 2011-2012 that a Bund
//     basket never carried.
//   - BUND-DAILY.csv   the same at daily granularity (1997-08→): the
//     Bundesbank's own daily 10-year point of the listed
//     federal securities term structure (Svensson), through
//     the same TreasuryTR. Daily shape for BUND-EUR.
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
//     accepted there and documented in the recipe.
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

	// German 10-year, monthly (1956-05→) and daily (1997-08→).
	bundYield := fetch(*base, "OECD/DSD_STES@DF_FINMARK/DEU.M.IRLT.PA._Z._Z._Z._Z.N")
	bund := simgen.TreasuryTR("German government bond total return (10y benchmark, monthly)", asSeries(bundYield), bondMaturity, 0)
	report("BUND-EUR", bund.Points)

	bundDailyYield := fetch(*base, "BUBA/BBSIS/D.I.ZST.ZI.EUR.S1311.B.A604.R10XX.R.A.A._Z._Z.A")
	bundDaily := simgen.TreasuryTR("German government bond total return (10y benchmark, daily)", asSeries(bundDailyYield), bondMaturity, 0)
	report("BUND-DAILY", bundDaily.Points)

	// Japanese 10-year, daily (1986-07→), and the yen call-money accrual.
	jgbYield := fetchJGB(*jgbURL)
	jgb := simgen.TreasuryTR("Japanese government bond total return (10y benchmark, daily)", jgbYield, bondMaturity, 0)
	report("JGB-JPY", jgb.Points)

	jpRate := fetch(*base, "OECD/DSD_STES@DF_FINMARK/JPN.M.IRSTCI.PA._Z._Z._Z._Z.N")
	jpCash := accrue(jpRate)
	report("JPCASH-JPY", jpCash)

	// British 10-year, monthly (1960-01→), and the sterling interbank accrual.
	giltYield := fetch(*base, "OECD/DSD_STES@DF_FINMARK/GBR.M.IRLT.PA._Z._Z._Z._Z.N")
	gilt := simgen.TreasuryTR("British government bond total return (10y benchmark, monthly)", asSeries(giltYield), bondMaturity, 0)
	report("GILT-GBP", gilt.Points)

	gbRate := fetch(*base, "OECD/DSD_STES@DF_FINMARK/GBR.M.IRSTCI.PA._Z._Z._Z._Z.N")
	gbCash := accrue(gbRate)
	report("GBCASH-GBP", gbCash)

	if *check {
		runChecks(*dir, bund, bundDaily, jgb, gilt, jpCash, gbCash)
	}
	if *dry {
		return
	}
	write(*dir, "BUND-EUR", "German government bond total return (10-year benchmark, EUR, monthly)",
		fmt.Sprintf("OECD long-term government bond yield DEU.M.IRLT (dataflow DSD_STES@DF_FINMARK, ~1956) run through the constant-maturity reconstruction (TreasuryTR, %.0fy par); via DBnomics. German leg of the NTSG global bond overlay; the euro-area aggregate EUROGOV-EUR is not reused, it carries periphery spreads a Bund basket never had.", bondMaturity), bund.Points)
	write(*dir, "BUND-DAILY", "German government bond total return (10-year benchmark, EUR, daily)",
		fmt.Sprintf("Bundesbank daily term structure of listed federal securities (Svensson), 10-year residual maturity, BBSIS D.I.ZST.ZI.EUR.S1311.B.A604.R10XX.R.A.A (~1997-08) run through TreasuryTR (%.0fy par); via DBnomics. Daily shape for BUND-EUR.", bondMaturity), bundDaily.Points)
	write(*dir, "JGB-JPY", "Japanese government bond total return (10-year benchmark, JPY, daily)",
		fmt.Sprintf("Japanese Ministry of Finance historical JGB interest rates (jgbcme_all.csv), %s column (~1986-07), run through TreasuryTR (%.0fy par). Japanese leg of the NTSG global bond overlay; the yield is negative on 453 days between 2016-02 and 2020-05 and the reconstruction prices those days rather than flat-lining them.", jgbTenor, bondMaturity), jgb.Points)
	write(*dir, "GILT-GBP", "British government bond total return (10-year benchmark, GBP, monthly)",
		fmt.Sprintf("OECD long-term government bond yield GBR.M.IRLT (dataflow DSD_STES@DF_FINMARK, ~1960) run through TreasuryTR (%.0fy par); via DBnomics. British leg of the NTSG global bond overlay; monthly texture accepted, the sleeve is ~1.8%% of net assets.", bondMaturity), gilt.Points)
	write(*dir, "JPCASH-JPY", "Japanese overnight money-market accrual (JPY, monthly)",
		"OECD immediate interest rate, call money JPN.M.IRSTCI (dataflow DSD_STES@DF_FINMARK, ~1985-07) compounded into a money-market level; via DBnomics. What the yen leg of the NTSG bond overlay finances at.", jpCash)
	write(*dir, "GBCASH-GBP", "British overnight money-market accrual (GBP, monthly)",
		"OECD immediate interest rate, interbank GBR.M.IRSTCI (dataflow DSD_STES@DF_FINMARK, ~1978-01) compounded into a money-market level; via DBnomics. What the sterling leg of the NTSG bond overlay finances at.", gbCash)
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

// accrue compounds a short-rate series (annualized percent) into a
// money-market level (base 100), pro rata temporis. Each step accrues at the
// rate that PREVAILED over it, i.e. the previous observation's level, so the
// index never earns a rate before it was published.
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
//   - German vs euro-area (1999-2010). Over the euro's first decade, before
//     the sovereign crisis reopened them, the spreads between the euro
//     aggregate and the Bund were a few tens of basis points. The two
//     reconstructions must therefore agree closely: monthly correlation above
//     0.95 and less than a point a year between their CAGRs. It is the one
//     check that grades the German series against something already bundled.
//   - German daily vs monthly (1998-2026). The daily Bundesbank curve and the
//     monthly OECD yield describe the same bond, so their monthly volatilities
//     must agree to within about a third. The daily one is expected to be the
//     higher of the two: the OECD publishes a monthly AVERAGE yield, which
//     damps whatever happened inside the month. What the check is really for is
//     a wrong curve point, since a 2-year or a 30-year residual maturity would
//     be off by a factor rather than by a fifth.
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
func runChecks(dir string, bund, bundDaily, jgb, gilt *marketdata.Series, jpCash, gbCash []marketdata.Point) {
	failed := 0
	fail := func(format string, a ...any) {
		failed++
		log.Printf("CHECK FAILED: "+format, a...)
	}

	if euro, err := readRefdata(dir, "EUROGOV-EUR"); err != nil {
		log.Printf("check: EUROGOV-EUR unavailable (%v), skipping the euro-area cross-check", err)
	} else {
		from, to := date(1999, 1), date(2010, 1)
		cb, cg := cagr(bund, from, to), cagr(euro, from, to)
		corr := monthlyCorr(bund, euro, from, to)
		log.Printf("check BUND-EUR vs EUROGOV-EUR 1999-2010: CAGR %.2f%% vs %.2f%% (gap %+.2f), monthly corr %.3f",
			cb*100, cg*100, (cb-cg)*100, corr)
		if corr < 0.95 || math.Abs(cb-cg) > 0.01 {
			fail("the German and euro-area reconstructions diverge over the years their spreads were thin")
		}
	}

	from, to := date(1998, 1), bundDaily.Last().Date
	vd, vm := vol(bundDaily, from, to), vol(bund, from, to)
	log.Printf("check BUND daily vs monthly 1998-2026: vol %.2f%% vs %.2f%%/yr (ratio %.2f)", vd*100, vm*100, vd/vm)
	if vd/vm < 0.8 || vd/vm > 1.35 {
		fail("the daily Bundesbank curve and the monthly OECD yield do not describe the same bond")
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
