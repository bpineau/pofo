# Reconstructing managed futures

How the bundled managed-futures histories (DBMF and its two UCITS share
classes, KMLM, Simplify CTA, the AQR UCITS pair, the Winton trend-equity
fund, and the trend overlay inside RSST and RSBT) are built, what was
measured, and what is still open. Read this before touching
`pkg/simgen/tsmom.go`, `pkg/simgen/trendanchor.go` or the trend recipes.

## The rule that comes first

**Where a real NAV of the same trade exists, use it.** A futures-price
reconstruction agrees with the fund it replicates at a monthly correlation
near 0.6; the NAV of another real managed-futures fund agrees at 0.7 to 0.97.
`DonorChain` therefore splices real NAVs behind each fund for as far back as
they go, nearest trade first, each volatility-matched to the fund over their
common window. The reconstruction below only covers what no real NAV reaches,
which for this family means before 2007.

| Fund | Real from | Donors, nearest first | Reconstruction only before |
|---|---|---|---|
| DBMF | 2019-05 | AlphaSimplex 2010-08 (0.81), Guggenheim 2007-02 (0.72) | 2007-02 |
| DBMF UCITS USD | 2025-03 | DBMF 2019-05 (0.97), then as above | 2007-02 |
| DBMF UCITS EUR | 2025-04 | same chain, converted at EURUSD spot | 2007-02 |
| AQR UCITS A | 2015-03 | AQMIX 2010-01 (0.93, same manager), Guggenheim 2007-02 | 2007-02 |
| KMLM | 2020-12 | AlphaSimplex 2010-08 (0.69), Guggenheim 2007-02 | 2007-02 |
| Simplify CTA | 2022-03 | AHL 2014-08, AlphaSimplex 2010-08, Guggenheim 2007-02 | 2007-02 |
| AQR UCITS RAEF EUR | 2021-04 | its B EUR sister 2015-03 (1.000, same fund, +0.45 %/yr fee uplift), then the AQR chain hedged to EUR | 2007-02 |

Measured effect on the reconstruction, before and after (daily / weekly
correlation with the fund, tracking error, CAGR gap):

| Fund | Before | After |
|---|---|---|
| DBMF | 0.55 / 0.58, 11.7 %, -5.6 pts | **0.69 / 0.75, 9.7 %, -4.2 pts** |
| DBMF UCITS USD | 0.29 / 0.54, 14.7 %, -13.7 pts | **0.65 / 0.85, 10.0 %, +0.1 pt** |
| DBMF UCITS EUR | 0.15 / 0.27, 19.2 %, -6.2 pts | **0.36 / 0.71, 17.3 %, -0.3 pt** |
| AQR UCITS A | 0.37 / 0.55, 10.6 %, -4.8 pts | **0.73 / 0.92, 6.8 %, +0.1 pt** |
| KMLM | 0.37 / 0.44, 16.6 %, -3.2 pts | **0.63 / 0.65, 12.7 %, -2.3 pts** |
| Simplify CTA | 0.19 / 0.24, 21.3 %, -6.8 pts | **0.44 / 0.45, 17.9 %, -4.6 pts** |

The euro-hedged AQR class runs the same chain and hedges it day by day, the
carry read off the two cash series themselves rather than off a frame: a donor
keeps its own trading calendar, and a date the frame does not hold would
otherwise be silently left unhedged (17 % of them, measured).

The donor brings its own manager and fee load with it, which is the honest
price: the reconstruction is no longer "what this fund would have done" but
"what this trade actually paid, run by the closest managers we can observe".

### The one donor whose fee load is worth correcting (measured, 2026-08)

The rule has an exception, and the euro-hedged AQR class is it. Its donor is
not another manager's fund but ANOTHER SHARE CLASS OF ITSELF, the legacy B EUR
class (LU1103258197): same portfolio, same hedge, same NAV cutoff. What
separates them is a price list. Class B pays a 10 % performance fee over an
EUR STR hurdle, high-on-high, crystallised each 31 March, that the flat-fee
RAEF class does not, and the audited accounts put it at 2.03 points of average
class NAV in the year to 31 March 2026.

Because it is the same trade, the wedge is measurable rather than inferred.
Both classes have real NAVs on four disjoint windows (RAEF CAGR minus B EUR
CAGR, common dates only):

| window | days | RAEF | wedge | daily corr |
|---|---|---|---|---|
| 2021-04-01 .. 2021-12-21 | 176 | -10.74 %/yr | +0.45 pts/yr | 1.000 |
| 2023-10-20 .. 2024-12-31 | 288 | +1.75 %/yr | +0.61 pts/yr | 0.999 |
| 2025-01-01 .. 2025-12-31 | 237 | +13.83 %/yr | +1.58 pts/yr | 0.999 |
| 2026-01-01 .. 2026-07-15 | 128 | +13.75 %/yr | +1.91 pts/yr | 0.999 |

The sign never changes and the magnitude tracks the fund's own return, which
is the fee's signature rather than a drift: folding the post-gap window by
month, class B lags RAEF by 1.92 pts/yr in the months RAEF gains and LEADS it
by 0.11 pts/yr in the months RAEF loses. A constant is the wrong shape for a
performance fee, so the donor segment is lifted by the smallest of the four,
0.45 pts/yr, which is also the only one measured while the fund was losing
money and therefore the one that matches the regime the donor segment covers
(2015-03 to 2021-12, a drought throughout, the class compounding at
-4.45 %/yr). The correction is deliberately too small in any better regime.
On the 2021 overlap it takes the reconstruction's CAGR gap against the real
class from 0.44 points to 0.04.

Two other candidates were measured and rejected. Extending the donor past its
2021-12 hole buys nothing, since real RAEF quotes cover everything after
2021-04, and it would import an artefact: the class resumes on 2023-10-19 with
one stale print at its pre-gap level before the re-seeded class prints 99.70
the next day, a -12.5 % day that by itself drags the post-gap daily
correlation from 0.999 to 0.777. Swapping the 2015-2021 donor for the same
manager's US fund hedged into EUR was tested against the criterion that it
must track RAEF better by at least 0.05 of monthly correlation: it tracks
WORSE, 0.957 against 1.000 on the 2021 overlap and 0.963 against 0.999 on the
2023-2026 one. The sister class keeps the slot.

### The EUR class's daily correlation is only partly a valuation convention (measured, 2026-08)

The unhedged EUR share class validates at a daily correlation of 0.36 against
the real FT NAV where its weekly figure is 0.71 and its monthly one is above
0.9. The natural suspicion is a valuation convention rather than an economic
error: a UCITS NAV is struck at a European valuation point with its own FX
fixing, so part of day t's US move should print in NAV t+1, and the FX leg
should be a fixing, not a Yahoo close. That was measured and found real but too
small to act on.

The test fits the real EUR return against `a * r_DBMF(t) + (1-a) * r_DBMF(t-1)`
divided by an FX return, over the live overlap only (2025-04-08 to 2026-07-31,
305 common dates, 287 consecutive-session pairs once two forward-filled FT
prints are dropped), with `a` on a 0.05 grid and four FX candidates (Yahoo
EURUSD close on t and on t-1, ECB reference fixing on t and on t-1).

| convention | daily | R2 | weekly | monthly |
|---|---|---|---|---|
| shipped (a = 1, Yahoo close on t) | 0.364 | -0.06 | 0.657 | 0.904 |
| a = 1, ECB fixing on t | 0.417 | 0.03 | 0.682 | 0.905 |
| best of the grid: a = 0.75, ECB fixing on t | 0.439 | 0.15 | 0.751 | 0.931 |
| both legs blended at a = 0.70, ECB fixing | 0.459 | 0.20 | 0.750 | 0.934 |
| unconstrained OLS on the four regressors | 0.460 | 0.21 | | |

Both effects exist and point the way theory says. About a quarter to a third of
a US session lands in the next EUR NAV print, and the ECB 16:00 CET fixing
beats the Yahoo close by itself (0.417 against 0.364). But the last row is the
one that decides: a free regression on the same four regressors reaches a
multiple correlation of 0.460, so no convention of this family, tuned or not,
can do better than about +0.10 over what is shipped, and the best constrained
one gains +0.095. That is at or under the adoption bar this was measured
against, for a statistic no user reads (real quotes are grafted from
inception, so the daily convention only ever governs the pre-2025 tail, where
no FT NAV exists to be right about). The same-day convention was kept.

The residual is the real finding: even the best convention explains a fifth of
the daily variance. It is not NAV rounding, which prints two decimals on a
level near 100 to 125, worth 0.4 % of the class's 1.07 % daily standard
deviation. What is left belongs to the fund, not to the clock: the UCITS
wrapper holds its own positions and cash, and its NAV is not a repackaged DBMF
close.

## The three layers

What is left to the reconstruction is assembled in three separable layers, each answering a
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
