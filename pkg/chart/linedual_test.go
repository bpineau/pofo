package chart

import (
	"strings"
	"testing"
)

func TestLineDualRendersBothAxes(t *testing.T) {
	left := XYSeries{Name: "Ruin %", Xs: []float64{0, 1, 2, 3}, Ys: []float64{20, 10, 5, 3}}
	right := XYSeries{Name: "Terminal k€", Xs: []float64{0, 1, 2, 3}, Ys: []float64{100, 300, 600, 900}}
	svg := LineDual(Options{Title: "Buffer arbitrage"}, "Buffer years", left, right)

	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("malformed SVG document")
	}
	for _, want := range []string{"Buffer arbitrage", "Ruin %", "Terminal k€", "Buffer years"} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG missing %q", want)
		}
	}
	if got := strings.Count(svg, "<path"); got < 2 {
		t.Errorf("want at least 2 line paths (one per axis), got %d", got)
	}
	if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
		t.Error("SVG contains NaN/Inf coordinates")
	}
}

func TestLineDualEmpty(t *testing.T) {
	svg := LineDual(Options{}, "x", XYSeries{}, XYSeries{})
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("empty input must still yield a valid SVG")
	}
}
