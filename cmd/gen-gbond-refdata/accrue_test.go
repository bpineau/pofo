package main

import (
	"math"
	"testing"
	"time"
)

func TestAccruePaysTheLastMonthOnceClosed(t *testing.T) {
	jan := time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)
	feb, mar := jan.AddDate(0, 1, 0), jan.AddDate(0, 2, 0)
	rates := []obs{{date: jan, val: 1}, {date: feb, val: 3}}
	years := func(from, to time.Time) float64 { return to.Sub(from).Hours() / 24 / 365.25 }
	janLevel := 100 * math.Pow(1.01, years(jan, feb))
	for _, tc := range []struct {
		end  time.Time
		want []float64
	}{
		{mar.AddDate(0, 0, -1), []float64{100, janLevel}},
		{mar, []float64{100, janLevel, janLevel * math.Pow(1.03, years(feb, mar))}},
	} {
		got := accrue(rates, tc.end)
		if len(got) != len(tc.want) {
			t.Fatalf("end %v: %d levels, want %d", tc.end, len(got), len(tc.want))
		}
		for i, p := range got {
			if math.Abs(p.Close-tc.want[i]) > 1e-12 {
				t.Errorf("end %v: level %d = %v, want %v", tc.end, i, p.Close, tc.want[i])
			}
		}
		if last := got[len(got)-1].Date; len(got) == 3 && !last.Equal(mar) {
			t.Errorf("last level dated %v, want %v", last, mar)
		}
	}
}
