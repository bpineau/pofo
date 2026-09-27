// Command gen-macropanel builds the bundled multi-country monthly macro panel
// (pkg/datasets/macropanel/oecd-monthly.csv) from the OECD's short-term
// statistics, read from the OECD's own SDMX API (refgen.OECD). It runs at
// data-generation time only (network); the pofo binary never fetches OECD, it
// embeds the committed CSV.
//
// The API admits 60 downloads an hour, so the panel is read one DATAFLOW at a
// time rather than one series at a time: every key a dataflow owes the panel,
// all countries and all fallbacks, travels in a single request, and the whole
// panel costs four downloads. It used to be read through the DBnomics mirror,
// one request per series, until the mirror's OECD provider stopped being
// indexed at 2026-06-16 and froze the panel around 2026-05 while the OECD
// already served 2026-08.
//
// The panel carries, per country and month, the macro drivers behind regime and
// factor analysis: industrial production (a growth proxy), consumer prices
// (inflation), the long and short interest rates (the monetary quadrant), and a
// share-price index. Growth/inflation breadth across these countries (the share
// accelerating) is the "world point" of the growth x inflation model; the rates
// drive the defensive rotation.
//
// Series (current OECD SDMX dataflows; the legacy OECD/MEI dataset these were
// read from until 2026-08 stopped being updated in 2024-01 while still
// answering HTTP 200, which is why the -check pass below leads with freshness):
//
//	ip         DSD_STES@DF_INDSERV  {ISO}.M.PRVM.IX.BTE.Y._Z._Z.N   production, industry B-to-E, index, seasonally adjusted
//	cpi        DSD_PRICES_COICOP2018@DF_PRICES_C2018_ALL {ISO}.M.N.CPI.IX._T.N._Z  consumer prices, all items, index
//	           DSD_PRICES@DF_PRICES_ALL {ISO}.M.N.CPI.IX._T.N._Z    the same key in the COICOP 1999 dataflow
//	shortrate  DSD_STES@DF_FINMARK  {ISO}.M.IR3TIB.PA._Z._Z._Z._Z.N 3-month interbank rate, per cent
//	longrate   DSD_STES@DF_FINMARK  {ISO}.M.IRLT.PA._Z._Z._Z._Z.N   long-term government bond yield, per cent
//	shareprice DSD_STES@DF_FINMARK  {ISO}.M.SHARE.IX._Z._Z._Z._Z.N  share prices, index
//
// Three columns have a documented fallback, and the merge between a column's
// sources is deterministic by construction: the sources are tried in priority
// order and the FIRST one that quotes a given country-month owns that cell,
// whatever order the concurrent fetches happen to complete in. Industrial
// production falls back from the whole industry aggregate (ACTIVITY BTE) to
// manufacturing alone (ACTIVITY C), which South Africa is the only country to
// need outright and which extends four more (their aggregate starts later than
// their manufacturing index); the short rate falls back from the 3-month
// interbank rate (IR3TIB) to the immediate/call money rate (IRSTCI), which most
// of the panel needs for its early decades; and consumer prices are read from
// the COICOP 2018 classification first, falling back to COICOP 1999.
//
// That last one is a migration, not a preference. The OECD is moving its price
// statistics from COICOP 1999 to COICOP 2018 country by country, and a country
// that has moved simply STOPS in the old dataflow while the old dataflow keeps
// answering HTTP 200 for the years it already holds. Japan stopped there at
// 2021-06, Mexico at 2024-07, South Africa at 2025-01 and eighteen European
// members at 2025-12, all of them current under COICOP 2018; nine countries
// (USA, DEU, GBR, AUS, KOR, POL, BRA, IND, NZL) are the other way round and
// answer 404 under COICOP 2018, which is why both are read and the newer one
// wins. The two classifications carry the same all-items index on the same base
// where they overlap (Japan to the last digit, France and Italy within 3e-4),
// so the rebasing below is a formality on this column.
//
// A rate is a rate, so the two short-rate sources are spliced as they come. Two
// index levels are not: they carry different bases, so a fallback filling in
// front of an index column is REBASED onto the source above it at the first
// month both quote, exactly as a donor series is lifted onto its target. Levels
// only ever enter the model as year-on-year ratios, so what this protects is the
// twelve months straddling the junction.
//
// Usage:
//
//	gen-macropanel [-oecd URL] [-o path] [-dry] [-check=false]
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/cmd/internal/refgen"
)

// countries covered by the panel: a broad OECD + large-emerging set, wide enough
// that growth/inflation breadth is a meaningful proxy for the world.
var countries = []string{
	"USA", "JPN", "DEU", "FRA", "GBR", "ITA", "CAN", "AUS", "ESP", "NLD",
	"BEL", "SWE", "CHE", "AUT", "DNK", "FIN", "NOR", "PRT", "GRC", "IRL",
	"POL", "CZE", "HUN", "KOR", "MEX", "TUR", "NZL", "ZAF", "BRA", "IND",
}

// column is one output column and the OECD series that may fill it, in
// PRIORITY order: the first source quoting a country-month owns that cell. The
// "{ISO}" of each key stands for the ISO country code.
type column struct {
	name string
	// level says the column carries an index level rather than a percentage,
	// so a lower-priority source must be rebased onto the one above it before
	// it may fill anything (see rebase).
	level   bool
	sources []source
}

// source is one candidate series for a column: the dataflow it lives in, its
// key, and the short label used in the logs when it had to stand in for the one
// above it.
type source struct {
	label string
	flow  refgen.Flow
	key   string
}

var columns = []column{
	{"ip", true, []source{
		{"industry B-to-E", refgen.IndServ, "{ISO}.M.PRVM.IX.BTE.Y._Z._Z.N"},
		{"manufacturing", refgen.IndServ, "{ISO}.M.PRVM.IX.C.Y._Z._Z.N"},
	}},
	{"cpi", true, []source{
		{"all items (COICOP 2018)", refgen.PricesC2018, "{ISO}.M.N.CPI.IX._T.N._Z"},
		{"all items (COICOP 1999)", refgen.Prices, "{ISO}.M.N.CPI.IX._T.N._Z"},
	}},
	{"shortrate", false, []source{
		{"3-month interbank", refgen.FinMark, "{ISO}.M.IR3TIB.PA._Z._Z._Z._Z.N"},
		{"immediate rate", refgen.FinMark, "{ISO}.M.IRSTCI.PA._Z._Z._Z._Z.N"},
	}},
	{"longrate", false, []source{
		{"long-term govt yield", refgen.FinMark, "{ISO}.M.IRLT.PA._Z._Z._Z._Z.N"},
	}},
	{"shareprice", true, []source{
		{"share prices", refgen.FinMark, "{ISO}.M.SHARE.IX._Z._Z._Z._Z.N"},
	}},
}

// cell identifies one fetched series: a country, a column and which of that
// column's sources it came from.
type cell struct {
	iso       string
	col, rank int
}

func main() {
	base := flag.String("oecd", refgen.OECDBase, "OECD SDMX API base URL")
	out := flag.String("o", "pkg/datasets/macropanel/oecd-monthly.csv", "output CSV path")
	dry := flag.Bool("dry", false, "print coverage and checks without writing")
	check := flag.Bool("check", true, "run the sanity checks before writing")
	flag.Parse()

	raw := fetchAll(*base)
	panel, contrib := merge(raw)
	dropRateDropouts(panel)
	recs := flatten(panel)
	logCoverage(recs, contrib)
	if *check {
		runChecks(panel, recs)
		compareShipped(*out, panel)
	}
	if *dry {
		return
	}
	if err := writeCSV(*out, recs); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d rows, %d countries)", *out, len(recs), len(panel))
}

// fetchAll downloads every candidate series, ONE download per dataflow: all the
// keys a dataflow owes the panel (every country, every source of every column
// living there) travel in a single request, which is what keeps a full refresh
// inside the OECD's hourly download budget. It only fills the map of RAW
// series, one entry per (country, column, source), so it cannot decide which
// source wins a cell: merge does that, afterwards, in priority order. A series
// the dataflow does not carry is a normal, documented absence; a failed
// download aborts the run rather than silently shipping a thinner panel.
func fetchAll(base string) map[cell]map[string]float64 {
	var flows []refgen.Flow
	keys := map[refgen.Flow][]string{}
	cells := map[refgen.Flow]map[string][]cell{}
	for _, iso := range countries {
		for ci, c := range columns {
			for ri, s := range c.sources {
				if cells[s.flow] == nil {
					flows = append(flows, s.flow)
					cells[s.flow] = map[string][]cell{}
				}
				k := strings.ReplaceAll(s.key, "{ISO}", iso)
				if cells[s.flow][k] == nil {
					keys[s.flow] = append(keys[s.flow], k)
				}
				cells[s.flow][k] = append(cells[s.flow][k], cell{iso, ci, ri})
			}
		}
	}
	raw := map[cell]map[string]float64{}
	for _, f := range flows {
		got, err := refgen.OECD(base, f, keys[f]...)
		if err != nil {
			log.Fatalf("%v; nothing written", err)
		}
		log.Printf("OECD %s: %d of %d series", f.ID, len(got), len(keys[f]))
		for k, pts := range got {
			obs := make(map[string]float64, len(pts))
			for _, p := range pts {
				obs[p.Date.Format("2006-01")] = p.Close
			}
			for _, c := range cells[f][k] {
				raw[c] = obs
			}
		}
	}
	return raw
}

// merge assembles the panel from the raw series, deterministically: for each
// country and column the sources are applied in priority order and a month
// already filled is never overwritten. Nothing here depends on the order the
// fetches completed in, so two runs over the same vintage of the provider's
// data produce the same bytes. It also returns, per column, how many
// country-months each source ended up supplying, so a fallback that quietly
// takes over shows in the log.
func merge(raw map[cell]map[string]float64) (map[string]map[string]map[string]float64, map[cell]prov) {
	panel := map[string]map[string]map[string]float64{}
	contrib := map[cell]prov{}
	for _, iso := range countries {
		for ci, c := range columns {
			for ri := range c.sources {
				k := cell{iso, ci, ri}
				obs, ok := raw[k]
				if !ok {
					continue
				}
				used := prov{factor: 1}
				if ri > 0 && c.level {
					obs, used.factor = rebase(obs, raw[cell{iso, ci, ri - 1}])
				}
				for m, v := range obs {
					if panel[iso] == nil {
						panel[iso] = map[string]map[string]float64{}
					}
					if panel[iso][m] == nil {
						panel[iso][m] = map[string]float64{}
					}
					if _, filled := panel[iso][m][c.name]; filled {
						continue // a higher-priority source already owns this month
					}
					panel[iso][m][c.name] = v
					used.months++
				}
				contrib[k] = used
			}
		}
	}
	return panel, contrib
}

// dropRateDropouts removes the rate cells that are a missing observation the
// provider published as a zero. A rate is genuinely allowed to be zero, and
// even negative: Switzerland charged for francs in 1978 and Japan's ten-year
// yield sat exactly on zero under yield-curve control. What is not a rate is a
// LONE zero between two months well away from it (Australia, 1969-11, between
// 5.6 % and 5.65 %), so that is the only shape removed, and each removal is
// named in the log rather than absorbed silently.
func dropRateDropouts(panel map[string]map[string]map[string]float64) {
	for _, iso := range countries {
		for _, col := range []string{"shortrate", "longrate"} {
			s := series(panel, iso, col)
			for _, m := range months(s) {
				if s[m] != 0 {
					continue
				}
				t, err := time.Parse("2006-01", m)
				if err != nil {
					continue
				}
				before, okb := s[t.AddDate(0, -1, 0).Format("2006-01")]
				after, oka := s[t.AddDate(0, 1, 0).Format("2006-01")]
				if !okb || !oka || math.Abs(before) < 1 || math.Abs(after) < 1 {
					continue
				}
				log.Printf("  %s %s %s: dropping a zero between %.2f%% and %.2f%%, a dropout rather than a rate", iso, col, m, before, after)
				delete(panel[iso][m], col)
			}
		}
	}
}

// prov records what one source ended up contributing to a column: how many
// country-months it owns, and the factor it was rebased by to get there.
type prov struct {
	months int
	factor float64
}

// rebase scales a fallback index onto the source above it, using their first
// common month, so the two never meet at a level jump. Without an overlap (the
// fallback is all this country has) the series is kept as it stands: it is then
// the column's only level, and only its own ratios are ever read.
func rebase(fallback, primary map[string]float64) (map[string]float64, float64) {
	var at string
	for _, m := range months(primary) {
		if _, ok := fallback[m]; ok {
			at = m
			break
		}
	}
	if at == "" || fallback[at] == 0 {
		return fallback, 1
	}
	f := primary[at] / fallback[at]
	out := make(map[string]float64, len(fallback))
	for m, v := range fallback {
		out[m] = v * f
	}
	return out, f
}

// compareShipped measures, column by column, how far the rebuilt panel
// reproduces the committed one it is about to replace: the cells both carry,
// how many of them read the same at the four decimals the file is written at,
// which countries the others belong to and the months they span, the largest
// change the model would see (a rate as it stands, an index level through its
// year-on-year ratio), and the cells the new panel no longer carries at all.
//
// Only the RATES are gated. A market rate is not revised once it is a year old,
// so a firm rate cell (older than the shipped panel's last firmRateMonths) that
// moves by more than maxFirmRateMove percentage point means the source changed
// what it serves, and nothing is written. The index columns are reported and
// not gated, because their publishers do revise them legitimately and deep: a
// seasonal-adjustment pass moves the whole of an industrial-production series,
// a national rebasing moves a consumer-price index back to its first year.
// Measured when the panel moved from the DBnomics mirror to the OECD's own API
// (2026-09): every common share-price and short-rate cell reproduced but four
// recent ones; the long rates reproduced but for France and Italy, whose whole
// histories now carry a sub-basis-point re-rounding; Japan's CPI moved end to
// end after the country's rebasing (at most 0.40 pt of year-on-year rate); and
// 15 % of the production cells moved with the usual revisions, Belgium's by up
// to 5.7 pt of year-on-year rate around 2000. The mirror's last Japanese
// three-month rate (2026-05) was also a provisional 0.73 % the OECD has since
// replaced with 1.27 %.
func compareShipped(path string, panel map[string]map[string]map[string]float64) {
	const (
		firmRateMonths  = 12
		maxFirmRateMove = 0.10 // percentage point
	)
	old, err := readPanel(path)
	if err != nil {
		log.Printf("check: no shipped panel to compare against (%v)", err)
		return
	}
	newest := ""
	for _, byMonth := range old {
		for m := range byMonth {
			newest = max(newest, m)
		}
	}
	firm := monthsBefore(newest, firmRateMonths)
	failed := 0
	for _, c := range columns {
		common, same, lost := 0, 0, 0
		first, last := "", ""
		moved := map[string]int{}
		worst, worstAt := 0.0, ""
		for iso, byMonth := range old {
			wasCol, nowCol := series(old, iso, c.name), series(panel, iso, c.name)
			for m, cols := range byMonth {
				was, ok := cols[c.name]
				if !ok {
					continue
				}
				now, ok := panel[iso][m][c.name]
				if !ok {
					lost++
					continue
				}
				common++
				if cellText(now) == cellText(was) {
					same++
					continue
				}
				moved[iso]++
				if first == "" || m < first {
					first = m
				}
				last = max(last, m)
				// What a move is worth is what the model reads: a rate as it
				// stands, an index level only through its year-on-year ratio.
				d := math.Abs(now - was)
				if !c.level && m < firm && d > maxFirmRateMove {
					failed++
					log.Printf("CHECK FAILED: %s %s %s moved from %.4f to %.4f, a year-old rate the source should not revise", iso, c.name, m, was, now)
				}
				if c.level {
					y0, ok0 := yoy(wasCol, m)
					y1, ok1 := yoy(nowCol, m)
					if !ok0 || !ok1 {
						continue
					}
					d = math.Abs(y1-y0) * 100
				}
				if d > worst {
					worst, worstAt = d, iso+" "+m
				}
			}
		}
		var who []string
		for iso, n := range moved {
			who = append(who, fmt.Sprintf("%s %d", iso, n))
		}
		sort.Strings(who)
		msg := fmt.Sprintf("check %-10s vs the shipped panel: %d of %d common cells reproduced (%.2f%%), %d cells no longer served",
			c.name, same, common, 100*float64(same)/float64(max(common, 1)), lost)
		if len(who) > 0 {
			what := "level"
			if c.level {
				what = "year-on-year rate"
			}
			msg += fmt.Sprintf("; moved %s..%s, worst %s change %.3f pt (%s): %s", first, last, what, worst, worstAt, strings.Join(who, ", "))
		}
		log.Print(msg)
	}
	if failed > 0 {
		log.Fatalf("%d firm rate cell(s) revised, nothing written", failed)
	}
}

// cellText is how a value is written in the panel.
func cellText(v float64) string { return strconv.FormatFloat(v, 'f', 4, 64) }

// readPanel reads a panel CSV written by writeCSV back into its
// country -> month -> column shape; empty cells are absent.
func readPanel(path string) (map[string]map[string]map[string]float64, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]map[string]map[string]float64{}
	var head []string
	sc := bufio.NewScanner(bytes.NewReader(body))
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, ",")
		if head == nil {
			head = f
			continue
		}
		if len(f) != len(head) {
			return nil, fmt.Errorf("%s: row %q has %d fields, want %d", path, line, len(f), len(head))
		}
		iso, m := f[0], f[1]
		for i := 2; i < len(f); i++ {
			if f[i] == "" {
				continue
			}
			v, err := strconv.ParseFloat(f[i], 64)
			if err != nil {
				return nil, fmt.Errorf("%s: row %q: %v", path, line, err)
			}
			if out[iso] == nil {
				out[iso] = map[string]map[string]float64{}
			}
			if out[iso][m] == nil {
				out[iso][m] = map[string]float64{}
			}
			out[iso][m][head[i]] = v
		}
	}
	return out, sc.Err()
}

// record is one country-month row of the panel.
type record struct {
	ISO, Month string
	Val        map[string]float64
}

func flatten(panel map[string]map[string]map[string]float64) []record {
	var out []record
	for iso, byMonth := range panel {
		for m, cols := range byMonth {
			out = append(out, record{iso, m, cols})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ISO != out[j].ISO {
			return out[i].ISO < out[j].ISO
		}
		return out[i].Month < out[j].Month
	})
	return out
}

func writeCSV(path string, recs []record) error {
	var b strings.Builder
	b.WriteString("# Multi-country monthly macro panel: OECD short-term statistics and prices.\n")
	b.WriteString("# Columns: iso,date(YYYY-MM),ip,cpi,shortrate,longrate,shareprice\n")
	b.WriteString("# ip/cpi/shareprice are index levels; shortrate/longrate are per-cent yields.\n")
	b.WriteString("# Source: OECD SDMX API (https://sdmx.oecd.org): production from\n")
	b.WriteString("# DSD_STES@DF_INDSERV, prices from DSD_PRICES_COICOP2018@DF_PRICES_C2018_ALL\n")
	b.WriteString("# falling back to DSD_PRICES@DF_PRICES_ALL, rates and share\n")
	b.WriteString("# prices from DSD_STES@DF_FINMARK (the legacy OECD/MEI froze at 2024-01).\n")
	b.WriteString("# ip = industry B-to-E (manufacturing alone where a country has no aggregate);\n")
	b.WriteString("# shortrate = 3-month interbank, immediate rate for the months it lacks.\n")
	b.WriteString("# Regenerate: make macropanel\n")
	b.WriteString("iso,date,ip,cpi,shortrate,longrate,shareprice\n")
	num := func(cols map[string]float64, k string) string {
		v, ok := cols[k]
		if !ok {
			return ""
		}
		return strconv.FormatFloat(v, 'f', 4, 64)
	}
	for _, r := range recs {
		fmt.Fprintf(&b, "%s,%s,%s,%s,%s,%s,%s\n", r.ISO, r.Month,
			num(r.Val, "ip"), num(r.Val, "cpi"), num(r.Val, "shortrate"),
			num(r.Val, "longrate"), num(r.Val, "shareprice"))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// logCoverage prints the panel's shape: rows per column, the countries a column
// does not reach at all, and every country-month a fallback source supplied.
func logCoverage(recs []record, contrib map[cell]prov) {
	byISO := map[string]int{}
	colCount := map[string]int{}
	have := map[string]map[string]bool{}
	for _, r := range recs {
		byISO[r.ISO]++
		for k := range r.Val {
			colCount[k]++
			if have[k] == nil {
				have[k] = map[string]bool{}
			}
			have[k][r.ISO] = true
		}
	}
	log.Printf("panel: %d country-month rows, %d countries", len(recs), len(byISO))
	for ci, c := range columns {
		var missing []string
		for _, iso := range countries {
			if !have[c.name][iso] {
				missing = append(missing, iso)
			}
		}
		log.Printf("  %-10s %5d rows, %2d countries%s", c.name, colCount[c.name], len(have[c.name]), missingNote(missing))
		for ri := 1; ri < len(c.sources); ri++ {
			var fell []string
			for _, iso := range countries {
				if p := contrib[cell{iso, ci, ri}]; p.months > 0 {
					fell = append(fell, fmt.Sprintf("%s(%d months, x%.4f)", iso, p.months, p.factor))
				}
			}
			if len(fell) > 0 {
				log.Printf("    fallback %-20s %s", c.sources[ri].label, strings.Join(fell, " "))
			}
		}
	}
}

func missingNote(missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	return ", none for " + strings.Join(missing, " ")
}

// runChecks grades the freshly downloaded panel before it is allowed to replace
// the committed one, and stops the generator when a check fails. A panel that
// downloaded cleanly has proved nothing: the house rule is that data is
// validated against something known BEFORE anything trusts it.
//
// The checks, and why each is the one to run:
//
//   - Freshness. The MEI dataflow this generator read until 2026-08 went on
//     answering HTTP 200 for two and a half years after it stopped being
//     updated, so a stale tail is the failure mode to catch first: every column
//     must reach within maxLag of today, somewhere in the panel. The bound is
//     three months, four for industrial production, which the OECD publishes
//     with a longer lag. It used to be six, which a MIRROR's freeze passes: the
//     DBnomics copy this generator read until 2026-09 stopped at 2026-05 and
//     still looked fresh four months later.
//   - Freshness, per country. The check above is blind to ONE country freezing:
//     the United States alone reaching the current month satisfies it, which is
//     exactly how Japanese CPI shipped stuck at 2021-06 for as long as it did,
//     with Mexico and South Africa behind it. So each country-column is also
//     compared to the newest month that column reaches ANYWHERE in the panel,
//     and may not trail it by more than maxCountryLag. The bound is generous
//     (18 months) because members genuinely publish at different cadences and a
//     few series legitimately end; what it catches is a country left behind by a
//     classification change, which is measured in years.
//   - Coverage. A column must still reach a wide majority of the countries, and
//     the panel a wide majority of its country list: a dataflow that quietly
//     drops half its reporters would otherwise pass every value-level check.
//   - Flat runs at the END of the rate columns. A degraded fetch, or a series
//     the provider stopped maintaining while still serving it, shows as one
//     level repeated up to the last month it publishes. Flatness in the MIDDLE
//     of a rate series is not evidence of anything (India's call money rate was
//     administered at one level for the whole of the 1980s, ten years to the
//     basis point), so it is logged and not gated; what is gated is a rate that
//     is still flat where the series ends.
//   - One anchor per column, all on the United States, where the answer is
//     public knowledge and independent of this repository: industrial production
//     fell ~16 % between February and April 2020; CPI year-on-year peaked at
//     ~9 % in mid-2022; the 3-month rate averaged ~5.3 % over 2023 after the
//     hiking cycle; the long yield topped 13 % in 1981 and bottomed under 1 % in
//     2020; the share-price index fell more than 40 % peak to trough in
//     2007-2009 and stands above its 2019 level.
func runChecks(panel map[string]map[string]map[string]float64, recs []record) {
	failed := 0
	fail := func(format string, a ...any) {
		failed++
		log.Printf("CHECK FAILED: "+format, a...)
	}

	thisMonth := time.Now().UTC().Format("2006-01")
	for _, c := range columns {
		maxLag := 3 // months
		if c.name == "ip" {
			maxLag = 4 // production is published later than prices and rates
		}
		horizon := monthsBefore(thisMonth, maxLag)
		last, n := "", 0
		for _, r := range recs {
			if _, ok := r.Val[c.name]; !ok {
				continue
			}
			n++
			if r.Month > last {
				last = r.Month
			}
		}
		log.Printf("check %-10s freshness: newest month %s", c.name, last)
		if last < horizon {
			fail("%s stops at %s (nothing since %s): its dataflow looks frozen", c.name, last, horizon)
		}
		if n < 5000 {
			fail("%s carries only %d country-months", c.name, n)
		}
	}

	const maxCountryLag = 18 // months behind the column's own newest month
	for _, c := range columns {
		newest := ""
		last := map[string]string{}
		for _, r := range recs {
			if _, ok := r.Val[c.name]; !ok {
				continue
			}
			if r.Month > last[r.ISO] {
				last[r.ISO] = r.Month
			}
			if r.Month > newest {
				newest = r.Month
			}
		}
		cut := monthsBefore(newest, maxCountryLag)
		var stale []string
		for iso, m := range last {
			if m < cut {
				stale = append(stale, iso+" "+m)
			}
		}
		sort.Strings(stale)
		log.Printf("check %-10s per-country freshness: newest %s, %d of %d countries behind %s%s",
			c.name, newest, len(stale), len(last), cut, listNote(stale))
		if len(stale) > 0 {
			fail("%s leaves %d countries frozen more than %d months behind the panel (%s): a source that moved rather than data that ended",
				c.name, len(stale), maxCountryLag, strings.Join(stale, ", "))
		}
	}

	if len(panel) < len(countries)-2 {
		log.Printf("check coverage: %d of %d countries", len(panel), len(countries))
		fail("the panel lost countries: %d of %d", len(panel), len(countries))
	}
	for _, c := range columns {
		n := 0
		for _, iso := range countries {
			for _, cols := range panel[iso] {
				if _, ok := cols[c.name]; ok {
					n++
					break
				}
			}
		}
		log.Printf("check %-10s coverage: %d of %d countries", c.name, n, len(countries))
		if n < len(countries)-3 {
			fail("%s reaches only %d of %d countries", c.name, n, len(countries))
		}
	}

	for _, col := range []string{"shortrate", "longrate"} {
		worstISO, worst, at := "", 0, ""
		tailISO, tail := "", 0
		for _, iso := range countries {
			s := series(panel, iso, col)
			if n, end := longestFlatRun(s); n > worst {
				worstISO, worst, at = iso, n, end
			}
			if n := trailingFlatRun(s); n > tail {
				tailISO, tail = iso, n
			}
		}
		log.Printf("check %-10s flat runs: %d months at the end of the longest one (%s); deepest anywhere %d months (%s, ending %s, not gated)",
			col, tail, tailISO, worst, worstISO, at)
		if tail > 24 {
			fail("%s ends on %d repeated months (%s): that series looks dead rather than quiet", col, tail, tailISO)
		}
	}

	ip := series(panel, "USA", "ip")
	if drop := ip["2020-04"]/ip["2020-02"] - 1; drop > -0.10 || drop < -0.25 {
		fail("US industrial production fell %.1f%% over the covid stop, expected about -16%%", drop*100)
	} else {
		log.Printf("check ip         anchor: US production %.1f%% between 2020-02 and 2020-04 (expected ~-16%%)", drop*100)
	}

	cpi := series(panel, "USA", "cpi")
	peak, at := 0.0, ""
	for m := range cpi {
		if m < "2022-01" || m > "2022-12" {
			continue
		}
		if y, ok := yoy(cpi, m); ok && y > peak {
			peak, at = y, m
		}
	}
	log.Printf("check cpi        anchor: US CPI year-on-year peaked at %.1f%% in %s (expected ~9.1%% in 2022-06)", peak*100, at)
	if peak < 0.08 || peak > 0.10 {
		fail("the US 2022 inflation peak reads %.1f%%, expected about 9%%", peak*100)
	}

	short := mean(series(panel, "USA", "shortrate"), "2023-01", "2023-12")
	log.Printf("check shortrate  anchor: US 3-month rate averaged %.2f%% over 2023 (expected ~5.3%%)", short)
	if short < 4.5 || short > 5.8 {
		fail("the US 2023 3-month rate averages %.2f%%, expected about 5.3%%", short)
	}

	long := series(panel, "USA", "longrate")
	hi := mean(long, "1981-09", "1981-10")
	lo := mean(long, "2020-07", "2020-08")
	log.Printf("check longrate   anchor: US long yield %.2f%% in autumn 1981, %.2f%% in mid-2020 (expected ~15%% and ~0.6%%)", hi, lo)
	if hi < 13 || lo > 1.0 {
		fail("the US long yield reads %.2f%% in 1981 and %.2f%% in 2020, expected about 15%% and 0.6%%", hi, lo)
	}

	share := series(panel, "USA", "shareprice")
	fallDD := drawdown(share, "2007-01", "2009-12")
	log.Printf("check shareprice anchor: US share prices fell %.0f%% over 2007-2009 and stand %.0f%% above 2019", fallDD*100,
		(mean(share, lastMonth(share), lastMonth(share))/mean(share, "2019-01", "2019-12")-1)*100)
	if fallDD > -0.40 {
		fail("the US share index only fell %.0f%% over 2007-2009, expected more than 40%%", fallDD*100)
	}
	if mean(share, lastMonth(share), lastMonth(share)) <= mean(share, "2019-01", "2019-12") {
		fail("the US share index no longer stands above its 2019 level")
	}

	if failed > 0 {
		log.Fatalf("%d sanity check(s) failed, nothing written", failed)
	}
}

// series extracts one country's column as a month-keyed map.
func series(panel map[string]map[string]map[string]float64, iso, col string) map[string]float64 {
	out := map[string]float64{}
	for m, cols := range panel[iso] {
		if v, ok := cols[col]; ok {
			out[m] = v
		}
	}
	return out
}

// months returns a series' months in ascending order.
func months(s map[string]float64) []string {
	out := make([]string, 0, len(s))
	for m := range s {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

func lastMonth(s map[string]float64) string {
	ms := months(s)
	if len(ms) == 0 {
		return ""
	}
	return ms[len(ms)-1]
}

// yoy is a level series' year-on-year rate at month m, as a fraction.
func yoy(s map[string]float64, m string) (float64, bool) {
	t, err := time.Parse("2006-01", m)
	if err != nil {
		return 0, false
	}
	prev, ok := s[t.AddDate(-1, 0, 0).Format("2006-01")]
	if !ok || prev == 0 {
		return 0, false
	}
	cur, ok := s[m]
	if !ok {
		return 0, false
	}
	return cur/prev - 1, true
}

// mean averages a series over the inclusive month range [from, to].
func mean(s map[string]float64, from, to string) float64 {
	sum, n := 0.0, 0
	for m, v := range s {
		if m >= from && m <= to {
			sum += v
			n++
		}
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// drawdown is the deepest peak-to-trough fall of a level series over the
// inclusive month range, as a negative fraction.
func drawdown(s map[string]float64, from, to string) float64 {
	peak, worst := math.Inf(-1), 0.0
	for _, m := range months(s) {
		if m < from || m > to {
			continue
		}
		peak = max(peak, s[m])
		worst = min(worst, s[m]/peak-1)
	}
	return worst
}

// trailingFlatRun counts the months at the very END of a series that repeat its
// last level. A series the provider abandoned goes on being served, flat, to its
// last month; a quiet policy rate eventually moves again.
func trailingFlatRun(s map[string]float64) int {
	ms := months(s)
	n := 0
	for i := len(ms) - 1; i > 0 && s[ms[i]] == s[ms[i-1]]; i-- {
		n++
	}
	return n
}

// monthsBefore shifts a "YYYY-MM" month back by n months.
func monthsBefore(month string, n int) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	return t.AddDate(0, -n, 0).Format("2006-01")
}

// listNote spells out a short list of offenders in a log line, or nothing when
// there are none.
func listNote(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return " (" + strings.Join(items, ", ") + ")"
}

// longestFlatRun returns the longest run of consecutive months carrying the
// exact same level, and the month it ends on.
func longestFlatRun(s map[string]float64) (int, string) {
	ms := months(s)
	best, at, run := 0, "", 0
	for i := 1; i < len(ms); i++ {
		if s[ms[i]] != s[ms[i-1]] {
			run = 0
			continue
		}
		run++
		if run > best {
			best, at = run, ms[i]
		}
	}
	return best, at
}
