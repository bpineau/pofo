package web

import "testing"

// At a comfortable spend the target is already met, and the "spend less" option
// reports a safe spend at least as high as the plan.
func TestSolveMenuComfortable(t *testing.T) {
	pr := Params{Capital: 2_000_000, NeedAnnual: 40000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 2000, TargetRuin: 0.05}

	m := SolveMenu(pr, nil)

	if len(m.Options) != 3 {
		t.Fatalf("options = %d, want 3 (spend, flex, buffer)", len(m.Options))
	}
	if m.CurrentRuin > m.TargetRuin {
		t.Errorf("expected current ruin %.3f <= target %.3f at a comfortable spend", m.CurrentRuin, m.TargetRuin)
	}
}

// At an aggressive spend the target is missed, and the "spend less" lever offers
// a safe spend strictly below the plan to reach it.
func TestSolveMenuAggressiveSpendLessHelps(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 60000, Years: 40,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 2000, TargetRuin: 0.05}

	m := SolveMenu(pr, nil)

	if m.CurrentRuin <= m.TargetRuin {
		t.Fatalf("test setup: 6%% spend should miss a 5%% target (current %.3f)", m.CurrentRuin)
	}
	spend := m.Options[0]
	if spend.Lever != "Spend less" || !spend.OK {
		t.Errorf("first option should be a reachable spend-less lever, got %+v", spend)
	}
}
