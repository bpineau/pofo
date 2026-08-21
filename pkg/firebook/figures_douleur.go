package firebook

import (
	"fmt"
	"strings"
)

// The pain-and-wage plate of the risk-premium article. The article's thesis is
// causal: a return is the WAGE of a risk carried, and the risk that gets paid
// is the one that hurts in the bad states of the world. Its ranked bars show
// the wages; nothing in the article shows the pain that buys them, so the
// causal claim stays a sentence one has to take on trust.
//
// Drawn as a cloud, pain on one axis and wage on the other, the claim becomes a
// shape: the assets that get paid line up, and the further left an asset sits
// (the more it loses when equities fall), the higher it is paid. Gold is the
// exception the article devotes a section to, and it lands visibly off that
// line: it does not hurt in those years, and what the window pays it is a
// monetary repricing rather than a wage.
//
// Both coordinates are measured over one common window, frozen here and
// recomputed by the guard test under frozenAgainstData.

// The measurement window: the fifty-seven calendar years 1969-2025, based at
// the end of 1968. Gold bounds it, and not by choice: the bundled spot series
// starts in April 1968 because before that there was no free gold price to
// record. Everything is a real dollar total return, deflated by the US CPI.
const (
	douleurFirst = 1969
	douleurLast  = 2025
)

// douleurWorstN is how many of the window's calendar years count as "the bad
// states of the world": the ten worst years of US equities, about one year in
// six, which covers every episode the book keeps returning to (1973 and 1974,
// 2002, 2008, 2022).
const douleurWorstN = 10

// The ten worst calendar years of US equities inside the window, worst first,
// in real dollars. Frozen, and recomputed by the guard test.
var douleurWorstYears = [douleurWorstN]int{2008, 1974, 2002, 2022, 1973, 1969, 1977, 2001, 1981, 2000}

// douleurAsset is one point: the name the plate prints, its mean real return
// over those ten years (pain, percent per year), its annualized real premium
// over cash over the whole window (premium, points per year), and whether it
// enters the fit. The assets that claim a wage do; gold does not, which is the
// question the plate asks. dx, dy and anchor place the label around the dot,
// by hand: six points in a small cloud leave no room for a rule.
type douleurAsset struct {
	name            string
	pain, premium   float64
	wage            bool
	dx, dy          float64
	anchor          string
	labelSuppressed bool // cash is named on the zero rule instead
}

// The six points, frozen from the bundled series. Cash is the origin by
// construction: the premium is measured against it, and it is its own pain in
// the bad years.
var douleurAssets = []douleurAsset{
	{"Actions US (S&amp;P 500)", -20.55, 5.85, true, -8, 17, "end", false},
	{"Portefeuille 60/40", -12.22, 4.79, true, 8, -16, "start", false},
	{"Treasuries 7-10 ans", 0.29, 2.16, true, 8, -22, "start", false},
	{"Treasuries 20 ans et +", -0.86, 1.83, true, -8, 22, "end", false},
	{"Cash (bons du Trésor 3 mois)", 0.07, 0, true, 0, 0, "middle", true},
	{"Or", 7.54, 3.75, false, -8, -18, "end", false},
}

// The least-squares line through the assets that earn a wage, gold excluded:
// premium = douleurFitA + douleurFitB * pain, with douleurFitR2 the share of
// their premiums it accounts for. The slope is NEGATIVE, which is the whole
// argument: the more an asset loses in the bad years, the more it is paid.
const (
	douleurFitA  = 1.3651
	douleurFitB  = -0.2346
	douleurFitR2 = 0.8631
)

// douleurPredict is the premium the fitted line expects at a given pain.
func douleurPredict(pain float64) float64 { return douleurFitA + douleurFitB*pain }

// douleurGold is the point the line does not explain.
func douleurGold() douleurAsset {
	for _, a := range douleurAssets {
		if !a.wage {
			return a
		}
	}
	return douleurAsset{}
}

// douleurGoldResidual is how far above the line gold sits, in points per year:
// the plate's most important reading, because it is the article's "aucune
// prime, on achète une corrélation" made visible.
func douleurGoldResidual() float64 {
	g := douleurGold()
	return g.premium - douleurPredict(g.pain)
}

// The plate's geometry.
const (
	douX0, douX1     = 80.0, 606.0
	douYTop, douYBot = 100.0, 320.0
	douPainMin       = -24.0
	douPainMax       = 9.0
	douPremMin       = -0.6
	douPremMax       = 6.4
)

func douleurScales() (x, y figScale) {
	return figScale{Min: douPainMin, Max: douPainMax, Px0: douX0, Px1: douX1},
		figScale{Min: douPremMin, Max: douPremMax, Px0: douYBot, Px1: douYTop}
}

func figDouleurPrime() string {
	x, y := douleurScales()
	var b strings.Builder
	b.WriteString(plateHead("douleur et salaire", "Ce qui est payé est ce qui fait mal au mauvais moment"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Chaque actif : ce qu'il a rendu dans les dix pires années des actions (horizontal), "+
			"ce qu'il paie au-dessus du cash (vertical)"))
	legendChips(&b, 74, [][2]string{{figBlue, "actifs à salaire"}, {figAccent, "or"}})

	// The premium axis, whose zero is the cash line, and the pain axis, whose
	// zero separates the assets that lose in those years from those that gain.
	axisTicks(&b, y, []float64{0, 2, 4, 6}, 0, "", douX0, douX1, false)
	b.WriteString(sTxt(douX0-10, douYTop-6, 9, figMuted, "end", "400", "points/an"))
	axisTicks(&b, x, []float64{-20, -15, -10, -5, 0, 5}, 0, " %", douYTop, douYBot, true)
	b.WriteString(sTxt(douX0, douYBot+37, 9.5, figMuted, "start", "400",
		"douleur : rendement réel moyen dans les dix pires années des actions, en % par an"))

	// The wage line: solid where the assets that earn one actually sit, dashed
	// where it is only an extrapolation toward gold.
	solidTo := 0.5
	b.WriteString(poly([][2]float64{
		{x.Map(douleurLineEnters()), douYTop},
		{x.Map(solidTo), y.Map(douleurPredict(solidTo))},
	}, figSoft, 1.8, ""))
	b.WriteString(poly([][2]float64{
		{x.Map(solidTo), y.Map(douleurPredict(solidTo))},
		{x.Map(8.2), y.Map(douleurPredict(8.2))},
	}, figSoft, 1.8, "5 4"))
	b.WriteString(sTxt(x.Map(-7)+8, y.Map(douleurPredict(-7))+22, 10, figSoft, "start", "600",
		"la droite des salaires"))

	// What the line does not explain: gold, and the drop from it to the line.
	g := douleurGold()
	gx := x.Map(g.pain)
	b.WriteString(poly([][2]float64{
		{gx, y.Map(g.premium) + 6}, {gx, y.Map(douleurPredict(g.pain))},
	}, figAccent, 1.4, "3 3"))
	b.WriteString(mTxt(gx-8, 250, 10.5, figAccent, "end", "600",
		"+"+frNum(douleurGoldResidual(), 1)+" pts"))

	// The points themselves.
	for _, a := range douleurAssets {
		col := figBlue
		if !a.wage {
			col = figAccent
		}
		px, py := x.Map(a.pain), y.Map(a.premium)
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5" fill="%s"/>`, px, py, col)
		if a.labelSuppressed {
			continue
		}
		b.WriteString(sTxt(px+a.dx, py+a.dy, 10, figSoft, a.anchor, "600", a.name))
		b.WriteString(mTxt(px+a.dx, py+a.dy+13, 10.5, col, a.anchor, "600",
			douleurPts(a.premium)))
		if !a.wage {
			b.WriteString(sTxt(px+a.dx, py+a.dy+27, 10, col, a.anchor, "600",
				"assurance, pas salaire"))
		}
	}
	// Cash is named on the rule it defines: its premium is zero because the
	// premium is measured against it.
	b.WriteString(sTxt(douX0+8, y.Map(0)-8, 10, figSoft, "start", "600",
		"Cash (bons du Trésor 3 mois) : l'étalon, prime nulle par construction"))

	b.WriteString(sTxt(24, 378, 10.5, figSoft, "start", "600",
		"Les actifs à salaire sont payés à proportion de ce qu'ils font perdre au mauvais moment ; "+
			"l'or est payé sans douleur."))
	b.WriteString(plateFoot(400, []string{
		"Fenêtre commune du 31/12/1968 au 31/12/2025 (57 années civiles), bornée par l'or : " +
			"avant 1968 son prix n'est pas libre.",
		"Douleur : rendement réel moyen de l'actif dans les dix pires années civiles des actions, " +
			"déterminées dans cette fenêtre :",
		douleurYearsLine() + " Prime : rendement réel annualisé moins celui du cash.",
		"Droite des moindres carrés sur les cinq actifs à salaire, or exclu (R² = " +
			frNum(douleurFitR2, 2) + "). Séries déflatées par l'IPC américain :",
		"S&amp;P 500 total return, Treasuries 7-10 et 20 ans et plus, 60/40 rééquilibré fin décembre, " +
			"bons du Trésor 3 mois, or au comptant.",
		"L'or part du prix fixe de 35 dollars abandonné en 1971 : la fenêtre lui compte une " +
			"revalorisation monétaire, pas un salaire.",
	}))
	return svg(640, 500, b.String())
}

// douleurLineEnters is the pain at which the wage line crosses the top of the
// plot, so the line is drawn inside its box and never over the title.
func douleurLineEnters() float64 { return (douPremMax - douleurFitA) / douleurFitB }

// douleurPts sets a premium: signed points per year, with the book's
// typographic minus and one decimal.
func douleurPts(v float64) string {
	if v > 0 {
		return "+" + frNum(v, 1) + " pts"
	}
	return frMinus(v, 1) + " pts"
}

// douleurYearsLine lists the ten worst equity years for the legend.
func douleurYearsLine() string {
	parts := make([]string, 0, douleurWorstN)
	for _, y := range douleurWorstYears {
		parts = append(parts, fmt.Sprint(y))
	}
	return strings.Join(parts, ", ") + "."
}
