package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The amortization plate. Tier two of the article is a formula, and a formula
// stated in prose hides its shape: the reader is told a 30-year horizon funds
// 5,8 % and a 50-year one 4,7 %, and concludes that a longer horizon costs a
// lot. The curve says the opposite. Almost the whole amortization bonus is
// spent in the first twenty years; from fifty years to eternity, what is left
// to lose is seven tenths of a point.
//
// Nothing here is measured: the plate is the annuity payout factor, drawn from
// one pure function that the guard test re-reads against the numbers the
// article's prose quotes.

// amortRate is the annuity payout factor r / (1 - (1+r)^-n): the withdrawal,
// in percent of the initial capital, that a real return of r percent funds
// over n years and that lands exactly on zero on the last day. The n -> +inf
// limit is r itself, which is the plate's asymptote.
func amortRate(r float64, n float64) float64 {
	x := r / 100
	if math.Abs(x) < 1e-12 {
		return 100 / n
	}
	return 100 * x / (1 - math.Pow(1+x, -n))
}

// The article's own tier two: a 4 % real return, horizons from ten to sixty
// years, and the three readings the plate annotates.
const (
	amortReal   = 4.0
	amortFirstN = 10.0
	amortLastN  = 60.0
)

// amortGap is what a plan gives up by stretching its horizon from a to b, in
// points of withdrawal rate; amortForever is what the last stretch costs, from
// b years all the way to a perpetuity, where the bonus is gone and only the
// real return is left.
func amortGap(a, b float64) float64 { return amortRate(amortReal, a) - amortRate(amortReal, b) }

func amortForever(a float64) float64 { return amortRate(amortReal, a) - amortReal }

// figAmortissementHorizon draws the payout factor against the horizon, with
// the perpetuity as an asymptote and the two stretches that carry the point.
func figAmortissementHorizon() string {
	const (
		x0, x1     = 70.0, 546.0
		xInf       = 596.0 // the perpetuity, past a deliberate gap in the axis
		yTop, yBot = 100.0, 300.0
		vLo, vHi   = 3.6, 13.0
	)
	x := func(n float64) float64 {
		return x0 + (n-amortFirstN)/(amortLastN-amortFirstN)*(x1-x0)
	}
	y := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("les maths du 4 %", "Le bonus d'amortissement fond avec l'horizon"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Le retrait qui épuise exactement le capital, à 4 % réel, selon le nombre d'années à couvrir"))

	// Grid and axes.
	for v := 4.0; v <= 12.0; v += 2 {
		b.WriteString(line(x0, y(v), xInf, y(v), figGrid, 1))
		b.WriteString(mTxt(x0-10, y(v)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f %%", v)))
	}
	b.WriteString(line(x0, yTop, x0, yBot, figRule, 1))
	b.WriteString(line(x0, yBot, xInf, yBot, figRule, 1))
	for n := amortFirstN; n <= amortLastN; n += 10 {
		b.WriteString(mTxt(x(n), 318, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", n)))
	}
	b.WriteString(mTxt(xInf, 318, 13, figMuted, "middle", "400", "∞"))
	b.WriteString(sTxt(x0, 344, 10.5, figMuted, "start", "400", "années à couvrir  →"))

	// The asymptote: the real return alone, which is what a perpetuity funds.
	b.WriteString(dashLine(x0, y(amortReal), xInf, y(amortReal), figSoft, 1.4, "5 4"))
	b.WriteString(sTxt(x0+6, y(amortReal)-8, 10.5, figSoft, "start", "600",
		"4 % : la perpétuité, le rendement seul"))

	// The curve.
	var pts [][2]float64
	for n := amortFirstN; n <= amortLastN+0.01; n += 0.5 {
		pts = append(pts, [2]float64{x(n), y(amortRate(amortReal, n))})
	}
	b.WriteString(poly(pts, figAccent, 2.8, ""))
	// And its end point, the perpetuity itself, drawn as an open circle on the
	// asymptote because no finite horizon ever reaches it.
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4" fill="none" stroke="%s" stroke-width="1.6"/>`,
		xInf, y(amortReal), figSoft)

	// The three readings, labelled on the side that keeps them clear of the
	// curve: the first above, the two later ones above and to the right.
	for _, rd := range []struct {
		n     float64
		dx    float64
		align string
	}{{10, 12, "start"}, {30, 0, "middle"}, {50, 0, "middle"}} {
		v := amortRate(amortReal, rd.n)
		cx, cy := x(rd.n), y(v)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, cx, cy, figDeep)
		b.WriteString(mTxt(cx+rd.dx, cy-11, 11.5, figDeep, rd.align, "600", frNum(v, 1)+" %"))
		b.WriteString(sTxt(cx+rd.dx, cy-25, 9.5, figMuted, rd.align, "400",
			fmt.Sprintf("%.0f ans", rd.n)))
	}

	// The two stretches, braced under the axis: the whole point of the plate
	// is that the second one is the cheaper.
	for _, br := range []struct {
		xa, xb float64
		label  string
	}{
		{x(30), x(50), fmt.Sprintf("30 → 50 ans : −%s point", frNum(amortGap(30, 50), 1))},
		{x(50), xInf, fmt.Sprintf("50 ans → toujours : −%s point", frNum(amortForever(50), 1))},
	} {
		b.WriteString(line(br.xa, 356, br.xb, 356, figMuted, 1.2))
		b.WriteString(line(br.xa, 351, br.xa, 356, figMuted, 1.2))
		b.WriteString(line(br.xb, 351, br.xb, 356, figMuted, 1.2))
		b.WriteString(sTxt((br.xa+br.xb)/2, 372, 10, figSoft, "middle", "600", br.label))
	}

	b.WriteString(sTxt(24, 400, 10.5, figSoft, "start", "600", fmt.Sprintf(
		"L'horizon se paie au début : de 10 à 30 ans le taux perd %s points, de 50 ans à l'éternité il en perd %s.",
		frNum(amortGap(10, 30), 1), frNum(amortForever(50), 1))))
	b.WriteString(sTxt(24, 416, 9.5, figMuted, "start", "400",
		"Arithmétique d'annuité, r / (1 − (1+r)^−n) à r = 4 % réel. Le capital finit à zéro le dernier jour, sans legs ni marge."))
	return svg(640, 432, b.String())
}
