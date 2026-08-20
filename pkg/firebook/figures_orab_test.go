package firebook

import (
	"math"
	"strings"
	"testing"
)

// Every number the plate draws is quoted by the article's example, so the
// article's prose is the fixture: if a value moves there, this fails and the
// plate is corrected, never the other way round.
func TestGoldABNumbersComeFromTheArticle(t *testing.T) {
	art, err := assets.ReadFile("assets/book/fr/or-en-retrait.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Plan : 1,5 M€, 51 000 €/an, corridor Vanguard, 45 ans",
		"Variante A : 70 % actions / 30 % obligations intermédiaires",
		"Variante B : 70 / 20 / 10 % or",
		"Au central, 4,1 % pour A contre 3,9 % pour B",
		"Au stress de séquence, on passe de 7,2 à 6,1 %",
		"Sur les modèles à inflation longue, de 10,8 à 8,9 %",
		"1966 passe de « épuisé à l'année 27 » à « traversé, amoché »",
		"2000 reste quasi inchangé",
		"Richesse médiane à 45 ans : −5 %",
		"::: figure or-ab-modeles",
	} {
		if !strings.Contains(string(art), want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The English edition carries the same example, amounts in dollars.
	en, err := assets.ReadFile("assets/book/en/gold-in-retirement.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Central case, 4.1% failure for A against 3.9% for B",
		"Sequence stress, from 7.2 down to 6.1%",
		"Long-inflation models, from 10.8 down to 8.9%",
		`1966 goes from "exhausted in year 27" to "ridden out, battered"`,
		"Median wealth at 45 years: down 5%",
		"::: figure or-ab-modeles",
	} {
		if !strings.Contains(string(en), want) {
			t.Errorf("the English article no longer says %q", want)
		}
	}
}

// The shape of the answer is the plate's whole claim: gold never hurts on
// these rows, it does almost nothing in the central case, and what it removes
// grows with the hostility of the model. If that ordering ever broke, the
// plate would be drawing an argument the numbers no longer make.
func TestGoldABGainGrowsWithHostility(t *testing.T) {
	if len(goldABRows) != 3 {
		t.Fatalf("%d modelled rows, expected the article's three", len(goldABRows))
	}
	for i, r := range goldABRows {
		if r.with >= r.without {
			t.Errorf("row %q: gold takes ruin from %.1f to %.1f, which is not an improvement",
				r.label, r.without, r.with)
		}
		if r.without > goldABAxisMax {
			t.Errorf("row %q: %.1f %% runs off the axis", r.label, r.without)
		}
		if i > 0 && goldABGain(r) <= goldABGain(goldABRows[i-1]) {
			t.Errorf("row %q removes %.1f points, no more than the friendlier row above it",
				r.label, goldABGain(r))
		}
	}
	// The three gaps the prose describes, to the tenth of a point.
	for i, want := range []float64{0.2, 1.1, 1.9} {
		if got := goldABGain(goldABRows[i]); math.Abs(got-want) > 0.001 {
			t.Errorf("row %q removes %.2f points, expected %.1f", goldABRows[i].label, got, want)
		}
	}
	// The central row is the "next to nothing" of the prose: a fifth of a
	// point, and by far the shortest bar of the plate.
	if goldABGain(goldABRows[0]) > 0.5 {
		t.Errorf("the central row shows %.1f points, which is no longer next to nothing",
			goldABGain(goldABRows[0]))
	}
	// The cover is not free, and the plate prints its price.
	if goldABMedianCost >= 0 {
		t.Errorf("the median-wealth cost is %.0f %%, which would make the cover a free lunch", goldABMedianCost)
	}
}

// The rendered plate obeys the book's drawing rules and carries every number
// and verdict it claims to draw.
func TestGoldABPlateRenders(t *testing.T) {
	svg := FigureSVG("or-ab-modeles")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"rgba(", "opacity", "\u2014", "\u2013", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, r := range goldABRows {
		for _, want := range []string{
			">" + r.label + "<",
			">" + frNum(r.without, 1) + "<",
			">" + frNum(r.with, 1) + "<",
			">−" + frNum(goldABGain(r), 1) + "<", // the column that reads the bar
		} {
			if !strings.Contains(svg, want) {
				t.Errorf("the plate does not draw %q", want)
			}
		}
	}
	for _, want := range []string{
		">" + goldABVintage + "<", ">" + goldABVerdictA + "<", ">" + goldABVerdictB + "<",
		">un verdict, pas une probabilité<", // the 1966 row is never a probability
		">avec or<", ">sans or<",
		"richesse médiane à 45 ans, −5 %.", // the price, always printed
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// One dot per variant per modelled row, and not a single one for the
	// vintage: that row is typeset, never plotted.
	if n := strings.Count(svg, "<circle"); n != 2*len(goldABRows) {
		t.Errorf("%d dots drawn, expected %d", n, 2*len(goldABRows))
	}
}
