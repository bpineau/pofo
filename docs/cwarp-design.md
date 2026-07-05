# CWARP support — design

CWARP (Cole Wins Above Replacement Portfolio, Artemis Capital Management,
"Moneyball for Modern Portfolio Theory", 2020) scores whether an asset or
portfolio *improves* a pre-existing replacement portfolio when layered on top
at a fixed weight, financed by borrowing. Unlike the Sharpe ratio it rewards
non-correlation and skew, because the combined portfolio's downside volatility
and drawdown depend on how the overlay interacts with the replacement.

## Formula

    CWARP = ( sqrt( (Sortino_new / Sortino_repl) x (RtMDD_new / RtMDD_repl) ) - 1 ) x 100

- New-portfolio return series (per period): `r_new = r_repl + w*(r_asset - financing)`.
- `Sortino = mean(r - rf) / downside_dev * sqrt(252)`, with
  `downside_dev = sqrt(mean( min(r - rf, 0)^2 ))` (MAR = rf).
- `RtMDD = (CAGR - rf) / |maxDrawdown|` (return to maximum drawdown; CAGR is the
  compound annualized growth of the value path).

Sign: CWARP > 0 means the overlay improves the replacement's risk-adjusted
returns (higher Sortino and/or return-to-drawdown); CWARP < 0 means it hurts.

Verified against the paper's worked example (p.5): Sortino 1.50 -> 1.75,
RtMDD 0.15 -> 0.25 gives CWARP = (sqrt(1.75/1.50 * 0.25/0.15) - 1) * 100 = +39.44.

## Conventions (match the reference Python and pofo)

- periodicity = 252 (daily simple returns, like the rest of `pkg/metrics`).
- risk-free rf = 0 (pofo's zero-risk-free annualization convention).
- financing = 0.
- overlay weight w = 0.25 (the paper's standard), replacement weight = 1.
- Replacement portfolio = the report benchmark (`-benchmark`, default `^GSPC` =
  S&P 500 equity beta, the paper's standard). Pointing `-benchmark` at a 60/40
  series yields the paper's 60/40 variant for free.

rf, financing and w are parameters of the `pkg/metrics` API (so the library
stays general) but are fixed at their defaults in the CLI for this version.

## Component 1 — `pkg/metrics.CWARP` (core)

```go
// CWARPParams configures the overlay; a zero value uses the paper's standard
// (Weight 0.25, RiskFree 0, Financing 0).
type CWARPParams struct { Weight, RiskFree, Financing float64 }

// CWARP scores overlaying `asset` at weight w on `replacement`. ok is false
// when the inputs are too short or a replacement denominator is non-positive
// (the ratio is then meaningless).
func CWARP(asset, replacement []float64, p CWARPParams) (score float64, ok bool)
```

Internals: build `r_new`, compute Sortino and return-to-max-drawdown for the
replacement and the new series, apply the geometric formula. Internal helpers
`sortinoRatio(returns, rf)` and `returnToMaxDrawdown(returns, rf)` implement the
exact definitions above (reusing `MaxDrawdown` and a compounded-CAGR helper
where they match). `ok=false` when `len<...`, or `Sortino_repl<=0`, or
`RtMDD_repl<=0`, or a `_new` denominator is non-positive.

Tests: the 39.44 example (formula level); an anti-correlated diversifier ->
CWARP>0; a correlated levered-beta overlay -> CWARP<0; the `ok=false` paths. A
golden anchor records the 39.44 example.

## Component 2 — portfolio-level CWARP column

Each report row is a portfolio; its CWARP answers "does layering 25% of this
portfolio on top of the benchmark improve it?". Computed in the CLI where the
benchmark returns are available (the same place Beta is set, `cmd/pofo/main.go`
around `VsBenchmark`), not inside `metrics.Compute` (which has no benchmark).

- `metrics.Stats` gains `CWARP float64` and `HasCWARP bool`.
- `buildStatRows` gains a "CWARP" `report.StatRow` with a tooltip explaining the
  metric, the 25% overlay, the replacement = benchmark, and the sign. The cell
  is blank when `!HasCWARP` (no benchmark, or a non-positive replacement
  denominator). Value formatted signed, e.g. `+39.4` / `-4.2`.

## Component 3 — per-asset CWARP

Each holding scored as a 25% overlay on the benchmark: `CWARP(assetReturns_i,
benchReturns)`. This is the paper's native use (screening diversifiers): which
sleeves actually diversify equity beta.

- `report.AssetRow` gains `CWARP float64` + `HasCWARP bool`.
- The per-portfolio holdings table (already built around `cmd/pofo/main.go`
  ~1349) fills CWARP from each holding's window-aligned returns and the
  benchmark, rendered as a new column, blank when unavailable.
- A legend note explains the column.

## Component 4 — `#meta optimize:cwarp`

The objective is path-dependent (max drawdown) and needs the replacement
series, so it cannot reuse the mean/covariance solver.

```go
// SolveCWARP maximizes the portfolio's CWARP against replacement over the
// capped simplex.
func SolveCWARP(returns [][]float64, replacement []float64, spec Spec) (Result, error)
```

- `f(w) = -CWARP( sum_i w_i * returns_i , replacement )`, minimized over the
  capped simplex by multi-start projected descent with a numerical
  (central-difference) gradient, reusing `minimizeSimplex` /
  `projectCappedSimplex`, keeping the best CWARP across the same deterministic
  starts as `maxSharpe` (equal, each single asset, inverse-variance). Documented
  as a heuristic optimum, since CWARP is non-convex and non-smooth.
- `optimize.Objective` gains `CWARP = "cwarp"`; `ParseSpec` accepts it (with the
  existing `max-weight` constraint). `Result` gains `CWARP float64`.
- The CLI passes the already-fetched benchmark returns into `SolveCWARP`; the
  optimizer note reports the achieved CWARP. Missing benchmark -> a clear error.

## Testing / verification

- Unit tests per component; `make check` and `make golden` stay green.
- `pkg/metrics` golden for the 39.44 example.
- `optimize:cwarp` test: a toy 2-asset case where the anti-correlated diversifier
  earns meaningful weight and the achieved CWARP is positive.
- End-to-end: an example portfolio rendered with the CWARP column, the per-asset
  column, and a `#meta optimize:cwarp` run.

## Out of scope (this version)

- Configurable overlay weight / financing / rf from the CLI (the `pkg/metrics`
  API exposes them; the CLI fixes the defaults).
- A monthly-periodicity option.
