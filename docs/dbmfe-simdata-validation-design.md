# Design: validating the DBMFE simdata (FIRE protection sleeve)

Date: 2026-06-28

Status: spec for a one-off validation campaign. No implementation started.
This document is a handoff: it carries enough context for a fresh session to
execute without re-reading the whole codebase first.

## Why this matters

DBMFE (iMGP DBi Managed Futures, unhedged EUR share class) is the main
crisis-protection / diversifying asset in the owner's FIRE portfolio, weighted
on the order of 25%. The portfolio runs semi-accumulation now and decumulation
later. If the simulated history that backs DBMFE before its real inception
(2025) overstates its diversification benefit, the optimizer and any future FIRE
engine will recommend a sleeve that looks robust on paper but is not, and a
decumulating portfolio is unforgiving of that error (sequence-of-returns risk
turns an optimistic backtest into a real probability of ruin).

ChatGPT's review (recorded in `/tmp/verif-sim-backdata.md`) lists a 10-level
battery to "try to break" the generator. This spec refines that battery to
pofo's *actual* construction, decides what to build versus what to compute
ad hoc, and sets pass/fail thresholds.

## What the DBMFE simdata actually is (read this first)

This is the single most important framing correction. ChatGPT assumed a
*synthetic data generator* calibrated on aggregate statistics (means,
covariance, volatility) that could "smooth" regimes and tails. pofo does
something different.

Source of truth: `pkg/simgen/recipes.go` (`dbmfeRecipe`, `dbmfeBuild`) and
`pkg/simgen/tsmom.go` (`TSMOM`).

- DBMFE is built by running one **deterministic** 12-month time-series-momentum
  strategy (Moskowitz-Ooi-Pedersen style) over a fixed cross-asset futures
  basket of *real* historical price series, then re-expressing the daily USD
  return as an unhedged EUR return via the real EURUSD spot path. The real
  DBMFE quotes are grafted on top from inception (`Splice` /
  `marketdata.ExtendBack`).
- The basket (`mfMarkets`) is only 7 markets: `VFINX, VTMGX, VEIEX` (US /
  developed / EM equities), `VFITX, VUSTX` (intermediate and long Treasuries),
  `GC=F` (gold), `CL=F` (crude). Cash is `^IRX`. No currencies, no ags, no
  industrial/precious metals beyond gold, no short-rate or non-US bond futures.
- Config (`mfConfig`): single 252-day lookback, 63-day vol window, 21-day
  rebalance, 10% annualized vol target, max 2x per market, 0.85%/yr fee,
  collateral earns cash.
- It is **one path, not an ensemble.** There is no distribution of simulated
  futures. The randomness/regimes in the path come entirely from the real
  underlying markets, not from a stochastic model.

Consequences for the validation plan:

1. The "smoothing of regimes/tails" risk is *lower* than for a parametric
   generator: real 2008, 2020, 2022 are inside the reconstructed window (the
   path starts 2004-10-12, set by EURUSD=X), so real autocorrelation and real
   fat tails of the underlying markets are inherited, not modeled away.
2. The *new* risks specific to this construction are:
   - **Basket narrowness.** Real DBMF / the SG Trend Index trade ~50+ markets
     across all sectors. A 7-market basket can both under-deliver crisis alpha
     and misstate the depth/length of trend "winters".
   - **Single-speed generic signal.** Real funds blend multiple lookbacks and
     add carry; one 12-month signal can sharpen or smooth turning points.
   - **Return shortfall, already visible.** The DBMF (USD parent, 7y real
     overlap) validation in its simdata header reads CAGR sim 4.94% vs real
     9.21%. The proxy under-earns by ~4%/yr. Understated return is conservative
     for "is the sleeve too attractive on return", but the sleeve is justified
     by *shape* (low equity correlation, crisis alpha), so the shape is what
     must be validated.
   - **The graft seam.** A discontinuity or scale error where simulation meets
     real data (2025 for DBMFE, 2019 for DBMF) would corrupt every downstream
     statistic.
   - **FX overlay realism.** EUR-unhedged adds the real EURUSD path (~8-10%/yr
     FX vol) on top of the strategy. Verify the conversion formula and that the
     resulting EUR vol is sane.

Because every level of ChatGPT's plan that says "compare the *distribution* of
simulations" must be reinterpreted: there is no Monte-Carlo ensemble to take a
distribution over. We instead compare the *single historical path's* in-sample
distributions (rolling windows, monthly/annual histograms, drawdown episodes)
against a real benchmark. Where a true ensemble view is wanted (e.g. for the
FIRE engine, Level 10), that is future work and is called out as such.

## The question to falsify

> Does the pre-2025 DBMFE reconstruction faithfully reproduce DBMF's real
> risk *shape*, especially (a) low/zero correlation to global equities, (b)
> positive crisis alpha during equity drawdowns, and (c) realistically long and
> deep trend-following winters, such that a 25% sleeve sized against it is not
> an artifact of a too-regular proxy?

Everything below serves that question. Secondary: quantify how sensitive the
optimal sleeve weight is to plausible errors in DBMF's CAGR / vol / equity
correlation (Level 9), because that tells us how much the answer even matters.

## Ground-truth data sources

The 286-day real DBMFE overlap (corr 0.154) is statistically meaningless. Do
not anchor conclusions on it. Use, in order of usefulness:

1. **Real DBMF (USD parent), 2019->**, daily, fetchable via pofo
   (`marketdata.Client.Fetch("DBMF", ...)`). Same USD strategy as DBMFE before
   FX. Best available daily real benchmark; covers 2020 and 2022 crisis alpha.
2. **SG CTA Index** (Societe Generale), daily, 1999-12-31 -> 2023-05-09,
   AVAILABLE in the repo at
   `docs/SG-CTA-Index-Daily-Returns-since-1999-12-31.csv`. The industry-standard
   managed-futures benchmark and the right yardstick for the long-winter /
   regime questions. Caveats the executing session must handle:
   - Format is `date,daily_return` (simple returns), NO header, NOT pofo's
     simdata format. Cumulate to a base-100 level series before feeding it to
     any value-series tool (`level[0]=100; level[k]=level[k-1]*(1+r[k])`).
     Optionally emit a simdata-format CSV (`# pofo simdata v1` header,
     `date,close`) into a `-refdata` dir so it loads via `marketdata.ReadSimdata`.
   - It is **USD**. Compare it against the **USD layer** of the construction
     (real DBMF, and/or the pre-FX TSMOM index), NOT against the EUR-converted
     DBMFE: that keeps the trend-shape validation clean and isolates the EURUSD
     overlay as its own check (pofo-specific check 4). For an EUR view, convert
     SG CTA to EUR with the same EURUSD path, but treat that as secondary.
   - It is the broad **SG CTA** index (trend plus some non-trend managers),
     marginally broader than the pure SG Trend sub-index; fine as the benchmark,
     just note the nuance.
   - It **ends 2023-05-09**, so 2024-2025 (incl. the 2023-24 trend chop tail and
     2025) are not covered. The long-winter (2011-2019), 2008, 2020 and 2022
     events that matter most ARE covered.
3. **Real KMLM (2020->) and CTA (2022->)** as secondary real managed-futures
   reals for triangulation (already bundled as recipes).
4. Global equity benchmark for correlation / crisis-alpha tests: **URTH** or
   **VT** (real, fetchable), in EUR terms where the comparison is for the EUR
   investor. Gold (`GC=F`) and Treasuries (`IEF`/`TLT`) for the other
   correlation legs ChatGPT lists.

How to feed a non-fetchable benchmark to the tools: `pkg/simgen.WithRefData`
wraps a `Fetcher` to serve CSVs (simdata format) from a directory before falling
back to the network; the CLI exposes it via `-refdata`. A throwaway script can
read any benchmark CSV the same way with `marketdata.ReadSimdata`.

## Inflation series for Level 4

Level 4 measures the longest stretch where DBMFE's trailing CAGR falls below an
inflation floor (real-return drought). The floor is **HICP France** (IPCH),
which is the relevant cost-of-living measure for a French FIRE investor.

Data: Eurostat `prc_hicp_midx`, France (geo=FR), monthly index base 2015=100,
public and free with no API key, available from **1996-01** so 2000-01-01 is
comfortably covered. (INSEE's national IPC via the BDM API is a defensible
alternative; HICP is chosen for EU comparability and a one-line CSV export.)

Granularity caveat: price indices are **monthly**; there is no realized daily
inflation (daily inflation swaps/breakevens are market *expectations*, not
realized). "Daily inflation" is therefore an interpolation. Use the **geometric
accrual** that matches pofo's existing rate handling (`BuildFrame` already turns
annualized rate levels into daily accruals): spread each month's inflation evenly
across its calendar days,

    daily_factor = (HICP[m] / HICP[m-1]) ^ (1 / days_in_month)

producing a smooth daily deflator that compounds cleanly against the assets'
daily returns (no month-boundary steps).

IMPLEMENTED (2026-06-28): `^HICP-FR` is now a first-class fetchable series
(`pkg/marketdata/eurostat.go`): `Client.Fetch("^HICP-FR", from)` returns the
daily-interpolated HICP France index (1996->), cached like every other source,
currency-less. Other geographies follow the `^HICP-<geo>` pattern (e.g.
`^HICP-EA`). Level 4 should fetch it directly; no bundled CSV needed. The same
series doubles as the real-return deflator the future FIRE engine will reuse.

(Historical note on what was considered: a tier-1 throwaway bundled CSV was the
fallback if a fetcher proved too costly; it turned out cheap enough to build the
proper fetcher straight away.)

## Execution policy: build vs. throwaway vs. by-hand

Per the owner's instruction, this is a one-off, so default to the cheapest
faithful method. Three modes:

- **L (library):** contribute a small, general, documented primitive to pofo
  (mostly `pkg/metrics`). Justified only when the primitive is generically
  useful to pofo beyond this campaign AND simplifies several checks here. These
  are the higher-moments, rolling-stat, drawdown-episode, autocorrelation and
  longest-underperformance gaps that pofo's reports/optimizer would benefit from
  anyway.
- **S (script):** a throwaway program under `scratchpad/` (or a `cmd/`-style
  one-file main kept out of the build) that wires pofo primitives + the chosen
  benchmark and prints a table/CSV. Not committed to the library surface.
- **C (executing model):** computed and judged directly by the executing model from CSVs
  / tool output, no committed code, when the check is a one-line stat or a
  visual/qualitative read.

The split below is deliberate: build the primitives that are reusable and that
several levels share; script the orchestration and the perturbation studies;
hand-judge the qualitative "do the curves look too pretty" reads.

## Proposed library contributions (pkg/metrics, one small chart helper)

These fill real gaps in pofo and each serves multiple validation levels. Keep
them in pofo's idiom: pure functions over `dates []time.Time, values []float64`
or over `[]float64` returns, 252-day annualization, rf=0, consistent with
`metrics.Compute`. Each ships with godoc and a table test. TDD per the repo's
skills.

1. **Higher moments** (`metrics`): add `Skewness` and `ExcessKurtosis` over a
   return slice, and surface them on `Stats` (new fields `Skew`, `Kurtosis`).
   Serves Level 1. ~15 lines + test. Clear keeper; reports should show them.

2. **Generic rolling window** (`metrics`): `Rolling(dates, values, years,
   fn func(window []float64) float64) (points []time.Time, out []float64, ok
   bool)` applying `fn` to each calendar-length window. Refactor the existing
   `RollingCAGR` onto it, and provide `RollingSharpe`, `RollingSortino`,
   `RollingUlcer`, `RollingVol` as thin wrappers. Serves Level 3 (the core of
   the "too pretty" test) and Level 6 (worst rolling 5y/10y). This is the
   highest-leverage contribution.

3. **Drawdown episodes** (`metrics`): `DrawdownEpisodes(dates, values)
   []Episode` where `Episode{PeakDate, TroughDate, RecoverDate, Depth,
   DrawdownDays, RecoveryDays, Ongoing}`. Today `Compute` only exposes the
   single longest underwater stretch; the full episode list is needed for the
   recovery-time *distribution* (Level 5) and the drawdown-depth distribution
   (Level 2), and pofo's report would benefit from a "deepest N drawdowns"
   table. Keep `Compute` as-is; this is additive.

4. **Autocorrelation** (`metrics`): `Autocorr(xs []float64, lags int) []float64`
   (lag-0 = 1). Serves Level 7 (return autocorrelation; also run on the running
   drawdown series). Small and generic.

5. **Longest underperformance** (`metrics`, relative): `LongestUnderperformance(
   dates, values, benchDates, benchValues, window years) (maxConsecutive
   ...)` measuring the longest stretch where the series' trailing-window CAGR is
   below the benchmark's (or below a constant floor, for the cash/inflation
   variants). Serves Level 4. Generalizes naturally from `VsBenchmark`.

6. **Histogram** (`metrics`): `Histogram(xs []float64, bins int) (edges,
   counts)` plus a tiny ASCII renderer (or reuse `pkg/chart`). Serves Level 2
   (monthly/annual return histograms). Borderline keeper; if it feels like
   over-build, downgrade to S and bucket inline. Recommendation: build the pure
   `Histogram` (trivial, testable), render ad hoc. QQ-plots stay S/C (no library
   value): emit sorted quantile pairs to CSV and eyeball, or one throwaway
   gnuplot/python plot.

Not built (explicitly): any Monte-Carlo ensemble engine, any FIRE
ruin-probability engine (Level 10). Those are real features, not validation
scaffolding, and belong to their own designs.

## The levels, refined and assigned

For each: what it actually tests here, the data, and the execution mode.

### Level 1. Base statistics  -> L + C
Compute CAGR, vol, Sharpe, Sortino, Ulcer, MaxDD, Skew, Kurtosis and
correlations vs equities / bonds / gold, for: (a) the DBMFE *reconstructed*
segment, (b) real DBMF, (c) SG Trend (if available). `metrics.Compute` already
covers most; add Skew/Kurtosis (contribution 1). Correlations via existing
`metrics.Beta` / `VsBenchmark` plus a plain Pearson on aligned returns.
Judge: are the reconstruction's moments in the same ballpark as real DBMF and
SG Trend? A Sharpe that is too high or a kurtosis that is too low is the first
smell. Necessary, not sufficient (ChatGPT is right).

### Level 2. Distributions  -> L + S
Monthly and annual return histograms; drawdown-depth distribution; QQ-plot of
reconstruction vs real DBMF monthly returns. Use contributions 3 (episodes) and
6 (histogram). Script aligns the series and emits histograms + a QQ CSV.
Judge: shape proximity, especially the left tail. A reconstruction whose monthly
distribution is visibly thinner-tailed than real DBMF is a red flag.

### Level 3. Temporal validation (most important per ChatGPT)  -> L + C
Rolling 10y (and 3y/5y) Sharpe, Sortino, CAGR, Ulcer on the reconstruction and
on real DBMF / SG Trend. Use contribution 2. Overlay and read by hand.
Judge: does the reconstruction ever sit in a long mediocre stretch, or is every
rolling curve "too pretty"? Specifically check the 2011-2019 trend winter,
which must show up as a depressed rolling Sharpe. If the proxy is smooth where
SG Trend was miserable, the sleeve is overstated. THIS is the falsification core.

### Level 4. Regime validation: longest underperformance  -> L + C
Longest consecutive span where trailing CAGR < cash (`^IRX`), < HICP France
inflation (fetch `^HICP-FR`, see "Inflation series for Level 4"), and < MSCI
World (URTH). Use contribution 5. Compare reconstruction vs real DBMF / SG CTA.
Judge: the reconstruction must contain multi-year "deserts". If its longest
underperformance is materially shorter than SG Trend's real one, it is too kind.

### Level 5. Recoveries  -> L + C
Longest recovery, average recovery, full recovery-time distribution, from
contribution 3 (episodes). Compare reconstruction vs real DBMF.
Judge: match the *tail* of recovery times, not just the mean.

### Level 6. Extremes  -> C (uses L)
Worst/best year, worst/best 5y and 10y CAGR (rolling min/max from contribution
2), worst decade. Pure reporting off the rolling primitive.
Judge: the proxy must not soften the worst outcomes relative to real DBMF /
SG Trend.

### Level 7. Temporal dependence  -> L + C
Return autocorrelation (lags 1..~20d and monthly), autocorrelation of the
running drawdown series, mean length of positive vs negative runs.
Use contribution 4. Trend strategies carry distinctive serial structure;
confirm the reconstruction has it and roughly matches real DBMF.

### Level 8. Portfolio properties  -> C + S (no new library code)
Build 100% World, 80/20 and 60/40 World/DBMFE and compare CAGR, vol, Sharpe,
MaxDD, Ulcer, correlation. pofo already simulates and reports portfolios
(`pkg/portfolio/sim.go`, `pkg/report`, `pkg/optimize`); create example portfolio
files (see `examples/*.txt`, e.g. `dragon-portfolio-artemis.txt`,
`alt-dragon.txt`) and run the CLI, OR a short script. Also run each portfolio
twice: once on the *spliced* history (with simdata) and once on *real-only*
history, to isolate how much of the diversification benefit comes from the
reconstructed segment.
Judge: the diversification gain from DBMFE must be of the same order whether
measured on real-only or spliced data; a benefit that exists *only* in the
reconstructed era is the exact artifact we fear.

### Level 9. Stress tests (most decision-relevant)  -> S
Perturb the DBMFE series and re-optimize: CAGR shift +/-1% and +/-20%
(`(1+r) -> (1+r-delta_daily)`), vol scale +/-20% (scale demeaned returns),
equity correlation +/-0.2 (blend a small fraction of URTH returns into DBMFE to
raise correlation, or orthogonalize to lower it). Re-run `pkg/optimize` and
record how the optimal DBMFE weight and the portfolio's Sharpe/MaxDD move.
Throwaway script; possibly a tiny reusable "perturb returns" helper if it reads
cleanly, but keep it out of the committed surface unless obviously general.
Judge (ChatGPT's key test): if a small, plausible error (e.g. equity
correlation 0 -> 0.15, or CAGR -1%) collapses or explodes the optimal weight,
then DBMF modeling precision is *critical* and conclusions must be hedged. If
the weight is stable, confidence rises a lot. This result should headline the
final report.

### Level 10. FIRE engine  -> deferred (separate future spec)
Probability of ruin, terminal wealth, 5th-percentile terminal wealth, years
underwater, failure rate, on spliced vs real-only. The FIRE engine does not
exist yet and is the subject of its own future spec (decumulation / withdrawal
ruin risk over multiple parameters). Out of scope here.

Note on Monte-Carlo, to avoid confusion with the framing above: the simdata
*generation* layer is and stays a single deterministic path (no ensemble). The
FIRE engine is a *consumption*-layer tool and is exactly where Monte-Carlo
belongs: it will draw many decumulation paths (historical bootstrap and/or
parametric) from the validated return series to estimate ruin probability. The
two are different layers; this validation campaign concerns only the generation
layer's faithfulness. Contributions 2-5 (rolling stats, drawdown episodes,
recovery distribution) are exactly the primitives that engine will reuse, so
building them now is not wasted.

## pofo-specific checks ChatGPT could not know to ask  -> S + C (do these too)

These are arguably higher priority than re-running the generic battery.

1. **Graft-seam continuity.** Inspect the return distribution in a window around
   each splice date (DBMFE 2025-03/04; for cross-checks, DBMF 2019). Confirm no
   abnormal jump and that `ExtendBack`'s rescaling is correct. A one-off plot +
   the largest daily returns near the seam. (S/C)
2. **Reconstruction-vs-real overlap on the parent.** Re-run `simgen.Validate`
   conceptually on DBMF (7y real) rather than DBMFE (286d): the DBMF header
   already shows corr 0.52 / beta 0.42 / CAGR 4.94 vs 9.21. Decompose the 4.3%
   CAGR shortfall: fee (0.85%), basket narrowness, vol target, missing carry.
   This tells us whether the *pre-2019 reconstructed* return is pessimistic
   (safe) and whether the *shape* (corr/beta) is right (what matters). (C, with
   small S helpers)
3. **Crisis-alpha events.** Zoom on 2008-H2, 2020-Q1, 2022 (the big trend year):
   does the reconstruction deliver positive return while URTH falls? Up/down
   capture vs URTH via `metrics.VsBenchmark`. This is the literal reason the
   sleeve exists. (C, uses L)
4. **FX overlay sanity.** Verify `rEUR = (1+rUSD)/(1+rFX)-1` against a hand
   example, and confirm DBMFE vol exceeds DBMF (USD) vol by roughly the EURUSD
   contribution. Confirm the unhedged EUR class really should carry that FX
   risk (it should). (C)

## Pass / fail thresholds (OPEN, propose then confirm with operator)

Suggested defaults; the executing session should propose these and the operator
confirms before drawing conclusions. Benchmark for the winter/tail bars is SG
CTA (USD), compared against the USD layer (real DBMF / pre-FX TSMOM):

- Shape: reconstruction vs real DBMF daily corr >= 0.45, beta in [0.3, 0.7],
  rolling 12m equity correlation centered near 0 and rarely above +0.3.
- Winters: longest underperformance vs cash within ~30% of SG Trend's (or, if
  SG unavailable, not shorter than real DBMF's observed worst stretch).
- Tails: reconstruction monthly skew/kurtosis not *more* benign than real DBMF;
  worst 12m and worst drawdown at least as bad as real DBMF's.
- Stability (Level 9, the decisive one): optimal DBMFE weight changes by less
  than ~5 percentage points under +/-1% CAGR and +/-0.15 equity correlation.
  If it does not, downgrade confidence and recommend a smaller / capped sleeve.

A FAIL on winters/tails or on stability is the actionable outcome: it would
argue for sizing the sleeve below 25% or adding a margin of safety, which is the
whole point of doing this.

## Open questions for the operator

1. RESOLVED: SG CTA daily history is in the repo
   (`docs/SG-CTA-Index-Daily-Returns-since-1999-12-31.csv`, USD, 1999->2023-05).
   Used as the winter/regime benchmark against the USD layer. See ground-truth
   source 2 for the format caveats.
2. RESOLVED: inflation floor for Level 4 is **HICP France** (IPCH), Eurostat
   series `prc_hicp_midx` (base 2015=100, monthly, 1996->, free, no API key).
   See "Inflation series for Level 4" below for the tiered approach.
3. Confirm the pass/fail thresholds above (especially the Level 9 stability bar,
   which drives the final recommendation).
4. Currency basis for Level 8 portfolios: EUR investor throughout (convert
   equity/bond legs to EUR), correct given the FIRE context. Confirm.

## Suggested execution order (for the handoff session)

1. pofo-specific checks 1-2 (seam + DBMF shortfall decomposition): cheapest, and
   if the seam is broken or the shape is wrong, stop and fix the recipe first.
2. Library contributions 1-5 with TDD (they unblock Levels 1-7).
3. Levels 1, 3, 4, 5, 7 against real DBMF and SG CTA (USD layer; cumulate the SG
   CTA returns CSV to a base-100 level first).
4. Crisis-alpha check 3 + Level 8 (real-only vs spliced).
5. Level 9 stress/sensitivity study (headline result).
6. Write findings back into a `dbmfe-simdata-validation-results.md`, and if any
   threshold fails, a recommendation on resizing the sleeve. Update the
   `# validation:` header line of `pkg/datasets/simdata/DBMFE.csv` if the work
   produces a better one-line summary.

## Key files

- `pkg/simgen/recipes.go` - `dbmfeRecipe`, `dbmfeBuild`, `mfMarkets`, `mfConfig`.
- `pkg/simgen/tsmom.go` - the TSMOM engine and `rollingVol`.
- `pkg/simgen/simgen.go` - `Frame`, `BuildFrame`, `Validate`, `Splice`,
  `WithRefData`.
- `pkg/metrics/metrics.go` - `Stats`, `Compute`, `Returns`, `Mean`, `Beta`.
- `pkg/metrics/relative.go` - `Drawdowns`, `RollingCAGR`, `VsBenchmark`.
- `pkg/datasets/simdata/DBMFE.csv`, `DBMF.csv` - the generated paths + headers.
- `cmd/pofo/main.go` - `-gen-simdata`, `-refdata`, `-simdata` wiring;
  `genOne` shows the build/validate/splice flow.
- `examples/*.txt` - portfolio file format for Level 8.
- `docs/SG-CTA-Index-Daily-Returns-since-1999-12-31.csv` - SG CTA benchmark
  (USD daily returns, no header; cumulate to base-100 before use).
