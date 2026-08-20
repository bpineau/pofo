package firebook

import (
	"fmt"
	"strings"
)

// The 2022 plate. The article's example prints three numbers for that year and
// leaves the reader to reconcile them: the naked 60/40 lost 17 %, the stacked
// core lost 21 %, and the whole plan got away with 10 %. Read as prose, the
// third number looks like luck. Decomposed as a waterfall, it is an addition
// with two terms, and both of them are the point of return stacking: the
// leverage bill is real, and it is paid by the layer the leverage freed.
//
// Every level is the article's own. The two intermediate steps are pure
// arithmetic on those levels and on the plan's stated 67 % core weight; no
// market series is recomputed here, and nothing is attributed that the article
// does not attribute. In particular the sleeves' contribution is a RESIDUAL,
// which the article credits to a very good trend year, and the plate says so
// rather than splitting it into instruments the prose never quantifies.

// The example's 2022, in percent of the plan's value, and the weight the
// stacked core carries in it.
const (
	stack22Naked      = -17.0 // the plain 60/40 that year
	stack22Core       = -21.0 // the stacked core, on its own
	stack22Plan       = -10.0 // the whole plan
	stack22CoreWeight = 0.67  // what the core weighs in the plan
)

// The three moves between those levels, all derived, all exact.
//
// stack22Lever is the leverage bill: what the stacked core lost beyond the
// naked 60/40. stack22Dilution is what the core NOT weighing the whole plan
// gives back, since a third of the book was freed without selling a share.
// stack22Sleeves is the residual, what the freed third actually earned.
func stack22Lever() float64 { return stack22Core - stack22Naked }

func stack22Dilution() float64 { return stack22Core*stack22CoreWeight - stack22Core }

func stack22Sleeves() float64 { return stack22Plan - stack22Core*stack22CoreWeight }

// stack22Step is one bar of the waterfall: where it starts, where it ends, and
// how it is named under the axis. A step whose start is zero is a LEVEL, one
// of the two states being compared; the others are moves between them.
type stack22Step struct {
	from, to   float64
	fill       string
	l1, l2     string
	labelBelow bool // the value sits under the bar rather than over it
}

// stack22Steps builds the waterfall from the constants above, so the drawing
// can never disagree with the arithmetic.
func stack22Steps() []stack22Step {
	mid := stack22Core * stack22CoreWeight
	return []stack22Step{
		{0, stack22Naked, figDeep, "60/40 nu", "l'année 2022", true},
		{stack22Naked, stack22Core, figBad, "la facture", "du levier", true},
		{stack22Core, mid, figBlue, "le cœur ne pèse", "que 67 % du plan", false},
		{mid, stack22Plan, figGreen, "l'étage libéré", "trend, or, cash", false},
		{0, stack22Plan, figDeep, "le plan empilé", "au complet", true},
	}
}

// figStacking2022 draws the waterfall.
func figStacking2022() string {
	const (
		x0, x1     = 90.0, 570.0
		yTop, yBot = 112.0, 330.0
		vHi, vLo   = 1.0, -24.0
		barW       = 76.0
	)
	steps := stack22Steps()
	slot := (x1 - x0) / float64(len(steps))
	left := func(i int) float64 { return x0 + slot*float64(i) + (slot-barW)/2 }
	mid := func(i int) float64 { return left(i) + barW/2 }
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("le return stacking",
		"L'anatomie de 2022, de la perte nue à la perte empilée"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Le millésime 2022 du plan de l'exemple, décomposé en une addition"))

	// The value axis, zero picked out: everything happens below it.
	for v := 0.0; v >= -24; v -= 4 {
		col := figGrid
		if v == 0 {
			col = figRule
		}
		b.WriteString(line(x0-16, y(v), x1, y(v), col, 1))
		lbl := fmt.Sprintf("%.0f %%", v)
		if v < 0 {
			lbl = fmt.Sprintf("−%.0f %%", -v)
		}
		b.WriteString(mTxt(x0-24, y(v)+3.5, 10, figMuted, "end", "400", lbl))
	}

	// The bars, each carrying its own move, with a dashed hand-over to the
	// next one at the level they share.
	for i, s := range steps {
		b.WriteString(barV(left(i), barW, y(s.from), y(s.to), s.fill))
		val := frMinus(s.to-s.from, 1)
		if s.from == 0 {
			val = frMinus(s.to, 1) + " %" // a level, not a move
		} else if s.to > s.from {
			val = "+" + val
		}
		if s.labelBelow {
			b.WriteString(mTxt(mid(i), y(s.to)+16, 11.5, figInk, "middle", "600", val))
		} else {
			b.WriteString(mTxt(mid(i), y(s.to)-9, 11.5, figInk, "middle", "600", val))
		}
		// The hand-over: a dash at the level this bar leaves, drawn whenever the
		// next bar starts there or, for the closing bar, lands there.
		if i < len(steps)-1 && (steps[i+1].from == s.to || steps[i+1].to == s.to) {
			b.WriteString(dashLine(left(i)+barW, y(s.to), left(i+1), y(s.to), figMuted, 1, "3 3"))
		}
		b.WriteString(sTxt(mid(i), 352, 10.5, figSoft, "middle", "600", s.l1))
		b.WriteString(sTxt(mid(i), 366, 10, figMuted, "middle", "400", s.l2))
	}

	// The sentence the plate exists to prove, under the bar that proves it.
	b.WriteString(sTxt(x1, 394, 11, figInk, "end", "600",
		"Le levier pique, et c'est l'étage libéré qui paie la facture."))

	b.WriteString(sTxt(24, 416, 9.5, figMuted, "start", "400", fmt.Sprintf(
		"Les trois niveaux sont ceux de l'article : %s %% pour le 60/40 nu, %s %% pour le cœur empilé seul, %s %% pour le plan entier.",
		frMinus(stack22Naked, 0), frMinus(stack22Core, 0), frMinus(stack22Plan, 0))))
	b.WriteString(sTxt(24, 430, 9.5, figMuted, "start", "400",
		"Les deux marches du milieu en découlent, le cœur pesant 67 % du plan. L'article crédite le solde au très bon millésime du trend."))
	return svg(640, 446, b.String())
}
