package web

import (
	"strings"
	"testing"
)

func TestFrontierRendersPerModelCurves(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500, TargetRuin: 0.04}

	res := Frontier(pr, nil)

	if !strings.HasPrefix(res.SVG, "<svg") {
		t.Fatalf("expected an SVG, got %.30q", res.SVG)
	}
	for _, name := range []string{"Student-t", "Regime", "Conservative"} {
		if !strings.Contains(res.SVG, name) {
			t.Errorf("frontier missing the %q curve", name)
		}
	}
	if !strings.Contains(res.SVG, "stroke-dasharray") {
		t.Errorf("expected the plan/target marker lines")
	}
}
