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
	LifeSVG     string `json:"lifeSvg"`     // funded, broke and dead shares per year
	RuinYearSVG string `json:"ruinYearSvg"` // when the ruined paths ran out
	CausesSVG   string `json:"causesSvg"`   // the shapes of ruin
	BequestSVG  string `json:"bequestSvg"`  // the estate distribution
	Cards       []Card `json:"cards"`       // the section's headline figures
	Note        string `json:"note"`        // caveat; "" when none
	// Annuitised says the plan bought an annuity, so this view is the only one
	// on the page carrying it. The front end shows the standing note that says
	// so, rather than leaving a reader to wonder why the sections above did
	// not move.
	Annuitised bool `json:"annuitised"`
}

// Lifecycle runs the central model at the planned spend with the death drawn
// INSIDE each path (decumul.Lifetime), so every figure here is counted rather
// than weighted after the fact: ruin means broke while alive, the ruin-year
// histogram holds the failures somebody lived through, and the terminal wealth
// is the estate at the household's own end.
//
// The same return draws are replayed a second time with mortality off, which
// is what the "ignoring mortality" card reports: the two runs differ by the
// drawn death alone, so the pair reads as one comparison rather than two
// simulations.
func Lifecycle(pr Params, panel *scenario.Panel) LifecycleResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	base := pr.plan()
	base.Monthly = false
	base.Source = pr.detailSource(panel, pr.Years)

	// A couple of the same age, each drawn from the bundled French law, and no
	// survivor adjustment: exactly the household the page's survival curve
	// assumed, now sampled per path instead of applied as a weight.
	age := pr.age()
	mortal := base
	mortal.Lifetime = &decumul.Lifetime{
		Self:    decumul.Life{Age: age},
		Partner: &decumul.Life{Age: age},
	}
	// The annuity lives here and nowhere else: this is the only view whose
	// households can outlive their money, which is the risk it insures.
	mortal.Annuity = pr.annuity()
	draws := mortal.Draw(pr.NPaths, simWorkers, 7)
	e := mortal.SimulateOn(draws, simWorkers)
	immortal := base.SimulateOn(draws, simWorkers) // same returns, nobody dies

	lifeSVG := lifeBandSVG(e.LifeBands(1, 2), age, pr, mortal.Annuity)

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
	ruinSVG := darkBars(chart.Options{Title: "When ruin happens", XLabel: "years into retirement, by year of failure",
		Width: 600, Height: 340}, bars)

	// Why plans fail: classify the ruined paths by trajectory shape (halved
	// early / never grew / prospered then outlived), sharper than the old
	// when-it-failed timing proxy.
	sh := e.RuinShapes()
	pct := roundShares100(sh.Crash, sh.Grind, sh.Longevity)
	causesSVG := darkCategoryBars(chart.Options{Title: "Why plans fail", Width: 460},
		[]chart.CatBar{
			{Label: "Early crash", Value: sh.Crash, Text: fmt.Sprintf("%d%%", pct[0]), Color: chart.ColorBad},
			{Label: "Slow grind", Value: sh.Grind, Text: fmt.Sprintf("%d%%", pct[1]), Color: chart.ColorWarn},
			{Label: "Longevity", Value: sh.Longevity, Text: fmt.Sprintf("%d%%", pct[2]), Color: chart.ColorDead},
		})

	// What you leave behind: the distribution of terminal real wealth across
	// paths (0 for the ruined). It shows the upside the broke/dead view hides:
	// most futures end far richer than they started, a few end with nothing.
	bequestSVG := darkBars(chart.Options{Title: "What's left at the end", XLabel: "real wealth at the household's own end",
		Width: 900, Height: 300}, bequestBuckets(e.Estates()))

	cards := lifecycleCards(e, immortal, mortal.Annuity != nil)
	// The third twin: the same households, the same deaths, no annuity. It is
	// what makes the purchase readable as a trade rather than a level.
	if a := mortal.Annuity; a != nil {
		plain := mortal
		plain.Annuity = nil
		cards = append(cards, annuityCards(e, plain.SimulateOn(draws, simWorkers), pr, a)...)
	}
	return LifecycleResult{LifeSVG: lifeSVG, RuinYearSVG: ruinSVG, CausesSVG: causesSVG, BequestSVG: bequestSVG,
		Cards: cards, Annuitised: mortal.Annuity != nil}
}

// lifeBandSVG draws the lifecycle band: at every age, the share of the
// simulated households that are gone, alive and broke, or alive and funded,
// the funded split by their wealth against the capital they started with.
//
// The two risks sit on the two edges, where each reads against an axis rather
// than floating mid-stack: broke while alive on the baseline, so its height is
// its share, and gone from the top, so the reader sees at a glance which of
// the two a plan is really exposed to (the comparison Engaging Data's "Rich,
// Broke or Dead" made famous). Those two frontiers are the only drawn lines;
// the funded bands between them are one hue stepped by wealth (an ordinal
// ramp, dimmest nearest the failure) parted by surface gaps, and the
// mortality layer is a neutral grey, an absorbing state rather than an outcome
// to grade. The x axis is the AGE of the household's first member, with the
// plan year in the hover header, since "broke at 92" and "broke at 61" are the
// distinction the section exists to draw.
func lifeBandSVG(bands []decumul.LifeBand, age float64, pr Params, annuity *decumul.Annuity) string {
	n := len(bands)
	broke, poor, mid, rich, dead := make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n), make([]float64, n)
	ticks, tips := make([]string, n), make([]string, n)
	step := 5
	if n > 50 {
		step = 10
	}
	for t, b := range bands {
		broke[t], dead[t] = b.Broke*100, b.Dead*100
		poor[t], mid[t], rich[t] = b.Funded[0]*100, b.Funded[1]*100, b.Funded[2]*100
		a := int(age) + t
		if a%step == 0 {
			ticks[t] = strconv.Itoa(a)
		}
		tips[t] = fmt.Sprintf("age %d · year %d", a, t)
	}
	var markers []chart.Marker
	if pr.PensionAnnual > 0 && pr.PensionYear > 0 && pr.PensionYear < n-1 {
		markers = append(markers, chart.Marker{Axis: 'x', Value: float64(pr.PensionYear), Label: "pension"})
	}
	if annuity != nil && annuity.Year > 0 && annuity.Year < n-1 {
		markers = append(markers, chart.Marker{Axis: 'x', Value: float64(annuity.Year), Label: "annuity"})
	}
	return darkStackedAreaWith(
		chart.Options{Title: fmt.Sprintf("Alive, broke or gone · a couple aged %.0f at retirement", age), Width: 900, Height: 380},
		chart.StackOptions{
			XLabel: "age", YLabel: "% of simulated households", XTicks: ticks, XTips: tips,
			Max: 100, Unit: "%", Gaps: true, EndLabels: true, Markers: markers,
		},
		[]chart.AreaSeries{
			{Name: "Broke", Values: broke, Color: chart.ColorBad, Weight: 0.85, Outline: true},
			{Name: "≤ start", Values: poor, Color: chart.ColorAccent, Weight: 0.45},
			{Name: "1-2x start", Values: mid, Color: chart.ColorAccent, Weight: 0.68},
			{Name: "> 2x start", Values: rich, Color: chart.ColorAccent, Weight: 0.95},
			{Name: "Gone", Values: dead, Color: chart.ColorDead, Weight: 1, Outline: true},
		})
}

// bequestBuckets buckets each path's estate (real wealth at the household's
// end, 0 for the ruined) into readable bands and returns their share of all
// paths.
func bequestBuckets(estates []float64) []chart.Bar {
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
	for _, w := range estates {
		for i, b := range bands {
			if (i == 0 && w < 1) || (w >= b.lo && w < b.hi) {
				counts[i]++
				break
			}
		}
	}
	n := float64(max(len(estates), 1))
	bars := make([]chart.Bar, 0, len(bands))
	for i, b := range bands {
		share := 100 * float64(counts[i]) / n
		// The ruined bucket is not "least wealth", it is a different outcome:
		// it wears the failure hue, the wealth bands the series accent.
		col := ""
		if i == 0 {
			col = chart.ColorBad
		}
		bars = append(bars, chart.Bar{Label: b.label, Value: share, Text: fmtPctShare(share), Color: col})
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

// brokeYearsIfRuined turns the unconditional mean broke-years into the
// conditional one, the length of a failure among the households that had one.
// Zero when nothing failed.
func brokeYearsIfRuined(lo decumul.LifeOutcome) float64 {
	if lo.RuinAlive <= 0 {
		return 0
	}
	return lo.BrokeYearsMean / lo.RuinAlive
}

// lifecycleCards summarises the mortality-aware risk against its
// mortality-free twin: the classic ruin figure (same return draws, nobody
// dies), the share that really ran out while alive, how long that failure
// lasted, and the odds of outliving the horizon. annuitised only affects what
// the first card says about itself: the mortality-free twin cannot carry an
// annuity, so with one bought it is no longer the same plan as the rest.
func lifecycleCards(e, immortal decumul.Ensemble, annuitised bool) []Card {
	lo := e.LifeOutcome()
	horizonHelp := "The headline ruin figure: the same futures run to the full horizon with nobody ever dying. It is the number every other section of this page reports."
	if annuitised {
		horizonHelp += " It carries no annuity: with everyone certain to reach the horizon there is no longevity left to insure, only an income paid for longer than it was priced for, so the fixed horizon cannot price the purchase at all."
	}
	return []Card{
		{Label: "Ruin (ignoring mortality)", Value: fmt.Sprintf("%.1f%%", immortal.Outcome().RuinProb*100),
			Help: horizonHelp},
		{Label: "Ever alive and broke", Value: fmt.Sprintf("%.1f%%", lo.RuinAlive*100),
			Help: "Share of households that ran out of money with somebody still there to feel it. The death is drawn inside each future, so this is counted, not estimated."},
		{Label: "Years broke, when it happens", Value: fmt.Sprintf("%.0f y", brokeYearsIfRuined(lo)),
			Help: "How long a failure lasts: mean years lived after running out, among the households it happens to. Running dry two years before the end and running dry at 70 are not the same event, and this is what separates them."},
		{Label: "Still alive at horizon", Value: fmt.Sprintf("%.0f%%", lo.OutlivedPlan*100),
			Help: "Share of households still alive when the plan's horizon ends, so their estate is read there rather than at a drawn death."},
	}
}
