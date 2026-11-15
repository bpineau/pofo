package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestHouseholdSpendingPensionTiming(t *testing.T) {
	for _, start := range []int{0, 1, 3} {
		p := Plan{Capital: 1000, NeedAnnual: 100, Years: 3, Cashflows: []Cashflow{{FromYear: start, Annual: 50}}}
		e := p.SimulateOn(Draws{Returns: []scenario.Sequence{{0, 0, 0}, {0, 0, 0}}}, 1)
		bands, err := p.HouseholdSpendBands(e, []float64{.05, .5, .95})
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range bands {
			for year, v := range b {
				if v != 100 {
					t.Errorf("pension starts %d year %d household %v", start, year, v)
				}
			}
		}
		cv, err := p.HouseholdSpendCV(e)
		if err != nil || cv != 0 {
			t.Fatalf("pension start %d household CV %v %v", start, cv, err)
		}
		if start == 1 && e.SpendCV() == 0 {
			t.Fatal("portfolio withdrawals should change at pension start")
		}
	}
}
func TestHouseholdSpendingUndefined(t *testing.T) {
	p := Plan{Years: 2}
	if cv, err := p.HouseholdSpendCV(Ensemble{Years: 2}); err != nil || !math.IsNaN(cv) {
		t.Fatalf("empty CV %v %v", cv, err)
	}
	p.Lifetime = &Lifetime{}
	if _, err := p.HouseholdSpendBands(Ensemble{Years: 2}, []float64{.5}); err == nil {
		t.Fatal("lifetime silently treated as fixed income")
	}
}
