package firebook

import (
	"fmt"
	"strings"
)

// The warning time. A simulator reports ruin as a cliff: the balance crosses
// zero and the story ends there, with no date attached to the moment the plan
// became recognizably doomed. The two dates are what this plate puts on one
// timeline, on the worst start year the record holds. Above, the capital of the
// 1966 vintage under a fixed 4 % inflation-indexed withdrawal. Below, the same
// plan's current withdrawal rate, drawn as the staircase it is, over the
// green / amber / red bands the book's dashboard article sets. The red light
// comes on in 1974 and is confirmed in 1975; the money runs out in 1994. The
// distance between those two facts is the plate's subject.
//
// The form is deliberately not that of retrait-deux-lectures, which reads the
// same vintage: there the question is what one rigid cheque looks like from two
// seats, drawn as bars against a curve. Here the question is how much notice a
// failing plan gives, so the upper register is a filled capital path and the
// lower one a staircase inside coloured bands, tied together by two dated rules
// and one brace that measures the gap.
//
// The numbers are frozen so the plate stays a pure function;
// figures_preavis_test.go re-runs the replay and recomputes all of them under
// the figure-drift gate.

// The replay behind both registers: 1966, thirty years, 1 M of capital and a
// 40 k cheque indexed to inflation, everything real.
const (
	preavisStart    = 1966
	preavisCapital0 = 1000.0
	preavisCheque   = 40.0
	preavisRuinYear = 1994 // the year the money ran out
)

// The dashboard's scale, quoted from quand-s-inquieter: green under 4,3 %,
// amber from 4,3 to 5,2 %, red above 5,2 % once confirmed on two spaced
// readings. They are the article's numbers, not the plate's.
const (
	preavisGreenMax  = 4.3
	preavisOrangeMax = 5.2
)

// preavisCeiling is where the lower register stops. The rate ends at 100 %, so
// a panel that held it whole would flatten the only part that matters: the
// crossing of the two thresholds. The staircase is drawn while it is on the
// scale and the rest is stated in words, at the top of the panel.
const preavisCeiling = 9.0

// preavisCapital is the capital the plan started each year with, in k of
// constant money, from 1966 to 1995: the denominator of that year's withdrawal
// rate, and the path the upper register draws.
var preavisCapital = []float64{
	1000.0, 885.8, 939.2, 932.2, 795.0, 782.1, 801.0, 829.3, 669.5, 486.1,
	521.4, 549.7, 458.0, 400.1, 356.3, 338.3, 281.7, 289.5, 277.4, 251.2,
	258.8, 254.1, 212.6, 185.4, 172.6, 127.5, 105.8, 68.4, 30.4, 0.0,
}

// preavisRate is the current withdrawal rate of each year from 1966 to 1994,
// in percent: the fixed cheque divided by the capital it was taken from. The
// last year of preavisCapital has no rate, there being nothing left to divide.
var preavisRate = []float64{
	4.00, 4.52, 4.26, 4.29, 5.03, 5.11, 4.99, 4.82, 5.97, 8.23,
	7.67, 7.28, 8.73, 10.00, 11.23, 11.82, 14.20, 13.82, 14.42, 15.92,
	15.46, 15.74, 18.82, 21.57, 23.17, 31.37, 37.80, 58.49, 100.00,
}

// preavisRateAt and preavisCapitalAt read the two series by calendar year.
func preavisRateAt(year int) float64 {
	i := year - preavisStart
	if i < 0 || i >= len(preavisRate) {
		return 0
	}
	return preavisRate[i]
}

func preavisCapitalAt(year int) float64 {
	i := year - preavisStart
	if i < 0 || i >= len(preavisCapital) {
		return 0
	}
	return preavisCapital[i]
}

// preavisFirstRed is the first year whose rate reads red, and
// preavisConfirmedRed the year that confirms it on the second spaced reading
// the article's hysteresis demands. Both are computed, so a refreshed replay
// moves the plate's markers with the data.
func preavisFirstRed() int {
	for i, v := range preavisRate {
		if v > preavisOrangeMax {
			return preavisStart + i
		}
	}
	return 0
}

func preavisConfirmedRed() int {
	for i := 1; i < len(preavisRate); i++ {
		if preavisRate[i] > preavisOrangeMax && preavisRate[i-1] > preavisOrangeMax {
			return preavisStart + i
		}
	}
	return 0
}

// preavisNotice is the plate's headline: the years between the first red
// reading and the year the money ran out.
func preavisNotice() int { return preavisRuinYear - preavisFirstRed() }

// preavisExit is the first year whose rate leaves the lower panel.
func preavisExit() int {
	for i, v := range preavisRate {
		if v > preavisCeiling {
			return preavisStart + i
		}
	}
	return preavisStart + len(preavisRate)
}

// figPreavis1966 draws the capital above, the warning lights below, and the
// distance between the two dates on the axis.
func figPreavis1966() string {
	const (
		x0, x1         = 76.0, 604.0
		firstYear      = preavisStart
		lastYear       = preavisStart + 29 // the start of 1995, where the path ends
		upTop, upBase  = 98.0, 202.0
		upMax          = 1000.0
		lowTop, lowBot = 262.0, 394.0
		card           = "#fffdf9" // the figure card the washes are blended onto
	)
	slot := (x1 - x0) / float64(lastYear-firstYear)
	xY := func(year int) float64 { return x0 + slot*float64(year-firstYear) }
	yUp := func(v float64) float64 { return upBase - v/upMax*(upBase-upTop) }
	yLow := func(v float64) float64 { return lowBot - v/preavisCeiling*(lowBot-lowTop) }

	var b strings.Builder
	b.WriteString(plateHead("le préavis",
		"Le millésime qui échoue prévient vingt ans à l'avance"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Millésime 1966, le pire connu : le capital au-dessus, le taux de retrait courant et ses voyants au-dessous"))

	// The upper register: the capital, in thousands of constant euros.
	b.WriteString(sTxt(24, 86, 11, figInk, "start", "600", "Le capital réel"))
	b.WriteString(sTxt(140, 86, 10, figMuted, "start", "400",
		"en milliers d'euros constants, capital de début d'année"))
	for i, v := range []float64{0, 250, 500, 750, 1000} {
		b.WriteString(line(x0, yUp(v), x1, yUp(v), figGrid, 1))
		b.WriteString(mTxt(x0-10, yUp(v)+3.5, 10, figMuted, "end", "400",
			[]string{"0", "250", "500", "750", "1 000"}[i]))
	}

	// The lower register: the dashboard, banded with the article's own scale.
	b.WriteString(sTxt(24, 226, 11, figInk, "start", "600", "Le voyant"))
	b.WriteString(sTxt(140, 226, 10, figMuted, "start", "400",
		"le retrait de l'année divisé par le capital du moment"))
	legendChips(&b, 238, [][2]string{
		{figGood, "vert : sous 4,3 %"},
		{figAccent, "orange : 4,3 à 5,2 %"},
		{figBad, "rouge : au-delà de 5,2 %, confirmé"},
	})
	band := func(lo, hi float64, col string, t float64) {
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			x0, yLow(hi), x1-x0, yLow(lo)-yLow(hi), mixHex(card, col, t))
	}
	band(0, preavisGreenMax, figGood, 0.17)
	band(preavisGreenMax, preavisOrangeMax, figAccent, 0.18)
	band(preavisOrangeMax, preavisCeiling, figBad, 0.16)
	for _, th := range []struct {
		v   float64
		col string
	}{{preavisGreenMax, figAccent}, {preavisOrangeMax, figBad}} {
		b.WriteString(line(x0, yLow(th.v), x1, yLow(th.v), mixHex(card, th.col, 0.55), 1))
		b.WriteString(mTxt(x0-10, yLow(th.v)+3.5, 10, figMuted, "end", "400",
			frNum(th.v, 1)+" %"))
	}
	b.WriteString(mTxt(x0-10, lowBot+3.5, 10, figMuted, "end", "400", "0"))
	b.WriteString(mTxt(x0-10, yLow(preavisCeiling)+3.5, 10, figMuted, "end", "400",
		frNum(preavisCeiling, 0)+" %"))

	// The two dates the plate exists to separate, ruled through both registers.
	for _, yr := range []int{preavisFirstRed(), preavisRuinYear} {
		b.WriteString(dashLine(xY(yr), upTop, xY(yr), lowBot, figMuted, 1, "3 4"))
	}

	// The capital path, filled: one point per year, from the opening million to
	// the empty account.
	var pts [][2]float64
	for i, v := range preavisCapital {
		pts = append(pts, [2]float64{xY(firstYear + i), yUp(v)})
	}
	var d strings.Builder
	for i, p := range pts {
		verb := "L"
		if i == 0 {
			verb = "M"
		}
		fmt.Fprintf(&d, "%s %.1f,%.1f ", verb, p[0], p[1])
	}
	fmt.Fprintf(&d, "L %.1f,%.1f L %.1f,%.1f Z", x1, upBase, x0, upBase)
	fmt.Fprintf(&b, `<path d="%s" fill="%s"/>`, d.String(), mixHex(card, figDeep, 0.14))
	b.WriteString(poly(pts, figDeep, 2.4, ""))
	b.WriteString(line(x0, upBase, x1, upBase, figRule, 1))

	// What was still in the account the day the light went red, which is the
	// reading that makes the notice worth having.
	red := preavisFirstRed()
	b.WriteString(dotLabel(xY(red), yUp(preavisCapitalAt(red)), figBad,
		fmt.Sprintf("premier rouge, %d", red),
		frNum(preavisCapitalAt(red), 0)+" k€", "start"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, x1, yUp(0), figBad)
	b.WriteString(sTxt(x1, 168, 10, figSoft, "end", "600", "épuisement"))
	b.WriteString(mTxt(x1, 181, 10.5, figBad, "end", "600",
		fmt.Sprintf("%d", preavisRuinYear)))

	// The staircase, one tread per year, drawn while it is on the scale.
	exit := preavisExit()
	var steps [][2]float64
	for i, v := range preavisRate {
		year := preavisStart + i
		if year == exit {
			steps = append(steps, [2]float64{xY(year), yLow(preavisCeiling)})
			break
		}
		steps = append(steps, [2]float64{xY(year), yLow(v)}, [2]float64{xY(year + 1), yLow(v)})
	}
	b.WriteString(poly(steps, figInk, 2.2, ""))
	// Beyond the panel the rate keeps climbing and never returns: a rail along
	// the top says so without pretending to a value.
	b.WriteString(dashLine(xY(exit), lowTop, x1, lowTop, figBad, 2, "5 4"))
	for _, yr := range []int{preavisFirstRed(), preavisConfirmedRed()} {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.6" fill="%s"/>`,
			xY(yr)+slot/2, yLow(preavisRateAt(yr)), figBad)
	}
	// The rest of the story, in the corner the staircase has left free, stopping
	// clear of the dated rule so nothing is written across it.
	for i, l := range []string{
		fmt.Sprintf("confirmé en %d, hors de l'échelle dès %d", preavisConfirmedRed(), exit),
		"le voyant ne repasse plus jamais au vert",
	} {
		b.WriteString(sTxt(xY(preavisRuinYear)-10, 284+15*float64(i), 9.5, figBad, "end", "400", l))
	}

	// One timeline under both registers, and the gap itself, measured.
	for _, yr := range []int{1966, 1970, 1974, 1980, 1985, 1990, 1994} {
		b.WriteString(mTxt(xY(yr), 412, 9.5, figMuted, "middle", "400",
			fmt.Sprintf("%d", yr)))
	}
	b.WriteString(braceH(xY(preavisFirstRed()), xY(preavisRuinYear), 428,
		fmt.Sprintf("préavis : %d ans", preavisNotice())))

	b.WriteString(sTxt(24, 470, 10.5, figSoft, "start", "600", fmt.Sprintf(
		"Le plan condamné a prévenu %d ans à l'avance : premier rouge en %d, dernière ligne du compte en %d.",
		preavisNotice(), preavisFirstRed(), preavisRuinYear)))
	b.WriteString(plateFoot(488, []string{
		"Millésime 1966, 60/40 américain réel, 1 M€ et un retrait de 4 % indexé sur l'inflation, appliqué sans aucune réaction. Tout est réel.",
		"Seuils cités tels quels de « Quand s'inquiéter » : vert sous 4,3 %, orange de 4,3 à 5,2 %, rouge au-delà de 5,2 %,",
		"confirmé sur deux points espacés. Le préavis se compte du premier rouge (1974) à l'année d'épuisement (1994).",
	}))
	return svg(640, 534, b.String())
}
