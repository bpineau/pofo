# The library face: audit and design

Status: spec, 2026-09-26. Implementation plan at the end, one milestone per
commit, each shippable on its own.

pofo has four faces (AGENTS.md), and the first one, the finance LIBRARY, is the
one a reader of `README.md` sees least of. This document audits what the
library actually offers, names where its API is inconsistent or thin, and
specifies the additions that make it a usable, consistent toolkit for
dissecting assets and portfolios, without breaking the two sibling programs
that pin it by tag.

## 1. Audit

### 1.1 What is there (measured with `go doc`, 2026-09-26)

| Package | Exported symbols | Runnable examples | What it is |
|---|---|---|---|
| `marketdata` | 35 top-level + 17 `Client` methods | 21 | resolution (ticker, ISIN, alias, catalog), daily/intraday/latest quotes, FX, fees, hygiene, simdata, `Align`, `SampleAt`, data doctor |
| `metrics` | 41 | 9 | `Compute` (CAGR, vol, Sharpe, Sortino, Ulcer, MDD, TTR, beta, CWARP, skew, kurtosis), IRR/TWR with flows, rolling CAGR/vol/Sharpe/Sortino/Ulcer, drawdown episodes, longest underperformance, variance ratio, attribution (Euler risk shares), moments, quantiles |
| `portfolio` | 9 types, 5 funcs | 7 | file format (`Parse`), `Build`, `Simulate` (rebalancing, fees, flows, leverage, per-holding contributions) |
| `optimize` | 10 objectives, bounds and limits, Black-Litterman (`ImpliedReturns`, `Posterior`) | 6 | weights |
| `suggest` | 25 | 8 | look-through splits (class, geography, currency, sector, duration), regime/factor coverage, redundancy, gap ranking |
| `scenario` | 11 sources | 0 | real-return paths: parametric t, block/stationary/pooled bootstrap, cohorts, Markov regimes, glidepath |
| `decumul` | 42 types, `Plan.Simulate/Solve/Sweep1D/Sweep2D`, `Ensemble` readers | 5 | FIRE engine, stochastic lifetime, annuities, taxes |
| `replay` | 8 | 2 | withdrawal rules replayed on real history |
| `simgen` | 26 | 5 | reconstruction engine (composites, cap-weighted, TSMOM, donor chains, Treasury pricing) and its audit |
| `compare` | `Compute`, `Sweep`, opaque `Comparison` | 3 | the CLI/web pipeline: fetch, build, simulate, common window, nominal and real stats, HTML page |
| `chart` | 20 renderers | 9 | SVG and terminal charts |
| `datasets` | catalog, simdata, refdata, panels | 7 | the bundled data |

So the reading "fetch, time series, charts and a few basic statistics" is
wrong. It is what the README's library chapter conveys: 100 lines out of 989,
a package listing and five calls.

### 1.2 Who consumes it

- `../finador` uses `marketdata` (`NewClient`, `Fetch`, `Latest`, `Quote`,
  `Series`, `Point`, `FetchOptions`, `QuoteOptions`, the two sentinel errors),
  `metrics` (`Report`, `Window`, `Flow`, `TWR`, `FlowReturns`, `Sharpe`,
  `Sortino`, `Volatility`, `MaxDrawdown`, `Episode`, `Annualize`) and `chart`
  (`Line`, `SVG`, `PiePalette`, `TimePoint`). It feeds `metrics` with its OWN
  valuation series, not with `marketdata.Series`.
- `../locador` uses `bookmd`, `epub`, `opds` only.

Consequence: the slice-based `metrics` functions are the contract a sibling
depends on, and they must stay. Everything below is additive to them.

### 1.3 Findings

**F1. Four shapes for "a dated series", no bridge.** `marketdata.Series`
holds `[]Point{Date, Close}`; `metrics` takes parallel `(dates, values)`
slices; `chart.Series` has its own `Dates, Values`; `scenario.Sequence` is a
bare `[]float64`; `simgen.Frame` is `Dates` plus a returns map. `Series` has
no `Dates()`, `Values()` or `Returns()`, so every consumer unzips `Points` by
hand before it can compute anything, and there is no constructor for a
consumer that owns its own data (finador builds a `Series` literal).

**F2. Duplicated or colliding small types.** `portfolio.Holding` (weight as
fraction, `RawWeight` in percent), `portfolio.Asset`, `suggest.Holding`,
`compare.SweepHolding` and `datasets.Asset` all describe a position.
`metrics.Flow` (a dated cash event) and `portfolio.Flow` (a periodic rule)
share a name for different things; `metrics.Window` and `optimize.Window`
likewise. `suggest.Correlation` is the only Pearson correlation in the tree,
in the package least likely to be looked in.

**F3. Units are a trap and nothing in the type system says so.** Fees are
percent per year in `portfolio` and `marketdata.Client.Fees`, fractions in
`simgen`; weights are fractions in memory and percent in files and
`RawWeight`; rates are percent LEVELS; `metrics.Stats.Ulcer` is in percent
points while every sibling field is a fraction. AGENTS.md calls it "the
number one trap" and it is documented, not prevented.

**F4. Analysis primitives an analyst expects are missing.** No correlation or
covariance MATRIX (optimize computes one privately); no calendar-period
returns table (monthly, yearly), the raw material of every heatmap and every
"annual returns" row (`metrics` resamples month-ends privately for the
variance ratio; the golden tests carry their own `calYearDaily`); no rolling
beta or rolling correlation; no historical VaR/CVaR though `Quantiles` is the
raw material; no way to get the ALIGNED daily returns of a portfolio's
holdings without knowing that `marketdata.Align` forward-fills ZEROS before a
series' first quote and that `metrics.Returns` must be mapped over it.

**F5. The high-level path is file-first and ends in an opaque value.**
`compare.Compute` takes `[]*portfolio.Spec`, which `Parse` produces from the
portfolio-file dialect; a Spec can be assembled by hand but `Holding` needs
both `Weight` and `RawWeight` set consistently and there is no constructor
that does what `Parse` does (`Single` covers one line). `Comparison` has
unexported fields and exposes `Columns`, `StatRows`, `HTMLPage`: the numbers
inside (per-holding series, correlation, attribution, composition) are
computed and then reachable only through the HTML page model. A consumer that
wants the numbers of "60/40 IWDA/AGGH in EUR since 2010" is six calls and
three documented traps away from them.

**F6. Public plumbing blurs the API.** `simgen.AnchorTrend`, `AnchorStart`,
`SeriesFromFrame`, `CapWeights` have no caller outside their package;
`simgen.DBiProjection`, `DBiReplication`, `Rebase`, `marketdata.WarmupIDs`,
`WriteSimdata` serve the data generators under `cmd/`. Listed next to
`Compute` and `Fetch` in a `go doc` page, they make the library look larger
and less designed than it is.

**F7. Examples are uneven.** `scenario` has none; `decumul` has five for
forty-two types; `marketdata` has twenty-one. The root `doc.go` is good and
nobody is pointed to it.

**F8. The README chapter is a table of contents, not a guide.** It lists
packages and shows one fetch, one `Compute`, one chart and the three-call
pipeline. Nothing shows what the library answers (an asset's profile, a
portfolio's risk budget, a FIRE plan, a reconstruction), and nothing ties a
snippet to a runnable example, so it can rot.

What is NOT a finding: the layering. `metrics` imports nothing of pofo,
`marketdata` only `datasets`, `optimize` only `metrics`, `portfolio` only
`marketdata` and `optimize`, `compare` sits on top of all of them. That is the
right shape and the design keeps it.

## 2. Principles

1. **Additive.** No existing exported signature changes. A sibling on
   v0.2.80 compiles on the next tag without edits.
2. **`metrics` stays slice-based and dependency-free.** It is the math; it
   must accept a valuation series finador built itself. The bridge to it
   lives on `marketdata.Series`, not the other way round.
3. **One convention per package, stated once, and every NEW API in
   fractions.** Percent stays where the package already speaks percent
   (portfolio files and `portfolio` fees, `Client.Fees`, rate levels); new
   functions take and return fractions and say so in their godoc.
4. **The high-level API returns numbers, not pages.** Presentation
   (`compare`, `report`, `chart`) consumes the numbers layer; it does not
   own it.
5. **Every new function has a runnable example, and the README snippets
   ARE those examples.** A snippet that does not compile under `go test`
   does not go in the README.
6. **Hard to misuse beats short.** A function that can silently return a
   wrong number (Align's zero fill) gets a strict sibling that errors
   instead, and the doc points at it.

## 3. The design

Six milestones. M1 to M3 are the library; M4 makes the CLI consume it; M5
and M6 are hygiene and documentation. Each is one commit, gated by `make
check` and `make golden`, and M4 additionally by a byte-for-byte report
diff.

### M1. `marketdata`: the series speaks to the rest of the tree (shipped 2026-09-26)

```go
// NewSeries builds a series from a consumer's own data. dates must be
// strictly ascending; they are normalized to 00:00 UTC. values are closes
// (or levels: a rate series is welcome) and must be finite. symbol is free.
func NewSeries(symbol string, dates []time.Time, values []float64) (*Series, error)

func (s *Series) Len() int
func (s *Series) Dates() []time.Time  // a fresh slice, one per point
func (s *Series) Values() []float64   // a fresh slice of closes
func (s *Series) Returns() []float64  // simple period returns, len-1, = metrics.Returns(s.Values())

// Rebase returns a copy scaled so its first close is base (100 for an
// index). Metadata (symbol, currency, source, SimulatedBefore, Junctions,
// Dividends) is carried over; dividends are NOT rescaled, they are cash.
func (s *Series) Rebase(base float64) *Series

// Frequency is a calendar period in months.
type Frequency int
const (
    Monthly   Frequency = 1
    Quarterly Frequency = 3
    Yearly    Frequency = 12
)

// Resample keeps the last close of each calendar period, dated on that
// close (a month-END series, the convention every bundled monthly anchor
// follows). The last period is kept even when it is not complete: the caller
// reads Last().Date to know. Junctions and Dividends that fall on a dropped
// date are dropped with it.
func (s *Series) Resample(f Frequency) *Series

// CommonWindow is the latest first quote and the earliest last quote across
// the series: the window on which all of them are defined. ok is false when
// the window is empty or list is.
func CommonWindow(list ...*Series) (start, end time.Time, ok bool)

// Aligned is several series on one calendar, forward-filled between their
// own quotes. Levels[i] parallels IDs[i] and Dates.
type Aligned struct {
    Dates  []time.Time
    IDs    []string
    Levels [][]float64
}

// AlignSeries is Align with its trap closed: from must be at or after every
// series' first quote, and a zero from means CommonWindow's start. It returns
// an error, naming the series, instead of forward-filling zeros. to zero is
// open-ended.
func AlignSeries(list []*Series, from, to time.Time) (*Aligned, error)

func (a *Aligned) Returns() [][]float64     // [asset][period], len(Dates)-1 each
func (a *Aligned) Series(i int) *Series      // column i as a Series (symbol = IDs[i])
```

`Align` keeps its contract (portfolio.Simulate relies on it and computes its
own start) and its godoc gains one line pointing at `AlignSeries`. `Trim`
stays the window operation (a method would duplicate it).

Tests: unit tests per method (empty, single point, unsorted dates rejected,
resample month-end on a holiday month-end, junction dropped/kept,
`AlignSeries` error names the offender). Examples: `ExampleNewSeries`,
`ExampleSeries_Resample`, `ExampleAlignSeries`, `ExampleCommonWindow`.

### M2. `metrics`: the matrices, the calendar table, the tails

All slice-based, fractions in and out, per-period inputs (daily unless said).

```go
// Corr is the Pearson correlation of two equal-length samples. suggest.Correlation
// becomes a one-line call to it and is kept.
func Corr(a, b []float64) float64

// CorrelationMatrix and Covariance take [asset][period] returns on ONE
// calendar (marketdata.Aligned.Returns). Covariance is the per-period sample
// covariance; multiply by 252 for an annualized daily one.
func CorrelationMatrix(returns [][]float64) [][]float64
func Covariance(returns [][]float64) [][]float64

// PeriodReturn is one calendar period of a value series.
type PeriodReturn struct {
    Start, End time.Time // first and last date of the series inside the period
    Return     float64   // level(End) / level(previous period's End) - 1
    Partial    bool      // the first period, measured from the series' first point
}

// CalendarReturns cuts a value series into calendar periods of the given
// length in months (1 monthly, 3 quarterly, 12 yearly) and returns each
// period's return. It is the table behind an "annual returns" row and a
// monthly heatmap; the first period is flagged Partial rather than dropped.
func CalendarReturns(dates []time.Time, values []float64, months int) []PeriodReturn

// RollingBeta and RollingCorr are Beta and Corr over every trailing window
// of the given length in years, on dates common to both series, in the
// shape of RollingVol.
func RollingBeta(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64, years float64) ([]time.Time, []float64, bool)
func RollingCorr(dates []time.Time, values []float64, benchDates []time.Time, benchValues []float64, years float64) ([]time.Time, []float64, bool)

// VaR is the historical value at risk of a return sample at confidence p
// (0.95: the loss exceeded in 5 % of periods), as a POSITIVE fraction;
// CVaR is the mean loss beyond it. Both read the sample as it is: daily
// returns give a daily figure.
func VaR(returns []float64, p float64) (float64, bool)
func CVaR(returns []float64, p float64) (float64, bool)
```

`metrics.Frequency` is deliberately NOT introduced: `months int` keeps the
package free of `marketdata` and free of a second enum; `marketdata.Frequency`
converts with a plain cast.

The golden package's private `calYearDaily` is replaced by `CalendarReturns`
so the two cannot disagree. Tests: matrices on a hand-built 3-asset sample
(symmetry, unit diagonal, agreement with `Corr`), calendar table on a series
with a partial first month and a 31-December quote, VaR against
`Quantiles`. Examples for each new function.

### M3. `portfolio.NewSpec` and `pkg/analyze`: the numbers layer

```go
// portfolio
// Line is one allocation, in the in-memory convention: a weight as a
// FRACTION (0.6 for 60 %), a TER in percent per year as Holding.Fees, negative
// when unknown. NewSpec validates and normalizes exactly as Parse does
// (weights renormalized to sum to 1 unless the spec later opts into
// Leverage), so a spec built here and the same lines in a file are the same
// spec.
type Line struct {
    ID     string
    Weight float64
    Fees   float64
}
func NewSpec(name string, lines ...Line) (*Spec, error)
```

`pkg/analyze` is new: the high-level, numbers-only API. It depends on
`marketdata`, `portfolio`, `metrics`, `suggest`, `datasets`, and on nothing
that renders.

```go
// Source is what analyze needs of a data client. *marketdata.Client
// satisfies it; tests supply a fake, and so can a consumer with its own
// store.
type Source interface {
    FetchExtended(ctx context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error)
    Fees(ctx context.Context, id string) (float64, bool)
}

type Options struct {
    Currency  string    // evaluation currency; "" = native
    Sim       bool      // splice bundled backcasts (the SIM convention)
    Benchmark string    // identifier for the relative statistics; "" = none
    From, To  time.Time // study window; zero = all available
    Rebalance int       // rebalancing period in days for a portfolio; 0 = 90
    Simdata   fs.FS     // nil = the embedded bundle
    ExactOnly bool      // marketdata.FetchOptions.ExactOnly
}

// AssetStudy is one asset dissected on the study window.
type AssetStudy struct {
    ID        string
    Meta      datasets.Asset // catalog record, zero when HasMeta is false
    HasMeta   bool
    Series    *marketdata.Series   // as evaluated: converted, spliced when asked, trimmed
    Stats     metrics.Stats
    Years     []metrics.PeriodReturn // calendar years
    Months    []metrics.PeriodReturn // calendar months
    Drawdowns []metrics.Episode      // chronological
    Relative  *metrics.Relative      // vs Options.Benchmark; nil without one
    Warnings  []string               // distributing share class, FX extrapolated before a date, simulated before a date
}

func Asset(ctx context.Context, src Source, id string, opt Options) (*AssetStudy, error)

// Composition is the look-through of a portfolio, as suggest computes it.
type Composition struct {
    AssetClass map[string]float64
    Geography  map[string]float64
    Currency   map[string]float64
    Sectors    map[string]float64 // of the equity share; Equity is that share
    Equity     float64
    Duration   suggest.DurationLedger
    Coverage   map[suggest.Category]float64 // RegimeFramework
    Unclassified float64
}

// PortfolioStudy is a portfolio dissected: the simulation, its statistics,
// each holding's own study on the same window, and what ties them.
type PortfolioStudy struct {
    Spec        *portfolio.Spec
    Portfolio   *portfolio.Portfolio
    Sim         *portfolio.SimResult
    Stats       metrics.Stats        // on Sim.Index
    Years       []metrics.PeriodReturn
    Months      []metrics.PeriodReturn
    Drawdowns   []metrics.Episode
    Relative    *metrics.Relative
    Holdings    []AssetStudy         // in Spec order, each on the simulation window
    Aligned     *marketdata.Aligned  // the holdings on the simulation calendar
    Correlation [][]float64          // of the holdings' daily returns, Spec order
    Attribution metrics.Attribution  // risk and return shares, Spec order
    Composition Composition
    Warnings    []string
}

func Portfolio(ctx context.Context, src Source, spec *portfolio.Spec, opt Options) (*PortfolioStudy, error)
```

Semantics to pin down in the godoc:
- The study window is `[From, To]` clipped to what the data covers; a
  portfolio's window is `Simulate`'s (first date where every holding
  quotes), and each `Holdings[i]` is studied on THAT window so their
  statistics are comparable, while `Asset` alone studies the longest window.
- `Rebalance` 0 means 90 days, the CLI's default, stated in one place
  (`analyze.DefaultRebalance`).
- Real (inflation-adjusted) statistics are NOT here in M3; they stay in
  `compare`, which owns the CPI choice per currency. M4 decides whether they
  move down.
- Errors: an unknown identifier surfaces `marketdata.ErrUnknownIdentifier`
  unwrapped; every other data problem is a `Warnings` line, never a silent
  number.

Tests: a fake `Source` serving three synthetic series (one with a late
start, one with a dividend list and a distributing name) checks the window
logic, the warnings, the correlation matrix against `metrics.Corr`, the
attribution shares summing to one, the composition against `suggest` called
directly. Examples: `ExampleAsset`, `ExamplePortfolio`, and one
`Example_sixtyForty` that is the README's opening snippet.

### M4. `compare` consumes `analyze`

`compare.Compute` builds each column through `analyze.Portfolio` and keeps
its own responsibilities: the common window across columns, the benchmark
column, the real (deflated) statistics, the HTML page model. `Comparison`
gains

```go
func (c *Comparison) Studies() []*analyze.PortfolioStudy
```

so a consumer of the pipeline reaches the numbers behind every chart.

Gate: the rendered report of `examples/dragon-decumulation-household.txt`
(and of two more example files, one with flows, one with leverage) must be
byte-identical before and after, with a warm cache; the diff is part of the
commit message. `make golden` unchanged. If the refactor cannot be made
identical (a rounding order changes a printed digit), the milestone stops
and reports the digit rather than moving a golden.

### M5. Plumbing hygiene

- Unexport `simgen.AnchorTrend`, `AnchorStart`, `SeriesFromFrame`,
  `CapWeights` (no caller outside the package).
- Keep exported, because `cmd/` needs them, but move under a "Generator
  plumbing" heading in each `doc.go` with one sentence saying they are not
  for consumers: `simgen.DBiProjection`, `DBiReplication`, `Rebase`,
  `Splice`; `marketdata.WarmupIDs`, `WriteSimdata`, `ProxySymbol`,
  `ExtendBack`, `MergeDividends`.
- `suggest.Correlation` delegates to `metrics.Corr` and its godoc says so.

No `internal/` move: the generators are part of the module and the churn
buys nothing a doc heading does not.

### M6. Documentation

- `README.md`, "Using it as a library", rewritten by TASK, each snippet
  copied from a runnable example and saying which (`// see
  analyze.Example_sixtyForty`): get a price history and its raw data; study
  one asset; build and simulate a portfolio in code; the correlation matrix
  and the risk budget; optimize weights; FIRE in ten lines
  (`scenario` + `decumul.Plan.Simulate` + `Outcome`); reconstruct a missing
  history (`simgen` recipe + `Validate`); render. Plus the units table
  (percent vs fraction, per package) and the "where does this number come
  from" pointers (golden tests, conventions).
- Root `doc.go`: the layering diagram (which package may import which), the
  same units table, and an "API versus plumbing" paragraph.
- Examples: `scenario` gets three (`ParametricSource`, `BlockBootstrap` on a
  bundled panel, `Deflate`); `decumul` gets three more (`Plan.Simulate` to
  `Outcome`, `Solve` on `WithdrawalAxis`, `Lifetime` with `LifeOutcome`);
  every M1 to M3 function has one.
- AGENTS.md: `analyze` row in the Map, the core pipeline shows the
  three-line `analyze` path first and the manual path second; the units
  paragraph points at the table.

## 4. Deliberately left out

- **Typed units** (a `Percent` type). It would touch every signature in
  `portfolio`; the gain is a compile error where a comment stands today.
  Revisit only with a v1 that is allowed to break the siblings.
- **A `metrics` API over `marketdata.Series`.** Principle 2.
- **Multi-factor regression, efficient frontier, rolling attribution.**
  Named here so the next reader knows they were weighed; each is a
  self-contained addition on top of M1 to M3 and none is needed to make the
  library consistent.
- **Merging the position types** (`portfolio.Holding`, `suggest.Holding`,
  `datasets.Asset`). They carry different information at different stages
  (written line, built asset, catalog record); `analyze.AssetStudy` is the
  one a consumer meets, and it holds the others.

## 5. Acceptance

- `make check` and `make golden` green at every milestone; M4 also passes
  the byte-identical report diff.
- No exported signature of v0.2.80 changed or removed, except the four
  `simgen` functions M5 names (no caller exists).
- `go doc` of every new symbol reads the unit it takes and returns.
- Every new exported function has a runnable example; every README snippet
  names the example it is copied from.
- `../finador` and `../locador` build and pass their gates on the resulting
  tag with only a `go.mod` bump.

## 6. Implementation plan

One commit per milestone, in order, each with its doc changes in the same
commit (AGENTS.md rule 9). Suggested commit subjects:

1. `marketdata: NewSeries, Dates/Values/Returns/Rebase/Resample, AlignSeries`
2. `metrics: correlation and covariance matrices, calendar returns, rolling beta and correlation, VaR/CVaR`
3. `portfolio: NewSpec; analyze: the numbers-only asset and portfolio studies`
4. `compare: columns are analyze studies (byte-identical reports)`
5. `simgen, marketdata: plumbing unexported or filed under its own heading`
6. `docs: the library chapter, by task, every snippet a runnable example`

A tag follows M6, and the two siblings bump to it.
