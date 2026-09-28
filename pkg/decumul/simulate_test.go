package decumul

import (
	"math"
	"sync/atomic"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestSimulateDeterministic(t *testing.T) {
	p := Plan{
		Capital: 1_500_000, NeedAnnual: 48000, Years: 35,
		Tax:    CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	a := p.Simulate(20000, 4, 7).RuinProb()
	b := p.Simulate(20000, 4, 7).RuinProb()
	if a != b {
		t.Errorf("not reproducible: %.4f vs %.4f", a, b)
	}
	if a < 0 || a > 1 {
		t.Errorf("ruin prob out of range: %.4f", a)
	}
}

func TestSimulateMoreCapitalLowerRuin(t *testing.T) {
	mk := func(c float64) Plan {
		return Plan{Capital: c, NeedAnnual: 48000, Years: 35, Tax: CTOFlatTax{Rate: 0.30},
			Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35}}
	}
	low := mk(1_200_000).Simulate(20000, 4, 7).RuinProb()
	high := mk(2_500_000).Simulate(20000, 4, 7).RuinProb()
	if !(high < low) {
		t.Errorf("more capital should lower ruin: low=%.4f high=%.4f", low, high)
	}
}

// TestSimulateOnMatchesRunPath pins the arena: the series a path gets out of
// SimulateOn's one shared block must be exactly what the standalone kernel
// builds for itself, path by path and point by point. It is the guard on a
// shared FIRE URL reproducing byte for byte.
func TestSimulateOnMatchesRunPath(t *testing.T) {
	p := Plan{
		Capital: 800_000, NeedAnnual: 32_000, Years: 25,
		Tax: CTOFlatTax{Rate: 0.30}, Buffer: BufferSleeve{Years: 2},
		Cashflows: []Cashflow{{FromYear: 10, Annual: 9_000}},
		Flex:      FlexRule{Threshold: 0.20, Cut: 0.10},
		Source:    scenario.ParametricSource{Mu: 0.05, Sigma: 0.15, Df: 5, Periods: 25},
	}
	draws := p.Draw(64, 4, 7)
	e := p.SimulateOn(draws, 4)
	for i, seq := range draws.Returns {
		want := p.RunPath(seq, Lives{})
		got := e.Paths[i]
		for k := range want.Wealth {
			if got.Wealth[k] != want.Wealth[k] {
				t.Fatalf("path %d wealth[%d] = %v, want %v", i, k, got.Wealth[k], want.Wealth[k])
			}
		}
		for k := range want.Spend {
			if got.Spend[k] != want.Spend[k] {
				t.Fatalf("path %d spend[%d] = %v, want %v", i, k, got.Spend[k], want.Spend[k])
			}
		}
		if got.Ruined != want.Ruined || got.TaxPaid != want.TaxPaid || got.Withdrawn != want.Withdrawn {
			t.Fatalf("path %d summary %+v, want %+v", i, got, want)
		}
	}
}

// TestShortcutsMatchEnsemble pins every driver that skips building the
// Ensemble to the Ensemble it stands for, bit for bit, over the whole plan
// matrix (each spending rule, the envelopes, the monthly and the lifetime
// kernels), a hand-assembled Draws without lifespans included: RuinProbOn and
// RuinProb to Ensemble.RuinProb, ruinAbove's early verdict to the comparison
// it settles, ruinAndMedianOn to the two Outcome fields a sweep reads,
// SolveOn to Solve and BestBuffer to the full sweep it used to read.
func TestShortcutsMatchEnsemble(t *testing.T) {
	for _, c := range determinismPlans() {
		p := c.plan
		draws := p.Draw(300, 3, 5)
		// The kernel does not care which goroutine ran a path, so a full
		// Draws replays identically at any worker count...
		if got, want := p.RuinProbOn(draws, 7), p.SimulateOn(draws, 1).RuinProb(); got != want {
			t.Errorf("%s: RuinProbOn at 7 workers = %v, want %v", c.name, got, want)
		}
		// ...but a Draws without lifespans gets them drawn at the caller's
		// worker count, so that case is compared at one count throughout.
		for _, d := range []Draws{draws, {Returns: draws.Returns}} {
			e := p.SimulateOn(d, 3)
			o := e.Outcome()
			if got := p.RuinProbOn(d, 3); got != o.RuinProb {
				t.Errorf("%s: RuinProbOn = %v, want %v", c.name, got, o.RuinProb)
			}
			for _, target := range []float64{-1, 0, 0.01, 0.05, o.RuinProb, 0.5, 1} {
				if got, want := p.ruinAbove(d, 3, target), o.RuinProb > target; got != want {
					t.Errorf("%s: ruinAbove(%v) = %v, want %v (ruin %v)", c.name, target, got, want, o.RuinProb)
				}
			}
			ruin, p50 := p.ruinAndMedianOn(d, 3)
			if ruin != o.RuinProb || p50 != o.TerminalP50 {
				t.Errorf("%s: ruinAndMedianOn = %v, %v, want %v, %v", c.name, ruin, p50, o.RuinProb, o.TerminalP50)
			}
		}
		if got, want := p.RuinProb(300, 3, 5), p.Simulate(300, 3, 5).RuinProb(); got != want {
			t.Errorf("%s: RuinProb = %v, want %v", c.name, got, want)
		}
		axis := WithdrawalAxis(0, 150_000)
		if got, want := p.SolveOn(0.05, axis, draws, 2), p.Solve(0.05, axis, 300, 3, 5); got != want {
			t.Errorf("%s: SolveOn = %v, want %v", c.name, got, want)
		}
		buffers := []float64{0, 1, 2, 4}
		years, ruin, err := p.BestBuffer(buffers, 300, 3, 5)
		if err != nil {
			t.Fatal(err)
		}
		wantYears, wantRuin := 0.0, math.Inf(1)
		for _, v := range buffers {
			q := p
			q.Buffer.Years = v
			if r := q.SimulateOn(draws, 3).RuinProb(); r < wantRuin {
				wantYears, wantRuin = v, r
			}
		}
		if years != wantYears || ruin != wantRuin {
			t.Errorf("%s: BestBuffer = %v, %v, want %v, %v", c.name, years, ruin, wantYears, wantRuin)
		}
	}
}

// TestRuinProbOnEmpty pins the degenerate run: no path, no ruin.
func TestRuinProbOnEmpty(t *testing.T) {
	p := determinismPlan(10)
	if got := p.RuinProbOn(Draws{}, 4); got != 0 {
		t.Errorf("RuinProbOn(empty) = %v, want 0", got)
	}
	if !p.ruinAbove(Draws{}, 4, -1) || p.ruinAbove(Draws{}, 4, 0) {
		t.Error("ruinAbove(empty) must read a zero ruin")
	}
	if r, m := p.ruinAndMedianOn(Draws{}, 4); r != 0 || m != 0 {
		t.Errorf("ruinAndMedianOn(empty) = %v, %v, want 0, 0", r, m)
	}
}

// TestForEachPathCoversEveryIndexOnce pins the chunked scheduler: every index
// is handed out exactly once, whatever the split.
func TestForEachPathCoversEveryIndexOnce(t *testing.T) {
	for _, n := range []int{0, 1, 7, 100, 1000, 4097} {
		for _, workers := range []int{0, 1, 3, 16} {
			seen := make([]int32, n)
			forEachPath(n, workers, func(_ int, loop func(func(int))) {
				loop(func(i int) { atomic.AddInt32(&seen[i], 1) })
			})
			for i, s := range seen {
				if s != 1 {
					t.Fatalf("n=%d workers=%d: index %d visited %d times", n, workers, i, s)
				}
			}
		}
	}
}
