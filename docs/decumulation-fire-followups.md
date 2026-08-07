# Decumulation / FIRE — follow-ups backlog

Running list of oddities to fix and enhancements to evaluate, from the fresh
`pkg/scenario` + `pkg/decumul` + `pkg/decumul/web` implementation (designs:
`decumulation-fire-design.md`, `decumulation-monthly-sampling-design.md`).
Not yet scheduled; each item should get its own brainstorm → spec → plan when
picked up. Priority: **P1** correctness, **P2** clarity/API, **P3** features.

> **Realism / conservatism** of the ruin figures (the model is optimistic vs the
> broad-sample SWR evidence) has its own spec:
> [`decumulation-fire-realism-spec.md`](decumulation-fire-realism-spec.md)
> (default flex off, honest return defaults + conservative prior, sequence-risk
> capture, longevity horizon).
>
> **Usability rewrite (2026-06-30):** the explorer is being rebuilt around a
> multi-model comparison, visible simulated markets, and an actionable safe-WR
> headline. Spec: [`decumulation-fire-rewrite-spec.md`](decumulation-fire-rewrite-spec.md)
> (synthesis of Ben's brief, a ChatGPT thread, and Claude's analysis in
> [`decumulation-fire-usability-proposals.md`](decumulation-fire-usability-proposals.md)).
> The regime mean-tanking bug it identified is already fixed (commit d96e4a1).
> Deferred to phase 2: bundle a broad-sample century-long real-return panel
> under `pkg/datasets/` so the historical models see 1900s-2020s regimes.

## Correctness oddities (P1)

1. ✅ **Done (2026-06-29).** **Ruin is tested on net need, not the grossed-up
   withdrawal.**
   `Plan.RunPath` flags ruin with `if need > total` (net), but the money that
   must actually be liquidated is the *gross* (need + tax, and buffer refills).
   With the cost-basis `CTOFlatTax`, gross can substantially exceed net at high
   embedded gains, so ruin can be under-flagged. The golden tests pass because
   they use a flat-12% stub and a no-buffer plan, which masks it. Fix: define
   ruin as "the gross required to deliver `need` exceeds available liquidity",
   computed consistently across the buffer-first and growth-first branches.

2. ✅ **Done (2026-06-29, with item 1).** **Partial-funding is silent.** `Tax.GrossUp` caps `gross` at `growth`, so a
   year that cannot fully fund `need` sells everything and delivers *less* net
   than needed, yet `RunPath` still does `Withdrawn += need` and may not latch
   ruin. The household "didn't get its money" without the path being marked.
   Fix together with (1): detect under-delivery (buffer drain + net deliverable
   from growth < need) and latch ruin; account the real net withdrawn.

3. ✅ **Done (2026-06-29).** **`0` is overloaded as "unset" in `BufferSleeve`.** `DrawThreshold == 0`
   and `RefillCap == 0` are treated as "use the default" (0.10 / 0.50), so a
   user cannot intentionally set a 0 threshold or a no-refill policy. Use
   explicit defaults at construction, sentinel `-1`, or pointer fields.

4. ✅ **Done (2026-06-29).** **`Sweep2D(_, Mu, …)` on a non-parametric plan is a silent no-op.**
   `Plan.set(Mu, …)` only mutates a `ParametricSource`; for bootstrap/cohort
   sources it does nothing, so the surface comes out flat with no error. The
   web layer dodges this by switching the y-axis, but the library API should
   either return an error or document the constraint loudly.

5. ✅ **Done (2026-06-29).** **`worst10y` uses a `-1` sentinel** to mean "hit
   zero within the window", then aggregates by `min`, so a single ruined path
   drives `Worst10yCAGR` to −100%. Arguably correct but the sentinel conflates
   "−100% realised" with "undefined"; revisit the representation (e.g. report
   separately the share of paths with a sub-`-x%` decade).
   Fixed: `worst10y` now returns `(cagr, ok)`, treating a decade that *ends* at
   zero as its realised −100% and *skipping* windows that *start* after ruin
   (the conflated case); `Outcome` keeps the honest min `Worst10yCAGR` and adds
   a robust `Worst10yP5` (5th-percentile of paths' worst decade) so one ruined
   path no longer defines the headline.

## Clarity / API (P2)

6. ✅ **Obsolete (2026-06-29).** **2D surface y-axis silently changes meaning**
   (real CAGR for parametric, spending floor for the historical models). The
   heatmap title says "scenario axis (y)" generically; label the axis
   dynamically and state the unit. The web heatmap was dropped in commit
   `fd5a75b`; `chart.Heatmap`/`Sweep2D` are no longer rendered anywhere, and the
   Mu-on-non-parametric ambiguity is now a loud library error (item 4). No
   surface axis remains to label.

7. ✅ **Done (2026-06-29).** **Monthly→annual compounding is
   rolling-from-window-start, not calendar years.** For cohorts a 15y window
   compounds months `[0..11], [12..23], …` from the cohort start, so the
   "annual" returns are not Jan–Dec. Statistically fine; document it so it is
   not mistaken for a bug. Documented on `scenario.Annualize`.

8. ✅ **Done (2026-07-01, obsolete).** **Common-window alignment truncates to
   the last N months/returns** assuming the month grids line up across holdings.
   Resolved by the monthly-panel migration (item 13): the only panel builder now
   is `web.BuildMonthlyPanel`, which aligns on the intersection of dense month
   keys (`monthKey`), keeps only calendar-consecutive one-month returns, and
   drops returns that span a gap. There is no positional-tail aligner left; a
   holding with internal gaps stays column-aligned. Nothing to do.

9. ✅ **Done (2026-06-29).** **`FitParametric` does not fit `df`** and estimates
   annual sigma from ~20 annualised points (noisy). Consider deriving annual vol
   from monthly returns (`σ_m·√12`) for stability, and optionally fitting `df`
   from the sample kurtosis to seed the fat-tail slider from history.
   Done: `FitParametric` now returns a `Fit{Mu, Sigma, Df}`; sigma is the
   monthly std × √12, and df is seeded from the monthly excess kurtosis
   (inverting the Student-t 6/(df−4)), clamped to the slider range and exposed
   via `/api/meta` + `/api/fit` so the df slider is seeded from history too.

## Features / "most useful FIRE info" (P3)

10. ✅ **Done (2026-06-29).** **Surface the computed-but-hidden metrics.**
    `Outcome` already produces median years underwater, worst-10y real CAGR and
    CDaR (and the recovery histogram is shown); add cards/panels for the rest,
    plus median cumulative tax and effective tax rate. Done: `Outcome` gained
    `MedianCumTax` and `EffectiveTaxRate`, and the web UI now shows cards for
    median years underwater, worst-10y real CAGR (both the robust p5 and the
    min), CDaR, median cumulative tax and the effective tax rate.

11. ✅ **Done (2026-06-29).** **Allocation A/B comparison.** Pin a baseline
    allocation and show a variant side by side (the original motivating
    question: "60/25/15 vs 20/20/…"), rather than only re-dragging one set of
    weights. Done: added `Compare` + `/api/compare`, each allocation re-fitted
    from the panel for a fair test; a "Pin allocation as baseline" button toggles
    A/B mode, rendering baseline vs variant cards side by side.

12. ✅ **Done (2026-06-29).** **Solve in the UI.** Expose `CapitalForRuin` and a
    buffer optimiser: given a target ruin %, show the required capital and the
    ruin-minimising buffer. Done: added `Plan.BestBuffer` (argmin ruin over
    candidate buffer years), a `/api/solve` endpoint returning the required
    capital (`CapitalForRuin`) and the ruin-minimising buffer, and a Solve panel
    in the UI (target-ruin input + button + result line).

13. ✅ **Done (2026-06-29).** Implemented `RunPathMonthly` (steps monthly,
    withdraws NeedAnnual/12, evaluates drawdown/flex/bucket monthly, applies one
    monthly real return; buffer years / horizon / cashflow years stay
    year-valued; buffer return applied as its 12th root; Wealth reported at
    annual granularity). A `Plan.Monthly` flag + `runPath` dispatcher route
    `Simulate`/sweeps through it while the annual `RunPath` stays the validated
    golden reference. The web exposes a "Monthly withdrawals" toggle: historical
    models feed the monthly source directly (no Compounded wrapper), parametric
    draws monthly i.i.d. shocks that compound to the annual mu/sigma. Below is
    the original note.

    **Monthly withdrawal kernel (stated requirement, P2-ish).** Ben's real
    use case is a **monthly** withdrawal, like a salary, and the buffer-vs-cut
    re-evaluation ("tap the buffer or cut this month's spend 25%?") is also a
    **monthly** decision. So the kernel should step monthly: withdraw
    NeedAnnual/12 each month, evaluate drawdown/flex/bucket monthly, apply one
    monthly real return per step. Crucially, **the durations that are naturally
    in years stay in years**: buffer size (years × annual spend), life horizon,
    and years-before-retirement are still year-valued inputs. This pairs
    naturally with the monthly return panel already built (a monthly Source
    feeds the kernel directly, no Compounded wrapper for this path). Keep the
    annual kernel + its golden tests as the validated reference; the monthly
    kernel needs its own validation targets. Bigger change, but it is what the
    real plan needs, so prioritise above the other P3 items.

14. ✅ **Done (2026-06-29).** **Richer policies.** Melting/glidepath buffer (stop
    refilling after the sequence-risk window), a distinct inflation-linked sleeve
    vs pure cash, side income (rental/activity) as another `Cashflow`, and a
    guardrails withdrawal rule (Guyton-Klinger style) beyond the single flex cut.
    Done: `BufferSleeve.RefillStopYear` (glidepath), `Cashflow.ToYear` (bounded
    side income), `Plan.Guard` Guyton-Klinger guardrails (yearly spending band on
    the withdrawal rate, replacing flex when set) — all in both kernels and
    exposed in the web (side-income + glidepath sliders, a guardrails toggle).
    The inflation-linked-vs-cash distinction is the existing `RealReturn` /
    "buffer real return" knob (≈0 for a linker, negative for cash), so no
    separate sleeve was needed.

15. ✅ **Done (2026-06-29).** **Shareable scenarios.** URL-encode the
    slider/allocation state so a configuration can be bookmarked or shared (the
    server is local, so this is cheap). Done: the slider values, return model and
    allocation weights round-trip through the URL hash; `run()` writes the hash
    after each compute and the page applies a shared hash on load (shared
    mu/sigma/df override the historical seed).

16. ✅ **Done (2026-06-29).** **Performance.** Each `/api/sim` runs Sweep1D + a
    full Simulate + Sweep2D, i.e. many independent Monte-Carlo passes per slider
    drag. Share pre-drawn paths across the sweep evaluations (as
    `CapitalForRuin` already does) to cut the per-request cost, especially at
    higher path counts. Done: split `Simulate` into `drawPaths` + `simulateOn`;
    `Sweep1D` (for every parameter but Mu, which rebuilds the Source) and
    `CapitalForRuin` now draw the paths once and reuse them. Behaviour is
    byte-identical, locked by a regression test comparing the shared-path sweep
    to a per-value `Simulate`. (Sweep2D is no longer run; the heatmap was
    dropped, see item 6.)

17. ✅ **Done (2026-07-02), as part of the v3 enrichment.** Full
    frontend-design pass on the new UI: the page is now a dense two-column
    desk (sticky grouped control rail + numbered analysis column), with a
    sticky "plan bar" echoing the live verdict and per-model ruin beads while
    scrolled, preset chips (pension scenarios), pervasive hover
    help on every control, and the shared warm-study theme throughout. See
    `decumulation-fire-v3-enrichment.md` for the whole feature drop (spending
    reality, mortality view, planning curves, envelopes, written rules…).

18. ✅ **Done (2026-06-29).** **Better charts for the buffer arbitrage and the
    recovery distribution.** Specific chart requirements (part of, or before,
    item 17):
    - Replace the two separate bar charts with a SINGLE dual-axis line chart:
      buffer years on x, ruin % on the LEFT y-axis, median terminal wealth on
      the RIGHT y-axis. Done via a new `chart.LineDual` primitive; the web shows
      one "Buffer arbitrage" chart instead of two bar charts.
    - The recovery-time distribution bars need readable numbers: add y-axis
      gridlines/ticks with labels and value labels on the bars. Done: `chart.Bars`
      now draws y-axis gridlines + tick labels and an optional per-bar value
      label (`Bar.Text`); the recovery chart passes shares as % with "NN%"
      labels.

## Portfolio analysis / report (not FIRE-specific)

17. **Volatility term structure in the comparison table (approved direction).**
    ✅ **Done (2026-06-29):** added `metrics.VarianceRatio` (Lo-MacKinlay) returning
    a `VolTermStructure` (daily vol, monthly-annualised vol, the ratio, sample
    size), and surfaced two new rows in the comparison table ("Volatility
    (monthly, annualised)" and "Variance ratio (monthly/daily)") with an
    explanatory footnote covering the interpretation and the small-sample caveat.
    ✅ **Monthly-Sharpe reuse done (2026-07-01):** `VolTermStructure` gained
    `MonthlySharpe`/`MonthlySortino` (annualised, rf 0, from the same month-end
    returns as the variance ratio, matching `Stats`' convention), surfaced as
    "Sharpe (monthly)" and "Sortino (monthly)" rows next to the daily ones. The
    FIRE-seeding reconcile is effectively closed too: `FitParametric.Sigma` is
    already `σ_m·√12` (item 9), i.e. the monthly-horizon dispersion, the same
    quantity as the report's `MonthlyVol`, so both now express risk at the
    monthly horizon rather than the daily-annualised one.
    The report currently ranks portfolios by **daily-annualised** volatility
    (and Sharpe/Sortino built on it), which over/understates the dispersion an
    investor actually realises at a multi-year horizon: it overstates when
    returns mean-revert (intraday/daily noise that never compounds) and
    understates when they trend (e.g. managed-futures sleeves). This biases the
    risk ranking and the daily-based Sharpe.
    - Add a reusable primitive **`metrics.VarianceRatio`** (Lo–MacKinlay): the
      ratio of a lower-frequency annualised variance to the daily one,
      e.g. monthly/daily. ≈1 → i.i.d.; <1 → mean reversion (daily vol
      overstates real risk); >1 → trending (it understates).
    - Surface in the comparison table: a **monthly annualised volatility**
      column **plus the variance ratio**, with an **explanatory legend/footnote**
      (what the ratio means, the small-sample caveat: weekly ~1000 pts and
      monthly ~240 are fine, annual ~20 is too noisy to show as a point
      estimate).
    - This is a recognised statistic (volatility term structure / variance
      ratio), not an ad-hoc home metric. Position it as **complementary** to the
      existing rolling-CAGR / drawdown / Ulcer / TTR metrics (which already
      capture long-horizon pain): the ratio specifically reveals the
      autocorrelation those do not show directly.
    - The same primitive would let the FIRE tool reconcile the report's daily
      vol with the annual sigma it seeds (see P2 item 9), and could feed a
      monthly-based Sharpe/Sortino variant. Note that `VarianceRatio` belongs in
      `pkg/metrics` (reusable), consumed by both the report and the FIRE seeding.

## Data history & performance (2026-07-01)

Work on extending the simulated histories back for a real 45-year FIRE backcast,
and on the currency conversion and fetch performance a EUR investor needs.

**Done:**
- **MSCI World real total return to 1969** for IWDA/URTH: Yahoo's MSCI symbols
  (`^990300-USD-STRD` etc.) return nothing to the client and MSCI's free tool
  caps at 1997, so the real monthly series (USD, 1969→, MSCI via a Curvo export)
  is embedded at `pkg/datasets/refdata/MSCIWORLD-USD.csv` (`go:embed` via
  `datasets.Refdata()`, layered into `-gen-simdata` automatically). The recipes
  use `simgen.longIndexOr` (net of TER) with the VFINX+VTMGX proxy blend as a
  fallback when the file is absent. Regenerate: `pofo -gen-simdata IE00B4L5Y983
  URTH && make`.
- **French CPI to 1955**: `^HICP-FR` is extended back with the OECD French CPI
  from FRED (`FRACPIALLMINMEI`), chained at the Eurostat overlap
  (`extendMonthlyBack`), best-effort/cached, short FRED timeout.
- **Gold real spot to 1968**: the `XAUUSD` recipe now splices the bundled
  monthly London/LBMA gold fix (`pkg/datasets/refdata/XAUUSD-LBMA.csv`, 1968→,
  from datahub `core/gold-prices`) behind the fetchable daily spot (`xauusdBuild`),
  and `longBack["GC=F"]` points at it too. Splice validated at the 2000 overlap
  (LBMA 2000-08 = $274.47 vs the daily quote $273.90). If the daily fetch fails,
  the monthly fix stands alone.
- **EUR/USD real cross to 1978**: `EURUSD=X`/`USDEUR=X` are extended back by a
  bundled monthly ECU/EUR proxy (`pkg/marketdata/data/eurusd-long.csv`: FRED
  `EXUSEC` ECU 1978-12→1998-12 chained 1:1 to `EXUSEU` euro 1999→), spliced in
  `Client.History` via `extendFXBack`. Benefits both `ConvertCurrency` (the
  convert-at-end path) and `dbmfeBuild`.
- **Perf**: the EUR/USD FX cross is fetched once per run (constant cache key,
  Yahoo). FRED was removed from the *live* FX path: it failed/stalled per USD
  asset and made runs take >1 min. FRED remains only for the (cached) French CPI;
  the euro long history is now a bundled snapshot, not a live fetch.

- **Crude to 1946 + Treasury TR to 1953** (done): `longBack["CL=F"]` → bundled
  monthly WTI spot (`WTI-USD.csv`, FRED WTISPLC). `longBack["VFITX"]`/`["VUSTX"]`
  → a constant-maturity par-bond total-return reconstruction (`simgen.TreasuryTR`,
  exact monthly repricing) bundled at `TREASURY-INT-USD.csv` (GS5, 5y) and
  `TREASURY-LONG-USD.csv` (GS20, 20y), 1953→. Stats: intermediate 5.12%/yr @
  4.3% vol, long 5.04%/yr @ 11.3% vol.
- **The real NTSG/DBMF cap was the intl-equity legs** (done). `BuildFrame` starts
  the frame at its YOUNGEST leg (`start = max` of first quotes), so extending the
  others is invisible. VTMGX (dev-ex-US) and VEIEX (EM) were capped at 1999/1994
  because their proxies were dead Yahoo MSCI symbols (`^990300-USD-STRD`,
  `^891800-USD-STRD`, which return no usable history). Fixed with bundled series:
  `DEVEXUS-USD.csv` (Ken French dev-ex-US 1990→, MSCI World before, ~1969) →
  `longBack["VTMGX"]`; `EM-USD.csv` (Ken French emerging, ~1989) →
  `longBack["VEIEX"]`. Now NTSG's start is VFINX (~1976), DBMF's is EM (~1989).
  `frame_start_test.go` reproduces the old 1999 cap and proves the fix.

- **Custom builders now extend too** (done): only `composite()`/`tsmom()` wrapped
  their fetch in `extend()`; the custom builders (`wintonBuild`, `dbmfeBuild`,
  `dpgtBuild`, `backcastBuild`) did not, so Winton stayed 2001 and a clean regen
  would have regressed DBMFE to ~2001. Every `BuildFrame` call site now uses
  `extend(f)`. Regenerate IE000O1VI174, DBMFE, GG00BQBFY362.
- **US legs extended** (done): `longBack["VFINX"]` → `SP500-USD.csv` (S&P 500 total
  return, Shiller price + reinvested dividends, ~1871 — the index VFINX tracks;
  replaced the earlier total-market proxy); `longBack["^IRX"]` → `TBILL-3M.csv`
  (FRED TB3MS, ~1934, a rate, rescaled ≈1 at the splice). NTSG's floor is now the
  dev-ex-US leg (~1969); NTSX reaches ~1953.
- **^HICP-FR long history embedded** (done): `hicp-fr.csv` now carries the OECD
  French CPI (FRED FRACPIALLMINMEI) chained before Eurostat, ~1955→, so the
  offline fallback deflates the high-inflation decades too.

- **Intl-equity proxies upgraded to true MSCI** (done): `DEVEXUS-USD.csv` is now
  MSCI World ex USA gross TR (Curvo, 1969-12→, the real ex-US universe, replacing
  the Ken-French/World approximation) and `EM-USD.csv` is MSCI Emerging Markets
  gross TR (Curvo, 1987-12→). DBMF/RSSB/VT's EM floor moves 1989→1988.
- **S&P 500 total return + VFINX proxy** (done): bundled `SP500-USD.csv` (Shiller
  price + reinvested dividends via datahub, ~1871→, ~9.4%/yr) and pointed
  `longBack["VFINX"]` at it (VFINX tracks the S&P 500, so this beats the earlier
  total-market proxy). Refines the pre-1976 US leg of NTSG/NTSX.
- **Real (inflation-adjusted) drawdown/TTR in the report** (done): the stats table
  now shows "Max Drawdown (real)" and "TTR real" beside the nominal ones,
  deflating each portfolio's window by the base-currency CPI (^HICP-FR for EUR;
  omitted for other currencies until a US CPI is wired). Makes the dot-com read
  ~13y real vs ~6.8y nominal. `deflate()` in cmd/pofo, unit-tested.
- **Yahoo 429 fix** (done): the client now falls back between Yahoo's twin
  query1/query2 hosts (`yahooGet`, yahoo.go), defaulting to query2. A per-host
  rate limit no longer sinks a run; fetch works again from constrained
  environments (verified AAPL + CSPX).

**Open:**
- **Cache expiry — resolved by analysis (2026-07-01), no code change.** The
  premise "old history never changes" does NOT hold for the dividend-adjusted
  close series pofo fetches: Yahoo retro-adjusts every prior close on each new
  dividend, so an incremental append (keep the deep history, fetch only the
  recent delta) would silently accumulate adjustment drift and bias exactly the
  long-run CAGR this tool is precise about. A full re-download is therefore the
  correct default. The lever for "refresh less often" already exists: the
  `-cache-age` flag (`MaxAge`, default 30d) can be raised. If bandwidth ever
  becomes the constraint, the only safe incremental design is delta-fetch +
  periodic full refresh to bound the drift; not worth it at present.

**Done (2026-07-01):**
- **Reference-validation pass (done).** New golden tests validate the bundled
  backcast series against public references over several December-to-December
  windows: `pkg/datasets/golden/refdata_test.go` covers SP500-USD (S&P 500 TR
  decade returns: 1970s 5.9%, 1980s 17.6%, 1990s 18.2%, 2000s −0.9%, 2010s 13.6%
  — all within ~0.35pt), XAUUSD-LBMA (gold 2000s ~14.9%, since-1971 ~8%; the
  datahub monthly *average* reads a touch below the year-end fixes at the
  volatile 1979/80 boundary, validated on the smoother windows), and
  MSCIWORLD-USD. **Finding: MSCIWORLD-USD (and the sibling Curvo exports
  DEVEXUS-USD, EM-USD) are NET total-return indices, not gross** as the CSV
  headers and recipe `Method` strings claimed. Verified numerically: the bundled
  Dec2012→Dec2024 CAGR is 10.82%/yr and Dec2014→Dec2024 is 9.95%/yr, matching
  MSCI World NET USD exactly, while the official GROSS figures are 11.41% and
  10.52% (the withholding drag). This is the *correct* proxy for an Irish-
  domiciled UCITS World ETF (IWDA/URTH are benchmarked against the net index) and
  the recipe only deducts the TER on top, so the reconstruction was already
  right; only the "gross" wording was wrong and has been corrected in the CSV
  headers, the recipe strings (`recipes.go`, `extend.go`) and the two generated
  simdata headers (URTH, IE00B4L5Y983 — data rows unchanged, no regen needed).
  The daily calc algorithms (vol/Sharpe/Sortino/TTR conventions) stay validated
  by the existing SPY/URTH daily fixtures.
- **Report window (done).** `-start` now defaults to empty = *earliest
  available* (zero time), so a plain `pofo` run surfaces the full backcast
  instead of clipping at 2006; the common window still aligns on the youngest
  holding's inception. Pass an explicit `-start` to clip.
- **Pre-1978 EUR/USD**: the ECU series starts 1978-12; earlier still needs a DM
  or EUA proxy. 1978 already covers a 45-year backcast, so deferred.
- **FX granularity pre-2003 is monthly** (the bundled ECU/EUR proxy). The uniform
  "convert at end in USD" stance stands; a fuller fix (per-segment currency
  conversion with a `# currency:` tag before `ExtendBack`, letting the EUR MSCI
  World Curvo export be used directly without double-counting FX) is still open
  but no longer blocking a long backcast.
- **Cache expiry**: resolved by analysis, see the "Cache expiry" item under the
  reorganised Open/Done lists above (Yahoo retro-adjusts, so a full refresh is
  correct; `-cache-age` already tunes the cadence).
- **Sandbox reachability (verified 2026-07-01, in-session):** Yahoo IS reachable
  from the Claude Code sandbox and fetch works — don't assume otherwise, TEST it.
  Verified this session: `./pofo -assets AAPL` and `URTHSIM` fetch live, and
  `./pofo -gen-simdata URTH IE00B4L5Y983` regenerates and validates end-to-end
  (sim CAGR 11.42% vs real URTH 11.73%). So reports, `-fire`, and `make simdata`
  all run here. Only FRED is flaky (the deep French-CPI extension can time out;
  it falls back to Eurostat 1996 gracefully, non-fatal). Stooq is PoW-gated but
  unused. If Yahoo ever 429s, the client falls back query1↔query2 (yahoo.go);
  check BOTH hosts before concluding it's blocked.

**Handover state (2026-07-01):** all the above is committed and pushed; the
embedded simdata is current with the refdata/recipes (regenerated after each
change). Yahoo fetch works from any environment again, so `make simdata` (which
rebuilds → gens → rebuilds; never run a stale `./pofo -gen-simdata`) can be run
in-session. Current backcast starts (with `-start 1968-01-01`): NTSG ~1969, NTSX
~1953, Winton ~1990, DBMF/DBMFE ~1988, VT/RSSB 1987-12, XAUUSD 1968, IWDA/URTH
1969, IEF/TLT/ZROZ 1962.

**To regenerate a bundled refdata series from source** (all sandbox-reachable):
FRED GS5/GS20 → `simgen.TreasuryTR` (treasury); Ken French US / MSCI-World-ex-US
via Curvo / MSCI-EM via Curvo (equity); datahub gold-prices (XAUUSD-LBMA) and
s-and-p-500 (SP500-USD, Shiller TR); FRED WTISPLC (crude), EXUSEC+EXUSEU (EUR),
TB3MS (^IRX), FRACPIALLMINMEI (HICP-FR). See the header of each refdata CSV.

**Next up:** all 18 numbered items above are done, including the `frontend-design`
pass on the `-fire` UI (item 17, shipped 2026-07-02 with the v3 enrichment,
`decumulation-fire-v3-enrichment.md`). The only remaining backlog items are
deliberately deferred (pre-1978 EUR/USD, per-segment FX conversion, the
broad-sample century panel for the historical models = rewrite-spec phase 2) or
resolved-by-analysis (cache expiry). No open work is currently scheduled.
