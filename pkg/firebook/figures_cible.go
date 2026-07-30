package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The two plates of combien-il-vous-faut. Both are pure arithmetic on the
// article's own numbers: the first draws the definition of the multiple, the
// second walks the worked case of étape 5 from the bank statement to the
// target. figures_cible_test.go recomputes every value from the closed
// formulas and fails the moment one of them leaves the prose.

// cibleMultiple is the whole content of the first plate: the target capital,
// counted in multiples of yearly spending, is the plain inverse of the
// withdrawal rate expressed in percent. No market assumption enters, which is
// why the article's "> 33x, rarement rationnel" line is a property of the
// curve rather than an opinion.
func cibleMultiple(rate float64) float64 { return 100 / rate }

// figCibleConvexite draws that inverse from a generous 5 % to a very prudent
// 2,4 %, with the three multiples the étape-4 table names posed on it and the
// cost of each half point bracketed on the left. The axis runs from generous
// to prudent, left to right, because that is the direction the reader travels
// when hesitating between 4 % and 3 %.
func figCibleConvexite() string {
	const (
		rMax, rMin = 5.0, 2.4   // the rate axis, drawn generous to prudent
		mLo, mHi   = 19.0, 44.0 // the multiple axis, a truncated ratio scale
		plotL      = 96.0
		plotR      = 600.0
		plotB      = 300.0
		plotT      = 96.0
	)
	m := mapper(rMax, rMin, mLo, mHi, plotL, plotR, plotB, plotT)
	x := func(r float64) float64 { return m(r, mLo)[0] }
	y := func(v float64) float64 { return m(rMax, v)[1] }

	var b strings.Builder
	b.WriteString(plateHead("combien il vous faut",
		"Chaque demi-point de prudence coûte plus cher que le précédent"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"le multiple est l'inverse exact du taux : les crochets donnent le prix de chaque demi-point de prudence"))

	// The zone the article calls rarely rational, drawn before the grid so the
	// hairlines stay on top of it.
	ceil := cibleMultiple(3)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		plotL, y(mHi), plotR-plotL, y(ceil)-y(mHi), figWash)

	for _, g := range []float64{20, 25, 30, 35, 40} {
		b.WriteString(line(plotL, y(g), plotR, y(g), figGrid, 1))
		b.WriteString(mTxt(plotL-10, y(g)+3.5, 10, figMuted, "end", "400", frNum(g, 0)))
	}
	b.WriteString(line(plotL, plotB, plotR, plotB, figRule, 1))
	b.WriteString(sTxt(plotL, 84, 10.5, figMuted, "start", "400",
		"capital cible, en multiples de dépenses annuelles"))

	// The four levels the brackets need. Only 28,6 and 33 need a guide of their
	// own: 25 and 40 already sit on a gridline.
	marks := []struct {
		rate  float64
		label string
		guide bool
	}{
		{4.0, "25x", false},
		{3.5, "28,6x", true},
		{3.0, "33x", true},
		{2.5, "40x", false},
	}
	for _, mk := range marks {
		px, py := x(mk.rate), y(cibleMultiple(mk.rate))
		if mk.guide {
			b.WriteString(dashLine(plotL, py, px, py, figMuted, 1, "3 3"))
		}
		b.WriteString(dashLine(px, py, px, plotB, figMuted, 1, "2 3"))
	}

	// What each half point of prudence buys, bracketed on the left where the
	// curve is nowhere near: the gaps grow, and that growth is the whole page.
	bracket := func(r0, r1 float64) {
		const bx = 124.0
		y0, y1 := y(cibleMultiple(r0)), y(cibleMultiple(r1))
		b.WriteString(line(bx, y0, bx, y1, figDeep, 1.2))
		b.WriteString(line(bx-4, y0, bx+4, y0, figDeep, 1.2))
		b.WriteString(line(bx-4, y1, bx+4, y1, figDeep, 1.2))
		d := cibleMultiple(r1) - cibleMultiple(r0)
		b.WriteString(mTxt(bx+8, (y0+y1)/2+3.5, 10.5, figDeep, "start", "600",
			"+ "+frNum(d, 1)+"x"))
	}
	bracket(4.0, 3.5)
	bracket(3.5, 3.0)
	bracket(3.0, 2.5)

	// The curve itself.
	var curve [][2]float64
	for r := rMax; r >= rMin-1e-9; r -= 0.05 {
		curve = append(curve, m(r, cibleMultiple(r)))
	}
	b.WriteString(smoothStroke(curve, figAccent, 2.4))

	for _, mk := range marks {
		px, py := x(mk.rate), y(cibleMultiple(mk.rate))
		if mk.rate == 2.5 { // the far end, labelled inside the zone
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="none" stroke="%s" stroke-width="1.8"/>`,
				px, py, figDeep)
			b.WriteString(mTxt(px-10, py-6, 11, figDeep, "end", "600", mk.label))
			continue
		}
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, px, py, figDeep)
		b.WriteString(mTxt(px+10, py+4, 11.5, figInk, "start", "600", mk.label))
	}

	b.WriteString(sTxt(plotL+8, 114, 10.5, figSoft, "start", "600",
		"au-delà de 33x, le risque dominant n'est plus la ruine :"))
	b.WriteString(sTxt(plotL+8, 128, 10.5, figMuted, "start", "400",
		"c'est d'avoir travaillé des années de trop"))

	for _, r := range []float64{5.0, 4.5, 4.0, 3.5, 3.0, 2.5} {
		b.WriteString(mTxt(x(r), 318, 10, figMuted, "middle", "400", frNum(r, 1)+" %"))
	}
	b.WriteString(sTxt(plotL, 340, 10.5, figMuted, "start", "400", "taux de retrait initial"))
	b.WriteString(sTxt(plotR, 340, 10.5, figMuted, "end", "400", "plus prudent, plus cher →"))

	b.WriteString(sTxt(24, 362, 9.5, figMuted, "start", "400",
		"De 25x à 33x, l'écart représente typiquement 3 à 6 ans de travail de plus pour un taux d'épargne de 40 à 50 %."))
	b.WriteString(sTxt(24, 374, 9.5, figMuted, "start", "400",
		"Courbe exacte, sans hypothèse de marché : c'est la définition du taux de retrait, pas un résultat de simulation."))
	return svg(640, 384, b.String())
}

// The worked case of étape 5, in the article's own words: Nadia and Marc read
// 3 400 €/month off 24 months of statements, add the life they aim at, gross
// the total up for tax friction, and multiply by the 3,5 % multiple. The
// pension is the one correction that points the other way: the article's
// simulation says the same plan ignoring it would have needed 200 000 € more.
const (
	cibleObserve  = 3400.0   // €/month, off 24 months of statements
	cibleLoisirs  = 350.0    // €/month, travel and leisure of the life aimed at
	cibleMutuelle = 220.0    // €/month, health cover now entirely theirs
	cibleFriction = 0.12     // tax + PUMa friction on the withdrawals
	cibleTaux     = 0.035    // the multiple chosen, as a rate: 28,6x
	ciblePension  = 200000.0 // what the same safety costs without the pension
)

// cibleCapital turns a monthly line of spending into the capital it demands:
// twelve months divided by the withdrawal rate. At 3,5 % one euro a month
// weighs 343 € of capital, which is why the article insists on the statements
// rather than on the multiple.
func cibleCapital(monthly float64) float64 { return monthly * 12 / cibleTaux }

// cibleNet is the capital the aimed-at spending demands before any tax, and
// cibleTarget the same after the friction is grossed up. cibleTarget is the
// article's 1 547 000 €.
func cibleNet() float64 {
	return cibleCapital(cibleObserve + cibleLoisirs + cibleMutuelle)
}
func cibleTarget() float64 { return cibleNet() / (1 - cibleFriction) }

// figCibleCascade walks that chain as one rising staircase in euros of
// capital, and closes on the only step that points down.
func figCibleCascade() string {
	const (
		kMax  = 1800.0 // the axis, in thousands of euros
		plotL = 76.0
		plotR = 616.0
		plotB = 336.0
		plotT = 96.0
		bw    = 68.0
	)
	m := mapper(0, 1, 0, kMax, 0, 1, plotB, plotT)
	y := func(euros float64) float64 { return m(0, euros/1000)[1] }
	slot := (plotR - plotL) / 5
	cx := func(i int) float64 { return plotL + slot*(float64(i)+0.5) }
	kTxt := func(v float64) string {
		return strings.TrimSuffix(euroFR(math.Round(v/1000)), " €")
	}

	var b strings.Builder
	b.WriteString(plateHead("combien il vous faut",
		"Du relevé bancaire au capital cible, marche par marche"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Nadia et Marc, étape 5 : à 3,5 %, un euro de dépense par mois pèse 343 € de capital"))

	for _, g := range []float64{0, 500, 1000, 1500} {
		gy := y(g * 1000)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(plotL, gy, plotR, gy, col, 1))
		b.WriteString(mTxt(plotL-10, gy+3.5, 10, figMuted, "end", "400", kTxt(g*1000)))
	}
	b.WriteString(sTxt(plotL, 86, 10.5, figMuted, "start", "400", "capital cible (k€)"))

	// The label block under each column: the section that produces the step,
	// its name, and the input the reader replaces with their own.
	label := func(i int, step, l1, l2, val, col string) {
		b.WriteString(sTxt(cx(i), 356, 9.5, figMuted, "middle", "400", step))
		b.WriteString(sTxt(cx(i), 371, 10.5, col, "middle", "600", l1))
		b.WriteString(sTxt(cx(i), 384, 10, figMuted, "middle", "400", l2))
		b.WriteString(mTxt(cx(i), 398, 10, figSoft, "middle", "600", val))
	}

	steps := []struct {
		amount       float64
		fill         string
		step, l1, l2 string
		val          string
	}{
		{cibleCapital(cibleObserve), figAccent, "étape 1", "dépenses observées", "24 mois de relevés", "3 400 €/mois"},
		{cibleCapital(cibleLoisirs), figBlue, "étape 1", "voyages et loisirs", "la vie visée", "+ 350 €/mois"},
		{cibleCapital(cibleMutuelle), figBlue, "étape 1", "mutuelle santé", "à votre charge", "+ 220 €/mois"},
		{cibleTarget() - cibleNet(), figBad, "étape 2", "friction fiscale", "impôts et PUMa", "+ 12 % du brut"},
	}
	level := 0.0
	for i, s := range steps {
		top := level + s.amount
		xb := cx(i) - bw/2
		b.WriteString(barV(xb, bw, y(level), y(top), s.fill))
		sign := "+ "
		if i == 0 {
			sign = ""
		}
		b.WriteString(mTxt(cx(i), y(top)-8, 11, figInk, "middle", "600", sign+kTxt(s.amount)))
		label(i, s.step, s.l1, s.l2, s.val, figSoft)
		b.WriteString(dashLine(xb+bw, y(top), cx(i+1)-bw/2, y(top), figMuted, 1, "3 3"))
		level = top
	}

	// The target, and above it the ghost of the same plan with the legal
	// pension left out of the simulation: 200 000 € taller.
	tgt := cibleTarget()
	xt := cx(4) - bw/2
	ghost := tgt + ciblePension
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="1.2" stroke-dasharray="4 3"/>`,
		xt, y(ghost), bw, y(tgt)-y(ghost), figWash, figGreen)
	b.WriteString(barV(xt, bw, y(0), y(tgt), figDeep))
	b.WriteString(mTxt(cx(4), y(tgt)+18, 11.5, "#fffdf9", "middle", "600", kTxt(tgt)))
	label(4, "étape 4", "cible", "× 28,6 (3,5 %)", euroFR(math.Round(tgt/1000)*1000), figDeep)

	// The credit, drawn as the descent it is.
	ax := cx(4)
	b.WriteString(line(ax, y(ghost)+3, ax, y(tgt)-6, figGreen, 1.4))
	fmt.Fprintf(&b, `<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		ax-4, y(tgt)-7, ax+4, y(tgt)-7, ax, y(tgt)-1, figGreen)
	b.WriteString(sTxt(plotR, 82, 10.5, figGreen, "end", "600",
		"− "+euroFR(ciblePension)+" : la retraite légale, comptée en revenu différé"))
	b.WriteString(sTxt(plotR, 95, 10, figMuted, "end", "400",
		"sans elle, le même plan aurait exigé "+euroFR(math.Round(ghost/1000)*1000)))

	b.WriteString(sTxt(24, 418, 9.5, figMuted, "start", "400",
		"Les deux marches que le calcul de comptoir oublie, la fiscalité et la pension, pèsent chacune plus lourd que le"))
	b.WriteString(sTxt(24, 430, 9.5, figMuted, "start", "400",
		"budget voyages, et en sens opposés. Les hypothèses sont celles de l'étape 5 : remplacez chacune par la vôtre."))
	return svg(640, 442, b.String())
}
