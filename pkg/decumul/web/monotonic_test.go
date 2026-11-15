package web

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// More capital must not worsen the fixed-rule probability on common draws,
// including the buffer cap and nominal gain-tax basis scaling.
func TestFixedCapitalMonotonicGrid(t *testing.T) {
	for _, years := range []int{20, 40, 60} {
		for _, buffer := range []float64{0, 3, 10} {
			for _, pension := range []float64{0, 30000} {
				pr := Params{Capital: 1e6, NeedAnnual: 50000, Years: years, BufferYears: buffer, BufferReturn: .005, TaxRate: .328, GainFrac: .5, Inflation: .02, PensionAnnual: pension, PensionYear: 10}
				p := pr.plan()
				p.Source = scenario.ParametricSource{Mu: .04, Sigma: .18, Df: 5, Periods: years}
				d := p.Draw(600, 4, 7)
				prev := 2.0
				for capital := 200000.; capital <= 2500000; capital += 50000 {
					pr.Capital = capital
					q := pr.plan()
					v := q.RuinProbOn(d, 4)
					if v > prev {
						t.Fatalf("horizon=%d buffer=%v pension=%v capital=%v ruin=%v previous=%v", years, buffer, pension, capital, v, prev)
					}
					prev = v
				}
			}
		}
	}
}

// Re-centering the adaptive bands does not restore monotonicity. This
// counterexample prevents treating an adaptive capital bisection as a minimum.
func TestAdaptiveCapitalCounterexample(t *testing.T) {
	pr := Params{Capital: 1300000, NeedAnnual: 50000, Years: 20, TaxRate: .328, GainFrac: .5, Inflation: .02, Guardrails: true}
	p := pr.plan()
	p.Source = scenario.ParametricSource{Mu: .04, Sigma: .18, Df: 5, Periods: 20}
	d := p.Draw(600, 4, 7)
	lower := p.RuinProbOn(d, 4)
	pr.Capital = 1350000
	q := pr.plan()
	upper := q.RuinProbOn(d, 4)
	if lower != 16./600 || upper != 17./600 {
		t.Fatalf("counterexample changed: %v -> %v", lower, upper)
	}
}
