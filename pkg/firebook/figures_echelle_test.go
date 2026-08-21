package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// The plate is a closed model: every number it draws is the arithmetic of the
// article's example. This checks that arithmetic year by year, against the
// rule the prose states.
func TestLadderYearFollowsTheCoverageRule(t *testing.T) {
	for n := 1; n <= ladderYears; n++ {
		matched, payer, portfolio := ladderYear(n)
		var wantMatched float64
		var wantPayer ladderPayer
		switch {
		case n <= 2:
			wantMatched, wantPayer = 45, payerEuros
		case n <= 6:
			wantMatched, wantPayer = 45, payerBond
		case n <= 8:
			wantMatched, wantPayer = 27, payerBond
		case n <= 13:
			wantMatched, wantPayer = 27, payerTips
		default:
			wantMatched, wantPayer = 24, payerPens
		}
		if math.Abs(matched-wantMatched) > 1e-9 || payer != wantPayer {
			t.Errorf("year %d: %.0f k from payer %d, expected %.0f k from payer %d",
				n, matched, payer, wantMatched, wantPayer)
		}
		if sum := matched + portfolio; math.Abs(sum-ladderComfort) > 1e-9 {
			t.Errorf("year %d: the bar sums to %.2f, not the comfort level %.0f", n, sum, ladderComfort)
		}
	}
	// The rule itself: the whole floor over the fragile window, 60 % of it
	// until the pensions, and never more than the comfort.
	for n := 1; n <= ladderFullYears; n++ {
		if m, _, _ := ladderYear(n); m != ladderFloor {
			t.Errorf("year %d matches %.0f k, the rule says the whole floor", n, m)
		}
	}
	for n := ladderFullYears + 1; n <= ladderBridge; n++ {
		if m, _, _ := ladderYear(n); math.Abs(m-ladderPartial*ladderFloor) > 1e-9 {
			t.Errorf("year %d matches %.0f k, the rule says 60 %% of the floor", n, m)
		}
	}
}

// The plate borrows three numbers from the household of choisir-sa-strategie
// and one ratio from the ladder article; the two must agree, or the plate
// would contradict a page of the book.
func TestLadderNumbersMatchTheArticles(t *testing.T) {
	strat, err := assets.ReadFile("assets/book/fr/choisir-sa-strategie.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(strat), "plancher 45 000 €, confort 58 000 €, pensions 24 000 €/an dans 13 ans") {
		t.Error("choisir-sa-strategie no longer states the household the plate draws")
	}
	art, err := assets.ReadFile("assets/book/fr/echelle-obligataire.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"plancher 45 000 €, pensions dans 13 ans qui le couvriront à ~53 %",
		"100 % du plancher des années 1-6",
		"60 % des années 7-13",
		"Années 1-2, le fonds euros",
		"fonds à échéance État/IG 2028-2033",
		"Années 9-13, des ETF linkers courts roulés",
		"taux de retrait effectif de ~1,8 %",
		"::: figure echelle-passif",
	} {
		if !strings.Contains(string(art), want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The pensions cover ~53 % of the floor: the article's ratio and the
	// household's 24 k EUR have to be the same fact.
	if got := ladderPension / ladderFloor; math.Abs(got-0.53) > 0.01 {
		t.Errorf("the pensions cover %.1f %% of the floor, the article says ~53 %%", got*100)
	}
	// The article's rungs mature 2028 to 2033, which dates year one.
	if ladderFirstYear+2 != 2028 || ladderFirstYear+7 != 2033 {
		t.Errorf("year 1 is %d, which puts the article's 2028-2033 rungs elsewhere", ladderFirstYear)
	}
	// And the effective rate the article quotes is the plate's own residual:
	// the average bridge draw over the 1,27 M EUR left outside the ladder.
	if got := ladderBridgeDraw(); math.Abs(got-22.69) > 0.01 {
		t.Errorf("the comfort portfolio serves %.2f k a year on average, expected 22,69 k", got)
	}
	if got := ladderBridgeDraw() / 1270 * 100; math.Abs(got-1.8) > 0.05 {
		t.Errorf("effective withdrawal rate %.2f %%, the article says ~1,8 %%", got)
	}
}

// The rendered plate obeys the book's drawing rules and carries every number
// and label it claims to draw.
func TestLadderPlateRenders(t *testing.T) {
	svg := FigureSVG("echelle-passif")
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
	for _, want := range []string{
		">45 k€<", ">58 k€<", ">24 k€/an<", // the three amounts of the prose
		">100 % du plancher adossé<", ">60 % du plancher adossé<",
		">fonds euros<", ">fonds à échéance<", ">ETF linkers courts roulés<", ">les pensions<",
		">2026<", ">2039<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// Every year is drawn, with both of its amounts.
	for n := 1; n <= ladderYears; n++ {
		matched, _, portfolio := ladderYear(n)
		for _, v := range []float64{matched, portfolio} {
			if !strings.Contains(svg, fmt.Sprintf(">%.0f<", v)) {
				t.Errorf("year %d: the plate does not draw the amount %.0f", n, v)
			}
		}
	}
}
