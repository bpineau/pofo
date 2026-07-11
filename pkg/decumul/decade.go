package decumul

import (
	"sort"

	"github.com/bpineau/pofo/pkg/metrics"
)

// DecadeBucket is one slice of the ensemble, grouped by the market return the
// path realised over its first decade. It quantifies sequence-of-returns risk,
// the core insight of the withdrawal literature (Bengen; Kitces): two
// retirements with the same average return but a different ordering have very
// different fates, and the ordering that matters is the first decade's.
type DecadeBucket struct {
	LoRet, HiRet float64 // first-decade annualized real return range of the bucket
	RuinProb     float64 // share of the bucket's paths that ran out
	TerminalP50  float64 // median terminal real wealth inside the bucket
	Paths        int     // number of paths in the bucket
}

// DecadeBuckets sorts the paths by their first-decade annualized real market
// return (PathResult.Ret10) and splits them into n equal-count buckets, worst
// decade first. It answers "how much of my ruin risk is decided in the first
// ten years?": the worst bucket's ruin is typically several times the
// headline figure while the best buckets almost never fail. Returns nil when
// the ensemble holds fewer than n paths.
func (e Ensemble) DecadeBuckets(n int) []DecadeBucket {
	if n <= 0 || len(e.Paths) < n {
		return nil
	}
	idx := make([]int, len(e.Paths))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool {
		return e.Paths[idx[a]].Ret10 < e.Paths[idx[b]].Ret10
	})
	out := make([]DecadeBucket, 0, n)
	for b := range n {
		lo, hi := b*len(idx)/n, (b+1)*len(idx)/n
		var ruined int
		terminals := make([]float64, 0, hi-lo)
		for _, i := range idx[lo:hi] {
			p := e.Paths[i]
			if p.Ruined {
				ruined++
			}
			terminals = append(terminals, p.Wealth[len(p.Wealth)-1])
		}
		out = append(out, DecadeBucket{
			LoRet:       e.Paths[idx[lo]].Ret10,
			HiRet:       e.Paths[idx[hi-1]].Ret10,
			RuinProb:    float64(ruined) / float64(hi-lo),
			TerminalP50: metrics.Quantiles(terminals, 0.50)[0],
			Paths:       hi - lo,
		})
	}
	return out
}
