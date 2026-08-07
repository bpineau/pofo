// Command gen-sgtrend-refdata rebuilds the two DAILY net managed-futures
// references this repository bundles from one publisher:
//
//   - pkg/datasets/refdata/TREND-PURE-NET-USD.csv, the PURE-TREND composite the
//     trend OVERLAY reconstructions (RSST, RSBT and the Winton trend sleeve)
//     take their monthly path and their level from, and the pre-inception donor
//     of the fund that names it as its benchmark (Simplify CTA);
//   - pkg/datasets/refdata/TREND-ALLSTYLES-NET-USD.csv, the ALL-STYLES
//     composite, which is the pre-inception donor of the DBi family: those
//     funds replicate this very index, and a fund's own index tracks it better
//     than any other manager's single fund does (measured, see
//     docs/trend-reconstruction-design.md).
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSVs and never fetches anything.
//
// # Why this exists
//
// The overlay builds used to anchor on the gross academic time-series-momentum
// factor and then force its level down with a constant daily drag calibrated to
// an information ratio. A drag reproduces an index's information ratio but not
// its drawdown: it turns a long trend drought into a long bleed. The
// diversified managed-futures funds escaped that in 2026-08 by anchoring on a
// record that is already investable (cmd/gen-trendnet-refdata); the overlays
// could not, because the pure-trend index they replicate was only public for
// five years. It is not: the whole daily history is served, free and at full
// precision, by the index's calculation agent. The same endpoint serves the
// publisher's all-styles index on the same terms, and the fund reconstructions
// use both.
//
// # The two series
//
// The SG Trend Index (NEIXCTAT) and the SG CTA Index (NEIXCTA), both published
// by SG Prime Services, both daily since 2000-01-03, both USD, both FUNDED
// total returns (the constituent programmes earn cash on their collateral).
// Each is an equally weighted composite reconstituted and rebalanced every
// January, each constituent entering NET of its own manager's fees, the index
// itself levying none. They differ in what they admit: the trend index takes
// the ten largest programmes that follow trends and nothing else, the CTA index
// takes the twenty largest managed-futures programmes whatever they trade. The
// trend index is the one the Return Stacked funds name on their trend sleeve;
// the CTA index is the one the DBi replication funds set out to reproduce.
//
// # Two channels, and why both are fetched
//
// The full-precision channel is the calculation agent's own daily dump, a
// legacy BIFF8 workbook (Dstamp as an Excel serial, ROR and VAMI at six
// decimals, VAMI base 1000 on 1999-12-31), obtained by one POST per index:
//
//	https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi
//	dump=excelDaily  prog_cod=FT90004127 (trend) or calyon (CTA)
//	return_option=since_inception
//
// The second channel is the publisher's own dashboard file, plain CSV with the
// daily levels of both indices rounded to two decimals:
//
//	https://wholesale.banking.societegenerale.com/fileadmin/indices_feeds/ti_screen/data/4.nav.csv
//
// Neither is a copy of the other, and a silent layout change in either would
// otherwise ship as data, so this generator downloads both and refuses to write
// a series unless every common daily return agrees to within that series' own
// tolerance over at least minCommonDays days, AND the disagreement compounded
// over the whole common window stays under its drift gate. The two tolerances
// are not the same, and the difference is a finding rather than a convenience:
// see the series table.
//
// # What was checked before this was trusted
//
// The trend index's calendar years were reconciled against six independent
// publications of them (the two channels above, the publisher's index website,
// its historical spreadsheet, its monthly report PDFs, and two archived
// captures of the calculation agent's own pages, the oldest from 2010). They
// agree everywhere; that index has never been restated by more than 5 bp, on
// 2018 alone. Over 2000-2026 it realizes about 13 % annualized volatility for a
// funded excess-over-T-bill information ratio near 0.27, which is the level a
// real trend programme delivers and the whole reason the overlays anchor on it.
// The CTA index realizes about 8 % for a near-identical ratio, and its two
// channels agree less tightly (see allStyles).
//
// # A note on what they are, and are not, licensed for
//
// The publisher's methodology carries an EU Benchmarks Regulation disclaimer:
// the indices are not to be used as a benchmark by financial products. pofo
// bundles them as reference series for a research reconstruction of fund
// histories, which is not that use.
//
// Usage: gen-sgtrend-refdata [-src URL] [-check URL] [-dir path] [-dry] [-only ID]
package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/cmd/internal/xls"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

const (
	// defaultSrc answers the POST below with the legacy workbook.
	defaultSrc = "https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi"
	// defaultCheck is the publisher's own dashboard copy, the second channel.
	defaultCheck = "https://wholesale.banking.societegenerale.com/fileadmin/indices_feeds/ti_screen/data/4.nav.csv"
	// browserUA is what the calculation agent's gateway expects; a bare Go
	// user agent is answered with an error page.
	browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// cashID is the bundled T-bill rate the funded excess is measured over.
	cashID = "TBILL-3M"

	// firstMonth is the inception month both indices are known to have:
	// anything later means a truncated dump, anything earlier the wrong series.
	firstMonth = "2000-01"
	// minCommonDays is the overlap the two channels must share before their
	// agreement means anything. Each full history holds about 6900 days.
	minCommonDays = 6500
)

// series is one index this generator ships: where each channel serves it, what
// the CSV says it is, and the gates it has to pass. The gates are per series
// because the two indices are not equally well behaved, and pretending
// otherwise would either wave one through or block the other.
type series struct {
	outID       string // the neutral bundled identifier
	progCode    string // the index's code in the calculation agent's catalog
	checkColumn string // the dashboard column holding the same index
	name        string // the CSV's name header
	source      string // the CSV's provenance header

	// maxReturnGap is how far a common DAILY return may differ between the two
	// channels, and maxDrift how far their disagreement may compound over the
	// whole common window. The first catches a layout change or a wrong column,
	// the second a slow restatement neither day would reveal on its own.
	maxReturnGap float64
	maxDrift     float64

	// The realized character the composite must have, on daily volatility and
	// on the funded excess-over-T-bill information ratio. Outside these bands
	// the file is not the index it claims to be.
	minVol, maxVol float64
	minIR, maxIR   float64
}

// shipped is what this generator writes, in output order.
var shipped = []series{pureTrend, allStyles}

// pureTrend is the composite of the ten largest trend-following programmes.
// Its two channels are as close as two independent publications of one index
// ever get: every one of ~6920 common daily returns agrees to within 2 bp,
// worst 1.16 bp, mean 0.16 bp, which is what rounding a level to two decimals
// costs and nothing more.
var pureTrend = series{
	outID:        "TREND-PURE-NET-USD",
	progCode:     "FT90004127",
	checkColumn:  "SG Trend Index",
	name:         "pure trend following, daily net composite index (USD)",
	source:       pureSource,
	maxReturnGap: 2e-4,
	maxDrift:     2e-3,
	minVol:       0.08,
	maxVol:       0.20,
	minIR:        0.10,
	maxIR:        0.60,
}

// allStyles is the broader composite: the largest managed-futures programmes
// whatever they trade, which is the index the DBi replication funds set out to
// reproduce and therefore their donor.
//
// Its channels agree less tightly than the trend index's, and the tolerances
// say so rather than hide it. Measured 2026-08 over 6923 common days: five days
// past 2 bp, four of them in the unrevised live tail (worst 15.6 bp) and one in
// 2024; and every month of calendar 2024 differs by 1 to 6 bp, 25 bp compounded
// over that year, which is a restatement in one channel and not rounding. Over
// the whole 2000-2026 window the two still compound to within 23 bp of each
// other, so the drift gate is the binding one and it is set at ten times what
// is measured, where the per-day gate is set to let a revising tail through.
var allStyles = series{
	outID:        "TREND-ALLSTYLES-NET-USD",
	progCode:     "calyon",
	checkColumn:  "SG CTA Index",
	name:         "all-styles managed futures, daily net composite index (USD)",
	source:       allStylesSource,
	maxReturnGap: 2e-3,
	maxDrift:     2e-2,
	minVol:       0.05,
	maxVol:       0.13,
	minIR:        0.10,
	maxIR:        0.60,
}

func main() {
	src := flag.String("src", defaultSrc, "daily index dump (POST endpoint)")
	check := flag.String("check", defaultCheck, "second channel, cross-checked against the first")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory")
	dry := flag.Bool("dry", false, "report without writing")
	only := flag.String("only", "", "rebuild this bundled id only (default: all)")
	flag.Parse()

	other, err := download(*check)
	if err != nil {
		log.Fatalf("cross-check download: %v", err)
	}
	cash, err := cashRates()
	if err != nil {
		log.Fatalf("cash reference %s: %v", cashID, err)
	}
	for _, s := range shipped {
		if *only != "" && *only != s.outID {
			continue
		}
		if err := build(s, *src, other, cash, *dir, *dry); err != nil {
			log.Fatalf("%s: %v", s.outID, err)
		}
	}
	if !*dry {
		log.Printf("rebuild (make simdata) to re-anchor the trend overlays and the fund donor chains")
	}
}

// build fetches one index from the calculation agent, grades it against the
// dashboard copy already downloaded, and writes it.
func build(s series, src string, dashboard []byte, cash []point, dir string, dry bool) error {
	raw, err := post(src, url.Values{
		"dump":          {"excelDaily"},
		"prog_cod":      {s.progCode},
		"return_option": {"since_inception"},
	})
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	cells, err := xls.Sheet(raw)
	if err != nil {
		return fmt.Errorf("read workbook: %w", err)
	}
	days, err := parseDaily(cells)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	log.Printf("%s: full-precision channel: %d days, %s to %s", s.outID, len(days),
		days[0].date.Format("2006-01-02"), days[len(days)-1].date.Format("2006-01-02"))

	second, err := parseDashboard(dashboard, s.checkColumn)
	if err != nil {
		return fmt.Errorf("cross-check parse: %w", err)
	}
	log.Printf("%s: cross-check channel: %d days, %s to %s", s.outID, len(second),
		second[0].date.Format("2006-01-02"), second[len(second)-1].date.Format("2006-01-02"))
	agreement, err := crossCheck(s, days, second)
	if err != nil {
		return fmt.Errorf("the two channels disagree: %w", err)
	}
	log.Printf("%s: %s", s.outID, agreement)

	kept, dropped := trimPartialMonth(days)
	if len(dropped) > 0 {
		log.Printf("%s: dropped %d days of the incomplete month %s",
			s.outID, len(dropped), dropped[0].date.Format("2006-01"))
	}
	days = kept

	if err := sanity(s, days, cash); err != nil {
		return fmt.Errorf("sanity: %w", err)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# pofo simdata v1\n# id: %s\n", s.outID)
	fmt.Fprintf(&b, "# name: %s\n", s.name)
	fmt.Fprintf(&b, "# source: %s\n", s.source)
	fmt.Fprintf(&b, "date,close\n")
	for _, p := range days {
		fmt.Fprintf(&b, "%s,%.6f\n", p.date.Format("2006-01-02"), p.level)
	}
	if dry {
		log.Printf("%s: dry run: %d bytes, %d points, final level %.2f",
			s.outID, b.Len(), len(days), days[len(days)-1].level)
		return nil
	}
	out := filepath.Join(dir, s.outID+".csv")
	if err := os.WriteFile(out, b.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	log.Printf("wrote %s (%d points)", out, len(days))
	return nil
}

// pureSource and allStylesSource are the provenance each CSV carries: what the
// series is, where both channels come from, what was checked, and the one
// licence restriction the publisher attaches to them.
const pureSource = "daily levels of the SG Trend Index (NEIXCTAT, SG Prime Services), an equally weighted composite of " +
	"the ten largest trend-following CTA programmes open to new investment, reconstituted each January, each " +
	"constituent NET of its own manager's fees and the index levying none; a funded total return, published as a " +
	"VAMI with base 1000 on 1999-12-31 and shipped here as published. Fetched at full precision from the index's " +
	"calculation agent (POST https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi with dump=excelDaily, " +
	"prog_cod=FT90004127, return_option=since_inception) and cross-checked day by day against the publisher's own " +
	"dashboard copy (https://wholesale.banking.societegenerale.com/fileadmin/indices_feeds/ti_screen/data/4.nav.csv), " +
	"which agrees on every common daily return to within 2 bp. The calendar years were reconciled against six " +
	"independent publications of them, the oldest an archived 2010 capture; the index has never been restated by " +
	"more than 5 bp, on 2018 alone. The publisher's methodology carries an EU Benchmarks Regulation disclaimer (not " +
	"to be used as a benchmark by financial products); it is bundled here as a reference series for a research " +
	"reconstruction of fund histories. Regenerate with cmd/gen-sgtrend-refdata."

const allStylesSource = "daily levels of the SG CTA Index (NEIXCTA, SG Prime Services), an equally weighted composite " +
	"of the twenty largest managed-futures programmes open to new investment whatever they trade, reconstituted " +
	"each January, each constituent NET of its own manager's fees and the index levying none; a funded total " +
	"return, published as a VAMI with base 1000 on 1999-12-31 and shipped here as published. Fetched at full " +
	"precision from the index's calculation agent (POST " +
	"https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi with dump=excelDaily, prog_cod=calyon, " +
	"return_option=since_inception) and cross-checked day by day against the publisher's own dashboard copy " +
	"(https://wholesale.banking.societegenerale.com/fileadmin/indices_feeds/ti_screen/data/4.nav.csv). This index " +
	"revises where its pure-trend sibling does not: measured 2026-08 over 6923 common days, five days differ by " +
	"more than 2 bp (four in the unrevised live tail, worst 15.6 bp) and every month of calendar 2024 differs by " +
	"1 to 6 bp, 25 bp compounded over that year. Over the whole window the two channels still compound to within " +
	"23 bp of each other. The publisher's methodology carries an EU Benchmarks Regulation disclaimer (not to be " +
	"used as a benchmark by financial products); it is bundled here as a reference series for a research " +
	"reconstruction of fund histories. Regenerate with cmd/gen-sgtrend-refdata."

// point is one day of the index: a date and the level it closed at.
type point struct {
	date  time.Time
	level float64
}

func post(endpoint string, form url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return send(req)
}

func download(endpoint string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	return send(req)
}

func send(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
}

// parseDaily reads the daily dump's single block: a title row, a header row
// naming the columns (Dstamp, ROR, VAMI, then period-to-date columns this
// generator ignores), then one row per index day. The date arrives as an Excel
// serial and the level as the VAMI. The header is read rather than assumed, so
// a column inserted upstream moves the reader with it instead of silently
// shifting the data by one.
func parseDaily(cells []xls.Cell) ([]point, error) {
	type key struct{ row, col int }
	grid := make(map[key]xls.Cell, len(cells))
	header, lastRow := -1, 0
	dateCol, levelCol := -1, -1
	for _, c := range cells {
		grid[key{c.Row, c.Col}] = c
		lastRow = max(lastRow, c.Row)
		if !c.IsText {
			continue
		}
		switch strings.ToUpper(strings.TrimSpace(c.Text)) {
		case "DSTAMP":
			header, dateCol = c.Row, c.Col
		case "VAMI":
			levelCol = c.Col
		}
	}
	if header < 0 || levelCol < 0 {
		return nil, fmt.Errorf("no Dstamp/VAMI header row in the sheet")
	}

	var out []point
	seen := make(map[time.Time]bool)
	for row := header + 1; row <= lastRow; row++ {
		d, ok := grid[key{row, dateCol}]
		if !ok || d.IsText {
			continue // a blank or annotated row: the block is not always dense
		}
		v, ok := grid[key{row, levelCol}]
		if !ok || v.IsText || v.Num <= 0 {
			return nil, fmt.Errorf("row %d: no usable level beside date serial %.0f", row, d.Num)
		}
		date := xls.SerialDate(d.Num)
		if seen[date] {
			return nil, fmt.Errorf("%s appears twice", date.Format("2006-01-02"))
		}
		seen[date] = true
		out = append(out, point{date, v.Num})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no rows under the header")
	}
	slices.SortFunc(out, func(a, b point) int { return a.date.Compare(b.date) })
	return out, nil
}

// parseDashboard reads the publisher's dashboard CSV and keeps one named
// column: a date column first, then one column per series it publishes.
func parseDashboard(raw []byte, column string) ([]point, error) {
	rows, err := csv.NewReader(bytes.NewReader(raw)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("%d rows, want a header and some data", len(rows))
	}
	col := slices.Index(rows[0], column)
	if col < 0 {
		return nil, fmt.Errorf("no %q column (found %s)", column, strings.Join(rows[0], ", "))
	}
	var out []point
	seen := make(map[time.Time]bool)
	for _, r := range rows[1:] {
		if len(r) <= col {
			continue
		}
		date, err := time.Parse("2006-01-02", strings.TrimSpace(r[0]))
		if err != nil {
			continue // a footer line, not a quote
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(r[col]), 64)
		if err != nil || v <= 0 {
			continue
		}
		if seen[date] {
			return nil, fmt.Errorf("%s appears twice", date.Format("2006-01-02"))
		}
		seen[date] = true
		out = append(out, point{date, v})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no usable rows in the %q column", column)
	}
	slices.SortFunc(out, func(a, b point) int { return a.date.Compare(b.date) })
	return out, nil
}

// crossCheck compares the two channels return by return over the days they
// share. Levels are not comparable (the two are published on different bases,
// and one is rounded), returns are, and a return is also what the reference is
// consumed for. It reports the agreement it found, or refuses it.
func crossCheck(s series, a, b []point) (string, error) {
	byDate := make(map[time.Time]float64, len(b))
	for _, p := range b {
		byDate[p.date] = p.level
	}
	var common []time.Time
	for _, p := range a {
		if _, ok := byDate[p.date]; ok {
			common = append(common, p.date)
		}
	}
	if len(common) < minCommonDays {
		return "", fmt.Errorf("only %d common days, want at least %d", len(common), minCommonDays)
	}
	levelA := make(map[time.Time]float64, len(a))
	for _, p := range a {
		levelA[p.date] = p.level
	}
	worst, worstDay, over := 0.0, common[0], 0
	drift := 1.0
	for i := 1; i < len(common); i++ {
		prev, cur := common[i-1], common[i]
		ra := levelA[cur]/levelA[prev] - 1
		rb := byDate[cur]/byDate[prev] - 1
		drift *= (1 + ra) / (1 + rb)
		gap := math.Abs(ra - rb)
		if gap > s.maxReturnGap {
			over++
		}
		if gap > worst {
			worst, worstDay = gap, cur
		}
	}
	if worst > s.maxReturnGap {
		return "", fmt.Errorf("%s: daily returns differ by %.1f bp, over the %.0f bp allowed (%d such days)",
			worstDay.Format("2006-01-02"), worst*1e4, s.maxReturnGap*1e4, over)
	}
	if d := math.Abs(drift - 1); d > s.maxDrift {
		return "", fmt.Errorf("the channels compound %.2f%% apart over %d days, over the %.1f%% allowed",
			d*100, len(common), s.maxDrift*100)
	}
	return fmt.Sprintf("the two channels agree over %d common days (%s to %s): "+
		"worst daily-return gap %.2f bp, on %s; compounded %.2f%% apart", len(common),
		common[0].Format("2006-01-02"), common[len(common)-1].Format("2006-01-02"),
		worst*1e4, worstDay.Format("2006-01-02"), (drift-1)*100), nil
}

// trimPartialMonth drops the tail of a month the index has not finished. A
// month is finished when no weekday is left in it after the last observation,
// which errs on the side of dropping a complete month whose last trading day
// was a holiday: one month at the end of a twenty-six-year file, and only for
// as long as the next one has not started.
func trimPartialMonth(days []point) (kept, dropped []point) {
	last := days[len(days)-1].date
	monthEnd := time.Date(last.Year(), last.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	complete := true
	for d := last.AddDate(0, 0, 1); !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			complete = false
			break
		}
	}
	if complete {
		return days, nil
	}
	cut := len(days)
	for cut > 0 && sameMonth(days[cut-1].date, last) {
		cut--
	}
	return days[:cut], days[cut:]
}

func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}

// sanity refuses to ship a series that is not the index it claims to be: an
// unbroken run of months from the inception both indices are known to have, at
// the volatility this composite realizes, earning what a real managed-futures
// programme earns over cash. Anything outside these bands means the wrong
// column, the wrong index or a truncated download.
func sanity(s series, days []point, cash []point) error {
	if got := days[0].date.Format("2006-01"); got != firstMonth {
		return fmt.Errorf("starts %s, want %s", got, firstMonth)
	}
	months := monthEnds(days)
	for i := 1; i < len(months); i++ {
		if gap := monthIndex(months[i].date) - monthIndex(months[i-1].date); gap != 1 {
			return fmt.Errorf("%s follows %s (%d months apart, want 1)",
				months[i].date.Format("2006-01"), months[i-1].date.Format("2006-01"), gap)
		}
	}
	if lag := monthIndex(time.Now().UTC()) - monthIndex(months[len(months)-1].date); lag > 3 {
		return fmt.Errorf("the last complete month is %s, %d months behind today",
			months[len(months)-1].date.Format("2006-01"), lag)
	}

	vol := annualVol(returns(days), 252)
	excess := excessReturns(months, cash)
	ir := annualGeoMean(excess, 12) / annualVol(excess, 12)
	log.Printf("%s: %d days over %d months, %s to %s, CAGR %.2f%%/yr, volatility %.1f%%/yr "+
		"(monthly %.1f%%), funded excess information ratio %.2f, max drawdown %.1f%%",
		s.outID, len(days), len(months), days[0].date.Format("2006-01-02"),
		days[len(days)-1].date.Format("2006-01-02"), cagr(days)*100, vol*100,
		annualVol(returns(months), 12)*100, ir, maxDrawdown(days)*100)
	if vol < s.minVol || vol > s.maxVol {
		return fmt.Errorf("volatility %.1f%%/yr outside [%.0f, %.0f]", vol*100, s.minVol*100, s.maxVol*100)
	}
	if ir < s.minIR || ir > s.maxIR {
		return fmt.Errorf("funded excess information ratio %.2f outside [%.2f, %.2f]", ir, s.minIR, s.maxIR)
	}
	return nil
}

// cashRates reads the bundled T-bill rate (annualized percent levels, monthly)
// the funded excess is measured over. It is embedded, so this gate needs no
// second network source of its own.
func cashRates() ([]point, error) {
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), cashID)
	if err != nil {
		return nil, err
	}
	if !ok || len(s.Points) < 12 {
		return nil, fmt.Errorf("not bundled, or too short")
	}
	out := make([]point, len(s.Points))
	for i, p := range s.Points {
		out[i] = point{p.Date, p.Close}
	}
	return out, nil
}

// monthEnds keeps the last observation of every calendar month.
func monthEnds(days []point) []point {
	var out []point
	for i, p := range days {
		if i+1 == len(days) || !sameMonth(days[i+1].date, p.date) {
			out = append(out, p)
		}
	}
	return out
}

func monthIndex(d time.Time) int { return d.Year()*12 + int(d.Month()) }

// returns is the simple return between consecutive points.
func returns(ps []point) []float64 {
	out := make([]float64, 0, len(ps))
	for i := 1; i < len(ps); i++ {
		out = append(out, ps[i].level/ps[i-1].level-1)
	}
	return out
}

// excessReturns is the month-by-month return of the index over the T-bill it
// is funded at. The rate series is monthly and annualized in percent, so a
// month's cash return is a twelfth of it; the last known rate is carried
// forward, since a rate publication lags an index by a month or two.
func excessReturns(months []point, cash []point) []float64 {
	out := make([]float64, 0, len(months))
	for i := 1; i < len(months); i++ {
		r := months[i].level/months[i-1].level - 1
		c := rateAt(cash, months[i].date) / 100 / 12
		out = append(out, (1+r)/(1+c)-1)
	}
	return out
}

// rateAt is the last rate published at or before d.
func rateAt(cash []point, d time.Time) float64 {
	rate := cash[0].level
	for _, p := range cash {
		if p.date.After(d) {
			break
		}
		rate = p.level
	}
	return rate
}

// annualGeoMean is the compounded mean of per-period returns, annualized over
// periods per year.
func annualGeoMean(rets []float64, periods float64) float64 {
	g := 1.0
	for _, r := range rets {
		g *= 1 + r
	}
	return math.Pow(g, periods/float64(len(rets))) - 1
}

func annualVol(rets []float64, periods float64) float64 {
	if len(rets) < 2 {
		return 0
	}
	var m float64
	for _, r := range rets {
		m += r
	}
	m /= float64(len(rets))
	var s float64
	for _, r := range rets {
		s += (r - m) * (r - m)
	}
	return math.Sqrt(s/float64(len(rets)-1)) * math.Sqrt(periods)
}

func cagr(days []point) float64 {
	years := days[len(days)-1].date.Sub(days[0].date).Hours() / 24 / 365.25
	return math.Pow(days[len(days)-1].level/days[0].level, 1/years) - 1
}

// maxDrawdown is the worst peak-to-trough fall, logged so a regeneration that
// changes the index's character is visible.
func maxDrawdown(days []point) float64 {
	peak, worst := 0.0, 0.0
	for _, p := range days {
		peak = max(peak, p.level)
		worst = min(worst, p.level/peak-1)
	}
	return worst
}
