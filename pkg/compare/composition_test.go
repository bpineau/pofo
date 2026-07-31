package compare

import (
	"reflect"
	"testing"
)

// breakdownSlices must be deterministic even when several categories carry
// the exact same weight: agg is a map, so ties are broken by label
// (ascending) rather than by map-iteration order, keeping the rendered
// report byte-stable run to run.
func TestBreakdownSlicesTiesDeterministic(t *testing.T) {
	agg := map[string]float64{"Delta": 25, "Alpha": 25, "Charlie": 25, "Bravo": 25}

	first := breakdownSlices(agg, 10)
	for i := range 50 {
		if got := breakdownSlices(agg, 10); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d differs from the first: %+v vs %+v", i, got, first)
		}
	}

	// All weights tie, so the order is purely the label tiebreak (ascending).
	want := []string{"Alpha", "Bravo", "Charlie", "Delta"}
	got := make([]string, len(first))
	for i, s := range first {
		got[i] = s.Label
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tie order = %v, want %v (label ascending)", got, want)
	}
}

// Distinct weights are ordered by weight descending, unchanged by the
// tiebreak, so existing reports keep their exact slice order.
func TestBreakdownSlicesWeightOrder(t *testing.T) {
	agg := map[string]float64{"low": 10, "high": 60, "mid": 30}
	got := breakdownSlices(agg, 10)
	want := []string{"high", "mid", "low"}
	labels := make([]string, len(got))
	for i, s := range got {
		labels[i] = s.Label
	}
	if !reflect.DeepEqual(labels, want) {
		t.Errorf("weight order = %v, want %v", labels, want)
	}
}
