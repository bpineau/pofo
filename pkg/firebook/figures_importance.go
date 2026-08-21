package firebook

import (
	"fmt"
	"strings"
)

// The importance profile of the sequence-risk article. The article states that
// the fate of a forty-year retirement is settled three quarters in its first
// decade, and describes the analysis that shows it: rank the paths by the
// return of their first ten years and read the outcomes. That is a two-bucket
// reading; nothing in the book shows the profile YEAR by year, and the profile
// is what makes the claim actionable, because it says where the line falls
// rather than that a line exists.
//
// Drawn as one bar per retirement year, the shape is the argument: the first
// two years alone carry a fifth of the outcome, the decade carries seven
// tenths, and the last fifteen years carry almost nothing at all. The plate has
// one series and no second axis, on purpose.
//
// # Method
//
// The central model of the simulator draws twenty thousand independent
// forty-year paths (Student-t fitted on the four bricks' common monthly
// history, the same panel and the same fit as the basket plate, seed 7, one
// worker so the draw is reproducible anywhere). Each path is run through the
// withdrawal kernel of pkg/decumul at a fixed inflation-indexed 4 % of the
// starting capital.
//
// The OUTCOME is the one the article names, "le succès final d'un plan": the
// indicator of whether the path ran out of money. For each retirement year k
// the plate measures the univariate R² between that year's return alone and the
// outcome across paths, then normalizes the forty of them to a hundred percent.
// A year's bar is therefore the share of the decided outcome that this single
// year's return accounts for.
//
// Read on terminal WEALTH instead of on success, the same profile is flatter
// (58 % for the first decade rather than 70 %); the legend carries that second
// reading, because the choice of outcome moves the number and hiding it would
// make the plate look more certain than it is.
//
// The guard test rebuilds the whole profile from the bundled series.

// The plan the profile is measured on: the book's type portfolio, a fixed
// inflation-indexed withdrawal, and the forty-year horizon the article's claim
// is stated for.
const (
	impYears   = 40
	impCapital = 1e6
	impRule    = 0.04
	impPaths   = 20000
	impWorkers = 1
	impSeed    = 7
	impDecade  = 10
)

// impWeights is the portfolio, over the same four bricks and in the same order
// as the basket plate (world equities, long government bonds, gold, trend).
var impWeights = []float64{0.60, 0.25, 0.075, 0.075}

// The measured profile: the share of the outcome each retirement year accounts
// for, in percent, year 1 first. Frozen, and recomputed by the guard test.
var impShares = [impYears]float64{
	10.51, 10.42, 8.17, 8.10, 6.65, 5.77, 5.52, 5.03, 5.24, 4.63,
	3.26, 3.78, 2.37, 3.05, 3.32, 2.01, 1.96, 1.52, 0.95, 1.53,
	1.38, 0.69, 0.56, 0.80, 0.72, 0.44, 0.55, 0.28, 0.26, 0.16,
	0.02, 0.01, 0.14, 0.05, 0.14, 0.00, 0.03, 0.00, 0.01, 0.00,
}

// impWealthDecade is the same first-decade share computed on the log of
// terminal wealth rather than on success, the second reading the legend
// carries. Frozen and recomputed alongside the profile.
const impWealthDecade = 58.3

// impDecadeShare is what the first decade carries, in percent.
func impDecadeShare() float64 {
	sum := 0.0
	for i := 0; i < impDecade; i++ {
		sum += impShares[i]
	}
	return sum
}

// impTailShare is what the last half of the horizon carries.
func impTailShare() float64 {
	sum := 0.0
	for i := impYears / 2; i < impYears; i++ {
		sum += impShares[i]
	}
	return sum
}

// The plate's geometry: one slot per retirement year, nothing else.
const (
	impX0, impX1   = 70.0, 608.0
	impTop, impBot = 118.0, 300.0
	impShareHi     = 11.0
)

func impScales() (slot float64, share figScale) {
	return (impX1 - impX0) / impYears,
		figScale{Min: 0, Max: impShareHi, Px0: impBot, Px1: impTop}
}

func figImportanceAnnees() string {
	slot, share := impScales()
	barW := slot - 4.5
	var b strings.Builder
	b.WriteString(plateHead("le profil d'importance",
		"Où se joue vraiment le sort d'une retraite"))
	b.WriteString(plateDeck(
		"Part de l'issue finale que le rendement de chaque année, à lui seul, explique"))

	// The decade the article points at, marked before the bars are drawn over
	// it, with what it carries written on it.
	bandX := impX0 + float64(impDecade)*slot - 4.5
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		impX0-2, impTop, bandX-impX0+2, impBot-impTop, mixHex("#fffdf9", figInk, 0.06))
	b.WriteString(sTxt((impX0+bandX)/2, 106, 11, figDeep, "middle", "600",
		fmt.Sprintf("les dix premières années : %s %% de l'issue", frNum(impDecadeShare(), 0))))

	axisTicks(&b, share, []float64{0, 2, 4, 6, 8, 10}, 0, " %", impX0, impX1, false)

	// One bar per year, the decade in full ink and the rest in a paler wash of
	// the same colour: one series, two weights of attention.
	pale := mixHex("#fffdf9", figDeep, 0.45)
	for i, v := range impShares {
		x := impX0 + float64(i)*slot + 2.25
		col := figDeep
		if i >= impDecade {
			col = pale
		}
		b.WriteString(barV(x, barW, impBot, share.Map(v), col))
	}
	for _, yr := range []int{1, 5, 10, 15, 20, 25, 30, 35, 40} {
		x := impX0 + float64(yr-1)*slot + 2.25 + barW/2
		b.WriteString(mTxt(x, impBot+18, 10, figMuted, "middle", "400", fmt.Sprint(yr)))
	}
	b.WriteString(sTxt(impX0, impBot+37, 9.5, figMuted, "start", "400",
		"année de retraite  →"))

	// The two readings that carry the argument, posed on the profile itself.
	b.WriteString(sTxt(impX0+2*slot+barW+10, share.Map(impShares[0])+6, 10, figSoft, "start", "600",
		fmt.Sprintf("les deux premières années : %s %%", frNum(impShares[0]+impShares[1], 0))))
	b.WriteString(sTxt(impX1, impBot-40, 10, figMuted, "end", "600",
		fmt.Sprintf("les vingt dernières : %s %% en tout", frNum(impTailShare(), 0))))

	b.WriteString(plateConclusion(364, fmt.Sprintf(
		"Le quart le plus court de l'horizon porte %s %% du résultat : la vigilance se date, elle ne s'étale pas.",
		frNum(impDecadeShare(), 0))))
	b.WriteString(plateFoot(386, []string{
		"Modèle central du simulateur : Student-t indépendant ajusté sur l'historique commun des quatre briques " +
			"(1987-2026),",
		"20 000 trajectoires, graine fixe. Plan : 60 % actions mondiales, 25 % obligations longues, 7,5 % or, " +
			"7,5 % suivi de tendance,",
		"retrait fixe de 4 % indexé, horizon 40 ans. Hauteur d'une barre : part de la variance de l'issue " +
			"(le plan tient ou non)",
		"expliquée par le seul rendement de cette année, R² univarié normalisé à 100 %. " +
			"Le « succès final » est l'issue que le texte nomme.",
		"Lue sur la richesse finale plutôt que sur le succès, la concentration tombe à " +
			frNum(impWealthDecade, 0) + " % pour la première décennie.",
		"Le texte annonce 70 % au moins, la mesure en donne " + frNum(impDecadeShare(), 0) +
			", stable de 67 à 77 % selon l'horizon, le retrait et le portefeuille.",
		"Un tirage indépendant sous-estime le risque de séquence : la concentration réelle est plutôt " +
			"au-dessus de cette planche.",
	}))
	return svg(640, 500, b.String())
}
