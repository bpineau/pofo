package firebook

import (
	"fmt"
	"strings"
)

// The second plate of the glidepath article: the tent's slope is about
// equities, but the tent is also MADE of something, and this measures what the
// material is worth. The protocol is the neighbouring tente-transfert plate's,
// unchanged: the same bundled real US history, the same thirty-year vintages,
// the same tent (58 % equity at departure, 85 % after a ten-year climb), the
// same maximum constant real withdrawal rate solved vintage by vintage. Only
// the defensive pocket moves.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.
//
// Three pockets are drawn against the incumbent, intermediate Treasuries: cash
// (3-month bills), the long bond (a 20-year par bond repriced off the bundled
// long constant-maturity yield) and, in the notes only, a fifty-fifty of bills
// and intermediates. The measured facts:
//
//   - The material does NOT change what the slope is worth. Run the tent
//     against its static twin on each of the four pockets and the DISTRIBUTION
//     of the difference barely moves: median vintage between −0,25 and −0,29
//     point, worst vintage between −1,09 and −1,14, nine to eleven winners out
//     of forty-three every time. The article's separation of the two decisions
//     is therefore a measured fact and not a rhetorical one.
//   - The material does change the LEVEL, and its sign flips once in
//     forty-three vintages. The long pocket subtracts from every departure from
//     1954 to 1979 and adds to every departure from 1980 to 1996, the pivot
//     falling within a year of the peak of yields.
//   - The side it lands on is the wrong one. The worst departure of the sample,
//     1966, sustains 3,97 % with intermediates, 4,03 % with bills and 3,69 %
//     with the long bond: the long pocket takes a quarter of a point from the
//     vintage that had none to give and hands three quarters of a point to
//     1982, which needed nothing.
//
// What the plate does NOT claim. This is not a proof that duration is bad; it
// is a measurement of one bond bear and one bond bull, the only two the record
// holds. The mechanism (a long nominal bond is deflation insurance, and the
// hostile regime of this sample was inflationary) is argued in the article from
// the correlation literature, not established here.

// The window and the geometry of the sample, shared with tente-transfert: the
// intermediate leg starts in 1953, and a thirty-year retirement leaves the
// 1954-1996 departures.
const (
	matiereFirst = 1954
	matiereLast  = 1996
)

// matiereVintage is one departure: the sustainable rate the incumbent pocket
// delivers, and what the two other materials add to it or take from it, all in
// points of maximum constant real withdrawal rate, under the tent.
type matiereVintage struct {
	year             int
	base, long, cash float64
}

// The forty-three vintages. Reproduce with the bundled month-end SP500-USD,
// TREASURY-INT-USD, TBILL-3M and TREASURY-LONG-YIELD from pkg/datasets refdata,
// deflated by the bundled ^CPI-US: fold each leg into December-to-December real
// calendar-year returns (bills accrue one twelfth of the previous month's
// annualized rate, the long bond is simgen.TreasuryTR over the month-end long
// yield, 20-year par, 0,10 %/yr), then solve each vintage's sustainable rate
// under tenteAlloc three times. figures_matiere_test.go re-solves all of it
// from the datasets and fails on any drift.
var matiereVintages = []matiereVintage{
	{1954, 7.790, -0.207, +0.046},
	{1955, 6.129, -0.260, +0.076},
	{1956, 5.369, -0.251, +0.021},
	{1957, 5.479, -0.223, -0.050},
	{1958, 6.010, -0.224, +0.030},
	{1959, 5.040, -0.173, -0.011},
	{1960, 4.915, -0.175, -0.089},
	{1961, 4.911, -0.192, +0.054},
	{1962, 4.368, -0.183, +0.055},
	{1963, 4.653, -0.234, +0.117},
	{1964, 4.290, -0.234, +0.092},
	{1965, 4.051, -0.257, +0.100},
	{1966, 3.974, -0.279, +0.055},
	{1967, 4.397, -0.318, +0.073},
	{1968, 4.164, -0.226, +0.015},
	{1969, 4.212, -0.211, -0.017},
	{1970, 4.922, -0.212, -0.187},
	{1971, 4.932, -0.144, +0.018},
	{1972, 4.771, -0.229, +0.102},
	{1973, 4.625, -0.304, +0.087},
	{1974, 5.624, -0.120, +0.051},
	{1975, 7.518, -0.077, -0.016},
	{1976, 6.862, -0.083, +0.007},
	{1977, 6.407, -0.180, +0.203},
	{1978, 7.487, -0.197, +0.114},
	{1979, 8.344, -0.190, -0.076},
	{1980, 8.996, +0.012, -0.284},
	{1981, 9.098, +0.269, -0.628},
	{1982, 10.553, +0.793, -1.120},
	{1983, 9.564, +0.400, -0.606},
	{1984, 9.387, +0.709, -0.840},
	{1985, 9.575, +0.832, -0.834},
	{1986, 8.457, +0.510, -0.482},
	{1987, 7.829, +0.176, -0.281},
	{1988, 8.311, +0.512, -0.501},
	{1989, 8.267, +0.500, -0.632},
	{1990, 7.421, +0.387, -0.515},
	{1991, 8.073, +0.638, -0.597},
	{1992, 7.104, +0.598, -0.375},
	{1993, 7.275, +0.637, -0.387},
	{1994, 7.176, +0.407, -0.283},
	{1995, 7.853, +0.686, -0.637},
	{1996, 6.652, +0.356, -0.358},
}

// The readings the plate prints, all of them re-derived by the guard test.
const (
	// The single sign change of the long pocket, and where it falls.
	matiereLongNeg   = 26   // departures 1954-1979, every one of them a loss
	matiereLongPos   = 17   // departures 1980-1996, every one of them a gain
	matierePivot     = 1979 // the last departure the long pocket costs
	matiereWorstYear = 1966 // the worst departure of the sample, on every pocket

	// The distribution of what the SLOPE is worth, across the four pockets
	// (bills, intermediate, fifty-fifty, long): the band, not a number, is the
	// point. Nothing here depends on the material.
	matiereGainMedLo   = -0.293
	matiereGainMedHi   = -0.248
	matiereGainWorstLo = -1.136
	matiereGainWorstHi = -1.086

	// The fifty-fifty of bills and intermediates at the worst departure, named
	// in the notes so the reader knows where the compromise lands.
	matiereMixWorst = 4.00
)

// figTenteMatiere draws the two panels the question needs. On top, how hard
// each departure was, read off the incumbent pocket: the eye finds the trough
// of the mid-sixties. Underneath, what each material adds to that rate or takes
// from it. The long pocket's curve is below zero for the whole left half, which
// is where the trough is, and above zero for the whole right half, which is
// where nothing was needed.
func figTenteMatiere() string {
	const (
		px0, px1 = 80.0, 596.0
		topBot   = 180.0 // pixel of a 3,5 % sustainable rate
		topTop   = 104.0 // ... and of 11 %
		botBot   = 362.0 // pixel of a gap of −1,15 point
		botTop   = 226.0 // ... and of +0,95
	)
	n := len(matiereVintages)
	x := func(i int) float64 { return px0 + float64(i)*(px1-px0)/float64(n-1) }
	top := figScale{Min: 3.5, Max: 11, Px0: topBot, Px1: topTop}
	bot := figScale{Min: -1.15, Max: 0.95, Px0: botBot, Px1: botTop}
	zero := bot.Map(0)

	// The era the long pocket never once paid in ends between the 1979 and the
	// 1980 departure; both panels are washed on that side of the cut, so the
	// trough of the top panel and the deficit of the bottom one read as one span.
	pivot, wi := 0, 0
	for i, v := range matiereVintages {
		if v.year == matierePivot {
			pivot = i
		}
		if v.year == matiereWorstYear {
			wi = i
		}
	}
	cut := (x(pivot) + x(pivot+1)) / 2
	wash := func(y0, y1 float64) string {
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			px0, y0, cut-px0, y1-y0, figWash)
	}

	var b strings.Builder
	b.WriteString(plateHead("de quoi la tente est faite",
		"La duration longue prend aux départs fragiles et donne aux autres"))
	b.WriteString(plateDeck(
		"même tente, même histoire : une seule chose change, la matière de la poche défensive"))

	// --- top panel: how hard each departure was
	b.WriteString(sTxt(24, 92, 10.5, figSoft, "start", "600",
		"Ce que chaque départ pouvait soutenir, avec la poche intermédiaire"))
	b.WriteString(wash(topTop, topBot))
	for _, g := range []float64{4, 7, 10} {
		gy := top.Map(g)
		b.WriteString(line(px0, gy, px1, gy, figGrid, 1))
		b.WriteString(mTxt(px0-10, gy+3.5, 10, figMuted, "end", "400", frNum(g, 0)+" %"))
	}
	tops := make([][2]float64, n)
	for i, v := range matiereVintages {
		tops[i] = [2]float64{x(i), top.Map(v.base)}
	}
	b.WriteString(poly(tops, figSoft, 1.8, ""))

	// the trough, marked: every pocket puts its worst departure on that year
	worst := matiereVintages[wi]
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, x(wi), top.Map(worst.base), figBad)
	b.WriteString(line(x(wi), 130, x(wi), top.Map(worst.base)-7, figBad, 1))
	b.WriteString(sTxt(x(wi), 126, 10.5, figBad, "middle", "600", "1966, le pire départ"))

	// --- bottom panel: what the material adds to that rate, or takes from it
	b.WriteString(sTxt(24, 214, 10.5, figSoft, "start", "600",
		"Ce que la matière ajoute à ce taux, ou lui retire (points de taux)"))
	b.WriteString(sTxt(px1, 214, 10, figAccent, "end", "600",
		"elle rapporte sur les dix-sept suivants"))
	b.WriteString(wash(botTop, botBot))
	for _, g := range []float64{0.5, -0.5, -1} {
		gy := bot.Map(g)
		b.WriteString(line(px0, gy, px1, gy, figGrid, 1))
		b.WriteString(mTxt(px0-10, gy+3.5, 10, figMuted, "end", "400", frMinus(g, 1)))
	}
	b.WriteString(line(px0, zero, px1, zero, figRule, 1.4))
	b.WriteString(mTxt(px0-10, zero+3.5, 10, figSoft, "end", "600", "0"))

	longs, cashes := make([][2]float64, n), make([][2]float64, n)
	for i, v := range matiereVintages {
		longs[i] = [2]float64{x(i), bot.Map(v.long)}
		cashes[i] = [2]float64{x(i), bot.Map(v.cash)}
	}
	b.WriteString(poly(cashes, figBlue, 1.8, ""))
	b.WriteString(poly(longs, figAccent, 2.4, ""))
	b.WriteString(dashLine(cut, topTop, cut, botBot, figDeep, 1.2, "4 3"))

	// the counts, one on each side of the cut: the sign changes once, and once only
	b.WriteString(sTxt(px0+10, botTop+13, 10, figAccent, "start", "600",
		"1954-1979 : elle coûte sur les vingt-six départs"))

	// the two curves, named where they run flat and far apart
	b.WriteString(sTxt(x(3), bot.Map(matiereVintages[3].cash)-13, 11, figBlue, "start", "600",
		"poche monétaire"))
	b.WriteString(sTxt(x(3), bot.Map(matiereVintages[3].long)+19, 11, figAccent, "start", "600",
		"poche longue (État 20 ans)"))

	// --- x axis: the departure years
	b.WriteString(line(px0, botBot, px1, botBot, figRule, 1))
	for i, v := range matiereVintages {
		if v.year%10 != 0 && v.year != matiereFirst && v.year != matiereLast {
			continue
		}
		b.WriteString(line(x(i), botBot, x(i), botBot+4, figRule, 1))
		b.WriteString(mTxt(x(i), botBot+17, 10, figMuted, "middle", "400", fmt.Sprintf("%d", v.year)))
	}
	b.WriteString(sTxt((px0+px1)/2, botBot+33, 10.5, figMuted, "middle", "400",
		"année de départ à la retraite"))

	b.WriteString(plateConclusion(421,
		"La matière ne change pas ce que la pente rapporte. Elle change le niveau, et du mauvais côté :"))
	b.WriteString(plateConclusion(437,
		"en 1966, "+frNum(worst.base, 2)+" % avec l'intermédiaire, "+
			frNum(worst.base+worst.cash, 2)+" % au monétaire, "+
			frNum(worst.base+worst.long, 2)+" % à la longue."))
	b.WriteString(plateFoot(461, []string{
		"Taux de retrait réel constant maximal sur trente ans, sous la même tente que plus haut " +
			"(58 → 85 % d'actions en dix ans),",
		"rééquilibrage annuel, hors fiscalité. Quarante-trois départs, de 1954 à 1996 : S&amp;P 500 réel, " +
			"et trois poches défensives,",
		"bons du Trésor 3 mois, Treasuries 5 ans, obligation d'État 20 ans reconstruite au taux long " +
			"(tout déflaté du CPI-U).",
		"Ce que la pente ajoute ne dépend pas de la matière, un cinquante-cinquante compris : millésime médian de " +
			frMinus(matiereGainMedHi, 2) + " à " + frMinus(matiereGainMedLo, 2) + ",",
		"pire millésime de " + frMinus(matiereGainWorstHi, 2) + " à " + frMinus(matiereGainWorstLo, 2) +
			". Au pire départ, ce cinquante-cinquante soutient " + frNum(matiereMixWorst, 2) + " %.",
	}))
	return svg(640, 542, b.String())
}
