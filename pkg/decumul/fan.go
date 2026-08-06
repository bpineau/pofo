package decumul

import (
	"sort"

	"github.com/bpineau/pofo/pkg/metrics"
)

// SamplePath is one individual simulated wealth path kept for display, so the
// user can see the texture of a single retirement (a couple of bear markets,
// the slope of the drawdown) rather than only the aggregate bands.
type SamplePath struct {
	Wealth []float64
	Ruined bool
}

// WealthFan summarises an Ensemble's wealth paths over time: percentile bands
// per year (the aleatory spread within one model) plus a handful of
// representative individual paths spanning the terminal-wealth distribution.
type WealthFan struct {
	Years   int
	Pcts    []float64    // percentile levels, ascending (e.g. 0.05, 0.50, 0.95)
	Bands   [][]float64  // Bands[p][year] = the Pcts[p] quantile of wealth that year
	Samples []SamplePath // representative paths, ascending by terminal wealth
}

// Fan computes the wealth fan: for each year the requested quantiles across all
// paths, and nSamples individual paths drawn at evenly spaced ranks of terminal
// wealth (so the set spans the worst, the typical and the best outcomes, ruin
// paths included). pcts should be ascending.
func (e Ensemble) Fan(pcts []float64, nSamples int) WealthFan {
	fan := WealthFan{Years: e.Years, Pcts: pcts}
	if len(e.Paths) == 0 {
		return fan
	}
	steps := len(e.Paths[0].Wealth)

	// Percentile bands, computed column by column (one year at a time).
	fan.Bands = make([][]float64, len(pcts))
	for p := range fan.Bands {
		fan.Bands[p] = make([]float64, steps)
	}
	col := make([]float64, len(e.Paths))
	for y := range steps {
		for i, path := range e.Paths {
			col[i] = path.Wealth[y]
		}
		q := metrics.Quantiles(col, pcts...)
		for p := range pcts {
			fan.Bands[p][y] = q[p]
		}
	}

	fan.Samples = sampleByTerminal(e.Paths, nSamples)
	return fan
}

// sampleByTerminal returns n paths picked at evenly spaced ranks of terminal
// wealth, ascending, so the selection spans the outcome distribution.
func sampleByTerminal(paths []PathResult, n int) []SamplePath {
	if n <= 0 || len(paths) == 0 {
		return nil
	}
	order := make([]int, len(paths))
	for i := range order {
		order[i] = i
	}
	terminal := func(p PathResult) float64 { return p.Wealth[len(p.Wealth)-1] }
	sort.Slice(order, func(a, b int) bool { return terminal(paths[order[a]]) < terminal(paths[order[b]]) })

	if n > len(order) {
		n = len(order)
	}
	out := make([]SamplePath, n)
	for k := range out {
		rank := 0
		if n > 1 {
			rank = k * (len(order) - 1) / (n - 1)
		}
		p := paths[order[rank]]
		out[k] = SamplePath{Wealth: p.Wealth, Ruined: p.Ruined}
	}
	return out
}
