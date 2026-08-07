package chart

import (
	"fmt"
	"math"
	"strings"
)

// HBars renders a horizontal signed bar chart: bars extend left (negative) or
// right (positive) from a central zero axis, each row labelled on the left with
// its Text drawn at the bar tip. It suits "sensitivity" displays of a signed
// change: negative bars (here, a fall in ruin) are green, positive ones red.
func HBars(opt Options, bars []Bar) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 60 + 28*len(bars) // height grows with the row count
	}
	const labelW, padR, padT, padB = 150.0, 60.0, 40.0, 16.0
	x0, x1 := labelW, float64(w)-padR
	zero := (x0 + x1) / 2

	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, Segoe UI, Helvetica, Arial, sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="#FFFFFF"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="8" y="22" font-size="14" font-weight="600" fill="#16181D">%s</text>`+"\n", esc(opt.Title))
	}
	if len(bars) == 0 {
		sb.WriteString("</svg>")
		return sb.String()
	}

	ext := 0.0
	for _, b := range bars {
		ext = math.Max(ext, math.Abs(b.Value))
	}
	if ext == 0 {
		ext = 1
	}
	// Reach the bars to 82% of the half-width so the value label outside each tip
	// stays clear of the row-label gutter on the left.
	half := (x1 - x0) / 2 * 0.82
	xAt := func(v float64) float64 { return zero + v/ext*half }

	plotH := float64(h) - padT - padB
	rowH := plotH / float64(len(bars))
	barH := math.Min(rowH*0.6, 20)

	// Zero axis.
	fmt.Fprintf(&sb, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%.1f" stroke="#CDD2DA"/>`+"\n", zero, padT-4, zero, float64(h)-padB)

	for i, b := range bars {
		cy := padT + (float64(i)+0.5)*rowH
		x := xAt(b.Value)
		left := math.Min(zero, x)
		width := math.Abs(x - zero)
		color := "#0C8A47" // negative: reduces ruin (good)
		anchor, tx := "start", x+4
		if b.Value > 0 {
			color = "#D2402F" // positive: increases ruin
		} else {
			anchor, tx = "end", x-4
		}
		fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="2" fill="%s"/>`+"\n",
			left, cy-barH/2, width, barH, color)
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="12" fill="#7A8294" text-anchor="end">%s</text>`+"\n",
			x0-8, cy, esc(b.Label))
		if b.Text != "" {
			fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="12" fill="#16181D" text-anchor="%s">%s</text>`+"\n",
				tx, cy, anchor, esc(b.Text))
		}
	}
	sb.WriteString("</svg>")
	return sb.String()
}
