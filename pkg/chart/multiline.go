package chart

import (
	"fmt"
	"math"
	"strings"
)

// Marker is an optional reference line on a MultiLine chart: a vertical line at
// an x value (Axis 'x') or a horizontal line at a y value (Axis 'y'), with a
// short label.
type Marker struct {
	Axis  rune // 'x' for a vertical line, 'y' for a horizontal line
	Value float64
	Label string
}

// MultiLine renders several numeric-x line series on shared x and y axes, with
// optional reference markers. It is meant for overlaying comparable curves, e.g.
// the ruin-vs-withdrawal-rate frontier of each model, with a vertical marker at
// the user's current rate and a horizontal one at the target ruin.
func MultiLine(opt Options, xLabel, yLabel string, series []XYSeries, markers ...Marker) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	marginL, marginR, top, bottom := 64.0, 16.0, 64.0, 48.0
	x0, x1 := marginL, float64(w)-marginR
	y0, y1 := top, float64(h)-bottom

	var xs, ys [][]float64
	for i := range series {
		if series[i].Color == "" {
			series[i].Color = PaletteColor(i)
		}
		xs = append(xs, series[i].Xs)
		ys = append(ys, series[i].Ys)
	}
	xmin, xmax := spanBounds(xs...)
	if xmin == xmax {
		xmax = xmin + 1
	}
	ymin, ymax := axisBounds(flatten(ys))

	xAt := func(x float64) float64 { return x0 + (x-xmin)/(xmax-xmin)*(x1-x0) }
	yAt := func(v float64) float64 { return y1 - (v-ymin)/(ymax-ymin)*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, Segoe UI, Helvetica, Arial, sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#FFFFFF"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="#16181D">%s</text>`+"\n", x0, esc(opt.Title))
	}

	// Horizontal grid and y labels.
	ystep := niceStep(ymax-ymin, 6)
	for v := math.Ceil(ymin/ystep) * ystep; v <= ymax+ystep/1e6; v += ystep {
		y := yAt(v)
		fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="#EDF0F3"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="#7A8294" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, ystep))
	}
	// X ticks and label.
	xstep := niceStep(xmax-xmin, 8)
	for v := math.Ceil(xmin/xstep) * xstep; v <= xmax+xstep/1e6; v += xstep {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#7A8294" text-anchor="middle">%s</text>`+"\n", xAt(v), y1+16, fmtTick(v, xstep))
	}
	if xLabel != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#7A8294" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, y1+34, esc(xLabel))
	}
	if yLabel != "" {
		fmt.Fprintf(&b, `<text x="14" y="%.1f" font-size="12" fill="#7A8294" text-anchor="middle" transform="rotate(-90 14 %.1f)">%s</text>`+"\n", (y0+y1)/2, (y0+y1)/2, esc(yLabel))
	}
	// Axes.
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#CDD2DA"/>`+"\n", x0, y1, x1, y1)
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#CDD2DA"/>`+"\n", x0, y0, x0, y1)

	// Reference markers (dashed), drawn under the series.
	for _, m := range markers {
		switch m.Axis {
		case 'x':
			x := xAt(m.Value)
			fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="#A8AEBC" stroke-dasharray="4 3"/>`+"\n", x, y0, x, y1)
			fmt.Fprintf(&b, `<text x="%.1f" y="%g" font-size="11" fill="#A8AEBC" text-anchor="middle">%s</text>`+"\n", x, y0-2, esc(m.Label))
		case 'y':
			y := yAt(m.Value)
			fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="#A8AEBC" stroke-dasharray="4 3"/>`+"\n", x0, y, x1, y)
			fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="-2" font-size="11" fill="#A8AEBC" text-anchor="end">%s</text>`+"\n", x1, y, esc(m.Label))
		}
	}

	for _, s := range series {
		drawXY(&b, s, xAt, yAt)
	}

	// Legend.
	lx := x0
	for _, s := range series {
		if s.Name == "" {
			continue
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="36" width="12" height="12" rx="2" fill="%s"/>`, lx, s.Color)
		fmt.Fprintf(&b, `<text x="%.1f" y="46" font-size="12" fill="#16181D">%s</text>`+"\n", lx+17, esc(s.Name))
		lx += 17 + 7.2*float64(len([]rune(s.Name))) + 18
	}
	b.WriteString("</svg>")
	return b.String()
}

// flatten concatenates slices into one.
func flatten(ss [][]float64) []float64 {
	var out []float64
	for _, s := range ss {
		out = append(out, s...)
	}
	return out
}
