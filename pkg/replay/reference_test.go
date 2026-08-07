package replay

import (
	"math"
	"testing"
	"time"
)

// The bundled 60/40 must land on the historical record, not merely compile.
// These bounds are wide enough to survive a data refresh and tight enough to
// catch a units, alignment or deflator mistake (the classic failures: an
// undeflated series would show a ~8.5% "real" CAGR, a misaligned bond leg a
// wrong 2022).
func TestReferenceMatchesTheRecord(t *testing.T) {
	r, err := Reference()
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Dates[0]; got.Year() > 1954 {
		t.Errorf("record starts %s, want 1953 or 1954", got.Format("2006-01"))
	}
	if got := r.Years[len(r.Years)-1]; got < time.Now().Year()-1 {
		t.Errorf("last complete year %d, want the record kept current", got)
	}

	for _, tc := range []struct {
		start            int
		years            int
		cagrLo, cagrHi   float64
		volLo, volHi     float64
		worstYear        int
		worstLo, worstHi float64
	}{
		// 1990-2025 real: ~5.8%/yr at ~10.5% volatility, worst year 2022.
		{1990, 36, 0.045, 0.070, 0.08, 0.13, 2022, -0.25, -0.15},
		// 2000-2025 real: the lost decade drags the whole span to ~4.0%/yr.
		{2000, 26, 0.025, 0.055, 0.08, 0.13, 2022, -0.25, -0.15},
	} {
		_, index, seq, ok := r.Window(tc.start, 60)
		if !ok {
			t.Fatalf("no window from %d", tc.start)
		}
		if len(seq) != tc.years {
			t.Errorf("%d: %d years, want %d", tc.start, len(seq), tc.years)
		}
		cagr := math.Pow(index[len(index)-1]/index[0], 1/float64(len(seq))) - 1
		if cagr < tc.cagrLo || cagr > tc.cagrHi {
			t.Errorf("%d: real CAGR %.2f%%, want %.1f-%.1f%%", tc.start, cagr*100, tc.cagrLo*100, tc.cagrHi*100)
		}
		if vol := annualVol(seq); vol < tc.volLo || vol > tc.volHi {
			t.Errorf("%d: volatility %.2f%%, want %.1f-%.1f%%", tc.start, vol*100, tc.volLo*100, tc.volHi*100)
		}
		// The chart window and the replayed years must describe the same span:
		// one base point plus twelve months per year.
		if want := 12*len(seq) + 1; len(index) != want {
			t.Errorf("%d: %d monthly points, want %d", tc.start, len(index), want)
		}
		worst, at := math.Inf(1), 0
		for i, v := range seq {
			if v < worst {
				worst, at = v, tc.start+i
			}
		}
		if at != tc.worstYear {
			t.Errorf("%d: worst year %d, want %d", tc.start, at, tc.worstYear)
		}
		if worst < tc.worstLo || worst > tc.worstHi {
			t.Errorf("%d: worst year %.1f%%, want %.0f to %.0f%%", tc.start, worst*100, tc.worstLo*100, tc.worstHi*100)
		}
	}
}

// The compounded monthly index and the folded calendar years must tell the same
// story: a fold that slipped a month is the bug this catches.
func TestReferenceYearsMatchTheIndex(t *testing.T) {
	r, err := Reference()
	if err != nil {
		t.Fatal(err)
	}
	_, index, seq, ok := r.Window(2000, 60)
	if !ok {
		t.Fatal("no window from 2000")
	}
	for k, want := range seq {
		got := index[12*(k+1)]/index[12*k] - 1
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("year %d: index says %.6f, sequence says %.6f", 2000+k, got, want)
		}
	}
}

func TestReferenceRejectsAnUncoveredStart(t *testing.T) {
	r, err := Reference()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, ok := r.Window(1900, 40); ok {
		t.Error("1900 is before the record; want ok=false")
	}
	if _, _, _, ok := r.Window(time.Now().Year()+5, 40); ok {
		t.Error("a future start has no complete year; want ok=false")
	}
}

func TestCommonMonthsStopsAtAHole(t *testing.T) {
	a := map[int]float64{10: 1, 11: 1, 12: 1, 14: 1}
	b := map[int]float64{9: 1, 10: 1, 11: 1, 12: 1, 13: 1, 14: 1}
	if got := commonMonths(a, b); len(got) != 3 || got[0] != 10 || got[2] != 12 {
		t.Errorf("commonMonths = %v, want [10 11 12] (the hole at 13 ends the run)", got)
	}
	if got := commonMonths(map[int]float64{1: 1}, map[int]float64{5: 1}); got != nil {
		t.Errorf("disjoint series = %v, want nil", got)
	}
}

func TestMonthEnd(t *testing.T) {
	for _, tc := range []struct {
		t    time.Time
		want string
	}{
		{time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC), "2024-02-29"},
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-12-31"},
		{time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-04-30"},
	} {
		if got := monthEnd(monthKey(tc.t)).Format("2006-01-02"); got != tc.want {
			t.Errorf("monthEnd(%s) = %s, want %s", tc.t.Format("2006-01-02"), got, tc.want)
		}
	}
}

// annualVol is the sample standard deviation of a set of annual returns.
func annualVol(seq []float64) float64 {
	n := float64(len(seq))
	var mean float64
	for _, v := range seq {
		mean += v
	}
	mean /= n
	var v float64
	for _, x := range seq {
		v += (x - mean) * (x - mean)
	}
	return math.Sqrt(v / (n - 1))
}
