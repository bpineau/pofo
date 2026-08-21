package firebook

import (
	"fmt"
	"strings"
)

// The slider plate of the all-weather article. The article closes on a
// recommendation that is not a portfolio but a DOSE: a growth core plus 30 to
// 40 % of a regime pocket, "le demi-tous-temps", which it says buys most of the
// tail shortening for three to six tenths of a point of expectation. That
// sentence contains two quantities moving at different speeds, and prose can
// only assert the ratio between them.
//
// Swept and drawn, the ratio becomes visible, and so does the reason the
// recommendation stops where it does: the worst path improves fast up to about
// thirty points of dose and then flattens, while the return keeps falling. The
// plateau is not an opinion, it is where the two curves stop trading.
//
// Both curves are measured on the bundled series over the article's own window
// and frozen here; the guard test rebuilds them, and rebuilds the family's
// control table alongside, so a convention drifting anywhere shows up as the
// four known portfolios missing their published values.

// curseurCore is the growth portfolio the slider starts from: the article's own
// variant A, "un 70/30 classique en obligations intermédiaires".
const (
	curseurCoreEquities = 0.70
	curseurCoreBonds    = 0.30
)

// curseurPocket splits the regime pocket. The article's worked example builds
// it from long duration, gold and linkers (10 / 10 / 5 points out of a hundred);
// no real-dollar linker series is bundled, so the pocket is renormalized over
// the two that are, half and half, and the legend says so. It leaves the
// pocket's inflation cover resting on gold alone, which is thinner than the
// article's three imperfect hedges and is the plate's main caveat.
const (
	curseurPocketGold = 0.50
	curseurPocketLong = 0.50
)

// curseurStep is one dose of the slider: the share of the portfolio moved from
// the growth core into the regime pocket, and what that does to the annualized
// real return and to the deepest real drawdown, both in percent.
type curseurStep struct {
	dose, cagr, drawdown float64
	worst                float64 // the worst calendar year, in percent
}

// The sweep, measured over December 1971 to December 2024 in real terms with a
// December rebalancing, on the same legs and the same engine as the family
// plate of this article (SP500, IEF, TLT, XAUUSD, deflator ^CPI-US).
var curseurSweep = []curseurStep{
	{0, 5.94, -41.3, -16.7},
	{10, 5.91, -35.3, -12.5},
	{20, 5.85, -28.9, -8.3},
	{30, 5.74, -25.6, -4.1},
	{40, 5.60, -25.4, -7.7},
}

// The plateau the article recommends, in points of dose.
const (
	curseurPlateauLo = 30.0
	curseurPlateauHi = 40.0
)

// curseurAt reads one step of the sweep.
func curseurAt(dose float64) curseurStep {
	for _, s := range curseurSweep {
		if s.dose == dose {
			return s
		}
	}
	return curseurStep{}
}

// curseurCost is what a dose costs in annualized real return against the naked
// core, in points; curseurBought is the depth of worst path it removes, also in
// points.
func curseurCost(dose float64) float64 {
	return curseurSweep[0].cagr - curseurAt(dose).cagr
}

func curseurBought(dose float64) float64 {
	return curseurAt(dose).drawdown - curseurSweep[0].drawdown
}

// The plate's geometry: one dose axis, and one value axis per curve, each
// keyed to its curve's colour because they measure different things.
const (
	curX0, curX1     = 90.0, 560.0
	curTop, curBot   = 115.0, 310.0
	curCagrLo        = 5.4
	curCagrHi        = 6.0
	curDrawLo        = -45.0
	curDrawHi        = -20.0
	curDoseLo        = 0.0
	curDoseHi        = 40.0
	curLabelOffsetUp = -11.0
	curLabelOffsetDn = 17.0
)

func curseurScales() (dose, cagr, draw figScale) {
	return figScale{Min: curDoseLo, Max: curDoseHi, Px0: curX0, Px1: curX1},
		figScale{Min: curCagrLo, Max: curCagrHi, Px0: curBot, Px1: curTop},
		figScale{Min: curDrawLo, Max: curDrawHi, Px0: curBot, Px1: curTop}
}

func figTousTempsCurseur() string {
	dose, cagr, draw := curseurScales()
	var b strings.Builder
	b.WriteString(plateHead("le curseur tous-temps", "Le plancher s'achète vite, et se paie peu"))
	b.WriteString(plateDeck(
		"Ce que chaque dose de poche de régimes retire au rendement, et ce qu'elle retire au pire recul"))
	legendChips(&b, 74, [][2]string{
		{figDeep, "rendement réel annualisé"},
		{figBlue, "pire recul réel"},
	})

	// The plateau the article recommends, marked before anything is drawn over
	// it: it is the ground the two curves are read against.
	px0, px1 := dose.Map(curseurPlateauLo), dose.Map(curseurPlateauHi)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		px0, curTop, px1-px0, curBot-curTop, mixHex("#fffdf9", figGreen, 0.10))
	b.WriteString(sTxt((px0+px1)/2, curTop-7, 10, figGreen, "middle", "600", "plateau recommandé"))

	// The return axis rules the plate; the drawdown axis only labels itself, on
	// the other side, in its own colour.
	axisTicks(&b, cagr, []float64{5.4, 5.6, 5.8, 6.0}, 1, "", curX0, curX1, false)
	b.WriteString(sTxt(curX0-10, curTop-14, 9, figDeep, "end", "400", "% par an"))
	for _, v := range []float64{-45, -40, -35, -30, -25, -20} {
		b.WriteString(mTxt(curX1+10, draw.Map(v)+3.5, 10, figMuted, "start", "400", frMinus(v, 0)+" %"))
	}
	b.WriteString(sTxt(curX1+10, curTop-14, 9, figBlue, "start", "400", "pire recul"))
	for _, s := range curseurSweep {
		b.WriteString(mTxt(dose.Map(s.dose), curBot+18, 10, figMuted, "middle", "400",
			fmt.Sprintf("%.0f %%", s.dose)))
	}
	b.WriteString(sTxt(curX0, curBot+37, 9.5, figMuted, "start", "400",
		"dose de poche de régimes, en % du portefeuille  →"))

	// The two curves, each with its readings on its own side of its points.
	var cagrPts, drawPts [][2]float64
	for _, s := range curseurSweep {
		cagrPts = append(cagrPts, [2]float64{dose.Map(s.dose), cagr.Map(s.cagr)})
		drawPts = append(drawPts, [2]float64{dose.Map(s.dose), draw.Map(s.drawdown)})
	}
	b.WriteString(poly(drawPts, figBlue, 2.6, ""))
	b.WriteString(poly(cagrPts, figDeep, 2.6, ""))
	for i, s := range curseurSweep {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`,
			cagrPts[i][0], cagrPts[i][1], figDeep)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`,
			drawPts[i][0], drawPts[i][1], figBlue)
		// The last pair sits against the right-hand axis, so its readings are
		// hung to the left instead of centred over the dots.
		anchor, dx := "middle", 0.0
		if i == len(curseurSweep)-1 {
			anchor, dx = "end", -9
		}
		b.WriteString(mTxt(cagrPts[i][0]+dx, cagrPts[i][1]+curLabelOffsetUp, 10.5, figDeep, anchor, "600",
			frNum(s.cagr, 2)))
		b.WriteString(mTxt(drawPts[i][0]+dx, drawPts[i][1]+curLabelOffsetDn, 10.5, figBlue, anchor, "600",
			frMinus(s.drawdown, 0)))
	}

	// What the whole sweep bought and what it cost, cotéd once.
	b.WriteString(plateConclusion(374, fmt.Sprintf(
		"Quarante points de poche retirent %s points de recul et coûtent %s point de rendement.",
		frNum(curseurBought(40), 0), frNum(curseurCost(40), 1))))
	b.WriteString(plateFoot(396, []string{
		"Cœur de croissance : 70 % actions US, 30 % Treasuries 7-10 ans (le 70/30 de l'article). " +
			"Poche de régimes : moitié or,",
		"moitié Treasuries 20 ans et plus. L'article partage sa poche entre or, duration longue et linkers ; " +
			"aucune série de linkers",
		"en dollars réels n'est disponible ici, donc la poche est renormalisée sur les deux autres, " +
			"et sa couverture d'inflation tient à l'or seul.",
		"US, décembre 1971 à décembre 2024, en réel (IPC américain), rééquilibrage chaque décembre, " +
			"recul maximal de l'indice mensuel réel.",
		"Le plateau : de 0 à 30 % de dose, le recul remonte de " + frNum(curseurBought(30), 0) +
			" points ; de 30 à 40 %, de " + frNum(curseurAt(40).drawdown-curseurAt(30).drawdown, 1) +
			" point de plus, pour " + frNum(curseurCost(40)-curseurCost(30), 2) + " de rendement.",
		"La pire année civile passe de " + frMinus(curseurAt(0).worst, 1) + " % à " +
			frMinus(curseurAt(30).worst, 1) + " % à 30 %, puis remonte à " + frMinus(curseurAt(40).worst, 1) +
			" % à 40 % : trop de duration longue a payé 2022.",
	}))
	return svg(640, 490, b.String())
}
