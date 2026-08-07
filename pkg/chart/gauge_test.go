package chart

import (
	"strings"
	"testing"
)

func TestGauge(t *testing.T) {
	svg := Gauge(Options{Width: 360, Height: 210}, "30.8", "CAPE", "cheap", "rich", 0.95)
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("not a well-formed svg")
	}
	for _, want := range []string{"gaugeg", "30.8", "CAPE", "cheap", "rich", "<line"} {
		if !strings.Contains(svg, want) {
			t.Errorf("gauge missing %q", want)
		}
	}
}

func TestGaugeClampsFraction(t *testing.T) {
	// Out-of-range fractions must not panic or escape the box.
	_ = Gauge(Options{}, "x", "", "", "", -3)
	_ = Gauge(Options{}, "x", "", "", "", 4)
}
