package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The lever keyboard of les-maths-du-4-pourcent: what each assumption moves,
// measured from the reference plan the article's cascade lands on (a historical
// 60/40 over 30 years with a fixed indexed withdrawal, which comes out at
// 4,0 %). Its companion plate cascade-4pct builds that 4,0 % stage by stage;
// this one holds the result still and pushes on one hypothesis at a time.

// leverBadTint and leverGoodTint are PRE-BLENDED solid hexes (figBad and
// figGreen at 38 % over the figure card background #fffdf9), never rgba: the
// EPUB renderer paints translucent fills solid black. They carry the part of a
// bar that lies between the two bounds of a range, so the eye separates "at
// least this much" from "possibly that much".
const (
	leverBadTint  = "#E7C3BD"
	leverGoodTint = "#B0D4C2"
)

// cascadeLever is one hypothesis of the cascade, with the effect it has on the
// withdrawal rate in points, signed. A point estimate has near == far; a range
// has near as the bound closest to the reference and far as the other one.
type cascadeLever struct {
	name  string  // the hypothesis, in the article's own words
	stage string  // the stage of the cascade it acts on
	near  float64 // points of withdrawal rate, signed, closest bound
	far   float64 // the far bound of the range (== near for a point estimate)
}

// cascadeBase is the reference the levers are measured from: the rate the
// article's cascade produces for a historical 60/40 over 30 years.
const cascadeBase = 4.0

// cascadeLevers freezes the plate, sorted by absolute weight (the midpoint of
// a range) so the ranking is the message.
//
// Every value is the article's own prose, in the "exemple" block that closes
// the cascade, and figures_leviers_test.go re-reads that block:
//
//   - CAPE élevé au départ, "étage 1 raboté d'un point": −1,0 ;
//   - échantillon mondial, "−0,5 à −1 point", the range Anarkulova-Cederburg
//     give in the article's "science" block ;
//   - horizon de 50 ans, "~3,4 %", so −0,6 from 4,0. Careful with this one: the
//     amortization BONUS alone falls by 1,1 point between 30 and 50 years
//     (5,8 % to 4,7 %, closed formula, checked in the test), but the sequence
//     penalty lightens over a longer horizon, and the article's net figure is
//     the one a reader can act on. The horizon-flatten plate of
//     horizon-et-esperance-de-vie measures the same move at 4,05 % on 30 years
//     against 3,37 % on 50, i.e. −0,68: the drawn −0,6 is its conservative
//     twin, not a third number ;
//   - 0,5 % de frais, "−0,5, presque un pour un" ;
//   - règle flexible à plancher, "+0,3 à +0,5".
//
// The sample lever also has a measured counterpart in the book: safemax-pays
// ranks a 30-year SAFEMAX per country on the JST panel, where the United States
// carry 3,75 % and the equal-weight world basket 3,28 %, a gap of 0,48 point.
// That sits at the low end of the drawn range, on a different construction (a
// basket of national experiences, no currency effect, against a bootstrap of
// the global sample), which the test checks as an order of magnitude.
var cascadeLevers = []cascadeLever{
	{"CAPE élevé au départ", "étage 1, le rendement", -1.0, -1.0},
	{"Échantillon mondial", "étage 1, le rendement", -0.5, -1.0},
	{"Horizon de 50 ans", "étage 2, le bonus", -0.6, -0.6},
	{"Frais de 0,5 % par an", "étage 1, le rendement", -0.5, -0.5},
	{"Règle flexible à plancher", "étage 3, la pénalité", 0.3, 0.5},
}

// leverPoints formats a signed number of points the French way.
func leverPoints(v float64) string {
	s := strings.Replace(fmt.Sprintf("%+.1f", v), ".", ",", 1)
	return strings.Replace(s, "-", "−", 1)
}

// leverLabel is the mono label printed at the far end of a bar.
func (l cascadeLever) label() string {
	if l.near == l.far {
		return leverPoints(l.near)
	}
	return leverPoints(l.near) + " à " + leverPoints(l.far)
}

// leverSeg draws one segment of a lever bar, from x0 to x1, on either side of
// the reference rule. The data end (x1) is rounded when round is set, the other
// end always square so two segments of the same bar read as one object.
func leverSeg(x0, x1, y, h float64, fill string, round bool) string {
	if !round {
		lo, hi := math.Min(x0, x1), math.Max(x0, x1)
		return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, lo, y, hi-lo, h, fill)
	}
	if x1 >= x0 {
		return barH(x0, x1, y, h, fill)
	}
	r := 3.5
	if x0-x1 < r*2 {
		r = (x0 - x1) / 2
	}
	// barH mirrored: the rounded end is the left one.
	return fmt.Sprintf(`<path d="M %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Q %.1f,%.1f %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		x0, y, x1+r, y, x1, y, x1, y+r, x1, y+h-r, x1, y+h, x1+r, y+h, x0, y+h, fill)
}

// figClavierLeviers draws one signed bar per lever, all of them leaving the
// same vertical rule at 4,0 %, sorted by absolute weight.
func figClavierLeviers() string {
	const (
		ruleX  = 436.0 // the 4,0 % reference
		perPt  = 150.0 // pixels per point of withdrawal rate
		yTop   = 118.0
		pitch  = 46.0
		bh     = 15.0
		labelR = 196.0 // right edge of the row-label column
		gap    = 8.0   // between a bar end and its number
		split  = 1.5   // hairline between the two segments of a range bar
	)
	x := func(v float64) float64 { return ruleX + v*perPt }
	yBot := yTop + pitch*float64(len(cascadeLevers)-1) + bh

	var b strings.Builder
	b.WriteString(plateHead("les maths du 4 %", "Le clavier des leviers : ce que chaque hypothèse déplace"))
	b.WriteString(sTxt(24, 64, 10.5, figMuted, "start", "400",
		"chaque hypothèse déplacée seule, à partir du plan de référence : 60/40 historique, 30 ans, retrait fixe indexé"))
	b.WriteString(sTxt(24, 80, 10, figMuted, "start", "400",
		"fourchettes : le plein s'arrête à la borne basse, la teinte va jusqu'à la haute"))

	// The two labels that straddle the reference rule (the rule itself is drawn
	// last, over the bars that leave it).
	b.WriteString(sTxt(ruleX-8, 104, 10.5, figDeep, "end", "600", "le plan de référence"))
	b.WriteString(mTxt(ruleX+8, 104, 10.5, figDeep, "start", "600", "4,0 %"))

	// Vertical grid and its mono ticks, in rate terms: a bar end is read as the
	// rate the lever leaves behind.
	for _, g := range []float64{3.0, 3.5, 4.5} {
		gx := x(g - cascadeBase)
		b.WriteString(line(gx, yTop-10, gx, yBot+10, figGrid, 1))
	}
	for _, g := range []float64{3.0, 3.5, 4.0, 4.5} {
		lbl := strings.Replace(fmt.Sprintf("%.1f", g), ".", ",", 1)
		col, weight := figMuted, "400"
		if g == cascadeBase {
			col, weight = figDeep, "600"
		}
		b.WriteString(mTxt(x(g-cascadeBase), yBot+26, 10, col, "middle", weight, lbl))
	}
	b.WriteString(sTxt((x(-1)+x(0.5))/2, yBot+46, 11, figMuted, "middle", "400",
		"taux de retrait obtenu, en % du capital de départ"))

	for i, l := range cascadeLevers {
		y := yTop + pitch*float64(i)
		b.WriteString(sTxt(labelR, y+11.5, 11.5, figSoft, "end", "600", l.name))
		b.WriteString(sTxt(labelR, y+25, 9.5, figMuted, "end", "400", l.stage))

		fill, tint := figBad, leverBadTint
		if l.near > 0 {
			fill, tint = figGreen, leverGoodTint
		}
		xNear, xFar := x(l.near), x(l.far)
		if l.near == l.far {
			b.WriteString(leverSeg(x(0), xFar, y, bh, fill, true))
		} else {
			// Solid up to the near bound, tinted from there to the far one.
			b.WriteString(leverSeg(x(0), xNear, y, bh, fill, false))
			step := split
			if xFar < xNear {
				step = -split
			}
			b.WriteString(leverSeg(xNear+step, xFar, y, bh, tint, true))
		}

		lx, anchor := xFar+gap, "start"
		if l.far < 0 {
			lx, anchor = xFar-gap, "end"
		}
		b.WriteString(mTxt(lx, y+11.5, 10.5, figInk, anchor, "600", l.label()))
	}
	b.WriteString(line(ruleX, 110, ruleX, yBot+14, figDeep, 2))
	return svg(640, 384, b.String())
}
