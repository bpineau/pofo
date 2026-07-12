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

## Shipped (2026-07-11 critique + enrichment drop)

See `docs/superpowers/specs/2026-07-11-fire-critique-enrichment-design.md`
for the full critique. Highlights:

- **Portfolio-mode doom bug fixed.** `runFire` now honours `#meta sim:on`
  (the de-suffix campaign had starved the panel to <12 common months:
  µ=0/σ=0 fit, 58% central ruin, absurd bootstrap columns). Plus guards:
  `Fit.Valid`, `minPanelMonths` (24) for the data-driven columns.
- **§02 The retirements that actually happened.** Deterministic replay of
  USA 1929/1966/2000 and Japan 1990 (JST year-indexed) through the user's
  exact plan; graded verdict cards.
- **§03 The decisive decade.** `PathResult.Ret10` + `Ensemble.DecadeBuckets`:
  ruin by first-decade-return quintile, with the concentration cards.
- **§04 income layers.** `/api/income` median funding mix (also fixed the
  annuity income missing from the spending fan's overlay).
- **CAPE refreshed and self-aware.** gen-cape multpl fallback (mirror's PE10
  died 2023-09); bundle now 2026-07 (CAPE 42.2); `Stale` flag + UI chip.
- Dual-axis buffer chart split into two single-axis panels; M€ formatting;
  survivor-conditioned detail stats; largest-remainder cause shares.
- **Broad-sample column moved to the literature's 60/40 (2026-07-12).** The
  pool was 100% single-market equity while the anchor it cites (Anarkulova
  ~2.26% SWR) is a 60/40 domestic mix: the column silently stressed the
  allocation on top of the data. Now `broadSampleMixed`: within-country
  60/40 real returns, contiguous runs split at the bond record's war gaps.
  Measured (fixed rule, no tax, 1M, 30y): geo 4.31%->3.71%, 4%-rule ruin
  27.1%->22.7%, SWR@5% 1.59%->1.69%. Still stricter than Anarkulova (fixed
  horizon, JST-16 disaster-heavy pool); anchors locked in
  `broadsample_test.go`. Vintages (§02) stay pure equity, labelled.

## Shipped (2026-07-12 clarity + interaction drop)

- **The market each model imagines** (§01 first row): growth-of-1 fans per
  model + bear-texture captions (`/api/market`); an average path would hide
  the crashes, individual draws + cones show them.
- **Hover/crosshair layer** on every chart: Fan/MultiLine/StackedArea embed
  their data as SVG metadata, one delegated JS handler snaps a hairline and
  lists every series at that x; Bars/Scatter/CategoryBars carry native
  titles with oversized hit targets.
- **Ruin causes by trajectory shape** (`decumul.RuinShapes`: crash / grind /
  longevity by wealth-path shape) replacing the timing proxy in the view.
- **Solvency-conditional spending CV** on the policy frontier; scatter label
  deconfliction; spending-fan markers (pension start, side income end).
- **In-product explanations**: 'How this machine works' fold (model wiring
  table, toggle scopes, mechanics FAQ), rewritten df/monthly/stress hovers,
  fan axis capped at 10x start, M€ formats, 60/40 broad-sample column.

## Shipped (2026-07-12, guardrails session)

- **Guardrails floor** (`Guardrails.Floor`, gkFloor slider): bounds the cut
  spiral at an incompressible standard; honestly re-creates ruin. Measured
  at 3%/broad-sample: floorless p5 income decays 60k->5k->0; 75% floor
  holds p5 at 45k and brings broad-sample ruin back to ~18%.
- **Monthly-stepped guardrails** in the monthly kernel
  (`Guardrails.stepped`): monthly evaluation at (1-c)^12 == 1-Cut, no
  anniversary-date lottery. Annual kernel keeps the yearly G-K cadence.

## Shipped (2026-07-12, feedback batch)

- **The strip is the model selector**: click a column, every detail section
  re-runs under that lens (Params.Central + detailSource unify the old
  regime checkbox and parametric/bootstrap/cohorts selector); realism
  ladder in the fold.
- Wealth-based rules net cashflows (household-budget semantics; ABW folds
  the PV of future pensions in, the TPAW treatment): fixes Bounded % looking
  inert next to ABW.
- One generic tax slider (envelopes stay in the library); drawer masonry +
  scroll-on-open; planbar overflow fixed; nPaths default 4000; TPAW spelled
  out; recovery-histogram hover.

## Remaining

- **E (remaining) — chart-system reskin.** The older charts (Fan,
  StackedArea, Bars, Frontier) could get the same proportion/margin finish
  pass as the newer primitives. Lower priority: the page frame already
  equals the mock's system.
- Keyboard-focus parity for the chart tooltips (hover-only today); a table
  view of the chart data for the accessibility-strict reading of the
  dataviz spec.
- SOTA rules: ABW/TPAW and Vanguard bounded % SHIPPED 2026-07-12 (six-rule
  frontier; ABW amortizes the after-tax liquidation value via
  pocketOps.liquidationNet). Still open: a CAPE-conditioned spending rule
  (WR = a + b/CAPE) needs a per-path valuation model, which no source
  simulates today; its planning-time content is covered by the CAPE anchor
  feeding ABW's assumed return.

## Design mock

`fire-desk-study` artifact (the dense reimagined page) and `fire-new-charts-live`
(the shipped gauge + frontier) are the visual reference.
