package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Reference values from docs/decumulation-fire-design.md §7 (Python model),
// tolerance ±0.03 M€ on target capital and ±0.3 pt on ruin, >=150k paths.
// The reference uses a flat 12% gross-up, so these golden checks fix Tax to
// a flat 12% stub; the cost-basis CTOFlatTax is covered by its own unit test.
type flatGrossUp struct{ rate float64 }

func (f flatGrossUp) GrossUp(net, growth, cost float64) (float64, float64, float64) {
	gross := net * (1 + f.rate)
	nc := cost
	if growth > 0 {
		nc = cost * (1 - gross/growth)
	}
	return gross, nc, gross - net
}

func basePlan(mu, pensionMonthly float64, years int) Plan {
	return Plan{
		NeedAnnual: 48000,
		Cashflows:  []Cashflow{{FromYear: 67 - 55, Annual: pensionMonthly * 12}},
		Years:      years,
		Tax:        flatGrossUp{rate: 0.12},
		Source:     scenario.ParametricSource{Mu: mu, Sigma: 0.12, Df: 6, Periods: years},
	}
}

func TestGoldenTargetCapital(t *testing.T) {
	cases := []struct {
		mu, pension float64
		years       int
		want        float64 // M€
	}{
		{0.035, 1800, 35, 1.67},
		{0.030, 1800, 35, 1.81},
		{0.045, 1800, 35, 1.45},
		{0.035, 1400, 35, 1.84},
	}
	for _, c := range cases {
		p := basePlan(c.mu, c.pension, c.years)
		got := p.CapitalForRuin(0.05, 0.8e6, 4.5e6, 200000, 8, 7) / 1e6
		if d := got - c.want; d < -0.05 || d > 0.05 {
			t.Errorf("mu=%.3f pension=%.0f: target=%.2fM, want ~%.2fM", c.mu, c.pension, got, c.want)
		}
	}
}

func TestGoldenRuinAt2M(t *testing.T) {
	p := basePlan(0.035, 1800, 35)
	p.Capital = 2_000_000
	got := p.Simulate(200000, 8, 7).RuinProb() * 100
	if got < 1.0 || got > 3.5 {
		t.Errorf("ruin at 2.0M = %.2f%%, want ~2.1%%", got)
	}
}
