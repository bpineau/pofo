// Command gen-trendnet-refdata rebuilds pkg/datasets/refdata/TREND-NET-USD.csv,
// the monthly NET reference path of managed futures that the deep tail of the
// trend reconstructions is anchored on.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSV and never fetches anything.
//
// WHY THIS EXISTS. The reconstructions already had a monthly reference, the
// gross academic time-series-momentum factor (cmd/gen-trend-refdata). It has
// the right SHAPE and the wrong LEVEL: it carries no fee, no slippage and no
// capacity limit, and realizes an information ratio near 0.97 where investable
// programmes deliver 0.25 to 0.5. The level was therefore forced down by a
// constant daily drag, which reproduces an index's information ratio but not
// its drawdown: pinning a gross path down by eleven points a year turns a long
// trend drought into a long bleed. The honest fix was always an anchor that is
// already investable, and this is it.
//
// THE SERIES. The Barclay BTOP50 Index, published monthly by BarclayHedge and
// downloadable as a legacy spreadsheet from
//
//	https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi?dump=excel&prog_cod=BTOP50
//
// It is a composite of REAL managed-futures programmes, selected each January
// as the largest investable ones that make up 50 % of the sector's assets,
// equally weighted and rebalanced annually, and each programme's return enters
// NET of that manager's own fees. So it is not a factor: it is what a book of
// real CTA programmes actually paid its investors, monthly from 1987-01.
//
// The last month or two of the live file are the source's own estimates and
// are dropped here rather than shipped. The CSV holds a month-end level index
// (base 100) of the surviving months, so simgen can rescale it to any
// volatility target.
//
// A NOTE ON WHAT IT IS NOT. The index is all-styles managed futures, not pure
// trend: it holds whatever the largest programmes trade. That is the right
// yardstick for the diversified funds reconstructed here (DBMF, KMLM, Simplify
// CTA, the AQR pair) and the wrong one for a pure-trend overlay, which keeps
// the gross factor and its own pin.
//
// Usage: gen-trendnet-refdata [-src URL] [-dir path] [-dry]
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const defaultSrc = "https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi?dump=excel&prog_cod=BTOP50"

// estimatedTail is how many months at the end of the live file the source
// still marks as estimated. They move afterwards, so they are not shipped.
const estimatedTail = 2

// minMonths is the shortest history worth shipping: the reconstruction it
// anchors starts in 1988, so anything under ~33 years means the download
// truncated or the layout moved.
const minMonths = 400

func main() {
	src := flag.String("src", defaultSrc, "monthly index spreadsheet")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	raw, err := download(*src)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	months, err := parse(raw)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	if len(months) <= estimatedTail {
		log.Fatalf("only %d monthly returns: the source layout probably moved", len(months))
	}
	dropped := months[len(months)-estimatedTail:]
	months = months[:len(months)-estimatedTail]
	log.Printf("dropped %d estimated months: %s", len(dropped), monthList(dropped))

	if err := check(months); err != nil {
		log.Fatalf("sanity: %v", err)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# pofo simdata v1\n# id: TREND-NET-USD\n")
	fmt.Fprintf(&b, "# name: managed futures, monthly net composite index (USD, month-end)\n")
	fmt.Fprintf(&b, "# source: monthly returns of the Barclay BTOP50 Index (BarclayHedge, "+
		"https://portal.barclayhedge.com/cgi-bin/barclay_stats/bcndx.cgi?dump=excel&prog_cod=BTOP50), the largest "+
		"investable managed-futures programmes making up 50%% of the sector's assets, equally weighted, rebalanced "+
		"annually, each NET of its own manager's fees, compounded into a base-100 index. The most recent months are "+
		"estimated at the source and are not shipped. Regenerate with cmd/gen-trendnet-refdata.\n")
	fmt.Fprintf(&b, "date,close\n")
	level := 100.0
	fmt.Fprintf(&b, "%s,%.6f\n", previousMonthEnd(months[0].date).Format("2006-01-02"), level)
	for _, p := range months {
		level *= 1 + p.ret
		fmt.Fprintf(&b, "%s,%.6f\n", p.date.Format("2006-01-02"), level)
	}
	if *dry {
		log.Printf("dry run: %d bytes, final level %.1f", b.Len(), level)
		return
	}
	out := filepath.Join(*dir, "TREND-NET-USD.csv")
	if err := os.WriteFile(out, b.Bytes(), 0o644); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d points)", out, len(months)+1)
	log.Printf("rebuild (make simdata) to re-anchor the trend reconstructions")
}

// point is one month of the index: a month-end date and the return over it.
type point struct {
	date time.Time
	ret  float64
}

func download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
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

// parse reads the sheet's rate-of-return block: a header row naming it, a row
// of month names, then one row per year holding the year in the first column
// and up to twelve monthly returns as fractions beside it. A month the index
// does not cover is simply absent. A second block below repeats the same
// history as a level index and is ignored: the two are checked against each
// other at the source, and returns are what this generator needs.
func parse(raw []byte) ([]point, error) {
	streams, err := ole2Streams(raw)
	if err != nil {
		return nil, err
	}
	wb, ok := streams["Workbook"]
	if !ok {
		wb, ok = streams["Book"]
	}
	if !ok {
		return nil, fmt.Errorf("no workbook stream (found %d others)", len(streams))
	}

	type key struct{ row, col int }
	grid := make(map[key]cell)
	header, lastRow := -1, 0
	for _, c := range biffCells(wb) {
		grid[key{c.row, c.col}] = c
		lastRow = max(lastRow, c.row)
		if header < 0 && c.isText && c.col < 14 && strings.Contains(c.text, "ROR") {
			header = c.row
		}
	}
	if header < 0 {
		return nil, fmt.Errorf("no rate-of-return block in the sheet")
	}

	var out []point
	for row := header + 2; row <= lastRow; row++ { // +2: the header, then the month names
		y, ok := grid[key{row, 0}]
		if !ok || y.isText || y.num < 1900 || y.num > 2200 {
			break // the block ends where the year column stops
		}
		for m := 1; m <= 12; m++ {
			c, ok := grid[key{row, m}]
			if !ok || c.isText {
				continue
			}
			if math.Abs(c.num) > 0.6 {
				return nil, fmt.Errorf("%d-%02d: %.4f is not a monthly return", int(y.num), m, c.num)
			}
			out = append(out, point{monthEndOf(int(y.num), time.Month(m)), c.num})
		}
	}
	slices.SortFunc(out, func(a, b point) int { return a.date.Compare(b.date) })
	return out, nil
}

// check refuses to ship a series that is not what the anchor expects: a long
// unbroken run of months, at a managed-futures volatility, earning a
// managed-futures information ratio over the whole period. Anything outside
// these bands means the wrong column, the wrong index, or a truncated file.
func check(months []point) error {
	if len(months) < minMonths {
		return fmt.Errorf("only %d months, want at least %d", len(months), minMonths)
	}
	for i := 1; i < len(months); i++ {
		if gap := monthIndex(months[i]) - monthIndex(months[i-1]); gap != 1 {
			return fmt.Errorf("%s follows %s (%d months apart, want 1)",
				months[i].date.Format("2006-01"), months[i-1].date.Format("2006-01"), gap)
		}
	}
	vol, ir := stats(months)
	log.Printf("%d months, %s to %s, vol %.1f%%/yr, information ratio %.2f, max drawdown %.1f%%",
		len(months), months[0].date.Format("2006-01"), months[len(months)-1].date.Format("2006-01"),
		vol*100, ir, maxDrawdown(months)*100)
	if vol < 0.05 || vol > 0.20 {
		return fmt.Errorf("volatility %.1f%%/yr outside [5, 20]", vol*100)
	}
	if ir < 0.3 || ir > 1.2 {
		return fmt.Errorf("information ratio %.2f outside [0.3, 1.2]", ir)
	}
	return nil
}

// monthIndex numbers months consecutively, so a gap or a duplicate is a
// subtraction away.
func monthIndex(p point) int {
	return p.date.Year()*12 + int(p.date.Month())
}

func monthList(ps []point) string {
	var names []string
	for _, p := range ps {
		names = append(names, p.date.Format("2006-01"))
	}
	return strings.Join(names, ", ")
}

// monthEndOf is the last day of the given month, at 00:00 UTC (day 0 of the
// next month, which time.Date normalizes for us).
func monthEndOf(year int, month time.Month) time.Time {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
}

// previousMonthEnd is the last day of the month before the one d falls in: the
// date the base-100 point carries, so the first return has a step to stand on.
func previousMonthEnd(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 0, 0, 0, 0, 0, time.UTC)
}

// stats returns the annualized volatility of the monthly returns and the
// information ratio the series earns over the whole period. The index is a
// total return, cash included, but over four decades its cash leg is small
// against its own variation, so this reads as an information ratio to within
// the band check accepts, which is all it is used for.
func stats(months []point) (vol, ir float64) {
	var m float64
	for _, p := range months {
		m += p.ret
	}
	m /= float64(len(months))
	var s float64
	for _, p := range months {
		s += (p.ret - m) * (p.ret - m)
	}
	vol = math.Sqrt(s/float64(len(months)-1)) * math.Sqrt(12)
	g := 1.0
	for _, p := range months {
		g *= 1 + p.ret
	}
	cagr := math.Pow(g, 12/float64(len(months))) - 1
	if vol > 0 {
		ir = cagr / vol
	}
	return vol, ir
}

// maxDrawdown is the worst peak-to-trough fall of the compounded series,
// logged so a regeneration that changes the index's character is visible.
func maxDrawdown(months []point) float64 {
	level, peak, worst := 1.0, 1.0, 0.0
	for _, p := range months {
		level *= 1 + p.ret
		peak = max(peak, level)
		worst = min(worst, level/peak-1)
	}
	return worst
}
