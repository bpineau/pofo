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
which for this family means before 1996-03.

| Fund | Real from | Donors, nearest first | Reconstruction only before |
|---|---|---|---|
| DBMF | 2019-05 | AlphaSimplex 2010-08 (0.81), Guggenheim 2007-02 (0.72), Man AHL Diversified 1996-03 (0.77) | 1996-03 |
| DBMF UCITS USD | 2025-03 | DBMF 2019-05 (0.97), then as above | 1996-03 |
| DBMF UCITS EUR | 2025-04 | same chain, converted at EURUSD spot | 1996-03 |
| AQR UCITS A | 2015-03 | AQMIX 2010-01 (0.93, same manager), Guggenheim 2007-02, Man AHL 1996-03 (0.71) | 1996-03 |
| KMLM | 2020-12 | AlphaSimplex 2010-08 (0.69), Guggenheim 2007-02, Man AHL 1996-03 (0.57) | 1996-03 |
| Simplify CTA | 2022-03 | AHL US 2014-08, AlphaSimplex 2010-08, Guggenheim 2007-02, Man AHL 1996-03 (0.48) | 1996-03 |
| AQR UCITS RAEF EUR | 2021-04 | its B EUR sister 2015-03 (1.000, same fund, +0.45 %/yr fee uplift), then the AQR chain hedged to EUR | 1996-03 |

Correlations are monthly, measured on each pair's own overlap.

Measured effect of the chain on the reconstruction, before and after (daily /
weekly correlation with the fund, tracking error, CAGR gap). These figures are
measured on each fund's own live window, so a donor that only reaches further
back cannot move them: they are what the chain bought over 2007-2019, and the
1996 donor is graded separately below.

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

### The deepest donor deals weekly, so the engine draws the days (measured, 2026-08)

Man AHL Diversified (IE0000360275, USD accumulating, real net NAVs from
1996-03-26) is the oldest donor of the family and the only one that reaches the
1990s. It is a real record, not a reconstruction of one: over 1996-03-26 to
2026-02-27 it compounds to +1649.50 % against the manager's own published
+1649.47 %, and its crisis years are the ones the trend literature describes,
1998 +41 %, 2008 +33 %, 2022 +12 %, for a maximum drawdown of -32 % across
thirty years.

It also deals WEEKLY until 2016 and daily only after, and the whole segment the
chains keep of it (1996-03 to 2007-02) falls in the weekly era. Spliced as it
stands, that segment ships week-sized steps inside a daily file, and every
statistic that annualizes per observation reads them as a fund five times more
volatile than it is: the raw DBi chain reports 31.9 % annualized volatility over
1997-2006 where the donor's own monthly record says 18.0 %.

The fix keeps every NAV and borrows nothing else. A donor segment whose median
spacing exceeds three calendar days is projected onto the daily calendar of the
reconstruction that already stands behind it, through the same `anchorShape`
that blends a monthly index with a daily one: the donor's NAVs are the anchors,
the reconstruction is the shape. The output passes exactly through every real
NAV, so the weekly and monthly truth is untouched, and takes its day-to-day
moves from the engine, which is the only thing here that can say what a trend
book did between two Wednesdays. Sparse donors carry the level, the engine
carries the texture.

Measured on the shipped files, annualized daily volatility over 1997-2006,
against each fund's own volatility target:

| File | target | before | after |
|---|---|---|---|
| DBMF, DBMF UCITS USD | 11.5 % | 11.9 % | **12.5 %** |
| DBMF UCITS EUR | 11.5 % + FX | 14.6 % | **15.1 %** |
| KMLM | 14 % | 14.5 % | **15.1 %** |
| Simplify CTA | 16 % | 16.6 % | **17.4 %** |
| AQR UCITS A | 9 % | 9.3 % | **9.4 %** |

The "before" column is the previous reconstruction, which had no cadence problem
because it was daily by construction; the point of the table is that replacing a
decade of it with real NAVs left the daily texture where it was, and 31.9 % is
what the same decade would have read without the projection.

What the decade's LEVEL does change, and should: over 1997-2006 the DBMF file
now compounds at 14.6 %/yr where the pinned reconstruction had 19.5 %/yr (KMLM
16.5 against 19.5, Simplify CTA 18.4 against 20.5, the AQR class 11.4 against
12.7). That is a real programme's record, fees and all, in place of an
information-ratio pin, and it is lower.

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
| Pin | the level, for the OVERLAY builds only (RSST, RSBT, Winton) | `pinTrendIR` |

Keeping them separate is the point. A criticism of the level is not a
criticism of the path, and the two are fixed in different places.

The third layer used to apply to everything. It no longer does: the
diversified managed-futures funds take their level from their anchor, because
their anchor is already an investable one (below). Only the funds whose trend
sleeve replicates a pure-trend INDEX still pin, since that index is published
gross and has to be brought down to what an investor keeps.

## The measuring sticks

There are two published monthly records here, and they are not
interchangeable.

**The gross factor**, bundled as `TREND-TSMOM-USD` (see
`cmd/gen-trend-refdata` for what it is, its terms, and how to regenerate it),
is an academic trend premium: no fee, no slippage, no capacity limit, and an
information ratio near 0.9 where investable programmes deliver 0.25 to 0.5. It
is the right yardstick for SHAPE and the wrong one for LEVEL. The overlay
builds anchor on it, because they replicate an index measured the same gross
way, and pin its level afterwards.

**The net composite**, bundled as `TREND-NET-USD` (see
`cmd/gen-trendnet-refdata` for its provenance and terms), is a published
monthly composite of real managed-futures programmes, each entering NET of its
own manager's fees, monthly since 1987 at 9.3 %/yr volatility for an
information ratio of 0.74 over the whole period. It is not a factor: it is what
a book of real programmes actually paid its investors. The deep tails of the
diversified funds anchor on it, and take their level from it too.

The net composite is a FUNDED total return: it earns cash on its collateral,
where the gross factor is published as an excess over cash. `AnchorTrend`
therefore strips its cash leg before rescaling it to a fund's volatility target
and funds it again afterwards (`TrendAnchor.Funded`). Reading it as an excess
index instead hands the deep tail a second cash leg, which in the 1990s is
worth six points a year: the DBMF reconstruction came out at 18 %/yr over
1988-1996 where the industry earned 8. Two unit tests hold that arithmetic in
place.

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
engine, only the month's total is fixed. Which reference it is decides
whether the level comes with the path (the net composite) or has to be pinned
separately afterwards (the gross factor).

Measured against the real funds, monthly correlation before and after:

| fund | months | engine alone | anchored |
|---|---|---|---|
| DBMF | 85 | 0.54 | 0.60 |
| KMLM | 66 | 0.33 | 0.67 |
| Simplify CTA | 51 | 0.27 | 0.41 |
| AQR UCITS | 135 | 0.43 | 0.72 |

## The level, and what is left open (measured, 2026-08)

The level of the deep tail used to be set by a constant daily drag calibrated
so the reconstruction realized a target information ratio over the 2000+
window. A constant drag reproduces an index's information ratio but not its
drawdown: pinning a path whose gross information ratio is 0.9 down to 0.24
costs about eleven points a year at a 17 % volatility target, which turns a
long drought into a long bleed. The honest fix was always an anchor that is
already investable, and this section used to close by saying no such record
could be sourced. One could.

The deep tails now carry the net composite's own months, rescaled to each
fund's volatility target, and NO drag: the anchor is already net of the
managers' fees, so a further drag would charge them twice. What that changes,
measured on the shipped files over 1988-11 to 1996-03:

| File | vol target | segment vol | segment CAGR before | after |
|---|---|---|---|---|
| DBMF, DBMF UCITS USD | 11.5 % | 12.2 % | 18.0 %/yr | **8.3 %/yr** |
| DBMF UCITS EUR | 11.5 % + FX | 17.1 % | 17.5 %/yr | **7.9 %/yr** |
| KMLM | 14 % | 14.9 % | 17.0 %/yr | **8.7 %/yr** |
| Simplify CTA | 16 % | 17.0 % | 17.4 %/yr | **8.9 %/yr** |
| AQR UCITS A | 9 % | 9.6 % | 12.0 %/yr | **7.9 %/yr** |
| AQR UCITS RAEF EUR | 9 % hedged | 9.7 % | 8.2 %/yr | **4.3 %/yr** |

Each fund's volatility target still holds over the segment, to within the two
tenths of a point the change moved it, and the maximum drawdown of every
shipped file is unchanged to the digit (DBMF and DBMF UCITS USD -26.5 %, DBMF
UCITS EUR -32.0 %, KMLM -31.0 %, Simplify CTA -27.4 %, the AQR class -21.7 %
and its EUR-hedged sister -27.6 %): every one of them falls in the donor
segment, not in the reconstruction.

Against the industry's own record, the deep tail's calendar years now track it
rather than doubling it. Compared with a second, independently published
managed-futures index (a different composite of the same industry, so the two
disagree between themselves by several points in several years):

| year | the industry index | DBMF before | DBMF after |
|---|---|---|---|
| 1989 | +1.8 % | +33.4 % | **+2.0 %** |
| 1990 | +21.0 % | +22.5 % | **+16.9 %** |
| 1991 | +3.7 % | +13.7 % | **+16.5 %** |
| 1992 | -0.9 % | +19.7 % | **+2.0 %** |
| 1993 | +10.4 % | +33.0 % | **+15.8 %** |
| 1994 | -0.7 % | -0.6 % | **-2.3 %** |
| 1995 | +13.6 % | +17.0 % | **+15.7 %** |

The mean absolute gap over those seven fully-reconstructed years falls from
13.3 points to 4.2, and the sign agrees in six of seven. The one wide year,
1991, is not the reconstruction's doing: the two industry indices themselves
read +14.7 % and +3.7 % that year, and the reconstruction sits 1.9 points above
its own anchor, which is what rescaling a 9.3 % index to an 11.5 % target
costs. 1996 is excluded from the table because the real donor takes over on
1996-03-26 and the year is mostly its record, not the reconstruction's.

**What is left open.** Two things, both structural.

First, rescaling. A 9.3 %-volatility industry index carried to a 16 or 17 %
target is levered about 1.8 times, and its drawdowns are levered with it. That
is arithmetically what a fund running at that target does, but it means the
deep tail of the hotter funds shows falls roughly twice the industry's own,
and no anchor can fix that: it is the target that says so.

Second, the very back of the file. The reconstructions begin 1988-11-17 and the
net composite begins 1987-01, so the reference covers every month they hold;
but an anchor governs the span BETWEEN two month ends, so the fortnight from
1988-11-17 to that month's end keeps the engine's own returns, as `AnchorTrend`
documents. Two unanchored weeks at the start of a thirty-eight-year file.

From 1996 on, these files are real fund NAVs and none of this applies.

## The error budget on the live overlaps

What agreement to EXPECT from these series, measured on each fund's own real
window (the only place a reconstruction can be graded), as of 2026-08. Columns:
daily, weekly and monthly return correlation, annualized tracking error, and
the CAGR gap (reconstruction minus fund) over the overlap.

| Fund | window | daily | weekly | monthly | TE | CAGR gap |
|---|---|---|---|---|---|---|
| DBMF | 7.2 y | 0.69 | 0.75 | 0.81 | 9.7 % | -4.2 pts |
| DBMF UCITS USD | 1.4 y | 0.65 | 0.85 | 0.98 | 10.0 % | +0.1 pt |
| DBMF UCITS EUR | 1.3 y | 0.36 | 0.71 | 0.90 | 17.3 % | -0.3 pt |
| KMLM | 5.7 y | 0.63 | 0.65 | 0.69 | 12.7 % | -2.3 pts |
| Simplify CTA | 4.4 y | 0.44 | 0.45 | 0.39 | 17.9 % | -4.6 pts |
| AQR UCITS A | 11.4 y | 0.73 | 0.92 | 0.93 | 6.8 % | +0.1 pt |
| AQR RAEF EUR | 0.7 y | none independent | | | 0.1 % | -0.04 pt |
| RSST (overlay) | 2.9 y | 0.90 | 0.90 | 0.88 | 11.8 % | -3.0 pts |
| RSBT (overlay) | 3.5 y | 0.49 | 0.50 | 0.54 | 12.7 % | +1.2 pts |
| Winton (overlay) | 1.2 y | 0.62 | 0.80 | 0.96 | 14.6 % | -7.1 pts |

How to read it, because each row's residual has a KNOWN decomposition:

- **The monthly column is the honest one** for a sleeve held for years. Daily
  and weekly figures are dominated by intra-month texture that no seven-market
  engine, and no donor of another manager, can match tick for tick.
- **A negative CAGR gap on DBMF, KMLM, CTA is not an error to fix.** Those
  funds beat the trade they run over their own live windows (DBMF by six
  points a year against every peer and index). Closing that gap would mean
  granting the manager's alpha to the backcast, which is curve fitting.
  Conversely DBMF UCITS USD, DBMFE and AQR sit within 0.3 pt because their
  nearest donor is the same manager running the same book.
- **The DBMFE daily 0.36 has a measured ceiling of 0.46** (the
  valuation-convention section below): four fifths of its daily residual is
  the UCITS wrapper's own positions and cash. Judge that class on weekly and
  monthly only.
- **Simplify CTA is the structurally worst fit and will stay so**: it tracks
  a pure-trend index at a 16 % vol target while every available donor and the
  net anchor are all-styles composites near 9-13 %. Its level is right to
  about 4.6 pts and its months only agree at 0.39.
- **Windows under two years** (both UCITS DBi classes, Winton, RAEF) make the
  CAGR-gap column mostly noise; the sub-year figures are reported because they
  are all there is, not because they are stable.
- London-listed funds measured against US-close donors (the bond twins, but
  also AHLPX inside the CTA chain) show daily correlation far under weekly for
  pure clock reasons; a weekly figure far above the daily one is the signature
  of a timing artefact, not of a bad reconstruction.

## The data that exists, and the data that does not (surveyed live, 2026-08)

The binding constraint of this whole file is data availability. The survey
below is what a full day of hunting established; re-verify before relying on
it, but do not re-run the dead ends blind.

**Fetchable today, used:**

- Real managed-futures fund NAVs through the client's own sources, reaching
  1996: the donor tables above. The deepest, Man AHL Diversified plc
  (IE0000360275, FT, weekly until 2016), reproduces its manager's published
  thirty-year cumulative to two decimals. The next-oldest candidates found
  (a 1995 EUR futures fund, a 1991 offshore multi-strategy vehicle) were
  REJECTED: the first for a bad programme (IR 0.11 full-period, -52 % maximum
  drawdown, monthly cadence), the second for a +102 % single print at a share
  restructuring and multi-strategy drift. Nothing older than RYMFX (2007)
  exists among US 1940-Act funds: it was the first of its kind.
- The gross academic trend factor, monthly 1985+ (cmd/gen-trend-refdata).
- The net industry composite, monthly 1987+ (cmd/gen-trendnet-refdata). Its
  publisher also serves a DAILY version from 2010, so far unused (see
  improvements).
- Annual (only) figures for the broader CTA index of the same publisher,
  1980-2017, recovered from a public web archive: the deep tail's independent
  calendar-year check.

**Exists but rejected, keep the reasons:**

- A rule-based futures index, monthly 1961-2003, recovered from a 2003 web
  archive of its sponsor's site (the archive is the only public copy; the
  live site went behind a login in 2003). Gross, fully collateralized (its
  returns INCLUDE T-bill interest, most of its 14 %/yr CAGR is 1960s-80s cash
  yield), the 1961-1988 segment is the sponsor's own backtest, and it
  correlates only 0.35 monthly with the net composite. Usable at most as a
  shape cross-check for KMLM's tail, whose benchmark descends from it.

**Walled or nonexistent, verified, do not chase blind:**

- The SG index family (Trend, CTA): public endpoints clip to five years by
  licence. Useful as a live cross-check of the overlay pins, nothing more.
- The broader CTA index's MONTHLY history: a paid product, always was.
- The bank-run liquid managed-futures index (1994+): the platform died with
  its owner's sale; every archived capture already redirected to a login.
- The crowdsourced CTA database: registration wall, bulk data at four
  figures a year, and it starts too late anyway.
- The academic "century of trend, net of fees" series: never published. The
  sitemap and the archive of the publisher's dataset directory were both
  enumerated; the absence is real, not a blocked probe.
- SEC EDGAR full-text search: clean public JSON, fully scriptable, and the
  right idea for pre-2001 partnership NAVs, EXCEPT full-text coverage starts
  in 2001 and the one ideal filing found prints "[To Come]" in place of every
  table. A composite of 1980s public futures partnerships remains possible
  through per-filing document retrieval, at high extraction cost and with
  worse survivorship than the net composite. Fallback only.

## Rebuilding it from scratch

The pipeline, in dependency order, with the invariants a re-implementation
must preserve. Every one of these was learned by breaking it.

1. **Legs.** Fetch the seven markets, extend each behind its long proxy
   (`extend`, `longBack`). Rates are annualized percent LEVELS converted to
   daily accruals (`BuildFrame`); a futures price return is ALREADY an excess
   return, a funded total return is not; mixing those two conventions
   double-charges financing.
2. **Engine** (`TSMOM`). Twelve-month sign signal per market, refreshed every
   21 trading days; inverse-vol weights; the whole book rescaled EVERY DAY to
   the volatility target against an EWMA covariance (half-life 23 days,
   RiskMetrics recursion, seeded from a flat 63-day window). The signal and
   the risk run on different clocks ON PURPOSE: sizing monthly off a flat
   window is how the book once lost 12.6 % in a day against a 10 % target
   (2020-03-12). A leg whose raw returns moved on under half the window's
   days is stale (a forward-filled proxy) and stays flat, or its zeros
   poison the covariance.
3. **Anchor** (`AnchorTrend`). Rewrite each calendar month of the engine's
   output so its total matches the reference month, rescaled to the target
   vol; spread the correction geometrically over the month's days. The
   reference can be an EXCESS factor or a FUNDED total return
   (`TrendAnchor.Funded`): strip the cash leg before rescaling a funded one
   and re-add it after, or the tail gains a phantom cash leg worth six points
   a year in the 1990s. Months the reference does not cover keep the engine's
   own returns.
4. **Level.** Nothing more for the deep tails (the net anchor carries its own
   level). For the overlay builds only, a constant daily drag
   (`pinTrendIR`/`trendDrag`) calibrated on 2000+ brings the gross anchor
   down to the replicated index's realized information ratio.
5. **Donor chain** (`DonorChain`). For each fund, volatility-match every real
   donor NAV to the fund on their common window (excess-over-cash returns,
   scale factor clamped to [0.5, 2], at least 120 common days), then splice
   nearest-first with `ExtendBack`, which rescales the incoming segment to
   the junction level. A donor whose median spacing exceeds three calendar
   days is first projected onto the engine's daily calendar (`densify`, via
   `anchorShape`): its NAVs are anchors, the engine is texture. NEVER splice
   a weekly series raw into a daily file; per-observation statistics will
   read the cadence as volatility (31.9 % measured where the truth was 12.5).
6. **Real quotes last** (`SpliceReal` in the recipe): the fund's own NAVs are
   grafted over everything from inception, byte-exact.
7. **Per-family constants.** Vol targets are the fund's own REALIZED
   volatility on the live overlap (measured anchors, not stated targets); do
   not re-tune them to minimize tracking error, because with correlation near
   0.5 the TE-minimizing reconstruction under-risks by that factor, and
   understating a sleeve's risk is the expensive error in portfolio work.
   The RAEF donor uplift (+0.45 %/yr) and the DBi fee-alpha reasoning are
   measured constants with their derivations in the recipe comments; they do
   not generalize.

Units traps inherited from the repo at large: fees in `simgen` are FRACTIONS
per year; `TrendAnchor` vol targets likewise; `portfolio`/`marketdata.Fees`
are percent. Dates are 00:00 UTC and matched by exact equality.

## Improvements worth attempting, ranked

1. **Validate the engine's daily texture against the net composite's DAILY
   series (2010+).** Free at the same endpoint (the daily dump), never yet
   used. The engine's texture is the one layer with no external check at all;
   this is the cheapest missing validation in the file.
2. **KMLM's deep tail on its own ancestor index.** The archived 1961-2003
   rule-based index is closer kin to KMLM's benchmark than the all-styles
   composite (which it correlates with at only 0.35): anchoring KMLM's
   1988-1996 stub on it (collateral stripped, level checked against the
   composite) might fit that one fund better. Measure before believing.
3. **Repair the third NAV fallback.** The Morningstar timeseries endpoint
   currently answers empty for every id, which narrows donor hunting to two
   sources; if it revives, re-run the 1990s donor survey, the rejected
   candidates deserve a second look through a second source.
4. **The EUR class's valuation convention** (a = 0.75 on the US session, ECB
   fixing): measured, real, and under the adoption bar because it buys only
   +0.08 of a daily correlation nobody consumes. Revisit only if the daily
   texture of the pre-2025 EUR tail ever starts to matter.
5. **A net PURE-TREND record for the overlays.** The overlay builds still pin
   a gross factor because the pure-trend index they replicate is only public
   for five years. If a long net record ever surfaces, the same funded-anchor
   swap applies and `pinTrendIR` retires completely.
6. **The first fortnight.** The reconstructions start 1988-11-17; the anchor
   governs whole months, so the stub to 1988-11-30 keeps raw engine returns.
   Two weeks out of thirty-eight years; fix only if the file's start date
   ever moves earlier, where the problem would grow.

## Traps

- The shipped `simdata` files graft the real quotes wherever they exist
  (100 % identical from each fund's inception), so a portfolio simulation
  only ever runs the chain and the reconstruction BEFORE inception. A
  validation figure measured on the overlap describes the engine, not what a
  user sees, and no donor older than the fund can move it: adding the 1996
  donor left every `-gen-simdata` validation line unchanged, to the digit.
- A donor is not required to quote daily, and the deepest one does not.
  Anything spliced into a daily file must be projected onto a daily calendar
  first (`densify`), or per-observation statistics will read its cadence as
  volatility.
- The funds beat the strategy over their own live windows (DBMF by about
  six points a year since 2019). No honest reconstruction of trend
  following reproduces that, and chasing it would be curve fitting.
- The reconstruction is not confined to the deep tail: it is also the daily
  texture the weekly-dealing 1996-2007 donor is projected onto. A change that
  only touches the pre-1996 level therefore still moves daily-frequency
  statistics over 1996-2007, at the third decimal. The anchor swap moved the
  twelve-month CTA/equity correlation by at most 0.0045 across 306 monthly
  windows, which is invisible in a chart and enough to break a literal frozen
  to two decimals.
- Any change to the CTA series breaks two FIRE-book plates, whose tests
  recompute from `pkg/datasets` and fail on drift. Their tolerance (0.005
  around a two-decimal literal) is finer than the series' own stability under
  any change to the reconstruction, so expect a handful of cells to need
  refreezing whenever this file is touched.
- Fee structures are the largest silent lever in this file. A performance fee
  shows up as a wedge that GROWS with the fund's own return and vanishes in
  down months (the RAEF/B EUR tables above); a composite "net of fees" means
  net of its constituents' 2-and-20, so a flat-fee replicator should sit
  ABOVE it, not on it; and correcting a fee twice (a drag on an already-net
  anchor, a wedge on an already-lagging donor) is the single easiest way to
  quietly lose several points a year. Every fee decision here carries its
  measurement next to it; keep that discipline.
- Index composites finalize late. The net reference's recent months revise
  for one to two years before freezing (measured against point-in-time
  archive captures), and its last one or two months are flagged estimated at
  the source, which is why the generator drops them. A regeneration that
  shifts recent tail months by a few basis points is revision, not breakage.
- Share classes get emptied and re-seeded. The B EUR donor resumes after its
  two-year hole with one stale print at the pre-gap level followed by a
  re-seeded NAV, a spurious -12.5 % day; `truncateAtGap` exists because of
  it. Any donor with a NAV gap needs its resumption inspected before use.
- FT NAVs carry occasional bad prints (the distributing bond class printed
  -13.1 % on a day its twin moved -5.6 %) and forward-filled stales; drop
  zero-return days before fitting anything to an FT series, and never graft
  a distributing class's price-only NAV into a total-return series.
- An unknown ISIN used to adopt an unrelated FT security through fuzzy
  search (a real case served a US media company's prices as an Austrian
  fund); the client now rejects FT hits whose symbol does not contain the
  queried ISIN. If a donor probe ever shows a delirious series, check the
  resolution line first and clear the cache entry.
