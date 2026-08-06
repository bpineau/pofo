package web

import (
	"testing"
)

// With no portfolio panel, Models returns the parametric family (Student-t,
// Regime, Conservative). Sequence risk and the conservative prior must each be
// at least as risky as the plain i.i.d. Student-t, and the conservative prior
// must not exceed its safe withdrawal: this is the calibrated gradient that
// replaces the single hypersensitive number.
func TestModelsParametricOrdering(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 30,
		Mu: 0.04, Sigma: 0.16, Df: 5, TaxRate: 0.30, NPaths: 3000}

	res := Models(pr, nil)

	by := map[string]ModelStat{}
	for _, m := range res.Models {
		by[m.Name] = m
	}
	for _, n := range []string{"Student-t", "Regime", "Conservative"} {
		if _, ok := by[n]; !ok {
			t.Fatalf("missing model %q (have %v)", n, by)
		}
	}
	st, rg, cons := by["Student-t"], by["Regime"], by["Conservative"]
	if rg.Ruin+1e-9 < st.Ruin {
		t.Errorf("regime ruin %.3f should be >= student-t %.3f (sequence risk)", rg.Ruin, st.Ruin)
	}
	if cons.Ruin+1e-9 < st.Ruin {
		t.Errorf("conservative ruin %.3f should be >= student-t %.3f", cons.Ruin, st.Ruin)
	}
	if cons.SafeWR > st.SafeWR+1e-9 {
		t.Errorf("conservative safe WR %.4f should be <= student-t %.4f", cons.SafeWR, st.SafeWR)
	}
	for _, m := range res.Models {
		if m.Ruin < 0 || m.Ruin > 1 {
			t.Errorf("%s ruin out of range: %.3f", m.Name, m.Ruin)
		}
		if m.Help == "" {
			t.Errorf("%s has no hover help", m.Name)
		}
	}
}

// With a panel long enough for the horizon, the historical and block-bootstrap
// columns are added, and a confidence level is reported.
func TestModelsWithPanelAddsHistorical(t *testing.T) {
	panel := monthlyPanel(360) // ~30y of monthly history
	pr := Params{Capital: 1_800_000, NeedAnnual: 60000, Years: 20,
		Mu: 0.04, Sigma: 0.16, Df: 5, TaxRate: 0.30, NPaths: 1500, Weights: []float64{1}}

	res := Models(pr, &panel)

	names := map[string]bool{}
	for _, m := range res.Models {
		names[m.Name] = true
	}
	for _, n := range []string{"Historical", "Block bootstrap", "Student-t", "Regime", "Conservative"} {
		if !names[n] {
			t.Errorf("expected model %q with a panel", n)
		}
	}
	if res.Confidence == "" {
		t.Errorf("expected a confidence level")
	}
	if res.Verdict == "" {
		t.Errorf("expected a verdict line")
	}
}
