package decumul

import (
	"math/rand/v2"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// stubSource is a minimal non-parametric Source (it carries no Mu) used to
// check that sweeping Mu against it is rejected rather than silently ignored.
type stubSource struct{ n int }

func (s stubSource) Len() int                          { return s.n }
func (s stubSource) Draw(*rand.Rand) scenario.Sequence { return make(scenario.Sequence, s.n) }

func sweepPlan() Plan {
	return Plan{Capital: 1_500_000, NeedAnnual: 48000, Years: 35, Tax: CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35}}
}

func TestSweep1DBufferMonotoneRuin(t *testing.T) {
	p := sweepPlan()
	pts, err := p.Sweep1D(BufferYears, []float64{0, 2, 4, 6}, 8000, 4, 7)
	if err != nil {
		t.Fatalf("Sweep1D: %v", err)
	}
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
	s, err := p.Sweep2D(BufferYears, Mu, []float64{0, 3}, []float64{0.03, 0.05}, 4000, 4, 7)
	if err != nil {
		t.Fatalf("Sweep2D: %v", err)
	}
	if len(s.Ruin) != 2 || len(s.Ruin[0]) != 2 {
		t.Fatalf("surface shape %dx%d, want 2x2", len(s.Ruin), len(s.Ruin[0]))
	}
}

// Sweeping Mu on a non-parametric source used to be a silent no-op (a flat
// surface). It must now report the constraint as an error rather than mislead.
func TestSweepMuOnNonParametricErrors(t *testing.T) {
	p := sweepPlan()
	p.Source = stubSource{n: 35}

	if _, err := p.Sweep1D(Mu, []float64{0.03, 0.05}, 1000, 2, 7); err == nil {
		t.Error("Sweep1D(Mu) on a bootstrap source: want error, got nil")
	}
	if _, err := p.Sweep2D(BufferYears, Mu, []float64{0, 3}, []float64{0.03, 0.05}, 1000, 2, 7); err == nil {
		t.Error("Sweep2D(_, Mu) on a bootstrap source: want error, got nil")
	}
	// A parameter that does apply to any source stays error-free.
	if _, err := p.Sweep1D(BufferYears, []float64{0, 2}, 1000, 2, 7); err != nil {
		t.Errorf("Sweep1D(BufferYears) on a bootstrap source: unexpected error %v", err)
	}
}
