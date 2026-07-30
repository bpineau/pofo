package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func sweepPlan() Plan {
	return Plan{Capital: 1_500_000, NeedAnnual: 48000, Years: 35, Tax: CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35}}
}

func TestSweep1DBufferMonotoneRuin(t *testing.T) {
	p := sweepPlan()
	pts := p.Sweep1D(BufferYears, []float64{0, 2, 4, 6}, 8000, 4, 7)
	if len(pts) != 4 {
		t.Fatalf("len = %d, want 4", len(pts))
	}
	for _, pt := range pts {
		if pt.RuinProb < 0 || pt.RuinProb > 1 {
			t.Errorf("ruin out of range: %.3f", pt.RuinProb)
		}
	}
}

func TestSweep2DShape(t *testing.T) {
	p := sweepPlan()
	s := p.Sweep2D(BufferYears, Mu, []float64{0, 3}, []float64{0.03, 0.05}, 4000, 4, 7)
	if len(s.Ruin) != 2 || len(s.Ruin[0]) != 2 {
		t.Fatalf("surface shape %dx%d, want 2x2", len(s.Ruin), len(s.Ruin[0]))
	}
}
