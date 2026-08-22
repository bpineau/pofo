package chart

import (
	"fmt"
	"math"
	"strings"
)

// HBars renders a signed horizontal bar chart (a tornado): each row's bar
// runs left or right of a zero axis, in the diverging good/bad pair, with the
// row named in the left gutter and its value at the bar's tip.
//
// The zero axis sits where the DATA puts it, not at the middle of the frame:
// its position splits the plot in proportion to how far the rows reach each
// way, so a chart where nine levers help and one hurts does not waste half
// its width. Each side that carries a row keeps a floor of the plot, so a
// lone small bar still has room to be seen and labelled.
func HBars(opt Options, bars []Bar) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 60 + 30*len(bars) // height grows with the row count
	}
	// The label gutter fits the longest row label (12px mono runs ~7.3px per
	// glyph), within a third of the chart so the bars keep the floor space.
	labelW := 120.0
	for _, b := range bars {
		labelW = math.Max(labelW, float64(len([]rune(b.Label)))*7.3+16)
	}
	labelW = math.Min(labelW, float64(w)/3)
	top, bottom := 28.0, 20.0
	if opt.Title != "" {
		top = 52
	}
	x0, x1 := labelW+10, float64(w)-20
	y0, y1 := top, float64(h)-bottom

	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`">`+"\n", w, h, w, h)
	// The background doubles as the capture surface: the whole frame answers
	// the pointer, so a hover layer reads a row from anywhere on it.
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="%.1f" y="24" font-size="16" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", 20.0, esc(opt.Title))
	}
	if len(bars) == 0 {
		sb.WriteString("</svg>")
		return finish(sb.String())
	}

	// Where zero falls: in proportion to how far the rows reach each way, on
	// ONE scale for both sides (two scales would make a small opposite bar
	// look like a large one). A gutter on each side holds the tip label of a
	// bar too short to hold it inside.
	negExt, posExt := 0.0, 0.0
	for _, b := range bars {
		negExt = math.Max(negExt, -b.Value)
		posExt = math.Max(posExt, b.Value)
	}
	const gutter = 64.0
	span := negExt + posExt
	scale := 0.0
	if span > 0 {
		scale = math.Max(x1-x0-2*gutter, 1) / span
	}
	zero := x0 + gutter + negExt*scale
	xAt := func(v float64) float64 { return zero + v*scale }

	rowH := (y1 - y0) / float64(len(bars))
	barH := math.Min(rowH*0.46, 14)

	// Zero axis: the reference every row is read against, so it is drawn as an
	// axis and not as a hairline lost among the bars.
	fmt.Fprintf(&sb, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="`+themeAxis+`"/>`+"\n", zero, y0-4, zero, y1)

	for i, b := range bars {
		cy := y0 + (float64(i)+0.5)*rowH
		x := xAt(b.Value)
		left, width := math.Min(zero, x), math.Abs(x-zero)
		color := themeGood // negative: reduces ruin (good)
		if b.Value > 0 {
			color = themeBad // positive: increases ruin
		}
		if width >= 1 {
			// Softened fills: a tornado is mostly one colour, and at full
			// strength that colour becomes the loudest thing on the page.
			fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="2" fill="%s" fill-opacity="0.82"/>`+"\n",
				left, cy-barH/2, width, barH, color)
		}
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="12" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n",
			x0-10, cy, esc(b.Label))
		if b.Text == "" {
			continue
		}
		// The value rides INSIDE a bar long enough to hold it, in the surface
		// colour, and just past the tip otherwise: it can then never collide
		// with the row label, whatever the bar's length.
		tx, anchor, fill := x, "start", themeInkSoft
		if b.Value > 0 {
			tx += 6
		} else {
			tx, anchor = x-6, "end"
		}
		if width > float64(len([]rune(b.Text)))*7.3+14 {
			fill = themeSurface
			if b.Value > 0 {
				tx, anchor = x-6, "end"
			} else {
				tx, anchor = x+6, "start"
			}
		}
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="12" fill="%s" text-anchor="%s">%s</text>`+"\n",
			tx, cy, fill, anchor, esc(b.Text))
	}
	if opt.XLabel != "" {
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, float64(h)-4, esc(opt.XLabel))
	}

	// Hover payload: one anchor per row, so the pointer is answered anywhere
	// on a row rather than only over the (often short) bar.
	hm := hoverMeta{Kind: "bars", Axis: "y", XLabel: opt.XLabel, YLabel: opt.Title,
		X0: 0, X1: float64(w), Y0: y0, Y1: y1}
	vals := hoverSeries{Name: "change"}
	for i, b := range bars {
		hm.Rows = append(hm.Rows, b.Label)
		hm.Marks = append(hm.Marks, hoverMark{X: xAt(b.Value), Y: y0 + (float64(i)+0.5)*rowH, Text: b.Text})
		vals.Ys = append(vals.Ys, b.Value)
	}
	hm.Series = append(hm.Series, vals)
	sb.WriteString(hoverBlock(hm))
	sb.WriteString("</svg>")
	return finish(sb.String())
}
