package main

import (
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// series builds a daily series from consecutive closes starting at day.
func series(day time.Time, closes ...float64) *marketdata.Series {
	s := &marketdata.Series{}
	for i, c := range closes {
		s.Points = append(s.Points, marketdata.Point{Date: day.AddDate(0, 0, i), Close: c})
	}
	return s
}

// day is a UTC midnight date, the only kind the library handles.
func day(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) }

// The common window is the overlap: the latest inception and the earliest
// last quote, whichever series they come from.
func TestCommonWindow(t *testing.T) {
	list := []*marketdata.Series{
		series(day(2000, 1, 1), 1, 2, 3, 4, 5), // 2000-01-01 .. 2000-01-05
		series(day(2000, 1, 3), 1, 2, 3),       // 2000-01-03 .. 2000-01-05
		series(day(2000, 1, 2), 1, 2, 3),       // 2000-01-02 .. 2000-01-04
	}
	start, end := commonWindow(list)
	if want := day(2000, 1, 3); !start.Equal(want) {
		t.Errorf("start = %s, want %s (the youngest inception)", start.Format("2006-01-02"), want.Format("2006-01-02"))
	}
	if want := day(2000, 1, 4); !end.Equal(want) {
		t.Errorf("end = %s, want %s (the earliest last quote)", end.Format("2006-01-02"), want.Format("2006-01-02"))
	}
	// One series is its own window.
	start, end = commonWindow(list[:1])
	if !start.Equal(day(2000, 1, 1)) || !end.Equal(day(2000, 1, 5)) {
		t.Errorf("single series window = %s..%s", start, end)
	}
}
