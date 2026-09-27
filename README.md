# pofo

pofo is a Go toolkit for deciding and checking how a long-horizon portfolio
is built, and a command built on it. It uses the standard library only, bundles
decades of market history, and runs offline for most questions.

It exists to answer four questions:

- What would this mix of funds have done over the longest window the evidence
  supports, and what did each line contribute to its risk and return?
- Is this allocation redundant, or blind to a macro regime, and what would
  fill the gap?
- Can this capital fund this retirement spending, for how long, and with what
  chance of running out?
- What did a fund that is only ten years old plausibly do in 1975?

A public instance runs at [pofo.zouh.org](https://pofo.zouh.org): no ads, no
accounts, no tracking cookies.

## Four faces

| Face | What you get | Start with |
|---|---|---|
| **A Go library** | market data with bundled history back to 1871, statistics, returns panels and regression, portfolio simulation, optimization, backcasts of young funds; every capability of the command is a call | [library guide](docs/usage/library/README.md) |
| **A FIRE engine** | ruin probabilities, spending rules, stochastic lifetimes, history replayed as it happened | [FIRE guide](docs/usage/fire.md), [the simulator](https://pofo.zouh.org/firesimulator/) |
| **A portfolio visualizer** | HTML and terminal reports comparing portfolios, their look-through composition, regime coverage and risk budget; a web app with shareable URLs | [command line](docs/usage/cli.md), [web app](docs/usage/web.md), [the visualizer](https://pofo.zouh.org/visualizer) |
| **A book** | a handbook of living off one's capital, in French and English, also as EPUB | [Le FIRE tranquille](https://pofo.zouh.org/firebook/fr/), [The Quiet FIRE](https://pofo.zouh.org/firebook/en/) |

The library is a peer of the command, not a by-product: the command, the web
app and the book are applications of it.

## Quick start

The command:

```sh
go install github.com/bpineau/pofo/cmd/pofo@latest
pofo -assets VOO,IWDA                                   # compare two funds: HTML report, opened
pofo -offline -cli -currency USD -assets SP500,MSCIWORLD  # bundled indices, in the terminal, no network
pofo -dump list                                         # every series bundled in the binary
pofo -serve                                             # the web app on http://127.0.0.1:8787/
```

The library, in a `main` importing `fmt`, `log` and
`github.com/bpineau/pofo/pkg/marketdata`:

```go
// from pofo.Example_quickStart
// The S&P 500 total return since 1871, bundled: no network, no API key.
sp500, err := marketdata.Bundled("SP500-USD")
if err != nil {
	log.Fatal(err)
}
st, err := sp500.Stats()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("%d-%d: CAGR %.1f %%/yr, volatility %.1f %%/yr, max drawdown %.1f %%\n",
	st.Start.Year(), st.End.Year(), st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
```

```text
1871-2026: CAGR 9.4 %/yr, volatility 16.2 %/yr, max drawdown -83.1 %
```

[Getting started](docs/usage/getting-started.md) goes one step further on
both sides.

## Two conventions to know

- **Units.** In Go, weights and returns are fractions (`0.6`, `0.04`); in a
  portfolio file, weights are percentages (`60`). Fees are percent per year,
  except in `pkg/simgen` and `Series.LessFee`. The
  [units table](docs/usage/library/README.md#units) lists every case.
- **SIM.** `IWDA` means a fund's real quotes only; `IWDASIM` puts its
  bundled reconstruction in front of them, for a longer history. The
  [data guide](docs/usage/data.md#backcasts) says how those are built and
  graded.

## Where to go next

| I want to... | Read |
|---|---|
| install and run a first report or program | [Getting started](docs/usage/getting-started.md) |
| find the right command-line mode | [The command line](docs/usage/cli.md) |
| write a portfolio file | [Portfolio files](docs/usage/portfolio-files.md) |
| run or deploy the web app | [The web app](docs/usage/web.md) |
| size a retirement | [FIRE and decumulation](docs/usage/fire.md) |
| know where the data comes from and how far to trust it | [Data and backcasts](docs/usage/data.md) |
| use the library, by task | [The library](docs/usage/library/README.md) |
| start from a runnable script | [`examples/code/`](examples/code/README.md) |
| start from a model portfolio | [`examples/portfolios/`](examples/portfolios/README.md) |
| read the API reference | `go doc github.com/bpineau/pofo`, [pkg.go.dev](https://pkg.go.dev/github.com/bpineau/pofo) |
| every guide at a glance | [`docs/usage/`](docs/usage/README.md) |

## Contributing

```sh
make check      # format, vet, staticcheck, tests, example scripts (no network)
make golden     # computations against frozen external references
make help       # every other target
```

No third-party dependency, ever: that is what lets other programs embed pofo
without inheriting a supply chain. [`AGENTS.md`](AGENTS.md) is the map for
contributors and coding agents (layout, conventions, traps, house rules), and
[`docs/specs/`](docs/specs/) holds the records behind design decisions and
data validations.
