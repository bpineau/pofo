package chart

import (
	"strings"
	"testing"
)

func TestStackedAreaRendersLayers(t *testing.T) {
	series := []AreaSeries{
		{Name: "Funded", Values: []float64{100, 90, 80}, Color: "#0E9384"},
		{Name: "Broke", Values: []float64{0, 5, 10}, Color: "#D92D20"},
		{Name: "Dead", Values: []float64{0, 5, 10}, Color: "#98A2B3"},
	}

	svg := StackedArea(Options{Title: "Alive, broke or dead"}, "Year", "%", series)

	if !strings.HasPrefix(svg, "<svg") {
		t.Fatalf("not an SVG: %.20q", svg)
	}
	if got := strings.Count(svg, "<polygon"); got != 3 {
		t.Errorf("polygons = %d, want one per layer (3)", got)
	}
	for _, name := range []string{"Funded", "Broke", "Dead"} {
		if !strings.Contains(svg, name) {
			t.Errorf("legend misses %q", name)
		}
	}
	if !strings.Contains(svg, "Alive, broke or dead") {
		t.Errorf("title missing")
	}
}

func TestStackedAreaEmptyIsSafe(t *testing.T) {
	if svg := StackedArea(Options{}, "x", "y", nil); !strings.HasPrefix(svg, "<svg") {
		t.Errorf("empty StackedArea should still return an SVG, got %.20q", svg)
	}
}
