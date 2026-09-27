// Command gen-french-refdata refreshes the two bundled series drawn from the Ken
// French Data Library besides the total-market factor (which
// cmd/gen-usmkt-refdata owns, with a validation of its own):
//
//   - USSCV-USD      US small-cap value total return, daily from 1963-07: the
//     value-weighted SMALL HiBM portfolio of the 2x3 size and book-to-market
//     sorts, cumulated from 100. It extends DFSVX (hence ZPRV) before 1993,
//     and the FIRE book's small-value plates read it directly. An academic
//     portfolio, so GROSS of costs: simgen charges USSCVGrossCost for that,
//     never this command.
//   - DEVEXUS-DAILY  developed-ex-US equity total return, daily from 1990-07:
//     the library's developed-ex-US market factor plus its bill rate, based at
//     100 on the first day. A DAILY SHAPE series only: simgen reads its
//     intra-month path behind the DEVEXUS-USD monthly anchors, and its levels
//     are not authoritative.
//
// Both files were refreshed by hand until 2026-09 and had frozen at 2026-05.
//
// The library REVISES its history as the underlying databases are corrected
// (CRSP, Compustat and Bloomberg rebuild past returns each year), so a refresh
// is not required to reproduce the bundled file point for point, only to stay
// the same series: daily returns correlated above 0.999 with the bundled ones
// over their whole overlap, and the level at the bundled file's last day within
// 3 % of it. Each series is also graded against an outside reference before it
// is written: USSCV-USD against the real NAVs of the fund it extends (the same
// trade minus a wrapper: monthly correlation, and a level ABOVE the fund by no
// more than a gross factor's costs), DEVEXUS-DAILY against the bundled MSCI
// World ex USA net monthly anchors it shapes. A refusal is loud and writes
// nothing.
//
// Usage: gen-french-refdata [-dir path] [-dry]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/bpineau/pofo/cmd/internal/refgen"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

const library = "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/"

// The two series, their sources and their references.
const (
	outSCV   = "USSCV-USD"
	scvZip   = library + "6_Portfolios_2x3_Daily_CSV.zip"
	scvCol   = "SMALL HiBM"
	scvStart = "1963-07-01" // the standard reliable-value start: earlier book equity is hand-collected
	scvFund  = "DFSVX"      // DFA US Small Cap Value, the fund the series extends

	outDev    = "DEVEXUS-DAILY"
	devZip    = library + "Developed_ex_US_3_Factors_daily_CSV.zip"
	devAnchor = "DEVEXUS-USD" // the MSCI World ex USA net monthly anchors it shapes
)

// Validation bounds. Each one states a property of the object, not a tolerance
// fitted to what the source prints today.
const (
	minRevisionCorr = 0.999                // a revised history is still the same series, day by day
	maxRevisionGap  = 0.03                 // ... and ends within 3 % of the bundled level
	maxMove         = 0.25                 // no diversified equity portfolio moved a quarter in a day
	maxStale        = 200 * 24 * time.Hour // the library republishes monthly, with a lag
	minFundCorr     = 0.95                 // measured 0.980 over 1993-2026
	fundGapLo       = -0.5                 // pts/yr: a gross factor does not lag its fund (the price list alone costs ~0.4)
	fundGapHi       = 2.5                  // pts/yr: measured +1.02, costs above that would not be the same trade
	minAnchorCorr   = 0.95                 // the same developed-ex-US market, two index providers
	anchorVolLo     = 0.85                 // monthly volatility ratio to the anchors
	anchorVolHi     = 1.15
)

func main() {
	dir := flag.String("dir", "pkg/datasets/refdata", "directory the reference series are written to")
	dry := flag.Bool("dry", false, "download and validate, write nothing")
	flag.Parse()

	scv, err := smallValue()
	if err != nil {
		log.Fatalf("%s: %v", outSCV, err)
	}
	dev, err := devExUS()
	if err != nil {
		log.Fatalf("%s: %v", outDev, err)
	}
	for _, s := range []struct {
		id  string
		pts []marketdata.Point
	}{{outSCV, scv}, {outDev, dev}} {
		log.Printf("%s: %d days %s → %s", s.id, len(s.pts),
			s.pts[0].Date.Format(time.DateOnly), s.pts[len(s.pts)-1].Date.Format(time.DateOnly))
		if err := checkShape(s.pts, time.Now()); err != nil {
			log.Fatalf("refusing to write %s: %v", s.id, err)
		}
		msg, err := checkRevision(*dir, s.id, s.pts)
		if err != nil {
			log.Fatalf("refusing to write %s: %v", s.id, err)
		}
		log.Printf("%s vs the bundled file: %s", s.id, msg)
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Logf = func(format string, args ...any) { log.Printf(format, args...) }
	fund, err := client.Fetch(context.Background(), scvFund, time.Time{})
	if err != nil {
		log.Fatalf("%s: %v", scvFund, err)
	}
	scvCheck, err := checkAgainstFund(scv, fund.Points)
	if err != nil {
		log.Fatalf("refusing to write %s: %v", outSCV, err)
	}
	log.Printf("%s vs %s: %s", outSCV, scvFund, scvCheck)

	anchor, ok, err := marketdata.ReadSimdata(*dir, devAnchor)
	if err != nil || !ok {
		log.Fatalf("%s: the bundled anchors are unreadable (ok=%v): %v", devAnchor, ok, err)
	}
	devCheck, err := checkAgainstAnchors(dev, anchor.Points)
	if err != nil {
		log.Fatalf("refusing to write %s: %v", outDev, err)
	}
	log.Printf("%s vs %s: %s", outDev, devAnchor, devCheck)

	if *dry {
		log.Printf("dry run: nothing written")
		return
	}
	for _, w := range []struct {
		h   refgen.Header
		pts []marketdata.Point
	}{
		{refgen.Header{ID: outSCV, Name: "US small-cap value total return (USD, daily)", Source: "Ken French Data Library, " +
			"6 Portfolios formed on Size and Book-to-Market (2x3) [Daily], value-weighted SMALL HiBM (small x high " +
			"book-to-market), daily returns cumulated to a total-return index from 1963-07 (the standard reliable-value " +
			"start; pre-1963 book equity is hand-collected). " + scvZip + " . Gross of fees/costs. Extends DFSVX (hence " +
			"ZPRV) before its 1993 inception. Regenerated by cmd/gen-french-refdata (make french-refdata); the library " +
			"revises its history, so a refresh moves old levels slightly. Validation: " + scvCheck + "."}, scv},
		{refgen.Header{ID: outDev, Name: "Developed markets ex-US equity total return, DAILY SHAPE series", Source: "Ken French " +
			"\"Developed ex US 3 Factors [Daily]\" (" + devZip + "), market TR = Mkt-RF + RF, cumulated from 100. Universe " +
			"close to (not identical to) MSCI World ex USA and gross of withholding: LEVELS ARE NOT AUTHORITATIVE. Used only " +
			"as the intra-month daily shape behind the DEVEXUS-USD monthly net anchors (simgen dailyShape/anchorShape). " +
			"Regenerated by cmd/gen-french-refdata (make french-refdata). Validation: " + devCheck + "."}, dev},
	} {
		if err := refgen.Write(*dir, w.h, w.pts); err != nil {
			log.Fatalf("write %s: %v", w.h.ID, err)
		}
		log.Printf("wrote %s (%d points)", filepath.Join(*dir, w.h.ID+".csv"), len(w.pts))
	}
}

// smallValue builds USSCV-USD: the value-weighted SMALL HiBM daily returns from
// scvStart, compounded from 100 the day before the first one.
func smallValue() ([]marketdata.Point, error) {
	body, err := download(scvZip)
	if err != nil {
		return nil, err
	}
	dates, cols, err := refgen.FrenchTable(body, scvCol)
	if err != nil {
		return nil, err
	}
	start, _ := time.Parse(time.DateOnly, scvStart)
	var out []marketdata.Point
	level := 100.0
	for i, d := range dates {
		if d.Before(start) {
			continue
		}
		level *= 1 + cols[0][i]/100
		out = append(out, marketdata.Point{Date: d, Close: level})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no return from %s on", scvStart)
	}
	return out, nil
}

// devExUS builds DEVEXUS-DAILY: the market factor plus the bill rate, based at
// 100 ON the first day (the file's convention since it was first bundled), so
// the first day's own return is not compounded.
func devExUS() ([]marketdata.Point, error) {
	body, err := download(devZip)
	if err != nil {
		return nil, err
	}
	dates, cols, err := refgen.FrenchTable(body, "Mkt-RF", "RF")
	if err != nil {
		return nil, err
	}
	out := make([]marketdata.Point, len(dates))
	level := 100.0
	for i, d := range dates {
		if i > 0 {
			level *= 1 + (cols[0][i]+cols[1][i])/100
		}
		out[i] = marketdata.Point{Date: d, Close: level}
	}
	return out, nil
}

func download(url string) (string, error) {
	raw, err := refgen.Get(url)
	if err != nil {
		return "", err
	}
	return refgen.FrenchCSV(raw)
}

// checkShape refuses a series that cannot be a daily equity history: out of
// order, non-positive, carrying a day no diversified portfolio had, or stale.
func checkShape(pts []marketdata.Point, now time.Time) error {
	for i, p := range pts {
		if p.Close <= 0 {
			return fmt.Errorf("%s: non-positive level %v", p.Date.Format(time.DateOnly), p.Close)
		}
		if i == 0 {
			continue
		}
		if !p.Date.After(pts[i-1].Date) {
			return fmt.Errorf("%s: dates out of order", p.Date.Format(time.DateOnly))
		}
		if move := math.Abs(p.Close/pts[i-1].Close - 1); move > maxMove {
			return fmt.Errorf("%s: %.1f %% in one day", p.Date.Format(time.DateOnly), move*100)
		}
	}
	last := pts[len(pts)-1].Date
	if age := now.Sub(last); age > maxStale {
		return fmt.Errorf("last return %s is %.0f days old: frozen source?", last.Format(time.DateOnly), age.Hours()/24)
	}
	return nil
}

// checkRevision grades a refresh against the bundled file it replaces. The
// library revises old returns, so the test is that it is still the same
// series: it covers every bundled day, its daily returns correlate above
// minRevisionCorr with the bundled ones, and it ends the bundled span within
// maxRevisionGap of the bundled level. A missing bundled file is not an error.
func checkRevision(dir, id string, pts []marketdata.Point) (string, error) {
	old, ok, err := marketdata.ReadSimdata(dir, id)
	if err != nil {
		return "", err
	}
	if !ok {
		return "no bundled file to compare with", nil
	}
	byDate := make(map[time.Time]float64, len(pts))
	for _, p := range pts {
		byDate[p.Date] = p.Close
	}
	var ro, rn []float64
	for i, p := range old.Points {
		v, ok := byDate[p.Date]
		if !ok {
			return "", fmt.Errorf("%s is no longer in the source", p.Date.Format(time.DateOnly))
		}
		if i > 0 {
			ro = append(ro, p.Close/old.Points[i-1].Close-1)
			rn = append(rn, v/byDate[old.Points[i-1].Date]-1)
		}
	}
	last := old.Last()
	c, gap := metrics.Corr(ro, rn), byDate[last.Date]/last.Close-1
	out := fmt.Sprintf("%d bundled days, daily-return correlation %.5f, level at %s %+.2f %%, %d new days",
		old.Len(), c, last.Date.Format(time.DateOnly), gap*100, len(pts)-old.Len())
	if c < minRevisionCorr {
		return "", fmt.Errorf("%s: a daily-return correlation of %.5f with the bundled file is another series", out, c)
	}
	if math.Abs(gap) > maxRevisionGap {
		return "", fmt.Errorf("%s: the revision moves the level by more than %.0f %%", out, maxRevisionGap*100)
	}
	return out, nil
}

// checkAgainstFund grades USSCV-USD on the real fund it extends, over their
// whole overlap: the same path (monthly correlation) and a level above the fund
// by no more than a gross factor's costs.
func checkAgainstFund(pts, fund []marketdata.Point) (string, error) {
	a, b, months := pairMonths(pts, fund)
	if months < 300 {
		return "", fmt.Errorf("only %d months shared with %s", months, scvFund)
	}
	c, gap := metrics.Corr(a, b), (annualized(a)-annualized(b))*100
	out := fmt.Sprintf("%d months, monthly correlation %.3f, %+.2f pts/yr above the fund (gross of its costs)", months, c, gap)
	if c < minFundCorr {
		return "", fmt.Errorf("%s: below the %.2f floor", out, minFundCorr)
	}
	if gap < fundGapLo || gap > fundGapHi {
		return "", fmt.Errorf("%s: outside the %+.1f..%+.1f pts/yr a gross factor can sit from its fund", out, fundGapLo, fundGapHi)
	}
	return out, nil
}

// checkAgainstAnchors grades DEVEXUS-DAILY on the monthly anchors it shapes:
// only the path matters (its levels are not authoritative), so the monthly
// correlation and the volatility ratio are what is checked.
func checkAgainstAnchors(pts, anchors []marketdata.Point) (string, error) {
	a, b, months := pairMonths(pts, anchors)
	if months < 300 {
		return "", fmt.Errorf("only %d months shared with %s", months, devAnchor)
	}
	c, ratio := metrics.Corr(a, b), stdev(a)/stdev(b)
	out := fmt.Sprintf("%d months, monthly correlation %.3f, volatility ratio %.2f", months, c, ratio)
	if c < minAnchorCorr {
		return "", fmt.Errorf("%s: below the %.2f floor", out, minAnchorCorr)
	}
	if ratio < anchorVolLo || ratio > anchorVolHi {
		return "", fmt.Errorf("%s: outside %.2f..%.2f", out, anchorVolLo, anchorVolHi)
	}
	return out, nil
}

// pairMonths returns the calendar-month returns both series can compute, in
// month order: a month counts when both quote it and the month before, each
// read at its last point of the month.
func pairMonths(x, y []marketdata.Point) (rx, ry []float64, n int) {
	mx, my := monthEnds(x), monthEnds(y)
	first, last := math.MaxInt, math.MinInt
	for k := range mx {
		first, last = min(first, k), max(last, k)
	}
	for k := first + 1; k <= last; k++ {
		x0, ok1 := mx[k-1]
		x1, ok2 := mx[k]
		y0, ok3 := my[k-1]
		y1, ok4 := my[k]
		if ok1 && ok2 && ok3 && ok4 {
			rx = append(rx, x1/x0-1)
			ry = append(ry, y1/y0-1)
		}
	}
	return rx, ry, len(rx)
}

// monthEnds keys each series' last level of a calendar month by year*12+month.
func monthEnds(pts []marketdata.Point) map[int]float64 {
	out := make(map[int]float64, len(pts)/20+1)
	for _, p := range pts {
		out[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close
	}
	return out
}

// annualized is the compound annual rate of a run of monthly returns.
func annualized(r []float64) float64 {
	growth := 1.0
	for _, x := range r {
		growth *= 1 + x
	}
	return math.Pow(growth, 12/float64(len(r))) - 1
}

func stdev(xs []float64) float64 {
	m := metrics.Mean(xs)
	var v float64
	for _, x := range xs {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v / float64(len(xs)-1))
}
