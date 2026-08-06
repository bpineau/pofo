package metrics

import "testing"

func TestLongestUnderperformanceOngoing(t *testing.T) {
	// The series is flat while the benchmark climbs: the value/benchmark
	// ratio peaks at day 0 and never recovers.
	dates := days(5)
	values := []float64{100, 100, 100, 100, 100}
	bench := []float64{100, 101, 102, 103, 104}
	days, _, _, ongoing, ok := LongestUnderperformance(dates, values, dates, bench)
	if !ok {
		t.Fatal("expected ok")
	}
	if !ongoing {
		t.Error("expected an ongoing underperformance spell")
	}
	if days != 4 {
		t.Errorf("days = %d, want 4", days)
	}
}

func TestLongestUnderperformanceRecovers(t *testing.T) {
	// The series lags then overtakes the benchmark: the spell is bounded.
	dates := days(6)
	values := []float64{100, 100, 100, 100, 120, 130}
	bench := []float64{100, 101, 102, 103, 104, 105}
	d, _, _, ongoing, ok := LongestUnderperformance(dates, values, dates, bench)
	if !ok {
		t.Fatal("expected ok")
	}
	if ongoing {
		t.Error("spell should have recovered")
	}
	if d <= 0 {
		t.Errorf("days = %d, want positive", d)
	}
}

func TestLongestUnderperformanceNever(t *testing.T) {
	// The series always beats the benchmark: no underperformance.
	dates := days(5)
	values := []float64{100, 102, 104, 106, 108}
	bench := []float64{100, 100, 100, 100, 100}
	d, _, _, _, ok := LongestUnderperformance(dates, values, dates, bench)
	if !ok {
		t.Fatal("expected ok")
	}
	if d != 0 {
		t.Errorf("days = %d, want 0", d)
	}
}
