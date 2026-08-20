package firebook

import (
	"strings"
)

// The rental cascade of immobilier-en-retrait: from the gross yield an advert
// shows to the net-net a landlord banks, one deduction at a time, and the fork
// the tax regime opens at the bottom.
//
// Every hypothesis below is stated in the article's own "Le chiffre honnête"
// paragraph, and figures_immobilier_test.go fails the moment one of them
// leaves the prose or the cascade stops landing inside the 2-4 % the article
// announces. The two rates are law rather than estimate: bare rental income
// (foncier nu) keeps its social levies at 17.2 %, being one of the incomes the
// LFSS 2026 (art. 12) left out of the rise to 18.6 %, and it is taxed at the
// income-tax barème, where the 6.8 deductible points of CSG give immoTMI of
// themselves back the following year. Furnished letting (LMNP) took the rise.
const (
	immoPrice        = 200000.0 // the flat, and the base of every yield drawn here
	immoRentMonthly  = 1000.0   // rent, excluding the recoverable charges
	immoCoproPNO     = 1150.0   // non-recoverable service charges + landlord insurance
	immoTaxeFonciere = 900.0    // yearly property tax
	immoEntretien    = 600.0    // running maintenance
	immoVacancyRate  = 0.05     // one month empty every twenty
	immoGestionRate  = 0.07     // letting agent, as a share of the rent
	immoImpayesRate  = 0.025    // rent-guarantee insurance, same base
	immoComptes      = 350.0    // yearly bookkeeping and filing
	immoTMI          = 0.30     // the article's marginal income-tax bracket
	immoPSFoncierNu  = 0.172    // social levies on bare rental income
	immoCSGDeduct    = 0.068    // deductible share of the CSG, at the barème
)

func immoGross() float64   { return 12 * immoRentMonthly }
func immoVacancy() float64 { return immoVacancyRate * immoGross() }

// immoGestion is the cost of not doing the work oneself: the agent, the
// unpaid-rent guarantee and the yearly filing.
func immoGestion() float64 {
	return (immoGestionRate+immoImpayesRate)*immoGross() + immoComptes
}

// immoPreTax is the rent left once every operating line is paid, which is also
// the taxable base of the bare-letting réel regime.
func immoPreTax() float64 {
	return immoGross() - immoCoproPNO - immoTaxeFonciere - immoEntretien -
		immoVacancy() - immoGestion()
}

// immoTaxRateNu is the effective bite on that base: the marginal bracket plus
// the social levies, less what the deductible CSG hands back.
func immoTaxRateNu() float64 {
	return immoTMI + immoPSFoncierNu - immoCSGDeduct*immoTMI
}

func immoNetNu() float64 { return immoPreTax() * (1 - immoTaxRateNu()) }

// immoYield turns a yearly amount into a yield on the price, in percent.
func immoYield(v float64) float64 { return v / immoPrice * 100 }

// figImmobilierNetNet draws the cascade as one descending staircase that forks
// in two at the tax step: bare letting on one side, LMNP au réel on the other.
func figImmobilierNetNet() string {
	m := mapper(0, 1, 0, 6.4, 0, 1, 330, 76)
	y := func(v float64) float64 { return m(0, v)[1] }
	const slotW, bw = 70.0, 50.0
	cx := func(i int) float64 { return 56 + slotW*(float64(i)+0.5) }

	var b strings.Builder
	b.WriteString(plateHead("l'immobilier locatif", "Du brut d'annonce au net-net : la cascade"))
	b.WriteString(plateDeck(
		"un logement de 200 000 € loué 1 000 € par mois, sans crédit, TMI 30 % : chaque marche se remplace par la vôtre"))

	for _, g := range []float64{0, 2, 4, 6} {
		gy := y(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(56, gy, 616, gy, col, 1))
		b.WriteString(mTxt(48, gy+3.5, 10, figMuted, "end", "400", frNum(g, 0)+" %"))
	}

	// The caption block under each column. Levels carry the yearly euro flow on
	// a third line; the deductions carry their cost above their own bar, so the
	// euros always add up on the page and a reader can substitute their own.
	label := func(i int, l1, l2, l3, col string) {
		b.WriteString(sTxt(cx(i), 348, 10.5, col, "middle", "600", l1))
		b.WriteString(sTxt(cx(i), 361, 10, figMuted, "middle", "400", l2))
		if l3 != "" {
			b.WriteString(mTxt(cx(i), 374, 9.5, figSoft, "middle", "400", l3))
		}
	}

	// Column 0: what the advert shows.
	lvl := immoYield(immoGross())
	b.WriteString(barV(cx(0)-bw/2, bw, y(0), y(lvl), figAccent))
	b.WriteString(mTxt(cx(0), y(lvl)-8, 11.5, figInk, "middle", "600", frNum(lvl, 1)+" %"))
	label(0, "brut", "d'annonce", euroFR(immoGross()), figSoft)

	// Columns 1 to 5: the deductions, in the order the article lists them.
	cuts := []struct {
		amount float64
		l1, l2 string
	}{
		{immoCoproPNO, "charges non", "récupérables"},
		{immoTaxeFonciere, "taxe", "foncière"},
		{immoEntretien, "entretien", "courant"},
		{immoVacancy(), "vacance", "(5 %)"},
		{immoGestion(), "gestion", "et impayés"},
	}
	for i, c := range cuts {
		top, col := lvl, i+1
		lvl -= immoYield(c.amount)
		x := cx(col) - bw/2
		b.WriteString(dashLine(cx(col-1)+bw/2, y(top), x, y(top), figMuted, 1, "3 3"))
		b.WriteString(barV(x, bw, y(top), y(lvl), figBad))
		b.WriteString(mTxt(cx(col), y(top)-8, 11, figBad, "middle", "600",
			"−"+euroFR(c.amount)))
		label(col, c.l1, c.l2, "", figSoft)
	}

	// The level the two regimes start from, called out between the two forks.
	pre := lvl
	b.WriteString(dashLine(cx(5)+bw/2, y(pre), 616, y(pre), figMuted, 1, "3 3"))
	b.WriteString(sTxt(616, 122, 10.5, figMuted, "end", "400", "loyer net avant impôt"))
	b.WriteString(mTxt(616, 136, 10.5, figSoft, "end", "600",
		frNum(pre, 1)+" % = "+euroFR(immoPreTax())))

	// Column 6: bare letting, where the tax step is the tallest of them all.
	nu := immoYield(immoNetNu())
	x6 := cx(6) - bw/2
	b.WriteString(barV(x6, bw, y(nu)-4, y(pre), figBad))
	b.WriteString(mTxt(cx(6), y(pre)-8, 11, figBad, "middle", "600",
		"−"+euroFR(immoPreTax()-immoNetNu())))
	b.WriteString(barV(x6, bw, y(0), y(nu), figDeep))
	b.WriteString(mTxt(cx(6), y(nu)+17, 11.5, "#fffdf9", "middle", "600", frNum(nu, 1)+" %"))
	label(6, "net-net", "foncier nu", euroFR(immoNetNu()), figDeep)

	// Column 7: the same flat, furnished at the réel regime, where the
	// amortisation absorbs the base and the tax step disappears.
	b.WriteString(barV(cx(7)-bw/2, bw, y(0), y(pre), figGreen))
	b.WriteString(sTxt(cx(7), y(pre)-8, 10.5, figGreen, "middle", "600", "aucun impôt"))
	b.WriteString(mTxt(cx(7), y(pre)+17, 11.5, "#fffdf9", "middle", "600", frNum(pre, 1)+" %"))
	label(7, "net-net", "LMNP au réel", euroFR(immoPreTax()), figGreen)

	b.WriteString(sTxt(24, 396, 9, figMuted, "start", "400",
		"Impôt du foncier nu : TMI 30 % plus 17,2 % de prélèvements sociaux, CSG déductible déduite, soit 45,2 % du net imposable."))
	b.WriteString(sTxt(24, 408, 9, figMuted, "start", "400",
		"En meublé, les prélèvements sociaux passent à 18,6 % (LFSS 2026), mais l'amortissement efface la base tant qu'il dure."))
	return svg(640, 420, b.String())
}
