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

5. **`worst10y` uses a `-1` sentinel** to mean "hit zero within the window",
   then aggregates by `min`, so a single ruined path drives `Worst10yCAGR` to
   −100%. Arguably correct but the sentinel conflates "−100% realised" with
   "undefined"; revisit the representation (e.g. report separately the share of
   paths with a sub-`-x%` decade).

## Clarity / API (P2)

6. **2D surface y-axis silently changes meaning** (real CAGR for parametric,
   spending floor for the historical models). The heatmap title says "scenario
   axis (y)" generically; label the axis dynamically and state the unit.

7. **Monthly→annual compounding is rolling-from-window-start, not calendar
   years.** For cohorts a 15y window compounds months `[0..11], [12..23], …`
   from the cohort start, so the "annual" returns are not Jan–Dec. Statistically
   fine; document it so it is not mistaken for a bug.

8. **Common-window alignment truncates to the last N months/returns** assuming
   the month grids line up across holdings. Dense simdata-extended series make
   this safe in practice, but a holding with internal gaps could misalign. A
   date-keyed alignment (intersect on month keys) would be robust.

9. **`FitParametric` does not fit `df`** and estimates annual sigma from ~20
   annualised points (noisy). Consider deriving annual vol from monthly returns
   (`σ_m·√12`) for stability, and optionally fitting `df` from the sample
   kurtosis to seed the fat-tail slider from history.

## Features / "most useful FIRE info" (P3)

10. **Surface the computed-but-hidden metrics.** `Outcome` already produces
    median years underwater, worst-10y real CAGR and CDaR (and the recovery
    histogram is shown); add cards/panels for the rest, plus median cumulative
    tax and effective tax rate.

11. **Allocation A/B comparison.** Pin a baseline allocation and show a variant
    side by side (the original motivating question: "60/25/15 vs 20/20/…"),
    rather than only re-dragging one set of weights.

12. **Solve in the UI.** Expose `CapitalForRuin` and a buffer optimiser: given a
    target ruin %, show the required capital and the ruin-minimising buffer.

13. **Fully monthly kernel (separate, bigger).** Monthly withdrawals and
    intra-year sequence-of-returns; would re-validate the golden numbers. Keep
    the annual kernel as the validated default; offer monthly as an option.

14. **Richer policies.** Melting/glidepath buffer (stop refilling after the
    sequence-risk window), a distinct inflation-linked sleeve vs pure cash,
    side income (rental/activity) as another `Cashflow`, and a guardrails
    withdrawal rule (Guyton-Klinger style) beyond the single flex cut.

15. **Shareable scenarios.** URL-encode the slider/allocation state so a
    configuration can be bookmarked or shared (the server is local, so this is
    cheap).

16. **Performance.** Each `/api/sim` runs Sweep1D + a full Simulate + Sweep2D,
    i.e. many independent Monte-Carlo passes per slider drag. Share pre-drawn
    paths across the sweep evaluations (as `CapitalForRuin` already does) to cut
    the per-request cost, especially at higher path counts.
