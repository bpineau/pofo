package firebook

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Figures are inline SVG diagrams generated in Go and themed to the book's
// warm palette, so they need no assets, no network and no build step, and a
// guard test can check that every "::: figure <id>" in an article resolves.
// FigureSVG returns the <svg> for an id, or an empty string for an unknown
// id (the caption still renders). It is exported so alternative frontends
// (a print export, another host app) can reuse the book's plates.

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
	// figRule and every band/area fill below are PRE-BLENDED solid hex, not
	// rgba: crengine (KOReader's EPUB SVG renderer) does not understand
	// rgba and paints any such fill solid black. Each value is the original
	// translucent color composited once onto the figure card background
	// #fffdf9 (bg*(1-a) + color*a per channel), which reads the same on the web
	// card and on near-white EPUB paper. figRule = ink 60,48,34 @ .22.
	figRule = "#D4D0CA"
)

func FigureSVG(id string) string {
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
	"sequence-risk":       figSequenceRisk,
	"cape-swr":            figCapeSWR,
	"retirement-smile":    figSmile,
	"regime-grid":         figRegimeGrid,
	"allocation-plateau":  figAllocationPlateau,
	"withdrawal-frontier": figWithdrawalFrontier,
	"fan-anatomy":         figFanAnatomy,
	"fan-two-plans":       figFanTwoPlans,
	"bond-tent":           figBondTent,
	"wr-signal":           figWrSignal,
	"fat-tails":           figFatTails,
	"horizon-flatten":     figHorizonFlatten,
	"vol-drag":            figVolDrag,
	"franc-decay":         figFrancDecay,
	"correl-sign":         figCorrelSign,
	"buffer-flat":         figBufferFlat,
	"cascade-4pct":        figCascade4pct,
	"utilite-ce":          figUtiliteCE,
	"correlation-vol":     figCorrelationVol,
	"primes-echelle":      figPrimesEchelle,
	"longvol-profil":      figLongvolProfil,
	"carry-courbes":       figCarryCourbes,
	"stacking-expo":       figStackingExpo,
}

// --- 5. The equity-allocation plateau: safe rate vs % equities ---
func figAllocationPlateau() string {
	m := mapper(0, 100, 2.4, 4.3, 66, 582, 286, 44)
	pts := [][2]float64{{0, 2.5}, {20, 3.05}, {30, 3.35}, {40, 3.7}, {50, 3.9}, {60, 4.0}, {70, 4.05}, {80, 4.0}, {90, 3.88}, {100, 3.66}}
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	// plateau wash (50-80 % equities)
	pl0, pl1 := m(50, 4.3), m(80, 2.4)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, pl0[0], pl0[1], pl1[0]-pl0[0], pl1[1]-pl0[1], figWash)
	// axes
	b.WriteString(line(66, 44, 66, 286, figRule, 1))
	b.WriteString(line(66, 286, 582, 286, figRule, 1))
	// 4 % reference line
	four := m(0, 4.0)
	b.WriteString(line(66, four[1], 582, four[1], figRule, 1))
	b.WriteString(txt(70, four[1]-5, 11, figMuted, "start", "400", "règle des 4 %"))
	b.WriteString(poly(px, figAccent, 2.6, ""))
	for _, p := range px {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3" fill="%s"/>`, p[0], p[1], figDeep)
	}
	// zone labels
	left := m(16, 2.75)
	b.WriteString(txt(left[0], left[1], 12, figBad, "middle", "600", "trop peu :"))
	b.WriteString(txt(left[0], left[1]+16, 12, figBad, "middle", "400", "l'érosion"))
	plat := m(65, 4.18)
	b.WriteString(txt(plat[0], plat[1], 12, figGood, "middle", "600", "le plateau"))
	right := m(97, 3.35)
	b.WriteString(txt(right[0], right[1], 12, figBad, "end", "600", "trop de vol :"))
	b.WriteString(txt(right[0], right[1]+16, 12, figBad, "end", "400", "le drag coûte"))
	// x ticks
	for _, c := range []float64{0, 25, 50, 75, 100} {
		p := m(c, 2.4)
		b.WriteString(txt(p[0], 300, 11, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(322, 320, 12, figMuted, "middle", "400", "part d'actions (%)  →"))
	b.WriteString(txt(46, 40, 11, figMuted, "start", "400", "taux sûr"))
	b.WriteString(txt(322, 24, 13, figInk, "middle", "600", "Le taux soutenable plonge des deux côtés du plateau"))
	return svg(620, 340, b.String())
}

// --- 6. The decumulation frontier: ruin vs lived-spending variability ---
func figWithdrawalFrontier() string {
	m := mapper(0, 10, 0, 10, 78, 566, 288, 48)
	var b strings.Builder
	// axes
	b.WriteString(line(78, 40, 78, 288, figRule, 1))
	b.WriteString(line(78, 288, 566, 288, figRule, 1))
	// the frontier arc (Bengen -> % fixe), bowed toward the ideal corner
	arc := [][2]float64{{1.2, 8.6}, {2.2, 5.2}, {3.4, 3.2}, {5.2, 2.1}, {7.2, 1.3}, {9, 0.8}}
	apx := make([][2]float64, len(arc))
	for i, p := range arc {
		apx[i] = m(p[0], p[1])
	}
	b.WriteString(poly(apx, figMuted, 1.6, "5 4"))
	// ideal corner marker (bottom-left = low ruin, low variability)
	ideal := m(0.6, 0.7)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.5" fill="none" stroke="%s" stroke-width="1.4"/>`, ideal[0], ideal[1], figGood)
	b.WriteString(txt(ideal[0]+10, ideal[1]+4, 11, figGood, "start", "600", "l'idéal (inatteignable)"))
	// named rules
	type pt struct {
		x, y   float64
		label  string
		anchor string
		dx, dy float64
		col    string
	}
	rules := []pt{
		{1.2, 8.6, "Bengen (fixe)", "start", 9, 4, figDeep},
		{3.4, 3.2, "guardrails", "start", 9, -8, figAccent},
		{4.3, 2.65, "ABW / TPAW", "end", -10, 16, figAccent},
		{9, 0.8, "% fixe", "end", -10, -8, figDeep},
	}
	for _, r := range rules {
		p := m(r.x, r.y)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, p[0], p[1], r.col)
		b.WriteString(txt(p[0]+r.dx, p[1]+r.dy, 12, r.col, r.anchor, "600", r.label))
	}
	// axis labels
	b.WriteString(txt(88, 44, 11, figMuted, "start", "400", "probabilité de ruine ↑"))
	b.WriteString(txt(322, 312, 12, figMuted, "middle", "400", "variabilité du niveau de vie  →"))
	b.WriteString(txt(322, 26, 13, figInk, "middle", "600", "On ne supprime pas le risque : on choisit sa forme"))
	return svg(620, 340, b.String())
}

// fanArea builds a filled percentile band between lo[] and hi[] (data coords).
func fanArea(m func(x, y float64) [2]float64, xs, lo, hi []float64, fill string) string {
	var sb strings.Builder
	for i := range xs {
		p := m(xs[i], hi[i])
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, "%.1f,%.1f", p[0], p[1])
	}
	for i := len(xs) - 1; i >= 0; i-- {
		p := m(xs[i], lo[i])
		fmt.Fprintf(&sb, " %.1f,%.1f", p[0], p[1])
	}
	return fmt.Sprintf(`<polygon points="%s" fill="%s"/>`, sb.String(), fill)
}

// fanLine maps a per-x data series to a themed polyline.
func fanLine(m func(x, y float64) [2]float64, xs, ys []float64, stroke string, w float64, dash string) string {
	pts := make([][2]float64, len(xs))
	for i := range xs {
		pts[i] = m(xs[i], ys[i])
	}
	return poly(pts, stroke, w, dash)
}

// band fills: warm-accent tints, PRE-BLENDED onto #fffdf9 (see figRule note;
// rgba renders black in crengine). Outer = accent 180,120,60 @ .12, inner
// @ .24; the inner is blended against plain paper (a slight double-blend over
// the outer band, visually negligible) and is drawn AFTER the outer so it
// still reads as the darker core.
const (
	figBandOuter = "#F6EDE2"
	figBandInner = "#EDDDCC"
)

// --- 7. Anatomy of a wealth cone (fan chart) ---
func figFanAnatomy() string {
	xs := []float64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45}
	p05 := []float64{1.00, 0.86, 0.77, 0.72, 0.70, 0.72, 0.77, 0.84, 0.93, 1.03}
	p25 := []float64{1.00, 1.00, 1.03, 1.08, 1.14, 1.21, 1.30, 1.38, 1.46, 1.54}
	p50 := []float64{1.00, 1.10, 1.22, 1.36, 1.50, 1.66, 1.83, 2.00, 2.18, 2.35}
	p75 := []float64{1.00, 1.18, 1.35, 1.52, 1.68, 1.85, 2.02, 2.18, 2.34, 2.50}
	p95 := []float64{1.00, 1.28, 1.55, 1.82, 2.08, 2.35, 2.60, 2.80, 2.95, 3.05}
	green := []float64{1.00, 1.18, 1.42, 1.30, 1.58, 1.86, 2.10, 2.34, 2.55, 2.78}
	red := []float64{1.00, 1.22, 1.10, 0.88, 0.68, 0.52, 0.37, 0.22, 0.10, 0.00}
	m := mapper(0, 47, 0, 3.15, 62, 498, 300, 54)
	var b strings.Builder
	b.WriteString(fanArea(m, xs, p05, p95, figBandOuter))
	b.WriteString(fanArea(m, xs, p25, p75, figBandInner))
	one := m(0, 1)
	b.WriteString(line(62, one[1], 498, one[1], figRule, 1))
	// zero line (the only hard frontier)
	z := m(0, 0)
	b.WriteString(line(62, z[1], 498, z[1], figBad, 1.2))
	b.WriteString(txt(66, z[1]-5, 11, figBad, "start", "600", "ruine (le zéro)"))
	// example paths, then the median on top
	b.WriteString(fanLine(m, xs, green, figGood, 2, ""))
	b.WriteString(fanLine(m, xs, red, figBad, 2, ""))
	rp := m(45, 0)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, rp[0], rp[1], figBad)
	b.WriteString(fanLine(m, xs, p50, figDeep, 2.6, ""))
	// right-gutter direct labels (line ends live at x=45, labels sit beyond)
	me := m(45, 2.35)
	b.WriteString(txt(me[0]+8, me[1]+4, 12, figDeep, "start", "600", "médiane"))
	ge := m(45, 2.78)
	b.WriteString(txt(ge[0]+8, ge[1]+3, 12, figGood, "start", "600", "prospère"))
	b.WriteString(txt(rp[0]+8, rp[1]+3, 12, figBad, "start", "600", "finit ruiné"))
	// band labels, placed inside the cone where it is wide
	ob := m(38, 2.60)
	b.WriteString(txt(ob[0], ob[1], 11, figMuted, "middle", "400", "5-95 %"))
	ib := m(38, 1.82)
	b.WriteString(txt(ib[0], ib[1], 11, figMuted, "middle", "400", "25-75 %"))
	// axes ticks and labels
	for _, c := range []float64{0, 15, 30, 45} {
		p := m(c, 0)
		b.WriteString(txt(p[0], 316, 11, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(280, 336, 12, figMuted, "middle", "400", "années de retraite  →"))
	b.WriteString(txt(62, 48, 11, figMuted, "start", "400", "× capital de départ"))
	b.WriteString(txt(280, 30, 13, figInk, "middle", "600", "Une pile de distributions par date, pas un faisceau de chemins"))
	return svg(620, 356, b.String())
}

// --- 8. Two cones: the first decade decides (defended vs tight plan) ---
func figFanTwoPlans() string {
	xs := []float64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45}
	// defended plan
	dp05 := []float64{1.00, 0.85, 0.76, 0.71, 0.70, 0.72, 0.76, 0.82, 0.89, 0.97}
	dp25 := []float64{1.00, 0.99, 1.00, 1.03, 1.07, 1.12, 1.18, 1.24, 1.30, 1.36}
	dp50 := []float64{1.00, 1.08, 1.18, 1.28, 1.38, 1.48, 1.58, 1.66, 1.74, 1.82}
	dp75 := []float64{1.00, 1.15, 1.28, 1.40, 1.51, 1.61, 1.70, 1.78, 1.85, 1.92}
	dp95 := []float64{1.00, 1.24, 1.46, 1.66, 1.84, 2.00, 2.14, 2.25, 2.34, 2.42}
	// tight plan
	tp05 := []float64{1.00, 0.80, 0.58, 0.38, 0.22, 0.10, 0.03, 0.00, 0.00, 0.00}
	tp25 := []float64{1.00, 0.93, 0.84, 0.74, 0.63, 0.52, 0.41, 0.31, 0.22, 0.14}
	tp50 := []float64{1.00, 1.04, 1.06, 1.06, 1.04, 1.00, 0.95, 0.88, 0.80, 0.72}
	tp75 := []float64{1.00, 1.10, 1.17, 1.21, 1.22, 1.21, 1.18, 1.14, 1.09, 1.04}
	tp95 := []float64{1.00, 1.22, 1.40, 1.55, 1.66, 1.74, 1.80, 1.83, 1.85, 1.86}
	panel := func(px0, px1 float64, p05, p25, p50, p75, p95 []float64, edge, title, note string) string {
		m := mapper(0, 45, 0, 2.55, px0, px1, 290, 66)
		var b strings.Builder
		b.WriteString(fanArea(m, xs, p05, p95, figBandOuter))
		b.WriteString(fanArea(m, xs, p25, p75, figBandInner))
		z := m(0, 0)
		b.WriteString(line(px0, z[1], px1, z[1], figRule, 1))
		one := m(0, 1)
		b.WriteString(line(px0, one[1], px1, one[1], figRule, 1))
		b.WriteString(fanLine(m, xs, p50, figDeep, 1.8, ""))
		// the 5th-percentile lower edge, emphasised in the panel's colour
		b.WriteString(fanLine(m, xs, p05, edge, 2.4, ""))
		b.WriteString(txt((px0+px1)/2, 54, 12, figInk, "middle", "600", title))
		nx := m(22, 0.28)
		b.WriteString(txt(nx[0], nx[1], 11, edge, "middle", "600", note))
		for _, c := range []float64{0, 15, 30, 45} {
			p := m(c, 0)
			b.WriteString(txt(p[0], 306, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
		}
		return b.String()
	}
	var b strings.Builder
	b.WriteString(panel(66, 300, dp05, dp25, dp50, dp75, dp95, figGood, "Plan défendu", "le bas s'enfonce lentement"))
	b.WriteString(panel(336, 570, tp05, tp25, tp50, tp75, tp95, figBad, "Plan tendu", "le 5e percentile pique à zéro"))
	b.WriteString(txt(318, 328, 12, figMuted, "middle", "400", "années  →  (la pente du bas, sur la première décennie, est l'exposition à la séquence)"))
	b.WriteString(txt(318, 30, 13, figInk, "middle", "600", "Le bas du cône, dans les dix premières années, décide"))
	return svg(636, 344, b.String())
}

// figSerif mirrors the book's display font (handler.go --serif), so a figure
// title reads as part of the editorial object rather than a system chart.
const figSerif = `Georgia,'Iowan Old Style',Palatino,'Times New Roman',serif`

func titleTxt(x, y float64, s string) string {
	return fmt.Sprintf(`<text x="%.1f" y="%.1f" font-size="14" fill="%s" text-anchor="middle" font-weight="600" font-family="%s">%s</text>`, x, y, figInk, figSerif, s)
}

// catmull turns a point list into a smooth SVG path (Catmull-Rom -> Bézier):
// curves, not straight segments, are what separate a real chart from a sketch.
func catmull(pts [][2]float64) string {
	if len(pts) < 2 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "M %.1f,%.1f", pts[0][0], pts[0][1])
	n := len(pts)
	for i := 0; i < n-1; i++ {
		p0, p1, p2, p3 := pts[max(0, i-1)], pts[i], pts[i+1], pts[min(n-1, i+2)]
		c1x, c1y := p1[0]+(p2[0]-p0[0])/6, p1[1]+(p2[1]-p0[1])/6
		c2x, c2y := p2[0]-(p3[0]-p1[0])/6, p2[1]-(p3[1]-p1[1])/6
		fmt.Fprintf(&sb, " C %.1f,%.1f %.1f,%.1f %.1f,%.1f", c1x, c1y, c2x, c2y, p2[0], p2[1])
	}
	return sb.String()
}

func smoothStroke(pts [][2]float64, stroke string, w float64) string {
	return fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="%.1f" stroke-linecap="round" stroke-linejoin="round"/>`, catmull(pts), stroke, w)
}

func smoothAreaBelow(pts [][2]float64, baseY float64, fill string) string {
	d := catmull(pts) + fmt.Sprintf(" L %.1f,%.1f L %.1f,%.1f Z", pts[len(pts)-1][0], baseY, pts[0][0], baseY)
	return fmt.Sprintf(`<path d="%s" fill="%s"/>`, d, fill)
}

func dashLine(x1, y1, x2, y2 float64, stroke string, w float64, dash string) string {
	return fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" stroke-dasharray="%s"/>`, x1, y1, x2, y2, stroke, w, dash)
}

// --- 12. Safe rate vs horizon: the curve that flattens ---
func figHorizonFlatten() string {
	pts := [][2]float64{{25, 4.35}, {30, 4.05}, {35, 3.72}, {40, 3.55}, {45, 3.44}, {50, 3.37}, {55, 3.32}, {60, 3.29}, {66, 3.27}}
	m := mapper(25, 66, 3.05, 4.5, 72, 556, 292, 60)
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	// steep-drop wash (30-40 ans)
	s0, s1 := m(30, 4.5), m(40, 3.05)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, s0[0], s0[1], s1[0]-s0[0], s1[1]-s0[1], figWash)
	// asymptote
	asy := m(25, 3.25)
	b.WriteString(dashLine(72, asy[1], 556, asy[1], figMuted, 1, "5 4"))
	b.WriteString(txt(553, asy[1]-16, 10, figMuted, "end", "400", "≈ perpétuité (~3,25 %)"))
	// curve
	b.WriteString(smoothStroke(px, figDeep, 2.8))
	for _, p := range px {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="2.8" fill="%s"/>`, p[0], p[1], figDeep)
	}
	// annotations
	sa := m(35, 4.22)
	b.WriteString(txt(sa[0], sa[1], 11, figAccent, "middle", "600", "chute rapide"))
	b.WriteString(txt(sa[0], sa[1]+14, 10, figAccent, "middle", "400", "(30 → 40 ans)"))
	fa := m(55, 3.62)
	b.WriteString(txt(fa[0], fa[1], 11, figGood, "middle", "600", "au-delà, quasi plat"))
	b.WriteString(txt(fa[0], fa[1]+14, 10, figGood, "middle", "400", "un plan qui tient 40 ans tient (presque) toujours"))
	// y ticks
	for _, v := range []float64{3.5, 4.0, 4.5} {
		p := m(25, v)
		b.WriteString(txt(68, p[1]+4, 10, figMuted, "end", "400", fmt.Sprintf("%.1f %%", v)))
	}
	// x ticks
	for _, c := range []float64{30, 40, 50, 60} {
		p := m(c, 3.05)
		b.WriteString(txt(p[0], 308, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(314, 330, 11, figMuted, "middle", "400", "horizon du plan (années)  →"))
	b.WriteString(txt(54, 56, 10, figMuted, "start", "400", "taux sûr"))
	b.WriteString(titleTxt(314, 30, "Le taux soutenable par horizon : la courbe qui s'aplatit"))
	return svg(620, 344, b.String())
}

// --- 13. Volatility drag: same arithmetic mean, opposite wealth ---
func figVolDrag() string {
	m := mapper(0, 30, 0, 8.6, 72, 548, 292, 58)
	var A, B [][2]float64
	for t := 0; t <= 30; t++ {
		A = append(A, m(float64(t), math.Pow(1.07, float64(t))))
	}
	v := 1.0
	B = append(B, m(0, 1))
	for t := 1; t <= 30; t++ {
		if t%2 == 1 {
			v *= 1.27
		} else {
			v *= 0.87
		}
		B = append(B, m(float64(t), v))
	}
	var b strings.Builder
	b.WriteString(line(72, m(0, 0)[1], 548, m(0, 0)[1], figRule, 1))
	b.WriteString(poly(B, figBad, 1.8, ""))
	b.WriteString(smoothStroke(A, figDeep, 2.6))
	// endpoints
	ea, eb := m(30, math.Pow(1.07, 30)), m(30, v)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, ea[0], ea[1], figDeep)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, eb[0], eb[1], figBad)
	b.WriteString(txt(ea[0]+6, ea[1]+4, 11, figDeep, "start", "600", "7,6×"))
	b.WriteString(txt(eb[0]+6, eb[1]+4, 11, figBad, "start", "600", "4,5×"))
	// legend, top-left in the open area (no labels riding the curves)
	lgx := 96.0
	b.WriteString(line(lgx, 70, lgx+22, 70, figDeep, 2.6))
	b.WriteString(txt(lgx+30, 74, 11, figSoft, "start", "600", "régulier : +7 % chaque année"))
	b.WriteString(line(lgx, 90, lgx+22, 90, figBad, 2.0))
	b.WriteString(txt(lgx+30, 94, 11, figSoft, "start", "600", "volatil : +27 % / −13 % (même moyenne)"))
	// y ticks
	for _, v := range []float64{2, 4, 6, 8} {
		p := m(0, v)
		b.WriteString(txt(68, p[1]+4, 10, figMuted, "end", "400", fmt.Sprintf("%.0f×", v)))
	}
	for _, c := range []float64{0, 10, 20, 30} {
		p := m(c, 0)
		b.WriteString(txt(p[0], 308, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(310, 330, 11, figMuted, "middle", "400", "années  →"))
	b.WriteString(txt(52, 52, 10, figMuted, "start", "400", "richesse (× capital de départ)"))
	b.WriteString(titleTxt(310, 30, "Le volatility drag : même moyenne, richesses opposées"))
	return svg(620, 344, b.String())
}

// --- 14. Purchasing power of the franc, 1914-1958 ---
func figFrancDecay() string {
	pts := [][2]float64{{1914, 100}, {1917, 58}, {1920, 25}, {1923, 27}, {1926, 20}, {1930, 23}, {1935, 25}, {1938, 18}, {1941, 11}, {1944, 6}, {1947, 2.4}, {1950, 1.0}, {1954, 0.8}, {1958, 0.7}}
	m := mapper(1913, 1959, 0, 105, 74, 556, 292, 60)
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	// area first (now an opaque pre-blend, accent 180,120,60 @ .18 on #fffdf9),
	// then the baseline rule on top so it stays visible along the fill's foot.
	b.WriteString(smoothAreaBelow(px, m(1913, 0)[1], "#F2E5D7"))
	b.WriteString(line(74, m(1913, 0)[1], 556, m(1913, 0)[1], figRule, 1))
	b.WriteString(smoothStroke(px, figDeep, 2.6))
	// episode annotations
	e1 := m(1917, 74)
	b.WriteString(txt(e1[0], e1[1], 10, figSoft, "middle", "600", "Grande Guerre"))
	b.WriteString(txt(e1[0], e1[1]+13, 9, figMuted, "middle", "400", "1914-1920"))
	e2 := m(1931, 40)
	b.WriteString(txt(e2[0], e2[1], 10, figSoft, "middle", "600", "entre-deux-guerres"))
	e3 := m(1948, 40)
	b.WriteString(txt(e3[0], e3[1], 10, figSoft, "middle", "600", "guerre et après"))
	b.WriteString(txt(e3[0], e3[1]+13, 9, figMuted, "middle", "400", "1940-1948"))
	// end label
	en := m(1958, 0.7)
	b.WriteString(txt(en[0]-4, en[1]-6, 10, figBad, "end", "600", "≈ 0,7 : plus de 99 % perdu"))
	// y ticks
	for _, v := range []float64{0, 50, 100} {
		p := m(1913, v)
		b.WriteString(txt(70, p[1]+4, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", v)))
	}
	for _, c := range []float64{1920, 1930, 1940, 1950} {
		p := m(c, 0)
		b.WriteString(txt(p[0], 308, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(315, 330, 11, figMuted, "middle", "400", "pouvoir d'achat de 100 francs de 1914, en francs constants"))
	b.WriteString(titleTxt(315, 30, "Le franc, 1914-1958 : la ruine silencieuse du rentier obligataire"))
	return svg(620, 344, b.String())
}

// --- 15. Stock-bond correlation flipping sign with the inflation regime ---
func figCorrelSign() string {
	pts := [][2]float64{{1965, 0.24}, {1970, 0.34}, {1974, 0.30}, {1979, 0.40}, {1984, 0.22}, {1990, 0.12}, {1995, 0.02}, {2000, -0.18}, {2003, -0.40}, {2008, -0.45}, {2012, -0.30}, {2016, -0.36}, {2020, -0.40}, {2021, -0.24}, {2022, 0.34}, {2024, 0.40}}
	m := mapper(1963, 2026, -0.62, 0.62, 76, 556, 300, 58)
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	// background bands: positive (they fall together, bad) vs negative (hedge, good)
	tp, zp, bp := m(1963, 0.62), m(1963, 0), m(2026, -0.62)
	// pre-blended tints on #fffdf9 (rgba renders black in crengine); the two
	// rects tile the top/bottom halves and are the background, drawn before the
	// zero rule, curve and labels. Red 192,101,91 @ .10; green 63,143,111 @ .10.
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#F9EEE9"/>`, tp[0], tp[1], bp[0]-tp[0], zp[1]-tp[1])
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#ECF2EB"/>`, zp[0], zp[1], bp[0]-zp[0], bp[1]-zp[1])
	b.WriteString(line(76, zp[1], 556, zp[1], figRule, 1.2))
	b.WriteString(txt(80, zp[1]-4, 10, figMuted, "start", "400", "corrélation 0"))
	b.WriteString(smoothStroke(px, figInk, 2.6))
	// top pair: same size, baseline-aligned
	yTop := m(1963, 0.5)[1] + 4
	b.WriteString(txt(m(1976, 0)[0], yTop, 11, figSoft, "middle", "600", "ère inflationniste"))
	b.WriteString(txt(552, yTop, 11, figBad, "end", "600", "positive : tombent ensemble"))
	// bottom band label + the golden-age era label (kept off the curve)
	b.WriteString(txt(552, m(1963, -0.5)[1]+4, 11, figGood, "end", "600", "négative : les obligations amortissent"))
	b.WriteString(txt(m(2010, 0)[0], m(2010, -0.13)[1], 11, figSoft, "middle", "600", "l'âge d'or du 60/40"))
	// the 2022 flip, pointing at the rise from clear space to its left
	b.WriteString(txt(m(2021.3, 0)[0], m(2021.3, 0.12)[1], 11, figSoft, "end", "600", "retour 2022 →"))
	// y ticks
	for _, v := range []float64{-0.5, 0.5} {
		p := m(1963, v)
		b.WriteString(txt(72, p[1]+4, 10, figMuted, "end", "400", fmt.Sprintf("%+.1f", v)))
	}
	for _, c := range []float64{1970, 1985, 2000, 2015} {
		p := m(c, -0.62)
		b.WriteString(txt(p[0], 316, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(316, 336, 11, figMuted, "middle", "400", "corrélation glissante actions / obligations  →"))
	b.WriteString(titleTxt(316, 30, "La corrélation actions / obligations change de signe"))
	return svg(620, 350, b.String())
}

// --- 16. The cash-buffer arbitrage: an almost-flat curve that turns up ---
func figBufferFlat() string {
	pts := [][2]float64{{0, 5.05}, {1, 4.78}, {2, 4.62}, {3, 4.62}, {4, 4.72}, {5, 4.88}, {6, 5.05}, {7, 5.22}, {8, 5.38}}
	m := mapper(0, 8, 4.0, 5.7, 78, 540, 288, 70)
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	b.WriteString(line(78, m(0, 4.0)[1], 540, m(0, 4.0)[1], figRule, 1))
	b.WriteString(smoothStroke(px, figAccent, 2.8))
	for _, p := range px {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="2.6" fill="%s"/>`, p[0], p[1], figDeep)
	}
	// min marker + annotation
	mn := m(2.5, 4.6)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, mn[0], mn[1], figGood)
	b.WriteString(txt(mn[0], mn[1]+18, 10, figGood, "middle", "600", "optimum mou (2-3 ans)"))
	// rising-right annotation
	rr := m(6.2, 5.52)
	b.WriteString(txt(rr[0], rr[1], 10, figBad, "middle", "600", "au-delà, le buffer appauvrit le moteur"))
	// flatness note
	b.WriteString(txt(m(0.2, 5.55)[0], m(0.2, 5.55)[1], 10, figSoft, "start", "400", "toute la plage tient en moins d'un point de ruine"))
	// y ticks
	for _, v := range []float64{4.5, 5.0, 5.5} {
		p := m(0, v)
		b.WriteString(txt(74, p[1]+4, 10, figMuted, "end", "400", fmt.Sprintf("%.1f %%", v)))
	}
	for _, c := range []float64{0, 2, 4, 6, 8} {
		p := m(c, 4.0)
		b.WriteString(txt(p[0], 304, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(309, 326, 11, figMuted, "middle", "400", "taille du matelas (années de dépenses)  →"))
	b.WriteString(txt(60, 62, 10, figMuted, "start", "400", "ruine"))
	b.WriteString(titleTxt(309, 30, "L'arbitrage du buffer : une courbe presque plate"))
	return svg(620, 340, b.String())
}

// --- 11. Fat tails: Normal vs Student-t density, same mean and variance ---
func figFatTails() string {
	mu, sigma, nu := 4.0, 15.0, 5.0
	s := sigma * math.Sqrt((nu-2)/nu) // scale so the t variable has std = sigma
	tc := math.Gamma((nu+1)/2) / (math.Sqrt(nu*math.Pi) * math.Gamma(nu/2)) / s
	normPdf := func(x float64) float64 {
		z := (x - mu) / sigma
		return math.Exp(-z*z/2) / (sigma * math.Sqrt(2*math.Pi))
	}
	tPdf := func(x float64) float64 {
		z := (x - mu) / s
		return tc * math.Pow(1+z*z/nu, -(nu+1)/2)
	}
	m := mapper(-44, 52, 0, 0.035, 70, 556, 288, 60)
	var nPts, tPts [][2]float64
	for x := -44.0; x <= 52.0001; x += 2 {
		nPts = append(nPts, m(x, normPdf(x)))
		tPts = append(tPts, m(x, tPdf(x)))
	}
	var b strings.Builder
	base := m(0, 0)[1]
	// fat left-tail fill under the t curve, beyond -30 %. Drawn first (it is an
	// opaque pre-blend now, red 192,101,91 @ .22 on #fffdf9; rgba renders black
	// in crengine), then the baseline rule on top so it stays visible along the
	// fill's foot.
	var tail [][2]float64
	for x := -44.0; x <= -30.0001; x += 1 {
		tail = append(tail, m(x, tPdf(x)))
	}
	var sb strings.Builder
	for i, p := range tail {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, "%.1f,%.1f", p[0], p[1])
	}
	last, first := m(-30, 0), m(-44, 0)
	fmt.Fprintf(&b, `<polygon points="%s %.1f,%.1f %.1f,%.1f" fill="#F1DCD6"/>`, sb.String(), last[0], last[1], first[0], first[1])
	// baseline
	b.WriteString(line(70, base, 556, base, figRule, 1))
	// the two densities
	b.WriteString(poly(nPts, figMuted, 1.8, "5 4"))
	b.WriteString(smoothStroke(tPts, figDeep, 2.6))
	// -30 % marker (label at the top of the marker, clear of the x-axis ticks)
	mk := m(-30, 0)
	top := m(-30, 0.024)
	b.WriteString(dashLine(mk[0], mk[1], mk[0], top[1], figBad, 1, "3 3"))
	b.WriteString(txt(mk[0], top[1]-5, 10, figBad, "middle", "600", "−30 % réel"))
	// fat-tail annotation in the empty upper-left, anchored start (never clipped)
	an := m(-43, 0.020)
	b.WriteString(txt(an[0], an[1], 11, figBad, "start", "600", "la queue épaisse"))
	b.WriteString(txt(an[0], an[1]+14, 10, figBad, "start", "400", "~10× plus d'années à −30 %"))
	// direct curve labels, in open space off the curves
	sl := m(23, 0.019)
	b.WriteString(txt(sl[0], sl[1], 12, figDeep, "start", "600", "Student-t (df 5)"))
	nl := m(32, 0.0075)
	b.WriteString(txt(nl[0], nl[1], 11, figMuted, "start", "400", "loi normale"))
	pk := m(4, 0.0345)
	b.WriteString(txt(pk[0], pk[1], 10, figMuted, "middle", "400", "les années ordinaires se ressemblent"))
	// x ticks
	for _, c := range []float64{-30, -15, 0, 15, 30, 45} {
		p := m(c, 0)
		lab := fmt.Sprintf("%+.0f", c)
		if c == 0 {
			lab = "0"
		}
		b.WriteString(txt(p[0], base+16, 10, figMuted, "middle", "400", lab))
	}
	b.WriteString(txt(313, base+34, 11, figMuted, "middle", "400", "rendement réel annuel (%)  →"))
	b.WriteString(txt(52, 56, 10, figMuted, "start", "400", "densité"))
	b.WriteString(titleTxt(313, 30, "À volatilité égale, deux mondes : la cloche et ses queues"))
	return svg(620, 344, b.String())
}

// --- 10. Current withdrawal rate as a green/amber/red traffic light ---
func figWrSignal() string {
	m := mapper(0, 30, 2.2, 5.8, 66, 556, 300, 56)
	var b strings.Builder
	// zone bands
	band := func(lo, hi float64, fill string) {
		p0, p1 := m(0, hi), m(30, lo)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, p0[0], p0[1], p1[0]-p0[0], p1[1]-p0[1], fill)
	}
	// pre-blended signal tints on #fffdf9 (rgba renders black in crengine);
	// the three bands tile the plot and are the background, drawn before the
	// threshold rules, trajectory and labels. Green @ .13, amber @ .16, red @ .16.
	band(2.2, 4.3, "#E6EFE7")
	band(4.3, 5.2, "#F3E8DB")
	band(5.2, 5.8, "#F5E5E0")
	// threshold rules
	for _, t := range []float64{4.3, 5.2} {
		p := m(0, t)
		b.WriteString(line(66, p[1], 556, p[1], figRule, 1))
	}
	// the current-WR trajectory
	wr := [][2]float64{{0, 3.6}, {3, 3.8}, {5, 4.4}, {7, 5.05}, {8, 5.3}, {10, 4.9}, {12, 4.4}, {14, 3.85}, {17, 3.4}, {21, 3.0}, {25, 2.7}, {30, 2.5}}
	px := make([][2]float64, len(wr))
	for i, p := range wr {
		px[i] = m(p[0], p[1])
	}
	b.WriteString(smoothStroke(px, figInk, 2.8))
	// event dots
	peak := m(8, 5.3)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, peak[0], peak[1], figBad)
	b.WriteString(txt(peak[0], peak[1]-9, 10, figBad, "middle", "600", "2 points de suite → coupe"))
	back := m(13.4, 4.3)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="%s"/>`, back[0], back[1], figGood)
	b.WriteString(txt(back[0]+6, back[1]+14, 10, figGood, "start", "600", "coupe levée"))
	// right-edge zone action labels (the paliers)
	zg, zo, zr := m(30, 3.65), m(30, 4.78), m(30, 5.55)
	b.WriteString(txt(zg[0]-6, zg[1]+4, 11, figGood, "end", "600", "vert · rien"))
	b.WriteString(txt(zo[0]-6, zo[1]+4, 11, figDeep, "end", "600", "orange · vigilance"))
	b.WriteString(txt(zr[0]-6, zr[1]+4, 11, figBad, "end", "600", "rouge · coupe écrite"))
	// y ticks
	for _, v := range []float64{3, 4, 5} {
		p := m(0, v)
		b.WriteString(txt(62, p[1]+4, 10, figMuted, "end", "400", fmt.Sprintf("%.0f %%", v)))
	}
	// x ticks
	for _, c := range []float64{0, 10, 20, 30} {
		p := m(c, 2.2)
		b.WriteString(txt(p[0], 316, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", c)))
	}
	b.WriteString(txt(311, 336, 11, figMuted, "middle", "400", "années de retraite  →"))
	b.WriteString(txt(46, 52, 10, figMuted, "start", "400", "taux courant"))
	b.WriteString(titleTxt(311, 30, "Le taux de retrait courant : le voyant qu'on pilote"))
	return svg(620, 350, b.String())
}

// --- 9. The bond tent: prudence concentrated on the fragile window ---
func figBondTent() string {
	eq := [][2]float64{{-8, 85}, {-6, 85}, {-4, 81}, {-2, 71}, {0, 57}, {2, 60}, {4, 67}, {6, 74}, {8, 81}, {10, 87}, {12, 90}, {15, 90}}
	m := mapper(-8, 15, 0, 100, 72, 556, 300, 58)
	px := make([][2]float64, len(eq))
	for i, p := range eq {
		px[i] = m(p[0], p[1])
	}
	var b strings.Builder
	tl, br := m(-8, 100), m(15, 0)
	// bonds fill the whole plot; equity area on top -> the bond band bulges at
	// departure, and that bulge IS the "tent". Both fills are pre-blended onto
	// #fffdf9 (rgba renders black in crengine): the bonds rect (muted
	// 135,124,109 @ .13) is the background, the equity area (accent
	// 180,120,60 @ .20, blended against plain paper) is drawn after as the
	// warmer core, and the frame lines and equity boundary land on top of both.
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#EFECE7"/>`, tl[0], tl[1], br[0]-tl[0], br[1]-tl[1])
	b.WriteString(smoothAreaBelow(px, br[1], "#F0E2D3"))
	// frame lines at 0 and 100 %
	b.WriteString(line(tl[0], tl[1], br[0], tl[1], figRule, 1))
	b.WriteString(line(tl[0], br[1], br[0], br[1], figRule, 1))
	// departure marker
	d0, d1 := m(0, 0), m(0, 100)
	b.WriteString(dashLine(d0[0], d0[1], d1[0], d1[1], figSoft, 1.1, "4 3"))
	b.WriteString(txt(d1[0], d1[1]-7, 11, figInk, "middle", "600", "départ"))
	// equity boundary (the hero line) + trough dot
	b.WriteString(smoothStroke(px, figDeep, 2.8))
	tr := m(0, 57)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, tr[0], tr[1], figDeep)
	// direct area labels
	ea := m(-5, 36)
	b.WriteString(txt(ea[0], ea[1], 13, figDeep, "middle", "600", "actions"))
	oa := m(1.6, 82)
	b.WriteString(txt(oa[0], oa[1], 13, figMuted, "middle", "600", "obligations"))
	// y ticks
	for _, v := range []struct {
		p float64
		s string
	}{{100, "100 %"}, {50, "50 %"}, {0, "0"}} {
		p := m(-8, v.p)
		b.WriteString(txt(68, p[1]+4, 10, figMuted, "end", "400", v.s))
	}
	// x ticks
	for _, c := range []float64{-8, -4, 0, 4, 8, 12} {
		p := m(c, 0)
		lab := fmt.Sprintf("%+.0f", c)
		if c == 0 {
			lab = "0"
		}
		b.WriteString(txt(p[0], 313, 10, figMuted, "middle", "400", lab))
	}
	// fragile-window bracket under the axis
	fl, fr := m(-2, 0), m(10, 0)
	b.WriteString(line(fl[0], 322, fr[0], 322, figAccent, 1.4))
	b.WriteString(line(fl[0], 322, fl[0], 318, figAccent, 1.4))
	b.WriteString(line(fr[0], 322, fr[0], 318, figAccent, 1.4))
	b.WriteString(txt((fl[0]+fr[0])/2, 336, 10, figAccent, "middle", "600", "fenêtre fragile (risque de séquence)"))
	b.WriteString(txt(556, 336, 10, figMuted, "end", "400", "années · 0 = départ  →"))
	b.WriteString(titleTxt(310, 30, "La tente obligataire : la prudence concentrée là où le danger est"))
	return svg(620, 350, b.String())
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
