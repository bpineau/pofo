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
	if !strings.Contains(svg, "#D2402F") {
		t.Errorf("expected the ruin sample drawn in red")
	}
}

func TestFanEmptyIsSafe(t *testing.T) {
	svg := Fan(Options{}, "Year", nil, nil)
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("empty fan should still return an SVG, got %.20q", svg)
	}
}

// A cone compounding far beyond the start must be capped at 10x the starting
// wealth so the zero line stays readable, with the clip flagged.
func TestFanYAxisCap(t *testing.T) {
	years := 40
	mk := func(mult float64) []float64 {
		out := make([]float64, years)
		for i := range out {
			out[i] = 1e6 * (1 + (mult-1)*float64(i)/float64(years-1))
		}
		return out
	}
	svg := Fan(Options{Width: 640, Height: 360}, "Year",
		[][]float64{mk(0.2), mk(5), mk(60)}, nil) // p95 ends at 60x start
	if !strings.Contains(svg, "upside clipped") {
		t.Errorf("cap marker missing when the upper band exceeds 10x start")
	}
	// The y-axis must not scale to 60M: no tick label at or above 20M.
	for _, tick := range []string{">20M<", ">30M<", ">40M<", ">60M<"} {
		if strings.Contains(svg, tick) {
			t.Errorf("axis not capped: found tick %s", tick)
		}
	}
	// A modest cone stays uncapped and unflagged.
	svg = Fan(Options{Width: 640, Height: 360}, "Year",
		[][]float64{mk(0.5), mk(2), mk(4)}, nil)
	if strings.Contains(svg, "upside clipped") {
		t.Errorf("cap marker drawn on an uncapped fan")
	}
}
