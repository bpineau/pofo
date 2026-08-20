package firebook

import (
	"fmt"
	"sort"
	"strings"
)

// The plate of the 4 % article: the sustainable rate, vintage by vintage. It
// exists to make one sentence visible, "c'est le taux du pire cas, pas une
// moyenne": the rule is not the middle of the record, it is its floor, and the
// floor is a handful of mid-sixties retirements.

// millesimesFirst is the first start year the plate draws, and
// millesimesRates the highest initial withdrawal rate (percent of the starting
// capital, then held constant in real terms) that a 50/50 American portfolio
// would have carried through the thirty years beginning that year.
//
// Source: pkg/datasets.BroadSample(), the bundled Jorda-Schularick-Taylor
// Macrohistory R6 panel, USA rows: real annual total returns of domestic
// equities and of government bonds, each already deflated by the US CPI.
//
// Recipe, per start year s in 1926..1991 (1991 is the last year whose thirty
// years all fit in the panel, which ends in 2020):
//
//  1. r(y) = 0.5*equity(y) + 0.5*bond(y), rebalanced every year, no fee, no tax;
//  2. the withdrawal happens at the START of each of the thirty years and is
//     held constant in real terms (the returns are already real, so nothing is
//     left to index), and the balance may reach exactly zero after the last one;
//  3. the highest rate that survives is then the exact annuity rate
//     100 / sum(t = 0..29) prod(i = 1..t) 1/(1+r_i), in percent.
//
// This is the same engine as figures_pays.go's SAFEMAX, run per vintage instead
// of taking the minimum over vintages, and on a 50/50 instead of a 60/40.
//
// figures_plancher_test.go recomputes all sixty-six values from pkg/datasets,
// plus the median, the minimum and its year, and fails on any drift.
const millesimesFirst = 1926

// millesimesBandFill tints everything the rule would not have carried. It is a
// PRE-BLENDED solid hex (figBad #c0655b at .14 composited onto the figure card
// background #fffdf9), never rgba: crengine paints rgba fills solid black.
const millesimesBandFill = "#F6E8E3"

var millesimesRates = []float64{
	7.138, 7.001, 6.106, 5.266, 5.476, 6.146, 7.598, 7.163, 5.719, 5.919,
	5.065, 4.387, 5.528, 5.000, 4.913, 5.197, 5.918, 6.309, 6.131, 5.864,
	5.004, 5.826, 6.856, 7.348, 6.866, 6.343, 6.396, 6.188, 6.313, 5.183,
	4.550, 4.720, 5.067, 4.636, 4.557, 4.485, 4.038, 4.235, 3.986, 3.744,
	3.672, 3.956, 3.883, 3.883, 4.617, 4.696, 4.399, 4.144, 5.061, 6.619,
	6.275, 5.846, 6.697, 7.467, 8.066, 8.517, 10.191, 9.346, 9.474, 9.807,
	8.550, 7.522, 8.417, 8.312, 7.519, 8.288,
}

// millesimesMedian is the median rate of the record. The sample is even-sized,
// so it is the average of the two middle vintages, 1945 and 1942.
func millesimesMedian() float64 {
	s := append([]float64(nil), millesimesRates...)
	sort.Float64s(s)
	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

// millesimesWorst returns the lowest rate of the record and the year that set
// it: the single vintage the whole rule is built on.
func millesimesWorst() (float64, int) {
	lo, at := millesimesRates[0], millesimesFirst
	for i, v := range millesimesRates {
		if v < lo {
			lo, at = v, millesimesFirst+i
		}
	}
	return lo, at
}

// millesimesBest returns the highest rate of the record and its year.
func millesimesBest() (float64, int) {
	hi, at := millesimesRates[0], millesimesFirst
	for i, v := range millesimesRates {
		if v > hi {
			hi, at = v, millesimesFirst+i
		}
	}
	return hi, at
}

// millesimesUnder counts the vintages that could not carry the given rate.
func millesimesUnder(rate float64) int {
	n := 0
	for _, v := range millesimesRates {
		if v < rate {
			n++
		}
	}
	return n
}

// figMillesimesSoutenables draws the sixty-six vintages as stems hanging off the
// 4 % rule itself: nearly all of them stand above the line, six consecutive
// mid-sixties retirements fall under it, and the deepest of those, 1966, is the
// number the whole rule is named after. The second reference is the median, and
// the distance between the two is the plate's message.
func figMillesimesSoutenables() string {
	const (
		x0, x1     = 68.0, 548.0
		yTop, yBot = 104.0, 372.0
		lo, hi     = 2.9, 10.6
		rule       = 4.0
	)
	n := len(millesimesRates)
	pitch := (x1 - x0) / float64(n)
	x := func(year int) float64 { return x0 + (float64(year-millesimesFirst)+0.5)*pitch }
	y := func(v float64) float64 { return yBot - (v-lo)/(hi-lo)*(yBot-yTop) }

	med := millesimesMedian()
	worst, worstYear := millesimesWorst()
	best, bestYear := millesimesBest()

	var b strings.Builder
	b.WriteString(plateHead("soixante-six départs en retraite",
		"Le taux qui a tenu trente ans, millésime par millésime"))
	b.WriteString(plateDeck(
		"Retrait initial maximal qu'un 50/50 américain aurait soutenu trente ans, en pouvoir d'achat constant"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		fmt.Sprintf("%d millésimes, de %d à %d, le dernier dont les trente années sont complètes",
			n, millesimesFirst, millesimesFirst+n-1)))

	// everything the rule would not have carried, tinted once
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, y(rule), x1-x0, yBot-y(rule), millesimesBandFill)

	// horizontal grid, in whole points of withdrawal rate
	for g := 4.0; g <= 10.0; g++ {
		gy := y(g)
		if g != rule {
			b.WriteString(line(x0, gy, x1, gy, figGrid, 1))
		}
		b.WriteString(mTxt(x0-9, gy+3.5, 10, figMuted, "end", "400", frNum(g, 0)+" %"))
	}
	b.WriteString(line(x0, yBot, x1, yBot, figRule, 1))

	// the decade ticks
	for yr := 1930; yr <= 1990; yr += 10 {
		b.WriteString(mTxt(x(yr), yBot+18, 10, figMuted, "middle", "400", fmt.Sprint(yr)))
	}

	// the stems, hanging off the 4 % rule: up in amber, down in red
	for i, v := range millesimesRates {
		col := figAccent
		if v < rule {
			col = figBad
		}
		b.WriteString(barV(x(millesimesFirst+i)-2.1, 4.2, y(rule), y(v), col))
	}

	// the two references, labelled in the right-hand gutter so no stem crosses
	// their text: the rule the article is about, and the middle of the record
	b.WriteString(line(x0, y(rule), x1, y(rule), figBad, 1.8))
	b.WriteString(sTxt(x1+10, y(rule)-2, 10, figBad, "start", "600", "la règle"))
	b.WriteString(mTxt(x1+10, y(rule)+11, 10.5, figBad, "start", "600", "4,0 %"))
	b.WriteString(dashLine(x0, y(med), x1, y(med), figDeep, 1.6, "5 4"))
	b.WriteString(sTxt(x1+10, y(med)-2, 10, figDeep, "start", "600", "médiane"))
	b.WriteString(mTxt(x1+10, y(med)+11, 10.5, figDeep, "start", "600", frNum(med, 1)+" %"))

	// the best vintage, labelled directly above its stem
	b.WriteString(mTxt(x(bestYear), y(best)-9, 10, figAccent, "middle", "600",
		fmt.Sprintf("%d · %s %%", bestYear, frNum(best, 1))))

	// the one vintage the rule is built on: the gap that separates it from the
	// median, drawn as a measure with caps, and a caption in the band under
	// every stem
	gx := x(worstYear)
	b.WriteString(line(gx, y(med), gx, y(worst), figDeep, 2))
	b.WriteString(line(gx-4, y(med), gx+4, y(med), figDeep, 2))
	b.WriteString(line(gx-4, y(worst), gx+4, y(worst), figDeep, 2))
	b.WriteString(mTxt(gx-9, 290, 10, figDeep, "end", "600", frNum(med-worst, 1)+" points"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.6" fill="none" stroke="%s" stroke-width="1.8"/>`,
		gx, y(worst), figBad)
	b.WriteString(sTxt(x0+8, 356, 10, figBad, "start", "600",
		"sous la règle : six départs, de 1964 à 1969"))
	b.WriteString(mTxt(gx+11, 356, 10.5, figBad, "start", "600",
		fmt.Sprintf("%d · %s %%", worstYear, frNum(worst, 2))))

	// the sentence the plate exists for
	b.WriteString(line(24, 410, 616, 410, figGrid, 1))
	b.WriteString(sTxt(24, 430, 11, figSoft, "start", "600",
		fmt.Sprintf("Le millésime médian aurait supporté %s %% de dépenses en plus que celui de %d.",
			frNum((med/worst-1)*100, 0), worstYear)))
	notes := []string{
		fmt.Sprintf("%d millésimes sur %d passent sous 4 %%, et ce sont six départs consécutifs, de 1964 à 1969. Les %d autres tiennent la règle.",
			millesimesUnder(rule), n, n-millesimesUnder(rule)),
		"Panel Jorda-Schularick-Taylor, États-Unis : actions et obligations d'État domestiques, rendements réels annuels déflatés du CPI américain.",
		"50/50 rééquilibré chaque année, prélèvement en début d'année, ni frais ni impôt ; le capital a le droit de finir exactement à zéro.",
		fmt.Sprintf("Bengen publie %s %% pour %d. Sa reconstruction n'est pas celle-ci : obligations à moyen terme, autre indice d'actions, données mensuelles.",
			frNum(4.15, 2), worstYear),
		"L'ordre de grandeur tient, la décimale non. Le pire millésime, lui, est le même des deux côtés.",
	}
	for i, s := range notes {
		b.WriteString(sTxt(24, 452+float64(i)*15, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, int(452+float64(len(notes)-1)*15+16), b.String())
}
