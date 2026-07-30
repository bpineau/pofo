package chart

import (
	"fmt"
	"strings"
)

// Bar is one labelled column of a Bars chart.
type Bar struct {
	Label string
	Value float64
}

// Bars renders a vertical bar chart as a standalone SVG, in the same visual
// style as Line. Bars are drawn left to right; the y-axis spans 0 to the
// largest value.
func Bars(opt Options, bars []Bar) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	const padL, padR, padT, padB = 50, 20, 40, 40
	plotW, plotH := w-padL-padR, h-padT-padB
	max := 0.0
	for _, b := range bars {
		if b.Value > max {
			max = b.Value
		}
	}
	if max == 0 {
		max = 1
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" font-family="sans-serif" font-size="12">`, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="20" font-size="14" font-weight="600">%s</text>`, padL, esc(opt.Title))
	}
	n := len(bars)
	if n == 0 {
		sb.WriteString(`</svg>`)
		return sb.String()
	}
	bw := float64(plotW) / float64(n) * 0.7
	gap := float64(plotW) / float64(n)
	for i, b := range bars {
		bh := b.Value / max * float64(plotH)
		x := float64(padL) + float64(i)*gap + (gap-bw)/2
		y := float64(padT+plotH) - bh
		fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, bw, bh, PaletteColor(0))
		fmt.Fprintf(&sb, `<text x="%.1f" y="%d" text-anchor="middle">%s</text>`, x+bw/2, padT+plotH+15, esc(b.Label))
	}
	sb.WriteString(`</svg>`)
	return sb.String()
}
