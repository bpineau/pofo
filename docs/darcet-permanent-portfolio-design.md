# Darcet's tactical Permanent Portfolio 2.0 — research notes & design

Status: research complete for a US + multi-country prototype; the permanent data
(`pkg/datasets/macropanel`) is shipped; the allocator itself is not yet a package
(candidate `pkg/permanent`). This document is the complete record so the method
can be refined, retuned, generalized (e.g. to the Artemis Dragon) or rewritten
from scratch without re-deriving anything.

## 0. Provenance

Didier Darcet (Gavekal / GaveTracks), video <https://www.youtube.com/watch?v=JRkJUznoszM>,
revisiting Harry Browne's Permanent Portfolio. Source material captured by
notebooklm (transcripts, ephemeral, were in `/tmp/darcet1.txt`, `darcet2.txt`,
`darcet-transcript.txt`). Everything below that is not attributed to Darcet is
our reconstruction or our empirical result.

## 1. Epistemic tags (READ THIS FIRST)

Ben's first requirement: never confuse what was *described in advance* (more
reliable) with what we *found by fitting data* (overfit risk). Every claim below
carries one tag:

- **[DARCET]** — asserted by Darcet in the video, i.e. specified before we
  touched any data. The most reliable layer, but note he does **not** disclose
  his exact formulas, thresholds or weights (he says so explicitly).
- **[RECON]** — our reconstruction of a gap Darcet left open (a weight function,
  poles, scales). Set **once** from his qualitative rules and **not** optimized
  against the data. Medium reliability: not tuned, but a judgement call.
- **[SELECTED]** — a specific parameter value we picked *after* seeing results
  (e.g. `wMax≈1.6` as the frontier sweet spot). **Highest overfit risk**; treat
  as illustrative, not validated.
- **[EMPIRICAL]** — a measured result from running on data. Reliability depends
  on whether it survived out-of-sample checks (see next tag).
- **[ROBUST]** — an empirical result that held across independent subperiods,
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

**Equity block** — driven by exactly two macro variables, **growth** and
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

**Defensive block** — for the non-equity sleeve, arbitrate gold / long bonds /
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

All work is REAL (inflation removed), monthly, no lookahead (macro signals lagged
one extra month for the publication delay).

Assets (via pofo `marketdata`, SIM suffix splices long history):
- Equity: `URTHSIM` (MSCI World TR, 1969→) for the global model; OECD MEI
  `SPASTT01` price index + flat dividend add-back for per-country.
- Long bonds: `TLTSIM` (1962→); per-country = synthetic from the long yield.
- Cash: from `^IRX` (13-week bill, 1960→) compounded; per-country from short rate.
- Gold: `XAUUSDSIM` (1968→), USD; converted to a currency by pofo FX.
- Deflator / inflation signal: `^CPI-US` (bundled) and per-country CPI.

Macro signals — **the key sourcing lesson**:
- **FRED is UNREACHABLE from the sandbox** (HTTP/2 INTERNAL_ERROR + timeout even
  on HTTP/1.1). Do not rely on it here.
- **DBnomics is reachable** and mirrors OECD MEI. The growth proxy is US/where-
  needed industrial production `OECD/MEI/<ISO>.PRINTO01.IXOBSA.M` (1919→ for US;
  the OECD MEI mirror is frozen ~late 2023).
- Interest rates: `IRLTLT01.ST` (long), `IR3TIB01.ST` (3-month, else
  `IRSTCI01.ST` immediate). Share prices: `SPASTT01.IXOB`. CPI: `CPALTT01.IXOB`.

**Now bundled permanently** (commit 8f649f3): `pkg/datasets/macropanel/oecd-
monthly.csv` — 30 economies, monthly `iso,date,ip,cpi,shortrate,longrate,
shareprice` from OECD MEI. Generator `cmd/gen-macropanel`, `make macropanel`,
accessor `datasets.MacroPanel()`. This is the offline, reproducible substrate for
the breadth model.

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
produced the return edge but *worse* drawdowns — superseded by the breadth
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
- **Clustering clean**: world-equity real return by breadth quadrant — paradise
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
  damping is what holds drawdown low** — Darcet's 1/d² is load-bearing, not
  decoration. **[ROBUST]** (whole sweep), the specific wMax=1.6 is **[SELECTED]**.

## 6. What we learned about regimes & invariants

- **[ROBUST] Static-PP real invariant** (~inflation + 3%) is real across
  countries and eras; the failure mode is war-on-soil (gold-only survival).
- **[ROBUST] Stagflation is the equity killer**, and it is cleanly separable
  *with breadth* (-9.9% vs +6/+15% elsewhere); single-country the signal is
  noisy. Breadth is not cosmetic — it is what makes the growth axis usable.
- **[ROBUST] The inflation axis dominates the growth axis.** Disinflation (even
  with slowing growth) is great for equities; the growth breadth mostly helps
  avoid the stagflation corner.
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

Findings:
- **Relevant for FIRE: yes, strongly.** The tactical PP2.0 gave **100% historical
  4% survival** with the best worst-case cushion (0.56x) among the low-risk
  options, at 7.5% vol and -23% drawdown. That is exactly the sequence-risk
  profile FIRE wants: little damage when you draw. 60/40 and equity match the
  return but are far more fragile to a bad start (0.32x / 98%).
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
5. **Caveat carried into any FIRE claim**: overlapping cohorts on one historical
   path (~24 near-independent starts) are encouraging but are **not** a Monte-Carlo
   ruin probability. The proper test is to feed these real-return series into
   `pkg/decumul` (bootstrap/parametric `scenario.Source`) for ruin bands — planned,
   not yet done.

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
quadrant). Long-vol and commodity-trend are *convexity/crisis* sleeves — they
want a **third axis**: a stress/volatility or trend-strength signal (e.g. VIX
level `^VIX`, or realized-vol / oil-momentum from the panel). A Dragon 2.0 would
size the crisis sleeves *up* as the world point approaches hell, exactly where
1/d² is cutting equity — a natural, testable extension. The harness, breadth
panel and damping generalize unchanged; only the pole map and the extra axis are
new.

Caution when generalizing: the pole placements (§4) and `wMax` are **[RECON]/
[SELECTED]** for the PP; re-fitting them per portfolio is where overfit creeps
in. Prefer keeping the *shape* (quadratic damping, breadth) fixed and changing
only the pole *positions* from a-priori economics, then validate with the same
subperiod/start-date/multi-country battery used here.

## 9. Honest overfit ledger

- Weight function is **[RECON]** (Darcet's is secret): we test the mechanism.
- Poles/scales set once, **not optimized** — but also not out-of-sample-validated
  per parameter. The frontier sweep is our robustness evidence; `wMax=1.6` is a
  post-hoc pick.
- In-sample, real-terms, pre-tax. Turnover cost tested (survives); taxes and ETF
  fees not modelled.
- OECD MEI mirror frozen ~2023; gold pre-1999 EUR/CAD FX unusable (worked around
  with global USD-real gold).
- Single realized path per country/global; no Monte-Carlo bands on the edge.

## 10. Next steps

Candidate `pkg/permanent`: read `datasets.MacroPanel()`, compute the breadth
world point + monetary means, apply §4 weights, backtest vs static PP and MSCI
World, expose a CLI mode. Port from the archived `darcet_breadth_faithful.go`.
Keep the epistemic tags in the godoc so the [RECON]/[SELECTED] parameters stay
visibly provisional. See memory `darcet-permanent-portfolio-2` and
`[[fire-decumulation-followups]]` (shared real-terms/scenario machinery).
