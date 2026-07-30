package firebook

import (
	"fmt"
	"math"
	"strings"
)

// This file holds the plate of the flat-tax-et-imposition article. Its subject
// is TIME: the sibling plate of enveloppes-francaises reads the friction of a
// withdrawal against the gain fraction of that withdrawal, here the gain
// fraction is itself a function of the age of the line, so the reader watches a
// nominal rate turn slowly into a felt cost. The rates are the dated constants
// of figures_enveloppes.go (1 January 2026, LFSS 2026), never quoted from
// memory; everything else is closed form and recomputed at render time, so
// nothing is frozen in this file and the guard test recomputes the same
// formulas.

// pmpRealGrowth is the real price growth the plate assumes: 5 % a year, the
// order of magnitude of a diversified equity sleeve once inflation is out.
const pmpRealGrowth = 0.05

// pmpMaxYears is the horizon the plate draws, one full retirement.
const pmpMaxYears = 35

// gainFractionLump returns the taxable share of a sale from a line bought in
// one go t years ago, that is 1 − PMP/cours when the price compounds at g a
// year in real terms. It starts at zero and creeps towards one, which is why
// the friction of a flow, rate × gain fraction, drifts upwards for decades.
func gainFractionLump(t, g float64) float64 {
	return 1 - math.Pow(1+g, -t)
}

// gainFractionRegular returns the same share for a line built by one constant
// real contribution at the start of each year, read at the end of year t and
// before the contribution of year t+1. The t contributions cost t units and are
// worth ((1+g)^(t+1) − (1+g))/g, so every new payment pulls the PMP back
// towards the price: the drift slows down without ever stopping. It is defined
// on whole years only, the plate joining the anniversaries.
func gainFractionRegular(t int, g float64) float64 {
	if t < 1 {
		return 0
	}
	n := float64(t)
	value := (math.Pow(1+g, n+1) - (1 + g)) / g
	return 1 - n/value
}

// The calibration band the article gives for the mixed rate of an organised
// plan. It is a rate on gains, hence the level a curve converges to, and the
// plate says so rather than passing it off as a friction of the day.
const (
	mixedRateLow  = 0.15
	mixedRateHigh = 0.25
)

// figFrictionDerivePmp draws the friction of a withdrawal as the line it comes
// from ages: three curves climbing towards the plateau of their own rate, the
// calibration band of the article's tax slider, and the two readings it quotes.
func figFrictionDerivePmp() string {
	const (
		x0, x1 = 76.0, 404.0
		yBase  = 300.0 // 0 %
		yTop   = 96.0  // 33 %
		vMax   = 0.33
	)
	px := func(t float64) float64 { return x0 + t/pmpMaxYears*(x1-x0) }
	py := func(f float64) float64 { return yBase - f/vMax*(yBase-yTop) }
	lump := func(t int) float64 { return gainFractionLump(float64(t), pmpRealGrowth) }
	regular := func(t int) float64 { return gainFractionRegular(t, pmpRealGrowth) }
	trace := func(rate float64, gain func(int) float64) [][2]float64 {
		pts := make([][2]float64, 0, pmpMaxYears+1)
		for t := 0; t <= pmpMaxYears; t++ {
			pts = append(pts, [2]float64{px(float64(t)), py(rate * gain(t))})
		}
		return pts
	}

	var b strings.Builder
	b.WriteString(plateHead("la dérive de la friction",
		"Un taux de 31,4 % ne prélève pas 31,4 % du retrait"))
	b.WriteString(sTxt(x0, 68, 10.5, figMuted, "start", "400",
		"impôt payé pour 100 € de flux extrait"))

	// the calibration band, behind everything: a band of PLATEAUX, not of
	// frictions, which the footnote states in so many words
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, py(mixedRateHigh), x1-x0, py(mixedRateLow)-py(mixedRateHigh), figWash))
	b.WriteString(sTxt(x0+8, py(mixedRateHigh)-7, 10.5, figDeep, "start", "600",
		"la fourchette de calibrage du curseur, 15 à 25 %"))

	// horizontal grid, the zero line doubling as the x axis
	for _, g := range []float64{0, 0.10, 0.20, 0.30} {
		gy := py(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(x0, gy, x1, gy, col, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f %%", g*100)))
	}
	b.WriteString(line(x0, yTop, x0, yBase, figRule, 1))

	// the two plateaux: neutral hairlines, labels in the colour of the curve
	// that climbs towards them
	for _, p := range []struct {
		rate float64
		col  string
		name string
	}{
		{ctoSlope, figBad, "plafond du PFU : "},
		{peaSlope, figAccent, "plafond du PEA : "},
	} {
		gy := py(p.rate)
		b.WriteString(dashLine(x0, gy, x1, gy, figRule, 1, "2 5"))
		b.WriteString(sTxt(x0+8, gy-6, 10.5, p.col, "start", "600", p.name+pctFR(p.rate)))
	}

	// the three traces: the flattest first, the headline on top
	b.WriteString(poly(trace(peaSlope, regular), figAccent, 2, "6 4"))
	b.WriteString(poly(trace(peaSlope, lump), figAccent, 2.2, ""))
	b.WriteString(poly(trace(ctoSlope, lump), figBad, 2.4, ""))

	// the two readings the article turns on, marked on the headline curve
	for _, r := range []struct {
		year   int
		dx, dy float64
	}{{10, -7, -7}, {30, 0, 19}} {
		f := ctoSlope * lump(r.year)
		cx, cy := px(float64(r.year)), py(f)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`,
			cx, cy, figBad)
		anchor := "middle"
		if r.dx != 0 {
			anchor = "end"
		}
		b.WriteString(mTxt(cx+r.dx, cy+r.dy, 10.5, figBad, anchor, "600", pctFR(f)))
	}

	// direct end labels, the value column aligned, no legend
	type endLabel struct {
		f     float64
		col   string
		name  string
		under string
	}
	for _, l := range []endLabel{
		{ctoSlope * lump(pmpMaxYears), figBad, "CTO au PFU, achat unique", ""},
		{peaSlope * lump(pmpMaxYears), figAccent, "PEA mûr, ou barème creux", ""},
		{peaSlope * regular(pmpMaxYears), figAccent, "le même, alimenté chaque année", "le PMP se recharge"},
	} {
		y := py(l.f) + 3.5
		b.WriteString(mTxt(x1+8, y, 10.5, l.col, "start", "600", pctFR(l.f)))
		b.WriteString(sTxt(x1+58, y, 10.5, l.col, "start", "600", l.name))
		if l.under != "" {
			b.WriteString(sTxt(x1+58, y+13, 10, figMuted, "start", "400", l.under))
		}
	}

	// what the eye should take away, right-aligned in the clear wedge under the
	// flattest trace, each line short enough to clear it
	b.WriteString(sTxt(x1, 262, 10.5, figSoft, "end", "600", "au départ, presque rien :"))
	b.WriteString(sTxt(x1, 276, 10.5, figSoft, "end", "400", "un CTO au PFU ne dépasse 15 % du flux"))
	b.WriteString(sTxt(x1, 290, 10.5, figSoft, "end", "400", "qu'après treize ans, et 25 % après trente-deux"))

	// x axis
	for _, t := range []float64{0, 5, 10, 15, 20, 25, 30, 35} {
		b.WriteString(mTxt(px(t), 317, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt((x0+x1)/2, 337, 11, figMuted, "middle", "400",
		"années depuis l'achat de la ligne (ou depuis le premier versement)"))

	// the plate must age honestly: the year of the rates, the growth hypothesis,
	// the contribution schedule, and what the band is really made of
	b.WriteString(sTxt(24, 360, 9.5, figMuted, "start", "400",
		"Taux au 1er janvier 2026 : PFU 31,4 % sur le CTO ; 18,6 % pour un PEA mûr, ou un CTO au barème une année à TMI nulle."))
	b.WriteString(sTxt(24, 373, 9.5, figMuted, "start", "400",
		"Cours en hausse de 5 % par an en réel. Ligne alimentée : un versement constant chaque année, qui recharge le PMP."))
	b.WriteString(sTxt(24, 386, 9.5, figMuted, "start", "400",
		"La fourchette 15-25 % est un taux sur les gains : c'est le plafond d'une courbe, pas la friction des premières années."))
	return svg(640, 398, b.String())
}
