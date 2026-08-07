package chart

import "strings"

import "testing"

func TestFanRendersBandsAndMedian(t *testing.T) {
	bands := [][]float64{
		{100, 95, 90, 88},    // p5
		{100, 105, 110, 115}, // p50 (median, central)
		{100, 120, 140, 165}, // p95
	}
	samples := [][]float64{
		{100, 80, 40, 0},     // a ruin path (ends at 0)
		{100, 110, 130, 150}, // a healthy path
	}

	svg := Fan(Options{Title: "Wealth fan"}, "Year", bands, samples)

	if !strings.HasPrefix(svg, "<svg") {
		t.Fatalf("not an SVG: %.20q", svg)
	}
	if !strings.Contains(svg, "<polygon") {
		t.Errorf("expected a shaded band polygon")
	}
	if !strings.Contains(svg, "Wealth fan") {
		t.Errorf("expected the title")
	}
	// The ruin sample ends at zero and should be flagged in red.
	if !strings.Contains(svg, "#D92D20") {
		t.Errorf("expected the ruin sample drawn in red")
	}
}

func TestFanEmptyIsSafe(t *testing.T) {
	svg := Fan(Options{}, "Year", nil, nil)
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("empty fan should still return an SVG, got %.20q", svg)
	}
}
