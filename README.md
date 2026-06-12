# portfodor

Go tool to visualize and compare investment portfolios over time — plus
reusable libraries to fetch price histories, compute risk/return metrics
and produce SVG charts.

The CLI reads allocation files, downloads price histories (Yahoo Finance,
Financial Times, Morningstar, Stooq), rebuilds the missing past (proxies and
simulated data), simulates each portfolio with periodic rebalancing and
generates a self-contained HTML report opened in the browser (per-portfolio
sections collapsed, comparison and statistics front and center).

## Usage

```sh
go build ./cmd/portfodor                       # self-contained binary (datasets embedded)
./portfodor my-portfolio.txt other.txt         # HTML report in /tmp + open
./portfodor -assets WPEA,NTSG,CSPX             # compare individual assets (100% each)
./portfodor -cli -assets VOO,IWDA              # quick check in the terminal
./portfodor -warmup                            # pre-warm the catalog cache
./portfodor -gen-simdata                       # regenerate datasets/simdata (then rebuild)
```

The binary can be installed anywhere: simulated histories and reference
series are **embedded at build time** (`go:embed` of `datasets/`), and the
quote cache lives in the standard user cache directory
(`~/Library/Caches/portfodor` on macOS, `~/.cache/portfodor` on Linux).

The `-assets` option treats each identifier as a portfolio invested 100% in
it — handy for comparing ETFs against each other without writing a file. It
can be combined with portfolio files.

## Portfolio file format

One line per asset: `<weight in %> <identifier> [free text]`. Everything
after a `#` is a comment; blank lines are ignored. The portfolio name is the
file name without its extension.

```
# Description, links, notes…
#meta rebalance:30   # directive: this portfolio rebalances every 30 days
#meta extra-fees:0.5 # wrapper/mandate fees, applied on top of the whole portfolio
60   VTI            US equities
25,5 IE00B4L5Y983   # ISIN resolved automatically (decimal comma accepted)
14.5 GOLD           # built-in alias → gold XAU/USD
```

`#meta key:value` lines carry per-portfolio directives: `rebalance:N` (days
between rebalances, `0` = never) and `extra-fees:X` (synonym
`envelope-fees:X`) — **additional fees in %/yr applied to the whole
portfolio**, on top of the assets' individual TERs: life-insurance/pension
wrappers, managed mandates, broker fees… Since they are not baked into the
quotes (unlike TERs), they are **deducted** from the simulated performance.

`#meta leverage:on` enables leveraged portfolios: weights are kept as
written (sum up to 500%) and the residual `100−sum` becomes a cash
position — earning the short rate (^IRX) if positive, **financed at the
rate + spread** if negative (`#meta borrow-spread:X`, default 1%/yr). A
NAV that reaches zero is a ruin: the series stops and the report flags it.
Without this directive, a weight > 100% is rejected (with a hint) and sums
≠ 100% are normalized as before.
An optional third numeric column declares an asset's TER
(e.g. `60 VOO 0.03`); otherwise it is fetched automatically (FT, justETF)
and cached for 6 months; `-no-fees` disables that lookup. The report shows
per-asset fees and a "Weighted ongoing fees" row in the statistics table.

The identifier can be a US ticker (`VTI`), a European ticker from the
embedded list (`IWDA`, `CSPX`, `CW8`…), an ISIN, or a built-in alias
(`GOLD`, `WTI`, `BHMG`, `AMUNDI-VOLATILITY`, `WINTON-TREND-EQUITY`…). If the
weights do not sum to 100, they are normalized with a warning.

**SIM convention**: a bare identifier (`DBMF`, `NTSG`, `VOO`) uses only the
asset's real quotes — the history starts at its inception date. The `SIM`
suffix (`DBMFSIM`, `NTSGSIM`, `VOOSIM`…) additionally allows extending the
uncovered period, via `datasets/simdata/` then the known proxies; real
quotes always keep priority wherever they exist. `-no-simulate` ignores SIM
suffixes globally.

## Main options

| Option | Default | Description |
|---|---|---|
| `-out` | `/tmp/portfodor-<timestamp>.html` | generated HTML file |
| `-data` | standard user cache | quote cache (JSON) |
| `-simdata` | embedded in the binary | source of simulated histories (directory for dev) |
| `-rebalance` | `90` | rebalance every N calendar days (0 = never) |
| `-start` | `2006-01-01` | desired start date |
| `-benchmark` | `^GSPC` | reference for Beta |
| `-cache-age` | `720h` (1 month) | cache freshness before re-downloading |
| `-assets` | | list `A,B,C`: each asset compared as a 100% portfolio |
| `-cli` | | curves and summary table in the terminal, no HTML |
| `-width` | `$COLUMNS` or 100 | width of the `-cli` chart (wider = more granularity) |
| `-warmup` | | pre-warm the built-in asset catalog then exit |
| `-no-open`, `-no-simulate` | | do not open the browser / ignore SIM suffixes |

## Data

- **Resolution**: aliases → embedded ticker→ISIN list (European ETFs/funds)
  → built-in catalog of pinned resolutions → multi-source search
  (Yahoo, FT, Morningstar via Boursorama), the deepest series winning.
- **Cache**: 1 month by default; a failed refresh **serves the stale data**
  with a stderr warning (charts may stop before today), and never deletes
  anything.
- **History extension** (`…SIM` identifiers only): first the
  `datasets/simdata/` files (below), otherwise a known proxy (VOO→^GSPC,
  BND→VBMFX, …), rescaled to the first real quote. The report flags every
  simulated portion.

## Simulated data (datasets/simdata/)

Complex assets (90/60 funds, managed futures…) are rebuilt by `cmd/simgen`
from long-history building blocks, validated against their real quotes,
then stored as self-documenting CSVs (method, validation, date) in
`datasets/simdata/`:

```sh
./portfodor -gen-simdata                   # regenerate everything (then make build to re-embed)
./portfodor -gen-simdata -dry NTSX         # validate without writing
```

Bundled recipes and measured quality (daily/weekly correlation of returns
vs the real series; the real series is always grafted on top of the
simulation wherever it exists):

| Asset | Method | Validation |
|---|---|---|
| NTSX (UCITS) | 0.90×VFINX + 0.60×(VFITX−cash) + 0.10×cash (1991→) | corr 0.96 / weekly 0.99 vs NTSX US |
| NTSG (UCITS) | global 60/40 US/intl variant | weekly 0.86 (thinly traded LSE listing) |
| URTH, IWDA | 0.60×VFINX + 0.40×VTMGX (1999→) | corr 0.90 / weekly 0.97 |
| ZROZ, IEF, TLT | imported refs derived from US yield curves (1962→) | corr 1.00 over 16–24 years of overlap |
| XAUUSD (GOLD) | imported spot gold (1968→), real GC=F grafted | corr 1.00 |
| KMLM | official MLM Index (1987→) + 0.90% ETF fees | corr 1.00 |
| DBMF | official SG CTA Index (2000→) | corr 0.68 / weekly 0.75, beta 0.96 |
| CTA | official SG Trend Index (2000→) | corr 0.54 — proprietary strategy, gap accepted |
| Winton Trend-Equity | global equities + 0.5×Winton Trend fund (real 2019→, sim before) | weekly 0.92 |
| Amundi Volatility, BH Macro | regression backcast **rejected** (R² 0.20 / 0.00) | real history only (2007→) |

Discretionary strategies cannot be honestly replicated with factors: rather
than inventing data, the generator rejects them below an R² floor.

## Reference data (datasets/refdata/)

`datasets/refdata/` holds reference series imported once and for all
(provenance and method at the top of each file): official SG Trend/SG CTA
indices, MLM Index history, 7-10/20+/25+ treasuries derived from US yield
curves since 1962, spot gold since 1968, Winton Trend fund.
`cmd/simgen` consumes them first (`-refdata`), before any network source.

## Using it as a library

The repository is also a toolkit for writing other portfolio-processing
applications. Layout:

```
pkg/marketdata/   data: resolution (aliases, ISIN, catalog), multi-provider
                  sources, cache, fees, simdata, alignment
pkg/metrics/      statistics (CAGR, Sharpe, Sortino, drawdowns, Beta…)
pkg/chart/        SVG charts (Line) and terminal (Term), shared palette
pkg/portfolio/    allocation file format + rebalanced simulation
pkg/report/       HTML and text rendering of the comparison model
pkg/simgen/       history reconstruction (composites, TSMOM, backcasts)
cmd/              the portfodor binary (report, warmup, gen-simdata)
datasets/         versioned data (embedded at build time) and its QA:
  simdata/          permanent simulated histories (spliced at runtime)
  refdata/          imported reference series (official indices…)
  golden/           golden tests + frozen fixtures vs external references
data/             old local cache (replaced by the user cache)
```

Everything consumable as a library lives under `pkg/`; `cmd/` only contains
the CLI wiring and `golden/` the golden test suite.

Each package has its documentation page — calculation conventions included
(`go doc github.com/bpineau/portfodor/pkg/metrics`) — and runnable examples:

```go
import (
	"github.com/bpineau/portfodor/pkg/chart"
	"github.com/bpineau/portfodor/pkg/marketdata"
	"github.com/bpineau/portfodor/pkg/metrics"
	"github.com/bpineau/portfodor/pkg/portfolio"
)

// Fetch a price history (transparent resolution + caching).
client := marketdata.NewClient("data")
series, err := client.Fetch("IWDA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

// Compute CAGR, Sharpe, Sortino, Ulcer, MaxDD, TTR, Beta…
stats, err := metrics.Compute(dates, values)

// Render a standalone SVG.
svg := chart.Line(chart.Options{Title: "Comparison"}, []chart.Series{{Name: "P1", Dates: dates, Values: values}})

// Parse and simulate a portfolio (N-day rebalancing).
spec, _ := portfolio.ParseFile("p.txt")
sim, _ := portfolio.Simulate(p, 90)
```

- `marketdata` — resolution (aliases, ISIN, catalog), multi-source
  downloads, cache, simdata, proxies.
- `metrics` — statistics over value series (returns, drawdowns, Beta).
- `chart` — pure-stdlib inline SVG charts.
- `portfolio` — allocation file parsing and rebalanced simulation.
- `report` — HTML report rendering.
- `simgen` — reconstruction engine (linear composites, imported
  references, TSMOM trend-following engine, regression backcasts) and
  validated recipes.

## Known limitations

- No currency conversion: mixing currencies triggers a warning, returns
  stay in each asset's own currency.
- Price-index proxies (^GSPC, ^NDX…) omit dividends over the simulated
  portion; managed-futures replications (corr ≈ 0.3–0.5) reflect those
  strategies' regime, not their daily positions.

## Golden tests

`datasets/golden/` replays the simulation on frozen real data (SPY
2006-2025, URTH 2012-2025) and compares CAGR, volatility, Sharpe, Sortino,
Ulcer, Max Drawdown and TTR against validated external references (official
S&P 500 TR annual returns, canonical GFC/COVID drawdowns,
LazyPortfolioETF).
Any calculation drift beyond the tolerances fails `go test ./datasets/golden`.

## Development

```sh
go test ./...   # unit tests + examples, no network
go vet ./...
```

No external dependencies: standard library only.
