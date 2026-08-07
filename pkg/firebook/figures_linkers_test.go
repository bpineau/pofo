package firebook

import (
	"math"
	"strings"
	"testing"
)

// ladderRateBySum prices the ladder the way a buyer would: rung k costs the
// withdrawal discounted k years at the real rate, and the rungs must add up to
// the capital. It is an independent derivation of the closed form the plate
// uses, so the two disagreeing means the formula drifted.
func ladderRateBySum(r float64, n int) float64 {
	x := r / 100
	cost := 0.0
	for k := 1; k <= n; k++ {
		cost += math.Pow(1+x, -float64(k))
	}
	return 100 / cost
}

func TestLadderRateMatchesTheDiscountedRungs(t *testing.T) {
	for _, n := range []int{20, 30, 40} {
		for r := linkersRateLo; r <= linkersRateHi+1e-9; r += 0.1 {
			got, want := ladderRate(r, n), ladderRateBySum(r, n)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("N=%d r=%.2f: closed form is %v", n, r, got)
			}
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("N=%d r=%.2f: closed form %.9f, discounted rungs %.9f", n, r, got, want)
			}
		}
	}
}

// The zero real rate is the degenerate case: the ladder just splits the capital
// into n equal rungs, so the rate must be exactly 1/n and never a division by a
// vanishing denominator.
func TestLadderRateAtZeroRealRate(t *testing.T) {
	for _, n := range []int{20, 30, 40} {
		want := 100 / float64(n)
		if got := ladderRate(0, n); math.Abs(got-want) > 1e-12 {
			t.Errorf("N=%d at 0 %% real: %.6f, want %.6f (= 1/N)", n, got, want)
		}
		// Just off zero must stay continuous with the limit.
		if got := ladderRate(1e-7, n); math.Abs(got-want) > 1e-4 {
			t.Errorf("N=%d just above 0 %%: %.6f, want ~%.6f", n, got, want)
		}
	}
}

// Every number the article's prose quotes about the ladder, and the two the
// plate adds, checked against the formula. A number moving here means the
// article and the figure must move together.
func TestLinkersPlateReadingsMatchTheArticle(t *testing.T) {
	for _, c := range []struct {
		what   string
		got    float64
		want   float64 // as the article rounds it, one decimal
		reason string
	}{
		{"30 ans à 0 % réel", ladderRate(0, 30), 3.3, "« à 0 % réel, l'échelle 30 ans ne finance plus que ~3,3 % »"},
		{"30 ans à 1 % réel", ladderRate(1, 30), 3.9, "« taux réels ~1 %, l'échelle de 30 ans finance ~3,9 % »"},
		{"30 ans à 2 % réel", ladderRate(2, 30), 4.5, "« avec des taux réels TIPS autour de 2 %, ... environ 4,5 % par an »"},
		{"40 ans à 1 % réel", ladderRate(1, 40), 3.0, "the plate's note on the FIRE horizon"},
		{"20 ans à 0 % réel", ladderRate(0, 20), 5.0, "the shortest curve's anchor"},
	} {
		if rounded := math.Round(c.got*10) / 10; rounded != c.want {
			t.Errorf("%s: %.4f rounds to %.1f %%, the book says %.1f %% (%s)", c.what, c.got, rounded, c.want, c.reason)
		}
	}

	// The crossings the plate marks: where each ladder catches the 4 % rule.
	if got := ladderCrossing(30, linkersRule); math.Abs(got-1.2191) > 5e-4 {
		t.Errorf("30-year ladder reaches 4 %% at %.4f %% real, plate marks 1,2 %%", got)
	}
	if got := ladderCrossing(40, linkersRule); math.Abs(got-2.5244) > 5e-4 {
		t.Errorf("40-year ladder reaches 4 %% at %.4f %% real, plate marks 2,5 %%", got)
	}
	// A crossing must be a crossing: below it the ladder pays less than the rule.
	for _, n := range []int{30, 40} {
		c := ladderCrossing(n, linkersRule)
		if ladderRate(c-0.05, n) >= linkersRule || ladderRate(c+0.05, n) <= linkersRule {
			t.Errorf("N=%d: %.3f %% is not where the curve crosses %.1f %%", n, c, linkersRule)
		}
	}
	// The 20-year ladder never needs the window: it beats the rule everywhere on
	// the plate, which is why it carries no crossing mark.
	if ladderRate(linkersRateLo, 20) <= linkersRule {
		t.Errorf("20-year ladder funds %.2f %% at %.1f %% real, the plate assumes it stays above %.1f %%",
			ladderRate(linkersRateLo, 20), linkersRateLo, linkersRule)
	}
}

func TestLinkersPlateRenders(t *testing.T) {
	svg := figLinkersEchelle()
	for _, want := range []string{
		"3,3 %", "3,9 %", "4,5 %", // the readings on the 30-year curve
		"1,2 % réel",    // the crossing
		"2,5 %",         // the 40-year crossing
		"3,0 %",         // the 40-year reading at 1 % real
		"règle des 4 %", // the reference rule
		"20 ans", "30 ans", "40 ans",
		"TIPS de 2023-2024", // the dated observation, never an undated live rate
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("plate is missing %q", want)
		}
	}
	for _, banned := range []string{"rgba", "opacity", "—", "linearGradient"} {
		if strings.Contains(svg, banned) {
			t.Errorf("plate contains the banned %q", banned)
		}
	}
}
