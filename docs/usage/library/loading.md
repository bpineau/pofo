# Loading data

Everything starts with a `*marketdata.Series`: dated closes, a currency, and
where it came from. There are four ways to get one.

| Call | Reads | Network |
|---|---|---|
| `marketdata.Bundled(id)` | a series embedded in the module | never |
| `Client.FetchExtended(ctx, id, opt)` | real quotes, cached on disk; the backcast in front for a `SIM` id | yes, unless `Client.Offline` |
| `marketdata.ReadCSV(r, name)` | a `date,value` file of your own | never |
| `Client.Load(ctx, id, opt)` | the first of the three that holds `id`: a path is a file, then the bundle, then the client | as the client allows |

`Client.Load` is the one call an exploration program needs.

## One call for any identifier

```go
// from marketdata.ExampleClient_Load
ctx := context.Background()
client := marketdata.NewClient("") // no disk cache: the bundle alone answers
client.Offline = true

// A reference series and a catalog index, both bundled; the index
// converted into euros through the bundled euro crosses.
for _, id := range []string{"SP500-USD", "MSCIWORLD"} {
	s, err := client.Load(ctx, id, marketdata.FetchOptions{
		From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(s.Symbol, s.Source, s.First().Date.Format(time.DateOnly))
}
eur, err := client.Load(ctx, "MSCIWORLD", marketdata.FetchOptions{Currency: "EUR"})
if err != nil {
	panic(err)
}
fmt.Println(eur.Currency)

// A quoted fund: "DBMFSIM" is its real quotes with the reconstruction in
// front, and offline, with nothing cached, the bundled SIM history alone.
// The bare "DBMF" would be its real quotes only, which offline and
// uncached do not exist.
sim, err := client.Load(ctx, "DBMFSIM", marketdata.FetchOptions{}) // DBMF: US25159K3095
if err != nil {
	panic(err)
}
fmt.Println(sim.First().Date.Year() < 2019)
_, err = client.Load(ctx, "DBMF", marketdata.FetchOptions{})
fmt.Println(errors.Is(err, marketdata.ErrOffline))
```

```text
SP500-USD refdata 2020-01-31
MSCIWORLD simdata 2020-01-02
EUR
true
true
```

`FetchOptions` shapes the result: `From` and `To` trim it, `Currency`
converts it. `pofo -dump list` (or `marketdata.BundledIDs`) names everything
the bundle holds.

## Real quotes and their raw data

`Client.Fetch` resolves a ticker, ISIN or alias and returns adjusted daily
closes. `FetchExtended` is the command's per-asset pipeline: the `SIM`
backcast spliced in front, then the currency conversion. `Raw` asks for
unadjusted closes with the distributions beside them as cash.

```go
// from marketdata.Example_priceHistory (compiled, not run: it needs the network)
ctx := context.Background()
client := marketdata.NewClient(marketdata.DefaultCacheDir())

// Real quotes, adjusted, native currency; the slices pkg/metrics takes.
iwda, err := client.Fetch(ctx, "IWDA", time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)) // IE00B4L5Y983
if err != nil {
	panic(err)
}
dates, closes, returns := iwda.Dates(), iwda.Values(), iwda.Returns()
monthEnds, err := iwda.Resample(marketdata.Monthly)
if err != nil {
	panic(err)
}
fmt.Println(len(dates), len(closes), len(returns), monthEnds.Len())

// The CLI's pipeline: the bundled backcast in front (SIM), in euros.
long, err := client.FetchExtended(ctx, "IWDASIM", marketdata.FetchOptions{Currency: "EUR"})
if err != nil {
	panic(err)
}
fmt.Println("simulated before", long.SimulatedBefore.Format(time.DateOnly))

// Unadjusted closes, the distributions beside them as cash.
vt, err := client.FetchExtended(ctx, "VT", marketdata.FetchOptions{Raw: true}) // US9220427424
if err != nil {
	panic(err)
}
fmt.Println(len(vt.Dividends), "distributions in", vt.Currency)

// Live: the freshest price, today's 5-minute path (ErrNotCovered off Yahoo).
if q, err := client.Latest(ctx, "IWDA"); err == nil {
	fmt.Println(q.Price, q.Currency, q.Live)
}
if today, err := client.Intraday(ctx, "IWDA"); err == nil {
	fmt.Println(len(today.Points), "ticks today")
}
```

Two rules keep the numbers right:

- Never pair `Series.Dividends` with adjusted closes: the income would count
  twice.
- A distributing share class publishes a price return, which misses its
  income. `marketdata.LooksDistributing` flags one.

`NewSeries` wraps data you already hold, and `Client.ConvertCurrency` reprices
any series into another currency.

## Files in and out

`ReadCSV` reads a `date,value` file, the layout of the bundled ones;
`ReadLongCSV` reads several series from one `id,date,value` file; `WriteCSV`
writes that long layout, with values that parse back exactly. An offline
client serves the quote cache whatever its age; its files are private, so
always read them through a client.

```go
// from marketdata.Example_loading (compiled, not run: it reads the local cache and a git checkout)
ctx := context.Background()

// Everything the binary bundles, no Client needed.
fmt.Println(len(marketdata.BundledIDs()), "bundled series")
tsy, err := marketdata.Bundled("TREASURY-LONG-USD")
if err != nil {
	panic(err)
}

// Any "date,value" file: the same series ten commits ago.
out, err := exec.Command("git", "show", "HEAD~10:pkg/datasets/refdata/TREASURY-LONG-USD.csv").Output()
if err != nil {
	panic(err)
}
old, err := marketdata.ReadCSV(bytes.NewReader(out), "TREASURY-LONG-USD@HEAD~10")
if err != nil {
	panic(err)
}

// What a previous run cached, whatever its age; never the network.
client := marketdata.NewClient(marketdata.DefaultCacheDir())
client.Offline = true
iwda, err := client.FetchExtended(ctx, "IWDASIM", marketdata.FetchOptions{Currency: "EUR"}) // IE00B4L5Y983
if errors.Is(err, marketdata.ErrOffline) {
	panic("never fetched: run it once online")
}
if err != nil {
	panic(err)
}

// Out to any tool, month-end closes, one long id,date,value file.
monthly, err := iwda.Resample(marketdata.Monthly)
if err != nil {
	panic(err)
}
if err := marketdata.WriteCSV(os.Stdout, tsy, old, monthly); err != nil {
	panic(err)
}
```

From the command line, `pofo -dump` writes the same long CSV.

## Live prices

`Client.Latest` returns the freshest price as a `Quote`, for a live
valuation. It degrades rather than fails: a second Yahoo host, then the daily
close with its fallbacks, then the stale cache, so it answers even offline.

```go
// from marketdata.ExampleClient_Latest (compiled, not run: it needs the network)
client := marketdata.NewClient(marketdata.DefaultCacheDir())
ctx := context.Background()

q, err := client.Latest(ctx, "VWCE")
if err != nil {
	panic(err)
}
rate, err := client.FXRate(ctx, q.Currency, "EUR", q.Time)
if err != nil {
	panic(err)
}
freshness := "close of"
if q.Live {
	freshness = "live at"
}
shares := 12.0
fmt.Printf("%.2f EUR (%s %s)\n", shares*q.Price*rate, freshness, q.Time.Format("2006-01-02 15:04"))
```

| Need | Call |
|---|---|
| today's 5-minute path | `Client.Intraday`; `ErrNotCovered` when Yahoo does not quote the asset; no caching, so throttle yourself |
| pre-market and after-hours prints of US venues | `Client.LatestBatchExtended`, or `QuoteOptions.ExtendedHours`; `Quote.Session` reads `pre` or `post` |
| many quotes at once | `Client.LatestBatch` |

Off-hours prints are thin: show them, but do not book a valuation on them.

Some funds publish their price once a day, with a lag (a French
employee-savings fund, for instance). When the catalog names a
`nowcast_proxy` for one, its intraday path and latest quote are estimated
from the proxy's moves, and its daily series runs on past its last published
value. The estimate is flagged (`IntradaySeries.Estimate`,
`Series.EstimatedFrom`, `Quote.Source == "nowcast"`), never cached, and
`Series.WithoutEstimates` strips it.
