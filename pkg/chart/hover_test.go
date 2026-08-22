package chart

import (
	"encoding/json"
	"regexp"
	"testing"
)

// parseHover pulls a chart's embedded hover payload back out of its SVG.
func parseHover(t *testing.T, svg string) hoverMeta {
	t.Helper()
	m := regexp.MustCompile(`(?s)<metadata class="hover">(.*?)</metadata>`).FindStringSubmatch(svg)
	if m == nil {
		t.Fatalf("no hover metadata in %.60q", svg)
	}
	var hm hoverMeta
	if err := json.Unmarshal([]byte(m[1]), &hm); err != nil {
		t.Fatalf("hover payload is not JSON: %v", err)
	}
	return hm
}

// Every discrete chart must publish the geometry a hover layer needs to answer
// the pointer ANYWHERE over its plot: a plot box, a layout axis, and one mark
// anchor per row, inside that box. Without it a tooltip can only fire on the
// drawn ink, which leaves the middle of a bar chart mute.
func TestDiscreteChartsCarryHoverGeometry(t *testing.T) {
	bars := []Bar{{Label: "a", Value: 3, Text: "3%"}, {Label: "b", Value: 7, Text: "7%"}, {Label: "c", Value: -2, Text: "-2%"}}
	cats := []CatBar{{Label: "x", Value: 0.6, Text: "60%"}, {Label: "y", Value: 0.4, Text: "40%"}}
	pts := []LabeledPoint{{X: 1, Y: 2, Label: "p"}, {X: 3, Y: 4, Label: "q"}}
	opt := Options{Title: "T", Width: 640, Height: 360}

	for _, tc := range []struct {
		name string
		svg  string
		axis string
		n    int
	}{
		{"bars", Bars(opt, bars), "x", len(bars)},
		{"hbars", HBars(opt, bars), "y", len(bars)},
		{"categorybars", CategoryBars(opt, cats), "y", len(cats)},
		{"scatter", Scatter(opt, "x", "y", pts), "xy", len(pts)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hm := parseHover(t, tc.svg)
			if hm.Axis != tc.axis {
				t.Errorf("axis = %q, want %q", hm.Axis, tc.axis)
			}
			if len(hm.Marks) != tc.n {
				t.Fatalf("got %d marks, want %d (one per row)", len(hm.Marks), tc.n)
			}
			if len(hm.Rows) != tc.n {
				t.Errorf("got %d row labels, want %d", len(hm.Rows), tc.n)
			}
			for _, s := range hm.Series {
				if len(s.Ys) != tc.n {
					t.Errorf("series %q has %d values, want %d", s.Name, len(s.Ys), tc.n)
				}
			}
			if !(hm.X1 > hm.X0) || !(hm.Y1 > hm.Y0) {
				t.Fatalf("empty plot box (%g,%g)-(%g,%g)", hm.X0, hm.Y0, hm.X1, hm.Y1)
			}
			for i, m := range hm.Marks {
				if m.X < hm.X0 || m.X > hm.X1 || m.Y < hm.Y0 || m.Y > hm.Y1 {
					t.Errorf("mark %d at (%.1f,%.1f) sits outside the plot box (%g,%g)-(%g,%g)",
						i, m.X, m.Y, hm.X0, hm.Y0, hm.X1, hm.Y1)
				}
			}
		})
	}
}

// The marks of a laid-out chart must be evenly spaced along their axis and in
// row order, since the hover layer derives the highlight band from their
// spacing.
func TestBarMarksAreOrderedAndEvenlySpaced(t *testing.T) {
	bars := []Bar{{Label: "a", Value: 1}, {Label: "b", Value: 4}, {Label: "c", Value: 2}, {Label: "d", Value: 3}}
	hm := parseHover(t, Bars(Options{Width: 640, Height: 360}, bars))
	step := hm.Marks[1].X - hm.Marks[0].X
	if step <= 0 {
		t.Fatalf("marks not left to right: %+v", hm.Marks)
	}
	for i := 2; i < len(hm.Marks); i++ {
		if d := hm.Marks[i].X - hm.Marks[i-1].X; d < step-0.01 || d > step+0.01 {
			t.Errorf("mark %d is %.2f px after its neighbour, want %.2f", i, d, step)
		}
	}
}

// A bar's own value label rides in the payload: it carries the unit that the
// raw number lost, and the tooltip prefers it.
func TestMarkTextCarriesTheFormattedValue(t *testing.T) {
	hm := parseHover(t, Bars(Options{Width: 640, Height: 360}, []Bar{{Label: "a", Value: 3, Text: "3%"}}))
	if hm.Marks[0].Text != "3%" {
		t.Errorf("mark text = %q, want %q", hm.Marks[0].Text, "3%")
	}
}
