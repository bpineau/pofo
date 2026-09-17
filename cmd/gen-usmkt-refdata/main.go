// Command gen-usmkt-refdata rebuilds USMKT-USD, the daily total return of the
// WHOLE US equity market since 1926-07, from the Fama/French market factor
// published by the Ken French Data Library.
//
// Why the series exists: a total-market fund (VTI and its share classes) has
// no deep history of its own, and the S&P 500 is not a stand-in for it. The
// S&P 500 is the large-cap segment, so a total-market backcast built on it
// drops the mid- and small-cap completion the fund holds, and with it the size
// premium of the eras where it was largest. Measured on the two funds' own
// NAVs over 1992-04 to 2001-06, the whole market ran 1.05 pts/yr BELOW the
// S&P 500; over 1962-2001 it ran above. The sign is not the point: the object
// is different, and a completion tilt worth a point a year in either direction
// has no business being assumed away.
//
// What the source is: "Fama/French 3 Factors [Daily]" carries Mkt-RF, the
// excess return of the CRSP value-weighted portfolio of all NYSE, AMEX and
// NASDAQ common stocks, and RF, the one-month Treasury bill return over the
// same day. Their sum is the market's own total return, which this command
// cumulates into a base-100 daily index. It is an ACADEMIC FACTOR and
// therefore gross of every cost of holding it: what it owes before it may
// stand in for a fund is stated and charged by pkg/simgen's longBackFee, never
// here.
//
// Validation happens BEFORE the file is written, against two references that
// know nothing about the factor: the bundled S&P 500 total return (SP500-USD,
// calendar-year returns over the whole span, the two indices being the same
// market minus a completion tail) and the real NAVs of the total-market fund
// the series extends (VTSMX, monthly over the whole overlap). A refusal is
// loud and writes nothing.
//
// Usage: gen-usmkt-refdata [-dir path] [-dry]
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// source is the zip of the daily three-factor file. The same library serves
// the repository's other two French series (USSCV-USD, DEVEXUS-DAILY).
const source = "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_Factors_daily_CSV.zip"

// seriesID is the bundled reference this command owns; fundRef is the real
// total-market fund it extends, and the reference its level and its path are
// checked against. That fund is the Investor share class of the fund the
// target ETF is a share class of, so the two are the same portfolio at
// different price lists. indexRef is the large-cap segment of the same market,
// already bundled: the one reference that needs no network.
const (
	seriesID = "USMKT-USD"
	fundRef  = "VTSMX"
	indexRef = "SP500-USD"
)

func main() {
	dir := flag.String("dir", "pkg/datasets/refdata", "directory the reference series is written to")
	dry := flag.Bool("dry", false, "download and validate, write nothing")
	flag.Parse()

	raw, err := download(source)
	if err != nil {
		log.Fatalf("download: %v", err)
	}
	points, err := parse(raw)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	log.Printf("%s: %d daily returns %s → %s", seriesID, len(points),
		points[0].Date.Format("2006-01-02"), points[len(points)-1].Date.Format("2006-01-02"))
	if err := checkShape(points); err != nil {
		log.Fatalf("refusing to write %s: %v", seriesID, err)
	}

	index, err := bundled(indexRef)
	if err != nil {
		log.Fatalf("%s: %v", indexRef, err)
	}
	years, err := checkAgainstIndex(points, index.Points)
	if err != nil {
		log.Fatalf("refusing to write %s: %v", seriesID, err)
	}
	log.Printf("vs %s: %s", indexRef, years)

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Logf = func(format string, args ...any) { log.Printf(format, args...) }
	fund, err := client.Fetch(context.Background(), fundRef, time.Time{})
	if err != nil {
		log.Fatalf("%s: %v", fundRef, err)
	}
	months, err := checkAgainstFund(points, fund.Points)
	if err != nil {
		log.Fatalf("refusing to write %s: %v", seriesID, err)
	}
	log.Printf("vs %s: %s", fundRef, months)

	if *dry {
		log.Printf("dry run: %d points, nothing written", len(points))
		return
	}
	if err := marketdata.WriteSimdata(*dir, &marketdata.SimdataFile{
		ID:   seriesID,
		Name: "US total market total return (USD, daily)",
		Method: "Ken French Data Library, Fama/French 3 Factors [Daily]: the CRSP value-weighted return of all " +
			"NYSE/AMEX/NASDAQ common stocks (Mkt-RF) plus the one-month Treasury bill return of the same day (RF), " +
			"cumulated to a total-return index from 1926-07. " + source +
			" . An academic factor, so GROSS of fees, commissions and spreads: what it owes before standing in for a " +
			"fund is charged by pkg/simgen's longBackFee. The deep proxy behind the total-US-market recipe (VTI).",
		Validation: fmt.Sprintf("%d daily returns %s → %s; vs %s: %s; vs %s: %s", len(points),
			points[0].Date.Format("2006-01-02"), points[len(points)-1].Date.Format("2006-01-02"),
			indexRef, years, fundRef, months),
		Generated: time.Now().UTC().Format("2006-01-02"),
		Points:    points,
	}); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s/%s.csv (%d points)", *dir, seriesID, len(points))
	log.Printf("rebuild (make simdata) to carry the total-market recipes back to %s", points[0].Date.Format("2006-01"))
}

// download GETs a URL with a browser User-Agent, which this host requires.
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

// parse turns the zipped CSV into the cumulated base-100 index. The file holds
// a few prose lines, then a header naming the factor columns, then one row per
// trading day with returns in PERCENT.
func parse(raw []byte) ([]marketdata.Point, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	if len(zr.File) != 1 {
		return nil, fmt.Errorf("zip holds %d files, want 1", len(zr.File))
	}
	f, err := zr.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, 64<<20))
	if err != nil {
		return nil, err
	}
	return cumulate(string(body))
}

// cumulate reads the CSV body: it locates the two columns it needs by NAME (the
// library has added columns before) and compounds their sum from a base of 100
// the day before the first return.
func cumulate(body string) ([]marketdata.Point, error) {
	excess, bill := -1, -1
	var out []marketdata.Point
	level := 100.0
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Split(strings.TrimRight(line, "\r"), ",")
		for i, f := range fields {
			switch strings.TrimSpace(f) {
			case "Mkt-RF":
				excess = i
			case "RF":
				bill = i
			}
		}
		if excess < 0 || bill < 0 || len(fields) <= max(excess, bill) {
			continue
		}
		date, ok := day(fields[0])
		if !ok {
			continue
		}
		mkt, err1 := strconv.ParseFloat(strings.TrimSpace(fields[excess]), 64)
		rf, err2 := strconv.ParseFloat(strings.TrimSpace(fields[bill]), 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if mkt <= -99 || rf <= -99 { // the library's missing-value sentinel
			return nil, fmt.Errorf("%s: missing value in the source", date.Format("2006-01-02"))
		}
		level *= 1 + (mkt+rf)/100
		out = append(out, marketdata.Point{Date: date, Close: level})
	}
	if excess < 0 || bill < 0 {
		return nil, fmt.Errorf("no Mkt-RF/RF header in the source")
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no daily return in the source")
	}
	return out, nil
}

// day parses the source's YYYYMMDD key, rejecting anything else (the monthly
// files of the same library key on YYYYMM, and the prose lines on nothing).
func day(field string) (time.Time, bool) {
	s := strings.TrimSpace(field)
	if len(s) != 8 {
		return time.Time{}, false
	}
	t, err := time.Parse("20060102", s)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

// bundled reads an embedded reference series.
func bundled(id string) (*marketdata.Series, error) {
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not bundled")
	}
	return s, nil
}

// Validation bounds. Each one states a property of the OBJECT, not a
// tolerance fitted to what the source currently prints.
const (
	minPoints   = 20000                // ~80 years of trading days
	maxStart    = "1926-08-01"         // the library's own first month is 1926-07
	maxMove     = 0.30                 // no US market day moved more; 1987-10-19 lost 0.20
	maxStale    = 200 * 24 * time.Hour // the library republishes monthly, with a lag
	minYears    = 90                   // calendar years shared with the large-cap index
	minYearCorr = 0.95                 // same market, minus a completion tail
	maxYearGap  = 1.5                  // pts/yr, mean calendar-year return difference
	minMonths   = 300                  // months shared with the real fund
	minFundCorr = 0.99                 // same portfolio, minus a wrapper
	maxFundGap  = 1.0                  // pts/yr, the factor's grossness ceiling
)

// checkShape refuses a series that cannot be the market's own daily history:
// too short, not starting at the library's first month, out of order,
// non-positive, carrying a day no market had, or stale.
func checkShape(points []marketdata.Point) error {
	if len(points) < minPoints {
		return fmt.Errorf("only %d daily returns, want at least %d", len(points), minPoints)
	}
	start, _ := time.Parse("2006-01-02", maxStart)
	if points[0].Date.After(start) {
		return fmt.Errorf("starts %s, after the library's own %s", points[0].Date.Format("2006-01-02"), maxStart)
	}
	for i, p := range points {
		if p.Close <= 0 {
			return fmt.Errorf("%s: non-positive level %v", p.Date.Format("2006-01-02"), p.Close)
		}
		if i == 0 {
			continue
		}
		if !p.Date.After(points[i-1].Date) {
			return fmt.Errorf("%s: dates out of order", p.Date.Format("2006-01-02"))
		}
		if move := math.Abs(p.Close/points[i-1].Close - 1); move > maxMove {
			return fmt.Errorf("%s: %.1f%% in one day", p.Date.Format("2006-01-02"), move*100)
		}
	}
	if age := time.Since(points[len(points)-1].Date); age > maxStale {
		return fmt.Errorf("last return %s is %.0f days old: frozen source?",
			points[len(points)-1].Date.Format("2006-01-02"), age.Hours()/24)
	}
	return nil
}

// checkAgainstIndex compares calendar-year returns with the bundled large-cap
// total return. The whole market and the S&P 500 are the same market minus a
// completion tail, so their years must move together and their mean must be
// close; a mis-parsed column, a percent/fraction slip or a lost era shows up
// as either check failing by a mile.
func checkAgainstIndex(points, index []marketdata.Point) (string, error) {
	a, b := yearEnds(points), yearEnds(index)
	ra, rb, years := commonReturns(a, b)
	if years < minYears {
		return "", fmt.Errorf("%d calendar years shared with %s, want at least %d", years, indexRef, minYears)
	}
	c := correlation(ra, rb)
	gap := (mean(ra) - mean(rb)) * 100
	out := fmt.Sprintf("%d calendar years, return correlation %.4f, mean year %+.2f pts", years, c, gap)
	if c < minYearCorr {
		return "", fmt.Errorf("calendar-year correlation %.4f with %s, want at least %.2f", c, indexRef, minYearCorr)
	}
	if math.Abs(gap) > maxYearGap {
		return "", fmt.Errorf("mean calendar year %+.2f pts vs %s, want within %.1f", gap, indexRef, maxYearGap)
	}
	return out, nil
}

// checkAgainstFund compares the factor with the real NAVs of the fund it
// extends, over their whole overlap. Two properties are asked of it: the same
// path (monthly correlation, the fund holding the same portfolio) and a level
// ABOVE the fund by less than a point a year, the factor being gross of the
// fee, commissions and spreads the fund actually paid. Below the fund would
// mean the reconstruction is not the market; far above it would mean the
// column is not the market either.
func checkAgainstFund(points, fund []marketdata.Point) (string, error) {
	a, b := monthEnds(points), monthEnds(fund)
	ra, rb, months := commonReturns(a, b)
	if months < minMonths {
		return "", fmt.Errorf("%d months shared with %s, want at least %d", months, fundRef, minMonths)
	}
	c := correlation(ra, rb)
	from, to := overlap(a, b)
	ga, gb := cagr(points, from, to), cagr(fund, from, to)
	out := fmt.Sprintf("%d months %s → %s, monthly correlation %.4f, CAGR %.2f vs %.2f %%/yr (%+.2f pts/yr gross)",
		months, from.Format("2006-01"), to.Format("2006-01"), c, ga*100, gb*100, (ga-gb)*100)
	if c < minFundCorr {
		return "", fmt.Errorf("monthly correlation %.4f with %s, want at least %.2f", c, fundRef, minFundCorr)
	}
	if ga < gb {
		return "", fmt.Errorf("%+.2f pts/yr BELOW %s: a gross factor cannot lag the fund it extends", (ga-gb)*100, fundRef)
	}
	if (ga-gb)*100 > maxFundGap {
		return "", fmt.Errorf("%+.2f pts/yr above %s, want within %.1f", (ga-gb)*100, fundRef, maxFundGap)
	}
	return out, nil
}

// yearEnds and monthEnds reduce a daily series to its last level per calendar
// year (resp. month), keyed so two series' periods can be matched.
func yearEnds(points []marketdata.Point) map[int]float64 {
	out := make(map[int]float64, len(points)/250+1)
	for _, p := range points {
		out[p.Date.Year()] = p.Close
	}
	return out
}

func monthEnds(points []marketdata.Point) map[int]float64 {
	out := make(map[int]float64, len(points)/20+1)
	for _, p := range points {
		out[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close
	}
	return out
}

// commonReturns pairs the period returns both series can compute, in period
// order. A period counts only when both series quote it AND the one before.
func commonReturns(a, b map[int]float64) (ra, rb []float64, n int) {
	keys := make([]int, 0, len(a))
	for k := range a {
		if _, ok := b[k]; ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	for _, k := range keys {
		pa, oka := a[k-1]
		pb, okb := b[k-1]
		if !oka || !okb || pa <= 0 || pb <= 0 {
			continue
		}
		ra = append(ra, a[k]/pa-1)
		rb = append(rb, b[k]/pb-1)
	}
	return ra, rb, len(ra)
}

// overlap is the window two period-keyed series share, as dates.
func overlap(a, b map[int]float64) (from, to time.Time) {
	lo, hi := math.MaxInt, math.MinInt
	for k := range a {
		if _, ok := b[k]; !ok {
			continue
		}
		lo, hi = min(lo, k), max(hi, k)
	}
	return time.Date(lo/12, time.Month(lo%12+1), 1, 0, 0, 0, 0, time.UTC),
		time.Date(hi/12, time.Month(hi%12+1), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1)
}

// cagr is the annualized growth of a daily series over [from, to].
func cagr(points []marketdata.Point, from, to time.Time) float64 {
	var first, last marketdata.Point
	for _, p := range points {
		if p.Date.Before(from) || p.Date.After(to) {
			continue
		}
		if first.Close == 0 {
			first = p
		}
		last = p
	}
	years := last.Date.Sub(first.Date).Hours() / 24 / 365.25
	if first.Close <= 0 || years <= 0 {
		return 0
	}
	return math.Pow(last.Close/first.Close, 1/years) - 1
}

func mean(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func correlation(xs, ys []float64) float64 {
	mx, my := mean(xs), mean(ys)
	var cov, vx, vy float64
	for i := range xs {
		cov += (xs[i] - mx) * (ys[i] - my)
		vx += (xs[i] - mx) * (xs[i] - mx)
		vy += (ys[i] - my) * (ys[i] - my)
	}
	if vx <= 0 || vy <= 0 {
		return 0
	}
	return cov / math.Sqrt(vx*vy)
}
