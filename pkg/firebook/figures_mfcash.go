package firebook

import (
	"fmt"
	"strings"
)

// The "cash works too" plate. The article's arithmetic is one line of prose,
// and it explains both the disappointment of the trend winter and the recent
// glow: a futures program returns cash plus the trend premium minus fees, so
// the SAME programme, unchanged, prints 2 % when short rates are zero and 5 to
// 6 % when they are 3 to 4 %. Drawn as two signed decompositions side by side,
// the eye sees that only one of the three bricks moved.
//
// No market data is involved. The three numbers of each regime, and the dates
// that name them, are the article's own; the plate holds them in one table and
// the guard test recomputes every net from it. Deliberately NOT a cascade: the
// two columns must stay strictly comparable brick by brick, which a running
// total would hide.

// mfRegime is one short-rate regime: what the collateral earns, what the
// programme earns gross, and what it charges, all in percent a year.
type mfRegime struct {
	name, dates string
	cash        float64
	gross       float64
	fees        float64
}

// mfNet is what the holder actually sees: cash plus premium minus fees.
func (r mfRegime) mfNet() float64 { return r.cash + r.gross - r.fees }

// The two regimes of the article, in its own numbers: a 3 % gross premium and
// 1 % of fees throughout, with only the collateral changing. The second regime
// quotes a RANGE of short rates, 3 to 4 %, so the plate draws its midpoint and
// the footnote carries the range.
var (
	mfZIRP = mfRegime{"taux courts à 0 %", "2015-2021", 0, 3, 1}
	mfHigh = mfRegime{"taux courts à 3-4 %", "2022 →", 3.5, 3, 1}
)

// mfRateLo and mfRateHi are the ends of the article's range, which produce the
// 5 to 6 % it quotes.
const (
	mfRateLo = 3.0
	mfRateHi = 4.0
)

// figMfCashPrimeFrais draws the two decompositions and their nets.
func figMfCashPrimeFrais() string {
	const (
		gridX0, gridX1 = 84.0, 560.0
		yTop, yBot     = 110.0, 330.0
		vLo, vHi       = -1.5, 7.0
		stackW, netW   = 110.0, 70.0
	)
	rate := figScale{Min: vLo, Max: vHi, Px0: yBot, Px1: yTop}
	y := rate.Map
	y0 := y(0)

	var b strings.Builder
	b.WriteString(plateHead("les managed futures",
		"Le même programme, deux régimes de taux courts"))
	b.WriteString(plateDeck(
		"Un fonds trend rapporte le cash, plus la prime du trend, moins les frais"))
	legendChips(&b, 76, [][2]string{
		{figBlue, "collatéral rémunéré"},
		{figAccent, "prime de trend brute"},
		{figBad, "frais"},
	})

	// The value axis, zero picked out from the rest by the kit.
	axisTicks(&b, rate, []float64{-1, 0, 1, 2, 3, 4, 5, 6, 7}, 0, " %", gridX0, gridX1, false)

	// One group per regime: the signed stack, then the net beside it.
	for i, r := range []mfRegime{mfZIRP, mfHigh} {
		cx := 200 + 240*float64(i)
		sx, nx := cx-100, cx+26

		// The collateral, then the gross premium on top of it. A regime that
		// pays nothing on collateral gets a hairline and its zero in writing,
		// because the absence IS the point of the left column.
		if r.cash > 0 {
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
				sx, y(r.cash), stackW, y0-y(r.cash), figBlue)
			b.WriteString(mTxt(sx+stackW/2, (y0+y(r.cash))/2+4, 11, "#fffdf9", "middle", "600",
				"+"+frNum(r.cash, 1)))
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			sx, y(r.cash+r.gross), stackW, y(r.cash)-y(r.cash+r.gross), figAccent)
		b.WriteString(mTxt(sx+stackW/2, (y(r.cash)+y(r.cash+r.gross))/2+4, 11, "#fffdf9", "middle", "600",
			"+"+frNum(r.gross, 1)))

		// A regime that pays nothing on collateral keeps the brick, flat on
		// zero: the absence is the point of the left column, and it is drawn
		// last so the premium above it cannot bury it.
		if r.cash == 0 {
			b.WriteString(line(sx, y0, sx+stackW, y0, figBlue, 2.6))
		}

		// The fees, hanging under zero, which is what makes this a signed
		// decomposition rather than a stack of good news.
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			sx, y0, stackW, y(-r.fees)-y0, figBad)
		b.WriteString(mTxt(sx+stackW/2, (y0+y(-r.fees))/2+4, 11, "#fffdf9", "middle", "600",
			"−"+frNum(r.fees, 1)))

		// The net, beside the bricks it comes from.
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			nx, y(r.mfNet()), netW, y0-y(r.mfNet()), figDeep)
		b.WriteString(sTxt(nx+netW/2, y(r.mfNet())-24, 10, figMuted, "middle", "400", "net"))
		b.WriteString(mTxt(nx+netW/2, y(r.mfNet())-9, 12.5, figDeep, "middle", "600",
			frNum(r.mfNet(), 1)+" %"))

		// The regime, named under its group.
		b.WriteString(sTxt(cx-16, 356, 11, figInk, "middle", "600", r.name))
		b.WriteString(sTxt(cx-16, 370, 9.5, figMuted, "middle", "400", r.dates))
	}

	b.WriteString(plateConclusion(396,
		"Rien n'a changé dans le programme : seule la rémunération du collatéral est passée de 0 à 3-4 %."))
	b.WriteString(sTxt(24, 412, 9.5, figMuted, "start", "400", fmt.Sprintf(
		"Prime brute %s %% et frais %s %% dans les deux régimes. Aux deux bouts de la fourchette de taux, le net vaut %s à %s %%.",
		frNum(mfZIRP.gross, 0), frNum(mfZIRP.fees, 0),
		frNum(mfRateLo+mfZIRP.gross-mfZIRP.fees, 0), frNum(mfRateHi+mfZIRP.gross-mfZIRP.fees, 0))))
	b.WriteString(sTxt(24, 426, 9.5, figMuted, "start", "400",
		"Les dates nomment les régimes comme l'article les date ; aucune série de marché n'entre dans la planche."))
	return svg(640, 442, b.String())
}
