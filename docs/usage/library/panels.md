# Panels, blends and regression

A question across several series (a blend, a regression, what gold did in
equity's worst months) needs their returns on one shared calendar.
`marketdata.NewPanel` builds that table: the periods all series share, cut on
a calendar you choose, labelled canonically (the month-end for `Monthly`).
Every column of a period then spans the same interval.

This is the table a notebook would build, with the return math already
tested. Keep ad hoc analysis in Go over a panel rather than redoing the math
elsewhere.

## The whole recipe

```go
// from marketdata.Example_explore (synthetic: threeLegs builds three made-up series)
// Monthly returns on the months all three share, labelled month-end.
p, err := marketdata.NewPanel(marketdata.Monthly, threeLegs()...)
if err != nil {
	panic(err)
}
fmt.Printf("%d months to %s\n", p.Len(), p.Ends[p.Len()-1].Format(time.DateOnly))

// A 60/40 rebalanced every month, as a level, then its statistics.
p, err = p.Mix("60/40", map[string]float64{"IWDA": 0.6, "AGGH": 0.4})
if err != nil {
	panic(err)
}
blend, err := p.Series("60/40")
if err != nil {
	panic(err)
}
st, err := blend.Stats()
if err != nil {
	panic(err)
}
fmt.Printf("60/40: CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)

// Gold regressed on equities: beta, t-statistic, annualized alpha.
eq, err := p.Col("IWDA")
if err != nil {
	panic(err)
}
gold, err := p.Col("IGLN")
if err != nil {
	panic(err)
}
reg, err := metrics.Regress(gold, eq)
if err != nil {
	panic(err)
}
fmt.Printf("IGLN on IWDA: beta %.2f (t %.1f), alpha %+.1f %%/yr, R2 %.2f\n",
	reg.Betas[0].Value, reg.Betas[0].T, reg.AnnualAlpha(p.PeriodsPerYear())*100, reg.R2)

// The three worst equity months, dated.
for _, t := range metrics.LowestK(eq, 3) {
	fmt.Printf("%s IWDA %+.1f %%\n", p.Ends[t].Format("2006-01"), eq[t]*100)
}

// What gold did in the worst tenth of equity months.
worst, err := p.Pick(metrics.LowestK(eq, p.Len()/10))
if err != nil {
	panic(err)
}
fmt.Printf("worst %d months, on average:", worst.Len())
for _, id := range []string{"IWDA", "IGLN"} {
	col, err := worst.Col(id)
	if err != nil {
		panic(err)
	}
	fmt.Printf(" %s %+.1f %%", id, metrics.Mean(col)*100)
}
fmt.Println()
```

```text
119 months to 2024-12-31
60/40: CAGR 7.4 %, volatility 8.2 %, max drawdown -7.1 %
IGLN on IWDA: beta -0.29 (t -4.7), alpha +6.3 %/yr, R2 0.16
2020-07 IWDA -7.6 %
2023-08 IWDA -6.6 %
2016-10 IWDA -6.5 %
worst 11 months, on average: IWDA -6.3 % IGLN +4.8 %
```

Step by step:

| Step | Call | Note |
|---|---|---|
| build the table | `NewPanel(freq, series...)` | `Daily`, `Monthly`, `Quarterly` or `Yearly`; a panel converts no currency, so convert first |
| add a blend | `Panel.Mix(name, weights)` | rebalanced every period; weights are fractions summing to 1; financing is an explicit cash column |
| read a column | `Panel.Col(id)` | returns, fractions, per period |
| back to a level | `Panel.Series(id)` | for `Stats` and every level statistic |
| regress | `metrics.Regress(y, x...)` | per-period alpha; `AnnualAlpha` annualizes it |
| rank periods | `metrics.LowestK`, `metrics.HighestK` | positions, so dates come with them |
| select periods | `Panel.Pick`, `Panel.Between` | the other columns on exactly those periods |

A regression's alpha is per period. Pass `Panel.PeriodsPerYear()` to
`AnnualAlpha` rather than guessing 12 or 252.

## One series: a fee, an episode

`Series.LessFee` charges a yearly fee, as a **fraction** (`0.0085` is
0.85 %/yr):

```go
// from marketdata.ExampleSeries_LessFee (synthetic)
start := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
index, _ := marketdata.NewSeries("MSCIWORLD", []time.Time{start, start.AddDate(8, 0, 0)}, []float64{100, 200})
fund, err := index.LessFee(0.0085)
if err != nil {
	panic(err)
}
fmt.Printf("index %.1f, fund %.1f\n", index.Last().Close, fund.Last().Close)
```

```text
index 200.0, fund 186.8
```

`Series.Change` measures a named episode between two dates:

```go
// from marketdata.ExampleSeries_Change (synthetic)
d := func(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }
s, _ := marketdata.NewSeries("SP500", []time.Time{d(2007, 12, 31), d(2008, 10, 10), d(2008, 12, 31), d(2009, 3, 9)},
	[]float64{100, 65, 63, 52})
c, err := s.Change(d(2008, 1, 1), d(2008, 12, 31))
if err != nil {
	panic(err)
}
fmt.Printf("2008: %+.0f %%\n", c*100)
```

```text
2008: -37 %
```

Scripts that do this on real data: `blend.go`, `regress.go`,
`worstmonths.go`, `episodes.go` and `fees.go` in
[`examples/code/`](../../../examples/code/README.md).
