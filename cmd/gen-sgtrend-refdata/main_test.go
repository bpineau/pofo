package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/cmd/internal/xls"
)

// The sources are fetched at generation time only, so everything below runs on
// sheets and files this file builds itself, laid out the way the published
// ones are.

// dailySheet is the dump's layout: a title row, a header row naming the
// columns, then one row per index day (date as an Excel serial, then the
// period return and the level, then period-to-date columns nobody reads).
func dailySheet(rows []point) []xls.Cell {
	cells := []xls.Cell{
		{Row: 0, Col: 0, Text: "TEST Index Daily Historical Data", IsText: true},
		{Row: 1, Col: 0, Text: "Dstamp", IsText: true},
		{Row: 1, Col: 1, Text: "ROR", IsText: true},
		{Row: 1, Col: 2, Text: "VAMI", IsText: true},
		{Row: 1, Col: 3, Text: "MTD", IsText: true},
	}
	epoch := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	for i, p := range rows {
		serial := p.date.Sub(epoch).Hours() / 24
		cells = append(cells,
			xls.Cell{Row: 2 + i, Col: 0, Num: serial},
			xls.Cell{Row: 2 + i, Col: 2, Num: p.level})
	}
	return cells
}

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestParseDaily(t *testing.T) {
	want := []point{
		{day("2000-01-03"), 997.199999},
		{day("2000-01-04"), 970.674479},
		{day("2000-01-05"), 962.132544},
	}
	got, err := parseDaily(dailySheet(want))
	if err != nil {
		t.Fatalf("parseDaily: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d days, want %d", len(got), len(want))
	}
	for i, w := range want {
		if !got[i].date.Equal(w.date) || math.Abs(got[i].level-w.level) > 1e-9 {
			t.Errorf("day %d = %s %.6f, want %s %.6f", i,
				got[i].date.Format("2006-01-02"), got[i].level,
				w.date.Format("2006-01-02"), w.level)
		}
	}
}

// The columns are read off the header, so a column inserted upstream moves the
// reader with it instead of shifting every level by one.
func TestParseDailyFollowsTheHeader(t *testing.T) {
	cells := dailySheet([]point{{day("2000-01-03"), 1000}, {day("2000-01-04"), 1010}})
	for i := range cells {
		if cells[i].Col >= 2 {
			cells[i].Col++ // an extra column slipped in before VAMI
		}
	}
	got, err := parseDaily(cells)
	if err != nil {
		t.Fatalf("parseDaily: %v", err)
	}
	if len(got) != 2 || got[1].level != 1010 {
		t.Errorf("got %v, want the VAMI column found at its new place", got)
	}
}

func TestParseDailyRejects(t *testing.T) {
	good := []point{{day("2000-01-03"), 1000}, {day("2000-01-04"), 1010}}
	noHeader := dailySheet(good)
	noHeader[1].Text = "Date" // no longer the header the dump publishes
	duplicate := dailySheet([]point{{day("2000-01-03"), 1000}, {day("2000-01-03"), 1010}})
	noLevel := dailySheet(good)
	noLevel = noLevel[:len(noLevel)-1] // the last row keeps its date, loses its level

	for _, c := range []struct {
		name  string
		cells []xls.Cell
		want  string
	}{
		{"no header", noHeader, "header"},
		{"a duplicate date", duplicate, "twice"},
		{"a row without a level", noLevel, "no usable level"},
		{"an empty sheet", nil, "header"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseDaily(c.cells)
			if err == nil {
				t.Fatalf("accepted %s", c.name)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err, c.want)
			}
		})
	}
}

const dashboard = `trade_date,Other Index,SG Trend Index,Equity
2000-01-01,100.00,100.00,100.00
2000-01-03,99.10,99.72,100.40
2000-01-04,98.20,97.07,100.90
not a date,1,2,3
`

func TestParseDashboard(t *testing.T) {
	got, err := parseDashboard([]byte(dashboard), "SG Trend Index")
	if err != nil {
		t.Fatalf("parseDashboard: %v", err)
	}
	want := []point{{day("2000-01-01"), 100}, {day("2000-01-03"), 99.72}, {day("2000-01-04"), 97.07}}
	if len(got) != len(want) {
		t.Fatalf("got %d days, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if !got[i].date.Equal(w.date) || math.Abs(got[i].level-w.level) > 1e-9 {
			t.Errorf("day %d = %v, want %v", i, got[i], w)
		}
	}
	if _, err := parseDashboard([]byte(dashboard), "No Such Index"); err == nil {
		t.Fatal("accepted a column the file does not carry")
	}
}

// crossCheck compares returns, not levels: the two channels are published on
// different bases and one of them is rounded.
func TestCrossCheckAcceptsADifferentBase(t *testing.T) {
	a := walk(minCommonDays+10, 0.0002, 0.008)
	b := make([]point, len(a))
	for i, p := range a {
		b[i] = point{p.date, p.level / 10} // the same index, published on a tenth of the base
	}
	msg, err := crossCheck(pureTrend, a, b)
	if err != nil {
		t.Fatalf("crossCheck: %v", err)
	}
	if !strings.Contains(msg, "agree") {
		t.Errorf("report %q does not report the agreement", msg)
	}
}

func TestCrossCheckRejects(t *testing.T) {
	a := walk(minCommonDays+10, 0.0002, 0.008)
	diverging := make([]point, len(a))
	for i, p := range a {
		diverging[i] = p
		if i >= 100 { // one 5 bp step, carried by every level after it
			diverging[i].level *= 1.0005
		}
	}
	if _, err := crossCheck(pureTrend, a, diverging); err == nil {
		t.Fatal("accepted two channels that disagree by 5 bp")
	}
	if _, err := crossCheck(pureTrend, a, a[:100]); err == nil {
		t.Fatal("accepted an overlap of 100 days")
	}
}

func TestTrimPartialMonth(t *testing.T) {
	// A month whose remaining days are all weekend is finished; one with a
	// weekday left in it is not.
	full := []point{{day("2026-01-30"), 100}, {day("2026-01-31"), 101}}
	if _, dropped := trimPartialMonth(full); len(dropped) != 0 {
		t.Errorf("dropped %d days of a finished month", len(dropped))
	}
	saturdayEnd := []point{{day("2026-02-26"), 100}, {day("2026-02-27"), 101}} // 02-28 is a Saturday
	if _, dropped := trimPartialMonth(saturdayEnd); len(dropped) != 0 {
		t.Errorf("dropped %d days of a month that ends on a weekend", len(dropped))
	}
	partial := []point{{day("2026-07-31"), 100}, {day("2026-08-03"), 101}, {day("2026-08-04"), 102}}
	kept, dropped := trimPartialMonth(partial)
	if len(kept) != 1 || !kept[0].date.Equal(day("2026-07-31")) || len(dropped) != 2 {
		t.Errorf("kept %v, dropped %v; want the two August days dropped", kept, dropped)
	}
}

func TestSanityAcceptsAPlausibleIndex(t *testing.T) {
	if err := sanity(pureTrend, index(), flatCash(2.0)); err != nil {
		t.Fatalf("rejected a plausible pure-trend index: %v", err)
	}
}

func TestSanityRejects(t *testing.T) {
	late := index()
	late = late[500:] // starts years after the index the file claims to be

	gap := index()
	var missing []point
	for _, p := range gap {
		if p.date.Year() == 2010 && p.date.Month() == time.June {
			continue
		}
		missing = append(missing, p)
	}

	for _, c := range []struct {
		name string
		days []point
		cash []point
		want string
	}{
		{"a late start", late, flatCash(2), "want 2000-01"},
		{"a missing month", missing, flatCash(2), "months apart"},
		{"a cash-like series", walkFrom(day("2000-01-03"), len(index()), 0.0001, 0.0005), flatCash(2), "volatility"},
		{"an equity-like series", walkFrom(day("2000-01-03"), len(index()), 0.0003, 0.02), flatCash(2), "volatility"},
		{"a backtest that is too good", index(), flatCash(-8), "information ratio"},
		{"a series that never pays", index(), flatCash(12), "information ratio"},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := sanity(pureTrend, c.days, c.cash)
			if err == nil {
				t.Fatalf("accepted %s", c.name)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err, c.want)
			}
		})
	}
}

// A stale download is the failure this gate is really for: the series parses,
// every month is there, and it simply stopped years ago.
func TestSanityRejectsAStaleDownload(t *testing.T) {
	days := index()
	cut := len(days)
	for cut > 0 && days[cut-1].date.After(time.Now().UTC().AddDate(-2, 0, 0)) {
		cut--
	}
	if err := sanity(pureTrend, days[:cut], flatCash(2)); err == nil || !strings.Contains(err.Error(), "behind today") {
		t.Fatalf("accepted a two-year-old download: %v", err)
	}
}

// The cash leg of the funded excess is bundled, so this gate needs no network
// of its own: check it is there and reads as an annualized percent rate.
func TestCashRates(t *testing.T) {
	cash, err := cashRates()
	if err != nil {
		t.Fatalf("cashRates: %v", err)
	}
	if len(cash) < 600 {
		t.Errorf("%d monthly rates, want the whole bundled history", len(cash))
	}
	if r := rateAt(cash, day("2007-06-15")); r < 1 || r > 10 {
		t.Errorf("the 2007 T-bill rate reads %.2f, want annualized percent", r)
	}
}

// index is a synthetic stand-in with the character the gates expect: weekdays
// from the index's own inception to the end of last month, at about 12.5 %/yr
// volatility for a funded excess information ratio near 0.3 over a 2 % cash
// rate. Deterministic: the same draw every run.
func index() []point {
	end := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 0, 0, 0, 0, 0, time.UTC)
	var days []point
	rnd := rand.New(rand.NewSource(7))
	level := 1000.0
	for d := day("2000-01-03"); !d.After(end); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		days = append(days, point{d, level})
		level *= 1 + 0.00024 + 0.0079*rnd.NormFloat64()
	}
	return days
}

// walk is index's shape with a chosen mean and deviation, over n weekdays.
func walk(n int, mean, sd float64) []point {
	return walkFrom(day("2000-01-03"), n, mean, sd)
}

func walkFrom(start time.Time, n int, mean, sd float64) []point {
	rnd := rand.New(rand.NewSource(11))
	var days []point
	level := 1000.0
	for d := start; len(days) < n; d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		days = append(days, point{d, level})
		level *= 1 + mean + sd*rnd.NormFloat64()
	}
	return days
}

// flatCash is a constant annualized percent rate, monthly from 1999.
func flatCash(pct float64) []point {
	var out []point
	for y := 1999; y <= time.Now().UTC().Year(); y++ {
		for m := time.January; m <= time.December; m++ {
			out = append(out, point{time.Date(y, m, 1, 0, 0, 0, 0, time.UTC), pct})
		}
	}
	return out
}

// A guard on the fixture itself: if the synthetic index drifts out of the
// bands, the rejection tests above would be passing for the wrong reason.
func TestIndexFixtureIsInsideTheBands(t *testing.T) {
	days := index()
	vol := annualVol(returns(days), 252)
	ir := annualGeoMean(excessReturns(monthEnds(days), flatCash(2)), 12) /
		annualVol(excessReturns(monthEnds(days), flatCash(2)), 12)
	if vol < pureTrend.minVol || vol > pureTrend.maxVol || ir < pureTrend.minIR || ir > pureTrend.maxIR {
		t.Errorf("fixture: vol %.1f%%, IR %.2f, outside the gates %v",
			vol*100, ir, fmt.Sprintf("[%.0f, %.0f]%% and [%.2f, %.2f]", pureTrend.minVol*100, pureTrend.maxVol*100, pureTrend.minIR, pureTrend.maxIR))
	}
}

// A slow restatement is invisible day by day and fatal in aggregate: every
// common return agrees well inside the per-day gate, and the two channels still
// end a third of a point apart.
func TestCrossCheckRejectsASlowDrift(t *testing.T) {
	a := walk(minCommonDays+10, 0.0002, 0.008)
	drifting := make([]point, len(a))
	k := 1.0
	for i, p := range a {
		k *= 1 + 5e-7 // 0.005 bp a day, a fortieth of the per-day gate
		drifting[i] = point{p.date, p.level * k}
	}
	if _, err := crossCheck(pureTrend, a, drifting); err == nil ||
		!strings.Contains(err.Error(), "compound") {
		t.Fatalf("accepted two channels drifting apart: %v", err)
	}
}

// The two shipped series must be two series: one wrong constant and this
// generator writes the same index twice under two names.
func TestShippedSeriesAreDistinct(t *testing.T) {
	if len(shipped) != 2 {
		t.Fatalf("%d shipped series, want 2", len(shipped))
	}
	for _, f := range []func(series) string{
		func(s series) string { return s.outID },
		func(s series) string { return s.progCode },
		func(s series) string { return s.checkColumn },
		func(s series) string { return s.name },
		func(s series) string { return s.source },
	} {
		if f(shipped[0]) == f(shipped[1]) {
			t.Errorf("both series share %q", f(shipped[0]))
		}
	}
	for _, s := range shipped {
		if s.maxReturnGap <= 0 || s.maxDrift <= 0 || s.minVol >= s.maxVol || s.minIR >= s.maxIR {
			t.Errorf("%s: unusable gates %+v", s.outID, s)
		}
	}
}
