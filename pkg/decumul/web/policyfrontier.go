package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
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
	cMu, cSigma, cDf := centralParams(pr, panel)

	// A clean baseline: the plan with every adaptive spending rule stripped, so
	// each policy below is applied to the same fixed starting point on the same
	// central return model.
	bare := func() decumul.Plan {
		p := pr.plan()
		p.Monthly = false
		p.Flex = decumul.FlexRule{}
		p.Guard = decumul.Guardrails{}
		p.Ratchet = decumul.Ratchet{}
		p.Percent = 0
		p.Source = scenario.ParametricSource{Mu: cMu, Sigma: cSigma, Df: cDf, Periods: pr.Years}
		return p
	}
	wr := pr.NeedAnnual / pr.Capital

	policies := []struct {
		name  string
		color string
		apply func(*decumul.Plan)
	}{
		{"Fixed", "#D2402F", func(p *decumul.Plan) {}},
		{"Flex -10%", "#C77E17", func(p *decumul.Plan) { p.Flex = decumul.FlexRule{Threshold: 0.20, Cut: 0.10} }},
		{"Guardrails", "#0C8A47", func(p *decumul.Plan) {
			p.Guard = decumul.Guardrails{Upper: wr * 1.2, Lower: wr * 0.8, Cut: 0.10, Raise: 0.10}
		}},
		{"VPW", "#0B7285", func(p *decumul.Plan) { p.Percent = wr }},
	}

	pts := make([]chart.LabeledPoint, 0, len(policies))
	for _, pol := range policies {
		p := bare()
		pol.apply(&p)
		e := p.Simulate(pr.NPaths, simWorkers, 7)
		pts = append(pts, chart.LabeledPoint{
			X:     e.SpendCV() * 100,
			Y:     e.RuinProb() * 100,
			Label: pol.name,
			Color: pol.color,
		})
	}
	svg := chart.Scatter(chart.Options{Width: 720, Height: 360},
		"lifestyle volatility (spending CV, %)", "ruin (%)", pts)
	return PolicyFrontierResult{
		SVG:  svg,
		Note: "Same plan, four spending rules, on the central model. Up and left is safe but rigid; down and right never runs out but the standard of living swings.",
	}
}
