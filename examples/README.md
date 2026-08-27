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

## Simple / lazy (accumulation-friendly)

- `lazy-all-world`: 100% All-World, one fund (Bogleheads "VT and chill").
- `lazy-world-plus-em`: World + EM, cap-weighted ACWI built cheaply.
- `lazy-bogleheads-3fund`: world + EM + global bonds, the canonical lazy build.
- `lazy-aggressive-80-20`: 80% equity / 20% bonds growth allocation.
- `lazy-value-tilt-twofund`: world core + a small-cap-value tilt.
- `coffeehouse-schultheis`: Bill Schultheis' value/small/REIT + bonds.

## Permanent / risk-parity / all-weather

- `permanent-portfolio-browne`: Harry Browne's four-environment 25/25/25/25.
- `all-weather-dalio`: Ray Dalio / Bridgewater All Weather.
- `golden-butterfly`: Portfolio Charts' Permanent + small-value wing.
- `pinwheel-portfolio`: Portfolio Charts' Pinwheel, broad and mildly tilted.
- `larry-portfolio-swedroe`: small-value + safe bonds (efficient equity risk).
- `desert-portfolio-bridges`: low-volatility 30/60/10.

## Tail-risk / trend (anti-fragile)

- `dragon-portfolio-artemis`: Chris Cole's Dragon (equity/bonds/gold/trend/long-vol).
- `claude-dragonlite`: the Dragon distilled to three buyable UCITS lines
  (NTSG/trend/gold, no long vol, growth-heavier than Cole); the family's
  simplicity flagship, with its own design notes and honest range check.
- `dragon-decumulation-household`: dragon-lite re-tuned for decumulation and deployed across French tax wrappers:
  the equity sleeve carved out of the 90/60 core into a PEA/PEA-PME section
  (world swap ETF + small-value boutique funds), the lost bond overlay
  rebuilt with a euro-native long-duration line (accumulating). Covers the wrapper logic
  (spending order as an implicit glidepath, rebalancing asymmetry) and
  measures as the family's best-of-both under the two lenses.
- `risk-budget-decumulation`: the household edition one generation further,
  sized by RISK rather than by capital symmetry. No risk is paid for without
  a documented mechanism, so every weight is labelled derived / mechanism /
  hedge / constraint: trend capped at 20 by the asymmetry of behavioral
  errors (not by the risk budget, which it barely uses), gold at 15 for the
  regime nothing else covers, the bond overlay rebuilt with two issuers
  (euro 25+ plus an unhedged US 20+ as the crisis dollar, capped at 5), the
  linkers dose derived by horizon immunization, and a defensive start meant
  to melt into a rising equity glidepath. Carries its own risk-budget
  reading (equity is 65% of the variance for 42 of capital), regime map,
  sane ranges and blind-spot ledger.
- `risk-budget-decumulation-longhist`: the same book with its euro linker line
  traded for euro cash, which buys thirty years instead of twenty: the Asian
  crisis, LTCM, the **dotcom top and bust**, 2008, 2011, 2020, 2022. Through
  2000-03 to 2003-03 it prints -13.5% against MSCI World's -55.2%. The header
  measures the substitution over the window where both lines exist, so the
  licence to extrapolate travels with the file.
- `risk-budget-decumulation-deephist`: forty years (1986-12), reached with
  three index stand-ins (the BTOP50 managed-futures index for the two trend
  funds, cash for the linkers, a US small-value index for the boutique funds).
  Both substitution errors point the same way (~0.9 pt less return, ~0.8 pt
  more volatility than the real book), so it reads as a floor case. Its
  finding: the worst drawdown of the whole period is neither 2008 nor the
  dotcom bust but the 1990 oil-spike recession, which is exactly the regime
  the parent file names as its worst covered.
- `cockroach-portfolio-mutiny`: Mutiny Fund's four-quadrant Cockroach.
- `risk-parity-plus-trend`: diversified set weighted by the **risk-parity optimizer**
  (`#meta optimize:risk-parity`; run it to see the computed weights).
- `hydra-five-engines-ucits`: one head per return engine (efficient core,
  small value, two trend models, gold, long duration), every line buyable
  from an EU/French retail account; over 1988-2026 it matches the
  dragon-lite blend on CAGR and beats it on vol, Sharpe, recovery time,
  worst rolling 5y and every stress window (2000-02, 2008, 2022). The file
  doubles as the design document: regime map, references, weight ranges,
  blind spots.
- `hydra-five-engines-capital-efficient`: the frontier build (~125% notional
  via stacked funds: NTSG, GDE); over 1996-2026 it matches the S&P 500's EUR
  CAGR on less than half its volatility, and beats the buyable hydra by
  1.2 pts of CAGR, but two pieces (GDE, ZROZ) are US-listed and an EU retail
  investor cannot buy them. Research file: exposure ledger, the volatility/
  CAGR frontier of the shape, alternatives-considered ledger (why RSBT was
  dropped for a EUR-hedged trend leg), and UCITS watch list.
- `cerberus-three-heads`: the hydra's sibling: the same engines regrouped
  into three even notional thirds (equity / trend / bonds+gold), euro
  duration (MTH) instead of the USD zero-coupon, global small value (AVWS),
  both trend legs EUR-hedged; direct US exposure a minority (40% geography,
  41% currency). Over 1996-2026 it matches the S&P 500's EUR return at 46%
  of its volatility, 31% of its drawdown and 15% of its time under water.
  One line (GDE) is not UCITS; the file carries the measured access ledger
  (professional opt-up, a priced CL2+IGLN synthesis, the watch list).

## Capital-efficient / return-stacking (modern)

- `return-stacked-modern`: stacked stocks + bonds + managed futures (RSSB/RSST).
- `ntsx-all-weather`: efficient core + gold/commodities/trend diversifiers.

## Factor-tilted (academic)

- `global-multifactor`: one diversified-factor world fund (JPGL).
- `merriman-style-tilt`: Paul Merriman's value/small worldwide tilt.
- `four-factor-blend`: equal-weight value/momentum/quality/min-vol.

## Income / decumulation

- `global-dividend-income`: diversified income core for a gentle drawdown.
- `yield-shield-decumulation`: income-tilted, drawn at 4%/yr (`#meta withdraw`).

## FIRE glidepath (curated, EUR-liability aware)

A coherent family for a European early retiree withdrawing ~3%/yr real:
one capital-efficient engine (NTSG), a two-engine trend sleeve, gold, and
a defensive pocket matched to a real-euro liability (short euro linkers +
cash). The stages are meant to be compared side by side:

```sh
./pofo examples/fire-bond-tent-departure.txt examples/fire-decumulation-core.txt examples/fire-glidepath-late.txt
./pofo -fire examples/fire-core-longhist.txt    # ruin-probability explorer
```

- `fire-accumulation-runway`: the last high-risk years before the switch
  (efficient core + small-value tilt + starter trend/gold, monthly DCA).
- `fire-bond-tent-departure`: year-0 build, defensive tent inflated to ~20%
  for the sequence-risk window (Kitces/Pfau bond tent).
- `fire-decumulation-core`: the cruise allocation (56/18/12/14 across
  equity+bonds / trend / gold / linkers+cash), drawn at 3%/yr.
- `fire-glidepath-late`: years ~8+, the tent consumed, equity drifted up
  (rising equity glidepath endpoint).
- `fire-core-longhist`: same cruise mix, every leg reaching the late 1980s;
  feed this one to `pofo -fire` so bootstrap/cohorts have real depth.
- `fire-macro-leg-bhmg`: the core plus a small discretionary-macro leg
  (BH Macro) carved out of the trend budget; compare and decide.
- `fire-simple-no-leverage`: the control group; plain unleveraged bricks
  only, so you can see what the complexity above actually buys.
- `fire-liability-buffer`: the defensive tent alone (cash + short euro
  linkers + hedged short TIPS); run it to trust it.
- `fire-trend-sleeve-lab`: one leg per trend ENGINE (DBi replication,
  Winton, AQR, MLM index); a correlation lab against engine-doubling.
- `miller-50-40-10`: the 2017 paper build that justifies the trend sleeve:
  50 stocks / 40 bonds / 10 trend, drawn at 4%/yr.
- `ern-75-25-jeske`: the opposite thesis, and the benchmark to beat: plain
  75/25 US equity/Treasuries drawn at the CAPE-implied rate (~2.7%), from
  the Early Retirement Now SWR series. Its notes carry the case against
  gold/trend/risk-parity sleeves, the window and horizon caveats that case
  depends on, and the small-value variant the series itself prefers. Run it
  next to `miller-50-40-10` and `golden-butterfly` to see the disagreement.
- `stagflation-bunker`: only what works in persistent inflation (trend,
  commodities, gold, short linkers); the regime lab (Neville et al. 2021).

## PEA-eligible (French equity wrapper)

- `pea-aggressive-growth`: World + S&P 500 + Nasdaq, PEA.

## Curated (built here)

- `modern-all-weather-ucits`: an all-weather you can buy today, UCITS-first.
- `aggressive-accumulation-ucits`: equity-heavy factor build for the growth phase.
- `gentle-decumulation-ucits`: balanced income build, 4%/yr withdrawal.

## Phase mechanics demos

- `balanced-decumulation-demo`: starting capital + 4%/yr withdrawal.

Older simple examples also live here: `tradi-60-40`, `permanent`,
`sp500`, `msci-world`, `optimized`.
