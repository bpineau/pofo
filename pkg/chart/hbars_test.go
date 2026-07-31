package chart

import (
	"strings"
	"testing"
)

func TestHBarsSignedColors(t *testing.T) {
	bars := []Bar{
		{Label: "Spend -5k", Value: -3.2, Text: "-3.2pp"},
		{Label: "Horizon -5y", Value: -1.1, Text: "-1.1pp"},
		{Label: "Spend +5k", Value: 4.0, Text: "+4.0pp"},
	}

	svg := HBars(Options{Title: "Sensitivity"}, bars)

	if !strings.HasPrefix(svg, "<svg") {
		t.Fatalf("not an SVG: %.20q", svg)
	}
	if !strings.Contains(svg, "#0C8A47") {
		t.Errorf("expected green for ruin-reducing (negative) bars")
	}
	if !strings.Contains(svg, "#D2402F") {
		t.Errorf("expected red for ruin-increasing (positive) bars")
	}
	if !strings.Contains(svg, "Spend -5k") || !strings.Contains(svg, "-3.2pp") {
		t.Errorf("expected row labels and value text")
	}
}

func TestHBarsEmptyIsSafe(t *testing.T) {
	if svg := HBars(Options{}, nil); !strings.HasPrefix(svg, "<svg") {
		t.Errorf("empty HBars should still return an SVG, got %.20q", svg)
	}
}
