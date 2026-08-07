package firebook

import (
	"math"
	"strings"
	"testing"
)

// The taxe-puma plate draws nothing but the closed formula of article D. 380-1.
// This pins the formula on every amount the article quotes in prose, so figure
// and article can never drift apart: change one and this test fails.
func TestCSMMatchesTheArticle(t *testing.T) {
	for _, tc := range []struct {
		name             string
		capital, activit float64
		want, tol        float64
	}{
		// "un rentier sans activité avec 60 000 € [...] paie une CSM ~ 2 340 €/an"
		{"60 k€ sans activité", 60000, 0, 2340, 5},
		// "avec 120 000 €, [...] elle grimpe à ~ 6 240 €"
		{"120 k€ sans activité", 120000, 0, 6240, 5},
		// "le maximum théorique, assiette plafonnée, atteint ~23 400 €/an"
		{"assiette plafonnée", 10 * csmPASS, 0, 23400, 40},
		// version A of the household example: "~70 k€/an [...] ~3 000 €/an"
		{"version A, 70 k€", 70000, 0, 3000, 15},
		// version B: "les revenus réalisés tombent à ~32 k€/an [...] ~520 €/an"
		{"version B, 32 k€", 32000, 0, 520, 5},
		// version C: "~10 k€/an [...] la CSM tombe à zéro"
		{"version C, 10 k€ d'activité", 70000, 10000, 0, 0},
		// "environ 9 600 €/an de revenus d'activité annulent la CSM intégralement"
		{"au seuil exact", 120000, csmSwitch * csmPASS, 0, 0},
		// below the franchise, nothing is owed whatever the activity
		{"sous la franchise", 24000, 0, 0, 0},
	} {
		if got := csm(tc.capital, tc.activit); math.Abs(got-tc.want) > tc.tol {
			t.Errorf("%s: CSM %.0f €, l'article dit %.0f €", tc.name, got, tc.want)
		}
	}

	// The thresholds the plate labels, and the rounded values the article gives
	// them ("~9 600 €/an", "~24 000 €/an", "~384 000 €").
	if got := csmSwitch * csmPASS; math.Abs(got-9612) > 0.5 {
		t.Errorf("seuil d'activité %.0f €, attendu 9 612 €", got)
	}
	if got := csmFranchise * csmPASS; math.Abs(got-24030) > 0.5 {
		t.Errorf("franchise %.0f €, attendu 24 030 €", got)
	}
	if got := csmBaseCap * csmPASS; math.Abs(got-384480) > 0.5 {
		t.Errorf("plafond d'assiette %.0f €, attendu 384 480 €", got)
	}

	// "trois années à 40 k€ de revenus du capital paient moins qu'une année à
	// 120 k€", the spreading argument of parade n° 2.
	if spread, once := 3*csm(40000, 0), csm(120000, 0); spread >= once {
		t.Errorf("étalement %.0f € contre %.0f € en une fois, l'article dit l'inverse", spread, once)
	}

	// Twelve years of bridge: "~3 000 €/an × 12 ans ~ 36 000 €" against
	// "~520 €/an ~ 6 200 €", so organizing is worth "~30 000 €".
	naive, tidy := 12*csm(70000, 0), 12*csm(32000, 0)
	if math.Abs(naive-36000) > 500 || math.Abs(tidy-6200) > 300 {
		t.Errorf("pont de 12 ans: %.0f € contre %.0f €, l'article dit 36 000 € contre 6 200 €", naive, tidy)
	}
	if gain := naive - tidy; math.Abs(gain-30000) > 800 {
		t.Errorf("l'organisation vaut %.0f €, l'article dit ~30 000 €", gain)
	}

	// The taper is linear and hits zero exactly at the threshold, never below.
	half := csm(120000, csmSwitch*csmPASS/2)
	if math.Abs(half-csm(120000, 0)/2) > 0.5 {
		t.Errorf("décote non linéaire: %.1f € à la moitié du seuil", half)
	}
	if got := csm(1e9, 2*csmSwitch*csmPASS); got != 0 {
		t.Errorf("CSM %.0f € au double du seuil d'activité, attendu 0", got)
	}
}

// The plate's fill ramp must be a monotone set of distinct SOLID hex steps, and
// the SVG must be free of the two things crengine paints solid black.
func TestPumaPlateRamp(t *testing.T) {
	seen := map[string]bool{}
	last := -1.0
	for _, s := range csmSteps {
		if seen[s.fill] {
			t.Errorf("step %s appears twice in the ramp", s.fill)
		}
		seen[s.fill] = true
		if !strings.HasPrefix(s.fill, "#") || len(s.fill) != 7 {
			t.Errorf("step %q is not a solid hex", s.fill)
		}
		if s.upTo <= last {
			t.Errorf("step bounds are not increasing at %v", s.upTo)
		}
		last = s.upTo
	}

	svg := figPumaInterrupteur()
	for _, banned := range []string{"rgba", "opacity", "—", "linearGradient"} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate contains %q", banned)
		}
	}
	// Every cell above the franchise and left of the switch is painted, and no
	// cell beyond the switch is: 40 of the 120 cells are blank by law.
	painted := strings.Count(svg, "<rect")
	if want := 80 + len(csmSteps); painted != want {
		t.Errorf("%d rects in the plate, expected %d cells plus %d key steps", painted, 80, len(csmSteps))
	}
	if figures["puma-interrupteur"] == nil {
		t.Error("puma-interrupteur is not registered in figures.go")
	}

	// The grid is sized so both legal thresholds land on a cell boundary, which
	// is what lets the hairlines sit flush with the fill and the blank zones be
	// exactly the zones the law describes. Keep it that way.
	colW, rowH := pumaActMax/pumaCols, pumaCapMax/pumaRows
	if off := math.Mod(csmSwitch*csmPASS, colW); off > 20 && colW-off > 20 {
		t.Errorf("le seuil d'activité tombe à %.0f € d'un bord de colonne", math.Min(off, colW-off))
	}
	if off := math.Mod(csmFranchise*csmPASS, rowH); off > 50 && rowH-off > 50 {
		t.Errorf("la franchise tombe à %.0f € d'un bord de ligne", math.Min(off, rowH-off))
	}
}

func TestFrSpace(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want string
	}{{0, "0"}, {500, "500"}, {2400, "2 400"}, {9612, "9 612"}, {384480, "384 480"}} {
		if got := frSpace(tc.in); got != tc.want {
			t.Errorf("frSpace(%.0f) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
