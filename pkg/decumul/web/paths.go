package web

import (
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Fan is one model's wealth fan chart, named for its column in the strip.
type Fan struct {
	Name string `json:"name"`
	SVG  string `json:"svg"`
}

// PathsResult is the set of wealth fan charts, one per planning model: the
// picture of the market each simulation actually produces (bands + a few
// sample/ruin paths), the answer to "is this an 80% mega-crash or March-2020
// repeated?". The four planning models are shown side by side so the central
// case and the successively grimmer stresses can be compared at a glance.
type PathsResult struct {
	Fans []Fan  `json:"fans"`
	Note string `json:"note"`
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
	base := pr.plan()
	base.Monthly = false

	var res PathsResult
	for _, name := range fanModels {
		ns, ok := pickModel(sources, name)
		if !ok {
			continue
		}
		p := base
		p.Source = ns.source
		fan := p.Simulate(pr.NPaths, simWorkers, 7).Fan(fanPercentiles, 8)
		svg := darkFan(
			chart.Options{Title: "Simulated wealth, real € (" + ns.name + ")", Width: 640, Height: 360},
			"Year", fan.Bands, sampleLines(fan.Samples))
		res.Fans = append(res.Fans, Fan{Name: ns.name, SVG: svg})
	}
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
