package firebook

import (
	"fmt"
	"math"
	"strings"
)

// --- What an inflation-linked ladder funds, by real rate and by horizon ---
//
// The plate is pure arithmetic: the payout factor of an annuity-certain,
// r / (1 - (1+r)^-N), which is what a ladder of linkers bought to maturity and
// consumed to the last rung can pay out every year in real terms. Nothing is
// frozen here; every curve, reading and crossing is computed from the constants
// below at render time, and figures_linkers_test.go re-derives the readings the
// article's prose quotes (3,3 / 3,9 / 4,5 %) plus the degenerate r = 0 case.

const (
	// The real-rate window on the x axis, in percent per year: deeply negative
	// rates (2015-2021) on the left, the post-2023 TIPS level on the right.
	linkersRateLo = -1.0
	linkersRateHi = 3.0
	// linkersRule is the 4 % rule, the horizontal reference the ladder is
	// judged against, in percent of the initial capital.
	linkersRule = 4.0
)

// ladderRate returns the annual real withdrawal, in percent of the capital,
// that an inflation-linked bond ladder of n yearly rungs funds at a market real
// rate of r percent per year. It is the payout factor of an annuity-certain,
// r / (1 - (1+r)^-n): the ladder is consumed entirely, so nothing is left at
// the end. The r = 0 limit is 1/n, handled separately to avoid dividing by a
// vanishing denominator.
func ladderRate(r float64, n int) float64 {
	x := r / 100
	if math.Abs(x) < 1e-9 {
		return 100 / float64(n)
	}
	return 100 * x / (1 - math.Pow(1+x, -float64(n)))
}

// ladderCrossing returns the real rate, in percent, at which an n-rung ladder
// starts funding target percent per year (bisection; ladderRate rises with r).
func ladderCrossing(n int, target float64) float64 {
	lo, hi := linkersRateLo, linkersRateHi
	for i := 0; i < 80; i++ {
		mid := (lo + hi) / 2
		if ladderRate(mid, n) < target {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

func figLinkersEchelle() string {
	const (
		x0, x1     = 76.0, 548.0
		yTop, yBot = 92.0, 268.0
		vLo, vHi   = 1.7, 7.4
	)
	x := func(r float64) float64 {
		return x0 + (r-linkersRateLo)/(linkersRateHi-linkersRateLo)*(x1-x0)
	}
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }
	curve := func(n int) [][2]float64 {
		var pts [][2]float64
		for i := int(linkersRateLo * 10); i <= int(linkersRateHi*10); i++ {
			r := float64(i) / 10
			pts = append(pts, [2]float64{x(r), y(ladderRate(r, n))})
		}
		return pts
	}

	var b strings.Builder
	b.WriteString(plateHead("l'échelle indexée",
		"Ce que finance une échelle de linkers, selon le taux réel et l'horizon"))
	b.WriteString(plateDeck(
		"Retrait annuel réel d'une échelle entièrement consommée : r / (1 − (1+r)^−N), en % du capital"))

	// The window where a 30-year ladder pays more than the rule: the wedge
	// between its curve and the 4 % line, from the crossing to the right edge.
	cross30 := ladderCrossing(30, linkersRule)
	var wedge []string
	for i := 0; ; i++ {
		r := cross30 + float64(i)/10
		if r >= linkersRateHi {
			r = linkersRateHi
		}
		wedge = append(wedge, fmt.Sprintf("%.1f,%.1f", x(r), y(ladderRate(r, 30))))
		if r >= linkersRateHi {
			break
		}
	}
	fmt.Fprintf(&b, `<path d="M %s L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
		strings.Join(wedge, " L "), x1, y(linkersRule), x(cross30), y(linkersRule), figWash)

	// grid, axes and ticks
	b.WriteString(sTxt(x0-46, 84, 10, figMuted, "start", "400", "retrait annuel"))
	for v := 2.0; v <= 7.0; v++ {
		b.WriteString(line(x0, y(v), x1, y(v), figGrid, 1))
		b.WriteString(mTxt(x0-8, y(v)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f %%", v)))
	}
	b.WriteString(line(x0, yTop, x0, yBot, figRule, 1))
	b.WriteString(line(x0, yBot, x1, yBot, figRule, 1))
	for i := -1; i <= 3; i++ {
		lbl := fmt.Sprintf("%d %%", i)
		if i < 0 {
			lbl = fmt.Sprintf("−%d %%", -i)
		}
		b.WriteString(mTxt(x(float64(i)), 286, 10, figMuted, "middle", "400", lbl))
	}
	b.WriteString(sTxt(312, 306, 11, figMuted, "middle", "400", "taux réel de marché des linkers  →"))

	// the rule to beat
	b.WriteString(dashLine(x0, y(linkersRule), x1, y(linkersRule), figSoft, 1.4, "5 4"))
	b.WriteString(sTxt(x0+6, y(linkersRule)+15, 10.5, figSoft, "start", "600", "règle des 4 %"))

	// the three horizons, direct end labels, no legend
	horizons := []struct {
		n   int
		col string
		w   float64
	}{
		{20, figBlue, 2.0},
		{30, figAccent, 2.8},
		{40, figBad, 2.0},
	}
	for _, h := range horizons {
		pts := curve(h.n)
		b.WriteString(poly(pts, h.col, h.w, ""))
		e := pts[len(pts)-1]
		b.WriteString(sTxt(e[0]+9, e[1]+3.5, 10.5, h.col, "start", "600", fmt.Sprintf("%d ans", h.n)))
	}

	// the readings the article quotes, on the 30-year curve; each label sits on
	// the side away from the 4 % line so the two never touch.
	for _, rd := range []struct {
		r  float64
		up bool
	}{{0, false}, {1, false}, {2, true}} {
		v := ladderRate(rd.r, 30)
		cx, cy := x(rd.r), y(v)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`, cx, cy, figAccent)
		dy := 15.0
		if rd.up {
			dy = -9
		}
		b.WriteString(mTxt(cx, cy+dy, 11, figDeep, "middle", "600", frNum(v, 1)+" %"))
	}

	// the crossing: where the 30-year ladder catches the rule up
	xc, yc := x(cross30), y(linkersRule)
	b.WriteString(dashLine(xc, yc, xc, yBot, figMuted, 1, "3 3"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.6" fill="none" stroke="%s" stroke-width="1.8"/>`, xc, yc, figInk)
	b.WriteString(mTxt(xc+10, 246, 10.5, figInk, "start", "600", frNum(cross30, 1)+" % réel"))
	b.WriteString(sTxt(xc+10, 260, 10.5, figSoft, "start", "400", "l'échelle de 30 ans rejoint la règle"))

	// the same crossing for 40 years, far to the right
	cross40 := ladderCrossing(40, linkersRule)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.8" fill="none" stroke="%s" stroke-width="1.6"/>`, x(cross40), yc, figBad)
	b.WriteString(mTxt(x(cross40), yc+16, 10.5, figBad, "middle", "600", frNum(cross40, 1)+" %"))

	b.WriteString(plateConclusion(338, fmt.Sprintf(
		"La ligne pointillée est la règle des 4 %%. Au-delà de %s %% réel, l'échelle de 30 ans fait mieux.",
		frNum(cross30, 1))))
	b.WriteString(sTxt(24, 354, 10.5, figMuted, "start", "400", fmt.Sprintf(
		"Repères : 0 %% réel finance %s %%, 1 %% (zone euro récente) %s %%, 2 %% (TIPS de 2023-2024) %s %%.",
		frNum(ladderRate(0, 30), 1), frNum(ladderRate(1, 30), 1), frNum(ladderRate(2, 30), 1))))
	b.WriteString(sTxt(24, 370, 10.5, figMuted, "start", "400", fmt.Sprintf(
		"Sur 40 ans, l'horizon FIRE, elle ne finance que %s %% à 1 %% réel et n'égale la règle qu'au-delà de %s %%.",
		frNum(ladderRate(1, 40), 1), frNum(cross40, 1))))
	return svg(640, 384, b.String())
}
