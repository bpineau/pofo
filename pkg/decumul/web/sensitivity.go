package web

import (
	"fmt"
	"sort"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// SensitivityResult is the "greeks" chart: the change in ruin (percentage
// points) from nudging one controllable lever at a time, under the central
// model. It answers "what is the most effective way to make my plan robust?".
type SensitivityResult struct {
	SVG  string `json:"sensitivitySvg"`
	Note string `json:"note"`
}

// Sensitivity nudges each lever once and measures the resulting change in ruin,
// at the user's planned spend under the calibrated central Student-t. Bars are
// signed: negative (green) lowers ruin, positive (red) raises it.
func Sensitivity(pr Params, panel *scenario.Panel) SensitivityResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	const seed = uint64(7)
	base := pr.plan()
	base.Monthly = false
	base.Source = scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years}
	baseRuin := base.Simulate(pr.NPaths, simWorkers, seed).RuinProb()

	// Each nudge is a single-lever change. The source's path length (Periods) is
	// at least Years, so shortening the horizon needs no source rebuild.
	nudges := []struct {
		label string
		apply func(decumul.Plan) decumul.Plan
	}{
		{"Spend -5 k€/yr", func(p decumul.Plan) decumul.Plan { p.NeedAnnual -= 5000; return p }},
		{"Capital +100 k€", func(p decumul.Plan) decumul.Plan { p.Capital += 100000; return p }},
		{"Horizon -5 y", func(p decumul.Plan) decumul.Plan { p.Years -= 5; return p }},
		{"Buffer +2 y", func(p decumul.Plan) decumul.Plan { p.Buffer.Years += 2; return p }},
		{"Cut 20% in downturns", func(p decumul.Plan) decumul.Plan { p.Flex = decumul.FlexRule{Threshold: 0.20, Cut: 0.20}; return p }},
	}

	bars := make([]chart.Bar, 0, len(nudges))
	for _, n := range nudges {
		ruin := n.apply(base).Simulate(pr.NPaths, simWorkers, seed).RuinProb()
		d := (ruin - baseRuin) * 100
		bars = append(bars, chart.Bar{Label: n.label, Value: d, Text: signedPP(d)})
	}
	// Most ruin-reducing levers first (most negative at the top).
	sort.SliceStable(bars, func(i, j int) bool { return bars[i].Value < bars[j].Value })

	svg := chart.HBars(chart.Options{Title: "Sensitivity: change in ruin (pp)"}, bars)
	return SensitivityResult{SVG: svg}
}

// signedPP formats a percentage-point delta with an explicit sign.
func signedPP(pp float64) string { return fmt.Sprintf("%+.1fpp", pp) }
