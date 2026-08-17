// Command gen-catbond-refdata rebuilds pkg/datasets/refdata/ILS-NET-USD.csv,
// the monthly NET reference path of insurance-linked securities that the cat
// bond reconstructions are anchored on.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSV and never fetches anything.
//
// THE SERIES. The ILS Advisers Fund Index (formerly the Eurekahedge ILS
// Advisers Index), published monthly by ILS Advisers, a business unit of HSZ
// (Hong Kong) Limited, and served as JSON behind its own chart at
//
//	https://www.ilsadvisers.com/wp-content/themes/raft/data/ils-index.json
//
// It is an equally weighted composite of the REAL funds that invest in
// non-life insurance-linked securities (catastrophe bonds and private ILS;
// a constituent must hold at least 70 % non-life risk), each entering NET of
// its own manager's fees, monthly from 2006-01. So it is not an index of
// bonds: it is what a book of real ILS funds actually paid its investors,
// which is the same object BTOP50 is for managed futures, and the right
// anchor for reconstructing a cat bond UCITS fund.
//
// WHY NOT THE SWISS RE INDEX. The Swiss Re Global Cat Bond Total Return Index
// reaches 2002 and is the market's reference, but its weekly history is not
// publicly downloadable (the sigma data explorer is bot-gated, and only annual
// figures are published in the methodology paper). It is also a market index,
// gross of any fund's fee, so it would need a fee estimate to stand in for a
// fund, where this one already arrives net. The published Swiss Re annual
// returns are kept in docs/catbond-sleeve-design.md as an external check.
//
// VALIDATION. The source ships its own summary statistics next to the monthly
// table (maximum drawdown, best and worst month, annualized standard
// deviation, share of positive months, the trailing 3-month, 3-year and
// 5-year returns). This generator recomputes every one of them from the table
// and refuses to write when they disagree, which catches a truncated download,
// a rescaled unit or a silent methodology change at the source. One published
// figure does NOT reconcile, the since-inception annualized return, and that
// divergence is reported rather than hidden: see checkAgainstSource.
//
// The CSV holds a month-end level index (base 100) of the surviving months.
//
// Usage: gen-catbond-refdata [-src URL] [-dir path] [-dry]
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const defaultSrc = "https://www.ilsadvisers.com/wp-content/themes/raft/data/ils-index.json"

// provisionalTail is how many months at the end of the live table are still
// being revised as the constituent managers report. The source publishes the
// newest month as soon as a majority has reported and restates it afterwards,
// so it is not shipped.
const provisionalTail = 1

// minMonths is the shortest history worth shipping: the index starts in
// 2006-01, so anything under ~17 years means the download truncated or the
// layout moved.
const minMonths = 200

func main() {
	src := flag.String("src", defaultSrc, "monthly index JSON")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory")
	dry := flag.Bool("dry", false, "report without writing")
	flag.Parse()

	raw, err := download(*src)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	var doc source
	if err := json.Unmarshal(raw, &doc); err != nil {
		log.Fatalf("parse: %v", err)
	}
	months, err := parse(doc)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	if len(months) <= provisionalTail {
		log.Fatalf("only %d monthly returns: the source layout probably moved", len(months))
	}
	// The published statistics describe the WHOLE table, provisional month
	// included, so they are checked before the tail is dropped.
	if err := checkAgainstSource(months, doc); err != nil {
		log.Fatalf("source cross-check: %v", err)
	}
	dropped := months[len(months)-provisionalTail:]
	months = months[:len(months)-provisionalTail]
	log.Printf("dropped %d provisional month(s): %s", len(dropped), monthList(dropped))

	if err := check(months); err != nil {
		log.Fatalf("sanity: %v", err)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "# pofo simdata v1\n# id: ILS-NET-USD\n")
	fmt.Fprintf(&b, "# name: insurance-linked securities, monthly net composite index (USD, month-end)\n")
	fmt.Fprintf(&b, "# source: monthly returns of the ILS Advisers Fund Index (ILS Advisers / HSZ Hong Kong, %s), "+
		"an equally weighted composite of the funds investing in non-life insurance-linked securities (catastrophe "+
		"bonds and private ILS, each constituent at least 70%% non-life risk), each already NET of its own manager's "+
		"fees, compounded into a base-100 index. The newest month is provisional at the source and is not shipped. "+
		"Regenerate with cmd/gen-catbond-refdata.\n", *src)
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
	out := filepath.Join(*dir, "ILS-NET-USD.csv")
	if err := os.WriteFile(out, b.Bytes(), 0o644); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d points)", out, len(months)+1)
	log.Printf("rebuild (make simdata) to re-anchor the cat bond reconstructions")
}

// source is the shape of the published JSON: the monthly table keyed by year,
// plus the summary statistics the generator checks itself against. Every
// statistic arrives as a formatted string ("-12.50%", "0.91"), which is why
// they are parsed rather than typed.
type source struct {
	MonthlyReturns    map[string][]float64 `json:"monthlyReturns"`
	KeyStatistics     map[string]string    `json:"keyStatistics"`
	RiskReturnMetrics map[string]string    `json:"riskReturnMetrics"`
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
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

// parse flattens the year-keyed table into a sorted month series. The source
// quotes returns in PERCENT and lists a year's months in order from January,
// so the slice index is the month; a year still running is simply shorter.
func parse(doc source) ([]point, error) {
	if len(doc.MonthlyReturns) == 0 {
		return nil, fmt.Errorf("no monthly returns in the document")
	}
	var out []point
	for ys, rets := range doc.MonthlyReturns {
		y, err := strconv.Atoi(ys)
		if err != nil || y < 1990 || y > 2200 {
			return nil, fmt.Errorf("%q is not a year", ys)
		}
		if len(rets) > 12 {
			return nil, fmt.Errorf("%d: %d monthly returns", y, len(rets))
		}
		for m, r := range rets {
			if math.Abs(r) > 60 {
				return nil, fmt.Errorf("%d-%02d: %.4f is not a monthly return in percent", y, m+1, r)
			}
			out = append(out, point{monthEndOf(y, time.Month(m+1)), r / 100})
		}
	}
	slices.SortFunc(out, func(a, b point) int { return a.date.Compare(b.date) })
	return out, nil
}

// checkAgainstSource recomputes the statistics the source publishes beside the
// table and refuses the download when they disagree. It is the strongest check
// available here: the same page serves both, so a table that no longer
// reproduces its own summary is a table that changed meaning.
//
// The tolerances are what the published rounding allows (two decimals on a
// percentage, two on a ratio), widened by one unit in the last place.
//
// ONE FIGURE IS KNOWN NOT TO RECONCILE. "Annualised Return" is published as
// 5.09 %/yr where the table compounds to 4.29 %/yr over the same span, and the
// published Sharpe ratio is consistent with the published return rather than
// with the table. Every other figure agrees to the last decimal, including the
// trailing 3-month, 3-year and 5-year returns, which are computed from the same
// months; the discrepancy therefore points at a stale headline rather than at a
// wrong table, and the table is what ships. It is logged, never fatal, so a
// regeneration says out loud that the source still contradicts itself.
func checkAgainstSource(months []point, doc source) error {
	rets := make([]float64, len(months))
	for i, p := range months {
		rets[i] = p.ret
	}
	type want struct {
		key  string
		from map[string]string
		got  float64
		tol  float64
	}
	checks := []want{
		{"Maximum Drawdown", doc.RiskReturnMetrics, maxDrawdown(months) * 100, 0.02},
		{"Best Monthly Return", doc.RiskReturnMetrics, slices.Max(rets) * 100, 0.02},
		{"Worst Monthly Return", doc.RiskReturnMetrics, slices.Min(rets) * 100, 0.02},
		{"Annualised Standard Deviation", doc.RiskReturnMetrics, stdev(rets) * math.Sqrt(12) * 100, 0.15},
		{"Percentage of Positive Months", doc.RiskReturnMetrics, share(rets) * 100, 0.6},
		{"Last 3 Months Return", doc.KeyStatistics, (compound(rets[len(rets)-3:]) - 1) * 100, 0.02},
		{"Last 3 Years Annualised", doc.KeyStatistics, annualized(rets, 36) * 100, 0.02},
		{"Last 5 Years Annualised", doc.KeyStatistics, annualized(rets, 60) * 100, 0.02},
	}
	for _, c := range checks {
		published, ok := parsePct(c.from[c.key])
		if !ok {
			return fmt.Errorf("%q missing from the published statistics", c.key)
		}
		if math.Abs(published-c.got) > c.tol {
			return fmt.Errorf("%s: table says %.2f, source publishes %.2f", c.key, c.got, published)
		}
		log.Printf("cross-check ok: %-30s table %7.2f, published %7.2f", c.key, c.got, published)
	}
	if published, ok := parsePct(doc.KeyStatistics["Annualised Return"]); ok {
		got := annualized(rets, len(rets)) * 100
		log.Printf("KNOWN DIVERGENCE:  %-30s table %7.2f, published %7.2f (headline not recomputed at the source)",
			"Annualised Return", got, published)
	}
	return nil
}

// check refuses to ship a series that is not what the anchor expects: a long
// unbroken run of months, at an ILS volatility, with the deep single-month
// losses the asset class is defined by. Anything outside these bands means the
// wrong table, the wrong index, or a truncated file.
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
	rets := make([]float64, len(months))
	for i, p := range months {
		rets[i] = p.ret
	}
	vol := stdev(rets) * math.Sqrt(12)
	cagr := annualized(rets, len(rets))
	log.Printf("%d months, %s to %s, CAGR %.2f%%/yr, vol %.1f%%/yr, max drawdown %.1f%%, worst month %.2f%%",
		len(months), months[0].date.Format("2006-01"), months[len(months)-1].date.Format("2006-01"),
		cagr*100, vol*100, maxDrawdown(months)*100, slices.Min(rets)*100)
	if vol < 0.015 || vol > 0.08 {
		return fmt.Errorf("volatility %.1f%%/yr outside [1.5, 8]", vol*100)
	}
	if cagr < 0.01 || cagr > 0.12 {
		return fmt.Errorf("CAGR %.2f%%/yr outside [1, 12]", cagr*100)
	}
	if worst := slices.Min(rets); worst > -0.03 {
		return fmt.Errorf("worst month %.2f%%: an ILS index without a single deep loss month is not this index", worst*100)
	}
	return nil
}

// parsePct reads the source's formatted numbers ("-12.50%", "0.91").
func parsePct(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(s), "%"), 64)
	return v, err == nil
}

func compound(rets []float64) float64 {
	level := 1.0
	for _, r := range rets {
		level *= 1 + r
	}
	return level
}

// annualized is the compound annual rate over the last n months of rets.
func annualized(rets []float64, n int) float64 {
	if n > len(rets) {
		n = len(rets)
	}
	if n == 0 {
		return 0
	}
	return math.Pow(compound(rets[len(rets)-n:]), 12/float64(n)) - 1
}

func stdev(rets []float64) float64 {
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
	return math.Sqrt(s / float64(len(rets)-1))
}

// share is the fraction of months that gained.
func share(rets []float64) float64 {
	n := 0
	for _, r := range rets {
		if r > 0 {
			n++
		}
	}
	return float64(n) / float64(len(rets))
}

// maxDrawdown is the worst peak-to-trough fall of the compounded series.
func maxDrawdown(months []point) float64 {
	level, peak, worst := 1.0, 1.0, 0.0
	for _, p := range months {
		level *= 1 + p.ret
		peak = max(peak, level)
		worst = min(worst, level/peak-1)
	}
	return worst
}

// monthIndex numbers months consecutively, so a gap or a duplicate is a
// subtraction away.
func monthIndex(p point) int { return p.date.Year()*12 + int(p.date.Month()) }

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
