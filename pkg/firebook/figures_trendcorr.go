package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The conditional-correlation plate of the managed-futures article. Its subject
// is one sentence of that article, "sa corrélation nulle est une moyenne, pas un
// état": a trend programme is always positioned, so its correlation of the
// moment follows its book, short equities in a long bear, long equities in a
// long bull. The famous zero only appears once the regimes are averaged.
//
// The form follows that message. The wandering line is a diverging ribbon around
// zero, amber above (the programme is long equities), blue below (short), so the
// eye reads the sign before the level. The period mean is a flat dashed rule, and
// it sits inside a hairline band at ±0,2, the "near zero" reading a practitioner
// would call flat. The two records are annotated where they happen. What the
// reader must catch is the excursions against the flat rule, not the rule.

// Pre-blended solid area fills, never rgba (crengine paints rgba solid black):
// each is its stroke colour composited once onto the figure card background
// #fffdf9 at .32, light enough to keep the labels readable through the ribbon.
// figTrendCorrUp = figAccent (180,120,60), figTrendCorrDown = figBlue (58,109,180).
const (
	figTrendCorrUp   = "#E7D2BD"
	figTrendCorrDown = "#C0CFE3"
)

// trendCorrStart is the month key (year*12 + month - 1) of the first plotted
// window END, January 2001: the first month end whose trailing 252 trading days
// fall in the window the SG Trend pin of the reconstruction was calibrated on.
const trendCorrStart = 2001*12 + 0

// trendCorrBand is the half-width of the "near zero" band, the correlation a
// practitioner reads as flat. The plate counts how many months fall outside it.
const trendCorrBand = 0.2

// trendCorrMinIdx and trendCorrMaxIdx are the offsets, into trendCorrPoints, of
// the period's most negative and most positive readings: May 2003 (the twelve
// months that contain the 2002 crash, the programme short equities) and January
// 2018 (twelve months of the quiet melt-up, the programme long equities). The
// trough is a plateau, not a spike: July 2009, the 2008 window, reads within a
// hundredth of it, which is the point the article makes.
const (
	trendCorrMinIdx = 28
	trendCorrMaxIdx = 204
)

// trendCorrPoints is the rolling correlation of the daily returns of a
// diversified trend programme and of the S&P 500, over a trailing window of 252
// common trading days (twelve months), one reading per month end from January
// 2001 to June 2026 (306 windows), each dated by the month that CLOSES it.
//
// Reproduce: read the bundled CTA (Simplify CTA, a diversified trend programme:
// a daily TSMOM reconstruction pinned to the SG Trend Index information ratio,
// with the fund's real quotes grafted from 2022) and SP500 (S&P 500 total
// return) with marketdata.ReadSimdataFS(datasets.Simdata(), id); keep the days
// both series quote; take simple daily returns; then at every month end from
// January 2001, excluding the final incomplete month, compute the Pearson
// correlation of the last 252 returns of each. figures_trendcorr_test.go
// recomputes all 306 values from pkg/datasets and fails if the plate and the
// record disagree.
//
// THE TREND LEG IS A RECONSTRUCTION before 2022, not a live track record: the
// plate's caption says so, as the article's own data callout already does.
var trendCorrPoints = []float64{
	0.02, -0.06, -0.22, -0.41, -0.53, -0.55, -0.58, -0.60, -0.66, -0.70, -0.68, -0.67,
	-0.65, -0.65, -0.64, -0.60, -0.63, -0.65, -0.71, -0.70, -0.70, -0.73, -0.76, -0.76,
	-0.78, -0.78, -0.79, -0.79, -0.79, -0.79, -0.71, -0.66, -0.59, -0.48, -0.42, -0.36,
	-0.27, -0.17, 0.08, 0.19, 0.31, 0.39, 0.45, 0.44, 0.47, 0.44, 0.49, 0.48,
	0.53, 0.51, 0.50, 0.54, 0.54, 0.54, 0.52, 0.53, 0.52, 0.57, 0.58, 0.60,
	0.62, 0.64, 0.63, 0.62, 0.60, 0.62, 0.62, 0.60, 0.62, 0.60, 0.59, 0.60,
	0.56, 0.63, 0.63, 0.63, 0.66, 0.66, 0.71, 0.72, 0.70, 0.71, 0.67, 0.64,
	0.58, 0.48, 0.34, 0.25, 0.19, 0.07, -0.09, -0.25, -0.47, -0.58, -0.65, -0.68,
	-0.71, -0.73, -0.74, -0.75, -0.77, -0.78, -0.78, -0.78, -0.76, -0.66, -0.54, -0.49,
	-0.37, -0.25, -0.10, 0.06, 0.31, 0.42, 0.56, 0.64, 0.70, 0.73, 0.73, 0.72,
	0.70, 0.66, 0.66, 0.64, 0.59, 0.60, 0.60, 0.55, 0.49, 0.36, 0.25, 0.22,
	0.20, 0.18, 0.13, 0.08, -0.01, -0.11, -0.18, -0.44, -0.51, -0.46, -0.34, -0.31,
	-0.26, -0.17, -0.13, -0.05, 0.03, 0.25, 0.38, 0.46, 0.51, 0.53, 0.52, 0.53,
	0.57, 0.57, 0.57, 0.65, 0.71, 0.70, 0.72, 0.70, 0.70, 0.64, 0.61, 0.48,
	0.33, 0.27, 0.27, 0.22, 0.23, 0.17, 0.07, -0.03, -0.15, -0.25, -0.28, -0.29,
	-0.33, -0.38, -0.48, -0.51, -0.58, -0.61, -0.62, -0.63, -0.46, -0.42, -0.41, -0.37,
	-0.26, -0.18, -0.07, 0.03, 0.13, 0.46, 0.52, 0.58, 0.54, 0.56, 0.71, 0.71,
	0.76, 0.75, 0.74, 0.71, 0.71, 0.71, 0.71, 0.70, 0.68, 0.62, 0.55, 0.42,
	0.34, 0.20, 0.07, 0.00, -0.13, -0.17, -0.20, -0.28, -0.30, -0.39, -0.40, -0.38,
	-0.29, -0.10, 0.15, 0.10, 0.10, 0.08, 0.08, 0.12, 0.16, 0.20, 0.21, 0.21,
	0.22, 0.20, 0.16, 0.32, 0.41, 0.57, 0.63, 0.63, 0.64, 0.65, 0.68, 0.70,
	0.67, 0.62, 0.47, 0.36, 0.28, 0.21, 0.14, 0.11, 0.01, -0.04, -0.16, -0.21,
	-0.24, -0.29, -0.19, -0.17, -0.17, -0.16, -0.16, -0.16, -0.12, -0.10, -0.09, -0.08,
	-0.07, -0.03, -0.15, -0.17, -0.17, -0.18, -0.17, -0.15, -0.16, -0.11, -0.07, -0.09,
	-0.07, -0.07, -0.10, -0.01, -0.03, -0.01, -0.01, 0.03, 0.04, 0.03, 0.06, 0.08,
	0.09, 0.08, 0.03, -0.12, -0.16, -0.15,
}

// trendCorrMean is the mean of the plotted readings, the "zero" of the article.
func trendCorrMean() float64 {
	var sum float64
	for _, v := range trendCorrPoints {
		sum += v
	}
	return sum / float64(len(trendCorrPoints))
}

// trendCorrOutside is the share of months, in percent, whose reading sits at or
// beyond ±trendCorrBand, i.e. outside what a practitioner would call flat.
func trendCorrOutside() float64 {
	var n float64
	for _, v := range trendCorrPoints {
		if math.Abs(v) >= trendCorrBand {
			n++
		}
	}
	return n / float64(len(trendCorrPoints)) * 100
}

// --- The rolling twelve-month correlation of trend and equities ---
func figTrendCorrelation() string {
	const (
		x0, x1     = 62.0, 548.0
		yTop, yBot = 96.0, 320.0
	)
	n := float64(len(trendCorrPoints) - 1)
	x := func(i float64) float64 { return x0 + i/n*(x1-x0) }
	y := func(v float64) float64 { return yTop + (1-v)/2*(yBot-yTop) }
	yz := y(0)
	dot := func(px, py float64, col string) string {
		return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`, px, py, col)
	}

	var b strings.Builder
	b.WriteString(plateHead("managed futures", "La corrélation aux actions n'est pas un état"))
	b.WriteString(sTxt(24, 62, 10.2, figMuted, "start", "400",
		"corrélation des rendements quotidiens du trend et du S&amp;P 500, fenêtre glissante de 12 mois"))

	// Grid, with the zero rule carrying the sign of the whole plate.
	for _, g := range []float64{0.8, 0.4, 0, -0.4, -0.8} {
		gy := y(g)
		col, lab := figGrid, "+"+frNum(g, 1)
		switch {
		case g == 0:
			col, lab = figRule, "0"
		case g < 0:
			lab = "−" + frNum(-g, 1)
		}
		b.WriteString(line(x0, gy, x1, gy, col, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", lab))
	}

	// The ribbon: one solid tint per side of zero, then the line over it.
	segs := scvSegments(trendCorrPoints, x, y)
	for _, s := range segs {
		fill := figTrendCorrDown
		if s.pos {
			fill = figTrendCorrUp
		}
		b.WriteString(scvArea(s.pts, yz, fill))
	}
	b.WriteString(line(x0, yz, x1, yz, figRule, 1.2))
	for _, s := range segs {
		col := figBlue
		if s.pos {
			col = figAccent
		}
		b.WriteString(poly(s.pts, col, 1.3, ""))
	}

	// The "flat" band and the period mean, the two flat references the
	// excursions are read against.
	for _, e := range []float64{trendCorrBand, -trendCorrBand} {
		gy := y(e)
		b.WriteString(dashLine(x0, gy, x1, gy, figRule, 1, "3 4"))
		sign := "+"
		if e < 0 {
			sign, e = "−", -e
		}
		b.WriteString(mTxt(x0-8, gy+3.5, 9.5, figMuted, "end", "400", sign+frNum(e, 1)))
	}
	mean := trendCorrMean()
	ym := y(mean)
	b.WriteString(dashLine(x0, ym, x1, ym, figDeep, 1.4, "6 4"))
	b.WriteString(sTxt(x1+6, ym-3, 10, figDeep, "start", "600", "moyenne"))
	b.WriteString(mTxt(x1+6, ym+10, 10.5, figDeep, "start", "600", "+"+frNum(mean, 2)))

	// The two records, annotated where they happen.
	lx, lv := x(trendCorrMaxIdx), trendCorrPoints[trendCorrMaxIdx]
	b.WriteString(dot(lx, y(lv), figDeep))
	b.WriteString(sTxt(lx-12, 106, 11, figSoft, "end", "600", "janvier 2018, douze mois de hausse"))
	b.WriteString(mTxt(lx-12, 120, 11, figDeep, "end", "600", "+"+frNum(lv, 2)))
	b.WriteString(sTxt(lx-52, 120, 10.5, figAccent, "end", "400", "acheteur d'actions"))

	tx, tv := x(trendCorrMinIdx), trendCorrPoints[trendCorrMinIdx]
	b.WriteString(dot(tx, y(tv), figBlue))
	b.WriteString(sTxt(tx+14, 300, 11, figSoft, "start", "600", "mai 2003, la fenêtre du krach"))
	b.WriteString(mTxt(tx+14, 314, 11, figBlue, "start", "600", "−"+frNum(-tv, 2)))
	b.WriteString(sTxt(tx+54, 314, 10.5, figBlue, "start", "400", "vendeur d'actions"))

	// The x axis says which of the window's two dates is plotted.
	for _, yr := range []int{2001, 2005, 2010, 2015, 2020, 2025} {
		b.WriteString(mTxt(x(float64(yr*12-trendCorrStart)), 336, 10, figMuted, "middle", "400", frNum(float64(yr), 0)))
	}

	for i, s := range []string{
		fmt.Sprintf("Moyenne de la période : +%s. Et pourtant %s %% des mois tombent hors de la bande ±%s, celle qu'un gérant lirait comme plate.",
			frNum(mean, 2), frNum(trendCorrOutside(), 0), frNum(trendCorrBand, 1)),
		"Trend : reconstruction quotidienne recalée sur le SG Trend Index. Actions : S&amp;P 500 total return.",
	} {
		b.WriteString(sTxt(24, 358+float64(i)*13, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 384, b.String())
}
