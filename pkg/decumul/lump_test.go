package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// flat is a zero-return sequence: on it a plan's wealth is pure arithmetic.
func flat(n int) scenario.Sequence { return make(scenario.Sequence, n) }

// A lump coming in lands in the portfolio, where a one-year income of the same
// size would mostly be received and lost: the very trap the type exists for.
func TestLumpInIsInvested(t *testing.T) {
	base := Plan{Capital: 500_000, NeedAnnual: 40_000, Years: 20}
	in := base
	in.Lumps = []Lump{{Year: 5, Amount: 200_000}}
	asIncome := base
	asIncome.Cashflows = []Cashflow{{FromYear: 5, ToYear: 6, Annual: 200_000}}
	for _, monthly := range []bool{false, true} {
		b, i, c := base, in, asIncome
		b.Monthly, i.Monthly, c.Monthly = monthly, monthly, monthly
		seq := flat(20)
		if monthly {
			seq = flat(240)
		}
		pb, pi, pc := b.runPath(seq, Lives{}, nil), i.runPath(seq, Lives{}, nil), c.runPath(seq, Lives{}, nil)
		// From year 5 on, the lump is exactly 200 000 more wealth.
		for k := 6; k <= 12; k++ {
			if got := pi.Wealth[k] - pb.Wealth[k]; math.Abs(got-200_000) > 1e-6 {
				t.Errorf("monthly=%v year %d: the lump added %.2f, want 200000", monthly, k, got)
			}
		}
		// As income it only covers year 5's need: 40 000 of it is used.
		if got := pc.Wealth[12] - pb.Wealth[12]; math.Abs(got-40_000) > 1e-6 {
			t.Errorf("monthly=%v: a one-year income kept %.2f, want the 40000 of need it covered", monthly, got)
		}
	}
}

// A lump going out is a sale like any other, taxed on its gain share, and one
// the portfolio cannot pay is that year's ruin.
func TestLumpOutIsATaxedSale(t *testing.T) {
	p := Plan{Capital: 500_000, NeedAnnual: 20_000, Years: 10,
		Envelopes: []Envelope{{Name: "CTO", Amount: 1, GainFrac: 0.5, Tax: CTOFlatTax{Rate: 0.3}}}}
	base := p.RunPath(flat(10), Lives{})
	p.Lumps = []Lump{{Year: 3, Amount: -100_000}}
	out := p.RunPath(flat(10), Lives{})
	if out.TaxPaid <= base.TaxPaid {
		t.Errorf("paying a lump out cost no tax: %.0f vs %.0f", out.TaxPaid, base.TaxPaid)
	}
	if out.Ruined {
		t.Errorf("a payable lump ruined the plan in year %d", out.RuinYear)
	}
	p.Lumps = []Lump{{Year: 3, Amount: -10_000_000}}
	broke := p.RunPath(flat(10), Lives{})
	if !broke.Ruined || broke.RuinYear != 3 {
		t.Errorf("an unpayable lump: ruined %v in year %d, want year 3", broke.Ruined, broke.RuinYear)
	}
}

// The amortization rule counts the inheritance still to come, as it counts a
// pension: it spends more today on it, and a dated expense makes it spend less.
func TestLumpsEnterTheAmortizationForecast(t *testing.T) {
	p := Plan{Capital: 500_000, NeedAnnual: 20_000, Years: 30, Amortize: true, AmortReturn: 0.03}
	spend0 := func(l []Lump) float64 {
		q := p
		q.Lumps = l
		return q.RunPath(flat(30), Lives{}).Spend[0]
	}
	none, in, out := spend0(nil), spend0([]Lump{{Year: 10, Amount: 200_000}}), spend0([]Lump{{Year: 10, Amount: -100_000}})
	if !(in > none && none > out) {
		t.Errorf("year-0 ABW spending: inheritance %.0f, none %.0f, expense %.0f", in, none, out)
	}
}

// The ledger books the lump in its own column and never reads money coming in
// as a negative sale.
func TestLedgerBooksLumps(t *testing.T) {
	p := Plan{Capital: 500_000, NeedAnnual: 30_000, Years: 12,
		Lumps: []Lump{{Year: 2, Amount: 150_000}, {Year: 6, Amount: -80_000}}}
	_, led := p.RunLedger(flat(12), Lives{})
	if led[2].Lump != 150_000 || led[6].Lump != -80_000 || led[3].Lump != 0 {
		t.Errorf("lumps booked %v %v %v", led[2].Lump, led[6].Lump, led[3].Lump)
	}
	for _, y := range led {
		if y.Sold < -1e-6 {
			t.Errorf("year %d sold %.2f", y.Year, y.Sold)
		}
	}
	if math.Abs(led[6].Sold-110_000) > 1e-6 {
		t.Errorf("year 6 sold %.2f, want the 80000 lump plus the 30000 need", led[6].Sold)
	}
}
