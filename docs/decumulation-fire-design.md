# Decumulation / FIRE analysis: design

Status: design, approved for planning (2026-06-28).

## 1. Goal

Extend pofo with the ability to evaluate, analyse and tune **decumulation**
(withdrawal / retirement / FIRE) portfolios: size a starting capital, a cash
buffer and a withdrawal rate against a target **probability of ruin**, under a
chosen return model, taking into account temporary spending cuts, a cash or
inflation-linked sleeve, pension/income cashflows, a finite life horizon and
the flat tax on capital gains.

Two questions the tool must answer directly:

- *To minimise ruin at a 4 k€/month real withdrawal on capital C at CAGR µ and
  vol σ, is it better to hold 5 years of cash or 3 years (more invested)?*
- *In "60% NTSGSIM, 25% DBMFESIM, 15% XAUUSD", would 20% DBMFE + 20% gold lower
  the ruin risk at a constant withdrawal rate?*

Everything is in **real euros** (today's purchasing power); the spending floor
is held constant in real terms, the return model is real, pensions are entered
in real terms.

The work is organised so that **the most generic capabilities live in
reusable, well-layered library packages**, a thinner FIRE-specific layer sits
on top, and a minimal embedded UI consumes them for live what-if exploration.

## 2. Constraints and fit with the existing library

- **Zero dependencies, stdlib only.** pofo ships no third-party modules. The
  reference specs assume `gonum` for the Student-t draw; we do **not** use it.
  Student-t is implemented by hand on `math/rand/v2`.
- **Reuse, don't duplicate.** `pkg/metrics` already computes drawdowns, TTR,
  Ulcer, rolling CAGR, quantiles and histograms; the FIRE outcome metrics
  compose those per path rather than re-deriving them. `pkg/chart` renders SVG;
  the UI returns chart fragments from it. `pkg/marketdata` fetches each
  holding's (SIM-extended) series and the `^HICP-FR` deflator.
- **Layered packages,** mirroring the existing `pkg/` style (short names,
  per-package `doc.go`, runnable examples, golden tests).
- **Real euros throughout.** A `Deflate` helper turns nominal closes into real
  returns via HICP; parametric inputs (µ, σ) are already real.

## 3. Architecture

Three layers plus small, generic additions to existing packages.

```
pkg/scenario/     LAYER 1: generic return-path generation (no FIRE knowledge)
pkg/decumul/      LAYER 2: withdrawal engine, FIRE metrics, sweeps, tax, cashflows
pkg/decumul/web/  LAYER 3: reusable http.Handler + go:embed html/js/css
cmd/pofo          -fire [portfolio.txt] wiring + portfolio→scenario adapter
pkg/chart         + Heatmap (2D ruin surface) + Bars (recovery-time histogram)
```

**Key seam.** The withdrawal engine consumes a `scenario.Source`, so the
parametric, block-bootstrap and historical-cohort projections are the *same*
withdrawal kernel with three interchangeable return providers. Switching the
return model is swapping one interface value.

**Data flow, portfolio mode.** `marketdata` fetches each holding (SIM-extended)
→ adapter deflates by `^HICP-FR`, aligns the holdings into a `scenario.Panel`
and fits real `µ/σ/df` → `decumul.Plan` wraps it with the withdrawal params →
the server runs `Simulate`/`Sweep` on each slider change → returns SVG/JSON →
the browser swaps it in. **Parametric mode** skips the portfolio and feeds a
`ParametricSource` straight from the sliders.

## 4. Layer 1: `pkg/scenario`

Generic synthetic return-path generation, reusable beyond FIRE (accumulation,
glidepaths, any path-dependent study).

```go
// Sequence is a periodic real-return path (e.g. 40 annual returns).
type Sequence []float64

// Source produces synthetic return paths of a fixed length.
type Source interface {
    Draw(rng *rand.Rand) Sequence
    Len() int
}
```

- **`ParametricSource{Mu, Sigma, Df, Periods}`**: i.i.d. draws
  `r = Mu + scale*T`, with `T` a standard Student-t at `Df` degrees of freedom
  standardised so that `stdev(r) == Sigma` exactly
  (`scale = Sigma / sqrt(Df/(Df-2))`, `Df>2`); `Df<=0` falls back to Normal.
  Student-t without gonum: `T = Z / sqrt(Chi2(Df)/Df)`, `Z~N(0,1)`,
  `Chi2(Df)` a sum of `Df` squared normals (general `Df` via a Gamma draw).
  Clamp `1+r >= 0` so an extreme draw cannot make capital negative.
- **`Panel`**: an aligned matrix of per-asset real returns `returns[asset][t]`
  plus a `Weights` vector; `Combine(weights) Sequence` produces the portfolio
  return path. Re-weighting is cheap, so live allocation changes never refetch.
- **`BlockBootstrap{Panel, BlockLen}`** and
  **`StationaryBootstrap{Panel, MeanBlock}`**: resample contiguous blocks *on
  the time axis* (preserving cross-asset correlations and historical regimes:
  the 70s inflation, the GFC, 2022), apply the current weights, and concatenate
  to a path of `Len()` periods.
- **`HistoricalCohorts{Panel}`**: yields each actual historical start window
  with no resampling: the deterministic "every retirement start date since the
  1970s" backtest. `Len()` paths is the number of available windows.
- **`Deflate(prices []marketdata.Point, hicp []marketdata.Point) Sequence`**:
  helper turning a nominal price series into real periodic returns using the
  HICP series; the bridge to "everything in real euros".

**Testing.** Feed a known panel and assert resampled mean/variance and
cross-asset correlation are preserved within tolerance; assert the standardised
Student-t has unit-scaled variance `Sigma^2`; round-trip `Deflate` on a
constant-inflation synthetic series.

## 5. Layer 2: `pkg/decumul`

The FIRE-specific layer: withdrawal kernel, outcome metrics, sweeps, tax,
cashflows.

```go
type Plan struct {
    Capital    float64         // starting real capital
    NeedAnnual float64         // net real floor spending
    Cashflows  []Cashflow      // pension/income as dated real flows
    Years      int             // life horizon (ruin = depleted before)
    Buffer     BufferSleeve    // years-of-spending sleeve, its real return, bucket thresholds
    Flex       FlexRule        // cut spending by Cut while drawdown > Threshold
    Tax        Tax             // pluggable; default CTOFlatTax{Rate}
    Source     scenario.Source // the return model
}

type Cashflow struct { FromYear int; Annual float64 } // e.g. pension from year 12
type BufferSleeve struct {
    Years         float64 // sleeve sized as Years * annual spending (capped at capital)
    RealReturn    float64 // sleeve's real return (cash ~0.5%, linkers a bit higher)
    DrawThreshold float64 // drawdown beyond which the buffer is drained first (default 0.10)
    RefillCap     float64 // max share of growth used to refill per year (default 0.5)
}
type FlexRule struct { Threshold, Cut float64 } // e.g. 0.20 drawdown -> cut 0.25

func Run(p Plan, rng *rand.Rand) PathResult        // one path
func Simulate(p Plan, nPaths, workers int) Ensemble // parallel MC, deterministic per worker
func (e Ensemble) Outcome() Outcome
func CapitalForRuin(p Plan, target float64) float64 // bisection on shared pre-drawn paths
func Sweep1D(p Plan, v Param, values []float64) []SweepPoint
func Sweep2D(p Plan, x, y Param, xs, ys []float64) Surface
```

**Per-path kernel (`Run`)**, generalising the v2 JS spec, all in real euros:

1. Split capital into `buffer = Years*spending` (capped) and a growth sleeve;
   initialise the tax cost basis to the growth sleeve.
2. For each year: `need = NeedAnnual - active cashflows`, floored at 0 (net).
3. `total = growth + buffer`; if `total <= 0` → ruin (latched), stop.
4. `dd = 1 - total/peak`; if `Flex` active and `dd > Threshold`, cut `need`.
5. Bucket rule: if `dd > DrawThreshold` and buffer > 0, take the net from the
   buffer first (no tax), the remainder from growth (gross-up tax below);
   otherwise take the net from growth (gross-up), then refill the buffer toward
   its target (`min(target-buffer, spending, growth*RefillCap)`).
6. Growth gross-up: tax only the realised **gain fraction**
   `gainFrac = max(0, 1 - cost/growth)`, `gross = need/(1 - rate*gainFrac)`,
   reduce `cost` pro rata. Effective rate starts low and drifts toward the flat
   rate as gains compound (intended).
7. Apply returns: `growth *= 1 + r` (from `Source`), `buffer *= 1 + RealReturn`;
   `cost` does not move with returns, only with sales.

`Run` returns the full yearly wealth path plus realised taxes and withdrawals,
so per-path `metrics` (drawdowns, TTR, rolling CAGR, Ulcer) compose directly.

**`Outcome`** bundles, reusing `pkg/metrics` where possible:

- ruin probability,
- terminal real wealth p5 / p50 (median over all paths, 0 for ruined),
- median years underwater (below the prior real high),
- worst 10-year real CAGR,
- withdrawal-failure rate, median realised withdrawal rate before depletion,
- Ulcer, **CDaR** (Conditional Drawdown at Risk) on the decumulation path,
- **`RecoveryTimeDistribution`**: the full histogram of years-to-regain a
  prior real high across all paths (not just the mean): the headline metric
  ("it's been 14 years below my initial wealth"), since the mean hides the
  tail.

**Tax** is an interface; `CTOFlatTax{Rate}` is the default French CTO flat-tax
implementation (pro-rata cost basis, gain-fraction only). **Pension/income** is
a `Cashflow` list (modelled as a future cashflow, not an asset). Both are
generic and swappable.

**`CapitalForRuin`** bisects `C0` in `[lo, hi]` over ~16-20 iterations,
**reusing the same pre-drawn paths** across evaluations so Monte-Carlo noise
does not break monotonicity.

**`Sweep1D`** varies one parameter to a `(value, ruin)` / `(value, terminal)`
curve; **`Sweep2D`** varies two (e.g. buffer years × CAGR) to a ruin
`Surface` for the heatmap.

## 6. Layer 3: UI (`pkg/decumul/web` + `-fire`)

A reusable `http.Handler` constructor, with `go:embed` for one html, one js and
one css file; `cmd/pofo` only wires it.

- `pofo -fire` → **parametric playground.** Sliders for every parameter
  (capital, net spending floor, buffer years, real growth return, vol, df,
  buffer real return, horizon, pension start year, pension annual, possible
  spending cut, flat-tax rate, n-sims), with the defaults and ranges from
  `valeurs-default.txt`. Live charts: **buffer arbitrage** (ruin % on the left
  axis and median terminal wealth on the right vs buffer years), **ruin vs
  capital**, **2D ruin surface** (buffer × CAGR heatmap), **recovery-time
  histogram**; summary cards (initial withdrawal rate, ruin at the selected
  buffer, ruin at 0 buffer, median terminal, median effective tax).
- `pofo -fire portfolio.txt` → **portfolio mode.** Same UI, but `µ/σ/df` and
  the historical `Panel` are derived from the portfolio. A **return-model
  toggle** chooses parametric / block-bootstrap / historical cohorts. Per
  holding **allocation sliders** re-weight the `Panel` and re-project live, so
  "60/25/15 vs 20/20/…" is a drag-and-compare.
- **Server protocol.** `/api/sim` takes the params as JSON and returns chart
  **SVG fragments** (rendered by `pkg/chart`) plus scalar cards as JSON. The JS
  is one small file: read sliders → debounced POST → swap `innerHTML`. The
  engine stays entirely in Go.

**Generic chart additions.** `chart.Heatmap` (the 2D ruin surface) and
`chart.Bars` (the recovery-time histogram) are added as small, dependency-free
SVG primitives reusable elsewhere, matching the existing `Line`/`Pie` style.

## 7. Validation and testing

- **Golden acceptance tests** translate the reference spec's validation table
  into `decumul` tests: ruin probabilities within ±0.3 pt and target capitals
  within ±0.03 M€ at ≥150 k paths, with workers fixed for reproducibility. This
  proves the kernel matches the Python reference and fits pofo's golden-test
  culture. Representative anchors to reproduce:
  - target capital @ 5% ruin, µ 3.5%, pension 1800, need 4 k€/m → ≈ 1.67 M€;
  - ruin @ C0 = 2.0 M€, µ 3.5%, need 4 k / pension 1800 → ≈ 2.1%;
  - horizon 95 / 90 / 85 → target ≈ 1.84 / 1.67 / 1.51 M€.
- **`scenario` tests:** resampled moment/correlation preservation; standardised
  Student-t variance; `Deflate` round-trip.
- **`chart` tests:** Heatmap/Bars render valid SVG, in the existing style.
- Each layer is independently testable; the web layer is smoke-tested via
  `httptest` on the handler.

## 8. Caveats to surface in the generated UI and docs

- The parametric model is **i.i.d. with fat tails**: it does not capture
  volatility clustering or long bear decades and is **probably optimistic** vs
  multi-country data (Anarkulova). The historical-cohort and block-bootstrap
  models exist precisely to temper this; treat 3.0–3.5% real as the planning
  case. Read ruin in relative orders of magnitude.
- Everything is real euros; the pension is an **input**, not computed here.
- **Not investment advice**; a hypothesis-exploration tool.

## 9. Phased implementation plan

1. `pkg/scenario`: `Source`, `ParametricSource`, `Panel`, bootstraps, cohorts,
   `Deflate`, with tests and `doc.go`.
2. `pkg/decumul`: `Plan`, `Run` kernel, `Simulate`, `CapitalForRuin`; golden
   acceptance tests vs the reference table.
3. `pkg/decumul` outcome metrics (recovery-time distribution, CDaR, …) and
   `Sweep1D`/`Sweep2D`, reusing `pkg/metrics`.
4. `pkg/chart`: `Heatmap` and `Bars`.
5. `pkg/decumul/web`: handler, embedded assets, parametric playground;
   `cmd/pofo -fire` wiring.
6. Portfolio adapter (holdings → `Panel`, fit `µ/σ/df`), the return-model
   toggle and live allocation sliders.

README and per-package `doc.go` updates land with each step.
