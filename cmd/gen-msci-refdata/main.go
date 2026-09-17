// Command gen-msci-refdata extends the three bundled MSCI monthly reference
// series past their manual export with a validated ETF proxy tail.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSVs and never fetches Yahoo.
//
// WHY THIS EXISTS. MSCIWORLD-USD, DEVEXUS-USD and EM-USD are month-end net
// total-return levels exported by hand from curvo.eu/backtest. They anchor the
// level of every world, developed-ex-US and emerging-market reconstruction in
// the bundle (MSCIWORLD, URTH, IWDA, WPEA, the VT/VTI legs), and having no
// generator they froze at the day of the last manual export: three months of
// staleness at the front of the longest series pofo ships. This generator
// closes that gap without touching a single exported point.
//
// THE TAIL POLICY. The exported point always wins where it exists. The
// generator reads the current CSV, keeps every anchor up to the month named by
// the "# tail-from:" header (the whole file when there is no such header, which
// is how the boundary is minted on the first run), throws away any tail a
// previous run appended, and rebuilds it from the proxy. The Curvo boundary is
// therefore frozen for good, the tail is always recomputed from the freshest
// quotes rather than accumulated, and a re-run with no new month is a no-op on
// the data.
//
// THE PROXY. Each series names one or more catalogued ETFs tracking its own
// index, in the currency of the series. Their adjusted closes are total
// returns already net of the fund's ongoing charge, so the charge is added back
// (the reference is an index, gross of any fund fee). Only complete months are
// appended: the proxy's own last month is dropped, since a month-end level
// struck mid-month is not a month-end level.
//
// VALIDATION, before anything is written and printed on every run:
//
//   - the proxy quotes in the series' currency, or the candidate is refused;
//   - at least minOverlapMonths of overlap with the anchors;
//   - annualized tracking difference over the whole overlap within maxTD,
//     after the fee add-back;
//   - monthly correlation with the anchors at or above minCorr;
//   - the proxy reaches a later month than the anchors;
//   - no appended monthly return beyond maxMonthly.
//
// Candidates are measured in turn and the smallest absolute tracking difference
// among those that pass wins. The measured figures are written into the file's
// "# tail-source:" header so the next reader knows what is export and what is
// proxy.
//
// Usage: gen-msci-refdata [-dir path] [-dry] [-only ID[,ID...]]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Validation bands, all measured before being set (2026-09-10, see
// docs/index-benchmarks-design.md):
//
// maxTD bounds the annualized tracking difference of a tracker against its
// own index once the ongoing charge is added back. A tracker of the RIGHT
// index still beats the NET index gross of fees, because the net index assumes
// the maximum statutory withholding on every dividend while a real fund
// reclaims at its domicile's treaty rates: measured +0.27 %/yr for a
// Dublin-domiciled World tracker (US treaty rate on 70 % of the index) and
// +0.41 for the US-listed sibling (no US withholding at all), both as domicile
// predicts. The band has to admit that, so 0.50; a tracker of a NEIGHBOURING
// index misses by whole points (ACWI ex USA against World ex USA, an IMI
// against a large-and-mid index) and is still caught.
//
// minCorr is a sanity floor on the monthly PATH. An ETF's month-end price is
// struck at its own exchange's close, on its own holiday calendar, while a
// global index is struck at each constituent market's local close: a London
// line tracking MSCI World therefore carries a month-boundary stub worth up to
// 1.5 points in a single month (August 2026, when the LSE was shut on the
// 31st), which caps its monthly correlation near 0.985. 0.98 accepts the stub
// and still catches what the gate is for, a stale, scale-broken or plainly
// wrong vendor series.
//
// SELECTION among the candidates that pass is by smallest monthly rmse, not by
// smallest tracking difference, because that stub is the larger error at the
// front edge of a series: three months of the +0.41 level bias cost 0.10 %,
// where one misaligned month-end costs ten times that until it reverses. The
// rmse is the one number that weighs both.
const (
	maxTD            = 0.50 // %/yr, absolute
	minCorr          = 0.98 // monthly returns, Pearson
	maxMonthly       = 0.25 // absolute monthly return of an appended point
	minOverlapMonths = 24
	gradeWindow      = 120 // months the bands are applied to, the recent overlap
)

// target is one bundled reference series and the proxies that may extend it.
type target struct {
	id       string
	name     string
	currency string
	source   string   // the export's own provenance, kept verbatim across runs
	proxies  []string // catalog ids of index trackers, network order
}

// The three Curvo exports. Each proxy tracks the SAME index as the series it
// extends: a tracker of a neighbouring index would pass the correlation gate
// and fail the tracking-difference one, which is the point of measuring both.
var targets = []target{{
	id:       "MSCIWORLD-USD",
	name:     "MSCI World net total return (USD, monthly month-end)",
	currency: "USD",
	source:   "MSCI via Curvo export (curvo.eu/backtest uses NET total-return indices), personal use; not for redistribution. Validated 2026-07-01: Dec2012->Dec2024 10.82%/yr and Dec2014->Dec2024 9.95%/yr match MSCI World NET USD exactly (gross was 11.41% / 10.52%). Net is the right proxy for an Irish-domiciled UCITS World ETF (IWDA/URTH track the net index). DATES: each point is dated the LAST calendar day of its month and holds that month-end level (relabeled from the first-of-month export 2026-07-18, so anchorShape pins it to the right month; values unchanged). Points from the month named by tail-from are NOT the export: they are the validated ETF proxy tail described by tail-source, rebuilt by cmd/gen-msci-refdata.",
	// IWDA is Irish-domiciled, the withholding treatment the MSCI NET index
	// assumes; URTH is the US-listed sibling, measured as the alternative.
	proxies: []string{"IE00B4L5Y983", "URTH"},
}, {
	id:       "DEVEXUS-USD",
	name:     "Developed markets ex-US equity total return (MSCI World ex USA, USD, monthly)",
	currency: "USD",
	source:   "MSCI World ex USA net total return via curvo.eu/backtest (MSCI-derived; curvo uses NET indices), USD, 1969-12->. True ex-US universe. Proxy behind VTMGX. Validated 2026-08-20: the Dec-to-Dec returns reproduce the published MSCI World ex USA NET USD calendar years 2012-2025 to the basis point (MSCI factsheet msci-world-ex-usa-index-net.pdf, Jul 31 2026). DATES: each point is dated the LAST calendar day of its month and holds that month-end level (relabeled from the first-of-month export 2026-08-20, so anchorShape pins it to the right month; values unchanged). Points from the month named by tail-from are NOT the export: they are the validated ETF proxy tail described by tail-source, rebuilt by cmd/gen-msci-refdata.",
	// The one catalogued tracker of MSCI World ex USA itself. ACWI ex USA
	// trackers (ACWX) and FTSE Developed ex-US ones (VEA) are NOT candidates:
	// they hold emerging markets or a different developed universe.
	proxies: []string{"IE0006WW1TQ4"},
}, {
	id:       "EM-USD",
	name:     "Emerging markets equity total return (MSCI Emerging Markets, USD, monthly)",
	currency: "USD",
	source:   "MSCI Emerging Markets net total return via curvo.eu/backtest (MSCI-derived; curvo uses NET indices), USD, index base 1987-12. Proxy behind VEIEX. Validated 2026-08-20: the Dec-to-Dec returns track the published MSCI Emerging Markets NET USD calendar years 2012-2025 within 0.5 point, running +0.31 point/yr rich on average (MSCI factsheet msci-emerging-markets-index-usd-net.pdf, Jul 31 2026); the export is MSCI-derived rather than the index itself, and its last months are chained from a tracking series. DATES: each point is dated the LAST calendar day of its month and holds that month-end level (relabeled from the first-of-month export 2026-08-20, so anchorShape pins it to the right month; values unchanged). Points from the month named by tail-from are NOT the export: they are the validated ETF proxy tail described by tail-source, rebuilt by cmd/gen-msci-refdata.",
	// XMME tracks MSCI EM (large and mid cap), the exported index. IMI
	// trackers (EIMI, IEMG) add small caps and are not candidates.
	proxies: []string{"IE00BTJRMP35"},
}}

func main() {
	dir := flag.String("dir", "pkg/datasets/refdata", "directory holding the reference CSVs")
	dry := flag.Bool("dry", false, "measure and report, but do not write")
	only := flag.String("only", "", "comma-separated series ids to rebuild (default all)")
	flag.Parse()

	ctx := context.Background()
	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	today := time.Now().UTC()
	failed := 0
	for _, t := range targets {
		if !selected(*only, t.id) {
			continue
		}
		if err := run(ctx, client, t, *dir, today, *dry); err != nil {
			log.Printf("%s: %v", t.id, err)
			failed++
		}
	}
	if failed > 0 {
		log.Fatalf("%d of %d series could not be extended", failed, len(targets))
	}
}

func selected(only, id string) bool {
	if only == "" {
		return true
	}
	for _, s := range strings.Split(only, ",") {
		if strings.EqualFold(strings.TrimSpace(s), id) {
			return true
		}
	}
	return false
}

// run extends one series: read, measure every candidate, append the best.
func run(ctx context.Context, client *marketdata.Client, t target, dir string, today time.Time, dry bool) error {
	path := filepath.Join(dir, t.id+".csv")
	file, err := readRefdata(path)
	if err != nil {
		return err
	}
	anchors := file.anchors
	log.Printf("%-13s %d export points %s..%s (tail-from %s)", t.id, len(anchors),
		month(anchors[0].Date), month(anchors[len(anchors)-1].Date), orDash(file.tailFrom))

	best, bestErr, candidates := -1, error(nil), 0
	var results []result
	for _, id := range t.proxies {
		candidates++
		r, err := measure(ctx, client, t, anchors, id)
		if err != nil {
			log.Printf("%-13s   %-14s REFUSED (%v)  %s", t.id, id, err, r)
			if bestErr == nil {
				bestErr = err
			}
			continue
		}
		log.Printf("%-13s   %-14s ok  %s", t.id, id, r)
		results = append(results, r)
		if best < 0 || r.gated.rmse < results[best].gated.rmse {
			best = len(results) - 1
		}
	}
	if best < 0 {
		return fmt.Errorf("no usable proxy: %w", bestErr)
	}
	r := results[best]
	if candidates > 1 {
		log.Printf("%-13s   picked %s (smallest monthly rmse of %d candidates)", t.id, r.proxy, candidates)
	}

	tail, err := buildTail(anchors, r.monthly, r.ter, maxMonthly)
	if err != nil {
		return err
	}
	log.Printf("%-13s   appending %d month(s) %s..%s, %s", t.id, len(tail),
		month(tail[0].Date), month(tail[len(tail)-1].Date), returnsLine(anchors, tail))

	out := refdata{
		id:         t.id,
		name:       t.name,
		source:     t.source,
		tailSource: r.provenance(today),
		tailFrom:   file.tailFrom,
		generated:  today.Format("2006-01-02"),
		anchors:    append(append([]marketdata.Point(nil), anchors...), tail...),
	}
	if out.tailFrom == "" {
		out.tailFrom = month(tail[0].Date)
	}
	if dry {
		return nil
	}
	return writeRefdata(path, out)
}

// returnsLine prints the appended monthly returns, the figures a reader wants
// to eyeball against the proxy's own published months.
func returnsLine(anchors, tail []marketdata.Point) string {
	var b strings.Builder
	b.WriteString("returns")
	prev := anchors[len(anchors)-1].Close
	for _, p := range tail {
		fmt.Fprintf(&b, " %s %+.2f%%", month(p.Date), (p.Close/prev-1)*100)
		prev = p.Close
	}
	return b.String()
}

func month(t time.Time) string { return t.Format("2006-01") }

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
