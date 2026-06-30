package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// PathsResult is the wealth fan chart for one chosen model: the picture of the
// market the simulation actually produces (bands + a few sample/ruin paths),
// the answer to "is this an 80% mega-crash or March-2020 repeated?".
type PathsResult struct {
	Model  string `json:"model"`
	FanSVG string `json:"fanSvg"`
	Note   string `json:"note"`
}

// fanPercentiles are the bands drawn: a 5-95 outer envelope, a 25-75 inner
// band, and the median.
var fanPercentiles = []float64{0.05, 0.25, 0.50, 0.75, 0.95}

// Paths renders the wealth fan for the requested model (pr.FanModel), defaulting
// to the calibrated central Student-t. It evaluates that one model at the user's
// planned spend, so the fan shows the spread of outcomes the user actually faces.
func Paths(pr Params, panel *scenario.Panel) PathsResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	ns, ok := pickModel(modelSources(pr, panel), pr.FanModel)
	if !ok {
		return PathsResult{Note: "no return model available"}
	}
	base := pr.plan()
	base.Monthly = false
	base.Source = ns.source

	fan := base.Simulate(pr.NPaths, simWorkers, 7).Fan(fanPercentiles, 8)
	svg := chart.Fan(
		chart.Options{Title: "Simulated wealth, real € (" + ns.name + ")"},
		"Year", fan.Bands, sampleLines(fan.Samples))
	return PathsResult{Model: ns.name, FanSVG: svg}
}

// pickModel returns the named source matching want, falling back to the central
// Student-t and then to the first available model.
func pickModel(sources []namedSource, want string) (namedSource, bool) {
	if len(sources) == 0 {
		return namedSource{}, false
	}
	for _, ns := range sources {
		if ns.name == want {
			return ns, true
		}
	}
	for _, ns := range sources {
		if ns.name == "Student-t" {
			return ns, true
		}
	}
	return sources[0], true
}

// sampleLines extracts the plain wealth slices the chart consumes.
func sampleLines(samples []decumul.SamplePath) [][]float64 {
	out := make([][]float64, len(samples))
	for i, s := range samples {
		out[i] = s.Wealth
	}
	return out
}
