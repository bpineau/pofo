# Optimization

`pkg/optimize` computes long-only weights for an objective, under per-line
bounds and portfolio limits. It is what `#meta optimize:` runs; the
[portfolio file guide](../portfolio-files.md#optimizing-weights) explains the
objectives and how to read their answers.

`optimize.Solve` takes three things:

- aligned returns, one slice per asset (`[asset][period]`);
- their cadence in periods per year: `metrics.TradingDaysPerYear` for daily
  returns, `metrics.PeriodsPerYear` of their dates in general;
- a `Spec`: the objective, the bounds and the limits.

## The most return under a volatility cap

```go
// from optimize.ExampleSolve_underAVolatilityCap (synthetic returns)
// A hot asset and a calm one, anti-correlated enough to blend well.
hot := make([]float64, 500)
calm := make([]float64, 500)
for i := range hot {
	swing := 0.012 * math.Sin(float64(i)*0.25)
	hot[i] = 0.0009 + swing
	calm[i] = 0.0003 - swing
}
res, err := optimize.Solve([][]float64{hot, calm}, metrics.TradingDaysPerYear, optimize.Spec{
	Objective: optimize.MaxReturn,
	MinWeight: 0.05, // keep both lines in the book
	Limits:    optimize.Limits{MaxVolatility: 0.08},
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("hot %.0f %%, calm %.0f %%, feasible %v\n",
	res.Weights[0]*100, res.Weights[1]*100, res.Feasible)
```

```text
hot 80 %, calm 20 %, feasible true
```

In a `Spec` built in code, everything is a **fraction** (`0.08` is 8 %/yr).
`res.Feasible` is false when the returned weights miss a limit. The heuristic
did not find a feasible answer, so always check it.

## The file's grammar, in code

`optimize.ParseSpec` reads the same text as `#meta optimize:`, in percent as
written. `Resolve` binds the identifiers it names to the columns, and
`black-litterman` takes the written weights as its `Prior`:

```go
// from optimize.ExampleSolve_boundedBlackLitterman (synthetic: exampleReturns builds daily returns)
spec, err := optimize.ParseSpec("black-litterman,view:TREND:8@70,bounds:TREND:10-40,max-vol:9")
if err != nil {
	log.Fatal(err)
}
if err := spec.Resolve([][]string{{"EQUITY"}, {"TREND"}, {"CASH"}}); err != nil {
	log.Fatal(err)
}
spec.Prior = []float64{0.5, 0.3, 0.2} // the weights written in the file

res, err := optimize.Solve(exampleReturns(750), metrics.TradingDaysPerYear, spec)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("EQUITY %.0f %%, TREND %.0f %%, CASH %.0f %%, feasible %v\n",
	res.Weights[0]*100, res.Weights[1]*100, res.Weights[2]*100, res.Feasible)
```

```text
EQUITY 51 %, TREND 40 %, CASH 9 %, feasible true
```

The view says trend earns 8 %/yr at 70 % confidence, so the weights move
toward trend until its bound stops them at 40 %.

## Traps

- **Zero risk-free rate.** Sharpe here subtracts nothing, so a cash-like line
  buys ratio for free. A volatility cap usually asks the intended question
  better than `max-sharpe`.
- **In-sample.** Weights fitted on a window describe that window. The
  command's `train:` constraint fits on one slice and reports a held-out
  stretch separately. In code, use disjoint panels with `FitEvaluate`.
- **Corners.** Without bounds, the optimum usually sits on a corner, the least
  durable answer there is.

The design and its measurements: [`docs/specs/weight-search-design.md`](../../specs/weight-search-design.md)
and [`docs/specs/black-litterman-design.md`](../../specs/black-litterman-design.md).
The script `optimize.go` in [`examples/code/`](../../../examples/code/README.md)
runs an objective on a portfolio file.

## A measured frontier

`Frontier` reuses `Solve` with `MaxReturn` at each volatility cap in an
explicit grid. It returns every requested point in grid order, including
infeasible points, with its `Result` (weights and feasibility) and `Measured`
path performance. Supply a `MaxReturn` spec with `Limits.MaxVolatility` unset;
weight bounds and other limits still apply.

```go
// from optimize.ExampleFrontier (synthetic returns)
returns := [][]float64{{.03, -.02, .04, -.01, .02, -.02}, {.003, -.001, .004, -.002, .003, -.001}}
points, err := optimize.Frontier(returns, 12, optimize.Spec{Objective: optimize.MaxReturn}, []float64{.0001, .2})
if err != nil {
	panic(err)
}
for _, p := range points {
	fmt.Printf("cap %.2f %%, feasible %v\n", p.MaxVolatility*100, p.Result.Feasible)
}
```

These are candidate points from a heuristic search. A false `Feasible`
flag says the returned weights miss a limit; it does not establish that
no feasible allocation exists. A true flag describes the fitted historical
window, with no guarantee about future risk.

## Fit once, evaluate separately

`FitEvaluate(training, heldOut, periodsPerYear, spec)` fits weights on the
training panel only, then returns `Fit`, `InSample` and `OutOfSample`.
The measured fields contain CAGR, volatility, maximum drawdown (negative)
and Sharpe (zero risk-free rate). This keeps Black-Litterman's expected
figures in `Fit` distinct from realized performance.

Both panels use `[asset][period]`, the same asset order and currency, and
the same cadence. They may have different observation counts. Select disjoint
date windows with `marketdata.Panel.Between` before passing their return
rows. This package cannot detect overlapping dates or reordered assets.
`FitEvaluate` does not accept CWARP, which also needs replacement returns.

```go
// from optimize.ExampleFitEvaluate (synthetic return panels)
training := [][]float64{{.02, -.02, .02, -.02}, {.01, .01, -.01, -.01}}
heldOut := [][]float64{{-.03, -.03, .02, .02}, {.01, -.01, .01, -.01}}
result, err := optimize.FitEvaluate(training, heldOut, 12, optimize.Spec{Objective: optimize.MinVolatility})
if err != nil {
	panic(err)
}
fmt.Printf("training weight A %.0f %%\n", result.Fit.Weights[0]*100)
fmt.Printf("in sample CAGR %.2f %%, out of sample CAGR %.2f %%\n", result.InSample.CAGR*100, result.OutOfSample.CAGR*100)
```

Use `Evaluate(returns, weights, periodsPerYear)` to measure an existing
long-only allocation without fitting. Measurements assume rebalancing every
period, with no fees or flows. A zero-variance Sharpe is reported as zero.

`Solve` and `SolveCWARP` reject a nonzero `Spec.Train`. After applying a
parsed `train:` window, clear that field on the solver copy. This makes a
forgotten date selection an error. The command's report handles it for you.

## Unequal risk budgets

`Spec.RiskBudgets` sets target Euler shares of total risk for `RiskParity`,
in asset order. Supply strictly positive finite fractions summing to 1
(tolerance 1e-6). Nil keeps equal shares. Zero budgets are rejected: remove
that asset from the input panel instead. Assets with zero sample variance
are rejected. The field is a library option, absent from the text grammar.

```go
// from optimize.ExampleSolve_riskBudgets (synthetic returns)
returns := [][]float64{{.01, .01, -.01, -.01}, {.02, -.02, .02, -.02}}
result, err := optimize.Solve(returns, 12, optimize.Spec{Objective: optimize.RiskParity, RiskBudgets: []float64{.8, .2}})
if err != nil {
	panic(err)
}
fmt.Printf("weights %.0f / %.0f %%\n", result.Weights[0]*100, result.Weights[1]*100)
```

Risk parity preserves its legacy risk-budget weights, then checks caps,
floors and per-asset bounds. Violations set `Feasible=false` with an explicit
`FeasibilityReason`; satisfied bounds remain feasible. Portfolio feasibility
limits are rejected, as are risk budgets on any other objective.

## An explicit risk-free rate

For `MaxSharpe` and `MaxSortino`, `Spec.RiskFree` supplies a finite annual
arithmetic fraction, for example `0.03` for 3 %/yr. The optimization and
reported ratios both use it. Negative rates are valid; zero preserves the
historical default. Other objectives reject a nonzero rate. This option is
library-only and does not change the text grammar.

`Evaluate` retains a zero risk-free rate. Use
`EvaluateWithRiskFree(returns, weights, periodsPerYear, riskFree)` for an
explicit measurement rate. `FitEvaluate` uses `Spec.RiskFree` for the fit and
the two measured windows. CAGR, volatility and drawdown remain nominal.

## Input errors and JSON

The public solvers validate requests built directly in Go, independently of
`ParseSpec`. Returns must be finite and greater than -1, with at least two
observations per asset and a positive finite cadence. Constraints must be
finite and within their documented ranges. Resolve named bounds and views
before solving. Unsupported combinations return errors, including limits,
floors and per-asset ranges on CWARP.

`Spec` and `Result` marshal undefined floats as JSON `null`. This includes
open bounds, represented internally by NaN. Their JSON is for reporting,
not specification round trips: keep the original spec or parse its text again
before solving, because JSON null does not decode back to a NaN sentinel.
