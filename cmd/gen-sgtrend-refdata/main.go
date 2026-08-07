// Command gen-sgtrend-refdata rebuilds
// pkg/datasets/refdata/TREND-PURE-NET-USD.csv, the DAILY net pure-trend
// reference the trend OVERLAY reconstructions (RSST, RSBT and the Winton trend
// sleeve) take their monthly path and their level from.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSV and never fetches anything.
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
// precision, by the index's calculation agent.
//
// # The series
//
// The SG Trend Index (NEIXCTAT), published by SG Prime Services. It is an
// equally weighted composite of the ten largest trend-following CTA programmes
// open to new investment, reconstituted and rebalanced each January, each
// constituent entering NET of its own manager's fees; the index itself levies
// none. It is a FUNDED total return (the programmes earn cash on their
// collateral), daily since 2000-01-03, and it is the very index the Return
// Stacked funds name on their trend sleeve.
//
// # Two channels, and why both are fetched
//
// The full-precision channel is the calculation agent's own daily dump, a
// legacy BIFF8 workbook (Dstamp as an Excel serial, ROR and VAMI at six
// decimals, VAMI base 1000 on 1999-12-31), obtained by one POST:
//
//	https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi
//	dump=excelDaily  prog_cod=FT90004127  return_option=since_inception
//
// The second channel is the publisher's own dashboard file, plain CSV with the
// daily levels rounded to two decimals:
//
//	https://wholesale.banking.societegenerale.com/fileadmin/indices_feeds/ti_screen/data/4.nav.csv
//
// Neither is a copy of the other, and a silent layout change in either would
// otherwise ship as data, so this generator downloads both and refuses to write
// unless every common daily return agrees to within two basis points over at
// least minCommonDays days.
//
// # What was checked before this was trusted
//
// The index's calendar years were reconciled against six independent
// publications of them (the two channels above, the publisher's index website,
// its historical spreadsheet, its monthly report PDFs, and two archived
// captures of the calculation agent's own pages, the oldest from 2010). They
// agree everywhere; the index has never been restated by more than 5 bp, on
// 2018 alone. Over 2000-2026 it realizes about 13 % annualized volatility for a
// funded excess-over-T-bill information ratio near 0.27, which is the level a
// real trend programme delivers and the whole reason the overlays now anchor
// on it.
//
// # A note on what it is, and is not, licensed for
//
// The publisher's methodology carries an EU Benchmarks Regulation disclaimer:
// the index is not to be used as a benchmark by financial products. pofo
// bundles it as a reference series for a research reconstruction of fund
// histories, which is not that use.
//
// Usage: gen-sgtrend-refdata [-src URL] [-check URL] [-dir path] [-dry]
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
	// progCode names the pure-trend index in the calculation agent's catalog.
	progCode = "FT90004127"
	// defaultCheck is the publisher's own dashboard copy, the second channel.
	defaultCheck = "https://wholesale.banking.societegenerale.com/fileadmin/indices_feeds/ti_screen/data/4.nav.csv"
	// checkColumn is the dashboard column holding this index (the file also
	// carries the publisher's broader CTA index, its transparent trend model
	// and an equity benchmark).
	checkColumn = "SG Trend Index"
	// browserUA is what the calculation agent's gateway expects; a bare Go
	// user agent is answered with an error page.
	browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// outID is the neutral bundled identifier: the reference is a published net
	// pure-trend composite, and the recipes name it as that and nothing else.
	outID = "TREND-PURE-NET-USD"
	// cashID is the bundled T-bill rate the funded excess is measured over.
	cashID = "TBILL-3M"
)

// The gates. Each one is a layout change, a truncated download or a wrong
// index caught before it becomes a bundled file.
const (
	// firstMonth is the index's inception month: anything later means a
	// truncated dump, anything earlier means the wrong series.
	firstMonth = "2000-01"
	// minCommonDays is the overlap the two channels must share before their
	// agreement means anything. The full history holds about 6900 days.
	minCommonDays = 6500
	// maxReturnGap is how far a common daily return may differ between the two
	// channels. The dashboard copy is rounded to two decimals on a level
	// between 100 and 450, which is worth about a basis point at the start of
	// the history and a fifth of one at its end.
	maxReturnGap = 2e-4
	// The realized character a pure-trend composite must have. Outside these
	// bands the file is not the index it claims to be.
	minVol = 0.08
	maxVol = 0.20
	minIR  = 0.10
	maxIR  = 0.60
)

func main() {
	src := flag.String("src", defaultSrc, "daily index dump (POST endpoint)")
	check := flag.String("check", defaultCheck, "second channel, cross-checked against the first")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	raw, err := post(*src, url.Values{
		"dump":          {"excelDaily"},
		"prog_cod":      {progCode},
		"return_option": {"since_inception"},
	})
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	cells, err := xls.Sheet(raw)
	if err != nil {
		log.Fatalf("read workbook: %v", err)
	}
	days, err := parseDaily(cells)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	log.Printf("full-precision channel: %d days, %s to %s", len(days),
		days[0].date.Format("2006-01-02"), days[len(days)-1].date.Format("2006-01-02"))

	other, err := download(*check)
	if err != nil {
		log.Fatalf("cross-check download: %v", err)
	}
	second, err := parseDashboard(other, checkColumn)
	if err != nil {
		log.Fatalf("cross-check parse: %v", err)
	}
	log.Printf("cross-check channel: %d days, %s to %s", len(second),
		second[0].date.Format("2006-01-02"), second[len(second)-1].date.Format("2006-01-02"))
	agreement, err := crossCheck(days, second)
	if err != nil {
		log.Fatalf("the two channels disagree: %v", err)
	}
	log.Printf("%s", agreement)

	kept, dropped := trimPartialMonth(days)
	if len(dropped) > 0 {
		log.Printf("dropped %d days of the incomplete month %s",
			len(dropped), dropped[0].date.Format("2006-01"))
	}
	days = kept

	cash, err := cashRates()
	if err != nil {
		log.Fatalf("cash reference %s: %v", cashID, err)
	}
	if err := sanity(days, cash); err != nil {
		log.Fatalf("sanity: %v", err)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# pofo simdata v1\n# id: %s\n", outID)
	fmt.Fprintf(&b, "# name: pure trend following, daily net composite index (USD)\n")
	fmt.Fprintf(&b, "# source: %s\n", sourceNote)
	fmt.Fprintf(&b, "date,close\n")
	for _, p := range days {
		fmt.Fprintf(&b, "%s,%.6f\n", p.date.Format("2006-01-02"), p.level)
	}
	if *dry {
		log.Printf("dry run: %d bytes, %d points, final level %.2f", b.Len(), len(days), days[len(days)-1].level)
		return
	}
	out := filepath.Join(*dir, outID+".csv")
	if err := os.WriteFile(out, b.Bytes(), 0o644); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d points)", out, len(days))
	log.Printf("rebuild (make simdata) to re-anchor the trend overlays")
}

// sourceNote is the provenance the CSV carries: what the series is, where both
// channels come from, what was checked, and the one licence restriction its
// publisher attaches to it.
const sourceNote = "daily levels of the SG Trend Index (NEIXCTAT, SG Prime Services), an equally weighted composite of " +
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
func crossCheck(a, b []point) (string, error) {
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
	worst, worstDay := 0.0, common[0]
	for i := 1; i < len(common); i++ {
		prev, cur := common[i-1], common[i]
		ra := levelA[cur]/levelA[prev] - 1
		rb := byDate[cur]/byDate[prev] - 1
		if gap := math.Abs(ra - rb); gap > worst {
			worst, worstDay = gap, cur
		}
	}
	if worst > maxReturnGap {
		return "", fmt.Errorf("%s: daily returns differ by %.1f bp, over the %.0f bp allowed",
			worstDay.Format("2006-01-02"), worst*1e4, maxReturnGap*1e4)
	}
	return fmt.Sprintf("the two channels agree over %d common days (%s to %s): "+
		"worst daily-return gap %.2f bp, on %s", len(common),
		common[0].Format("2006-01-02"), common[len(common)-1].Format("2006-01-02"),
		worst*1e4, worstDay.Format("2006-01-02")), nil
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
// unbroken run of months from the inception the index is known to have, at a
// pure-trend volatility, earning what a real trend programme earns over cash.
// Anything outside these bands means the wrong column, the wrong index or a
// truncated download.
func sanity(days []point, cash []point) error {
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
	log.Printf("%d days over %d months, %s to %s, CAGR %.2f%%/yr, volatility %.1f%%/yr "+
		"(monthly %.1f%%), funded excess information ratio %.2f, max drawdown %.1f%%",
		len(days), len(months), days[0].date.Format("2006-01-02"),
		days[len(days)-1].date.Format("2006-01-02"), cagr(days)*100, vol*100,
		annualVol(returns(months), 12)*100, ir, maxDrawdown(days)*100)
	if vol < minVol || vol > maxVol {
		return fmt.Errorf("volatility %.1f%%/yr outside [%.0f, %.0f]", vol*100, minVol*100, maxVol*100)
	}
	if ir < minIR || ir > maxIR {
		return fmt.Errorf("funded excess information ratio %.2f outside [%.2f, %.2f]", ir, minIR, maxIR)
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
