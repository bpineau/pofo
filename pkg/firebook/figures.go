package firebook

import (
	"fmt"
	"sort"
	"strings"
)

// Figures are inline SVG diagrams generated in Go and themed to the book's
// warm palette, so they need no assets, no network and no build step, and a
// guard test can check that every "::: figure <id>" in an article resolves.
// figureSVG returns the <svg> for an id, or an empty string for an unknown id
// (the caption still renders).

// book palette (mirrors bookCSS)
const (
	figInk    = "#211c16"
	figSoft   = "#4c4438"
	figMuted  = "#877c6d"
	figAccent = "#b4783c"
	figDeep   = "#8a5526"
	figGood   = "#3f8f6f"
	figBad    = "#c0655b"
	figWash   = "#f2ebdd"
	figRule   = "rgba(60,48,34,.22)"
)

func figureSVG(id string) string {
	if f, ok := figures[id]; ok {
		return f()
	}
	return ""
}

// FigureIDs lists the diagrams available (for the guard test).
func FigureIDs() []string {
	ids := make([]string, 0, len(figures))
	for k := range figures {
		ids = append(ids, k)
	}
	sort.Strings(ids)
	return ids
}

var figures = map[string]func() string{
	"sequence-risk":    figSequenceRisk,
	"cape-swr":         figCapeSWR,
	"retirement-smile": figSmile,
	"regime-grid":      figRegimeGrid,
}

// svg wraps content in a responsive, theme-aware <svg>.
func svg(vbW, vbH int, body string) string {
	return fmt.Sprintf(`<svg viewBox="0 0 %d %d" role="img" xmlns="http://www.w3.org/2000/svg" `+
		`font-family="Georgia,serif" style="width:100%%;height:auto;display:block">%s</svg>`, vbW, vbH, body)
}

// poly builds an SVG polyline from (x,y) pixel points.
func poly(pts [][2]float64, stroke string, w float64, dash string) string {
	var sb strings.Builder
	for i, p := range pts {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, "%.1f,%.1f", p[0], p[1])
	}
	d := ""
	if dash != "" {
		d = ` stroke-dasharray="` + dash + `"`
	}
	return fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="%.1f" stroke-linejoin="round" stroke-linecap="round"%s/>`,
		sb.String(), stroke, w, d)
}

func txt(x, y float64, size int, color, anchor, weight, s string) string {
	return fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="%d" fill="%s" text-anchor="%s" font-weight="%s" font-family="ui-sans-serif,system-ui,sans-serif">%s</text>`,
		x, y, size, color, anchor, weight, s)
}

func line(x1, y1, x2, y2 float64, stroke string, w float64) string {
	return fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f"/>`, x1, y1, x2, y2, stroke, w)
}

// mapper returns a function from data coords to pixel coords in a plot box.
func mapper(x0, x1, y0, y1, px0, px1, py0, py1 float64) func(x, y float64) [2]float64 {
	return func(x, y float64) [2]float64 {
		return [2]float64{
			px0 + (x-x0)/(x1-x0)*(px1-px0),
			py0 + (y-y0)/(y1-y0)*(py1-py0),
		}
	}
}

// --- 1. Sequence-of-returns risk: same average, opposite order & outcome ---
func figSequenceRisk() string {
	m := mapper(0, 30, 0, 2.0, 62, 582, 288, 40)
	good := [][2]float64{{0, 1}, {5, 1.28}, {10, 1.55}, {15, 1.78}, {20, 1.92}, {25, 1.28}, {30, 1.42}}
	bad := [][2]float64{{0, 1}, {3, 0.72}, {6, 0.56}, {10, 0.64}, {15, 0.52}, {20, 0.38}, {25, 0.18}, {30, 0.03}}
	toPx := func(pts [][2]float64) [][2]float64 {
		out := make([][2]float64, len(pts))
		for i, p := range pts {
			out[i] = m(p[0], p[1])
		}
		return out
	}
	var b strings.Builder
	// axes
	z := m(0, 0)
	b.WriteString(line(62, 40, 62, 288, figRule, 1))
	b.WriteString(line(62, z[1], 582, z[1], figRule, 1))
	// gridline at 1.0 (starting capital)
	one := m(0, 1)
	b.WriteString(line(62, one[1], 582, one[1], figRule, 1))
	b.WriteString(txt(66, one[1]-5, 11, figMuted, "start", "400", "capital de départ"))
	b.WriteString(poly(toPx(good), figGood, 2.5, ""))
	b.WriteString(poly(toPx(bad), figBad, 2.5, ""))
	// labels at line ends
	ge := m(30, 1.42)
	be := m(30, 0.03)
	b.WriteString(txt(ge[0]-4, ge[1]-6, 12, figGood, "end", "600", "krach tardif : survit"))
	b.WriteString(txt(be[0]-4, be[1]-8, 12, figBad, "end", "600", "krach précoce : ruine"))
	// axis labels
	b.WriteString(txt(322, 320, 12, figMuted, "middle", "400", "années de retraite  →"))
	b.WriteString(txt(322, 24, 13, figInk, "middle", "600", "Deux séquences, même rendement moyen, retraits identiques"))
	return svg(620, 340, b.String())
}

// --- 2. CAPE at start vs safe withdrawal rate ---
func figCapeSWR() string {
	m := mapper(8, 42, 2.8, 6.2, 66, 582, 286, 44)
	pts := [][2]float64{{10, 5.8}, {15, 5.0}, {20, 4.45}, {25, 3.95}, {30, 3.55}, {35, 3.3}, {40, 3.12}}
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	// axes
	b.WriteString(line(66, 44, 66, 286, figRule, 1))
	b.WriteString(line(66, 286, 582, 286, figRule, 1))
	// 4% reference line
	four := m(8, 4.0)
	b.WriteString(line(66, four[1], 582, four[1], figRule, 1))
	b.WriteString(txt(70, four[1]-5, 11, figMuted, "start", "400", "règle des 4 %"))
	b.WriteString(poly(px, figAccent, 2.6, ""))
	for _, p := range px {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="%s"/>`, p[0], p[1], figDeep)
	}
	// zone labels
	cheap := m(12, 5.7)
	rich := m(37, 3.5)
	b.WriteString(txt(cheap[0], cheap[1], 12, figGood, "start", "600", "marché bon marché"))
	b.WriteString(txt(rich[0], rich[1]-10, 12, figBad, "end", "600", "marché cher"))
	// x ticks
	for _, c := range []float64{10, 20, 30, 40} {
		p := m(c, 2.8)
		b.WriteString(txt(p[0], 300, 11, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(322, 320, 12, figMuted, "middle", "400", "CAPE au départ  →"))
	b.WriteString(txt(46, 40, 11, figMuted, "start", "400", "taux sûr"))
	b.WriteString(txt(322, 24, 13, figInk, "middle", "600", "Plus le marché est cher au départ, plus le taux soutenable baisse"))
	return svg(620, 340, b.String())
}

// --- 3. The retirement spending smile ---
func figSmile() string {
	m := mapper(55, 95, 80, 110, 66, 582, 280, 48)
	pts := [][2]float64{{55, 102}, {60, 98}, {65, 93}, {70, 89}, {75, 87}, {80, 88}, {85, 93}, {90, 100}, {95, 108}}
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	b.WriteString(line(66, 48, 66, 280, figRule, 1))
	b.WriteString(line(66, 280, 582, 280, figRule, 1))
	hundred := m(55, 100)
	b.WriteString(line(66, hundred[1], 582, hundred[1], figRule, 1))
	b.WriteString(poly(px, figAccent, 2.6, ""))
	// phase labels
	p1 := m(60, 100)
	p2 := m(75, 87)
	p3 := m(92, 104)
	b.WriteString(txt(p1[0], p1[1]-10, 12, figSoft, "middle", "600", "go-go"))
	b.WriteString(txt(p2[0], p2[1]+22, 12, figSoft, "middle", "600", "ralentissement"))
	b.WriteString(txt(p3[0], p3[1]-6, 12, figSoft, "end", "600", "santé"))
	for _, a := range []float64{55, 65, 75, 85, 95} {
		p := m(a, 80)
		b.WriteString(txt(p[0], 296, 11, figMuted, "middle", "400", fmt.Sprintf("%.0f", a)))
	}
	b.WriteString(txt(322, 316, 12, figMuted, "middle", "400", "âge  →"))
	b.WriteString(txt(40, 44, 11, figMuted, "start", "400", "dépenses réelles"))
	b.WriteString(txt(322, 26, 13, figInk, "middle", "600", "Les dépenses réelles font un « sourire », pas une ligne plate"))
	return svg(620, 340, b.String())
}

// --- 4. Growth x inflation regime grid ---
func figRegimeGrid() string {
	var b strings.Builder
	x0, x1, y0, y1 := 90.0, 560.0, 58.0, 300.0
	xm, ym := (x0+x1)/2, (y0+y1)/2
	// quadrant washes
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="%s"/>`, x0, y0, x1-x0, y1-y0, figWash)
	b.WriteString(line(xm, y0, xm, y1, figRule, 1))
	b.WriteString(line(x0, ym, x1, ym, figRule, 1))
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="none" stroke="%s" stroke-width="1"/>`, x0, y0, x1-x0, y1-y0, figRule)
	quad := func(cx, cy float64, title, win string) {
		b.WriteString(txt(cx, cy-6, 14, figDeep, "middle", "600", title))
		b.WriteString(txt(cx, cy+16, 12, figSoft, "middle", "400", win))
	}
	quad((x0+xm)/2, (y0+ym)/2, "Prospérité", "actions, obligations")
	quad((xm+x1)/2, (y0+ym)/2, "Surchauffe", "matières 1res, or")
	quad((x0+xm)/2, (ym+y1)/2, "Déflation", "obligations longues")
	quad((xm+x1)/2, (ym+y1)/2, "Stagflation", "or, linkers, trend")
	// axis labels
	b.WriteString(txt(x0-10, (y0+ym)/2, 11, figMuted, "end", "400", "croissance +"))
	b.WriteString(txt(x0-10, (ym+y1)/2, 11, figMuted, "end", "400", "croissance −"))
	b.WriteString(txt((x0+xm)/2, y1+22, 11, figMuted, "middle", "400", "inflation basse"))
	b.WriteString(txt((xm+x1)/2, y1+22, 11, figMuted, "middle", "400", "inflation haute"))
	b.WriteString(txt(322, 30, 13, figInk, "middle", "600", "Les quatre régimes : un gagnant par saison"))
	return svg(620, 340, b.String())
}
