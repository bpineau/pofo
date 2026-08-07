package chart

import (
	"strings"
	"testing"
)

func TestScatter(t *testing.T) {
	pts := []LabeledPoint{
		{X: 2, Y: 14, Label: "Fixed", Color: "#D2402F"},
		{X: 18, Y: 1, Label: "VPW", Color: "#0B7285"},
	}
	svg := Scatter(Options{Width: 640, Height: 360}, "volatility", "ruin", pts)
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("not a well-formed svg")
	}
	for _, want := range []string{"Fixed", "VPW", "volatility", "ruin", "<circle", "stroke-dasharray"} {
		if !strings.Contains(svg, want) {
			t.Errorf("scatter missing %q", want)
		}
	}
}

func TestScatterEmpty(t *testing.T) {
	svg := Scatter(Options{}, "x", "y", nil)
	if !strings.Contains(svg, "<svg") {
		t.Error("empty scatter should still be a valid svg")
	}
}
