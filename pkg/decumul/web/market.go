package web

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/scenario"
)

// MarketResult shows what each return model's MARKET looks like before any
// withdrawal: the growth of 1 real euro (percentile cone + a few individual
// draws) and the texture of its bear markets. It answers "what economy am I
// actually facing under this lens?" the way a long S&P 500 chart does for a
// US investor. An average path would be a visual lie (crashes land on
// different dates per path, so averaging erases them); individual draws and
// percentiles are the honest picture.
type MarketResult struct {
	Fans []Fan  `json:"fans"` // one market fan per planning model, fanModels order
	Note string `json:"note"`
}

// marketDraws is the number of paths behind the market bands and bear stats:
// enough for stable p5/p95 and a smooth cone, cheap to draw (no kernel).
const marketDraws = 400

// Market renders the market-alone view for each planning model.
func Market(pr Params, panel *scenario.Panel) MarketResult {
	sources := modelSources(pr, panel)
	var res MarketResult
	for _, name := range fanModels {
		ns, ok := pickModel(sources, name)
		if !ok {
			continue
		}
		res.Fans = append(res.Fans, Fan{Name: ns.name, SVG: marketFan(ns, pr.Years)})
	}
	if len(res.Fans) == 0 {
		res.Note = "no return model available"
	}
	return res
}

// marketFan draws one model's market cone: percentile bands of the cumulative
// real index across draws, three sample draws (worst / median / best by
// terminal value), and a caption with the model's bear texture.
func marketFan(ns namedSource, years int) string {
	rng := rand.New(rand.NewPCG(11, 7))
	indexes := make([][]float64, marketDraws)
	depths := make([]float64, marketDraws)
	spells := make([]float64, marketDraws)
	for i := range indexes {
		indexes[i] = cumIndex(ns.source.Draw(rng))
		depths[i], spells[i] = bearTexture(indexes[i])
	}

	// Percentile bands of the index, year by year.
	steps := years + 1
	bands := make([][]float64, len(fanPercentiles))
	for p := range bands {
		bands[p] = make([]float64, steps)
	}
	col := make([]float64, len(indexes))
	for y := range steps {
		for i, idx := range indexes {
			col[i] = at(idx, y)
		}
		sort.Float64s(col)
		for p, q := range fanPercentiles {
			bands[p][y] = col[int(q*float64(len(col)-1))]
		}
	}

	// Three representative draws: worst, median and best terminal value.
	order := make([]int, len(indexes))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		ia, ib := indexes[order[a]], indexes[order[b]]
		return ia[len(ia)-1] < ib[len(ib)-1]
	})
	samples := [][]float64{
		indexes[order[0]],
		indexes[order[len(order)/2]],
		indexes[order[len(order)-1]],
	}

	title := fmt.Sprintf("Market alone, growth of 1 real € (%s)", ns.name)
	// A market index at zero is a wipeout year the model admits (an extreme
	// fat-tail draw), not a plan ruin: label the baseline accordingly.
	svg := darkFan(chart.Options{Title: title, Width: 640, Height: 320,
		Style: chart.Style{ZeroLabel: "wipeout · 0"}}, "Year", bands, samples)
	// The bear texture: the typical draw's worst bear and the 1-in-20 draw's.
	sort.Float64s(depths)
	sort.Float64s(spells)
	q := func(s []float64, p float64) float64 { return s[int(p*float64(len(s)-1))] }
	caption := fmt.Sprintf("worst bear per draw · typical: −%.0f%% and %.0fy under water · 1-in-20: −%.0f%% and %.0fy",
		q(depths, 0.50)*100, q(spells, 0.50), q(depths, 0.95)*100, q(spells, 0.95))
	return svg + fmt.Sprintf(`<div class="mcap">%s</div>`, caption)
}

// cumIndex compounds a return sequence into a cumulative index starting at 1.
func cumIndex(seq scenario.Sequence) []float64 {
	out := make([]float64, len(seq)+1)
	out[0] = 1
	for i, r := range seq {
		v := out[i] * (1 + r)
		if v < 0 {
			v = 0
		}
		out[i+1] = v
	}
	return out
}

// at reads a series at y, holding the last value past the end.
func at(s []float64, y int) float64 {
	if y >= len(s) {
		return s[len(s)-1]
	}
	return s[y]
}

// bearTexture returns the worst peak-to-trough depth of an index path and the
// longest spell spent below a prior peak (in periods).
func bearTexture(idx []float64) (depth, spell float64) {
	peak := idx[0]
	run, worstRun := 0, 0
	for _, v := range idx {
		if v >= peak {
			peak = v
			run = 0
			continue
		}
		run++
		if run > worstRun {
			worstRun = run
		}
		if peak > 0 {
			if d := 1 - v/peak; d > depth {
				depth = d
			}
		}
	}
	return depth, float64(worstRun)
}
