package chart

import (
	"fmt"
	"strings"
)

// CatBar is one row of a CategoryBars chart: a label, a value (fraction of the
// total width, 0..1), a right-hand value text, and a bar color.
type CatBar struct {
	Label string
	Value float64
	Text  string
	Color string
}

// CategoryBars renders labelled horizontal bars sharing a common 0..1 scale,
// each in its own color, with the row label to the left and a value at the bar
// tip. It suits a small composition (e.g. the causes of failure) where each
// category is qualitatively distinct, so a single-hue scale would mislead.
func CategoryBars(opt Options, bars []CatBar) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 460
	}
	if h == 0 {
		h = 40 + 34*len(bars)
	}
	labelW, valueW, pad := 108.0, 48.0, 12.0
	x0 := labelW
	x1 := float64(w) - valueW - pad
	rowH := 30.0
	top := 16.0

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeSans+`">`+"\n", w, h, w, h)
	for i, bar := range bars {
		y := top + float64(i)*rowH
		col := bar.Color
		if col == "" {
			col = themeAccent
		}
		bw := bar.Value * (x1 - x0)
		if bw < 0 {
			bw = 0
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="20" rx="4" fill="`+themeWell+`"/>`+"\n", x0, y, x1-x0)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="20" rx="4" fill="%s"/>`+"\n", x0, y, bw, col)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.02em" font-size="12" fill="`+themeInkSoft+`" text-anchor="end">%s</text>`+"\n", labelW-10, y+14, esc(bar.Label))
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.02em" font-size="12" font-family="'Spline Sans Mono',monospace" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", x1+8, y+14, esc(bar.Text))
	}
	b.WriteString("</svg>")
	return finish(b.String())
}
