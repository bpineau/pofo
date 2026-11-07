# Currency, inflation and cash benchmarks

Run the complete [offline workflow](../../../examples/code/workflow.go) with
`go run examples/code/workflow.go -out /tmp/study.csv`.
It loads bundled references, converts to EUR, studies purchasing power with
French HICP, measures excess returns over a euro cash index, fits weights on a
training window, evaluates a later window, and exports three curves. The
portfolio simulation rebalances monthly with a 0.1 percent cost per traded
side; optimizer training and evaluation are explicitly frictionless. It runs
from another Go module too and needs neither a disk cache nor private helpers.

## Choose the economic questions separately

Currency conversion chooses the denomination of the investment. Inflation
chooses a consumption basket. A EUR investment can be deflated by a French or
another price index: the library does not silently equate those choices.
Use Client.Load with FetchOptions.Currency, then Series.Deflate, or pass a CPI
to analyze.Options.Inflation for a Real study beside the nominal one.

A CPI is a positive price-level index, not an inflation rate. The zero
AdjustmentOptions refuses coverage beyond the index. AllowExtrapolation holds
the first/last level and records a warning; MaxAge caps the age of observations
held forward inside the window. Deflation uses observations labelled on or
before each valuation date. It does not model publication lags or data vintages:
a historical decision backtest must supply data available on each decision date.

## Analyze caller-owned data

The following example requires only fmt, time, pkg/analyze and pkg/marketdata.
Its entire data construction is visible, with no test fixture helper.

```go
// from analyze.ExampleSeries
dates := []time.Time{
	time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
}
asset, err := marketdata.NewSeries("valuation", dates, []float64{100, 110, 121})
if err != nil {
	panic(err)
}
cpi, err := marketdata.NewSeries("basket", dates, []float64{100, 105, 110.25})
if err != nil {
	panic(err)
}
cash, err := marketdata.NewSeries("cash", dates, []float64{1, 1.02, 1.0404})
if err != nil {
	panic(err)
}
asset.Currency, cash.Currency = "EUR", "EUR"
study, err := analyze.Series(asset, nil, analyze.Options{Inflation: cpi, RiskFreeIndex: cash})
if err != nil {
	panic(err)
}
fmt.Printf("nominal %.2f, purchasing power %.2f, excess growth %.4f\n",
	study.Series.Last().Close, study.Real.Series.Last().Close, study.Excess.Series.Last().Close)
```

Series.Excess subtracts the cash index's arithmetic return in each interval,
then compounds the differences from one. Dividing the asset by cash is a
relative-wealth ratio, a different quantity. A loss of 100 percent or more in
the excess curve is an error. Known currency mismatches are refused. A
portfolio cash comparison requires all holdings in one known denomination,
even when Options.Currency is left empty for native-currency analysis.

RiskFreeIndex is a total-return cash index. It changes nominal Sharpe and
Sortino and adds an Excess study, leaving nominal CAGR, volatility and drawdown
unchanged. Alternatively RiskFree is an annual arithmetic fraction (0.03 for
3 percent), divided by cadence for each interval. The two inputs are exclusive.
Relative beta and capture remain nominal; scalar RiskFree adjusts Jensen alpha,
while a cash-index study's Relative remains the nominal zero-rate comparison.
For regression of excess returns, construct excess curves for both assets,
then build a shared panel and regress its columns. Real and Excess statistics
have a zero risk-free rate in their own units.

## Turn a quoted rate into cash

CashIndex accrues an annual nominal percent rate on actual calendar days,
compounded daily with an explicit denominator. Negative rates are allowed when
each day's capital remains positive. Changes between output dates are included.
The input must already be a nominal accrual rate: a Treasury bill discount yield
requires its instrument-specific conversion first. Published cash total-return
indices, such as EURCASH-EUR, avoid that conversion when available.

```go
// from marketdata.ExampleCashIndex
start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
end := start.AddDate(0, 0, 360)
rates, err := marketdata.NewSeries("deposit", []time.Time{start, end}, []float64{3.6, 3.6})
if err != nil {
	panic(err)
}
rates.Currency = "EUR"
cash, err := marketdata.CashIndex(rates, []time.Time{start, end}, marketdata.Actual360)
if err != nil {
	panic(err)
}
fmt.Printf("360 calendar days: %.4f %%\n", 100*(cash.Last().Close-1))
```

Leveraged-portfolio financing is another explicit choice: analyze.Options.Financing
is an annual percent rate series, FinancingID loads one through the Source,
including the prevailing observation before the study window.
Both unset preserve the historical ^IRX default and the spec's borrowing spread.
A missing explicitly requested series is an error; the legacy default retains
its warning and zero-rate fallback. Never pass a cash total-return index as
Financing.

For calendar rebalancing or execution costs, pass a complete
`analyze.Options.Simulation` with `portfolio.SimOptions`. A nonnil value
cannot be mixed with `Options.Rebalance` or a spec rebalancing directive;
zero settings mean no scheduled rebalance. TransactionCost is a fraction per
gross traded side. Costs and Turnover are available on the returned simulation.
The run starts invested: entry and terminal liquidation costs are excluded.

## Preserve qualifications

Display Series.Warnings and study.Warnings. FX conversions retain the date
before which the first known cross was held flat in FXExtrapolatedBefore.
Panels and their blends retain qualifications and refuse mixing known different
currencies. WriteCSV and ReadSeriesCSV preserve warnings, currency and explicit
cadence. Load reads a single-series export; a multiple-series export requires
ReadSeriesCSV and selection by Symbol.

Undefined ratios are NaN in Go and JSON null in Stats and Relative. A measured
zero remains zero. JSON is a reporting format, not a lossless round trip for
NaN-bearing optimizer specifications. CSV is the series interchange format.
