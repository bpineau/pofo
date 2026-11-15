package fire

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

func TestBuildPlanPensionAndExplicitZero(t *testing.T) {
	for _, start := range []int{0, 1, 10} {
		s := PlanSpec{Capital: 1000, AnnualSpending: 100, Years: 10, Pension: &Pension{AnnualAmount: 100, StartYear: start}}
		p, err := BuildPlan(s)
		if err != nil {
			t.Fatal(err)
		}
		if p.Buffer.Years != 0 || p.Inflation != 0 || p.Tax.(decumul.CTOFlatTax).Rate != 0 {
			t.Fatalf("explicit zero defaulted: %+v", p)
		}
		path := p.RunPath(make(scenario.Sequence, 10), decumul.Lives{})
		want := 1000 - float64(start)*100
		if path.Wealth[10] != want {
			t.Errorf("pension start %d: terminal %v want %v", start, path.Wealth[10], want)
		}
	}
}

func TestBuildPlanZeroGainTaxesNominalGains(t *testing.T) {
	s := PlanSpec{Capital: 500000, AnnualSpending: 20000, Years: 10, TaxRate: .314, Inflation: .02}
	p, err := BuildPlan(s)
	if err != nil {
		t.Fatal(err)
	}
	flat := make(scenario.Sequence, 10)
	got := p.RunPath(flat, decumul.Lives{}).TaxPaid
	if math.Abs(got-5461) > 1 {
		t.Fatalf("zero initial gain lost nominal erosion: tax %v", got)
	}
	s.Inflation = 0
	p, err = BuildPlan(s)
	if err != nil {
		t.Fatal(err)
	}
	if got = p.RunPath(flat, decumul.Lives{}).TaxPaid; got != 0 {
		t.Fatalf("flat real gains tax %v", got)
	}
	s.InitialGain = .5
	p, err = BuildPlan(s)
	if err != nil {
		t.Fatal(err)
	}
	if got = p.RunPath(flat, decumul.Lives{}).TaxPaid; got <= 0 {
		t.Fatal("initial gain ignored")
	}
}

func TestBuildPlanRebuildsBandsAndCapsBuffer(t *testing.T) {
	s := Defaults()
	s.Capital = 100
	s.AnnualSpending = 50
	s.Years = 2
	s.Policy = GuytonKlinger
	p, err := BuildPlan(s)
	if err != nil {
		t.Fatal(err)
	}
	if p.Guard.Upper != .6 || p.Guard.Lower != .4 || p.Envelopes[0].Amount != 0 {
		t.Fatalf("bands or buffer: %+v", p)
	}
	s.Capital = 200
	q, err := BuildPlan(s)
	if err != nil {
		t.Fatal(err)
	}
	if q.Guard.Upper != .3 || q.Guard.Lower != .2 || q.Envelopes[0].Amount != 50 {
		t.Fatalf("candidate settings not rebuilt: %+v", q)
	}
}
