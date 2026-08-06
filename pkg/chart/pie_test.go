package chart

import (
	"strings"
	"testing"
)

func TestPieRendersWedgesAndLegend(t *testing.T) {
	svg := Pie(PieOptions{Title: "Geo <x> & co"}, []Slice{
		{Label: "US", Value: 60},
		{Label: "Japan & co", Value: 30},
		{Label: "Other", Value: 10, Color: NeutralColor},
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
	if !strings.Contains(svg, NeutralColor) {
		t.Error("the neutral slice color must be used")
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
