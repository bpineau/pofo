# The catastrophe bond sleeve

What is bundled for insurance-linked securities (ILS), where the data comes
from, what a 5 to 10 % sleeve was measured to do to a euro decumulation book,
and what the measurement cannot say. Read this before touching
`pkg/simgen/catbond.go` or `cmd/gen-catbond-refdata`.

## What the asset is

A catastrophe bond is a floating-rate note whose principal is forfeited when a
named natural event exceeds a defined trigger. The coupon is the money-market
rate plus an insurance premium; the collateral sits in short-term instruments.
Two consequences run through everything below:

- the return driver is the weather and the insurance cycle, not the economic
  cycle, so the correlation to every other sleeve is genuinely low and the low
  correlation is CAUSAL rather than a statistical accident;
- the loss arrives as a single step when a season goes wrong, and it does not
  mean-revert. The asset has no equivalent of "the market recovers": a burnt
  bond is gone, and what recovers is the next vintage's wider spread.

It is therefore a diversifier and a carry asset, and it is NOT a crisis hedge.
The catalog maps `asset_class: "insurance-linked"` onto the Crisis quadrant
because that is the closest label the four macro regimes offer for "pays a
premium that owes nothing to this quadrant", not as a claim of protection.

## The bundled series

| id | what | currency | cadence | from |
|---|---|---|---|---|
| `ILS-NET-USD` (refdata) | ILS Advisers Fund Index monthly returns | USD | month-end | 2006-01 |
| `ILSFUND` | that index served as a benchmark, unrescaled | USD | month-end | 2005-12 |
| `ILSFUNDE` | the same hedged to EUR | EUR | month-end returns, daily cash legs | 2005-12 |
| `LI0049587301` | Solidum Cat Bond R EUR hedged, backcast | EUR | semi-monthly | 2005-12 |
| `IE00B3Q8M574` | GAM Star Cat Bond EUR hedged, backcast | EUR | weekly | 2005-12 |
| `LI0115208543` | Plenum CAT Bond Defensive R EUR, backcast | EUR | weekly | 2005-12 |
| `LU0951570927` | Schroder GAIA Cat Bond IF EUR hedged, real only | EUR | weekly | 2013-10 |
| `IE000UWJUW87` | KRC Cat Bond UCITS ETF, real only | EUR line | daily | 2025-12 |

### Why this index and not the market's own

The market's reference is the Swiss Re Global Cat Bond Total Return Index
(Bloomberg `SRGLTRR`), weekly since 2002. It is not bundled, for two reasons
that both stand on their own. It is not publicly downloadable: the publisher's
data explorer is bot-gated and only annual figures appear in the methodology
paper. And it is a MARKET index, gross of every fund fee, so standing it in
front of a fund would need a fee estimate on top of a currency hedge.

The ILS Advisers Fund Index (formerly the Eurekahedge ILS Advisers Index) is
the other kind of object, and the useful one here: an equally weighted
composite of the REAL funds investing in non-life ILS, each constituent at
least 70 % non-life risk, each entering NET of its own manager's fees. It is
what a book of real ILS funds actually paid its investors, which is exactly the
relationship `TREND-NET-USD` (BTOP50) has to the managed-futures family, and it
is the right anchor for reconstructing a cat bond UCITS class.

It is served as JSON behind the publisher's own chart, which is what
`cmd/gen-catbond-refdata` reads.

### How the download is validated

The source publishes its summary statistics next to the table it serves, so the
generator recomputes every one of them and refuses to write on a disagreement.
Measured 2026-08-17:

| statistic | recomputed from the table | published |
|---|---|---|
| maximum drawdown | -12.52 % | -12.50 % |
| worst month | -8.61 % | -8.61 % |
| best month | +2.26 % | +2.26 % |
| annualized standard deviation | 3.37 % | 3.47 % |
| share of positive months | 87.45 % | 87.04 % |
| last 3 months | +2.32 % | +2.32 % |
| last 3 years annualized | +11.84 % | +11.84 % |
| last 5 years annualized | +8.07 % | +8.06 % |
| **since inception annualized** | **+4.26 %** | **+5.09 %** |

Eight figures out of nine reconcile, three of them to the last decimal on
returns computed from the same months. The ninth does not, and it is treated as
a stale headline rather than as evidence against the table: the generator logs
the divergence on every run instead of hiding it, and the shipped series
compounds to 4.26 %/yr.

A second, fully independent check lives in `pkg/datasets/golden/catbond_test.go`:
the four deepest months of the shipped series must be the months of named
events, in order (Harvey/Irma/Maria 2017-09, Ian 2022-09, the California
wildfires and Michael 2018-11, the Lehman collateral shock 2008-10), and the
calendar years 2006-2013 must sit BELOW the Swiss Re market index's published
years by 0.5 to 8 points, further below in 2008. Those eight Swiss Re figures
are quoted from its methodology paper as an external anchor; no Swiss Re series
is fetched or bundled.

### Reliability bounds length

Nothing reaches before 2006-01. Cat bonds have traded since 1997 and the market
index starts in 2002, but no public monthly record of what an ILS FUND paid its
investors reaches further, and 2002-2005 would have to be spliced onto an
estimate. Hurricane Katrina (2005) therefore sits outside every file here, and
so does the 1990s market. The 2008 collateral shock is inside, which is the
single most valuable stretch the extension buys: it is the only period on
record where cat bond funds fell alongside everything else.

### Cadence, and the trap it sets

The index is monthly and nothing invents a daily texture for it: no ILS series
quotes daily anywhere, and a single day of cat bond return is not a path a
reader could look up. The funds themselves deal weekly (Solidum semi-monthly).

`pkg/metrics` annualizes by 252 observations, so **every per-observation
statistic on these lines is wrong by roughly sqrt(5)** and the monthly columns
are the ones to read. The GAM class reports 7.0 %/yr "volatility" and 3.0 %/yr
monthly volatility; the second is the truth. The doctor (`Verify`) is already
cadence-aware and judges them at their own pace.

### The euro hedge, and what it costs

Every cat bond class a European household can hold is EUR-hedged, and hedging a
funded total return means swapping its dollar cash leg for the euro one. The
identity is unusually exact for this asset (the collateral really does earn the
money-market rate), and `hedgeToEUR` applies it day by day, shared with BTOP50E.

The swap is not free, and its sign changes with the rate differential:

| window | index in USD | hedged to EUR | give-up |
|---|---|---|---|
| 2006-01 to 2026-06 | 4.26 %/yr | 3.58 %/yr | 0.68 pt |
| 2006-01 to 2011-09 | 4.55 %/yr | 5.06 %/yr | **-0.51 pt (it paid)** |
| 2011-10 to 2026-06 | 4.12 %/yr | 3.05 %/yr | 1.07 pt |

The model is optimistic by roughly a third of a point: measured on four fund
families' own USD and EUR-hedged classes over their live windows, the realized
give-up is 1.3 to 1.6 points a year, against the 1.07 the clean identity
produces over the same era. The difference is forward-roll friction, imperfect
hedge ratios and the basis between a three-month bill (`^IRX`) and an overnight
euro rate. Treat every EUR figure here as a mild overstatement.

The unhedged alternative is not a way out. Held in USD by a euro investor, the
same fund's volatility roughly triples (about 3 % becomes about 8 %) because
EURUSD supplies most of the variation, which destroys the only property the
sleeve was bought for.

### How the fund files are built

`catBondBuild` takes the EUR-hedged index, lifts it to the target class's own
ongoing charge and rescales it to the target's OWN monthly volatility, then the
recipe splices the fund's real NAVs on top.

The volatility match is what separates a fund file from the benchmark: these
classes are not the same risk, a defensive mandate holding the low-expected-loss
end of the market realizing about four fifths of the whole market's volatility,
and splicing the market's history in front of it unchanged would credit it with
losses it was built not to take. It is measured on MONTH-END returns, because
`volMatch` cannot serve here: a monthly donor and a weekly fund share almost no
observation dates, so a per-observation match would find nothing to measure and
skip the donor in silence.

The fee alignment reads published ongoing charges only, never an observed return
gap, exactly as `feeAligned` does for the trend family. The index's own load is
the single ESTIMATE of the table (1.50 %/yr: an index of funds levies nothing
itself, but every return in it arrives net of a constituent's charge, and those
constituents publish no schedule). Since the retail classes here run 1.20 to
1.75 %, the resulting uplift is small by construction.

Measured against the funds' own live windows:

| class | real from | backcast CAGR | fund CAGR | gap |
|---|---|---|---|---|
| Solidum EUR hedged | 2009-09 | 3.43 % | 3.52 % | -0.09 pt |
| Plenum Defensive | 2010-09 | 2.37 % | 1.88 % | +0.49 pt |
| GAM Star EUR hedged | 2011-10 | 2.53 % | 4.29 % | -1.76 pt |

The audit harness (`pofo -verify-simdata`) grades them on the same windows:
Solidum level ok / path warn (monthly correlation 0.83), GAM level warn / path
warn (0.84), Plenum Defensive level ok / path bad (0.66, its defensive mandate
diverging from the whole market it is reconstructed from). Read that report's
VOLATILITY and tracking-error columns with the cadence trap in mind: it
annualizes per observation, so the engine (monthly steps spread over a daily
calendar, which annualizes correctly) is compared against a weekly fund NAV
whose figure is inflated by about sqrt(5). The monthly correlations and the
CAGR gaps are the columns to trust here.

The GAM gap is manager selection, not a modelling error, and it is deliberately
NOT closed: the class beat the index of its peers by about 1.8 points a year
over its own window, and granting that to fifteen years it did not live through
would be curve fitting. The backcast carries the index's return, not the
manager's, which is the conservative direction.

## What a 5 to 10 % sleeve does to a decumulation book

Measured on `examples/dragon-decumulation-household.txt` with `sim:on`, in EUR,
rebalanced every 90 days, over **2005-12-31 to 2026-08-06** (the deepest window
the file's own lines and this sleeve share), with the Solidum class as the
sleeve. Two funding rules: pro rata from every line, and out of the bond block
(the 4 % long-duration line plus the 13 % linkers line).

| | base | 5 % pro rata | 10 % pro rata | 5 % from bonds | 10 % from bonds |
|---|---|---|---|---|---|
| CAGR | 8.15 % | 7.85 % | 7.64 % | 8.14 % | 8.23 % |
| volatility (monthly) | 6.82 % | 6.53 % | 6.22 % | 6.77 % | 6.71 % |
| max drawdown | -17.92 % | -17.42 % | -16.58 % | -17.91 % | -17.56 % |
| Sharpe (monthly) | 1.17 | 1.19 | 1.21 | 1.19 | 1.21 |
| worst rolling 5y | 1.13 % | 1.02 % | 1.02 % | 1.07 % | 1.11 % |
| ongoing charges | 0.43 % | 0.47 % | 0.51 % | 0.48 % | 0.54 % |

The weight sweep says the same thing without a corner: every point of cat bond
funded pro rata costs about 0.04 point of CAGR and buys about 0.08 point of
volatility and 0.17 point of drawdown, monotonically, with no interior optimum
anywhere between 0 and 45 %. It is a de-risking line, not a return engine.

The sleeve's own statistics on that window: CAGR 3.80 %, monthly volatility
3.18 %, worst monthly drawdown -8.48 %. Its monthly correlations to the book's
other lines are the reason to look at it at all: 0.23 to the efficient core,
0.22 to world equity, 0.24 to the small-value sleeve, 0.26 to the linkers, 0.10
to long duration, 0.09 to gold, and -0.13 / -0.12 to the two trend lines.

### Under the FIRE model

Monthly stationary bootstrap (mean block 24 months) over the same panel, real
terms, 40 years, 20 000 paths, rigid real withdrawal:

| | ruin at 4 % | safe rate at 5 % ruin | ruin at 5.5 % | 5th-pct worst 10y |
|---|---|---|---|---|
| base | 0.01 % | 5.74 % | 2.88 % | -5.2 % |
| 5 % pro rata | 0.01 % | 5.58 % | 4.08 % | -12.0 % |
| 10 % pro rata | 0.01 % | 5.43 % | 6.01 % | ruin |
| 5 % from bonds | 0.01 % | 5.80 % | 2.44 % | -4.4 % |
| 10 % from bonds | 0.01 % | 5.86 % | 2.08 % | -4.0 % |

### Reading it honestly

Funded PRO RATA, the sleeve is a small, reliable loss of expected return bought
for a small, reliable reduction in risk. Every risk-adjusted ratio improves and
every level statistic worsens; under a fixed real withdrawal the level is what
matters, so the FIRE model reads it as a mild negative.

Funded OUT OF THE BOND BLOCK, it is roughly neutral to mildly positive, and
that verdict is mostly a verdict on euro bonds over 2006-2026 rather than on cat
bonds. It also removes the two things the bond block was bought for and no
backtest of this window prices: the contractual link to euro inflation, and the
long-duration convexity that answers a 2008-style deflation. Reading the swap as
an upgrade means accepting that trade knowingly.

## What the measurement cannot say

- **Twenty years hold few big cat years.** The record contains one very bad
  season (2017), two bad ones (2018, 2022) and one collateral shock (2008). The
  distribution's left tail is defined by events rarer than that.
- **The tail is concentrated.** About two thirds of the market's expected loss
  is US hurricane and US quake. This is one risk factor with a season, not a
  diversified book, and a Miami landfall is a plausible single event that takes
  a fifth of a whole ILS fund.
- **The euro investor's drawdown is deeper than the asset's.** The EUR-hedged
  index fell 18.75 % peak to trough between 2017-07 and 2022-10, against the
  12.5 % the USD index shows: six thin years with the hedge give-up running
  underneath them. A spending household must be able to sit through a five-year
  underwater stretch in a line it bought for stability.
- **Access is the binding constraint.** Of the classes measured here, the
  cheapest and deepest are Liechtenstein and Irish share classes bought at NAV
  through a platform, not instruments a normal broker lists. The only readily
  brokered vehicle is the KRC Cat Bond UCITS ETF (`IE000UWJUW87`, 1.28 %),
  which is USD UNHEDGED and launched in 2025-12, so it has neither the hedge nor
  a record.
- **Retail fees eat a large share of a small premium.** The Plenum defensive
  class charges 1.75 % against a sleeve whose whole excess over cash is 3 to 4
  points, and it returned 1.88 %/yr over fifteen years. Inside this asset class
  the share class choice is not a detail.
- **The window is kind to the funding argument.** 2006-2026 is a period in
  which euro sovereign bonds returned very little; a window containing a real
  bond bull would reverse the bond-funded verdict.

## If this is revisited

Two things would move the answer more than anything else here: a Swiss Re index
history back to 2002 (Katrina, and a market-level view free of fund fees), and
an EUR-hedged share class cheap enough and brokered widely enough to make the
sleeve buyable at a fee that leaves the premium intact.
