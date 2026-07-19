package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// FrontierResult is the ruin-vs-withdrawal-rate chart: one curve per model, with
// markers at the user's planned rate and the target ruin. It shows the slope of
// risk (how fast ruin rises as you spend more) and where the plan sits across
// models, rather than a single point.
type FrontierResult struct {
	SVG  string `json:"frontierSvg"`
	Note string `json:"note"`
}

// frontierWRs are the withdrawal rates (fractions of capital) sampled along the
// frontier's x-axis.
var frontierWRs = []float64{0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.055, 0.06, 0.065, 0.07}

// Frontier sweeps each model's ruin across withdrawal rates and renders them
// together. The y-axis is ruin %, the x-axis the withdrawal rate %.
func Frontier(pr Params, panel *scenario.Panel) FrontierResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	if pr.Capital <= 0 {
		return FrontierResult{Note: "set a capital"}
	}
	base := pr.plan()
	base.Monthly = false

	var series []chart.XYSeries
	for i, ns := range modelSources(pr, panel) {
		xs := make([]float64, len(frontierWRs))
		ys := make([]float64, len(frontierWRs))
		for j, wr := range frontierWRs {
			p := base
			p.Source = ns.source
			p.NeedAnnual = wr * pr.Capital
			// Under guardrails the spending band is centred on the initial
			// withdrawal rate, so it must be re-anchored at each swept rate;
			// otherwise every point keeps the band of the user's current spend
			// and the whole curve barely moves when guardrails are toggled.
			if pr.Guardrails {
				p.Guard = decumul.Guardrails{Upper: wr * 1.2, Lower: wr * 0.8, Cut: 0.10, Raise: 0.10}
			}
			xs[j] = wr * 100
			ys[j] = p.Simulate(pr.NPaths, simWorkers, 7).RuinProb() * 100
		}
		series = append(series, chart.XYSeries{Name: ns.name, Xs: xs, Ys: ys, Color: chart.PaletteColor(i)})
	}

	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	svg := darkMultiLine(
		chart.Options{Title: "Ruin vs withdrawal rate", Width: 720, Height: 360}, "Withdrawal rate %", "Ruin %", series,
		chart.Marker{Axis: 'x', Value: pr.NeedAnnual / pr.Capital * 100, Label: "your plan"},
		chart.Marker{Axis: 'y', Value: target * 100, Label: "target"},
	)
	return FrontierResult{SVG: svg}
}
