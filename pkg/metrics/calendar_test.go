package metrics

import (
	"math"
	"testing"
	"time"
)

// weekdays is a weekday calendar from..to inclusive, with a deterministic
// wobbling level path starting at 100.
func weekdays(from, to time.Time) ([]time.Time, []float64) {
	var dates []time.Time
	var values []float64
	v := 100.0
	for d, i := from, 0; !d.After(to); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		if i > 0 {
			v *= 1 + 0.0004 + 0.01*math.Sin(float64(i)*0.7)
		}
		dates, values = append(dates, d), append(values, v)
		i++
	}
	return dates, values
}

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func TestCalendarReturns(t *testing.T) {
	// Starts mid-March 2020 (a partial first month), quotes 31 December
	// 2020 and 2021 (a Thursday and a Friday), ends mid-June 2022.
	dates, values := weekdays(date(2020, 3, 11), date(2022, 6, 15))
	level := make(map[time.Time]float64, len(dates))
	for i, d := range dates {
		level[d] = values[i]
	}
	total := values[len(values)-1]/values[0] - 1

	for _, months := range []int{1, 3, 12} {
		table := CalendarReturns(dates, values, months)
		chained := 1.0
		for k, r := range table {
			chained *= 1 + r.Return
			if r.Partial != (k == 0) {
				t.Errorf("months=%d period %d: Partial=%v", months, k, r.Partial)
			}
			base := values[0]
			if k > 0 {
				base = level[table[k-1].End]
				if !r.Start.After(table[k-1].End) {
					t.Errorf("months=%d period %d starts %s, not after the previous end", months, k, r.Start)
				}
			}
			near(t, "period return", r.Return, level[r.End]/base-1, 1e-15)
		}
		near(t, "chained total", chained-1, total, 1e-12)
		first, last := table[0], table[len(table)-1]
		if !first.Start.Equal(date(2020, 3, 11)) || !last.End.Equal(date(2022, 6, 15)) {
			t.Errorf("months=%d: first starts %s, last ends %s", months, first.Start, last.End)
		}
		if last.Partial {
			t.Errorf("months=%d: the last period is flagged Partial", months)
		}
	}

	monthly := CalendarReturns(dates, values, 1)
	if len(monthly) != 28 { // Mar 2020 .. Jun 2022
		t.Errorf("%d monthly periods, want 28", len(monthly))
	}
	if !monthly[0].End.Equal(date(2020, 3, 31)) || !monthly[1].Start.Equal(date(2020, 4, 1)) {
		t.Errorf("first month %s..%s, second starts %s", monthly[0].Start, monthly[0].End, monthly[1].Start)
	}
	if n := len(CalendarReturns(dates, values, 3)); n != 10 { // 2020Q1 .. 2022Q2
		t.Errorf("%d quarterly periods, want 10", n)
	}

	yearly := CalendarReturns(dates, values, 12)
	if len(yearly) != 3 {
		t.Fatalf("%d yearly periods, want 3", len(yearly))
	}
	if !yearly[0].End.Equal(date(2020, 12, 31)) || !yearly[1].Start.Equal(date(2021, 1, 1)) || !yearly[1].End.Equal(date(2021, 12, 31)) {
		t.Errorf("2020 ends %s, 2021 runs %s..%s", yearly[0].End, yearly[1].Start, yearly[1].End)
	}
	// The yearly row is the monthly rows of that year chained.
	months2021 := 1.0
	for _, r := range monthly {
		if r.End.Year() == 2021 {
			months2021 *= 1 + r.Return
		}
	}
	near(t, "2021 from its months", yearly[1].Return, months2021-1, 1e-12)
}

func TestCalendarReturnsGap(t *testing.T) {
	// No quote in February: March's return spans it.
	dates := []time.Time{date(2021, 1, 15), date(2021, 1, 29), date(2021, 3, 5), date(2021, 3, 31)}
	values := []float64{100, 110, 99, 121}
	got := CalendarReturns(dates, values, 1)
	if len(got) != 2 {
		t.Fatalf("%d periods, want 2", len(got))
	}
	near(t, "January", got[0].Return, 0.10, 1e-12)
	near(t, "March over January's end", got[1].Return, 0.10, 1e-12)
	if !got[1].Start.Equal(date(2021, 3, 5)) {
		t.Errorf("March starts %s", got[1].Start)
	}
}

func TestCalendarReturnsEdgeCases(t *testing.T) {
	if got := CalendarReturns(nil, nil, 12); got != nil {
		t.Errorf("empty: %v", got)
	}
	if got := CalendarReturns([]time.Time{date(2021, 1, 4)}, []float64{1, 2}, 12); got != nil {
		t.Errorf("mismatched: %v", got)
	}
	got := CalendarReturns([]time.Time{date(2021, 1, 4)}, []float64{100}, 12)
	if len(got) != 1 || got[0].Return != 0 || !got[0].Partial || !got[0].Start.Equal(got[0].End) {
		t.Errorf("one point: %+v", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("months=0 did not panic")
		}
	}()
	CalendarReturns([]time.Time{date(2021, 1, 4)}, []float64{100}, 0)
}
