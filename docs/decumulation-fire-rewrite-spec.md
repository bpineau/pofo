# FIRE explorer — rewrite spec

Status: spec, ready for brainstorm → plan (2026-06-30).

Supersedes the UX direction in `decumulation-fire-realism-spec.md` and
`decumulation-fire-usability-proposals.md`; keeps their engine analysis. This is
the synthesis of three inputs: Ben's brief and objections, a ChatGPT thread
(shared 2026-06-30), and Claude's code-level analysis
(`decumulation-fire-usability-proposals.md`). Read those for the long-form
rationale; this document is the build target.

The engine lives in `pkg/scenario` (return models), `pkg/decumul` (the
withdrawal kernel + outcomes), `pkg/decumul/web` (server + assets), `pkg/chart`
(server-side SVG). This spec changes all four.

---

## 1. The problem in one paragraph

The tool answers "what is my risk of ruin?" with a single number that swings
from 0.1% to 92.3% as checkboxes are toggled, gives no central case, no picture
of the simulated market, and a paragraph of disclaimers instead of a position.
Two root causes: (a) **a framing error**, ruin is treated as one knowable number
when it is irreducibly a *range across modelling choices*; and (b) **a
calibration bug**, the "Stress regimes" toggle silently turns the portfolio's
expected real return negative (see §2), so part of the spread is not legitimate
model disagreement but a coding error. The rewrite fixes both: it presents an
honest **range across calibrated models**, makes the market **visible**, and
makes the headline **actionable** (safe withdrawal rate and the levers the user
controls).

---

## 2. Non-negotiable engine fix first: mean-preserving sequence risk

This is the most important correction and ChatGPT missed it entirely. Today
`web/model.go:109` builds the regime bear state as `BearMu = mu - 2*sigma`,
which yields (verified numerically):

| Sliders | Bear mean | Blended long-run real mean with regime ON |
|---|---|---|
| default (μ 4%, σ 16%) | −28%/yr | **−1.95%/yr** |
| conservative (μ 3%, σ 18%) | −33%/yr | **−3.70%/yr** |

So "cluster bad years (sequence risk)" secretly converts the portfolio into one
with a **negative expected real return held forever**, worse than any real
diversified equity history. That, compounded with the conservative prior (which
already lowers the mean), is the 92.3%.

**Sequence risk is about the ordering and clustering of returns around an
unchanged long-run mean, not about lowering the mean.** The rewrite must
separate two orthogonal axes and never let them silently multiply:

- **Return level** (an *epistemic* choice about the mean): fund-fit vs honest
  forward-looking vs broad-sample-conservative. Lower mean is a legitimate
  pessimistic assumption, but it must be labelled as "lower expected return",
  not hidden inside a "sequence risk" switch.
- **Return dynamics** (sequencing): i.i.d. vs block-persistent vs
  regime-switching. These must be **mean-preserving**: re-calibrate
  `MarkovRegime` so its blended mean equals the active mean, and choose its
  persistence/skew so the *worst rolling 10y/30y real CAGR* matches historical
  broad-sample statistics, not an ad-hoc −2σ bear.

Acceptance: at equal target mean, the regime source produces a worse 5th-pctile
worst-decade than i.i.d. (it already does) **while keeping the blended mean
within a small tolerance of the target** (it currently does not). Locked by a
test (§7).

---

## 3. Conceptual model: two kinds of uncertainty, shown separately

State-of-the-art retirement research (Pfau, Kitces, Blanchett, Cederburg &
O'Doherty / Anarkulova et al., EarlyRetirementNow) has converged on: *there is
no single FIRE number; you compare models.* We adopt this explicitly and make
the distinction visible:

- **Aleatory uncertainty** (Monte-Carlo randomness *within* one model): shown as
  the fan band (p5–p95) on the wealth chart and as the ruin% of that model.
- **Epistemic uncertainty** (*which* model / assumptions): shown as the spread
  *across* the model-comparison columns and summarised as the "envelope".

The headline is never one number. It is: a **central estimate**, a **plausible
range**, and a **confidence badge** stating why the range is wide (e.g. short
fund history vs a 50y horizon).

---

## 4. The model set (the columns)

All models run in a single request and are shown side by side. Each is
calibrated and anchored (§7). Plain-language hover text is mandatory.

| Column | Source | What it captures | Honest caveat (hover) | Anchor |
|---|---|---|---|---|
| **Historical** | `HistoricalCohorts` (rolling windows) | Real observed sequences, no resampling | Few independent retirements; optimistic for long horizons | own panel |
| **Block bootstrap** | `StationaryBootstrap` (already block, mean-block 24m) | Resamples multi-year blocks: preserves clustered bear markets and cross-asset correlation | Anchored to one favourable window | own panel |
| **Student-t i.i.d.** | `ParametricSource` | Fat tails matching mean/vol/kurtosis | No persistence; under-produces early-crash paths | fitted μ/σ/df |
| **Regime** | `MarkovRegime` (mean-preserving, §2) | Persistent bull/bear; long real drawdowns (sequence risk) | Synthetic persistence, calibrated to historical worst-decade | same mean as Student-t |
| **Conservative (broad sample)** | parametric/regime with broad-sample prior | Lower forward real return, higher vol, fat left tail | Forward-looking pessimism, not this fund's history | Anarkulova-class |

Notes from the critical reading:

- **Address Ben's "coward warnings" complaint.** The short-history problem is
  solved *by method, not by refusing*: the block bootstrap manufactures many
  synthetic full-length retirements from a short panel (legitimate), and the
  Conservative column supplies the deep-past behaviour the fund lacks. The wall
  of warnings collapses into one compact confidence badge. A future, optional
  upgrade bundles a broad-sample real-return panel under `pkg/datasets/` so the
  historical models see 1900s–2020s regimes directly (the gold standard; data
  licensing TBD, see §9).
- **Do not invent historical labels for synthetic paths** ("1973 style", "Japan
  style") unless the path is a real cohort. For synthetic models, describe the
  mechanism honestly ("two bear markets three years apart"), never a fake date.
  (ChatGPT proposed the labels; rejected as misleading.)
- **CAPE-conditioned returns**: deferred (pofo has no CAPE series wired).

---

## 5. UI / information architecture (dense but legible)

Ben likes **dense information**, so the goal is not minimalism. The real
constraint is *legibility of change*: however much is on screen, the impact of
moving a slider must be **immediately visible** without scrolling kilometres or
re-running. Resolution: a fixed **hero strip that updates live on every drag**
(always in view), a **persistent main layout** (the sliders, the model strip and
the active chart never move), and rich detail shown by default rather than
hidden. Tabs switch the *main chart* only; they do not gate the headline
numbers. Keep the sliders and the allocation bar.

**Pervasive hover help (Ben's request).** Every model column, every toggle,
every metric and every chart axis carries a plain-language mouse-hover
explanation aimed at a non-specialist ("what this is, why it differs, how to
read it"). The long methodology stays one click away; the inline hovers carry
the day-to-day understanding.

```
┌───────────────────────────────────────────────────────────────────────┐
│  Safe spend ≈ €59k/yr (3.3%) at 95% success · you plan €48k (4.0%)      │  ← hero verdict (live)
│                                                                         │
│            Historical  Bootstrap  Student-t  Regime  Conservative       │  ← model strip (live)
│  Ruin        0.3%        0.8%       2.5%      9.1%      18.0%   ⓘ        │
│  Safe WR     3.9%        3.8%       3.7%      3.4%       3.0%            │
│  Med. wealth 420k        390k       340k      240k       120k           │
│  ░ confidence: MEDIUM — fund history 18y vs 45y horizon                 │
├───────────────────────────────────────────────────────────────────────┤
│  CONTROLS (left rail)        │   MAIN  [ Paths | Frontier | Sensitivity ]│
│  ── You control ──           │                                          │
│  Horizon, Retire-in years,   │     (one server-rendered SVG, ~480px)    │
│  Capital/Savings, Spending,  │                                          │
│  Allocation (drag bar)       │                                          │
│  ── Market assumptions ──    │                                          │
│  μ, σ, tail df, buffer, tax  │   Solve for [Safe WR ▾] @ [5% ruin] →    │
│  ── Advanced (collapsed) ──  │   "Retire in 2033 (vs 2030)"             │
│  monthly · guardrails ·      │                                          │
│  conservative · regime       │   ▸ More metrics (collapsed)             │
└──────────────────────────────┴──────────────────────────────────────────┘
```

### Hero strip (always visible, recomputed live, ~200ms debounce)
- **Verdict line**: the safe withdrawal rate in €/yr at the **target ruin**, vs
  the user's planned spend. The target ruin is a control (default 5% ruin /
  95% success; the user can dial it to 2%, 10%, etc.), so the verdict answers
  "at the risk level *I* accept, what can I spend?" not a single fixed
  definition of safe. This is the single most useful sentence; promote it from
  the buried "Solve" panel. (Both Ben and ChatGPT: SWR + ruin are the two things
  that matter; Claude's original solver targeting *starting capital* was wrong
  because capital is the least controllable input.)
- **Model strip**: the §4 table, three rows (Ruin, Safe WR, Median wealth),
  one column per model, cells colour-graded (green→red). Column headers carry
  the plain-language hover (§4). This is the epistemic-uncertainty view and the
  core of the redesign.

### Main chart (tabs, one at a time, server-side SVG)
1. **Paths** (default): wealth fan chart (median + p25–p75 + p5–p95 bands) for
   the selected model, overlaid with ~8 representative sample paths including
   real ruin paths. Answers Ben's question directly ("is it a −80% mega-crash or
   March-2020 repeated?"). A model selector switches the underlying model;
   default to the central one.
2. **Frontier**: ruin % vs withdrawal rate, **one curve per model**, with a
   vertical marker at the user's current WR and a horizontal marker at the
   target ruin. Reference ticks for the classic 4% rule and the broad-sample
   ~3% so the user sees where the plan sits. Replaces the buffer-arbitrage chart
   as the hero analytic chart.
3. **Sensitivity** ("what makes my plan robust"): a horizontal bar chart of
   Δruin (or ΔsafeWR) per one-step move of each *controllable* lever: +1 year
   working, +100k capital, −5k spending, +10% bonds, buffer +2y, guardrails on.
   ChatGPT's "greeks" idea; the highest-value view for someone still
   accumulating. Computed by finite differences over the central model.

A fourth **"Why it fails"** view (ruin-cause attribution: early-crash vs
inflation vs longevity, plus the drawdown-duration histogram per model) is
specified but **phase 2** to protect compactness.

### Controls
- Keep all sliders; regroup into **"You control"** (horizon, years-to-retire,
  capital/savings, spending, allocation) and **"Market assumptions"**
  (μ/σ/df/buffer/tax). Advanced toggles (monthly, guardrails, conservative,
  regime) collapse into a disclosure: with the model strip always showing
  conservative and regime as columns, these become "make this assumption the
  active scenario for the Paths/Sensitivity views", not the primary control.
- **Keep the allocation drag bar and make it actually drive results.** Today a
  drag only matters via a μ/σ/df refit, which is skipped when "conservative" is
  on and is invisible when no panel is loaded; verify and fix so every model's
  inputs move with the allocation (refit the parametric/regime moments, reweight
  the panel for historical/bootstrap), and show a quick A/B against a pinned
  baseline as today.

### Solver: "what do I need to hit my target ruin?" (multi-lever)
The user sets a **target ruin** (e.g. "keep it < 2%") and the solver answers
*per controllable lever*, showing the menu of equivalent ways to get there, not
one number. No accumulation phase (Ben's decision), so it never targets starting
capital, retirement year or required savings.
```
Keep ruin below  [ 2% ▾ ]   (under the [ central ▾ ] model)

To reach it, any one of:
  • Withdrawal      3.1%  (€56k/yr)      vs your 4.0% (€48k)   ← spend less
  • Temporary cut   accept a 25% spending cut in downturns      ← flex / guardrails
  • Allocation      shift ~15% equity → bonds/buffer
  • Buffer          hold 5y cash instead of 3y
  • (or combine: 3.6% withdrawal + a 15% downturn cut)
```
The "temporary cut" lever is first-class: the user explicitly wants to trade a
**reversible drop in living standard during bad years** (the flex cut depth, or
the Guyton-Klinger guardrail band width) against a lower ruin, instead of only
permanently spending less. So the solver root-finds on the lever the user
chooses, including the flex/guardrail depth, to meet the chosen target.

Engine: `CapitalForRuin` is generalised to `Solve(targetRuin, lever)` that
root-finds the chosen lever (withdrawal rate, flex-cut depth, equity weight,
buffer years; capital demoted) at the chosen target and model. The live "menu"
runs one root-find per lever and is cheap with shared pre-drawn paths.

### Detail metrics (shown, visually secondary)
Terminal p5, years underwater, worst-10y CAGR, CDaR, taxes: kept on screen (Ben
likes dense information) but in a smaller, lower-priority band under the hero, so
they inform without competing with ruin/safe-WR. Each carries a hover
explanation. They update live like everything else.

---

## 6. Engine / API changes (`pkg/decumul`, `pkg/scenario`, `web`)

1. **`MarkovRegime` mean-preserving re-calibration** (§2). Add a constructor
   that, given a target mean/σ/df and a persistence target, derives bear/calm
   parameters so the blended mean matches. Keep the raw struct for tests.
2. **Multi-model in one pass.** New `web` endpoint `/api/models` returns, for a
   single `Params`, the `Outcome` of each model in §4 (ruin, safeWR, median
   wealth, confidence). Reuse pre-drawn paths where the source allows
   (`drawPaths`/`simulateOn` already exist); the per-model cost at 2000 paths is
   acceptable. The live hero strip calls this.
3. **Multi-lever solver.** Generalise `CapitalForRuin` into
   `Solve(targetRuin, lever, model)` that root-finds a selectable lever
   (withdrawal rate, flex-cut depth / guardrail-band width, equity weight,
   buffer years; capital demoted) to meet a **user-set target ruin** (not a
   fixed 5%). Drives both the verdict line and the solver "menu" (one root-find
   per lever). No accumulation pre-phase (Ben's decision), so retirement-year and
   required-savings targets are out of scope.
4. **Ruin-frontier series**: `Sweep1D` over withdrawal rate per model, returned
   for the Frontier chart.
5. **Sensitivity**: finite-difference Δruin/ΔsafeWR per controllable lever,
   computed on the central model, returned for the Sensitivity chart.
6. **Ruin-cause attribution** (phase 2): tag each ruined path by dominant cause
   (early-crash if max drawdown in first third, inflation/return-shortfall,
   longevity/horizon) and return shares.
7. **Confidence metric**: a single function of (history length, horizon, model)
   returning {HIGH, MEDIUM, LOW} + a one-line reason, replacing the prose
   caveats. The long methodological note moves to a collapsed "Methodology"
   disclosure.

API shape (sketch):
```
POST /api/models   → { models: [{name, ruin, safeWR, medianWealth, confidence}], envelope, verdict }
POST /api/paths    → { model, fanSvg }          // bands + sample paths
POST /api/frontier → { frontierSvg }            // ruin vs WR, per model
POST /api/sensitivity → { sensitivitySvg }      // greeks bars
POST /api/solve    → { variable, target, result }  // generalised solver
```

---

## 7. Calibration & validation (earn trust)

The tool currently has no anchor, which is why no number is believable. Add Go
tests that pin the engine to published results within tolerance:

- **Trinity/Bengen anchor**: a US-historical-ish parameterisation, fixed 4% real
  over 30y, must give ≈95% success (±a few points).
- **Broad-sample anchor**: the Conservative parameterisation, fixed 4% real with
  a couple's longevity, must give a materially higher failure consistent with
  Anarkulova et al. (their 5%-ruin SWR ≈ **2.26%**; the 4% rule ≈ **17.4%**
  failure for 60/40 domestic, improving to a ~**3.0%** SWR with ~90% global
  equity). Source: Anarkulova, Cederburg & O'Doherty, *Journal of Pension
  Economics & Finance*, 2025 (SSRN 4227132).
- **Mean-preservation test** for the regime (§2).
- **On-screen anchor line** (compact): "Reference: classic US 4%/30y ≈ 95%
  success; broad century-long samples ≈ 75–80% for the same fixed rule. Your
  central case: N%."

Reference points to cite in the methodology disclosure and to seed defaults:
- Morningstar *State of Retirement Income* 2025: **3.9%** base (30y, 90%, 30–50%
  equity), 3.7% in 2024; flexible strategies push the start toward ~6%.
- EarlyRetirementNow (Karsten): ~**3.25–3.5%** for 50–60y FIRE horizons (CAPE
  aware).
- Anarkulova et al.: international diversification *raises* the broad-sample SWR
  (2.26% domestic 60/40 → ~3.0% at 90% global equity), supporting a globally
  diversified equity sleeve rather than heavy domestic bonds.

---

## 8. `pkg/chart` additions

Stay server-side SVG (project ethos: Go, stdlib, embedded; no JS chart deps).
New primitives:

- **`chart.Fan`**: a median line with shaded percentile bands (p5–p95, p25–p75)
  plus optional overlaid individual paths in distinct colours (the spaghetti /
  ruin paths). Drives the Paths tab.
- **`chart.MultiLine`**: several labelled XY series on shared axes with markers
  (vertical at current WR, horizontal at target ruin) and reference ticks.
  Drives the Frontier tab. (Generalises the existing `LineDual`.)
- **`chart.HBars`**: horizontal bar chart with signed values and value labels,
  for the Sensitivity "greeks". (The existing `Bars` is vertical; either extend
  it with an orientation option or add `HBars`.)
- **Reuse** `Bars` (with the y-tick/label work already done) for the
  drawdown-duration histogram in the phase-2 "Why it fails" view.

Colour: a single perceptual green→amber→red scale for the model-strip cells and
consistent per-model series colours across charts (a model keeps its colour
everywhere).

---

## 9. What was adopted, rejected, deferred (the critical reading)

**Adopted (converged across ChatGPT + Claude + Ben):**
- Multi-model comparison shown simultaneously; ruin **and** safe-WR per model.
- Envelope/central-estimate framing instead of a single number.
- Wealth fan + representative/ruin paths ("show me the market").
- Ruin-vs-withdrawal-rate frontier as the hero analytic chart.
- Sensitivity "greeks" over the levers the user controls (ChatGPT).
- Solver targets safe-WR / retirement-year / savings, not starting capital.
- Keep sliders (grouped) and the allocation drag bar; fix the bar to drive
  results (Ben).
- Plain-language model hovers; one compact confidence badge instead of a wall of
  warnings (Ben's "coward" complaint).
- Make historical models usable via block bootstrap + a broad-sample backbone,
  rather than refusing on short history (Ben).

**Claude's additions ChatGPT missed (kept, high priority):**
- The regime mean-tanking **bug fix** (§2): the spread is partly a coding error,
  not pure model disagreement; fixing it is prerequisite to an honest model
  strip.
- Calibration/validation tests + on-screen anchor (§7).
- Explicit separation of "lower mean" vs "worse ordering"; forbid silent
  multiplication of pessimism toggles.

**Rejected / downweighted:**
- ChatGPT's first-pass "remove almost all sliders" (Ben rejected; ChatGPT later
  walked it back).
- Shipping every panel at once (greeks + drivers + decomposition + cause +
  frontier + paths + confidence). Compactness is a hard constraint; phase the
  lower-value views behind tabs/collapsibles.
- Fake historical labels for synthetic paths (misleading).

**Deferred to a later phase:**
- "Why it fails" ruin-cause attribution + drawdown-duration histogram (phase 2).
- Risk decomposition by turning factors on/off (taxes/inflation/sequence).
- Deep integration with the rest of pofo ("Explore improvements" jumping into
  portfolio analyses) — good direction, scope-controlled out of v1.
- Bundled broad-sample historical panel under `pkg/datasets/` and CAPE-aware
  returns (data/licensing work).

---

## 10. Phasing

1. **Engine truth**: §2 regime fix + §7 calibration tests. Nothing else is
   trustworthy until this lands.
2. **Multi-model core**: `/api/models`, the live hero strip (verdict + model
   strip + confidence), and the multi-lever `Solve(targetRuin, lever, model)`
   with a user-set target. The conceptual redesign.
3. **Show the market**: `chart.Fan` + Paths tab; `chart.MultiLine` + Frontier
   tab.
4. **Robustness**: `chart.HBars` + Sensitivity tab; the solver "menu" (per-lever
   answers to hit the target ruin, incl. the temporary-cut lever); fix the
   allocation bar so it drives every model.
5. **Polish**: confidence/anchor copy, methodology disclosure, demoted detail
   metrics, colour system.
6. **Phase 2**: "Why it fails", risk decomposition, pofo integration, bundled
   broad-sample panel.

---

## 11. Decisions (resolved with Ben, 2026-06-30)

- **Return-model stance (resolved after a calibration pass):** i.i.d., calibrated,
  accepting long-horizon caution. Pessimism comes from tails, volatility and
  sequence clustering, NOT from an implausibly low mean. The earlier "μ 3.25% /
  σ 17%" defaults were wrong: read as the model's arithmetic mean and 1-year
  vol, they imply ~2% real geometric and reproduce the "FIRE is impossible"
  artifact. Two corrections were essential and are now in the engine:
  - **Geometric vs arithmetic.** `Mu` is the arithmetic per-period mean;
    geometric ≈ Mu − σ²/2. Target a realistic ~4.5% real geometric.
  - **i.i.d. vs the 1-year vol.** i.i.d. draws have no cross-year mean reversion,
    so feeding the ~17% one-year equity vol into a 30-50y i.i.d. model
    double-counts dispersion. Use the long-horizon (variance-ratio-consistent)
    volatility, ~11%.
- **Calibrated defaults / model family:**
  - Central **Student-t (i.i.d.)**: μ 5% / σ 11% / df 5 (geometric ~4.4% real).
    Reproduces the literature at 30y (~3.4% safe WR, Trinity/Morningstar range)
    and reads honestly tougher at 45-50y. **This is the verdict source.**
  - **Regime**: a sequence-risk *stress* at the same mean (mean-preserving
    `NewMarkovRegime`, `bearGapFactor` 0.6 so it costs a realistic ~0.5% of safe
    WR, not ~1%).
  - **Conservative (broad-sample prior)**: μ 4.5% / σ 13% / df 4 with clustering,
    ~3.5% real geometric, landing ~1.9% safe WR at 30y (Anarkulova ballpark).
  - At 45-50y the i.i.d. family reads stricter than the mean-reverting historical
    evidence by design; the Historical / Block-bootstrap columns (when a panel is
    loaded) show the mean-reverting counterpoint.
- **Target ruin is a control**: default 5% ruin / 95% success for the headline
  safe-WR, but the user can dial the target (2%, 10%, ...) and the solver answers
  per lever for that target. Not a single fixed definition of "safe".
- **Broad-sample historical panel**: **deferred to phase 2** (block bootstrap +
  the Conservative column cover the dark past for v1). Tracked as a follow-up.
- **Accumulation phase**: **out of scope.** The solver targets safe-WR /
  spending / allocation, never retirement-year or required-savings, and never
  starting capital as the default.

Still to settle during planning (low stakes):

- **Block bootstrap mean-block**: keep the 24-month default, or expose it as a
  slider? (Lean: keep fixed for v1.)
