package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// With Inflation set, a flat real path still pays tax: the basis erodes by
// 1/(1+pi) each year, so year 0 sells at a gain fraction of pi/(1+pi) and
// year 1 at 1-1/(1+pi)^2. Without it, the same path pays nothing.
func TestRunPathInflationErodesBasis(t *testing.T) {
	const capital, need, rate, pi = 100_000.0, 10_000.0, 0.5, 0.02
	p := Plan{Capital: capital, NeedAnnual: need, Years: 2, Tax: CTOFlatTax{Rate: rate}}
	flat := scenario.Sequence{0, 0}
	if res := p.RunPath(flat, Lives{}); res.TaxPaid != 0 {
		t.Fatalf("real-gain taxation on a flat path paid %.2f of tax, want 0", res.TaxPaid)
	}
	p.Inflation = pi
	res := p.RunPath(flat, Lives{})

	// Year 0: the year's sale happens before the first erosion (the kernel
	// grows then erodes after the withdrawal), so it is untaxed; year 1 sees
	// a basis of (capital-10k)/(1+pi) against a value of capital-10k.
	g1 := 1 - 1/(1+pi)
	want := (need / (1 - rate*g1)) * rate * g1
	if math.Abs(res.TaxPaid-want) > 0.01 {
		t.Errorf("TaxPaid = %.2f, want %.2f (erosion 1/(1+%.2f) on the year-1 sale)", res.TaxPaid, want, pi)
	}
	if res.Spend[0] != need || math.Abs(res.Spend[1]-need) > 1e-6 {
		t.Errorf("Spend = %v, want %.0f delivered in full both years", res.Spend, need)
	}
}

// basisErosion: 1 without inflation (the kernels then skip the pass), the
// annual factor at one period a year, its twelfth root at twelve, so twelve
// monthly erosions equal one annual one.
func TestBasisErosionFactors(t *testing.T) {
	p := Plan{}
	if p.basisErosion(1) != 1 || p.basisErosion(12) != 1 {
		t.Fatalf("zero inflation must erode nothing")
	}
	p.Inflation = -0.01
	if p.basisErosion(1) != 1 {
		t.Fatalf("deflation must not inflate a basis")
	}
	p.Inflation = 0.03
	if y := p.basisErosion(1); math.Abs(y-1/1.03) > 1e-12 {
		t.Errorf("annual factor %.6f, want %.6f", y, 1/1.03)
	}
	if m := math.Pow(p.basisErosion(12), 12); math.Abs(m-1/1.03) > 1e-12 {
		t.Errorf("twelve monthly erosions compound to %.6f, want %.6f", m, 1/1.03)
	}
}

// The monthly kernel erodes at the twelfth root, twelve times a year. A year
// of sales spread over the months therefore pays more than the annual
// kernel's single sale at the start of that year (3 erosions at year 3) and
// less than a sale one year later (4 erosions): the monthly bill must sit
// strictly between the two annual ones.
func TestRunPathMonthlyInflationBracket(t *testing.T) {
	// Side income covers years 0-2, so the first sale is year 3's.
	sale := func(years, sellYear int, monthly bool) float64 {
		p := Plan{Capital: 300_000, NeedAnnual: 12_000, Years: years, Inflation: 0.025,
			Tax:       CTOFlatTax{Rate: 0.3},
			Cashflows: []Cashflow{{FromYear: 0, ToYear: sellYear, Annual: 12_000}}}
		if monthly {
			p.Monthly = true
			return p.RunPathMonthly(make(scenario.Sequence, 12*years), Lives{}).TaxPaid
		}
		return p.RunPath(make(scenario.Sequence, years), Lives{}).TaxPaid
	}
	lo, hi, m := sale(4, 3, false), sale(5, 4, false), sale(4, 3, true)
	if !(lo > 0 && lo < m && m < hi) {
		t.Errorf("monthly tax %.2f must sit between the annual sales at year 3 (%.2f) and year 4 (%.2f)", m, lo, hi)
	}
}

// Every pocket erodes, whatever its tax: an envelope book under inflation
// pays more from its sheltered pocket too, and a zero-inflation book is
// bit-identical to the plan before the field existed.
func TestEnvelopesInflationEveryPocket(t *testing.T) {
	seq := scenario.Sequence{0.05, -0.1, 0.02, 0, 0.03}
	p := Plan{Capital: 400_000, NeedAnnual: 30_000, Years: 5,
		Envelopes: []Envelope{
			{Name: "CTO", Amount: 1, GainFrac: 0, Tax: CTOFlatTax{Rate: 0.314}},
			{Name: "PEA", Amount: 1, GainFrac: 0, Tax: CTOFlatTax{Rate: 0.186}},
		}}
	none := p.RunPath(seq, Lives{})
	p.Inflation = 0.02
	some := p.RunPath(seq, Lives{})
	if some.TaxPaid <= none.TaxPaid {
		t.Errorf("inflation must raise the tax bill: %.2f vs %.2f", some.TaxPaid, none.TaxPaid)
	}
	p.Inflation = 0
	again := p.RunPath(seq, Lives{})
	if again.TaxPaid != none.TaxPaid || again.Withdrawn != none.Withdrawn {
		t.Errorf("zero inflation is not the legacy plan: tax %.2f vs %.2f", again.TaxPaid, none.TaxPaid)
	}
}
