package marketdata

import (
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
