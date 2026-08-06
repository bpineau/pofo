# Example portfolios

Ready-to-run model portfolios — famous strategies and well-regarded
investors' builds, modernized with the bundled (mostly UCITS) catalog. Each
file's header gives the name, the idea and a link to the source. Run any of
them:

```sh
./pofo examples/all-weather-dalio.txt          # HTML report
./pofo -cli examples/dragon-portfolio-artemis.txt
./pofo -coverage examples/cockroach-portfolio-mutiny.txt   # regime coverage, offline
./pofo examples/*.txt                           # compare them all
```

**Conventions.** Classic American models (All Weather, Permanent, Dragon…)
use their US building blocks with the `SIM` suffix so the backtest reaches
back decades; the UCITS you would actually buy is named in each line's
comment. Modern, European and PEA models use real UCITS quotes. `SIM`
histories before a fund's inception are simulated (see `pkg/datasets/simdata/`).

## Using the optimizers

Any portfolio can hand its weights to an optimizer with a single meta line.
The report then shows two versions side by side — `name (as written)` and
`name (objective)` — so you compare the optimizer's allocation against your
baseline before adopting anything.

```
#meta optimize:risk-parity                 # equalize each asset's risk
#meta optimize:min-volatility              # lowest-variance mix
#meta optimize:max-sharpe,max-weight:35    # best in-sample Sharpe, capped at 35%
```

Pick the objective by what you trust:

- **risk-parity** — every holding contributes the same share of total risk.
  Uses only the covariance (not past returns), so it does not chase whatever
  happened to win the backtest. The most robust choice and the natural one
  for all-weather / diversified sleeves. `max-weight` is ignored here (the
  weights follow from the equal-risk condition).
- **min-volatility** — the calmest mix. Tends to pile into bonds / low-vol
  assets; useful to anchor a withdrawal phase, but it ignores return
  entirely, so cap it or it concentrates.
- **max-sharpe** — the best risk-adjusted return *over the fitted window*.
  This one overfits: it leans hard on the past winner. Always cap it
  (`,max-weight:30`–`40`) and read the result as a hint, not a target.

The weights are fitted **in-sample**, over the period where every asset has a
quote, and the note under the optimized portfolio reports the in-sample
expected return / volatility / Sharpe. Past-fitted figures are a starting
point, not a promise — check that the common window (printed in the report)
is long, and that the allocation makes economic sense, before moving real
weights. `optimize` cannot be combined with `#meta leverage`.

Workflow tip: keep your file's hand-written weights as the baseline and add
an `optimize` line only while exploring. Once you have decided, write the
tuned weights in directly, so the file documents your actual allocation
rather than recomputing it (and drifting) on every run.

## Simple / lazy (accumulation-friendly)

- `lazy-all-world` — 100% All-World, one fund (Bogleheads "VT and chill").
- `lazy-world-plus-em` — World + EM, cap-weighted ACWI built cheaply.
- `lazy-bogleheads-3fund` — world + EM + global bonds, the canonical lazy build.
- `lazy-aggressive-80-20` — 80% equity / 20% bonds growth allocation.
- `lazy-value-tilt-twofund` — world core + a small-cap-value tilt.
- `coffeehouse-schultheis` — Bill Schultheis' value/small/REIT + bonds.

## Permanent / risk-parity / all-weather

- `permanent-portfolio-browne` — Harry Browne's four-environment 25/25/25/25.
- `all-weather-dalio` — Ray Dalio / Bridgewater All Weather.
- `golden-butterfly` — Portfolio Charts' Permanent + small-value wing.
- `pinwheel-portfolio` — Portfolio Charts' Pinwheel, broad and mildly tilted.
- `larry-portfolio-swedroe` — small-value + safe bonds (efficient equity risk).
- `desert-portfolio-bridges` — low-volatility 30/60/10.

## Tail-risk / trend (anti-fragile)

- `dragon-portfolio-artemis` — Chris Cole's Dragon (equity/bonds/gold/trend/long-vol).
- `cockroach-portfolio-mutiny` — Mutiny Fund's four-quadrant Cockroach.
- `risk-parity-plus-trend` — diversified set weighted by the **risk-parity optimizer**
  (`#meta optimize:risk-parity` — run it to see the computed weights).

## Capital-efficient / return-stacking (modern)

- `efficient-core-9060` — WisdomTree NTSX/NTSG 90/60 efficient core.
- `return-stacked-modern` — stacked stocks + bonds + managed futures (RSSB/RSST).
- `ntsx-all-weather` — efficient core + gold/commodities/trend diversifiers.

## Factor-tilted (academic)

- `global-multifactor` — one diversified-factor world fund (JPGL).
- `merriman-style-tilt` — Paul Merriman's value/small worldwide tilt.
- `four-factor-blend` — equal-weight value/momentum/quality/min-vol.

## Income / decumulation

- `global-dividend-income` — diversified income core for a gentle drawdown.
- `yield-shield-decumulation` — income-tilted, drawn at 4%/yr (`#meta withdraw`).

## PEA-eligible (French equity wrapper)

- `pea-all-world-100` — single World PEA ETF.
- `pea-aggressive-growth` — World + S&P 500 + Nasdaq, PEA.
- `pea-world-emerging` — World + emerging markets, PEA.
- `pea-core-satellite` — World core + US tech + Europe satellites, PEA.

## Curated (built here)

- `modern-all-weather-ucits` — an all-weather you can buy today, UCITS-first.
- `aggressive-accumulation-ucits` — equity-heavy factor build for the growth phase.
- `gentle-decumulation-ucits` — balanced income build, 4%/yr withdrawal.

## Phase mechanics demos

- `dca-accumulation-demo` — starting capital + monthly contributions.
- `balanced-decumulation-demo` — starting capital + 4%/yr withdrawal.

Older simple examples also live here: `classique-60-40`, `permanent`,
`sp500`, `world-equity`, `optimized`.
