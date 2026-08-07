package firebook

import (
	"fmt"
	"strings"
)

// This file holds the timeline plate of the enveloppes-francaises article. The
// two French fiscal clocks (5 years for a PEA, 8 years for an assurance-vie)
// run from the day the envelope is opened and never from the day it is funded,
// so the only thing that separates the article's two households is a date. The
// plate puts both on one age axis. Everything it states about a clock (the age
// it starts, the age it matures, its state on the departure day, how many taps
// flow then) is derived from the named constants below, and the guard test
// recomputes those states and checks that the article still quotes the same
// maturities, the same ages and the same rates.

// The legal maturities, as the article states them: "Après 5 ans, les retraits
// sont exonérés d'impôt sur le revenu" for the PEA (art. 150-0 A CGI), "Après
// 8 ans, un abattement annuel s'applique" for the assurance-vie (art. 125-0 A
// CGI). Both count from the opening date of the plan or of the contract, which
// is the whole point of the plate.
const (
	peaClockYears = 5
	avClockYears  = 8
)

// The ages come from the article's own worked example and its own prose: the
// couple leaves at 58 ("Prenons un couple de 58 ans"), the late household
// discovers the subject "à deux ans du départ", and the prepared one opens as
// the introduction says a rentier should ("le rentier de 50 ans hérite des
// clics du trentenaire"). The gain share is the example's own, "un patrimoine
// de 1,7 M€ gainé à 45 % environ".
const (
	envDepartureAge = 58
	envLateLead     = 2
	envEarlyOpenAge = 30
	envLateOpenAge  = envDepartureAge - envLateLead
	envGainShare    = 0.45
	envAxisFirstAge = 25
	envAxisLastAge  = 66
)

// envMatureAge is the age at which a clock started at openAge has run its
// course.
func envMatureAge(openAge, clockYears int) int { return openAge + clockYears }

// envClockState returns, for a clock opened at openAge, the age at which it
// matures, whether it has matured by the departure age, and the number of years
// that separates the departure from the maturity (years of ripeness when it has
// matured, years still missing when it has not).
func envClockState(openAge, clockYears int) (matureAge int, mature bool, years int) {
	matureAge = envMatureAge(openAge, clockYears)
	if matureAge <= envDepartureAge {
		return matureAge, true, envDepartureAge - matureAge
	}
	return matureAge, false, matureAge - envDepartureAge
}

// envTapsFlowing counts the envelopes a household can actually draw on at the
// departure age: every matured clock, plus the CTO, which has none.
func envTapsFlowing(openAge int) int {
	n := 1
	for _, l := range envLanes {
		if _, mature, _ := envClockState(openAge, l.years); mature {
			n++
		}
	}
	return n
}

// envLane describes one clocked envelope: its gutter label, the French forms
// the sentences need (an assurance-vie comes in pairs, so it agrees in the
// plural), its legal maturity and its colour, kept identical to the
// frictions-enveloppes plate of the same article.
type envLane struct {
	gutter  string // left gutter label, with the legal maturity
	opened  string // "ouvert" / "ouvertes"
	ripe    string // "mûr" / "mûres"
	since   string // subject of "... depuis N ans", when the clock has rung
	missing string // object of "il manque N ans ...", when it has not
	years   int
	col     string
}

var envLanes = []envLane{
	{"PEA (5 ans)", "ouvert", "mûr", "le PEA est mûr", "au PEA", peaClockYears, figAccent},
	{"assurances-vie (8 ans)", "ouvertes", "mûres", "les assurances-vie", "aux assurances-vie", avClockYears, figBlue},
}

// envTaps phrases the tap count, agreeing in number.
func envTaps(n, total int) string {
	if n <= 1 {
		return fmt.Sprintf("%d robinet sur %d coule", n, total)
	}
	return fmt.Sprintf("%d robinets sur %d coulent", n, total)
}

// euroFR formats a whole number of euros the French way, "9 200 €".
func euroFR(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String() + " €"
}

// figHorlogesEnveloppes draws the two clocks on two tracks: the household that
// opens at 30 and the one that discovers the subject two years before leaving.
// Same departure day, same portfolio, three taps against one.
func figHorlogesEnveloppes() string {
	const (
		xLeft, xRight = 176.0, 600.0
		laneH         = 16.0
		lanePitch     = 26.0
		gutterX       = 166.0
		yAxis         = 388.0
	)
	px := func(age float64) float64 {
		return xLeft + (age-envAxisFirstAge)/(envAxisLastAge-envAxisFirstAge)*(xRight-xLeft)
	}
	rect := func(b *strings.Builder, x0, x1, y, h float64, fill string) {
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="2" fill="%s"/>`,
			x0, y, x1-x0, h, fill)
	}
	// tap draws the state of one envelope on the departure line: a filled disc
	// when the household can draw on it that day, an empty ring when it cannot.
	tap := func(b *strings.Builder, y float64, col string, flowing bool) {
		if flowing {
			fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="4.6" fill="%s" stroke="#fffdf9" stroke-width="1.8"/>`,
				px(envDepartureAge), y, col)
			return
		}
		fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="#fffdf9" stroke="%s" stroke-width="1.8"/>`,
			px(envDepartureAge), y, figMuted)
	}

	var b strings.Builder
	b.WriteString(plateHead("les horloges fiscales",
		fmt.Sprintf("Deux dates d'ouverture, un même départ à %d ans", envDepartureAge)))

	// The departure, marked once per track, on the same abscissa. The line is
	// cut between the two blocks rather than run through the plate, so it never
	// strikes through the sentences that read the tracks.
	for _, seg := range [][2]float64{{74, 180}, {236, 336}} {
		b.WriteString(dashLine(px(envDepartureAge), seg[0], px(envDepartureAge), seg[1], figDeep, 1.4, "4 4"))
	}
	b.WriteString(sTxt(px(envDepartureAge), 66, 11, figDeep, "middle", "600",
		fmt.Sprintf("départ à %d ans", envDepartureAge)))
	// how to read a lane, once, in the band the departure label leaves free
	b.WriteString(sTxt(24, 66, 10.5, figMuted, "start", "400",
		"En pâle, l'horloge tourne ; en couleur, le robinet coule."))

	// track draws one household: its two clocks, then the CTO, then the verdict
	// of the departure day. Every number below is read off envClockState.
	track := func(y0 float64, openAge int, title string, tail string) {
		b.WriteString(sTxt(24, y0, 11.5, figSoft, "start", "600", title))
		var ripe, missing []string
		for i, l := range envLanes {
			y := y0 + 12 + lanePitch*float64(i)
			mid := y + laneH/2 + 3.5
			matureAge, mature, years := envClockState(openAge, l.years)
			b.WriteString(sTxt(gutterX, mid, 10.5, figSoft, "end", "600", l.gutter))
			// the clock running, then the envelope mature and flowing for good
			rect(&b, px(float64(openAge)), px(float64(matureAge)), y, laneH, figWash)
			rect(&b, px(float64(matureAge)), xRight, y, laneH, l.col)
			if mature {
				b.WriteString(mTxt((px(float64(openAge))+px(float64(matureAge)))/2, mid, 10,
					figMuted, "middle", "400", fmt.Sprintf("%d ans", l.years)))
				b.WriteString(sTxt(px(float64(matureAge))+10, mid, 10.5, "#fffdf9", "start", "600",
					fmt.Sprintf("%s à %d ans, et pour toujours", l.ripe, matureAge)))
				ripe = append(ripe, fmt.Sprintf("%s depuis %d ans", l.since, years))
			} else {
				// A clock opened that late still runs across the departure line,
				// so its length is stated in the note rather than inside a
				// segment the departure marker already occupies.
				b.WriteString(sTxt(px(float64(openAge))-10, mid, 10.5, figMuted, "end", "400",
					fmt.Sprintf("%s à %d ans, %s à %d ans : %d ans trop tard",
						l.opened, openAge, l.ripe, matureAge, years)))
				missing = append(missing, fmt.Sprintf("%d ans %s", years, l.missing))
			}
			tap(&b, y+laneH/2, l.col, mature)
		}
		// the CTO: no clock, open from day one, taxed at the full rate
		y := y0 + 12 + lanePitch*float64(len(envLanes))
		mid := y + laneH/2 + 3.5
		b.WriteString(sTxt(gutterX, mid, 10.5, figSoft, "end", "600", "CTO (aucune horloge)"))
		rect(&b, px(envAxisFirstAge), xRight, y, laneH, figBad)
		b.WriteString(sTxt(px(envAxisFirstAge)+10, mid, 10.5, "#fffdf9", "start", "600",
			"ouvert et taxé au taux plein dès le premier jour"))
		tap(&b, y+laneH/2, figBad, true)

		// the verdict, in words, at the departure age
		verdict := "Au départ, " + envTaps(envTapsFlowing(openAge), len(envLanes)+1) + " : "
		if len(missing) > 0 {
			verdict += "il manque " + strings.Join(missing, " et ") + "."
		} else {
			verdict += strings.Join(ripe, ", ") + "."
		}
		b.WriteString(sTxt(24, y0+104, 11, figSoft, "start", "600", verdict))
		b.WriteString(sTxt(24, y0+119, 10.5, figMuted, "start", "400", tail))
	}

	track(92, envEarlyOpenAge,
		fmt.Sprintf("Il ouvre un PEA et deux assurances-vie à %d ans, même à 100 €", envEarlyOpenAge),
		"Le flux se répartit : "+pctFR(peaSlope)+" sur la part de gain retirée du PEA, abattement de "+
			euroFR(avAllowanceCouple)+" consommé chaque année.")
	b.WriteString(line(24, 224, 616, 224, figGrid, 1))
	track(248, envLateOpenAge,
		fmt.Sprintf("Il découvre le sujet à %d ans, deux ans avant le départ", envLateOpenAge),
		"Tout sort du CTO, au taux plein de "+pctFR(ctoSlope)+", soit "+pctFR(envGainShare*ctoSlope)+
			" de friction sur le flux retiré.")

	// the age axis, shared by both tracks
	b.WriteString(line(xLeft, yAxis, xRight, yAxis, figRule, 1))
	for age := envAxisFirstAge; age <= 65; age += 5 {
		b.WriteString(line(px(float64(age)), yAxis, px(float64(age)), yAxis+4, figRule, 1))
		b.WriteString(mTxt(px(float64(age)), yAxis+16, 10, figMuted, "middle", "400", fmt.Sprintf("%d", age)))
	}
	b.WriteString(sTxt((xLeft+xRight)/2, yAxis+34, 11, figMuted, "middle", "400", "âge du titulaire"))

	// the plate must be readable alone: the two legal durations and the status
	// of the example it draws
	b.WriteString(sTxt(24, 446, 9.5, figMuted, "start", "400",
		fmt.Sprintf("Durées légales : %d ans pour le PEA, %d ans pour l'assurance-vie, comptées depuis la date d'ouverture et non depuis les versements.",
			peaClockYears, avClockYears)))
	b.WriteString(sTxt(24, 459, 9.5, figMuted, "start", "400",
		fmt.Sprintf("Frise illustrative de l'exemple de l'article (départ à %d ans, patrimoine gainé à 45 %%). Taux au 1er janvier 2026.", envDepartureAge)))
	return svg(640, 472, b.String())
}
