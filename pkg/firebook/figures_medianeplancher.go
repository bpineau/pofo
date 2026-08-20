package firebook

import (
	"fmt"
	"strings"
)

// The basket plate of the diversification article. Its decumulation section
// makes a two-handed claim: going from all equities to a basket of four bricks
// lowers the median wealth a little AND raises the safe withdrawal rate,
// because a retiree sizes a plan on the fifth percentile rather than on the
// average. Two numbers moving in opposite directions is exactly what prose
// cannot show and a slopegraph can: two vertical axes, one segment per
// portfolio, and the segments cross.
//
// The four numbers come from the simulator's own central model run over the
// bundled series; the guard test rebuilds them from end to end.

// The plan both portfolios are run through: a thirty-year horizon, a fixed
// inflation-indexed withdrawal, and the conventional definition of a safe rate,
// the one that leaves 95 % of the paths solvent. The capital is arbitrary (the
// plate reports a multiple of it and a rate), and the seed and the worker count
// are fixed so the reading is reproducible to the last digit.
const (
	plancherYears   = 30
	plancherRule    = 0.04 // the withdrawal the median wealth is read at
	plancherSuccess = 0.95
	plancherCapital = 1e6
	plancherPaths   = 20000
	plancherWorkers = 1
	plancherSeed    = 7
)

// plancherWindow is the common history the model is fitted on: the months all
// four bricks quote. The trend index bounds it at the start of 1987, which is
// why the plate says so rather than claiming a century.
const plancherWindow = "1987-2026"

// plancherPortfolio is one segment: its name, its weights over the article's
// four bricks (world equities, long government bonds, gold, trend), and the two
// ends of the segment, both frozen from the model.
type plancherPortfolio struct {
	name    string
	short   string // the name repeated at the right end of the segment
	weights []float64
	median  float64 // terminal real wealth at the 4 % rule, as a multiple of the capital
	swr     float64 // withdrawal rate leaving 95 % of the paths solvent, as a fraction
	color   string
	width   float64 // the basket's segment is the heavier of the two, on purpose
}

// The two portfolios. The basket's weights are the book's own type portfolio
// (60 % world equities, 25 % bonds, 7.5 % gold, 7.5 % trend), read here with
// the long government-bond brick the diversification article names.
var plancherPortfolios = []plancherPortfolio{
	{"100 % actions mondiales", "actions", []float64{1, 0, 0, 0}, 2.2070, 0.0337, figBlue, 2.2},
	{"Panier de quatre briques", "panier", []float64{0.60, 0.25, 0.075, 0.075}, 1.6656, 0.0415, figDeep, 3.4},
}

// plancherWealth sets a median terminal wealth: a multiple of the starting
// capital, one decimal.
func plancherWealth(v float64) string { return "×" + frNum(v, 1) }

// plancherRate sets a withdrawal rate as a percent with two decimals, the
// resolution at which the two portfolios actually differ.
func plancherRate(v float64) string { return frNum(v*100, 2) + " %" }

// The plate's geometry: two vertical axes, and nothing between them but the
// segments.
const (
	plaLeftX, plaRightX = 210.0, 470.0
	plaTop, plaBot      = 110.0, 330.0
	plaMedLo, plaMedHi  = 1.4, 2.4
	plaSwrLo, plaSwrHi  = 0.032, 0.043
)

func plancherScales() (med, swr figScale) {
	return figScale{Min: plaMedLo, Max: plaMedHi, Px0: plaBot, Px1: plaTop},
		figScale{Min: plaSwrLo, Max: plaSwrHi, Px0: plaBot, Px1: plaTop}
}

func figMedianePlancher() string {
	med, swr := plancherScales()
	var b strings.Builder
	b.WriteString(plateHead("médiane et plancher",
		"Ce que le panier vend, et ce qu'il achète avec"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Le même plan de trente ans, tenu par deux portefeuilles : ce qui reste au milieu des cas, "+
			"ce qu'on peut en retirer"))

	// The two axes, each with its own question written above it.
	b.WriteString(line(plaLeftX, plaTop, plaLeftX, plaBot, figRule, 1))
	b.WriteString(line(plaRightX, plaTop, plaRightX, plaBot, figRule, 1))
	b.WriteString(sTxt(plaLeftX, 96, 10.5, figInk, "middle", "600", "richesse médiane à 30 ans"))
	b.WriteString(sTxt(plaRightX, 96, 10.5, figInk, "middle", "600", "retrait sûr à 95 % de succès"))
	b.WriteString(sTxt(plaLeftX, plaBot+20, 9.5, figMuted, "middle", "400",
		"en multiple du capital de départ"))
	b.WriteString(sTxt(plaRightX, plaBot+20, 9.5, figMuted, "middle", "400",
		"en % du capital, retrait fixe indexé"))

	// The segments, drawn before their ends so no line crosses a dot.
	for _, p := range plancherPortfolios {
		b.WriteString(poly([][2]float64{
			{plaLeftX, med.Map(p.median)}, {plaRightX, swr.Map(p.swr)},
		}, p.color, p.width, ""))
	}
	for _, p := range plancherPortfolios {
		ly, ry := med.Map(p.median), swr.Map(p.swr)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5" fill="%s"/>`, plaLeftX, ly, p.color)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5" fill="%s"/>`, plaRightX, ry, p.color)
		b.WriteString(sTxt(plaLeftX-12, ly-3, 10.5, figSoft, "end", "600", p.name))
		b.WriteString(mTxt(plaLeftX-12, ly+12, 11.5, p.color, "end", "600", plancherWealth(p.median)))
		b.WriteString(mTxt(plaRightX+12, ry-2, 11.5, p.color, "start", "600", plancherRate(p.swr)))
		b.WriteString(sTxt(plaRightX+12, ry+12, 10, figMuted, "start", "400", p.short))
	}

	// What each axis costs and pays, cotéd right against the axis that carries
	// it: the whole trade, in two numbers.
	eq, basket := plancherPortfolios[0], plancherPortfolios[1]
	b.WriteString(plancherBracket(plaLeftX+20, med.Map(eq.median), med.Map(basket.median),
		figMinus+frNum(plancherDrop()*100, 0)+" %", "start"))
	b.WriteString(plancherBracket(plaRightX-20, swr.Map(basket.swr), swr.Map(eq.swr),
		"+"+frNum(plancherGain()*100, 2)+" pt", "end"))

	b.WriteString(sTxt(24, 374, 10.5, figSoft, "start", "600", fmt.Sprintf(
		"Le panier échange %s %% de médiane contre %s point de retrait sûr : "+
			"on se dimensionne sur le percentile 5.",
		frNum(plancherDrop()*100, 0), frNum(plancherGain()*100, 2))))
	b.WriteString(plateFoot(396, []string{
		"Modèle central du simulateur : Student-t indépendant d'une année sur l'autre, ajusté sur l'historique " +
			"commun des quatre briques,",
		"horizon 30 ans, 20 000 trajectoires, graine fixe. Richesse médiane : ce qu'il reste après 30 ans " +
			"de retrait fixe à 4 %.",
		"Retrait sûr : le taux fixe indexé sur l'inflation qui laisse 95 % des trajectoires solvables, " +
			"sans flexibilité ni garde-fous.",
		"Panier : 60 % actions mondiales, 25 % obligations d'État longues, 7,5 % or, 7,5 % suivi de tendance, " +
			"rééquilibré chaque mois.",
		"Fenêtre commune " + plancherWindow + ", bornée par l'indice de suivi de tendance ; " +
			"séries en dollars réels, déflatées par l'IPC américain.",
		"Un ordre de grandeur, pas un résultat universel : d'autres briques, d'autres poids ou une autre " +
			"fenêtre déplacent les deux nombres.",
	}))
	return svg(640, 506, b.String())
}

// plancherDrop is the share of median wealth the basket gives up, and
// plancherGain the extra safe withdrawal it buys, both as fractions.
func plancherDrop() float64 {
	return 1 - plancherPortfolios[1].median/plancherPortfolios[0].median
}

func plancherGain() float64 {
	return plancherPortfolios[1].swr - plancherPortfolios[0].swr
}

// plancherBracket cotes the gap between two points of one axis: a hairline with
// two ticks and its value beside it, on the side the anchor names.
func plancherBracket(x, yTop, yBot float64, label, anchor string) string {
	var b strings.Builder
	b.WriteString(line(x, yTop, x, yBot, figMuted, 1.2))
	tick := 5.0
	if anchor == "end" {
		tick = -5
	}
	b.WriteString(line(x, yTop, x+tick, yTop, figMuted, 1.2))
	b.WriteString(line(x, yBot, x+tick, yBot, figMuted, 1.2))
	dx := 8.0
	if anchor == "end" {
		dx = -8
	}
	b.WriteString(mTxt(x+dx, (yTop+yBot)/2+4, 10.5, figSoft, anchor, "600", label))
	return b.String()
}
