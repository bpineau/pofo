# Comparing series

Two kinds of comparison come up. Several series side by side (correlations,
calendar years, betas) need one calendar. One series against its reference
(a backcast against the real fund, a fund against its index, a refreshed file
against its previous version) needs a dedicated study: `analyze.Pair`.

## Several series on one calendar

`marketdata.NewPanel` pairs returns over shared periods: each column covers
the same start and end dates. Use its `R` slices for cross-series statistics.
`AlignSeries` uses the forward-filled union calendar for valuation and
simulation; it does not select the intersection of quote dates. A panel is
the appropriate input for correlations and covariance.

```go
// from metrics.Example_oneCalendar (synthetic: threeFunds builds three made-up series)
p, err := marketdata.NewPanel(marketdata.Daily, threeFunds()...)
if err != nil {
	panic(err)
}
r := p.R
corr, err := metrics.CorrelationMatrix(r)
if err != nil {
	panic(err)
}
cov, err := metrics.Covariance(r)
if err != nil {
	panic(err)
}
fmt.Printf("from %s: %s/%s %.2f, %s/%s %.2f\n", p.Starts[0].Format(time.DateOnly),
	p.IDs[0], p.IDs[1], corr[0][1], p.IDs[0], p.IDs[2], corr[0][2])
fmt.Printf("%s volatility %.1f %%/yr\n", p.IDs[0], math.Sqrt(cov[0][0]*p.PeriodsPerYear())*100)

// Rebuild levels from the same paired returns for dated calculations.
first, err := p.Series(p.IDs[0])
if err != nil {
	panic(err)
}
second, err := p.Series(p.IDs[1])
if err != nil {
	panic(err)
}

// Partial flags a first year measured from the first quote; the last row
// ends on the last quote, which its End says.
years, err := metrics.CalendarReturns(first.Dates(), first.Values(), 12)
if err != nil {
	panic(err)
}
for _, y := range years {
	fmt.Printf("year to %s %+5.1f %%, partial=%v\n", y.End.Format(time.DateOnly), y.Return*100, y.Partial)
}
if _, betas, ok := metrics.RollingBeta(second.Dates(), second.Values(), first.Dates(), first.Values(), 1); ok {
	fmt.Printf("one-year beta of %s on %s, last: %.2f\n", p.IDs[1], p.IDs[0], betas[len(betas)-1])
}
if v, ok := metrics.VaR(r[0], 0.95); ok {
	fmt.Printf("daily 95 %% VaR of %s: %.2f %%\n", p.IDs[0], v*100)
}
```

```text
from 2020-07-20: IWDA/VUAA 0.88, IWDA/IGLN -0.16
IWDA volatility 11.2 %/yr
year to 2020-12-31  +5.1 %, partial=true
year to 2021-12-31  +8.0 %, partial=false
year to 2022-12-30 +11.3 %, partial=false
year to 2023-04-14  +1.9 %, partial=false
one-year beta of VUAA on IWDA, last: 1.10
daily 95 % VaR of IWDA: 0.95 %
```

A covariance is per period: annualize it with `Panel.PeriodsPerYear`, which
measures the shared calendar's cadence. For monthly work, build a Monthly
[panel](panels.md) and read the same `[asset][period]` layout in `Panel.R`.
`Panel.Series` rebuilds a column's levels on those same periods for calendar
returns and rolling statistics, refusing to chain across a gap.

## One series against its reference

`analyze.Pair` measures a candidate A against a reference B on the window
both cover:

- the level: CAGR gap with its standard error, the gap once both are rebased,
  the first date the ratio A/B moves;
- per cadence both series support (`Daily`, `Monthly`): correlation,
  volatilities, tracking error, beta, alpha, the largest divergences, dated;
- the calendar years side by side, and `Warnings` for anything that makes a
  figure unreliable (a short window, a cadence or currency mismatch, a change
  of definition in a series).

```go
// from analyze.ExamplePair (synthetic: backcastAndFund builds two made-up series)
backcast, fund := backcastAndFund()
st, err := analyze.Pair(backcast, fund, analyze.PairOptions{Divergences: 2})
if err != nil {
	panic(err)
}
fmt.Printf("%s to %s: CAGR gap %+.2f pt/yr (standard error %.2f)\n",
	st.Start.Format(time.DateOnly), st.End.Format(time.DateOnly), st.CAGRGap*100, st.GapSE*100)
m := st.Monthly
fmt.Printf("monthly: corr %.3f, tracking error %.2f %%, beta %.2f\n", m.Corr, m.TrackingError*100, m.Beta)
for _, d := range m.Divergences {
	fmt.Printf("%s: %+.2f %% against %+.2f %%\n", d.End.Format("2006-01"), d.A*100, d.B*100)
}
```

```text
2018-01-01 to 2023-12-29: CAGR gap +0.30 pt/yr (standard error 0.15)
monthly: corr 1.000, tracking error 0.38 %, beta 1.00
2018-08: +5.62 % against +5.39 %
2019-05: +0.48 % against +0.26 %
```

`PairStudy.WriteText` prints the whole study as aligned text, and the struct
marshals to JSON; `pofo -pair` is exactly that. When the two closes are struck
hours apart (a European listing against a US index), set
`PairOptions.LeadLag` so the daily divergences forgive a one-session offset,
and read the monthly block first.

To compare a bundled file with an older version, read the old one with
`marketdata.ReadCSV` over `git show HEAD~1:<path>` (see
[Loading data](loading.md#files-in-and-out)).

## Tracking inside a panel

Each cadence block of a pair is a `metrics.Tracking`. `Panel.Track` returns the
same for any two columns of a panel you built yourself:

```go
// from marketdata.ExamplePanel_Track (synthetic)
index := threeLegs()[0]
fund, err := index.LessFee(0.0020)
if err != nil {
	panic(err)
}
fund.Symbol = "FUND"
p, err := marketdata.NewPanel(marketdata.Monthly, fund, index)
if err != nil {
	panic(err)
}
tr, err := p.Track("FUND", "IWDA")
if err != nil {
	panic(err)
}
fmt.Printf("%d months: correlation %.3f, tracking difference %+.2f %%/yr, beta %.2f\n",
	tr.Periods, tr.Corr, tr.Difference*100, tr.Beta)
```

```text
119 months: correlation 1.000, tracking difference -0.22 %/yr, beta 1.00
```

`metrics.Track` does it for two bare return slices. Scripts: `pair.go`,
`oldnew.go`, `correl.go`, `currency.go` in
[`examples/code/`](../../../examples/code/README.md).

## Coefficient uncertainty under serial dependence

`Regress` reports classical OLS standard errors. `RegressHAC(y, lags, xs...)`
keeps the same coefficients and fit, replacing their standard errors and
t-statistics with the Newey-West Bartlett estimator. Samples must be in time
order on one regular cadence. Choose `lags` explicitly from 0 to N-1; zero
is heteroskedasticity-robust HC0. No degrees-of-freedom correction is applied
to the sandwich covariance. `ResidualSD` retains the ordinary residual
standard deviation with N-k-1 degrees of freedom.

```go
// from metrics.ExampleRegressHAC
// Synthetic ordered observations, with heteroskedastic and serial residuals.
x := []float64{-3, -2, -1, 1, 2, 3}
y := []float64{1.5, 3, -1.5, -.5, 5, 4.5}
fit, err := metrics.RegressHAC(y, 1, x)
if err != nil {
	panic(err)
}
fmt.Printf("alpha %.2f (HAC SE %.3f), beta %.2f (HAC SE %.3f)\n", fit.Alpha.Value, fit.Alpha.SE, fit.Betas[0].Value, fit.Betas[0].SE)
```

The standard errors describe asymptotic coefficient uncertainty under weak
dependence. They do not predict future returns or repair a misspecified mean
model. Units and annualization remain those of `Regression`; multiply an
alpha standard error by periods per year to annualize it arithmetically.
The implementation uses centered Householder QR and inverse triangular
factors. The weights follow equation (5) of the
[original Newey-West paper](https://users.ssc.wisc.edu/~behansen/718/NeweyWest1987.pdf).

## Attribute a period's active return

`Brinson` accepts aligned portfolio weights, benchmark weights, portfolio
sector returns and benchmark sector returns, all as fractions for one period.
It returns Brinson-Fachler allocation, selection and interaction, per sector
and in total. Both weight vectors sum to 1, within 1e-9; weights are finite
and nonnegative. Map the returned positions to sector names in the caller.

```go
// from metrics.ExampleBrinson
// Two aligned sectors, with a 60/40 portfolio and 50/50 benchmark.
attribution, err := metrics.Brinson(
	[]float64{.6, .4}, []float64{.5, .5},
	[]float64{.12, .02}, []float64{.10, .04},
)
if err != nil {
	panic(err)
}
fmt.Printf("active %.1f pt: allocation %.1f, selection cancels %v, interaction %.1f\n", attribution.ActiveReturn*100, attribution.Allocation*100, math.Abs(attribution.Selection) < 1e-12, attribution.Interaction*100)
```

The three effects reconcile to portfolio return minus benchmark return up to
floating-point rounding. Zero-weight sectors are supported, without division
by their weights. An absent sector still needs a finite return convention on
that side: the convention can shift the individual effects while leaving
the active total unchanged. This is single-period arithmetic attribution;
it does not link effects over multiple periods or attribute fees, flows or
currencies. The formulas are those of Brinson-Fachler, reproduced as equations
(10)-(12) in the CFA Institute's
[Performance Attribution review](https://rpc.cfainstitute.org/sites/default/files/-/media/documents/book/rf-lit-review/2019/rflr-performance-attribution.pdf).
