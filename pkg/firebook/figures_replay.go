package firebook

import (
	"fmt"
	"strings"
)

// The plates of "Sept façons de vivre du même portefeuille": three retirements
// replayed on the real record, two figures each.
//
// The first shows the weather (the reference portfolio's real value, with not a
// single withdrawal taken out of it), the second shows the seven lives lived
// inside that weather, as small multiples. Small multiples rather than seven
// overlaid curves: on paper and on a reader, seven lines on one grid is a
// thicket, while seven panels sharing one scale let the eye compare shapes,
// which is the whole question here. A long slow slide and an abrupt cut look
// nothing alike, and that difference is the article's subject.
//
// Every number below is produced by pkg/replay from the bundled record, and
// figures_replay_test.go recomputes them and fails if a series drifts. They are
// frozen here on purpose: the book's figures must stay pure, dependency-free
// functions, exactly like the rest of the plates.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.

// replayTarget is the household's planned income, in k€ per year, the reference
// line every income panel is read against.
const replayTarget = 24.0

// replayRuleNames labels the seven panels, in the order pkg/replay returns them.
var replayRuleNames = []string{
	"Retrait fixe", "Flex −10 %", "Guardrails (GK)", "Guardrails par risque",
	"% borné", "Amortissement (ABW)", "% du portefeuille (VPW)",
}

// --- 1973: the crisis arrives first ---

// replayIndex1973 is the reference 60/40's real value at each year end, base 100
// at the end of 1972, with no withdrawals taken.
var replayIndex1973 = []float64{
	100.0, 85.0, 65.6, 77.0, 87.5, 78.6, 75.0, 74.6, 80.4, 74.7, 90.1, 100.0, 105.8, 130.1, 150.1,
	149.7, 160.4, 190.1, 183.3, 222.3, 230.2, 246.0, 237.2, 299.3, 333.5, 405.2, 483.5, 525.9,
	504.6, 477.9, 425.2, 490.0, 510.0, 506.4, 549.2, 565.7, 467.3, 521.8, 570.1, 577.0, 625.1,
}

// replayIncome1973 is the real income each rule delivered, k€ per year.
var replayIncome1973 = [][]float64{
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6},
	{24.0, 21.6, 19.4, 17.5, 17.5, 17.5, 15.7, 14.2, 14.2, 14.2, 14.2, 14.2, 14.2, 14.2, 15.6, 15.6, 17.1, 18.9, 18.9, 20.7, 20.7, 20.7, 20.7, 22.8, 25.1, 27.6, 30.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4},
	{21.6, 19.4, 17.5, 15.7, 15.7, 14.2, 12.8, 12.8, 12.8, 12.8, 12.8, 12.8, 12.8, 14.0, 15.4, 17.0, 18.7, 20.5, 22.6, 24.9, 27.3, 30.1, 30.1, 33.1, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0},
	{24.0, 23.4, 22.8, 22.2, 21.7, 21.1, 20.6, 20.1, 19.6, 19.1, 18.6, 18.2, 17.7, 17.3, 16.8, 16.4, 16.0, 16.8, 16.4, 17.2, 18.1, 19.0, 18.5, 19.4, 20.4, 21.4, 22.5, 23.6, 24.8, 26.0, 25.4, 26.5, 26.5, 25.9, 26.3, 26.0, 25.4, 24.7, 24.1, 23.5},
	{29.1, 23.8, 17.7, 19.9, 21.8, 18.8, 17.3, 16.5, 17.1, 15.3, 17.7, 18.9, 19.3, 22.8, 25.3, 24.2, 25.0, 28.4, 26.4, 30.8, 30.6, 31.5, 29.2, 35.4, 37.9, 44.3, 50.8, 53.2, 49.1, 44.7, 38.2, 42.3, 42.4, 40.5, 42.2, 41.8, 33.2, 35.6, 37.4, 36.4},
	{24.0, 19.6, 14.5, 16.3, 17.8, 15.4, 14.1, 13.4, 13.9, 12.4, 14.4, 15.3, 15.6, 18.4, 20.3, 19.5, 20.0, 22.8, 21.1, 24.6, 24.4, 25.1, 23.2, 28.1, 30.0, 35.0, 40.1, 41.9, 38.6, 35.1, 30.0, 33.2, 33.1, 31.6, 32.9, 32.5, 25.8, 27.7, 29.0, 28.2},
}

// --- 1985: the opposite problem ---

var replayIndex1985 = []float64{
	100.0, 122.9, 141.8, 141.4, 151.5, 179.6, 173.2, 210.1, 217.5, 232.5, 224.1, 282.8, 315.1,
	382.8, 456.8, 496.9, 476.8, 451.5, 401.8, 463.0, 481.8, 478.4, 518.9, 534.5, 441.6, 493.0,
	538.7, 545.2, 590.6, 685.6, 747.9, 746.5, 781.8, 867.9, 836.1, 988.9, 1110.3, 1196.4, 960.2,
	1095.6, 1230.4,
}

var replayIncome1985 = [][]float64{
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 24.0},
	{24.0, 24.0, 26.4, 26.4, 26.4, 29.0, 29.0, 31.9, 31.9, 35.1, 35.1, 38.7, 42.5, 46.8, 51.4, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 56.0, 56.0, 61.6, 61.6, 61.6},
	{21.6, 21.6, 21.6, 21.6, 21.6, 23.8, 23.8, 26.1, 28.7, 31.6, 31.6, 34.8, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0},
	{24.0, 25.2, 26.5, 27.8, 29.2, 30.6, 32.2, 33.8, 35.5, 37.2, 37.1, 39.0, 40.9, 43.0, 45.1, 47.4, 49.7, 52.2, 51.1, 53.7, 56.3, 54.9, 56.1, 55.5, 54.1, 52.8, 51.5, 50.2, 49.0, 51.5, 54.1, 55.1, 55.4, 58.2, 56.7, 59.6, 62.5, 65.7, 64.0, 62.4},
	{29.1, 34.4, 38.2, 36.6, 37.8, 43.0, 39.9, 46.5, 46.3, 47.6, 44.1, 53.5, 57.4, 67.0, 76.9, 80.4, 74.2, 67.6, 57.8, 64.1, 64.1, 61.2, 63.8, 63.2, 50.2, 53.9, 56.6, 55.1, 57.4, 64.1, 67.2, 64.5, 65.0, 69.3, 64.2, 73.0, 78.9, 81.7, 63.1, 69.2},
	{24.0, 28.3, 31.4, 30.0, 30.9, 35.2, 32.5, 37.9, 37.7, 38.6, 35.8, 43.3, 46.3, 54.0, 61.9, 64.6, 59.6, 54.1, 46.2, 51.2, 51.1, 48.7, 50.7, 50.2, 39.8, 42.6, 44.7, 43.5, 45.2, 50.4, 52.7, 50.5, 50.8, 54.2, 50.1, 56.9, 61.3, 63.4, 48.9, 53.5},
}

// --- 2000: the one still running ---

var replayIndex2000 = []float64{
	100.0, 96.0, 90.9, 80.9, 93.2, 97.0, 96.3, 104.4, 107.6, 88.9, 99.2, 108.4, 109.7, 118.9, 138.0,
	150.5, 150.2, 157.3, 174.7, 168.3, 199.0, 223.4, 240.8, 193.2, 220.5, 247.6, 274.4,
}

var replayIncome2000 = [][]float64{
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 24.0, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6},
	{24.0, 24.0, 24.0, 21.6, 21.6, 21.6, 19.4, 19.4, 19.4, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 19.2, 19.2, 19.2, 19.2},
	{21.6, 19.4, 17.5, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 17.3, 19.1, 19.1, 19.1, 21.0, 21.0, 23.1, 25.4, 27.9, 27.9, 27.9, 30.7},
	{24.0, 23.4, 22.8, 22.2, 21.7, 21.1, 20.6, 20.1, 19.6, 19.1, 18.6, 18.2, 17.7, 17.3, 17.0, 17.8, 17.3, 17.1, 18.0, 17.5, 18.4, 19.3, 20.3, 19.8, 19.3, 19.3},
	{29.1, 26.9, 24.5, 21.0, 23.2, 23.2, 22.2, 23.1, 22.9, 18.2, 19.5, 20.5, 20.0, 20.8, 23.2, 24.4, 23.4, 23.5, 25.1, 23.3, 26.5, 28.6, 29.6, 22.9, 25.1, 27.1},
	{24.0, 22.1, 20.1, 17.2, 19.0, 19.0, 18.1, 18.8, 18.6, 14.8, 15.8, 16.6, 16.1, 16.8, 18.7, 19.6, 18.8, 18.9, 20.1, 18.6, 21.1, 22.8, 23.5, 18.1, 19.9, 21.4},
}

// figReplayMarche1973 and its siblings draw the weather: what the reference
// portfolio itself did, with not one euro withdrawn.
func figReplayMarche1973() string {
	return replayMarketPlate("janvier 1973", "La crise d'abord : le portefeuille sans un seul retrait",
		1973, replayIndex1973, 700,
		"Creux à 65 fin 1974, et il faut attendre 1983 pour repasser durablement au-dessus de 100.")
}

func figReplayMarche1985() string {
	return replayMarketPlate("janvier 1985", "Le problème inverse : quarante ans de vent arrière",
		1985, replayIndex1985, 1300,
		"Le capital est multiplié par 12 en termes réels, 1987, 2000, 2008 et 2022 compris.")
}

func figReplayMarche2000() string {
	return replayMarketPlate("janvier 2000", "Celle qu'on vit : une décennie perdue, puis le rattrapage",
		2000, replayIndex2000, 300,
		"Treize ans pour retrouver la valeur réelle de départ, et le plan en compte encore quatorze.")
}

// replayMarketPlate draws one era's untouched portfolio: a single line, year
// ends, base 100 at the opening December.
func replayMarketPlate(kicker, title string, start int, idx []float64, ymax float64, note string) string {
	n := len(idx) - 1
	x := func(i int) float64 { return 74 + float64(i)/float64(n)*(600-74) }
	y := func(v float64) float64 { return 268 - v/ymax*(268-76) }

	var b strings.Builder
	b.WriteString(plateHead(kicker, title))
	b.WriteString(sTxt(74, 62, 10.5, figMuted, "start", "400",
		"valeur réelle du portefeuille, base 100 au départ, aucun retrait"))
	step := ymax / 4
	for g := 0.0; g <= ymax+1; g += step {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(74, gy, 600, gy, col, 1))
		b.WriteString(mTxt(66, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	// The curve opens on the December before the retirement, so point i is the
	// end of year start-1+i. One tick every decade, plus the last year.
	for i := 0; i <= n; i++ {
		yr := start - 1 + i
		if i == n || (yr%10 == 0 && n-i > 3) { // the last label wins over a neighbouring decade
			b.WriteString(mTxt(x(i), 286, 10, figMuted, "middle", "400", fmt.Sprintf("%d", yr)))
		}
	}
	pts := make([][2]float64, len(idx))
	for i, v := range idx {
		pts[i] = [2]float64{x(i), y(v)}
	}
	b.WriteString(dashLine(74, y(100), 600, y(100), figMuted, 1, "4 4"))
	b.WriteString(sTxt(604, y(100)+3.5, 9.5, figMuted, "start", "400", "100"))
	b.WriteString(poly(pts, figAccent, 2.4, ""))
	// The deepest point of the opening decade, named.
	lo, at := idx[0], 0
	for i, v := range idx[:min(11, len(idx))] {
		if v < lo {
			lo, at = v, i
		}
	}
	if at > 0 {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, x(at), y(lo), figBad)
		b.WriteString(mTxt(x(at)+8, y(lo)+3.5, 10, figBad, "start", "600", frNum(lo, 0)))
	}
	b.WriteString(mTxt(600, y(idx[n])-9, 10, figDeep, "end", "600", frNum(idx[n], 0)))
	b.WriteString(sTxt(74, 308, 10.5, figMuted, "start", "400", note))
	return svg(640, 322, b.String())
}

// figReplayRevenus1973 and its siblings draw the seven lives lived inside that
// weather: one panel per rule, one common scale, the plan as a dotted line.
func figReplayRevenus1973() string {
	return replayIncomePlate("janvier 1973", "Sept revenus dans la même crise",
		1973, replayIncome1973, 56,
		"Le fixe ne bouge jamais et finit à 19 k€ de capital. Les autres coupent, plus ou moins tôt, plus ou moins fort.")
}

func figReplayRevenus1985() string {
	return replayIncomePlate("janvier 1985", "Sept revenus dans le même vent arrière",
		1985, replayIncome1985, 85,
		"Personne ne manque d'argent. La question devient : qui a osé dépenser ce que le portefeuille pouvait payer ?")
}

func figReplayRevenus2000() string {
	return replayIncomePlate("janvier 2000", "Sept revenus dans la décennie perdue",
		2000, replayIncome2000, 34,
		"Vingt-six ans sur quarante. Six règles sur sept ont déjà fait vivre le ménage en dessous de son plan.")
}

// replayIncomePlate draws the seven income paths as small multiples: four
// panels then three, sharing one scale, each with the planned income as a
// dotted reference and its own average and worst year in the sub-title.
func replayIncomePlate(kicker, title string, start int, income [][]float64, ymax float64, note string) string {
	const (
		left    = 24.0
		colW    = 142.0
		gap     = 12.0
		panelH  = 74.0
		row0Top = 96.0
		row1Top = 218.0
	)
	n := len(income[0]) - 1

	var b strings.Builder
	b.WriteString(plateHead(kicker, title))
	b.WriteString(plateDeck(
		fmt.Sprintf("revenu réel servi, k€ par an, de %d à %d ; même échelle partout, pointillé = les 24 k€ prévus",
			start, start+n)))

	for i, vals := range income {
		col, row := i%4, i/4
		px := left + float64(col)*(colW+gap)
		top := row0Top
		if row == 1 {
			top = row1Top
		}
		x := func(k int) float64 { return px + float64(k)/float64(n)*colW }
		y := func(v float64) float64 { return top + panelH - v/ymax*panelH }

		// Panel title and its two numbers.
		lo, hi, sum := vals[0], vals[0], 0.0
		for _, v := range vals {
			lo, hi, sum = min(lo, v), max(hi, v), sum+v
		}
		b.WriteString(sTxt(px, top-22, 10, figInk, "start", "600", replayRuleNames[i]))
		b.WriteString(mTxt(px, top-10, 9.5, figMuted, "start", "400",
			fmt.Sprintf("moy. %s · pire %s", frNum(sum/float64(len(vals)), 1), frNum(lo, 1))))

		// A hairline every decade, so a reader can place 2000 or 2008 inside a
		// panel, with the last two digits of the year underneath.
		for k := 0; k <= n; k++ {
			if (start+k)%10 != 0 {
				continue
			}
			b.WriteString(line(x(k), top, x(k), y(0), figGrid, 1))
			b.WriteString(mTxt(x(k), top+panelH+11, 8.5, figMuted, "middle", "400",
				fmt.Sprintf("%02d", (start+k)%100)))
		}
		// Baseline, the planned income, and the path itself.
		b.WriteString(line(px, y(0), px+colW, y(0), figRule, 1))
		b.WriteString(dashLine(px, y(replayTarget), px+colW, y(replayTarget), figMuted, 1, "3 3"))
		pts := make([][2]float64, len(vals))
		for k, v := range vals {
			pts[k] = [2]float64{x(k), y(v)}
		}
		b.WriteString(poly(pts, figAccent, 1.8, ""))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="2.6" fill="%s"/>`, x(n), y(vals[n]), figDeep)
	}
	// The eighth cell carries the scale rather than an eighth rule.
	px := left + 3*(colW+gap)
	top := row1Top
	b.WriteString(sTxt(px, top-22, 10, figInk, "start", "600", "L'échelle"))
	b.WriteString(mTxt(px, top-10, 9.5, figMuted, "start", "400", fmt.Sprintf("0 à %s k€ par an", frNum(ymax, 0))))
	for _, g := range []float64{0, replayTarget, ymax} {
		gy := top + panelH - g/ymax*panelH
		b.WriteString(line(px, gy, px+34, gy, figGrid, 1))
		b.WriteString(mTxt(px+38, gy+3.5, 9.5, figMuted, "start", "400", frNum(g, 0)))
	}
	b.WriteString(sTxt(24, 328, 10.5, figMuted, "start", "400", note))
	return svg(640, 342, b.String())
}

// frNum formats a number the French way, with a comma for the decimal mark.
func frNum(v float64, decimals int) string {
	return strings.Replace(fmt.Sprintf("%.*f", decimals, v), ".", ",", 1)
}
