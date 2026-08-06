package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// The flex cut can also trigger on the current withdrawal rate (Ben's written
// rule: cut when drawdown > 20% OR current rate > 3.6%): here there is no
// drawdown at all but the rate is above the trigger, so the cut applies.
func TestFlexCutTriggersOnWithdrawalRate(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 5000, Years: 2,
		Flex: FlexRule{Threshold: 0.99, WRThreshold: 0.04, Cut: 0.20},
		Tax:  CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0})
	// wr = 5000/100000 = 5% > 4%: spend 4000, not 5000.
	if math.Abs(res.Spend[0]-4000) > 1e-6 {
		t.Errorf("Spend[0] = %.0f, want 4000 (WR-triggered cut)", res.Spend[0])
	}
}

// Without the WR trigger and with no drawdown, the same plan spends the full
// need: WRThreshold zero keeps today's drawdown-only semantics.
func TestFlexCutWRThresholdZeroInactive(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 5000, Years: 2,
		Flex: FlexRule{Threshold: 0.99, Cut: 0.20}, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0})
	if math.Abs(res.Spend[0]-5000) > 1e-6 {
		t.Errorf("Spend[0] = %.0f, want 5000 (no trigger)", res.Spend[0])
	}
}

// The ratchet raises the spending level by Step once real wealth exceeds
// Trigger times the initial capital, and keeps compounding while it stays
// above (Kitces-style only-up spending).
func TestRatchetRaisesSpending(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3,
		Ratchet: Ratchet{Trigger: 1.2, Step: 1000, Cap: 10000},
		Tax:     CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0.5, 0, 0})
	// Year 0: 100k < 120k, spend 4000. The +50% year lifts wealth to 144k:
	// years 1 and 2 each ratchet up by 1000.
	want := []float64{4000, 5000, 6000}
	for k, w := range want {
		if math.Abs(res.Spend[k]-w) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want %.0f", k, res.Spend[k], w)
		}
	}
}

// Cooldown spaces the ratchet steps: with a 2-year cooldown the second raise
// cannot follow the first immediately.
func TestRatchetCooldown(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3,
		Ratchet: Ratchet{Trigger: 1.2, Step: 1000, Cap: 10000, Cooldown: 2},
		Tax:     CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0.5, 0, 0})
	want := []float64{4000, 5000, 5000}
	for k, w := range want {
		if math.Abs(res.Spend[k]-w) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want %.0f", k, res.Spend[k], w)
		}
	}
}

// The cap bounds the ratcheted level, and MaxWR vetoes a raise when the
// current withdrawal rate is already above it (Ben's rule: only ratchet when
// the rate is comfortable).
func TestRatchetCapAndMaxWR(t *testing.T) {
	capped := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3,
		Ratchet: Ratchet{Trigger: 1.2, Step: 1000, Cap: 5000},
		Tax:     CTOFlatTax{Rate: 0}}
	res := capped.RunPath(scenario.Sequence{0.5, 0, 0})
	if math.Abs(res.Spend[2]-5000) > 1e-6 {
		t.Errorf("Spend[2] = %.0f, want 5000 (cap)", res.Spend[2])
	}

	vetoed := Plan{Capital: 100000, NeedAnnual: 4000, Years: 2,
		// After +50%, wr = 4000/144000 ≈ 2.8% which is above the 2% comfort
		// bound, so no raise happens despite wealth > trigger.
		Ratchet: Ratchet{Trigger: 1.2, Step: 1000, Cap: 10000, MaxWR: 0.02},
		Tax:     CTOFlatTax{Rate: 0}}
	res = vetoed.RunPath(scenario.Sequence{0.5, 0})
	if math.Abs(res.Spend[1]-4000) > 1e-6 {
		t.Errorf("Spend[1] = %.0f, want 4000 (MaxWR veto)", res.Spend[1])
	}
}

// SpendSchedule scales the base need year by year (health-cost drift, or a
// retirement smile), before cashflows are netted.
func TestSpendSchedule(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3,
		SpendSchedule: []float64{1, 1.1, 1.21}, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0})
	want := []float64{4000, 4400, 4840}
	for k, w := range want {
		if math.Abs(res.Spend[k]-w) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want %.0f", k, res.Spend[k], w)
		}
	}
}

// A schedule shorter than the horizon keeps a factor of 1 for missing years.
func TestSpendScheduleShort(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4000, Years: 3,
		SpendSchedule: []float64{1.5}, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0})
	want := []float64{6000, 4000, 4000}
	for k, w := range want {
		if math.Abs(res.Spend[k]-w) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want %.0f", k, res.Spend[k], w)
		}
	}
}

// The monthly kernel honours the ratchet (a yearly decision) and the spend
// schedule like the annual reference.
func TestRunPathMonthlyRatchetAndSchedule(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 4800, Years: 2,
		Ratchet:       Ratchet{Trigger: 1.2, Step: 1200, Cap: 12000},
		SpendSchedule: []float64{1, 1},
		Tax:           CTOFlatTax{Rate: 0}}
	seq := make(scenario.Sequence, 24)
	for i := 0; i < 12; i++ {
		seq[i] = 0.05 // ≈ +80% over year 0: well above the ratchet trigger
	}
	res := p.RunPathMonthly(seq)
	if math.Abs(res.Spend[0]-4800) > 1e-6 {
		t.Errorf("Spend[0] = %.0f, want 4800", res.Spend[0])
	}
	if math.Abs(res.Spend[1]-6000) > 1e-6 {
		t.Errorf("Spend[1] = %.0f, want 6000 (one ratchet step)", res.Spend[1])
	}
}
