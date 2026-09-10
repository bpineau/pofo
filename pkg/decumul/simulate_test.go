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

// TestSimulateOnMatchesRunPath pins the arena: the series a path gets out of
// SimulateOn's one shared block must be exactly what the standalone kernel
// builds for itself, path by path and point by point. It is the guard on a
// shared FIRE URL reproducing byte for byte.
func TestSimulateOnMatchesRunPath(t *testing.T) {
	p := Plan{
		Capital: 800_000, NeedAnnual: 32_000, Years: 25,
		Tax: CTOFlatTax{Rate: 0.30}, Buffer: BufferSleeve{Years: 2},
		Cashflows: []Cashflow{{FromYear: 10, Annual: 9_000}},
		Flex:      FlexRule{Threshold: 0.20, Cut: 0.10},
		Source:    scenario.ParametricSource{Mu: 0.05, Sigma: 0.15, Df: 5, Periods: 25},
	}
	draws := p.Draw(64, 4, 7)
	e := p.SimulateOn(draws, 4)
	for i, seq := range draws.Returns {
		want := p.RunPath(seq, Lives{})
		got := e.Paths[i]
		for k := range want.Wealth {
			if got.Wealth[k] != want.Wealth[k] {
				t.Fatalf("path %d wealth[%d] = %v, want %v", i, k, got.Wealth[k], want.Wealth[k])
			}
		}
		for k := range want.Spend {
			if got.Spend[k] != want.Spend[k] {
				t.Fatalf("path %d spend[%d] = %v, want %v", i, k, got.Spend[k], want.Spend[k])
			}
		}
		if got.Ruined != want.Ruined || got.TaxPaid != want.TaxPaid || got.Withdrawn != want.Withdrawn {
			t.Fatalf("path %d summary %+v, want %+v", i, got, want)
		}
	}
}
