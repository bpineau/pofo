// Command gen-sp500-refdata rebuilds pkg/datasets/refdata/SP500-USD.csv as a
// month-END S&P 500 total-return series, so the shaped S&P 500 reconstruction
// (simgen.sp500Index behind SP500, SPX/VUAA and every VFINX-extended composite)
// carries the correct intra-year path, not a smeared one.
//
// It runs at data-generation time only (network); the pofo binary embeds the
// CSV and never fetches Yahoo or datahub.
//
// WHY THIS EXISTS. The previous SP500-USD came straight from Shiller's monthly
// file, whose price column is the MONTHLY AVERAGE of daily closes, dated the
// first of the month. Averaged month levels smear every turning point: rebuilt
// calendar returns were off by up to ~5 points (1987 came out -0.1 % against the
// real +5.3 %), and once anchorShape pinned those first-of-month levels under the
// daily ^GSPC shape the whole reconstruction slipped ~1 month. This generator
// instead takes point-in-time month-END levels from the best source available
// for each era and dates every point on the actual last trading day, so the
// anchor a shape is pinned to is the real month-end close.
//
// Sourcing, newest era first (each month's return comes from the first source
// that covers both its and the previous month's end):
//
//   - 1988-> : ^SP500TR, the S&P 500 Total Return INDEX itself (Yahoo daily).
//     Exact: dividends are already inside the index, so no reinvestment
//     assumption is needed and calendar returns match the published figures to
//     the basis point.
//   - 1928-1988 : ^GSPC month-end price (Yahoo daily) with Shiller's monthly
//     dividend reinvested (annual rate / 12 added to the month's price return).
//     Reproduces the published S&P 500 TR to ~0.2 %/yr across the era.
//   - 1871-1928 : Shiller's own monthly price (the pre-^GSPC average) with the
//     same dividend reinvestment. No daily data exists this far back, so the
//     averaging is unavoidable here; it only affects the deep pre-Depression
//     tail, dated on the calendar month-end.
//
// Returns are chained (never levels spliced), so the era boundaries carry no
// discontinuity: each month's return is computed wholly within one source.
//
// Usage: gen-sp500-refdata [-shiller URL] [-dir path] [-dry]
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// defaultShiller is Robert Shiller's monthly S&P 500 data as mirrored by the
// datahub "s-and-p-500" package: Date, SP500 (monthly-average price), Dividend
// (annualized), Earnings, CPI, ... one row per month from 1871.
const defaultShiller = "https://raw.githubusercontent.com/datasets/s-and-p-500/main/data/data.csv"

func main() {
	shillerURL := flag.String("shiller", defaultShiller, "URL of the Shiller monthly S&P 500 CSV (price + dividend)")
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory for the refdata CSV")
	dry := flag.Bool("dry", false, "build and report, but do not write the file")
	flag.Parse()

	ctx := context.Background()
	client := marketdata.NewClient("")
	from := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

	gspc := monthEnd(mustFetch(ctx, client, "^GSPC", from))    // price index, ~1927
	sptr := monthEnd(mustFetch(ctx, client, "^SP500TR", from)) // total-return index, ~1988
	shPrice, shDiv := fetchShiller(*shillerURL)                // monthly avg price + annual dividend, ~1871

	pts := build(gspc, sptr, shPrice, shDiv)
	report("SP500-USD", pts)
	// Spot-check a few published S&P 500 TR calendar returns so a bad build is
	// caught before it is ever written.
	for _, c := range []struct {
		y         int
		want, tol float64
	}{
		{1958, 43.36, 0.5}, {1987, 5.25, 0.5}, {2008, -37.00, 0.4},
		{2019, 31.49, 0.3}, {2022, -18.11, 0.3}, {2024, 25.02, 0.3},
	} {
		if got := calYear(pts, c.y); math.Abs(got-c.want) > c.tol {
			log.Fatalf("sanity: %d TR = %.2f%%, want %.2f%% (tol %.2f)", c.y, got, c.want, c.tol)
		}
	}
	log.Print("sanity: calendar-year checks passed (1958/1987/2008/2019/2022/2024)")

	if *dry {
		return
	}
	write(*dir, "SP500-USD", "S&P 500 total return (USD, monthly month-end)",
		"month-END S&P 500 total return: ^SP500TR index (Yahoo, 1988->), ^GSPC month-end price + Shiller dividend reinvested (1928-1988), Shiller monthly price + dividend (1871-1928). Point-in-time month-end levels dated on the last trading day; returns chained across eras. Proxy behind VFINX and the S&P 500 UCITS ETFs.",
		pts)
}

// dated is a month-end level: the last close of a month and the date it fell on.
type dated struct {
	date  time.Time
	close float64
}

// monthEnd reduces a daily series to one point per calendar month, the last
// close in that month, keyed "YYYY-MM".
func monthEnd(s *marketdata.Series) map[string]dated {
	m := make(map[string]dated, len(s.Points)/20)
	for _, p := range s.Points {
		k := p.Date.Format("2006-01")
		if cur, ok := m[k]; !ok || p.Date.After(cur.date) {
			m[k] = dated{p.Date, p.Close}
		}
	}
	return m
}

// build assembles the month-end total-return level series from 1871 to the last
// month all price sources still cover, chaining each month's return from the
// best source that reaches both it and the previous month.
func build(gspc, sptr map[string]dated, shPrice, shDiv map[string]float64) []marketdata.Point {
	months := monthKeys(shPrice, gspc, sptr)
	out := make([]marketdata.Point, 0, len(months))
	level := 100.0
	var prev string
	for _, k := range months {
		date, ok := endDate(k, gspc, sptr)
		if !ok {
			date = calendarMonthEnd(k)
		}
		if prev == "" {
			out = append(out, marketdata.Point{Date: date, Close: level})
			prev = k
			continue
		}
		r, ok := monthReturn(prev, k, gspc, sptr, shPrice, shDiv)
		if !ok {
			continue // a gap the sources cannot bridge; skip without breaking the chain level
		}
		level *= 1 + r
		out = append(out, marketdata.Point{Date: date, Close: level})
		prev = k
	}
	return out
}

// monthReturn is the total return from month a to month b, taken from the first
// source covering both ends: the TR index, else month-end price plus that
// month's reinvested dividend, else the Shiller average price plus dividend.
func monthReturn(a, b string, gspc, sptr map[string]dated, shPrice, shDiv map[string]float64) (float64, bool) {
	if pb, ok := sptr[b]; ok {
		if pa, ok := sptr[a]; ok {
			return pb.close/pa.close - 1, true
		}
	}
	div := shDiv[b] / 12 // Shiller dividend is an annual rate; one month reinvests a twelfth
	if pb, ok := gspc[b]; ok {
		if pa, ok := gspc[a]; ok {
			return (pb.close+div)/pa.close - 1, true
		}
	}
	if pb, ok := shPrice[b]; ok {
		if pa, ok := shPrice[a]; ok && pa > 0 {
			return (pb+div)/pa - 1, true
		}
	}
	return 0, false
}

// endDate returns the actual last-trading-day date for a month from the daily
// sources (TR index preferred, else the price index); ok is false pre-^GSPC.
func endDate(k string, gspc, sptr map[string]dated) (time.Time, bool) {
	if p, ok := sptr[k]; ok {
		return p.date, true
	}
	if p, ok := gspc[k]; ok {
		return p.date, true
	}
	return time.Time{}, false
}

// calendarMonthEnd is the last calendar day of the month "YYYY-MM".
func calendarMonthEnd(k string) time.Time {
	t, _ := time.Parse("2006-01", k)
	return t.AddDate(0, 1, -1)
}

// monthKeys is the sorted union of every month any source covers, from the
// first Shiller month to the last month all still overlap on.
func monthKeys(shPrice map[string]float64, gspc, sptr map[string]dated) []string {
	seen := map[string]bool{}
	for k := range shPrice {
		seen[k] = true
	}
	for k := range gspc {
		seen[k] = true
	}
	for k := range sptr {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	// "YYYY-MM" sorts chronologically as a string.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// calYear is the December-to-December total return (%) of calendar year y.
func calYear(pts []marketdata.Point, y int) float64 {
	dec := func(yr int) float64 {
		var v float64
		for _, p := range pts {
			if p.Date.Year() == yr && p.Date.Month() == 12 {
				v = p.Close
			}
		}
		return v
	}
	a, b := dec(y-1), dec(y)
	if a == 0 {
		return 0
	}
	return (b/a - 1) * 100
}

func mustFetch(ctx context.Context, c *marketdata.Client, id string, from time.Time) *marketdata.Series {
	s, err := c.Fetch(ctx, id, from)
	if err != nil {
		log.Fatalf("fetch %s: %v", id, err)
	}
	if s == nil || len(s.Points) < 100 {
		log.Fatalf("fetch %s: too few points", id)
	}
	return s
}

// fetchShiller downloads the Shiller monthly CSV and returns the price and
// annualized-dividend columns keyed "YYYY-MM".
func fetchShiller(url string) (price, div map[string]float64) {
	cl := &http.Client{Timeout: 60 * time.Second}
	resp, err := cl.Get(url)
	if err != nil {
		log.Fatalf("shiller: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("shiller: HTTP %d", resp.StatusCode)
	}
	price, div = map[string]float64{}, map[string]float64{}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	first := true
	for sc.Scan() {
		if first { // header: Date,SP500,Dividend,Earnings,...
			first = false
			continue
		}
		cols := strings.Split(sc.Text(), ",")
		if len(cols) < 3 {
			continue
		}
		t, err := time.Parse("2006-01-02", cols[0])
		if err != nil {
			continue
		}
		k := t.Format("2006-01")
		if p, err := strconv.ParseFloat(cols[1], 64); err == nil {
			price[k] = p
		}
		if d, err := strconv.ParseFloat(cols[2], 64); err == nil {
			div[k] = d
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("shiller: %v", err)
	}
	if len(price) < 1000 {
		log.Fatalf("shiller: only %d monthly rows", len(price))
	}
	return price, div
}

func report(id string, pts []marketdata.Point) {
	if len(pts) == 0 {
		log.Fatalf("%s: empty", id)
	}
	first, last := pts[0], pts[len(pts)-1]
	yrs := last.Date.Sub(first.Date).Hours() / 24 / 365.25
	cagr := math.Pow(last.Close/first.Close, 1/yrs) - 1
	log.Printf("%-13s %5d points  %s..%s  CAGR %.2f%%/yr",
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
