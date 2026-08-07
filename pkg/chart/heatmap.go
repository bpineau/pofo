package chart

import (
	"fmt"
	"strings"
)

// HeatmapData is a grid of values Z[y][x] in [0,1] over the axes Xs and Ys.
type HeatmapData struct {
	Xs, Ys         []float64
	Z              [][]float64
	XLabel, YLabel string
}

// Heatmap renders Z as a coloured grid (green = low, red = high), suitable
// for a ruin surface (buffer years x expected return). Values are clamped to
// [0,1].
func Heatmap(opt Options, d HeatmapData) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	const padL, padR, padT, padB = 60, 20, 40, 50
	plotW, plotH := w-padL-padR, h-padT-padB
	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" font-family="sans-serif" font-size="12">`, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="20" font-size="14" font-weight="600">%s</text>`, padL, esc(opt.Title))
	}
	ny, nx := len(d.Ys), len(d.Xs)
	if ny == 0 || nx == 0 {
		sb.WriteString(`</svg>`)
		return finish(sb.String())
	}
	cw := float64(plotW) / float64(nx)
	ch := float64(plotH) / float64(ny)
	for j := 0; j < ny; j++ {
		for i := 0; i < nx; i++ {
			x := float64(padL) + float64(i)*cw
			// y axis drawn bottom-up.
			y := float64(padT) + float64(ny-1-j)*ch
			fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, cw+0.5, ch+0.5, heatColor(d.Z[j][i]))
		}
	}
	if d.XLabel != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle">%s</text>`, padL+plotW/2, h-15, esc(d.XLabel))
	}
	sb.WriteString(`</svg>`)
	return finish(sb.String())
}

// heatColor maps v in [0,1] to the risk ramp, green through amber to red,
// interpolating between the theme's status colors at reduced opacity so the
// grid stays readable under labels.
func heatColor(v float64) string {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	// #0C8A47 (safe) -> #C77E17 (caution) -> #D2402F (danger).
	lerp := func(a, b int, t float64) int { return a + int(t*float64(b-a)) }
	var r, g, b int
	if v < 0.5 {
		t := v * 2
		r, g, b = lerp(0x12, 0xF7, t), lerp(0xB7, 0x90, t), lerp(0x6A, 0x09, t)
	} else {
		t := (v - 0.5) * 2
		r, g, b = lerp(0xF7, 0xD9, t), lerp(0x90, 0x2D, t), lerp(0x09, 0x20, t)
	}
	return fmt.Sprintf("#%02x%02x%02x59", r, g, b)
}
