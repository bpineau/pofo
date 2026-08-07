package web

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The market view must render one fan per planning model, growth-of-1 based,
// with a bear-texture caption, and the stress models must show meaner bears
// than the central case.
func TestMarket(t *testing.T) {
	pr := Params{Capital: 1_500_000, NeedAnnual: 60000, Years: 40,
		Mu: 0.05, Sigma: 0.11, Df: 5}
	res := Market(pr, nil)
	if len(res.Fans) != len(fanModels) {
		t.Fatalf("want %d market fans, got %d", len(fanModels), len(res.Fans))
	}
	depth := map[string]float64{}
	capRe := regexp.MustCompile(`typical: −([0-9]+)% and ([0-9]+)y`)
	for _, f := range res.Fans {
		if !strings.Contains(f.SVG, "<svg") || !strings.Contains(f.SVG, "growth of 1 real €") {
			t.Errorf("%s: market fan malformed", f.Name)
		}
		m := capRe.FindStringSubmatch(f.SVG)
		if m == nil {
			t.Errorf("%s: bear-texture caption missing", f.Name)
			continue
		}
		d, _ := strconv.ParseFloat(m[1], 64)
		depth[f.Name] = d
		if d < 5 || d > 90 {
			t.Errorf("%s: implausible typical worst bear %.0f%%", f.Name, d)
		}
	}
	// The sequence-clustering stress must show a deeper typical bear than the
	// i.i.d. central case; the lost decade deeper still (its whole point).
	if depth["Sequence stress"] <= depth["Student-t"] {
		t.Errorf("sequence stress typical bear %.0f%% should exceed student-t %.0f%%",
			depth["Sequence stress"], depth["Student-t"])
	}
	if depth["Lost decade"] <= depth["Student-t"] {
		t.Errorf("lost decade typical bear %.0f%% should exceed student-t %.0f%%",
			depth["Lost decade"], depth["Student-t"])
	}
	// Deterministic across calls (fixed seed).
	if again := Market(pr, nil); again.Fans[0].SVG != res.Fans[0].SVG {
		t.Errorf("market view is not deterministic")
	}
}

// cumIndex and bearTexture agree on a hand-checkable path.
func TestBearTexture(t *testing.T) {
	idx := cumIndex([]float64{0.10, -0.50, 0.20, 0.50, 0.20})
	// peaks: 1, 1.10; trough 0.55 (-50%); 0.66 and 0.99 stay under water;
	// 1.188 finally clears the 1.10 peak, ending a 3-year spell.
	depth, spell := bearTexture(idx)
	if depth < 0.499 || depth > 0.501 {
		t.Errorf("depth = %v, want 0.50", depth)
	}
	if spell != 3 {
		t.Errorf("spell = %v, want 3 (years below the 1.10 peak)", spell)
	}
}
