package chart

import (
	"strings"
	"testing"
)

func TestBarsSVG(t *testing.T) {
	svg := Bars(Options{Title: "Recovery"}, []Bar{
		{Label: "0y", Value: 0.4}, {Label: "1y", Value: 0.3}, {Label: "2y", Value: 0.3},
	})
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("not an SVG: %.20q", svg)
	}
	if !strings.Contains(svg, "Recovery") {
		t.Errorf("title missing")
	}
	if strings.Count(svg, "<rect") < 3 {
		t.Errorf("expected at least 3 bars")
	}
}

// Value labels and a y-axis grid are alternatives, never both: bars that all
// carry their own number drop the axis, bars that do not get it back.
func TestBarsValueLabelsReplaceTheYAxis(t *testing.T) {
	bars := []Bar{
		{Label: "0y", Value: 40, Text: "40%"},
		{Label: "1y", Value: 25, Text: "25%"},
		{Label: "2y", Value: 10, Text: "10%"},
	}
	svg := Bars(Options{Title: "Recovery"}, bars)
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("malformed SVG document")
	}
	for _, want := range []string{"40%", "25%", "10%"} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG missing value label %q", want)
		}
	}
	if n := strings.Count(svg, `stroke="`+themeGrid+`"`); n != 0 {
		t.Errorf("fully labelled bars should carry no gridline, got %d", n)
	}

	bars[1].Text = "" // one label missing: the axis has to carry the reading
	svg = Bars(Options{Title: "Recovery"}, bars)
	if n := strings.Count(svg, `stroke="`+themeGrid+`"`); n < 2 {
		t.Errorf("expected y-axis gridlines, got %d", n)
	}
}

// A bar sits on the baseline: its tip is rounded, its foot is not.
func TestBarsSitOnTheBaseline(t *testing.T) {
	svg := Bars(Options{Width: 400, Height: 200}, []Bar{{Label: "a", Value: 1}})
	if !strings.Contains(svg, "<path d=\"M") {
		t.Error("bars should be drawn as paths, so the tip can round while the foot stays square")
	}
	if strings.Contains(svg, `rx="`) {
		t.Error("no capsule bars: rounding is per-corner, not a uniform radius")
	}
}

func TestBarsEmpty(t *testing.T) {
	svg := Bars(Options{}, nil)
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("empty input must still yield a valid SVG")
	}
}
