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
- The spending rules incl. VPW, guardrails with an incompressible
  floor (monthly-stepped in the monthly kernel), ABW/TPAW with
  after-tax-liquidation amortization and pension PV folded in.
- Risk-based guardrails (Kitces & Tharp, Morningstar; added 2026-07-25).
  Same architecture as the 2006 rule, different sensor: the band tracks
  the safe rate of the REMAINING horizon rather than a band around the
  initial rate, and the rate is read on total wealth (portfolio + PV of
  the cashflows still to come), so a pension about to start no longer
  triggers a cut. `decumul.RiskGuardrails` takes the safe-rate table as
  data; the web layer solves it (`riskband.go`) on a pension-free,
  unit-capital plan so one table serves any wealth level, at five
  interpolated anchors, cached per assumption set (~23 ms cold). It is
  the household's planning table: computed once under the central
  parametric assumptions and deliberately NOT re-derived per strip
  column, so the reader sees how a table written under their central
  case behaves inside a harsher world. Expressing the sensor in
  withdrawal-rate space rather than re-running a success probability
  inside every year of every path is what makes it affordable; it is the
  same signal read on the other side of the solve.
- Historical replays (USA 1929/1966/2000, Japan 1990), decisive-decade
  decomposition, ruin causes by trajectory shape, income layers, the
  policy frontier, per-model market fans.
- Glidepath and partial annuity as central-case toggles, framed honestly:
  under the full-need ruin metric they can raise headline ruin
  (Cederburg), which the tool shows rather than hides; a utility/floor
  metric that would flatter them was deliberately rejected.
- The model strip as selector, hover/crosshair + keyboard + table view on
  every chart, in-product mechanics explanations.

- Stochastic lifetime (2026-08-22). `Plan.Lifetime` draws the household's
  lifespan inside every path, so ruin is broke-WHILE-ALIVE counted rather
  than weighted, the estate at death is a first-class output, couples
  carry a survivor budget and a per-cashflow reversion, and `Plan.Annuity`
  realises real mortality credits. It is opt-in and moved no golden. The
  design, its calibration and its deferrals are
  `docs/stochastic-lifetime-kernel-design.md`.

## Open

- Wire the lifecycle view and the annuity toggle in `pkg/decumul/web` to the
  exact kernel: `LifeStates` in place of `LifeCurve`, the estate distribution
  as its own panel, and the annuity as a `Plan.Annuity` rather than a
  hand-built cashflow. The kernel is complete without it; this is UI work
  (layout, copy, the horizon-versus-planning-horizon control) and it was kept
  out of the kernel pass on purpose.
- CAPE-conditioned spending rule (WR = a + b/CAPE): needs a per-path
  valuation model, which no scenario source simulates today; its
  planning-time content is already covered by the CAPE anchor feeding
  ABW's assumed return. Revisit only with a valuation-path model.

- Envelope tax model: DECIDED 2026-09-10, not to build. A per-envelope tax
  model (CTO 31.4 %, PEA 18.6 %, assurance-vie 24.7 % under its allowance,
  PEE 18.6 %) is worth 0.015 point of sustainable withdrawal rate over a
  correctly blended single rate, and the drain order another 0.03; the rate's
  CALIBRATION is worth 0.12 and the embedded gain fraction 0.30. The kernel
  keeps the `Plan.Envelopes` machinery it already has; the blended slider stays
  the shipped control, with the calibration recipe (gain-weighted rate,
  capital-weighted gain fraction) and the PEE match break-even in
  `docs/fire-envelopes-tax-model-design.md`. Unlock dates (PEE and PEA 5 years,
  AV 8 years) and per-envelope allocations were refused there too.
- Cheap UI item that spike exposed: SHIPPED 2026-09-10. `Params.GainFrac`,
  `Params.PEACapital` and `Params.AVCapital` existed in
  `pkg/decumul/web/model.go` and reached the kernel, but none of the three was
  in the page's control list (`GROUPS` in `assets/app.js`), so the shipped page
  always ran a single sleeve with a ZERO embedded gain, which flattered the plan
  by 0.30 point of withdrawal rate. The Taxes group now carries the embedded
  gain (default 50 %, the spike's base case and the recipe's worked example
  rounded) beside the rate, and the two envelope amounts behind an `envelopes`
  disclosure, each with the recipe in its help text; both amounts at zero keeps
  the single blended sleeve, and the rate's help now says the rate applies to
  the gain share of every sale.
