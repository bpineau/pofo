package firebook

import (
	"fmt"
	"strings"
)

// The plate of the glidepath article's first thesis: a bond tent does not
// manufacture return, it moves it. Every vintage is run twice on the same
// history, once under the tent and once under the static allocation of the same
// average equity exposure, and the plate shows the DIFFERENCE of the two
// maximum sustainable withdrawal rates, vintages sorted from the one the tent
// favours most to the one it favours least. The curve is monotone by
// construction and crosses zero once: left of the crossing the tent pays, right
// of it the tent costs, and the areas are what "transfer" means.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.
//
// The measured facts the plate carries, all of them uncomfortable and all of
// them drawn rather than asserted: only nine of the forty-three vintages gain
// anything, the median vintage pays a quarter of a point, and the winners are
// the hard vintages (the ten with the lowest static sustainable rate average
// +0,07 point, the ten easiest −0,32 point).
//
// The sample is thin and the plate says so on its own surface: the bond leg
// starts in 1953, and a thirty-year retirement then leaves only the 1954-1996
// departures. The shape survives the horizon (25 years: 13 winners out of 48,
// range +0,52 to −1,19; 40 years: 9 out of 33, +0,31 to −1,10), which is why a
// forty-three point curve is worth drawing at all.
//
// What the plate deliberately does NOT draw: the article's second thesis, that
// the benefit grows with the starting CAPE. It cannot be tested here. No
// departure of this window ever started above a CAPE of 25 (the maximum is 24,8
// in January 1996), and over the window the correlation between the starting
// CAPE and the gain is +0,09, i.e. nothing. Marking "CAPE > 25" cohorts would
// have meant marking an empty set, and marking the dearest cohorts instead
// would have shown a split verdict (the 1960s vintages gain, the mid-1990s ones
// are among the worst losers). The plate marks the hardness of the vintage
// instead, which is what the data does support.

// The tent and its static twin, both read off this chapter's prose: the
// household example leaves at 58 % equity and climbs to 85 % ("tous les
// retraits sur la poche obligataire jusqu'à 85 % d'actions"), at the pace the
// chapter writes down for the climb ("de 60 à 90 % d'actions, +2 à +3 points
// par an", so ten years here), and it is compared with "les allocations
// statiques équivalentes", the constant weight of the same average exposure.
const (
	tenteStart  = 0.58 // equity share on the day of departure
	tenteEnd    = 0.85 // cruising equity share
	tenteClimb  = 10   // years of the climb, +2,7 points a year
	tenteYears  = 30   // length of the retirement
	tenteStatic = 0.8005
)

// tenteAlloc is the equity share of the tent in year t of the retirement.
func tenteAlloc(t int) float64 {
	if t >= tenteClimb {
		return tenteEnd
	}
	return tenteStart + (tenteEnd-tenteStart)*float64(t)/float64(tenteClimb)
}

// tenteCohort is one departure vintage: the gain of the tent over its static
// twin, in points of maximum sustainable withdrawal rate, and whether the
// vintage belongs to the ten hardest of the sample (the ten with the lowest
// static sustainable rate, 3,90 % to 4,69 %; the eleventh sits at 4,84 %, so
// the cut is not on a knife edge).
type tenteCohort struct {
	year int
	gain float64
	hard bool
}

// The forty-three vintages, sorted by gain. Reproduce with the two bundled
// legs, month-end SP500-USD and TREASURY-INT-USD from pkg/datasets refdata,
// deflated by the bundled ^CPI-US, folded into December-to-December real
// calendar-year returns; then for each vintage bisect the constant real
// withdrawal rate (taken at the start of the year, portfolio rebalanced to the
// year's target weight) that survives thirty years, once with tenteAlloc and
// once with tenteStatic, and subtract. figures_tente_test.go re-solves all
// forty-three from the datasets and fails on any drift.
var tenteCohorts = []tenteCohort{
	{1973, +0.327, true},
	{1981, +0.289, false},
	{1969, +0.228, true},
	{1970, +0.214, true},
	{1972, +0.172, true},
	{1974, +0.166, false},
	{1968, +0.131, true},
	{1971, +0.094, false},
	{1966, +0.059, true},
	{1965, -0.050, true},
	{1977, -0.067, false},
	{1967, -0.072, true},
	{1984, -0.091, false},
	{1982, -0.117, false},
	{1962, -0.130, true},
	{1976, -0.145, false},
	{1960, -0.171, false},
	{1964, -0.181, true},
	{1980, -0.211, false},
	{1990, -0.221, false},
	{1957, -0.244, false},
	{1956, -0.251, false},
	{1986, -0.255, false},
	{1959, -0.258, false},
	{1978, -0.293, false},
	{1987, -0.314, false},
	{1983, -0.316, false},
	{1961, -0.334, false},
	{1979, -0.337, false},
	{1985, -0.353, false},
	{1989, -0.354, false},
	{1963, -0.360, false},
	{1988, -0.389, false},
	{1992, -0.453, false},
	{1996, -0.454, false},
	{1975, -0.498, false},
	{1993, -0.529, false},
	{1955, -0.547, false},
	{1991, -0.610, false},
	{1994, -0.622, false},
	{1958, -0.669, false},
	{1995, -0.709, false},
	{1954, -1.132, false},
}

// The three readings the plate prints, all of them derived from tenteCohorts
// and re-derived by the guard test: how many vintages gain, what the median
// vintage pays, and the gap between the ten hardest vintages and the ten
// easiest ones.
const (
	tenteWinners  = 9
	tenteMedian   = -0.251
	tenteHardMean = +0.070
	tenteEasyMean = -0.319
)

// figTenteTransfert draws the sorted gain curve. One monotone line, one zero
// crossing, and a second colour on the markers of the hard vintages: the gains
// pile up exactly where the plan was in danger, and they are paid for by the
// vintages that never needed help.
func figTenteTransfert() string {
	const (
		px0, px1 = 62.0, 604.0
		py0, py1 = 128.0, 356.0 // pixels of gain +0.45 and gain -1.25
		gTop     = 0.45
		gBot     = -1.25
	)
	m := mapper(1, float64(len(tenteCohorts)), gTop, gBot, px0, px1, py0, py1)
	x := func(rank int) float64 { return m(float64(rank), 0)[0] }
	y := func(g float64) float64 { return m(1, g)[1] }

	var b strings.Builder
	b.WriteString(plateHead("le glidepath, ce qu'il rapporte",
		"La tente déplace le résultat, elle n'en fabrique pas"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"écart de taux de retrait soutenable, en points de taux : tente 58 → 85 % d'actions en dix ans,"))
	b.WriteString(sTxt(24, 76, 10.5, figMuted, "start", "400",
		"moins la statique de même exposition moyenne (80/20)"))

	// The gaining half of the plane, washed once so the crossing reads as a
	// surface and not just as a line. Pre-blended solid, never rgba.
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		px0, y(gTop), px1-px0, y(0)-y(gTop), figWash)

	for _, g := range []float64{0.25, -0.25, -0.5, -0.75, -1} {
		b.WriteString(line(px0, y(g), px1, y(g), figGrid, 1))
		b.WriteString(mTxt(px0-8, y(g)+3.5, 10, figMuted, "end", "400", tentePts(g)))
	}
	b.WriteString(line(px0, y(0), px1, y(0), figRule, 1.4))
	b.WriteString(mTxt(px0-8, y(0)+3.5, 10, figSoft, "end", "600", "0"))

	// The curve, then the markers on top of it.
	pts := make([][2]float64, len(tenteCohorts))
	for i, c := range tenteCohorts {
		pts[i] = m(float64(i+1), c.gain)
	}
	b.WriteString(poly(pts, figSoft, 1.6, ""))
	for i, c := range tenteCohorts {
		col, r := figBlue, 3.0
		if c.hard {
			col, r = figAccent, 4.2
		}
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s"/>`, pts[i][0], pts[i][1], r, col)
	}

	// Legend, above the plot on the side the curve leaves empty.
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, px0+236, 98.0, figAccent)
	b.WriteString(sTxt(px0+246, 102, 10, figSoft, "start", "400",
		"les dix millésimes les plus durs (taux soutenable statique le plus bas)"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.0" fill="%s"/>`, px0+236, 114.0, figBlue)
	b.WriteString(sTxt(px0+246, 118, 10, figSoft, "start", "400", "les trente-trois autres"))

	// The two ends, named and numbered, each label set clear of the curve.
	b.WriteString(mTxt(x(1)+10, y(tenteCohorts[0].gain)-26, 11, figDeep, "start", "600",
		"1973 : "+tentePts(tenteCohorts[0].gain)+" pt"))
	b.WriteString(sTxt(x(1)+10, y(tenteCohorts[0].gain)-14, 10, figMuted, "start", "400",
		"le meilleur cas de la tente"))
	last := tenteCohorts[len(tenteCohorts)-1]
	b.WriteString(sTxt(px1-24, y(last.gain)-30, 10, figMuted, "end", "400",
		"le pire cas : un millésime qui n'avait besoin de rien"))
	b.WriteString(mTxt(px1-24, y(last.gain)+2, 11, figBad, "end", "600",
		"1954 : "+tentePts(last.gain)+" pt"))

	// The crossing, marked in the band it closes, and counted rather than felt.
	cx := (x(tenteWinners) + x(tenteWinners+1)) / 2
	b.WriteString(dashLine(cx, y(gTop), cx, y(-0.06), figDeep, 1.2, "4 3"))
	b.WriteString(sTxt(px1-8, 152, 10.5, figDeep, "end", "600",
		fmt.Sprintf("%d millésimes sur %d y gagnent ; les %d autres paient la note",
			tenteWinners, len(tenteCohorts), len(tenteCohorts)-tenteWinners)))

	// The transfer itself, in three lines of arithmetic under the curve.
	b.WriteString(sTxt(px0+24, y(-0.72), 10.5, figAccent, "start", "600",
		"les dix millésimes les plus durs : "+tentePts(tenteHardMean)+" pt en moyenne"))
	b.WriteString(sTxt(px0+24, y(-0.86), 10.5, figBlue, "start", "600",
		"les dix plus faciles : "+tentePts(tenteEasyMean)+" pt"))
	b.WriteString(sTxt(px0+24, y(-1.00), 10.5, figSoft, "start", "600",
		"millésime médian : "+tentePts(tenteMedian)+" pt"))

	// x axis: ranks are the only sensible ticks, the years are on the marks.
	b.WriteString(line(px0, y(gBot), px1, y(gBot), figRule, 1))
	for _, r := range []int{1, 10, 20, 30, 43} {
		b.WriteString(line(x(r), y(gBot), x(r), y(gBot)+4, figRule, 1))
		b.WriteString(mTxt(x(r), y(gBot)+16, 10, figMuted, "middle", "400", fmt.Sprintf("%d", r)))
	}
	b.WriteString(sTxt((px0+px1)/2, y(gBot)+32, 10.5, figMuted, "middle", "400",
		"millésimes de départ, classés du plus favorisé au moins favorisé par la tente"))

	b.WriteString(sTxt(24, 420, 9.5, figMuted, "start", "400",
		"Taux de retrait réel constant maximal sur trente ans, calculé millésime par millésime, "+
			"rééquilibrage annuel, hors fiscalité."))
	b.WriteString(sTxt(24, 434, 9.5, figMuted, "start", "400",
		"60/40 américain réel reconstruit (S&amp;P 500 + Treasuries 5 ans, déflatés CPI-U) : "+
			"43 départs seulement, de 1954 à 1996."))
	return svg(640, 446, b.String())
}

// tentePts formats a gain in points of withdrawal rate the French way: a signed
// number with a comma decimal separator and a true minus sign.
func tentePts(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	switch {
	case v < 0:
		s = "−" + s[1:]
	case v > 0:
		s = "+" + s
	}
	return strings.Replace(s, ".", ",", 1)
}
