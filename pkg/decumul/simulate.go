package decumul

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Ensemble is the result of many simulated paths sharing a horizon.
type Ensemble struct {
	Paths []PathResult // one per simulated path, in draw order
	Years int          // the plan's horizon, the length of every path
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
	Returns []scenario.Sequence // one real-return path per simulated path
	Lives   []Lives             // nil when the plan has no Lifetime
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
	// The source's rng-independent setup (a data-driven source collapses its
	// panel into one weighted history) is hoisted out of the path loop; the
	// draws are byte-identical, the rng being consumed in the same order.
	src := scenario.Prepare(p.Source)
	forEachWorker(nPaths, workers, func(w int, lo func(func(int))) {
		rng := rand.New(rand.NewPCG(seed, uint64(w)+1))
		var lives *rand.Rand
		if d.Lives != nil {
			lives = rand.New(rand.NewPCG(seed, uint64(w)+1+lifeStream))
		}
		lo(func(i int) {
			d.Returns[i] = src.Draw(rng)
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
	p, lives := p.forRun(d, workers)
	paths := make([]PathResult, len(d.Returns))
	// Every path's two series (Wealth and Spend) come out of ONE arena rather
	// than an allocation per path: a render simulates thousands of paths, and
	// the allocator and the collector, not the kernel, were where that time
	// went. The windows are disjoint, so the workers stay independent.
	stride := seriesLen(p.Years)
	arena := make([]float64, len(d.Returns)*stride)
	forEachPath(len(d.Returns), workers, func(_ int, lo func(func(int))) {
		lo(func(i int) {
			var lv Lives
			if lives != nil {
				lv = lives[i]
			}
			paths[i] = p.runPath(d.Returns[i], lv, arena[i*stride:(i+1)*stride:(i+1)*stride])
		})
	})
	return Ensemble{Paths: paths, Years: p.Years}
}

// forRun is what a driver runs d under: the plan with the per-year tables its
// paths share (withTables), and the lifespans, d's own or, for a Draws without
// them on a plan with a Lifetime, the fixed-seed fallback draw.
func (p Plan) forRun(d Draws, workers int) (Plan, []Lives) {
	lives := d.Lives
	if p.Lifetime != nil && len(lives) < len(d.Returns) {
		lives = p.drawLives(len(d.Returns), workers, fallbackLifeSeed)
	}
	return p.withTables(), lives
}

// RuinProb is Simulate(nPaths, workers, seed).RuinProb(), bit for bit, without
// building the Ensemble (see RuinProbOn).
func (p Plan) RuinProb(nPaths, workers int, seed uint64) float64 {
	return p.RuinProbOn(p.Draw(nPaths, workers, seed), workers)
}

// RuinProbOn is SimulateOn(d, workers).RuinProb(), bit for bit, without
// building the Ensemble. The bisections (Solve, CapitalForRuin) and the ruin
// grids (Sweep2D, a frontier, a sensitivity table) read nothing else from a
// run, yet each SimulateOn allocates every path's series, about a megabyte at
// two thousand thirty-year paths: at eighteen runs a solve, the heap churn,
// not the kernel, was the cost. Here each worker runs its paths through one
// reused scratch window and keeps only a count.
func (p Plan) RuinProbOn(d Draws, workers int) float64 {
	n := len(d.Returns)
	if n == 0 {
		return 0
	}
	p, lives := p.forRun(d, workers)
	counts := make([]int, max(workers, 1))
	forEachPath(n, workers, func(w int, lo func(func(int))) {
		// The monthly kernel accumulates into Spend, so the window is cleared
		// before every path, as a fresh arena window would be.
		buf := make([]float64, seriesLen(p.Years))
		ruined := 0
		lo(func(i int) {
			clear(buf)
			var lv Lives
			if lives != nil {
				lv = lives[i]
			}
			if p.runPath(d.Returns[i], lv, buf).Ruined {
				ruined++
			}
		})
		counts[w] = ruined
	})
	ruined := 0
	for _, c := range counts {
		ruined += c
	}
	return float64(ruined) / float64(n)
}

// ruinAbove reports RuinProbOn(d, workers) > target, the one bit a bisection
// step reads, and stops running paths as soon as the count settles it: once
// enough paths have failed to exceed the target whatever the rest do, or too
// few remain to reach it. The count does not depend on which paths ran first,
// so the answer is exactly RuinProbOn's, and a step far from the crossing
// (where half a bisection's steps land) costs a fraction of a full run.
func (p Plan) ruinAbove(d Draws, workers int, target float64) bool {
	n := len(d.Returns)
	above := func(ruined int64) bool { return float64(ruined)/float64(n) > target }
	if n == 0 {
		return 0 > target
	}
	p, lives := p.forRun(d, workers)
	// The shared tallies are published every few paths rather than at each
	// one, to keep the goroutines off one contended cache line. A batch's
	// failures are published before the batch itself, and the test reads
	// them the other way round, so the bounds it forms can only be wider than
	// the truth (lower <= failures of the whole run <= upper): never a
	// premature verdict.
	const batch = 8
	var ruined, done atomic.Int64
	var settled atomic.Bool
	publish := func(r, k int64) {
		ruined.Add(r)
		finished := done.Add(k)
		lower := ruined.Load()
		if above(lower) || !above(lower+int64(n)-finished) {
			settled.Store(true)
		}
	}
	forEachPath(n, workers, func(_ int, lo func(func(int))) {
		buf := make([]float64, seriesLen(p.Years))
		var r, k int64
		lo(func(i int) {
			if settled.Load() {
				return
			}
			clear(buf)
			var lv Lives
			if lives != nil {
				lv = lives[i]
			}
			if p.runPath(d.Returns[i], lv, buf).Ruined {
				r++
			}
			if k++; k == batch {
				publish(r, k)
				r, k = 0, 0
			}
		})
		if k > 0 {
			publish(r, k)
		}
	})
	return above(ruined.Load())
}

// aggregateWorkers is how many goroutines a statistic over n independent
// units of work (paths, or path-years) spreads across: one per core, but none
// beyond one per few hundred units, below which starting them costs more than
// they save.
func aggregateWorkers(n int) int {
	return max(1, min(runtime.GOMAXPROCS(0), n/256))
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
		if q.ruinAbove(shared, workers, target) {
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

// forEachPath runs body on workers goroutines like forEachWorker, but the
// paths are not dealt in a fixed stride: each goroutine claims the next chunk
// of consecutive indices from a shared counter until none is left. It is for
// the loops whose result does not depend on which goroutine ran a path (the
// kernel is a pure function of its draw), never for Draw, whose values are
// tied to a worker's RNG stream. A fixed stride makes every run wait for its
// slowest goroutine, and on a machine that mixes fast and slow cores (or
// shares them with other requests) that goroutine finished far behind the
// others; claimed chunks keep every core busy to the end, and consecutive
// indices keep two goroutines off the same cache lines of the results.
func forEachPath(n, workers int, body func(w int, loop func(func(int)))) {
	workers = max(workers, 1)
	// A chunk small enough to balance the tail, large enough that the
	// counter is touched a few hundred times a run rather than once a path.
	chunk := int64(max(1, min(64, n/(workers*16))))
	var next atomic.Int64
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body(w, func(f func(int)) {
				for {
					hi := next.Add(chunk)
					lo := hi - chunk
					if lo >= int64(n) {
						return
					}
					for i := lo; i < min(hi, int64(n)); i++ {
						f(int(i))
					}
				}
			})
		}()
	}
	wg.Wait()
}
