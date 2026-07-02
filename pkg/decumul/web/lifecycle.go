package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// LifecycleResult is the mortality-aware view ("rich, broke or dead"): at each
// age, the share of simulated households that are gone, alive-but-broke, or
// alive-and-funded, plus the distribution of the year ruin happens. Mortality
// puts the ruin risk in perspective: a failure at 93 is not the same life
// event as a failure at 61.
type LifecycleResult struct {
	LifeSVG     string `json:"lifeSvg"`
	RuinYearSVG string `json:"ruinYearSvg"`
	Cards       []Card `json:"cards"`
	Note        string `json:"note"`
}

// Lifecycle runs the central model at the planned spend and crosses the ruin
// timing with a French couple survival curve at the user's age.
func Lifecycle(pr Params, panel *scenario.Panel) LifecycleResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	base := pr.plan()
	base.Monthly = false
	cMu, cSigma, cDf := centralParams(pr, panel)
	base.Source = scenario.ParametricSource{Mu: cMu, Sigma: cSigma, Df: cDf, Periods: pr.Years}
	e := base.Simulate(pr.NPaths, simWorkers, 7)

	age := pr.age()
	surv := func(years float64) float64 {
		return decumul.FrenchMortality.CoupleSurvival(age, years)
	}
	curve := e.LifeCurve(surv)

	funded := make([]float64, len(curve))
	broke := make([]float64, len(curve))
	dead := make([]float64, len(curve))
	for i, pt := range curve {
		funded[i] = pt.Funded * 100
		broke[i] = pt.Broke * 100
		dead[i] = pt.Dead * 100
	}
	lifeSVG := chart.StackedArea(
		chart.Options{Title: fmt.Sprintf("Alive, broke or gone (couple aged %.0f at retirement)", age), Width: 720, Height: 440},
		"Years into retirement", "% of simulated households",
		[]chart.AreaSeries{
			{Name: "Funded", Values: funded, Color: "#12B76A"},
			{Name: "Broke", Values: broke, Color: "#D92D20"},
			{Name: "Gone", Values: dead, Color: "#D5DBE5"},
		})

	// Ruin-year histogram, folded into 5-year buckets so it reads at a glance.
	hist := e.RuinYearHistogram()
	var bars []chart.Bar
	for from := 0; from < len(hist); from += 5 {
		share := 0.0
		for k := from; k < min(from+5, len(hist)); k++ {
			share += hist[k]
		}
		bars = append(bars, chart.Bar{
			Label: fmt.Sprintf("%d-%d", from, min(from+4, len(hist)-1)),
			Value: share * 100,
			Text:  fmt.Sprintf("%.1f%%", share*100),
		})
	}
	ruinSVG := chart.Bars(chart.Options{Title: "When ruin happens (share of all paths, by year of failure)", Width: 480, Height: 280}, bars)

	cards := lifecycleCards(e, age)
	return LifecycleResult{LifeSVG: lifeSVG, RuinYearSVG: ruinSVG, Cards: cards}
}

// lifecycleCards summarises the mortality-adjusted risk: the classic ruin
// figure, the probability of ever being alive AND broke, and the odds of
// outliving the horizon.
func lifecycleCards(e decumul.Ensemble, age float64) []Card {
	o := e.Outcome()
	// Probability of experiencing ruin while alive: the ruin year must be
	// reached alive, i.e. weight each ruined path by survival to its ruin year.
	pRuinAlive := 0.0
	if n := len(e.Paths); n > 0 {
		for _, p := range e.Paths {
			if p.Ruined && p.RuinYear >= 0 {
				pRuinAlive += decumul.FrenchMortality.CoupleSurvival(age, float64(p.RuinYear))
			}
		}
		pRuinAlive /= float64(n)
	}
	horizon := float64(e.Years)
	return []Card{
		{"Ruin (ignoring mortality)", fmt.Sprintf("%.1f%%", o.RuinProb*100)},
		{"Ever alive and broke", fmt.Sprintf("%.1f%%", pRuinAlive*100)},
		{"Still alive at horizon", fmt.Sprintf("%.0f%%", decumul.FrenchMortality.CoupleSurvival(age, horizon)*100)},
	}
}
