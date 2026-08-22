package chart

import (
	"fmt"
	"math"
	"strings"
)

// Bar is one labelled column of a Bars chart (or one row of an HBars chart).
// Text, when set, is drawn at the bar's tip as a value label (the caller
// formats it, e.g. "34%"). Color overrides the default accent when a bar is
// qualitatively apart from its neighbours (a "ruined" bucket among wealth
// bands); leave it empty for an ordinary series.
type Bar struct {
	Label string
	Value float64
	Text  string
	Color string
}

// Bars renders a vertical bar chart in the same instrument dialect as Line
// and MultiLine: a chart surface, a title in ink, mono labels in muted grey,
// and one calm accent hue for the whole series (color is identity, and there
// is only one thing here).
//
// A chart whose every bar carries a Text needs no y axis: the numbers already
// sit on the marks, and a grid behind them would be decoration. Bars drops
// the axis in that case and keeps only the baseline. Where any value label is
// missing, the y-axis grid comes back and carries the reading instead. One or
// the other, never both.
func Bars(opt Options, bars []Bar) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	labelled := len(bars) > 0
	for _, b := range bars {
		if b.Text == "" {
			labelled = false
			break
		}
	}
	marginL, marginR := 56.0, 20.0
	if labelled {
		marginL = 20 // no tick column to make room for
	}
	top, bottom := 28.0, 36.0
	if opt.Title != "" {
		top = 52
	}
	if opt.XLabel != "" {
		bottom += 18
	}
	x0, x1 := marginL, float64(w)-marginR
	y0, y1 := top, float64(h)-bottom

	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`" font-size="12">`+"\n", w, h, w, h)
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="%g" y="24" font-size="16" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", x0, esc(opt.Title))
	}
	n := len(bars)
	if n == 0 {
		sb.WriteString(`</svg>`)
		return finish(sb.String())
	}

	maxV := 0.0
	for _, b := range bars {
		maxV = math.Max(maxV, b.Value)
	}
	if maxV <= 0 {
		maxV = 1
	}
	// Headroom above the tallest bar so its value label is not glued to the
	// frame; without labels the axis is stepped instead.
	vmax := maxV
	step := niceStep(maxV, 5)
	if labelled {
		vmax = maxV / 0.88
	} else {
		vmax = math.Ceil(maxV/step) * step
	}
	yAt := func(v float64) float64 { return y1 - v/vmax*(y1-y0) }

	if !labelled {
		for v := 0.0; v <= vmax+step/1e6; v += step {
			y := yAt(v)
			fmt.Fprintf(&sb, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="`+themeGrid+`"/>`+"\n", x0, y, x1, y)
			fmt.Fprintf(&sb, `<text x="%g" y="%.1f" dy="0.35em" font-size="11" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, step))
		}
	}

	slot := (x1 - x0) / float64(n)
	bw := math.Min(slot*0.56, 72)
	for i, b := range bars {
		cx := x0 + (float64(i)+0.5)*slot
		y := yAt(math.Max(b.Value, 0))
		col := b.Color
		if col == "" {
			col = themeAccent
		}
		tip := b.Label
		if b.Text != "" {
			tip += ": " + b.Text
		}
		// The whole column strip is a hit target, so a thin bar is reachable
		// and a hover layer has a surface to read from.
		fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="transparent"><title>%s</title></rect>`+"\n",
			cx-slot/2, y0, slot, y1-y0, esc(tip))
		if hgt := y1 - y; hgt >= 0.5 {
			fmt.Fprintf(&sb, `<path d="%s" fill="%s"><title>%s</title></path>`+"\n",
				topRounded(cx-bw/2, y, bw, hgt, 4), col, esc(tip))
		}
		if b.Text != "" {
			fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="12" fill="`+themeInkSoft+`">%s</text>`+"\n", cx, y-7, esc(b.Text))
		}
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="12" fill="`+themeMuted+`">%s</text>`+"\n", cx, y1+17, esc(b.Label))
	}
	// Baseline, and the axis labels.
	fmt.Fprintf(&sb, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="`+themeAxis+`"/>`+"\n", x0, y1, x1, y1)
	if opt.XLabel != "" {
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, y1+37, esc(opt.XLabel))
	}
	if opt.YLabel != "" && !labelled {
		fmt.Fprintf(&sb, `<text x="14" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle" transform="rotate(-90 14 %.1f)">%s</text>`+"\n",
			(y0+y1)/2, (y0+y1)/2, esc(opt.YLabel))
	}

	// Hover payload: the plot box plus one column anchor per bar, so the
	// pointer is answered anywhere over the plot, not only on the ink.
	yl := opt.YLabel
	if yl == "" {
		yl = opt.Title
	}
	hm := hoverMeta{Kind: "bars", Axis: "x", XLabel: opt.XLabel, YLabel: yl,
		X0: x0, X1: x1, Y0: y0, Y1: y1}
	vals := hoverSeries{Name: "value"}
	for i, b := range bars {
		hm.Rows = append(hm.Rows, b.Label)
		// The anchor is the bar's top, kept inside the plot box: Bars charts a
		// non-negative quantity, and a stray negative must not put a mark (and
		// with it a hover band) outside the plot.
		y := math.Min(math.Max(yAt(b.Value), y0), y1)
		hm.Marks = append(hm.Marks, hoverMark{X: x0 + (float64(i)+0.5)*slot, Y: y, Text: b.Text})
		vals.Ys = append(vals.Ys, b.Value)
	}
	hm.Series = append(hm.Series, vals)
	sb.WriteString(hoverBlock(hm))
	sb.WriteString(`</svg>`)
	return finish(sb.String())
}

// topRounded is the path of a bar rounded at its tip and square on its
// baseline: a data end that reads as a mark rather than as a pill, and that
// stays honest about where the value stops.
func topRounded(x, y, w, h, r float64) string {
	r = math.Min(r, math.Min(w/2, h))
	return fmt.Sprintf("M%.1f %.1fV%.1fQ%.1f %.1f %.1f %.1fH%.1fQ%.1f %.1f %.1f %.1fV%.1fZ",
		x, y+h, y+r, x, y, x+r, y, x+w-r, x+w, y, x+w, y+r, y+h)
}
