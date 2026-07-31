package web

import (
	"strings"
	"testing"
)

// TestChartsRenderDark guards the -serve regression: the FIRE charts must
// carry the terminal-dark theme independently of the chart process-global,
// so they stay dark even in the same process as the light /view report. A
// dark chart shows the dark surface (#17130D) and never the light one
// (#FFFFFF); a chart added without going through theme.go's wrappers would
// fail here.
func TestChartsRenderDark(t *testing.T) {
	pr := Params{Capital: 1_000_000, NeedAnnual: 50000, Years: 40,
		Mu: 0.05, Sigma: 0.11, Df: 5, TaxRate: 0.30, NPaths: 1500}

	svgs := map[string]string{
		"sensitivity":    Sensitivity(pr, nil).SVG,
		"curves-horizon": func() string { c := Curves(pr, nil); return c.HorizonSVG }(),
	}
	for name, svg := range svgs {
		if !strings.HasPrefix(svg, "<svg") {
			t.Fatalf("%s: expected an SVG, got %.30q", name, svg)
		}
		if !strings.Contains(svg, "#17130D") {
			t.Errorf("%s: missing the dark surface (#17130D); chart not darkened", name)
		}
		if strings.Contains(svg, "#FFFFFF") {
			t.Errorf("%s: contains the light surface (#FFFFFF); chart rendered light", name)
		}
	}
}
