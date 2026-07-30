package chart_test

import (
	"fmt"
	"strings"
	"time"

	"portfodor/chart"
)

// Line produit un document SVG autonome, embarquable tel quel dans une page
// HTML, à partir d'une ou plusieurs séries datées.
func ExampleLine() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var a, b []float64
	for i := range 500 {
		dates = append(dates, start.AddDate(0, 0, i))
		a = append(a, 100+float64(i)*0.1)
		b = append(b, 100-float64(i)*0.02)
	}
	svg := chart.Line(chart.Options{Title: "Comparaison", Width: 800, Height: 400}, []chart.Series{
		{Name: "Croissant", Dates: dates, Values: a, Color: "#1f77b4"},
		{Name: "Déclinant", Dates: dates, Values: b},
	})
	fmt.Println(strings.HasPrefix(svg, "<svg"), strings.HasSuffix(svg, "</svg>"))
	fmt.Println(strings.Count(svg, "<path"))
	// Output:
	// true true
	// 2
}
