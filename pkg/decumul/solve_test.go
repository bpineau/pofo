package decumul

import (
	"math"
	"testing"
)

// Solve on the withdrawal axis must return the spending level whose ruin sits at
// the target, and a materially higher spend must be riskier (the axis is
// monotonic increasing in ruin).
func TestSolveWithdrawalHitsTargetRuin(t *testing.T) {
	p := sweepPlan() // Capital 1.5M, 35y, parametric
	const target = 0.10
	const n, w, seed = 6000, 4, uint64(7)

	got := p.Solve(target, WithdrawalAxis(10000, 150000), n, w, seed)

	at := p
	at.NeedAnnual = got
	ruinAt := at.Simulate(n, w, seed).RuinProb()
	if math.Abs(ruinAt-target) > 0.03 {
		t.Errorf("ruin at solved withdrawal %.0f = %.3f, want ~%.2f", got, ruinAt, target)
	}

	hi := p
	hi.NeedAnnual = got * 1.3
	if hi.Simulate(n, w, seed).RuinProb() <= ruinAt {
		t.Errorf("a 30%% higher withdrawal should be riskier than the solved one")
	}
}

// The capital axis must reproduce CapitalForRuin exactly: same bisection, same
// shared draws, so Solve is validated against the trusted existing solver.
func TestSolveCapitalMatchesCapitalForRuin(t *testing.T) {
	p := sweepPlan()
	const target = 0.05
	const n, w, seed = 4000, 4, uint64(7)

	want := p.CapitalForRuin(target, 500000, 5_000_000, n, w, seed)
	got := p.Solve(target, CapitalAxis(500000, 5_000_000), n, w, seed)

	if math.Abs(want-got) > 1e-6 {
		t.Errorf("Solve(CapitalAxis) = %.4f, want CapitalForRuin = %.4f", got, want)
	}
}

// The flex-cut axis is decreasing in ruin (a deeper downturn cut lowers ruin),
// so Solve returns the smallest cut that reaches the target: applying it must
// hit the target while no cut at all stays riskier.
func TestSolveFlexCutReachesTarget(t *testing.T) {
	p := sweepPlan()
	p.NeedAnnual = 66000 // a high enough spend that the fixed rule misses the target
	p.Flex.Threshold = 0.20
	const target = 0.10
	const n, w, seed = 6000, 4, uint64(7)

	noCut := p
	noCut.Flex.Cut = 0
	if noCut.Simulate(n, w, seed).RuinProb() <= target {
		t.Fatalf("test setup: fixed rule already meets the target, nothing to solve")
	}

	cut := p.Solve(target, FlexCutAxis(0, 0.60), n, w, seed)
	at := p
	at.Flex.Cut = cut
	if ruinAt := at.Simulate(n, w, seed).RuinProb(); ruinAt > target+0.03 {
		t.Errorf("ruin at solved flex cut %.2f = %.3f, want <= ~%.2f", cut, ruinAt, target)
	}
}
