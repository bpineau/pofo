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

## Not done, and why

- The forward nowcast uses Xetra-free daily closes of URTH, so no attempt was
  made to correct Xetra ticks for the US afternoon; a European-hours estimate
  would need the ETFs' official NAVs (DWS publishes them, no public feed was
  found) or an intraday model of the US afternoon, both more machinery than
  the half-day of latency they would buy.
- The other share classes (H, P) are the same portfolio at other charges and
  are not catalogued; the generator and the source are generic over the
  catalog, so adding one is a record.
