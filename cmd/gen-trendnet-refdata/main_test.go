package main

import (
	"math"
	"strings"
	"testing"
	"time"
)

// series builds n consecutive months from 1987-01, each returning ret with an
// alternating swing of amp around it, which is enough to give the sanity gates
// something to measure.
func series(n int, ret, amp float64) []point {
	out := make([]point, n)
	for i := range out {
		r := ret + amp
		if i%2 == 1 {
			r = ret - amp
		}
		out[i] = point{monthEndOf(1987+i/12, time.Month(i%12+1)), r}
	}
	return out
}

func TestCheckAcceptsAPlausibleIndex(t *testing.T) {
	// ~5.5 %/yr at ~9 %/yr volatility: the shape the real index has.
	if err := check(series(474, 0.0045, 0.026)); err != nil {
		t.Fatalf("rejected a plausible index: %v", err)
	}
}

func TestCheckRejects(t *testing.T) {
	gap := series(474, 0.0045, 0.026)
	gap = append(gap[:100], gap[101:]...)
	duplicate := series(474, 0.0045, 0.026)
	duplicate[100] = duplicate[99]

	for _, c := range []struct {
		name   string
		months []point
		want   string
	}{
		{"too short", series(399, 0.0045, 0.026), "only 399 months"},
		{"a missing month", gap, "2 months apart"},
		{"a duplicated month", duplicate, "0 months apart"},
		{"a cash-like series", series(474, 0.0025, 0.001), "volatility"},
		{"an equity-like series", series(474, 0.006, 0.06), "volatility"},
		{"a series that never pays", series(474, 0.0, 0.026), "information ratio"},
		{"a backtest that is too good", series(474, 0.012, 0.026), "information ratio"},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := check(c.months)
			if err == nil {
				t.Fatalf("accepted %s", c.name)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err, c.want)
			}
		})
	}
}

func TestStats(t *testing.T) {
	// A constant 1 %/month has no volatility to speak of.
	if vol, _ := stats(series(120, 0.01, 0)); vol > 1e-12 {
		t.Errorf("constant series: vol %.4g, want none", vol)
	}
	// Alternating +3 %/-1 % around +1 %: the sample deviation of n such months
	// is 2 %*sqrt(n/(n-1)), annualized by sqrt(12).
	const n = 120
	want := 0.02 * math.Sqrt(n/(n-1.0)) * math.Sqrt(12)
	vol, ir := stats(series(n, 0.01, 0.02))
	if math.Abs(vol-want) > 1e-9 {
		t.Errorf("vol = %.6f, want %.6f", vol, want)
	}
	if ir <= 0 {
		t.Errorf("ir = %.4f, want a positive information ratio", ir)
	}
}

func TestMaxDrawdown(t *testing.T) {
	months := []point{{ret: 0.10}, {ret: -0.20}, {ret: -0.10}, {ret: 0.50}}
	want := 0.80*0.90 - 1 // the two falls compound
	if got := maxDrawdown(months); math.Abs(got-want) > 1e-12 {
		t.Errorf("maxDrawdown = %.6f, want %.6f", got, want)
	}
	if got := maxDrawdown([]point{{ret: 0.01}, {ret: 0.02}}); got != 0 {
		t.Errorf("a series that only rises has drawdown %.6f, want 0", got)
	}
}

func TestMonthEnd(t *testing.T) {
	for _, c := range []struct {
		year  int
		month time.Month
		want  string
	}{
		{1987, time.January, "1987-01-31"},
		{1988, time.February, "1988-02-29"}, // a leap year
		{2026, time.June, "2026-06-30"},
	} {
		if got := monthEndOf(c.year, c.month).Format("2006-01-02"); got != c.want {
			t.Errorf("monthEndOf(%d, %s) = %s, want %s", c.year, c.month, got, c.want)
		}
	}
	// The level series opens one month before the first return: the CSV's
	// base-100 point must land on the previous month's end, never on the
	// first return's own date, which would ship two rows for one month.
	for _, c := range []struct{ from, want string }{
		{"1987-01-31", "1986-12-31"},
		{"1987-03-31", "1987-02-28"},
		{"2026-06-30", "2026-05-31"},
	} {
		d, err := time.Parse("2006-01-02", c.from)
		if err != nil {
			t.Fatal(err)
		}
		if got := previousMonthEnd(d).Format("2006-01-02"); got != c.want {
			t.Errorf("previousMonthEnd(%s) = %s, want %s", c.from, got, c.want)
		}
	}
}
