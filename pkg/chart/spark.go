package chart

import (
	"fmt"
	"math"
	"strings"
)

// SparkOptions controls a Sparkline's rendering. Width and Height default
// to 72x20 pixels; Color defaults to the first palette color.
type SparkOptions struct {
	Width, Height int
	Color         string
}

// Sparkline renders a bare inline curve: one polyline, no axes, no labels,
// no background, stretched to its box (preserveAspectRatio "none") so it
// fits table cells and summary rows. Non-finite values are skipped; fewer
// than two finite values yield an empty string so callers can omit the
// chart entirely.
func Sparkline(opt SparkOptions, values []float64) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 72
	}
	if h == 0 {
		h = 20
	}
	color := opt.Color
	if color == "" {
		color = PaletteColor(0)
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	finite := 0
	for _, v := range values {
		if !isFinite(v) {
			continue
		}
		finite++
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	if finite < 2 {
		return ""
	}
	if hi == lo {
		hi = lo + 1 // flat series: a horizontal line in the middle
	}
	const pad = 2.0
	x := func(i int) float64 {
		return pad + float64(i)/float64(len(values)-1)*(float64(w)-2*pad)
	}
	y := func(v float64) float64 {
		return pad + (hi-v)/(hi-lo)*(float64(h)-2*pad)
	}
	var pts strings.Builder
	for i, v := range values {
		if !isFinite(v) {
			continue
		}
		fmt.Fprintf(&pts, "%.1f,%.1f ", x(i), y(v))
	}
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" preserveAspectRatio="none">`+
			`<polyline points="%s" fill="none" stroke="%s" stroke-width="1.3"/></svg>`,
		w, h, w, h, strings.TrimSpace(pts.String()), color)
}
