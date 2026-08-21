package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The plate's engine, rebuilt: the same panel and the same fit as the basket
// plate (plancherPanel / plancherFit, in figures_medianeplancher_test.go), the
// withdrawal kernel of pkg/decumul over it, and the univariate R² decomposition
// the plate's method comment describes.

// impR2 is the univariate R² between one year's returns and the outcome, across
// paths: the square of their correlation.
func impR2(x, y []float64) float64 {
	n := float64(len(x))
	var sx, sy, sxx, syy, sxy float64
	for i := range x {
		sx, sy = sx+x[i], sy+y[i]
		sxx, syy, sxy = sxx+x[i]*x[i], syy+y[i]*y[i], sxy+x[i]*y[i]
	}
	num := n*sxy - sx*sy
	den := math.Sqrt((n*sxx - sx*sx) * (n*syy - sy*sy))
	if den == 0 {
		return 0
	}
	return (num / den) * (num / den)
}

// impProfile returns the normalized importance profile for one outcome, in
// percent, plus the share the first decade carries.
func impProfile(seqs []scenario.Sequence, outcome []float64) ([]float64, float64) {
	raw := make([]float64, impYears)
	total := 0.0
	for k := 0; k < impYears; k++ {
		year := make([]float64, len(seqs))
		for i, s := range seqs {
			year[i] = s[k]
		}
		raw[k] = impR2(year, outcome)
		total += raw[k]
	}
	shares, decade := make([]float64, impYears), 0.0
	for k := range raw {
		shares[k] = raw[k] / total * 100
		if k < impDecade {
			decade += shares[k]
		}
	}
	return shares, decade
}

// impMeasure runs the plan and returns both readings: the profile on the
// article's own outcome (the plan holds or it does not), and the first-decade
// share of the same decomposition run on terminal wealth instead.
func impMeasure(t *testing.T) (shares []float64, decade, wealthDecade float64) {
	t.Helper()
	frozenAgainstData(t)
	panel, _ := plancherPanel(t)
	mu, sigma, df := plancherFit(panel, impWeights)
	plan := decumul.Plan{
		Capital: impCapital, NeedAnnual: impCapital * impRule, Years: impYears,
		Source: scenario.ParametricSource{Mu: mu, Sigma: sigma, Df: df, Periods: impYears},
	}
	draws := plan.Draw(impPaths, impWorkers, impSeed)
	e := plan.SimulateOn(draws, impWorkers)

	ruin := make([]float64, len(e.Paths))
	logWealth := make([]float64, len(e.Paths))
	// A ruined path has no terminal wealth to take a logarithm of, so the wealth
	// reading is censored at one year of spending: a floor decided with the
	// method, not after seeing the numbers.
	floor := impCapital * impRule
	for i, p := range e.Paths {
		if p.Ruined {
			ruin[i] = 1
		}
		logWealth[i] = math.Log(math.Max(p.Wealth[len(p.Wealth)-1], floor))
	}
	shares, decade = impProfile(draws.Returns, ruin)
	_, wealthDecade = impProfile(draws.Returns, logWealth)
	return shares, decade, wealthDecade
}

// The forty frozen bars and the legend's second reading, rebuilt from the
// bundled series and the simulator's own engine.
func TestImportanceProfileMatchesTheModel(t *testing.T) {
	shares, decade, wealthDecade := impMeasure(t)
	for k, got := range shares {
		if math.Abs(got-impShares[k]) > 0.05 {
			t.Errorf("year %d: the model gives %.2f %%, the plate draws %.2f", k+1, got, impShares[k])
		}
	}
	if math.Abs(decade-impDecadeShare()) > 0.1 {
		t.Errorf("the first decade carries %.2f %%, the plate says %.2f", decade, impDecadeShare())
	}
	if math.Abs(wealthDecade-impWealthDecade) > 0.2 {
		t.Errorf("read on wealth the first decade carries %.2f %%, the legend says %.1f",
			wealthDecade, impWealthDecade)
	}
}

// The claim the plate is built to test, checked on the frozen profile. The
// article says at least 70 %; the measurement has to land in the same
// neighbourhood or the plate would be contradicting the sentence it
// illustrates, and no drawing choice could repair that.
func TestImportanceFirstDecadeIsNearTheArticlesClaim(t *testing.T) {
	got := impDecadeShare()
	if got < 65 || got > 85 {
		t.Errorf("the first decade carries %.1f %% of the outcome, the article says at least 70", got)
	}
	// The shares are shares: they add up.
	sum := 0.0
	for _, v := range impShares {
		sum += v
	}
	if math.Abs(sum-100) > 0.15 {
		t.Errorf("the forty bars add up to %.2f %%, not to a hundred", sum)
	}
	// The profile decays: no year of the second half matters more than any year
	// of the first five, and the tail is negligible.
	early := impShares[4]
	for k := impYears / 2; k < impYears; k++ {
		if impShares[k] >= early {
			t.Errorf("year %d carries %.2f %%, as much as year 5 (%.2f)", k+1, impShares[k], early)
		}
	}
	if impTailShare() > 10 {
		t.Errorf("the last twenty years carry %.1f %%, which is no longer negligible", impTailShare())
	}
	// The two first years alone carry about a fifth, which is the plate's other
	// posed reading.
	if two := impShares[0] + impShares[1]; two < 15 || two > 25 {
		t.Errorf("the first two years carry %.1f %%, the plate poses about a fifth", two)
	}
}

// The plate against the article that carries it.
func TestImportanceAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "sequence-des-rendements")
	for _, want := range []string{
		"::: figure importance-annees",
		"le sort d'une retraite de 40 ans se joue **à 70 % au moins dans sa première décennie**",
		"la corrélation entre le succès final d'un plan et les rendements réalisés est écrasante pour les 5 à 10 premières années",
		"Tout modèle qui tire les années indépendamment (Monte-Carlo naïf, y compris le modèle Student-t central) sous-estime légèrement le risque de séquence",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The plate runs the horizon the claim is stated for.
	if impYears != 40 {
		t.Errorf("the plate runs %d years, the article's claim is about a forty-year retirement", impYears)
	}
	if impDecade != 10 {
		t.Errorf("the plate bands %d years, the article's claim is about the first decade", impDecade)
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestImportancePlateRenders(t *testing.T) {
	svg := FigureSVG("importance-annees")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"\u2014", "\u2013", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, want := range []string{
		">les dix premières années : 70 % de l'issue<",
		">les deux premières années : 21 %<",
		">les vingt dernières : 6 % en tout<",
		">année de retraite  →<",
		"Le quart le plus court de l'horizon porte 70 % du résultat",
		"la concentration tombe à 58 % pour la première décennie",
		"Le texte annonce 70 % au moins, la mesure en donne 70",
		"sous-estime le risque de séquence",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// One bar per retirement year, and the decade band: no other shape.
	if n := strings.Count(svg, "<path"); n != impYears {
		t.Errorf("%d bars drawn, expected one per retirement year (%d)", n, impYears)
	}
	if n := strings.Count(svg, "<rect"); n != 1 {
		t.Errorf("%d rectangles drawn, expected the decade band alone", n)
	}
	if strings.Contains(svg, "<polyline") || strings.Contains(svg, "<circle") {
		t.Error("the plate grew a second series; it is meant to stay bare")
	}
}
