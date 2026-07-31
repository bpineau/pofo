package chart

import (
	"os"
	"strings"
	"testing"
)

// TestChromeColorsCentralised guards the reskin surface: no chart source may
// hardcode one of the theme's chrome colors as a string literal. New charts must
// draw with the tokens in theme.go, so a future reskin stays a single-file edit
// (this is the third time the consolidation has been needed; the guard keeps it
// from scattering again). Hex in comments is allowed, for documentation.
func TestChromeColorsCentralised(t *testing.T) {
	chrome := []string{
		themeInk, themeInkSoft, themeMuted, themeFaint, themeGrid, themeWell,
		themeAxis, themeSurface, themeAccent, themeGood, themeWarn, themeBad, themeDead,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "theme.go" {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(src), "\n") {
			code := line
			if i := strings.Index(code, "//"); i >= 0 {
				code = code[:i] // ignore hex in comments
			}
			for _, hx := range chrome {
				if strings.Contains(code, `"`+hx+`"`) {
					t.Errorf("%s hardcodes chrome color %s; use the matching token from theme.go", name, hx)
				}
			}
		}
	}
}
