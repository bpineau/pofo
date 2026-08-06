package chart

import (
	"strings"
	"testing"
)

func TestMultiLineRendersSeriesAndMarkers(t *testing.T) {
	a := XYSeries{Name: "Student-t", Xs: []float64{2, 3, 4, 5}, Ys: []float64{1, 4, 12, 28}}
	b := XYSeries{Name: "Conservative", Xs: []float64{2, 3, 4, 5}, Ys: []float64{5, 14, 30, 52}}

	svg := MultiLine(Options{Title: "Ruin vs withdrawal"}, "Withdrawal %", "Ruin %",
		[]XYSeries{a, b}, Marker{Axis: 'x', Value: 3.3, Label: "you"}, Marker{Axis: 'y', Value: 5, Label: "target"})

	if !strings.HasPrefix(svg, "<svg") {
		t.Fatalf("not an SVG: %.20q", svg)
	}
	if !strings.Contains(svg, "Student-t") || !strings.Contains(svg, "Conservative") {
		t.Errorf("expected both series names in the legend")
	}
	if !strings.Contains(svg, "stroke-dasharray") {
		t.Errorf("expected dashed marker lines")
	}
	if !strings.Contains(svg, "target") {
		t.Errorf("expected the marker label")
	}
}
