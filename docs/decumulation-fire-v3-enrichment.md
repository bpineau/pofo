# FIRE explorer v3 — the enrichment drop (2026-07-02)

One autonomous session, on top of the 2026-07-01 rewrite
(`decumulation-fire-rewrite-spec.md`). Goal, per Ben's brief: make the FIRE
view **rich** — much denser in features, models and visuals, and genuinely
useful for the decumulation decisions in the consolidated reference doc
(`/tmp/Decumulation_FIRE_Reference_Consolidee.md`, §11 "cahier des charges").
Everything below shipped in this drop, `go test ./...` green, verified live
against both the parametric and the portfolio (`-fire <file>`) modes.

## Engine (`pkg/decumul`)

- **Per-path timing and lived spending.** `PathResult` gained `RuinYear`
  (first ruin year, -1 when none), `Spend` (net real spending delivered per
  year), `FirstCut`/`CutYears` (years lived below the uncut standard). Both
  kernels (annual reference and monthly) record them.
- **Written-rules spending policy** (reference doc §10):
  - `FlexRule.WRThreshold`: the cut also triggers when the *current
    withdrawal rate* exceeds a bound (e.g. 3.6%), OR-ed with the drawdown
    trigger.
  - `Ratchet`: Kitces-style only-up rule — raise the level by `Step` when
    real wealth exceeds `Trigger`× the initial capital, at most every
    `Cooldown` years, capped at `Cap`, vetoed while the current rate is above
    `MaxWR`. Ignored while guardrails are active.
- **Spending schedules.** `Plan.SpendSchedule` scales the base need year by
  year: health-cost drift and/or a Blanchett retirement smile (web builds the
  multipliers).
- **Multi-envelope taxes** (§11.6). `Plan.Envelopes` splits the growth sleeve
  into ordered tax pockets (drained CTO → PEA → AV); `Envelope.GainFrac`
  models embedded unrealised gains. New `AVTax` implements the assurance-vie
  annual allowance (9 200 €/couple tax-free realised gains, then 24.7%) as a
  stateful per-path, per-year tax (`YearlyTax` interface); the PEA is
  `CTOFlatTax{0.172}`. A nil `Envelopes` is byte-identical to the legacy
  single-CTO sleeve (locked by a parity test); the golden Python-anchored
  tests and the Trinity/broad-sample calibration anchors still pass.
- **Mortality.** `Gompertz` survival (INSEE-fitted `FrenchMortality`, mode 88
  / dispersion 10), `CoupleSurvival` (either member alive), and
  `Ensemble.LifeCurve` → per-year Dead / Broke / Funded shares (the "Rich,
  Broke or Dead" decomposition). `RuinYearHistogram` gives the failure-timing
  distribution.
- **Spending statistics.** `Ensemble.SpendStats` (share of paths ever cut,
  median first-cut year, median and p90 years lived cut) and
  `Ensemble.SpendBands` (per-year quantiles of delivered spending).

## Charts (`pkg/chart`)

- `StackedArea`: part-to-whole layers over time (the alive/broke/gone view),
  warm-study styled like the rest of the package.

## Web (`pkg/decumul/web`)

- **New endpoints** (all POST-a-Params, like the rest; the server handler
  boilerplate was factored into one `post()` helper):
  - `/api/spending` — household real spending fan (delivered spend + pension
    and side income added back) + the cut statistics cards.
  - `/api/lifecycle` — alive-broke-gone stacked area against the French
    couple mortality at the `age` param, ruin-year histogram (5-year
    buckets), and mortality-adjusted cards ("ever alive and broke" weights
    each ruin by survival to its ruin year).
  - `/api/curves` — safe WR vs horizon (central + broad-sample, solved at the
    user's target ruin) and required capital vs spending (central), both on
    the fixed rule.
- **Params** gained `age`, `peaCapital`, `avCapital`, `gainFrac`, `ratchet`,
  `wrTrigger`, `spendDrift`, `smile`; `plan()` translates them (ratchet
  defaults follow the written rules: 1.2× trigger, +10% steps, 1.2× cap,
  2-year cooldown, 2.2% MaxWR).
- **Sensitivity** gained four levers: pension +500 €/m, side income
  12 k€×8 y, the WR-trigger cut, and the ratchet (the one lever that
  *raises* ruin — it prices the lifestyle option).

## UI (`pkg/decumul/web/assets`) — the frontend-design pass (backlog #17)

Two-column desk on the shared warm-study theme: a **sticky control rail**
(grouped fieldsets: situation, pension & side income with **preset chips**
for the three pension scenarios, spending policy, market
model, cash buffer, taxes & envelopes, simulation) and a long **numbered
analysis column**: 01 simulated futures (fans), 02 the spending you actually
live, 03 alive-broke-or-gone, 04 what moves the risk (frontier, levers, the
two planning curves), 05 buffer & recovery, 06 plan detail + the methodology
fold. Signature element: the **plan bar**, a sticky condensed verdict with
one colour-graded ruin bead per model, so a slider's consequence stays
visible however deep the page is scrolled (IntersectionObserver toggles it
when the hero leaves the viewport). Pervasive `data-help` hovers on every
control; the whole state (including the new params and checkboxes)
round-trips through the URL hash; a run-id guard drops stale fetch responses;
the two solver-heavy curves live on a slower debounce lane.

## Deliberately not done (still open)

- Two parametric growth sleeves with correlation (§11.5) — the portfolio
  panel mode already answers the multi-asset question with real
  correlations; a synthetic two-sleeve model would duplicate it.
- Broad-sample bundled panel (1900s+) for the historical models — tracked in
  the rewrite spec's phase 2.
- Flex "max cumulative cut duration" (§11.2) — the new `SpendStats` measure
  the realised cut durations instead, which answers the underlying question
  ("how hard is the floor actually hit?") without another kernel knob.
- Ruin-cause attribution (early crash vs longevity) — the ruin-year
  histogram covers most of the value; full attribution stays phase 2.
