package web

import "testing"

// At a comfortable spend the target is already met, so the menu reports the
// spending headroom (a single "Room to spare" option) rather than the nonsense
// reach-the-target levers (a 0% cut, a 0-year buffer).
func TestSolveMenuComfortable(t *testing.T) {
	pr := Params{Capital: 2_000_000, NeedAnnual: 40000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 2000, TargetRuin: 0.05}

	m := SolveMenu(pr, nil)

	if !m.Met || m.CurrentRuin > m.TargetRuin {
		t.Errorf("expected the target met at a comfortable spend (ruin %.3f, target %.3f)", m.CurrentRuin, m.TargetRuin)
	}
	if len(m.Options) != 1 || m.Options[0].Lever != "Room to spare" {
		t.Fatalf("expected a single headroom option, got %+v", m.Options)
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
