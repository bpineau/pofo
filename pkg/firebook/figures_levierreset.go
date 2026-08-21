package firebook

import (
	"fmt"
	"strings"
)

// The reset-frequency plate of the leverage chapter. Its subject is the one
// thing a leverage table never shows: two portfolios can carry the SAME ×2
// exposure and end a decade far apart, because the day the leverage is put back
// on target is itself a decision.
//
// The form is a frequency axis, three positions, from the fastest reset to the
// slowest, with one line per decade. What the reader has to catch is the SHAPE
// of each line, not its level: both peak in the middle and fall away on either
// side, steeply in a choppy decade and barely at all in a trending one. The two
// ends lose for opposite reasons, which is the finding, and a plain "daily is
// worse" would be a lie.
//
// The vertical scale is what makes the two decades comparable at all. Each
// point is read as a share of the FRICTIONLESS benchmark: twice the index's own
// compounded growth, financing deducted, i.e. the L·(mu - sigma^2/2) of
// rendements-arithmetiques-geometriques. Against that yardstick, the daily
// column's shortfall IS the beta slippage of that article, exactly
// (L^2-L)·sigma^2/2 per year, which for L = 2 is the realized variance. The
// benchmark is a measuring stick, not a product: no implementable strategy
// reaches it.
//
// Every number is frozen from pkg/datasets (the bundled S&P 500 total return
// and the 3-month bill), and figures_levierreset_test.go recomputes all of them.

// leverResetArms are the three reset frequencies the plate poses, fastest
// first: the daily-reset ETF, a monthly rebalanced position, and a one-year
// term loan rolled every January.
var leverResetArms = []string{"quotidien", "mensuel", "annuel"}

// leverResetL is the exposure every arm carries. It is the same in all three
// columns on purpose: only the reset frequency varies.
const leverResetL = 2.0

// leverResetWindow is one decade of the plate.
//
// vol and bill are the index's annualized volatility and the average 3-month
// bill rate over the window, in percent. index is the unlevered index's
// terminal multiple, mult the three arms' terminal multiples, and share their
// share of the frictionless benchmark, in percent.
type leverResetWindow struct {
	years, tag string
	color      string
	vol, bill  float64
	index      float64
	mult       [3]float64
	share      [3]float64
}

// The two decades, frozen from pkg/datasets. 2000-2009 is the choppy one, two
// bear markets around a rally at 22 % volatility; 2010-2019 the trending one,
// one long rise at 15 %.
var leverResetWindows = []leverResetWindow{
	{
		years: "2000-2009", tag: "décennie heurtée", color: figBad,
		vol: 22.23, bill: 2.69, index: 0.9090,
		mult:  [3]float64{0.3849, 0.4760, 0.3051},
		share: [3]float64{61.0, 75.4, 48.3},
	},
	{
		years: "2010-2019", tag: "décennie porteuse", color: figBlue,
		vol: 14.77, bill: 0.57, index: 3.5666,
		mult:  [3]float64{9.648, 10.203, 9.753},
		share: [3]float64{80.3, 84.9, 81.1},
	},
}

// leverResetSlippage is the beta slippage the window's volatility implies, in
// points of compounded return per year: (L^2 - L)/2 × sigma^2, which at L = 2
// is the variance itself. It is what separates the daily column from the 100 %
// benchmark, and the plate's conclusion prints it for both decades.
func (w leverResetWindow) slippage() float64 {
	return (leverResetL*leverResetL - leverResetL) / 2 * w.vol * w.vol / 100
}

// The term arm's worst moment, the honest cost of never resetting: left alone,
// a ×2 position gears itself up as it falls. These are the equity multiple and
// the effective leverage it reached on 20 November 2008, quoted on the plate
// next to the point they explain.
const (
	leverResetTermTrough = 0.026
	leverResetTermGear   = 32.56
)

// leverResetMultiple sets a terminal multiple the way the plate prints it: two
// decimals under one, where the difference between 0,38 and 0,31 is the whole
// reading, and one above it, where a third digit would be noise.
func leverResetMultiple(m float64) string {
	if m < 1 {
		return "×" + frNum(m, 2)
	}
	return "×" + frNum(m, 1)
}

// figLevierReset draws the frequency axis, the two decades on it, and the term
// arm's 2008 gearing where it happened.
func figLevierReset() string {
	const (
		xAxis, xRight = 150.0, 616.0
		yTop, yBot    = 108.0, 300.0
		vLo, vHi      = 40.0, 104.0
	)
	xs := []float64{195, 320, 445}
	rate := figScale{Min: vLo, Max: vHi, Px0: yBot, Px1: yTop}
	y := rate.Map

	var b strings.Builder
	b.WriteString(plateHead("levier et remise à niveau",
		"Le levier n'a pas qu'une taille, il a un rythme"))
	b.WriteString(plateDeck(
		"Capital final d'une exposition ×2 sur le S&amp;P 500, en part de ce que rendrait un levier sans frottement de chemin"))

	var legend [][2]string
	for _, w := range leverResetWindows {
		legend = append(legend, [2]string{w.color,
			fmt.Sprintf("%s, %s (σ %.0f %%)", w.years, w.tag, w.vol)})
	}
	legendChips(&b, 78, legend)

	// Axes and grid.
	axisTicks(&b, rate, []float64{40, 50, 60, 70, 80, 90, 100}, 0, " %", xAxis, xRight, false)
	b.WriteString(line(xAxis, yTop, xAxis, yBot, figRule, 1))
	b.WriteString(line(xAxis, yBot, xRight, yBot, figRule, 1))
	for i, name := range leverResetArms {
		b.WriteString(sTxt(xs[i], 320, 10.5, figSoft, "middle", "600", name))
	}
	b.WriteString(sTxt(xAxis, 342, 10.5, figMuted, "start", "400",
		"fréquence de remise à niveau du levier, de la plus rapide à la plus lente  →"))

	// The frictionless benchmark: the rule everything is read against.
	b.WriteString(dashLine(xAxis, y(100), xRight, y(100), figMuted, 1.4, "5 4"))
	b.WriteString(sTxt(xAxis+4, y(100)-7, 10, figMuted, "start", "400",
		"repère 100 % : 2 × la croissance composée de l'indice, financement déduit"))

	// One line per decade, its points labelled with the terminal multiple: the
	// choppy decade below its dots, the trending one above, so the two label
	// bands never meet.
	for _, w := range leverResetWindows {
		var pts [][2]float64
		for i := range leverResetArms {
			pts = append(pts, [2]float64{xs[i], y(w.share[i])})
		}
		b.WriteString(poly(pts, w.color, 2.6, ""))
		dy := 17.0
		if w.color == figBlue {
			dy = -10
		}
		for i := range leverResetArms {
			cx, cy := xs[i], y(w.share[i])
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.4" fill="%s"/>`, cx, cy, w.color)
			b.WriteString(mTxt(cx, cy+dy, 10.5, w.color, "middle", "600",
				leverResetMultiple(w.mult[i])))
		}
	}

	// The term arm's 2008, written where it is read: a ×2 left alone geared
	// itself to 32 and would have been called long before the decade ended.
	term := leverResetWindows[0]
	b.WriteString(sTxt(xs[2]+15, y(term.share[2])-4, 10, figBad, "start", "600",
		"creux de 2008 : "+leverResetMultiple(leverResetTermTrough)))
	b.WriteString(sTxt(xs[2]+15, y(term.share[2])+9, 10, figBad, "start", "600",
		fmt.Sprintf("levier effectif ×%.0f", leverResetTermGear)))

	b.WriteString(plateConclusion(372, fmt.Sprintf(
		"Le point « quotidien » est le repère moins σ² par an, le beta slippage : %s points en %s, %s en %s.",
		frNum(leverResetWindows[0].slippage(), 1), leverResetWindows[0].years,
		frNum(leverResetWindows[1].slippage(), 1), leverResetWindows[1].years)))
	b.WriteString(plateFoot(392, []string{
		"S&amp;P 500, rendement total nominal. Les trois colonnes portent la même exposition ×2, financée à 1 × le bon du Trésor 3 mois.",
		"Seule la fréquence de remise à niveau change. La colonne annuelle est un prêt à terme d'un an, renouvelé chaque 1er janvier.",
		"Le repère à 100 % double la croissance composée de l'indice, financement déduit. C'est un étalon, pas un produit achetable.",
		fmt.Sprintf("L'indice nu, sans levier, finit à %s sur %s, mieux que les trois colonnes, et à %s sur %s.",
			leverResetMultiple(leverResetWindows[0].index), leverResetWindows[0].years,
			leverResetMultiple(leverResetWindows[1].index), leverResetWindows[1].years),
		"Un ETF à levier réel paie en plus ses frais et un écart de financement au-dessus du taux du bon du Trésor.",
	}))
	return svg(640, 470, b.String())
}
