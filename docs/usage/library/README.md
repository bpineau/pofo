# The library

Everything the `pofo` command does is a Go call you can make yourself. The
packages under `pkg/` use the standard library only, bundle decades of market
history, and document every exported name.

```sh
go get github.com/bpineau/pofo
```

This guide walks the library by task. Each page shows short snippets, and each
snippet is the body of a runnable `Example` in the package's tests, named on
its first line, so `go test` keeps it working. Snippets marked "synthetic"
read made-up series built by a helper in that test file; the others read the
bundled data or the network.

## Start here

| I want to... | Call | Guide |
|---|---|---|
| load a series | `marketdata.Bundled` (no network), `Client.Load`, `Client.FetchExtended`, `ReadCSV` | [Loading data](loading.md) |
| describe it | `Series.Stats`, `metrics.CalendarReturns`, `metrics.RollingCAGRs`, `analyze.Asset` | [Statistics](statistics.md) |
| blend, regress, study the worst months | `marketdata.NewPanel`, `Panel.Mix`, `metrics.Regress`, `metrics.LowestK` | [Panels and blends](panels.md) |
| compare series | `marketdata.AlignSeries`, `metrics.CorrelationMatrix`, `analyze.Pair` | [Comparing series](comparing.md) |
| study a portfolio | `analyze.Portfolio`; `portfolio.Parse`, `Build`, `Simulate` | [Portfolios](portfolios.md) |
| optimize weights | `optimize.ParseSpec`, `optimize.Solve` | [Optimization](optimization.md) |
| rebuild a missing past | `simgen.Find`, `simgen.Validate` | [Backcasts](backcasts.md) |
| draw charts or the report | `chart.Line`, `compare.Compute`, `report.Render` | [Rendering](rendering.md) |
| plan a withdrawal (FIRE) | `decumul.Plan`, `scenario` sources, `replay.Run` | [FIRE](../fire.md#from-go) |
| export | `marketdata.WriteCSV` | [Loading data](loading.md#files-in-and-out) |

## A first program

Two bundled series, a 60/40 of them rebalanced monthly, and the statistics of
all three, offline. It is the body of a `main` importing `fmt`, `log` and
`github.com/bpineau/pofo/pkg/marketdata`:

```go
// from pofo.Example
// The S&P 500 total return since 1871, bundled with the module.
sp500, err := marketdata.Bundled("SP500-USD")
if err != nil {
	log.Fatal(err)
}
bonds, err := marketdata.Bundled("TREASURY-INT-USD")
if err != nil {
	log.Fatal(err)
}

// Monthly returns on the months both share, and a 60/40 rebalanced
// every month (weights are fractions).
p, err := marketdata.NewPanel(marketdata.Monthly, sp500, bonds)
if err != nil {
	log.Fatal(err)
}
p, err = p.Mix("60/40", map[string]float64{"SP500-USD": 0.6, "TREASURY-INT-USD": 0.4})
if err != nil {
	log.Fatal(err)
}
for _, id := range p.IDs {
	s, err := p.Series(id)
	if err != nil {
		log.Fatal(err)
	}
	st, err := s.Stats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%-16s CAGR %5.2f %%/yr, volatility %4.1f %%/yr, max drawdown %5.1f %%\n",
		id, st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
}
```

## Scripts to start from

[`examples/code/`](../../../examples/code/README.md) holds nineteen
single-file programs, one question each: describe a series, compare several,
blend, regress, the worst months, the crises, rolling outcomes, calendar
years, the currency effect, a fee, a reconstruction against its reference, a
portfolio's correlations, simulation and optimized weights, a FIRE run, the
withdrawal rules through history, a CSV export.

```sh
go run examples/code/describe.go SP500-USD
mkdir -p scratch && cp examples/code/blend.go scratch/mine.go   # scratch/ is gitignored
```

## Units

Units are the library's first trap. This table matches the root package
documentation (`go doc github.com/bpineau/pofo`), and a test keeps the two in
step.

| Quantity | Unit | Where |
|---|---|---|
| weights | FRACTION (0.6) in memory; PERCENT (60) in files | fractions in `portfolio.Line`, `Holding.Weight`, `Asset.Weight`, `Panel.Mix`, `optimize`, `analyze`; percent in portfolio files and `Holding.RawWeight` |
| fees (TER) | PERCENT per year (0.20), except FRACTION per year (0.0020) in two places | percent in `portfolio` (`Line.Fees`, `Holding.Fees`, `EnvelopeFees`, `BorrowSpread`), `Client.Fees`, `datasets.Asset.Fees`; fraction in `simgen` and `Series.LessFee` |
| returns | FRACTION (0.04 = +4 %) | `metrics`, `analyze`, `scenario`, `decumul`; the last two in REAL terms |
| statistics | FRACTION, with two exceptions | `metrics.Stats` and the `analyze` studies; `Stats.Ulcer` in percent points, `Stats.CWARP` in percent |
| per period or annualized | a return, a regression, a covariance, a VaR are PER PERIOD; CAGR, volatility, Sharpe, Sortino, tracking error are ANNUALIZED | `Regression.AnnualAlpha` converts |
| cadence | periods per year: 252 daily, 52 weekly, 12 monthly | `metrics.PeriodsPerYear`, `Stats.PeriodsPerYear`, `Panel.PeriodsPerYear`, the `periodsPerYear` argument of `optimize.Solve` and of the bare-return functions |
| dates | a session date at 00:00 UTC; a monthly series is labelled by its month-END | every `Point.Date`; series match by exact `time.Time` equality |
| closes | ADJUSTED (dividends reinvested) by default | `FetchOptions.Raw` gives unadjusted closes with `Series.Dividends` (never both); a distributing fund's NAV is a PRICE return |
| SIM suffix | `IWDA` = real quotes only; `IWDASIM` = the bundled backcast spliced in front | `Client.FetchExtended`, `Series.SimulatedBefore`; `#meta sim:on` for a whole file |
| rates | annualized PERCENT LEVELS, never a return | `^IRX`, `^ESTR`, `^SOFR`, `portfolio.Portfolio.Cash` |
| #meta directives | PERCENT as written | `max-vol:9`, `view:ID:8@70`; FRACTION once parsed into `optimize.Spec` |

Volatility and ratios annualize at each series' own cadence, with a zero
risk-free rate; CAGR counts 365.25-day years. The conventions section of
`go doc github.com/bpineau/pofo/pkg/metrics` says why figures differ from
other tools'.

## The packages

| Package | What it does |
|---|---|
| `pkg/marketdata` | fetch, cache, resolve and convert prices; bundled series; panels of returns; CSV in and out |
| `pkg/metrics` | statistics on plain slices: CAGR, volatility, Sharpe, drawdowns, IRR, correlation, regression, VaR, tracking, attribution |
| `pkg/analyze` | one call for everything about an asset, a portfolio, or a pair of series; numbers only |
| `pkg/portfolio` | the file format, specs built in code, the rebalanced and fee-aware simulation |
| `pkg/optimize` | long-only weights for ten objectives, under bounds and limits |
| `pkg/suggest` | regime coverage, look-through composition, redundancy, gap-filling suggestions |
| `pkg/simgen` | reconstruction of the missing past of young funds, and its audit |
| `pkg/scenario`, `pkg/decumul`, `pkg/replay` | return paths, the withdrawal engine, history as it happened |
| `pkg/chart`, `pkg/report`, `pkg/compare` | SVG and terminal charts, the HTML report, the command's comparison pipeline |
| `pkg/datasets` | the embedded data and its catalog |
| `pkg/firebook`, `pkg/bookmd`, `pkg/epub`, `pkg/opds`, `pkg/seo` | the book, and generic Markdown, EPUB, OPDS and site-file writers |

Every package's documentation opens on the calls to start with:
`go doc github.com/bpineau/pofo/pkg/metrics`, or
[pkg.go.dev](https://pkg.go.dev/github.com/bpineau/pofo). The root package
documentation adds the layering (which package may import which).
`pkg/simgen` and `pkg/marketdata` end theirs with a "Generator plumbing"
section: exports that maintain the bundled data and are not meant for
consumers.
