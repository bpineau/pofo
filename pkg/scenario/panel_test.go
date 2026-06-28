package scenario

import (
	"math"
	"testing"
)

func TestPanelCombine(t *testing.T) {
	p := Panel{
		Returns: [][]float64{
			{0.10, -0.05, 0.20},
			{0.00, 0.02, -0.01},
		},
		Weights: []float64{0.6, 0.4},
	}
	got := p.Combine(nil)
	want := Sequence{0.06, -0.022, 0.116}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("period %d: got %.4f, want %.4f", i, got[i], want[i])
		}
	}
}

func TestPanelCombineReweight(t *testing.T) {
	p := Panel{Returns: [][]float64{{0.10}, {0.00}}, Weights: []float64{0.6, 0.4}}
	if got := p.Combine([]float64{1, 0}); math.Abs(got[0]-0.10) > 1e-9 {
		t.Errorf("reweighted got %.4f, want 0.10", got[0])
	}
}
