package firebook

import (
	"math"
	"strings"
	"testing"
)

// The plate is a closed model, so the formula is the fixture: the three
// readings it annotates and the two stretches it braces are all recomputed
// here, and confronted with the numbers the article's prose quotes.
func TestAmortRateMatchesTheArticle(t *testing.T) {
	for _, c := range []struct {
		n    float64
		want float64
	}{
		{10, 0.12329}, {30, 0.05783}, {50, 0.04655},
	} {
		if got := amortRate(amortReal, c.n) / 100; math.Abs(got-c.want) > 5e-5 {
			t.Errorf("%.0f years funds %.5f, expected %.5f", c.n, got, c.want)
		}
	}
	// The prose of tier two, word for word: 5,8 % over thirty years for a
	// bonus of +1,8, and 4,7 % over fifty for a bonus of +0,7.
	art := bookArticle(t, "les-maths-du-4-pourcent")
	for _, want := range []string{
		"à 4 % réels sur 30 ans, le retrait qui épuise exactement le capital au dernier jour est d'environ 5,8 % par an",
		"Le droit de finir à zéro vaut donc +1,8 point",
		"À 50 ans d'horizon, le même calcul donne 4,7 % et le bonus fond à +0,7",
		"À l'infini, il disparaît",
		"::: figure amortissement-horizon",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	en, err := assets.ReadFile("assets/book/en/the-math-of-4-percent.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"at 4% real over 30 years, the withdrawal that lands exactly on zero on the last day is about 5.8% a year",
		"Stretch it to 50 years and the same formula gives 4.7%: the bonus melts to +0.7",
		"::: figure amortissement-horizon",
	} {
		if !strings.Contains(string(en), want) {
			t.Errorf("the English article no longer says %q", want)
		}
	}
	// The bonus the article names is the plate's own gap to the asymptote.
	if got := amortForever(30); math.Abs(got-1.8) > 0.05 {
		t.Errorf("the 30-year bonus is %.2f points, the article says +1,8", got)
	}
	if got := amortForever(50); math.Abs(got-0.7) > 0.05 {
		t.Errorf("the 50-year bonus is %.2f points, the article says +0,7", got)
	}
}

// The two braces are the plate's argument, and the closing line its
// conclusion: both are computed, never typed.
func TestAmortBracesCarryTheArgument(t *testing.T) {
	late := amortForever(50)       // fifty years to a perpetuity
	middle := amortGap(30, 50)     // thirty years to fifty
	early := amortGap(10, 30)      // ten years to thirty
	if frNum(middle, 1) != "1,1" { // the first brace
		t.Errorf("the 30 to 50 stretch costs %s points, the plate braces 1,1", frNum(middle, 1))
	}
	if frNum(late, 1) != "0,7" { // the second brace
		t.Errorf("the 50 to forever stretch costs %s points, the plate braces 0,7", frNum(late, 1))
	}
	if frNum(early, 1) != "6,5" { // the closing line
		t.Errorf("the 10 to 30 stretch costs %s points, the plate says 6,5", frNum(early, 1))
	}
	// The shape of the whole plate: each further stretch of horizon costs
	// less than the one before, which is the counterintuitive part.
	if !(early > middle && middle > late) {
		t.Errorf("the stretches cost %.2f, %.2f then %.2f: the curve no longer flattens",
			early, middle, late)
	}
	// The asymptote is the real return itself, approached and never reached.
	if got := amortRate(amortReal, 1e6); math.Abs(got-amortReal) > 1e-6 {
		t.Errorf("an endless horizon funds %.6f %%, expected the real return %.1f %%", got, amortReal)
	}
	for n := 10.0; n < 200; n++ {
		if amortRate(amortReal, n) <= amortReal {
			t.Errorf("a %.0f-year horizon funds no more than the perpetuity", n)
			break
		}
	}
	// A zero real return is the degenerate case the formula must not divide
	// through: spending one nth of the capital every year.
	if got := amortRate(0, 25); math.Abs(got-4) > 1e-9 {
		t.Errorf("at 0 %% real over 25 years the plan funds %.2f %%, expected 4 %%", got)
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestAmortPlateRenders(t *testing.T) {
	svg := FigureSVG("amortissement-horizon")
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
		">12,3 %<", ">5,8 %<", ">4,7 %<", // the three readings
		">10 ans<", ">30 ans<", ">50 ans<",
		">30 → 50 ans : −1,1 point<", ">50 ans → toujours : −0,7 point<",
		">∞<", // the perpetuity, which the axis names
		">4 % : la perpétuité, le rendement seul<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// One curve, one perpetuity marker and three read dots: nothing else.
	if n := strings.Count(svg, "<circle"); n != 4 {
		t.Errorf("%d circles drawn, expected the three readings plus the perpetuity", n)
	}
	if n := strings.Count(svg, "<polyline"); n != 1 {
		t.Errorf("%d curves drawn, the plate has exactly one", n)
	}
}
