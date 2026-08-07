package chart

import (
	"strings"
	"testing"
)

func TestCategoryBars(t *testing.T) {
	svg := CategoryBars(Options{Width: 460}, []CatBar{
		{Label: "Early crash", Value: 0.52, Text: "52%", Color: "#D2402F"},
		{Label: "Lost decade", Value: 0.31, Text: "31%", Color: "#C77E17"},
		{Label: "Longevity", Value: 0.17, Text: "17%", Color: "#9AA2B1"},
	})
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("not a well-formed svg")
	}
	for _, want := range []string{"Early crash", "52%", "Longevity", "#D2402F", "<rect"} {
		if !strings.Contains(svg, want) {
			t.Errorf("category bars missing %q", want)
		}
	}
}
