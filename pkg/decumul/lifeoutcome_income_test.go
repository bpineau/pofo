package decumul

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// flatSrc yields a constant periodic return.
type flatSrc struct {
	r float64
	n int
}

func (f flatSrc) Len() int { return f.n }
func (f flatSrc) Draw(*rand.Rand) scenario.Sequence {
	s := make(scenario.Sequence, f.n)
	for i := range s {
		s[i] = f.r
	}
	return s
}

// A pension outlives the portfolio. Running out of money stops the sleeves,
// not the pension cheque: a household with 200 k of capital, a 40 k need and a
// 50 k pension from year 10 is broke for five years and then lives better than
// it planned, and Received/IncomeMean must say so. Stopping the accounting at
// the ruin year reported a mean income of 5 k on a 50 k pension.
func TestIncomeOutlivesTheRuinedPortfolio(t *testing.T) {
	for _, monthly := range []bool{false, true} {
		src := scenario.Source(flatSrc{n: 40})
		if monthly {
			src = flatSrc{n: 40 * 12}
		}
		p := Plan{Capital: 200_000, NeedAnnual: 40_000, Years: 40, Tax: CTOFlatTax{Rate: 0},
			Cashflows: []Cashflow{{FromYear: 10, Annual: 50_000}},
			Monthly:   monthly, Source: src}
		e := p.Simulate(1, 1, 7)
		r := e.Paths[0]
		const want = 30 * 50_000.0 // years 10..39
		if math.Abs(r.Received-want) > 1e-6 {
			t.Errorf("monthly=%v: Received = %.0f, want %.0f", monthly, r.Received, want)
		}
		// The five funded years delivered 40 k each, so the lived income is
		// (5*40k + 30*50k) over the forty years of the plan.
		wantMean := (5*40_000.0 + want) / 40
		if got := e.LifeOutcome().IncomeMean; math.Abs(got-wantMean) > 1 {
			t.Errorf("monthly=%v: IncomeMean = %.0f, want %.0f", monthly, got, wantMean)
		}
	}
}

// The income a path never lives to see is not collected: a household that dies
// at year 12 collects two years of the pension, not thirty.
func TestIncomeAfterRuinStopsAtDeath(t *testing.T) {
	p := Plan{Capital: 200_000, NeedAnnual: 40_000, Years: 40, Tax: CTOFlatTax{Rate: 0},
		Cashflows: []Cashflow{{FromYear: 10, Annual: 50_000}},
		Lifetime:  &Lifetime{Self: Life{Age: 60, Law: diesAt{12}}},
		Source:    flatSrc{n: 40}}
	r := p.Simulate(1, 1, 7).Paths[0]
	if want := 2 * 50_000.0; math.Abs(r.Received-want) > 1e-6 {
		t.Errorf("Received = %.0f, want %.0f (alive for years 0..11)", r.Received, want)
	}
}
