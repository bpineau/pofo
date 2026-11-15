# Preparing and solving FIRE scenarios

`pkg/fire` prepares the same annual models and plan settings used by the
simulator. `pkg/decumul` runs the withdrawal kernel. Both are independent of
HTTP and MCP, and neither needs a server's startup portfolio.

Money is real purchasing power at plan start; returns, taxes and weights are
fractions. `MarketSpec.Mu` is an arithmetic annual real mean, `Sigma` an annual
volatility, and `Nu` the Student-t degrees of freedom. `PlanSpec.Capital`
includes the cash buffer, and `AnnualSpending` includes pension-funded needs.

```go
// from fire.ExampleBuildPlan
spec := fire.Defaults()
spec.Capital, spec.AnnualSpending, spec.Years = 1000000, 30000, 30
plan, err := fire.BuildPlan(spec)
if err != nil {
	panic(err)
}
model, err := fire.PrepareModel(fire.Central, fire.MarketSpec{Mu: .04, Sigma: .12, Nu: 5}, spec.Years)
if err != nil {
	panic(err)
}
plan.Source = model.Source
ensemble, err := plan.SimulateContext(context.Background(), 100, 2, 7)
if err != nil {
	panic(err)
}
fmt.Println(model.Basis, len(ensemble.Paths), ensemble.Years)
```

The output is `parametric_assumptions 100 30`. `Defaults` supplies current
buffer and tax assumptions, including their labeled French taxable-account
scope through `TaxAssumptionLabel` and `TaxAssumptionYear`. It is not a legal
calculation or a frozen API contract. Start with a zero `PlanSpec` when no
such assumptions are wanted; explicit zero values are preserved. `BuildPlan`
leaves `Source` unset and permits an incomplete zero-capital/zero-horizon
template, which must be filled before simulation.

`Pension.StartYear` is zero-based: year zero pays immediately, while a start
at the horizon pays nothing. `RealDrift` compounds the first real annual
payment only after its start. The buffer is the lesser of capital and
`BufferYears * AnnualSpending`. `InitialGain` describes the starting gain
fraction; positive inflation erodes nominal cost basis even when that initial
fraction is zero.

## Portfolio history and model meaning

Pass already currency-converted nominal `AssetSeries` histories and an
inflation index to `BuildMonthlyPanel`. It retains the longest uninterrupted
common monthly stretch, excludes incomplete trailing months and reports
exclusions in `Coverage`. `FitParametric` derives annual assumptions from the
monthly real panel. Supply the panel, normalized weights and fitted values in
`MarketSpec`; `PrepareModel` rejects fewer than 24 months and does not silently
replace an unavailable model.

| Model ID | Basis |
|---|---|
| `central` | Student-t assumptions, calibrated when portfolio history is present |
| `bootstrap` | Stationary blocks of the portfolio's monthly real history |
| `historical` | Uniform sampling of complete contiguous cohorts, with replacement |
| `stress` | Mean-preserving regime model with clustered volatile bear years |
| `broad_sample` | External domestic 60/40 multi-country reference |
| `lost_decade` | Prolonged bear regime that lowers realized returns |

Every prepared source produces annual real returns. The history models
compound monthly scenarios before the annual withdrawal kernel. Requested
historical horizons must fit the panel. Available cohort count describes
windows in one finite record, not independent historical evidence.

The central model blends short histories toward its conservative prior.
`CAPE: true` bypasses that blend and replaces the mean in central, stress and
lost-decade models with the bundled CAPE-implied arithmetic return. It leaves
bootstrap, historical and broad-sample draws unchanged. Invalid CAPE prevents
an adjustment; a valid stale observation remains usable with an explicit
`Stale` flag. Prepared results identify raw/effective parameters and the input
basis. Broad-sample is not a simulation of the supplied allocation.

## Reusing draws and solving

Call `DrawContext` once, then reuse its `Draws` in `SimulateOnContext`,
`RuinProbOnContext` and `SolveOnContext`. Worker count does not affect draw
order or successful results. Cancellation discards partial results; custom
sources can supply a context-aware draw method as documented by
`scenario.DrawContext`.

Structured solvers evaluate both endpoints and verify their selected feasible
candidate. `solved` reports a crossing; `boundary_satisfies` reports only the
limiting search bound; `target_unreachable_in_bounds` has no solution.
The default tolerance is `max(10, 0.001 * abs(feasible candidate))` in amount
units. An explicit `SolveOptions` changes this numerical precision separately
from the sampled ruin probability and its descriptive Monte-Carlo error.

Solving assumes monotonicity. Fixed-real spending and capital axes support
buffers, pensions and gain taxes. Guyton-Klinger capital searches can reverse
even with correctly rebuilt initial bands, so compare adaptive capital
choices by simulation. A fixed-rule safe-spending comparison must be labeled
as a changed policy when the main simulation uses an adaptive rule.

`Outcome` and `SpendStats` retain undefined estimates as NaN in Go and null in
JSON. Empty samples are undefined, as are decade returns without a complete
positive-start decade, effective-tax ratios with no gross withdrawals, and
conditional cut metrics without cuts. Effective-tax medians exclude paths
with zero gross withdrawals; cumulative tax includes zero-tax paths.
`SpendCV` is undefined without lived spending observations or with zero mean.
For household income, use `Plan.HouseholdSpendBands` and
`Plan.HouseholdSpendCV`: these include deterministic pension and other
cashflows on the same fixed-horizon ensemble. The raw ensemble counterparts
measure portfolio-funded spending only. The household methods reject a
non-nil lifetime because its individual income paths are not recoverable
from the ensemble summary.
See the [FIRE guide](../fire.md) for broader kernel policies and web behavior.
