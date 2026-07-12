package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
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

// A guardrails floor bounds the cut spiral: in a relentless bear, spending
// steps down 10% a year but never below the floor.
func TestGuardrailsFloor(t *testing.T) {
	crash := make(scenario.Sequence, 20)
	for i := range crash {
		crash[i] = -0.15 // a persistent bear: the upper guardrail stays breached
	}
	p := Plan{Capital: 1e6, NeedAnnual: 40000, Years: 20,
		Guard: Guardrails{Upper: 0.048, Lower: 0.032, Cut: 0.10, Raise: 0.10, Floor: 30000}}
	res := p.RunPath(crash)
	minSpend := res.Spend[0]
	for k, s := range res.Spend {
		if res.Ruined && k >= res.RuinYear {
			break // after depletion nothing is delivered
		}
		if s < minSpend {
			minSpend = s
		}
	}
	if minSpend < 30000-1 {
		t.Errorf("spending fell below the floor: %.0f", minSpend)
	}
	// Same plan without a floor must go materially below it.
	p.Guard.Floor = 0
	res = p.RunPath(crash)
	below := false
	for k, s := range res.Spend {
		if res.Ruined && k >= res.RuinYear {
			break
		}
		if s < 25000 {
			below = true
		}
	}
	if !below {
		t.Errorf("floorless guardrails should cut below 25k in a relentless bear")
	}
}

// The monthly kernel evaluates guardrails monthly at the pace-preserving
// step: after one fully-breached year, the level lands near the annual
// kernel's single -10% (not -72%), without waiting for an anniversary.
func TestGuardrailsMonthlyStepped(t *testing.T) {
	months := make(scenario.Sequence, 240)
	for i := range months {
		months[i] = -0.02 // ~-21%/yr: breached from the start, every month
	}
	p := Plan{Capital: 1e6, NeedAnnual: 40000, Years: 20, Monthly: true,
		Guard: Guardrails{Upper: 0.048, Lower: 0.032, Cut: 0.10, Raise: 0.10}}
	res := p.RunPathMonthly(months)
	y0 := res.Spend[0]
	if y0 > 40000+1 || y0 < 36000 {
		t.Errorf("first-year delivered spending %.0f: monthly steps should land between the full level and one annual cut", y0)
	}
	if res.Spend[1] >= y0 {
		t.Errorf("second year should keep stepping down in a persistent bear (%.0f then %.0f)", y0, res.Spend[1])
	}
	// The stepped rule preserves the annual intensity exactly.
	g := Guardrails{Cut: 0.10, Raise: 0.10, Upper: 1, Lower: 0.5}.stepped(12)
	if got := math.Pow(1-g.Cut, 12); math.Abs(got-0.9) > 1e-9 {
		t.Errorf("12 monthly cuts compound to %.4f, want 0.90", got)
	}
	if got := math.Pow(1+g.Raise, 12); math.Abs(got-1.1) > 1e-9 {
		t.Errorf("12 monthly raises compound to %.4f, want 1.10", got)
	}
}
