package chart

import (
	"fmt"
	"math"
	"strings"
)

// AreaSeries is one layer of a stacked-area chart: values over the x index,
// stacked bottom-up in slice order.
type AreaSeries struct {
	Name   string
	Values []float64
	Color  string // CSS color; picked from the default palette when empty
}

// StackedArea renders layers stacked bottom-up over the x index (year 0..N):
// layer i fills the space between the running sum below it and itself. It is
// meant for part-to-whole shares over time, e.g. the alive-broke-dead
// decomposition of a retirement (pass shares already scaled to percent). The
// fills are solid but soft; a hairline in the surface colour separates
// adjacent layers so they read distinct without borrowing data ink.
func StackedArea(opt Options, xLabel, yLabel string, series []AreaSeries) string {
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

	steps := 0
	for i := range series {
		if series[i].Color == "" {
			series[i].Color = PaletteColor(i)
		}
		steps = max(steps, len(series[i].Values))
	}

	// Cumulative sums: cum[0] is the zero baseline, cum[i] tops layer i-1.
	cum := make([][]float64, len(series)+1)
	cum[0] = make([]float64, steps)
	for i, s := range series {
		row := make([]float64, steps)
		for j := range row {
			v := 0.0
			if j < len(s.Values) {
				v = s.Values[j]
			}
			row[j] = cum[i][j] + v
		}
		cum[i+1] = row
	}
	vmax := 1.0
	for _, v := range cum[len(series)] {
		vmax = math.Max(vmax, v)
	}

	xmax := float64(max(steps-1, 1))
	xAt := func(i int) float64 { return x0 + float64(i)/xmax*(x1-x0) }
	yAt := func(v float64) float64 { return y1 - v/vmax*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, Segoe UI, Helvetica, Arial, sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#FFFFFF"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="#16181D">%s</text>`+"\n", x0, esc(opt.Title))
	}

	// Layers first, so the grid stays visible on top of the soft fills.
	for i := range series {
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="0.55" stroke="#FFFFFF" stroke-width="1"/>`+"\n",
			bandPolygon(cum[i], cum[i+1], xAt, yAt), series[i].Color)
	}

	// Horizontal grid and y labels, drawn over the fills as hairlines.
	ystep := niceStep(vmax, 5)
	for v := 0.0; v <= vmax+ystep/1e6; v += ystep {
		y := yAt(v)
		fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="#EDF0F3" stroke-opacity="0.6"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="#7A8294" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, ystep))
	}
	// X ticks and labels.
	for _, i := range axisTicks(steps) {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#7A8294" text-anchor="middle">%d</text>`+"\n", xAt(i), y1+16, i)
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

	// Legend row.
	lx := x0
	for _, s := range series {
		if s.Name == "" {
			continue
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="36" width="12" height="12" rx="2" fill="%s" fill-opacity="0.75"/>`, lx, s.Color)
		fmt.Fprintf(&b, `<text x="%.1f" y="46" font-size="12" fill="#16181D">%s</text>`+"\n", lx+17, esc(s.Name))
		lx += 17 + 7.2*float64(len([]rune(s.Name))) + 18
	}
	b.WriteString("</svg>")
	return b.String()
}
