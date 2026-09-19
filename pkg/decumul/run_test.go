package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func ptr(f float64) *float64 { return &f }

// An explicit zero DrawThreshold must mean "always tap the buffer first", not
// be read as "unset" and replaced by the 0.10 default. At a shallow ~9%
// drawdown the default sells from growth (taxed) while the explicit zero drains
// the tax-free buffer, so the latter pays strictly less tax.
func TestBufferDrawThresholdZeroHonored(t *testing.T) {
	base := Plan{Capital: 100000, NeedAnnual: 5000, Years: 3,
		Buffer: BufferSleeve{Years: 2}, Tax: CTOFlatTax{Rate: 0.5}}
	seq := scenario.Sequence{0.5, -0.05, 0}

	def := base.RunPath(seq, Lives{})
	eager := base
	eager.Buffer.DrawThreshold = ptr(0)
	got := eager.RunPath(seq, Lives{})

	if !(got.TaxPaid < def.TaxPaid) {
		t.Errorf("explicit DrawThreshold 0 should tap the buffer at a shallow drawdown and pay less tax: got=%.0f default=%.0f", got.TaxPaid, def.TaxPaid)
	}
}

// An explicit zero RefillCap must mean "never refill", not be read as "unset"
// and replaced by the 0.50 default. After a deep year drains the buffer, the
// default tops it back up from growth (a taxed sale) while the explicit zero
// skips it, so the latter pays strictly less tax.
func TestBufferRefillCapZeroHonored(t *testing.T) {
	base := Plan{Capital: 100000, NeedAnnual: 5000, Years: 4,
		Buffer: BufferSleeve{Years: 2}, Tax: CTOFlatTax{Rate: 0.5}}
	seq := scenario.Sequence{0.3, -0.2, 0.25, 0}

	on := base.RunPath(seq, Lives{})
	off := base
	off.Buffer.RefillCap = ptr(0)
	got := off.RunPath(seq, Lives{})

	if !(got.TaxPaid < on.TaxPaid) {
		t.Errorf("explicit RefillCap 0 should skip the refill and pay less tax: got=%.0f default=%.0f", got.TaxPaid, on.TaxPaid)
	}
}

// Guyton-Klinger guardrails cut spending when a market drop pushes the
// withdrawal rate above the upper guardrail, so the path withdraws less than
// the fixed-spending plan on the same declining sequence.
func TestRunPathGuardrailsCutSpending(t *testing.T) {
	base := Plan{Capital: 100000, NeedAnnual: 5000, Years: 3, Tax: CTOFlatTax{Rate: 0}}
	seq := scenario.Sequence{-0.3, -0.3, 0}

	fixed := base.RunPath(seq, Lives{})
	guarded := base
	guarded.Guard = Guardrails{Upper: 0.06, Lower: 0.03, Cut: 0.10, Raise: 0.10}
	got := guarded.RunPath(seq, Lives{})

	if !(got.Withdrawn < fixed.Withdrawn) {
		t.Errorf("guardrails should cut spending after the drop: guarded=%.0f fixed=%.0f", got.Withdrawn, fixed.Withdrawn)
	}
}

// A glidepath buffer stops refilling after the sequence-risk window: with
// RefillStopYear set, the late-path refill is skipped, so less tax is paid than
// when the buffer keeps topping up.
func TestBufferRefillStopYearHonored(t *testing.T) {
	base := Plan{Capital: 100000, NeedAnnual: 5000, Years: 4,
		Buffer: BufferSleeve{Years: 2}, Tax: CTOFlatTax{Rate: 0.5}}
	seq := scenario.Sequence{0.3, -0.2, 0.25, 0}

	on := base.RunPath(seq, Lives{}) // refills throughout
	stop := base
	stop.Buffer.RefillStopYear = 3 // no refill from year 3 onward
	got := stop.RunPath(seq, Lives{})

	if !(got.TaxPaid < on.TaxPaid) {
		t.Errorf("RefillStopYear should skip the year-3 refill and pay less tax: got=%.0f default=%.0f", got.TaxPaid, on.TaxPaid)
	}
}

// With zero returns, no tax and no pension, capital depletes by need/year.
func TestRunPathDepletion(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 25000, Years: 5, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0, 0, 0}, Lives{})
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
	res := p.RunPath(scenario.Sequence{1.0, 0}, Lives{})
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
	res := p.RunPath(scenario.Sequence{0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05}, Lives{})
	if res.Ruined {
		t.Errorf("did not expect ruin")
	}
	if res.Wealth[10] <= 0 {
		t.Errorf("final wealth = %.0f, want > 0", res.Wealth[10])
	}
}

// A year that delivers its whole need and leaves the portfolio at exactly zero
// is a FUNDED year, not the ruin year: the failure belongs to the first year
// that goes unfunded. Four withdrawals of 50k out of 200k at a zero real
// return pay years 0 to 3 in full, so ruin latches at year 4.
func TestRuinYearIsTheFirstUnfundedYear(t *testing.T) {
	p := Plan{Capital: 200000, NeedAnnual: 50000, Years: 6, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0, 0, 0, 0}, Lives{})
	if res.RuinYear != 4 {
		t.Errorf("RuinYear = %d, want 4", res.RuinYear)
	}
	for k := 0; k < 4; k++ {
		if math.Abs(res.Spend[k]-50000) > 1e-6 {
			t.Errorf("Spend[%d] = %.0f, want 50000", k, res.Spend[k])
		}
	}
}

// ...and a plan whose LAST withdrawal empties the pot exactly never failed at
// all: every year of the horizon was paid in full, so the path is not ruined.
// Latching ruin on an emptied-but-funded year used to report this plan, and
// every plan that lands on zero at its horizon, as a failure.
func TestExactlyExhaustedAtTheHorizonIsNotRuin(t *testing.T) {
	for _, monthly := range []bool{false, true} {
		p := Plan{Capital: 200000, NeedAnnual: 50000, Years: 4, Tax: CTOFlatTax{Rate: 0},
			Monthly: monthly}
		seq := make(scenario.Sequence, 4)
		if monthly {
			seq = make(scenario.Sequence, 48)
		}
		res := p.runPath(seq, Lives{}, nil)
		if res.Ruined {
			t.Errorf("monthly=%v: every year was funded in full, yet the path reads as ruined (year %d)",
				monthly, res.RuinYear)
		}
		if math.Abs(res.Withdrawn-200000) > 1e-6 {
			t.Errorf("monthly=%v: withdrew %.2f, want 200000", monthly, res.Withdrawn)
		}
	}
}
