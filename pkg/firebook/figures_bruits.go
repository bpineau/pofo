package firebook

import (
	"fmt"
	"strings"
)

// The three-noise ruler of the ruin-probability article. The article's
// "précision illusoire" section names three sources of uncertainty around a
// displayed ruin figure and RANKS them: sampling is the least, parameters are
// worse, the choice of model dominates everything. It gives the ranking in
// words and illustrative numbers; what it cannot do in prose is put the three
// on the same scale, which is the only way to see that the first is a hair and
// the third is the whole page.
//
// Drawn as one graduated ruler with three bars under it, all measured on ONE
// reference plan, the ranking becomes a proportion. The marker for the figure
// a simulator would actually display sits inside all three, and inside the
// narrowest by a wide margin: that is the article's "lisez la ruine en ordinal,
// pas en cardinal" made visible.
//
// The numbers are this plan's, not the article's example: its 2 / 5 / 9 / 14
// are thrown to illustrate a shape, and the legend says so.

// The reference plan, the same family the campaign's other computed plates
// use: the book's type portfolio through the central model, a rigid
// inflation-indexed withdrawal, one fixed seed.
const (
	bruitYears   = 30
	bruitCapital = 1e6
	bruitRule    = 0.04
	bruitPaths   = 20000
	bruitRefPath = 200000 // the path count the "true" ruin is measured at
	bruitDisplay = 2000   // the path count a simulator commonly draws
	bruitWorkers = 1
	bruitSeed    = 7
	bruitMuShift = 0.005 // half a real point, the article's own probe
)

// bruitWeights is the portfolio, over the four bricks of the basket plate.
var bruitWeights = []float64{0.60, 0.25, 0.075, 0.075}

// bruitRef is the plan's ruin measured with enough paths for the sampling
// noise to vanish, in percent: the value the three noises are read around.
const bruitRef = 3.67

// bruitBand is one of the three noises: its name, the two bounds it puts on the
// displayed figure, and the note printed under its bar.
type bruitBand struct {
	name     string
	lo, hi   float64
	detail   string
	color    string
	loLabel  string
	hiLabel  string
	showEnds bool
}

// The three noises, measured on the reference plan and ordered as the plate
// draws them: narrowest first, which is also the article's ranking.
var bruitBands = []bruitBand{
	{
		name: "Le bruit d'échantillonnage",
		lo:   2.84, hi: 4.49,
		detail: "2 000 trajectoires, intervalle binomial à 95 %",
		color:  figGreen,
	},
	{
		name: "La sensibilité aux paramètres",
		lo:   2.38, hi: 5.96,
		detail: "le rendement réel espéré déplacé d'un demi-point",
		color:  figBlue,
	},
	{
		name: "Le choix du modèle",
		lo:   0.00, hi: 35.42,
		detail:  "les six colonnes du simulateur, même plan partout",
		color:   figBad,
		loLabel: "fenêtres historiques", hiLabel: "décennie perdue",
		showEnds: true,
	},
}

// bruitColumns are the six model columns behind the widest band, in the order
// the page runs them: the ticks the plate poses inside that bar.
var bruitColumns = []struct {
	name string
	ruin float64
}{
	{"Fenêtres historiques", 0.00},
	{"Bootstrap par blocs", 2.68},
	{"Student-t central", 3.96},
	{"Stress de séquence", 7.99},
	{"Échantillon mondial", 22.59},
	{"Décennie perdue", 35.42},
}

// bruitWidth is one noise's span, in points of ruin.
func (b bruitBand) width() float64 { return b.hi - b.lo }

// The plate's geometry: one ruler, three bars.
const (
	bruX0, bruX1 = 90.0, 600.0
	bruRuler     = 132.0
	bruRuinHi    = 36.0
	bruRow0      = 176.0
	bruRowGap    = 62.0
	bruBarH      = 17.0
)

func bruitScale() figScale {
	return figScale{Min: 0, Max: bruRuinHi, Px0: bruX0, Px1: bruX1}
}

func figTroisBruits() string {
	s := bruitScale()
	var b strings.Builder
	b.WriteString(plateHead("trois bruits, un seul chiffre",
		"Ce que vous lisez est le plus petit des trois"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Le même plan, la même ruine affichée, et les trois incertitudes qui l'entourent, sur la même règle"))

	// The ruler itself, graduated in points of ruin.
	b.WriteString(line(bruX0, bruRuler, bruX1, bruRuler, figRule, 1.4))
	for v := 0.0; v <= bruRuinHi; v += 5 {
		x := s.Map(v)
		b.WriteString(line(x, bruRuler, x, bruRuler+5, figRule, 1))
		b.WriteString(mTxt(x, bruRuler-8, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", v)))
	}
	b.WriteString(sTxt(bruX1, bruRuler-24, 9.5, figMuted, "end", "400",
		"probabilité de ruine, en %"))

	// The figure a simulator would print, marked once and carried down through
	// the three bars so the eye can see it sits inside all of them.
	// The marker drops through the BARS only, never through their names: a
	// hairline crossing a word reads as a strike-through.
	mx := s.Map(bruitRef)
	b.WriteString(poly([][2]float64{{mx, bruRuler}, {mx, bruRuler + 14}}, figInk, 1, "3 3"))
	for i := range bruitBands {
		y := bruRow0 + float64(i)*bruRowGap
		b.WriteString(poly([][2]float64{{mx, y - 5}, {mx, y + bruBarH + 5}}, figInk, 1, "3 3"))
	}
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.5" fill="%s"/>`, mx, bruRuler, figInk)
	b.WriteString(sTxt(mx+10, 100, 10.5, figInk, "start", "600",
		"le chiffre affiché : "+frNum(bruitRef, 1)+" %"))

	for i, band := range bruitBands {
		y := bruRow0 + float64(i)*bruRowGap
		x0, x1 := s.Map(band.lo), s.Map(band.hi)
		b.WriteString(sTxt(bruX0, y-8, 10.5, figSoft, "start", "600", band.name))
		b.WriteString(barH(x0, x1, y, bruBarH, mixHex("#fffdf9", band.color, 0.55)))
		// The bounds, in mono, outside the bar when it is too narrow to hold
		// them and inside when it is wide enough.
		if x1-x0 > 150 {
			b.WriteString(mTxt(x0+8, y+bruBarH-4.5, 10.5, figInk, "start", "600", frNum(band.lo, 1)+" %"))
			b.WriteString(mTxt(x1-8, y+bruBarH-4.5, 10.5, figInk, "end", "600", frNum(band.hi, 1)+" %"))
		} else {
			b.WriteString(mTxt(x0-8, y+bruBarH-4.5, 10.5, figSoft, "end", "600", frNum(band.lo, 1)+" %"))
			b.WriteString(mTxt(x1+8, y+bruBarH-4.5, 10.5, figSoft, "start", "600", frNum(band.hi, 1)+" %"))
		}
		b.WriteString(sTxt(bruX0, y+bruBarH+15, 9.5, figMuted, "start", "400", band.detail))
		if !band.showEnds {
			continue
		}
		// The widest band carries its six columns as ticks, and its two ends
		// are named: they are models, not error bars.
		for _, c := range bruitColumns {
			cx := s.Map(c.ruin)
			b.WriteString(line(cx, y+2, cx, y+bruBarH-2, mixHex("#fffdf9", band.color, 0.95), 1.2))
		}
		b.WriteString(sTxt(x0, y+bruBarH+28, 9.5, band.color, "start", "600", band.loLabel))
		b.WriteString(sTxt(x1, y+bruBarH+28, 9.5, band.color, "end", "600", band.hiLabel))
	}

	b.WriteString(sTxt(24, 372, 10.5, figSoft, "start", "600",
		"Les décimales sont du bruit, les écarts entre modèles sont du signal : "+
			"la ruine se lit en ordinal, pas en cardinal."))
	b.WriteString(plateFoot(394, []string{
		"Plan de référence : 1 M€, portefeuille type du livre (60 % actions mondiales, 25 % obligations longues, " +
			"7,5 % or, 7,5 % suivi de tendance),",
		"retrait rigide de 4 % indexé sur l'inflation, horizon 30 ans, 20 000 tirages et graine fixe " +
			"pour chaque mesure.",
		"Bruit d'échantillonnage : intervalle binomial à 95 % autour de la ruine vraie (" + frNum(bruitRef, 2) +
			" %, mesurée sur 200 000 tirages), pour les 2 000",
		"trajectoires que tire couramment un simulateur. Paramètres : le même plan avec un rendement réel espéré " +
			"déplacé d'un demi-point",
		"vers le haut, puis vers le bas. Modèle : les six colonnes de la page FIRE, de la plus optimiste " +
			"à la plus sombre.",
		"Les trois bruits se cumulent, ils ne s'annulent pas. Les chiffres sont ceux de ce plan-là ; " +
			"ceux du texte",
		"(2, 5, 9 et 14 %) illustrent la même forme sur un autre plan.",
	}))
	return svg(640, 502, b.String())
}

// bruitOrdered reports whether the three noises come in the order the article
// claims, each strictly wider than the one before.
func bruitOrdered() bool {
	for i := 1; i < len(bruitBands); i++ {
		if bruitBands[i].width() <= bruitBands[i-1].width() {
			return false
		}
	}
	return true
}
