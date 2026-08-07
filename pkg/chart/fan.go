package chart

import (
	"fmt"
	"math"
	"strings"
)

// Fan renders a wealth fan chart: shaded percentile bands, a median line, and a
// few individual sample paths overlaid. bands are ascending-percentile series
// over the x index (year 0..N); they are shaded in symmetric pairs (outer to
// inner) and the central band, when the count is odd, is drawn as the median
// line. samples are individual paths; one that ends at (or below) zero is a ruin
// path and is drawn in red so the reader sees what failure looks like.
func Fan(opt Options, xLabel string, bands [][]float64, samples [][]float64) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	marginL, marginR, top, bottom := 64.0, 16.0, 40.0, 40.0
	x0, x1 := marginL, float64(w)-marginR
	y0, y1 := top, float64(h)-bottom

	steps := 0
	for _, b := range bands {
		steps = max(steps, len(b))
	}
	for _, s := range samples {
		steps = max(steps, len(s))
	}

	// Scale the y-axis to the percentile bands only: a single lucky sample path
	// can reach many times the p95 over a long horizon and would otherwise
	// squash the informative bands. Sample lines are clipped to this range.
	scaleBands := bands
	if len(bands) > 2 {
		scaleBands = bands[:len(bands)-1] // drop the outermost (p95) so the low region breathes
	}
	var scale []float64
	for _, b := range scaleBands {
		scale = append(scale, b...)
	}
	if len(scale) == 0 {
		for _, s := range samples {
			scale = append(scale, s...)
		}
	}
	vmin, vmax := axisBounds(scale) // starts at 0 for non-negative wealth
	clamp := func(v float64) float64 { return math.Max(vmin, math.Min(v, vmax)) }

	xmax := float64(max(steps-1, 1))
	xAt := func(i int) float64 { return x0 + float64(i)/xmax*(x1-x0) }
	yAt := func(v float64) float64 { return y1 - (v-vmin)/(vmax-vmin)*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, Segoe UI, Helvetica, Arial, sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#FFFFFF"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="#101828">%s</text>`+"\n", x0, esc(opt.Title))
	}

	// Horizontal grid and y-axis labels (wealth).
	step := niceStep(vmax-vmin, 6)
	for v := math.Ceil(vmin/step) * step; v <= vmax+step/1e6; v += step {
		y := yAt(v)
		fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="#E9EDF3"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="#667085" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, step))
	}
	// X-axis ticks (years) and label.
	for _, i := range axisTicks(steps) {
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="#E9EDF3"/>`+"\n", xAt(i), y0, xAt(i), y1)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#667085" text-anchor="middle">%d</text>`+"\n", xAt(i), y1+16, i)
	}
	if xLabel != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#667085" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, y1+32, esc(xLabel))
	}
	// Bottom axis is zero wealth: draw it as a bold red ruin line, labelled, so
	// paths reaching 0 (running out of money) are unmistakable.
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#D92D20" stroke-width="1.8"/>`+"\n", x0, y1, x1, y1)
	fmt.Fprintf(&b, `<text x="%g" y="%g" font-size="11" font-weight="600" fill="#D92D20">ruin · 0</text>`+"\n", x0+5, y1-4)
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#C6CEDA"/>`+"\n", x0, y0, x0, y1)

	// Shaded bands, outermost first so inner pairs paint on top.
	const bandFill = "#2E4BE0"
	n := len(bands)
	for i := 0; i < n/2; i++ {
		opacity := 0.10 + 0.10*float64(i) // inner pairs a touch more opaque
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="%.2f" stroke="none"/>`+"\n",
			bandPolygon(bands[i], bands[n-1-i], xAt, yAt), bandFill, opacity)
	}

	// Individual sample paths (clipped to the band range). Ruin paths (ending at
	// or below 0) are drawn thicker and saturated so failures stand out; the rest
	// are faint grey context.
	for _, s := range samples {
		color, width, op := "#98A2B3", "1", "0.55"
		if len(s) > 0 && s[len(s)-1] <= 0 {
			color, width, op = "#D92D20", "1.8", "0.9"
		}
		clipped := make([]float64, len(s))
		for i, v := range s {
			clipped[i] = clamp(v)
		}
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="%s" stroke-opacity="%s"/>`+"\n",
			linePath(clipped, xAt, yAt), color, width, op)
	}

	// Median line: the central band when the percentile count is odd.
	if n%2 == 1 {
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="#101828" stroke-width="2.2" stroke-linejoin="round"/>`+"\n",
			linePath(bands[n/2], xAt, yAt))
	}
	b.WriteString("</svg>")
	return b.String()
}

// bandPolygon builds the closed polygon between a lower and an upper band: along
// the lower forward, back along the upper.
func bandPolygon(lower, upper []float64, xAt func(int) float64, yAt func(float64) float64) string {
	var p strings.Builder
	for i, v := range lower {
		fmt.Fprintf(&p, "%.1f,%.1f ", xAt(i), yAt(v))
	}
	for i := len(upper) - 1; i >= 0; i-- {
		fmt.Fprintf(&p, "%.1f,%.1f ", xAt(i), yAt(upper[i]))
	}
	return strings.TrimSpace(p.String())
}

// linePath builds an SVG path for one series over the x index.
func linePath(ys []float64, xAt func(int) float64, yAt func(float64) float64) string {
	var p strings.Builder
	for i, v := range ys {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&p, "%s%.1f %.1f", cmd, xAt(i), yAt(v))
	}
	return p.String()
}

// axisTicks returns up to ~8 evenly spaced x indices for the year axis,
// including the first and last.
func axisTicks(steps int) []int {
	if steps <= 1 {
		return []int{0}
	}
	last := steps - 1
	stride := max((last+7)/8, 1)
	var out []int
	for i := 0; i < last; i += stride {
		out = append(out, i)
	}
	return append(out, last)
}
