package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The three-phases plate: the single variable the article says governs the
// whole plan, the year's net flow divided by the portfolio it lands in.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed.
//
// Nothing here is a measurement or a simulation. It is a deterministic
// projection of the article's own example (Inès), computed at render time from
// the constants below, with no market noise at all. The plate says so twice,
// in its subtitle and in its caption.
const (
	// The article's numbers. Inès saves 34 % of her income and holds 90 %
	// equities while accumulating, then de-risks to 55 % at the transition.
	// The two real returns are the book's own anchors for those two mixes:
	// 5 % for an equity-heavy portfolio, 4 % for the 60/40 (see the
	// "cascade-4pct" plate). The target is 25 years of spending, the 4 % rule.
	phaseSavingsRate  = 0.34
	phaseAccumReturn  = 0.05
	phaseRetireReturn = 0.04
	phaseTargetYears  = 25.0

	// The ages. The article dates the departure (Inès is 47 and "partie depuis
	// deux ans") and the transition band ("cible atteinte moins deux ans" to
	// "départ plus trois ans"). phaseStartAge is the only derived age: reaching
	// 25 years of spending at 34 % saved and 5 % real takes
	// ln(1+r*target/rate)/ln(1+r) = 25,2 years, so a departure at 45 means a
	// first serious saving year at 20. The projection then reproduces the
	// article's own checkpoint without being told it: 92 % of the target at 44,
	// where the article writes "à 90 % de sa cible".
	phaseStartAge    = 20.0
	phaseDepartAge   = 45.0
	phaseTransFrom   = 43.0
	phaseTransTo     = 48.0
	phaseLastAge     = 85.0
	phaseFirstFigAge = 22.0 // first year the portfolio holds a year of spending

	// The y axis is cut at +15 %. Below the first few years of saving the
	// ratio divides the year's contribution by an almost empty portfolio, so it
	// diverges (100 % after one year, 49 % after two) and its left end says
	// nothing about the plan: it is an artefact of the starting balance, not a
	// fact about saving. The plate draws the curve only where it is meaningful
	// and states the cut on the face of the chart rather than pretending to an
	// asymptote.
	phaseAxisTop = 15.0
	phaseAxisBot = -5.5
)

// phaseRedBand is the transition band's fill: figBad (192,101,91) at .18,
// PRE-BLENDED onto the figure card background #fffdf9 (never rgba, crengine
// paints those solid black). It runs deeper than the book's usual red wash
// because the band is only five years wide and has to hold its own between two
// tinted neighbours.
const phaseRedBand = "#F4E2DD"

// phaseYear is one year of the projection: the age, the portfolio at that age
// (in years of gross income) and the year's net flow as a percentage of it.
type phaseYear struct {
	age   float64
	value float64
	ratio float64 // percent, positive while paying in, negative once drawing
}

// phaseProjection runs the deterministic plan year by year. It returns the two
// branches, which share their first portfolio value: at the departure age the
// same portfolio carries a last contribution and a first withdrawal, which is
// exactly the discontinuity the plate is about.
func phaseProjection() (accum, retire []phaseYear) {
	spend := 1 - phaseSavingsRate
	w := 0.0
	for a := phaseStartAge + 1; a <= phaseDepartAge; a++ {
		w = w*(1+phaseAccumReturn) + phaseSavingsRate
		accum = append(accum, phaseYear{a, w, phaseSavingsRate / w * 100})
	}
	for a := phaseDepartAge; a <= phaseLastAge; a++ {
		if a > phaseDepartAge {
			w = (w - spend) * (1 + phaseRetireReturn)
		}
		retire = append(retire, phaseYear{a, w, -spend / w * 100})
	}
	return accum, retire
}

// phaseAccumRatio is the accumulation branch in closed form: the portfolio
// after k yearly contributions is rate*((1+r)^k-1)/r, so the ratio is
// r/((1+r)^k-1), independent of the contribution's size. It agrees with the
// recurrence above at every integer age (the guard test checks it) and lets the
// plate sample the steep early years at a quarter of a year.
func phaseAccumRatio(age float64) float64 {
	k := age - phaseStartAge
	return phaseAccumReturn / (math.Pow(1+phaseAccumReturn, k) - 1) * 100
}

// phaseAgeAtRatio inverts phaseAccumRatio: the age at which the accumulation
// ratio falls to pct. Used to enter the plot exactly on the cut axis.
func phaseAgeAtRatio(pct float64) float64 {
	return phaseStartAge + math.Log(1+phaseAccumReturn*100/pct)/math.Log(1+phaseAccumReturn)
}

// figFluxRelatifPhases draws that single ratio from the first meaningful year
// to 85, over the three phase bands. The reading is the article's own thesis:
// the flow shrinks against the portfolio for twenty-five years, reaches its
// smallest value of the whole plan the year of the departure, changes sign
// there, and never again weighs more than a few percent. Past the transition
// no flow can repair an accident, because there is no flow left worth the name.
func figFluxRelatifPhases() string {
	const (
		x0, x1     = 62.0, 616.0
		yTop, yBot = 122.0, 338.0
		ageLo      = 25.0
	)
	x := func(age float64) float64 { return x0 + (age-ageLo)/(phaseLastAge-ageLo)*(x1-x0) }
	y := func(v float64) float64 {
		return yTop + (phaseAxisTop-v)/(phaseAxisTop-phaseAxisBot)*(yBot-yTop)
	}

	var b strings.Builder
	b.WriteString(plateHead("les trois phases",
		"Le flux de l'année, rapporté au portefeuille où il tombe"))
	b.WriteString(sTxt(24, 62, 11, figSoft, "start", "400",
		"Une seule variable gouverne les trois phases : ce que vous versez ou prélevez dans l'année, en % du portefeuille."))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		"Projection déterministe sur les hypothèses de l'article, pas une simulation : 34 % d'épargne, 5 % de rendement réel"))
	b.WriteString(sTxt(24, 92, 10.5, figMuted, "start", "400",
		"jusqu'au départ à 45 ans, puis un portefeuille dé-risqué à 4 % réel. Ni krach, ni bonne surprise."))

	// The three phase bands, the background of the plate. Pre-blended solid
	// tints (never rgba, crengine paints those black): the accumulation keeps
	// the accent hue of the paying-in branch, the transition the red of the
	// article's "zone rouge", the retirement the blue of the drawing branch.
	band := func(from, to float64, fill string) {
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			x(from), yTop, x(to)-x(from), yBot-yTop, fill)
	}
	band(ageLo, phaseTransFrom, figBandOuter)
	band(phaseTransFrom, phaseTransTo, phaseRedBand)
	band(phaseTransTo, phaseLastAge, figBlueWash)
	for _, a := range []float64{phaseTransFrom, phaseTransTo} {
		b.WriteString(line(x(a), yTop, x(a), yBot, figRule, 1))
	}
	for _, s := range []struct {
		at    float64
		color string
		label string
	}{
		{(ageLo + phaseTransFrom) / 2, figDeep, "accumulation"},
		{(phaseTransFrom + phaseTransTo) / 2, figBad, "transition"},
		{(phaseTransTo + phaseLastAge) / 2, figBlue, "retrait"},
	} {
		b.WriteString(sTxt(x(s.at), 112, 9.5, s.color, "middle", "600",
			`<tspan letter-spacing="1.6">`+strings.ToUpper(s.label)+`</tspan>`))
	}

	// Grid, zero rule and the cut at the top of the axis.
	for _, g := range []float64{15, 10, 5, 0, -5} {
		gy := y(g)
		switch {
		case g == 0:
			b.WriteString(line(x0, gy, x1, gy, figRule, 1.4))
		case g == phaseAxisTop:
			b.WriteString(dashLine(x0, gy, x1, gy, figMuted, 1, "2 4"))
		default:
			b.WriteString(line(x0, gy, x1, gy, figGrid, 1))
		}
		lbl := "0"
		if g > 0 {
			lbl = fmt.Sprintf("+%.0f %%", g)
		} else if g < 0 {
			lbl = fmt.Sprintf("−%.0f %%", -g)
		}
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", lbl))
	}
	b.WriteString(sTxt(x1, 142, 10, figMuted, "end", "400",
		fmt.Sprintf("axe coupé à +%.0f %% : le ratio vaut %.0f %% à %.0f ans, et diverge au premier euro versé",
			phaseAxisTop, phaseAccumRatio(phaseFirstFigAge), phaseFirstFigAge)))

	// What the sign means, once, on the axis itself.
	b.WriteString(sTxt(x0+8, y(0)-9, 10, figAccent, "start", "600", "vous versez"))
	b.WriteString(sTxt(x0+8, y(0)+18, 10, figBlue, "start", "600", "vous prélevez"))

	accum, retire := phaseProjection()

	// The paying-in branch, entered exactly on the cut and sampled every
	// quarter so the steep early years read as a curve, not as a polygon.
	var up [][2]float64
	for a := phaseAgeAtRatio(phaseAxisTop); a < phaseDepartAge; a += 0.25 {
		up = append(up, [2]float64{x(a), y(phaseAccumRatio(a))})
	}
	last := accum[len(accum)-1]
	up = append(up, [2]float64{x(last.age), y(last.ratio)})
	b.WriteString(poly(up, figAccent, 2.4, ""))

	// The day of the departure: the same portfolio, the opposite flow. The
	// vertical is the curve itself, split at zero so each half keeps its phase
	// colour, with the crossing marked as the event it is.
	zx, zy := x(phaseDepartAge), y(0)
	b.WriteString(line(zx, y(last.ratio), zx, zy, figAccent, 2.4))
	b.WriteString(line(zx, zy, zx, y(retire[0].ratio), figBlue, 2.4))

	// The drawing branch, one point per year.
	var down [][2]float64
	for _, p := range retire {
		down = append(down, [2]float64{x(p.age), y(p.ratio)})
	}
	b.WriteString(poly(down, figBlue, 2.4, ""))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`,
		zx, zy, figDeep)

	// Direct labels. The thirty-year-old first: the contribution still repairs
	// anything the market does.
	var at30 phaseYear
	for _, p := range accum {
		if p.age == 30 {
			at30 = p
		}
	}
	p30 := [2]float64{x(at30.age), y(at30.ratio)}
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s" stroke="#fffdf9" stroke-width="1.4"/>`,
		p30[0], p30[1], figDeep)
	b.WriteString(line(p30[0]+6, p30[1]-4, 170, 180, figMuted, 1))
	b.WriteString(sTxt(176, 168, 11, figSoft, "start", "600",
		fmt.Sprintf("à 30 ans, le versement de l'année vaut encore %s du capital :",
			frPct(at30.ratio, 0))))
	b.WriteString(sTxt(176, 182, 10.5, figMuted, "start", "400",
		"un krach est effacé par les versements suivants."))

	// The thesis, in the empty quarter of the plate.
	b.WriteString(sTxt(316, 212, 11, figSoft, "start", "600",
		"À la transition, le portefeuille n'a jamais été"))
	b.WriteString(sTxt(316, 226, 11, figSoft, "start", "600",
		"aussi gros par rapport au flux qui pourrait le réparer."))

	// The crossing, and the far end of the drawing branch.
	b.WriteString(sTxt(zx+10, 292, 11, figDeep, "start", "600",
		fmt.Sprintf("%.0f ans : le flux change de signe,", phaseDepartAge)))
	b.WriteString(sTxt(zx+10, 306, 10.5, figMuted, "start", "400",
		fmt.Sprintf("de +%s à %s du portefeuille, du jour au lendemain.",
			frPct(last.ratio, 1), frPct(retire[0].ratio, 1))))
	end := retire[len(retire)-1]
	b.WriteString(mTxt(x1, 318, 10.5, figBlue, "end", "600", frPct(end.ratio, 1)))
	b.WriteString(sTxt(x1-46, 318, 10.5, figMuted, "end", "400",
		fmt.Sprintf("à %.0f ans :", end.age)))

	// The age axis, with the departure called out among the round decades.
	b.WriteString(line(x0, yBot, x1, yBot, figRule, 1))
	for _, a := range []float64{30, 40, 50, 60, 70, 80} {
		b.WriteString(mTxt(x(a), yBot+16, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", a)))
	}
	b.WriteString(mTxt(zx, yBot+16, 10, figDeep, "middle", "600", fmt.Sprintf("%.0f", phaseDepartAge)))
	b.WriteString(sTxt((x0+x1)/2, yBot+36, 11, figMuted, "middle", "400",
		"âge, de la première épargne sérieuse au grand âge"))
	return svg(640, 392, b.String())
}

// frPct formats a signed percentage the French way: a comma decimal separator,
// a true minus sign, a space before the unit.
func frPct(v float64, dec int) string {
	s := fmt.Sprintf("%.*f", dec, math.Abs(v))
	s = strings.Replace(s, ".", ",", 1)
	if v < 0 {
		s = "−" + s
	}
	return s + " %"
}
