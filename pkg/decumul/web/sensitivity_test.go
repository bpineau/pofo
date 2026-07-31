package web

import (
	"strings"
	"testing"
)

func TestSensitivityRendersSignedBars(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 50000, Years: 40,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500}

	res := Sensitivity(pr, nil)

	if !strings.HasPrefix(res.SVG, "<svg") {
		t.Fatalf("expected an SVG, got %.30q", res.SVG)
	}
	// Spending less and adding capital both reduce ruin (green); they must appear.
	for _, label := range []string{"Spend -5 k€/yr", "Capital +100 k€"} {
		if !strings.Contains(res.SVG, label) {
			t.Errorf("missing lever %q", label)
		}
	}
	if !strings.Contains(res.SVG, "#12B76A") {
		t.Errorf("expected at least one ruin-reducing (green) bar")
	}
}
