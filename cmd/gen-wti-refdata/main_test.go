package main

import (
	"archive/zip"
	"bytes"
	"math"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// tradingDays lists the weekdays of a span, which is the exchange calendar the
// tests need: the real holidays only shift a roll by a day and are exercised by
// the recorded expiry dates in TestRenumberDaysMatchesCMEExpiries.
func tradingDays(from, to time.Time) []time.Time {
	var out []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if wd := d.Weekday(); wd != time.Saturday && wd != time.Sunday {
			out = append(out, d)
		}
	}
	return out
}

func TestRollWeight(t *testing.T) {
	// Nothing moves before the fifth business day, a fifth moves on each of
	// the five roll days, and the position is wholly in the next contract
	// from the ninth on.
	want := map[int]float64{1: 0, 4: 0, 5: 0.2, 6: 0.4, 7: 0.6, 8: 0.8, 9: 1, 15: 1, 23: 1}
	for n, w := range want {
		if got := rollWeight(n); math.Abs(got-w) > 1e-12 {
			t.Errorf("rollWeight(%d) = %.2f, want %.2f", n, got, w)
		}
	}
}

func TestBusinessDayOfMonth(t *testing.T) {
	dates := tradingDays(day(2015, time.January, 28), day(2015, time.February, 6))
	bd := businessDayOfMonth(dates)
	if got := bd[day(2015, time.January, 30)]; got != 3 {
		t.Errorf("2015-01-30 is business day %d, want 3", got)
	}
	// The count restarts with the month.
	if got := bd[day(2015, time.February, 2)]; got != 1 {
		t.Errorf("2015-02-02 is business day %d, want 1", got)
	}
	if got := bd[day(2015, time.February, 6)]; got != 5 {
		t.Errorf("2015-02-06 is business day %d, want 5", got)
	}
}

// TestRenumberDaysMatchesCMEExpiries pins the slot calendar to dates that are
// matters of record. CME terminates WTI trading three business days before the
// 25th (or before the last business day preceding it), and EIA's contract 1
// becomes the next delivery month on the following trading day.
//
// The April 2020 case is the one that admits no doubt: the May contract settled
// at -37.63 on 2020-04-20 and at 10.01 on the 21st, its last day, and EIA
// carried both under contract 1, so the slot can only have moved on the 22nd.
func TestRenumberDaysMatchesCMEExpiries(t *testing.T) {
	// A calendar with the real NYMEX holidays of the tested months.
	holidays := map[string]bool{
		"2015-01-01": true, "2015-01-19": true, "2015-02-16": true,
		"2020-01-01": true, "2020-01-20": true, "2020-02-17": true,
		"2020-04-10": true, "2020-05-25": true,
	}
	var dates []time.Time
	for _, d := range tradingDays(day(2014, time.December, 1), day(2020, time.July, 31)) {
		if !holidays[d.Format("2006-01-02")] {
			dates = append(dates, d)
		}
	}
	renum := renumberDays(dates)

	want := []time.Time{
		day(2015, time.February, 23), // March 2015 expired Friday 2015-02-20
		day(2020, time.April, 22),    // May 2020 expired Tuesday 2020-04-21
		day(2020, time.May, 20),      // June 2020 expired Tuesday 2020-05-19
	}
	for _, d := range want {
		if !renum[d] {
			t.Errorf("%s should renumber the contract slots", d.Format("2006-01-02"))
		}
	}
	notWant := []time.Time{
		day(2015, time.February, 20), // the expiry itself, still the old slot
		day(2020, time.April, 21),
		day(2020, time.April, 23),
	}
	for _, d := range notWant {
		if renum[d] {
			t.Errorf("%s should not renumber the contract slots", d.Format("2006-01-02"))
		}
	}
	// Exactly one renumbering per month, no more.
	perMonth := map[string]int{}
	for d := range renum {
		perMonth[d.Format("2006-01")]++
	}
	for m, n := range perMonth {
		if n != 1 {
			t.Errorf("%s renumbers %d times, want 1", m, n)
		}
	}
	if len(perMonth) < 60 {
		t.Errorf("only %d months renumber over a 5-year span", len(perMonth))
	}
}

// TestRenumberDaysStopsAtTruncation guards the trap that a month whose data
// simply ends before the 25th is not an expiry: the series was cut there.
func TestRenumberDaysStopsAtTruncation(t *testing.T) {
	dates := tradingDays(day(2024, time.March, 1), day(2024, time.April, 5))
	renum := renumberDays(dates)
	for d := range renum {
		if d.Month() == time.April {
			t.Errorf("April 2024 renumbers on %s although the data stops on the 5th", d.Format("2006-01-02"))
		}
	}
	if !renum[day(2024, time.March, 21)] { // April 2024 expired 2024-03-20
		t.Error("2024-03-21 should renumber the contract slots")
	}
}

// TestBuildFlatCurvePaysNothing is the heart of the method. Both contracts hold
// the same motionless price, so there is no spread to earn or to pay and every
// honest return is zero. A build that mishandled either the roll days or the
// slot renumbering would book the day's price ratio as a move.
func TestBuildFlatCurvePaysNothing(t *testing.T) {
	dates := tradingDays(day(2015, time.January, 1), day(2015, time.June, 30))
	c1 := map[time.Time]float64{}
	c2 := map[time.Time]float64{}
	for _, d := range dates {
		c1[d] = 50
		c2[d] = 50
	}
	idx, err := build(c1, c2)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(idx) != len(dates) {
		t.Fatalf("built %d points from %d days", len(idx), len(dates))
	}
	for _, p := range idx {
		if math.Abs(p.level-100) > 1e-9 {
			t.Fatalf("%s: level %.6f, want 100 (a flat curve cannot pay a roll)", p.date.Format("2006-01-02"), p.level)
		}
	}
}

// TestBuildPricesTheRollYield checks the sign and the size of the carry, which
// is the whole reason this series exists.
//
// The roll days themselves are value-neutral: the position exchanges one
// contract for an equal VALUE of the next, which is why the daily return prices
// both legs at the weights carried into the day. The carry is realized instead
// as the held contract converges, which on a curve pinned in place happens all
// at once when the slots renumber. So a static curve pays exactly its spread,
// once a month, with the sign of the backwardation.
func TestBuildPricesTheRollYield(t *testing.T) {
	dates := tradingDays(day(2015, time.January, 1), day(2015, time.December, 31))
	for _, tc := range []struct {
		name       string
		front, nxt float64
	}{
		{"backwardation", 50, 49},
		{"contango", 50, 52},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c1 := map[time.Time]float64{}
			c2 := map[time.Time]float64{}
			for _, d := range dates {
				c1[d] = tc.front
				c2[d] = tc.nxt
			}
			idx, err := build(c1, c2)
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			renum := renumberDays(dates)
			want := math.Pow(tc.front/tc.nxt, float64(len(renum)))
			got := idx[len(idx)-1].level / idx[0].level
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("%d rolls at %.0f/%.0f returned %.6fx, want %.6fx",
					len(renum), tc.front, tc.nxt, got, want)
			}
		})
	}
}

// TestBuildSkipsUnpairedDays: a day only one leg quotes cannot price a roll and
// must be dropped, not forward-filled.
func TestBuildSkipsUnpairedDays(t *testing.T) {
	dates := tradingDays(day(2015, time.January, 1), day(2015, time.March, 31))
	c1 := map[time.Time]float64{}
	c2 := map[time.Time]float64{}
	for _, d := range dates {
		c1[d] = 50
		if !d.Equal(day(2015, time.February, 10)) {
			c2[d] = 50
		}
	}
	idx, err := build(c1, c2)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, p := range idx {
		if p.date.Equal(day(2015, time.February, 10)) {
			t.Error("2015-02-10 quotes only one leg and should have been skipped")
		}
	}
	if len(idx) != len(dates)-1 {
		t.Errorf("built %d points from %d pairable days", len(idx), len(dates)-1)
	}
}

// TestExtract reads the wanted series out of an archive shaped like EIA's:
// one zip member of JSON lines, one line per series, newest observation first,
// nulls for days the series does not quote.
func TestExtract(t *testing.T) {
	lines := `{"series_id":"PET.RWTC.D","name":"spot","units":"Dollars per Barrel","data":[["20240102",70.1]]}
{"series_id":"PET.RCLC1.D","name":"Contract 1","units":"Dollars per Barrel","data":[["20240103",72.5],["20240102",null],["20240101",71.0]]}
{"series_id":"PET.RCLC2.D","name":"Contract 2","units":"Dollars per Barrel","data":[["20240103",73.5],["20240101",72.0]]}
`
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("PET.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(lines)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := extract(buf.Bytes(), front, next)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("extracted %d series, want 2", len(got))
	}
	if n := len(got[front]); n != 2 {
		t.Errorf("%s has %d observations, want 2 (the null day is dropped)", front, n)
	}
	if v := got[front][day(2024, time.January, 3)]; v != 72.5 {
		t.Errorf("2024-01-03 contract 1 = %v, want 72.5", v)
	}
	if _, ok := got[front][day(2024, time.January, 2)]; ok {
		t.Error("the null observation was kept")
	}
	if _, err := extract(buf.Bytes(), front, "PET.NOPE.D"); err == nil {
		t.Error("a missing series should be an error, not a silent gap")
	}
}

func TestCheckRejectsAShortSeries(t *testing.T) {
	idx := []point{{day(2020, time.January, 2), 100}, {day(2020, time.January, 3), 101}}
	if err := check(idx); err == nil {
		t.Error("a two-day series should not pass the sanity gate")
	}
}

// TestTbillDaily reads the bundled rate and checks the convention: a discount
// rate in percent on a 91-day bill, compounded per calendar day.
func TestTbillDaily(t *testing.T) {
	f, err := tbillDaily()
	if err != nil {
		t.Fatalf("tbillDaily: %v", err)
	}
	r := f(day(2019, time.June, 15))
	annual := math.Pow(1+r, 365) - 1
	if annual < 0.015 || annual > 0.035 {
		t.Errorf("mid-2019 T-bill compounds to %.2f%%/yr, want roughly 2 to 3", annual*100)
	}
	if got := f(day(2015, time.June, 15)); got >= r {
		t.Errorf("2015 accrual %.8f is not below 2019's %.8f", got, r)
	}
}
