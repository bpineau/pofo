package firebook

import (
	"math"
	"strings"
	"testing"
)

// The kit is drawn on by every plate, so a silent change here would move a
// hundred figures at once. These check the conventions the plates rely on.

func TestFigScaleMaps(t *testing.T) {
	s := figScale{Min: 0, Max: 10, Px0: 100, Px1: 500}
	for v, want := range map[float64]float64{0: 100, 5: 300, 10: 500, 15: 700} {
		if got := s.Map(v); math.Abs(got-want) > 1e-9 {
			t.Errorf("%.0f maps to %.1f, expected %.1f", v, got, want)
		}
	}
	// A vertical axis passes the bottom pixel first, which inverts it.
	up := figScale{Min: 0, Max: 10, Px0: 300, Px1: 100}
	if got := up.Map(10); got != 100 {
		t.Errorf("the top of an inverted scale lands at %.1f, expected 100", got)
	}
	// A degenerate range must not divide by zero.
	if got := (figScale{Min: 4, Max: 4, Px0: 7, Px1: 9}).Map(4); got != 7 {
		t.Errorf("a flat scale maps to %.1f, expected its first pixel", got)
	}
}

func TestAxisTicksConventions(t *testing.T) {
	s := figScale{Min: -2, Max: 6, Px0: 300, Px1: 100}
	var b strings.Builder
	axisTicks(&b, s, []float64{-2, 0, 2, 4, 6}, 0, " %", 80, 560, false)
	out := b.String()
	if n := strings.Count(out, "<line"); n != 5 {
		t.Errorf("%d gridlines for 5 ticks", n)
	}
	// Zero is a fact, not a guide: it takes the rule colour, the others the
	// hairline one.
	if n := strings.Count(out, figRule); n != 1 {
		t.Errorf("%d ticks drawn in the rule colour, expected the zero alone", n)
	}
	if n := strings.Count(out, figGrid); n != 4 {
		t.Errorf("%d hairline gridlines, expected four", n)
	}
	// A vertical axis labels to the left of the plot, right-aligned.
	if !strings.Contains(out, `text-anchor="end"`) || !strings.Contains(out, ">0 %<") {
		t.Error("the labels are not set to the left of a vertical axis")
	}
	// Passing one extent twice draws the labels alone.
	var bare strings.Builder
	axisTicks(&bare, figScale{Min: 0, Max: 10, Px0: 80, Px1: 560}, []float64{0, 5, 10}, 0, "", 300, 300, true)
	if strings.Contains(bare.String(), "<line") {
		t.Error("an axis given no extent still ruled the plate")
	}
	if !strings.Contains(bare.String(), `text-anchor="middle"`) {
		t.Error("a horizontal axis does not centre its labels under the ticks")
	}
}

// The plates are single-source French: a tick label carries a decimal COMMA and
// the book's typographic minus, never a point or an ASCII hyphen. The English
// edition reformats them mechanically, which only works from French input.
func TestAxisTicksLabelsSpeakFrench(t *testing.T) {
	var b strings.Builder
	axisTicks(&b, figScale{Min: -2, Max: 2, Px0: 100, Px1: 500},
		[]float64{-1.5, 0, 1.5}, 1, " %", 60, 60, true)
	out := b.String()
	for _, want := range []string{">" + figMinus + "1,5 %<", ">0,0 %<", ">1,5 %<"} {
		if !strings.Contains(out, want) {
			t.Errorf("the axis does not carry %q", want)
		}
	}
	if strings.Contains(out, "1.5") || strings.Contains(out, ">-") {
		t.Error("a tick label fell back to a point decimal or an ASCII hyphen")
	}
	// The suffix is appended verbatim, and an empty one leaves the bare number.
	var bare strings.Builder
	axisTicks(&bare, figScale{Min: 0, Max: 10, Px0: 100, Px1: 500},
		[]float64{5}, 0, "", 60, 60, true)
	if !strings.Contains(bare.String(), ">5<") {
		t.Error("an axis given no suffix still dressed its labels")
	}
}

func TestPlateFootStacksItsLines(t *testing.T) {
	out := plateFoot(400, []string{"une convention", "une réserve"})
	if n := strings.Count(out, "<text"); n != 2 {
		t.Errorf("%d lines drawn for two notes", n)
	}
	for _, want := range []string{`y="400.0"`, `y="415.0"`, `x="24.0"`} {
		if !strings.Contains(out, want) {
			t.Errorf("the notes block does not carry %q", want)
		}
	}
	if plateFoot(400, nil) != "" {
		t.Error("an empty notes block still drew something")
	}
}

func TestBraceHMeasuresAStretch(t *testing.T) {
	out := braceH(100, 300, 350, "30 → 50 ans")
	if n := strings.Count(out, "<line"); n != 3 {
		t.Errorf("%d lines: a brace is its span and two ticks", n)
	}
	// The label sits centred under the span.
	if !strings.Contains(out, `x="200.0"`) || !strings.Contains(out, ">30 → 50 ans<") {
		t.Error("the brace does not centre its label under the stretch it measures")
	}
}

func TestDotLabelNudgesAwayFromItsMark(t *testing.T) {
	right := dotLabel(200, 150, figBlue, "actions", "5,9 %", "start")
	if !strings.Contains(right, `cx="200.0"`) || !strings.Contains(right, `x="208.0"`) {
		t.Error("a start-anchored label is not set to the right of its dot")
	}
	left := dotLabel(200, 150, figBlue, "actions", "5,9 %", "end")
	if !strings.Contains(left, `x="192.0"`) {
		t.Error("an end-anchored label is not set to the left of its dot")
	}
	mid := dotLabel(200, 150, figBlue, "actions", "5,9 %", "middle")
	if !strings.Contains(mid, `x="200.0"`) {
		t.Error("a middle-anchored label is not set straight above its dot")
	}
	// The value wears the mark's colour, the name the book's soft ink.
	if !strings.Contains(mid, `fill="`+figSoft+`"`) || strings.Count(mid, figBlue) != 2 {
		t.Error("the dot and its value do not share one colour, with the name in soft ink")
	}
}

func TestFrMinusUsesTheTypographicSign(t *testing.T) {
	if got := frMinus(-1.7, 1); got != figMinus+"1,7" {
		t.Errorf("frMinus(-1.7) = %q", got)
	}
	if got := frMinus(4, 0); got != "4" {
		t.Errorf("frMinus(4) = %q, a positive number keeps no sign", got)
	}
	if strings.Contains(frMinus(-1.7, 1), "-") {
		t.Error("frMinus fell back to an ASCII hyphen")
	}
}
