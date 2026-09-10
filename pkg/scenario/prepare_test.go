package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

// testPanel is a small two-asset panel with distinct rows, so a wrong set of
// weights would show up in the combined history.
func testPanel() Panel {
	a := make([]float64, 60)
	b := make([]float64, 60)
	for t := range a {
		a[t] = 0.01 + 0.02*math.Sin(float64(t)/5)
		b[t] = -0.005 + 0.01*math.Cos(float64(t)/3)
	}
	return Panel{Returns: [][]float64{a, b}, Weights: []float64{0.6, 0.4}}
}

// TestPrepareDrawsIdenticalPaths is the contract Prepare lives by: hoisting a
// source's setup must not move a single drawn return, since a shared FIRE URL
// has to reproduce byte for byte.
func TestPrepareDrawsIdenticalPaths(t *testing.T) {
	panel := testPanel()
	srcs := map[string]Source{
		"block":      BlockBootstrap{Panel: panel, BlockLen: 12, Periods: 40},
		"stationary": StationaryBootstrap{Panel: panel, MeanBlock: 8, Periods: 40},
		"cohorts":    HistoricalCohorts{Panel: panel, Periods: 24},
		"weighted":   BlockBootstrap{Panel: panel, Weights: []float64{0.2, 0.8}, BlockLen: 6, Periods: 40},
		"compounded": Compounded{Inner: StationaryBootstrap{Panel: panel, MeanBlock: 8, Periods: 36}, Group: 12},
		"parametric": ParametricSource{Mu: 0.05, Sigma: 0.15, Df: 5, Periods: 40},
	}
	for name, src := range srcs {
		prepared := Prepare(src)
		if prepared.Len() != src.Len() {
			t.Errorf("%s: prepared Len %d, want %d", name, prepared.Len(), src.Len())
		}
		plain := rand.New(rand.NewPCG(3, 4))
		hoisted := rand.New(rand.NewPCG(3, 4))
		for i := 0; i < 5; i++ {
			want, got := src.Draw(plain), prepared.Draw(hoisted)
			if len(got) != len(want) {
				t.Fatalf("%s: path %d has %d periods, want %d", name, i, len(got), len(want))
			}
			for k := range want {
				if got[k] != want[k] {
					t.Fatalf("%s: path %d period %d = %v, want %v (bit-for-bit)", name, i, k, got[k], want[k])
				}
			}
		}
	}
}

// TestPrepareLeavesPlainSourcesAlone: a source with no setup to hoist comes
// back as it went in, so a driver can always call Prepare.
func TestPrepareLeavesPlainSourcesAlone(t *testing.T) {
	src := ParametricSource{Mu: 0.04, Sigma: 0.12, Df: 4, Periods: 10}
	if got, ok := Prepare(src).(ParametricSource); !ok || got != src {
		t.Errorf("Prepare(%v) = %v, want it unchanged", src, Prepare(src))
	}
}

// TestPrepareIsIdempotent: preparing an already prepared source is a no-op,
// which is what makes the wrappers (Compounded) safe to prepare twice.
func TestPrepareIsIdempotent(t *testing.T) {
	src := Prepare(HistoricalCohorts{Panel: testPanel(), Periods: 24})
	once, twice := rand.New(rand.NewPCG(9, 9)), rand.New(rand.NewPCG(9, 9))
	a, b := src.Draw(once), Prepare(src).Draw(twice)
	for k := range a {
		if a[k] != b[k] {
			t.Fatalf("period %d = %v, want %v", k, b[k], a[k])
		}
	}
}
