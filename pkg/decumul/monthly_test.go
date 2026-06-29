package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// zeros returns n zero monthly returns.
func zeros(n int) scenario.Sequence { return make(scenario.Sequence, n) }

// With zero returns, no tax and no buffer, monthly withdrawals of NeedAnnual/12
// deplete the capital on the same schedule as the annual kernel.
func TestRunPathMonthlyDepletion(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 25000, Years: 5, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPathMonthly(zeros(5 * 12))
	if !res.Ruined {
		t.Errorf("expected ruin: 100k - 5*25k < 0")
	}
	if len(res.Wealth) != 6 {
		t.Fatalf("Wealth len = %d, want 6 (annual granularity)", len(res.Wealth))
	}
	if math.Abs(res.Wealth[0]-100000) > 1e-6 {
		t.Errorf("Wealth[0] = %.0f, want 100000", res.Wealth[0])
	}
}

// The growth sleeve compounds one monthly return per step: with no withdrawals
// and no buffer, a year of +1%/month ends at capital × 1.01¹².
func TestRunPathMonthlyGrowthCompounding(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 0, Years: 1, Tax: CTOFlatTax{Rate: 0}}
	seq := scenario.Sequence{0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01}
	res := p.RunPathMonthly(seq)
	want := 100000 * math.Pow(1.01, 12)
	if math.Abs(res.Wealth[1]-want) > 1e-6 {
		t.Errorf("Wealth[1] = %.2f, want %.2f", res.Wealth[1], want)
	}
}

// The buffer's annual real return is applied monthly as its 12th root, so an
// all-buffer year compounds to exactly the annual figure.
func TestRunPathMonthlyBufferCompounding(t *testing.T) {
	p := Plan{
		Capital:    100000,
		NeedAnnual: 24000,
		Years:      1,
		// A pension fully covers the need (net need 0), and Years×Need > capital
		// so the whole capital is the buffer: it compounds untouched.
		Cashflows: []Cashflow{{FromYear: 0, Annual: 24000}},
		Buffer:    BufferSleeve{Years: 100, RealReturn: 0.1268},
		Tax:       CTOFlatTax{Rate: 0},
	}
	res := p.RunPathMonthly(zeros(12))
	want := 100000 * 1.1268
	if math.Abs(res.Wealth[1]-want) > 1e-6 {
		t.Errorf("Wealth[1] = %.2f, want %.2f (annual buffer return applied monthly)", res.Wealth[1], want)
	}
}

// A large capital with positive returns survives, like the annual kernel.
func TestRunPathMonthlySurvives(t *testing.T) {
	p := Plan{Capital: 1_000_000, NeedAnnual: 20000, Years: 10, Tax: CTOFlatTax{Rate: 0}}
	seq := make(scenario.Sequence, 120)
	for i := range seq {
		seq[i] = 0.004 // ~4.9%/yr
	}
	res := p.RunPathMonthly(seq)
	if res.Ruined {
		t.Errorf("did not expect ruin")
	}
	if res.Wealth[10] <= 0 {
		t.Errorf("final wealth = %.0f, want > 0", res.Wealth[10])
	}
}

// A monthly Plan simulates end to end through the shared dispatcher: Simulate
// must run the monthly kernel and produce a plausible ruin probability.
func TestSimulateMonthly(t *testing.T) {
	p := Plan{
		Capital: 1_000_000, NeedAnnual: 45000, Years: 30, Buffer: BufferSleeve{Years: 2},
		Tax:     CTOFlatTax{Rate: 0.30}, Monthly: true,
		Source:  scenario.ParametricSource{Mu: 0.003, Sigma: 0.035, Df: 6, Periods: 30 * 12},
	}
	o := p.Simulate(3000, 4, 7).Outcome()
	if o.RuinProb < 0 || o.RuinProb > 1 {
		t.Errorf("ruin out of range: %.3f", o.RuinProb)
	}
	if len(p.Simulate(10, 1, 1).Paths[0].Wealth) != 31 {
		t.Errorf("monthly path wealth should be annual-granular (Years+1 points)")
	}
}

// Cashflows are year-valued: a pension from year 2 reduces the monthly need
// only once the second year begins.
func TestRunPathMonthlyCashflowYearValued(t *testing.T) {
	p := Plan{Capital: 500000, NeedAnnual: 24000, Years: 3,
		Cashflows: []Cashflow{{FromYear: 2, Annual: 12000}}, Tax: CTOFlatTax{Rate: 0}}
	// Year 0 and 1 withdraw 2000/mo; year 2 withdraws (24000-12000)/12 = 1000/mo.
	res := p.RunPathMonthly(zeros(36))
	wantWithdrawn := 2*24000.0 + 1*12000.0
	if math.Abs(res.Withdrawn-wantWithdrawn) > 1e-6 {
		t.Errorf("Withdrawn = %.0f, want %.0f", res.Withdrawn, wantWithdrawn)
	}
}
