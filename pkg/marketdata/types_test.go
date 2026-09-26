package marketdata

import (
	"reflect"
	"testing"
	"time"
)

func TestSeriesAt(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2024, 1, day, 0, 0, 0, 0, time.UTC) }
	s := &Series{Points: []Point{
		{Date: d(2), Close: 100},
		{Date: d(3), Close: 101},
		{Date: d(8), Close: 105}, // gap: 4th to 7th missing
	}}

	if v, on, ok := s.At(d(3)); !ok || v != 101 || !on.Equal(d(3)) {
		t.Errorf("exact day: %v %v %v", v, on, ok)
	}
	if v, on, ok := s.At(d(5)); !ok || v != 101 || !on.Equal(d(3)) {
		t.Errorf("forward fill across the gap: %v %v %v", v, on, ok)
	}
	if v, on, ok := s.At(d(20)); !ok || v != 105 || !on.Equal(d(8)) {
		t.Errorf("forward fill after the end: %v %v %v", v, on, ok)
	}
	if _, _, ok := s.At(d(1)); ok {
		t.Error("before the first point should not be ok")
	}
	if _, _, ok := (&Series{}).At(d(1)); ok {
		t.Error("empty series should not be ok")
	}
	var nilSeries *Series
	if _, _, ok := nilSeries.At(d(1)); ok {
		t.Error("nil series should not be ok")
	}
}

func TestMergeDividends(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2024, 1, day, 0, 0, 0, 0, time.UTC) }
	got := MergeDividends(nil, Dividend{Date: d(10), Amount: 1})
	got = MergeDividends(got,
		Dividend{Date: d(5), Amount: 2},    // insert before
		Dividend{Date: d(10), Amount: 1.5}, // upsert existing
		Dividend{Date: d(20), Amount: 3},   // append after
	)
	want := []Dividend{{Date: d(5), Amount: 2}, {Date: d(10), Amount: 1.5}, {Date: d(20), Amount: 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestSeriesFirstLast(t *testing.T) {
	if (&Series{}).First() != (Point{}) || (&Series{}).Last() != (Point{}) {
		t.Error("an empty series must yield the zero point at both ends")
	}
	s := &Series{Points: []Point{
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Close: 1},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Close: 2},
	}}
	if s.First().Close != 1 || s.Last().Close != 2 {
		t.Errorf("first/last = %v/%v", s.First(), s.Last())
	}
}

// TestAlignWindowAndForwardFill pins the two rules of the calendar merge: the
// union of dates is clamped to the window, and each series is forward-filled
// on days it did not trade. A start before a series' first quote would
// forward-fill zeros, which is why callers must not pass one.
func TestAlignWindowAndForwardFill(t *testing.T) {
	day := func(i int) time.Time { return time.Date(2024, 1, i, 0, 0, 0, 0, time.UTC) }
	a := &Series{Symbol: "A", Points: []Point{
		{Date: day(1), Close: 10}, {Date: day(2), Close: 11}, {Date: day(5), Close: 12},
	}}
	b := &Series{Symbol: "B", Points: []Point{
		{Date: day(1), Close: 100}, {Date: day(3), Close: 101}, {Date: day(9), Close: 102},
	}}
	dates, levels := Align([]*Series{a, b}, day(2), day(5))
	want := []time.Time{day(2), day(3), day(5)}
	if len(dates) != len(want) {
		t.Fatalf("dates = %v, want %v", dates, want)
	}
	for i := range want {
		if !dates[i].Equal(want[i]) {
			t.Fatalf("dates = %v, want %v", dates, want)
		}
	}
	if got := levels[0]; got[0] != 11 || got[1] != 11 || got[2] != 12 {
		t.Errorf("A forward-filled = %v, want [11 11 12]", got)
	}
	if got := levels[1]; got[0] != 100 || got[1] != 101 || got[2] != 101 {
		t.Errorf("B forward-filled = %v, want [100 101 101]", got)
	}
	// A zero end bound is open on that side.
	if dates, _ := Align([]*Series{a, b}, day(1), time.Time{}); len(dates) != 5 {
		t.Errorf("open-ended window kept %d dates, want 5", len(dates))
	}
}
