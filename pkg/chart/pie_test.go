package chart

import (
	"strings"
	"testing"
)

func TestPieRendersWedgesAndLegend(t *testing.T) {
	const customColor = "#cbcbcb"
	svg := Pie(PieOptions{Title: "Geo <x> & co"}, []Slice{
		{Label: "US", Value: 60},
		{Label: "Japan & co", Value: 30},
		{Label: "Other", Value: 10, Color: customColor},
	})
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("malformed SVG document")
	}
	if got := strings.Count(svg, "<path"); got != 3 {
		t.Errorf("want 3 wedges, got %d", got)
	}
	if strings.Contains(svg, "NaN") {
		t.Error("the SVG contains NaN")
	}
	// Title and labels must be escaped.
	if !strings.Contains(svg, "Geo &lt;x&gt; &amp; co") {
		t.Error("the title must be escaped")
	}
	if !strings.Contains(svg, "Japan &amp; co") {
		t.Error("the legend must be escaped")
	}
	// Shares are normalized to percentages: 60/30/10 of a total of 100.
	for _, want := range []string{">60%<", ">30%<", ">10%<"} {
		if !strings.Contains(svg, want) {
			t.Errorf("missing legend percentage %q", want)
		}
	}
	if !strings.Contains(svg, customColor) {
		t.Error("an explicit slice color must be honored")
	}
}

func TestPieNormalizesArbitraryUnits(t *testing.T) {
	// Values need not sum to 100; they are normalized.
	svg := Pie(PieOptions{}, []Slice{{Label: "A", Value: 3}, {Label: "B", Value: 1}})
	if !strings.Contains(svg, ">75%<") || !strings.Contains(svg, ">25%<") {
		t.Error("values must be normalized to percentages (75/25)")
	}
}

func TestPieEmptyWhenNoPositiveValue(t *testing.T) {
	if got := Pie(PieOptions{Title: "x"}, []Slice{{Label: "A", Value: 0}, {Label: "B", Value: -2}}); got != "" {
		t.Errorf("want empty string for a pie with no positive value, got %q", got)
	}
	if got := Pie(PieOptions{}, nil); got != "" {
		t.Errorf("want empty string for a nil pie, got %q", got)
	}
}

func TestPieHole(t *testing.T) {
	slices := []Slice{{Label: "a", Value: 3}, {Label: "b", Value: 1}}
	def := Pie(PieOptions{}, slices)
	if !strings.Contains(def, "A70 70 0") || !strings.Contains(def, "A42 42 0") {
		t.Fatalf("default geometry changed: %q", def)
	}
	wide := Pie(PieOptions{Hole: 0.9}, slices)
	if !strings.Contains(wide, "A63 63 0") {
		t.Errorf("Hole 0.9 should shrink the ring to innerR=63, got %q", wide)
	}
}

func TestPieHideLegend(t *testing.T) {
	slices := []Slice{{Label: "a", Value: 3, Color: "#111111"}, {Label: "b", Value: 1}}
	svg := Pie(PieOptions{Width: 190, HideLegend: true}, slices)
	if strings.Contains(svg, "<text") || strings.Contains(svg, "<rect") {
		t.Error("legend-less pie should carry no text or swatches")
	}
	if !strings.Contains(svg, `viewBox="0 0 190 190"`) {
		t.Errorf("legend-less pie should be square, got %q", svg)
	}
	if !strings.Contains(svg, "#111111") {
		t.Error("caller colors should be honored")
	}
}
