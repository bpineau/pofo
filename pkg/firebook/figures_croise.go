package firebook

import (
	"fmt"
	"strings"
)

// The crossed-samples plate of the Anarkulova-Cederburg article. The article's
// whole point is that one plan, unchanged, gets two different verdicts
// depending on which century it is judged against, and that the gap opens
// exactly where everyone sizes their plan. It states that in a table of three
// samples; a table cannot show WHERE the two verdicts separate, and the
// separation is the argument.
//
// Two curves on one axis do: the same 60/40, the same rigid withdrawal, the
// same thirty years, judged first on the American record and then on the
// developed century. They start almost together at two percent and are seven
// times apart at four.
//
// # Which anchors this plate is held to
//
// The article's own table separates three samples and warns that their
// published numbers "ne mesurent pas la même chose". The two curves here are
// its rows two and three, not row one: the book's broad-sample model (sixteen
// JST countries, blocks drawn inside one country, the horizon of your plan) and
// the American record (rolling thirty-year windows). Row one is the 2023 paper
// itself, on thirty-eight countries with real mortality rather than a fixed
// horizon, and its published 17 % / 2.26 % therefore belong to a third setup;
// the legend prints them as such, next to what this machinery reads.
//
// The anchor that DOES apply is the article's own sentence about this panel:
// "les États-Unis ne tiennent que 3,75 %". Measured here, the American curve
// carries zero failures up to 3.75 % and its first failure at 3.80 %, which is
// the guard test's tightest check.

// The plan both curves judge, identical on both sides.
const (
	croiseYears   = 30
	croiseEquity  = 0.60
	croisePaths   = 20000
	croiseWorkers = 1
	croiseSeed    = 7
	croiseTarget  = 5.0 // the failure rate the plate rules
)

// croiseStep is one withdrawal rate and the two verdicts on it, in percent.
type croiseStep struct {
	rate, us, broad float64
}

// The measured grid. The range starts below the article's two percent so that
// BOTH crossings of the five percent rule are inside the plate: the broad
// sample is already over it at two.
var croiseGrid = []croiseStep{
	{1.50, 0.00, 4.20}, {1.75, 0.00, 5.15}, {2.00, 0.00, 6.35}, {2.25, 0.00, 7.55},
	{2.50, 0.00, 9.11}, {2.75, 0.00, 10.49}, {3.00, 0.00, 12.21}, {3.25, 0.00, 14.16},
	{3.50, 0.00, 16.46}, {3.75, 0.00, 19.37}, {4.00, 3.33, 22.59}, {4.25, 9.17, 26.11},
	{4.50, 12.50, 30.09}, {4.75, 20.00, 34.42}, {5.00, 25.83, 39.05},
}

// The two safe rates the plate marks, in percent: where each curve crosses the
// five percent failure rule. Frozen from the solver, recomputed by the guard.
const (
	croiseSafeUS    = 4.07
	croiseSafeBroad = 1.72
)

// The band where the article says the gap explodes, and where every plan is
// sized.
const (
	croiseBandLo = 3.5
	croiseBandHi = 4.5
)

// croiseAt reads both verdicts at one rate of the grid.
func croiseAt(rate float64) croiseStep {
	for _, s := range croiseGrid {
		if s.rate == rate {
			return s
		}
	}
	return croiseStep{}
}

// croiseRatio is how many times likelier failure is on the developed century
// than on the American one, at the rate everyone quotes.
func croiseRatio() float64 {
	s := croiseAt(4)
	return s.broad / s.us
}

// The plate's geometry.
const (
	croX0, croX1   = 84.0, 566.0
	croTop, croBot = 112.0, 312.0
	croRuinHi      = 40.0
)

func croiseScales() (rate, ruin figScale) {
	return figScale{Min: 1.5, Max: 5, Px0: croX0, Px1: croX1},
		figScale{Min: 0, Max: croRuinHi, Px0: croBot, Px1: croTop}
}

func figEchantillonCroise() string {
	rate, ruin := croiseScales()
	var b strings.Builder
	b.WriteString(plateHead("deux échantillons, un plan",
		"Le même retrait, jugé sur deux siècles"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Probabilité d'épuiser le capital en trente ans, pour un 60/40 et un retrait rigide indexé"))

	// The band where plans are actually sized, marked first.
	bx0, bx1 := rate.Map(croiseBandLo), rate.Map(croiseBandHi)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		bx0, croTop, bx1-bx0, croBot-croTop, mixHex("#fffdf9", figInk, 0.06))
	b.WriteString(sTxt((bx0+bx1)/2, croTop-6, 10, figMuted, "middle", "600",
		"là où tout le monde se dimensionne"))

	axisTicks(&b, ruin, []float64{0, 10, 20, 30, 40}, 0, " %", croX0, croX1, false)
	for r := 2.0; r <= 5.001; r += 0.5 {
		b.WriteString(mTxt(rate.Map(r), croBot+18, 10, figMuted, "middle", "400", frNum(r, 1)+" %"))
	}
	b.WriteString(sTxt(croX0, croBot+37, 9.5, figMuted, "start", "400",
		"taux de retrait rigide, en % du capital de départ  →"))

	// The rule the whole profession sizes against.
	ry := ruin.Map(croiseTarget)
	b.WriteString(poly([][2]float64{{croX0, ry}, {croX1, ry}}, figRule, 1.4, "5 4"))
	b.WriteString(sTxt(croX1, ry-6, 9.5, figMuted, "end", "400", "5 % d'échec"))

	var usPts, broadPts [][2]float64
	for _, s := range croiseGrid {
		usPts = append(usPts, [2]float64{rate.Map(s.rate), ruin.Map(s.us)})
		broadPts = append(broadPts, [2]float64{rate.Map(s.rate), ruin.Map(s.broad)})
	}
	b.WriteString(poly(broadPts, figBad, 2.6, ""))
	b.WriteString(poly(usPts, figDeep, 2.6, ""))
	b.WriteString(sTxt(rate.Map(4.6), ruin.Map(35), 10.5, figBad, "start", "600",
		"le siècle développé"))
	b.WriteString(sTxt(rate.Map(2.5), ruin.Map(3), 10.5, figDeep, "start", "600",
		"le siècle américain"))

	// The two safe rates, where each curve meets the rule.
	for _, c := range []struct {
		rate  float64
		color string
		label string
		dy    float64
	}{
		{croiseSafeBroad, figBad, "sûr ici : " + frNum(croiseSafeBroad, 2) + " %", -16},
		{croiseSafeUS, figDeep, "sûr ici : " + frNum(croiseSafeUS, 2) + " %", 18},
	} {
		x := rate.Map(c.rate)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.5" fill="%s"/>`, x, ry, c.color)
		// Each rate is written on the side of the rule its curve leaves free.
		anchor, dx := "middle", 0.0
		if c.dy > 0 {
			anchor, dx = "start", 8
		}
		b.WriteString(sTxt(x+dx, ry+c.dy, 10, c.color, anchor, "600", c.label))
	}

	b.WriteString(sTxt(24, 376, 10.5, figSoft, "start", "600", fmt.Sprintf(
		"À 4 %%, le même plan échoue %s fois sur cent sur le siècle américain, et %s fois sur cent "+
			"sur le siècle développé.",
		frNum(croiseAt(4).us, 0), frNum(croiseAt(4).broad, 0))))
	b.WriteString(plateFoot(398, []string{
		"Un seul plan des deux côtés : 60/40 domestique, retrait rigide indexé sur l'inflation, " +
			"horizon 30 ans, mêmes conventions.",
		"Siècle américain : panel Jorda-Schularick-Taylor des États-Unis (1872-2020), " +
			"120 fenêtres glissantes de trente ans.",
		"Siècle développé : le modèle broad-sample du livre, 16 pays du même panel, blocs tirés à l'intérieur " +
			"d'un même pays, 20 000 tirages, graine fixe.",
		"Le panel donne aux États-Unis 3,75 % sans aucun échec, exactement ce que dit le texte ; " +
			"à 5 % d'échec toléré il monte à " + frNum(croiseSafeUS, 2) + " %.",
		"Le papier de 2023 publie 17 % d'échec à 4 % et 2,26 % à 5 % d'échec, sur 38 pays, en blocs tous pays " +
			"et à mortalité réelle,",
		"quand cette planche tient l'horizon à trente ans fermes sur 16 pays : plus sévère (" +
			frNum(croiseAt(4).broad, 0) + " % et " + frNum(croiseSafeBroad, 1) + " %).",
		"Trois questions différentes, comme le dit le tableau de l'article : ces nombres ne mesurent pas " +
			"la même chose.",
	}))
	return svg(640, 490, b.String())
}
