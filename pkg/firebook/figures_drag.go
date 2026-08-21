package firebook

import (
	"fmt"
	"strings"
)

// The volatility-drag plate. The article states the drag as a table, one row
// per asset, and a table cannot show that the cost is QUADRATIC: read down the
// column, 10 % of volatility costs half a point and 30 % costs four and a
// half, and the eye registers "more" where the truth is "nine times more".
// Drawn as the parabola it is, with the article's own assets posed on it, the
// non-linearity is the picture.
//
// The whole plate is one pure function, dragGeometric, applied to the
// volatilities the article's table quotes; the guard test recomputes every
// posed point and both arrows from it.

// dragArith is the arithmetic mean every row of the article's table starts
// from, in percent: holding it constant is what isolates the volatility.
const dragArith = 7.0

// dragGeometric is the compounded return an asset delivers when its
// arithmetic mean is dragArith and its annualized volatility is sigma, both in
// percent: mu - sigma^2 / 2, the drag being the second term. It is the plate's
// only formula.
func dragGeometric(sigma float64) float64 { return dragArith - dragDrag(sigma) }

// dragDrag is the volatility drag itself, in points per year.
func dragDrag(sigma float64) float64 { return sigma * sigma / 200 }

// dragAsset is one row of the article's table: a name, the volatility it
// quotes, and where the label sits when two assets land on the same point.
type dragAsset struct {
	sigma float64
	names []string
}

// The six assets of the article's table, in its order. Global equities and the
// stacked 90/60 share a volatility, which is the comparison the article draws
// its conclusion from, so they share a point and are named together on it.
var dragAssets = []dragAsset{
	{1, []string{"fonds monétaire"}},
	{10, []string{"portefeuille 60/40"}},
	{15, []string{"actions mondiales", "empilement 90/60"}},
	{22, []string{"actions émergentes"}},
	{30, []string{"levier ×2 quotidien"}},
}

// The two moves the article's prose describes, both of them moves along the
// volatility axis: diversifying walks left with the arithmetic mean untouched,
// daily leverage walks right and pays the square.
const (
	dragDiversifyFrom = 22.0
	dragDiversifyTo   = 10.0
	dragLeverFrom     = 15.0
	dragLeverTo       = 30.0
)

// dragMove is what a move from one volatility to another does to the
// compounded return, in points: positive when the move pays.
func dragMove(from, to float64) float64 { return dragGeometric(to) - dragGeometric(from) }

// figDragVolatilite draws the parabola, the article's assets on it, and the
// two moves under the axis.
func figDragVolatilite() string {
	const (
		x0, x1     = 70.0, 500.0
		sigMax     = 31.0
		yTop, yBot = 100.0, 290.0
		vLo, vHi   = 2.0, 7.4
		arrowA     = 352.0 // the diversifying move
		arrowB     = 378.0 // the leverage move
	)
	x := func(s float64) float64 { return x0 + s/sigMax*(x1-x0) }
	rate := figScale{Min: vLo, Max: vHi, Px0: yBot, Px1: yTop}
	y := rate.Map

	var b strings.Builder
	b.WriteString(plateHead("arithmétique et géométrique",
		"Doubler la volatilité quadruple ce qu'elle coûte"))
	b.WriteString(plateDeck(
		"Le rendement réellement composé quand la moyenne arithmétique reste 7 % et que seule la volatilité change"))

	// Grid and axes.
	axisTicks(&b, rate, []float64{2, 3, 4, 5, 6, 7}, 0, " %", x0, x1, false)
	b.WriteString(line(x0, yTop, x0, yBot, figRule, 1))
	b.WriteString(line(x0, yBot, x1, yBot, figRule, 1))
	for s := 0.0; s <= 30; s += 5 {
		b.WriteString(mTxt(x(s), 308, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", s)))
	}
	b.WriteString(sTxt(x0, 330, 10.5, figMuted, "start", "400",
		"volatilité annualisée σ (%)  →"))

	// The parabola.
	var pts [][2]float64
	for s := 0.0; s <= sigMax+0.01; s += 0.25 {
		pts = append(pts, [2]float64{x(s), y(dragGeometric(s))})
	}
	b.WriteString(poly(pts, figAccent, 2.8, ""))

	// The article's assets, each dropped onto the axis and named above right,
	// where the parabola leaves the room free.
	for _, a := range dragAssets {
		cx, cy := x(a.sigma), y(dragGeometric(a.sigma))
		b.WriteString(line(cx, cy, cx, yBot, figGrid, 1))
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, cx, cy, figDeep)
		ly := cy - 9 - 13*float64(len(a.names))
		for _, n := range a.names {
			b.WriteString(sTxt(cx+8, ly, 10, figSoft, "start", "600", n))
			ly += 13
		}
		b.WriteString(mTxt(cx+8, ly, 10.5, figDeep, "start", "600",
			frNum(dragGeometric(a.sigma), 1)+" %"))
	}

	// The two moves, under the axis, where they can be read as what they are:
	// steps along the volatility axis, one paid and one paying.
	arrow := func(y, from, to float64, col string) {
		xa, xb := x(from), x(to)
		b.WriteString(line(xa, y, xb, y, col, 1.8))
		d := 7.0
		if xb < xa {
			d = -7
		}
		fmt.Fprintf(&b, `<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
			xb, y, xb-d, y-4, xb-d, y+4, col)
	}
	arrow(arrowA, dragDiversifyFrom, dragDiversifyTo, figGreen)
	b.WriteString(sTxt(x(dragDiversifyFrom)+10, arrowA+4, 10, figGreen, "start", "600",
		fmt.Sprintf("diversifier : moins de σ, %s point de géométrique",
			"+"+frNum(dragMove(dragDiversifyFrom, dragDiversifyTo), 1))))
	arrow(arrowB, dragLeverFrom, dragLeverTo, figBad)
	b.WriteString(sTxt(x(dragLeverFrom)-10, arrowB+4, 10, figBad, "end", "600",
		fmt.Sprintf("levier quotidien : deux fois le σ, %s points",
			"−"+frNum(-dragMove(dragLeverFrom, dragLeverTo), 1))))

	b.WriteString(plateConclusion(402, fmt.Sprintf(
		"Le drag est le carré de la volatilité : %s point à 10 %% de σ, %s points à 30 %%, pour la même moyenne annoncée.",
		frNum(dragDrag(10), 1), frNum(dragDrag(30), 1))))
	b.WriteString(plateFoot(418, []string{
		"Moyenne arithmétique tenue à 7 %/an, drag = σ² / 2. Les volatilités sont celles du tableau de l'article, hors frais de levier.",
	}))
	return svg(640, 434, b.String())
}
