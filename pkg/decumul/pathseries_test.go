package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// A depleting path records the year ruin latched and the net spending actually
// delivered each year: 25k for four years, then nothing.
func TestRunPathRuinYearAndSpend(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 25000, Years: 5, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0, 0, 0})
	if !res.Ruined {
		t.Fatalf("expected ruin")
	}
	if res.RuinYear != 3 {
		t.Errorf("RuinYear = %d, want 3 (wealth reaches 0 at the 4th withdrawal)", res.RuinYear)
	}
	if len(res.Spend) != 5 {
		t.Fatalf("Spend len = %d, want 5 (one per year)", len(res.Spend))
	}
	want := []float64{25000, 25000, 25000, 25000, 0}
	for k, w := range want {
		if math.Abs(res.Spend[k]-w) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want %.0f", k, res.Spend[k], w)
		}
	}
}

// A surviving path keeps RuinYear at -1 and delivers the full need every year.
func TestRunPathNoRuinYear(t *testing.T) {
	p := Plan{Capital: 1_000_000, NeedAnnual: 20000, Years: 3, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0.05, 0.05, 0.05})
	if res.Ruined {
		t.Fatalf("did not expect ruin")
	}
	if res.RuinYear != -1 {
		t.Errorf("RuinYear = %d, want -1 on a surviving path", res.RuinYear)
	}
	for k, s := range res.Spend {
		if math.Abs(s-20000) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want 20000", k, s)
		}
	}
}

// The spend series reflects the flex cut: a deep drawdown year delivers the
// reduced spending, not the nominal need.
func TestRunPathSpendReflectsFlexCut(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3,
		Flex: FlexRule{Threshold: 0.20, Cut: 0.25}, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{-0.5, 0, 0})
	// Year 0: no drawdown yet, full 4000. Year 1: ~52% drawdown, cut to 3000.
	if math.Abs(res.Spend[0]-4000) > 1e-6 {
		t.Errorf("Spend[0] = %.0f, want 4000", res.Spend[0])
	}
	if math.Abs(res.Spend[1]-3000) > 1e-6 {
		t.Errorf("Spend[1] = %.0f, want 3000 (25%% flex cut)", res.Spend[1])
	}
}

// The monthly kernel aggregates the twelve monthly deliveries into the year's
// spend and records the ruin year at annual granularity.
func TestRunPathMonthlyRuinYearAndSpend(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 25000, Years: 5, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPathMonthly(zeros(5 * 12))
	if !res.Ruined {
		t.Fatalf("expected ruin")
	}
	if res.RuinYear != 4 {
		t.Errorf("RuinYear = %d, want 4 (the buffer-less capital covers exactly 48 months)", res.RuinYear)
	}
	if len(res.Spend) != 5 {
		t.Fatalf("Spend len = %d, want 5", len(res.Spend))
	}
	if math.Abs(res.Spend[0]-25000) > 1e-6 {
		t.Errorf("Spend[0] = %.0f, want 25000 (12 monthly deliveries)", res.Spend[0])
	}
	if res.Spend[4] > 1e-6 {
		t.Errorf("Spend[4] = %.0f, want 0 after ruin", res.Spend[4])
	}
}
