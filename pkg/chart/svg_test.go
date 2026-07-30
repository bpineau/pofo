package chart

import (
	"strings"
	"testing"
	"time"
)

func TestLineRendersAllSeries(t *testing.T) {
	n := 3000 // force decimation
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := make([]time.Time, n)
	v1 := make([]float64, n)
	v2 := make([]float64, n)
	for i := range dates {
		dates[i] = start.AddDate(0, 0, i)
		v1[i] = 100 + float64(i)*0.05
		v2[i] = 100 - float64(i)*0.01
	}
	svg := Line(Options{Title: "Test <chart> & co"}, []Series{
		{Name: "A & B", Dates: dates, Values: v1, Color: "#ff0000"},
		{Name: "C", Dates: dates, Values: v2},
	})
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(svg, "</svg>") {
		t.Fatal("malformed SVG document")
	}
	if got := strings.Count(svg, "<path"); got != 2 {
		t.Errorf("want 2 curves, got %d", got)
	}
	if strings.Contains(svg, "NaN") {
		t.Error("the SVG contains NaN")
	}
	if !strings.Contains(svg, "Test &lt;chart&gt; &amp; co") {
		t.Error("the title must be escaped")
	}
	if !strings.Contains(svg, "A &amp; B") {
		t.Error("the legend must be escaped")
	}
	// Years must appear on the x-axis over a 2000-2008 span.
	if !strings.Contains(svg, ">2004<") {
		t.Error("missing year tick")
	}
}

func TestLineSingleSeriesHasNoLegend(t *testing.T) {
	dates := []time.Time{
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	svg := Line(Options{Title: "solo"}, []Series{{Name: "P1", Dates: dates, Values: []float64{100, 110}}})
	if strings.Contains(svg, ">P1<") {
		t.Error("no legend expected for a single series")
	}
}
