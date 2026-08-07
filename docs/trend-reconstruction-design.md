# Reconstructing managed futures

How the bundled managed-futures histories (DBMF and its two UCITS share
classes, KMLM, Simplify CTA, the AQR UCITS pair, the Winton trend-equity
fund, and the trend overlay inside RSST and RSBT) are built, what was
measured, and what is still open. Read this before touching
`pkg/simgen/tsmom.go`, `pkg/simgen/trendanchor.go` or the trend recipes.

## Reliability bounds length (2026-08, maintainer decision)

Every file described here now stops where its evidence stops, and no earlier
history is shipped even though the engine can compute one:

- the donor chains (DBMF and its UCITS classes, KMLM, Simplify CTA, the AQR
  pair) start at their deepest REAL donor's first NAV, **1996-03-26**. The
  anchored engine tail that used to stand in front of it, back to 1988-11, is
  no longer shipped.
- the overlay builds (RSST, RSBT, Winton) start where their reference does,
  **2000-01-03**, and no longer carry an information-ratio pin at all.

The decision is deliberate and it is a trade: a very reliable twenty-plus-year
backcast is worth more than a heavily simulated forty-year one. An engine
anchored on a monthly composite is a decent account of a decade in aggregate
and a poor account of any single month a reader might look up, and a portfolio
tool invites exactly that lookup. What the removed tail was, and what it
measured, is kept below: the measurements are the argument for the decision,
not a description of what ships.

## The rule that comes first

**Where a real NAV of the same trade exists, use it.** A futures-price
reconstruction agrees with the fund it replicates at a monthly correlation
near 0.6; the NAV of another real managed-futures fund agrees at 0.7 to 0.97.
`DonorChain` therefore splices real NAVs behind each fund for as far back as
they go, nearest trade first, each volatility-matched to the fund over their
common window and lifted to the fund's own fee load (below). Where no real NAV
reaches, the file ends.

| Fund | Real from | Donors, nearest first | File starts |
|---|---|---|---|
| DBMF | 2019-05 | AlphaSimplex 2010-08 (0.81), Guggenheim 2007-02 (0.72), Man AHL Diversified 1996-03 (0.77) | 1996-03-26 |
| DBMF UCITS USD | 2025-03 | DBMF 2019-05 (0.97), then as above | 1996-03-26 |
| DBMF UCITS EUR | 2025-04 | same chain, converted at EURUSD spot | 1996-03-26 |
| AQR UCITS A | 2015-03 | AQMIX 2010-01 (0.93, same manager), Guggenheim 2007-02, Man AHL 1996-03 (0.71) | 1996-03-26 |
| KMLM | 2020-12 | AlphaSimplex 2010-08 (0.69), Guggenheim 2007-02, Man AHL 1996-03 (0.57) | 1996-03-26 |
| Simplify CTA | 2022-03 | AHL US 2014-08, AlphaSimplex 2010-08, Guggenheim 2007-02, Man AHL 1996-03 (0.48) | 1996-03-26 |
| AQR UCITS RAEF EUR | 2021-04 | its B EUR sister 2015-03 (1.000, same fund, +0.45 %/yr fee uplift), then the AQR chain hedged to EUR | 1996-03-26 |

Correlations are monthly, measured on each pair's own overlap.

Full-period statistics of the shipped files after the truncation (daily
returns, the fund's own real quotes included from its inception):

| File | window | CAGR | volatility | max drawdown |
|---|---|---|---|---|
| DBMF | 1996-03-26 .. 2026-07-31 | 9.66 % | 11.8 % | -24.9 % |
| DBMF UCITS USD | 1996-03-26 .. 2026-07-30 | 9.75 % | 11.9 % | -24.6 % |
| DBMF UCITS EUR | 1996-03-26 .. 2026-08-03 | 10.01 % | 15.5 % | -30.3 % |
| KMLM | 1996-03-26 .. 2026-07-31 | 9.88 % | 13.9 % | -31.0 % |
| Simplify CTA | 1996-03-26 .. 2026-07-31 | 12.49 % | 16.2 % | -25.3 % |
| AQR UCITS A | 1996-03-26 .. 2026-07-31 | 7.04 % | 9.0 % | -21.7 % |
| AQR RAEF EUR | 1996-03-26 .. 2026-07-15 | 6.89 % | 9.2 % | -27.6 % |
| RSST | 2000-01-03 .. 2026-07-31 | 10.11 % | 22.1 % | -48.8 % |
| RSBT | 2000-01-03 .. 2026-07-31 | 5.45 % | 12.6 % | -26.3 % |
| Winton | 2000-01-03 .. 2026-08-06 | 7.88 % | 19.3 % | -53.8 % |

Every maximum drawdown falls inside the donor era or the real quotes, not in
what was removed: the truncation took away length, not the worst days.

Measured effect of the chain on the reconstruction, before and after (daily /
weekly correlation with the fund, tracking error, CAGR gap). These figures are
measured on each fund's own live window, so a donor that only reaches further
back cannot move them: they are what the chain bought over 2007-2019, and the
1996 donor is graded separately below.

| Fund | Before | After |
|---|---|---|
| DBMF | 0.55 / 0.58, 11.7 %, -5.6 pts | **0.69 / 0.75, 9.7 %, -3.5 pts** |
| DBMF UCITS USD | 0.29 / 0.54, 14.7 %, -13.7 pts | **0.65 / 0.85, 10.0 %, +0.2 pt** |
| DBMF UCITS EUR | 0.15 / 0.27, 19.2 %, -6.2 pts | **0.36 / 0.71, 17.3 %, -0.2 pt** |
| AQR UCITS A | 0.37 / 0.55, 10.6 %, -4.8 pts | **0.73 / 0.92, 6.8 %, +0.1 pt** |
| KMLM | 0.37 / 0.44, 16.6 %, -3.2 pts | **0.63 / 0.65, 12.7 %, -1.7 pts** |
| Simplify CTA | 0.19 / 0.24, 21.3 %, -6.8 pts | **0.44 / 0.45, 17.9 %, -3.4 pts** |

The euro-hedged AQR class runs the same chain and hedges it day by day, the
carry read off the two cash series themselves rather than off a frame: a donor
keeps its own trading calendar, and a date the frame does not hold would
otherwise be silently left unhedged (17 % of them, measured).

The donor brings its own manager with it, which is the honest price: the
reconstruction is no longer "what this fund would have done" but "what this
trade actually paid, run by the closest managers we can observe". It no longer
brings its own FEE load, which was never honest, only inherited (below).

### Every donor is fee-aligned, from price lists only (2026-08)

A donor is spliced for its returns, and those returns are net of ITS price
list. The donors of this family are old-school 1940-Act funds and one offshore
vehicle charging 1.3 to 2.7 %/yr; the funds they stand in for are modern ETFs
and UCITS classes at 0.75 to 0.90 %. Left alone, every donor segment runs about
a point a year colder than the fund it covers for, and that cold belongs to the
wrapper, not to the trade.

Each donor segment is therefore lifted by a constant equal to the donor's
MANAGEMENT-AND-EXPENSE load minus the target's, floored at zero. The loads are
read off published fee tables and NEVER off an observed return gap: a gap
between two managers contains their skill, and this family has a lot of it
(DBi beats every peer over its own live window by about six points a year), so
a "fee" constant fitted to a wedge would quietly grant that alpha to the
backcast.

| vehicle | load | source |
|---|---|---|
| ASFYX, Virtus AlphaSimplex class I | 1.45 % | 1.59 % total, 1.45 % after the contractual cap running to 2027 (summary prospectus, 2026-04) |
| RYMFX, Guggenheim class P | 1.99 % | 2.18 % total, 1.99 % after waiver (summary prospectus, 2026) |
| AHLPX, American Beacon AHL Investor | 1.91 % | 1.91 % total, no waiver in force (prospectus supplement, 2025-08) |
| AQMIX, AQR Managed Futures class I | 1.29 % | 1.29 % total and after reimbursement (summary prospectus, 2024-05) |
| Man AHL Diversified plc, USD acc | 2.74 % | ongoing charge for the ISIN (fund database, 2026) |
| DBMF | 0.85 % | fund page, 2026 |
| DBi UCITS classes | 0.75 % | fund pages, 2026 |
| KMLM | 0.90 % | fund page, 2026 |
| Simplify CTA | 0.75 % | fund page, 2026 |
| AQR UCITS class A | 0.79 % + 1.58 % | 0.60 % management + 0.18 % expense cap + 0.01 % subscription tax (prospectus supplement), plus a 10 % performance fee over the T-bill hurdle worth 1.58 points of average class NAV in the year to 31 March 2026 (audited accounts) |

Performance fees enter asymmetrically, and both directions are deliberately
conservative, meaning both can only make the uplift too SMALL:

- the DONOR's performance fee is ignored. Man AHL's audited accounts add a
  1.00 %/yr introducing broker fee and a 20 % performance fee on net new
  profits to the 2.74 % above, and Guggenheim's swap counterparties charge
  their own management and performance fees inside the swap returns, outside
  the fee table. None of that is claimed back.
- the TARGET's performance fee is subtracted, because the target's own record
  is already net of it and the donor owes only the DIFFERENCE. That is what
  keeps the AQR chain honest: class A's base list is 0.50 points under AQMIX's,
  but class A also pays a performance fee AQMIX does not, so the aligned uplift
  is zero rather than half a point. Measurement agrees, the AQMIX chain already
  tracking class A's live CAGR to within 0.1 point.

Applied, per chain (nearest donor first, the one that governs each fund's own
validation window):

| chain | nearest | then | then | deepest (Man AHL) |
|---|---|---|---|---|
| DBMF | ASFYX +0.60 | RYMFX +1.14 | | +1.89 |
| DBMF UCITS USD, DBMF UCITS EUR | DBMF +0.10 | ASFYX +0.70 | RYMFX +1.24 | +1.99 |
| KMLM | ASFYX +0.55 | RYMFX +1.09 | | +1.84 |
| Simplify CTA | AHLPX +1.16 | ASFYX +0.70 | RYMFX +1.24 | +1.99 |
| AQR UCITS A, and the USD leg of RAEF | AQMIX 0 | RYMFX 0 | | +0.37 |

What it moves, on each fund's own live overlap (CAGR gap, reconstruction minus
fund):

| Fund | before | after |
|---|---|---|
| DBMF | -4.2 pts | **-3.5 pts** |
| DBMF UCITS USD | +0.1 pt | **+0.2 pt** |
| DBMF UCITS EUR | -0.3 pt | **-0.2 pt** |
| KMLM | -2.3 pts | **-1.7 pts** |
| Simplify CTA | -4.6 pts | **-3.4 pts** |
| AQR UCITS A | +0.1 pt | **+0.1 pt** (unchanged, uplift 0) |
| AQR RAEF EUR | -0.04 pt | **-0.04 pt** (unchanged, see below) |

The three funds whose donor is another manager close a little; the three whose
nearest donor is the same manager and nearly the same price list barely move,
which is the test that the correction is a fee correction and not a fudge. No
gap crosses zero: what remains negative is the managers' own edge, and it stays
open.

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

## The two layers

What is left to the reconstruction is assembled in two separable layers, each
answering a different question, each validated on its own.

| Layer | What it decides | Where |
|---|---|---|
| Engine | the daily texture: how a trend book moves inside a month | `TSMOM` |
| Anchor | the monthly path AND the level | `AnchorTrend` |

Keeping them separate is the point. A criticism of the level is not a
criticism of the texture, and the two are fixed in different places.

There used to be a third layer, a constant daily drag that pulled a gross
index's level down to the information ratio an investable programme realizes
(`pinTrendIR`). It is gone. A drag reproduces an index's information ratio but
not its drawdown: it leaves every crisis spike where it was and turns every
drought into a bleed. Both families now anchor on a record that is already net
of the constituent managers' fees, so the level arrives with the path and
nothing corrects it afterwards. The information ratios that used to do the
levelling are kept as a measured record below.

## The measuring sticks

There are three published records here, and they are not interchangeable.

**The gross factor**, bundled as `TREND-TSMOM-USD` (see
`cmd/gen-trend-refdata` for what it is, its terms, and how to regenerate it),
is an academic trend premium: no fee, no slippage, no capacity limit, and an
information ratio near 0.9 where investable programmes deliver 0.25 to 0.5. It
is the right yardstick for SHAPE and the wrong one for LEVEL. Nothing anchors
on it any more; it is kept as the yardstick it is.

**The net all-styles composite**, bundled as `TREND-NET-USD` (see
`cmd/gen-trendnet-refdata` for its provenance and terms), is a published
monthly composite of real managed-futures programmes, each entering NET of its
own manager's fees, monthly since 1987 at 9.3 %/yr volatility for an
information ratio of 0.74 over the whole period. It is not a factor: it is what
a book of real programmes actually paid its investors. It is the daily
texture's monthly anchor inside the donor chains, and it is the independent
cross-check the chains are read against.

**The net pure-trend composite**, bundled as `TREND-PURE-NET-USD` (see
`cmd/gen-sgtrend-refdata` for its provenance, its two sources, its licence
restriction and how to regenerate it), is the same kind of record for the
narrower trade an overlay runs: a published composite of the largest
trend-following programmes open to new investment, each net of its own
manager's fees, DAILY since 2000-01-03, at 13.3 %/yr volatility for a funded
excess-over-T-bill information ratio of 0.26 over the whole period. The overlay
builds take both their months and their level from it, and it is why they no
longer pin anything.

The two net records are FUNDED total returns: they earn cash on their
collateral, where the gross factor is published as an excess over cash.
`AnchorTrend` therefore strips the cash leg before rescaling one to a
volatility target and funds it again afterwards, or not, depending on whether
the OUTPUT is a fund or an overlay (`TrendAnchor.Funded`, `earnCash`). Reading
a funded record as an excess index hands the reconstruction a second cash leg,
which in the 1990s was worth six points a year. Two unit tests hold that
arithmetic in place.

A reference may also be published daily, and what is anchored is still a
month: `AnchorTrend` reduces any reference to its month ends first. Read a
daily index as it stands and the anchor takes each month's FIRST day for the
month itself, which is a different index; a third unit test holds that.

### The information ratios, kept as a record (measured, 2026-08)

Nothing applies these any more. They are what any future claim about these
funds' level has to be argued against, and each was triangulated from an
index's public long-run record and cross-checked against the fund's own real
quotes.

| ratio | what it measures |
|---|---|
| 0.24 | the pure-trend index the overlays replicate: ~4.9 %/yr at ~13.6 % volatility over a ~1.6 % cash rate since 2000. It used to pin RSST, RSBT and the Winton sleeve, because the bare engine earns an in-sample ratio near 0.7 over the same window and grafting that onto a funded core overstated the backcast by ~4 %/yr |
| 0.50 | the iMGP DBi family. The all-styles hedge-fund index those funds replicate is measured net of the underlying managers' 2-and-20 while DBi copies their positions at 0.85 % flat, so pinning them to the index's own 0.25 would take that fee off twice: adding back the ~3.5 points a 2-and-20 costs lands near 0.55, and the live DBMF record (9.2 %/yr at 12.4 % volatility over ~2.5 % cash since 2019) lands at 0.54 from an independent direction |
| 0.25 | the same all-styles index as published, the underlying managers' fees included |
| 0.30 | the rule-based index KMLM tracks, whose crisis-alpha profile (2000 +38 %, 2008 +40 %, 2022 +45 % at a ~15 % target) is genuinely richer than the composites above; the live ETF window (2020+, ratio ~0.19) set the floor |
| 0.20 | AQR's own managed-futures record: AQMIX realized ~3.4 %/yr (ratio ~0.15) over a winter-heavy 2015-2026, and 0.20 credited the strong pre-2008 era a backcast also covers |

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
engine, only the month's total is fixed. With a net reference the level comes
with the path, so nothing corrects it afterwards.

Measured against the real funds, monthly correlation before and after:

| fund | months | engine alone | anchored |
|---|---|---|---|
| DBMF | 85 | 0.54 | 0.60 |
| KMLM | 66 | 0.33 | 0.67 |
| Simplify CTA | 51 | 0.27 | 0.41 |
| AQR UCITS | 135 | 0.43 | 0.72 |

### The overlays, on their own reference (measured, 2026-08)

The three overlay builds swapped the gross factor plus a pin for the net
pure-trend record, and dropped the pin. Measured against each fund's real
quotes over its own live overlap:

| build | overlap | monthly correlation | CAGR gap |
|---|---|---|---|
| RSST | 2.9 y | 0.879 -> **0.957** | -3.0 pts -> **-1.3 pts** |
| RSBT | 3.5 y | 0.544 -> **0.838** | +1.2 pts -> **+1.4 pts** |
| Winton | 1.2 y | 0.956 -> **0.972** | -7.1 pts -> **-4.2 pts** |

The months improve everywhere, RSBT's transformationally, and two of the three
level gaps close. RSBT's widens by 0.2 of a point, which is noise on that
window and not a finding: at a 12.6 %/yr tracking error over 3.5 years the
standard error of the gap is about 7 points, so a two-tenths move measures
nothing. The mechanism, for the record, is that the overlay runs at a 10 %
volatility target against the reference's 13.3 %, so it is scaled to about
three quarters of the index; in a stretch the index spent losing money, that
makes the reconstruction lose slightly less than the fund.

The chains were untouched by the swap, and their validation lines are
identical to the digit before and after (correlation, beta, tracking error,
CAGR), which is the check that the anchor change reached only the overlays.

## The tail that was removed, and why it went (measured, 2026-08)

Until 2026-08 the donor chains shipped an engine tail back to 1988-11, anchored
on the net all-styles composite. That tail is no longer shipped. The
measurements below are why the decision was possible to make on evidence rather
than on taste, and they remain the record of what such a tail is worth.

The tail's own level was honest as far as it went. Anchored on the net
composite rather than pinned to an information ratio, over 1988-11 to 1996-03
it compounded at 8.3 %/yr for DBMF (7.9 % for the EUR class, 8.7 % KMLM, 8.9 %
Simplify CTA, 7.9 % the AQR class, 4.3 % its hedged sister) against 18.0, 17.5,
17.0, 17.4, 12.0 and 8.2 %/yr for the pinned version it replaced. Each fund's
volatility target held over the segment to within two tenths of a point, and
no fund's maximum drawdown moved: every one of them falls in the donor era or
in the real quotes.

Its calendar years also tracked a second, independently published
managed-futures index rather than doubling it (the two industry indices
disagree between themselves by several points in several years):

| year | the industry index | DBMF pinned | DBMF anchored |
|---|---|---|---|
| 1989 | +1.8 % | +33.4 % | **+2.0 %** |
| 1990 | +21.0 % | +22.5 % | **+16.9 %** |
| 1991 | +3.7 % | +13.7 % | **+16.5 %** |
| 1992 | -0.9 % | +19.7 % | **+2.0 %** |
| 1993 | +10.4 % | +33.0 % | **+15.8 %** |
| 1994 | -0.7 % | -0.6 % | **-2.3 %** |
| 1995 | +13.6 % | +17.0 % | **+15.7 %** |

The mean absolute gap over those seven fully-reconstructed years fell from 13.3
points to 4.2, and the sign agreed in six of seven.

Four points a year of residual error on a calendar year, and a monthly path
that is a rescaled industry aggregate rather than the fund's own, is a decent
account of a decade and a poor account of a month. Two structural problems
stayed open on top of it, and neither had a fix:

- **Rescaling.** A 9.3 %-volatility industry index carried to a 16 or 17 %
  target is levered about 1.8 times, and its drawdowns are levered with it.
  That is arithmetically what a fund at that target does, but it means the tail
  of the hotter funds showed falls roughly twice the industry's own, and no
  anchor can fix that: it is the target that says so.
- **The unanchored fortnight.** An anchor governs the span BETWEEN two month
  ends, so the stub from the tail's first day to that month's end kept the
  engine's own returns, as `AnchorTrend` documents.

Given the choice between eight more years of that and a file that is real
throughout, the maintainer took the shorter file (see the decision at the top).
The engine that produced the tail is still built and still matters: it is the
daily texture the weekly-dealing 1996-2007 donor is projected onto.

## The error budget on the live overlaps

What agreement to EXPECT from these series, measured on each fund's own real
window (the only place a reconstruction can be graded), as of 2026-08. Columns:
daily, weekly and monthly return correlation, annualized tracking error, and
the CAGR gap (reconstruction minus fund) over the overlap.

| Fund | window | daily | weekly | monthly | TE | CAGR gap |
|---|---|---|---|---|---|---|
| DBMF | 7.2 y | 0.69 | 0.75 | 0.81 | 9.7 % | -3.5 pts |
| DBMF UCITS USD | 1.4 y | 0.65 | 0.85 | 0.98 | 10.0 % | +0.2 pt |
| DBMF UCITS EUR | 1.3 y | 0.36 | 0.71 | 0.90 | 17.3 % | -0.2 pt |
| KMLM | 5.7 y | 0.63 | 0.65 | 0.69 | 12.7 % | -1.7 pts |
| Simplify CTA | 4.4 y | 0.44 | 0.45 | 0.39 | 17.9 % | -3.4 pts |
| AQR UCITS A | 11.4 y | 0.73 | 0.92 | 0.93 | 6.8 % | +0.1 pt |
| AQR RAEF EUR | 0.7 y | none independent | | | 0.1 % | -0.04 pt |
| RSST (overlay) | 2.9 y | 0.88 | 0.89 | 0.96 | 11.7 % | -1.3 pts |
| RSBT (overlay) | 3.5 y | 0.50 | 0.47 | 0.84 | 12.6 % | +1.4 pts |
| Winton (overlay) | 1.2 y | 0.62 | 0.84 | 0.97 | 14.6 % | -4.2 pts |

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
- **The three overlay rows are the newest and the best-behaved monthly fits**
  (0.84 to 0.97), because their reference is a record of the very trade their
  sleeve runs rather than an all-styles composite. Their daily columns stay
  ordinary: an anchor governs months, not days.
- **Windows under two years** (both UCITS DBi classes, Winton, RAEF) make the
  CAGR-gap column mostly noise; the sub-year figures are reported because they
  are all there is, not because they are stable. The same arithmetic applies to
  RSBT at 3.5 years and a 12.6 % tracking error: its gap carries a standard
  error near 7 points.
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
- The net all-styles industry composite, monthly 1987+
  (cmd/gen-trendnet-refdata). Its publisher also serves a DAILY version from
  2010, so far unused (see improvements).
- The net PURE-TREND composite, DAILY 2000-01-03+ (cmd/gen-sgtrend-refdata,
  which holds its provenance). This was the file's headline gap and it is
  closed. Two independent channels serve it: the calculation agent's own
  full-precision dump (six decimals, one POST) and the publisher's dashboard
  copy (two decimals). They are not copies of one another and they agree on
  every one of 6921 common daily returns to within 2 bp, worst gap 1.16 bp,
  mean 0.16 bp, which is what rounding a level to two decimals costs; the
  generator refuses to write unless that still holds over at least 6500 days.
  Its calendar years were reconciled against six independent publications of
  them, the oldest an archived 2010 capture, and the index has never been
  restated by more than 5 bp, on 2018 alone. Its publisher attaches an EU
  Benchmarks Regulation restriction to it (not to be used as a benchmark by a
  financial product); it is bundled as a reference series for a research
  reconstruction, which is not that use.
- Annual (only) figures for the broader CTA index of the same publisher,
  1980-2017, recovered from a public web archive: an independent
  calendar-year check.

**Exists but rejected, keep the reasons:**

- A rule-based futures index, monthly 1961-2003, recovered from a 2003 web
  archive of its sponsor's site (the archive is the only public copy; the
  live site went behind a login in 2003). Gross, fully collateralized (its
  returns INCLUDE T-bill interest, most of its 14 %/yr CAGR is 1960s-80s cash
  yield), the 1961-1988 segment is the sponsor's own backtest, and it
  correlates only 0.35 monthly with the net composite. A shape cross-check at
  most, and with the deep tails gone it has nothing left to anchor.
- A bank's rule-based CROSS-ASSET TREND index, symmetrical long/short, daily
  and freely served from 1994-12, excess return in USD. It is the only free
  daily trend series found that predates the pure-trend composite, so it was
  graded as a possible deeper reference. It tracks WORSE than what the repo
  already has, everywhere it was measured: 0.659 monthly against KMLM's live
  quotes where the pure-trend composite reaches 0.684, 0.70 to 0.72 against the
  shipped chains over 1996-2007 where the chains' own donor NAVs are the truth,
  and its 2023-2024 (-12.5 %, -10.0 %) diverges from every managed-futures
  index of those years. Kept as a cross-check only; it anchors nothing.
- Its LONG-BIASED sibling (multi-asset trend, daily from 1998-12) is rejected
  outright. It correlates 0.482 monthly with KMLM's live window and lost
  3.9 %/yr over it while trend earned 9; it returned -1.7 % in 2022, the year
  every trend index earned +16 to +27 %, which is the long bias showing; and
  its public feed stops in 2025-06.

**Walled or nonexistent, verified, do not chase blind:**

- The pure-trend index's own publisher-side endpoints clip to five years by
  licence. That was long believed to be the whole of it; it is not, and the
  two channels above are the answer. This entry stays as a warning that a
  licence-clipped front end says nothing about what the calculation agent
  serves.
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
4. **Level.** Nothing. Both net references carry their own investable level,
   and a drag on top of one would charge the constituent managers' fees twice.
   A build that takes its level from a reference must not ship a day in front
   of it either: `trimToAnchor`/`AnchorStart` cut the overlays at their
   reference's first date.
5. **Donor chain** (`DonorChain`). For each fund, volatility-match every real
   donor NAV to the fund on their common window (excess-over-cash returns,
   scale factor clamped to [0.5, 2], at least 120 common days), lift each
   segment by its documented fee uplift (`feeAligned`, donor load minus target
   load, performance fees only ever subtracted), then splice nearest-first with
   `ExtendBack`, which rescales the incoming segment to the junction level. A donor whose median spacing exceeds three calendar
   days is first projected onto the engine's daily calendar (`densify`, via
   `anchorShape`): its NAVs are anchors, the engine is texture. NEVER splice
   a weekly series raw into a daily file; per-observation statistics will
   read the cadence as volatility (31.9 % measured where the truth was 12.5).
   The chain is where the file STARTS: nothing is spliced behind the deepest
   donor.
6. **Real quotes last** (`SpliceReal` in the recipe): the fund's own NAVs are
   grafted over everything from inception, byte-exact.
7. **Per-family constants.** Vol targets are the fund's own REALIZED
   volatility on the live overlap (measured anchors, not stated targets); do
   not re-tune them to minimize tracking error, because with correlation near
   0.5 the TE-minimizing reconstruction under-risks by that factor, and
   understating a sleeve's risk is the expensive error in portfolio work.
   The RAEF donor uplift (+0.45 %/yr) and the DBi fee-alpha reasoning are
   measured constants with their derivations in the recipe comments; they do
   not generalize. The donor fee uplifts (`trendFeeLoad`, `trendPerfFee`) are
   the opposite: published prices, not measurements, and a new donor must have
   one before it may be spliced (`feeLoad` panics otherwise).

Units traps inherited from the repo at large: fees in `simgen` are FRACTIONS
per year; `TrendAnchor` vol targets likewise; `portfolio`/`marketdata.Fees`
are percent. Dates are 00:00 UTC and matched by exact equality.

## Improvements worth attempting, ranked

1. **Validate the engine's daily texture against a daily reference.** Two are
   now available: the all-styles composite's daily version (2010+, free at the
   same endpoint, never used) and the bundled pure-trend one (2000+, already in
   the repo). The engine's texture is the one layer with no external check at
   all, and it is now the layer that carries the whole 1996-2007 donor era
   between its weekly NAVs. This is the cheapest missing validation in the file.
2. **Repair the third NAV fallback.** The Morningstar timeseries endpoint
   currently answers empty for every id, which narrows donor hunting to two
   sources; if it revives, re-run the 1990s donor survey, the rejected
   candidates deserve a second look through a second source.
3. **The EUR class's valuation convention** (a = 0.75 on the US session, ECB
   fixing): measured, real, and under the adoption bar because it buys only
   +0.08 of a daily correlation nobody consumes. Revisit only if the daily
   texture of the pre-2025 EUR tail ever starts to matter.
4. **A deeper REAL donor, which is now the only way any of these files grows.**
   Length no longer comes from the engine, so the whole question is whether a
   1980s or early-1990s managed-futures NAV series exists that survives the
   grading applied to the current donors (a real programme, a sane drawdown, no
   restructuring artefacts, a documented fee load). The rejected candidates are
   listed in the survey; the SEC EDGAR route is the one that was never fully
   walked.

Two entries retired here rather than being solved: the pure-trend record the
overlays needed was found (survey above), and the unanchored fortnight at the
start of the engine tail went with the tail.

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
- The engine is still in the shipped files even though its own tail is not:
  it is the daily texture the weekly-dealing 1996-2007 donor is projected onto.
  A change that seems to touch only a level therefore still moves
  daily-frequency statistics over 1996-2007, at the third decimal, and that is
  enough to break a literal frozen to two decimals.
- Two FIRE-book plates recompute the CTA series from `pkg/datasets` and fail on
  drift, at a tolerance (0.005 around a two-decimal literal) finer than the
  series' own stability. They read 2001 onward, so a change confined to the
  pre-1996 tail or to the overlays leaves them alone: the 2026-08 truncation
  and anchor swap moved no plate cell. A change to the donor era or to the
  texture will, and then expect a handful of cells to need refreezing.
- Fee structures are the largest silent lever in this file. A performance fee
  shows up as a wedge that GROWS with the fund's own return and vanishes in
  down months (the RAEF/B EUR tables above); a composite "net of fees" means
  net of its constituents' 2-and-20, so a flat-fee replicator should sit
  ABOVE it, not on it; and correcting a fee twice (a drag on an already-net
  anchor, a wedge on an already-lagging donor) is the single easiest way to
  quietly lose several points a year. Every fee decision here carries its
  measurement next to it; keep that discipline.
  The rule for a donor segment is fixed and lives in `feeAligned`: lift it by
  the donor's management-and-expense load minus the target's, floored at zero,
  both read off published fee tables (`trendFeeLoad`) and never off an observed
  return gap, which would bake a manager's alpha into a "fee". Performance fees
  move the number in one direction only: the donor's are ignored, the target's
  are subtracted (`trendPerfFee`), so the uplift errs small. Applied today:
  ASFYX +0.60/+0.70/+0.55 into DBMF/the UCITS classes and CTA/KMLM, RYMFX
  +1.14/+1.24/+1.09, AHLPX +1.16 into CTA, DBMF +0.10 into its UCITS classes,
  Man AHL +1.84 to +1.99 everywhere except the AQR chain (+0.37), and 0 for
  AQMIX into AQR UCITS A, whose own performance fee already covers the gap.
  The RAEF B EUR donor is corrected by its own measured wedge and takes no
  uplift, or it would be corrected twice.
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
