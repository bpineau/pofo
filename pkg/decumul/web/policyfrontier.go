package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/replay"
	"github.com/bpineau/pofo/pkg/scenario"
)

// PolicyFrontierResult is the decumulation trade-off chart: each withdrawal
// policy as one point of ruin (y) versus lifestyle volatility (x), so the cost
// of every rule is legible at once. A fixed rule risks running out; adaptive
// and percentage-of-portfolio rules trade that risk for a wobbling standard of
// living. No single ruin number can show this; the frontier is the point.
type PolicyFrontierResult struct {
	SVG  string `json:"policyFrontierSvg"`
	Note string `json:"note"`
}

// PolicyFrontier evaluates the same plan under four withdrawal policies on the
// central model and plots them as a ruin-versus-spending-volatility frontier.
func PolicyFrontier(pr Params, panel *scenario.Panel) PolicyFrontierResult {
	if pr.NPaths == 0 {
		pr.NPaths = 3000
	}
	if pr.Capital <= 0 || pr.NeedAnnual <= 0 {
		return PolicyFrontierResult{Note: "set a capital and a spending floor"}
	}
	// A clean baseline: the plan with every adaptive spending rule stripped, so
	// each policy below is applied to the same fixed starting point on the same
	// return model (the strip column the user selected, central by default).
	bare := func() decumul.Plan {
		p := fixedRule(pr.plan())
		p.Monthly = false
		p.Source = pr.detailSource(panel, pr.Years)
		return p
	}
	wr := pr.NeedAnnual / pr.Capital
	// Every policy shares this Source (they differ only in the spending rule),
	// so the paths are drawn once and replayed for each of the six rules.
	base := bare()
	seqs := base.DrawPaths(min(pr.NPaths, shapePaths), simWorkers, 7)

	// The seven rules, their names and their colours live in pkg/replay, so a
	// rule reads the same here and in the book's historical replays. Note the
	// two neighbouring greens: the guardrails family, where only the sensor
	// changes.
	policies := replay.Policies(replay.PolicyConfig{
		WR: wr, SafeWR: pr.safeRateTable(), RaiseCap: pr.raiseCap(),
		PVRate: pr.pvRate(), AmortReturn: pr.abwReturn(),
	})

	pts := make([]chart.LabeledPoint, 0, len(policies))
	for _, pol := range policies {
		p := base
		pol.Apply(&p)
		e := p.SimulateOn(seqs, simWorkers)
		// Lifestyle volatility on the SURVIVING paths only: post-ruin zeros
		// would inflate the fixed rule's CV with what is really ruin, the
		// quantity the y axis already carries. The x axis then measures pure
		// standard-of-living swing among the futures that work.
		pts = append(pts, chart.LabeledPoint{
			X:     survivors(e).SpendCV() * 100,
			Y:     e.RuinProb() * 100,
			Label: pol.Name,
			Color: pol.Color,
		})
	}
	svg := darkScatter(chart.Options{Width: 720, Height: 360},
		"lifestyle volatility (spending CV among surviving futures, %)", "ruin (%)", pts)
	return PolicyFrontierResult{
		SVG:  svg,
		Note: "Same plan, seven spending rules, under the selected return model. Up and left is safe but rigid; down and right rarely runs out but the standard of living moves. Volatility is measured among the surviving futures, so ruin is not double-counted on both axes.",
	}
}
