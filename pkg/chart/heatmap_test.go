package chart

import (
	"strings"
	"testing"
)

func TestHeatmapSVG(t *testing.T) {
	d := HeatmapData{Xs: []float64{0, 1}, Ys: []float64{0.03, 0.05}, Z: [][]float64{{0.1, 0.2}, {0.3, 0.4}}}
	svg := Heatmap(Options{Title: "Ruin"}, d)
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("not an SVG")
	}
	if strings.Count(svg, "<rect") < 4 {
		t.Errorf("expected 4 cells, got %d", strings.Count(svg, "<rect"))
	}
}
