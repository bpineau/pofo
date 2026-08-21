package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// The plate is a closed model: every number it draws is the discounting
// arithmetic of the article's example. This checks that arithmetic against the
// numbers the prose quotes, which is the only thing the plate is allowed to
// say.
func TestFloorLadderArithmetic(t *testing.T) {
	// Fourteen rungs pay the floor once each, so the ladder delivers 560 k of
	// spending; it costs less because the far rungs are discounted.
	if got := float64(floorLadderRungs) * floorLadderFloor; got != 560 {
		t.Errorf("the ladder pays %.0f k in all, expected 560 k", got)
	}
	if got := floorLadderCost(); math.Abs(got-520) > 0.5 {
		t.Errorf("the ladder costs %.2f k, the article says ~520 k", got)
	}
	// The two panels are one balance sheet: what the ladder costs plus what
	// stays invested is the couple's whole capital.
	if got := floorLadderCost() + floorLadderRest(); math.Abs(got-floorLadderWealth) > 1e-9 {
		t.Errorf("the two bars sum to %.2f k, not the %.0f k of capital", got, floorLadderWealth)
	}
	if got := floorLadderRest(); math.Abs(got-1080) > 0.5 {
		t.Errorf("%.2f k stays outside the ladder, the article says 1,08 M", got)
	}
	// The number the whole example exists to produce.
	if got := floorLadderRate(); math.Abs(got-1.4) > 0.05 {
		t.Errorf("the comfort layer asks %.2f %% of the rest, the article says 1,4 %%", got)
	}
	// The rungs run from the dearest (nearest) to the cheapest (furthest),
	// every one of them below the floor it pays.
	for k := 1; k <= floorLadderRungs; k++ {
		c := floorLadderRungCost(k)
		if c >= floorLadderFloor {
			t.Errorf("rung %d costs %.2f k to pay %.0f k, which is not a discount", k, c, floorLadderFloor)
		}
		if k > 1 && c >= floorLadderRungCost(k-1) {
			t.Errorf("rung %d is not cheaper than rung %d", k, k-1)
		}
	}
	if got := floorLadderRungCost(floorLadderRungs); math.Abs(got-34.8) > 0.05 {
		t.Errorf("the last rung costs %.2f k, the caption says 34,8 k", got)
	}
	// The uncovered phase is what dates the hand-over: the last rung pays the
	// year before the pensions do.
	if floorLadderAge+floorLadderRungs != floorLadderPension {
		t.Errorf("%d rungs from age %d end at %d, not at the pensions of %d",
			floorLadderRungs, floorLadderAge, floorLadderAge+floorLadderRungs, floorLadderPension)
	}
}

// The plate borrows its whole household from the article's example; the two
// must agree, or the figure would contradict the page it sits on.
func TestFloorLadderMatchesTheArticle(t *testing.T) {
	art, err := assets.ReadFile("assets/book/fr/obligations-indexees.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Couple, 52 ans, plancher 40 000 €/an, confort 55 000 €, pensions couvrant le plancher à 66 ans",
		"La phase à découvert du plancher dure donc 14 ans",
		"~14 barreaux de ~40 000 € réels",
		"Au taux réel de 1 %, cela coûte ~520 000 € en linkers détenus à terme",
		"(1,08 M€ sur 1,6 M€)",
		"(15 000 €/an, taux de retrait 1,4 % !)",
		"::: figure linkers-plancher-adosse",
	} {
		if !strings.Contains(string(art), want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The English edition adapts the amounts one for one, in dollars.
	en, err := assets.ReadFile("assets/book/en/inflation-linked-bonds.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"a floor of $40,000 a year, comfort at $55,000, pensions covering the floor at 66",
		"$1.08M out of $1.6M",
		"($15,000 a year, a withdrawal rate of 1.4%)",
		"::: figure linkers-plancher-adosse",
	} {
		if !strings.Contains(string(en), want) {
			t.Errorf("the English article no longer says %q", want)
		}
	}
}

// mixHex is the plate's own colour ramp: it must stay on the opaque-hex path
// the book's figures are restricted to, and end exactly on its two anchors.
func TestMixHexStaysOpaqueHex(t *testing.T) {
	if got := mixHex(floorLadderLight, floorLadderDark, 0); got != strings.ToLower(floorLadderLight) {
		t.Errorf("t = 0 gives %q, expected the light anchor", got)
	}
	if got := mixHex(floorLadderLight, floorLadderDark, 1); got != strings.ToLower(floorLadderDark) {
		t.Errorf("t = 1 gives %q, expected the dark anchor", got)
	}
	if got := mixHex("#000000", "#ffffff", 0.5); got != "#808080" {
		t.Errorf("the midpoint of black and white is %q", got)
	}
	seen := map[string]bool{}
	for k := 1; k <= floorLadderRungs; k++ {
		c := floorLadderTint(k)
		if len(c) != 7 || c[0] != '#' {
			t.Errorf("rung %d is painted %q, which is not an opaque hex", k, c)
		}
		if seen[c] {
			t.Errorf("rung %d repeats the tint %q: the correspondence needs one tint per rung", k, c)
		}
		seen[c] = true
	}
}

// The rendered plate obeys the book's drawing rules and carries every number
// and label it claims to draw.
func TestFloorLadderPlateRenders(t *testing.T) {
	svg := FigureSVG("linkers-plancher-adosse")
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
		">520 k€<", ">1,08 M€<", // the two halves of the balance sheet
		">le confort, 15 k€ par an<", ">15/1080 = 1,4 % du portefeuille<",
		">échelle indexée<", ">portefeuille<", ">de confort<",
		">les pensions<", ">prennent le relais<",
		">14 barreaux de 40 k€ réels<",
		">52<", ">75<", // the window of the right panel
		">55<", ">40<", // the comfort level and the floor
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// A shape carries its fill at the very end of its tag, where a label
	// carries it in the middle, so this counts blocks and never words.
	shapes := func(fill string) int {
		return strings.Count(svg, fmt.Sprintf(`fill="%s"/>`, fill))
	}
	// Every rung is drawn three times in its own tint: what it costs on the
	// left, the year it pays on the right, and its cell of the legend ramp.
	for k := 1; k <= floorLadderRungs; k++ {
		if n := shapes(floorLadderTint(k)); n != 3 {
			t.Errorf("rung %d is painted %d times, expected the stock, the flow and the legend", k, n)
		}
	}
	// The pension years wear the book's green, the comfort layer its amber:
	// one green block per year past the ladder, one amber layer per year of
	// the window, plus the amber bar of the portfolio itself.
	years := floorLadderLast - floorLadderAge + 1
	if n := shapes(figGreen); n != years-floorLadderRungs {
		t.Errorf("%d pension years drawn, expected %d", n, years-floorLadderRungs)
	}
	if n := shapes(figAccent); n != years+1 {
		t.Errorf("%d amber blocks drawn, expected %d comfort layers plus the portfolio bar", n, years+1)
	}
}
