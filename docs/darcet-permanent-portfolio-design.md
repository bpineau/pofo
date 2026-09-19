# Darcet's tactical Permanent Portfolio 2.0 - research notes & design

Status: shipped as `pkg/datasets/macropanel` + `pkg/permanent` + the `pofo
-permanent` CLI (backtest and decumul ruin bands). This document is the complete
record so the method can be refined, retuned, generalized (e.g. to the Artemis
Dragon) or rewritten from scratch without re-deriving anything.

## 0. Provenance

Didier Darcet (Gavekal / GaveTracks), video <https://www.youtube.com/watch?v=JRkJUznoszM>,
revisiting Harry Browne's Permanent Portfolio. Source material captured by
notebooklm (transcripts, ephemeral, were in `/tmp/darcet1.txt`, `darcet2.txt`,
`darcet-transcript.txt`). Everything below that is not attributed to Darcet is
our reconstruction or our empirical result.

## 1. Epistemic tags (READ THIS FIRST)

Cardinal rule of this document: never confuse what was *described in advance* (more
reliable) with what we *found by fitting data* (overfit risk). Every claim below
carries one tag:

- **[DARCET]** - asserted by Darcet in the video, i.e. specified before we
  touched any data. The most reliable layer, but note he does **not** disclose
  his exact formulas, thresholds or weights (he says so explicitly).
- **[RECON]** - our reconstruction of a gap Darcet left open (a weight function,
  poles, scales). Set **once** from his qualitative rules and **not** optimized
  against the data. Medium reliability: not tuned, but a judgement call.
- **[SELECTED]** - a specific parameter value we picked *after* seeing results
  (e.g. `wMax≈1.6` as the frontier sweet spot). **Highest overfit risk**; treat
  as illustrative, not validated.
- **[EMPIRICAL]** - a measured result from running on data. Reliability depends
  on whether it survived out-of-sample checks (see next tag).
- **[ROBUST]** - an empirical result that held across independent subperiods,
  start dates, cost levels and/or countries. The empirical claims we trust.

## 2. The method as Darcet describes it  [DARCET]

Harry Browne's original: 25% each of equities, long government bonds, cash
(T-bills), gold; rebalanced on a calendar; never re-weighted. **[DARCET]** claims
this earns a remarkably stable **inflation + 3-4% real**, an *invariant* across
~40 countries and 150 years, *except* when war is fought on home soil, which
destroys everything but gold (~-75%). The four assets are chosen as pure,
liquid, and structurally opposed: contracts (bonds, cash) vs property titles
(equities, gold); fiat money (cash) vs ancestral money (gold); private long-
duration (equities) vs public long-duration (bonds).

Darcet's "2.0" keeps the four assets but **tilts them tactically**, in two
independent blocks:

**Equity block** - driven by exactly two macro variables, **growth** and
**inflation**, forming four quadrants. **[DARCET]**:
- Paradise = growth accelerating + inflation decelerating (profit growth × multiple
  expansion): maximum equity.
- Hell = stagflation (growth down + inflation up): equities get massacred (-50 to
  -70%); flee.
- The world's position is measured by **breadth across ~40 countries**: the share
  whose growth is accelerating and the share whose inflation is accelerating give
  a single, slowly-moving "world point".
- Allocation falls with the **square of the distance** to the optimal pole
  ("distance doubles → allocation ÷4"): the 1/d² damping is presented as the
  guard against entropy/"avalanches" (markets collapse like sandpiles; crash
  amplitude ∝ 1/frequency, verified on the S&P since 1927). **[DARCET]**

**Defensive block** - for the non-equity sleeve, arbitrate gold / long bonds /
cash from the **monetary quadrant**, short rate × long rate. **[DARCET]**:
- Bonds when the curve is steep (10y − 3m > ~1%): duration is paid.
- Cash when the curve is flat/inverted but short rates are well remunerated.
- Gold ("juge de paix") when real rates are negative / currencies are debased.

Cadence: measured **monthly**, but reallocation happens in slow **waves of 5-6
years**; "no frenzy". Master variable to watch: the **oil price** (~$80-85 is the
danger threshold; above it, margins compress and inflation rises → both quadrants
turn). Philosophy: purely **reactive, never predictive** ("measure, adapt, flee").

**Not disclosed by Darcet** (the IP he keeps): the function mapping a distance to
a weight, the exact thresholds, and the final aggregation of the two blocks. So
any reproduction tests his *mechanism*, not his numbers. His only quoted
backtest: PP2.0 ≈ 10.3% CAGR nominal, maxDD 9.4%, Sharpe 0.75 over **2023-2025**
(too short to validate) plus the long-run invariant claims.

## 3. Data & reproducibility

All work is REAL (inflation removed), monthly, with the macro signal lagged one
month by REFERENCE date: month M is allocated on the regime of reference month
M-1 (`Simulate` takes the most recent regime dated strictly before the month it
pays). That rules out the contemporaneous reading, but it is not a lag by
PUBLICATION date, and the two differ here: OECD industrial production for
reference month M-1 is released about forty days after that month ends, so the
allocation of month M reads a number that did not exist when it was struck.
Section 5.8 measures what that is worth and section 9 keeps the caveat; the
honest information set is available as `SignalConfig.ReleaseLags`
(`permanent.PublicationLags()`), and every figure in this document that does not
say otherwise is computed at the reference-date lag.

Assets (via pofo `marketdata`, SIM suffix splices long history):
- Equity: `URTHSIM` (MSCI World TR, 1969→) for the global model; the OECD
  `SHARE` price index + flat dividend add-back for per-country.
- Long bonds: `TLTSIM` (1962→); per-country = synthetic from the long yield.
- Cash: from `^IRX` (13-week bill, 1960→) compounded; per-country from short rate.
- Gold: `XAUUSDSIM` (1968→), USD; converted to a currency by pofo FX.
- Deflator / inflation signal: `^CPI-US` (bundled) and per-country CPI.

Macro signals - **the key sourcing lesson**:
- **FRED is UNREACHABLE from the sandbox** (HTTP/2 INTERNAL_ERROR + timeout even
  on HTTP/1.1). Do not rely on it here.
- **DBnomics is reachable** and mirrors the OECD. The growth proxy is industrial
  production `OECD/DSD_STES@DF_INDSERV/<ISO>.M.PRVM.IX.BTE.Y._Z._Z.N` (1919→ for
  the US), the whole industry B-to-E aggregate.
- Interest rates, `OECD/DSD_STES@DF_FINMARK`: `IRLT` (long), `IR3TIB` (3-month,
  else `IRSTCI` immediate). Share prices: `SHARE`. CPI:
  `OECD/DSD_PRICES_COICOP2018@DF_PRICES_C2018_ALL/<ISO>.M.N.CPI.IX._T.N._Z`,
  falling back to the COICOP 1999 dataflow `OECD/DSD_PRICES@DF_PRICES_ALL` with
  the same key for the countries that have not migrated.
- These are the CURRENT dataflows; the panel read the legacy `OECD/MEI` codes
  (`PRINTO01.IXOBSA`, `CPALTT01.IXOB`, `IR3TIB01.ST`, `IRLTLT01.ST`,
  `SPASTT01.IXOB`) until the 2026-08-19 migration below.

**Now bundled permanently** (commit 8f649f3): `pkg/datasets/macropanel/oecd-
monthly.csv` - 30 economies, monthly `iso,date,ip,cpi,shortrate,longrate,
shareprice`. Generator `cmd/gen-macropanel`, `make macropanel`, accessor
`datasets.MacroPanel()`. This is the offline, reproducible substrate for the
breadth model.

### The 2026-08-19 migration off OECD MEI (and the merge that was not reproducible)

The MEI dataset the panel was built on stopped being updated in 2024-01 while
still answering HTTP 200, so the tactical allocation had been re-reading a
January 2024 macro state for two and a half years (the backtest holds the last
regime dated before each month, so a frozen panel freezes the allocation, it does
not stop it). The generator now reads the three current dataflows above.

What changed, measured old against new over their shared window:

- **Coverage is a strict superset.** Not one country-month was lost; every
  country and column now runs to 2026-03/2026-05 instead of 2024-01. Australia
  gains a (short, 2024→) monthly CPI, South Africa an industrial production
  series, Turkey's production goes back to 1958 and France's, Sweden's and
  Greece's to 1955. Still absent at the source: production for Australia and New
  Zealand, CPI for New Zealand, a long yield for Turkey, and Japanese CPI after
  2021-06 (all four were absent under MEI too).
- **The index columns are a different vintage**, so `ip` moves in 99 % of shared
  cells (new base year and revisions) and `cpi` in 11 %, `shareprice` in under
  4 %. Only ratios of these are ever read, so a rebasing is not a change of
  signal.
- **The short rate merge was order-dependent and is not any more.** The 3-month
  interbank rate and the immediate-rate fallback used to be fetched concurrently
  into one map, so either could win a month depending on which HTTP response
  arrived first; ~4300 cells flipped between two runs of the same binary. The
  sources are now applied in priority order after every fetch has landed, the
  primary owning every month it quotes. 2082 shared cells changed because the
  shipped file had been carrying the fallback where the 3-month rate exists
  (USA 708, DEU 504, KOR 395, POL 391 months), and two runs of the generator now
  produce byte-identical files.
- **Regimes barely move.** Over 1960-2024, 33 of 769 months (4.3 %) change
  quadrant, all isolated boundary crossings on the growth axis; mean absolute
  change in growth breadth 0.019 (max 0.099), in inflation breadth 0.0009, in
  the monetary slope 0.12 points. The backtest reads 5.00 %/yr real (vol 7.3 %,
  maxDD -22.5 %) where it read 5.25 % (7.5 %, -22.8 %): the same portrait, and
  the 2024-2026 years are now driven by live macro rather than a held signal.

Two lessons were folded back into the generator. A dataflow that freezes must
fail loudly, so `-check` (on by default) refuses to write a panel whose columns
do not reach within six months of today, or whose rate series end on a run of
repeated levels, or that misses one of five public anchors (US production over
the 2020 stop, the 2022 US inflation peak, the 2023 US 3-month rate, the 1981
and 2020 US long-yield extremes, the 2007-2009 fall in US share prices). And
level columns are spliced like donor chains: where manufacturing fills in front
of the industry aggregate, it is rebased onto it at the first month both quote.

Known data artifacts:
- **Gold in EUR/CAD before ~1999/1971 is wrong** (pofo returns identical bogus
  1970 values for EUR and CAD; GBP/JPY/USD check out). Use *global USD-real gold*
  (USD gold ÷ US CPI) for all countries to avoid it.
- JST broadsample (`datasets.BroadSample()`) has NO gold column and is annual;
  it covers 16 countries incl. USA/DEU/FRA/GBR/JPN but **not CAN**.
- Equity source materially changes drawdown conclusions (see §6): MSCI World
  (diversified TR) flatters drawdown vs a single-country price index.

## 4. The algorithm, precisely (so it can be rewritten)

Everything works on month-end real total-return series. `r_x[t]` is asset x's
real return over `[t-1, t]`; signals are read at `t-1` (rates) or `t-2` (macro).

### 4.1 Growth × inflation → equity weight

Per-country signals:
- `g_yoy(c,t)` = IP(c,t)/IP(c,t-12) − 1; `i_yoy(c,t)` similarly on CPI.
- **Accelerating** = the YoY rate is higher than 3 months earlier:
  `gAcc(c,t) = g_yoy(c,t) > g_yoy(c,t-3)`, likewise `iAcc`.  **[RECON]** (Darcet
  says "accelerating"; the 3-month lookback is our choice.)

Breadth "world point" over the country set (≥8 reporting), 3-month smoothed:
- `G(t)` = share of countries with `gAcc` true ∈ [0,1].
- `I(t)` = share with `iAcc` true ∈ [0,1].
- Paradise = (G,I) = (1,0); Hell = (0,1).  **[DARCET]** (the poles),
  **[RECON]** (using acceleration-breadth as the coordinate).

Equity weight, quadratic distance-to-hell damping:
```
dHell = hypot(G - 0, I - 1)                 # 0 at hell, sqrt2 at paradise
wEq   = min(1, wMax * (dHell / sqrt2)^p)     # p=2 → Darcet's 1/d² damping
```
- `p = 2` (quadratic) is **[DARCET]** in spirit and **[ROBUST]** (see frontier).
- `wMax` scales aggression. `wMax≈0.75` is conservative; `wMax≈1.6` is the
  return/drawdown sweet spot **[SELECTED]**.

(The earlier *single-country* variant instead placed a hell pole at
`(g%,i%)=(-3, 8)`, scaled each axis by 5 pts, and used
`wEq = wMax*(1 - 1/(1+(d/dRef)^2))`, `dRef=1.2`, `wMax=0.75`. **[RECON]**. It
produced the return edge but *worse* drawdowns - superseded by the breadth
version.)

### 4.2 Short × long rate → defensive split (gold / bonds / cash)

Signals (per-country then averaged over a G8 set for the global model):
- `slope = longRate − shortRate` (percentage points).
- `realShort = shortRate − i_yoy·100`.

Three poles in `(slope, realShort)` space and inverse-square (1/d²) weights:
```
poles: bonds=(2.0, 1.0), cash=(0.0, 2.5), gold=(0.5, -2.5)     # [RECON]
w_k = 1 / ((slope-px_k)^2 + (realShort-py_k)^2 + 0.25)          # eps=0.25
normalize w over {bonds,cash,gold}
```
Poles encode Darcet's rules: bonds when steep + positive real short; cash when
short rates high; gold when real rates negative. **[RECON]** placements,
**[DARCET]** intent.

### 4.3 Combination

```
defRet = w_bonds·r_bond + w_cash·r_cash + w_gold·r_gold
portRet = wEq·r_equity + (1 - wEq)·defRet
```
Monthly rebalance to these targets. Static Browne PP = 0.25 each, same assets.
Turnover ≈ 6%/month (~73%/yr); costs at 25 bps/turn shave ~19 bps CAGR. **[EMPIRICAL]**

## 5. Results ledger

All REAL. "DD" = max peak-to-trough. Prototypes archived (session scratchpad):
`darcet_*.go` (ppinvariant, defensive, full_pp2, regime_clustering, multicountry,
multicountry_jst, breadth_faithful).

### 5.1 Static PP invariant (JST broadsample, 16 countries, ~1871-2020)  [EMPIRICAL]
- Real CAGR ~3-4% hors-guerre (USA 4.1, DNK 4.4, SWE 3.9, GBR/CHE/NLD ~3.0);
  war countries collapse (DEU 1.8/-98%, JPN 1.7/-96%, FRA 1.1/-95%, PRT 1.0/-91%).
- Confirms **[DARCET]** invariant, and confirms the gold-saves-in-war claim **by
  absence**: these are goldless (JST has no gold), so -95/-98% is what a PP
  *without* the one war-proof asset looks like. **[ROBUST]** (16 countries).

### 5.2 Defensive block alone (1968-2026)  [EMPIRICAL]
- Tactical rotation 4.62% vs static equal-weight 2.90%, similar DD (~-30%).
  +1.7% real for the monetary rotation.

### 5.3 US full prototype (1970-2024), growth=INDPRO/OECD IP  [ROBUST]
- Tactical **6.06%**, DD -23.6, ret/vol 0.66; static PP 3.61%/-25.4/0.52; MSCI
  World 4.76%/-54/0.32.
- Robustness: edge in **every** third and half; edge at **every** start date
  1970-2010 (+1.5 to +2.9%); survives 25-50 bps costs. **[ROBUST]**
- Caveat: this "US" run used MSCI World as the equity leg → its low drawdown was
  partly a diversification artifact (see §6).

### 5.4 Regime clustering, single country (US/DE/GB/JP)  [EMPIRICAL]
- Paradise > Hell in all four, but weak (US 9.9 vs 5.8; DE 25.9 vs -2.6; GB 4.0
  vs 2.5; JP 0.7 vs -1.4). Disinflation quadrants (G↓I↓) also strong (US 13, JP
  19). Lesson: **inflation axis is the cleaner driver; growth is noisy single-
  country.** **[EMPIRICAL]**

### 5.5 Multi-country full backtest, crude data (6 countries)  [ROBUST/mixed]
- Return edge in **5/6** (+0.7 to +1.1%; JPN tie). Static invariant holds.
- **Drawdown WORSE than static PP in all 6** (still ≪ equity). "No gamelle" does
  not reproduce with crude single-country proxies.

### 5.6 Multi-country with JST-calibrated asset TR (5 countries)  [ROBUST]
- Return edge 4/5. **Drawdown still worse than static in all 5.** ⇒ the drawdown
  problem is **not** a data-quality artifact; it is intrinsic to a *noisy single-
  country signal* + concentration.

### 5.7 Faithful breadth construction (global, 1970-2024, 30-country breadth)  [ROBUST]
- **Clustering clean**: world-equity real return by breadth quadrant - paradise
  +6.4%, G+I+ +7.0%, G−I− +15.4%, **Hell (stagflation) −9.9%** (isolated,
  negative). This is the cleanest confirmation of **[DARCET]**'s core claim.
- **Drawdown restored**: with quadratic damping, DD ≈ static PP.
- **Frontier** (equity aggressiveness, quadratic p=2):
  | wMax | avgEq | CAGR | DD | ret/vol |
  |---|---|---|---|---|
  | 0.75 | 21% | 3.56% | -24.8% | 0.54 |
  | 1.00 | 29% | 4.11% | -23.0% | 0.60 |
  | 1.30 | 37% | 4.75% | -22.2% | 0.64 |
  | 1.60 | 46% | **5.35%** | **-21.5%** | **0.66** |
  (static PP 3.61%/-25.4/0.51; MSCI World 4.83%/-54/0.32)
- **Linear** damping (p=1) at wMax=1.6 gives 6.30% but DD -37.6%: the **quadratic
  damping is what holds drawdown low** - Darcet's 1/d² is load-bearing, not
  decoration. **[ROBUST]** (whole sweep), the specific wMax=1.6 is **[SELECTED]**.

### 5.8 The publication-honest information set  [ROBUST]

Everything above reads the macro panel by REFERENCE date. A real-time allocator
reads it by PUBLICATION date, and the two are not the same month.
`SignalConfig.ReleaseLags` holds each driver back by its publisher's delay;
`PublicationLags()` is the honest setting and the zero value is the historical
one, so both are reproducible from the same code.

#### The honest information set, per driver

`k` below reads: *at the start of month M, the latest reference month whose
value is published is M-k*. Verified against the publishers' own calendars on
2026-09-19; the `ReleaseLags` column is `k-1`, because `Simulate` already
applies one month of its own.

| driver | k | `ReleaseLags` | binding evidence |
|---|---|---|---|
| industrial production | **3** | `IP: 2` | Eurostat publishes the monthly production index "between 5 and 10 weeks after the end of the reference period", measured at t+43 to t+47 days over 2026 (ref March out 13 May, ref July out 16 September). The US Federal Reserve is faster ("about 15 days after the reference month ends", t+15 to t+18 over 2026), which alone would give k=2, but a breadth panel needs Europe, so Europe binds. |
| consumer prices | **2** | `CPI: 1` | BLS publishes the US CPI at t+10 to t+14 days; Eurostat's full HICP "usually between 15 and 18 days after the end of the reference month". Both are out before the start of M+1 for reference month M-1, so at the start of M the freshest is M-2. |
| short and long rates | **1** | `Rates: 0` | The monthly figure is an average of DAILY market quotes ("the monthly average interest rates for long-term government bonds issued by each country", ECB), so it is fully determined at the close of the reference month and an allocator reads it off a screen without waiting for any statistical agency. |

Sources: Federal Reserve G.17 release policy and performance evaluation
<https://www.federalreserve.gov/releases/g17/OMB/2026/OMB_2026.htm>; Eurostat
short-term business statistics quality and scope
<https://ec.europa.eu/eurostat/statistics-explained/index.php?title=Short-term_business_statistics_-_quality_and_scope>;
BLS CPI release schedule <https://www.bls.gov/schedule/news_release/cpi.htm>;
Eurostat HICP metadata
<https://ec.europa.eu/eurostat/cache/metadata/en/prc_hicp_esms.htm>; ECB
long-term interest rate statistics
<https://www.ecb.europa.eu/stats/financial_markets_and_interest_rates/long_term_interest_rates/html/index.en.html>.
Every `oecd.org` HTML page is behind a Cloudflare challenge and could not be
read; the OECD's own news-release PDFs and its live SDMX endpoint were used
instead, which is stronger evidence anyway.

Two of the three can be argued LONGER, and both readings are measured below:

- The OECD's own CPI news release is issued in the first days of month M and
  carries reference month M-2 (Paris, 8 September 2026 for July 2026), so a user
  of the OECD AGGREGATE rather than of the national releases is at k=3.
- The OECD database carries a month's rate average around day 10 to 15 of the
  next month (FRED's OECD-sourced copy `IRLTLT01USM156N` received the August
  2026 average on 15 September 2026), so a user who waits for the DATABASE
  rather than reading the market is at k=2.

`PublicationLags()` takes the shorter, defensible reading (3/2/1) because the
allocator does not have to go through the OECD: it can read its own CPI release
and its own screen. The longer reading (3/3/2, "conservative" in the tables) is
what someone who only ever touches the OECD database would have had.

**Two hazards found while checking the calendars, which concern the panel's
generator rather than the backtest.** First, `DSD_STES@DF_INDSERV` **version 4.0
still answers HTTP 200 and is frozen at 2024-03**, while the live version is
4.3: the exact failure mode of the `OECD/MEI` freeze documented in section 3,
one dataflow later. Second, the DBnomics mirror the generator reads is itself
BEHIND the OECD's SDMX endpoint: on 2026-09-19 DBnomics carried 2026-04/2026-05
where `sdmx.oecd.org` carried 2026-08, and the bundled panel ends accordingly
(`ip` at 2026-04, the other columns at 2026-05). That is three to four months of
silent staleness on top of every `k` above, and it is inside the generator's
six-month freshness tolerance, so nothing complains. A live allocation is
therefore stale by more than even the conservative setting; a backtest is not,
since it only ever reads months that are long since final.

The whole battery was re-run at every setting below, on today's data (global window
1970-01..2026-07, monthly real, `wMax=1.3` unless stated). The reference-date
row is the recomputation of the published model on the current panel and the
current quotes, so the before/after comparison is like for like; it sits a
little above the figures of 5.3/5.7, which were measured on earlier vintages and
shorter windows.

**Global model** (MSCI World / TLT / T-bills / gold, US-real):

| information set | CAGR | vol | maxDD | %UW | edge vs static | turnover/mo | net 25bp | net 50bp |
|---|---|---|---|---|---|---|---|---|
| reference date (M-1) - the published figures | 5.12% | 7.5% | -22.5% | 73% | **+1.29%** | 7.2% | 4.89% | 4.67% |
| **publication-honest** (IP M-3, CPI M-2, rates M-1) | **4.64%** | 7.6% | **-27.4%** | 75% | **+0.80%** | 7.1% | 4.42% | 4.19% |
| conservative (IP M-3, CPI M-3, rates M-2) | 4.44% | 7.6% | -30.8% | 76% | +0.60% | 7.2% | 4.21% | 3.99% |
| uniform lag 2 | 4.85% | 7.6% | -27.0% | 74% | +1.02% | 7.2% | 4.63% | 4.40% |
| uniform lag 3 | 4.52% | 7.6% | -30.1% | 76% | +0.69% | 7.2% | 4.30% | 4.07% |
| static Browne PP | 3.84% | 7.3% | -25.2% | 79% | - | 1.2% | 3.80% | 3.76% |
| MSCI World (equity) | 5.07% | 14.8% | -54.7% | 83% | - | - | - | - |

**Two thirds of the edge survives, and the drawdown does not.** The tactical
line keeps +0.80 points a year over the static PP instead of +1.29, but its
worst drawdown deepens from -22.5 % to -27.4 %, i.e. from *better* than the
static PP's -25.2 % to *worse*. Darcet's quadratic damping still cuts equity
before the stagflation corner; it now arrives late enough to be caught in part
of the fall. TURNOVER IS UNCHANGED (7.1 % vs 7.2 % a month): a staler signal is
not a busier one, so the honest setting costs nothing extra in fees.

**The decay is smooth, not a cliff** (global, edge over static by TOTAL
information lag; lag 0 is a deliberate look-ahead, the regime of the very month
being paid, and lag 1 is the published convention):

| total lag (months) | 0 | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|---|
| CAGR | 5.07% | 5.12% | 4.85% | 4.52% | 4.20% | 4.12% | 3.96% |
| edge vs static | +1.23% | +1.29% | +1.02% | +0.69% | +0.36% | +0.28% | +0.12% |
| maxDD | -21.3% | -22.5% | -27.0% | -30.1% | -34.8% | -36.0% | -36.5% |

Reading the very month the portfolio earns adds NOTHING (+1.23 vs +1.29): there
is no discontinuity at the look-ahead boundary, which is what a genuine
look-ahead artefact would show. From there the edge bleeds off at roughly a
third of a point per extra month of staleness and is gone by lag 6. That is the
signature of a real, slow-moving signal whose half-life is short: two extra
months of staleness halve it. It also sets the practical bound, and it is a
narrow one - this method tolerates being a month or two late, and nothing more.

**Per country** (each country's own equity, bond and cash sleeves from the panel
- share prices plus a flat 3 %/yr dividend add-back, a 20-year par-bond total
return from the long yield, the short rate accrued, all deflated by the
country's own CPI - global USD-real gold, and the same global breadth signal):

| country | window | static | tactical ref | tactical honest | edge ref | edge honest | maxDD ref | maxDD honest |
|---|---|---|---|---|---|---|---|---|
| USA | 1969-02..2026-05 | 3.86% | 5.14% | 4.79% | +1.28% | +0.93% | -19.8% | -23.0% |
| JPN | 1989-02..2026-04 | 2.91% | 4.64% | 3.75% | +1.73% | +0.84% | -24.9% | -25.3% |
| DEU | 1969-02..2026-04 | 4.16% | 5.93% | 5.54% | +1.77% | +1.38% | -26.0% | -24.6% |
| FRA | 1969-02..2026-04 | 4.40% | 6.18% | 5.99% | +1.78% | +1.59% | -32.7% | -40.4% |
| GBR | 1978-03..2026-04 | 4.27% | 6.05% | 5.95% | +1.77% | +1.67% | -20.1% | -19.0% |
| ITA | 1991-04..2026-04 | 4.66% | 7.22% | 7.37% | +2.56% | +2.71% | -22.3% | -19.3% |
| CAN | 1969-02..2026-04 | 4.30% | 5.85% | 5.41% | +1.54% | +1.10% | -25.1% | -26.6% |
| **average** | | | | | **+1.78%** | **+1.46%** | | |

The edge survives in all seven, and 82 % of it survives on average (+1.46
against +1.78; +1.29, or 73 %, at the conservative setting). What the honest lag
costs varies widely and not in an order this document can explain: Britain loses
0.10 points a year and Italy GAINS 0.15, while Japan loses 0.89 and the United
States 0.35. Turnover is 7.1 to 7.5 % a month at every setting in every country.
Australia has no per-country run: its monthly CPI only starts in 2024, so the
four sleeves never overlap for long enough (25 months). These runs vary the
ASSETS, not the signal; the single-country SIGNAL prototypes of 5.5 and 5.6 were
never in the repository and were not re-run.

**Subperiods and start dates.** The edge over the static PP, in points a year:

| information set | 1st third | 2nd third | 3rd third | 1st half | 2nd half |
|---|---|---|---|---|---|
| reference date | -0.29% | +1.10% | +2.95% | +0.56% | +1.99% |
| publication-honest | -1.44% | +1.27% | +2.47% | -0.25% | +1.84% |
| conservative | -1.75% | +1.18% | +2.26% | -0.48% | +1.67% |

| information set (start year, to 2026) | 1970 | 1980 | 1990 | 2000 | 2010 | 2015 |
|---|---|---|---|---|---|---|
| reference date | +1.29% | +2.33% | +2.08% | +1.79% | +2.94% | +3.67% |
| publication-honest | +0.80% | +2.10% | +1.93% | +1.65% | +2.29% | +3.38% |
| conservative | +0.60% | +1.90% | +1.78% | +1.45% | +1.76% | +2.59% |

Section 5.3's "edge in every third and half" DOES NOT SURVIVE: the first third
(the 1970s and early 1980s, the inflation era the model is supposed to be best
at) goes from -0.29 to -1.44 points a year, and the first half turns negative.
The start-date claim survives: every start year from 1970 to 2015 still shows a
positive honest edge, from +0.80 to +3.38. So the honest reading is that the
edge is real but concentrated in the last forty years, and that the model was
LATE, not wrong, during the great inflation.

**FIRE metrics** (4 % real, 30-year overlapping cohorts; the section 7 table
recomputed):

| information set | CAGR | vol | maxDD | %UW | longest UW | 4% survival | p10 terminal |
|---|---|---|---|---|---|---|---|
| tactical, reference date | 5.12% | 7.5% | -22.5% | 73% | 10.1y | 100% | 0.80x |
| **tactical, publication-honest** | 4.64% | 7.6% | -27.4% | 75% | 12.4y | 100% | **0.29x** |
| tactical, conservative | 4.44% | 7.6% | -30.8% | 76% | 12.8y | 98% | 0.18x |
| static Browne PP | 3.84% | 7.3% | -25.2% | 79% | 6.1y | 100% | 0.47x |

40-year ruin probability at a fixed real withdrawal (same stationary bootstrap,
mean block 24 months, 3000 paths):

| withdrawal | 3.0% | 3.5% | 4.0% | 4.5% |
|---|---|---|---|---|
| tactical, reference date | 0.1% | 1.7% | 5.5% | 14.6% |
| **tactical, publication-honest** | 1.5% | 4.5% | 11.9% | 22.6% |
| tactical, conservative | 2.0% | 5.8% | 14.1% | 26.6% |
| static Browne PP | 1.6% | 6.7% | 18.9% | 37.5% |

This is where the honest lag bites hardest, and it changes a conclusion.
Section 7 sold the overlay on two things: the best worst-case cushion among the
low-risk options, and a ROUGHLY HALVED ruin probability. The cushion goes: at
the honest lag the 10th-percentile terminal wealth is 0.29x, BELOW the static
PP's 0.47x, because the deeper drawdown lands on the withdrawals. The ruin
advantage survives but shrinks from a halving to about a third off (11.9 % vs
18.9 % at 4 %), and at a 3 % withdrawal it disappears into the noise (1.5 % vs
1.6 %). The overlay remains the better of the two at 3.5 % and above; it is no
longer the better CUSHION in a bad sequence.

**Revisions are a second gap, and a lag cannot close it.** The panel holds
today's revised index levels, not the vintage the allocator would have read. A
real-time backtest needs real-time VINTAGES, which this repository does not have
and DBnomics does not serve, so the size of this one is stated from the
publishers' own revision studies rather than measured here:

- **Industrial production is the exposed column.** The Federal Reserve measures
  the average revision to the percent change in total IP, disregarding sign, at
  **0.24 percentage point** from the first to the fourth estimate over
  1987-2025, and **0.9 point** mean absolute once the annual benchmarks are
  folded in (benchmarks are mildly biased downward, -0.4 point on average).
  Eurostat measures the mean absolute revision of the euro-area production
  year-on-year rate at **0.3 point** after one month and **0.8 point** after
  thirty-six. The regime reads an ACCELERATION of a year-on-year rate, i.e. a
  difference of two such numbers, so a few tenths of a point is exactly the
  scale at which a country flips side in the breadth count.
- **Consumer prices are safe.** The BLS states that the CPI-U and CPI-W "are
  final when issued"; only the SEASONALLY ADJUSTED series is revised, for five
  rolling years each February, and the panel reads the non-seasonally-adjusted
  level. (The chained C-CPI-U is the one genuinely provisional US price index,
  and the panel does not use it.)
- **Market rates are safe**, being transaction data: the ECB revises them only
  to correct errors, which Eurostat's metadata calls "extremely infrequent".

Sources: <https://www.federalreserve.gov/releases/g17/OMB/2026/OMB_2026.htm>,
<https://ec.europa.eu/eurostat/statistics-explained/index.php?title=Short-term_business_statistics_-_revisions>,
<https://www.bls.gov/cpi/questions-and-answers.htm>,
<https://ec.europa.eu/eurostat/cache/metadata/en/irt_st_esms.htm>.

So the figures above remain an upper bound even at the honest lag, by an amount
concentrated entirely in the growth axis. That cuts the same way as everything
else in this section: the growth breadth is the fragile half of the signal, and
section 6's standing finding that the inflation axis dominates it is, if
anything, reinforced.

## 6. What we learned about regimes & invariants

- **[ROBUST] Static-PP real invariant** (~inflation + 3%) is real across
  countries and eras; the failure mode is war-on-soil (gold-only survival).
- **[ROBUST] Stagflation is the equity killer**, and it is cleanly separable
  *with breadth* (-9.9% vs +6/+15% elsewhere); single-country the signal is
  noisy. Breadth is not cosmetic - it is what makes the growth axis usable.
- **[ROBUST] The inflation axis dominates the growth axis.** Disinflation (even
  with slowing growth) is great for equities; the growth breadth mostly helps
  avoid the stagflation corner. The publication lag says the same thing from the
  other side (5.8): production is the slowest driver to be released AND the only
  one materially revised, so the growth axis is where both honesty costs are
  concentrated.
- **[ROBUST] Quadratic (1/d²) damping trades almost no return for a large
  drawdown reduction**; linear scaling of the same signal concentrates and blows
  drawdown. This is the single most important mechanical finding.
- **[EMPIRICAL] Equity leg diversification matters**: a global (MSCI World) leg
  has far lower drawdown than any single-country leg, independent of the tactical
  overlay. Do not attribute a diversified leg's calm to the timing model.
- **[EMPIRICAL] Data quality changed levels but not the drawdown conclusion**;
  the fix was the *signal* (breadth + damping), not the *assets*.

## 7. FIRE / decumulation relevance  [EMPIRICAL]

The question that matters for a retiree is not CAGR or maxDD but **sequence
risk**: withdrawing during a drawdown permanently impairs capital, and long
stretches under water are where FIRE plans die. We measured the FIRE metrics
directly (REAL, monthly, 1970-2024; global breadth model at a **moderate**
`wMax=1.3`, not the cherry-picked 1.6; 4% real withdrawal over overlapping 30-year
cohorts, Bengen-style). Prototype archived as `darcet_fire_relevance.go`.

| strategy | CAGR | vol | maxDD | %UW | longest UW | 4% survival | p10 terminal |
|---|---|---|---|---|---|---|---|
| GLOBAL tactical PP2.0 | 4.83% | 7.5% | -22.7% | 75% | 10.4y | **100%** | **0.56x** |
| GLOBAL static Browne PP | 3.61% | 7.0% | -25.4% | 79% | **6.2y** | **100%** | 0.46x |
| GLOBAL 60/40 (World/bond) | 4.38% | 10.6% | -44.7% | 79% | 12.8y | 97% | 0.32x |
| MSCI World (100% equity) | 4.83% | 14.9% | -54.1% | 85% | 12.7y | 98% | 0.59x |
| JAPAN home-bias static PP | 3.35% | 6.1% | -24.6% | 81% | 16.4y | 85% | **0.00x** |
| JAPAN equity only | 3.68% | 15.1% | -66.2% | 92% | **31.1y** | 76% | 0.00x |

%UW = share of months below the prior real peak; longest UW = worst underwater
stretch; 4% survival = share of 30y cohorts never ruined at a 4% real draw; p10
terminal = 10th-percentile real wealth left after 30y at 4% (1.00x = starting
capital, 0 = ruined).

READ THIS TABLE WITH 5.8 OPEN: its tactical rows are at the reference-date lag,
i.e. optimistic. At the publication-honest lag the tactical line's drawdown
deepens to -27.4 %, its longest underwater stretch to 12.4 years and its p10
terminal cushion falls to 0.29x, BELOW the static PP's. The findings below are
amended accordingly.

Findings:
- **Relevant for FIRE: yes, but on ruin, not on cushion.** The tactical PP2.0
  gave **100% historical 4% survival** at both lags. At the reference-date lag
  it also had the best worst-case cushion among the low-risk options (0.56x
  then, 0.80x on today's longer window), at 7.5% vol and -23% drawdown. **That
  cushion does not survive an honest information set** (0.29x against the static
  PP's 0.47x, 5.8): arriving two months late on the growth signal deepens the
  drawdown, and a drawdown taken while withdrawing is permanent. 60/40 and
  equity remain far more fragile to a bad start (0.32x / 98%).
- **Japan lost decades are cured by GLOBAL diversification, not by the overlay.**
  Japan-equity-only = 31y under water, -66%, 24% of retirements ruined; a Japan-
  concentrated static PP still ruins 15% of cohorts. Any *global* PP → 100%
  survival. The tactical tilt does not fix single-country concentration (and the
  single-country signal is noisy anyway, §5.4); diversification does.
- **Long underwater periods: nuanced.** The tactical spends *less* total time
  under water (75% vs 79%) and crushes equity/60-40 (~13y), but its single
  *longest* stretch (10.4y) is **longer** than the static PP's (6.2y): it
  compounds to higher peaks that take longer to reclaim. If the specific fear is
  "years below my high-water mark," the **static PP is calmer**; the tactical
  trades a longer worst-case underwater for more return, survival and terminal
  cushion.

**General lessons for any regime-quadrant portfolio (reusable for Dragon, All-
Weather, etc.)**  [EMPIRICAL/ROBUST]:
1. For decumulation, judge a quadrant strategy on **survival, worst-case terminal
   wealth and time-under-water**, not CAGR/Sharpe. A quadrant overlay earns its
   keep by *shrinking left-tail sequence risk* (shallow drawdowns), which these
   portfolios do well, more than by raising the mean.
2. **Diversification dominates timing for tail risk.** No quadrant overlay rescues
   a single-country concentration (Japan); breadth/global exposure does. Build the
   asset base globally first, then let the regime overlay shape the tilt.
3. **The overlay can lengthen the *worst* underwater stretch even while lowering
   average drawdown**, because higher compounding raises the bar to reclaim. Track
   *longest* underwater, not just %-under-water or maxDD, when selling "smoothness".
4. **Quadratic (1/d²) damping is a decumulation-friendly shape**: it de-risks
   hardest exactly at the stagflation corner where real sequence risk concentrates.
5. **Monte-Carlo ruin bands (done)**: the realized tactical and static real
   series are block-bootstrapped (stationary bootstrap, mean block 24 months)
   through `pkg/decumul`, giving proper ruin probabilities rather than overlapping
   cohorts. 40-year ruin at a fixed real withdrawal (`pofo -permanent`):

   | withdrawal | 3.0% | 3.5% | 4.0% | 4.5% |
   |---|---|---|---|---|
   | tactical PP2.0 (reference-date lag) | 0.3% | 2.2% | 7.2% | 15.7% |
   | tactical PP2.0 (publication-honest) | 1.5% | 4.5% | 11.9% | 22.6% |
   | static Browne PP | 1.4% | 6.0% | 17.3% | 37.2% |

   At the reference-date lag the tactical overlay roughly **halves the ruin
   probability at every rate** (4%: 7.2% vs 17.3%). At the publication-honest lag
   the halving becomes **about a third off** (11.9% vs 18.9% on the same rerun,
   5.8), and at a 3% withdrawal the advantage disappears into the noise. The
   mechanism still holds: shallower drawdowns convert directly into lower
   sequence-of-returns ruin, and the overlay still wins at 3.5% and above. The
   honest row is from the 5.8 rerun (the reference-date and static rows are the
   original run; the rerun reads 5.5% and 18.9% at 4%, i.e. the same picture).
   Still a single historical path (one economy's realized series), pre-tax and
   pre-fee.

## 8. Generalizing the framework (e.g. Artemis Dragon)

Abstract the method into three reusable pieces, all now data-supported here:

1. **A regime signal** = a point in a low-dim macro space (growth×inflation,
   short×long rate), ideally as **cross-country breadth** (smoother, less
   overfit-prone) rather than single-country levels. `datasets.MacroPanel()`
   already provides the breadth substrate.
2. **A distance-to-danger, quadratically-damped weight** per asset toward its
   favorable pole. The 1/d² damping is the drawdown control.
3. **A real-terms monthly backtest harness** with a lookahead guard and a
   turnover/cost model.

Mapping onto the **Artemis Dragon** (≈ equity, fixed income/long-duration, gold,
commodity-trend/CTA, long-volatility): the same two quadrants drive four of the
five buckets (equity from growth×inflation; bonds/cash/gold from the monetary
quadrant). Long-vol and commodity-trend are *convexity/crisis* sleeves - they
want a **third axis**: a stress/volatility or trend-strength signal (e.g. VIX
level `^VIX`, or realized-vol / oil-momentum from the panel). A Dragon 2.0 would
size the crisis sleeves *up* as the world point approaches hell, exactly where
1/d² is cutting equity - a natural, testable extension. The harness, breadth
panel and damping generalize unchanged; only the pole map and the extra axis are
new.

Caution when generalizing: the pole placements (§4) and `wMax` are **[RECON]/
[SELECTED]** for the PP; re-fitting them per portfolio is where overfit creeps
in. Prefer keeping the *shape* (quadratic damping, breadth) fixed and changing
only the pole *positions* from a-priori economics, then validate with the same
subperiod/start-date/multi-country battery used here.

## 9. Honest overfit ledger

- Weight function is **[RECON]** (Darcet's is secret): we test the mechanism.
- Poles/scales set once, **not optimized** - but also not out-of-sample-validated
  per parameter. The frontier sweep is our robustness evidence; `wMax=1.6` is a
  post-hoc pick.
- In-sample, real-terms, pre-tax. Turnover cost tested (survives); taxes and ETF
  fees not modelled.
- Gold pre-1999 EUR/CAD FX unusable (worked around with global USD-real gold).
  The OECD MEI mirror froze at 2024-01 and the panel was migrated off it on
  2026-08-19 (see the migration note in section 3); the regime series moved by a
  handful of boundary months.
- Single realized path per country/global; no Monte-Carlo bands on the edge.
- **The lag is by reference date, not by publication date** (section 3), and the
  gap is now MEASURED (5.8), not merely flagged. It was worth measuring: over
  1960-2026 consecutive regimes flip quadrant in 28 % of months and move the four
  sleeves by 14.5 points of allocation on average (L1 distance, 58.9 points at
  the worst month). Under the honest information set two thirds of the global
  edge survives (+0.80 against +1.29 points a year) and 82 % of the per-country
  average (+1.46 against +1.78), the decay with staleness is smooth rather than a
  cliff, and turnover does not move. What does NOT survive: the drawdown claim
  (-27.4 % against the static PP's -25.2 %, i.e. worse rather than better), the
  "edge in every third and half" of 5.3, and the FIRE cushion of section 7. Every
  figure in this document that does not say "publication-honest" is still at the
  reference-date lag and is therefore an upper bound; the tables of 5.8 are the
  ones to quote.
- **Revisions remain unmeasured, and cannot be fixed by a lag.** The panel holds
  today's revised levels; a real-time backtest needs real-time vintages, which
  this repository does not have. Industrial production is the exposed column,
  consumer prices much less so (5.8).

## 10. Status & next steps

Shipped:
- `pkg/datasets/macropanel` (OECD monthly macro panel, `make macropanel`).
- `pkg/permanent`: Panel -> Regime (breadth world point + monetary means) ->
  quadratically-damped four-sleeve Allocation -> monthly-real Simulate/Compute.
  Pure core, offline tests, epistemic tags kept in the godoc.
- `pofo -permanent` CLI: fetches the four sleeves, deflates to real, drives the
  allocation from the panel, prints the backtest stats and the decumul ruin
  table (§7 caveat 5), each tactical line at BOTH information sets.
- `SignalConfig.ReleaseLags` + `PublicationLags()`: the publication-honest
  information set, measured across the whole battery in §5.8.

Not made the default, deliberately. `DefaultSignalConfig` keeps the zero lag
because `Regime` has two jobs: it drives a backtest (where the honest lag is the
only defensible reading) and it DESCRIBES a past month's macro state for
reporting (`pkg/compare`'s regime strip, `Regime.Quadrant`), where the
reference-date reading is the right one and a publication lag would mislabel the
month. Callers running a backtest should set `PublicationLags()`, as
`pofo -permanent` now does for its honest row. Flipping the default would touch:
`DefaultSignalConfig` and its doc, the regime strip in `pkg/compare/contrib.go`
(which would silently start labelling months by what was known rather than by
what happened) and its chart snapshots, `cmd/pofo/permanent.go`, this document's
sections 5.3 to 5.7 and 7 in full, and the examples in
`pkg/permanent/example_test.go`.

Open:
- **Add the measured drivers of the honest lag to the generator's guards**
  (§5.8): `DSD_STES@DF_INDSERV` 4.0 answers HTTP 200 while frozen at 2024-03,
  and the DBnomics mirror runs three to four months behind `sdmx.oecd.org`.
  Neither trips the six-month freshness check.
- Retune only pole *positions* and `wMax` from a-priori economics, not by fitting;
  re-run the subperiod/start-date/multi-country/frontier battery on any change.
- Generalize to the Artemis Dragon (§8): add the third stress/vol axis and the
  crisis sleeves; the panel, damping and harness carry over.
- Deeper decumul work: per-country ruin (single-market sequence risk), tax and
  fees, and a `scenario.Source` that regenerates the tactical stream inside the
  bootstrap rather than resampling one realized path.

See memory `darcet-permanent-portfolio-2` and `[[fire-decumulation-followups]]`
(shared real-terms / scenario machinery).
