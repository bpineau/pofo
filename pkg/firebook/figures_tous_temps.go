package firebook

import (
	"fmt"
	"strings"
)

// The all-weather article's plate: the trade the family sells, drawn as a
// return / worst-path cloud. See figTousTempsEchange.

// tousTempsPoint is one portfolio measured over the reconstruction window:
// annualized real return and deepest real drawdown, both in percent.
type tousTempsPoint struct {
	name   string
	cagr   float64
	drawdn float64
}

// The six portfolios of the article plus the rest of the equity/bond ladder,
// measured on this repository's bundled series over 1972-2024 (December 1971
// to December 2024), in real terms, rebalanced every December.
//
// Legs: equities SP500, long Treasuries TLT, intermediate Treasuries IEF,
// short Treasuries SHY (T-bills before it starts in 1991-10), gold XAUUSD
// (all pkg/datasets/simdata); small-cap value USSCV-USD, cash TBILL-3M (the
// previous month's annualized rate, divided by 12), commodities WTI-USD (all
// pkg/datasets/refdata); deflator ^CPI-US (the snapshot embedded in
// pkg/marketdata). Compositions: Browne 25/25/25/25 equities, long, gold,
// cash; Golden Butterfly 20 equities, 20 small-cap value, 20 long, 20 short,
// 20 gold; All-Weather (the Robbins version) 30 equities, 40 long, 15
// intermediate, 7.5 gold, 7.5 commodities; the ladder equities against
// intermediate Treasuries, by steps of 10 points.
//
// TestTousTempsFigureMatchesTheData recomputes every pair from those series
// and fails when the plate and the data disagree.
var (
	tousTempsFamily = []tousTempsPoint{
		{"Browne 4 × 25", 4.41, -22.2},
		{"All-Weather *", 5.01, -28.7},
		{"Golden Butterfly", 5.78, -21.6},
	}
	// tousTempsLadder runs from 100 % intermediate Treasuries to 100 %
	// equities, by steps of 10 points of equities.
	tousTempsLadder = []tousTempsPoint{
		{"0/100", 2.66, -41.2},
		{"10/90", 3.23, -38.8},
		{"20/80", 3.77, -36.9},
		{"30/70", 4.27, -35.4},
		{"40/60", 4.74, -34.5},
		{"50/50", 5.18, -33.8},
		{"60/40", 5.58, -37.6},
		{"70/30", 5.94, -41.3},
		{"80/20", 6.27, -45.0},
		{"90/10", 6.55, -48.5},
		{"100/0", 6.80, -54.0},
	}
)

// tousTempsLadderAt reads the ladder's worst drawdown at a given real return,
// linearly between the two steps that bracket it. The plate's central
// annotation compares the Golden Butterfly with the ladder at its own return,
// so the comparison must sit on the drawn line and not beside it.
func tousTempsLadderAt(cagr float64) float64 {
	for i := 1; i < len(tousTempsLadder); i++ {
		lo, hi := tousTempsLadder[i-1], tousTempsLadder[i]
		if cagr >= lo.cagr && cagr <= hi.cagr {
			t := (cagr - lo.cagr) / (hi.cagr - lo.cagr)
			return lo.drawdn + t*(hi.drawdn-lo.drawdn)
		}
	}
	return tousTempsLadder[len(tousTempsLadder)-1].drawdn
}

// ttNum is frNum with a typographic minus sign (U+2212), the house rule for
// negative numbers in a figure label.
func ttNum(v float64, dec int) string {
	return strings.Replace(frNum(v, dec), "-", "−", 1)
}

// figTousTempsEchange draws the trade at the heart of the all-weather family:
// a real return the equity/bond ladder matches easily, against a worst path
// the ladder never comes close to. The family is a cloud of three points, the
// ladder a line of eleven, and one vertical rule reads the Golden Butterfly
// against the ladder at the very same return.
func figTousTempsEchange() string {
	m := mapper(2, 7.2, -60, -14, 84, 600, 322, 96)
	var b strings.Builder
	b.WriteString(plateHead("portefeuilles tous-temps", "À rendement égal, un pire chemin presque deux fois moins profond"))
	legendChips(&b, 56, [][2]string{
		{figAccent, "famille tous-temps"},
		{figBlue, "échelle actions / obligations (pas de 10 points)"},
	})

	// horizontal grid and the two axis titles
	for _, g := range []float64{-20, -30, -40, -50} {
		gy := m(2, g)[1]
		b.WriteString(line(84, gy, 600, gy, figGrid, 1))
		b.WriteString(mTxt(76, gy+3.5, 10, figMuted, "end", "400", ttNum(g, 0)+" %"))
	}
	b.WriteString(sTxt(84, 86, 10.5, figMuted, "start", "400", "pire drawdown réel (%)"))
	b.WriteString(line(84, 322, 600, 322, figRule, 1))
	for _, t := range []float64{2, 3, 4, 5, 6, 7} {
		b.WriteString(mTxt(m(t, 0)[0], 338, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt(342, 358, 11, figMuted, "middle", "400", "rendement réel annualisé (%)"))

	// the ladder, as the continuum it is
	lad := make([][2]float64, len(tousTempsLadder))
	for i, p := range tousTempsLadder {
		lad[i] = m(p.cagr, p.drawdn)
	}
	b.WriteString(poly(lad, figBlue, 2.2, ""))
	for _, p := range lad {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="%s" stroke="#fffdf9" stroke-width="1.4"/>`, p[0], p[1], figBlue)
	}
	b.WriteString(sTxt(lad[0][0]+9, lad[0][1]+16, 10.5, figBlue, "start", "400", "100 % obligations"))
	sixty := m(tousTempsLadder[6].cagr, tousTempsLadder[6].drawdn)
	b.WriteString(sTxt(sixty[0]-8, sixty[1]+15, 10.5, figBlue, "end", "400", "60/40"))
	last := lad[len(lad)-1]
	b.WriteString(sTxt(last[0], last[1]+17, 10.5, figBlue, "middle", "400", "100 % actions"))

	// the ladder's own floor, named where the eye already sees the curve turn
	floor := tousTempsLadder[5] // 50 % equities, 50 % intermediate Treasuries
	fp := m(floor.cagr, floor.drawdn)
	b.WriteString(dashLine(300, 250, fp[0]-3, fp[1]+8, figMuted, 1, "2 3"))
	b.WriteString(sTxt(140, 262, 10.5, figSoft, "start", "600", "le moins profond de toute l'échelle, vers 50 % d'actions"))
	b.WriteString(mTxt(140, 278, 10.5, figSoft, "start", "600", ttNum(floor.drawdn, 0)+" %"))

	// the Golden Butterfly read against the ladder at its own return
	gb := tousTempsFamily[2]
	gp := m(gb.cagr, gb.drawdn)
	iso := tousTempsLadderAt(gb.cagr)
	ip := m(gb.cagr, iso)
	b.WriteString(dashLine(gp[0], gp[1]+8, ip[0], ip[1]-5, figDeep, 1.2, "4 4"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="none" stroke="%s" stroke-width="1.6"/>`, ip[0], ip[1], figDeep)
	b.WriteString(mTxt(ip[0]+9, ip[1]+4, 10.5, figDeep, "start", "600", ttNum(iso, 0)+" %"))
	// The block hangs to the right of the fund's own dashed line, far enough
	// from it to clear the All-Weather label on its left: the two points sit on
	// the same line once the small-value leg pays its costs.
	b.WriteString(sTxt(gp[0]+22, gp[1]+45, 11, figDeep, "start", "600", "à rendement égal,"))
	b.WriteString(sTxt(gp[0]+22, gp[1]+59, 11, figDeep, "start", "600", "un plongeon presque"))
	b.WriteString(sTxt(gp[0]+22, gp[1]+73, 11, figDeep, "start", "600", "deux fois plus profond"))

	// the family, above the whole ladder. Every name sits over its point; the
	// three pairs of numbers go where the neighbours leave room, Browne's to
	// the left because the All-Weather label needs the space on its right.
	nameDy := []float64{-19, -18, -19}
	valLeft := []bool{true, false, false}
	for i, p := range tousTempsFamily {
		q := m(p.cagr, p.drawdn)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`, q[0], q[1], figAccent)
		b.WriteString(sTxt(q[0], q[1]+nameDy[i], 11, figSoft, "middle", "600", p.name))
		val := ttNum(p.cagr, 1) + " % · " + ttNum(p.drawdn, 0) + " %"
		if valLeft[i] {
			b.WriteString(mTxt(q[0]-11, q[1]+4, 10.5, figDeep, "end", "600", val))
			continue
		}
		b.WriteString(mTxt(q[0]+11, q[1]+4, 10.5, figDeep, "start", "600", val))
	}

	b.WriteString(sTxt(24, 380, 9.5, figMuted, "start", "400",
		"US, 1972-2024, en réel, rééquilibrage annuel. Reconstruction sur le S&amp;P 500, les Treasuries, l'or, les T-bills et le small value,"))
	b.WriteString(sTxt(24, 394, 9.5, figMuted, "start", "400",
		"ce dernier net de 1,0 point de coûts par an, le prix qu'un vrai fonds paie et pas le portefeuille académique."))
	b.WriteString(sTxt(24, 408, 9.5, figMuted, "start", "400",
		"* All-Weather : la jambe matières premières est approchée par le pétrole, son pire chemin dépend de ce choix."))
	return svg(640, 420, b.String())
}
