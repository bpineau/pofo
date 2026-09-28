# FIRE Monte Carlo: performance record

What the profiler found in `pkg/decumul` and `pkg/decumul/web`, what was
changed, what was refused, and how a change is proven harmless. The design
itself is the godoc (`go doc ./pkg/decumul`, section "Performance and
determinism").

## The invariant

Determinism is a feature: a shared `/view` or `-fire` URL must reproduce byte
for byte, and the solvers rely on common random numbers. So every optimisation
here is held to the same bits, never to a tolerance:

- untouchable: the float summation order, the RNG calls and their order, the
  seeding scheme, and the fixed stride that deals `Plan.Draw`'s paths to the
  workers' RNG streams (a draw's value is tied to its worker);
- free: which goroutine runs a path through the kernel, and in which order,
  since the kernel is a pure function of its draw and the aggregates index
  their per-path figures by path.

`TestDeterminismPlans` and `TestDeterminismSources` pin the bits of everything
the package produces; `TestShortcutsMatchEnsemble` pins each shortcut below to
the full computation it stands for. A change that moves a digest is a model
change, not an optimisation.

## Proving a change harmless

Beyond the suite, a change to the engine is checked end to end: a JSON dump of
all fifteen `/api/*` endpoints over sixteen parameter sets (every return model,
monthly, ABW, both guardrails, VPW, bounded, ratchet, annuity, envelopes,
glidepath, CAPE, no panel), hashed, before and after, at `GOMAXPROCS` 1, 4 and
10; and the same at the kernel level (draws, per-path series, `Outcome`, `Fan`,
`LifeOutcome`, every `Solve` axis, `CapitalForRuin`, both sweeps, a `Draws`
without lifespans) over twenty-one plans at worker counts 1, 3 and 8. Both
were identical for every change below, also under `-race`. The harness is a
throwaway `_test.go` over the exported API, run once on a worktree of the
previous commit and once on the change.

## Where the time went (2026-09, M1 Max, 8 performance + 2 efficiency cores)

Profiling the endpoints, not the kernel benchmarks, is what showed it:

- **Allocation, not arithmetic.** Every `SimulateOn` allocated all its paths'
  series (about 1 MB at 2 000 thirty-year paths), and the solvers call it
  eighteen times a bisection: on one core, `madvise` (the heap returning pages
  to the OS and taking them back) was 55 % of `/api/sim`. The bisections, the
  ruin grids and the sweeps read only the ruin, or the ruin and the median
  terminal: `RuinProbOn`, `ruinAbove` and `ruinAndMedianOn` run the kernel
  through one reused scratch window per goroutine and keep a count.
- **A fixed stride waits for its slowest core.** Path `i` went to worker
  `i mod W`, so each barrier waited for the goroutines on the efficiency cores:
  ten cores ran the kernel 4.2 times faster than one. The drivers over drawn
  paths now hand out chunks of consecutive paths from an atomic counter
  (`forEachPath`); `Plan.Draw` keeps the stride.
- **A bisection step reads one bit.** `ruinAbove` stops as soon as the count
  settles `ruin > target` (enough failures, or too few paths left to reach
  it). The count does not depend on the order, so the verdict is exact; it
  saves the steps far above the crossing, about half of them.
- **Per-year figures shared by every path.** Without a `Lifetime`, a year's
  income and its present value are the same on every path; the drivers
  tabulate them once per run (`Plan.withTables`) instead of rescanning the
  cashflows each year of each path.
- **Serial aggregates.** `Outcome` and `Fan` cost more than the simulation
  they summarised (0.58 and 1.47 ms against 1.1 ms). Their per-path and
  per-year work is now spread over the cores, and the drawdown walk divides
  once per underwater episode rather than once per point (under one peak the
  deepest loss is the lowest trough, division and subtraction being monotone
  once rounded).
- **Independent work queued in series.** An endpoint's simulations (the
  strip's six models, the curves' twenty-five solves, the frontier's models,
  the sensitivity levers) now run concurrently, and the endpoints that drew
  the same paths several times (`solvemenu` up to thirteen) draw once and ask
  every question `On` them (`Plan.SolveOn`).

Totals over the sixteen parameter sets and fifteen endpoints (the sum of one
render per set), before and after, results identical:

| GOMAXPROCS | before | after | change |
|---|---|---|---|
| 1 (CPU work) | 24.0 s | 18.2 s | -24 % |
| 4 (a small server) | 8.4 s | 5.1 s | -40 % |
| 10 | 7.1 s | 2.75 s | -61 % |

Benchmarks at ten cores: `BenchmarkSolve` 11.9 to 6.6 ms/op (23.9 to 0.65
MB/op), `BenchmarkComputeParametric` (`/api/sim`) 16.1 to 6.4 ms/op,
`BenchmarkModelsStrip` 92 to 38 ms/op, `/api/curves` about 210 to 80 ms,
`BenchmarkOutcome` 0.59 to 0.26 ms/op, `BenchmarkFan` 1.47 to 0.60 ms/op.

## What was refused

- **Per-path monotonic pruning of the bisections.** If each path's ruin were
  monotone in the solved value, a step could skip the paths an earlier step
  already settled, most of the work. It is not guaranteed: the buffer target
  scales with the spending, so a higher spend carves out a larger buffer that
  can spare a path a sale at the bottom. Refused, since it would change
  results.
- **A cheaper kernel for the ruin-only drivers** (skipping `Ret10`'s `Pow`,
  stopping a path at its first failure): about 4 %, for a second contract on
  `PathResult`. Not taken.
- **Earlier findings (2026-09-10):** blocked rather than striped assignment
  inside `Draw` (nothing, and it would move the draws), an arena for the
  drawn sequences behind a `DrawInto` on `Source` (1 to 2 %, not worth the
  public API), hoisting the regime source's per-period sigma (the compiler
  already had).

## What is left

The kernel is a chain of dependent floating-point operations at about 25 ns a
path-year, with three divisions per taxed sale. Further CPU gains would need a
structure-of-arrays kernel running many paths in lockstep, which Go cannot
vectorise portably; the next large gain would be not recomputing, i.e. caching
responses per parameter set, which touches the web layer's no-stored-state
rule and is a product decision rather than an optimisation.
