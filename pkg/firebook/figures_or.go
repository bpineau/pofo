package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The gold plate of the withdrawal-portfolio part. The article states gold's
// double thesis in prose (a secular real return near zero, yet decades that
// never look like zero); this draws it as one signed bar per ten-year holding
// window, so the regimes and the flat long-run average sit in the same frame.
//
// EVERYTHING IS REAL, in dollars: gold quotes in USD and is deflated here by
// the US CPI, the same deflator pkg/replay uses. A euro investor's experience
// differs by the dollar move, which the plate says out loud.

// The decades of gold, in real dollars. Bars are ANNUALIZED real returns over
// each ten-year window (%/yr), not cumulative ones: the annualized form keeps
// the +21.6 %/yr of the 1970s from crushing the other ten bars, puts the last
// (six-year) window on the same axis, and compares directly with the "0 to 1 %
// real per year" the article claims for the secular record. The cumulative
// totals of the loudest windows are quoted under the plate.
//
// How to reproduce: read pkg/datasets/simdata/XAUUSD.csv, keep the last close
// of each month, divide it by the ^CPI-US level of the same month (the bundled
// snapshot marketdata serves offline, i.e. pkg/replay's deflator), then take
// December-to-December ratios ten years apart. Windows start every five years
// from 1970 (the record opens in April 1968, so December 1969 is the first
// usable base). figures_or_test.go recomputes all of it and fails on drift.
var (
	orDecadeFirst = 1970 // first window: December 1969 to December 1979
	orDecadeStep  = 5    // one window every five years
	orDecadeAnn   = []float64{21.60, -2.01, -7.17, -1.35, -5.81, -1.07, 11.36, 8.24, 1.54, 5.04}
	orDecadeTotal = []float64{607.0, -18.4, -52.5, -12.7, -45.1, -10.2, 193.3, 120.8, 16.5, 63.4}

	// The running window, kept as its own case: December 2019 to December 2025
	// is six years, not ten, so only its annualized form is comparable.
	orPartialStart = 2020
	orPartialEnd   = 2025
	orPartialAnn   = 14.54
	orPartialTotal = 125.8

	// Averages of the ten complete windows, with and without the single window
	// that holds the 1970s. The second one is what is left of gold once the
	// post-1971 repricing leaves the sample.
	orDecadeMean     = 3.04
	orDecadeMeanEx70 = 0.97
)

// orDecadeLabel names a window the way the axis shows it ("1970-79").
func orDecadeLabel(start int) string { return fmt.Sprintf("%d-%02d", start, (start+9)%100) }

// orSigned formats a signed percentage the French way, with a true minus sign.
func orSigned(v float64, decimals int) string {
	s := frNum(math.Abs(v), decimals)
	if v < 0 {
		return "−" + s
	}
	return "+" + s
}

// barVOutlined draws the barV shape hollow, for a bar that is not the same kind
// of measurement as its neighbours (here: an incomplete window). barV emits a
// single <path .../>, so appending the stroke to its only self-closing tag is
// enough.
func barVOutlined(x, w, yBase, yEnd float64, fill, stroke string) string {
	return strings.Replace(barV(x, w, yBase, yEnd, fill), "/>",
		fmt.Sprintf(` stroke="%s" stroke-width="1.6"/>`, stroke), 1)
}

func figOrDecennies() string {
	const (
		x0, x1   = 62.0, 616.0
		yTop     = 96.0  // +23.5 %/yr
		yBot     = 306.0 // −9.5 %/yr
		vLo, vHi = -9.5, 23.5
		bw       = 30.0
	)
	n := len(orDecadeAnn) + 1 // the ten complete windows plus the running one
	pitch := (x1 - x0) / float64(n)
	cx := func(i int) float64 { return x0 + pitch*(float64(i)+0.5) }
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("l'or en portefeuille",
		"Les décennies de l'or, en réel : cinq dans le rouge, trois à deux chiffres"))
	b.WriteString(plateDeck(
		"Une fenêtre de dix ans tous les cinq ans. Or en dollars, déflaté par le CPI américain."))

	// The purgatory: five consecutive windows underwater, 1975 to 2004.
	wx0, wx1 := cx(1)-pitch/2, cx(5)+pitch/2
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		wx0, yTop, wx1-wx0, yBot-yTop, figWash)
	b.WriteString(sTxt((wx0+wx1)/2, 108, 10.5, figSoft, "middle", "600",
		"cinq fenêtres d'affilée dans le rouge"))
	b.WriteString(mTxt((wx0+wx1)/2, 130, 10, figMuted, "middle", "400", "1975-2004"))

	// Grid, then the zero rule on top of it.
	for _, g := range []float64{-5, 0, 5, 10, 15, 20} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(x0, gy, x1, gy, col, 1))
		lbl := "0"
		if g != 0 {
			lbl = orSigned(g, 0)
		}
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", lbl))
	}

	// The average of the ten complete windows, and what is left of it without
	// the 1970s. Both are quoted; only the first is drawn, with a short leader.
	my := y(orDecadeMean)
	b.WriteString(dashLine(x0, my, x1, my, figDeep, 1.2, "5 4"))
	b.WriteString(sTxt(cx(1)+4, 176, 10.5, figDeep, "start", "600",
		"moyenne des dix fenêtres : "+orSigned(orDecadeMean, 1)+" %/an"))
	b.WriteString(sTxt(cx(1)+4, 190, 10.5, figMuted, "start", "400",
		"sans la fenêtre 1970-79 : "+orSigned(orDecadeMeanEx70, 1)+" %/an"))
	b.WriteString(dashLine(cx(3), 198, cx(3), my, figMuted, 1, "2 3"))

	// One bar per window, the running one hollow.
	bar := func(i int, v float64, partial bool) {
		x, top := cx(i)-bw/2, y(v)
		fill, col := figAccent, figDeep
		if v < 0 {
			fill, col = figBad, figBad
		}
		if partial {
			b.WriteString(barVOutlined(x, bw, y(0), top, figBandOuter, figAccent))
		} else {
			b.WriteString(barV(x, bw, y(0), top, fill))
		}
		ly := top - 8
		switch {
		case v < 0:
			ly = top + 13
		case math.Abs(ly-my) < 9:
			ly = my - 8 // keep the average rule out of a short bar's label
		}
		b.WriteString(mTxt(cx(i), ly, 10.5, col, "middle", "600", orSigned(v, 1)))
		lbl := orDecadeLabel(orDecadeFirst + orDecadeStep*i)
		if partial {
			lbl = fmt.Sprintf("%d-%02d", orPartialStart, orPartialEnd%100)
		}
		b.WriteString(mTxt(cx(i), 322, 9.5, figSoft, "middle", "400", lbl))
	}
	for i, v := range orDecadeAnn {
		bar(i, v, false)
	}
	bar(len(orDecadeAnn), orPartialAnn, true)
	b.WriteString(sTxt(cx(len(orDecadeAnn)), 334, 9, figMuted, "middle", "400", "partielle"))

	b.WriteString(sTxt((x0+x1)/2, 354, 11, figMuted, "middle", "400",
		"rendement réel annualisé sur la fenêtre, en % par an"))
	b.WriteString(sTxt(24, 376, 10, figMuted, "start", "400",
		fmt.Sprintf("En cumulé : %s %% réel sur 1970-79, %s %% sur 1980-89. Dernière barre partielle : déc. 2019 à déc. 2025.",
			orSigned(math.Round(orDecadeTotal[0]), 0), orSigned(math.Round(orDecadeTotal[2]), 0))))
	b.WriteString(sTxt(24, 390, 10, figMuted, "start", "400",
		"En euros, l'expérience diffère : le change du dollar s'ajoute au prix de l'or."))
	return svg(640, 402, b.String())
}
