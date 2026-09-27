# Example programs

Small runnable programs over the pofo library, one question each. They read
only the data the module bundles (`marketdata.Bundled`, `replay.Reference`),
so every one runs offline, from the module root:

```sh
go run ./examples/lib/describe SP500-USD
```

| Program | The question | What it shows |
|---|---|---|
| [describe](describe/main.go) | What is this series? | `marketdata.Bundled`, `Series.Stats`, `metrics.CalendarReturns`, `metrics.DrawdownEpisodes` |
| [blend](blend/main.go) | What would a 60/40 have done? | `marketdata.NewPanel` (monthly), `Panel.Between`, `Panel.Mix`, `Panel.Series`, `metrics.Corr` |
| [regress](regress/main.go) | How does gold move with equities and bonds? | a daily and two monthly series on one calendar, `metrics.Regress`, t-statistics, annualized alpha |
| [worstmonths](worstmonths/main.go) | What did the hedges do in the worst equity months? | `metrics.LowestK` (dated extremes), `Panel.Pick`, conditional means |
| [pair](pair/main.go) | Does a reconstruction match its reference? | `analyze.Pair`, `PairStudy.WriteText`, the study as JSON |
| [oldnew](oldnew/main.go) | Did a data refresh move the past? | `git show` piped into `marketdata.ReadCSV`, then `analyze.Pair` |
| [simulate](simulate/main.go) | How does this portfolio file behave? | `portfolio.Parse`, `Build`, `Simulate`, `metrics.Compute`, `metrics.Attribute` |
| [fire](fire/main.go) | Can this capital fund this spending? | `replay.Reference`, `scenario.StationaryBootstrap`, `decumul.Plan.Simulate`, `Plan.Solve`, historical cohorts |
| [export](export/main.go) | Get the data out to another tool | `Series.Resample`, `marketdata.WriteCSV` (long `id,date,value`) |

`marketdata.BundledIDs()` lists every identifier they accept. For live
quotes, swap `marketdata.Bundled(id)` for a client:

```go
client := marketdata.NewClient(marketdata.DefaultCacheDir())
s, err := client.FetchExtended(ctx, "IWDASIM", marketdata.FetchOptions{Currency: "EUR"})
```

and set `client.Offline = true` to read what a previous run cached without
touching the network.

To explore, copy one into the gitignored `scratch/` directory at the module
root (`scratch/<name>/main.go`, run with `go run ./scratch/<name>`) and
edit it there. The entry point of the library documentation is the root
package (`go doc github.com/bpineau/pofo`), and each package's `go doc` page
opens on the calls to start with.

Their output depends on the bundled data, so it is not pinned: `go test
./examples/lib/` builds them all and runs each one offline, checking only
that it exits cleanly and prints something.
