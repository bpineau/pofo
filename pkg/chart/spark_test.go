package chart

import (
	"math"
	"strings"
	"testing"
)

func TestSparkline(t *testing.T) {
	svg := Sparkline(SparkOptions{}, []float64{1, 2, math.NaN(), 3, 2.5})
	if !strings.Contains(svg, "<polyline") {
		t.Fatal("Sparkline should draw a polyline")
	}
	if !strings.Contains(svg, `preserveAspectRatio="none"`) {
		t.Error("Sparkline should stretch to its box")
	}
	if strings.Contains(svg, "<text") || strings.Contains(svg, "<rect") {
		t.Error("Sparkline should carry no labels or background")
	}
	if !strings.Contains(svg, `viewBox="0 0 72 20"`) {
		t.Errorf("default dimensions should be 72x20, got %q", svg)
	}
	if !strings.Contains(svg, PaletteColor(0)) {
		t.Error("default color should be the first palette color")
	}
	// The NaN point is skipped, not drawn at zero.
	if strings.Count(svg, ",") < 4 {
		t.Error("finite points should all be drawn")
	}
}

func TestSparklineOptions(t *testing.T) {
	svg := Sparkline(SparkOptions{Width: 120, Height: 32, Color: "#123456"}, []float64{5, 1, 4})
	if !strings.Contains(svg, `viewBox="0 0 120 32"`) || !strings.Contains(svg, "#123456") {
		t.Errorf("options not honored: %q", svg)
	}
}

func TestSparklineDegenerate(t *testing.T) {
	if Sparkline(SparkOptions{}, nil) != "" {
		t.Error("no values should yield an empty string")
	}
	if Sparkline(SparkOptions{}, []float64{1, math.NaN()}) != "" {
		t.Error("fewer than two finite values should yield an empty string")
	}
	if svg := Sparkline(SparkOptions{}, []float64{2, 2, 2}); svg == "" {
		t.Error("a flat series should still render")
	}
}
