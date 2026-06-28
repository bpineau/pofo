package decumul

import (
	"math"
	"testing"
)

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
