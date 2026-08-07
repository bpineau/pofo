package firebook

import (
	"fmt"
	"math"
	"strings"
)

// Plates for the insurance-premium article (cat bonds, merger arbitrage) and
// for the cash-ladder article. Both follow the v2 plate system: sans labels,
// mono numbers, horizontal text only, solid hex fills (no rgba: KOReader's
// EPUB renderer paints translucent fills solid black).

// ilsMonthlyBp holds the monthly REAL returns, in basis points, of a
// EUR-hedged catastrophe bond fund (Schroder GAIA Cat Bond IF EUR hedged,
// LU0951570927), deflated by the French HICP, from 2013-11 to 2025-12.
// Frozen from the toolkit's own panel so the plate stays deterministic and
// offline; regenerate with the same fetch + deflate pipeline if extended.
var ilsMonthlyBp = []int{
	-34, 59, 38, -16, 29, 18, -45, 54, -12, 136, 83, 125,
	-41, 99, -46, -67, -29, -12, -12, 19, 11, 175, 96, 11,
	16, 110, -22, -62, 10, 3, 15, 35, 4, 111, 68, 37,
	-31, -1, 1, -56, -15, -2, 4, 44, 4, 64, -594, 213,
	25, 27, 100, -117, -23, 18, -25, 0, -15, -23, 9, 25,
	-164, 19, 91, -131, -83, -106, -137, 74, 33, -37, 153, 88,
	-101, 96, 30, 32, -207, 72, -17, 49, 109, 159, 112, 0,
	-24, -14, 60, -32, -7, -23, -16, 3, -40, -110, 165, -60,
	10, -32, -108, -158, -69, -87, -107, -67, -88, 98, -940, 187,
	129, 98, 76, -43, 41, 125, 89, 125, -3, 242, 54, 164,
	70, 52, 44, 44, 25, 10, -64, 52, 70, 312, 177, 77,
	80, 103, -23, -8, 10, 36, -8, 37, 47, 239, 152, 81,
	36, 19,
}

// eqMonthlyBp holds the monthly REAL returns, in basis points, of world
// equity in euros (MSCI World tracker, backcast series) over the same months,
// deflated by the same index.
var eqMonthlyBp = []int{
	223, 83, -236, 342, -50, 24, 402, 176, 37, 363, 225, 62,
	331, 265, 390, 625, 284, -227, 219, -320, 353, -841, -403, 1022,
	324, -354, -752, 32, 171, 53, 302, -76, 430, -23, 21, 15,
	435, 357, 19, 375, 2, -32, -146, -90, -131, -73, 350, 305,
	-17, 145, 66, -309, -374, 324, 350, 151, 112, 139, 107, -488,
	1, -674, 690, 338, 200, 350, -482, 406, 356, -245, 397, -23,
	411, 131, 137, -933, -1126, 982, 287, 106, -76, 807, -216, -326,
	941, 147, 94, 178, 654, 79, 98, 387, 97, 344, -219, 357,
	129, 361, -539, -336, 401, -323, -471, -672, 991, -134, -710, 358,
	110, -582, 354, -50, -86, 77, 183, 457, 109, -34, -161, -379,
	548, 498, 256, 311, 325, -213, 113, 473, -39, 92, 108, 134,
	719, -45, 377, -274, -854, -374, 608, 47, 424, 66, 252, 438,
	-40, 7,
}

// --- 30. Insurance premium: the bad months land on different dates ---
func figSinistresCalendrier() string {
	const (
		x0, x1   = 96.0, 596.0
		topBase  = 152.0 // baseline of the cat bond panel
		botBase  = 330.0 // baseline of the equity panel
		half     = 62.0  // pixels for the full scale
		scaleBp  = 1200.0
		firstYr  = 2013 // the series starts in 2013-11
		firstMon = 10   // 0-based month index of 2013-11
	)
	w := (x1 - x0) / float64(len(ilsMonthlyBp))
	y := func(base float64, bp int) float64 {
		v := math.Max(-scaleBp, math.Min(scaleBp, float64(bp)))
		return base - v/scaleBp*half
	}
	xAt := func(i int) float64 { return x0 + float64(i)*w }

	var b strings.Builder
	b.WriteString(plateHead("primes d'assurance", "Les mauvais mois ne tombent pas aux mêmes dates"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"rendements mensuels réels en euros, novembre 2013 à décembre 2025, même échelle sur les deux panneaux (±12 %)"))

	panel := func(base float64, data []int, pos, label string) {
		b.WriteString(sTxt(x0, base-half-10, 10.5, figSoft, "start", "600", label))
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="0.8" fill="%s"/>`,
			x0, base, x1-x0, figGrid)
		for i, bp := range data {
			fill := pos
			if bp < 0 {
				fill = figBad
			}
			fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s"/>`,
				xAt(i), math.Min(base, y(base, bp)), w*0.78, math.Abs(y(base, bp)-base), fill)
		}
	}
	panel(topBase, ilsMonthlyBp, figDeep, "cat bonds, couverts en euro")
	panel(botBase, eqMonthlyBp, figBlue, "actions mondiales")

	// the two catastrophe months, labelled above their own bars
	mark := func(idx int, text string) {
		x := xAt(idx) + w/2
		b.WriteString(dashLine(x, topBase-8, x, topBase-26, figMuted, 1, "2 3"))
		b.WriteString(sTxt(x, topBase-32, 10.5, figSoft, "middle", "600", text))
	}
	mark(46, "Irma et Maria") // 2017-09
	mark(106, "Ian")          // 2022-09

	// March 2020, the liquidity bridge: a guide spanning both panels
	xc := xAt(76) + w/2
	b.WriteString(dashLine(xc, topBase+14, xc, botBase-half, figBad, 1, "2 3"))
	b.WriteString(sTxt(xc+9, 232, 10.5, figSoft, "start", "600", "mars 2020,"))
	b.WriteString(sTxt(xc+9, 246, 10.5, figMuted, "start", "400", "les vendeurs forcés"))
	b.WriteString(sTxt(xc+9, 260, 10.5, figMuted, "start", "400", "font tout tomber"))

	for yr := 2014; yr <= 2025; yr += 2 {
		i := (yr-firstYr)*12 - firstMon
		if i < 0 || i >= len(ilsMonthlyBp) {
			continue
		}
		b.WriteString(mTxt(xAt(i), 412, 10, figMuted, "middle", "400", fmt.Sprintf("%d", yr)))
	}
	b.WriteString(sTxt(24, 438, 10.5, figMuted, "start", "400",
		"Deux mois sur 146 voient les deux panneaux perdre ensemble : mars 2020, par la liquidité, et septembre 2022, par coïncidence."))
	return svg(640, 452, b.String())
}

// cashRung is one step of the cash ladder plate: what the rung paid above
// the overnight rate, and the worst drawdown it took, both in basis points.
type cashRung struct {
	name, detail string
	gainBp       int
	lossBp       int
}

// --- 31. The cash ladder: extra yield and what it can cost, in basis points ---
func figEchelleDuCash() string {
	rungs := []cashRung{
		{"monétaire (ESTR)", "XEON, l'étalon", 0, -8},
		{"obligataire ultra-court", "ERNX, crédit investment grade court", 27, -20},
		{"CLO AAA en euro", "JCL0, depuis 2024-12", 127, -70},
		{"CLO AAA en dollar", "JAAA, depuis 2020-10", 133, -260},
	}
	const (
		rowH   = 46.0
		top    = 96.0
		midX   = 330.0
		gainX1 = 596.0
		lossX0 = 188.0
	)
	maxGain, maxLoss := 140.0, 280.0
	var b strings.Builder
	b.WriteString(plateHead("poche de liquidités", "Le supplément se compte en points de base, le risque aussi"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"écart de rendement annualisé au monétaire, et pire baisse subie, sur la fenêtre de chaque ligne"))
	b.WriteString(sTxt(midX+14, 82, 10.5, figSoft, "start", "600", "ce que ça rapporte en plus"))
	b.WriteString(sTxt(midX-14, 82, 10.5, figSoft, "end", "600", "ce que ça peut perdre"))

	for i, r := range rungs {
		y := top + float64(i)*rowH
		b.WriteString(sTxt(24, y+4, 10.5, figSoft, "start", "600", r.name))
		b.WriteString(sTxt(24, y+18, 10, figMuted, "start", "400", r.detail))
		// gain, to the right of the spine
		gw := float64(r.gainBp) / maxGain * (gainX1 - midX - 40)
		if gw > 0 {
			b.WriteString(barH(midX+14, midX+14+gw, y-8, 16, figGreen))
			b.WriteString(mTxt(midX+22+gw, y+4, 10.5, figSoft, "start", "600",
				fmt.Sprintf("+%d bp", r.gainBp)))
		} else {
			b.WriteString(mTxt(midX+22, y+4, 10.5, figMuted, "start", "400", "référence"))
		}
		// loss, to the left
		lw := float64(-r.lossBp) / maxLoss * (midX - lossX0 - 46)
		b.WriteString(barHL(midX-14-lw, midX-14, y-8, 16, figBad))
		b.WriteString(mTxt(midX-22-lw, y+4, 10.5, figSoft, "end", "600",
			fmt.Sprintf("%d bp", r.lossBp)))
	}
	// the spine
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="1" height="%.1f" fill="%s"/>`,
		midX, top-20, float64(len(rungs))*rowH-8, figGrid)
	b.WriteString(sTxt(24, 300, 10.5, figMuted, "start", "400",
		"Les deux barres n'ont pas la même unité de temps : le gain est annuel, la perte est un accident bref."))
	b.WriteString(sTxt(24, 316, 10.5, figMuted, "start", "400",
		"Le dollar occupe la dernière ligne pour son historique plus long, pas pour un investisseur en euros."))
	return svg(640, 332, b.String())
}
