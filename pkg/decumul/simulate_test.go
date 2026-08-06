package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestSimulateDeterministic(t *testing.T) {
	p := Plan{
		Capital: 1_500_000, NeedAnnual: 48000, Years: 35,
		Tax:    CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	a := p.Simulate(20000, 4, 7).RuinProb()
	b := p.Simulate(20000, 4, 7).RuinProb()
	if a != b {
		t.Errorf("not reproducible: %.4f vs %.4f", a, b)
	}
	if a < 0 || a > 1 {
		t.Errorf("ruin prob out of range: %.4f", a)
	}
}

func TestSimulateMoreCapitalLowerRuin(t *testing.T) {
	mk := func(c float64) Plan {
		return Plan{Capital: c, NeedAnnual: 48000, Years: 35, Tax: CTOFlatTax{Rate: 0.30},
			Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35}}
	}
	low := mk(1_200_000).Simulate(20000, 4, 7).RuinProb()
	high := mk(2_500_000).Simulate(20000, 4, 7).RuinProb()
	if !(high < low) {
		t.Errorf("more capital should lower ruin: low=%.4f high=%.4f", low, high)
	}
}
