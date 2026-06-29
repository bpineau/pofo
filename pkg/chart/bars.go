package chart

import (
	"fmt"
	"strings"
)

// Bar is one labelled column of a Bars chart. Text, when set, is drawn above
// the bar as a value label (the caller formats it, e.g. "34%").
type Bar struct {
	Label string
	Value float64
	Text  string
}

// Bars renders a vertical bar chart as a standalone SVG, in the same visual
// style as Line. Bars are drawn left to right; the y-axis spans 0 to the
// largest value, with light gridlines and tick labels so the heights are
// readable, and each bar's Text (when set) above it.
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
	yAt := func(v float64) float64 { return float64(padT+plotH) - v/max*float64(plotH) }
	// Y-axis gridlines and tick labels, from 0 to the largest value.
	step := niceStep(max, 5)
	for v := 0.0; v <= max+step/1e6; v += step {
		y := yAt(v)
		fmt.Fprintf(&sb, `<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#e6e6e6"/>`, padL, y, w-padR, y)
		fmt.Fprintf(&sb, `<text x="%d" y="%.1f" dy="0.35em" font-size="11" fill="#555" text-anchor="end">%s</text>`, padL-6, y, fmtTick(v, step))
	}
	bw := float64(plotW) / float64(n) * 0.7
	gap := float64(plotW) / float64(n)
	for i, b := range bars {
		x := float64(padL) + float64(i)*gap + (gap-bw)/2
		y := yAt(b.Value)
		fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, bw, float64(padT+plotH)-y, PaletteColor(0))
		if b.Text != "" {
			fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="11" fill="#333">%s</text>`, x+bw/2, y-4, esc(b.Text))
		}
		fmt.Fprintf(&sb, `<text x="%.1f" y="%d" text-anchor="middle">%s</text>`, x+bw/2, padT+plotH+15, esc(b.Label))
	}
	sb.WriteString(`</svg>`)
	return sb.String()
}
