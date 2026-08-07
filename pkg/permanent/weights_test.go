package permanent

import (
	"math"
	"testing"
)

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestEquityWeightPoles(t *testing.T) {
	pm := DefaultParams()
	// Paradise (1,0) is the far corner: equity saturates at the cap (clamped 1).
	if w := (Regime{GrowthBreadth: 1, InflationBreadth: 0}).EquityWeight(pm); !almost(w, 1) {
		t.Fatalf("paradise equity = %v, want 1", w)
	}
	// Hell (0,1) is the pole: zero equity.
	if w := (Regime{GrowthBreadth: 0, InflationBreadth: 1}).EquityWeight(pm); !almost(w, 0) {
		t.Fatalf("hell equity = %v, want 0", w)
	}
	// Monotone: more growth breadth and less inflation breadth => more equity.
	mild := Regime{GrowthBreadth: 0.5, InflationBreadth: 0.5}.EquityWeight(pm)
	good := Regime{GrowthBreadth: 0.9, InflationBreadth: 0.1}.EquityWeight(pm)
	if !(good > mild && mild > 0) {
		t.Fatalf("not monotone: mild=%v good=%v", mild, good)
	}
}

func TestQuadraticDampsMoreThanLinear(t *testing.T) {
	r := Regime{GrowthBreadth: 0.6, InflationBreadth: 0.4}
	quad := DefaultParams()
	lin := DefaultParams()
	lin.Damping = 1
	// Away from paradise, the quadratic (1/d^2) shape cuts equity harder.
	if !(r.EquityWeight(quad) < r.EquityWeight(lin)) {
		t.Fatalf("quadratic %v should be below linear %v away from paradise",
			r.EquityWeight(quad), r.EquityWeight(lin))
	}
}

func TestAllocateSumsToOne(t *testing.T) {
	pm := DefaultParams()
	for _, r := range []Regime{
		{GrowthBreadth: 1, InflationBreadth: 0, Slope: 2, RealShort: 1},
		{GrowthBreadth: 0, InflationBreadth: 1, Slope: -1, RealShort: -3},
		{GrowthBreadth: 0.5, InflationBreadth: 0.5, Slope: 0.5, RealShort: 0.2},
	} {
		a := r.Allocate(pm)
		sum := a.Equity + a.Bonds + a.Cash + a.Gold
		if !almost(sum, 1) {
			t.Fatalf("weights sum to %v, want 1 (%+v)", sum, a)
		}
		if a.Equity < 0 || a.Bonds < 0 || a.Cash < 0 || a.Gold < 0 {
			t.Fatalf("negative weight: %+v", a)
		}
	}
}

func TestDefensiveSplitFollowsPoles(t *testing.T) {
	pm := DefaultParams()
	// Deeply negative real rates => gold dominates the defensive sleeve.
	gold := Regime{Slope: 0.5, RealShort: -2.5}.defensiveSplit(pm)
	if !(gold[2] > gold[0] && gold[2] > gold[1]) {
		t.Fatalf("gold should dominate near its pole: %v", gold)
	}
	// Steep curve, positive real short => bonds dominate.
	bonds := Regime{Slope: 2, RealShort: 1}.defensiveSplit(pm)
	if !(bonds[0] > bonds[1] && bonds[0] > bonds[2]) {
		t.Fatalf("bonds should dominate near their pole: %v", bonds)
	}
}
