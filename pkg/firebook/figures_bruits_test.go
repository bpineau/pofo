package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// bruitPlan is the reference plan every measurement below runs, differing only
// in the return model it is handed.
func bruitPlan(src scenario.Source) decumul.Plan {
	return decumul.Plan{
		Capital: bruitCapital, NeedAnnual: bruitCapital * bruitRule,
		Years: bruitYears, Source: src,
	}
}

// bruitCentral fits the central model on the same panel as the basket plate
// (plancherPanel / plancherFit), which is what makes this plate's plan the same
// plan the rest of the campaign measures.
func bruitCentral(t *testing.T) (panel scenario.Panel, mu, sigma, df float64) {
	t.Helper()
	frozenAgainstData(t)
	panel, _ = plancherPanel(t)
	mu, sigma, df = plancherFit(panel, bruitWeights)
	return panel, mu, sigma, df
}

// The three bands, rebuilt end to end: this is what "make figure-drift"
// reports when the bundled series move.
func TestBruitsBandsMatchTheModel(t *testing.T) {
	panel, mu, sigma, df := bruitCentral(t)
	central := scenario.ParametricSource{Mu: mu, Sigma: sigma, Df: df, Periods: bruitYears}

	// The reference ruin, measured with enough paths that the sampling noise
	// the first band measures is itself negligible here.
	ref := bruitPlan(central).Simulate(bruitRefPath, bruitWorkers, bruitSeed).RuinProb()
	if math.Abs(ref*100-bruitRef) > 0.05 {
		t.Errorf("the plan's ruin is %.2f %%, the plate marks %.2f", ref*100, bruitRef)
	}
	// Band one: the closed-form 95 % binomial interval at the path count a
	// simulator commonly draws.
	half := 1.96 * math.Sqrt(ref*(1-ref)/bruitDisplay) * 100
	if got := bruitBands[0]; math.Abs((ref*100-half)-got.lo) > 0.05 ||
		math.Abs((ref*100+half)-got.hi) > 0.05 {
		t.Errorf("the sampling interval is [%.2f, %.2f], the plate draws [%.2f, %.2f]",
			ref*100-half, ref*100+half, got.lo, got.hi)
	}
	// Band two: half a real point of expected return, up and down.
	up := bruitPlan(scenario.ParametricSource{Mu: mu + bruitMuShift, Sigma: sigma, Df: df,
		Periods: bruitYears}).Simulate(bruitPaths, bruitWorkers, bruitSeed).RuinProb()
	down := bruitPlan(scenario.ParametricSource{Mu: mu - bruitMuShift, Sigma: sigma, Df: df,
		Periods: bruitYears}).Simulate(bruitPaths, bruitWorkers, bruitSeed).RuinProb()
	if got := bruitBands[1]; math.Abs(up*100-got.lo) > 0.05 || math.Abs(down*100-got.hi) > 0.05 {
		t.Errorf("the parameter interval is [%.2f, %.2f], the plate draws [%.2f, %.2f]",
			up*100, down*100, got.lo, got.hi)
	}
	// Band three: every column the FIRE page runs, on this same plan.
	months := bruitYears * 12
	for i, c := range []scenario.Source{
		scenario.Compounded{Inner: scenario.HistoricalCohorts{
			Panel: panel, Weights: bruitWeights, Periods: months}, Group: 12},
		scenario.Compounded{Inner: scenario.StationaryBootstrap{
			Panel: panel, Weights: bruitWeights, MeanBlock: 24, Periods: months}, Group: 12},
		central,
		scenario.NewMarkovRegime(mu, sigma, df, bruitYears),
		broadPool(t, bruitYears),
		scenario.NewLostDecadeRegime(mu, sigma, df, bruitYears),
	} {
		ruin := bruitPlan(c).Simulate(bruitPaths, bruitWorkers, bruitSeed).RuinProb() * 100
		if math.Abs(ruin-bruitColumns[i].ruin) > 0.3 {
			t.Errorf("%s: ruin %.2f %%, the plate draws %.2f",
				bruitColumns[i].name, ruin, bruitColumns[i].ruin)
		}
	}
}

// The ranking the article states, checked on the frozen bands. It is the whole
// plate: if the three noises came in any other order, the drawing would be
// contradicting the sentence it illustrates.
func TestBruitsRankingIsTheArticles(t *testing.T) {
	if !bruitOrdered() {
		t.Fatalf("the three noises are %.2f, %.2f and %.2f points wide: not the article's order",
			bruitBands[0].width(), bruitBands[1].width(), bruitBands[2].width())
	}
	// And not marginally: the article calls sampling "la moindre" and says the
	// model choice "domine tout".
	if bruitBands[1].width() < 1.5*bruitBands[0].width() {
		t.Errorf("the parameter noise (%.2f) barely exceeds the sampling one (%.2f)",
			bruitBands[1].width(), bruitBands[0].width())
	}
	if bruitBands[2].width() < 5*bruitBands[1].width() {
		t.Errorf("the model noise (%.2f) does not dominate the parameter one (%.2f)",
			bruitBands[2].width(), bruitBands[1].width())
	}
	// The displayed figure sits inside all three.
	for _, band := range bruitBands {
		if bruitRef < band.lo || bruitRef > band.hi {
			t.Errorf("%q does not bracket the displayed figure %.2f %%", band.name, bruitRef)
		}
	}
	// The columns are the widest band's own ends, and they are sorted as the
	// plate draws them.
	for i := 1; i < len(bruitColumns); i++ {
		if bruitColumns[i].ruin <= bruitColumns[i-1].ruin {
			t.Errorf("%s is not grimmer than %s", bruitColumns[i].name, bruitColumns[i-1].name)
		}
	}
	if bruitColumns[0].ruin != bruitBands[2].lo ||
		bruitColumns[len(bruitColumns)-1].ruin != bruitBands[2].hi {
		t.Error("the model band's ends are no longer its own extreme columns")
	}
}

// The plate against the article that carries it.
func TestBruitsAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "ruine-et-probabilites")
	for _, want := range []string{
		"::: figure trois-bruits",
		"**Le bruit d'échantillonnage** est la moindre.",
		"**La sensibilité aux paramètres** est bien pire.",
		"**Le choix du modèle domine tout**",
		"lisez la ruine en ordinal, pas en cardinal",
		"Les décimales sont du bruit ; les écarts entre scénarios et entre modèles sont du signal.",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The plate probes the parameter axis the article names, half a real point.
	if bruitMuShift != 0.005 {
		t.Errorf("the plate shifts the mean by %.3f, the article probes half a point", bruitMuShift*100)
	}
	// And it draws the three noises in the article's own order.
	for i, want := range []string{"échantillonnage", "paramètres", "modèle"} {
		if !strings.Contains(bruitBands[i].name, want) {
			t.Errorf("band %d is %q, the article's order puts %q there", i+1, bruitBands[i].name, want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestBruitsPlateRenders(t *testing.T) {
	svg := FigureSVG("trois-bruits")
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
	for _, band := range bruitBands {
		if !strings.Contains(svg, ">"+band.name+"<") {
			t.Errorf("the plate does not name %q", band.name)
		}
		if !strings.Contains(svg, ">"+band.detail+"<") {
			t.Errorf("the plate does not carry the note %q", band.detail)
		}
		for _, v := range []float64{band.lo, band.hi} {
			if want := ">" + frNum(v, 1) + " %<"; !strings.Contains(svg, want) {
				t.Errorf("the plate does not print the bound %q of %q", want, band.name)
			}
		}
	}
	for _, want := range []string{
		">le chiffre affiché : 3,7 %<",
		">fenêtres historiques<",
		">décennie perdue<",
		">probabilité de ruine, en %<",
		"la ruine se lit en ordinal, pas en cardinal",
		"Les trois bruits se cumulent, ils ne s'annulent pas",
		"ceux du texte",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// Three bars, and the six column ticks inside the widest one.
	if n := strings.Count(svg, "<path"); n != len(bruitBands) {
		t.Errorf("%d bars drawn, expected one per noise", n)
	}
	if n := strings.Count(svg, "<circle"); n != 1 {
		t.Errorf("%d dots drawn, expected the displayed figure alone", n)
	}
}
