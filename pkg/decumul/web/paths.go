package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Fan is one model's wealth fan chart, named for its column in the strip.
type Fan struct {
	Name string `json:"name"` // the planning model
	SVG  string `json:"svg"`  // its wealth fan chart
	// Samples name the futures drawn as thin lines, in the order of the
	// chart's data-pick indices: what the page posts back to /api/path to
	// open one of them year by year.
	Samples []SampleRef `json:"samples"`
}

// SampleRef identifies one sample future of a fan: its path index in the
// ensemble (decumul.SamplePath.Index, the stable key) and what it came to.
type SampleRef struct {
	Index  int     `json:"index"`  // path index, the decumul.Plan.DrawPath key
	Ruined bool    `json:"ruined"` // it ran out
	Final  float64 `json:"final"`  // real wealth at the horizon, euros
}

// PathsResult is the set of wealth fan charts, one per planning model: the
// picture of the market each simulation actually produces (bands + a few
// sample/ruin paths), the answer to "is this an 80% mega-crash or March-2020
// repeated?". The four planning models are shown side by side so the central
// case and the successively grimmer stresses can be compared at a glance.
type PathsResult struct {
	Fans []Fan  `json:"fans"` // one per planning model
	Note string `json:"note"` // caveat; "" when none
}

// fanPercentiles are the bands drawn: a 5-95 outer envelope, a 25-75 inner
// band, and the median.
var fanPercentiles = []float64{0.05, 0.25, 0.50, 0.75, 0.95}

// fanModels are the four planning models shown as fans, from the central case to
// the grimmest stress. They are the synthetic family, always present (no panel
// needed), so the 2x2 grid is stable across parametric and portfolio modes.
var fanModels = []string{"Student-t", "Sequence stress", "Broad-sample", "Lost decade"}

// Paths renders the wealth fan for each planning model at the user's planned
// spend, so the fans show the spread of outcomes the user actually faces under
// each lens.
func Paths(pr Params, panel *scenario.Panel) PathsResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	sources := modelSources(pr, panel)
	base := pr.plan() // once: a risk-guardrail plan solves its safe-rate table here
	var picked []string
	for _, name := range fanModels {
		if _, ok := pickModel(sources, name); ok {
			picked = append(picked, name)
		}
	}
	var res PathsResult
	if len(picked) > 0 {
		res.Fans = make([]Fan, len(picked))
	}
	concurrently(len(picked), func(i int) {
		name := picked[i]
		p, _ := fanPlan(base, sources, name)
		fan := p.Simulate(pr.NPaths, simWorkers, fanSeed).Fan(fanPercentiles, 8)
		svg := darkFan(
			chart.Options{Title: "Simulated wealth, real € (" + name + ")", Width: 640, Height: 360},
			"Year", fan.Bands, sampleLines(fan.Samples), pr.lumpMarkers(pr.Years)...)
		refs := make([]SampleRef, len(fan.Samples))
		for k, s := range fan.Samples {
			refs[k] = SampleRef{Index: s.Index, Ruined: s.Ruined, Final: s.Wealth[len(s.Wealth)-1]}
		}
		res.Fans[i] = Fan{Name: name, SVG: svg, Samples: refs}
	})
	if len(res.Fans) == 0 {
		res.Note = "no return model available"
	}
	return res
}

// pickModel returns the named source matching want exactly.
func pickModel(sources []namedSource, want string) (namedSource, bool) {
	for _, ns := range sources {
		if ns.name == want {
			return ns, true
		}
	}
	return namedSource{}, false
}

// sampleLines extracts the plain wealth slices the chart consumes.
func sampleLines(samples []decumul.SamplePath) [][]float64 {
	out := make([][]float64, len(samples))
	for i, s := range samples {
		out[i] = s.Wealth
	}
	return out
}
