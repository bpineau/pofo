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
// window END, January 2001: the quarter century the article argues over, and far
// enough inside the trend leg's real net asset values (1996 on) that every
// trailing 252-day window is made of them.
const trendCorrStart = 2001*12 + 0

// trendCorrBand is the half-width of the "near zero" band, the correlation a
// practitioner reads as flat. The plate counts how many months fall outside it.
const trendCorrBand = 0.2

// trendCorrMinIdx and trendCorrMaxIdx are the offsets, into trendCorrPoints, of
// the period's most negative and most positive readings: April 2003 (the twelve
// months that contain the 2002 crash, the programme short equities) and February
// 2006 (twelve months of a steady bull market, the programme long equities). The
// trough is a plateau, not a spike: the six readings from January to June 2003
// all sit within four hundredths of it, which is the point the article makes.
const (
	trendCorrMinIdx = 27
	trendCorrMaxIdx = 61
)

// trendCorrPoints is the rolling correlation of the daily returns of a
// diversified trend programme and of the S&P 500, over a trailing window of 252
// common trading days (twelve months), one reading per month end from January
// 2001 to June 2026 (306 windows), each dated by the month that CLOSES it.
//
// Reproduce: read the bundled CTA (Simplify CTA, a diversified trend programme:
// real net asset values of trend programmes over the whole plotted window, the
// fund's own quotes from 2022) and SP500 (S&P 500 total
// return) with marketdata.ReadSimdataFS(datasets.Simdata(), id); keep the days
// both series quote; take simple daily returns; then at every month end from
// January 2001, excluding the final incomplete month, compute the Pearson
// correlation of the last 252 returns of each. figures_trendcorr_test.go
// recomputes all 306 values from pkg/datasets and fails if the plate and the
// record disagree.
//
// The trend leg is made of REAL net asset values across the whole plotted
// window: the fund itself from 2022, and the net asset values of other managers'
// programmes before it, back to 1996. The futures-price reconstruction only
// covers what those cannot reach, which is earlier than this plate starts.
var trendCorrPoints = []float64{
	-0.01, -0.07, -0.21, -0.39, -0.49, -0.49, -0.51, -0.54, -0.59, -0.62, -0.60, -0.61,
	-0.60, -0.59, -0.58, -0.52, -0.57, -0.61, -0.65, -0.66, -0.66, -0.70, -0.72, -0.72,
	-0.73, -0.74, -0.75, -0.76, -0.74, -0.72, -0.65, -0.58, -0.53, -0.43, -0.37, -0.33,
	-0.23, -0.14, 0.14, 0.23, 0.29, 0.35, 0.36, 0.32, 0.31, 0.24, 0.29, 0.28,
	0.33, 0.32, 0.31, 0.34, 0.37, 0.37, 0.36, 0.41, 0.41, 0.52, 0.53, 0.54,
	0.57, 0.58, 0.57, 0.56, 0.55, 0.57, 0.55, 0.52, 0.56, 0.53, 0.53, 0.53,
	0.47, 0.44, 0.40, 0.39, 0.37, 0.29, 0.27, 0.24, 0.17, 0.16, 0.05, 0.01,
	-0.05, -0.07, -0.04, -0.07, -0.08, -0.09, -0.13, -0.11, -0.04, -0.24, -0.32, -0.34,
	-0.36, -0.38, -0.42, -0.42, -0.39, -0.34, -0.31, -0.31, -0.33, -0.26, -0.10, -0.00,
	0.11, 0.18, 0.37, 0.44, 0.43, 0.28, 0.18, 0.11, 0.07, 0.12, 0.17, 0.19,
	0.17, 0.25, 0.29, 0.31, 0.33, 0.42, 0.48, 0.45, 0.33, 0.18, 0.07, 0.04,
	0.00, -0.03, -0.04, -0.03, -0.06, -0.11, -0.14, -0.28, -0.17, -0.00, 0.09, 0.16,
	0.21, 0.25, 0.18, 0.20, 0.18, 0.30, 0.39, 0.45, 0.44, 0.44, 0.44, 0.48,
	0.51, 0.51, 0.55, 0.52, 0.56, 0.54, 0.56, 0.53, 0.51, 0.38, 0.35, 0.28,
	0.17, 0.12, 0.08, 0.07, 0.08, 0.12, 0.10, -0.08, -0.21, -0.26, -0.26, -0.24,
	-0.31, -0.35, -0.39, -0.41, -0.42, -0.45, -0.46, -0.41, -0.28, -0.23, -0.19, -0.19,
	-0.02, 0.16, 0.24, 0.29, 0.32, 0.46, 0.49, 0.49, 0.44, 0.44, 0.44, 0.46,
	0.46, 0.45, 0.42, 0.40, 0.38, 0.37, 0.35, 0.31, 0.33, 0.31, 0.17, 0.07,
	-0.06, -0.20, -0.25, -0.27, -0.23, -0.20, -0.19, -0.22, -0.24, -0.28, -0.20, -0.14,
	0.06, 0.16, -0.25, -0.31, -0.35, -0.38, -0.38, -0.36, -0.32, -0.30, -0.28, -0.28,
	-0.25, -0.28, 0.05, 0.24, 0.34, 0.44, 0.47, 0.48, 0.49, 0.50, 0.55, 0.52,
	0.45, 0.38, 0.22, 0.14, 0.08, 0.02, -0.03, -0.06, -0.13, -0.17, -0.30, -0.29,
	-0.32, -0.32, -0.19, -0.17, -0.17, -0.16, -0.16, -0.16, -0.12, -0.10, -0.09, -0.08,
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
	// The peak sits in the first fifth of the plate, so its label reads to the
	// RIGHT of the dot: anchored the other way it runs off the left edge.
	b.WriteString(sTxt(lx+12, 106, 11, figSoft, "start", "600", "février 2006, douze mois de hausse"))
	b.WriteString(mTxt(lx+12, 120, 11, figDeep, "start", "600", "+"+frNum(lv, 2)))
	b.WriteString(sTxt(lx+52, 120, 10.5, figAccent, "start", "400", "acheteur d'actions"))

	tx, tv := x(trendCorrMinIdx), trendCorrPoints[trendCorrMinIdx]
	b.WriteString(dot(tx, y(tv), figBlue))
	b.WriteString(sTxt(tx+14, 300, 11, figSoft, "start", "600", "avril 2003, la fenêtre du krach"))
	b.WriteString(mTxt(tx+14, 314, 11, figBlue, "start", "600", "−"+frNum(-tv, 2)))
	b.WriteString(sTxt(tx+54, 314, 10.5, figBlue, "start", "400", "vendeur d'actions"))

	// The x axis says which of the window's two dates is plotted.
	for _, yr := range []int{2001, 2005, 2010, 2015, 2020, 2025} {
		b.WriteString(mTxt(x(float64(yr*12-trendCorrStart)), 336, 10, figMuted, "middle", "400", frNum(float64(yr), 0)))
	}

	for i, s := range []string{
		fmt.Sprintf("Moyenne de la période : +%s. Et pourtant %s %% des mois tombent hors de la bande ±%s, celle qu'un gérant lirait comme plate.",
			frNum(mean, 2), frNum(trendCorrOutside(), 0), frNum(trendCorrBand, 1)),
		"Trend : valeurs liquidatives réelles de programmes de suivi de tendance, dont le fonds lui-même depuis 2022. Actions : S&amp;P 500 total return.",
	} {
		b.WriteString(sTxt(24, 358+float64(i)*13, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 384, b.String())
}
