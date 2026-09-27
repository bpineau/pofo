# Example scripts

Single-file Go programs over the pofo library, one question each. Every
file is a self-contained `package main` behind a `//go:build ignore` line,
so they share this directory and each one runs on its own, from the module
root:

```sh
go run examples/code/describe.go SP500-USD
go run examples/code/describe.go -h          # its flags
```

They read their data through `marketdata.Client.Load`: a CSV path is a
file, an identifier the module bundles (a catalog fund's backcast, an index,
a yield: `pofo -dump list` names them all) comes from the bundle, offline
and the same on every machine, and anything else is fetched and cached on
disk. `IWDA` therefore reads the fund's bundled backcast, while `IWDASIM`
reads its live quotes with that backcast in front. `-offline` keeps every
script off the network: the bundle and the quote cache only. Log lines
(identifier resolution, currency warnings) go to standard error, the answer
to standard output.

| Script | The question | Example |
|---|---|---|
| [describe](describe.go) | What is this series: statistics, calendar years, deepest drawdowns? | `go run examples/code/describe.go SP500-USD` |
| [stats](stats.go) | How do these series compare, side by side? | `go run examples/code/stats.go -monthly SP500-USD MSCIWORLD-USD TREASURY-LONG-USD XAUUSD-LBMA` |
| [blend](blend.go) | What would this mix, rebalanced monthly, have done? | `go run examples/code/blend.go SP500-USD=60 TREASURY-INT-USD=40` |
| [regress](regress.go) | What is this series made of: alpha, betas, t, R2? | `go run examples/code/regress.go -from 1972-01-01 XAUUSD-LBMA SP500-USD TREASURY-LONG-USD` |
| [worstmonths](worstmonths.go) | Does this hedge work in the reference's worst months? | `go run examples/code/worstmonths.go SP500-USD TREASURY-LONG-USD XAUUSD-LBMA TREND-NET-USD` |
| [episodes](episodes.go) | What did each series do in the crises (1973-74, 1987, 2000-02, 2008, 2020, 2022, your own)? | `go run examples/code/episodes.go -dd SP500-USD TREASURY-LONG-USD XAUUSD-LBMA` |
| [rolling](rolling.go) | What did holding it N years deliver, at worst, and when? | `go run examples/code/rolling.go -years 15 SP500-USD MSCIWORLD-USD XAUUSD-LBMA` |
| [calendar](calendar.go) | What did each year look like, and does it match a published table? | `go run examples/code/calendar.go -from 2000-01-01 IWDA SP500-USD XAUUSD-LBMA` |
| [currency](currency.go) | What did the currency do to this investment? | `go run examples/code/currency.go -in EUR IWDA` |
| [fees](fees.go) | What does a fee cost; what does a share class really charge? | `go run examples/code/fees.go -fee 0.5 MSCIWORLD-USD` |
| [pair](pair.go) | Does this series match its reference? | `go run examples/code/pair.go IWDA MSCIWORLD-USD` |
| [oldnew](oldnew.go) | Did a data refresh move the past of a bundled file? | `go run examples/code/oldnew.go -rev HEAD~20 TREASURY-LONG-USD` |
| [scanbundle](scanbundle.go) | Is the bundled data sound: cadence, age, gaps, spikes? | `go run examples/code/scanbundle.go -stale 45` |
| [correl](correl.go) | Do these holdings diversify each other? | `go run examples/code/correl.go examples/portfolios/all-weather-dalio.txt` |
| [simulate](simulate.go) | How does this portfolio file behave, holding by holding? | `go run examples/code/simulate.go examples/portfolios/golden-butterfly.txt` |
| [optimize](optimize.go) | What weights would an objective pick for these holdings? | `go run examples/code/optimize.go -objective min-volatility,max-weight:40 examples/portfolios/golden-butterfly.txt` |
| [fire](fire.go) | Can this capital fund this spending, and what spending is safe? | `go run examples/code/fire.go -capital 1200000 -spend 42000 -years 40` |
| [replay](replay.go) | What life would each withdrawal rule have given from that year? | `go run examples/code/replay.go -start 1966 -years 35` |
| [export](export.go) | How do I get this data into another tool? | `go run examples/code/export.go -monthly -o /tmp/series.csv IWDA XAUUSD-LBMA` |

Their output moves with the bundled data, so none is pinned; `make
examples` (part of `make check`) builds, vets and lints every file, so a
library change that breaks one fails the gate.

## Write your own

Copy the nearest script into the gitignored `scratch/` directory at the
module root, inside the module so pofo's packages import with no `go.mod`
or `replace` of your own, and edit it there:

```sh
mkdir -p scratch && cp examples/code/blend.go scratch/volcheck.go
go run scratch/volcheck.go SP500-USD=60 TREASURY-INT-USD=40
```

`go doc github.com/bpineau/pofo` is the library's entry point (which package
answers which question, a complete program, the units table), and every
package's `go doc` page opens on the calls to start with. What turns out to
be worth keeping becomes a script here, a `pofo` mode, or library code with
tests.
