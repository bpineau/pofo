package firebook

import (
	"fmt"
	"strings"
)

// The plate of the flexibility article. Its central mechanism ("the duration
// beats the depth") is stated in prose and nowhere shown, so the plate puts the
// two quantities on one pair of axes: how deep a household must cut to survive
// a hostile vintage, against how many years it must hold that cut.
//
// The plate mixes two kinds of object and must never let the reader confuse
// them. The descending curve is MEASURED: a minimum-depth solver run on the
// bundled real US 60/40 record. The shaded zone is EDITORIAL: it is this
// chapter's own written claim about what a household can hold (a 15 % cut for
// eighteen months with morale, five years with effort, never twelve years),
// which no simulation can establish. The legend says so, and so does the
// caption in the article.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.

// The minimum cut a household must accept to carry the 1966 vintage through a
// thirty-year plan, as a percentage of its standard of living, by the number of
// years the cut is held. Two plans on the same vintage: 45 000 EUR a year on
// 1 M EUR (4,5 %) and 40 000 EUR a year on the same capital (4,0 %). The only
// difference between the two panels is that half point of initial rate.
//
// The cut starts in 1970, the year after the household's own portfolio first
// sat more than 20 % below its running peak. That is the trigger this chapter
// writes down ("drawdown > 20 %"), read on the wealth the household actually
// watches. The answer barely depends on it: starting the cut in 1966, 1970 or
// 1975 moves the required depth by at most three points, which is the honest
// reason the plate can quote a single curve.
//
// Reproduce with pkg/replay and pkg/decumul. Take the real annual returns of
// the vintage, then bisect on the depth for each duration:
//
//	ref, _ := replay.Reference()
//	_, _, seq, _ := ref.Window(1966, 30)
//	// sched: 30 real spending multipliers, 1 everywhere except (1-depth)
//	// over the coupeTrigger-1966 .. +duration window.
//	p := decumul.Plan{Capital: 1e6, NeedAnnual: spend, Years: 30, SpendSchedule: sched}
//	survives := !p.RunPath(seq).Ruined
//
// figures_flex_test.go re-runs the whole solver and fails on any drift.
//
// The 4,5 % plan has no answer at all below four years: even spending NOTHING
// for one, two or three years leaves it short, because a plan that far behind
// needs more than three years of saved withdrawals to catch up. That is why its
// curve starts at four years, with the shorter durations drawn off the scale.
var (
	coupeVintage    = 1966
	coupeTrigger    = 1970 // first year of the cut
	coupeYears      = 30
	coupeCapital    = 1e6
	coupeSpend45    = 45000.0
	coupeSpend40    = 40000.0
	coupeImpossible = 3 // at 4,5 %, no cut of 1 to 3 years saves the plan, at any depth
	coupeFirst45    = 4 // the first duration with an answer, in years
	coupeDepth45    = []float64{
		83.0, 65.7, 51.8, 43.8, 38.6, 34.1, 30.4, 27.4, 25.1, 23.0, 21.5,
		20.4, 19.4, 18.6, 18.0, 17.4, 16.9, 16.6, 16.2, 15.9, 15.6, 15.3, 15.0,
	}
	coupeFirst40 = 1
	coupeDepth40 = []float64{
		42.5, 21.6, 14.9, 11.5, 9.1, 7.2, 6.1, 5.4, 4.7, 4.2, 3.8, 3.5, 3.2,
		3.0, 2.8, 2.7, 2.6, 2.5, 2.4, 2.4, 2.3, 2.2, 2.2, 2.2, 2.1, 2.1,
	}
)

// The editorial half of the plate, read straight off this chapter's key block:
// "Une coupe de 15 % se tient dix-huit mois avec du moral, cinq ans avec effort,
// jamais douze ans." Nothing here is measured, and the plate says so on its own
// surface. It is drawn as a zone rather than a curve precisely because the
// chapter only ever commits to one depth: a shallower cut held longer is
// neither claimed nor denied, and the plate must not invent the difference.
var (
	coupeTenableDepth = 15.0 // % of the standard of living
	coupeTenableEasy  = 5    // years, "cinq ans avec effort"
	coupeTenableWall  = 12   // years, "jamais douze ans"
)

// coupeEUR renders a whole euro amount the French way, with a space every three
// digits.
func coupeEUR(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte(' ')
		}
		out.WriteByte(s[i])
	}
	return out.String() + " €"
}

// figCoupeExigeeTenable draws the two quantities the chapter argues about, on
// one pair of axes: the depth of the cut and the number of years it lasts. The
// hostile vintage asks for a shallow cut, which is the intuitive half, but it
// asks for it over a span nobody holds, which is the half the prose has to
// assert. Here the geometry does it: the measured curve enters the tenable zone
// or it does not, and half a point of initial withdrawal rate decides which.
func figCoupeExigeeTenable() string {
	const (
		yBot, yTop = 340.0, 132.0
		vMax       = 118.0 // headroom above the 100 % line for the off-scale marks
		dMax       = 26.6
	)
	type panel struct {
		px0, px1 float64
		first    int
		depths   []float64
		title    string
		color    string
		verdict  string
	}
	rate := func(spend float64) string { return "Plan à " + frPct(spend/coupeCapital*100, 1) }
	panels := []panel{
		{62, 320, coupeFirst45, coupeDepth45, rate(coupeSpend45), figBad, "aucune coupe tenable ne sauve ce plan"},
		{358, 616, coupeFirst40, coupeDepth40, rate(coupeSpend40), figBlue, "la coupe tenable suffit dès trois ans"},
	}

	var b strings.Builder
	b.WriteString(plateHead("la flexibilité, mythe et réalité",
		"La coupe exigée par le marché contre la coupe tenable par un ménage"))
	b.WriteString(plateDeck(
		fmt.Sprintf("millésime %d, 60/40 américain réel, plan de trente ans", coupeVintage)))

	// The legend carries the plate's honesty: one measured object, one claim.
	b.WriteString(line(24, 77, 38, 77, figBad, 2.6))
	b.WriteString(line(42, 77, 56, 77, figBlue, 2.6))
	b.WriteString(sTxt(62, 81, 10, figSoft, "start", "400",
		"exigé par le marché : la coupe minimale qui évite la ruine (calculé, en % du train de vie)"))
	fmt.Fprintf(&b, `<rect x="24" y="90" width="32" height="11" fill="%s"/>`, figWash)
	b.WriteString(line(24, 90, 56, 90, figAccent, 1.6))
	b.WriteString(sTxt(62, 99, 10, figSoft, "start", "400",
		"tenable par un ménage : 15 % au plus, douze ans au plus (affirmation de ce chapitre, pas une mesure)"))

	for pi, p := range panels {
		m := mapper(0.4, dMax, 0, vMax, p.px0, p.px1, yBot, yTop)
		x := func(d float64) float64 { return m(d, 0)[0] }
		y := func(v float64) float64 { return m(0.4, v)[1] }

		b.WriteString(sTxt(p.px0, 116, 11.5, figInk, "start", "600", p.title))

		// The tenable zone, first, so every mark sits on top of it.
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			p.px0, y(coupeTenableDepth), x(float64(coupeTenableWall))-p.px0,
			y(0)-y(coupeTenableDepth), figWash)

		for _, g := range []float64{25, 50, 75} {
			b.WriteString(line(p.px0, y(g), p.px1, y(g), figGrid, 1))
			b.WriteString(mTxt(p.px0-8, y(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
		}
		b.WriteString(mTxt(p.px0-8, y(0)+3.5, 10, figMuted, "end", "400", "0"))
		b.WriteString(mTxt(p.px0-8, y(100)+3.5, 10, figMuted, "end", "400", "100"))
		b.WriteString(line(p.px0, y(0), p.px1, y(0), figRule, 1))

		// Spending nothing at all: the ceiling of the whole question.
		b.WriteString(dashLine(p.px0, y(100), p.px1, y(100), figRule, 1.2, "4 4"))

		// The zone's two edges: solid where the chapter says the cut holds,
		// dashed where it says it only holds with effort, then the wall.
		b.WriteString(line(p.px0, y(coupeTenableDepth), x(float64(coupeTenableEasy)),
			y(coupeTenableDepth), figAccent, 2.2))
		b.WriteString(dashLine(x(float64(coupeTenableEasy)), y(coupeTenableDepth),
			x(float64(coupeTenableWall)), y(coupeTenableDepth), figAccent, 1.5, "3 3"))
		b.WriteString(dashLine(x(float64(coupeTenableWall)), y(coupeTenableDepth),
			x(float64(coupeTenableWall)), y(0), figAccent, 1.6, "3 3"))

		// The measured curve.
		pts := make([][2]float64, len(p.depths))
		for i, d := range p.depths {
			pts[i] = m(float64(p.first+i), d)
		}
		b.WriteString(poly(pts, p.color, 2.4, ""))

		// Duration ticks: the chapter's own two numbers among them.
		for _, t := range []int{1, coupeTenableEasy, coupeTenableWall, 20, 26} {
			col := figMuted
			if t == coupeTenableEasy || t == coupeTenableWall {
				col = figDeep
			}
			b.WriteString(line(x(float64(t)), yBot, x(float64(t)), yBot+4, figRule, 1))
			b.WriteString(mTxt(x(float64(t)), 356, 10, col, "middle", "400", fmt.Sprintf("%d", t)))
		}

		if pi == 0 {
			b.WriteString(mTxt(p.px0+4, y(coupeTenableDepth)-6, 10, figDeep, "start", "600", frPct(coupeTenableDepth, 0)))
			b.WriteString(sTxt(p.px1-2, y(100)+14, 10, figMuted, "end", "400", "arrêt total des dépenses"))

			// The durations with no answer at all, drawn off the scale.
			for d := 1; d <= coupeImpossible; d++ {
				fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="none" stroke="%s" stroke-width="1.6"/>`,
					x(float64(d)), y(112), figBad)
			}
			b.WriteString(sTxt(x(5)+10, y(112)+3, 10.5, figBad, "start", "600", "1 à 3 ans : impossible,"))
			b.WriteString(sTxt(x(5)+10, y(112)+15, 10, figMuted, "start", "400", "même sans dépenser un euro"))

			// The three readings that carry the panel: the shortest cut that
			// exists at all, what the plan still asks at the household's wall,
			// and how long the cut must run to come back down to 15 %.
			b.WriteString(mTxt(x(float64(p.first))+10, y(p.depths[0])-4, 10.5,
				figBad, "start", "600", frPct(p.depths[0], 0)))
			last := len(p.depths) - 1
			wall := coupeTenableWall - p.first
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`,
				x(float64(coupeTenableWall)), y(p.depths[wall]), figBad)
			b.WriteString(mTxt(x(float64(coupeTenableWall)), y(p.depths[wall])-8, 10.5,
				figBad, "middle", "600", frPct(p.depths[wall], 0)))
			b.WriteString(sTxt(p.px1-6, y(45), 10.5, figSoft, "end", "600",
				"pour ne demander que "+frPct(p.depths[last], 0)+","))
			b.WriteString(sTxt(p.px1-6, y(37), 10, figMuted, "end", "400", "il faut tenir vingt-six ans"))
			b.WriteString(dashLine(p.px1-6, y(32), p.px1-6, y(19), figMuted, 1, "2 3"))
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`,
				x(26), y(p.depths[last]), figBad)
		} else {
			// The crossing: the year the requirement drops into the zone.
			cross := 3
			cd := p.depths[cross-p.first]
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, x(float64(cross)), y(cd), figBlue)
			b.WriteString(sTxt(x(5.2), y(86), 10.5, figSoft, "start", "600", "l'exigence entre dans le tenable"))
			b.WriteString(sTxt(x(5.2), y(76), 10, figMuted, "start", "400",
				fmt.Sprintf("un an : %s. Trois ans : %s.", frPct(p.depths[0], 0), frPct(p.depths[2], 0))))
			b.WriteString(sTxt(x(5.2), y(66), 10, figMuted, "start", "400",
				fmt.Sprintf("Cinq ans : %s. Dix ans : %s.", frPct(p.depths[4], 0), frPct(p.depths[9], 0))))
			b.WriteString(dashLine(x(5.0), y(60), x(3.3), y(20), figMuted, 1, "2 3"))
		}

		b.WriteString(sTxt((p.px0+p.px1)/2, 398, 10.5, p.color, "middle", "600", p.verdict))
	}

	b.WriteString(sTxt(320, 374, 10, figMuted, "middle", "400",
		fmt.Sprintf("durée de la coupe, en années ; elle démarre en %d, l'année où le portefeuille passe 20 %% sous son sommet", coupeTrigger)))
	b.WriteString(sTxt(24, 420, 9.5, figMuted, "start", "400",
		fmt.Sprintf("Plan de %d ans sur 1 M€ de capital : %s par an à gauche, %s à droite, en euros constants, hors fiscalité.",
			coupeYears, coupeEUR(coupeSpend45), coupeEUR(coupeSpend40))))
	b.WriteString(sTxt(24, 434, 9.5, figMuted, "start", "400",
		"60/40 américain réel (S&amp;P 500, Treasuries 5 ans, déflatés CPI-U), reconstruction du livre."))
	return svg(640, 446, b.String())
}
