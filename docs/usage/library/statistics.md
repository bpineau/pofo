# Statistics

`pkg/metrics` computes risk and return statistics on plain slices of dates
and values, so it works on any series, including one you built yourself.
`Series.Stats` is the shortcut from a loaded series, and `analyze.Asset`
gathers everything about one identifier in one call.

## The statistics of a series

```go
// from marketdata.ExampleSeries_Stats (synthetic: threeLegs builds three made-up series)
iwda := threeLegs()[0]
st, err := iwda.Stats()
if err != nil {
	panic(err)
}
fmt.Printf("%s, %s to %s, %.0f periods a year\n", iwda.Symbol,
	st.Start.Format(time.DateOnly), st.End.Format(time.DateOnly), st.PeriodsPerYear)
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, Sharpe %.2f, max drawdown %.1f %%\n",
	st.CAGR*100, st.Volatility*100, st.Sharpe, st.MaxDrawdown*100)
```

```text
IWDA, 2015-01-01 to 2024-12-31, 252 periods a year
CAGR 10.7 %, volatility 10.8 %, Sharpe 0.97, max drawdown -14.2 %
```

`metrics.Stats` also holds Sortino, the Ulcer Index, the longest time to
recovery, skew and kurtosis. Every figure is a fraction (`0.107` is 10.7 %),
except `Ulcer` (percent points) and `CWARP` (percent).

## Cadence: daily, weekly, monthly

Volatility and every ratio annualize at the series' own cadence, which
`metrics.Compute` measures from the dates: 252 periods a year on daily closes,
52 on weekly ones, 12 on month-ends. A monthly index therefore compares with a
daily fund without any flag:

```go
// from metrics.ExampleCompute_monthly (synthetic)
var dates []time.Time
var values []float64
v := 100.0
for m := range 120 { // ten years of month-end closes
	dates = append(dates, time.Date(2010, time.Month(m+2), 0, 0, 0, 0, 0, time.UTC))
	values = append(values, v)
	if m%2 == 0 {
		v *= 1.03
	} else {
		v *= 0.97
	}
}
stats, err := metrics.Compute(dates, values)
if err != nil {
	panic(err)
}
fmt.Printf("%.0f periods a year, volatility %.1f %%/yr\n", stats.PeriodsPerYear, stats.Volatility*100)
```

```text
12 periods a year, volatility 10.4 %/yr
```

Daily and monthly volatility of the same asset can still differ, because
daily moves partly reverse (or trend) within a month. The report prints both.
A fund quoted weekly has its daily statistics wrong by about the square root
of five; read its monthly figures.

## Everything about one asset

`analyze.Asset` studies one identifier on its longest window: statistics,
calendar years and months, drawdown episodes, and relative statistics against
a benchmark. `src` is any `analyze.Source`; a `*marketdata.Client` is one.

```go
// from analyze.ExampleAsset (synthetic: newFake serves made-up series)
ctx := context.Background()
src := newFake()

a, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{
	Currency:  "EUR",
	Benchmark: "MSCIWORLD",
	From:      time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
})
if err != nil {
	panic(err)
}
fmt.Printf("%s (%s), %s to %s, in %s\n", a.ID, a.Meta.Name,
	a.Stats.Start.Format(time.DateOnly), a.Stats.End.Format(time.DateOnly), a.Series.Currency)
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, beta %.2f\n", a.Stats.CAGR*100, a.Stats.Volatility*100, a.Relative.Beta)
for _, y := range a.Years[:2] {
	fmt.Printf("%d: %+.1f %% (partial: %v)\n", y.End.Year(), y.Return*100, y.Partial)
}
fmt.Printf("%d drawdown episodes, the deepest %.1f %%\n", len(a.Drawdowns), a.Stats.MaxDrawdown*100)
```

```text
IWDA (iShares Core MSCI World UCITS ETF USD (Acc)), 2015-01-01 to 2019-12-31, in EUR
CAGR 13.1 %, volatility 15.4 %, beta 0.92
2015: +28.6 % (partial: true)
2016: +17.3 % (partial: false)
37 drawdown episodes, the deepest -23.9 %
```

A study never hides what it cannot know: its `Warnings` name simulated spans,
distributing share classes and unconverted currencies.

## Holding periods

`metrics.RollingCAGRs` returns every N-year holding period, dated, ending on
each anniversary. It answers "what did holding it ten years deliver, at worst,
and when?":

```go
// from metrics.ExampleRollingCAGRs (synthetic)
// Yearly closes through a bust and a recovery.
var dates []time.Time
for y := 2000; y <= 2006; y++ {
	dates = append(dates, time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC))
}
values := []float64{100, 110, 55, 60, 110, 121, 133}

windows := metrics.RollingCAGRs(dates, values, 2)
worst, lost := windows[0], 0
for _, w := range windows {
	if w.CAGR < worst.CAGR {
		worst = w
	}
	if w.CAGR < 0 {
		lost++
	}
}
fmt.Printf("%d two-year windows, %d lost money; the worst ran %d to %d at %.1f %%/yr\n",
	len(windows), lost, worst.Start.Year(), worst.End.Year(), worst.CAGR*100)
```

```text
5 two-year windows, 2 lost money; the worst ran 2001 to 2003 at -26.2 %/yr
```

## More in pkg/metrics

| Question | Call |
|---|---|
| calendar years or months | `CalendarReturns` (the first period flagged `Partial`) |
| every drawdown, dated | `DrawdownEpisodes`, `MaxDrawdown` |
| tail losses | `VaR`, `CVaR` (positive per-period losses) |
| rolling risk | `RollingBeta`, `RollingCorr` |
| money-weighted and time-weighted returns | `IRR`, `TWR` |
| a value with deposits and sales: its flow-neutral growth, period table and drawdown | `FlowIndex`, `Report` (inflows booked at the start of the day, outflows at its close) |
| a return quantile | `Quantiles` |

Scripts that put these together: `describe.go`, `stats.go`, `rolling.go`,
`calendar.go`, `episodes.go` in
[`examples/code/`](../../../examples/code/README.md).

## Explicit covariance shrinkage

`ShrinkageCovariance(returns, shrinkage)` keeps sample variances and multiplies
cross-covariances by `1-shrinkage`. Its target is the sample diagonal, with a
caller-chosen coefficient from 0 to 1. It does not estimate an optimal
coefficient. Inputs are finite `[asset][period]` returns on one calendar,
with at least two observations. Output is per-period covariance: multiply by
the cadence to annualize it.

```go
// from metrics.ExampleShrinkageCovariance
// Sample variances are 1 and 4; sample covariance is 2.
cov, err := metrics.ShrinkageCovariance([][]float64{{1, 2, 3}, {2, 4, 6}}, .25)
if err != nil {
	panic(err)
}
fmt.Printf("variances %.1f / %.1f, covariance %.1f\n", cov[0][0], cov[1][1], cov[0][1])
```

## Uncertainty from historical blocks

`BlockBootstrap` returns point estimates and central percentile intervals for
CAGR, volatility and Sharpe. Specify cadence, block length, repetitions,
confidence and a fixed seed. It draws overlapping circular blocks from the
observed return sample, preserving dependence inside each block. A block
length of 1 gives independent draws; a full-length block only rotates the
original observations and cannot produce useful uncertainty for these
order-independent statistics.

```go
// from metrics.ExampleBlockBootstrap
// Synthetic monthly observations; two-month circular blocks retain
// consecutive pairs. This measures estimation uncertainty, not a forecast.
returns := []float64{.02, -.01, .03, .01, -.02, .04, .01, -.01, .02, .01, -.03, .02}
result, err := metrics.BlockBootstrap(returns, metrics.BootstrapOptions{
	PeriodsPerYear: 12, BlockLength: 2, Repetitions: 1000, Confidence: .9, Seed: 7,
})
if err != nil {
	panic(err)
}
fmt.Printf("CAGR %.2f %%, 90 %% interval [%.2f, %.2f] %%\n", result.CAGR.Estimate*100, result.CAGR.Lower*100, result.CAGR.Upper*100)
fmt.Printf("defined Sharpe replicates: %d/%d\n", result.Sharpe.Defined, result.Options.Repetitions)
```

These intervals describe estimation uncertainty under resampling of that
past, not future outcomes or retirement failure probabilities. Choosing a
block length remains the caller's responsibility. All samples have the
original length; no future data are used. Work grows with observations times
repetitions. Storage uses one sample buffer and three statistic slices,
with repetitions capped at `MaxBootstrapRepetitions` (100000).

`Interval.Estimate` is measured on the original sample; `Lower` and `Upper`
are interpolated percentiles of finite replicate estimates. `Defined` counts
those estimates. Constant returns have zero volatility and undefined Sharpe:
the Go result retains NaN and JSON writes `null`. Undefined replicates are
excluded only from their own statistic's interval, so inspect `Defined`
before interpreting a band. `BootstrapOptions.RiskFree` supplies an annual
arithmetic fraction for Sharpe, defaulting to zero; it changes neither CAGR
nor volatility.
