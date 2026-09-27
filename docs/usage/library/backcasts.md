# Backcasts

A backcast rebuilds the missing past of a young fund from older series: its
own earlier share classes, the index it tracks, long-lived mutual funds,
bundled references. The [data guide](../data.md#backcasts) explains the
principles; this page shows the Go side.

## Reading a shipped backcast

You rarely need `pkg/simgen` to use a backcast: every shipped one is a `SIM`
identifier.

| Call | Returns |
|---|---|
| `marketdata.Bundled("IWDASIM")` | the reconstruction alone, offline |
| `Client.FetchExtended(ctx, "IWDASIM", opt)` | the real quotes with the reconstruction in front; `Series.SimulatedBefore` marks the join |
| `Client.Load(ctx, "IWDASIM", opt)` | the same, falling back on the bundle alone when offline and uncached |

The bare `IWDA` always means the real quotes only.

## Rebuilding one

`simgen.Find` returns the recipe `pofo -gen-simdata` ships for an asset, and
`Recipe.Build` runs it on any fetcher. Here it runs on bundled reference data
only, offline:

```go
// from simgen.ExampleFind (offline: bundled reference data only)
r, ok := simgen.Find("ZROZ")
if !ok {
	panic("no recipe")
}
s, err := r.Build(simgen.WithRefData(datasets.Refdata(), offline{}), time.Time{})
if err != nil {
	panic(err)
}
fmt.Println(r.Name)
fmt.Printf("rebuilt from %s, graded against %s\n", s.First().Date.Format(time.DateOnly), r.ValidateAgainst)
```

```text
PIMCO 25+Y zero-coupon: 27y Treasury STRIP
rebuilt from 1953-04-30, graded against ZROZ
```

## Grading one

`simgen.Validate` measures a reconstruction on its overlap with the real
quotes: correlation, beta, the CAGR of each.

```go
// from simgen.ExampleValidate (synthetic)
var dates []time.Time
var real, sim []float64
r, s := 100.0, 100.0
for i := range 750 {
	move := 0.01 * math.Sin(float64(i)*0.7)
	r *= 1 + 0.0003 + move
	s *= 1 + 0.0002 + 0.9*move + 0.002*math.Cos(float64(i)*1.3)
	dates = append(dates, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
	real, sim = append(real, r), append(sim, s)
}
realS, _ := marketdata.NewSeries("FUND", dates, real)
simS, _ := marketdata.NewSeries("FUND (rebuilt)", dates, sim)

v, err := simgen.Validate(simS, realS)
if err != nil {
	panic(err)
}
fmt.Printf("%d common returns, correlation %.2f, beta %.2f\n", v.Overlap, v.Corr, v.Beta)
fmt.Printf("CAGR %.1f %% rebuilt vs %.1f %% real\n", v.CAGRSim*100, v.CAGRReal*100)
```

```text
749 common returns, correlation 0.98, beta 0.90
CAGR 8.1 % rebuilt vs 12.1 % real
```

A high correlation with a large CAGR gap is the classic failure: the shape is
right and the level is wrong, and every backtest built on it inherits the
gap. That is why the full audit, `pofo -verify-simdata`, grades **level** and
**path** separately.

## Building blocks

| Engine | Use |
|---|---|
| `simgen.Composite` | constant-weight blends of legs, some as excess returns over cash (stacked funds) |
| `simgen.CapWeighted` | blends whose weights drift with their legs, from a published split (world equity) |
| `simgen.TreasuryZeroTR` | a zero-coupon Treasury priced off a yield series |
| the TSMOM engine (`simgen.Example_tsmom`) | a time-series momentum (trend) strategy |

Each has a runnable example in `go doc github.com/bpineau/pofo/pkg/simgen`.
The exports listed under "Generator plumbing" at the end of that page
maintain the bundled data and are not meant for consumers.
