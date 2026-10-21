package decumul

import (
	"math"
	"slices"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// ledgerPlan is a plan that exercises every ledger column: a cash buffer drawn
// in drawdowns and refilled, a taxed sleeve with an embedded gain, a pension
// starting mid-way and a flex cut.
func ledgerPlan() Plan {
	return Plan{
		Capital: 1_000_000, NeedAnnual: 48_000, Years: 35,
		Buffer:    BufferSleeve{Years: 2, RealReturn: -0.005},
		Flex:      FlexRule{Threshold: 0.20, Cut: 0.10},
		Cashflows: []Cashflow{{FromYear: 10, Annual: 15_000, Drift: -0.005}},
		Envelopes: []Envelope{{Name: "CTO", Amount: 904_000, GainFrac: 0.5, Tax: CTOFlatTax{Rate: 0.314}}},
		Inflation: 0.02,
		Source:    scenario.ParametricSource{Mu: 0.035, Sigma: 0.14, Df: 5, Periods: 35},
	}
}

// The ledger is the kernel's own reading, not a re-simulation: the path it
// returns is RunPath's bit for bit, and every row agrees with the path's
// series and totals.
func TestRunLedgerMatchesRunPath(t *testing.T) {
	p := ledgerPlan()
	d := p.Draw(300, 4, 7)
	sawBuffer, sawCut, sawRuin := false, false, false
	for i, seq := range d.Returns {
		want := p.RunPath(seq, Lives{})
		got, led := p.RunLedger(seq, Lives{})
		if !slices.Equal(got.Wealth, want.Wealth) || !slices.Equal(got.Spend, want.Spend) ||
			got.TaxPaid != want.TaxPaid || got.Ruined != want.Ruined || got.RuinYear != want.RuinYear {
			t.Fatalf("path %d: the ledger run is not RunPath's", i)
		}
		if len(led) != p.Years {
			t.Fatalf("path %d: %d ledger rows, want %d", i, len(led), p.Years)
		}
		tax := 0.0
		for k, y := range led {
			if y.Year != k || y.Spent != got.Spend[k] {
				t.Fatalf("path %d year %d: row %+v vs spend %v", i, k, y, got.Spend[k])
			}
			if !y.Broke && (y.Close != got.Wealth[k+1] || y.Open != got.Wealth[k]) {
				t.Fatalf("path %d year %d: open/close %v/%v vs wealth %v/%v", i, k, y.Open, y.Close, got.Wealth[k], got.Wealth[k+1])
			}
			if y.Income != p.income(k, p.life(Lives{})) {
				t.Fatalf("path %d year %d: income %v", i, k, y.Income)
			}
			if y.FromBuffer > y.Spent+1e-9 || y.Tax < 0 || y.Sold < y.Tax-1e-9 {
				t.Fatalf("path %d year %d: incoherent row %+v", i, k, y)
			}
			tax += y.Tax
			sawBuffer = sawBuffer || y.FromBuffer > 0
			sawCut = sawCut || y.Cut
			sawRuin = sawRuin || y.Broke
		}
		if math.Abs(tax-got.TaxPaid) > 1e-6*max(1, got.TaxPaid) {
			t.Fatalf("path %d: ledger tax %v vs TaxPaid %v", i, tax, got.TaxPaid)
		}
		if got.Ruined && !led[len(led)-1].Broke {
			t.Fatalf("path %d ruined in year %d but its last row is not broke", i, got.RuinYear)
		}
	}
	if !sawBuffer || !sawCut || !sawRuin {
		t.Errorf("300 paths never exercised a column: buffer %v, cut %v, ruin %v", sawBuffer, sawCut, sawRuin)
	}
}

// DrawPath replays one path of an ensemble alone, lifespans included, and a
// fan's sample names the path it shows.
func TestDrawPathReplaysTheEnsemble(t *testing.T) {
	p := couplePlan(60)
	d := p.Draw(500, 3, 11)
	for _, i := range []int{0, 1, 137, 499} {
		seq, lives := p.DrawPath(11, i)
		if !slices.Equal(seq, d.Returns[i]) || lives != d.Lives[i] {
			t.Errorf("path %d: DrawPath differs from Draw", i)
		}
	}
	e := ledgerPlan().Simulate(500, 3, 11)
	for _, s := range e.Fan([]float64{0.5}, 8).Samples {
		if !slices.Equal(s.Wealth, e.Paths[s.Index].Wealth) {
			t.Errorf("sample index %d does not name the path it shows", s.Index)
		}
	}
}
