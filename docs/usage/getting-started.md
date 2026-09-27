# Getting started

pofo is two things you can use separately: a command, `pofo`, and the Go
library it is built on. This page gets both running in a few minutes.

## Install

pofo needs Go and nothing else: no database, no API key, no third-party
module.

```sh
go install github.com/bpineau/pofo/cmd/pofo@latest
```

To follow the examples in these guides, clone the repository instead, since
the model portfolios and the example scripts live in it:

```sh
git clone https://github.com/bpineau/pofo && cd pofo
make build          # ./pofo, with every bundled dataset embedded
```

The binary is self-contained. Reference series, backcasts, the asset catalog
and the book are embedded at build time, so it can be copied anywhere. Quotes
it downloads are cached in the standard user cache directory
(`~/Library/Caches/pofo` on macOS, `~/.cache/pofo` on Linux).

## A first report

Compare two funds, each held at 100 %, and open the HTML report in your
browser:

```sh
./pofo -assets VOO,IWDA
```

The same in the terminal, without a browser. The two bundled indices below
need no download at all, so this works offline:

```sh
./pofo -offline -cli -currency USD -assets SP500,MSCIWORLD
```

```text
Portfolios: SP500, MSCIWORLD
Common period: 1969-12-31 → 2026-08-31

Metric                                        SP500          MSCIWORLD
──────────────────────────────────────────────────────────────────────
CAGR (annualized return)                   *11.14 %             9.38 %
Volatility (annualized)                     17.13 %           *14.44 %
Max Drawdown                              *-55.25 %           -57.81 %
Worst rolling 5y CAGR                       -8.20 %           *-6.81 %
...
```

A portfolio is a small text file, one holding per line. The repository ships
about forty model portfolios:

```sh
./pofo examples/portfolios/golden-butterfly.txt examples/portfolios/tradi-60-40.txt
```

Next: [the command line](cli.md) for every mode, and
[portfolio files](portfolio-files.md) for the format.

## A first Go program

The library is a peer of the command, not a by-product: everything the
command does is a call you can make. This program reads two bundled series,
builds a 60/40 of them rebalanced monthly and prints the statistics of all
three. It runs offline. It is the body of a `main` importing `fmt`, `log` and
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

```text
SP500-USD        CAGR 11.36 %/yr, volatility 14.6 %/yr, max drawdown -50.9 %
TREASURY-INT-USD CAGR  5.08 %/yr, volatility  5.0 %/yr, max drawdown -14.7 %
60/40            CAGR  9.11 %/yr, volatility  9.2 %/yr, max drawdown -30.3 %
```

The window starts in 1953, the first month both series share. The figures
move a little with every data refresh.

To go further without writing code from scratch, copy one of the
[example scripts](../../examples/code/README.md), each one file answering one
question. The [library guide](library/README.md) walks the packages by task.

## Where the data comes from

pofo reads two kinds of data.

| Kind | What it is | Network |
|---|---|---|
| Bundled | long reference series (the S&P 500 since 1871, Treasury yields and total returns, gold, MSCI indices, trend and cash references) and a reconstructed past for young funds | never |
| Quotes | real daily prices from public sources (Yahoo Finance, Financial Times, Morningstar, Stooq, the ECB), by ticker, ISIN or alias | once, then cached |

`./pofo -dump list` names every bundled series. Add `-offline` to any command
to forbid downloads: the cache and the bundle then answer, and anything else
is an error.

A young fund can borrow a longer past. `IWDA` means the fund's real quotes
only; `IWDASIM` puts its bundled reconstruction in front of them. The
[data guide](data.md) explains how those reconstructions are built and graded.

## Three things to know before reading any number

- **Returns are total returns.** Quotes are adjusted for dividends, so a CAGR
  includes the income, net of the fund's own fees.
- **Statistics are in the report currency.** The command converts everything
  to euros by default (`-currency USD` or `-currency ""` to change it); the
  library keeps each series in its own currency until you convert it.
- **Units differ between files and code.** A portfolio file writes weights in
  percent (`60`), the library in fractions (`0.6`). The
  [units table](library/README.md#units) lists every case.
