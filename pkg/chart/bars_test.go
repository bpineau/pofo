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

func TestBarsYAxisTicksAndValueLabels(t *testing.T) {
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
	if strings.Count(svg, `stroke="#e6e6e6"`) < 2 {
		t.Errorf("expected y-axis gridlines, got %d", strings.Count(svg, `stroke="#e6e6e6"`))
	}
}

func TestBarsEmpty(t *testing.T) {
	svg := Bars(Options{}, nil)
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("empty input must still yield a valid SVG")
	}
}
