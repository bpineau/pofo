package firebook

import (
	"fmt"
	"math"
	"strings"
)

// This file holds the second generation of book figures ("plates"), with a
// tighter editorial system than the first batch: a left-aligned serif title
// under a letterspaced kicker, axis labels in the UI sans (Instrument Sans),
// numbers in the UI mono (Spline Sans Mono), hairline grids, thin marks with
// rounded data ends, and direct labels instead of legends wherever possible.
// The palette is the book's, with a categorical trio validated for CVD and
// normal-vision separation on the card surface: amber, blue, red (+ green as
// a third stacked segment, always paired with gaps and direct labels).
// Labels are always horizontal (house rule).

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

// --- 17. The 4 % cascade: real return + amortization − sequence penalty ---
func figCascade4pct() string {
	m := mapper(0, 1, 0, 6.4, 0, 1, 330, 76)
	y := func(v float64) float64 { return m(0, v)[1] }
	var b strings.Builder
	b.WriteString(plateHead("les maths du 4 %", "Du rendement au taux de retrait : la cascade"))
	// horizontal grid + mono ticks
	for _, g := range []float64{0, 2, 4, 6} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(56, gy, 616, gy, col, 1))
		b.WriteString(mTxt(48, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f %%", g)))
	}
	// bars
	type step struct {
		v0, v1 float64
		fill   string
		val    string
		l1, l2 string
	}
	steps := []step{
		{0, 4.0, figAccent, "4,0 %", "rendement réel", "géométrique du 60/40"},
		{4.0, 5.8, figBlue, "+1,8", "bonus d'amortissement", "(consommer le capital, 30 ans)"},
		{5.8, 4.0, figBad, "−1,8", "pénalité de séquence", "(survivre au pire ordre)"},
		{0, 4.0, figDeep, "4,0 %", "taux de retrait sûr", "(la règle de Bengen)"},
	}
	bw, span := 84.0, 560.0/4
	for i, s := range steps {
		cx := 56 + span*(float64(i)+0.5)
		x := cx - bw/2
		b.WriteString(barV(x, bw, y(s.v0), y(s.v1), s.fill))
		top := math.Min(y(s.v0), y(s.v1))
		b.WriteString(mTxt(cx, top-8, 11.5, figInk, "middle", "600", s.val))
		b.WriteString(sTxt(cx, 352, 11, figSoft, "middle", "600", s.l1))
		b.WriteString(sTxt(cx, 366, 10.5, figMuted, "middle", "400", s.l2))
		// connector to the next bar
		if i < len(steps)-1 {
			yc := y(s.v1)
			nx := 56 + span*(float64(i)+1.5) - bw/2
			b.WriteString(dashLine(x+bw, yc, nx, yc, figMuted, 1, "3 3"))
		}
	}
	return svg(640, 384, b.String())
}

// --- 18. Concave utility and the certainty equivalent ---
func figUtiliteCE() string {
	// x: wealth 10..70 k€ ; y: utility ln(w), plotted on [ln 12, ln 74]
	u := math.Log
	m := mapper(10, 70, u(12), u(74), 64, 600, 316, 78)
	pt := func(w float64) [2]float64 { return m(w, u(w)) }
	var b strings.Builder
	b.WriteString(plateHead("décider sous incertitude", "L'équivalent certain : ce que vaut vraiment un plan risqué"))
	// axes
	b.WriteString(line(64, 70, 64, 316, figRule, 1))
	b.WriteString(line(64, 316, 600, 316, figRule, 1))
	b.WriteString(sTxt(70, 80, 10.5, figMuted, "start", "400", "utilité (bien-être)"))
	b.WriteString(sTxt(598, 338, 11, figMuted, "end", "400", "revenu annuel (k€)"))
	// the utility curve
	var curve [][2]float64
	for w := 12.0; w <= 70.01; w += 2 {
		curve = append(curve, pt(w))
	}
	b.WriteString(smoothStroke(curve, figAccent, 2.2))
	b.WriteString(sTxt(96, 104, 10.5, figMuted, "start", "400", "l'utilité croît de"))
	b.WriteString(sTxt(96, 118, 10.5, figMuted, "start", "400", "moins en moins vite"))
	// the 50/50 lottery: 20 or 65
	w1, w2 := 20.0, 65.0
	p1, p2 := pt(w1), pt(w2)
	b.WriteString(dashLine(p1[0], p1[1], p2[0], p2[1], figBlue, 1.5, "5 4"))
	for _, p := range [][2]float64{p1, p2} {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, p[0], p[1], figInk)
	}
	b.WriteString(sTxt(p1[0]+9, p1[1]+13, 11, figSoft, "start", "600", "mauvais monde (20 k€)"))
	b.WriteString(sTxt(p2[0]-9, p2[1]-10, 11, figSoft, "end", "600", "bon monde (65 k€)"))
	// expected value of the lottery, on the chord
	ew := (w1 + w2) / 2
	em := m(ew, (u(w1)+u(w2))/2)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, em[0], em[1], figBlue)
	b.WriteString(sTxt(em[0]+11, em[1]+16, 11, figBlue, "start", "600", "la loterie 50/50"))
	// certainty equivalent: same utility, on the curve
	ce := math.Exp((u(w1) + u(w2)) / 2)
	cm := pt(ce)
	b.WriteString(dashLine(em[0], em[1], cm[0], cm[1], figMuted, 1.2, "3 3"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="none" stroke="%s" stroke-width="1.8"/>`, cm[0], cm[1], figDeep)
	// drop lines to the x axis
	b.WriteString(dashLine(em[0], em[1], em[0], 316, figMuted, 1, "2 3"))
	b.WriteString(dashLine(cm[0], cm[1], cm[0], 316, figMuted, 1, "2 3"))
	b.WriteString(mTxt(em[0]+4, 330, 10, figBlue, "start", "600", "E = 42,5"))
	b.WriteString(mTxt(cm[0]-4, 330, 10, figDeep, "end", "600", fmt.Sprintf("ÉC ≈ %.0f", ce)))
	// the price of risk, bracketed under the axis
	b.WriteString(line(cm[0], 342, em[0], 342, figBad, 1.4))
	b.WriteString(line(cm[0], 339, cm[0], 345, figBad, 1.4))
	b.WriteString(line(em[0], 339, em[0], 345, figBad, 1.4))
	b.WriteString(sTxt((cm[0]+em[0])/2, 358, 10.5, figBad, "middle", "600", "le prix du risque"))
	return svg(640, 372, b.String())
}

// --- 19. Two-asset portfolio volatility as a function of correlation ---
func figCorrelationVol() string {
	sig := func(rho float64) float64 { return 20 * math.Sqrt((1+rho)/2) }
	m := mapper(-1, 1, 0, 22, 72, 600, 300, 74)
	var b strings.Builder
	b.WriteString(plateHead("pourquoi la diversification marche", "Le seul levier est la corrélation"))
	// grid
	for _, g := range []float64{0, 5, 10, 15, 20} {
		gy := m(0, g)[1]
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(72, gy, 600, gy, col, 1))
		b.WriteString(mTxt(64, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(72, 64, 10.5, figMuted, "start", "400", "volatilité du panier 50/50 (%)"))
	// each asset's own volatility, as a reference
	ry := m(0, 20)[1]
	b.WriteString(dashLine(72, ry, 600, ry, figRule, 1, "4 4"))
	b.WriteString(sTxt(80, ry-8, 10.5, figMuted, "start", "400", "volatilité de chaque actif (20 %)"))
	// the curve
	var curve [][2]float64
	for r := -1.0; r <= 1.001; r += 0.05 {
		curve = append(curve, m(r, sig(r)))
	}
	b.WriteString(smoothStroke(curve, figAccent, 2.4))
	// markers with mono values
	type mk struct {
		rho    float64
		dx, dy float64
		anchor string
	}
	for _, k := range []mk{{1, -6, -10, "end"}, {0.5, 4, -12, "middle"}, {0, 4, -12, "middle"}, {-0.5, 6, -12, "middle"}, {-1, 12, -8, "start"}} {
		p := m(k.rho, sig(k.rho))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`, p[0], p[1], figDeep)
		lbl := strings.Replace(fmt.Sprintf("%.1f", sig(k.rho)), ".", ",", 1)
		b.WriteString(mTxt(p[0]+k.dx, p[1]+k.dy, 10.5, figDeep, k.anchor, "600", lbl))
	}
	// x ticks
	for _, r := range []float64{-1, -0.5, 0, 0.5, 1} {
		p := m(r, 0)
		lbl := strings.Replace(fmt.Sprintf("%g", r), ".", ",", 1)
		b.WriteString(mTxt(p[0], 316, 10, figMuted, "middle", "400", lbl))
	}
	b.WriteString(sTxt(336, 336, 11, figMuted, "middle", "400", "corrélation ρ entre les deux actifs"))
	// the free-lunch annotation
	b.WriteString(sTxt(212, 250, 11, figGood, "middle", "600", "même rendement moyen,"))
	b.WriteString(sTxt(212, 264, 11, figGood, "middle", "600", "risque en moins : le free lunch"))
	return svg(640, 352, b.String())
}

// --- 20. The risk-premia ladder: what each risk pays above cash ---
func figPrimesEchelle() string {
	x := func(v float64) float64 { return 208 + v*(600-208)/6.5 }
	var b strings.Builder
	b.WriteString(plateHead("primes de risque", "Ce que chaque risque paie au-dessus du cash"))
	type row struct {
		name   string
		lo, hi float64
		fill   string
		note   string
	}
	rows := []row{
		{"Actions mondiales", 4, 6, figAccent, ""},
		{"Suivi de tendance (brut)", 2, 4, figAccent, ""},
		{"Terme (obligations longues)", 1, 2, figAccent, ""},
		{"Crédit IG (net des défauts)", 0.5, 1, figAccent, ""},
		{"Or", 0, 0, "", "pas de prime : une monnaie, pas un risque rémunéré"},
		{"Cash (l'étalon)", 0, 0, "", "zéro réel par définition, négatif en répression"},
	}
	y0, dy, bh := 78.0, 36.0, 16.0
	// vertical grid
	for _, g := range []float64{0, 2, 4, 6} {
		gx := x(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(gx, y0-12, gx, y0+dy*float64(len(rows))-10, col, 1))
		b.WriteString(mTxt(gx, y0+dy*float64(len(rows))+8, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(404, y0+dy*float64(len(rows))+24, 11, figMuted, "middle", "400", "points de rendement réel par an (ordres de grandeur historiques)"))
	for i, r := range rows {
		y := y0 + dy*float64(i)
		b.WriteString(sTxt(198, y+bh-3.5, 11.5, figSoft, "end", "600", r.name))
		if r.fill != "" {
			mid := (r.lo + r.hi) / 2
			b.WriteString(barH(x(0), x(mid), y, bh, r.fill))
			// the plausible range, as a whisker
			wy := y + bh/2
			b.WriteString(line(x(r.lo), wy, x(r.hi), wy, figDeep, 1.6))
			b.WriteString(line(x(r.hi), wy-4, x(r.hi), wy+4, figDeep, 1.6))
			lbl := strings.ReplaceAll(fmt.Sprintf("%g à %g", r.lo, r.hi), ".", ",")
			b.WriteString(mTxt(x(r.hi)+8, y+bh-3.5, 10.5, figDeep, "start", "600", lbl))
		} else {
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="none" stroke="%s" stroke-width="1.6"/>`, x(0), y+bh/2, figMuted)
			b.WriteString(sTxt(x(0)+12, y+bh-3.5, 10.5, figMuted, "start", "400", r.note))
		}
	}
	return svg(640, 356, b.String())
}

// --- 21. The long-vol P&L profile: bleed slowly, spike rarely ---
func figLongvolProfil() string {
	m := mapper(0, 14, 55, 175, 60, 608, 306, 72)
	var b strings.Builder
	b.WriteString(plateHead("long volatility", "La poche de puts : perdre souvent peu, gagner rarement beaucoup"))
	// crisis bands, behind everything
	for _, c := range []struct {
		x0, x1 float64
		lbl    string
	}{{5.7, 6.6, "krach 2008"}, {11.5, 12.2, "krach 2020"}} {
		p0, p1 := m(c.x0, 175), m(c.x1, 55)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, p0[0], p0[1], p1[0]-p0[0], p1[1]-p0[1], figWash)
		b.WriteString(sTxt((p0[0]+p1[0])/2, 66, 10, figMuted, "middle", "600", c.lbl))
	}
	// reference at 100
	ry := m(0, 100)[1]
	b.WriteString(dashLine(60, ry, 608, ry, figRule, 1.2, "4 4"))
	b.WriteString(mTxt(54, ry+3.5, 10, figMuted, "end", "400", "100"))
	// the sleeve's value: linear bleed, sharp spikes
	pts := [][2]float64{{0, 100}, {1.5, 96}, {3, 92}, {4.5, 88}, {5.7, 85}, {6.2, 128}, {6.6, 124}, {8, 116}, {9.5, 108}, {11, 101}, {11.5, 99}, {12, 146}, {12.2, 143}, {13, 138}, {14, 134}}
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	b.WriteString(poly(px, figAccent, 2.2, ""))
	// annotations
	b.WriteString(sTxt(96, 262, 11, figSoft, "start", "600", "le saignement :"))
	b.WriteString(sTxt(96, 276, 10.5, figMuted, "start", "400", "la prime versée, année après année"))
	sp := m(12, 146)
	b.WriteString(sTxt(sp[0]-13, sp[1]-2, 11, figDeep, "end", "600", "le pic : la convexité paie"))
	b.WriteString(sTxt(sp[0]-13, sp[1]+12, 10.5, figMuted, "end", "400", "(à condition de vendre)"))
	// x axis
	b.WriteString(line(60, 306, 608, 306, figRule, 1))
	for _, yv := range []float64{0, 5, 10, 14} {
		p := m(yv, 55)
		b.WriteString(mTxt(p[0], 320, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", yv)))
	}
	b.WriteString(sTxt(334, 338, 11, figMuted, "middle", "400", "années (profil stylisé, base 100)"))
	return svg(640, 352, b.String())
}

// --- 22. Futures curves: backwardation pays the roll, contango charges it ---
func figCarryCourbes() string {
	m := mapper(0, 24, 80, 122, 72, 590, 296, 76)
	var b strings.Builder
	b.WriteString(plateHead("commodity carry", "Deux pentes, deux destins : la courbe des contrats à terme"))
	// axes and the spot level
	sp := m(0, 100)
	b.WriteString(line(72, 76, 72, 296, figRule, 1))
	b.WriteString(line(72, 296, 590, 296, figRule, 1))
	b.WriteString(dashLine(72, sp[1], 590, sp[1], figRule, 1, "4 4"))
	b.WriteString(sTxt(586, sp[1]+14, 10.5, figMuted, "end", "400", "niveau du spot"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, sp[0], sp[1], figInk)
	b.WriteString(sTxt(sp[0]+9, sp[1]-9, 11, figInk, "start", "600", "spot (100)"))
	// the two curves
	var up, down [][2]float64
	for t := 0.0; t <= 24.01; t += 2 {
		up = append(up, m(t, 100+14*(1-math.Exp(-t/10))))
		down = append(down, m(t, 100-13*(1-math.Exp(-t/10))))
	}
	b.WriteString(smoothStroke(up, figBlue, 2.2))
	b.WriteString(smoothStroke(down, figAccent, 2.2))
	// direct labels, in the clear zones above/below each curve
	b.WriteString(sTxt(452, 100, 11.5, figBlue, "start", "600", "contango :"))
	b.WriteString(sTxt(452, 114, 11, figBlue, "start", "400", "rouler coûte"))
	b.WriteString(sTxt(452, 262, 11.5, figDeep, "start", "600", "backwardation :"))
	b.WriteString(sTxt(452, 276, 11, figDeep, "start", "400", "rouler rapporte"))
	// how the roll works, in the empty mid-left band
	b.WriteString(sTxt(262, 208, 10.5, figMuted, "middle", "400", "à l'échéance, chaque contrat"))
	b.WriteString(sTxt(262, 222, 10.5, figMuted, "middle", "400", "converge vers le spot"))
	// x ticks
	for _, t := range []float64{0, 6, 12, 18, 24} {
		p := m(t, 80)
		b.WriteString(mTxt(p[0], 310, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt(331, 330, 11, figMuted, "middle", "400", "échéance du contrat (mois)"))
	return svg(640, 344, b.String())
}

// --- 23. Return stacking: exposure per 100 € invested ---
func figStackingExpo() string {
	m := mapper(0, 1, 0, 165, 0, 1, 330, 84)
	y := func(v float64) float64 { return m(0, v)[1] }
	var b strings.Builder
	b.WriteString(plateHead("return stacking", "Pour 100 € investis : l'exposition, pas la mise"))
	// legend (chips + labels), one quiet row under the title
	lx := 24.0
	chip := func(col, lbl string) {
		fmt.Fprintf(&b, `<rect x="%.1f" y="56" width="10" height="10" rx="2.5" fill="%s"/>`, lx, col)
		b.WriteString(sTxt(lx+15, 65, 10.5, figSoft, "start", "400", lbl))
		lx += 15 + 7.2*float64(len(lbl)) + 22
	}
	chip(figAccent, "actions")
	chip(figBlue, "obligations (via futures au-delà de 100)")
	chip(figGreen, "diversifiants (trend, or)")
	// grid
	for _, g := range []float64{0, 50, 100, 150} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(56, gy, 616, gy, col, 1))
		b.WriteString(mTxt(48, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	// the invested-capital line
	cy := y(100)
	b.WriteString(dashLine(56, cy, 616, cy, figDeep, 1.4, "5 4"))
	b.WriteString(sTxt(60, cy-6, 10.5, figDeep, "start", "600", "100 € investis"))
	type seg struct {
		v    float64
		fill string
	}
	type col struct {
		segs   []seg
		total  string
		l1, l2 string
	}
	cols := []col{
		{[]seg{{60, figAccent}, {40, figBlue}}, "100", "60/40", "classique"},
		{[]seg{{90, figAccent}, {60, figBlue}}, "150", "100 € de fonds 90/60", "(un 60/40 à levier 1,5)"},
		{[]seg{{60, figAccent}, {40, figBlue}, {33, figGreen}}, "133", "67 € de 90/60", "+ 33 € de diversifiants"},
	}
	bw, span := 96.0, 560.0/3
	for i, c := range cols {
		cx := 56 + span*(float64(i)+0.5)
		x := cx - bw/2
		acc := 0.0
		for j, s := range c.segs {
			y0, y1 := y(acc), y(acc+s.v)
			if j < len(c.segs)-1 { // 2px surface gap between stacked segments
				fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y1+1, bw, y0-y1-1, s.fill)
			} else {
				b.WriteString(barV(x, bw, y0+0, y1, s.fill))
			}
			if y0-y1 > 18 {
				b.WriteString(mTxt(cx, (y0+y1)/2+3.5, 10.5, "#fffdf9", "middle", "600", fmt.Sprintf("%.0f", s.v)))
			}
			acc += s.v
		}
		b.WriteString(mTxt(cx, y(acc)-8, 11.5, figInk, "middle", "600", c.total))
		b.WriteString(sTxt(cx, 352, 11, figSoft, "middle", "600", c.l1))
		b.WriteString(sTxt(cx, 366, 10.5, figMuted, "middle", "400", c.l2))
	}
	b.WriteString(sTxt(56, y(160)-6, 10.5, figMuted, "start", "400", "exposition totale (€)"))
	return svg(640, 384, b.String())
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

// --- 24. Bond primer: one rate point, five durations ---
func figDurationChoc() string {
	m := mapper(0, 1, -32, 32, 0, 1, 336, 88)
	y := func(v float64) float64 { return m(0, v)[1] }
	var b strings.Builder
	b.WriteString(plateHead("l'atlas des obligations", "Un point de taux, et le prix bouge de sa duration"))
	legendChips(&b, 56, [][2]string{{figBlue, "taux −1 point : le prix monte"}, {figBad, "taux +1 point : le prix baisse"}})
	// horizontal grid + mono ticks
	for _, g := range []float64{-30, -15, 0, 15, 30} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(56, gy, 616, gy, col, 1))
		lbl := fmt.Sprintf("%+.0f", g)
		if g == 0 {
			lbl = "0"
		}
		b.WriteString(mTxt(48, gy+3.5, 10, figMuted, "end", "400", lbl+" %"))
	}
	type it struct {
		d      float64
		l1, l2 string
	}
	items := []it{
		{0.2, "monétaire", "duration 0,2"},
		{2, "État 2 ans", "duration 2"},
		{7, "aggregate euro", "duration 7"},
		{15, "État 30 ans", "duration 15"},
		{29, "zéro-coupon 30 ans", "duration 29"},
	}
	bw, span := 32.0, 560.0/5
	for i, it := range items {
		cx := 56 + span*(float64(i)+0.5)
		b.WriteString(barV(cx-bw-2, bw, y(0), y(it.d), figBlue))
		b.WriteString(barV(cx+2, bw, y(0), y(-it.d), figBad))
		lbl := strings.Replace(fmt.Sprintf("±%g", it.d), ".", ",", 1)
		b.WriteString(mTxt(cx, y(it.d)-7, 10.5, figInk, "middle", "600", lbl+" %"))
		b.WriteString(sTxt(cx, 358, 11, figSoft, "middle", "600", it.l1))
		b.WriteString(mTxt(cx, 372, 10, figMuted, "middle", "400", it.l2))
	}
	return svg(640, 388, b.String())
}

// --- 25. Bond primer: what each type pays, in real terms ---
func figObligationsRendements() string {
	x := func(v float64) float64 { return 218 + v*(600-218)/4.0 }
	var b strings.Builder
	b.WriteString(plateHead("l'atlas des obligations", "Ce que chaque espèce paie, en réel"))
	type row struct {
		name   string
		lo, hi float64
		fill   string
	}
	rows := []row{
		{"Monétaire", 0, 0.5, figAccent},
		{"État euro 2 ans", 0.5, 1, figAccent},
		{"État euro 10 ans", 1, 1.5, figAccent},
		{"État euro 30 ans", 1, 2, figAccent},
		{"Linkers euro (réel affiché)", 0.5, 1.5, figAccent},
		{"Crédit IG euro", 1.5, 2.5, figAccent},
		{"High yield (corrélé actions)", 2, 3.5, figBad},
	}
	y0, dy, bh := 76.0, 34.0, 15.0
	for _, g := range []float64{0, 1, 2, 3, 4} {
		gx := x(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(gx, y0-12, gx, y0+dy*float64(len(rows))-8, col, 1))
		b.WriteString(mTxt(gx, y0+dy*float64(len(rows))+10, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(409, y0+dy*float64(len(rows))+26, 11, figMuted, "middle", "400", "points de rendement réel par an (ordres de grandeur 2024-2026)"))
	for i, r := range rows {
		y := y0 + dy*float64(i)
		b.WriteString(sTxt(208, y+bh-3.5, 11.5, figSoft, "end", "600", r.name))
		mid := (r.lo + r.hi) / 2
		b.WriteString(barH(x(0), x(mid), y, bh, r.fill))
		wy := y + bh/2
		b.WriteString(line(x(r.lo), wy, x(r.hi), wy, figDeep, 1.6))
		b.WriteString(line(x(r.hi), wy-4, x(r.hi), wy+4, figDeep, 1.6))
		lbl := strings.ReplaceAll(fmt.Sprintf("%g à %g", r.lo, r.hi), ".", ",")
		b.WriteString(mTxt(x(r.hi)+8, y+bh-3.5, 10.5, figDeep, "start", "600", lbl))
	}
	return svg(640, 352, b.String())
}

// --- 26. Bond primer: two shocks, five answers ---
func figObligationsRegimes() string {
	xz := 448.0 // zero axis; scale −45..+30 over 208..608
	x := func(v float64) float64 { return xz + v*(608-208)/75.0 }
	var b strings.Builder
	b.WriteString(plateHead("l'atlas des obligations", "Deux chocs, cinq réponses"))
	legendChips(&b, 56, [][2]string{{figBlue, "choc déflationniste (2008)"}, {figBad, "choc d'inflation (2022)"}})
	type row struct {
		name       string
		defl, infl float64
	}
	rows := []row{
		{"État court", 5, -3},
		{"État long", 25, -40},
		{"Linker court", -1, -2},
		{"Crédit IG", -5, -14},
		{"High yield", -26, -11},
	}
	y0, pitch, bh := 88.0, 43.0, 11.0
	// zero axis
	b.WriteString(line(xz, y0-10, xz, y0+pitch*float64(len(rows))-12, figRule, 1))
	bar := func(v, y float64, fill string) {
		if v >= 0 {
			b.WriteString(barH(xz, x(v), y, bh, fill))
			b.WriteString(mTxt(x(v)+7, y+bh-2, 10.5, figDeep, "start", "600", fmt.Sprintf("%+.0f", v)))
		} else {
			b.WriteString(barHL(x(v), xz, y, bh, fill))
			b.WriteString(mTxt(x(v)-7, y+bh-2, 10.5, figDeep, "end", "600", fmt.Sprintf("%.0f", v)))
		}
	}
	for i, r := range rows {
		y := y0 + pitch*float64(i)
		b.WriteString(sTxt(198, y+bh+3, 11.5, figSoft, "end", "600", r.name))
		bar(r.defl, y, figBlue)
		bar(r.infl, y+bh+3, figBad)
	}
	b.WriteString(sTxt(408, y0+pitch*float64(len(rows))+8, 11, figMuted, "middle", "400", "rendement nominal sur l'année du choc (%, stylisé)"))
	return svg(640, 332, b.String())
}

// --- 27. Managed futures: the trend smile ---
func figTrendSmile() string {
	m := mapper(-50, 50, -14, 32, 64, 600, 296, 72)
	var b strings.Builder
	b.WriteString(plateHead("managed futures", "Le sourire du trend : payé aux deux extrêmes"))
	// axes: y ticks + grid
	for _, g := range []float64{-10, 0, 10, 20, 30} {
		gy := m(0, g)[1]
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(64, gy, 600, gy, col, 1))
		b.WriteString(mTxt(58, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%+.0f", g)))
	}
	// x ticks
	for _, g := range []float64{-50, -25, 0, 25, 50} {
		p := m(g, -14)
		b.WriteString(mTxt(p[0], 312, 10, figMuted, "middle", "400", fmt.Sprintf("%+.0f", g)))
	}
	zx := m(0, 0)[0]
	b.WriteString(dashLine(zx, 72, zx, 296, figRule, 1, "3 4"))
	b.WriteString(sTxt(332, 330, 11, figMuted, "middle", "400", "rendement des actions mondiales sur 12 mois (%, profil stylisé)"))
	b.WriteString(sTxt(64, 62, 10.5, figMuted, "start", "400", "rendement du trend sur la même fenêtre (%)"))
	// the smile
	pts := [][2]float64{{-50, 25}, {-40, 19}, {-30, 13}, {-20, 6}, {-12, 1}, {-6, -2}, {0, -4}, {6, -3}, {12, -1}, {20, 3}, {30, 8}, {40, 13}, {50, 18}}
	px := make([][2]float64, len(pts))
	for i, p := range pts {
		px[i] = m(p[0], p[1])
	}
	b.WriteString(smoothStroke(px, figAccent, 2.4))
	// annotations: the two tails and the middle trough
	l := m(-40, 19)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`, l[0], l[1], figDeep)
	b.WriteString(sTxt(l[0]+12, l[1]-6, 11, figSoft, "start", "600", "les grands krachs"))
	b.WriteString(sTxt(l[0]+12, l[1]+8, 10.5, figMuted, "start", "400", "sont des tendances (2008)"))
	r := m(40, 13)
	b.WriteString(sTxt(r[0]-10, r[1]-16, 11, figSoft, "end", "600", "les grands bulls aussi"))
	b.WriteString(sTxt(r[0]-10, r[1]-2, 10.5, figMuted, "end", "400", "(fin des années 1990)"))
	tr := m(3, -4)
	b.WriteString(sTxt(tr[0]+6, tr[1]+22, 11, figSoft, "start", "600", "le creux : marchés sans direction,"))
	b.WriteString(sTxt(tr[0]+6, tr[1]+36, 10.5, figMuted, "start", "400", "faux départs (2011-2019)"))
	// 2022, off the smile: gains came from rates & energy, not equities
	s := m(-18, 27)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, s[0], s[1], figBlue)
	b.WriteString(dashLine(s[0], s[1]+6, m(-18, 4.8)[0], m(-18, 4.8)[1], figMuted, 1, "2 3"))
	b.WriteString(sTxt(s[0]+10, s[1]+3, 11, figBlue, "start", "600", "2022 : hors du sourire actions,"))
	b.WriteString(sTxt(s[0]+10, s[1]+17, 10.5, figMuted, "start", "400", "gagné sur les taux et l'énergie"))
	return svg(640, 344, b.String())
}

// --- 28. Managed futures: a quarter century of SG Trend, year by year ---
func figTrendAnnees() string {
	m := mapper(2000, 2025, -12, 30, 56, 616, 300, 64)
	y := func(v float64) float64 { return m(2000, v)[1] }
	type yr struct {
		year int
		v    float64
	}
	years := []yr{
		{2000, 6}, {2001, 4}, {2002, 19}, {2003, 9}, {2004, 5}, {2005, 0}, {2006, 6}, {2007, 8},
		{2008, 21}, {2009, -5}, {2010, 7}, {2011, -8}, {2012, -3}, {2013, 3}, {2014, 20}, {2015, 0},
		{2016, -6}, {2017, 2}, {2018, -8}, {2019, 9}, {2020, 3}, {2021, 9}, {2022, 27}, {2023, -4}, {2024, 2},
	}
	var b strings.Builder
	b.WriteString(plateHead("managed futures", "Un quart de siècle de SG Trend, année par année"))
	// the winter wash, behind everything
	w0, w1 := m(2011, 30)[0], m(2020, -12)[0]
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, w0, 64.0, w1-w0, 236.0, figWash)
	b.WriteString(sTxt((w0+w1)/2, 78, 10.5, figMuted, "middle", "600", "l'hiver : ≈ 0 % cumulé"))
	// grid
	for _, g := range []float64{-10, 0, 10, 20, 30} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(56, gy, 616, gy, col, 1))
		b.WriteString(mTxt(48, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%+.0f", g)))
	}
	// bars
	bw := 14.0
	for _, e := range years {
		cx := m(float64(e.year)+0.5, 0)[0]
		if e.v == 0 { // flat year: a quiet tick on the axis
			b.WriteString(line(cx-bw/2, y(0)-1, cx+bw/2, y(0)-1, figMuted, 2))
			continue
		}
		fill := figAccent
		if e.v < 0 {
			fill = figBad
		}
		b.WriteString(barV(cx-bw/2, bw, y(0), y(e.v), fill))
	}
	// direct labels on the memorable years
	for _, e := range []yr{{2008, 21}, {2014, 20}, {2022, 27}} {
		cx := m(float64(e.year)+0.5, 0)[0]
		b.WriteString(mTxt(cx, y(e.v)-6, 10.5, figDeep, "middle", "600", fmt.Sprintf("+%.0f", e.v)))
	}
	b.WriteString(mTxt(m(2018.5, 0)[0], y(-8)+14, 10.5, figBad, "middle", "600", "−8"))
	// x ticks
	for _, t := range []int{2000, 2005, 2010, 2015, 2020, 2024} {
		cx := m(float64(t)+0.5, 0)[0]
		b.WriteString(mTxt(cx, 316, 10, figMuted, "middle", "400", fmt.Sprintf("%d", t)))
	}
	b.WriteString(sTxt(336, 336, 11, figMuted, "middle", "400", "rendement annuel, net de frais (ordres de grandeur)"))
	return svg(640, 352, b.String())
}

// --- 29. Risk-based guardrails: the sensor decides, the income follows ---
func figGuardrailsCapteur() string {
	xr := func(rev float64) float64 { return 96 + (rev-1)*(596-96)/6.5 }
	// top panel: success probability, 70..103 % over py 200..70
	ys := func(v float64) float64 { return 200 - (v-70)/33*130 }
	// bottom panel: income, 42..57 k€ over py 380..240
	yi := func(v float64) float64 { return 380 - (v-42)/15*140 }
	var b strings.Builder
	b.WriteString(plateHead("guardrails par risque", "Le capteur décide, le revenu suit"))
	// vertical guides at the two confirmed decisions, spanning both panels
	for _, r := range []float64{3, 7} {
		b.WriteString(dashLine(xr(r), 70, xr(r), 380, figMuted, 1, "2 4"))
	}
	// -- top panel --
	b.WriteString(sTxt(96, 60, 10.5, figMuted, "start", "400", "le capteur : probabilité de succès recalculée (%)"))
	// corridor wash between the two guardrails
	fmt.Fprintf(&b, `<rect x="96" y="%.1f" width="500" height="%.1f" fill="%s"/>`, ys(99), ys(85)-ys(99), figWash)
	b.WriteString(sTxt(160, ys(96), 10.5, figMuted, "start", "400", "le corridor : on ne touche à rien"))
	// guardrail thresholds
	b.WriteString(dashLine(96, ys(85), 596, ys(85), figBad, 1.2, "5 4"))
	b.WriteString(sTxt(592, ys(85)+14, 10.5, figBad, "end", "600", "coupe sous 85 %"))
	b.WriteString(dashLine(96, ys(99), 596, ys(99), figGood, 1.2, "5 4"))
	b.WriteString(sTxt(100, ys(99)-7, 10.5, figGood, "start", "600", "hausse au-dessus de 99 %"))
	for _, g := range []float64{70, 85, 99} {
		b.WriteString(mTxt(88, ys(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	// the sensor path
	vals := []float64{93, 82, 76, 88, 91, 99.2, 99.4}
	pts := make([][2]float64, len(vals))
	for i, v := range vals {
		pts[i] = [2]float64{xr(float64(i + 1)), ys(v)}
	}
	b.WriteString(poly(pts, figAccent, 2, ""))
	for i, p := range pts {
		switch i {
		case 1: // first low alert: open circle, on hold
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="#fffdf9" stroke="%s" stroke-width="1.8"/>`, p[0], p[1], figBad)
		case 2: // confirmed: cut
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, p[0], p[1], figBad)
		case 5: // first high alert: open circle
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="#fffdf9" stroke="%s" stroke-width="1.8"/>`, p[0], p[1], figGood)
		case 6: // confirmed: raise
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, p[0], p[1], figGood)
		default:
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s" stroke="#fffdf9" stroke-width="1.4"/>`, p[0], p[1], figDeep)
		}
	}
	b.WriteString(sTxt(xr(3)+10, ys(76)+5, 10.5, figBad, "start", "600", "confirmée : coupe −10 %"))
	b.WriteString(sTxt(xr(7)-2, ys(85)-10, 10.5, figGood, "end", "600", "confirmée : hausse +10 %"))
	// -- bottom panel --
	b.WriteString(sTxt(96, 232, 10.5, figMuted, "start", "400", "le revenu : retrait réel servi (k€)"))
	for _, g := range []float64{44, 48, 52, 56} {
		gy := yi(g)
		b.WriteString(line(96, gy, 596, gy, figGrid, 1))
		b.WriteString(mTxt(88, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	// the floor, never approached
	b.WriteString(dashLine(96, yi(44), 596, yi(44), figRule, 1.4, "5 4"))
	b.WriteString(sTxt(592, yi(44)-6, 10.5, figMuted, "end", "400", "le plancher (44 k€), jamais approché"))
	// the income staircase
	steps := [][2]float64{{1, 54}, {3, 54}, {3, 48.6}, {7, 48.6}, {7, 53.46}, {7.5, 53.46}}
	spx := make([][2]float64, len(steps))
	for i, s := range steps {
		spx[i] = [2]float64{xr(s[0]), yi(s[1])}
	}
	b.WriteString(poly(spx, figDeep, 2.2, ""))
	b.WriteString(mTxt(xr(1)+4, yi(54)-7, 10.5, figDeep, "start", "600", "54,0"))
	b.WriteString(mTxt(xr(5), yi(48.6)+16, 10.5, figDeep, "middle", "600", "48,6 (quatre ans)"))
	b.WriteString(mTxt(xr(7.5), yi(53.46)-7, 10.5, figDeep, "end", "600", "53,5"))
	// shared x axis
	b.WriteString(line(96, 380, 596, 380, figRule, 1))
	for r := 1; r <= 7; r++ {
		b.WriteString(mTxt(xr(float64(r)), 394, 10, figMuted, "middle", "400", fmt.Sprintf("%d", r)))
	}
	b.WriteString(sTxt(346, 412, 11, figMuted, "middle", "400", "revues annuelles (chiffres illustratifs de la table)"))
	return svg(640, 424, b.String())
}

// figMcEntreesVsTirages contrasts the two things one can change in a
// Monte-Carlo: the assumptions and the number of draws. Both blocks share one
// ruin axis, so the eye reads the asymmetry directly, half a point of mu
// scatters the dots across the axis while a tenfold N only shortens a whisker.
//
// The three ruin figures are computed with this repository's own engine
// (decumul.Plan over a scenario.ParametricSource, 400 000 paths): 1 M EUR,
// 32 k EUR/yr real (3.2 %), 35 years, no tax, Student-t sigma 11 % df 5, mu
// 4.5/5.0/5.5 % arithmetic real. The whiskers are the 95 % sampling interval
// 1.96*sqrt(p(1-p)/N) at p = 5.9 %, the very formula the article's callout gives.
func figMcEntreesVsTirages() string {
	x := func(pct float64) float64 { return 150 + pct/10*(596-150) }
	var b strings.Builder
	b.WriteString(plateHead("monte-carlo", "Ce qui déplace la ruine : les hypothèses, pas le nombre de tirages"))

	// shared grid and axis
	for _, g := range []float64{0, 2, 4, 6, 8, 10} {
		b.WriteString(line(x(g), 66, x(g), 296, figGrid, 1))
		b.WriteString(mTxt(x(g), 314, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(line(150, 296, 596, 296, figRule, 1))
	b.WriteString(sTxt(373, 334, 11, figMuted, "middle", "400", "probabilité de ruine (%)"))
	// the central estimate, the spine both blocks are read against
	b.WriteString(dashLine(x(5.93), 66, x(5.93), 296, figMuted, 1, "2 4"))

	// -- block A: the assumption moves -------------------------------------
	b.WriteString(sTxt(24, 78, 10.5, figSoft, "start", "600", "On bouge μ de ±0,5 point (un écart indétectable dans les données), à N = 10 000"))
	muRows := []struct {
		label string
		ruin  float64
		value string
	}{
		{"μ 4,5 %", 8.66, "8,7 %"},
		{"μ 5,0 %", 5.93, "5,9 %"},
		{"μ 5,5 %", 4.05, "4,1 %"},
	}
	for i, r := range muRows {
		y := 102 + float64(i)*24
		b.WriteString(mTxt(138, y+3.5, 10.5, figSoft, "end", "400", r.label))
		// the sampling whisker at N = 10 000, drawn to the same scale as block B
		b.WriteString(line(x(r.ruin-0.46), y, x(r.ruin+0.46), y, figAccent, 1.4))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, x(r.ruin), y, figDeep)
		b.WriteString(mTxt(x(r.ruin+0.46)+9, y+3.5, 10.5, figDeep, "start", "600", r.value))
	}
	// the span bracket: how far the assumption alone carries the answer
	b.WriteString(line(x(4.05), 172, x(8.66), 172, figDeep, 1.2))
	b.WriteString(line(x(4.05), 168, x(4.05), 176, figDeep, 1.2))
	b.WriteString(line(x(8.66), 168, x(8.66), 176, figDeep, 1.2))
	b.WriteString(sTxt(x(7.4), 190, 10.5, figDeep, "middle", "600", "×2,1 sur la ruine"))

	// -- block B: the number of draws moves --------------------------------
	b.WriteString(sTxt(24, 218, 10.5, figSoft, "start", "600", "On multiplie N par dix, à hypothèses figées (μ = 5,0 %)"))
	nRows := []struct {
		label string
		ci    float64
		value string
	}{
		{"N 1 000", 1.46, "± 1,5 pt"},
		{"N 4 000", 0.73, "± 0,7 pt"},
		{"N 10 000", 0.46, "± 0,5 pt"},
	}
	for i, r := range nRows {
		y := 242 + float64(i)*24
		b.WriteString(mTxt(138, y+3.5, 10.5, figSoft, "end", "400", r.label))
		b.WriteString(line(x(5.93-r.ci), y, x(5.93+r.ci), y, figBlue, 1.4))
		b.WriteString(line(x(5.93-r.ci), y-4, x(5.93-r.ci), y+4, figBlue, 1.4))
		b.WriteString(line(x(5.93+r.ci), y-4, x(5.93+r.ci), y+4, figBlue, 1.4))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, x(5.93), y, figBlue)
		b.WriteString(mTxt(x(5.93+r.ci)+9, y+3.5, 10.5, figBlue, "start", "600", r.value))
	}

	b.WriteString(sTxt(24, 356, 10.5, figMuted, "start", "400",
		"1 M€, 32 k€/an réels (3,2 %), 35 ans, Student-t σ 11 %, df 5. Barres : erreur d'échantillonnage à 95 %."))
	return svg(640, 372, b.String())
}
