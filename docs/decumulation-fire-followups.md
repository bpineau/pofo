# Decumulation / FIRE — follow-ups backlog

Running list of oddities to fix and enhancements to evaluate, from the fresh
`pkg/scenario` + `pkg/decumul` + `pkg/decumul/web` implementation (designs:
`decumulation-fire-design.md`, `decumulation-monthly-sampling-design.md`).
Not yet scheduled; each item should get its own brainstorm → spec → plan when
picked up. Priority: **P1** correctness, **P2** clarity/API, **P3** features.

## Correctness oddities (P1)

1. **Ruin is tested on net need, not the grossed-up withdrawal.**
   `Plan.RunPath` flags ruin with `if need > total` (net), but the money that
   must actually be liquidated is the *gross* (need + tax, and buffer refills).
   With the cost-basis `CTOFlatTax`, gross can substantially exceed net at high
   embedded gains, so ruin can be under-flagged. The golden tests pass because
   they use a flat-12% stub and a no-buffer plan, which masks it. Fix: define
   ruin as "the gross required to deliver `need` exceeds available liquidity",
   computed consistently across the buffer-first and growth-first branches.

2. **Partial-funding is silent.** `Tax.GrossUp` caps `gross` at `growth`, so a
   year that cannot fully fund `need` sells everything and delivers *less* net
   than needed, yet `RunPath` still does `Withdrawn += need` and may not latch
   ruin. The household "didn't get its money" without the path being marked.
   Fix together with (1): detect under-delivery (buffer drain + net deliverable
   from growth < need) and latch ruin; account the real net withdrawn.

3. **`0` is overloaded as "unset" in `BufferSleeve`.** `DrawThreshold == 0`
   and `RefillCap == 0` are treated as "use the default" (0.10 / 0.50), so a
   user cannot intentionally set a 0 threshold or a no-refill policy. Use
   explicit defaults at construction, sentinel `-1`, or pointer fields.

4. **`Sweep2D(_, Mu, …)` on a non-parametric plan is a silent no-op.**
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

8. **Common-window alignment truncates to the last N months/returns** assuming
   the month grids line up across holdings. Dense simdata-extended series make
   this safe in practice, but a holding with internal gaps could misalign. A
   date-keyed alignment (intersect on month keys) would be robust.

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

11. **Allocation A/B comparison.** Pin a baseline allocation and show a variant
    side by side (the original motivating question: "60/25/15 vs 20/20/…"),
    rather than only re-dragging one set of weights.

12. **Solve in the UI.** Expose `CapitalForRuin` and a buffer optimiser: given a
    target ruin %, show the required capital and the ruin-minimising buffer.

13. **Monthly withdrawal kernel (stated requirement, P2-ish).** Ben's real
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

14. **Richer policies.** Melting/glidepath buffer (stop refilling after the
    sequence-risk window), a distinct inflation-linked sleeve vs pure cash,
    side income (rental/activity) as another `Cashflow`, and a guardrails
    withdrawal rule (Guyton-Klinger style) beyond the single flex cut.

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

17. **Redesign the `-fire` UI with the frontend-design skill (later).** The
    current explorer (`pkg/decumul/web/assets/`) was built functionally, not
    designed: layout, typography, colour, chart styling and the overall visual
    hierarchy were never given a deliberate pass. Redo it with the
    `frontend-design` skill for a distinctive, intentional look (and reconsider
    chart rendering, the allocation bar styling, mobile layout, and surfacing
    the hidden metrics from item 10 as part of the same pass).

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
    The FIRE-seeding / monthly-Sharpe reuse (the last sub-bullet) is still open.
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
