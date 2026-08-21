package firebook

import (
	"fmt"
	"strings"
)

// The gold A/B plate. The article's closing example runs the same plan twice,
// once with a gold sleeve and once without, and reports the answer as a list
// of paired numbers. Read as a list, the pairs look interchangeable; drawn as
// dumbbells on one ruin axis, they say the thing the prose is really about:
// the two plans are the same plan in the central case, and they come apart
// exactly as the model turns hostile.
//
// The fourth line of the example is not a probability at all: replaying the
// 1966 vintage gives a verdict, "exhausted in year 27" against "ridden out".
// Forcing it onto the axis would invent two numbers, so it lives in a band of
// its own under the axis, typeset rather than plotted.
//
// Every number is the article's own, and no market data is involved. The
// guard test re-reads them from the French article's prose.

// goldABRow is one model, with the ruin probability of the two variants: A is
// 70 % equities / 30 % intermediate bonds, B is 70 / 20 / 10 with gold.
type goldABRow struct {
	label   string
	without float64 // A, no gold
	with    float64 // B, 10 % gold
}

// The three modelled rows of the article's example, in percent of ruin, from
// the friendliest model to the most hostile. The gap between the two variants
// is what the plate exists to show: near nothing in the central case, a full
// point under sequence stress, two under a long inflation.
var goldABRows = []goldABRow{
	{"scénario central", 4.1, 3.9},
	{"stress de séquence", 7.2, 6.1},
	{"modèles à inflation longue", 10.8, 8.9},
}

// The vintage replay, which is a verdict and not a probability, and the price
// of the cover, which the plate prints so it never reads as a free lunch.
const (
	goldABVintage     = "millésime 1966 rejoué"
	goldABVerdictA    = "épuisé à l'année 27"
	goldABVerdictB    = "traversé, amoché"
	goldABMedianCost  = -5.0 // median wealth at 45 years, in percent
	goldABAxisMax     = 12.0
	goldABPlotX0      = 200.0
	goldABPlotX1      = 528.0
	goldABRowSpacing  = 48.0
	goldABFirstRowY   = 124.0
	goldABWithoutFill = figBlue
	goldABWithFill    = figAccent
)

// goldABGain is the ruin probability the gold sleeve removes on one row, in
// points. It is the length of that row's dumbbell, and the only quantity the
// plate derives rather than quotes.
func goldABGain(r goldABRow) float64 { return r.without - r.with }

// figOrAbModeles draws the three modelled rows as dumbbells on a single ruin
// axis, then the vintage verdict in a band of its own, then the price.
func figOrAbModeles() string {
	x := func(v float64) float64 {
		return goldABPlotX0 + v/goldABAxisMax*(goldABPlotX1-goldABPlotX0)
	}
	rowY := func(i int) float64 { return goldABFirstRowY + goldABRowSpacing*float64(i) }

	var b strings.Builder
	b.WriteString(plateHead("l'or en retrait", "Avec ou sans or, modèle par modèle"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Le même plan joué deux fois : 1,5 M€, 51 000 €/an, 45 ans, corridor Vanguard"))

	// The axis: a hairline per even point, and the ticks under the last row.
	yTop, yBot := 104.0, 240.0
	for v := 0.0; v <= goldABAxisMax; v += 2 {
		col := figGrid
		if v == 0 {
			col = figRule
		}
		b.WriteString(line(x(v), yTop, x(v), yBot, col, 1))
		b.WriteString(mTxt(x(v), 260, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f %%", v)))
	}
	b.WriteString(line(goldABPlotX0, 244, goldABPlotX1, 244, figRule, 1))
	b.WriteString(sTxt((goldABPlotX0+goldABPlotX1)/2, 280, 10.5, figMuted, "middle", "400",
		"probabilité de ruine  →"))

	// The column that reads the dumbbells as a number.
	b.WriteString(sTxt(616, 104, 10, figSoft, "end", "600", "l'or retire"))
	b.WriteString(sTxt(616, 117, 9.5, figMuted, "end", "400", "(points de ruine)"))

	// The three rows. The bar between the two dots is the whole message, so it
	// is drawn first and the dots sit on top of it.
	for i, r := range goldABRows {
		y := rowY(i)
		b.WriteString(sTxt(24, y+4, 10.5, figSoft, "start", "600", r.label))
		b.WriteString(line(x(r.with), y, x(r.without), y, figRule, 5))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5.5" fill="%s"/>`, x(r.without), y, goldABWithoutFill)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5.5" fill="%s"/>`, x(r.with), y, goldABWithFill)
		b.WriteString(mTxt(x(r.with)-10, y+3.5, 10.5, figDeep, "end", "600", frNum(r.with, 1)))
		b.WriteString(mTxt(x(r.without)+10, y+3.5, 10.5, figBlue, "start", "600", frNum(r.without, 1)))
		b.WriteString(mTxt(616, y+3.5, 11, figInk, "end", "600", "−"+frNum(goldABGain(r), 1)))
	}

	// The two variants, named once, over the row where they stand furthest
	// apart: colour carries the mapping everywhere else.
	last := goldABRows[len(goldABRows)-1]
	yLast := rowY(len(goldABRows) - 1)
	b.WriteString(sTxt(x(last.with), yLast-14, 10.5, goldABWithFill, "middle", "600", "avec or"))
	b.WriteString(sTxt(x(last.without), yLast-14, 10.5, goldABWithoutFill, "middle", "600", "sans or"))

	// The vintage row: a verdict, in its own band, under the axis and off it.
	fmt.Fprintf(&b, `<rect x="24" y="302" width="592" height="50" rx="4" fill="%s"/>`, figWash)
	b.WriteString(sTxt(38, 322, 10.5, figInk, "start", "600", goldABVintage))
	b.WriteString(sTxt(38, 336, 9.5, figMuted, "start", "400", "un verdict, pas une probabilité"))
	b.WriteString(sTxt(210, 318, 9.5, goldABWithoutFill, "start", "600", "sans or"))
	b.WriteString(sTxt(210, 336, 11.5, figBad, "start", "600", goldABVerdictA))
	b.WriteString(sTxt(394, 331, 13, figMuted, "middle", "400", "→"))
	b.WriteString(sTxt(420, 318, 9.5, goldABWithFill, "start", "600", "avec or"))
	b.WriteString(sTxt(420, 336, 11.5, goldABWithFill, "start", "600", goldABVerdictB))

	// The price of the cover, which the plate is not allowed to leave out.
	b.WriteString(sTxt(24, 376, 10.5, figSoft, "start", "600", fmt.Sprintf(
		"Le prix de cette couverture : richesse médiane à 45 ans, \u2212%s %%.", frNum(-goldABMedianCost, 0))))
	b.WriteString(sTxt(24, 392, 9.5, figMuted, "start", "400",
		"A = 70 % actions / 30 % obligations intermédiaires. B = 70 / 20 / 10 avec or, ETC physique en CTO."))
	b.WriteString(sTxt(24, 406, 9.5, figMuted, "start", "400",
		"Mêmes hypothèses des deux côtés ; seule l'allocation change. Le millésime 2000, lui, ne bouge quasiment pas."))
	return svg(640, 422, b.String())
}
