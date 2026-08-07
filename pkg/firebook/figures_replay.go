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
	100.0, 84.8, 65.5, 76.5, 87.4, 78.6, 75.2, 74.4, 79.6, 75.1, 90.0, 100.1,
	105.9, 129.7, 150.7, 149.6, 160.7, 190.8, 183.4, 221.8, 230.4, 246.8, 237.8,
	299.4, 334.9, 405.8, 485.6, 528.5, 503.9, 478.8, 424.0, 490.5, 510.9, 506.7,
	551.4, 566.0, 468.5, 525.8, 571.9, 577.3, 626.4,
}

// replayIncome1973 is the real income each rule delivered, k€ per year.
var replayIncome1973 = [][]float64{
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6},
	{24.0, 21.6, 19.4, 17.5, 17.5, 17.5, 15.7, 14.2, 14.2, 14.2, 14.2, 14.2, 14.2, 14.2, 15.6, 15.6, 17.1, 18.9, 18.9, 20.7, 20.7, 20.7, 20.7, 22.8, 25.1, 27.6, 30.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4, 33.4},
	{21.6, 19.4, 17.5, 15.7, 15.7, 14.2, 12.8, 12.8, 12.8, 12.8, 12.8, 12.8, 12.8, 14.0, 15.4, 17.0, 18.7, 20.5, 22.6, 24.9, 27.3, 30.1, 30.1, 33.1, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0},
	{24.0, 23.4, 22.8, 22.2, 21.7, 21.1, 20.6, 20.1, 19.6, 19.1, 18.6, 18.2, 17.7, 17.3, 16.8, 16.4, 16.0, 16.8, 16.4, 17.2, 18.1, 19.0, 18.5, 19.4, 20.4, 21.4, 22.5, 23.6, 24.8, 26.0, 25.4, 26.5, 26.5, 25.9, 26.4, 26.0, 25.3, 24.7, 24.1, 23.5},
	{29.1, 23.8, 17.7, 19.8, 21.8, 18.8, 17.3, 16.5, 16.9, 15.4, 17.7, 18.9, 19.3, 22.7, 25.4, 24.2, 25.0, 28.5, 26.4, 30.7, 30.7, 31.6, 29.3, 35.4, 38.1, 44.4, 51.1, 53.4, 49.0, 44.8, 38.1, 42.4, 42.4, 40.5, 42.4, 41.8, 33.3, 35.9, 37.6, 36.5},
	{24.0, 19.5, 14.5, 16.3, 17.8, 15.4, 14.1, 13.4, 13.8, 12.5, 14.4, 15.3, 15.6, 18.3, 20.4, 19.5, 20.1, 22.9, 21.1, 24.5, 24.4, 25.1, 23.3, 28.1, 30.2, 35.1, 40.3, 42.1, 38.6, 35.2, 29.9, 33.2, 33.2, 31.6, 33.0, 32.5, 25.9, 27.9, 29.1, 28.2},
}

// --- 1985: the opposite problem ---

var replayIndex1985 = []float64{
	100.0, 122.5, 142.3, 141.3, 151.8, 180.2, 173.2, 209.5, 217.6, 233.1, 224.6,
	282.7, 316.3, 383.2, 458.6, 499.1, 475.9, 452.2, 400.4, 463.2, 482.5, 478.5,
	520.8, 534.5, 442.5, 496.5, 540.1, 545.2, 591.6, 688.3, 748.7, 748.2, 782.4,
	869.3, 834.5, 990.1, 1111.1, 1198.4, 965.3, 1094.8, 1234.9,
}

var replayIncome1985 = [][]float64{
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 24.0, 24.0, 24.0, 24.0, 24.0, 21.6, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 26.4, 26.4, 26.4, 29.0, 29.0, 31.9, 31.9, 35.1, 35.1, 38.7, 42.5, 46.8, 51.4, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 56.6, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 50.9, 56.0, 56.0, 61.6, 61.6, 61.6},
	{21.6, 21.6, 21.6, 21.6, 21.6, 23.8, 23.8, 26.1, 28.7, 31.6, 31.6, 34.8, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0, 36.0},
	{24.0, 25.2, 26.5, 27.8, 29.2, 30.6, 32.2, 33.8, 35.5, 37.2, 37.2, 39.1, 41.0, 43.1, 45.2, 47.5, 49.8, 52.3, 51.0, 53.6, 56.3, 54.8, 56.4, 55.5, 54.1, 52.8, 51.5, 50.2, 49.2, 51.6, 54.2, 55.3, 55.5, 58.3, 56.8, 59.7, 62.6, 65.8, 64.1, 62.5},
	{29.1, 34.3, 38.3, 36.6, 37.8, 43.2, 39.9, 46.4, 46.4, 47.7, 44.2, 53.5, 57.6, 67.1, 77.2, 80.8, 74.1, 67.7, 57.6, 64.1, 64.2, 61.2, 64.1, 63.2, 50.3, 54.3, 56.8, 55.1, 57.5, 64.3, 67.3, 64.7, 65.0, 69.5, 64.1, 73.1, 78.9, 81.8, 63.4, 69.1},
	{24.0, 28.2, 31.5, 30.0, 30.9, 35.3, 32.5, 37.8, 37.7, 38.7, 35.8, 43.3, 46.5, 54.1, 62.1, 64.9, 59.4, 54.2, 46.1, 51.2, 51.2, 48.7, 50.9, 50.2, 39.9, 42.9, 44.8, 43.5, 45.3, 50.6, 52.8, 50.7, 50.9, 54.2, 50.0, 56.9, 61.3, 63.5, 49.1, 53.5},
}

// --- 2000: the one still running ---

var replayIndex2000 = []float64{
	100.0, 95.4, 90.6, 80.2, 92.8, 96.7, 95.9, 104.4, 107.1, 88.7, 99.5, 108.2,
	109.3, 118.5, 137.9, 150.0, 149.9, 156.8, 174.2, 167.2, 198.4, 222.6, 240.1,
	193.4, 219.4, 247.5, 273.8,
}

var replayIncome2000 = [][]float64{
	{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	{24.0, 24.0, 24.0, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6, 21.6},
	{24.0, 24.0, 21.6, 19.4, 19.4, 19.4, 19.4, 19.4, 19.4, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 17.5, 19.2, 19.2, 19.2, 19.2},
	{21.6, 19.4, 17.5, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 15.7, 17.3, 19.1, 19.1, 19.1, 21.0, 21.0, 23.1, 25.4, 27.9, 27.9, 27.9, 30.7},
	{24.0, 23.4, 22.8, 22.2, 21.7, 21.1, 20.6, 20.1, 19.6, 19.1, 18.6, 18.2, 17.7, 17.3, 16.9, 17.7, 17.2, 17.0, 17.8, 17.4, 18.3, 19.2, 20.1, 19.6, 19.1, 19.2},
	{29.1, 26.7, 24.4, 20.8, 23.1, 23.2, 22.1, 23.1, 22.8, 18.2, 19.6, 20.5, 19.9, 20.8, 23.2, 24.3, 23.3, 23.5, 25.1, 23.1, 26.4, 28.5, 29.5, 22.9, 24.9, 27.1},
	{24.0, 22.0, 20.0, 17.0, 18.9, 18.9, 18.0, 18.8, 18.5, 14.7, 15.9, 16.6, 16.1, 16.7, 18.7, 19.5, 18.7, 18.8, 20.0, 18.5, 21.0, 22.7, 23.5, 18.2, 19.8, 21.4},
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
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
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
