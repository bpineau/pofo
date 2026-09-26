# Example portfolios

Ready-to-run model portfolios: famous strategies and well-regarded
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
use their US building blocks with backcast-extended history so the test reaches
back decades; the UCITS you would actually buy is named in each line's
comment. Modern, European and PEA models use real UCITS quotes. Histories
before a fund's inception are simulated (see `pkg/datasets/simdata/`). Files
that want that extension now opt in for the whole file with one `#meta sim:on`
line, so holdings read as what you would actually buy (`IWDA`, `NTSG`,
`GOLD`…); a holding with no simulated history simply falls back to its real
quotes. The per-line `SIM` suffix (`IWDASIM`) still works for extending a
single holding ad hoc.

## Using the optimizers

Any portfolio can hand its weights to an optimizer with a single meta line.
The report then shows two versions side by side: `name (as written)` and
`name (objective)`, so you compare the optimizer's allocation against your
baseline before adopting anything.

```
#meta optimize:risk-parity                 # equalize each asset's risk
#meta optimize:min-volatility              # lowest-variance mix
#meta optimize:max-sharpe,max-weight:35    # best in-sample Sharpe, capped at 35%
#meta optimize:max-return,max-vol:9.5      # the most return inside a volatility budget
```

Pick the objective by what you trust:

- **risk-parity**: every holding contributes the same share of total risk.
  Uses only the covariance (not past returns), so it does not chase whatever
  happened to win the backtest. The most robust choice and the natural one
  for all-weather / diversified sleeves. `max-weight` is ignored here (the
  weights follow from the equal-risk condition), and the `max-vol` /
  `min-return` / `max-drawdown` limits are refused outright, since this
  solver cannot enforce them.
- **min-volatility**: the calmest mix. Tends to pile into bonds / low-vol
  assets; useful to anchor a withdrawal phase, but it ignores return
  entirely, so cap it or it concentrates.
- **max-sharpe**: the best risk-adjusted return *over the fitted window*.
  This one overfits: it leans hard on the past winner. Always cap it
  (`,max-weight:30`-`40`) and read the result as a hint, not a target.

The weights are fitted **in-sample**, over the period where every asset has a
quote, and the note under the optimized portfolio reports the in-sample
expected return / volatility / Sharpe. Past-fitted figures are a starting
point, not a promise; check that the common window (printed in the report)
is long, and that the allocation makes economic sense, before moving real
weights. `optimize` cannot be combined with `#meta leverage`.

Three constraints make that warning less necessary, and
`optimized-constrained.txt` is the worked example of all three:

- **bounds per line** (`min-weight:5`, `bounds:NTSG:10-35`) keep the search
  inside ranges you can defend for reasons the backtest cannot see. Without
  them an optimum is a corner solution, which is the least durable thing an
  optimizer produces.
- **limits** (`max-vol:9.5`, `min-return:10.5`, `max-drawdown:20`) ask the
  real question instead of a proxy for it. They also route around a trap:
  Sharpe here is computed at a **zero** risk-free rate, so a cash-like sleeve
  buys ratio for free.
- **`train:..2015`** fits on that window only, while the report measures the
  result over the whole history: the column you read is then out of sample,
  and the note prints what the weights promised in-sample next to what they
  actually did afterwards. The gap is usually the most useful number on the
  page.

To see what a single weight is worth without any optimizer at all, run
`pofo -sweep <file>`: one table per line, its weight moved across a grid, the
others keeping their proportions.

Workflow tip: keep your file's hand-written weights as the baseline and add
an `optimize` line only while exploring. Once you have decided, write the
tuned weights in directly, so the file documents your actual allocation
rather than recomputing it (and drifting) on every run.


## Classics, modernized

Famous builds with their US bricks and backcast-extended history, so the
test reaches back decades; the UCITS you would buy is in each line's comment.

- `tradi-60-40`: the 60/40 every other file is arguing with.
- `bogleheads-3fund`: world + EM + global bonds, the canonical lazy build.
- `value-tilt-twofund`: a world core with a mild small-cap-value tilt.
- `coffeehouse-schultheis`: Bill Schultheis' value/small/REIT + bonds.
- `permanent-portfolio-browne`: Harry Browne's four-environment 25/25/25/25.
- `all-weather-dalio`: Ray Dalio / Bridgewater All Weather.
- `golden-butterfly`: Portfolio Charts' Permanent + small-value wing.
- `pinwheel-portfolio`: Portfolio Charts' Pinwheel, broad and mildly tilted.
- `larry-portfolio-swedroe`: small-value + safe bonds (efficient equity risk).
- `desert-portfolio-bridges`: low-volatility 30/60/10.
- `merriman-style-tilt`: Paul Merriman's value/small worldwide tilt.
- `four-factor-blend`: equal-weight value / momentum / quality / min-vol.
- `sp500`, `msci-world`: the two single-index benchmarks, history-extended.

## Tail-risk, trend and return engines

- `dragon-portfolio-artemis`: Chris Cole's Dragon (equity, bonds, gold,
  trend, long vol).
- `cockroach-portfolio-mutiny`: Mutiny Fund's four-quadrant Cockroach.
- `stagflation-bunker`: only what works in persistent inflation (trend,
  commodities, gold, short linkers); the regime lab (Neville et al. 2021).
- `risk-parity-plus-trend`: a diversified set weighted by the risk-parity
  optimizer (`#meta optimize:risk-parity`; run it to see the weights).
- `hydra-five-engines-ucits`: one head per return engine (efficient core,
  small value, two trend models, gold, long duration), every line buyable
  from an EU retail account; over 1988-2026 it matches the dragon family on
  CAGR and beats it on volatility, recovery time and every stress window.
  The file doubles as its own design document: regime map, weight ranges,
  blind spots.
- `hydra-five-engines-capital-efficient`: the frontier build (~125 %
  notional through stacked funds, NTSG and GDE); the research file with the
  exposure ledger, the volatility/CAGR frontier and the alternatives ledger.
  Two lines are US-listed and not buyable by an EU retail investor.
- `cerberus-three-heads`: the hydra's sibling, the same engines regrouped
  into three even notional thirds (equity / trend / bonds + gold), euro
  duration, global small value; over 1996-2026 it matches the S&P 500's EUR
  return at 46 % of its volatility. One line (GDE) is not UCITS; the file
  carries the measured access ledger.

## Capital-efficient / return stacking

- `return-stacked-modern`: stacked stocks + bonds + trend + gold (RSSB, RSST).
- `ntsx-all-weather`: efficient core + gold, commodities, trend diversifiers.
- `modern-all-weather-ucits`: an all-weather you can buy today, UCITS-first.

## Decumulation

The curated family for a European early retiree drawing ~3 %/yr real: one
capital-efficient engine, a two-engine trend sleeve, gold, and a defensive
pocket matched to a real-euro liability. Compare the stages side by side:

```sh
./pofo examples/fire-bond-tent-departure.txt examples/fire-decumulation-core.txt
./pofo -fire examples/fire-core-longhist.txt    # ruin-probability explorer
```

- `fire-bond-tent-departure`: the year-0 build, defensive tent inflated to
  ~20 % for the sequence-risk window (Kitces/Pfau bond tent).
- `fire-decumulation-core`: the cruise allocation (equity + bonds / trend /
  gold / linkers + cash), drawn at 3 %/yr.
- `fire-core-longhist`: the same cruise mix with every leg reaching the late
  1980s; feed this one to `pofo -fire` so bootstrap and cohorts have depth.
- `dragon-decumulation-household`: the dragon design re-tuned for
  decumulation and deployed across French tax wrappers (the equity sleeve
  carved into a PEA/PEA-PME section, the lost bond overlay rebuilt with a
  euro-native long-duration line); covers the wrapper logic (spending order
  as an implicit glidepath, rebalancing asymmetry).
- `risk-budget-decumulation`: the household edition one generation further,
  sized by RISK rather than by capital symmetry: every weight is labelled
  derived / mechanism / hedge / constraint, trend capped by the asymmetry of
  behavioral errors, gold sized for the regime nothing else covers, the bond
  overlay rebuilt with two issuers, the linkers dose derived by horizon
  immunization. Carries its own risk-budget reading, regime map, sane ranges,
  blind-spot ledger, scenario map and running rules. The reference file of
  the whole collection.
- `miller-50-40-10`: the 2017 paper build that justifies a trend sleeve:
  50 stocks / 40 bonds / 10 trend, drawn at 4 %/yr.
- `ern-75-25-jeske`: the opposite thesis, and the benchmark to beat: plain
  75/25 US equity / Treasuries drawn at the CAPE-implied rate, from the Early
  Retirement Now series; its notes carry the case against gold, trend and
  risk-parity sleeves and the caveats that case depends on.
- `gentle-decumulation-ucits`: a balanced income build, 4 %/yr withdrawal.
- `global-dividend-income`: a diversified income core for a gentle drawdown.
- `yield-shield-decumulation`: income-tilted, drawn at 4 %/yr (`#meta withdraw`).

## Optimizer and foil

- `optimized-constrained`: the worked example of bounds, limits and a
  held-out `train:` window (see "Using the optimizers" above).
- `predictis`: the odd one out, a real French assurance-vie unit-linked
  allocation, kept as a foil for the builds above (a bond core plus thematic
  satellites, fitted to the previous few years' winners).
