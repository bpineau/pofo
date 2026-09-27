// Command gen-bcom-refdata writes pkg/datasets/refdata/BCOM-ER-USD.csv, the
// daily Bloomberg Commodity Index (BCOM) since its first print, 100 on
// 1991-01-02.
//
// BCOM is an EXCESS return index: the rolled front futures of 23 commodities,
// no collateral. The commodity backcast (simgen's icomRecipe, the iShares
// Bloomberg Commodity swap ETF IE00BDFL4P12) funds it with the 3-month T-bill
// to rebuild the total return the fund tracks, over the 26 years before the
// fund's own quotes begin in 2017-07.
//
// WHY A BUNDLED FILE. The recipe used to read Yahoo's ^BCOM, which Yahoo
// withdrew in 2026-09 together with the whole BCOM family (HTTP 404 on every
// symbol form, ^BCOMTR and the subindices included). A warm quote cache hid
// it; on a fresh machine the identifier fell through to the FT full-text
// search and resolved to an unrelated GBP exchange-traded commodity quoting
// since 2016, which would have been served as the index.
//
// THE SOURCE is the Financial Times' chart API, which carries the index under
// its own tearsheet symbol BCOM:IOM from 1991-01-02 on. The instrument id is
// looked up by that exact symbol on every run and the answer's name must be
// the index's, so a renumbered or renamed listing stops the refresh instead of
// feeding it. Levels are carried as published, with two exceptions, each
// logged: FT's record holds a few one-session round trips no index could make
// (2011-09-05 read 140.40 between 162.51 and 161.22; 2023-12-25 read 1496.35
// between 99.35 and 99.91), which marketdata.FindSpikes names and this command
// drops; and FT reprints the previous level, rounded to the cent, on US
// holidays when the index was not calculated, which dropReprints removes
// (169 rows in 2026-09) because a funded composite would accrue a session of
// cash on each of them.
//
// VALIDATED before anything is written, each check naming a failure a
// completed download would hide:
//
//   - the first print is 100 on 1991-01-02, the index's base;
//   - at most maxSpikes round trips were dropped, no remaining session moves
//     beyond maxMove (the record's largest is -8.8 %, 1991-01-17, the first
//     day of the Gulf War air campaign), and every level stays in a band the
//     index never left (59.48 on 2020-03-18 to 237.95 on 2008-07-02);
//   - the same source's TOTAL return index (BCOMTR:IOM, from 2009-06) must
//     exceed it each calendar year by the 3-month T-bill yield of that year
//     (the bundled TBILL-3M), within maxCarryGap: it is the identity the
//     recipe relies on, measured within 0.29 point over 2010-2025;
//   - BCOMTR's 2025 return must be the +15.77 % Bloomberg publishes in its
//     BCOM factsheet (August 31, 2026 edition), which ties the pair to the
//     publisher rather than to FT alone;
//   - no frozen run, a last print no older than maxAge days, and, when the
//     bundle already holds the file, its whole history reproduced (an index
//     level is never revised).
//
// Measured 2026-09-27 against the Yahoo ^BCOM copy the recipe was built on
// (1991-01-02 to 2026-07-17): 8922 common sessions, 11 of them apart by more
// than 0.01 % (at most 0.48 %, 2015-02-02, where Yahoo repeated the previous
// close), daily-return correlation 0.99995 and the same compound return to
// the fourth decimal. Funded with the bundled T-bill, the file reproduces the
// publisher's BCOMTR figures within 0.01 point a year (pkg/datasets/golden,
// TestGoldenBCOM), and the commodity backcast built on it follows FT's BCOMTR
// over 2009-06 to 2017-07 at -0.016 point a year with a 0.015 % tracking
// error.
//
// Usage: gen-bcom-refdata [-dir path] [-dry]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/bpineau/pofo/cmd/internal/refgen"
	"github.com/bpineau/pofo/pkg/marketdata"
)

const (
	ftBase = "https://markets.ft.com"
	outID  = "BCOM-ER-USD"

	// erSymbol and trSymbol are FT's tearsheet symbols for the index and its
	// total return sibling; erName and trName the names FT must answer with.
	erSymbol, erName = "BCOM:IOM", "Bloomberg Commodity Index"
	trSymbol, trName = "BCOMTR:IOM", "Bloomberg Commodity Index Total Return"

	// firstDay and baseLevel are the index's first print.
	firstDay  = "1991-01-02"
	baseLevel = 100.0

	// maxSpikes bounds the round trips dropped: two measured, more would say
	// the source degraded rather than slipped.
	maxSpikes = 4
	// maxMove is the largest one-session move a clean record may keep.
	maxMove = 0.15
	// lowest and highest bound every level (the record: 59.48 and 237.95).
	lowest, highest = 40.0, 400.0
	// maxCarryGap bounds, in points, how far a year's total return over the
	// excess return may sit from that year's mean 3-month T-bill yield. The
	// residual is the difference between the bill's discount yield and the
	// index's accrual, 0.29 point at most over 2010-2025.
	maxCarryGap = 0.6
	// published2025 is BCOMTR's 2025 return in the publisher's factsheet, in
	// percent, and published2025Tol the tolerance on it (two decimals).
	published2025, published2025Tol = 15.77, 0.02
	// maxRun is the longest run of one repeated level: 4 in the record, the
	// four reprints of the markets' closure after 2001-09-11.
	maxRun = 5
	// reprintTol is how close to the previous level a print must sit to be
	// read as a reprint: FT's holiday reprints round the level to the cent.
	reprintTol = 0.01
	// maxAge bounds the age of the last print in days.
	maxAge = 10
	// revisable is how many of the bundled file's last levels may differ.
	revisable = 5
)

// source is the file's "# source:" header.
const source = "Bloomberg Commodity Index (BCOM, excess return, USD), daily closing levels as published by the " +
	"Financial Times (" + ftBase + ", tearsheet " + erSymbol + "); base 100 on 1991-01-02. Carried as published " +
	"except one-session round trips (marketdata.FindSpikes) and holiday reprints (a level within a cent of the " +
	"previous one), dropped. Regenerated by cmd/gen-bcom-refdata " +
	"(make bcom-refdata), which checks it against BCOMTR less the 3-month T-bill and refuses a refresh that " +
	"does not reproduce this file's history."

func main() {
	dir := flag.String("dir", "pkg/datasets/refdata", "output directory (also where TBILL-3M is read)")
	dry := flag.Bool("dry", false, "download and validate, write nothing")
	flag.Parse()

	er, err := fetch(ftBase, erSymbol, erName)
	if err != nil {
		log.Fatalf("%s: %v", erSymbol, err)
	}
	if err := frozen(er); err != nil {
		log.Fatalf("refusing to write %s: %v", outID, err)
	}
	er, dropped := dropSpikes(er)
	for _, s := range dropped {
		log.Printf("dropped %s: a one-session round trip (%+.1f %% then %+.1f %%)", s.Date.Format(time.DateOnly), 100*s.In, 100*s.Out)
	}
	er, reprints := dropReprints(er)
	log.Printf("dropped %d reprints (a level within %g index point of the previous one)", reprints, reprintTol)
	log.Printf("%s: %d levels, %s to %s", outID, len(er), er[0].Date.Format(time.DateOnly), er[len(er)-1].Date.Format(time.DateOnly))
	if err := check(er, len(dropped), time.Now()); err != nil {
		log.Fatalf("refusing to write %s: %v", outID, err)
	}
	tr, err := fetch(ftBase, trSymbol, trName)
	if err != nil {
		log.Fatalf("%s: %v", trSymbol, err)
	}
	tr, _ = dropSpikes(tr)
	bill, ok, err := marketdata.ReadSimdata(*dir, "TBILL-3M")
	if err != nil || !ok {
		log.Fatalf("reading TBILL-3M in %s: %v (found %v)", *dir, err, ok)
	}
	if err := carry(er, tr, bill.Points); err != nil {
		log.Fatalf("refusing to write %s: %v", outID, err)
	}
	old, ok, err := marketdata.ReadSimdata(*dir, outID)
	switch {
	case err != nil:
		log.Fatalf("reading the bundled %s: %v", outID, err)
	case !ok:
		log.Printf("no bundled %s in %s to compare with", outID, *dir)
	default:
		msg, err := refgen.SameHistory(old.Points, er, revisable)
		if err != nil {
			log.Fatalf("refusing to write %s: it does not reproduce the bundled history: %v", outID, err)
		}
		log.Printf("history: %s", msg)
	}
	if *dry {
		log.Printf("dry run: nothing written")
		return
	}
	h := refgen.Header{ID: outID, Name: "Bloomberg Commodity Index (excess return, USD, daily)", Source: source}
	if err := refgen.Write(*dir, h, er); err != nil {
		log.Fatalf("write: %v", err)
	}
	log.Printf("wrote %s (%d levels)", filepath.Join(*dir, outID+".csv"), len(er))
}

// fetch looks symbol up on FT, requires the answer to carry that exact symbol
// under the expected name, and downloads its whole daily closing history.
func fetch(base, symbol, name string) ([]marketdata.Point, error) {
	xid, err := lookup(base, symbol, name)
	if err != nil {
		return nil, err
	}
	days := int(time.Since(time.Date(1990, 12, 1, 0, 0, 0, 0, time.UTC)).Hours()/24) + 2
	payload, err := json.Marshal(map[string]any{
		"days": days, "dataPeriod": "Day", "dataInterval": 1,
		"timeServiceFormat": "JSON", "returnDateType": "ISO8601",
		"elements": []map[string]any{{"Type": "price", "Symbol": xid}},
	})
	if err != nil {
		return nil, err
	}
	raw, err := refgen.Post(base+"/data/chartapi/series", "application/json", payload)
	if err != nil {
		return nil, err
	}
	return parse(raw)
}

// lookup returns FT's instrument id for symbol, refusing a hit that does not
// carry the symbol exactly or answers under another name.
func lookup(base, symbol, name string) (string, error) {
	raw, err := refgen.Get(base + "/data/searchapi/searchsecurities?query=" + url.QueryEscape(symbol))
	if err != nil {
		return "", err
	}
	var resp struct {
		Data struct {
			Security []struct{ Name, Symbol, Xid string } `json:"security"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("unreadable FT search: %w", err)
	}
	for _, s := range resp.Data.Security {
		if s.Symbol != symbol {
			continue
		}
		if !strings.EqualFold(s.Name, name) || s.Xid == "" {
			return "", fmt.Errorf("FT lists %s as %q (id %q), want %q", symbol, s.Name, s.Xid, name)
		}
		return s.Xid, nil
	}
	return "", fmt.Errorf("FT search has no %s", symbol)
}

// parse reads an FT chart API answer: a date list and, under the one element,
// a Close series of the same length. A day without a positive close is
// skipped rather than guessed.
func parse(raw []byte) ([]marketdata.Point, error) {
	var resp struct {
		Dates    []string
		Elements []struct {
			Currency        string
			ComponentSeries []struct {
				Type   string
				Values []*float64
			}
		}
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unreadable FT series: %w", err)
	}
	if len(resp.Elements) != 1 {
		return nil, fmt.Errorf("FT answered %d elements, want 1", len(resp.Elements))
	}
	if c := resp.Elements[0].Currency; c != "USD" {
		return nil, fmt.Errorf("FT quotes it in %q, want USD", c)
	}
	var closes []*float64
	for _, cs := range resp.Elements[0].ComponentSeries {
		if cs.Type == "Close" {
			closes = cs.Values
		}
	}
	if len(closes) != len(resp.Dates) {
		return nil, fmt.Errorf("FT answered %d closes for %d dates", len(closes), len(resp.Dates))
	}
	var out []marketdata.Point
	for i, label := range resp.Dates {
		if closes[i] == nil || *closes[i] <= 0 {
			continue
		}
		t, err := time.Parse("2006-01-02T15:04:05", label)
		if err != nil {
			return nil, fmt.Errorf("bad date %q", label)
		}
		if n := len(out); n > 0 && !t.After(out[n-1].Date) {
			return nil, fmt.Errorf("%s does not follow %s", label, out[n-1].Date.Format(time.DateOnly))
		}
		out = append(out, marketdata.Point{Date: t, Close: *closes[i]})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no close in the FT answer")
	}
	return out, nil
}

// dropSpikes removes the one-session round trips marketdata.FindSpikes names
// and returns them.
func dropSpikes(pts []marketdata.Point) ([]marketdata.Point, []marketdata.Spike) {
	spikes := marketdata.FindSpikes(&marketdata.Series{Points: pts})
	if len(spikes) == 0 {
		return pts, nil
	}
	bad := make(map[time.Time]bool, len(spikes))
	for _, s := range spikes {
		bad[s.Date] = true
	}
	out := make([]marketdata.Point, 0, len(pts)-len(spikes))
	for _, p := range pts {
		if !bad[p.Date] {
			out = append(out, p)
		}
	}
	return out, spikes
}

// dropReprints removes every level within reprintTol of the level kept before
// it, and returns how many it removed.
//
// FT prints the previous level, rounded to the cent, on some US holidays when
// the index was not calculated. Kept, such a row is a session with a zero
// return, and a zero return is not free downstream: a composite that funds the
// index with a cash rate accrues that rate per session, so the eighty or so
// reprints of the record once credited the commodity backcast a full point of
// T-bill it never earned. A genuine session that moved less than a cent is
// dropped with them, which costs nothing: its return folds into the next
// session's, and the cash leg's own calendar still carries the day.
func dropReprints(pts []marketdata.Point) ([]marketdata.Point, int) {
	out := make([]marketdata.Point, 0, len(pts))
	for _, p := range pts {
		if n := len(out); n > 0 && math.Abs(p.Close-out[n-1].Close) < reprintTol {
			continue
		}
		out = append(out, p)
	}
	return out, len(pts) - len(out)
}

// frozen refuses a record in which one level repeats for more than maxRun
// sessions: a feed that stopped, which dropReprints would otherwise hide by
// removing the repeats.
func frozen(pts []marketdata.Point) error {
	run, end := refgen.FlatRun(pts, time.Time{})
	log.Printf("check flat runs: longest %d sessions (ending %s)", run, end.Format(time.DateOnly))
	if run > maxRun {
		return fmt.Errorf("one level repeats for %d sessions, a frozen feed rather than an index", run)
	}
	return nil
}

// check grades the index on its own: its base, the round trips dropped, the
// band and the largest session move, and freshness.
func check(pts []marketdata.Point, dropped int, now time.Time) error {
	if first := pts[0].Date.Format(time.DateOnly); first != firstDay || pts[0].Close != baseLevel {
		return fmt.Errorf("starts %s at %g, want %s at %g", first, pts[0].Close, firstDay, baseLevel)
	}
	if dropped > maxSpikes {
		return fmt.Errorf("%d one-session round trips dropped, more than %d: a degraded record", dropped, maxSpikes)
	}
	for i, p := range pts {
		if !(p.Close > lowest && p.Close < highest) {
			return fmt.Errorf("%.2f on %s, outside (%g, %g)", p.Close, p.Date.Format(time.DateOnly), lowest, highest)
		}
		if i > 0 && math.Abs(p.Close/pts[i-1].Close-1) > maxMove {
			return fmt.Errorf("%+.1f %% on %s, beyond any session the index has made", (p.Close/pts[i-1].Close-1)*100, p.Date.Format(time.DateOnly))
		}
	}
	last := pts[len(pts)-1].Date
	if age := now.Sub(last).Hours() / 24; age > maxAge {
		return fmt.Errorf("the last level %s is %.0f days old, the source has stopped", last.Format(time.DateOnly), age)
	}
	return nil
}

// carry checks the excess return against its total return sibling: over every
// whole calendar year both cover, the total return must beat the excess
// return by that year's mean 3-month T-bill yield (bill, in percent) within
// maxCarryGap points, and the total return's 2025 must be the published one.
func carry(er, tr, bill []marketdata.Point) error {
	ye, yt := yearEnds(er), yearEnds(tr)
	rates := map[int][]float64{}
	for _, p := range bill {
		rates[p.Date.Year()] = append(rates[p.Date.Year()], p.Close)
	}
	first, last := tr[0].Date.Year()+1, tr[len(tr)-1].Date.Year()-1
	if last-first < 10 {
		return fmt.Errorf("%s covers %d..%d, too short to check the carry", trSymbol, first, last)
	}
	var worst float64
	for y := first; y <= last; y++ {
		if ye[y-1] == 0 || ye[y] == 0 || len(rates[y]) == 0 {
			return fmt.Errorf("%d: no year-end level or no T-bill yield to compare", y)
		}
		var mean float64
		for _, r := range rates[y] {
			mean += r
		}
		mean /= float64(len(rates[y]))
		gap := 100*((yt[y]/yt[y-1])/(ye[y]/ye[y-1])-1) - mean
		if math.Abs(gap) > math.Abs(worst) {
			worst = gap
		}
		if math.Abs(gap) > maxCarryGap {
			return fmt.Errorf("%d: the total return beats the excess return by %.2f points more than the %.2f %% T-bill", y, gap, mean)
		}
	}
	log.Printf("check carry %d..%d: total less excess return within %+.2f point of the T-bill every year", first, last, worst)
	if yt[2024] == 0 || yt[2025] == 0 {
		return fmt.Errorf("%s has no 2025 to check against the published return", trSymbol)
	}
	got := 100 * (yt[2025]/yt[2024] - 1)
	log.Printf("check %s 2025: %+.2f %%, published %+.2f %%", trSymbol, got, published2025)
	if math.Abs(got-published2025) > published2025Tol {
		return fmt.Errorf("%s returned %+.2f %% in 2025, the publisher says %+.2f %%", trSymbol, got, published2025)
	}
	return nil
}

// yearEnds maps each calendar year to the last level quoted in it.
func yearEnds(pts []marketdata.Point) map[int]float64 {
	out := map[int]float64{}
	for _, p := range pts {
		out[p.Date.Year()] = p.Close
	}
	return out
}
