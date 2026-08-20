package firebook

import (
	"fmt"
	"strings"
)

// The two readings of one rigid rule. A fixed inflation-indexed withdrawal
// gives the household a cheque that never moves, which is the whole appeal of
// the rule and the reason it is so hard to abandon in time. The plan, looking
// at the same withdrawal, sees something else entirely: the share of the
// remaining capital it takes every year, and that share climbs. The two panels
// are the same 4 % rule, from the two seats.
//
// The vintage is 1966, the worst start year on record, replayed on the bundled
// US 60/40 (pkg/replay) with a 4 % initial withdrawal indexed to inflation.
// Everything is REAL: constant purchasing power, inflation removed.
//
// The numbers are frozen here so the book's figures stay pure functions;
// figures_deuxlectures_test.go recomputes them from pkg/replay under the
// figure-drift gate.

// The replay behind both panels: 1966, thirty years, 1 M of capital and a
// 40 k cheque, in thousands.
const (
	deuxStart   = 1966
	deuxCapital = 1000.0
	deuxCheque  = 40.0
	deuxRuin    = 1994 // the year the money ran out, in the 29th year of the plan
)

// deuxSpend is what the rule actually paid each year, in k of constant money,
// from 1966. It is the cheque the household received: flat, then the last
// partial year, then nothing.
var deuxSpend = []float64{
	40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0,
	40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0,
	40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 40.0, 30.4, 0.0,
}

// deuxRate is that same cheque as a percentage of the capital it was taken
// from, year by year: the withdrawal rate the plan actually ran at. The last
// entry of deuxSpend has no rate, the capital being gone, so this series is one
// year shorter.
var deuxRate = []float64{
	4.00, 4.52, 4.26, 4.29, 5.03, 5.11, 4.99, 4.82, 5.97, 8.23,
	7.67, 7.28, 8.73, 10.00, 11.23, 11.82, 14.20, 13.82, 14.42, 15.92,
	15.46, 15.74, 18.82, 21.57, 23.17, 31.37, 37.80, 58.49, 100.00,
}

// deuxCeiling is where the lower panel stops drawing: the rate runs to 100 %,
// so the curve is drawn while it is on the scale and the rest is written out.
// deuxTint is the level above which the panel is tinted.
const (
	deuxCeiling = 20.0
	deuxTint    = 8.0
)

// deuxExit is the first year whose rate leaves the panel, and deuxRateAt reads
// the series by calendar year. Both are computed, so a refreshed replay moves
// the annotation with the data.
func deuxExit() int {
	for i, v := range deuxRate {
		if v > deuxCeiling {
			return deuxStart + i
		}
	}
	return deuxStart + len(deuxRate)
}

func deuxRateAt(year int) float64 {
	i := year - deuxStart
	if i < 0 || i >= len(deuxRate) {
		return 0
	}
	return deuxRate[i]
}

// figRetraitDeuxLectures draws the cheque above and the rate below, on one
// timeline.
func figRetraitDeuxLectures() string {
	const (
		x0, x1         = 80.0, 600.0
		topHi, topBase = 112.0, 206.0 // the cheque panel, 0 to 48 k
		topMax         = 48.0
		lowTop, lowBot = 252.0, 384.0 // the rate panel, 0 to deuxCeiling
	)
	years := len(deuxSpend)
	slot := (x1 - x0) / float64(years)
	barW := slot - 6
	left := func(i int) float64 { return x0 + slot*float64(i) + 3 }
	mid := func(i int) float64 { return left(i) + barW/2 }
	yTop := func(v float64) float64 { return topBase - v/topMax*(topBase-topHi) }
	yLow := func(v float64) float64 { return lowBot - v/deuxCeiling*(lowBot-lowTop) }

	var b strings.Builder
	b.WriteString(plateHead("la règle des 4 %",
		"Un retrait rigide, et les deux lectures qu'on en fait"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Millésime 1966, 60/40 américain réel, 1 M€ et un retrait initial de 4 % indexé sur l'inflation"))

	// The upper panel: the cheque, one bar per year.
	b.WriteString(sTxt(24, 88, 11, figInk, "start", "600", "Ce que voit le retraité"))
	b.WriteString(sTxt(180, 88, 10, figMuted, "start", "400", "le virement annuel, en euros constants"))
	b.WriteString(dashLine(x0, yTop(deuxCheque), x1, yTop(deuxCheque), figMuted, 1, "4 3"))
	for i, v := range deuxSpend {
		if v <= 0 {
			continue
		}
		fill := figDeep
		if v < deuxCheque {
			fill = figBad // the last, partial cheque
		}
		b.WriteString(barV(left(i), barW, topBase, yTop(v), fill))
	}
	b.WriteString(line(x0, topBase, x1, topBase, figRule, 1))
	b.WriteString(mTxt(x0-10, yTop(deuxCheque)+3.5, 10, figMuted, "end", "400", "40 k€"))
	b.WriteString(mTxt(x0-10, topBase+3.5, 10, figMuted, "end", "400", "0"))
	b.WriteString(sTxt(x1, 100, 10, figBad, "end", "600",
		fmt.Sprintf("le dernier virement tombe en %d", deuxRuin)))

	// The lower panel: the same withdrawal, as a share of what is left.
	b.WriteString(sTxt(24, 230, 11, figInk, "start", "600", "Ce que voit le plan"))
	b.WriteString(sTxt(180, 230, 10, figMuted, "start", "400", "le même retrait, en % du capital restant"))
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, yLow(deuxCeiling), x1-x0, yLow(deuxTint)-yLow(deuxCeiling), figWash)
	for v := 0.0; v <= deuxCeiling; v += 4 {
		b.WriteString(line(x0, yLow(v), x1, yLow(v), figGrid, 1))
		b.WriteString(mTxt(x0-10, yLow(v)+3.5, 10, figMuted, "end", "400",
			fmt.Sprintf("%.0f %%", v)))
	}
	b.WriteString(line(x0, lowBot, x1, lowBot, figRule, 1))

	// The curve, drawn while it is on the scale: it crosses the ceiling and
	// keeps going, which the label under it says in words.
	var pts [][2]float64
	for i, v := range deuxRate {
		if v > deuxCeiling {
			prev := deuxRate[i-1]
			t := (deuxCeiling - prev) / (v - prev)
			pts = append(pts, [2]float64{mid(i-1) + t*(mid(i)-mid(i-1)), yLow(deuxCeiling)})
			break
		}
		pts = append(pts, [2]float64{mid(i), yLow(v)})
	}
	b.WriteString(poly(pts, figAccent, 2.6, ""))
	end := pts[len(pts)-1]
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, end[0], end[1], figAccent)
	// The rest of the story, in the corner the curve has long left behind: it
	// climbs out of the panel and never comes back.
	b.WriteString(sTxt(x1, 352, 10.5, figBad, "end", "600", fmt.Sprintf(
		"le taux quitte l'échelle en %d, passe %s %% en 1992,", deuxExit(), frNum(deuxRateAt(1992), 0))))
	b.WriteString(sTxt(x1, 366, 10.5, figBad, "end", "600",
		fmt.Sprintf("et il n'y a plus rien à retirer en %d", deuxRuin)))

	// The two readings the panels are named after, marked on the curve.
	for _, rd := range []struct {
		year int
		dy   float64
	}{{deuxStart, 16}, {1975, -12}} {
		v := deuxRateAt(rd.year)
		cx, cy := mid(rd.year-deuxStart), yLow(v)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, cx, cy, figDeep)
		b.WriteString(mTxt(cx+8, cy+rd.dy, 10.5, figDeep, "start", "600",
			frNum(v, 1)+" %"))
		b.WriteString(sTxt(cx+8, cy+rd.dy+13, 9.5, figMuted, "start", "400",
			fmt.Sprintf("%d", rd.year)))
	}

	// One timeline, under both panels.
	for _, yr := range []int{1966, 1970, 1975, 1980, 1985, 1990, 1994} {
		b.WriteString(mTxt(mid(yr-deuxStart), 402, 9.5, figMuted, "middle", "400",
			fmt.Sprintf("%d", yr)))
	}

	b.WriteString(sTxt(24, 428, 10.5, figSoft, "start", "600",
		"Le virement ne bouge pas d'un euro pendant vingt-huit ans, et c'est bien ce qui rend la règle si difficile à quitter à temps."))
	b.WriteString(sTxt(24, 444, 9.5, figMuted, "start", "400",
		"Tout est réel, inflation retirée. Le capital s'épuise dans la 29e année : 4 % dépassait ce que ce millésime pouvait porter."))
	return svg(640, 460, b.String())
}
