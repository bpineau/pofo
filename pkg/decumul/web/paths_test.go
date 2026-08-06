package web

import (
	"strings"
	"testing"
)

func TestPathsRendersFan(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500}

	res := Paths(pr, nil)

	if res.Model != "Student-t" {
		t.Errorf("default model = %q, want the central Student-t", res.Model)
	}
	if !strings.HasPrefix(res.FanSVG, "<svg") || !strings.Contains(res.FanSVG, "<polygon") {
		t.Errorf("expected a fan SVG with bands, got %.30q", res.FanSVG)
	}
}

func TestPathsHonorsModelChoice(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500, FanModel: "Conservative"}

	if res := Paths(pr, nil); res.Model != "Conservative" {
		t.Errorf("model = %q, want Conservative", res.Model)
	}
}
