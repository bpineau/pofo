package chart

import (
	"strings"
	"testing"
)

func TestBarsSVG(t *testing.T) {
	svg := Bars(Options{Title: "Recovery"}, []Bar{{"0y", 0.4}, {"1y", 0.3}, {"2y", 0.3}})
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("not an SVG: %.20q", svg)
	}
	if !strings.Contains(svg, "Recovery") {
		t.Errorf("title missing")
	}
	if strings.Count(svg, "<rect") < 3 {
		t.Errorf("expected at least 3 bars")
	}
}
