package metrics

import (
	"math"
	"testing"
	"time"
)

// calendar returns the dates of [from, to) that keep says to keep.
func calendar(from, to time.Time, keep func(time.Time) bool) []time.Time {
	var out []time.Time
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		if keep(d) {
			out = append(out, d)
		}
	}
	return out
}

func weekday(d time.Time) bool { return d.Weekday() != time.Saturday && d.Weekday() != time.Sunday }

// holiday is a rough exchange calendar: New Year, Good Friday-ish, May Day,
// Christmas and Boxing Day off, about 250 sessions a year.
func holiday(d time.Time) bool {
	switch {
	case d.YearDay() == 1, d.Month() == time.May && d.Day() == 1,
		d.Month() == time.April && d.Day() == 10,
		d.Month() == time.December && (d.Day() == 25 || d.Day() == 26):
		return true
	}
	return false
}

func TestPeriodsPerYear(t *testing.T) {
	from := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	monthEnds := func(step int) []time.Time {
		var out []time.Time
		for m := 0; m < 120; m += step {
			out = append(out, time.Date(2000, time.Month(m+2), 0, 0, 0, 0, 0, time.UTC))
		}
		return out
	}
	var semiMonthly []time.Time
	for m := range 120 {
		semiMonthly = append(semiMonthly,
			time.Date(2000, time.Month(m+1), 15, 0, 0, 0, 0, time.UTC),
			time.Date(2000, time.Month(m+2), 0, 0, 0, 0, 0, time.UTC))
	}
	// A daily series with a two-year hole (a suspended fund) is still daily.
	holed := calendar(from, to, func(d time.Time) bool {
		return weekday(d) && (d.Year() < 2003 || d.Year() > 2004)
	})
	// A sparse monthly deep history in front of a dense daily one reads as
	// the regime most of its returns belong to.
	deep := append(monthEnds(1)[:60], calendar(time.Date(2005, 1, 3, 0, 0, 0, 0, time.UTC), to, weekday)...)
	// A fund that dealt weekly, then daily: whichever regime holds most
	// returns wins, the other part being misread either way.
	switchedAt := func(year int) []time.Time {
		return calendar(from, to, func(d time.Time) bool {
			return d.Year() >= year && weekday(d) || d.Weekday() == time.Friday
		})
	}
	for _, tc := range []struct {
		name  string
		dates []time.Time
		want  float64
	}{
		{"trading days", calendar(from, to, weekday), TradingDaysPerYear},
		{"trading days with holidays", calendar(from, to, func(d time.Time) bool { return weekday(d) && !holiday(d) }), TradingDaysPerYear},
		{"every calendar day", calendar(from, to, func(time.Time) bool { return true }), 365},
		{"weekly", calendar(from, to, func(d time.Time) bool { return d.Weekday() == time.Friday }), 52},
		{"fortnightly", calendar(from, to, func(d time.Time) bool {
			return d.Weekday() == time.Friday && (d.YearDay()/7)%2 == 0
		}), 26},
		{"semi-monthly", semiMonthly, 24},
		{"monthly", monthEnds(1), 12},
		{"quarterly", monthEnds(3), 4},
		{"yearly", monthEnds(12), 1},
		{"daily with a hole", holed, TradingDaysPerYear},
		{"monthly then daily", deep, TradingDaysPerYear},
		{"weekly, daily for the last year", switchedAt(2009), 52},
		{"weekly, daily for the last six years", switchedAt(2004), TradingDaysPerYear},
		// A six-day trading week is no canonical cadence: it keeps its own.
		{"six-day week", calendar(from, to, func(d time.Time) bool { return d.Weekday() != time.Sunday }), 6 * 365.25 / 7},
	} {
		got := PeriodsPerYear(tc.dates)
		if math.Abs(got-tc.want) > 0.5 {
			t.Errorf("%s: PeriodsPerYear = %v, want %v", tc.name, got, tc.want)
		}
		if tc.want == TradingDaysPerYear && got != TradingDaysPerYear {
			t.Errorf("%s: PeriodsPerYear = %v, want exactly %v", tc.name, got, TradingDaysPerYear)
		}
	}
	for _, dates := range [][]time.Time{nil, {from}, {from, from}} {
		if got := PeriodsPerYear(dates); !math.IsNaN(got) {
			t.Errorf("PeriodsPerYear(%v) = %v, want NaN", dates, got)
		}
	}
}

// TestComputeMonthlyAnnualizesAtTwelve pins the bug this cadence fixes: a
// monthly series annualized at 252 reported its volatility sqrt(21) too high.
func TestComputeMonthlyAnnualizesAtTwelve(t *testing.T) {
	var dates []time.Time
	var values []float64
	v := 100.0
	for m := range 240 {
		dates = append(dates, time.Date(2000, time.Month(m+2), 0, 0, 0, 0, 0, time.UTC))
		values = append(values, v)
		v *= 1 + 0.01 + 0.04*math.Sin(float64(m)*1.7)
	}
	s, err := Compute(dates, values)
	if err != nil {
		t.Fatal(err)
	}
	r := Returns(values)
	monthlyStd := sampleStdev(r)
	near(t, "PeriodsPerYear", s.PeriodsPerYear, 12, 0)
	near(t, "Volatility", s.Volatility, monthlyStd*math.Sqrt(12), 1e-12)
	near(t, "Sharpe", s.Sharpe, Mean(r)*12/(monthlyStd*math.Sqrt(12)), 1e-12)
}

// TestComputeDailyUnchanged holds a trading-day series to the fixed
// 252-day formulas bit for bit: annualizing at the measured cadence must not
// move a single daily number.
func TestComputeDailyUnchanged(t *testing.T) {
	dates := calendar(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		func(d time.Time) bool { return weekday(d) && !holiday(d) })
	values := make([]float64, len(dates))
	v := 100.0
	for i := range values {
		v *= 1 + 0.0003 + 0.011*math.Sin(float64(i)*0.9)
		values[i] = v
	}
	s, err := Compute(dates, values)
	if err != nil {
		t.Fatal(err)
	}
	r := Returns(values)
	mean := Mean(r)
	variance, downSq := 0.0, 0.0
	for _, x := range r {
		variance += (x - mean) * (x - mean)
		if x < 0 {
			downSq += x * x
		}
	}
	vol := math.Sqrt(variance/float64(len(r)-1)) * math.Sqrt(252)
	sortino := mean * 252 / (math.Sqrt(downSq/float64(len(r))) * math.Sqrt(252))
	if s.PeriodsPerYear != 252 || s.Volatility != vol || s.Sharpe != mean*252/vol || s.Sortino != sortino {
		t.Errorf("daily stats moved: ppy %v vol %v sharpe %v sortino %v, want 252, %v, %v, %v",
			s.PeriodsPerYear, s.Volatility, s.Sharpe, s.Sortino, vol, mean*252/vol, sortino)
	}
}

func TestPairedReturnsCadence(t *testing.T) {
	// Two monthly series: the relative statistics annualize at twelve.
	var dates []time.Time
	var own, bench []float64
	o, b := 100.0, 100.0
	for m := range 60 {
		dates = append(dates, time.Date(2000, time.Month(m+2), 0, 0, 0, 0, 0, time.UTC))
		own, bench = append(own, o), append(bench, b)
		x := 0.03 * math.Sin(float64(m)*2.1)
		o *= 1 + 0.002 + 1.5*x + 0.004*math.Cos(float64(m))
		b *= 1 + x
	}
	p := pairReturns(dates, own, dates, bench)
	if got := p.cadence(); got != 12 {
		t.Fatalf("paired cadence = %v, want 12", got)
	}
	rel, ok := VsBenchmark(dates, own, dates, bench)
	if !ok {
		t.Fatal("VsBenchmark not ok")
	}
	near(t, "Alpha", rel.Alpha, (Mean(p.own)-rel.Beta*Mean(p.bench))*12, 1e-12)
}
