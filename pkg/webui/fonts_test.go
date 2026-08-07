package webui

import (
	"strings"
	"testing"
)

func TestFontsCSS(t *testing.T) {
	for _, want := range []string{
		"font-family:'Instrument Sans'",
		"font-family:'Spline Sans Mono'",
		"font-weight:400 700",
		"font-display:swap",
		"data:font/woff2;base64,",
	} {
		if !strings.Contains(FontsCSS, want) {
			t.Errorf("FontsCSS missing %q", want)
		}
	}
	if n := strings.Count(FontsCSS, "@font-face"); n != 2 {
		t.Errorf("FontsCSS has %d @font-face rules, want 2", n)
	}
}
