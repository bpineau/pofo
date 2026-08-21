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

// Draws is the exogenous randomness of one ensemble: a return sequence per
// path and, when the plan carries a Lifetime, the household lifespans drawn
// alongside them. Both depend only on the Source and the Lifetime, never on
// Capital, BufferYears or the spending rule, so a caller sweeping those can
// draw once and reuse the Draws across many SimulateOn calls instead of
// re-sampling at every point (which is where the sampling cost, a third of a
// path's total, would otherwise be paid again and again). Solve, Sweep1D and
// CapitalForRuin do exactly this internally, which is also what keeps their
// bisections free of Monte-Carlo noise.
type Draws struct {
	Returns []scenario.Sequence
	Lives   []Lives // nil when the plan has no Lifetime
}

// lifeStream offsets the lifespan RNG's second word so lifespans and returns
// are drawn from independent streams of the same seed.
const lifeStream = 1 << 40

// fallbackLifeSeed draws the lifespans of a Draws assembled by hand, without
// them, on a plan that has a Lifetime. Fixed, so the result stays reproducible
// and the solvers still see common random numbers.
const fallbackLifeSeed = 0x5eed11fe

// Simulate runs nPaths Monte-Carlo paths across workers goroutines. Each
// worker derives its RNG from (seed, workerID) so the result is reproducible
// for a fixed worker count.
func (p Plan) Simulate(nPaths, workers int, seed uint64) Ensemble {
	return p.SimulateOn(p.Draw(nPaths, workers, seed), workers)
}

// Draw samples nPaths return sequences from the plan's Source, and nPaths
// household lifespans from its Lifetime when it has one, with the same
// per-worker RNG split as Simulate so the draws are identical.
func (p Plan) Draw(nPaths, workers int, seed uint64) Draws {
	d := Draws{Returns: make([]scenario.Sequence, nPaths)}
	if p.Lifetime != nil {
		d.Lives = make([]Lives, nPaths)
	}
	var sampler lifeSampler
	if p.Lifetime != nil {
		sampler = p.Lifetime.sampler(p.Years)
	}
	forEachWorker(nPaths, workers, func(w int, lo func(func(int))) {
		rng := rand.New(rand.NewPCG(seed, uint64(w)+1))
		var lives *rand.Rand
		if d.Lives != nil {
			lives = rand.New(rand.NewPCG(seed, uint64(w)+1+lifeStream))
		}
		lo(func(i int) {
			d.Returns[i] = p.Source.Draw(rng)
			if lives != nil {
				d.Lives[i] = sampler.draw(lives)
			}
		})
	})
	return d
}

// drawLives samples nPaths lifespans alone, for a Draws that arrived without
// them.
func (p Plan) drawLives(nPaths, workers int, seed uint64) []Lives {
	out := make([]Lives, nPaths)
	sampler := p.Lifetime.sampler(p.Years)
	forEachWorker(nPaths, workers, func(w int, lo func(func(int))) {
		rng := rand.New(rand.NewPCG(seed, uint64(w)+1+lifeStream))
		lo(func(i int) { out[i] = sampler.draw(rng) })
	})
	return out
}

// SimulateOn runs the kernel on already-drawn paths (from Draw) across workers
// goroutines, without re-sampling. The kernel is deterministic, so the
// Ensemble is identical whatever the worker count. The plan may differ from
// the one that drew them in any field the draws do not depend on (capital,
// spending rule…).
//
// A Draws assembled by hand, with returns but no lifespans, on a plan that has
// a Lifetime, gets its lifespans drawn here from a fixed seed rather than
// being run as if the household were immortal: reproducible, and never a
// silent fixed horizon.
func (p Plan) SimulateOn(d Draws, workers int) Ensemble {
	lives := d.Lives
	if p.Lifetime != nil && len(lives) < len(d.Returns) {
		lives = p.drawLives(len(d.Returns), workers, fallbackLifeSeed)
	}
	paths := make([]PathResult, len(d.Returns))
	forEachWorker(len(d.Returns), workers, func(_ int, lo func(func(int))) {
		lo(func(i int) {
			var lv Lives
			if lives != nil {
				lv = lives[i]
			}
			paths[i] = p.runPath(d.Returns[i], lv)
		})
	})
	return Ensemble{Paths: paths, Years: p.Years}
}

// forEachWorker runs body on workers goroutines, each handed its worker index
// and a loop over the path indices it owns (a stride, so a worker's draws stay
// tied to its own RNG stream).
func forEachWorker(n, workers int, body func(w int, loop func(func(int)))) {
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			body(w, func(f func(int)) {
				for i := w; i < n; i += workers {
					f(i)
				}
			})
		}(w)
	}
	wg.Wait()
}

// CapitalForRuin returns the smallest starting capital in [lo, hi] whose
// ruin probability is at most target, by ~18 bisection steps. The same seed
// is reused at every capital so Monte-Carlo noise does not break
// monotonicity. Buffer.Years scales with NeedAnnual, not with capital, so
// only Capital varies between evaluations.
func (p Plan) CapitalForRuin(target, lo, hi float64, nPaths, workers int, seed uint64) float64 {
	shared := p.Draw(nPaths, workers, seed) // Capital affects neither the returns nor the lifespans
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

// RuinProb is the fraction of paths that ran out of money. With a
// Plan.Lifetime it is the fraction that ran out WHILE ALIVE, since the kernel
// stops at the household's end: everything built on it (Solve, Sweep1D,
// CapitalForRuin) therefore targets the alive-ruin without any further change,
// and a target that was demanding over a fixed horizon becomes far more
// demanding once the horizon can run to a hundred.
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
