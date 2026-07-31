# FIRE improvement program (2026-07-05)

Follow-up to the assessment thread: what was missing in the FIRE engine/report
versus the state of the art, decomposed into shippable increments. Each has its
own design spec under `docs/superpowers/specs/`.

## Shipped

- **A — Broad-sample empirical model.** `cmd/gen-broadsample` builds a
  per-country real-return table (JST Macrohistory, 18 economies 1870-2020) into
  `pkg/datasets/broadsample/country-real.csv` (`make broadsample`).
  `scenario.PooledBootstrap` resamples single-market runs (not a diversified
  world index, which was too rosy), so national disasters survive: pooled geo
  ~4.4%, 4%/30y ruin in the Anarkulova band. Replaces the synthetic
  "Broad-sample" column. Sanity guard `decumul.GeoMean`/`Plausibility` locks the
  doom-loop.
- **B — CAPE valuation anchoring.** `cmd/gen-cape` bundles Shiller CAPE
  (`pkg/datasets/cape/shiller-cape.csv`, `make cape`). A `capeAdjust` toggle sets
  the central return to the CAPE-implied estimate (1/CAPE + vol drag);
  `/api/meta` serves the valuation snapshot.
- **C — VPW rule.** `Plan.Percent`: percentage-of-portfolio spending, never
  ruins but variable, the opposite end of the frontier. Exposed as a slider.
- **E (partial) — three new visuals.** `chart.Gauge` (CAPE valuation gauge,
  section 00), `chart.Scatter` + `PolicyFrontier` (ruin vs lifestyle-volatility
  frontier over the four rules, section 04), `chart.CategoryBars` +
  `Ensemble.RuinTiming` ("why plans fail" early/mid/late decomposition, section
  03).

- **D — allocation glidepath + partial annuity (shipped).**
  `scenario.Glidepath`: a correlation-preserving two-asset Student-t source whose
  equity weight glides 30%->75% over the horizon (bond tent), wired as a
  central-case toggle via a shared `centralSource` helper.
  `decumul.AnnuityFactor`/`AnnuityIncome` + an `annuityShare` control: a share of
  capital buys a joint-life inflation-linked income floor (premium out of
  capital, lifelong cashflow). Honest framing on both: they trade growth for
  security, so under the tool's full-need ruin metric they can *raise* headline
  ruin (Cederburg; and the user declined the utility/floor metric that would
  reward them), which the tool shows rather than hides.

## Remaining

- **E (remaining) — chart-system reskin.** The new primitives (Gauge, Scatter,
  CategoryBars) match the mock; the older charts (Fan, StackedArea, Bars,
  Frontier) are competent but could get the same finish pass. Lower priority:
  the page frame already equals the mock's system.
- **Refinements.** RuinTiming is a timing proxy for cause; a wealth-trajectory
  classifier (crash depth vs grind) would be more precise. The frontier's
  spending-CV for the fixed rule is inflated by post-ruin zeros; a
  solvency-conditional CV would separate ruin from lifestyle swing more cleanly.

## Design mock

`fire-desk-study` artifact (the dense reimagined page) and `fire-new-charts-live`
(the shipped gauge + frontier) are the visual reference.
