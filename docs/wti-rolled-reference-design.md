# A rolled WTI crude reference (WTI-ER-USD)

Dated record of the 2026-08-21 hunt, the method that was chosen, the validation
it passed and the bounds it ships under. The generator's godoc
(`go doc ./cmd/gen-wti-refdata`) is the operational description; this file keeps
the evidence and the roads that turned out to be closed.

## The gap this fills

The repo carried WTI **spot** only: `WTI-USD` (monthly, 1946 onwards, FRED
WTISPLC) and `WTI-DAILY` (daily, 1986-2000, a shape series whose levels are
explicitly not authoritative). A spot price is not investable. What a futures
holder earns is the spot move plus the roll yield, and for crude that term has
been worth far more than the spot move itself:

| Window | Rolled excess return | Spot | Roll yield |
|---|---|---|---|
| 1986-01 to 2000-12 | +9.80 %/yr | +0.30 %/yr | **+9.51 pt/yr** |
| 2000-12 to 2005-12 | +21.14 %/yr | +17.97 %/yr | +3.17 pt/yr |
| 2005-12 to 2016-12 | -14.00 %/yr | -1.15 %/yr | **-12.84 pt/yr** |
| 2016-12 to 2024-04 | +1.09 %/yr | +6.97 %/yr | -5.88 pt/yr |
| 1986-01 to 2024-04 | +2.06 %/yr | +3.27 %/yr | -1.21 pt/yr |

Anything that prices an oil position off spot is therefore wrong by a
double-digit annual rate whose **sign flips by era**, which is the worst kind of
error: it flatters a backwardation backtest and punishes a contango one.

## What is bundled

`pkg/datasets/refdata/WTI-ER-USD.csv`, daily, 1985-01-02 to 2024-04-05, 9875
points, base 100.

It is an **excess return**: the futures position alone, with no collateral
interest, on the same convention as `TREND-TSMOM-USD` against `TREND-NET-USD`.
Fund it with `TBILL-3M` to get a total return; the golden test
(`TestGoldenWTIRolled`) shows how, in about twenty lines.

Method: a long first-nearby NYMEX WTI contract, rolled into the second nearby
over the fifth through ninth business day of each month, a fifth of the position
per day. That is the S&P GSCI roll schedule, and for a single-commodity index
the GSCI contract daily return reduces to the return of that basket.

Source: EIA's daily NYMEX settlement prices for the first and second nearby
contracts (series `RCLC1`, `RCLC2`), from EIA's own key-free bulk archive
`https://api.eia.gov/bulk/PET.zip`.

### The two mechanical traps

1. **EIA's "contract 1" is a slot, not a contract.** It renumbers to the next
   delivery month on the first trading day after the expiring contract's last
   trading day. On that day the position, already wholly in slot 2, becomes
   wholly slot 1, so the return is `P1(t)/P2(t-1) - 1`. Using `P1(t)/P1(t-1)`
   would book the entire calendar spread as a price move, once a month, forever.
   The dates come from the CME rule (trading terminates three business days
   before the 25th calendar day of the month preceding delivery, or before the
   last business day preceding the 25th when the 25th is not one), computed over
   the exchange calendar the data itself carries, so no holiday table is needed.

   April 2020 settles the convention beyond argument: EIA's slot 1 reads -37.63
   on 2020-04-20 and 10.01 on the 21st, the expiring May contract's last two
   settlements, and only becomes the June contract on the 22nd.

2. **Where the carry is realized.** The roll days are value-neutral: the
   position exchanges one contract for an equal *value* of the next, which is
   why the daily return prices both legs at the weights carried into the day.
   The carry accrues instead as the held contract converges, and on a curve
   pinned in place it lands entirely on the renumbering day. This looks
   surprising and is correct; `TestBuildPricesTheRollYield` pins it.

Because the roll completes around the 13th, the position is already in the next
contract when the front month expires around the 21st. That is why this index,
like the real S&P GSCI Crude Oil index, never touches the negative May-2020
settlement.

## Validation

Against the published **S&P GSCI Crude Oil Total Return Index** year-end levels,
1986-2015, which are a matter of SEC record: Barclays Bank PLC, iPath S&P GSCI
Crude Oil ETN pricing supplement, form 424B2 filed 2016-11-17, accession
`0001104659-16-157831`, section "Historical and Hypothetical Historical
Performance of the Index". The sponsor began calculating the index on
1991-05-01; earlier levels are the sponsor's own backtest.

That series is **funded**, so the check compounds the reconstruction with the
bundled `TBILL-3M` (a 91-day discount rate, accruing every calendar day, the
GSCI's own convention) before comparing calendar-year returns.

| Year | Published | Rebuilt | Divergence |
|---|---|---|---|
| 1987 | +11.26 % | +14.25 % | +2.69 |
| 1988 | +15.82 % | +13.95 % | -1.62 |
| 1989 | +94.27 % | +94.35 % | +0.04 |
| 1990 | +46.49 % | +70.67 % | **+16.51** |
| 1991 | -16.57 % | -12.82 % | +4.50 |
| 1992 | +3.65 % | +3.50 % | -0.15 |
| 1993 | -35.25 % | -34.47 % | +1.20 |
| 1994 | +38.81 % | +30.27 % | **-6.15** |
| 1995 | +33.61 % | +29.96 % | -2.73 |
| 1996 | +108.31 % | +106.93 % | -0.67 |
| 1997 | -31.06 % | -29.48 % | +2.29 |
| 1998 | -47.61 % | -48.59 % | -1.87 |
| 1999 | +122.69 % | +127.69 % | +2.24 |
| 2000 | +50.69 % | +47.70 % | -1.98 |
| 2001 | -25.44 % | -25.23 % | +0.28 |
| 2002 | +58.30 % | +62.68 % | +2.76 |
| 2003 | +27.47 % | +27.20 % | -0.21 |
| 2004 | +44.79 % | +48.11 % | +2.29 |
| 2005 | +21.19 % | +26.77 % | +4.60 |
| 2006 | -16.95 % | -13.38 % | +4.30 |
| 2007 | +47.45 % | +47.23 % | -0.15 |
| 2008 | -55.47 % | -53.84 % | +3.66 |
| 2009 | +7.15 % | +9.13 % | +1.85 |
| 2010 | -0.11 % | +0.01 % | +0.12 |
| 2011 | -1.31 % | +0.14 % | +1.47 |
| 2012 | -11.52 % | -11.61 % | -0.10 |
| 2013 | +6.01 % | +7.16 % | +1.08 |
| 2014 | -42.56 % | -42.14 % | +0.72 |
| 2015 | -45.34 % | -45.42 % | -0.16 |

Divergences are relative, `(1+rebuilt)/(1+published)-1`, in points.

27 of the 29 years land within five points, the median divergence is 1.85
points and the cumulative divergence is +1.20 %/yr over the 29 years, most of it
1990 alone. 2009 is the year that matters most for what the series is for: the
published index returned +7.15 % while spot rose 78 %, and the reconstruction
says +9.13 %. The contango bleed is priced.

**1990 (+16.5 points) is not explained.** It is the Gulf War, when the
front-to-second spread ran into double digits a month and any roll-timing
difference is maximally levered; 1994 (-6.2) is the other outlier. Both sit in
the era before the index sponsor published anything, so the published numbers
there are themselves a backtest. The generator gates on the aggregate (25 of 29
years within five points, median under three, cumulative under 2 %/yr) rather
than on any single year, so an ordinary source revision passes and a broken
download or a regressed roll does not.

Roll conventions were swept against the same anchor before settling. Rolling on
the weights carried into the day, over business days 5-9, was the best of six
combinations tested (S&P's published formula applies the same day's roll weights
to both legs, placing the roll a day earlier; that variant drifts +1.68 %/yr
instead of +1.20 %, and the BD6-10 Bloomberg window is worse again).

**This is a reconstruction that behaves like the published index, not a copy of
one, and it is not labelled as one.**

## Reliability bounds length

The file ends 2024-04-05 because **EIA discontinued `RCLC1` to `RCLC4` after
that day**; it still publishes the WTI spot price (`RWTC`) daily, so the freeze
is a decision at the source, not a broken fetch. No engine tail is shipped in
front of it. The same rule the trend reconstructions live under.

## Sources probed, and what answered

| Source | Result |
|---|---|
| **EIA bulk `api.eia.gov/bulk/PET.zip`** | **answered**: key-free, 52 MB, daily `RCLC1`-`RCLC4` 1983/1985 to 2024-04-05. The series bundled here. |
| **SEC EDGAR** (archives + `efts.sec.gov` full-text search) | **answered**: the published S&P GSCI Crude Oil TR year-end table, 1986-2015, in ETN pricing supplements. The validation anchor. |
| AQR public data library, "Commodities for the Long Run" | reachable, but the free file is the **equal-weight aggregate** portfolio only (monthly, 1877 onwards): no per-commodity crude series. |
| DBnomics `EIA/PET` | reachable and mirrors the same series, but its copy stops at the same 2024-04-05 and its scraper cannot notice a resumption. Not used, so the generator reads EIA directly. |
| Yahoo `^SPGSCLP`, `^BCOMCL`, `CL=F`, individual contracts (`CLZ26.NYM`) | **walled**: HTTP 429 on both `query1` and `query2` throughout the session, with and without a cookie jar. The archive for `^SPGSCLP` was already known to serve a single point at every range. |
| CME settlement API (`cmegroup.com/CmeWS/mvc/Settlements/...`) | **walled**: "This IP address is blocked due to suspected web scraping activity", for recent and historical trade dates alike. |
| S&P Dow Jones (`spglobal.com/spdji`) | **walled**: HTTP 403 on the index page, so its performance widgets were never reachable. |
| Stooq | known JavaScript/proof-of-work wall; not retried. |
| WSJ michelangelo timeseries API | known dead with the old public token; not retried once EIA answered. |

The two published index families that would carry the series past 2024
(S&P GSCI Crude Oil ER, base 1987; Bloomberg Crude Oil Subindex ER, base 1991)
have **no free historical download**. If that ever changes, the right move is to
splice the real index in front of nothing and keep this reconstruction as the
deep tail, not to graft an engine onto it.

## Open

- **The tail, 2024-04 to today.** Rebuildable in principle from individual
  contract histories (about thirty `CL` monthly contracts), which needs a
  futures data source that answers; none did in this session.
- **1990 and 1994.** Named above, unresolved, and gated around rather than
  tuned away.
- **`pkg/simgen`'s DBi replica prices its crude leg off `CL=F`.** Measured
  against this series on 2026-08-21 and left untouched: the two correlate 0.957
  daily over 5913 shared days, and the difference is not the slow drift the
  projection's discarded intercept would absorb (removing each 60-day window's
  own mean shrinks the residual by -0.5 %). It sits on the days the continuous
  series switches contract, RMS 0.0258 there against 0.0056 elsewhere, with 24
  of the 40 largest differences on the 4.8 % of days that are switch days and
  +23.6 % on 2008-12-22 alone. `CL=F` books the calendar spread as a return once
  a month, which is a monthly phantom in a regression leg rather than a level
  error. Repricing it still fails the gate, because this series stops at
  2024-04-05 while the blend is graded against the fund through 2026; the full
  record is in `docs/trend-reconstruction-design.md`.
