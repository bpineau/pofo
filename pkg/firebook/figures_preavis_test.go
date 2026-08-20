package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/replay"
)

// Both registers are a snapshot of the bundled record, so this is what keeps
// them honest: the replay is run again and every capital and every rate is
// recomputed, along with the two dated markers and the notice they measure.
// Gated with the other data-frozen checks, since a refreshed dataset is allowed
// to move a plate without breaking the build.
func TestPreavisMatchesTheReplay(t *testing.T) {
	frozenAgainstData(t)
	res, err := replay.Run(replay.Setup{
		Start: preavisStart, Capital: preavisCapital0 * 1000, Spend: preavisCheque * 1000, Years: 30,
		Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	rule := res.Rules[0] // the fixed inflation-indexed withdrawal
	if rule.NameFR != "Retrait fixe" {
		t.Fatalf("the plate reads rule %q, not the fixed withdrawal", rule.NameFR)
	}
	if !rule.Ruined || rule.RuinYear != preavisRuinYear {
		t.Errorf("the replay runs out in %d (ruined=%v), the plate says %d",
			rule.RuinYear, rule.Ruined, preavisRuinYear)
	}
	// The capital the plan opened each year with: the starting million, then
	// every year end the replay reports.
	capital := preavisCapital0 * 1000
	for i, want := range preavisCapital {
		if got := capital / 1000; math.Abs(got-want) > 0.05 {
			t.Errorf("%d: the replay opens on %.1f k, the plate draws %.1f k",
				preavisStart+i, got, want)
		}
		if i < len(preavisRate) {
			rate := rule.Spend[i] / capital * 100
			if math.Abs(rate-preavisRate[i]) > 0.01 {
				t.Errorf("%d: the replay withdraws %.2f %% of what is left, the plate draws %.2f %%",
					preavisStart+i, rate, preavisRate[i])
			}
		}
		capital = rule.Wealth[i]
	}
}

// The measurement the plate is built on, and the reason the articles' "8 to 15
// years" had to move: on the worst vintage the record holds, every threshold of
// the book's own dashboard fires two decades before the money runs out.
func TestPreavisMeasuresTwentyYearsOfNotice(t *testing.T) {
	if got := preavisFirstRed(); got != 1974 {
		t.Errorf("the light first reads red in %d, the plate marks 1974", got)
	}
	if got := preavisConfirmedRed(); got != 1975 {
		t.Errorf("the red is confirmed in %d, the plate marks 1975", got)
	}
	if got := preavisNotice(); got != 20 {
		t.Errorf("the notice is %d years, the plate cotes 20", got)
	}
	if got := preavisExit(); got != 1979 {
		t.Errorf("the staircase leaves the panel in %d, the annotation says 1979", got)
	}
	// The amber threshold fires earlier still, and never clears again: the two
	// readings the corrected prose rests on.
	firstAmber, lastGreen := 0, 0
	for i, v := range preavisRate {
		year := preavisStart + i
		if v >= preavisGreenMax && firstAmber == 0 {
			firstAmber = year
		}
		if v < preavisGreenMax {
			lastGreen = year
		}
	}
	if firstAmber != 1967 || preavisRuinYear-firstAmber != 27 {
		t.Errorf("the first amber lands in %d, %d years ahead: the measurement says 1967 and 27",
			firstAmber, preavisRuinYear-firstAmber)
	}
	if lastGreen != 1969 || preavisRuinYear-(lastGreen+1) != 24 {
		t.Errorf("the light is green for the last time in %d: the measurement says 1969", lastGreen)
	}
	// Shape: the plan starts at the rule's own rate, the red never clears, and
	// the capital ends empty.
	if preavisRate[0] != 4.0 {
		t.Errorf("the plan starts at %.2f %%, not at the 4 %% of the rule", preavisRate[0])
	}
	for i, v := range preavisRate {
		if year := preavisStart + i; year > preavisFirstRed() && v <= preavisOrangeMax {
			t.Errorf("%d: the rate fell back to %.2f %%, under the red threshold it had crossed", year, v)
		}
	}
	if len(preavisCapital) != len(preavisRate)+1 {
		t.Fatalf("%d capitals for %d rates: the last year has nothing left to divide",
			len(preavisCapital), len(preavisRate))
	}
	if preavisCapital[0] != preavisCapital0 || preavisCapital[len(preavisCapital)-1] != 0 {
		t.Error("the capital path no longer runs from the opening million to zero")
	}
}

// The bands are the article's own scale, quoted rather than invented.
func TestPreavisQuotesTheDashboardThresholds(t *testing.T) {
	art := bookArticle(t, "quand-s-inquieter")
	for _, want := range []string{
		"**Vert** (< ~4,3 %",
		"**Orange** (~4,3-5,2 %)",
		"**Rouge** (> ~5,2 % confirmé sur deux points espacés",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the dashboard article no longer sets %q", want)
		}
	}
	// And the corrected claim, which this plate is the evidence for.
	for _, want := range []string{
		"la ruine prévient 10 à 20 ans à l'avance",
		"C'est un processus de 10 à 20 ans",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the dashboard article no longer says %q", want)
		}
	}
	host := bookArticle(t, "ruine-et-probabilites")
	for _, want := range []string{
		"::: figure preavis-1966",
		"et vingt ans sur le pire millésime connu",
		"Le rouge du millésime 1966 se confirme en 1974-1975 ; le capital ne s'épuise qu'en 1994.",
	} {
		if !strings.Contains(host, want) {
			t.Errorf("the host article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestPreavisPlateRenders(t *testing.T) {
	svg := FigureSVG("preavis-1966")
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
		">Le capital réel<", ">Le voyant<",
		">vert : sous 4,3 %<", ">orange : 4,3 à 5,2 %<", ">rouge : au-delà de 5,2 %, confirmé<",
		">4,3 %<", ">5,2 %<", ">9 %<", ">1 000<",
		">premier rouge, 1974<", ">670 k€<",
		">épuisement<", ">1994<", ">1966<", ">1974<",
		">préavis : 20 ans<",
		">confirmé en 1975, hors de l'échelle dès 1979<",
		">le voyant ne repasse plus jamais au vert<",
		">Le plan condamné a prévenu 20 ans à l'avance : premier rouge en 1974, dernière ligne du compte en 1994.<",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	// The three bands and the three legend chips that name them, the filled
	// capital path, and no other filled shape.
	if n := strings.Count(svg, "<rect"); n != 6 {
		t.Errorf("%d rectangles drawn, expected the three bands and their three chips", n)
	}
	if n := strings.Count(svg, "<path"); n != 1 {
		t.Errorf("%d filled paths drawn, expected the capital area alone", n)
	}
	// The capital line and the rate staircase, and nothing else drawn as a run
	// of segments.
	if n := strings.Count(svg, "<polyline"); n != 2 {
		t.Errorf("%d polylines drawn, the plate has the capital path and the staircase", n)
	}
	// One dot at each end of the story, plus the two readings that confirm the
	// red: the exhaustion, the first red on the capital path, and the pair.
	if n := strings.Count(svg, "<circle"); n != 4 {
		t.Errorf("%d dots drawn, expected the four the plate poses", n)
	}
	// The two dated rules through both registers, the off-scale rail, and the
	// staircase never leaves its panel.
	if n := strings.Count(svg, "stroke-dasharray"); n != 3 {
		t.Errorf("%d dashed strokes drawn, expected the two dated rules and the rail", n)
	}
}
