package chart_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
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

// StyleMinimal renders the bare dialect: no background, no grid, no axes,
// an area fill under the curve and the date range at the corners.
func ExampleStyleMinimal() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var v []float64
	for i := range 90 {
		dates = append(dates, start.AddDate(0, 0, i))
		v = append(v, 100+float64(i%17))
	}
	opt := chart.Options{Width: 400, Height: 160, Style: chart.StyleMinimal()}
	svg := chart.Line(opt, []chart.Series{{Name: "net", Dates: dates, Values: v}})
	fmt.Println(strings.Contains(svg, "<rect"), strings.Contains(svg, "fill-opacity"))
	// Output:
	// false true
}

// Braille mode packs 2x4 dots per terminal cell for a smoother curve.
func ExampleTerm_braille() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	var v []float64
	for i := range 120 {
		dates = append(dates, start.AddDate(0, 0, i))
		v = append(v, 100+float64(i))
	}
	out := chart.Term(chart.TermOptions{Width: 40, Height: 6, Braille: true},
		[]chart.Series{{Name: "up", Dates: dates, Values: v}})
	braille := false
	for _, r := range out {
		if r > 0x2800 && r <= 0x28FF {
			braille = true
		}
	}
	fmt.Println(braille)
	// Output:
	// true
}

// Sparkline packs a value trail into a bare inline curve for table cells.
func ExampleSparkline() {
	svg := chart.Sparkline(chart.SparkOptions{Width: 72, Height: 20, Color: "#2E6E63"},
		[]float64{100, 104, 101, 108, 112})
	fmt.Println(strings.Count(svg, "<polyline"), strings.Contains(svg, "<text"))
	// Output:
	// 1 false
}

// Term plots the same series for the terminal: ANSI colors on a TTY,
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

// Pie renders a self-contained SVG donut with a legend, for composition
// breakdowns. Slice shares are normalized internally.
func ExamplePie() {
	svg := chart.Pie(chart.PieOptions{Title: "Allocation"}, []chart.Slice{
		{Label: "Equity", Value: 60},
		{Label: "Bonds", Value: 40},
	})
	fmt.Println(strings.HasPrefix(svg, "<svg"), strings.Contains(svg, "Equity"), strings.Contains(svg, "60%"))
	// Output:
	// true true true
}

// Line renders an intraday path the same way as a daily one: feed it a series
// of timestamps and prices spanning a single day, and the time axis switches
// to clock-time labels.
func ExampleLine_intraday() {
	open := time.Date(2024, 3, 1, 9, 30, 0, 0, time.UTC)
	var dates []time.Time
	var v []float64
	for i := range 78 { // a 6.5h session at 5-minute resolution
		dates = append(dates, open.Add(time.Duration(i)*5*time.Minute))
		v = append(v, 500+float64(i)*0.05)
	}
	svg := chart.Line(chart.Options{Title: "VOO today"}, []chart.Series{
		{Name: "price USD", Dates: dates, Values: v},
	})
	fmt.Println(strings.HasPrefix(svg, "<svg"), strings.Contains(svg, ":"))
	// Output:
	// true true
}
