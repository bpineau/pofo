package firebook

import (
	"math"
	"strings"
	"testing"
)

// The whole plate is an addition, so the addition is the fixture: the three
// levels are quoted by the article, the two middle steps are derived from
// them, and the chain has to land exactly on the plan the article states. A
// waterfall that does not close is a lie the eye cannot catch.
func TestStack22ChainClosesOnTheArticlesNumbers(t *testing.T) {
	art := bookArticle(t, "return-stacking")
	for _, want := range []string{
		"67 % en fonds 90/60",
		"18 % trend, 10 % or, 5 % cash",
		"Le cœur empilé y perd environ 21 % là où le 60/40 nu perd 17 %",
		"le plan complet ne s'en tire à environ −10 % que parce que le trend a très bien travaillé cette année-là",
		"::: figure stacking-2022",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	sum := stack22Naked + stack22Lever() + stack22Dilution() + stack22Sleeves()
	if math.Abs(sum-stack22Plan) > 1e-9 {
		t.Errorf("the waterfall lands on %.3f %%, the article says %.0f %%", sum, stack22Plan)
	}
	// The two derived steps, to the tenth the plate prints: the leverage bill
	// and what the core's 67 % weight gives back.
	if got := stack22Lever(); math.Abs(got+4) > 1e-9 {
		t.Errorf("the leverage bill is %.2f points, expected −4", got)
	}
	if got := stack22Dilution(); math.Abs(got-6.93) > 1e-9 {
		t.Errorf("the core's weight gives back %.2f points, expected 6,93", got)
	}
	if got := stack22Sleeves(); math.Abs(got-4.07) > 1e-9 {
		t.Errorf("the freed layer contributes %.2f points, expected 4,07", got)
	}
	// Rounded to the tenth, as drawn, the chain still closes: no rounding of
	// the plate's own labels may open a gap.
	rounded := stack22Naked
	for _, v := range []float64{stack22Lever(), stack22Dilution(), stack22Sleeves()} {
		rounded += math.Round(v*10) / 10
	}
	if math.Abs(rounded-stack22Plan) > 1e-9 {
		t.Errorf("the drawn labels add up to %.2f %%, not to %.0f %%", rounded, stack22Plan)
	}
	// The shape of the argument: the leverage costs, and the freed layer more
	// than pays for it.
	if stack22Lever() >= 0 {
		t.Error("the leverage step no longer costs anything")
	}
	if stack22Dilution()+stack22Sleeves() <= -stack22Lever() {
		t.Error("the freed layer no longer pays the leverage bill")
	}
	// The English edition carries the same three levels.
	en, err := assets.ReadFile("assets/book/en/return-stacking.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"The stacked core loses about 21% there, where the naked 60/40 loses 17%",
		"the whole plan gets away with about −10% only because trend worked very well that year",
		"::: figure stacking-2022",
	} {
		if !strings.Contains(string(en), want) {
			t.Errorf("the English article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every step of
// the chain, its two states, and the sentence it exists to prove.
func TestStack22PlateRenders(t *testing.T) {
	svg := FigureSVG("stacking-2022")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"\u2014", "\u2013", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, want := range []string{
		">−17,0 %<", ">−10,0 %<", // the two states
		">−4,0<", ">+6,9<", ">+4,1<", // the three steps
		">60/40 nu<", ">la facture<", ">le cœur ne pèse<", ">l'étage libéré<", ">le plan empilé<",
		">Le levier pique, et c'est l'étage libéré qui paie la facture.<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// Negative numbers use the book's typographic minus, never a hyphen.
	if strings.Contains(svg, ">-") {
		t.Error("the plate prints an ASCII hyphen where the book uses a minus sign")
	}
	// One bar per step, and one hand-over dash between consecutive bars.
	if n := strings.Count(svg, "<path"); n != len(stack22Steps()) {
		t.Errorf("%d bars drawn, expected one per step (%d)", n, len(stack22Steps()))
	}
	if n := strings.Count(svg, `stroke-dasharray="3 3"`); n != len(stack22Steps())-1 {
		t.Errorf("%d hand-overs drawn, expected one between each pair of bars", n)
	}
}
