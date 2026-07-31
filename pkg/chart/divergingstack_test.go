package chart

import (
	"strings"
	"testing"
)

func TestDivergingStack(t *testing.T) {
	svg := DivergingStack(DivergingStackOptions{
		Title:   "contrib",
		XLabels: []string{"2020", "", "", "2021"},
		XLabel:  "month",
		YLabel:  "pts",
		Total:   []float64{1, -1, 2, 3},
		Strip: []StripBand{
			{From: 0, To: 1, Label: "growth", Color: "#D5DBE5"},
			{From: 2, To: 3, Label: "crisis", Color: "#D2402F"},
		},
		StripLegend: []Slice{{Label: "growth", Color: "#D5DBE5"}, {Label: "crisis", Color: "#D2402F"}},
	}, []DivergingStackSeries{
		{Name: "EQ", Values: []float64{2, -3, 1, 2}},
		{Name: "GOLD", Values: []float64{-1, 2, 1, 1}},
	})
	for _, want := range []string{
		"<polygon", "EQ", "GOLD", `metadata class="hover"`, `"kind":"stack"`,
		"growth", "crisis", "2021", "<polyline", `data-tip="growth"`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("DivergingStack output lacks %q", want)
		}
	}
	// One positive and one negative polygon per series.
	if got := strings.Count(svg, "<polygon"); got != 4 {
		t.Errorf("want 4 polygons (pos+neg per series), got %d", got)
	}
}

func TestDivergingStackEmpty(t *testing.T) {
	if svg := DivergingStack(DivergingStackOptions{}, nil); svg != "" {
		t.Errorf("empty input must render nothing, got %d bytes", len(svg))
	}
	if svg := DivergingStack(DivergingStackOptions{}, []DivergingStackSeries{{Name: "X", Values: []float64{1}}}); svg != "" {
		t.Errorf("a single point cannot make an area, got %d bytes", len(svg))
	}
}

func TestBarMatrix(t *testing.T) {
	svg := BarMatrix(BarMatrixOptions{
		Title:        "per regime",
		RowLabels:    []string{"EQ", "GOLD"},
		Unit:         "pts/yr",
		Summary:      []float64{30, -1}, // 30 exceeds the row scale: clamped bar, true label
		SummaryLabel: "portfolio",
	}, []MatrixColumn{
		{Title: "growth", Subtitle: "75 months", Color: "#D5DBE5", Values: []float64{5.3, 1.5}},
		{Title: "crisis", Subtitle: "58 months", Color: "#D2402F", Values: []float64{-2.0, 1.8}},
	})
	for _, want := range []string{
		"per regime", "growth", "crisis", "75 months",
		`data-tip="EQ · growth: +5.3 pts/yr"`,
		`data-tip="portfolio · growth: +30.0 pts/yr"`,
		"portfolio", "+5.3",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("BarMatrix output lacks %q", want)
		}
	}
}

func TestSafeStackOrder(t *testing.T) {
	cases := []struct {
		n    int
		want []int
	}{
		{3, []int{0, 1, 2}},
		{8, []int{0, 1, 2, 3, 4, 6, 7, 5}},
		{9, []int{0, 8, 1, 2, 3, 4, 6, 7, 5}}, // the wrapped twin follows its color
	}
	for _, c := range cases {
		got := SafeStackOrder(c.n)
		if len(got) != c.n {
			t.Fatalf("n=%d: %d indices", c.n, len(got))
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("SafeStackOrder(%d) = %v, want %v", c.n, got, c.want)
			}
		}
	}
}

func TestBarMatrixEmpty(t *testing.T) {
	if svg := BarMatrix(BarMatrixOptions{}, nil); svg != "" {
		t.Errorf("empty input must render nothing, got %d bytes", len(svg))
	}
}
