package chart

import (
	"fmt"
	"math"
	"strings"
)

// Slice is one wedge of a Pie. Value is in arbitrary positive units (shares
// are normalized internally); Color is optional and falls back to the
// default palette in slice order.
type Slice struct {
	Label string
	Value float64
	Color string
}

// PieOptions controls a Pie's rendering. Width defaults to 300px; the height
// is derived from the legend length.
type PieOptions struct {
	Title string
	Width int
}

// Pie renders a donut chart with a title and a legend beneath it. Slices of
// non-positive value are dropped; an empty result (no positive value) yields
// an empty string so callers can omit the chart entirely.
func Pie(opt PieOptions, slices []Slice) string {
	w := opt.Width
	if w == 0 {
		w = 300
	}
	// Keep positive slices and resolve each color once (palette in order).
	total := 0.0
	clean := make([]Slice, 0, len(slices))
	for _, s := range slices {
		if s.Value <= 0 {
			continue
		}
		if s.Color == "" {
			s.Color = defaultPalette[len(clean)%len(defaultPalette)]
		}
		clean = append(clean, s)
		total += s.Value
	}
	if total <= 0 {
		return ""
	}

	cx := float64(w) / 2
	cy, outerR, innerR := 100.0, 70.0, 42.0
	legendY := 184.0
	const rowH = 17.0
	h := int(legendY + float64(len(clean))*rowH + 6)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica, Arial, sans-serif">`, w, h, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="16" text-anchor="middle" font-size="13" font-weight="600" fill="#14232B">%s</text>`, cx, esc(opt.Title))
	}

	// Wedges, clockwise from the top (−90°).
	angle := -math.Pi / 2
	for _, s := range clean {
		frac := s.Value / total
		if frac > 0.99999 { // a lone full slice still needs a hairline gap to render
			frac = 0.99999
		}
		next := angle + frac*2*math.Pi
		b.WriteString(donutPath(cx, cy, outerR, innerR, angle, next, s.Color))
		angle = next
	}

	// Legend: swatch + "Label 42%".
	y := legendY
	for _, s := range clean {
		fmt.Fprintf(&b, `<rect x="6" y="%g" width="10" height="10" rx="2" fill="%s"/>`, y-9, s.Color)
		fmt.Fprintf(&b, `<text x="22" y="%g" font-size="12" fill="#55666E">%s</text>`, y, esc(s.Label))
		fmt.Fprintf(&b, `<text x="%g" y="%g" text-anchor="end" font-size="12" fill="#55666E" font-variant-numeric="tabular-nums">%s</text>`, float64(w)-6, y, esc(fmtPctSlice(100*s.Value/total)))
		y += rowH
	}
	b.WriteString(`</svg>`)
	return b.String()
}

// donutPath returns one filled donut wedge between angles a0 and a1 (radians).
func donutPath(cx, cy, outerR, innerR, a0, a1 float64, color string) string {
	large := 0
	if a1-a0 > math.Pi {
		large = 1
	}
	ox0, oy0 := cx+outerR*math.Cos(a0), cy+outerR*math.Sin(a0)
	ox1, oy1 := cx+outerR*math.Cos(a1), cy+outerR*math.Sin(a1)
	ix1, iy1 := cx+innerR*math.Cos(a1), cy+innerR*math.Sin(a1)
	ix0, iy0 := cx+innerR*math.Cos(a0), cy+innerR*math.Sin(a0)
	return fmt.Sprintf(
		`<path d="M%.2f %.2f A%.0f %.0f 0 %d 1 %.2f %.2f L%.2f %.2f A%.0f %.0f 0 %d 0 %.2f %.2f Z" fill="%s"/>`,
		ox0, oy0, outerR, outerR, large, ox1, oy1, ix1, iy1, innerR, innerR, large, ix0, iy0, color)
}

func fmtPctSlice(p float64) string {
	if p < 1 {
		return fmt.Sprintf("%.1f%%", p)
	}
	return fmt.Sprintf("%.0f%%", p)
}
