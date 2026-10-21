package chart

import (
	"strings"
	"testing"
)

func TestStackedAreaRendersLayers(t *testing.T) {
	series := []AreaSeries{
		{Name: "Funded", Values: []float64{100, 90, 80}, Color: "#0E9384"},
		{Name: "Broke", Values: []float64{0, 5, 10}, Color: "#D2402F"},
		{Name: "Dead", Values: []float64{0, 5, 10}, Color: "#A8AEBC"},
	}

	svg := StackedArea(Options{Title: "Alive, broke or dead"}, "Year", "%", series)

	if !strings.HasPrefix(svg, "<svg") {
		t.Fatalf("not an SVG: %.20q", svg)
	}
	if got := strings.Count(svg, "<polygon"); got != 3 {
		t.Errorf("polygons = %d, want one per layer (3)", got)
	}
	for _, name := range []string{"Funded", "Broke", "Dead"} {
		if !strings.Contains(svg, name) {
			t.Errorf("legend misses %q", name)
		}
	}
	if !strings.Contains(svg, "Alive, broke or dead") {
		t.Errorf("title missing")
	}
}

func TestStackedAreaEmptyIsSafe(t *testing.T) {
	if svg := StackedArea(Options{}, "x", "y", nil); !strings.HasPrefix(svg, "<svg") {
		t.Errorf("empty StackedArea should still return an SVG, got %.20q", svg)
	}
}

// Under Gaps the layers are parted by the surface, only the Outline layers
// draw their two frontiers, the grid stays off the fills, an end label is
// dropped where its band is too thin to carry it, and a marker outside the
// x range is ignored.
func TestStackedAreaWithLayers(t *testing.T) {
	series := []AreaSeries{
		{Name: "Broke", Values: []float64{0, 5, 10}, Color: ColorBad, Outline: true},
		{Name: "Thin", Values: []float64{1, 1, 1}, Color: ColorAccent},
		{Name: "Funded", Values: []float64{99, 84, 49}, Color: ColorAccent, Weight: 0.3},
		{Name: "Gone", Values: []float64{0, 10, 40}, Color: ColorDead, Outline: true},
	}
	svg := StackedAreaWith(Options{Width: 600, Height: 300}, StackOptions{
		XLabel: "age", XTicks: []string{"60", "", "70"}, XTips: []string{"a60", "a65", "a70"},
		Max: 100, Unit: "%", Gaps: true, EndLabels: true,
		Markers: []Marker{{Axis: 'x', Value: 1, Label: "pension"}, {Axis: 'x', Value: 9, Label: "beyond"}},
	}, series)
	if got := strings.Count(svg, "<path"); got != 4 {
		t.Errorf("frontier paths = %d, want 2 per outlined layer (4)", got)
	}
	if strings.Contains(svg, `stroke-opacity="0.6"`) {
		t.Errorf("a grid line runs across the fills under Gaps")
	}
	if !strings.Contains(svg, ">40%") || !strings.Contains(svg, ">49%") || !strings.Contains(svg, ">10%") {
		t.Errorf("end labels missing")
	}
	if strings.Contains(svg, ">1%") {
		t.Errorf("a 1%% band got an end label it cannot carry")
	}
	if !strings.Contains(svg, "pension") || strings.Contains(svg, "beyond") {
		t.Errorf("markers: in-range one missing or out-of-range one drawn")
	}
	if !strings.Contains(svg, `"rows":["a60","a65","a70"]`) {
		t.Errorf("hover headers missing")
	}
	if !strings.Contains(svg, `fill-opacity="0.30"/><text`) {
		t.Errorf("legend swatch does not wear the layer's weight")
	}
}
