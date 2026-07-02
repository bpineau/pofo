package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Within the annual allowance, an assurance-vie withdrawal is tax-free: the
// realised gain share of the sale stays below the allowance.
func TestAVTaxWithinAllowance(t *testing.T) {
	tax := AVTax{Rate: 0.247, Allowance: 9200}.NewPath()
	// growth 200k, cost 100k -> gain fraction 0.5; net 10k realises 5k of gain.
	gross, _, paid := tax.GrossUp(10000, 200000, 100000)
	if math.Abs(gross-10000) > 1e-6 {
		t.Errorf("gross = %.2f, want 10000 (no tax inside the allowance)", gross)
	}
	if paid != 0 {
		t.Errorf("paid = %.2f, want 0", paid)
	}
}

// Beyond the allowance, only the excess gain is taxed, and the gross-up solves
// the circular net/tax dependency exactly.
func TestAVTaxBeyondAllowance(t *testing.T) {
	tax := AVTax{Rate: 0.247, Allowance: 9200}.NewPath()
	gross, _, paid := tax.GrossUp(30000, 200000, 100000)
	// gross = (net - rate*allowance) / (1 - rate*gainFrac)
	wantGross := (30000 - 0.247*9200) / (1 - 0.247*0.5)
	if math.Abs(gross-wantGross) > 1e-6 {
		t.Errorf("gross = %.2f, want %.2f", gross, wantGross)
	}
	wantPaid := 0.247 * (wantGross*0.5 - 9200)
	if math.Abs(paid-wantPaid) > 1e-6 {
		t.Errorf("paid = %.2f, want %.2f", paid, wantPaid)
	}
	if math.Abs((gross-paid)-30000) > 1e-6 {
		t.Errorf("net delivered = %.2f, want 30000", gross-paid)
	}
}

// The allowance is consumed within a year and restored by NewYear.
func TestAVTaxAllowanceResetsYearly(t *testing.T) {
	tax := AVTax{Rate: 0.247, Allowance: 9200}.NewPath()
	// Two 10k withdrawals at gain fraction 0.5 realise 5k gain each: the first
	// fits the allowance, the second exceeds the 4.2k remainder and pays tax.
	_, _, paid1 := tax.GrossUp(10000, 200000, 100000)
	_, _, paid2 := tax.GrossUp(10000, 200000, 100000)
	if paid1 != 0 {
		t.Errorf("first withdrawal paid = %.2f, want 0", paid1)
	}
	if paid2 <= 0 {
		t.Errorf("second withdrawal paid = %.2f, want > 0 (allowance exhausted)", paid2)
	}
	tax.NewYear()
	_, _, paid3 := tax.GrossUp(10000, 200000, 100000)
	if paid3 != 0 {
		t.Errorf("after NewYear paid = %.2f, want 0 (allowance restored)", paid3)
	}
}

// Envelopes drain in slice order: the CTO empties first (paying its tax), then
// the tax-free pocket takes over, and the path ruins when both are gone.
func TestRunPathEnvelopeDrainOrder(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 30000, Years: 3,
		Envelopes: []Envelope{
			{Name: "CTO", Amount: 30000, GainFrac: 1, Tax: CTOFlatTax{Rate: 0.5}},
			{Name: "PEA", Amount: 70000, GainFrac: 1, Tax: CTOFlatTax{Rate: 0}},
		}}
	res := p.RunPath(scenario.Sequence{0, 0, 0})
	// Year 0: the CTO can only gross 30k (tax 15k, net 15k); the PEA tops up
	// the remaining 15k tax-free. Year 1: 30k tax-free from the PEA.
	// Year 2: only 25k left -> under-delivery, ruin.
	if math.Abs(res.TaxPaid-15000) > 1 {
		t.Errorf("TaxPaid = %.0f, want 15000 (CTO fully liquidated year 0)", res.TaxPaid)
	}
	if math.Abs(res.Spend[0]-30000) > 1 || math.Abs(res.Spend[1]-30000) > 1 {
		t.Errorf("Spend[0..1] = %.0f, %.0f, want 30000 each", res.Spend[0], res.Spend[1])
	}
	if !res.Ruined || res.RuinYear != 2 {
		t.Errorf("Ruined=%v RuinYear=%d, want ruin in year 2", res.Ruined, res.RuinYear)
	}
}

// Envelope amounts are relative: the growth sleeve is split pro-rata, so
// doubling the capital doubles every pocket (and the tax paid on an all-gain
// CTO liquidation).
func TestEnvelopesScaleWithCapital(t *testing.T) {
	p := Plan{Capital: 200000, NeedAnnual: 30000, Years: 1,
		Envelopes: []Envelope{
			{Name: "CTO", Amount: 30000, GainFrac: 1, Tax: CTOFlatTax{Rate: 0.5}},
			{Name: "PEA", Amount: 70000, GainFrac: 1, Tax: CTOFlatTax{Rate: 0}},
		}}
	res := p.RunPath(scenario.Sequence{0})
	// Pockets scale to 60k/140k. The 30k net need grosses 60k from the CTO
	// (exactly its scaled size) at a 50% effective rate: 30k of tax.
	if math.Abs(res.TaxPaid-30000) > 1 {
		t.Errorf("TaxPaid = %.0f, want 30000 (pockets scaled pro-rata)", res.TaxPaid)
	}
}

// A nil Envelopes slice must behave exactly like a single explicit envelope on
// the plan's Tax with no embedded gain: the historical single-sleeve model.
func TestEnvelopesNilParity(t *testing.T) {
	seq := scenario.Sequence{0.3, -0.2, 0.1, 0, 0.05}
	base := Plan{Capital: 500000, NeedAnnual: 30000, Years: 5,
		Buffer: BufferSleeve{Years: 2}, Tax: CTOFlatTax{Rate: 0.30}}
	env := base
	env.Tax = nil
	env.Envelopes = []Envelope{{Name: "CTO", Amount: 1, Tax: CTOFlatTax{Rate: 0.30}}}

	a, b := base.RunPath(seq), env.RunPath(seq)
	if math.Abs(a.TaxPaid-b.TaxPaid) > 1e-6 || math.Abs(a.Withdrawn-b.Withdrawn) > 1e-6 {
		t.Errorf("single envelope diverges from the legacy sleeve: tax %.2f vs %.2f, withdrawn %.2f vs %.2f",
			a.TaxPaid, b.TaxPaid, a.Withdrawn, b.Withdrawn)
	}
	for k := range a.Wealth {
		if math.Abs(a.Wealth[k]-b.Wealth[k]) > 1e-6 {
			t.Fatalf("Wealth[%d] = %.2f vs %.2f", k, a.Wealth[k], b.Wealth[k])
		}
	}
}

// The AV allowance works inside a full path run and resets every year, with
// per-path state isolated across paths (the template Envelope is not mutated).
func TestRunPathAVAllowanceYearly(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 10000, Years: 2,
		Envelopes: []Envelope{
			{Name: "AV", Amount: 1, GainFrac: 0.5, Tax: AVTax{Rate: 0.247, Allowance: 9200}},
		}}
	res := p.RunPath(scenario.Sequence{0, 0})
	// Each year realises 5k of gain, inside the 9.2k allowance: no tax at all.
	if res.TaxPaid != 0 {
		t.Errorf("TaxPaid = %.2f, want 0 (allowance covers each year)", res.TaxPaid)
	}
	// Run again: the template must be untouched (fresh allowance per path).
	res2 := p.RunPath(scenario.Sequence{0, 0})
	if res2.TaxPaid != 0 {
		t.Errorf("second path TaxPaid = %.2f, want 0 (per-path state)", res2.TaxPaid)
	}

	noAllowance := p
	noAllowance.Envelopes = []Envelope{
		{Name: "AV", Amount: 1, GainFrac: 0.5, Tax: AVTax{Rate: 0.247}}}
	if got := noAllowance.RunPath(scenario.Sequence{0, 0}); got.TaxPaid <= 0 {
		t.Errorf("TaxPaid = %.2f, want > 0 without an allowance", got.TaxPaid)
	}
}

// The monthly kernel drains envelopes in order too, and resets the AV
// allowance at year boundaries only.
func TestRunPathMonthlyEnvelopes(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 30000, Years: 3,
		Envelopes: []Envelope{
			{Name: "CTO", Amount: 30000, GainFrac: 1, Tax: CTOFlatTax{Rate: 0.5}},
			{Name: "PEA", Amount: 70000, GainFrac: 1, Tax: CTOFlatTax{Rate: 0}},
		}}
	res := p.RunPathMonthly(zeros(36))
	if math.Abs(res.TaxPaid-15000) > 1 {
		t.Errorf("TaxPaid = %.0f, want 15000", res.TaxPaid)
	}
	if !res.Ruined {
		t.Errorf("expected ruin in year 2")
	}
}
