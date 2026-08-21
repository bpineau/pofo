package firebook

import (
	"math"
	"strings"
	"testing"
)

// The four curves, rebuilt from the bundled panel: this is what
// "make figure-drift" reports when the record is regenerated.
func TestTrinityCurvesMatchTheData(t *testing.T) {
	us := jstUSA(t)
	if us.first != trinityFirst || us.first+len(us.equity)-1 != trinityLast {
		t.Errorf("the USA record runs %d-%d, the plate says %d-%d",
			us.first, us.first+len(us.equity)-1, trinityFirst, trinityLast)
	}
	for _, a := range trinityAllocs {
		for i, r := range trinityRates {
			got, n := jstSuccess(us, a.equity, r/100, trinityYears)
			if n != trinityWindow {
				t.Fatalf("%d rolling windows, the plate says %d", n, trinityWindow)
			}
			if math.Abs(got*100-a.success[i]) > 0.1 {
				t.Errorf("%s at %.2f %%: %.1f %% succeed, the plate draws %.1f",
					a.name, r, got*100, a.success[i])
			}
		}
	}
}

// The article's three lasting lessons, checked on the frozen curves. If any of
// them failed on this deeper sample the plate would be contradicting the
// sentence it illustrates.
func TestTrinityLessonsHoldOnThisSample(t *testing.T) {
	// One: the cliff. Every allocation loses far more between four and five
	// than between three and four.
	for _, a := range trinityAllocs {
		three, four := trinityCliff(a)
		if four <= three {
			t.Errorf("%s: the step 4->5 costs %.1f points, the step 3->4 costs %.1f: no cliff",
				a.name, four, three)
		}
	}
	if trinityCliffFour() < 3*trinityCliffThree() {
		t.Errorf("on the 75/25 the cliff is only %.1f against %.1f: too shallow to be called one",
			trinityCliffFour(), trinityCliffThree())
	}
	// Two: the asymmetry. At the article's own four percent, too few equities
	// is far more dangerous than too many.
	tooFew := trinityAt(trinityAllocs[3], 4)  // 25/75
	tooMany := trinityAt(trinityAllocs[0], 4) // 100/0
	if tooFew >= tooMany-10 {
		t.Errorf("at 4 %%, 25/75 succeeds %.1f %% against %.1f for all equity: the asymmetry is gone",
			tooFew, tooMany)
	}
	// Three: at the top of the range all equity is the safest of the four, which
	// is the same lesson read where it surprises.
	for _, a := range trinityAllocs[1:] {
		if trinityAt(a, 8) >= trinityAt(trinityAllocs[0], 8) {
			t.Errorf("at 8 %%, %s does as well as all equity", a.name)
		}
	}
	if cr := trinityCross(); cr < 4 || cr > 5.5 {
		t.Errorf("all equity overtakes the 75/25 at %.2f %%, outside the range the plate marks", cr)
	}
	// Every curve is monotone: more withdrawal never helps.
	for _, a := range trinityAllocs {
		for i := 1; i < len(a.success); i++ {
			if a.success[i] > a.success[i-1] {
				t.Errorf("%s: success rises from %.2f %% to %.2f %%", a.name,
					trinityRates[i-1], trinityRates[i])
			}
		}
	}
}

// The plate against the article that carries it.
func TestTrinityAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "etude-trinity")
	for _, want := range []string{
		"::: figure trinity-falaise",
		"Trois enseignements durables sortent de cette grille.",
		"Entre 4 et 5 %, le succès s'effondre",
		"trop peu d'actions est bien plus dangereux que trop",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestTrinityPlateRenders(t *testing.T) {
	svg := FigureSVG("trinity-falaise")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"—", "–", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, a := range trinityAllocs {
		if !strings.Contains(svg, ">"+a.name+"<") {
			t.Errorf("the plate does not name %q", a.name)
		}
	}
	for _, want := range []string{
		">la falaise<",
		">à partir de 4,50 %, le 100 % actions<",
		">repasse devant le 75/25<",
		"Un point de retrait en plus, de 4 à 5 %, coûte 20 points de succès au 75/25",
		"Aucun chiffre publié par Trinity n'est repris ici",
		"1872-2020), soit 120 fenêtres glissantes de trente ans",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	if n := strings.Count(svg, "<polyline"); n != len(trinityAllocs) {
		t.Errorf("%d curves drawn, expected one per allocation", n)
	}
	if n := strings.Count(svg, "<circle"); n != 1 {
		t.Errorf("%d dots drawn, expected the crossing alone", n)
	}
}
