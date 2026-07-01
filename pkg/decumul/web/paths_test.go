package web

import (
	"strings"
	"testing"
)

// Paths renders one fan per planning model (the four synthetic columns), each a
// bands-and-samples SVG.
func TestPathsRendersFans(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 30,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500}

	res := Paths(pr, nil)

	if len(res.Fans) != len(fanModels) {
		t.Fatalf("fans = %d, want %d (one per planning model)", len(res.Fans), len(fanModels))
	}
	for i, f := range res.Fans {
		if f.Name != fanModels[i] {
			t.Errorf("fan %d = %q, want %q", i, f.Name, fanModels[i])
		}
		if !strings.HasPrefix(f.SVG, "<svg") || !strings.Contains(f.SVG, "<polygon") {
			t.Errorf("fan %q: expected a fan SVG with bands, got %.30q", f.Name, f.SVG)
		}
	}
}
