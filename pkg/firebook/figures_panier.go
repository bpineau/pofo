package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The two plates of "Suivre l'inflation". They answer two different questions
// and take deliberately different forms: the first lays a weighted sum flat,
// post by post, so the reader can repeat the gesture on their own statement;
// the second is a wedge that opens with time, measured in years of spending
// rather than in index points.
//
// Both are closed-form arithmetic. The basket is the article's own worked
// example (the "exemple" block), the drift plate is compound interest.
// figures_panier_test.go re-derives every number.

// panierFR formats a number the French way, with a comma decimal separator.
func panierFR(v float64, dec int) string {
	return strings.Replace(fmt.Sprintf("%.*f", dec, v), ".", ",", 1)
}

// panierPoste is one budget line of the household's basket: its weight in the
// budget and the yearly rise of its own prices, both in percent. The
// contribution to the household's inflation is the product, in points a year.
type panierPoste struct {
	name         string
	weight, rate float64
	lead         bool // the line the article singles out (health)
}

func (p panierPoste) contribution() float64 { return p.weight * p.rate / 100 }

// panierPostes is Denise and Paul's basket exactly as the article states it,
// ordered by contribution, largest first. The weights add up to 100 % and the
// contributions to 2,63 points, the "~2,6 %" of the prose.
var panierPostes = []panierPoste{
	{"Services, aide, assurances", 22, 3.0, false},
	{"Santé et mutuelle", 14, 4.5, true},
	{"Loisirs et voyages", 24, 2.0, false},
	{"Alimentation", 18, 2.0, false},
	{"Énergie et transport", 12, 2.5, false},
	{"Divers", 10, 2.0, false},
}

// panierIPCH is the published index the household is compared with (%/yr).
const panierIPCH = 2.1

// figPanierContributions lays the weighted sum flat: one bar per budget line,
// its length the contribution to the household's inflation, each starting
// where the previous one ended so the six of them land on the total. The
// index is a vertical marker, and the excess is what the plan has to budget.
func figPanierContributions() string {
	const (
		xZero, xFull = 210.0, 566.0
		vMax         = 2.8
		y0, pitch    = 96.0, 38.0
		bh           = 15.0
		yTot         = 330.0 // the summary bar
		yGrid        = 352.0 // where the vertical grid stops
	)
	x := func(v float64) float64 { return xZero + v/vMax*(xFull-xZero) }

	var b strings.Builder
	b.WriteString(plateHead("votre inflation", "Le panier de Denise et Paul, poste par poste"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"chaque barre : le poids du poste × la hausse de ses prix, en points d'inflation par an"))

	// vertical grid, ticked at the top so the bottom stays free for the total
	for _, g := range []float64{0, 1, 2} {
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(x(g), 86, x(g), yGrid, col, 1))
		b.WriteString(mTxt(x(g), 80, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	// the published index, the line the household is measured against
	b.WriteString(dashLine(x(panierIPCH), 86, x(panierIPCH), yGrid, figBlue, 1.4, "5 4"))
	b.WriteString(sTxt(x(panierIPCH)+7, 80, 10.5, figBlue, "start", "600", "l'IPCH : 2,1 %"))

	cum := 0.0
	for i, p := range panierPostes {
		y := y0 + pitch*float64(i)
		c := p.contribution()
		fill := figAccent
		if p.lead {
			fill = figDeep
		}
		b.WriteString(sTxt(200, y+11, 11.5, figSoft, "end", "600", p.name))
		b.WriteString(mTxt(200, y+26, 9.5, figMuted, "end", "400",
			fmt.Sprintf("%.0f %% × %s %%", p.weight, panierFR(p.rate, 1))))
		b.WriteString(barH(x(cum), x(cum+c), y, bh, fill))
		b.WriteString(mTxt(x(cum+c)+7, y+11.5, 10.5, figInk, "start", "600", panierFR(c, 2)))
		cum += c
	}

	// the note the plate exists for, in the empty band the staircase leaves,
	// tied to the health bar by a chip of its own colour
	fmt.Fprintf(&b, `<rect x="234" y="258" width="10" height="10" rx="2.5" fill="%s"/>`, figDeep)
	b.WriteString(sTxt(252, 267, 11, figDeep, "start", "600", "La santé ne pèse que 14 % du budget,"))
	b.WriteString(sTxt(252, 283, 10.5, figSoft, "start", "400", "et fait un quart de l'inflation du ménage."))

	// the total, on the same scale
	b.WriteString(line(24, 316, xFull, 316, figGrid, 1))
	b.WriteString(sTxt(200, yTot+13, 11.5, figInk, "end", "600", "Inflation du ménage"))
	b.WriteString(barH(x(0), x(cum), yTot, 18, figSoft))
	b.WriteString(mTxt(x(cum)+7, yTot+13, 11, figInk, "start", "600", panierFR(cum, 2)+" %"))

	// the gap to the index, bracketed under the total bar
	gapL, gapR := x(panierIPCH), x(cum)
	b.WriteString(line(gapL, 360, gapR, 360, figBlue, 1.4))
	b.WriteString(line(gapL, 356, gapL, 364, figBlue, 1.4))
	b.WriteString(line(gapR, 356, gapR, 364, figBlue, 1.4))
	b.WriteString(sTxt((gapL+gapR)/2, 376, 10.5, figBlue, "middle", "600", "+0,5 point au-dessus de l'indice"))
	return svg(640, 388, b.String())
}

// panierDrifts are the two yearly drifts above the index the article asks the
// reader to plan for, in fraction per year (+0,3 and +0,5 point).
var panierDrifts = []float64{0.003, 0.005}

// panierManque returns what a household whose basket drifts drift above the
// index every year fails to cover after n years, when its income is indexed on
// the index alone: the sum of the yearly shortfalls (1+drift)^t − 1, expressed
// in years of today's spending. At 30 years it is 1,44 year for +0,3 point and
// 2,44 years for +0,5, the "année et demie" of the prose.
func panierManque(drift float64, n int) float64 {
	tot := 0.0
	for t := 1; t <= n; t++ {
		tot += math.Pow(1+drift, float64(t)) - 1
	}
	return tot
}

// figEcartCompose plots that shortfall over thirty years for both drifts, in
// years of spending rather than in index points: a tenth of a point a year is
// nothing before year ten and one to two whole years of spending by year
// thirty.
func figEcartCompose() string {
	const (
		years          = 30
		px0, px1       = 90.0, 516.0
		pyBase, pyTop  = 300.0, 86.0
		vMax           = 2.6
		labelX         = 524.0
		endLabelOffset = 2.0
	)
	m := mapper(0, years, 0, vMax, px0, px1, pyBase, pyTop)
	curve := func(d float64) [][2]float64 {
		pts := make([][2]float64, 0, years+1)
		for t := 0; t <= years; t++ {
			pts = append(pts, m(float64(t), panierManque(d, t)))
		}
		return pts
	}
	low, up := curve(panierDrifts[0]), curve(panierDrifts[1])

	var b strings.Builder
	b.WriteString(plateHead("l'écart composé", "Un demi-point par an, compté en années de dépenses"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"ce qu'une indexation sur le seul indice ne paie pas, cumulé année après année"))

	// the cost zone: the whole area under the worse drift, the softer tint
	// under the recommended one, so the wedge between the two stands out
	b.WriteString(smoothAreaBelow(up, pyBase, figBadWash))
	b.WriteString(smoothAreaBelow(low, pyBase, figWash))
	// two references drawn on top of the tints, at one and two years
	for _, g := range []float64{1, 2} {
		gy := m(0, g)[1]
		b.WriteString(dashLine(px0, gy, px1, gy, figRule, 1, "4 4"))
	}
	for _, tick := range []struct {
		v     float64
		label string
	}{{0, "0"}, {1, "1 année"}, {2, "2 années"}} {
		b.WriteString(mTxt(px0-8, m(0, tick.v)[1]+3.5, 10, figMuted, "end", "400", tick.label))
	}

	// the anchors: a guide, a dot and a value at 10, 20 and 30 years
	for _, t := range []int{10, 20, 30} {
		b.WriteString(dashLine(m(float64(t), 0)[0], pyBase, m(float64(t), 0)[0],
			m(float64(t), panierManque(panierDrifts[1], t))[1], figRule, 1, "2 4"))
	}
	b.WriteString(smoothStroke(up, figBad, 2.2))
	b.WriteString(smoothStroke(low, figAccent, 2.2))
	for i, d := range panierDrifts {
		col := figAccent
		if i == 1 {
			col = figBad
		}
		for _, t := range []int{10, 20, 30} {
			p := m(float64(t), panierManque(d, t))
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s" stroke="#fffdf9" stroke-width="1.4"/>`,
				p[0], p[1], col)
		}
		// year 20: both readings stacked under the curves, right of the guide,
		// where no stroke can run into them (worse drift on top, as drawn)
		b.WriteString(mTxt(m(20, 0)[0]+9, 278-16*float64(i), 10.5, col, "start", "600",
			panierFR(panierManque(d, 20), 1)+" année"))
		// year 30, direct labels off the last point
		p30 := m(float64(years), panierManque(d, years))
		unit := " année"
		if panierManque(d, years) >= 2 {
			unit = " années"
		}
		b.WriteString(mTxt(labelX, p30[1]-endLabelOffset, 11, col, "start", "600",
			panierFR(panierManque(d, years), 1)+unit))
		b.WriteString(sTxt(labelX, p30[1]+12, 10.5, figMuted, "start", "400",
			"+"+panierFR(d*100, 1)+" point/an"))
	}

	// what the plate is for: the same gap, read twice
	b.WriteString(sTxt(106, 152, 11, figSoft, "start", "600", "Avant l'année 10, l'écart est invisible :"))
	b.WriteString(sTxt(106, 168, 10.5, figMuted, "start", "400", "deux à trois mois de dépenses en tout."))
	b.WriteString(sTxt(106, 188, 11, figSoft, "start", "600", "Après l'année 25, il coûte une à deux années."))
	b.WriteString(sTxt(106, 204, 10.5, figMuted, "start", "400", "En niveau de prix, l'écart n'est que de 9 à 16 %."))

	// the index itself is the baseline
	b.WriteString(line(px0, pyBase, px1, pyBase, figRule, 1))
	b.WriteString(sTxt(labelX, pyBase+2, 10.5, figMuted, "start", "400", "l'indice officiel"))
	for _, t := range []int{0, 10, 20, 30} {
		b.WriteString(mTxt(m(float64(t), 0)[0], 318, 10, figMuted, "middle", "400", fmt.Sprintf("%d", t)))
	}
	b.WriteString(sTxt((px0+px1)/2, 340, 11, figMuted, "middle", "400", "années de retrait"))
	return svg(640, 360, b.String())
}
