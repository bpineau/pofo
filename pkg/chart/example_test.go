package chart_test

import (
	"fmt"
	"strings"
	"time"

	"portfodor/pkg/chart"
)

// Line produces a self-contained SVG document, embeddable as-is in an HTML
// page, from one or more dated series.
func ExampleLine() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var a, b []float64
	for i := range 500 {
		dates = append(dates, start.AddDate(0, 0, i))
		a = append(a, 100+float64(i)*0.1)
		b = append(b, 100-float64(i)*0.02)
	}
	svg := chart.Line(chart.Options{Title: "Comparison", Width: 800, Height: 400}, []chart.Series{
		{Name: "Rising", Dates: dates, Values: a, Color: "#1f77b4"},
		{Name: "Declining", Dates: dates, Values: b},
	})
	fmt.Println(strings.HasPrefix(svg, "<svg"), strings.HasSuffix(svg, "</svg>"))
	fmt.Println(strings.Count(svg, "<path"))
	// Output:
	// true true
	// 2
}

// Term plots the same series for the terminal — ANSI colors on a TTY,
// distinct markers otherwise.
func ExampleTerm() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var v []float64
	for i := range 300 {
		dates = append(dates, start.AddDate(0, 0, i))
		v = append(v, 100+float64(i)*0.1)
	}
	out := chart.Term(chart.TermOptions{Title: "Demo", Width: 60, Height: 8},
		[]chart.Series{{Name: "P", Dates: dates, Values: v}})
	fmt.Println(strings.Contains(out, "┤"), strings.Contains(out, "2020"))
	// Output:
	// true true
}
