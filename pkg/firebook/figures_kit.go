package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The editorial system of the book's figures ("plates"), tighter than the
// first batch of them was: a left-aligned serif title
// under a letterspaced kicker, axis labels in the UI sans (Instrument Sans),
// numbers in the UI mono (Spline Sans Mono), hairline grids, thin marks with
// rounded data ends, and direct labels instead of legends wherever possible.
// The palette is the book's, with a categorical trio validated for CVD and
// normal-vision separation on the card surface: amber, blue, red (+ green as
// a third stacked segment, always paired with gaps and direct labels).
// Labels are always horizontal (house rule).
//
// This file is the toolbox itself: the palette, the type primitives and the
// small utilities every plate draws with. The plates themselves live in
// figures_<theme>.go, one theme per file, and doc.go's "Adding a plate"
// section walks through writing one.

const (
	figBlue  = "#3a6db4"
	figGreen = "#2f9068"
	// figGrid is a PRE-BLENDED solid hex (ink 60,48,34 @ .10 composited onto the
	// figure card background #fffdf9), not rgba: crengine (KOReader's EPUB SVG
	// renderer) paints rgba fills solid black. As a hairline gridline drawn
	// behind the marks, the opaque value reads the same as the old translucent one.
	figGrid = "#ECE9E4"
	figSans = `'Instrument Sans',-apple-system,'Segoe UI',Roboto,sans-serif`
	figMono = `'Spline Sans Mono',ui-monospace,Menlo,Consolas,monospace`
)

// figMinus is the typographic minus sign the book sets negative numbers with.
// The ASCII hyphen stays what Go writes in source ("-4.0"), but nothing drawn
// on a plate uses it: a hyphen is narrower and lighter than the digits it sits
// against, and reads as a dash rather than a sign. frMinus formats with it.
const figMinus = "\u2212"

// sTxt sets a label in the UI sans.
func sTxt(x, y, size float64, color, anchor, weight, s string) string {
	return fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="%.1f" fill="%s" text-anchor="%s" font-weight="%s" font-family="%s">%s</text>`,
		x, y, size, color, anchor, weight, figSans, s)
}

// mTxt sets a number in the UI mono.
func mTxt(x, y, size float64, color, anchor, weight, s string) string {
	return fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="%.1f" fill="%s" text-anchor="%s" font-weight="%s" font-family="%s">%s</text>`,
		x, y, size, color, anchor, weight, figMono, s)
}

// plateHead renders the kicker + left-aligned serif title of a v2 plate.
func plateHead(kicker, title string) string {
	var b strings.Builder
	b.WriteString(sTxt(24, 21, 9.5, figDeep, "start", "600",
		`<tspan letter-spacing="1.8">`+strings.ToUpper(kicker)+`</tspan>`))
	fmt.Fprintf(&b, `<text x="24" y="43" font-size="15.5" fill="%s" text-anchor="start" font-weight="600" font-family="%s">%s</text>`,
		figInk, figSerif, title)
	return b.String()
}

// barV draws a vertical bar with the data end rounded (r), anchored at base.
func barV(x, w, yBase, yEnd float64, fill string) string {
	r := 3.5
	if math.Abs(yBase-yEnd) < r*2 {
		r = math.Abs(yBase-yEnd) / 2
	}
	if yEnd < yBase { // grows upward, round the top
		return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
			x, yBase, x, yEnd+r, x, yEnd, x+r, yEnd, x+w-r, yEnd, x+w, yEnd, x+w, yEnd+r, x+w, yBase, fill)
	}
	// grows downward, round the bottom
	return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		x, yBase, x, yEnd-r, x, yEnd, x+r, yEnd, x+w-r, yEnd, x+w, yEnd, x+w, yEnd-r, x+w, yBase, fill)
}

// barH draws a horizontal bar from x0 with the right (data) end rounded.
func barH(x0, x1, y, h float64, fill string) string {
	r := 3.5
	if x1-x0 < r*2 {
		r = (x1 - x0) / 2
	}
	return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		x0, y, x1-r, y, x1, y, x1, y+r, x1, y+h-r, x1, y+h, x1-r, y+h, x0, y+h, fill)
}

// barHL draws a horizontal bar ending at x1 with the LEFT (data) end rounded,
// for bars growing leftward from a zero axis.
func barHL(x0, x1, y, h float64, fill string) string {
	r := 3.5
	if x1-x0 < r*2 {
		r = (x1 - x0) / 2
	}
	return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		x1, y, x0+r, y, x0, y, x0, y+r, x0, y+h-r, x0, y+h, x0+r, y+h, x1, y+h, fill)
}

// legendChips renders one quiet row of legend chips under the plate title.
func legendChips(b *strings.Builder, y float64, items [][2]string) {
	lx := 24.0
	for _, it := range items {
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="10" height="10" rx="2.5" fill="%s"/>`, lx, y, it[0])
		b.WriteString(sTxt(lx+15, y+9, 10.5, figSoft, "start", "400", it[1]))
		lx += 15 + 6.4*float64(len(it[1])) + 22
	}
}

// mixHex blends two "#rrggbb" colours, t = 0 giving the first and t = 1 the
// second, and returns an opaque hex: the book's figures never ship a
// translucent fill.
func mixHex(a, b string, t float64) string {
	var ar, ag, ab, br, bg, bb int
	fmt.Sscanf(a, "#%02x%02x%02x", &ar, &ag, &ab)
	fmt.Sscanf(b, "#%02x%02x%02x", &br, &bg, &bb)
	mix := func(x, y int) int { return int(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return fmt.Sprintf("#%02x%02x%02x", mix(ar, br), mix(ag, bg), mix(ab, bb))
}

// frMinus formats a number the French way and with the book's typographic
// minus, so a negative value in a plate never falls back to an ASCII hyphen.
func frMinus(v float64, decimals int) string {
	if v < 0 {
		return figMinus + frNum(-v, decimals)
	}
	return frNum(v, decimals)
}

// figScale is the linear value-to-pixel map every plate needs: Min and Max are
// the ends of the value range, Px0 and Px1 the pixels they land on. A vertical
// axis simply passes the BOTTOM pixel as Px0 and the top one as Px1, which
// inverts the direction without a special case.
type figScale struct {
	Min, Max float64
	Px0, Px1 float64
}

// Map places one value on the scale. Values outside [Min, Max] extrapolate,
// which is the caller's business: a plate that lets a series run off its panel
// has to say so in words (see the rate curve of retrait-deux-lectures).
func (s figScale) Map(v float64) float64 {
	if s.Max == s.Min {
		return s.Px0
	}
	return s.Px0 + (v-s.Min)/(s.Max-s.Min)*(s.Px1-s.Px0)
}

// axisTicks draws one value axis: a hairline gridline at every tick and its
// label in the UI mono, in the book's standing convention.
//
// The scale carries the axis; horizontal says which way it runs. A horizontal
// axis (values along x) gets vertical gridlines and its labels centred
// underneath; a vertical one (values along y) gets horizontal gridlines and
// its labels to the left of the plot, right-aligned. from and to are the
// gridlines' extent in the OTHER direction, so a vertical axis passes the
// plot's left and right edges and a horizontal one its top and bottom; passing
// the same value twice draws the labels alone, which is what a plate wants for
// a time axis it does not want ruled.
//
// A tick at exactly zero is drawn in figRule rather than figGrid: the zero of a
// signed plate is a fact, not a guide.
//
// The label is the tick set the book's way: decimals digits after a French
// decimal COMMA, a negative value carrying the typographic minus, and suffix
// appended verbatim for the unit a plate wants on every tick (" %", " ans", or
// "" when an axis title carries it instead). The plates are single-source
// French and the English edition reformats their numbers mechanically, so the
// kit has no business emitting a point decimal or an ASCII hyphen.
func axisTicks(b *strings.Builder, s figScale, ticks []float64, decimals int, suffix string, from, to float64, horizontal bool) {
	for _, v := range ticks {
		p := s.Map(v)
		if from != to {
			col := figGrid
			if v == 0 {
				col = figRule
			}
			if horizontal {
				b.WriteString(line(p, from, p, to, col, 1))
			} else {
				b.WriteString(line(from, p, to, p, col, 1))
			}
		}
		label := frMinus(v, decimals) + suffix
		if horizontal {
			b.WriteString(mTxt(p, math.Max(from, to)+18, 10, figMuted, "middle", "400", label))
		} else {
			b.WriteString(mTxt(math.Min(from, to)-10, p+3.5, 10, figMuted, "end", "400", label))
		}
	}
}

// plateFoot sets the block of notes under a plate: the window, the conventions,
// the caveats, in small grey sans lines from the plate's text column, fifteen
// pixels apart. y is the baseline of the first line, so a plate sizes itself as
// y + 15*len(lines) plus its bottom margin.
//
// The conclusive sentence some plates carry just above these notes is not part
// of the block: it is louder on purpose and stays the plate's own business.
func plateFoot(y float64, lines []string) string {
	var b strings.Builder
	for i, l := range lines {
		b.WriteString(sTxt(24, y+15*float64(i), 9.5, figMuted, "start", "400", l))
	}
	return b.String()
}

// braceH draws the horizontal brace a plate hangs under its axis to measure a
// stretch of it, with its label centred underneath: the idiom that writes "from
// here to there costs this much" without a second axis. The brace's ticks point
// up, toward the thing being measured.
func braceH(x0, x1, y float64, label string) string {
	var b strings.Builder
	b.WriteString(line(x0, y, x1, y, figMuted, 1.2))
	b.WriteString(line(x0, y-5, x0, y, figMuted, 1.2))
	b.WriteString(line(x1, y-5, x1, y, figMuted, 1.2))
	b.WriteString(sTxt((x0+x1)/2, y+16, 10, figSoft, "middle", "600", label))
	return b.String()
}

// dotLabel marks one data point and names it: the dot, its name in the UI sans
// above, and its value in the UI mono just over the dot. anchor is the usual
// SVG text anchor, and the labels are nudged to the side they are anchored on
// so they never sit on top of the mark ("start" puts them to its right, "end"
// to its left, "middle" straight above it).
//
// The value takes the mark's own colour and the name the book's soft ink, so a
// plate can pose a dozen points without turning into a colour chart.
func dotLabel(x, y float64, color, name, value, anchor string) string {
	dx := 0.0
	switch anchor {
	case "start":
		dx = 8
	case "end":
		dx = -8
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, x, y, color)
	b.WriteString(sTxt(x+dx, y-20, 10, figSoft, anchor, "600", name))
	b.WriteString(mTxt(x+dx, y-7, 10.5, color, anchor, "600", value))
	return b.String()
}
