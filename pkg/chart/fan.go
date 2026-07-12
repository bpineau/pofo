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
// Optional markers draw dashed reference lines (e.g. the year a pension
// starts); the band data is embedded as hover metadata for the crosshair.
func Fan(opt Options, xLabel string, bands [][]float64, samples [][]float64, markers ...Marker) string {
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

	// Scale the y-axis to the percentile bands (a single lucky sample path can
	// reach many times the p95 and would squash everything), then cap it at
	// 10x the starting wealth: over long horizons the upper cone compounds to
	// many times the start, and an uncapped axis crushes the region that
	// matters (the start and the zero line) into the bottom pixels. Anything
	// above the axis range is drawn clamped and the clip is flagged.
	var scale []float64
	for _, b := range bands {
		scale = append(scale, b...)
	}
	if len(scale) == 0 {
		for _, s := range samples {
			scale = append(scale, s...)
		}
	}
	vmin, vmax := axisBounds(scale) // starts at 0 for non-negative wealth
	if start := fanStart(bands, samples); start > 0 && vmax > 10*start {
		vmax = 10 * start
	}
	clipped := false
	for _, vs := range append(append([][]float64{}, bands...), samples...) {
		for _, v := range vs {
			if v > vmax {
				clipped = true
			}
		}
	}
	clamp := func(v float64) float64 { return math.Max(vmin, math.Min(v, vmax)) }
	clampAll := func(vs []float64) []float64 {
		out := make([]float64, len(vs))
		for i, v := range vs {
			out[i] = clamp(v)
		}
		return out
	}

	xmax := float64(max(steps-1, 1))
	xAt := func(i int) float64 { return x0 + float64(i)/xmax*(x1-x0) }
	yAt := func(v float64) float64 { return y1 - (v-vmin)/(vmax-vmin)*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", x0, esc(opt.Title))
	}

	// Horizontal grid and y-axis labels (wealth).
	step := niceStep(vmax-vmin, 6)
	for v := math.Ceil(vmin/step) * step; v <= vmax+step/1e6; v += step {
		y := yAt(v)
		fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="`+themeGrid+`"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, step))
	}
	// X-axis ticks (years) and label.
	for _, i := range axisTicks(steps) {
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="`+themeGrid+`"/>`+"\n", xAt(i), y0, xAt(i), y1)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%d</text>`+"\n", xAt(i), y1+16, i)
	}
	if xLabel != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, y1+32, esc(xLabel))
	}
	// Bottom axis is zero: draw it as a bold red line, labelled, so paths
	// reaching 0 are unmistakable. The label defaults to the wealth-fan
	// reading ("ruin"); a market-only fan overrides it via Style.ZeroLabel.
	zeroLabel := opt.Style.ZeroLabel
	if zeroLabel == "" {
		zeroLabel = "ruin · 0"
	}
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="`+themeBad+`" stroke-width="1.8"/>`+"\n", x0, y1, x1, y1)
	fmt.Fprintf(&b, `<text x="%g" y="%g" font-size="11" font-weight="600" fill="`+themeBad+`">%s</text>`+"\n", x0+5, y1-4, esc(zeroLabel))
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="`+themeAxis+`"/>`+"\n", x0, y0, x0, y1)

	// Reference markers (dashed), e.g. the pension start year.
	xAtF := func(v float64) float64 { return x0 + v/xmax*(x1-x0) }
	for _, m := range markers {
		if m.Axis != 'x' {
			continue // fans only carry vertical (year) references
		}
		x := xAtF(m.Value)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="`+themeFaint+`" stroke-dasharray="4 3"/>`+"\n", x, y0, x, y1)
		fmt.Fprintf(&b, `<text x="%.1f" y="%g" font-size="11" fill="`+themeFaint+`" text-anchor="middle">%s</text>`+"\n", x, y0-2, esc(m.Label))
	}

	// Hover payload: the percentile bands, high to low so the tooltip reads
	// top-down like the chart.
	names := fanBandNames(len(bands))
	hm := hoverMeta{Kind: "fan", X0: x0, X1: x1, Y0: y0, Y1: y1, Xmin: 0, Xmax: xmax, XLabel: xLabel}
	for i := len(bands) - 1; i >= 0; i-- {
		color := bandFillColor
		if len(bands)%2 == 1 && i == len(bands)/2 {
			color = themeInk
		}
		hm.Series = append(hm.Series, hoverSeries{Name: names[i], Color: color, Ys: bands[i]})
	}
	b.WriteString(hoverBlock(hm))

	// Shaded bands, outermost first so inner pairs paint on top, values
	// clamped to the (possibly capped) axis range.
	n := len(bands)
	for i := 0; i < n/2; i++ {
		opacity := 0.10 + 0.10*float64(i) // inner pairs a touch more opaque
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="%.2f" stroke="none"/>`+"\n",
			bandPolygon(clampAll(bands[i]), clampAll(bands[n-1-i]), xAt, yAt), bandFillColor, opacity)
	}

	// Individual sample paths (clipped to the axis range). Ruin paths (ending
	// at or below 0) are drawn thicker and saturated so failures stand out;
	// the rest are faint grey context.
	for _, s := range samples {
		color, width, op := themeFaint, "1", "0.55"
		if len(s) > 0 && s[len(s)-1] <= 0 {
			color, width, op = themeBad, "1.8", "0.9"
		}
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="%s" stroke-opacity="%s" stroke-linejoin="round" stroke-linecap="round"/>`+"\n",
			linePath(clampAll(s), xAt, yAt), color, width, op)
	}

	// Median line: the central band when the percentile count is odd, finished
	// with an emphasized endpoint so the eye lands on the expected outcome.
	if n%2 == 1 {
		med := clampAll(bands[n/2])
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="`+themeInk+`" stroke-width="2.2" stroke-linejoin="round" stroke-linecap="round"/>`+"\n",
			linePath(med, xAt, yAt))
		if k := len(med) - 1; k >= 0 {
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3" fill="`+themeInk+`"/>`+"\n", xAt(k), yAt(med[k]))
			// Direct label on the one emphasized line, so the dark stroke
			// never needs a legend.
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="11" fill="`+themeInkSoft+`" text-anchor="end">median</text>`+"\n",
				xAt(k)-7, yAt(med[k])-7)
		}
	}
	if clipped {
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" font-size="11" fill="`+themeMuted+`" text-anchor="end">upside clipped ↑</text>`+"\n", x1, y0+12)
	}
	b.WriteString("</svg>")
	return finish(b.String())
}

// fanStart is the common starting wealth of the fan (every band and sample
// begins at the deployed capital); 0 when nothing is drawn.
func fanStart(bands, samples [][]float64) float64 {
	for _, b := range bands {
		if len(b) > 0 {
			return b[0]
		}
	}
	for _, s := range samples {
		if len(s) > 0 {
			return s[0]
		}
	}
	return 0
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
