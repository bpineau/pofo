# Reconstructing managed futures

How the bundled managed-futures histories (DBMF and its two UCITS share
classes, KMLM, Simplify CTA, the AQR UCITS pair, the Winton trend-equity
fund, and the trend overlay inside RSST and RSBT) are built, what was
measured, and what is still open. Read this before touching
`pkg/simgen/tsmom.go`, `pkg/simgen/trendanchor.go` or the trend recipes.

## The three layers

A reconstruction is assembled in three separable layers, each answering a
different question, each validated on its own.

| Layer | What it decides | Where |
|---|---|---|
| Engine | the daily texture: how a trend book moves inside a month | `TSMOM` |
| Anchor | the monthly path: which months trend paid and which it did not | `AnchorTrend` |
| Pin | the level: what an investable programme keeps of it | `pinTrendIR` |

Keeping them separate is the point. A criticism of the level is not a
criticism of the path, and the two are fixed in different places.

## The measuring stick

Every claim below is measured against a published monthly record of the
trend premium, bundled as the `TREND-TSMOM-USD` reference series (see
`cmd/gen-trend-refdata` for what it is, its terms, and how to regenerate
it). It is a gross academic factor: no fee, no slippage, no capacity
limit, and an information ratio near 0.9 where investable programmes
deliver 0.25 to 0.5. It is the right yardstick for SHAPE and the wrong one
for LEVEL, which is exactly how it is used here.

## What the engine can and cannot do

The engine trades seven markets (three equity legs, two Treasury legs,
gold, crude), because those are the ones the client can fetch with enough
history to reach 1989. Measured over 451 months against the reference:

- monthly correlation 0.59, similar maximum drawdown, the right sign in
  almost every crisis year: it is a genuine trend book, not noise;
- about half the amplitude of the reference in those crisis years (1998
  +20 % against +32 %, 2008 +11 % against +20 %, 2022 +11 % against +20 %,
  all volatility-matched): seven markets cannot reproduce what a programme
  trading fifty does. That is a diversification deficit, not a defect.

### Widening the basket does not fix it (measured, 2026-08)

The obvious answer is more markets. It was tried and rejected on evidence.
Adding a dollar index, six currency pairs and eight more commodities
(US-calendar instruments only, each joining the book when its own history
begins) moved the monthly correlation with the reference from 0.63 to 0.67
but cut the information ratio from 0.68 to 0.49 and deepened the maximum
drawdown from -25 % to -31 %. Three findings explain it, and they are worth
keeping:

- **The added history is the wrong history.** The currency pairs the client
  can fetch start in 2003-2006, which is precisely the trend drought. Four
  of the six carry a NEGATIVE standalone trend information ratio over their
  own window (CHF -0.30, GBP -0.16, AUD -0.07, CAD -0.02) where the
  reference's currency sub-factor earns 0.48 over the full sample.
- **Some commodity series carry roll artefacts.** Front-month contracts are
  spliced without back-adjustment: corn shows -24 %, -17 % and -17 % single
  days, every one of them in July, its expiry month. Grains, softs and
  natural gas were dropped for that reason.
- **Foreign equity indices cannot be mixed in naively.** A market closed on
  a US trading day contributes a zero return, which biases its volatility
  estimate down and its position up. They were left out.

A market must also have a full live signal window before it may be traded;
below that, both the twelve-month signal and the covariance are computed on
padding rather than on prices.

## The anchor

Since the engine cannot manufacture the real sequence of trend months, it
is given one. `AnchorTrend` rewrites each month so its return equals the
reference's, rescaled to the book's volatility target, spreading the
correction evenly across the days of that month: the daily texture, the
intra-month drawdown shape and the crisis timing all still come from the
engine, only the month's total is pinned.

Measured against the real funds, monthly correlation before and after:

| fund | months | engine alone | anchored |
|---|---|---|---|
| DBMF | 85 | 0.54 | 0.60 |
| KMLM | 66 | 0.33 | 0.67 |
| Simplify CTA | 51 | 0.27 | 0.41 |
| AQR UCITS | 135 | 0.43 | 0.72 |

## The pin, and what is still open

The level is set by a constant daily drag calibrated so the reconstruction
realizes a target information ratio over the 2000+ window. The targets are
the realized record of the index each fund tracks, except for the iMGP DBi
family: the SG CTA index is measured net of the underlying managers'
2-and-20 while DBi copies their positions for 0.85 % flat, so pinning that
family to the index would deduct the fee twice (see the constants block in
`recipes.go` for the full argument and the two independent routes to 0.50).

**The open problem.** A constant drag reproduces an index's information
ratio but not its drawdown. Pinning a path whose gross information ratio is
0.9 down to 0.24 costs about eleven points a year at a 17 % volatility
target, which turns the 2011-2021 trend drought into a long bleed: the
Simplify CTA reconstruction draws down 57 % where the SG Trend index it
tracks never went past about 18 %. The arithmetic is unavoidable in that
form (at a fixed volatility target, only a constant can move the
information ratio), so the honest fix is not another drag formula but an
anchor that is already investable, that is, a real CTA index. None could be
sourced: the SG, BarclayHedge and Mount Lucas records are all behind
registration walls. Until one is, read the deep past of these series as
right in shape and conservative in level, with drawdowns deeper than the
industry's own indices.

## Traps

- The shipped `simdata` files graft the real quotes wherever they exist
  (100 % identical from each fund's inception), so a portfolio simulation
  only ever runs the reconstruction BEFORE inception. A validation figure
  measured on the overlap describes the engine, not what a user sees.
- The funds beat the strategy over their own live windows (DBMF by about
  six points a year since 2019). No honest reconstruction of trend
  following reproduces that, and chasing it would be curve fitting.
- Any change to the CTA series breaks two FIRE-book plates, whose tests
  recompute from `pkg/datasets` and fail on drift.
