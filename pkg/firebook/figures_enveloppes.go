package firebook

import (
	"fmt"
	"math"
	"strings"
)

// This file holds the plate of the enveloppes-francaises article: the four
// withdrawal frictions of the French envelopes, read against the gain fraction
// of the withdrawal (that is, the age of the portfolio). Everything it draws
// comes from the dated rate constants and the closed formulas below, so nothing
// is frozen here: the plate is recomputed at render time and the guard test
// checks the same formulas against every friction the article quotes in prose.

// The rates in force on 1 January 2026, each one verified against the tax
// reference set rather than quoted from memory:
//
//   - LFSS 2026 (loi n° 2025-1403 du 30 décembre 2025, art. 12) raises the CSG
//     on capital income from 9,2 to 10,6 points, so the social levies go from
//     17,2 to 18,6 %. It covers the "revenus du patrimoine" of art. L. 136-6
//     CSS (a CTO's capital gains, from the 2025 income year) and the "produits
//     de placement" of art. L. 136-7 CSS (a PEA's gains cashed from 01/01/2026).
//   - Assurance-vie and capitalisation contracts were explicitly excluded from
//     that rise and stay at 17,2 %.
//   - Art. 125-0 A CGI: after 8 years, an assurance-vie pays 7,5 % of income tax
//     on premiums up to 150 000 €, after a yearly allowance of 4 600 € (single)
//     or 9 200 € (couple). The allowance is an INCOME-TAX allowance only, and
//     that is the whole point of the fourth curve: social levies are due on the
//     entire taxable gain of the redemption, allowance or not.
const (
	socialLevies2026   = 0.186 // PEA and CTO, after the LFSS 2026 rise
	socialLeviesAV2026 = 0.172 // assurance-vie and capitalisation, excluded from it
	pfuIncomeTax       = 0.128 // the income-tax leg of the PFU
	avIncomeTax8Y      = 0.075 // an assurance-vie older than 8 years
	avAllowanceCouple  = 9200.0
	avAllowanceSingle  = 4600.0
)

// The three linear frictions, in tax paid per euro of gain contained in the
// withdrawal. A PEA older than 5 years owes no income tax at all, which is the
// entire distance between it and a CTO.
const (
	ctoSlope = pfuIncomeTax + socialLevies2026    // 31,4 %, the PFU
	peaSlope = socialLevies2026                   // 18,6 %, social levies only
	avSlope  = avIncomeTax8Y + socialLeviesAV2026 // 24,7 %, outside the allowance
)

// avRedemption is the yearly redemption the fourth curve is drawn for: the
// couple of the article, at 25 000 € a year.
const avRedemption = 25000.0

// avFriction returns the tax owed on a redemption from a mature (8 years)
// assurance-vie whose gain fraction is g, as a fraction of the amount
// withdrawn. Because the allowance spares the income-tax leg only, the curve
// has one kink and one kink only: below it the friction is the bare social
// levies, above it the full slope, shifted down by what the allowance saves.
func avFriction(g, redemption, allowance float64) float64 {
	gain := g * redemption
	taxed := math.Max(0, gain-allowance)
	return (socialLeviesAV2026*gain + avIncomeTax8Y*taxed) / redemption
}

// avAllowanceExhausted is the gain fraction at which the allowance runs out,
// the abscissa of the kink: 9 200 € of gain inside a 25 000 € redemption.
func avAllowanceExhausted(redemption, allowance float64) float64 {
	return allowance / redemption
}

// pctFR formats a rate as a French percentage, one decimal.
func pctFR(v float64) string {
	return strings.Replace(fmt.Sprintf("%.1f %%", v*100), ".", ",", 1)
}

// figFrictionsEnveloppes draws the four frictions as a fan: the ranking of the
// French envelopes is nothing but a ranking of slopes, which is why a young CTO
// (little gain inside each sale) costs less than a mature assurance-vie.
func figFrictionsEnveloppes() string {
	const (
		x0, x1 = 74.0, 400.0
		yBase  = 300.0 // 0 %
		yTop   = 84.0  // 32 %
		vMax   = 0.32
	)
	px := func(g float64) float64 { return x0 + g*(x1-x0) }
	py := func(f float64) float64 { return yBase - f/vMax*(yBase-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("pea, assurance-vie, cto",
		"Le classement des enveloppes tient dans une pente"))
	b.WriteString(sTxt(x0, 68, 10.5, figMuted, "start", "400",
		"impôt payé, en % du flux retiré (et non du gain)"))

	// horizontal grid, the zero line doubling as the x axis
	for _, g := range []float64{0, 0.10, 0.20, 0.30} {
		gy := py(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(x0, gy, x1+8, gy, col, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f %%", g*100)))
	}
	b.WriteString(line(x0, yTop, x0, yBase, figRule, 1))

	// the kink of the allowance curve, and the guide that carries the eye from
	// the annotation down to the x axis
	gStar := avAllowanceExhausted(avRedemption, avAllowanceCouple)
	kx, ky := px(gStar), py(avFriction(gStar, avRedemption, avAllowanceCouple))
	b.WriteString(dashLine(kx, 136, kx, yBase, figMuted, 1, "2 4"))

	// the four traces, straight lines but for the fourth
	straight := []struct {
		slope float64
		col   string
	}{{ctoSlope, figBad}, {avSlope, figBlue}, {peaSlope, figAccent}}
	for _, s := range straight {
		b.WriteString(poly([][2]float64{{px(0), py(0)}, {px(1), py(s.slope)}}, s.col, 2.2, ""))
	}
	b.WriteString(poly([][2]float64{
		{px(0), py(0)},
		{kx, ky},
		{px(1), py(avFriction(1, avRedemption, avAllowanceCouple))},
	}, figGreen, 2.4, ""))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`,
		kx, ky, figGreen)

	// direct end labels, the value column aligned, no legend
	for _, l := range []struct {
		f         float64
		col, name string
	}{
		{ctoSlope, figBad, "CTO au PFU"},
		{avSlope, figBlue, "assurance-vie, hors abattement"},
		{avFriction(1, avRedemption, avAllowanceCouple), figGreen, "assurance-vie, avec abattement"},
		{peaSlope, figAccent, "PEA après 5 ans"},
	} {
		y := py(l.f) + 3.5
		b.WriteString(mTxt(x1+8, y, 10.5, l.col, "start", "600", pctFR(l.f)))
		b.WriteString(sTxt(x1+52, y, 10.5, l.col, "start", "600", l.name))
	}

	// what the fourth curve is, and what the allowance does not buy
	b.WriteString(sTxt(80, 96, 10.5, figGreen, "start", "600",
		"le couple qui rachète 25 000 € par an"))
	b.WriteString(sTxt(80, 110, 10.5, figMuted, "start", "400",
		"l'abattement de 9 200 € efface l'IR jusqu'à 37 %"))
	b.WriteString(sTxt(80, 124, 10.5, figMuted, "start", "400",
		"de part de gain, jamais les prélèvements sociaux"))

	// the reading the article turns on, in the clear zone under the flattest
	// trace: the ranking follows the slope, so the age of the portfolio decides
	young, old := 0.20, 0.80
	b.WriteString(sTxt(x1, 276, 10.5, figSoft, "end", "400",
		"un CTO jeune ("+fmt.Sprintf("%.0f", young*100)+" % de gain) coûte "+pctFR(young*ctoSlope)+","))
	b.WriteString(sTxt(x1, 290, 10.5, figSoft, "end", "400",
		"une assurance-vie mûre ("+fmt.Sprintf("%.0f", old*100)+" %) coûte "+pctFR(old*avSlope)))

	// x axis, with the kink's abscissa called out in the curve's colour
	for _, g := range []float64{0, 0.25, 0.50, 0.75, 1} {
		b.WriteString(mTxt(px(g), 316, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g*100)))
	}
	b.WriteString(mTxt(kx, 316, 10, figGreen, "middle", "600", fmt.Sprintf("%.0f", gStar*100)))
	b.WriteString(sTxt((x0+x1)/2, 336, 11, figMuted, "middle", "400",
		"part de gain dans le retrait, en % (l'âge du portefeuille)"))

	// the plate must age honestly: the year and the basis of every rate
	b.WriteString(sTxt(24, 358, 9.5, figMuted, "start", "400",
		"Taux au 1er janvier 2026 (LFSS 2026). PEA après 5 ans, prélèvements sociaux de 18,6 %. CTO au PFU, 31,4 %."))
	b.WriteString(sTxt(24, 371, 9.5, figMuted, "start", "400",
		"Assurance-vie après 8 ans, versements sous 150 000 €, IR de 7,5 % après abattement et 17,2 % de prélèvements sociaux."))
	return svg(640, 384, b.String())
}
