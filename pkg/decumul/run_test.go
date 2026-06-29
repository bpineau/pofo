package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// With zero returns, no tax and no pension, capital depletes by need/year.
func TestRunPathDepletion(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 25000, Years: 5, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0, 0, 0})
	if !res.Ruined {
		t.Errorf("expected ruin: 100k - 5*25k < 0")
	}
	if len(res.Wealth) != 6 {
		t.Fatalf("Wealth len = %d, want 6", len(res.Wealth))
	}
	if math.Abs(res.Wealth[0]-100000) > 1e-6 {
		t.Errorf("Wealth[0] = %.0f, want 100000", res.Wealth[0])
	}
}

// When embedded gains make the gross-up exceed the available growth, the year
// under-delivers: it must latch ruin and account only the net actually
// withdrawn (with a non-negative tax), not the full requested need.
func TestRunPathUnderDelivery(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 70000, Years: 2, Tax: CTOFlatTax{Rate: 0.5}}
	res := p.RunPath(scenario.Sequence{1.0, 0})
	if !res.Ruined {
		t.Errorf("expected ruin: year 2 cannot gross up 70k net from 60k of growth")
	}
	// Year 1 delivers 70k (no gain yet); year 2 caps at 60k gross, 15k tax,
	// so only 45k net reaches the household: 70k + 45k = 115k withdrawn.
	if math.Abs(res.Withdrawn-115000) > 1 {
		t.Errorf("Withdrawn = %.0f, want 115000 (real net, not the requested 140000)", res.Withdrawn)
	}
	if res.TaxPaid < 0 {
		t.Errorf("TaxPaid = %.0f, must never be negative", res.TaxPaid)
	}
	if math.Abs(res.TaxPaid-15000) > 1 {
		t.Errorf("TaxPaid = %.0f, want 15000", res.TaxPaid)
	}
}

// A high enough capital with positive returns survives.
func TestRunPathSurvives(t *testing.T) {
	p := Plan{Capital: 1_000_000, NeedAnnual: 20000, Years: 10, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05})
	if res.Ruined {
		t.Errorf("did not expect ruin")
	}
	if res.Wealth[10] <= 0 {
		t.Errorf("final wealth = %.0f, want > 0", res.Wealth[10])
	}
}
