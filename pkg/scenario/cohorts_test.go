package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestHistoricalCohorts(t *testing.T) {
	p := Panel{Returns: [][]float64{{0.1, 0.2, 0.3, 0.4, 0.5}}, Weights: []float64{1}}
	h := HistoricalCohorts{Panel: p, Periods: 3}
	if h.Count() != 3 { // windows starting at 0,1,2
		t.Fatalf("Count = %d, want 3", h.Count())
	}
	got := h.Cohort(1)
	want := Sequence{0.2, 0.3, 0.4}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("cohort[1][%d] = %.4f, want %.4f", i, got[i], want[i])
		}
	}
	rng := rand.New(rand.NewPCG(1, 1))
	if len(h.Draw(rng)) != 3 {
		t.Errorf("Draw len wrong")
	}
}
