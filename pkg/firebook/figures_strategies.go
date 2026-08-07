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

// --- The same shock, three transmissions: raw, Yale, corridor ---

// pourcentageReturns are the reference 60/40's real calendar-year returns from
// 1973, the regime the article's worked example names. The three income paths
// are derived from them in the plate rather than frozen, so the smoothing rules
// stay visible as code; figures_strategies_test.go recomputes the returns from
// pkg/replay.
var pourcentageReturns = []float64{
	-0.1518, -0.2278, 0.1687, 0.1419, -0.1014, -0.0429,
	-0.0105, 0.0697, -0.0557, 0.1976, 0.1119, 0.0584,
}

// smoothingPaths runs a 4 % rule through those returns three ways: raw
// percentage, Yale (70 % memory + 30 % of the proportional target), and
// Vanguard's corridor (+5 %/-2.5 % a year). Wealth is drawn at the start of the
// year, then the year's return applies. Everything is real, in k EUR.
func smoothingPaths() (raw, yale, corridor []float64, endWealth [3]float64) {
	const w, p0 = 0.04, 1400.0
	wealth := [3]float64{p0, p0, p0}
	var last [3]float64
	for i, r := range pourcentageReturns {
		cur := [3]float64{w * wealth[0], w * wealth[1], w * wealth[2]}
		if i > 0 {
			cur[1] = 0.7*last[1] + 0.3*w*wealth[1]
			cur[2] = math.Min(math.Max(cur[2], last[2]*0.975), last[2]*1.05)
		}
		raw, yale, corridor = append(raw, cur[0]), append(yale, cur[1]), append(corridor, cur[2])
		for k := range wealth {
			wealth[k] = (wealth[k] - cur[k]) * (1 + r)
		}
		last = cur
	}
	return raw, yale, corridor, wealth
}

func figPourcentageLissages() string {
	raw, yale, corridor, end := smoothingPaths()
	const (
		x0, x1     = 76.0, 520.0
		yTop, yBot = 96.0, 268.0
		vLo, vHi   = 26.0, 58.0
		lean       = 45.0 // 80 % of the plan, the article's admissibility bar
	)
	n := len(raw) - 1
	x := func(i int) float64 { return x0 + float64(i)/float64(n)*(x1-x0) }
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("le lissage du pourcentage",
		"Le même krach, trois façons de le transmettre au ménage"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"1,4 M€, w = 4 %, à travers les rendements réels du 60/40 de 1973 à 1984 ; revenu servi, k€"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"pointillé : les 45 k€ sous lesquels le ménage entame son incompressible"))

	for _, g := range []float64{30, 40, 50} {
		b.WriteString(line(x0, y(g), x1, y(g), figGrid, 1))
		b.WriteString(mTxt(x0-8, y(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(dashLine(x0, y(lean), x1, y(lean), figMuted, 1, "3 3"))

	type series struct {
		v     []float64
		color string
		label string
		short string
		dy    float64 // where the end label sits, the three ends being close
	}
	all := []series{
		{corridor, figBlue, "corridor borné", "corridor", -12},
		{yale, figAccent, "règle de Yale", "Yale", 14},
		{raw, figBad, "pourcentage brut", "brut", -16},
	}
	for _, s := range all {
		pts := make([][2]float64, len(s.v))
		for i, v := range s.v {
			pts[i] = [2]float64{x(i), y(v)}
		}
		b.WriteString(poly(pts, s.color, 2.2, ""))
		e := pts[len(pts)-1]
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, e[0], e[1], s.color)
		b.WriteString(sTxt(e[0]+9, e[1]+s.dy, 10.5, s.color, "start", "600", s.label))
		b.WriteString(mTxt(e[0]+9, e[1]+s.dy+13, 10, s.color, "start", "400", frNum(s.v[len(s.v)-1], 1)))
	}
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, x(0), y(56), figInk)
	b.WriteString(mTxt(x(0)+8, y(56)-7, 10.5, figInk, "start", "600", "56,0 au départ"))

	for i := range raw {
		if i%2 == 0 {
			b.WriteString(mTxt(x(i), 286, 10, figMuted, "middle", "400", fmt.Sprintf("%d", 1973+i)))
		}
	}

	// One square per year, filled when the household lives under the bar: the
	// count is the answer to "how long would this last?", which a curve alone
	// makes the reader do by hand.
	b.WriteString(sTxt(24, 314, 10.5, figMuted, "start", "400",
		"années passées sous 45 k€, une case par an"))
	for k, s := range all {
		ty := 330 + float64(k)*22
		b.WriteString(sTxt(x0-8, ty+9, 10, s.color, "end", "600", s.short))
		count := 0
		for i, v := range s.v {
			fill := "#fffdf9"
			if v < lean {
				fill, count = s.color, count+1
			}
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="11" height="11" rx="2" fill="%s" stroke="%s" stroke-width="1"/>`,
				x0+float64(i)*15, ty, fill, s.color)
		}
		b.WriteString(mTxt(x0+12*15+8, ty+9, 10.5, s.color, "start", "600", fmt.Sprintf("%d / 12", count)))
	}
	b.WriteString(sTxt(24, 404, 10.5, figMuted, "start", "400", fmt.Sprintf(
		"Le confort du corridor se paie sur le capital : fin 1984 il reste %.0f k€ contre %.0f au brut. Il a emprunté à sa propre suite.",
		end[2], end[0])))
	return svg(640, 418, b.String())
}

// --- Choosing w: the geometric bound, drawn ---

func figBorneGeometrique() string {
	const (
		g          = 0.04 // real geometric return of a diversified portfolio
		p0         = 1000.0
		years      = 30
		x0, x1     = 84.0, 556.0
		yTop, yBot = 92.0, 268.0
		vLo, vHi   = 18.0, 64.0
	)
	x := func(t float64) float64 { return x0 + t/years*(x1-x0) }
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }
	// income after t years of withdrawing w of the portfolio each year
	income := func(w, t float64) float64 { return w * p0 * math.Pow((1+g)*(1-w), t) }

	var b strings.Builder
	b.WriteString(plateHead("choisir le pourcentage",
		"La borne géométrique : au-delà, le revenu s'érode pour toujours"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"1 M€, rendement réel géométrique de 4 %/an, revenu servi en k€, sans aléa"))

	for _, v := range []float64{20, 30, 40, 50, 60} {
		b.WriteString(line(x0, y(v), x1, y(v), figGrid, 1))
		b.WriteString(mTxt(x0-8, y(v)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", v)))
	}
	for _, t := range []float64{0, 10, 20, 30} {
		b.WriteString(mTxt(x(t), 286, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt(320, 306, 11, figMuted, "middle", "400", "années de retraite"))

	// The boundary itself: w* leaves the income flat for ever.
	wStar := g / (1 + g)
	for _, w := range []float64{0.03, 0.04, 0.05, 0.06} {
		var pts [][2]float64
		for t := 0.0; t <= years+0.01; t += 1 {
			pts = append(pts, [2]float64{x(t), y(income(w, t))})
		}
		col := figBad
		if w <= wStar {
			col = figGood
		}
		b.WriteString(smoothStroke(pts, col, 2.2))
		st := pts[0]
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, st[0], st[1], col)
		b.WriteString(sTxt(st[0]+9, st[1]-6, 10.5, col, "start", "600",
			fmt.Sprintf("w = %.0f %%", w*100)))
		if w == 0.03 || w == 0.06 {
			e := pts[len(pts)-1]
			b.WriteString(mTxt(e[0]+8, e[1]+3.5, 10.5, col, "start", "600", frNum(income(w, years), 1)))
		}
	}

	// The crossing: the generous rate starts twice as high and ends lower.
	tc := math.Log(2) / math.Log(((1+g)*(1-0.03))/((1+g)*(1-0.06)))
	cx, cy := x(tc), y(income(0.03, tc))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="none" stroke="%s" stroke-width="1.8"/>`, cx, cy, figInk)
	b.WriteString(mTxt(cx+9, cy+16, 10, figInk, "start", "600", fmt.Sprintf("%.0f ans", tc)))

	b.WriteString(sTxt(24, 338, 10.5, figSoft, "start", "600", fmt.Sprintf(
		"La bascule est à w = g / (1 + g), soit %s %% ici.", frNum(wStar*100, 1))))
	b.WriteString(sTxt(24, 354, 10.5, figMuted, "start", "400",
		"En dessous, le portefeuille croît plus vite qu'on ne le ponctionne et le revenu monte doucement."))
	b.WriteString(sTxt(24, 370, 10.5, figMuted, "start", "400",
		"Au-dessus, chaque année prélève plus que la croissance : le 6 % part au double du 3 % et finit en dessous."))
	return svg(640, 384, b.String())
}

// --- What the rule the household would have been told, every January since 1881 ---

// capeJanuaries is the Shiller CAPE of every 1st of January from 1881 to 2026,
// from the bundled pkg/datasets/cape snapshot; figures_strategies_test.go reads
// the dataset back and fails on any drift. The plate turns each into the rate
// ERN's rule would have quoted that year, so the formula stays visible.
var capeJanuaries = []float64{
	18.47, 15.68, 15.27, 14.43, 13.13, 16.69, 17.51, 15.36, 15.80, 17.22, 15.43, 19.02,
	17.66, 15.74, 16.52, 16.58, 17.03, 19.25, 22.93, 18.67, 20.98, 22.34, 20.32, 15.86,
	18.46, 20.13, 17.22, 11.90, 14.76, 14.55, 14.05, 13.79, 13.15, 11.64, 10.36, 12.54,
	10.99, 6.64, 6.10, 5.99, 5.12, 6.29, 8.15, 8.07, 9.69, 11.34, 13.19, 18.81,
	27.08, 22.31, 16.71, 9.31, 8.73, 13.03, 11.50, 17.09, 21.62, 13.51, 15.60, 16.38,
	13.90, 10.10, 10.15, 11.05, 11.96, 15.62, 11.47, 10.42, 10.25, 10.75, 11.90, 12.53,
	13.01, 12.00, 15.99, 18.29, 16.72, 13.79, 17.98, 18.34, 18.47, 21.20, 19.26, 21.63,
	23.27, 24.06, 20.43, 21.51, 21.19, 17.09, 16.46, 17.26, 18.71, 13.53, 8.92, 11.19,
	11.44, 9.24, 9.26, 8.85, 9.26, 7.39, 8.76, 9.89, 10.00, 11.72, 14.92, 13.90,
	15.09, 17.05, 15.61, 19.77, 20.32, 21.41, 20.22, 24.76, 28.33, 32.86, 40.58, 43.77,
	36.98, 30.28, 22.90, 27.66, 26.59, 26.47, 27.21, 24.02, 15.17, 20.53, 22.98, 21.21,
	21.90, 24.86, 26.49, 24.21, 28.06, 33.31, 28.38, 30.99, 34.51, 36.94, 28.33, 31.97,
	37.14, 40.03,
}

const capeFirstYear = 1881

func figCapeDepuis1881() string {
	const (
		x0, x1     = 76.0, 596.0
		yTop, yBot = 88.0, 270.0
		vLo, vHi   = 2.0, 12.0
	)
	n := len(capeJanuaries) - 1
	x := func(year int) float64 { return x0 + float64(year-capeFirstYear)/float64(n)*(x1-x0) }
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("le taux n'est pas une constante",
		"Ce que la règle CAPE aurait servi, chaque janvier depuis 1881"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"taux de retrait a + b / CAPE (a = 1,75, b = 0,5) appliqué au CAPE du 1er janvier, en % par an"))

	for _, g := range []float64{2, 4, 6, 8, 10, 12} {
		gy := y(g)
		col := figGrid
		if g == 4 {
			col = figRule
		}
		b.WriteString(line(x0, gy, x1, gy, col, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(x0+8, y(4)+15, 10, figMuted, "start", "400", "les 4 % de Bengen"))

	pts := make([][2]float64, len(capeJanuaries))
	for i, c := range capeJanuaries {
		v := capeRate(c)
		if v > vHi {
			v = vHi
		}
		pts[i] = [2]float64{x(capeFirstYear + i), y(v)}
	}
	b.WriteString(poly(pts, figAccent, 1.8, ""))

	// The years the book keeps naming, each with what the rule would have said.
	type mark struct {
		year   int
		dy     float64
		anchor string
		dx     float64
		note   string
	}
	for _, m := range []mark{
		{1921, 16, "end", -8, "l'après-guerre : CAPE 5"},
		{1929, 16, "middle", 0, "la bulle de 1929"},
		{1966, 16, "middle", 0, "le pire millésime"},
		{1982, -16, "start", 8, "la fin de la stagflation"},
		{2000, 16, "end", -8, "la bulle internet"},
		{2026, -14, "end", -6, ""},
	} {
		v := capeRate(capeJanuaries[m.year-capeFirstYear])
		p := [2]float64{x(m.year), y(math.Min(v, vHi))}
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.8" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`,
			p[0], p[1], figDeep)
		b.WriteString(mTxt(p[0]+m.dx, p[1]+m.dy, 10.5, figDeep, m.anchor, "600", frNum(v, 1)+" %"))
		if m.note != "" {
			b.WriteString(sTxt(p[0]+m.dx, p[1]+m.dy+13, 9.5, figMuted, m.anchor, "400", m.note))
		}
	}

	for _, yr := range []int{1881, 1900, 1920, 1940, 1960, 1980, 2000, 2026} {
		b.WriteString(mTxt(x(yr), 290, 10, figMuted, "middle", "400", fmt.Sprintf("%d", yr)))
	}
	b.WriteString(sTxt(24, 320, 10.5, figMuted, "start", "400",
		"La règle n'a presque jamais dit 4 %. Elle a dit 11,5 % en 1921 et 2,9 % en 2000, parce que le taux soutenable"))
	b.WriteString(sTxt(24, 336, 10.5, figMuted, "start", "400",
		"est une fonction du prix payé, pas une constante de la nature. Aux extrêmes elle déraille, faute de borne :"))
	b.WriteString(sTxt(24, 352, 10.5, figMuted, "start", "400",
		"personne n'aurait dû retirer 11 % en 1921, et c'est le défaut d'une formule linéaire en 1 / CAPE."))
	return svg(640, 366, b.String())
}

// --- Total wealth: what the amortisation rule looks at, and what the others miss ---

func figRichesseTotale() string {
	const (
		portfolio = 1550.0 // k EUR, the article's worked household
		pension   = 310.0  // discounted value of the future pension
		bequest   = 90.0   // discounted value of the bequest aimed at
		rate      = 0.0405 // the annuity rate over the remaining horizon
		x0, x1    = 120.0, 480.0
		vMax      = 1900.0
		barH      = 40.0
	)
	x := func(v float64) float64 { return x0 + v/vMax*(x1-x0) }
	seg := func(lo, hi, top float64, fill string) string {
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			x(lo), top, x(hi)-x(lo), barH, fill)
	}

	var b strings.Builder
	b.WriteString(plateHead("la richesse totale",
		"Ce que l'amortissement regarde, et ce que les autres règles ignorent"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"un ménage de 49 ans : 1,55 M€ de portefeuille, 19 k€/an de pension à 67 ans, 200 k€ de legs visé"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"barres à l'échelle en k€, segment bleu = la pension future actualisée ; à droite, la rente servie"))

	// Row one: the visible portfolio, which is all most rules ever see.
	const t1 = 126.0
	b.WriteString(sTxt(24, t1-12, 11, figInk, "start", "600", "Ce que la plupart des règles amortissent"))
	b.WriteString(seg(0, portfolio, t1, figWash))
	b.WriteString(mTxt(x(portfolio/2), t1+barH/2+4, 11.5, figSoft, "middle", "600", "1 550"))
	b.WriteString(mTxt(x(portfolio)+14, t1+barH/2+4, 12.5, figMuted, "start", "600", "62,8 k€"))

	// Row two: total wealth, the pension added and the bequest carved out of it.
	const t2 = 226.0
	b.WriteString(sTxt(24, t2-12, 11, figInk, "start", "600", "Ce que l'amortissement amortit vraiment"))
	b.WriteString(seg(0, portfolio, t2, figWash))
	b.WriteString(seg(portfolio, portfolio+pension, t2, figBlueWash))
	b.WriteString(line(x(portfolio), t2, x(portfolio), t2+barH, figRule, 1))
	b.WriteString(mTxt(x(portfolio/2), t2+barH/2+4, 11.5, figSoft, "middle", "600", "1 550"))
	b.WriteString(mTxt(x(portfolio+(pension-bequest)/2), t2+barH/2+4, 10.5, figBlue, "middle", "600", "+310"))
	b.WriteString(mTxt(x(portfolio+pension)+14, t2+barH/2+4, 12.5, figDeep, "start", "600", "71,7 k€"))

	// The bequest is the slice of that wealth already spoken for.
	bq := x(portfolio + pension - bequest)
	b.WriteString(dashLine(bq, t2-6, bq, t2+barH+22, figDeep, 1.4, "4 3"))
	b.WriteString(line(bq, t2+barH+16, x(portfolio+pension), t2+barH+16, figDeep, 1.2))
	b.WriteString(sTxt(x(portfolio+pension)+8, t2+barH+20, 10, figDeep, "start", "400", "−90, le legs visé"))
	b.WriteString(mTxt(bq, t2-14, 10.5, figDeep, "middle", "600", "1 770 amortis"))

	// The two legends, aligned under the numbers they explain.
	b.WriteString(sTxt(x(portfolio)+14, t1+barH/2+20, 10, figMuted, "start", "400", "par an"))
	b.WriteString(sTxt(x(portfolio+pension)+14, t2+barH/2+20, 10, figDeep, "start", "400", "par an, soit +14 %"))

	b.WriteString(sTxt(24, 320, 10.5, figMuted, "start", "400",
		"La pension n'arrive que dans dix-huit ans, mais elle est certaine : la consommer par anticipation est"))
	b.WriteString(sTxt(24, 336, 10.5, figMuted, "start", "400",
		"le lissage que les économistes prescrivent depuis Modigliani. Et le legs cesse d'être un résidu."))
	return svg(640, 350, b.String())
}

// --- The two floors: what has to be covered, and who covers it ---

func figEtagesDuPlancher() string {
	const (
		yBase = 292.0
		yTop  = 108.0
		total = 50.0 // k EUR of yearly need
	)
	y := func(v float64) float64 { return yBase - v/total*(yBase-yTop) }
	bar := func(cx, lo, hi float64, fill string) string {
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="120" height="%.1f" fill="%s"/>`,
			cx-60, y(hi), y(lo)-y(hi), fill)
	}

	var b strings.Builder
	b.WriteString(plateHead("safety first",
		"Deux colonnes qui doivent s'aligner : le besoin et sa source"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"un couple de 74 et 72 ans, 1,1 M€, 38 k€ de plancher et 50 k€ de confort, k€ par an"))

	const lx, rx = 190.0, 450.0
	// left: the need, split at the floor
	b.WriteString(bar(lx, 0, 38, figWash))
	b.WriteString(bar(lx, 38, 50, figBlueWash))
	b.WriteString(sTxt(lx, y(19)+4, 11, figSoft, "middle", "600", "le plancher, 38"))
	b.WriteString(sTxt(lx, y(44)+4, 10.5, figBlue, "middle", "600", "le confort, 12"))
	b.WriteString(sTxt(lx, yTop-22, 11, figInk, "middle", "600", "Ce qu'il faut chaque année"))

	// right: who pays for it, in the order safety-first prescribes
	type layer struct {
		lo, hi float64
		fill   string
		color  string
		label  string
	}
	for _, l := range []layer{
		{0, 26, figWash, figSoft, "pensions, 26"},
		{26, 38, figDeepWash, figDeep, "rente viagère, 12"},
		{38, 50, figBlueWash, figBlue, "portefeuille, 12"},
	} {
		b.WriteString(bar(rx, l.lo, l.hi, l.fill))
		b.WriteString(sTxt(rx, y((l.lo+l.hi)/2)+4, 10.5, l.color, "middle", "600", l.label))
		b.WriteString(line(rx-60, y(l.hi), rx+60, y(l.hi), figRule, 1))
	}
	b.WriteString(sTxt(rx, yTop-22, 11, figInk, "middle", "600", "Qui le paie, et avec quel risque"))

	// the floor line, drawn across both columns: the whole doctrine in one rule
	b.WriteString(dashLine(lx-60, y(38), rx+60, y(38), figDeep, 1.6, "5 4"))
	b.WriteString(sTxt(320, y(38)-8, 10.5, figDeep, "middle", "600", "le plancher,"))
	b.WriteString(sTxt(320, y(38)+16, 10.5, figDeep, "middle", "600", "couvert à vie"))
	b.WriteString(line(80, yBase, 600, yBase, figRule, 1))

	// what is left exposed, which is the real product of the operation
	b.WriteString(sTxt(24, yBase+28, 10.5, figMuted, "start", "400",
		"Le portefeuille de 874 k€ restant ne finance plus que 12 k€ par an, soit un taux de retrait de 1,4 %."))
	b.WriteString(sTxt(24, yBase+44, 10.5, figMuted, "start", "400",
		"La variabilité n'a plus de prise sur le plancher : elle ne peut plus toucher qu'aux voyages."))
	return svg(640, int(yBase)+58, b.String())
}

// --- Where the attention belongs: ranked by what actually moves the ruin ---

func figHierarchieAttention() string {
	const (
		x0, x1 = 250.0, 560.0
		vMax   = 11.0
	)
	x := func(v float64) float64 { return x0 + v/vMax*(x1-x0) }

	var b strings.Builder
	b.WriteString(plateHead("où porter son attention",
		"Ce qui déplace vraiment la ruine, du plus lourd au plus léger"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"effet sur la probabilité de ruine, en points (ordres de grandeur de la littérature)"))

	for _, g := range []float64{0, 2, 4, 6, 8, 10} {
		b.WriteString(line(x(g), 94, x(g), 300, figGrid, 1))
		b.WriteString(mTxt(x(g), 88, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}

	type row struct {
		label, sub string
		lo, hi     float64
		color      string
	}
	rows := []row{
		{"Le taux initial", "4,5 % au lieu de 3,7 % en marché cher", 5, 10, figBad},
		{"La pension oubliée", "un flux futur non compté", 3, 8, figBad},
		{"Les dépenses sous-estimées", "10 % de budget en moins que la réalité", 3, 6, figAccent},
		{"Le choix de la règle", "entre deux bonnes règles bien calibrées", 1, 3, figGood},
	}
	for i, r := range rows {
		y := 118.0 + float64(i)*42
		b.WriteString(sTxt(x0-14, y+2, 11, figInk, "end", "600", r.label))
		b.WriteString(sTxt(x0-14, y+16, 10, figMuted, "end", "400", r.sub))
		b.WriteString(barH(x(r.lo), x(r.hi), y-8, 18, r.color))
		b.WriteString(line(x(r.lo), y-8, x(r.lo), y+10, r.color, 2))
		b.WriteString(mTxt(x(r.hi)+8, y+5, 10.5, r.color, "start", "600",
			fmt.Sprintf("%.0f à %.0f", r.lo, r.hi)))
	}

	// the one that has no scale, and dominates all of them
	y := 118.0 + 4*42
	b.WriteString(sTxt(x0-14, y+2, 11, figInk, "end", "600", "L'abandon de la règle en panique"))
	b.WriteString(sTxt(x0-14, y+16, 10, figMuted, "end", "400", "vendre au creux, refaire le plan à chaud"))
	b.WriteString(arrow(x(0), y+1, x(10.4), y+1, figDeep, 2.4, "6 4"))
	b.WriteString(sTxt(x(5), y+18, 10.5, figDeep, "middle", "600", "incalculable"))

	b.WriteString(sTxt(24, 330, 10.5, figMuted, "start", "400",
		"L'ordre de la liste est l'ordre du travail. Le choix fin entre deux finalistes arrive en dernier, sereinement,"))
	b.WriteString(sTxt(24, 346, 10.5, figMuted, "start", "400",
		"parce qu'à ce stade on ne peut plus beaucoup se tromper."))
	return svg(640, 360, b.String())
}

// --- The loss-tolerance test the VPW doctrine mandates, and everyone skips ---

func figVpwTestDePerte() string {
	const (
		yBase    = 288.0
		yTop     = 100.0
		vMax     = 78.0
		confort  = 52.0
		plancher = 38.0
	)
	y := func(v float64) float64 { return yBase - v/vMax*(yBase-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("le test de tolérance à la perte",
		"Le même choc, avec et sans le pont de pension"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"un couple de 47 ans, 1,6 M€ en 60/40, pensions de 21,6 k€ à 67 ans ; revenu servi, k€ par an"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"pointillé bleu : le confort visé, 52 ; pointillé brun : le plancher, 38"))

	for _, g := range []float64{0, 20, 40, 60} {
		b.WriteString(line(72, y(g), 600, y(g), figGrid, 1))
		b.WriteString(mTxt(64, y(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	for _, r := range []struct {
		v     float64
		lbl   string
		color string
	}{{confort, "le confort visé, 52", figBlue}, {plancher, "le plancher, 38", figDeep}} {
		b.WriteString(dashLine(72, y(r.v), 600, y(r.v), r.color, 1.4, "5 4"))
	}

	type col struct {
		cx          float64
		vpw, bridge float64
		label       string
	}
	cols := []col{
		{140, 50.0, 21.6, "normal"},
		{260, 35.0, 21.6, "actions −50 %"},
		{420, 64.3, 0, "normal"},
		{540, 45.0, 0, "actions −50 %"},
	}
	for _, c := range cols {
		total := c.vpw + c.bridge
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="76" height="%.1f" fill="%s"/>`,
			c.cx-38, y(c.vpw), y(0)-y(c.vpw), figWash)
		if c.bridge > 0 {
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="76" height="%.1f" fill="%s"/>`,
				c.cx-38, y(total), y(c.vpw)-y(total), figDeepWash)
			b.WriteString(line(c.cx-38, y(c.vpw), c.cx+38, y(c.vpw), figRule, 1))
			b.WriteString(mTxt(c.cx, y((c.vpw+total)/2)+4, 10, figDeep, "middle", "600", "pont"))
		}
		b.WriteString(mTxt(c.cx, y(c.vpw/2)+4, 10.5, figSoft, "middle", "600", "VPW"))
		col := figGood
		if total < confort {
			col = figBad
		}
		lbl := frNum(total, 1)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="16" fill="#fffdf9"/>`,
			c.cx-float64(len(lbl))*4, y(total)-23, float64(len(lbl))*8)
		b.WriteString(mTxt(c.cx, y(total)-10, 12, col, "middle", "600", lbl))
		b.WriteString(sTxt(c.cx, yBase+18, 10.5, figMuted, "middle", "400", c.label))
	}
	b.WriteString(line(72, yBase, 600, yBase, figRule, 1))
	b.WriteString(sTxt(200, yBase+38, 11, figInk, "middle", "600", "Avec le pont de pension"))
	b.WriteString(sTxt(200, yBase+52, 10.5, figGood, "middle", "600", "le choc laisse le ménage au-dessus du confort"))
	b.WriteString(sTxt(480, yBase+38, 11, figInk, "middle", "600", "Sans le pont"))
	b.WriteString(sTxt(480, yBase+52, 10.5, figBad, "middle", "600", "le même choc passe sous le confort"))
	b.WriteString(line(340, 96, 340, yBase+56, figRule, 1))
	return svg(640, int(yBase)+66, b.String())
}
