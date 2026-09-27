# Comparing series

Two kinds of comparison come up. Several series side by side (correlations,
calendar years, betas) need one calendar. One series against its reference
(a backcast against the real fund, a fund against its index, a refreshed file
against its previous version) needs a dedicated study: `analyze.Pair`.

## Several series on one calendar

`marketdata.AlignSeries` puts series on the dates where all of them quote.
It fails, naming the series, where the looser `Align` would forward-fill
zeros. `pkg/metrics` then reads the aligned slices.

```go
// from metrics.Example_oneCalendar (synthetic: threeFunds builds three made-up series)
a, err := marketdata.AlignSeries(threeFunds(), time.Time{}, time.Time{})
if err != nil {
	panic(err)
}
r := a.Returns()
corr, err := metrics.CorrelationMatrix(r)
if err != nil {
	panic(err)
}
cov, err := metrics.Covariance(r)
if err != nil {
	panic(err)
}
fmt.Printf("from %s: %s/%s %.2f, %s/%s %.2f\n", a.Dates[0].Format(time.DateOnly),
	a.IDs[0], a.IDs[1], corr[0][1], a.IDs[0], a.IDs[2], corr[0][2])
fmt.Printf("%s volatility %.1f %%/yr\n", a.IDs[0], math.Sqrt(cov[0][0]*252)*100)

// Partial flags a first year measured from the first quote; the last row
// ends on the last quote, which its End says.
years, err := metrics.CalendarReturns(a.Dates, a.Levels[0], 12)
if err != nil {
	panic(err)
}
for _, y := range years {
	fmt.Printf("year to %s %+5.1f %%, partial=%v\n", y.End.Format(time.DateOnly), y.Return*100, y.Partial)
}
if _, betas, ok := metrics.RollingBeta(a.Dates, a.Levels[1], a.Dates, a.Levels[0], 1); ok {
	fmt.Printf("one-year beta of %s on %s, last: %.2f\n", a.IDs[1], a.IDs[0], betas[len(betas)-1])
}
if v, ok := metrics.VaR(r[0], 0.95); ok {
	fmt.Printf("daily 95 %% VaR of %s: %.2f %%\n", a.IDs[0], v*100)
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

A covariance is per period: annualize it with the cadence (here 252 daily
returns a year). For monthly work, build a [panel](panels.md) instead and read
`Panel.R`, which is the same `[asset][period]` layout.

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
