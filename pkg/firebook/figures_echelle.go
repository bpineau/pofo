package firebook

import (
	"fmt"
	"strings"
)

// The bond-ladder plate. The article ends on a couple whose pension bridge is
// built rung by rung, and states that build as a list of costs; what the list
// hides is the shape of the thing. The liability is FLAT in today's euros, and
// what moves underneath it is the payer. This plate draws the need year by
// year and colours each year by who settles it.
//
// No market data is involved: everything is the arithmetic of the article's
// own example, held in one pure function (ladderYear) that the guard test
// confronts with the numbers the prose quotes.

// Pre-blended solid fills, never rgba (crengine, KOReader's EPUB SVG renderer,
// paints rgba solid black). The ladder's three contractual bricks share one
// blue family, light to dark, because they are one structure bought three
// ways: ladderCashFill is figBlue composited onto the figure card at .55,
// ladderTipsFill figBlue darkened toward the ink at .30. The pensions keep the
// book's green (guaranteed for life) and the comfort portfolio its amber, the
// one brick a market can take away.
const (
	ladderCashFill = "#8FAAD4"
	ladderBondFill = figBlue
	ladderTipsFill = "#2A4C7C"
	ladderPensFill = figGreen
	ladderRiskFill = figAccent
	ladderPaper    = "#fffdf9"
)

// The couple of the article's example (and of choisir-sa-strategie), in k EUR
// of today's money: a floor they will not go under, the comfort they aim at,
// and the pensions that arrive after thirteen years and cover ~53 % of the
// floor. The ladder matches the whole floor over the fragile window and 60 %
// of it afterwards, which is the coverage rule the plate's two braces name.
const (
	ladderFirstYear = 2026 // year 1: the article's rungs mature 2028 to 2033, its years 3 to 8
	ladderYears     = 14   // thirteen bridge years, plus the first year the pensions pay
	ladderFloor     = 45.0
	ladderComfort   = 58.0
	ladderPension   = 24.0
	ladderFullYears = 6  // years 1 to 6: 100 % of the floor matched
	ladderBridge    = 13 // the pensions arrive in year 14
	ladderPartial   = 0.60
	ladderEurosLast = 2 // the fonds euros pay years 1 and 2
	ladderBondLast  = 8 // the maturity funds pay years 3 to 8
)

// ladderPayer names who settles the contractual part of one year.
type ladderPayer int

const (
	payerEuros ladderPayer = iota // fonds euros, years 1-2
	payerBond                     // maturity funds, years 3-8
	payerTips                     // rolled short linkers, years 9-13
	payerPens                     // the pensions, year 14 on
)

// ladderYear returns, for the year of rank n (1-based), the amount some
// contract pays that year, who pays it, and what the comfort portfolio has
// left to serve, all in k EUR of today's money. The need itself never moves:
// the floor is a floor and the comfort above it is stable, so the whole story
// is in how the payers hand over.
func ladderYear(n int) (matched float64, payer ladderPayer, portfolio float64) {
	switch {
	case n > ladderBridge:
		matched, payer = ladderPension, payerPens
	case n > ladderFullYears:
		matched = ladderPartial * ladderFloor
		payer = payerBond
		if n > ladderBondLast {
			payer = payerTips
		}
	default:
		matched, payer = ladderFloor, payerEuros
		if n > ladderEurosLast {
			payer = payerBond
		}
	}
	return matched, payer, ladderComfort - matched
}

// ladderBridgeDraw is the average yearly draw the comfort portfolio serves
// over the thirteen bridge years, in k EUR. Divided by the 1,27 M EUR the
// couple keeps outside the ladder, it is the effective withdrawal rate the
// article quotes (~1,8 %), which is the arithmetic check on the whole plate.
func ladderBridgeDraw() float64 {
	sum := 0.0
	for n := 1; n <= ladderBridge; n++ {
		_, _, portfolio := ladderYear(n)
		sum += portfolio
	}
	return sum / ladderBridge
}

// figEchellePassif draws one stacked bar per year, 2026 to 2039: the flat
// annual need, cut into what a contract pays and what the portfolio still
// owes. Under the timeline each payer names the years it covers, two braces
// write the coverage rule over the years it applies to, and a rule between
// 2038 and 2039 marks the hand-over to the pensions.
func figEchellePassif() string {
	const (
		x0, x1 = 62.0, 556.0
		gap    = 4.0
		yBase  = 332.0
		yTop   = 126.0 // the comfort level, and the flat top of every bar
	)
	slot := (x1 - x0) / ladderYears
	barW := slot - gap
	left := func(i int) float64 { return x0 + slot*float64(i) + gap/2 }
	mid := func(i int) float64 { return left(i) + barW/2 }
	y := func(v float64) float64 { return yBase - v*(yBase-yTop)/ladderComfort }

	var b strings.Builder
	b.WriteString(plateHead("l'échelle obligataire",
		"Le passif année par année, et qui le paie"))
	b.WriteString(sTxt(24, 64, 10.5, figMuted, "start", "400",
		"Claire et Idris : 45 k€ de plancher, 58 k€ de confort, les pensions dans 13 ans"))

	// The bars: the matched brick at the base, the portfolio's residual on top,
	// each carrying its own amount so every year can be read on its own.
	fill := map[ladderPayer]string{
		payerEuros: ladderCashFill, payerBond: ladderBondFill,
		payerTips: ladderTipsFill, payerPens: ladderPensFill,
	}
	for i := 0; i < ladderYears; i++ {
		matched, payer, portfolio := ladderYear(i + 1)
		ink := ladderPaper
		if payer == payerEuros {
			ink = figInk // the lightest brick of the family takes dark numbers
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			left(i), y(matched), barW, yBase-y(matched), fill[payer])
		b.WriteString(barV(left(i), barW, y(matched)-1.5, y(ladderComfort), ladderRiskFill))
		b.WriteString(mTxt(mid(i), (y(matched)+yBase)/2+3.5, 10, ink, "middle", "600",
			fmt.Sprintf("%.0f", matched)))
		b.WriteString(mTxt(mid(i), (yTop+y(matched))/2+3.5, 10, ladderPaper, "middle", "600",
			fmt.Sprintf("%.0f", portfolio)))
	}

	// The two levels the article names, read on the left margin.
	b.WriteString(dashLine(x0, y(ladderFloor), x1, y(ladderFloor), figInk, 1, "3 3"))
	b.WriteString(line(x0, yBase, x1, yBase, figRule, 1))
	for _, l := range []struct {
		v          float64
		num, label string
	}{
		{ladderComfort, "58 k€", "confort"},
		{ladderFloor, "45 k€", "plancher"},
	} {
		b.WriteString(mTxt(x0-8, y(l.v)+3.5, 10, figInk, "end", "600", l.num))
		b.WriteString(sTxt(x0-8, y(l.v)+16, 9.5, figMuted, "end", "400", l.label))
	}
	b.WriteString(mTxt(x0-8, yBase+3.5, 10, figMuted, "end", "400", "0"))

	// The coverage rule, braced over the years it applies to.
	for _, br := range []struct {
		from, to int
		label    string
	}{
		{0, ladderFullYears - 1, "100 % du plancher adossé"},
		{ladderFullYears, ladderBridge - 1, "60 % du plancher adossé"},
	} {
		xa, xb := left(br.from), left(br.to)+barW
		b.WriteString(line(xa, 106, xb, 106, figMuted, 1.2))
		b.WriteString(line(xa, 106, xa, 111, figMuted, 1.2))
		b.WriteString(line(xb, 106, xb, 111, figMuted, 1.2))
		b.WriteString(sTxt((xa+xb)/2, 98, 10, figSoft, "middle", "600", br.label))
	}

	// The hand-over: the pensions arrive, and the ladder stops.
	xr := x0 + slot*float64(ladderBridge)
	b.WriteString(dashLine(xr, 112, xr, 356, ladderPensFill, 1.4, "4 4"))

	// The two stacked layers, named once in the right margin, against the last
	// bar: the residual the portfolio always carries, and the pension under it.
	pens, _, _ := ladderYear(ladderYears)
	b.WriteString(sTxt(x1+9, (yTop+y(pens))/2-3, 10.5, ladderRiskFill, "start", "600", "portefeuille"))
	b.WriteString(sTxt(x1+9, (yTop+y(pens))/2+11, 10.5, ladderRiskFill, "start", "600", "de confort"))
	b.WriteString(sTxt(x1+9, (yTop+y(pens))/2+25, 9.5, figMuted, "start", "400", "non garanti"))
	b.WriteString(mTxt(x1+9, (yBase+y(pens))/2+3.5, 9.5, ladderPensFill, "start", "600", "24 k€/an"))

	// The timeline, and under it the payer that settles each stretch of years.
	for i := 0; i < ladderYears; i++ {
		b.WriteString(mTxt(mid(i), 350, 9.5, figMuted, "middle", "400",
			fmt.Sprintf("%d", ladderFirstYear+i)))
	}
	for _, g := range []struct {
		from, to  int
		fill      string
		name, sub string
	}{
		{0, ladderEurosLast - 1, ladderCashFill, "fonds euros", "garantis, liquides"},
		{ladderEurosLast, ladderBondLast - 1, ladderBondFill, "fonds à échéance", "État ou IG, 2028-2033"},
		{ladderBondLast, ladderBridge - 1, ladderTipsFill, "ETF linkers courts roulés", "indexés sur les prix"},
		{ladderBridge, ladderYears - 1, ladderPensFill, "les pensions", "le relais"},
	} {
		xa, xb := left(g.from), left(g.to)+barW
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="3" rx="1.5" fill="%s"/>`,
			xa, 364.0, xb-xa, g.fill)
		b.WriteString(sTxt((xa+xb)/2, 382, 10.5, g.fill, "middle", "600", g.name))
		b.WriteString(sTxt((xa+xb)/2, 395, 9.5, figMuted, "middle", "400", g.sub))
	}

	b.WriteString(sTxt(24, 419, 9.5, figMuted, "start", "400",
		"Besoins et barreaux en euros d'aujourd'hui ; à l'achat, les barreaux nominaux sont gonflés de 2,5 %/an et les linkers sont indexés."))
	return svg(640, 436, b.String())
}
