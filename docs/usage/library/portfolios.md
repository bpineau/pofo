# Portfolios and simulation

A portfolio study takes a spec (holdings and weights), fetches each series,
simulates the rebalanced portfolio day by day, and measures it.
`analyze.Portfolio` does it all in one call; `pkg/portfolio` exposes each
step.

## A portfolio in one call

`portfolio.NewSpec` builds a spec in code, with weights as **fractions**.
`analyze.Portfolio` returns the numbers the command's report is drawn from.
`src` is offline here; live, it is a `*marketdata.Client`.

```go
// from analyze.Example_sixtyForty (synthetic: newFake serves made-up series)
ctx := context.Background()
src := newFake()

spec, _ := portfolio.NewSpec("60/40",
	portfolio.Line{ID: "IWDA", Weight: 0.6}, // IE00B4L5Y983
	portfolio.Line{ID: "AGGH", Weight: 0.4}) // IE00BDBRDM35
study, err := analyze.Portfolio(ctx, src, spec, analyze.Options{Currency: "EUR"})
if err != nil {
	panic(err)
}

st := study.Stats
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
fmt.Printf("correlation %s/%s: %.2f\n", study.Aligned.IDs[0], study.Aligned.IDs[1], study.Correlation[0][1])
fmt.Printf("risk budget: %.0f %% / %.0f %% of the variance\n", study.Attribution.Risk[0]*100, study.Attribution.Risk[1]*100)
```

```text
CAGR 7.8 %, volatility 9.2 %, max drawdown -14.1 %
correlation IWDA/AGGH: -0.16
risk budget: 101 % / -1 % of the variance
```

The study (`analyze.PortfolioStudy`) holds:

| Field | What |
|---|---|
| `Stats`, `Years`, `Months`, `Drawdowns` | the portfolio's statistics, calendar tables and drawdown episodes |
| `Holdings` | each holding studied on the same window |
| `Aligned`, `Correlation` | the holdings on one calendar, and their correlation matrix |
| `Attribution` | each holding's share of the risk and of the return |
| `Composition` | the look-through: asset classes, geography, currencies, sectors |
| `Warnings` | simulated spans, distributing classes, unconverted currencies |

The risk budget above is the point of the example: a 60/40 whose bonds
barely correlate with its equities carries nearly all its variance in the
equities.

## From a file

`portfolio.Parse` reads the [file format](../portfolio-files.md), including
`#meta` directives; `portfolio.ParseFile` opens a path.

```go
// from portfolio.ExampleParse
spec, err := portfolio.Parse("my-portfolio", strings.NewReader(`
# Comment lines and blank lines are ignored.
60   VTI    0.03            # optional TER, then a free-text comment
25,5 IE00B4L5Y983           # decimal comma accepted
14.5 GLD
`))
if err != nil {
	panic(err)
}
for _, h := range spec.Holdings {
	fmt.Printf("%5.1f %% %s\n", h.Weight*100, h.ID)
}
```

```text
 60.0 % VTI
 25.5 % IE00B4L5Y983
 14.5 % GLD
```

## Step by step

What `analyze.Portfolio` wires, one call at a time: `Build` fetches each
holding through your callback, `Simulate` rebalances every N days. With
flows, `SimResult.Index` is the time-weighted series (for statistics) and
`SimResult.Values` follows the money (for the IRR).

```go
// from portfolio.Example_byHand (synthetic: a fetch callback serving made-up series)
spec, _ := portfolio.NewSpec("60/40",
	portfolio.Line{ID: "IWDA", Weight: 0.6}, // IE00B4L5Y983
	portfolio.Line{ID: "AGGH", Weight: 0.4}) // IE00BDBRDM35
p, err := portfolio.Build(spec, portfolio.BuildOptions{Fetch: synthetic})
if err != nil {
	panic(err)
}
p.Capital = 10_000
p.Contribute = portfolio.Flow{Amount: 500, Period: portfolio.Monthly}
sim, err := portfolio.Simulate(p, 90) // rebalance every 90 days
if err != nil {
	panic(err)
}

stats, _ := metrics.Compute(sim.Dates, sim.Index) // the strategy, flows stripped out
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", stats.CAGR*100, stats.Volatility*100, stats.MaxDrawdown*100)

// The saver's own rate: money going in is negative, the final value closes the account.
dates, flows := []time.Time{sim.Dates[0]}, []float64{-p.Capital}
var booked []metrics.Flow // the same flows the other way round, for TWR
for i, d := range sim.FlowDates {
	dates, flows = append(dates, d), append(flows, -sim.FlowAmounts[i])
	booked = append(booked, metrics.Flow{Date: d, Amount: sim.FlowAmounts[i]})
}
last := len(sim.Dates) - 1
irr, _ := metrics.IRR(dates, flows, sim.Dates[last], sim.Values[last])
twr, _ := metrics.TWR(sim.Dates, sim.Values, booked)
fmt.Printf("put in %.0f, worth %.0f, money-weighted %.1f %%/yr\n", p.Capital+sim.Contributed, sim.Values[last], irr*100)
fmt.Printf("time-weighted %+.1f %%, as the index says: %+.1f %%\n", twr*100, sim.Index[last]-100)
```

```text
CAGR 10.7 %, volatility 11.4 %, max drawdown -5.3 %
put in 27500, worth 33664, money-weighted 10.2 %/yr
time-weighted +35.7 %, as the index says: +35.7 %
```

For live data, the fetch callback is usually
`client.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "EUR"})`.
Fund fees (TER) are already in the prices and never deducted;
`Portfolio.EnvelopeFees` (percent per year) is.

## What the portfolio is missing

`pkg/suggest` reads what holdings **are** before what they returned: regime
coverage and gaps, redundancies, and candidates validated on walk-forward
windows. It is what `pofo -coverage` and `pofo -suggest` run.

```go
// from suggest.ExampleAnalyze (synthetic returns)
const n = 500
held := make([]float64, n)
diversifier := make([]float64, n)
for i := range n {
	held[i] = 0.004 * math.Sin(float64(i)/5)
	diversifier[i] = 0.004*math.Cos(float64(i)/5) + 0.0003
}
holdings := []suggest.Holding{
	{ID: "IWDA", Weight: 1, HasMeta: true, Meta: suggest.Meta{AssetClass: "equity"}}, // IE00B4L5Y983
}
candidates := []suggest.Candidate{{
	Meta:        suggest.Meta{ID: "IGLN", AssetClass: "gold"}, // IE00B4ND3602
	PortReturns: held,
	Returns:     diversifier,
	Years:       12,
}}

res := suggest.Analyze(holdings, [][]float64{held}, candidates, suggest.DefaultOptions(), suggest.RegimeFramework())
fmt.Println("gaps:", res.Gaps)
for _, s := range res.Suggestions {
	fmt.Printf("%s at %.0f %% fills %s (%d/%d windows)\n",
		s.Meta.ID, s.Weight*100, s.Fills, s.SharpeWins, s.Windows)
}
```

```text
gaps: [deflation inflation crisis]
IGLN at 20 % fills inflation (8/8 windows)
```

The look-through splits (`AssetClassSplit`, `CurrencySplit`,
`DurationSplit`...) are what fills `PortfolioStudy.Composition`.
`compare.Sweep` is the library side of `pofo -sweep`.

Scripts: `correl.go`, `simulate.go` in
[`examples/code/`](../../../examples/code/README.md).
