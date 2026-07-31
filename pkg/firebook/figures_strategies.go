package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The plates of the withdrawal-strategies part. Each one carries the central
// claim of an article that had, until then, only stated it in prose, and each
// takes the form its own question asks for rather than a house recipe: a
// warning light that climbs for years (Bengen), a plane of constant incomes
// crossed two different ways (the CAPE rule), two curves that swap ranks at a
// nameable age (annuities), a narrowing procedure (the decision), and four
// instrument strips read side by side (the two thermometers).
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.

// Pre-blended solid fills, never rgba: crengine (KOReader's EPUB SVG renderer)
// paints rgba solid black. Each is its colour composited once onto the figure
// card background #fffdf9. figBadWash = figBad at .10, figBlueWash = figBlue at
// .10, figDeepWash = figDeep at .09.
const (
	figBadWash  = "#F9EEE9"
	figBlueWash = "#EBEEF2"
	figDeepWash = "#F5EFE9"
)

// arrow draws a straight arrow, the head sitting at the far end.
func arrow(x0, y0, x1, y1 float64, stroke string, w float64, dash string) string {
	dx, dy := x1-x0, y1-y0
	n := math.Hypot(dx, dy)
	if n == 0 {
		return ""
	}
	ux, uy := dx/n, dy/n
	const h = 7.0
	bx, by := x1-ux*h, y1-uy*h
	var b strings.Builder
	b.WriteString(dashLine(x0, y0, bx, by, stroke, w, dash))
	fmt.Fprintf(&b, `<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		x1, y1, bx-uy*3.4, by+ux*3.4, bx+uy*3.4, by-ux*3.4, stroke)
	return b.String()
}

// --- Bengen's silent cliff: the warning light that climbs for nineteen years ---

// bengenRates is the current withdrawal rate (spend / portfolio, %) the 1966
// household reads every 1st of January, and bengenWealth what is still standing
// on that date, in k EUR. Both come from pkg/replay on the bundled real US
// 60/40, for the book's reference household: 600 k EUR, 24 k EUR a year (4.0 %)
// indexed to inflation, never adjusted. figures_strategies_test.go recomputes
// them and fails if either drifts.
//
// The rates run 1966 to 1993, the wealth 1966 to 1993 as well (the reading of
// January 1994 is the last one: 18.2 k EUR left against a 24 k EUR plan, so the
// year is only part paid and the money is gone).
var (
	bengenStart  = 1966
	bengenRuin   = 1994
	bengenRates  = []float64{4.00, 4.52, 4.26, 4.29, 5.03, 5.11, 4.99, 4.82, 5.97, 8.23, 7.67, 7.28, 8.73, 10.00, 11.23, 11.82, 14.20, 13.82, 14.42, 15.92, 15.46, 15.74, 18.82, 21.57, 23.17, 31.37, 37.80, 58.49}
	bengenWealth = []float64{600.0, 531.5, 563.5, 559.3, 477.0, 469.3, 480.6, 497.6, 401.7, 291.7, 312.8, 329.8, 274.8, 240.0, 213.8, 203.0, 169.0, 173.7, 166.4, 150.7, 155.3, 152.5, 127.5, 111.3, 103.6, 76.5, 63.5, 41.0}
)

func figBengenFalaise() string {
	const (
		x0, x1 = 74.0, 600.0
		yTop   = 76.0 // 24 %
		yBot   = 286.0
		vMax   = 24.0
	)
	n := float64(bengenRuin - bengenStart) // 28 January readings, 1966 to 1994
	x := func(year int) float64 { return x0 + float64(year-bengenStart)/n*(x1-x0) }
	y := func(v float64) float64 { return yBot - v/vMax*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("le voyant du retrait fixe",
		"La falaise silencieuse : dix-neuf ans de préavis"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"taux de retrait courant lu chaque 1er janvier (retrait de l'année / portefeuille), en %"))

	// The two severity bands the article names, drawn under everything else.
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, yTop, x1-x0, y(10)-yTop, figBadWash)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, y(10), x1-x0, y(8)-y(10), figWash)
	b.WriteString(sTxt(x0+8, y(9)+3.5, 10, figDeep, "start", "600", "la zone d'alerte : 8 à 10 %"))
	b.WriteString(sTxt(x0+8, yTop+16, 10, figBad, "start", "600", "au-delà, aucun marché ne rattrape"))

	for _, g := range []float64{0, 4, 8, 12, 16, 20, 24} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(x0, gy, x1, gy, col, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}

	// The reading itself, up to the last year that still fits the frame.
	var pts [][2]float64
	for i, v := range bengenRates {
		if v > vMax {
			break
		}
		pts = append(pts, [2]float64{x(bengenStart + i), y(v)})
	}
	b.WriteString(poly(pts, figAccent, 2.4, ""))

	// Year one, and the year the light turns red.
	p0 := pts[0]
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, p0[0], p0[1], figDeep)
	b.WriteString(mTxt(p0[0]+8, p0[1]+4, 10.5, figDeep, "start", "600", "4,0 %"))

	alert := 1975 // the first January read above 8 %
	ax, ay := x(alert), y(bengenRates[alert-bengenStart])
	b.WriteString(dashLine(ax, ay, ax, yBot, figBad, 1, "2 3"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, ax, ay, figBad)
	b.WriteString(sTxt(ax+9, ay-10, 10.5, figBad, "start", "600", "janvier 1975 : 8,2 %"))

	// The end of the frame is not the end of the story.
	last := pts[len(pts)-1]
	rx := x(bengenRuin)
	b.WriteString(dashLine(last[0], last[1], last[0], yTop, figAccent, 1.6, "3 3"))

	b.WriteString(line(rx, yTop, rx, yBot, figBad, 1.4))
	b.WriteString(sTxt(rx-8, y(6), 10.5, figAccent, "end", "600", "1991 à 1993 : 31, 38 puis 58 %"))
	b.WriteString(sTxt(rx-8, y(3), 10.5, figBad, "end", "600", "1994 : plus rien"))

	// Years, then what was still standing on those Januaries.
	for _, yr := range []int{1966, 1970, 1975, 1980, 1985, 1990, 1994} {
		b.WriteString(mTxt(x(yr), 302, 10, figMuted, "middle", "400", fmt.Sprintf("%d", yr)))
	}
	b.WriteString(sTxt(24, 324, 10.5, figMuted, "start", "400",
		"capital encore là ce jour-là, en k€ réels"))
	for _, yr := range []int{1966, 1970, 1975, 1980, 1985, 1990} {
		b.WriteString(mTxt(x(yr), 340, 10.5, figSoft, "middle", "600",
			frNum(bengenWealth[yr-bengenStart], 0)))
	}
	b.WriteString(mTxt(x(1994), 340, 10.5, figBad, "middle", "600", "0"))

	// The punchline, bracketed under the whole thing.
	by := 358.0
	b.WriteString(line(ax, by, rx, by, figBad, 1.2))
	b.WriteString(line(ax, by-4, ax, by+4, figBad, 1.2))
	b.WriteString(line(rx, by-4, rx, by+4, figBad, 1.2))
	b.WriteString(sTxt((ax+rx)/2, by+17, 10.5, figBad, "middle", "600",
		"dix-neuf ans entre le premier signal et le zéro"))
	return svg(640, 384, b.String())
}

// --- The CAPE rule's double countercyclicality, on the plane of constant incomes ---

// capeRate is ERN's rule: a + b / CAPE, in percent per year.
func capeRate(cape float64) float64 { return 1.75 + 0.5*(100/cape) }

func figCapeContracyclique() string {
	const (
		pLo, pHi = 0.9, 2.25
		wLo, wHi = 2.6, 4.5
		left     = 96.0
		right    = 560.0
		top      = 96.0
		bottom   = 288.0
	)
	m := mapper(pLo, pHi, wLo, wHi, left, right, bottom, top)
	// income in k EUR of a portfolio of p M EUR served at w percent
	inc := func(p, w float64) float64 { return 10 * p * w }

	var b strings.Builder
	b.WriteString(plateHead("la double contracyclicité",
		"Le revenu, produit de deux facteurs en sens inverse"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"un portefeuille de 1,5 M€ à CAPE 30, frappé par un krach puis par une euphorie"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"les courbes fines joignent les couples qui servent le même revenu, en k€ par an"))

	// The plane: iso-income hyperbolas, w = R / (10 P).
	for r := 30.0; r <= 70.1; r += 10 {
		var pts [][2]float64
		for p := pLo; p <= pHi+0.001; p += 0.03 {
			w := r / (10 * p)
			if w > wHi || w < wLo {
				continue
			}
			pts = append(pts, m(p, w))
		}
		if len(pts) < 2 {
			continue
		}
		b.WriteString(smoothStroke(pts, figGrid, 1.4))
	}

	// Axes.
	b.WriteString(line(left, bottom, right, bottom, figRule, 1))
	for _, p := range []float64{1.0, 1.5, 2.0} {
		q := m(p, wLo)
		b.WriteString(line(q[0], bottom, q[0], bottom+4, figRule, 1))
		b.WriteString(mTxt(q[0], bottom+22, 10, figMuted, "middle", "400", frNum(p, 1)))
	}
	b.WriteString(sTxt(328, bottom+42, 11, figMuted, "middle", "400", "portefeuille (M€ réels)"))
	for _, w := range []float64{3.0, 3.5, 4.0} {
		q := m(pLo, w)
		b.WriteString(line(left-4, q[1], left, q[1], figRule, 1))
		b.WriteString(mTxt(left-8, q[1]+3.5, 10, figMuted, "end", "400", frNum(w, 1)))
	}
	b.WriteString(sTxt(left-8, bottom+22, 10, figMuted, "end", "400", "taux servi (%)"))

	type node struct {
		p, cape float64
		label   string
		lx, ly  float64 // where its three lines of annotation go
		anchor  string
	}
	base := node{1.50, 30, "aujourd'hui", 0, -16, "middle"}
	krach := node{1.05, 21, "krach", -11, -30, "end"}
	boom := node{2.10, 38, "euphorie", 0, 20, "middle"}

	bw := capeRate(base.cape)
	bp := m(base.p, bw)

	// The fixed percentage keeps w and crosses every iso-income curve.
	for _, t := range []node{krach, boom} {
		q := m(t.p, bw)
		b.WriteString(arrow(bp[0], bp[1], q[0]+math.Copysign(8, bp[0]-q[0]), q[1], figMuted, 1.6, "4 4"))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="none" stroke="%s" stroke-width="1.6"/>`,
			q[0], q[1], figMuted)
		b.WriteString(mTxt(q[0], q[1]-10, 10.5, figMuted, "middle", "600",
			fmt.Sprintf("%s k€", frNum(inc(t.p, bw), 1))))
		b.WriteString(sTxt(q[0], q[1]-24, 10, figMuted, "middle", "400", "pourcentage fixe"))
	}

	// The CAPE rule moves across the plane, almost along one curve.
	for _, t := range []node{krach, boom} {
		w := capeRate(t.cape)
		q := m(t.p, w)
		dx := math.Copysign(8, bp[0]-q[0])
		b.WriteString(arrow(bp[0], bp[1], q[0]+dx, q[1]-math.Copysign(4, bp[1]-q[1]), figAccent, 2.2, ""))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, q[0], q[1], figDeep)
		delta := inc(t.p, w)/inc(base.p, bw) - 1
		b.WriteString(mTxt(q[0]+t.lx, q[1]+t.ly, 11.5, figDeep, t.anchor, "600",
			fmt.Sprintf("%s k€", frNum(inc(t.p, w), 1))))
		b.WriteString(mTxt(q[0]+t.lx, q[1]+t.ly+14, 10, figDeep, t.anchor, "400",
			strings.Replace(fmt.Sprintf("%+.0f %%", delta*100), "-", "\u2212", 1)))
		b.WriteString(sTxt(q[0]+t.lx, q[1]+t.ly+28, 10, figMuted, t.anchor, "400",
			fmt.Sprintf("%s, CAPE %.0f", t.label, t.cape)))
	}

	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.6" fill="%s"/>`, bp[0], bp[1], figInk)
	b.WriteString(mTxt(bp[0], bp[1]+base.ly, 11.5, figInk, "middle", "600",
		fmt.Sprintf("%s k€", frNum(inc(base.p, bw), 1))))
	b.WriteString(sTxt(bp[0], bp[1]+base.ly-13, 10, figMuted, "middle", "400", "aujourd'hui, CAPE 30"))

	b.WriteString(sTxt(24, 352, 10.5, figMuted, "start", "400",
		"Trait plein : la règle CAPE, qui longe presque une courbe de revenu constant."))
	b.WriteString(sTxt(24, 368, 10.5, figMuted, "start", "400",
		"Pointillé : le pourcentage fixe, qui les traverse toutes et prend le choc entier."))
	return svg(640, 384, b.String())
}

// --- Mortality credits: the age at which pooling beats the portfolio ---

// gompertzCredit is the yearly mortality credit of an annuitant pool, that is
// the force of mortality of the group, under a Gompertz law calibrated on
// annuitant lives: modal age 91, dispersion 10. It returns percent per year,
// and lands on the orders of magnitude the article quotes (0.7 % at 65, 3.3 %
// at 80, 5.5 % at 85).
func gompertzCredit(age float64) float64 { return math.Exp((age-91)/10) / 10 * 100 }

func figCreditsMortalite() string {
	const (
		aLo, aHi = 60.0, 92.0
		vMax     = 10.0
	)
	m := mapper(aLo, aHi, 0, vMax, 76, 600, 286, 76)
	y := func(v float64) float64 { return m(aLo, v)[1] }
	x := func(a float64) float64 { return m(a, 0)[0] }

	var b strings.Builder
	b.WriteString(plateHead("l'économie de la rente",
		"Le crédit de mortalité, et l'âge où il prend le dessus"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"ce que la mortalité du groupe verse aux survivants, en % du capital et par an"))

	for _, g := range []float64{0, 2, 4, 6, 8, 10} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(76, gy, 600, gy, col, 1))
		b.WriteString(mTxt(68, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}

	// The window where the credit curve leaves the band: mortality takes over.
	// creditAge inverts gompertzCredit, so the two bounds are read off the model
	// rather than asserted.
	creditAge := func(v float64) float64 { return 91 + 10*math.Log(v/10) }
	lo, hi := creditAge(2.5), creditAge(3.5)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x(lo), 76.0, x(hi)-x(lo), 286.0-76, figWash)
	b.WriteString(sTxt((x(lo)+x(hi))/2, 92, 10.5, figDeep, "middle", "600",
		fmt.Sprintf("%.0f à %.0f ans", lo, hi)))

	// What you give up by converting a growth portfolio into an insurer's
	// bond-backed promise: the equity risk premium, 2.5 to 3.5 points a year.
	fmt.Fprintf(&b, `<rect x="76" y="%.1f" width="524" height="%.1f" fill="%s"/>`,
		y(3.5), y(2.5)-y(3.5), figBlueWash)
	b.WriteString(sTxt(82, (y(2.5)+y(3.5))/2+3.5, 10.5, figBlue, "start", "600",
		"la prime de risque cédée : 2,5 à 3,5 %/an"))

	// The credit curve, drawn while it fits the frame.
	var pts [][2]float64
	for a := aLo; a <= aHi+0.01; a += 0.5 {
		v := gompertzCredit(a)
		if v > vMax {
			break
		}
		pts = append(pts, m(a, v))
	}
	b.WriteString(smoothStroke(pts, figAccent, 2.6))

	// Three readings the article quotes in words.
	for _, a := range []float64{65, 75, 85} {
		v := gompertzCredit(a)
		p := m(a, v)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.8" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`,
			p[0], p[1], figDeep)
		b.WriteString(mTxt(p[0]-8, p[1]-8, 10.5, figDeep, "end", "600", frNum(v, 1)+" %"))
	}
	last := pts[len(pts)-1]
	b.WriteString(mTxt(last[0]-8, last[1]+6, 10, figAccent, "end", "600", "11 % à 92 ans"))

	for _, a := range []float64{60, 65, 70, 75, 80, 85, 90} {
		b.WriteString(mTxt(x(a), 304, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", a)))
	}
	b.WriteString(sTxt(338, 324, 11, figMuted, "middle", "400", "âge à l'achat de la rente"))
	b.WriteString(sTxt(24, 348, 10.5, figMuted, "start", "400",
		"À 65 ans, convertir un portefeuille en rente cède trois points de prime pour en toucher moins d'un."))
	b.WriteString(sTxt(24, 364, 10.5, figMuted, "start", "400",
		"Après 80 ans, aucun portefeuille prudent ne réplique ce que verse la mortalité du groupe."))
	return svg(640, 380, b.String())
}

// --- The decision, as the narrowing procedure it is ---

func figArbreDecision() string {
	type step struct {
		n, q, out string
	}
	steps := []step{
		{"1", "Le socle : quel plancher, quel confort, quelle phase à découvert ?", "trois chiffres écrits"},
		{"2", "L'admissibilité : chaque famille passe-t-elle son test éliminatoire ?", "une ou deux familles éliminées"},
		{"3", "Le tempérament, la gouvernance : quelle forme de risque, et qui exécutera dans vingt ans ?", "deux finalistes"},
		{"4", "Les hybrides : une règle par phase, un plancher garanti ?", "l'assemblage retenu"},
		{"5", "Calibrer, tester, écrire : le plan tient-il dans le faisceau ?", "une page signée"},
	}

	var b strings.Builder
	b.WriteString(plateHead("la procédure", "Cinq étapes, dans cet ordre : le champ se referme"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"on entre avec neuf règles candidates, on sort avec une seule, écrite"))

	const (
		top  = 80.0
		hRow = 52.0
		gap  = 10.0
	)
	for i, s := range steps {
		ty := top + float64(i)*(hRow+gap)
		x0 := 24 + float64(i)*18
		x1 := 616 - float64(i)*18
		fill, ink, sub := figWash, figInk, figSoft
		if i == len(steps)-1 {
			fill, ink, sub = figDeepWash, figDeep, figDeep
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="6" fill="%s"/>`,
			x0, ty, x1-x0, hRow, fill)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="11" fill="%s"/>`, x0+22, ty+hRow/2, ink)
		b.WriteString(mTxt(x0+22, ty+hRow/2+4, 11.5, "#fffdf9", "middle", "600", s.n))

		// The question breaks after its label, which is also where it reads best.
		l1, l2 := splitOnColon(s.q)
		b.WriteString(sTxt(x0+42, ty+21, 10.5, ink, "start", "600", l1))
		if l2 != "" {
			b.WriteString(sTxt(x0+42, ty+35, 10.5, ink, "start", "400", l2))
		}
		b.WriteString(sTxt(x1-14, ty+hRow/2+4, 10.5, sub, "end", "400", "→ "+s.out))

		if i < len(steps)-1 {
			b.WriteString(arrow(320, ty+hRow+1, 320, ty+hRow+gap-1, figMuted, 1.4, ""))
		}
	}
	b.WriteString(sTxt(24, top+5*(hRow+gap)+16, 10.5, figMuted, "start", "400",
		"Les trois premières étapes se font sans simulateur. Les deux dernières tiennent en une séance."))
	return svg(640, int(top+5*(hRow+gap))+30, b.String())
}

// splitOnColon breaks "Label : question ?" after the colon, which is both the
// natural reading break and the shortest first line.
func splitOnColon(s string) (string, string) {
	if i := strings.Index(s, " : "); i > 0 {
		return s[:i+2], s[i+3:]
	}
	return s, ""
}

// wrapTwo splits a sentence onto at most two lines of about n characters, on
// word boundaries. It balances the two lines rather than filling the first
// greedily, which leaves a lonely word on the second line.
func wrapTwo(s string, n int) (string, string) {
	words := strings.Fields(s)
	if len([]rune(s)) <= n {
		return s, ""
	}
	best, bestCost := 1, math.MaxFloat64
	for k := 1; k < len(words); k++ {
		a := len([]rune(strings.Join(words[:k], " ")))
		z := len([]rune(strings.Join(words[k:], " ")))
		if a > n {
			break
		}
		cost := math.Abs(float64(a - z))
		if z > n {
			cost += float64(z-n) * 10
		}
		if cost < bestCost {
			best, bestCost = k, cost
		}
	}
	return strings.Join(words[:best], " "), strings.Join(words[best:], " ")
}

// --- Two thermometers, four instrument strips ---

func figDeuxThermometres() string {
	type strip struct {
		cx            float64
		unit          string
		lo, hi        float64 // scale
		bLo, bHi      float64 // corridor
		before, after float64
		dec           int
		ticks         []float64
		verdict       string
		cut           bool
	}
	type pair struct {
		x0, x1 float64
		who    string
		what   string
		a, b   strip
	}
	pairs := []pair{{
		x0: 24, x1: 312,
		who:  "Le retraité de 62 ans, pension à 64 ans",
		what: "un krach emporte 25 % du portefeuille",
		a: strip{cx: 104, unit: "taux de retrait courant (%)", lo: 3, hi: 7, bLo: 3.4, bHi: 5.2,
			before: 4.0, after: 5.3, dec: 1, ticks: []float64{3, 4, 5, 6, 7}, verdict: "coupe de 10 %", cut: true},
		b: strip{cx: 236, unit: "probabilité de succès (%)", lo: 60, hi: 100, bLo: 85, bHi: 99,
			before: 93, after: 90, ticks: []float64{60, 70, 80, 90, 100}, verdict: "on ne touche à rien"},
	}, {
		x0: 328, x1: 616,
		who:  "Le FIRE de 48 ans, marché très cher",
		what: "une décennie plate, sans krach",
		a: strip{cx: 408, unit: "taux de retrait courant (%)", lo: 3, hi: 7, bLo: 3.4, bHi: 5.2,
			before: 3.6, after: 3.8, dec: 1, ticks: []float64{3, 4, 5, 6, 7}, verdict: "on ne touche à rien"},
		b: strip{cx: 540, unit: "probabilité de succès (%)", lo: 60, hi: 100, bLo: 85, bHi: 99,
			before: 88, after: 74, ticks: []float64{60, 70, 80, 90, 100}, verdict: "coupe de 10 %", cut: true},
	}}

	const (
		sTop = 138.0
		sBot = 300.0
	)
	var b strings.Builder
	b.WriteString(plateHead("kitces-tharp", "Deux thermomètres, deux verdicts opposés"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"chaque ménage lu par le ratio brut de 2006, puis par la probabilité de succès du plan complet ;"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"la bande beige est le corridor de la règle, le point vide l'avant, le point plein l'après"))
	b.WriteString(line(320, 92, 320, 356, figRule, 1))

	for _, p := range pairs {
		b.WriteString(sTxt(p.x0, 108, 11, figInk, "start", "600", p.who))
		b.WriteString(sTxt(p.x0, 122, 10.5, figMuted, "start", "400", p.what))
		for _, s := range []strip{p.a, p.b} {
			y := func(v float64) float64 { return sBot - (v-s.lo)/(s.hi-s.lo)*(sBot-sTop) }
			// the corridor the rule watches
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="52" height="%.1f" rx="3" fill="%s"/>`,
				s.cx-26, y(s.bHi), y(s.bLo)-y(s.bHi), figWash)
			b.WriteString(line(s.cx-26, y(s.bHi), s.cx+26, y(s.bHi), figRule, 1))
			b.WriteString(line(s.cx-26, y(s.bLo), s.cx+26, y(s.bLo), figRule, 1))
			b.WriteString(line(s.cx, sTop, s.cx, sBot, figRule, 1))
			for _, t := range s.ticks {
				b.WriteString(mTxt(s.cx-32, y(t)+3.5, 9.5, figMuted, "end", "400", frNum(t, 0)))
			}
			// before, after, and the move between them
			bp, ap := y(s.before), y(s.after)
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="none" stroke="%s" stroke-width="1.8"/>`,
				s.cx, bp, figMuted)
			b.WriteString(mTxt(s.cx+32, bp+3.5, 9.5, figMuted, "start", "400", frNum(s.before, s.dec)))
			col := figGood
			if s.cut {
				col = figBad
			}
			b.WriteString(arrow(s.cx, bp+math.Copysign(7, ap-bp), s.cx, ap-math.Copysign(6, ap-bp), col, 1.8, ""))
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.6" fill="%s"/>`, s.cx, ap, col)
			b.WriteString(mTxt(s.cx+32, ap+3.5, 10, col, "start", "600", frNum(s.after, s.dec)))
			// the instrument, and what it decides
			l1, l2 := wrapTwo(s.unit, 18)
			b.WriteString(sTxt(s.cx, 322, 10, figSoft, "middle", "600", l1))
			b.WriteString(sTxt(s.cx, 335, 10, figSoft, "middle", "400", l2))
			b.WriteString(sTxt(s.cx, 353, 10.5, col, "middle", "600", s.verdict))
		}
	}
	b.WriteString(sTxt(24, 378, 10.5, figMuted, "start", "400",
		"Le ratio brut ignore les deux ans de pont du premier ménage, et le prix payé pour les actions par le second."))
	b.WriteString(sTxt(24, 394, 10.5, figMuted, "start", "400",
		"Le capteur par risque voit les deux. Valeurs illustratives."))
	return svg(640, 408, b.String())
}
