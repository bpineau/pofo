package firebook

import (
	"strings"
	"testing"
)

// The stock-bond correlation plate is the one place in the book where a value
// axis runs through zero and labels both ends, so it is also the one place a
// negative tick can quietly fall back to Go's ASCII hyphen. It must set the
// book's typographic minus like every other negative number on a plate, in
// both editions: the tick is a wordless payload, so the English render takes
// it through the mechanical number reformat rather than the dictionary, and
// the sign has to survive that trip too.
func TestFigCorrelSign(t *testing.T) {
	fr := FigureSVG("correl-sign")
	for _, want := range []string{
		">+0,5<", ">" + figMinus + "0,5<", // the two ticks, signed at both ends
		">corrélation 0<",
		">La corrélation actions / obligations change de signe<",
	} {
		if !strings.Contains(fr, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	en := FigureSVGEnglish("correl-sign")
	for _, want := range []string{">+0.5<", ">" + figMinus + "0.5<"} {
		if !strings.Contains(en, want) {
			t.Errorf("the English render does not draw %q", want)
		}
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them. The ASCII hyphen is banned in front of a number for a
	// different reason: it is narrower and lighter than the digits it signs.
	for _, svg := range []string{fr, en} {
		for _, banned := range []string{"\u2014", "\u2013", ">-"} {
			if strings.Contains(svg, banned) {
				t.Errorf("the plate uses %q, which the book's figures never do", banned)
			}
		}
	}
}
