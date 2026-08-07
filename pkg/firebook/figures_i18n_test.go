package firebook

import (
	"strings"
	"testing"
)

func TestAnglicizeNumbers(t *testing.T) {
	cases := map[string]string{
		"6,6 %":       "6.6%",
		"1 000 000":   "1,000,000",
		"1 000":       "1,000",
		"1966-1995":   "1966-1995",
		"+4,1 % / an": "+4.1% / an", // words untouched: dictionary's job
		"7,6×":        "7.6×",
		"0":           "0",
	}
	for in, want := range cases {
		if got := anglicizeNumbers(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestIsNeutralPayload(t *testing.T) {
	for _, s := range []string{"7,6×", "0", "1966-1995", "+27 % / −13 %", "(4,5 %)"} {
		if !isNeutralPayload(s) {
			t.Errorf("%q should be neutral", s)
		}
	}
	for _, s := range []string{"années  →", "richesse", "", "an"} {
		if isNeutralPayload(s) {
			t.Errorf("%q should not be neutral", s)
		}
	}
}

func TestFigureSVGEnglish(t *testing.T) {
	// vol-drag is small and stable; pin one dictionary entry end to end.
	fr, en := FigureSVG("vol-drag"), FigureSVGEnglish("vol-drag")
	if fr == en {
		t.Error("translation pass changed nothing")
	}
	if !strings.Contains(en, "Volatility drag: same average, opposite outcomes") {
		t.Error("the dictionary title did not reach the SVG")
	}
	for _, node := range figureTextNodes(en) {
		if hasFrenchDecimal(node) {
			t.Errorf("French decimal survived: %q", node)
		}
		if reFrenchPercent.MatchString(node) {
			t.Errorf("French percent spacing survived: %q", node)
		}
	}
	// Structure is untouched: the pass only rewrites text payloads.
	if len(figureTextNodes(fr)) != len(figureTextNodes(en)) {
		t.Error("the translation pass changed the node count")
	}
	if strings.Count(fr, "<path") != strings.Count(en, "<path") {
		t.Error("the translation pass touched the drawing")
	}
}

// The English edition renders through the translation pass.
func TestEnglishEditionUsesTheFigurePass(t *testing.T) {
	if English.Figure == nil {
		t.Fatal("English.Figure is nil: figure blocks would silently vanish")
	}
	if English.Figure("vol-drag") != FigureSVGEnglish("vol-drag") {
		t.Error("English.Figure is not the English pass")
	}
}

func TestHasFrenchDecimal(t *testing.T) {
	for _, s := range []string{"3,3 %", "27,0 k€", "a 4,6 M"} {
		if !hasFrenchDecimal(s) {
			t.Errorf("%q carries a French decimal", s)
		}
	}
	// English thousands separators are commas between digits too.
	for _, s := range []string{"EUR 1,243k", "1,000,000", "1966-1995", "no digits"} {
		if hasFrenchDecimal(s) {
			t.Errorf("%q is fine English", s)
		}
	}
}
