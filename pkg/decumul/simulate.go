package decumul

import (
	"math/rand/v2"
	"sync"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Ensemble is the result of many simulated paths sharing a horizon.
type Ensemble struct {
	Paths []PathResult
	Years int
}

// Simulate runs nPaths Monte-Carlo paths across workers goroutines. Each
// worker derives its RNG from (seed, workerID) so the result is
// reproducible for a fixed worker count.
func (p Plan) Simulate(nPaths, workers int, seed uint64) Ensemble {
	return p.SimulateOn(p.DrawPaths(nPaths, workers, seed), workers)
}

// DrawPaths samples nPaths return sequences from the plan's Source, with the
// same per-worker RNG split as Simulate so the draws are identical. The
// sequences depend only on the Source (not on Capital, BufferYears, the
// spending rule…), so a caller sweeping those parameters can draw once and
// reuse the sequences across many SimulateOn calls instead of re-sampling at
// every point (which is where the sampling cost, a third of a path's total,
// would otherwise be paid again and again). Sweep1D and CapitalForRuin do
// exactly this internally.
func (p Plan) DrawPaths(nPaths, workers int, seed uint64) []scenario.Sequence {
	if workers < 1 {
		workers = 1
	}
	seqs := make([]scenario.Sequence, nPaths)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(seed, uint64(w)+1))
			for i := w; i < nPaths; i += workers {
				seqs[i] = p.Source.Draw(rng)
			}
		}(w)
	}
	wg.Wait()
	return seqs
}

// SimulateOn runs the kernel on already-drawn sequences (from DrawPaths)
// across workers goroutines, without re-sampling. RunPath is deterministic, so
// the Ensemble is identical whatever the worker count. The plan may differ
// from the one that drew the sequences in any non-Source field (capital,
// horizon, spending rule…); the draws depend only on the Source.
func (p Plan) SimulateOn(seqs []scenario.Sequence, workers int) Ensemble {
	if workers < 1 {
		workers = 1
	}
	paths := make([]PathResult, len(seqs))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := w; i < len(seqs); i += workers {
				paths[i] = p.runPath(seqs[i])
			}
		}(w)
	}
	wg.Wait()
	return Ensemble{Paths: paths, Years: p.Years}
}

// CapitalForRuin returns the smallest starting capital in [lo, hi] whose
// ruin probability is at most target, by ~18 bisection steps. The same seed
// is reused at every capital so Monte-Carlo noise does not break
// monotonicity. Buffer.Years scales with NeedAnnual, not with capital, so
// only Capital varies between evaluations.
func (p Plan) CapitalForRuin(target, lo, hi float64, nPaths, workers int, seed uint64) float64 {
	shared := p.DrawPaths(nPaths, workers, seed) // Capital does not affect the Source
	for i := 0; i < 18; i++ {
		mid := (lo + hi) / 2
		q := p
		q.Capital = mid
		if q.SimulateOn(shared, workers).RuinProb() > target {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// RuinProb is the fraction of paths that ran out of money.
func (e Ensemble) RuinProb() float64 {
	if len(e.Paths) == 0 {
		return 0
	}
	n := 0
	for _, r := range e.Paths {
		if r.Ruined {
			n++
		}
	}
	return float64(n) / float64(len(e.Paths))
}
