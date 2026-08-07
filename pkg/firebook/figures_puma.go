package firebook

import (
	"fmt"
	"math"
	"strings"
)

// This file holds the plate of the taxe-puma article: a two-dimensional map of
// the cotisation subsidiaire maladie (CSM) over activity income and realized
// capital income. Everything it draws comes from the closed formula below, so
// there is nothing frozen here: the plate is recomputed at render time and the
// guard test checks the formula against every amount the article quotes.

// The CSM parameters of the 2019 reform, exactly as article D. 380-1 of the
// code de la Sécurité sociale writes them and as taxe-puma.md quotes them.
const (
	csmPASS      = 48060.0 // plafond annuel de la Sécurité sociale, 2026
	csmRate      = 0.065   // 6,5 % of the base
	csmFranchise = 0.5     // × PASS: capital income below it owes nothing
	csmSwitch    = 0.2     // × PASS: activity income above it switches the charge off
	csmBaseCap   = 8.0     // × PASS: the capital-income base is capped there
)

// csm returns the yearly cotisation subsidiaire maladie owed on realized
// capital income and activity income, both in euros:
//
//	CSM = 6,5 % × (min(capital, 8 PASS) − 0,5 PASS) × (1 − activité / (0,2 PASS))
//
// The second factor is a linear taper, not a cliff: it shrinks the charge all
// the way down and lands exactly on zero at 0,2 PASS, where it stays. That is
// the switch the plate is about, and the reason the plate must show a fade and
// not a fake vertical wall.
func csm(capital, activity float64) float64 {
	base := math.Min(capital, csmBaseCap*csmPASS) - csmFranchise*csmPASS
	if base <= 0 {
		return 0
	}
	taper := 1 - activity/(csmSwitch*csmPASS)
	if taper <= 0 {
		return 0
	}
	return csmRate * base * taper
}

// csmSteps is the fill ramp of the map: six SOLID hex steps, never a gradient
// and never an rgba fill or an opacity (crengine, KOReader's EPUB renderer,
// paints those solid black). Each step is figAccent (180,120,60) composited
// onto the figure card background #fffdf9 (255,253,249) at alpha a, channel by
// channel: c = bg + a×(accent − bg), rounded.
//
//	0.12 -> #F6EDE2   0.26 -> #ECDAC8   0.42 -> #E0C5AA
//	0.60 -> #D2AD88   0.80 -> #C39362   1.00 -> #B4783C (= figAccent)
//
// upTo is the step's inclusive upper bound in euros per year; a cell at exactly
// zero gets no rect at all, so the two zero zones read as blank.
var csmSteps = []struct {
	upTo float64
	fill string
}{
	{500, "#F6EDE2"},
	{1500, "#ECDAC8"},
	{2500, "#E0C5AA"},
	{3500, "#D2AD88"},
	{5000, "#C39362"},
	{math.Inf(1), figAccent},
}

func csmFill(v float64) string {
	for _, s := range csmSteps {
		if v <= s.upTo {
			return s.fill
		}
	}
	return figAccent
}

// The grid: 10 columns of 1 200 € of activity income (0 to 12 000) and 12 rows
// of 12 000 € of realized capital income (0 to 144 000, that is 3 PASS). Those
// widths are chosen so both legal thresholds fall on a cell boundary: 0,2 PASS
// = 9 612 € sits 12 € above the 9 600 € column edge, and 0,5 PASS = 24 030 €
// sits 30 € above the 24 000 € row edge. Each cell carries the cotisation
// computed at its centre.
const (
	pumaCols     = 10
	pumaRows     = 12
	pumaActMax   = 12000.0
	pumaCapMax   = 144000.0
	pumaPlotL    = 100.0
	pumaPlotR    = 520.0
	pumaPlotT    = 92.0
	pumaPlotB    = 356.0
	pumaCellW    = (pumaPlotR - pumaPlotL) / pumaCols
	pumaCellH    = (pumaPlotB - pumaPlotT) / pumaRows
	pumaKeyX     = 540.0
	pumaKeyW     = 14.0
	pumaKeyStepH = 22.0
)

// figPumaInterrupteur maps the CSM over the two incomes that decide it. It
// reads in one look as the three facts the article states in prose: a flat
// floor below 0,5 PASS of capital income, a regular climb above it, and a
// charge that fades away with activity income until it is exactly zero from
// 0,2 PASS on, whatever the capital income.
func figPumaInterrupteur() string {
	px := func(activity float64) float64 {
		return pumaPlotL + activity/pumaActMax*(pumaPlotR-pumaPlotL)
	}
	py := func(capital float64) float64 {
		return pumaPlotB - capital/pumaCapMax*(pumaPlotB-pumaPlotT)
	}
	xSwitch, yFranchise := px(csmSwitch*csmPASS), py(csmFranchise*csmPASS)

	var b strings.Builder
	b.WriteString(plateHead("taxe puma", "La carte de la CSM, et l'interrupteur qui l'éteint"))

	// axis captions and the callout of the activity threshold, all horizontal
	b.WriteString(sTxt(pumaPlotL, 68, 10.5, figMuted, "start", "400", "revenus du capital réalisés (k€ par an)"))
	b.WriteString(sTxt(xSwitch, 68, 10.5, figDeep, "middle", "600", "seuil d'activité, 0,2 PASS"))
	b.WriteString(mTxt(xSwitch, 82, 10.5, figDeep, "middle", "600", frSpace(csmSwitch*csmPASS)+" €"))

	// the cell grid, drawn first so the blank zones keep a structure
	for c := 0; c <= pumaCols; c++ {
		x := pumaPlotL + float64(c)*pumaCellW
		b.WriteString(line(x, pumaPlotT, x, pumaPlotB, figGrid, 1))
	}
	for r := 0; r <= pumaRows; r++ {
		y := pumaPlotT + float64(r)*pumaCellH
		b.WriteString(line(pumaPlotL, y, pumaPlotR, y, figGrid, 1))
	}

	// the filled cells, each inset by a hairline so the discrete grid stays
	// legible where neighbours land in the same step
	var top float64
	for r := 0; r < pumaRows; r++ {
		capital := (float64(r) + 0.5) * (pumaCapMax / pumaRows)
		for c := 0; c < pumaCols; c++ {
			activity := (float64(c) + 0.5) * (pumaActMax / pumaCols)
			v := csm(capital, activity)
			if v <= 0 {
				continue
			}
			top = math.Max(top, v)
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
				pumaPlotL+float64(c)*pumaCellW+0.7, pumaPlotB-float64(r+1)*pumaCellH+0.7,
				pumaCellW-1.4, pumaCellH-1.4, csmFill(v))
		}
	}

	// the two legal thresholds, as hairlines
	b.WriteString(line(xSwitch, 88, xSwitch, pumaPlotB+2, figDeep, 1.4))
	b.WriteString(line(pumaPlotL-2, yFranchise, pumaPlotR+2, yFranchise, figDeep, 1.4))

	// the switch: zero all the way up the column, said once per third of it
	for _, r := range []int{3, 6, 9} {
		y := pumaPlotB - (float64(r)+0.5)*pumaCellH + 3.5
		b.WriteString(mTxt((xSwitch+pumaPlotR)/2, y, 10.5, figMuted, "middle", "600", "0 €"))
	}

	// the flat floor, annotated inside its own blank band
	b.WriteString(sTxt(pumaPlotL+8, 330, 10.5, figDeep, "start", "600", "franchise, 0,5 PASS"))
	b.WriteString(mTxt(pumaPlotL+8, 345, 10.5, figDeep, "start", "600", frSpace(csmFranchise*csmPASS)+" €"))
	b.WriteString(sTxt(pumaPlotR-6, 340, 10.5, figSoft, "end", "400", "en dessous, aucune cotisation"))

	// axes: capital ticks in k€, activity ticks in €
	b.WriteString(line(pumaPlotL, pumaPlotB, pumaPlotR, pumaPlotB, figRule, 1))
	b.WriteString(line(pumaPlotL, pumaPlotT, pumaPlotL, pumaPlotB, figRule, 1))
	for c := 0; c <= pumaCols; c++ {
		x := pumaPlotL + float64(c)*pumaCellW
		b.WriteString(line(x, pumaPlotB, x, pumaPlotB+4, figRule, 1))
	}
	for r := 0; r <= pumaRows; r++ {
		y := pumaPlotT + float64(r)*pumaCellH
		b.WriteString(line(pumaPlotL-4, y, pumaPlotL, y, figRule, 1))
	}
	for _, k := range []float64{0, 36, 72, 108, 144} {
		b.WriteString(mTxt(pumaPlotL-8, py(k*1000)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", k)))
	}
	for _, a := range []float64{0, 2400, 4800, 7200, 9600, 12000} {
		b.WriteString(mTxt(px(a), 372, 9.5, figMuted, "middle", "400", frSpace(a)))
	}
	b.WriteString(sTxt(pumaPlotL, 392, 10.5, figMuted, "start", "400", "revenus d'activité déclarés (€ par an)"))

	// the key: a vertical ramp with horizontal numbers beside it, darkest on top.
	// The open top step is closed by the largest cotisation the map reaches, so
	// the reader can price the darkest cells without leaving the plate.
	b.WriteString(sTxt(pumaKeyX, 70, 9.5, figMuted, "start", "400", "cotisation"))
	b.WriteString(sTxt(pumaKeyX, 82, 9.5, figMuted, "start", "400", "(€ par an)"))
	for i := range csmSteps {
		s := csmSteps[len(csmSteps)-1-i]
		y := pumaPlotT + float64(i)*pumaKeyStepH
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			pumaKeyX, y, pumaKeyW, pumaKeyStepH, s.fill)
		bound := s.upTo
		if math.IsInf(bound, 1) {
			bound = math.Round(top)
		}
		b.WriteString(mTxt(pumaKeyX+pumaKeyW+4, y+3.5, 9.5, figMuted, "start", "400", frSpace(bound)))
	}
	keyBot := pumaPlotT + float64(len(csmSteps))*pumaKeyStepH
	b.WriteString(mTxt(pumaKeyX+pumaKeyW+4, keyBot+3.5, 9.5, figMuted, "start", "400", "0"))
	b.WriteString(sTxt(pumaKeyX, keyBot+24, 9.5, figMuted, "start", "400", "blanc"))
	b.WriteString(mTxt(pumaKeyX, keyBot+37, 9.5, figMuted, "start", "400", "0 €"))

	b.WriteString(sTxt(pumaPlotL, 416, 10, figSoft, "start", "600",
		"L'interrupteur : passé le seuil d'activité, la colonne entière tombe à zéro."))
	b.WriteString(sTxt(pumaPlotL, 432, 9.5, figMuted, "start", "400",
		"PASS 2026 = "+frSpace(csmPASS)+" €. Décote linéaire de l'article D. 380-1 du code de la Sécurité sociale."))
	b.WriteString(sTxt(pumaPlotL, 446, 9.5, figMuted, "start", "400",
		"Assiette plafonnée à 8 PASS ("+frSpace(csmBaseCap*csmPASS)+" €), au-delà de cette carte. Chaque case est calculée en son centre."))
	return svg(640, 464, b.String())
}

// frSpace formats a whole number the French way, with a thousands space (the
// plain space the other plates use, so the mono face renders it identically).
func frSpace(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	var out []byte
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, s[i])
	}
	return string(out)
}
