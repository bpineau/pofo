package chart

import (
	"strings"
	"testing"
)

// 2-4 named series get direct end labels (colored key + ink text) on top of
// the line-key legend; a single unnamed series gets neither.
func TestMultiLineEndLabels(t *testing.T) {
	svg := MultiLine(Options{Width: 720, Height: 360}, "x", "y", []XYSeries{
		{Name: "Alpha", Xs: []float64{0, 1, 2}, Ys: []float64{1, 2, 3}},
		{Name: "Beta", Xs: []float64{0, 1, 2}, Ys: []float64{3, 2.9, 3.1}},
	})
	if got := strings.Count(svg, ">Alpha<"); got != 2 {
		t.Errorf("Alpha should appear in the legend AND as an end label, got %d occurrences", got)
	}
	// Close end values must be deconflicted (Beta ends at 3.1, Alpha at 3).
	if !strings.Contains(svg, ">Beta<") {
		t.Errorf("Beta end label missing")
	}
	single := MultiLine(Options{Width: 720, Height: 360}, "x", "Ruin %", []XYSeries{
		{Xs: []float64{0, 1}, Ys: []float64{1, 2}},
	})
	if strings.Contains(single, `width="14" height="3"`) {
		t.Errorf("single unnamed series must not draw a legend key")
	}
}
