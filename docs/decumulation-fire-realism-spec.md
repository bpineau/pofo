# Decumulation / FIRE — realism & conservatism spec

Status: spec, not yet scheduled (2026-06-29).

Companion to `decumulation-fire-design.md` and `decumulation-fire-followups.md`.
This spec captures work to make the FIRE explorer's headline ruin figures
**trustworthy for a real life plan**, after a diagnosis found them materially
more optimistic than the empirical safe-withdrawal-rate (SWR) evidence.

## 1. Why (the diagnosis)

Running the engine directly (parametric, 50k paths, €1M capital, 40y horizon)
and toggling one assumption at a time, ruin at a 4% withdrawal rate moves as:

| Change from the current web defaults | Ruin at 4% |
|---|---|
| Baseline (mu 4.5%, sigma 12%, flex 25%, 3y buffer, 40y) | **21.8%** |
| Turn off the flex cut (a *true* fixed 4% rule) | 48.5% |
| mu 3.0% real (broad-sample) instead of 4.5% | 45.6% |
| sigma 17% (single-region equity) instead of 12% | 37.8% |
| horizon 50y (longevity) | 38.9% |
| Pessimistic stack (fixed, mu 3%, sigma 17%, df 4, 50y) | **82.8%** |

The kernel itself is sound (after the P1 fixes). The optimism is **defaults +
fitted data + the i.i.d. return model**, not a bug. For MSCI World specifically
the reported ~1.8% is driven by (a) the flex cut halving ruin, (b) mu/sigma
fitted from a short, US-heavy, post-2009-favourable 1999–2026 window, and (c)
i.i.d. draws that cannot produce the persistent multi-decade real bear markets
(Japan post-1990, 1910s–40s Europe) that actually cause ruin.

Reference: Anarkulova, Cederburg & O'Doherty, "The Safe Withdrawal Rate:
Evidence from a Broad Sample of Developed Markets" (2023, SSRN 4227132): a
block bootstrap over ~38 developed markets, 1890–2019, with realistic
longevity, finds a fixed 4% rule fails far more often than US-only backtests
suggest (their 5%-failure SWR is ~2.26% real).

Already shipped (honesty only, numbers unchanged): a short-history warning
(`reliabilityCaveat`) and an expandable methodological note in the page.

## 2. Goals / non-goals

- **Goal:** the out-of-the-box headline should equal the *canonical fixed
  withdrawal* result on *honest* return assumptions, with optimism-reducing
  levers (flex, guardrails, favourable fit) made explicit opt-ins.
- **Goal:** let the user reach broad-sample-style pessimism in one click.
- **Non-goal:** match Anarkulova's exact numbers. We model a EUR investor with
  the CTO flat tax and our own data; the aim is calibrated realism and
  transparency, not replication.
- **Non-goal (this spec):** full mortality/longevity modelling (tracked
  separately; here we only nudge the horizon).

## 3. Workstreams

### W1 — Default the flex cut OFF (highest impact)

**Problem.** `flexCut` defaults to 0.25, so the headline is "4% *and I cut 25%
in a crash*", which roughly halves ruin vs the fixed 4% rule everyone means.

**Change.** Default `flexCut` slider to 0. Keep flex and the Guyton-Klinger
guardrails as explicit "I will adapt spending" levers. Optionally relabel the
result so it states whether spending is fixed or adaptive.

**Files.** `pkg/decumul/web/assets/app.js` (SLIDERS default), maybe a label in
`model.go` cards.

**Acceptance.** With default sliders and no flex, the 4% headline matches the
fixed-withdrawal figure; turning flex on visibly lowers ruin.

**Priority: P1.**

### W2 — More honest return defaults + a "conservative prior" toggle

**Problem.** Defaults `mu 4.5% / sigma 12%` and, worse, the per-portfolio fit
from a short favourable window are optimistic: low vol, high mean, thin tails.

**Change.**
- Raise the default `sigma` toward ~0.16 and lower default `mu` to ~0.030–0.035
  real for a 100% equity sleeve.
- Add a **"Conservative (broad-sample) prior"** toggle that overrides the fit
  with long-run global real-equity assumptions (e.g. mu ≈ 0.03, sigma ≈ 0.17,
  df ≈ 4–5) regardless of the portfolio's own rosy history, with a one-line
  explanation.
- Consider deriving the seeded `df` and a small negative skew from the broad
  prior rather than only the (thin) fitted sample.

**Files.** `pkg/decumul/web/assets/app.js` (SLIDERS, a toggle), `model.go`
(apply the prior when set), `portfolio.go` (`FitParametric` — optionally blend
toward the prior, or expose both fitted and prior values).

**Acceptance.** The toggle moves the headline up to the broad-sample regime;
documented and round-tripped through the URL hash.

**Priority: P1.**

### W3 — Capture sequence risk and the fat left tail

**Problem.** I.i.d. symmetric Student-t draws have no autocorrelation and no
skew, so they under-produce the early-crash, prolonged-real-drawdown paths that
cause ruin. The historical bootstrap is only as good as its (short, favourable)
sample.

**Change (options, pick during planning).**
- Bundle a **long, broad-sample real-return panel** (developed-market index or
  a curated multi-country series) to bootstrap from, so the historical models
  see 1900s–2020s regimes, not just 1999–. This is the most faithful fix and
  the closest to Anarkulova.
- And/or add **negative skew** and mild **autocorrelation/regime persistence**
  to the parametric source (e.g. a two-state or AR(1)-in-vol generator) so the
  i.i.d. path is not the only synthetic option.

**Files.** `pkg/scenario/` (a new source or a skewed/regime variant; a bundled
panel under `pkg/datasets/`), `pkg/decumul/web/` wiring + a model-picker entry.

**Acceptance.** At equal mu/sigma, the new model yields higher ruin than i.i.d.
(captures sequence risk); validated against published broad-sample SWR figures
within a stated tolerance.

**Priority: P2** (larger; the highest-fidelity item).

### W4 — Horizon / longevity nudge

**Problem.** A fixed 40y understates an early-retirement (FIRE) horizon; ruin
rises steeply with years (21.8% → 38.9% from 40y → 50y).

**Change.** Raise the default horizon for FIRE (or label it as "years from
today, set past your life expectancy"), and optionally add a simple longevity
note. Full mortality modelling is out of scope here.

**Files.** `pkg/decumul/web/assets/app.js` (years default/label).

**Priority: P3.**

## 4. Decisions needed before implementation

- W1: flip the default to flex-off, or keep flex-on but add a prominent "fixed
  vs adaptive" headline toggle?
- W2: exact broad-sample prior values (mu/sigma/df/skew) and their source.
- W3: bundle a broad-sample panel (which dataset, licensing/availability under
  the stdlib-only, embedded-data constraint) vs a synthetic skew/regime source.
