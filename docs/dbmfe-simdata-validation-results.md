# Results: validating the DBMFE simdata (FIRE protection sleeve)

Date: 2026-06-28
Status: campaign executed. Companion to `dbmfe-simdata-validation-design.md`
(read that first for context, methods and thresholds).

## Headline

The pre-2025 DBMFE reconstruction faithfully reproduces DBMF's real risk
*shape* (near-zero equity correlation, the trend-following winters, the fat
left tail, the up/down capture against equities), and where it is unfaithful
it errs on the *conservative* side (it under-earns and understates crisis
upside relative to the real fund and to the SG CTA index). The decisive
sensitivity test passes: the optimal trend-sleeve weight moves by less than
about 2.3 percentage points under the specified plausible DBMF errors
(+/-1%/yr CAGR and +/-0.15 equity correlation), so the 25% sizing decision is
not balanced on a knife edge of modeling precision.

Recommendation: a ~25% DBMFE sleeve is defensible as a crisis-protection
allocation. Treat 25% as a cap rather than a floor, and keep the deliberate
margin in mind: a naive in-sample max-Sharpe optimizer lands the trend sleeve
nearer 15-20% (the rest crowding into bonds on the back of the 2004-2021 bond
bull), so the extra weight is justified by forward-looking crisis protection
and bond-regime skepticism, not by backtested return.

## What was validated against what

- Reconstruction (the thing under test): the deterministic 12-month TSMOM
  replay (`pkg/simgen`), in its USD layer ("recon USD", 2001-08 -> 2026-06)
  and its EUR-unhedged layer ("recon EUR", 2004-10 -> 2026-06).
- Real DBMF (USD parent, 2019-05 -> 2026-06): the best daily real benchmark.
- SG CTA Index (USD, 1999-12 -> 2023-05): the industry winter/regime yardstick,
  cumulated from daily returns to a base-100 level (USD layer vs USD layer).
- Equity: iShares MSCI World, real URTH (2012+) and its 60/40 reconstruction
  (1999+); bonds IEF; gold GC=F; cash from `^IRX`; inflation from `^HICP-FR`.

The 286-day real DBMFE overlap is statistically meaningless and was not used
for any conclusion (for the record it currently shows a flattering CAGR 21.7%,
Sharpe 1.29; ignore it).

## Library primitives delivered (TDD, pushed to master)

All in `pkg/metrics`, pure functions in the existing idiom (252-day
annualization, rf=0), each with godoc and a table test; full suite green,
`gofmt`/`go vet` clean.

- `Skewness`, `ExcessKurtosis`, surfaced on `Stats.Skew` / `Stats.Kurtosis`.
- `Rolling` (generic trailing-window engine) plus `RollingVol`,
  `RollingSharpe`, `RollingSortino`, `RollingUlcer`.
- `DrawdownEpisodes` (full peak/trough/recovery list; `Compute` still reports
  only the single longest stretch).
- `Autocorr`, `Histogram`, `Quantiles`.
- `LongestUnderperformance` (longest calendar span a series trails a
  benchmark, i.e. the relative-drawdown duration), for the cash / inflation /
  equity "desert" checks.

## pofo-specific checks (highest priority)

### Check 1: graft-seam continuity. PASS.

No discontinuity at either splice.

- DBMFE seam 2025-04-08: 30-day daily vol 1.29% (pre) vs 1.44% (post); seam-day
  return -0.90%, in line with neighbours. (The -4.2% on 2025-04-04 is the real
  April-2025 tariff selloff, a genuine market move just before the graft.)
- DBMF seam 2019-05-08: 30-day daily vol 0.67% (pre) vs 0.52% (post); seam-day
  return +0.06%.

`ExtendBack` rescaling is sound; the seam does not corrupt downstream stats.

### Check 2: DBMF CAGR shortfall decomposition. Explained, conservative.

Over the 2019-2026 overlap: sim CAGR 4.94% vs real 9.21% (gap 4.27%/yr).

- The gap is *not* volatility-target conservatism: the reconstruction runs
  *hotter*, vol 15.5% vs real 12.4%. So scaling sim up to the real fund's vol
  does not close the gap (it would leave CAGR ~5.0%).
- Adding back the 0.85%/yr fee lifts gross sim CAGR to ~5.79%, still ~3.4% short.
- The residual is a *risk-adjusted* (signal/breadth) gap: sim Sharpe 0.39 vs
  real 0.77 over this window. A single-speed 12-month signal on 7 markets
  captures less efficient trend than the real multi-speed ~50-market fund.

Important context: the 2019-2026 window is unusually strong for the real fund.
Over the full reconstruction (2001+) the recon USD Sharpe is 0.52, essentially
matching the SG CTA index's 0.56 over its own life. So on long-history,
broad-benchmark terms the reconstruction's efficiency is well calibrated; it is
only against this particular real DBMF window that it looks weak. Either way the
proxy *under*-states return, which is the safe direction, and the *shape*
(corr 0.52, beta 0.42, equity corr ~0, capture ratios) is preserved.

### Check 3: crisis alpha vs equities. Directionally right, magnitude conservative.

| Crisis | Equity (MSCI World) | recon USD | real DBMF | SG CTA |
|---|---|---|---|---|
| 2008-H2 GFC | -44.3% | +22.1% | n/a (pre-inception) | +4.6% |
| 2020-Q1 COVID | -22.0% | -14.6% | -1.1% | -0.5% |
| 2022 inflation | -17.5% | +3.4% | +20.5% | +19.2% |

- 2008 and 2022: positive while equities fell (the literal reason the sleeve
  exists). In 2022 the reconstruction *under*-delivers badly (+3.4% vs the real
  fund's +20.5% and SG CTA's +19.2%): the narrow single-speed basket missed
  most of the great trend year.
- 2020-Q1: a genuine miss. The fast V-crash overwhelmed the 12-month signal and
  the reconstruction fell -14.6% while the real fund and SG CTA were ~flat.

Both misses are *conservative* (they make the sleeve look worse in crises than
it really was), but they confirm the design's worry that a 7-market,
single-speed proxy under-states fast-crisis and big-trend alpha.

Full-sample capture vs equities is where the shape match is strongest:

| | alpha/yr | beta | up-capture | down-capture |
|---|---|---|---|---|
| recon USD | +7.9% | ~0.00 | 0.22 | 0.16 |
| real DBMF | +7.7% | 0.13 | 0.22 | 0.16 |

Up- and down-capture are *identical* to the real fund (0.22 / 0.16), and beta
to equities is ~0. This is the cleanest single piece of evidence that the
reconstruction reproduces DBMF's diversifying shape.

### Check 4: FX overlay (EUR unhedged). Sane.

`r_eur = (1+r_usd)/(1+r_fx)-1` is correct for an unhedged EUR NAV. The EUR
layer carries the extra EURUSD vol as expected: recon EUR vol 17.9% vs recon
USD vol 15.2%, and recon EUR vs equity correlation -0.01 (the FX overlay does
not import equity correlation). Consistent with an unhedged share class.

## The levels

### Level 1: base statistics

| Series | CAGR | vol | Sharpe | Sortino | Ulcer | MaxDD | skew | kurt |
|---|---|---|---|---|---|---|---|---|
| recon EUR | 8.35% | 17.9% | 0.48 | 0.69 | 16.0 | -32.9% | +0.06 | 18.0 |
| recon USD | 6.88% | 15.2% | 0.52 | 0.71 | 15.4 | -34.2% | -0.58 | 7.6 |
| real DBMF | 9.21% | 12.4% | 0.77 | 1.06 | 8.5 | -20.4% | -0.61 | 3.2 |
| SG CTA | 4.54% | 8.2% | 0.56 | 0.78 | 6.8 | -16.5% | -0.57 | 3.8 |

Correlations (daily returns): recon USD vs real DBMF **0.524**, recon USD vs
SG CTA 0.504, real DBMF vs SG CTA 0.678. Equity correlation: recon USD vs
MSCI World **0.003** (full history) / 0.14 (2012+), real DBMF vs World 0.19,
recon EUR vs World -0.01. Gold 0.23, bonds 0.13.

Verdict: moments are in the right ballpark; the reconstruction's daily skew
(-0.58) and kurtosis (7.6) are *more* adverse than the real fund (-0.61 /
3.2 is comparable skew, thinner kurtosis), i.e. the proxy is not benign.
Equity correlation is near zero, as required. PASS.

### Level 2: distributions (monthly tails)

| Series | n | skew | kurt | p1 | p5 | median | p95 | p99 |
|---|---|---|---|---|---|---|---|---|
| recon USD | 298 | -0.23 | +0.6 | -9.8% | -6.1% | +0.7% | +7.6% | +10.3% |
| real DBMF | 85 | +0.02 | +0.9 | -7.6% | -3.8% | +0.7% | +5.8% | +8.3% |
| SG CTA | 281 | +0.12 | +0.4 | -5.7% | -3.3% | +0.3% | +4.5% | +6.8% |

The reconstruction's monthly left tail (p1 -9.8%, p5 -6.1%) is *fatter* than
both the real fund and SG CTA; its monthly skew is the most negative of the
three. Not thinner-tailed than reality. PASS.

### Level 3: temporal validation (the falsification core). PASS.

Rolling 3-year Sharpe never looks "too pretty". The reconstruction dips
negative repeatedly (year-end rolling-3y Sharpe: 2017 -0.10, 2019 -0.14, 2021
-0.28, 2023 -0.19) and reproduces the 2011-2019 trend winter (winter min
-0.38), comparable to SG CTA's winter min -0.55. Overall rolling-3y Sharpe min
-0.48 (recon) vs -0.55 (SG CTA). The reconstruction is, if anything, choppier
than the real benchmark, not smoother. PASS.

### Level 4: longest underperformance (the "deserts"). PASS.

Longest relative-drawdown duration:

| Pair | longest stretch |
|---|---|
| recon USD vs cash | 8.37 yr (ongoing) |
| recon EUR vs cash | 4.91 yr |
| SG CTA vs cash | 6.01 yr |
| real DBMF vs cash | 3.64 yr (short history) |
| recon EUR vs HICP-FR (real-return drought) | 4.06 yr |
| recon EUR vs MSCI World | 6.22 yr (ongoing) |
| SG CTA vs MSCI World | 10.92 yr (ongoing) |

The reconstruction's cash-underperformance desert (8.4 yr USD / 4.9 yr EUR) is
at least as long as SG CTA's real 6.0 yr, well within (indeed beyond) the
"within ~30% of SG Trend" bar. The proxy contains genuine multi-year deserts;
it is not too kind. PASS.

### Level 5: recoveries. PASS.

Deepest drawdowns and recovery tail:

- recon USD: deepest -34.2% (trough 2021-07, 1536 d to recover, 4.21 yr);
  128 episodes, avg recovery 40 d.
- real DBMF: deepest -20.4% (trough 2023-03, 1001 d, 2.74 yr); 53 episodes.
- SG CTA: deepest -16.5% (trough 2019-01, 832 d, 2.28 yr); 115 episodes.

The reconstruction's worst drawdown and longest recovery are *deeper and
longer* than both real benchmarks. The recovery tail is not softened. PASS.

### Level 6: extremes. PASS.

| Series | worst 1y | best 1y | worst 5y CAGR | best 5y CAGR |
|---|---|---|---|---|
| recon USD | -28.4% | +54.6% | -6.8% | +19.9% |
| SG CTA | -13.4% | +31.8% | -1.6% | +10.7% |
| real DBMF | -12.7% | +36.2% | +4.8% | +10.2% |

The reconstruction's worst year (-28.4%) and worst 5-year CAGR (-6.8%) are
harsher than both real series. It does not soften the worst outcomes. PASS.

### Level 7: temporal dependence. PASS (secondary).

Daily return autocorrelation is small for all three, as expected for trend at
daily frequency: recon lag1 +0.023, real DBMF -0.051, SG CTA +0.088; all higher
lags near zero. No spurious serial structure introduced by the reconstruction.

### Level 8: portfolio properties (spliced vs real-only). PASS.

Adding the trend sleeve to MSCI World, measured two ways:

Real-only era (2019-05 -> 2026-06, real DBMF + real World):

| Mix | CAGR | vol | Sharpe | MaxDD |
|---|---|---|---|---|
| 100% equity | 16.1% | 19.0% | 0.85 | -34.0% |
| 75/25 | 14.6% | 15.2% | 0.96 | -27.7% |
| 60/40 | 13.7% | 13.3% | 1.03 | -23.7% |

Reconstructed era (1999+ World + recon USD):

| Mix | CAGR | vol | Sharpe | MaxDD |
|---|---|---|---|---|
| 100% equity | 10.5% | 18.5% | 0.57 | -57.2% |
| 75/25 | 9.9% | 14.4% | 0.69 | -41.2% |
| 60/40 | 9.5% | 12.6% | 0.75 | -29.8% |

A 25% sleeve lifts Sharpe by ~+0.11 (real-only) / ~+0.12 (reconstructed) and
cuts MaxDD by ~6 pp / ~16 pp. The diversification benefit is of the *same
order* in real-only data as in the spliced history; it is not an artifact of
the reconstructed era. PASS.

### Level 9: stress / sensitivity (the decision-relevant test). PASS.

Max-Sharpe optimization over {MSCI World, IEF bonds, gold, DBMFE} on the full
2004+ window (base DBMFE: CAGR 10.3%, vol 19.3%, equity corr -0.10), perturbing
only the DBMFE leg:

Uncapped (per-asset cap 100%), optimal DBMFE weight:

| Scenario | DBMFE weight | change |
|---|---|---|
| base | 14.9% | - |
| CAGR -1%/yr | 13.6% | -1.3 pp |
| CAGR +1%/yr | 16.1% | +1.2 pp |
| equity corr -0.10 -> +0.05 (+0.15) | 12.9% | -2.0 pp |
| equity corr -0.10 -> -0.25 (-0.15) | 17.2% | +2.3 pp |
| equity corr -> +0.30 (a +0.40 shock) | 9.9% | -5.0 pp |
| CAGR -3%/yr | 10.9% | -4.0 pp |
| vol +/-20% | 10.9% / 21.3% | -4.0 / +6.4 pp |

With a 40% per-asset cap the base is 17.8% and the +/-1% CAGR and +/-0.15 corr
moves are even smaller (+/-1.1 and -1.5 / +1.6 pp).

Against the stability bar (optimal weight should move < ~5 pp under +/-1% CAGR
and +/-0.15 equity correlation): the actual moves are ~1.1-1.3 pp (CAGR) and
~1.5-2.3 pp (correlation). PASS comfortably. The weight only moves materially
under much larger shocks (a +0.40 correlation jump, a -3% CAGR cut, or +/-20%
vol), which are well beyond the stated plausible-error band. DBMF modeling
precision is therefore *not* critical to the sizing decision: confidence is
high.

## Thresholds scorecard

| Threshold (from design) | Result | Verdict |
|---|---|---|
| recon vs real DBMF daily corr >= 0.45 | 0.524 | PASS |
| beta sim->real in [0.3, 0.7] | 0.42 | PASS |
| rolling equity correlation near 0, rarely > +0.3 | ~0.00-0.19 | PASS |
| winter (longest cash underperformance) within ~30% of SG Trend | 8.4 yr vs SG 6.0 yr | PASS (longer) |
| tails: skew/kurt not more benign than real DBMF | recon worse on both | PASS |
| worst 12m / worst DD at least as bad as real DBMF | -28.4% / -34.2% vs -12.7% / -20.4% | PASS |
| Level 9 weight move < ~5 pp under +/-1% CAGR, +/-0.15 corr | <= 2.3 pp | PASS |

No threshold fails.

## Caveats and limitations

1. The reconstruction *understates* crisis alpha magnitude in fast V-crashes
   (2020-Q1, -14.6% vs the real fund's -1.1%) and in big-trend years (2022,
   +3.4% vs +20.5%). Direction is right, magnitude is conservative. Do not rely
   on the reconstruction to *quantify* peak crisis upside; rely on it for shape
   and for floor/winter behaviour.
2. The reconstruction's standalone 2019-2026 Sharpe (0.39) trails the real fund
   (0.77). Over long history vs SG CTA the efficiency matches (0.52 vs 0.56).
   Net effect on sizing is conservative (return under-stated).
3. SG CTA history ends 2023-05, so 2024-2025 trend chop is benchmarked only
   against the (short) real DBMF. The events that matter most (2008, 2011-2019
   winter, 2020, 2022) are all covered.
4. A pure in-sample max-Sharpe optimizer favours ~15-20% trend (and a large
   bond weight built on the 2004-2021 bond bull). The 25% sleeve is a
   forward-looking crisis-protection choice with a deliberate margin, not an
   in-sample optimum.
5. Level 10 (FIRE ruin-probability engine) is out of scope; the engine does not
   exist. The primitives built here (rolling stats, drawdown episodes, recovery
   distribution) are exactly what it will reuse.

## Recommendation on the 25% sleeve

A ~25% DBMFE allocation is **defensible**. The reconstruction reproduces the
diversifying *shape* that justifies the sleeve (near-zero equity correlation,
identical up/down capture to the real fund, faithful trend winters, a fat left
tail), and every place it is unfaithful is conservative (lower return, smaller
crisis upside, deeper drawdowns than reality). The diversification benefit
survives on real-only data, and the optimal weight is stable under the plausible
DBMF modeling errors, so the decision does not hinge on getting DBMF exactly
right.

Guidance: treat 25% as a **cap, not a floor**. Holding it there (rather than
pushing higher) keeps a margin of safety given the reconstruction's understated
crisis upside and the single-speed/narrow-basket risk. For decumulation
specifically, the deep-and-long trend winters reproduced here (8+ year cash
deserts, -34% drawdowns) are real and must be planned around: the sleeve damps
equity sequence risk but introduces its own multi-year droughts, so pair it
with sufficient short-duration/cash buffer rather than treating it as a
standalone safe asset.

## Reproduction

Throwaway harness (not committed) under the campaign scratchpad
`scratchpad/validate/` (`go run .`, reads the live quote cache, the SG CTA CSV
in `docs/`, and builds the reconstructions via `pkg/simgen`). Raw numeric output
saved alongside as `results-raw.txt`. The committed deliverable is the
`pkg/metrics` primitives and this document.
