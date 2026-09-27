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
`res.Feasible` is false when no allocation meets the limits: the weights are
then the least-violating point found, not an answer, so always check it.

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
  command's `train:` constraint fits on one slice and measures on the rest;
  in code, fit on `Panel.Between` of the training years and measure on the
  whole panel.
- **Corners.** Without bounds, the optimum usually sits on a corner, the least
  durable answer there is.

The design and its measurements: [`docs/specs/weight-search-design.md`](../../specs/weight-search-design.md)
and [`docs/specs/black-litterman-design.md`](../../specs/black-litterman-design.md).
The script `optimize.go` in [`examples/code/`](../../../examples/code/README.md)
runs an objective on a portfolio file.
