package firebook

import (
	"fmt"
	"strings"
)

// The buckets plate. The article's first count of the indictment says the
// bucket recipe is an allocation in disguise, then leaves the arithmetic to
// the reader. This plate does the arithmetic once, continuously, over the
// withdrawal rates a retiree actually chooses between: the two lower buckets
// are counted in YEARS OF SPENDING, so their weight in the portfolio is that
// number of years times the withdrawal rate, and the equity sleeve is
// whatever is left. Nothing is estimated and no market data is involved.

const (
	// The canonical recipe the article judges: two years of spending in cash
	// (bucket 1) and six in bonds (bucket 2), the rest in equities.
	bucketsCashYears = 2.0
	bucketsBondYears = 6.0
	// The rate the article's own key block prices the recipe at (~7/21/72).
	bucketsRateRef = 3.5
	bucketsRateMin = 2.5
	bucketsRateMax = 5.0
)

// bucketsSplit returns the cash, bond and equity weights (percent of the
// portfolio) the canonical bucket recipe produces at an initial withdrawal
// rate of rate percent. The first two buckets hold a fixed number of years of
// spending, so they cost that many times the rate; equities are the residual.
func bucketsSplit(rate float64) (cash, bonds, equity float64) {
	cash = bucketsCashYears * rate
	bonds = bucketsBondYears * rate
	return cash, bonds, 100 - cash - bonds
}

// bucketsReadout formats one column of the plate, "5 / 15 / 80".
func bucketsReadout(rate float64) string {
	c, b, e := bucketsSplit(rate)
	return fmt.Sprintf("%.0f / %.0f / %.0f", c, b, e)
}

// figBucketsAllocation draws the three sleeves as continuous stacked bands
// over a withdrawal-rate axis. The two bucket bands widen linearly with the
// rate, so the equity band, drawn on top, is seen to be pushed down: it is
// read as a residual, which is the whole point of the article's first count.
func figBucketsAllocation() string {
	const (
		xL, xR     = 88.0, 452.0
		yTop, yBot = 100.0, 300.0
		gap        = 1.2 // surface gap between two stacked sleeves
	)
	x := func(rate float64) float64 {
		return xL + (rate-bucketsRateMin)/(bucketsRateMax-bucketsRateMin)*(xR-xL)
	}
	y := func(v float64) float64 { return yBot - v/100*(yBot-yTop) }
	quad := func(yTL, yTR, yBR, yBL float64, fill string) string {
		return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
			xL, yTL, xR, yTR, xR, yBR, xL, yBL, fill)
	}

	var b strings.Builder
	b.WriteString(plateHead("les buckets",
		"Le taux de retrait décide de l'allocation, pas les buckets"))
	b.WriteString(sTxt(24, 64, 10.5, figMuted, "start", "400",
		"recette canonique, 2 ans de dépenses en cash, 6 ans en obligations, le reste en actions"))

	// The three sleeves, at the two ends of the rate axis.
	cLo, bLo, _ := bucketsSplit(bucketsRateMin)
	cHi, bHi, _ := bucketsSplit(bucketsRateMax)
	b.WriteString(quad(y(cLo)+gap, y(cHi)+gap, y(0), y(0), figGreen))
	b.WriteString(quad(y(cLo+bLo)+gap, y(cHi+bHi)+gap, y(cHi), y(cLo), figBlue))
	b.WriteString(quad(y(100), y(100), y(cHi+bHi), y(cLo+bLo), figAccent))

	// What the eye should take away, written where the sleeve is widest.
	b.WriteString(sTxt(106, 132, 11.5, "#fffdf9", "start", "600",
		"la part actions n'est jamais choisie,"))
	b.WriteString(sTxt(106, 149, 11, "#fffdf9", "start", "400",
		"elle est le solde, 100 − 8 × le taux"))

	// The rate the book prices the recipe at, marked from the axis up through
	// the two buckets and named at the top of its own line.
	xr := x(bucketsRateRef)
	b.WriteString(dashLine(xr, yBot, xr, y(55), "#fffdf9", 1.4, "4 4"))
	b.WriteString(mTxt(xr+9, 172, 11, "#fffdf9", "start", "600", "3,5 %"))
	b.WriteString(sTxt(xr+9, 187, 10.5, "#fffdf9", "start", "400",
		"le taux de référence du livre"))

	// Axes: a quiet frame drawn over the bands, ticks outside.
	b.WriteString(line(xL, yTop, xL, yBot, figRule, 1))
	b.WriteString(line(xL, yBot, xR, yBot, figRule, 1))
	for _, g := range []float64{0, 25, 50, 75, 100} {
		b.WriteString(line(xL-4, y(g), xL, y(g), figRule, 1))
		b.WriteString(mTxt(xL-9, y(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(24, 88, 10, figMuted, "start", "400", "part du patrimoine (%)"))

	// Direct labels, each against its own sleeve at the 5 % end.
	type tag struct {
		mid    float64
		fill   string
		l1, l2 string
	}
	for _, t := range []tag{
		{(cHi + bHi + 100) / 2, figAccent, "actions", "le solde, jamais choisi"},
		{cHi + bHi/2, figBlue, "obligations", "6 ans de dépenses"},
		{cHi / 2, figGreen, "cash", "2 ans de dépenses"},
	} {
		my := y(t.mid)
		b.WriteString(line(xR+4, my, xR+11, my, t.fill, 2))
		b.WriteString(sTxt(xR+17, my-3, 11.5, t.fill, "start", "600", t.l1))
		b.WriteString(sTxt(xR+17, my+12, 10.5, figMuted, "start", "400", t.l2))
	}

	// The rate axis, and under three of its ticks the allocation that follows.
	for _, r := range []float64{2.5, 3, 3.5, 4, 4.5, 5} {
		lbl := strings.Replace(fmt.Sprintf("%.1f", r), ".", ",", 1)
		col, weight := figMuted, "400"
		if r == bucketsRateRef {
			col, weight = figDeep, "600"
		}
		b.WriteString(mTxt(x(r), 318, 10, col, "middle", weight, lbl))
	}
	for _, r := range []float64{bucketsRateMin, bucketsRateRef, bucketsRateMax} {
		col := figSoft
		if r == bucketsRateRef {
			col = figDeep
		}
		b.WriteString(mTxt(x(r), 340, 11, col, "middle", "600", bucketsReadout(r)))
	}
	b.WriteString(sTxt(320, 362, 11, figMuted, "middle", "400",
		"taux de retrait initial (%), et l'allocation qui en découle : cash / obligations / actions"))
	return svg(640, 384, b.String())
}
