package decumul

import (
	"math"
	"testing"
)

// A single total-loss path drives the min worst-decade CAGR to -100%, but the
// robust p5 worst-decade must not be dragged all the way down by it.
func TestOutcomeWorst10yRobust(t *testing.T) {
	paths := make([]PathResult, 20)
	flat := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100}
	for i := range paths {
		paths[i] = PathResult{Wealth: append([]float64(nil), flat...)}
	}
	// One path loses everything within the decade (ends at 0): a -100% decade.
	paths[0] = PathResult{Wealth: []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 0}, Ruined: true}
	o := Ensemble{Years: 10, Paths: paths}.Outcome()

	if math.Abs(o.Worst10yCAGR-(-1)) > 1e-9 {
		t.Errorf("Worst10yCAGR = %.4f, want -1 (the realised total loss)", o.Worst10yCAGR)
	}
	// p5 of [-1, 0, 0, …] (n=20) interpolates to -0.05: robust to the one ruin.
	if math.Abs(o.Worst10yP5-(-0.05)) > 1e-9 {
		t.Errorf("Worst10yP5 = %.4f, want -0.05", o.Worst10yP5)
	}
	if !(o.Worst10yP5 > o.Worst10yCAGR) {
		t.Errorf("p5 (%.4f) should sit above the min (%.4f)", o.Worst10yP5, o.Worst10yCAGR)
	}
}

func TestOutcomeBasics(t *testing.T) {
	// two paths: one survives flat at 100, one ruined.
	e := Ensemble{Years: 2, Paths: []PathResult{
		{Wealth: []float64{100, 100, 100}, Ruined: false},
		{Wealth: []float64{100, 50, 0}, Ruined: true},
	}}
	o := e.Outcome()
	if math.Abs(o.RuinProb-0.5) > 1e-9 {
		t.Errorf("RuinProb = %.3f, want 0.5", o.RuinProb)
	}
	if o.TerminalP5 > o.TerminalP50 {
		t.Errorf("p5 (%.1f) should be <= p50 (%.1f)", o.TerminalP5, o.TerminalP50)
	}
	if o.CDaR < 0 || o.CDaR > 1 {
		t.Errorf("CDaR out of range: %.3f", o.CDaR)
	}
}
