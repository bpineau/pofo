package firebook

import (
	"math"
	"strings"
	"testing"
)

// The plate carries the article's numbers, so the article must still carry
// them. Each fragment is the sentence the corresponding bar is drawn from.
func TestClavierLeviersMatchesTheArticle(t *testing.T) {
	article := bookArticle(t, "les-maths-du-4-pourcent")
	for _, want := range []string{
		"::: figure clavier-leviers",
		"Horizon de 50 ans → le bonus fond, la pénalité s'allège un peu, ~3,4 %",
		"→ −0,5 à −1 point",
		"étage 1 raboté d'un point",
		"0,5 % de frais → −0,5, presque un pour un",
		"Règle flexible à plancher → +0,3 à +0,5",
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer reads %q: the plate draws that lever", want)
		}
	}
	if !strings.Contains(article, "::: figure cascade-4pct") {
		t.Error("the companion cascade plate must stay: this one measures deviations from its result")
	}
}

// The arithmetic the plate rests on: a lever is the distance between the
// reference rate and what the article says the hypothesis leaves behind.
func TestClavierLeviersArithmetic(t *testing.T) {
	if cascadeBase != 4.0 {
		t.Fatalf("the reference is %.1f %%, the article's cascade lands on 4,0 %%", cascadeBase)
	}
	horizon := cascadeLevers[2]
	if horizon.name != "Horizon de 50 ans" {
		t.Fatalf("row 3 is %q, expected the horizon lever", horizon.name)
	}
	if got := cascadeBase + horizon.near; math.Abs(got-3.4) > 1e-9 {
		t.Errorf("the horizon lever leaves %.2f %%, the article reads ~3,4 %%", got)
	}
	// Ranges point away from the reference and never change sign inside.
	for _, l := range cascadeLevers {
		if math.Abs(l.far) < math.Abs(l.near) {
			t.Errorf("%s: far bound %.1f is closer to the reference than the near bound %.1f", l.name, l.far, l.near)
		}
		if l.near*l.far <= 0 {
			t.Errorf("%s: the two bounds must share one sign, got %.1f and %.1f", l.name, l.near, l.far)
		}
	}
	// Sorted by absolute weight: the ranking is the message of the plate.
	weight := func(l cascadeLever) float64 { return math.Abs(l.near+l.far) / 2 }
	for i := 1; i < len(cascadeLevers); i++ {
		if weight(cascadeLevers[i]) > weight(cascadeLevers[i-1]) {
			t.Errorf("lever %d (%s) outweighs the one above it (%s): re-sort the plate",
				i, cascadeLevers[i].name, cascadeLevers[i-1].name)
		}
	}
}

// The horizon bar is the one that could be drawn wrong, so it gets its own
// guard: the amortization bonus alone loses 1,1 point between 30 and 50 years,
// but the sequence penalty lightens over the longer horizon and the net move
// the plate draws is smaller. Recompute both from the closed formula.
func TestClavierLeviersHorizonAgainstTheAmortizationFormula(t *testing.T) {
	// amortRate is the plate-side formula of figures_amortissement.go, so the
	// two plates of this article can never drift apart.
	at30, at50 := amortRate(amortReal, 30), amortRate(amortReal, 50)
	if math.Abs(at30-5.8) > 0.06 {
		t.Errorf("30 years at 4 %% real amortizes at %.2f %%, the article says 5,8 %%", at30)
	}
	if math.Abs(at50-4.7) > 0.06 {
		t.Errorf("50 years at 4 %% real amortizes at %.2f %%, the article says 4,7 %%", at50)
	}
	bonusLoss := at30 - at50 // 1,1 point, the stage-2 effect alone
	if bonusLoss < 1.0 || bonusLoss > 1.25 {
		t.Fatalf("the bonus falls by %.2f point between 30 and 50 years, expected about 1,1", bonusLoss)
	}
	drawn := math.Abs(cascadeLevers[2].near)
	if drawn >= bonusLoss {
		t.Errorf("the plate draws −%.1f for the horizon, which is at least the %.2f the bonus alone loses: "+
			"the article's net figure must stay below it", drawn, bonusLoss)
	}
	if drawn < 0.4 {
		t.Errorf("the plate draws −%.1f for the horizon, far below the bonus arithmetic", drawn)
	}
}

// The sample lever has a measured counterpart in the book: the safemax-pays
// plate ranks a 30-year SAFEMAX per country on the JST panel. The gap between
// the United States and the equal-weight world basket is a different
// construction from Anarkulova-Cederburg, so this only checks the order of
// magnitude the article claims.
func TestClavierLeviersSampleAgainstTheJSTPanel(t *testing.T) {
	var usa float64
	for _, c := range safemaxCountries {
		if c.ISO == "USA" {
			usa = c.Rate
		}
	}
	if usa == 0 {
		t.Fatal("the safemax panel no longer carries the United States")
	}
	gap := usa - safemaxWorld
	if gap < 0.3 || gap > 1.1 {
		t.Errorf("the JST panel puts the world basket %.2f point below the United States; "+
			"the plate draws a sample lever of −0,5 à −1, which no longer covers it", gap)
	}
	sample := cascadeLevers[1]
	if sample.name != "Échantillon mondial" {
		t.Fatalf("row 2 is %q, expected the sample lever", sample.name)
	}
	if math.Abs(sample.near) > gap+0.1 {
		t.Errorf("the drawn near bound −%.1f is harsher than the measured gap %.2f", math.Abs(sample.near), gap)
	}
}

// What the rendered plate must contain, and the house rules it must respect.
func TestClavierLeviersPlateReads(t *testing.T) {
	svg := figClavierLeviers()
	for _, want := range []string{
		"Le clavier des leviers : ce que chaque hypothèse déplace",
		">4,0 %<",
		">CAPE élevé au départ<",
		">−1,0<", ">−0,5 à −1,0<", ">−0,6<", ">−0,5<", ">+0,3 à +0,5<",
		"taux de retrait obtenu, en % du capital de départ",
		`viewBox="0 0 640 384"`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// The em-dash the book bans is spelled as an escape so this source file
	// stays free of it.
	for _, banned := range []string{"rgba(", "rotate(", "\u2014", "opacity"} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate contains %q, which the book's figure rules forbid", banned)
		}
	}
	if figures["clavier-leviers"] == nil {
		t.Error("the plate is not registered in figures.go")
	}
}
