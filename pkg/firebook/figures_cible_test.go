package firebook

import (
	"math"
	"strings"
	"testing"
)

// cibleK mirrors how the cascade prints a capital amount in thousands.
func cibleK(v float64) string {
	return strings.TrimSuffix(euroFR(math.Round(v/1000)), " €")
}

// The first plate claims to draw a definition, so the definition is what the
// test checks: the multiple is exactly 100 over the rate, the three multiples
// posed on the curve are the ones the article's table names, and the three
// bracketed costs are the differences between them.
func TestCibleCourbeEstUnSurX(t *testing.T) {
	for _, r := range []float64{5, 4.5, 4, 3.5, 3, 2.5, 2.4} {
		if got := cibleMultiple(r) * r; math.Abs(got-100) > 1e-9 {
			t.Errorf("multiple(%.2f) x rate = %.6f, expected 100", r, got)
		}
	}
	for _, tc := range []struct {
		rate  float64
		label string
	}{
		{4.0, "25x"},
		{3.5, "28,6x"},
		{3.0, "33x"},
		{2.5, "40x"},
	} {
		decimals := 0
		if strings.Contains(tc.label, ",") {
			decimals = 1
		}
		if got := frNum(cibleMultiple(tc.rate), decimals) + "x"; got != tc.label {
			t.Errorf("at %.1f %% the plate draws %q, the curve says %q", tc.rate, tc.label, got)
		}
	}
	// The convexity, which is the whole argument: each half point costs more
	// than the one before it.
	var prev float64
	for _, tc := range []struct {
		r0, r1 float64
		want   string
	}{
		{4.0, 3.5, "3,6"},
		{3.5, 3.0, "4,8"},
		{3.0, 2.5, "6,7"},
	} {
		d := cibleMultiple(tc.r1) - cibleMultiple(tc.r0)
		if got := frNum(d, 1); got != tc.want {
			t.Errorf("from %.1f %% to %.1f %%: +%sx, the plate says +%sx", tc.r0, tc.r1, got, tc.want)
		}
		if d <= prev {
			t.Errorf("the curve is not convex: step %.2fx follows %.2fx", d, prev)
		}
		prev = d
	}
}

// The article's ceiling line has to be the level the plate shades.
func TestCiblePlafondEstTrenteTrois(t *testing.T) {
	if got := cibleMultiple(3); math.Abs(got-33.333333) > 1e-5 {
		t.Errorf("the 3 %% line sits at %.4fx, expected 33,33x", got)
	}
	if cibleMultiple(2.5) <= cibleMultiple(3) {
		t.Error("the zone beyond 33x must lie above the 3 % marker")
	}
}

// The cascade must reconcile from the first marche to the last, both in exact
// euros and at the precision it prints: a reader who adds the printed steps
// has to land on the printed target.
func TestCibleCascadeReconciles(t *testing.T) {
	steps := []float64{
		cibleCapital(cibleObserve),
		cibleCapital(cibleLoisirs),
		cibleCapital(cibleMutuelle),
		cibleTarget() - cibleNet(),
	}
	var sum float64
	for _, s := range steps {
		sum += s
	}
	if math.Abs(sum-cibleTarget()) > 1e-6 {
		t.Errorf("the four marches add up to %.2f €, the target is %.2f €", sum, cibleTarget())
	}
	var kSum float64
	for _, s := range steps {
		kSum += math.Round(s / 1000)
	}
	if want := math.Round(cibleTarget() / 1000); kSum != want {
		t.Errorf("the printed marches add up to %.0f k€, the printed target is %.0f k€", kSum, want)
	}
	// The friction is a gross-up on the withdrawal, not a share of the target.
	if got := cibleNet() / (1 - cibleFriction); math.Abs(got-cibleTarget()) > 1e-9 {
		t.Errorf("the friction step does not gross up: %.2f vs %.2f", got, cibleTarget())
	}
	// The two numbers the plate prints in full.
	if got := euroFR(math.Round(cibleTarget()/1000) * 1000); got != "1 547 000 €" {
		t.Errorf("target printed %q, the article says 1 547 000 €", got)
	}
	if got := euroFR(math.Round((cibleTarget()+ciblePension)/1000) * 1000); got != "1 747 000 €" {
		t.Errorf("the plan without the pension prints %q, expected 1 747 000 €", got)
	}
	// The subtitle's rule of thumb, and the étape-1 claim it explains.
	if got := math.Round(12 / cibleTaux); got != 343 {
		t.Errorf("one euro a month weighs %.0f € of capital, the plate says 343 €", got)
	}
	forgotten := cibleCapital(200)
	if forgotten < 60000 || forgotten > 80000 {
		t.Errorf("200 €/mois forgotten weigh %.0f €, outside the 60 000 to 80 000 € the article announces", forgotten)
	}
}

// Figure and prose must agree. Every number the two plates draw is one the
// article states, so a reader can audit the plate against the text; this fails
// the moment one of them leaves the markdown.
func TestCiblePlatesAgreeWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/combien-il-vous-faut.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	for _, want := range []string{
		"::: figure cible-convexite",
		"::: figure cible-cascade",
		// the étape-4 table the curve carries
		"| 25 (4 %) | 4,0 % |",
		"| 28-29 (3,5 %) | ~3,5 % |",
		"| 33 (3 %) | 3,0 % |",
		"| > 33 | < 3 % |",
		"3 à 6 ans de travail supplémentaires",
		// the worked case the cascade walks
		"3 400 €/mois",
		"+ 350 €/mois",
		"+ 220 €/mois",
		"3 970 €/mois",
		"12 %",
		"28,6x",
		"1 547 000 €",
		"200 000 €",
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer carries %q, which a plate draws", want)
		}
	}
}

func TestCiblePlatesRender(t *testing.T) {
	for _, tc := range []struct {
		id    string
		draws []string
	}{
		{"cible-convexite", []string{
			"Chaque demi-point de prudence coûte plus cher que le précédent",
			"capital cible, en multiples de dépenses annuelles",
			"au-delà de 33x, le risque dominant n'est plus la ruine :",
			"25x", "28,6x", "33x", "40x",
			"+ 3,6x", "+ 4,8x", "+ 6,7x",
			"taux de retrait initial",
		}},
		{"cible-cascade", []string{
			"Du relevé bancaire au capital cible, marche par marche",
			"dépenses observées", "voyages et loisirs", "mutuelle santé",
			"friction fiscale", "× 28,6 (3,5 %)",
			"3 400 €/mois", "+ 350 €/mois", "+ 220 €/mois", "+ 12 % du brut",
			"1 547 000 €",
			"− 200 000 € : la retraite légale, comptée en revenu différé",
			cibleK(cibleTarget()),
		}},
	} {
		svg := FigureSVG(tc.id)
		if svg == "" {
			t.Fatalf("%s is not registered", tc.id)
		}
		for _, want := range tc.draws {
			if !strings.Contains(svg, want) {
				t.Errorf("%s does not draw %q", tc.id, want)
			}
		}
		if strings.Contains(svg, "—") { // the em-dash, banned book-wide
			t.Errorf("%s: no em-dash in a figure", tc.id)
		}
		if strings.Contains(svg, "rgba(") || strings.Contains(svg, "opacity") {
			t.Errorf("%s: no rgba and no opacity, crengine paints them solid black", tc.id)
		}
	}
}
