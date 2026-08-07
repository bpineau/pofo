package web

import (
	"testing"
)

// With no portfolio panel, Models returns the synthetic family (Student-t,
// Sequence stress, Broad-sample, Lost decade). Each stress must be at least as
// risky as the plain i.i.d. Student-t, and the broad-sample prior must not
// exceed its safe withdrawal: this is the calibrated gradient that replaces the
// single hypersensitive number.
func TestModelsParametricOrdering(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 3000}

	res := Models(pr, nil)

	by := map[string]ModelStat{}
	for _, m := range res.Models {
		by[m.Name] = m
	}
	for _, n := range []string{"Student-t", "Sequence stress", "Broad-sample", "Lost decade"} {
		if _, ok := by[n]; !ok {
			t.Fatalf("missing model %q (have %v)", n, by)
		}
	}
	st, rg, cons, lost := by["Student-t"], by["Sequence stress"], by["Broad-sample"], by["Lost decade"]
	if rg.Ruin+1e-9 < st.Ruin {
		t.Errorf("sequence-stress ruin %.3f should be >= student-t %.3f (sequence risk)", rg.Ruin, st.Ruin)
	}
	if cons.Ruin+1e-9 < st.Ruin {
		t.Errorf("broad-sample ruin %.3f should be >= student-t %.3f", cons.Ruin, st.Ruin)
	}
	if lost.Ruin+1e-9 < st.Ruin {
		t.Errorf("lost-decade ruin %.3f should be >= student-t %.3f", lost.Ruin, st.Ruin)
	}
	if cons.SafeWR > st.SafeWR+1e-9 {
		t.Errorf("broad-sample safe WR %.4f should be <= student-t %.4f", cons.SafeWR, st.SafeWR)
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

// With a short history relative to the horizon, the central (Student-t) model
// shrinks the rosy fit toward the broad-sample prior: its safe withdrawal must
// land below the raw fit (no shrink) yet stay above the conservative floor, a
// believable middle, not a collapse to doom.
func TestCentralShrinksTowardPriorOnShortHistory(t *testing.T) {
	panel := monthlyPanel(240) // 20y of history
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 45,
		Mu: 0.07, Sigma: 0.10, Df: 8, TaxRate: 0.30, NPaths: 3000, Weights: []float64{1}}

	get := func(r ModelsResult, name string) ModelStat {
		for _, m := range r.Models {
			if m.Name == name {
				return m
			}
		}
		t.Fatalf("model %q missing", name)
		return ModelStat{}
	}
	rawSafe := get(Models(pr, nil), "Student-t").SafeWR // no panel: the rosy fit, no shrink
	withPanel := Models(pr, &panel)
	blendedSafe := get(withPanel, "Student-t").SafeWR
	consSafe := get(withPanel, "Broad-sample").SafeWR

	if !(blendedSafe < rawSafe) {
		t.Errorf("blended central safe WR %.3f should be below the raw fit %.3f (shrunk toward the prior)", blendedSafe, rawSafe)
	}
	if !(blendedSafe > consSafe) {
		t.Errorf("blended central safe WR %.3f should stay above the conservative floor %.3f (not doom)", blendedSafe, consSafe)
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
	for _, n := range []string{"Historical windows", "Block bootstrap", "Student-t", "Sequence stress", "Broad-sample", "Lost decade"} {
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

// The risk-based guardrail's table loosens as the horizon shortens (that is
// its whole point), and ticking the rule changes the plan's outcome, i.e. it
// is really wired into the kernel and not merely parsed.
func TestRiskGuardrailsWired(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 42_000, Years: 35,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1000, TargetRuin: 0.05,
		Weights: []float64{1}}
	tab := pr.safeRateTable()
	if len(tab) != pr.Years {
		t.Fatalf("table has %d entries, want %d", len(tab), pr.Years)
	}
	if tab[0] <= 0 || tab[0] > 0.10 {
		t.Errorf("the long-horizon safe rate looks wrong: %.3f", tab[0])
	}
	for k := 1; k < len(tab); k++ {
		if tab[k] < tab[k-1]-1e-9 {
			t.Fatalf("the safe rate must never fall as the horizon shortens (year %d: %.4f then %.4f)",
				k, tab[k-1], tab[k])
		}
	}
	if tab[len(tab)-1] <= tab[0] {
		t.Errorf("the band must open with the shortening horizon: %.4f then %.4f", tab[0], tab[len(tab)-1])
	}

	fixedRuin := Compute(pr).Cards[0].Value
	pr.RiskGuard = true
	guardedRuin := Compute(pr).Cards[0].Value
	if fixedRuin == guardedRuin {
		t.Errorf("ticking the risk guardrail changed nothing (ruin %s both ways)", fixedRuin)
	}
}
