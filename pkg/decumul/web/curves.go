package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// CurvesResult carries the two planning curves: the safe withdrawal rate as
// the horizon stretches (how much a longer retirement costs), and the capital
// required at the target ruin as the planned spending moves (what one more
// k€/yr of lifestyle costs in capital).
type CurvesResult struct {
	HorizonSVG string `json:"horizonSvg"`
	CapitalSVG string `json:"capitalSvg"`
	Note       string `json:"note"`
}

// curveHorizons are the retirement lengths sampled on the safe-WR curve.
var curveHorizons = []int{25, 30, 35, 40, 45, 50, 55, 60}

// Curves computes both planning curves on the fixed rule (the conventional
// definition of a safe withdrawal), for the central and the broad-sample
// models so the epistemic spread stays visible.
func Curves(pr Params, panel *scenario.Panel) CurvesResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	const seed = uint64(7)
	cMu, cSigma, cDf := centralParams(pr, panel)

	// Safe WR vs horizon, central Student-t and broad-sample regime curves.
	// slot pins each model to its hero-strip palette slot (Student-t 0,
	// Broad-sample 2) so a model keeps its color across every chart.
	type curveModel struct {
		name   string
		slot   int
		source func(years int) scenario.Source
	}
	models := []curveModel{
		{"Central (Student-t)", 0, func(years int) scenario.Source {
			return centralSource(pr, cMu, cSigma, cDf, years)
		}},
		{"Broad-sample", 2, func(years int) scenario.Source {
			return broadSampleSource(years)
		}},
	}
	var series []chart.XYSeries
	for _, m := range models {
		xs := make([]float64, len(curveHorizons))
		ys := make([]float64, len(curveHorizons))
		for j, years := range curveHorizons {
			p := fixedRule(pr.plan())
			p.Monthly = false
			p.Years = years
			p.Source = m.source(years)
			safe := p.Solve(target, decumul.WithdrawalAxis(0, pr.Capital*0.15), pr.NPaths, simWorkers, seed)
			xs[j], ys[j] = float64(years), safe/pr.Capital*100
		}
		series = append(series, chart.XYSeries{Name: m.name, Xs: xs, Ys: ys, Color: chart.PaletteColor(m.slot)})
	}
	horizonSVG := chart.MultiLine(
		chart.Options{Title: "Safe withdrawal rate vs horizon (at your target ruin)", Width: 720, Height: 360},
		"Horizon (years)", "Safe WR %", series,
		chart.Marker{Axis: 'x', Value: float64(pr.Years), Label: "your horizon"},
	)

	// Required capital vs annual spending, central model, at the target ruin.
	spends := []float64{36000, 42000, 48000, 54000, 60000, 66000, 72000, 78000, 84000}
	xs := make([]float64, len(spends))
	ys := make([]float64, len(spends))
	for j, spend := range spends {
		p := fixedRule(pr.plan())
		p.Monthly = false
		p.NeedAnnual = spend
		p.Source = centralSource(pr, cMu, cSigma, cDf, pr.Years)
		// Ruin falls as capital rises: the smallest capital meeting the target.
		cap := p.Solve(target, decumul.CapitalAxis(solveLo, solveHi), pr.NPaths, simWorkers, seed)
		xs[j], ys[j] = spend/1000, cap/1e6
	}
	capitalSVG := chart.MultiLine(
		chart.Options{Title: "Capital required vs spending (central model, target ruin)", Width: 720, Height: 360},
		"Net spending k€/yr", "Required capital M€",
		[]chart.XYSeries{{Name: "", Xs: xs, Ys: ys, Color: chart.PaletteColor(0)}},
		chart.Marker{Axis: 'x', Value: pr.NeedAnnual / 1000, Label: "your plan"},
	)
	return CurvesResult{HorizonSVG: horizonSVG, CapitalSVG: capitalSVG}
}
