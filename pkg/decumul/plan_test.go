package decumul

import (
	"math"
	"testing"
)

func TestCTOFlatTaxGrossUp(t *testing.T) {
	tax := CTOFlatTax{Rate: 0.30}
	// growth 200k, cost 100k -> gain fraction 0.5; net 10k.
	gross, newCost, paid := tax.GrossUp(10000, 200000, 100000)
	// effective rate = 0.30*0.5 = 0.15 -> gross = 10000/0.85.
	wantGross := 10000 / 0.85
	if math.Abs(gross-wantGross) > 1e-6 {
		t.Errorf("gross = %.2f, want %.2f", gross, wantGross)
	}
	if math.Abs(paid-(gross-10000)) > 1e-6 {
		t.Errorf("paid = %.2f, want %.2f", paid, gross-10000)
	}
	// cost reduced pro rata of the sale: cost * (1 - gross/growth).
	wantCost := 100000 * (1 - gross/200000)
	if math.Abs(newCost-wantCost) > 1e-6 {
		t.Errorf("newCost = %.2f, want %.2f", newCost, wantCost)
	}
}

// When the sale is capped at the available growth, the gain fraction (and thus
// the tax) must be computed on the gross actually sold, not implied from the
// requested net; otherwise the reported tax can even go negative.
func TestCTOFlatTaxGrossUpCapped(t *testing.T) {
	tax := CTOFlatTax{Rate: 0.5}
	// Want 70k net but only 60k of growth available, all at a 0.5 gain
	// fraction (cost 30k): the sale is capped at the 60k market value.
	gross, newCost, paid := tax.GrossUp(70000, 60000, 30000)
	if math.Abs(gross-60000) > 1e-6 {
		t.Errorf("gross = %.2f, want 60000 (capped at growth)", gross)
	}
	// Effective rate 0.5*0.5 = 0.25 on the 60k sold -> 15k tax.
	if math.Abs(paid-15000) > 1e-6 {
		t.Errorf("paid = %.2f, want 15000", paid)
	}
	if paid < 0 {
		t.Errorf("paid = %.2f, must never be negative", paid)
	}
	// Net actually delivered is gross - tax = 45k, below the 70k requested.
	if net := gross - paid; net >= 70000 {
		t.Errorf("net delivered = %.2f, must be below the 70000 requested", net)
	}
	if math.Abs(newCost-0) > 1e-6 {
		t.Errorf("newCost = %.2f, want 0 (whole sleeve sold)", newCost)
	}
}

func TestNeedAtAppliesCashflows(t *testing.T) {
	p := Plan{NeedAnnual: 48000, Cashflows: []Cashflow{{FromYear: 12, Annual: 18000}}}
	if got := p.needAt(0); math.Abs(got-48000) > 1e-9 {
		t.Errorf("needAt(0) = %.0f, want 48000", got)
	}
	if got := p.needAt(12); math.Abs(got-30000) > 1e-9 {
		t.Errorf("needAt(12) = %.0f, want 30000", got)
	}
}

// A bounded cashflow (side income with a ToYear) applies only over [FromYear,
// ToYear); a zero ToYear keeps running to the horizon.
func TestNeedAtBoundedCashflow(t *testing.T) {
	p := Plan{NeedAnnual: 48000, Cashflows: []Cashflow{{FromYear: 0, ToYear: 5, Annual: 12000}}}
	if got := p.needAt(0); math.Abs(got-36000) > 1e-9 {
		t.Errorf("needAt(0) = %.0f, want 36000", got)
	}
	if got := p.needAt(4); math.Abs(got-36000) > 1e-9 {
		t.Errorf("needAt(4) = %.0f, want 36000", got)
	}
	if got := p.needAt(5); math.Abs(got-48000) > 1e-9 {
		t.Errorf("needAt(5) = %.0f, want 48000 (side income ended)", got)
	}
}
