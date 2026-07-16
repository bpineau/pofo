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
	figGrid  = "rgba(60,48,34,.10)"
	figSans  = `'Instrument Sans',-apple-system,'Segoe UI',Roboto,sans-serif`
	figMono  = `'Spline Sans Mono',ui-monospace,Menlo,Consolas,monospace`
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
