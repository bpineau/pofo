package web

import (
	"fmt"
	"strconv"

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
	CausesSVG   string `json:"causesSvg"`
	BequestSVG  string `json:"bequestSvg"`
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
	base.Source = centralSource(pr, cMu, cSigma, cDf, pr.Years)
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
		chart.Options{Title: fmt.Sprintf("Alive, broke or gone (couple aged %.0f at retirement)", age), Width: 900, Height: 360},
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
	ruinSVG := chart.Bars(chart.Options{Title: "When ruin happens (share of all paths, by year of failure)", Width: 600, Height: 360}, bars)

	// Why plans fail: decompose the ruined paths by the timing that stands in for
	// the cause (early crash / lost decade / longevity).
	rt := e.RuinTiming()
	pct := roundShares100(rt.Early, rt.Mid, rt.Late)
	causesSVG := chart.CategoryBars(chart.Options{Width: 460},
		[]chart.CatBar{
			{Label: "Early crash", Value: rt.Early, Text: fmt.Sprintf("%d%%", pct[0]), Color: "#D2402F"},
			{Label: "Lost decade", Value: rt.Mid, Text: fmt.Sprintf("%d%%", pct[1]), Color: "#C77E17"},
			{Label: "Longevity", Value: rt.Late, Text: fmt.Sprintf("%d%%", pct[2]), Color: "#9AA2B1"},
		})

	// What you leave behind: the distribution of terminal real wealth across
	// paths (0 for the ruined). It shows the upside the broke/dead view hides:
	// most futures end far richer than they started, a few end with nothing.
	bequestSVG := chart.Bars(chart.Options{Title: "What's left at the end · terminal real wealth across futures", Width: 900, Height: 300}, bequestBuckets(e))

	cards := lifecycleCards(e, age)
	return LifecycleResult{LifeSVG: lifeSVG, RuinYearSVG: ruinSVG, CausesSVG: causesSVG, BequestSVG: bequestSVG, Cards: cards}
}

// bequestBuckets buckets each path's terminal real wealth into readable bands
// and returns their share of all paths.
func bequestBuckets(e decumul.Ensemble) []chart.Bar {
	type band struct {
		label  string
		lo, hi float64
	}
	bands := []band{
		{"0 (ruined)", -1, 1},
		{"<0.5M", 1, 0.5e6},
		{"0.5-1M", 0.5e6, 1e6},
		{"1-2M", 1e6, 2e6},
		{"2-4M", 2e6, 4e6},
		{"4-8M", 4e6, 8e6},
		{"8M+", 8e6, 1e18},
	}
	counts := make([]int, len(bands))
	for _, p := range e.Paths {
		w := 0.0
		if len(p.Wealth) > 0 {
			w = p.Wealth[len(p.Wealth)-1]
		}
		for i, b := range bands {
			if (i == 0 && w < 1) || (w >= b.lo && w < b.hi) {
				counts[i]++
				break
			}
		}
	}
	n := float64(max(len(e.Paths), 1))
	bars := make([]chart.Bar, 0, len(bands))
	for i, b := range bands {
		share := 100 * float64(counts[i]) / n
		bars = append(bars, chart.Bar{Label: b.label, Value: share, Text: fmtPctShare(share)})
	}
	return bars
}

// roundShares100 rounds fractional shares to integer percentages that sum to
// exactly 100 (largest-remainder method), so a composition never reads 101%.
// All-zero shares stay all zero.
func roundShares100(shares ...float64) []int {
	total := 0.0
	for _, s := range shares {
		total += s
	}
	out := make([]int, len(shares))
	if total == 0 {
		return out
	}
	rem := make([]float64, len(shares))
	sum := 0
	for i, s := range shares {
		exact := 100 * s / total
		out[i] = int(exact)
		rem[i] = exact - float64(out[i])
		sum += out[i]
	}
	for sum < 100 {
		best := 0
		for i := range rem {
			if rem[i] > rem[best] {
				best = i
			}
		}
		out[best]++
		rem[best] = -1
		sum++
	}
	return out
}

func fmtPctShare(v float64) string {
	if v < 0.05 {
		return "0%"
	}
	return strconv.FormatFloat(v, 'f', 1, 64) + "%"
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
		{Label: "Ruin (ignoring mortality)", Value: fmt.Sprintf("%.1f%%", o.RuinProb*100)},
		{Label: "Ever alive and broke", Value: fmt.Sprintf("%.1f%%", pRuinAlive*100)},
		{Label: "Still alive at horizon", Value: fmt.Sprintf("%.0f%%", decumul.FrenchMortality.CoupleSurvival(age, horizon)*100)},
	}
}
