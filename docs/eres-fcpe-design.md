# ERESMONDEM: an employee-savings fund served from its own NAV, nowcast in between

Decided and measured 2026-08-24. Governs `pkg/marketdata/airfund.go`,
`pkg/marketdata/nowcast.go`, the `ERESMONDEM` recipe in
`pkg/simgen/recipes.go`, `cmd/gen-eres-refdata` and the catalog record.

## The object

"ERES XTRACKERS ACTIONS MONDE, Part M" is a French employee-savings fund
(FCPE) offered inside Eres PEE/PER plans, launched 2024-03-05 at 50.00 EUR.
It holds permanently 75 % Xtrackers MSCI World Swap UCITS ETF 1D
(LU2263803533, synthetic, TER 0.19 %) and 25 % Xtrackers MSCI World UCITS ETF
1D (IE00BK1PV551, physical, TER 0.12 %), reinvests their distributions
(capitalisation pure) and hedges nothing (the KID allows 0-110 % of non-euro
currency). It has NO ISIN: the company's share code 990000135629 stands in
for one (sicavonline's QS0009135623 is a pseudo-ISIN), the newer A/AM classes
(FR00140148R2, FR0014014UB1) are the only ones admitted to Euroclear and
publish no NAV. No quote site covers it.

Charges, per the FY2025 annual report's fee table (Part M column): 0.35 %
management (the maximum, charged in full), 0.21 % induced by the two ETFs,
0.06 % transaction costs, 0.62 % all-in. The KID states 0.56 % recurring plus
0.06 % transactions. The 0.47 % that circulates is Part H's total, not M's.
Entry fee up to 5 % in the KID, waived by some plans.

## The feed

The fund page renders its chart and its "Exporter les VLs" button from
web components served by airfund.io, and the delivery API behind them is
public: no login, cookie or key, one POST.

    POST https://core.communicate.airfund.io/api/v1/navs-evolution-chart/data
    {"locale":"fr","sId":"<widget id>","isinCode":"990000135629",
     "maxPeriodCode":"inception","debug":null,"displayBenchmark":false}
    -> 201 {"fundName":"...","navs":[{"date":"2024-03-05","value":50},...]}

`sId` is the id of the chart widget embedded on the page and is REQUIRED
(the API answers 500 without it); it lives in the catalog record's `xid`,
the share code in `symbol`, `source: "airfund"`. The page itself sits behind
a Cloudflare challenge, the API does not. The series matches the page's CSV
export to the cent (612 NAVs on 2026-08-24, 50.00 to 70.79). The client
accepts any 2xx since this API answers a POST with 201.

`cmd/gen-eres-refdata` (`make eres-refdata`, part of `make refresh`) writes
`refdata/ERESMONDEM-NAV.csv` from the same call after validating it (first
NAV at the catalog inception, ordered dates, positive, no 15 % daily move,
not older than 45 days). That file is the offline fallback of the live
source (`embeddedNAV`, reached through the same stale-fallback path as the
bundled CPI and FX snapshots) and the real series the recipe splices at
generation time, so an offline run and the shipped reconstruction agree.

## The clock: what the NAV of day D actually prices

Measured on the 612 NAVs against daily references, correlation of the fund's
daily return with the reference's return shifted by k trading days:

| reference | k = -1 | k = 0 | k = +1 |
|---|---|---|---|
| URTH (NY-listed MSCI World, converted to EUR) | 0.03 | **0.875** | 0.05 |
| XDWD.DE (Xetra, EUR) | 0.32 | 0.62 | 0.21 |
| IWDA.L (LSE, converted) | 0.36 | 0.56 | 0.19 |
| LU2263803533 via FT (NAV) | 0.37 | 0.54 | 0.20 |

The fund strikes its NAV on the two ETFs' official NAVs, which value every
market at its own close of the day, New York included. On 2025-04-09 (the
tariff-pause rally, +9.5 % on the S&P after 19:20 CET) the fund printed
+5.28 %, the Xetra line -4.73 % (closed before the rally), URTH +9.22 % (a
US ETF's price at 22:00 CET, premium included), the index in EUR about
+5.5 %. Only a reference struck after New York closes shares the fund's
calendar, and no Xetra or LSE line does. The FX convention could not be
pinned (Yahoo's daily cross shifted by a day correlates 0.853 against 0.868
unshifted); the standard conversion is kept.

Consequences: URTH is the nowcast proxy (below); the daily correlation of
the recipe, whose donor years are Xetra closes, reads 0.62 by construction
while its weekly correlation reads 0.95 and its monthly 0.98; and the level
of any Xetra-based comparison wobbles by up to half a percent around the
truth from one year-end to the next.

## The recipe

Daily-rebalanced 75/25 blend of the two ETFs' total-return paths, each leg a
donor chain nearest first, every donor lifted to the class's TER by the fee
difference and never to close a gap:

- swap leg: XWD1.DE (the fund's own 1D class, 2021-03) <- DBXW.DE (the 1C
  sibling, same swap, 0.45 % TER against 0.19: +0.26 %/yr credited, 2008-01)
  <- the MSCI World net-TR-in-EUR path of `wpeaBuild` (real IWDA from 2009,
  MSCIWORLD-USD refdata + daily index shape before, EURUSD spot to 1971),
  lifted from IWDA's 0.20 % to 0.19;
- physical leg: XDWL.DE (1D, 2015-04) <- XDWD.DE (1C, same 0.12 % TER,
  2014-08) <- the same world path lifted to 0.12;
- less the wrapper charge 0.41 %/yr (0.35 management + 0.06 transactions:
  what the fund bears that its ETFs do not; the 0.21 induced cost is already
  inside the ETF NAVs);
- the real NAVs grafted on top from 2024-03-05, so the level is the fund's.

Each donor is optional: a leg that cannot read one reads the next series
behind it, which is what keeps the recipe building offline (the offline
universe carries synthetic stand-ins for the four classes) and what would
keep it building should a Xetra listing die.

Validation over the real window (2024-03-05 to 2026-08-20, 2.46 years),
`pofo -verify-simdata ERESMONDEM`: level ok, engine CAGR 15.50 % against
15.19 % real, +0.31 pt/yr; path warn on the daily clock above. The residual
was NOT tuned away. Measured against the two 1C Xetra classes the fund
lagged by 0.53 %/yr, against the 1D classes' true total return by ~0.75 to
0.8 %/yr, per calendar year +0.13 / -1.77 / +0.32 points (2024 partial, 2025,
2026 partial), which is not the signature of a steady fee: the 2025 figure
is a slow bleed plus the year-end timing wobble, and a day-by-day trace of
January 2025 shows noise of +-0.9 % around it, no step. The a-priori charge
stays; the audit caveat records the measurement.

Two limits are stated in the recipe's godoc: the half-session timing smear
of the donor years, and the swap leg's edge over the net index (XWD1.DE
outran XDWD.DE by ~0.35 %/yr over 2024-2026) which the pre-2008 tail does
not model, so the deep tail is a touch conservative.

## The nowcast

The NAV of day D is published around D+2, and a portfolio tracker wants a
value now. The catalog names `nowcast_proxy: "URTH"` and `nowcast.go` reads
it in three places, all estimates, all flagged, none stored:

- `Fetch` extends the daily series from the last NAV to the proxy's last
  close, each day carrying the proxy's return converted into EUR through the
  ordinary daily conversion; `Series.EstimatedFrom` marks the first estimated
  day and `EstimateProxy` the proxy. The tail is added after the cache layer
  on a copy, so the disk cache and the memo hold published NAVs only. It ends
  where the proxy's cached history ends: a caller whose cache policy keeps
  URTH a week old gets a week-old nowcast, by design.
- `Intraday` returns today's path: the fund's last daily value BEFORE the
  session (a NAV or the forward estimate standing on the proxy's previous
  close) scaled by the proxy's intraday move, each tick converted at the
  intraday USD/EUR cross (`IntradaySeries.Estimate`, `Proxy`, `Source
  "nowcast"`). Anchoring on the previous close rather than the last close
  is what keeps a session from being counted twice once the proxy's daily
  close for the same day has landed.
- `Latest` quotes the last tick of that path (`Live` true, `Source
  "nowcast"`), or the last published NAV when the proxy cannot be read.

`Series.WithoutEstimates` is the one door every storing consumer walks
through: `simgen`'s fetcher (recipes, audit), the simdata generator's
splice, the refdata generator. Nothing bundled carries an estimate.

What the estimate ignores, stated: the wrapper charge (0.41 %/yr is below a
cent over the days involved), the proxy's own tracking of the fund (the 0.875
daily correlation above, i.e. ~0.4 % of daily residual, mostly URTH's
premium/discount and FX timing, mean-reverting rather than cumulative), and
European trading hours: the estimate moves when New York trades and holds
the previous close otherwise, which is honest about what the fund's own
clock knows.

## The second fund: ERES_DATADOG (added 2026-08-25)

"Actions Datadog, Part C" (share code 990000124099) is the single-stock FCPE
of the Datadog France plan: 90-100 % DDOG class A shares, the rest cash, NAV
in EUR, unhedged, launched 2021-07-22 at 100.00, 0.61 %/yr charged in FY2025
(max 1.50 %), no entry, exit or transaction cost. Catalogued as
`ERES_DATADOG`, deliberately apart from the listed `DDOG` record, since a
household can hold both. Everything above applies unchanged (the widget id
is the Eres site's, not the fund's: the same `xid` serves every fund on the
site), with three differences worth a line each:

- WEEKLY, then daily. One NAV a week (Fridays) until 2026-07-13, when the
  fund switched to daily valuation (fund page, "changements intervenus");
  293 NAVs over five years, so every per-observation statistic over most of
  the line is off by ~sqrt(5) (the cat bond cadence trap); the
  audit's "real vol 115 %" is that artefact, the monthly correlation (0.96)
  and the level are what to read. The refdata generator's per-NAV move bound
  is a corruption detector (45 %) for this reason: the fund printed -24.6 %
  in the week of 2026-08-06 and that is real.
- THE CLOCK is the OPENING price, established two ways. Fitting the fund's
  returns to DDOG in EUR under each convention (Yahoo daily opens and closes,
  292 spans): open of the valuation day 2.00 % rmse, close of the day
  3.58 %, close of the previous day 3.48 %, high or low of the day 2.6 %; a
  scan of the fraction alpha of the open-to-close path bottoms at 1.77 % for
  alpha 0.2-0.3, which the second test resolves. On the 5-minute history
  (60 days, 34 NAVs, the daily era) the best-fitting instant is 09:30 New
  York at 2.45 % and every later time fits worse, monotonically, to 5.21 %
  at 16:00; the daily NAVs match the open to the tenth of a percent
  (2026-07-27: +2.47 % against +2.47 % at the open and +2.07 % at the close;
  2026-08-11: +10.32 / +10.46 / -5.49; 2026-08-06: -24.63 / -24.50 / -20.82).
  The FX bar alignment changes nothing (1.96 % with the next bar), and a
  price-date shift is excluded (NAV of X on the open of X-1: 5.2 % rmse,
  X-2: 7.4 %). The 2 % rmse itself is three regimes, not a tracking error,
  and the fund's FY2025 annual report names the first: its valuation rules
  (and the auditor's observation) state that the share is valued at the
  NASDAQ OPENING price from 2022-04-11, under an expert's method dated
  2021-04-06, and at the CLOSING price before. The refit agrees to the week:
  over 2021-07 to 2022-04-11 (37 spans) the close of the valuation day fits
  at 0.76 % rmse with a beta of 1.0 where the open gives 5.3 % and 0.74 (that
  low beta once read as "a fund still filling up" was the wrong clock, not a
  cash pocket); the weekly years 2022-04 to 2026-07 (225 spans) fit the open
  at 0.58 %, beta 1.00, the close at 3.6 %; the daily era (40 spans) the open
  at 0.69 %, MEDIAN 0.08 %, beta 0.99, 75 % of the days within 0.5 %. The
  fund IS the closing print in EUR before 2022-04-11 and the opening print
  since. The charge is visible once the regimes are kept apart: on the open,
  the 225 weekly spans cumulate -3.0 pt, about -0.7 %/yr against 0.61 %
  charged (an earlier reading of "no visible fee" mixed the two clocks). The
  same report closes the other candidates: no securities financing operation
  (SFTR) in the year and no financial income on the cash pocket, and its
  FY2025 tracking checks to the basis point on the fund's own clock (NAV
  -17.06 %, share at the open in EUR at the 16:00 London fixing -16.54 %,
  gap -0.62 pt; the report's own benchmark line, -16.07 %, is struck on
  another clock and is not the one to read). The
  NOWCAST IS ANCHORED THERE (added 2026-09-17): the record carries
  `nowcast_anchor: "open"` (default `close`, `ERESMONDEM` unchanged) and the
  estimate is the last NAV times the proxy's move since the proxy's OPENING
  print of that NAV's day, `DDOG` converted as before. Anchoring on the close
  instead carried the valuation day's open-to-close move as an offset until the
  next NAV: measured on the 264 NAV spans since 2022-04-11 (published NAVs,
  Yahoo bars, same-day cross), the one-span-ahead estimate lands at 3.81 %
  rmse anchored on the close against 2.78 % anchored on the open, and 4.52 %
  against 3.22 % over the 39 spans of the daily era. The same measurement
  reads the other way over the 37 pre-2022-04-11 spans (0.75 % on the close
  against 3.33 % on the open), which is the clock change again and costs
  nothing here: a nowcast only ever anchors on the LAST published NAV, whose
  era is the current one. The opening prices reach no further than a factor:
  `Client.fetchYahooOpenFactors` reads the chart API's `open` column next to
  the `close` and stores the OPEN-TO-CLOSE RATIO of each session as a series of
  its own (cached under the `SYMBOL~open` view, alongside `~raw`), so `Point`
  keeps carrying closes only and no cache file written before this existed
  changed meaning. A day the proxy did not trade, a proxy quoted by a source
  with no opening print and a failed fetch all fall back on the close anchor
  with a warning, never on an error, and an estimated forward day always
  anchors on the proxy's close of its own day since that is the print it was
  built from.
- THE RECIPE is the share itself: DDOG in EUR (no dividend, so price is total
  return) less 0.61 %/yr from the 2019-09-19 IPO, real NAVs grafted from
  2021-07-22, nothing before the IPO (a single stock has no donor). Measured
  over the real window (5.08 years): level ok, -0.22 pt/yr, the fund's cash
  pocket and the intraday clock accounting for the sign either way.

## Not done, and why

- The forward nowcast uses Xetra-free daily closes of URTH, so no attempt was
  made to correct Xetra ticks for the US afternoon; a European-hours estimate
  would need the ETFs' official NAVs (DWS publishes them, no public feed was
  found) or an intraday model of the US afternoon, both more machinery than
  the half-day of latency they would buy.
- The other share classes (H, P) are the same portfolio at other charges and
  are not catalogued; the generator and the source are generic over the
  catalog, so adding one is a record.
