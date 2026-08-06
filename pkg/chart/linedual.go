package chart

import (
	"fmt"
	"math"
	"strings"
)

// XYSeries is a numeric-x line for LineDual: paired Xs and Ys, an optional
// display name and CSS color.
type XYSeries struct {
	Name   string
	Xs, Ys []float64
	Color  string
}

// LineDual renders two numeric-x line series that share the x-axis but each
// scale to their own y-axis (left and right). It is meant for superimposing two
// quantities measured in different units, e.g. ruin % and median terminal
// wealth against buffer years: the two curves together read as one trade-off.
// xLabel titles the shared x-axis; each axis is drawn and labelled in its
// series' color.
func LineDual(opt Options, xLabel string, left, right XYSeries) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	if left.Color == "" {
		left.Color = PaletteColor(0)
	}
	if right.Color == "" {
		right.Color = PaletteColor(1)
	}
	marginL, marginR, top, bottom := 64.0, 70.0, 64.0, 48.0
	x0, x1 := marginL, float64(w)-marginR
	y0, y1 := top, float64(h)-bottom

	xmin, xmax := spanBounds(left.Xs, right.Xs)
	if xmin == xmax {
		xmax = xmin + 1
	}
	lmin, lmax := axisBounds(left.Ys)
	rmin, rmax := axisBounds(right.Ys)

	xAt := func(x float64) float64 { return x0 + (x-xmin)/(xmax-xmin)*(x1-x0) }
	yAtL := func(v float64) float64 { return y1 - (v-lmin)/(lmax-lmin)*(y1-y0) }
	yAtR := func(v float64) float64 { return y1 - (v-rmin)/(rmax-rmin)*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, Segoe UI, Helvetica, Arial, sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#ffffff"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="#14232B">%s</text>`+"\n", x0, esc(opt.Title))
	}

	// Left-axis gridlines and labels (the shared horizontal grid).
	lstep := niceStep(lmax-lmin, 6)
	for v := math.Ceil(lmin/lstep) * lstep; v <= lmax+lstep/1e6; v += lstep {
		y := yAtL(v)
		fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="#E8ECEA"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="%s" text-anchor="end">%s</text>`+"\n", x0-8, y, left.Color, fmtTick(v, lstep))
	}
	// Right-axis labels only (no extra gridlines, to avoid clutter).
	rstep := niceStep(rmax-rmin, 6)
	for v := math.Ceil(rmin/rstep) * rstep; v <= rmax+rstep/1e6; v += rstep {
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="%s" text-anchor="start">%s</text>`+"\n", x1+8, yAtR(v), right.Color, fmtTick(v, rstep))
	}
	// X-axis ticks at each distinct x of the (left) series.
	for _, x := range left.Xs {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#55666E" text-anchor="middle">%s</text>`+"\n", xAt(x), y1+16, esc(fmt.Sprintf("%g", x)))
	}
	if xLabel != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#55666E" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, y1+34, esc(xLabel))
	}
	// Axes: left, right and bottom.
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="%s"/>`+"\n", x0, y0, x0, y1, left.Color)
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="%s"/>`+"\n", x1, y0, x1, y1, right.Color)
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#AEB9B8"/>`+"\n", x0, y1, x1, y1)

	drawXY(&b, left, xAt, yAtL)
	drawXY(&b, right, xAt, yAtR)

	// Legend, each entry in its series' color.
	lx := x0
	for _, s := range []XYSeries{left, right} {
		if s.Name == "" {
			continue
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="36" width="12" height="12" rx="2" fill="%s"/>`, lx, s.Color)
		fmt.Fprintf(&b, `<text x="%.1f" y="46" font-size="12" fill="#14232B">%s</text>`+"\n", lx+17, esc(s.Name))
		lx += 17 + 7.2*float64(len([]rune(s.Name))) + 18
	}
	b.WriteString("</svg>")
	return b.String()
}

// drawXY renders one numeric-x series as a polyline plus point markers.
func drawXY(b *strings.Builder, s XYSeries, xAt, yAt func(float64) float64) {
	var p strings.Builder
	pen := false
	for i := range s.Xs {
		if i >= len(s.Ys) || !isFinite(s.Ys[i]) {
			pen = false
			continue
		}
		cmd := "L"
		if !pen {
			cmd, pen = "M", true
		}
		fmt.Fprintf(&p, "%s%.1f %.1f", cmd, xAt(s.Xs[i]), yAt(s.Ys[i]))
	}
	if p.Len() > 0 {
		fmt.Fprintf(b, `<path d="%s" fill="none" stroke="%s" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"/>`+"\n", p.String(), s.Color)
	}
	for i := range s.Xs {
		if i >= len(s.Ys) || !isFinite(s.Ys[i]) {
			continue
		}
		fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="2.6" fill="%s"/>`, xAt(s.Xs[i]), yAt(s.Ys[i]), s.Color)
	}
}

// spanBounds returns the min and max finite value across the given slices,
// or (0, 1) when none is finite.
func spanBounds(slices ...[]float64) (float64, float64) {
	min, max := math.Inf(1), math.Inf(-1)
	for _, s := range slices {
		for _, v := range s {
			if !isFinite(v) {
				continue
			}
			min, max = math.Min(min, v), math.Max(max, v)
		}
	}
	if math.IsInf(min, 1) {
		return 0, 1
	}
	return min, max
}

// axisBounds returns a padded [min, max] for a y-axis: it starts at 0 for
// non-negative data and adds 5% headroom at the top.
func axisBounds(ys []float64) (float64, float64) {
	min, max := spanBounds(ys)
	if min == max {
		max = min + 1
	}
	if min > 0 {
		min = 0
	}
	return min, max + (max-min)*0.05
}
