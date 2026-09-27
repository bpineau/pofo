# Using pofo

Guides for people using pofo: the command, the web app and the Go library.
Start with the first one; the others stand alone.

| Guide | Read it to... |
|---|---|
| [Getting started](getting-started.md) | install pofo, run a first report and a first Go program, know where the data comes from |
| [The command line](cli.md) | find the mode that answers your question: reports, sweeps, coverage, dumps, pairs, data checks |
| [Portfolio files](portfolio-files.md) | write a portfolio: weights, identifiers, `SIM`, `#meta` directives, fees, leverage, flows, optimizers |
| [The web app](web.md) | run `pofo -serve`: the visualizer, share URLs, the FIRE simulator, the book, a public deploy |
| [FIRE and decumulation](fire.md) | size a retirement: the simulator, the scripts, the engine in Go, how to read a ruin probability |
| [Data and backcasts](data.md) | know where every series comes from, how young funds get a longer past, and how far to trust it |

## The library, by task

| Guide | Read it to... |
|---|---|
| [Overview](library/README.md) | find the call for your question, a first program, the units table, the packages |
| [Loading data](library/loading.md) | read bundled, fetched, cached and file series; export them; live prices |
| [Statistics](library/statistics.md) | describe a series: statistics, cadence, one asset in one call, holding periods |
| [Panels and blends](library/panels.md) | blend, regress, and study the worst months across series |
| [Comparing series](library/comparing.md) | correlate series on one calendar; measure a series against its reference |
| [Portfolios](library/portfolios.md) | study and simulate a portfolio, with flows; find what it is missing |
| [Optimization](library/optimization.md) | compute weights under bounds, limits and views |
| [Backcasts](library/backcasts.md) | read, rebuild and grade a reconstructed history |
| [Rendering](library/rendering.md) | draw SVG charts and the HTML report |

## Elsewhere

- [`examples/code/`](../../examples/code/README.md): nineteen runnable
  scripts, one question each.
- [`examples/portfolios/`](../../examples/portfolios/README.md): about forty
  model portfolios.
- `go doc github.com/bpineau/pofo` and
  [pkg.go.dev](https://pkg.go.dev/github.com/bpineau/pofo): the API
  reference; every package opens on the calls to start with.
- [`docs/specs/`](../specs/): the records behind design decisions and data
  validations, for whoever changes pofo.
