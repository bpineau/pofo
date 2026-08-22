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
// each in its own color, with the row label to the left and a value at the
// right. It suits a small composition (e.g. the causes of failure) where each
// category is qualitatively distinct, so a single-hue scale would mislead.
//
// Each row sits in a recessive track showing the whole it is a part of, and
// the chart draws in the same mono dialect as the rest of the family: the bar
// is a mark, not a capsule, and the value is read, not shouted.
func CategoryBars(opt Options, bars []CatBar) string {
	const (
		rowH   = 28.0
		barH   = 13.0
		labelW = 112.0
		valueW = 52.0
	)
	top := 14.0
	if opt.Title != "" {
		top = 46
	}
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 460
	}
	if h == 0 {
		h = int(top+rowH*float64(len(bars))) + 12
	}
	x0 := labelW
	x1 := float64(w) - valueW - 12

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`">`+"\n", w, h, w, h)
	// Capture surface: the whole frame answers the pointer, so a hover layer
	// reads a row from anywhere on it and not only off the drawn bar.
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="transparent"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="24" font-size="16" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", 8.0, esc(opt.Title))
	}
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
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%g" rx="2" fill="`+themeWell+`"><title>%s: %s</title></rect>`+"\n", x0, y, x1-x0, barH, esc(bar.Label), esc(bar.Text))
		if bw >= 1 {
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%g" rx="2" fill="%s"><title>%s: %s</title></rect>`+"\n", x0, y, bw, barH, col, esc(bar.Label), esc(bar.Text))
		}
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="12" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", labelW-12, y+barH/2, esc(bar.Label))
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="12" fill="`+themeInkSoft+`">%s</text>`+"\n", x1+10, y+barH/2, esc(bar.Text))
	}
	// Hover payload: the rows' band, so the pointer is answered anywhere on a
	// row (label gutter and value included), not only on the bar.
	hm := hoverMeta{Kind: "cat", Axis: "y", YLabel: opt.Title,
		X0: 0, X1: float64(w), Y0: top, Y1: top + float64(len(bars))*rowH}
	vals := hoverSeries{Name: "share"}
	for i, bar := range bars {
		hm.Rows = append(hm.Rows, bar.Label)
		hm.Marks = append(hm.Marks, hoverMark{X: (x0 + x1) / 2, Y: top + float64(i)*rowH + barH/2, Text: bar.Text})
		vals.Ys = append(vals.Ys, bar.Value)
	}
	hm.Series = append(hm.Series, vals)
	b.WriteString(hoverBlock(hm))
	b.WriteString("</svg>")
	return finish(b.String())
}
