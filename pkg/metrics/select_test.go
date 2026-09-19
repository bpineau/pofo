package metrics

import (
	"math"
	"math/rand/v2"
	"sort"
	"testing"
)

// fullSortQuantiles is the definition Quantiles must match exactly: sort
// everything, then interpolate. Quantiles places only the ranks it needs, so
// this reference is what keeps that shortcut honest.
func fullSortQuantiles(xs []float64, qs ...float64) []float64 {
	if len(xs) == 0 {
		return nil
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	out := make([]float64, len(qs))
	for i, q := range qs {
		pos := q * float64(len(sorted)-1)
		lo, hi := int(math.Floor(pos)), int(math.Ceil(pos))
		if lo < 0 {
			lo = 0
		}
		if hi >= len(sorted) {
			hi = len(sorted) - 1
		}
		out[i] = sorted[lo] + (pos-float64(lo))*(sorted[hi]-sorted[lo])
	}
	return out
}

// samples covers the shapes the selection has to survive: continuous values,
// heavy ties (a wealth column where every ruined path sits at zero), an
// all-identical sample, an already sorted one, and sizes on both sides of the
// cutoff below which the code falls back to a full sort.
func samples(rng *rand.Rand, n, kind int) []float64 {
	xs := make([]float64, n)
	for i := range xs {
		switch kind {
		case 0:
			xs[i] = rng.NormFloat64()
		case 1:
			xs[i] = float64(rng.IntN(3)) // heavy ties
		case 2:
			xs[i] = 42 // all identical
		case 3:
			xs[i] = float64(i) // already sorted
		default:
			if rng.IntN(3) == 0 {
				xs[i] = 0 // a third of the paths ruined
			} else {
				xs[i] = rng.Float64() * 1e6
			}
		}
	}
	return xs
}

func TestQuantilesMatchesFullSort(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	qs := []float64{0, 0.01, 0.05, 0.25, 0.5, 0.75, 0.95, 0.99, 1, 0.33}
	for trial := range 2000 {
		xs := samples(rng, 1+rng.IntN(400), trial%5)
		want := fullSortQuantiles(xs, qs...)
		got := Quantiles(xs, qs...)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("n=%d kind=%d q=%v: got %v, want %v", len(xs), trial%5, qs[i], got[i], want[i])
			}
		}
	}
}

// TestQuantilesKeepsInput guards the documented promise that the sample is not
// reordered under the caller's feet: the selection works on a copy.
func TestQuantilesKeepsInput(t *testing.T) {
	xs := samples(rand.New(rand.NewPCG(3, 4)), 500, 0)
	before := append([]float64(nil), xs...)
	Quantiles(xs, 0.05, 0.5, 0.95)
	for i := range xs {
		if xs[i] != before[i] {
			t.Fatalf("input reordered at %d", i)
		}
	}
}

// TestQuantilesNaN pins the fallback: a sample holding a NaN cannot be
// partitioned (every comparison is false), so it is fully sorted, exactly as
// before the selection existed.
func TestQuantilesNaN(t *testing.T) {
	xs := make([]float64, 200)
	for i := range xs {
		xs[i] = float64(200 - i)
	}
	xs[7] = math.NaN()
	qs := []float64{0, 0.5, 1}
	want, got := fullSortQuantiles(xs, qs...), Quantiles(xs, qs...)
	for i := range want {
		if got[i] != want[i] && !(math.IsNaN(got[i]) && math.IsNaN(want[i])) {
			t.Fatalf("q=%v: got %v, want %v", qs[i], got[i], want[i])
		}
	}
}

func TestTopKMatchesFullSort(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	for trial := range 2000 {
		xs := samples(rng, 1+rng.IntN(400), trial%5)
		n := 1 + rng.IntN(len(xs))
		sorted := append([]float64(nil), xs...)
		sort.Sort(sort.Reverse(sort.Float64Slice(sorted)))
		got := TopK(xs, n)
		if len(got) != n {
			t.Fatalf("TopK(n=%d) returned %d values", n, len(got))
		}
		for i := range got {
			if got[i] != sorted[i] {
				t.Fatalf("n=%d kind=%d at %d: got %v, want %v", n, trial%5, i, got[i], sorted[i])
			}
		}
	}
}

func TestTopKBounds(t *testing.T) {
	xs := []float64{3, 1, 2}
	if got := TopK(xs, 0); got != nil {
		t.Errorf("TopK(_, 0) = %v, want nil", got)
	}
	if got := TopK(nil, 2); got != nil {
		t.Errorf("TopK(nil, 2) = %v, want nil", got)
	}
	got := TopK(xs, 9)
	if len(got) != 3 || got[0] != 3 || got[1] != 2 || got[2] != 1 {
		t.Errorf("TopK(_, 9) = %v, want [3 2 1]", got)
	}
	if xs[0] != 3 || xs[1] != 1 || xs[2] != 2 {
		t.Errorf("input reordered: %v", xs)
	}
}
