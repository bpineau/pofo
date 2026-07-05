package chart

import (
	"strings"
	"testing"
)

// Darken rewrites the chrome from the light theme to the terminal-dark theme:
// the dark surface replaces white, and no light chrome color survives.
func TestDarken(t *testing.T) {
	light := Fan(Options{Width: 240, Height: 160}, "y", [][]float64{{100, 90}, {100, 110}}, nil)
	dark := Darken(light)
	if !strings.Contains(dark, "#17130D") {
		t.Error("dark chart should carry the dark surface #17130D")
	}
	for _, lightHex := range []string{themeSurface, themeGrid, themeMuted, themeInk} {
		if strings.Contains(dark, lightHex) {
			t.Errorf("dark chart must not carry light chrome %s", lightHex)
		}
	}
	if light == dark {
		t.Error("Darken should change the output")
	}
}
