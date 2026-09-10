package marketdata

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	day := func(i int) time.Time {
		return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
	}
	s := &Series{Symbol: "X"}
	for i := 0; i < 30; i++ {
		s.Points = append(s.Points, Point{Date: day(i), Close: 100 + float64(i)})
	}
	now := day(31)
	if issues := Verify(s, now); len(issues) != 0 {
		t.Fatalf("clean series: unexpected issues %v", issues)
	}

	// One huge move, one gap, one zero price, one stale ending.
	s.Points[10].Close = 200                                      // +90 % vs day 9, then -45 %
	s.Points[20].Close = 0                                        // error
	s.Points = append(s.Points, Point{Date: day(60), Close: 130}) // 31-day gap
	issues := Verify(s, day(80))                                  // 20 days after the last quote
	var moves, gaps, zeros, stale int
	for _, is := range issues {
		switch {
		case strings.Contains(is.Message, "in one observation"):
			moves++
		case strings.Contains(is.Message, "no quotes for"):
			gaps++
		case strings.Contains(is.Message, "non-positive"):
			zeros++
			if is.Severity != "error" {
				t.Fatalf("zero price must be an error, got %q", is.Severity)
			}
		case strings.Contains(is.Message, "days old"):
			stale++
		}
	}
	if moves < 2 || gaps != 1 || zeros != 1 || stale != 1 {
		t.Fatalf("moves=%d gaps=%d zeros=%d stale=%d in %v", moves, gaps, zeros, stale, issues)
	}

	if issues := Verify(&Series{Symbol: "E"}, now); len(issues) != 1 || issues[0].Severity != "error" {
		t.Fatalf("empty series: %v", issues)
	}
}

func TestVerifyFlatRun(t *testing.T) {
	s := &Series{Symbol: "F"}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 25; i++ {
		s.Points = append(s.Points, Point{Date: start.AddDate(0, 0, i), Close: 50})
	}
	issues := Verify(s, start.AddDate(0, 0, 26))
	found := false
	for _, is := range issues {
		if strings.Contains(is.Message, "unchanged") {
			found = true
		}
	}
	if !found {
		t.Fatalf("flat run not reported: %v", issues)
	}
}

// TestVerifyMoveLimitFollowsCadence pins the two ways the single-observation
// check adapts: to the series' own pace, and to a caller-supplied class limit.
func TestVerifyMoveLimitFollowsCadence(t *testing.T) {
	monthly := &Series{Symbol: "SP500"}
	daily := &Series{Symbol: "SPY"}
	start := time.Date(1932, 1, 31, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 40; i++ {
		v := 100.0
		if i == 20 {
			v = 140 // the S&P 500's real +40 % month, August 1932
		}
		monthly.Points = append(monthly.Points, Point{Date: start.AddDate(0, i, 0), Close: v})
		daily.Points = append(daily.Points, Point{Date: start.AddDate(0, 0, i), Close: v})
	}
	now := start.AddDate(4, 0, 0)
	if got := countMoves(verify(monthly, now, 0)); got != 0 {
		t.Errorf("a +40 %% month-end observation must pass the cadence-scaled limit, got %d finding(s)", got)
	}
	if got := countMoves(verify(daily, start.AddDate(0, 0, 45), 0)); got != 2 {
		t.Errorf("the same jump between two SESSIONS must be flagged twice (up and back), got %d", got)
	}
	// A class limit replaces the blanket quarter, in both directions.
	if got := countMoves(verify(daily, start.AddDate(0, 0, 45), 0.60)); got != 0 {
		t.Errorf("a 60 %% class limit must absorb a +40 %% session, got %d finding(s)", got)
	}
	if got := countMoves(verify(daily, start.AddDate(0, 0, 45), 0.035)); got != 2 {
		t.Errorf("a money-market limit must catch it, got %d finding(s)", got)
	}
}

func countMoves(issues []Issue) int {
	n := 0
	for _, is := range issues {
		if strings.Contains(is.Message, "in one observation") {
			n++
		}
	}
	return n
}

// TestVerifyRejectsNonMonotonicDates: metrics match series by exact date
// equality, so a repeated or backwards date is an ERROR, not a warning.
func TestVerifyRejectsNonMonotonicDates(t *testing.T) {
	day := func(i int) time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &Series{Symbol: "X", Points: []Point{
		{Date: day(0), Close: 100},
		{Date: day(1), Close: 101},
		{Date: day(1), Close: 101}, // repeated
		{Date: day(0), Close: 100}, // backwards
	}}
	var errs int
	for _, is := range Verify(s, day(2)) {
		if is.Severity == "error" && strings.Contains(is.Message, "strictly increasing") {
			errs++
		}
	}
	if errs != 2 {
		t.Fatalf("got %d date errors, want 2 in %v", errs, Verify(s, day(2)))
	}
}

// TestVerifyNaNAndInfArePriceErrors: a NaN close breaks every computation
// downstream and must never pass as a mere warning.
func TestVerifyNaNAndInfArePriceErrors(t *testing.T) {
	day := func(i int) time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &Series{Symbol: "X", Points: []Point{
		{Date: day(0), Close: 100},
		{Date: day(1), Close: math.NaN()},
		{Date: day(2), Close: math.Inf(1)},
		{Date: day(3), Close: 101},
	}}
	var errs int
	for _, is := range Verify(s, day(4)) {
		if is.Severity == "error" && strings.Contains(is.Message, "non-positive price") {
			errs++
		}
	}
	if errs != 2 {
		t.Fatalf("got %d price errors, want 2", errs)
	}
}

// TestLocalSpacingDays: the cadence a series is judged against. Too short to
// measure reads as daily; a weekly NAV and a monthly index each report their
// own pace, and a series that changes pace is judged by the one in force.
func TestLocalSpacingDays(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	pts := func(n, step int) []Point {
		out := make([]Point, n)
		for i := range out {
			out[i] = Point{Date: base.AddDate(0, 0, i*step), Close: 100}
		}
		return out
	}
	for _, n := range []int{0, 1, 2} {
		for i, got := range localSpacingDays(pts(n, 30)) {
			if got != 1 {
				t.Errorf("a %d-point series reads %v days at index %d, want the daily default", n, got, i)
			}
		}
	}
	if got := localSpacingDays(pts(40, 7)); got[0] != 7 || got[20] != 7 {
		t.Errorf("a weekly series reads %v days, want 7", got[20])
	}
	daily := localSpacingDays(pts(40, 1))
	if daily[0] != 1 || daily[39] != 1 {
		t.Errorf("a daily series reads %v days, want 1", daily[39])
	}
	// Monthly for eighty points, then daily: each half keeps its own pace.
	mixed := append(pts(80, 30), pts(40, 1)...)
	for i := 80; i < len(mixed); i++ {
		mixed[i].Date = mixed[79].Date.AddDate(0, 0, i-79)
	}
	spacing := localSpacingDays(mixed)
	if spacing[40] != 30 {
		t.Errorf("the monthly stretch reads %v days, want 30", spacing[40])
	}
	if spacing[len(spacing)-1] != 1 {
		t.Errorf("the daily stretch reads %v days, want 1", spacing[len(spacing)-1])
	}
}
