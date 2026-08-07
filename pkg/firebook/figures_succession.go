package firebook

import (
	"fmt"
	"strconv"
	"strings"
)

// The plate of the transmission article: the staircase of the fifteen-year
// gift reloads, read against the age at which the first wave is made.
//
// Nothing here is frozen. The steps are recomputed at render time from the
// four dated constants below, so the day an allowance moves the figure, the
// guard test and the prose all move together.
//
// Sources, all verified 2026-07-30 against service-public.gouv.fr and the
// articles of the CGI they cite:
//   - art. 779 I CGI: 100 000 EUR of allowance per parent and per child,
//     rebuilt in full after fifteen years.
//   - art. 790 G CGI: 31 865 EUR of "don familial de sommes d'argent" on top,
//     per parent and per child, same fifteen-year clock, but ONLY while the
//     donor is under eighty and the recipient is an adult.
//   - art. 784 CGI: the fifteen-year "rappel fiscal" that resets both.
//
// The household is the one the article names: a couple, two adult children,
// both parents taken at the same age. So a wave made before eighty is worth
// 2 x 2 x (100 000 + 31 865) = 527 460 EUR, and a wave made at eighty or later
// loses the family cash gift and is worth 2 x 2 x 100 000 = 400 000 EUR.
const (
	succAbattementDirect = 100000.0 // art. 779 I CGI, per parent, per child
	succDonFamilial      = 31865.0  // art. 790 G CGI, per parent, per child
	succDonAgeMax        = 80       // art. 790 G CGI: donor strictly under 80
	succRappelAns        = 15       // art. 784 CGI
	succParents          = 2
	succEnfants          = 2
	succHorizon          = 90 // the age the plate reads the cumulative total at
)

// succStep is what one wave of gifts passes free of duty, for the whole
// household, when the parents are aged age.
func succStep(age int) float64 {
	n := float64(succParents * succEnfants)
	s := n * succAbattementDirect
	if age < succDonAgeMax {
		s += n * succDonFamilial
	}
	return s
}

// succWaves lists the ages at which a household starting at start can give,
// the fifteen-year clock of art. 784 CGI running from the first wave.
func succWaves(start, horizon int) []int {
	var ages []int
	for a := start; a <= horizon; a += succRappelAns {
		ages = append(ages, a)
	}
	return ages
}

// succCumul returns the wave ages and the running total transmitted free of
// duty, one entry per wave.
func succCumul(start, horizon int) ([]int, []float64) {
	ages := succWaves(start, horizon)
	cum := make([]float64, len(ages))
	acc := 0.0
	for i, a := range ages {
		acc += succStep(a)
		cum[i] = acc
	}
	return ages, cum
}

// succGrouped formats a whole number of thousands the French way, "1 455".
func succGrouped(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	head := len(s) % 3
	if head > 0 {
		b.WriteString(s[:head])
	}
	for i := head; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// figEscalierRecharges draws three cumulative staircases of duty-free
// transmission over an age axis, one per age of the first wave. The steps are
// all the same height, so the eye reads the only thing that separates the
// three households: how many of them they had time to climb.
func figEscalierRecharges() string {
	const (
		xLeft, xRight   = 76.0, 520.0
		ageMin, ageMax  = 50, succHorizon
		yBase, yTop     = 326.0, 96.0
		vTop            = 1600.0 // k EUR at the top gridline
		lateWaveIsAtEnd = succHorizon
	)
	x := func(age float64) float64 {
		return xLeft + (age-ageMin)/(ageMax-ageMin)*(xRight-xLeft)
	}
	// y takes k EUR. off is a purely visual nudge (see below).
	y := func(k, off float64) float64 {
		return yBase - k/vTop*(yBase-yTop) + off
	}

	var b strings.Builder
	b.WriteString(plateHead("donations rechargeables",
		"Un flux, pas un événement : l'âge du premier don fait tout"))
	b.WriteString(sTxt(24, 62, 10, figSoft, "start", "400",
		"Couple, deux enfants majeurs. Une marche = 527 k€ : quatre abattements de 100 000 € (art. 779 CGI)"))
	b.WriteString(sTxt(24, 76, 10, figSoft, "start", "400",
		"et quatre dons familiaux de 31 865 € (art. 790 G), rechargeables tous les 15 ans (art. 784 CGI)."))

	// grid and value ticks
	for _, g := range []float64{0, 400, 800, 1200, 1600} {
		gy := y(g, 0)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(xLeft, gy, xRight, gy, col, 1))
		b.WriteString(mTxt(xLeft-8, gy+3.5, 10, figMuted, "end", "400", succGrouped(int(g))))
	}
	b.WriteString(sTxt(xLeft, 110, 10.5, figMuted, "start", "400",
		"cumul transmis en franchise totale de droits (k€)"))

	// The three households. Amber, blue and red are the book's CVD-validated
	// trio; the offsets are a legibility nudge of three pixels, because the
	// three staircases share the 527 k€ tread for a few years and would
	// otherwise hide one another. Every exact value is written out in a label.
	type track struct {
		start int
		off   float64
		col   string
	}
	tracks := []track{
		{55, -3, figAccent},
		{65, 0, figBlue},
		{75, 3, figBad},
	}
	for _, t := range tracks {
		ages, cum := succCumul(t.start, succHorizon)
		pts := [][2]float64{{x(float64(t.start)), y(0, 0)}}
		solid := len(ages)
		// A wave landing exactly on the last year of the axis is drawn as a
		// dashed riser: it is legally open, but only to a household still
		// there to make it.
		if ages[len(ages)-1] == lateWaveIsAtEnd && len(ages) > 1 {
			solid--
		}
		for i := 0; i < solid; i++ {
			pts = append(pts,
				[2]float64{x(float64(ages[i])), y(cum[i]/1000, t.off)},
				[2]float64{x(float64(ages[i]) + succRappelAns), y(cum[i]/1000, t.off)})
		}
		// clip the last tread at the right edge of the frame
		last := pts[len(pts)-1]
		if last[0] > xRight {
			pts[len(pts)-1] = [2]float64{xRight, last[1]}
		}
		b.WriteString(poly(pts, t.col, 2, ""))

		// each riser's own size, so the reader sees the treads are equal
		for i := 0; i < solid; i++ {
			from := 0.0
			if i > 0 {
				from = cum[i-1] / 1000
			}
			y0, y1 := y(from, t.off), y(cum[i]/1000, t.off)
			if i == 0 {
				y0 = y(0, 0)
			}
			b.WriteString(mTxt(x(float64(ages[i]))-6, (y0+y1)/2+3.5, 10, figMuted, "end", "400",
				fmt.Sprintf("+%.0f", succStep(ages[i])/1000)))
		}

		// the dashed riser of the late household, and its total so far
		endK := cum[solid-1] / 1000
		if solid < len(ages) {
			b.WriteString(dashLine(xRight, y(endK, t.off), xRight, y(cum[len(ages)-1]/1000, t.off),
				t.col, 1.6, "4 3"))
			b.WriteString(sTxt(xRight-6, 278, 10, t.col, "end", "400",
				fmt.Sprintf("2ᵉ vague seulement à %d ans", lateWaveIsAtEnd)))
		}
		ey := y(endK, t.off)
		vagues := "vagues"
		if solid == 1 {
			vagues = "vague"
		}
		b.WriteString(mTxt(xRight+8, ey+4, 11.5, t.col, "start", "600",
			succGrouped(int(endK+0.5))+" k€"))
		b.WriteString(sTxt(xRight+8, ey+18, 10, figMuted, "start", "400",
			fmt.Sprintf("départ à %d ans", t.start)))
		b.WriteString(sTxt(xRight+8, ey+31, 10, figMuted, "start", "400",
			fmt.Sprintf("%d %s", solid, vagues)))
	}

	// why three risers out of six are shorter than the others
	b.WriteString(sTxt(xLeft, 132, 10, figMuted, "start", "400",
		"après 80 ans, le don familial s'éteint : la marche tombe à 400 k€"))

	// age axis
	b.WriteString(line(xLeft, yBase, xRight, yBase, figRule, 1))
	for a := ageMin; a <= ageMax; a += 5 {
		b.WriteString(mTxt(x(float64(a)), 342, 10, figMuted, "middle", "400", strconv.Itoa(a)))
	}
	b.WriteString(sTxt((xLeft+xRight)/2, 362, 11, figMuted, "middle", "400",
		"âge des parents (supposés du même âge)"))
	return svg(640, 380, b.String())
}
