package chart

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Coincident points must not stack their labels on the same pixel row.
func TestScatterLabelDeconflict(t *testing.T) {
	svg := Scatter(Options{Width: 640, Height: 360}, "x", "y", []LabeledPoint{
		{X: 1, Y: 1, Label: "Fixed"},
		{X: 1, Y: 1, Label: "Guardrails"},
		{X: 1.02, Y: 1.01, Label: "Flex -10%"},
	})
	re := regexp.MustCompile(`<text x="([0-9.]+)" y="([0-9.]+)" font-size="12"`)
	var ys []float64
	for _, m := range re.FindAllStringSubmatch(svg, -1) {
		v, _ := strconv.ParseFloat(m[2], 64)
		ys = append(ys, v)
	}
	if len(ys) != 3 {
		t.Fatalf("want 3 point labels, got %d\n%s", len(ys), svg)
	}
	for i := range ys {
		for j := i + 1; j < len(ys); j++ {
			if d := ys[i] - ys[j]; d > -14 && d < 14 {
				t.Errorf("labels %d and %d overlap vertically (y %.1f vs %.1f)", i, j, ys[i], ys[j])
			}
		}
	}
	if !strings.Contains(svg, "Guardrails") {
		t.Errorf("label text missing")
	}
}
