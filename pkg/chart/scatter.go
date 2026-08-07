package chart

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// LabeledPoint is one point of a Scatter: a position, a text label and a color.
type LabeledPoint struct {
	X, Y  float64
	Label string
	Color string
}

// Scatter plots a handful of labeled points against two axes, joined in x-order
// by a faint dashed line to read as a trade-off frontier. Axes auto-scale to the
// data with a little headroom; the y-axis starts at 0. It is meant for a few
// annotated points (e.g. one withdrawal policy each), not a dense cloud.
func Scatter(opt Options, xlab, ylab string, pts []LabeledPoint) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 640
	}
	if h == 0 {
		h = 360
	}
	if len(pts) == 0 {
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"></svg>`, w, h, w, h)
	}
	marginL, marginR, top, bottom := 56.0, 120.0, 24.0, 44.0
	x0, x1 := marginL, float64(w)-marginR
	y0, y1 := top, float64(h)-bottom

	xmax, ymax := 0.0, 0.0
	for _, p := range pts {
		xmax = math.Max(xmax, p.X)
		ymax = math.Max(ymax, p.Y)
	}
	xmax *= 1.15
	ymax *= 1.2
	if xmax == 0 {
		xmax = 1
	}
	if ymax == 0 {
		ymax = 1
	}
	xAt := func(v float64) float64 { return x0 + v/xmax*(x1-x0) }
	yAt := func(v float64) float64 { return y1 - v/ymax*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeSans+`">`+"\n", w, h, w, h)
	// Grid + y ticks.
	ystep := niceStep(ymax, 5)
	for v := 0.0; v <= ymax+ystep/1e6; v += ystep {
		y := yAt(v)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="`+themeGrid+`"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="11" font-family="'Spline Sans Mono',monospace" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", x0-6, y, fmtTick(v, ystep))
	}
	// Axis labels.
	fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="11" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, float64(h)-8, esc(xlab))
	fmt.Fprintf(&b, `<text x="14" y="%.1f" font-size="11" fill="`+themeMuted+`" text-anchor="middle" transform="rotate(-90 14 %.1f)">%s</text>`+"\n", (y0+y1)/2, (y0+y1)/2, esc(ylab))

	// Frontier line through the points in x-order.
	ordered := append([]LabeledPoint(nil), pts...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].X < ordered[j].X })
	var path strings.Builder
	for i, p := range ordered {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&path, "%s%.1f %.1f", cmd, xAt(p.X), yAt(p.Y))
	}
	fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="`+themeAxis+`" stroke-width="1.5" stroke-dasharray="4 4"/>`+"\n", path.String())

	// Points and labels.
	for _, p := range pts {
		col := p.Color
		if col == "" {
			col = themeAccent
		}
		px, py := xAt(p.X), yAt(p.Y)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="6" fill="%s"/>`+"\n", px, py, col)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.32em" font-size="12" fill="`+themeInk+`">%s</text>`+"\n", px+11, py, esc(p.Label))
	}
	b.WriteString("</svg>")
	return finish(b.String())
}
