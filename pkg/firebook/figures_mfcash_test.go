package firebook

import (
	"math"
	"strings"
	"testing"
)

// The plate exists to show one addition, so the addition is the fixture: each
// regime's three bricks must produce the net the article prints, and the two
// ends of the article's rate range must produce the two ends of the net range
// it quotes. If any of that stopped holding, the plate would be drawing an
// arithmetic the prose does not make.
func TestMfRegimesDecomposeIntoTheStatedNet(t *testing.T) {
	art := bookArticle(t, "managed-futures")
	for _, want := range []string{
		"Un fonds trend rapporte donc « cash + prime du trend − frais »",
		"À taux courts nuls (2015-2021), un programme à prime brute de 3 % et frais de 1 % affichait ~2 %",
		"À taux courts de 3-4 %, le même programme affiche 5-6 % sans avoir rien changé",
		"doit donc se faire en excess return, au-dessus du cash",
		"::: figure mf-cash-prime-frais",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	if got := mfZIRP.mfNet(); math.Abs(got-2) > 1e-9 {
		t.Errorf("the zero-rate regime nets %.2f %%, the article says ~2 %%", got)
	}
	// The high-rate column draws the midpoint of the article's 3-4 % range,
	// and that range's two ends are exactly the 5-6 % the article quotes.
	if got := mfHigh.cash; math.Abs(got-(mfRateLo+mfRateHi)/2) > 1e-9 {
		t.Errorf("the drawn collateral is %.2f %%, not the midpoint of %.0f-%.0f %%", got, mfRateLo, mfRateHi)
	}
	if got := mfHigh.mfNet(); math.Abs(got-5.5) > 1e-9 {
		t.Errorf("the high-rate regime nets %.2f %%, expected the 5,5 %% midpoint", got)
	}
	for _, c := range []struct{ rate, net float64 }{{mfRateLo, 5}, {mfRateHi, 6}} {
		if got := (mfRegime{cash: c.rate, gross: mfHigh.gross, fees: mfHigh.fees}).mfNet(); math.Abs(got-c.net) > 1e-9 {
			t.Errorf("at %.0f %% of short rates the programme nets %.2f %%, the article says %.0f %%", c.rate, got, c.net)
		}
	}
	// Only the collateral moves between the two columns: that identity is the
	// entire point, and the plate would lie without it.
	if mfZIRP.gross != mfHigh.gross || mfZIRP.fees != mfHigh.fees {
		t.Error("the two columns no longer draw the same programme")
	}
	if mfZIRP.cash != 0 {
		t.Errorf("the zero-rate regime pays %.2f %% on collateral", mfZIRP.cash)
	}
	// The English edition carries the same arithmetic, in its own words.
	en, err := assets.ReadFile("assets/book/en/managed-futures.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"a program earning a 3% gross premium and charging 1% showed about 2%",
		"With short rates at 3% to 4%, the same program shows 5% to 6% without changing a thing",
		"::: figure mf-cash-prime-frais",
	} {
		if !strings.Contains(string(en), want) {
			t.Errorf("the English article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules, keeps the two columns
// strictly comparable, and never turns into a cascade.
func TestMfPlateRenders(t *testing.T) {
	svg := FigureSVG("mf-cash-prime-frais")
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
		">+3,0<", ">−1,0<", ">+3,5<", // the bricks, signed
		">2,0 %<", ">5,5 %<", // the two nets
		">taux courts à 0 %<", ">taux courts à 3-4 %<",
		">2015-2021<", ">2022 →<",
		">collatéral rémunéré<", ">prime de trend brute<", ">frais<", ">net<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// Both columns draw the same gross premium, so the value appears twice,
	// as do the fees: that repetition IS the comparability.
	for _, shared := range []string{">+3,0<", ">−1,0<"} {
		if n := strings.Count(svg, shared); n != 2 {
			t.Errorf("%q appears %d times, expected once per column", shared, n)
		}
	}
	// Five bricks (the zero-rate column has no collateral bar to draw) and two
	// nets, counted by their widths so the legend chips do not enter. Nothing
	// connects them, which is what keeps the plate a signed decomposition
	// rather than a cascade.
	if n := strings.Count(svg, `width="110.0"`); n != 5 {
		t.Errorf("%d bricks drawn, expected two for the zero-rate column and three for the other", n)
	}
	if n := strings.Count(svg, `width="70.0"`); n != 2 {
		t.Errorf("%d net bars drawn, expected one per regime", n)
	}
	if strings.Contains(svg, `stroke-dasharray="3 3"`) {
		t.Error("the plate draws connectors: it is a decomposition, not a cascade")
	}
}
