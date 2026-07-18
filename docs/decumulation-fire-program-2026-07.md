# FIRE program: state and open items

The 2026-07 improvement program (what the FIRE engine/report lacked versus
the state of the art) is fully shipped except one item. The code and its
tests are the reference for what exists; highlights of what the program
added, so nobody re-proposes them:

- Broad-sample empirical model (`scenario.PooledBootstrap` over the JST
  per-country panel, `pkg/datasets/broadsample`), anchored to the
  Anarkulova/Cederburg band and locked by `broadsample_test.go`; the
  column resamples a within-country 60/40 mix, matching the anchor it
  cites (pure-equity pooling silently stressed the allocation on top of
  the data).
- CAPE valuation anchoring (`pkg/datasets/cape`, `make cape`, multpl
  fallback, staleness chip) feeding the central return and ABW's assumed
  return.
- The six spending rules incl. VPW, guardrails with an incompressible
  floor (monthly-stepped in the monthly kernel), ABW/TPAW with
  after-tax-liquidation amortization and pension PV folded in.
- Historical replays (USA 1929/1966/2000, Japan 1990), decisive-decade
  decomposition, ruin causes by trajectory shape, income layers, the
  policy frontier, per-model market fans.
- Glidepath and partial annuity as central-case toggles, framed honestly:
  under the full-need ruin metric they can raise headline ruin
  (Cederburg), which the tool shows rather than hides; a utility/floor
  metric that would flatter them was deliberately rejected.
- The model strip as selector, hover/crosshair + keyboard + table view on
  every chart, in-product mechanics explanations.

## Open

- CAPE-conditioned spending rule (WR = a + b/CAPE): needs a per-path
  valuation model, which no scenario source simulates today; its
  planning-time content is already covered by the CAPE anchor feeding
  ABW's assumed return. Revisit only with a valuation-path model.
