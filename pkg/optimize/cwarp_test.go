package optimize

import (
	"math"
	"testing"
)

// cwarpSeries builds a replacement (equity-like, with a drawdown), a perfectly
// correlated equity-beta asset, and an anti-correlated diversifier with carry.
func cwarpSeries() (repl, equity, diversifier []float64) {
	const n = 300
	repl = make([]float64, n)
	equity = make([]float64, n)
	diversifier = make([]float64, n)
	for i := 0; i < n; i++ {
		repl[i] = 0.0010 + 0.006*math.Sin(float64(i)*0.3)
		if i >= 100 && i < 115 {
			repl[i] = -0.010
		}
		equity[i] = repl[i]                // adds equity beta on top of equity beta
		diversifier[i] = -repl[i] + 0.0007 // hedge plus positive carry
	}
	return
}

// TestSolveCWARPFavorsDiversifier: given the choice between more equity beta and
// an anti-correlated diversifier, the CWARP optimizer loads the diversifier and
// achieves a positive score.
func TestSolveCWARPFavorsDiversifier(t *testing.T) {
	repl, equity, div := cwarpSeries()
	res, err := SolveCWARP([][]float64{equity, div}, repl, Spec{Objective: CWARP})
	if err != nil {
		t.Fatal(err)
	}
	if res.CWARP <= 0 {
		t.Fatalf("achieved CWARP = %.2f, want > 0", res.CWARP)
	}
	if res.Weights[1] <= res.Weights[0] {
		t.Fatalf("diversifier weight %.3f should exceed equity weight %.3f", res.Weights[1], res.Weights[0])
	}
	if s := res.Weights[0] + res.Weights[1]; math.Abs(s-1) > 1e-6 {
		t.Fatalf("weights sum %.6f, want 1", s)
	}
}

func TestParseSpecCWARP(t *testing.T) {
	s, err := ParseSpec("cwarp,max-weight:50")
	if err != nil {
		t.Fatal(err)
	}
	if s.Objective != CWARP {
		t.Fatalf("objective = %q, want cwarp", s.Objective)
	}
	if math.Abs(s.MaxWeight-0.5) > 1e-9 {
		t.Fatalf("max-weight = %.3f, want 0.5", s.MaxWeight)
	}
}
