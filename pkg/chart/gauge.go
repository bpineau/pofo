package chart

import (
	"fmt"
	"math"
	"strings"
)

// Gauge renders a half-circle valuation gauge: a good-to-bad arc, a needle at
// frac (0 at the left/low end, 1 at the right/high end), a large central value
// and a caption, with left and right end labels. It is used for readings whose
// position on a scale matters as much as the number (e.g. today's CAPE in its
// historical range). frac is clamped to [0,1].
func Gauge(opt Options, value, caption, left, right string, frac float64) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 360
	}
	if h == 0 {
		h = 210
	}
	frac = math.Max(0, math.Min(1, frac))
	cx, cy := float64(w)/2, float64(h)-34
	r := math.Min(cx-24, cy-16)

	// Point on the arc at a gauge fraction f (0 = left = 180deg, 1 = right = 0deg).
	pt := func(f, radius float64) (float64, float64) {
		a := math.Pi * (1 - f)
		return cx + radius*math.Cos(a), cy - radius*math.Sin(a)
	}
	arc := func(radius float64) string {
		x0, y0 := pt(0, radius)
		x1, y1 := pt(1, radius)
		return fmt.Sprintf("M%.1f %.1f A%.1f %.1f 0 0 0 %.1f %.1f", x0, y0, radius, radius, x1, y1)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="'Instrument Sans',system-ui,sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<defs><linearGradient id="gaugeg" x1="0" y1="0" x2="1" y2="0">`+
		`<stop offset="0" stop-color="#0C8A47"/><stop offset="0.5" stop-color="#C77E17"/><stop offset="1" stop-color="#D2402F"/></linearGradient></defs>`+"\n")
	// Track then the coloured arc.
	fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="#EEF0F3" stroke-width="18" stroke-linecap="round"/>`+"\n", arc(r))
	fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="url(#gaugeg)" stroke-width="18" stroke-linecap="round"/>`+"\n", arc(r))

	// Needle.
	nx, ny := pt(frac, r-4)
	fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#16181D" stroke-width="3" stroke-linecap="round"/>`+"\n", cx, cy, nx, ny)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5.5" fill="#16181D"/>`+"\n", cx, cy)

	// Central value and caption.
	fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-family="'Spline Sans Mono',ui-monospace,monospace" font-size="30" font-weight="600" fill="#16181D">%s</text>`+"\n", cx, cy-16, esc(value))
	if caption != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="12" fill="#4A5160">%s</text>`+"\n", cx, cy+4, esc(caption))
	}
	lx, ly := pt(0, r)
	rx, ry := pt(1, r)
	fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="11" fill="#7A8294">%s</text>`+"\n", lx, ly+20, esc(left))
	fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="11" fill="#7A8294">%s</text>`+"\n", rx, ry+20, esc(right))
	b.WriteString("</svg>")
	return b.String()
}
