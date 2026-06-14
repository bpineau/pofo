# portfodor

Go tool to visualize and compare investment portfolios over time — plus
reusable libraries to fetch price histories, compute risk/return metrics
and produce SVG charts.

The CLI reads allocation files, downloads price histories (Yahoo Finance,
Financial Times, Morningstar, Stooq), rebuilds the missing past (proxies and
simulated data), simulates each portfolio with periodic rebalancing and
generates a self-contained HTML report opened in the browser (comparison and
statistics front and center; per-portfolio sections collapsed, each with a
performance curve, three composition pies — geography, sector, asset type —
and its macro-regime coverage).

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
`#meta capital:10000` sets a starting amount, and unlocks periodic external
flows: `#meta contribute:500/month` (fixed amount added every week, month,
quarter or year) and `#meta withdraw:4%/year` (fixed amount, or a
percentage of the current value). Flows are invested or sold pro rata on
the first trading day of each new period. Statistics and comparison charts
stay on a **time-weighted index** (flows don't distort returns), while the
money rows — starting capital, total contributed/withdrawn, final value and
a **money-weighted IRR** — follow the actual cash. Withdrawing a depleted
portfolio is a ruin: the series stops and the report flags it.

`#meta optimize:max-sharpe` (or `min-volatility`, `risk-parity`) lets the
optimizer compute the weights instead of using the ones you wrote. It works
out the long-only allocation that maximizes the Sharpe ratio, minimizes
volatility, or equalizes each asset's risk contribution, over the period
where all the assets quote. An optional cap diversifies the result:
`#meta optimize:max-sharpe,max-weight:40`. The report then shows **two
portfolios side by side** — `name (as written)` and `name (max-sharpe)` —
so the optimizer's choice is compared with your baseline; the computed
weights and their in-sample expected return/volatility/Sharpe appear as a
note under the optimized portfolio. Those figures are fitted on the past,
so treat them as a starting point, not a promise. `max-weight` does not
apply to `risk-parity` (its weights follow from the equal-risk condition),
and `optimize` cannot be combined with `leverage`.

Without `#meta leverage:on`, a weight > 100% is rejected (with a hint) and sums
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

## Suggesting assets to add

`portfodor -suggest portfolio.txt` analyses a portfolio's **macro-regime
coverage** and recommends catalog assets to add that fill the gaps. The four
regimes are the growth × inflation quadrants behind All-Weather- and
Dragon-style portfolios — `growth`, `deflation`, `inflation`, `crisis` — and
each catalog asset is mapped to the regimes it helps in from its factual tags
(asset class, strategy; see `datasets/assetmeta/`). A regime with little
weight is a gap.

It is **structure-first**: only assets that fill a gap are considered, and
each is then validated **out-of-sample** — the candidate is added at a
modest weight and a walk-forward checks that Sharpe and max-drawdown improve
*consistently across periods*, not in one lucky stretch. Because adding an
asset at a fixed weight fits nothing to the data, this measures robustness
rather than an over-fitted optimum. Suggestions are kept diverse (at most one
per asset class) and reported with the gap they fill, a suggested weight,
their correlation to the portfolio, and the out-of-sample win counts.

`-suggest` also flags **redundancies** — holdings that move almost
identically and share an asset class (three S&P 500 trackers are one bet, not
three). It prints to the terminal and exits, like `-verify-data`.

For a quick, **offline** read, `portfodor -coverage portfolio.txt` shows the
same coverage chart and then, for each gap, lists the catalog assets that
fill it (grouped by asset class) — no price downloads, no ranking, just the
menu of options. Run `-suggest` afterwards to rank and validate them.

By default coverage is organized by the four **macro regimes**.
`-framework factors` switches to a **risk-factor** lens (market, size,
value, momentum, quality, term, credit, alternative, cash) for both
`-coverage` and `-suggest`. The factor mapping is coarser — this catalog
holds many diversifiers (gold, commodities, managed futures, volatility)
that are not Fama-French factors and all land in *alternative* — so the
regime view stays the default.

## Main options

| Option | Default | Description |
|---|---|---|
| `-out` | `/tmp/portfodor-<timestamp>.html` | generated HTML file |
| `-data` | standard user cache | quote cache (JSON) |
| `-simdata` | embedded in the binary | source of simulated histories (directory for dev) |
| `-rebalance` | `90` | rebalance every N calendar days (0 = never) |
| `-start` | `2006-01-01` | desired start date |
| `-benchmark` | `^GSPC` | reference for Beta |
| `-currency` | `EUR` | convert every series (and the benchmark) to this currency; empty disables |
| `-cache-age` | `720h` (1 month) | cache freshness before re-downloading |
| `-assets` | | list `A,B,C`: each asset compared as a 100% portfolio |
| `-cli` | | curves and summary table in the terminal, no HTML |
| `-width` | `$COLUMNS` or 100 | width of the `-cli` chart (wider = more granularity) |
| `-warmup` | | pre-warm the built-in asset catalog then exit |
| `-verify-data` | | data doctor: check the referenced assets' quotes (or the whole catalog) for anomalies — bad points, gaps, stale feeds — then exit |
| `-suggest` | | recommend catalog assets to add for better regime coverage, flag redundant holdings, then exit |
| `-coverage` | | offline advisor: show which regimes/factors a portfolio misses and the catalog assets that fill them, then exit |
| `-framework` | `regimes` | classification for coverage and `-suggest`: `regimes` (macro quadrants) or `factors` (risk factors) |
| `-no-open`, `-no-simulate` | | do not open the browser / ignore SIM suffixes |

## Data

- **Resolution**: aliases → embedded ticker→ISIN list (European ETFs/funds)
  → built-in catalog of pinned resolutions → multi-source search
  (Yahoo, FT, Morningstar via Boursorama), the deepest series winning.
- **Currency**: every series is converted to the `-currency` (default EUR)
  using daily Yahoo FX crosses, so USD ETFs and EUR funds compare fairly;
  the earliest known rate is held flat before the FX history starts (with a
  warning), and unconverted (unknown-currency) assets are flagged.
- **Cache**: 1 month by default; a failed refresh **serves the stale data**
  with a stderr warning (charts may stop before today), and never deletes
  anything.
- **History extension** (`…SIM` identifiers only): first the
  `datasets/simdata/` files (below), otherwise a known proxy (VOO→^GSPC,
  BND→VBMFX, …), rescaled to the first real quote. The report flags every
  simulated portion.

## Simulated data (datasets/simdata/)

Complex assets (90/60 funds, managed futures…) are rebuilt by `pkg/simgen`
from long-history building blocks, validated against their real quotes,
then stored as self-documenting CSVs (method, validation, date) in
`datasets/simdata/`:

```sh
./portfodor -gen-simdata                   # regenerate everything (then make build to re-embed)
./portfodor -gen-simdata -dry NTSX         # validate without writing
```

Every series is built **only from quotes the tool itself can fetch**
(Vanguard/Yahoo funds with decades of history, the `^IRX` cash rate, gold and
oil futures) combined by the in-house composite, TSMOM trend, and regression
backcast engines — no third-party data is bundled. External index series are
used solely to cross-check quality during development, never shipped.

Bundled recipes and measured quality (daily / weekly correlation of returns
vs the real series; the real series is always grafted on top of the
simulation wherever it exists):

| Asset | Method (building blocks) | Validation (daily / weekly corr) |
|---|---|---|
| NTSX (UCITS) | 0.90×VFINX + 0.60×(VFITX−cash) + 0.10×cash (1991→) | 0.96 / 0.99 |
| NTSG (UCITS) | global 90/60 US/intl variant (1999→) | 0.39 / 0.86 (thin LSE listing, short overlap) |
| URTH (MSCI World) | 0.60×VFINX + 0.40×VTMGX (1999→) | 0.90 / 0.97 |
| IWDA (MSCI World) | 0.60×VFINX + 0.40×VTMGX (1999→) | 0.60 / 0.85 (GBP listing, short overlap) |
| VT (total world) | 0.60×VFINX + 0.30×VTMGX + 0.10×VEIEX (1999→) | 0.98 / 0.99 |
| RSSB (100/100 stocks+bonds) | VT composite + 1.0×(VFITX−cash) (1999→) | 0.95 / 0.99 |
| ZPRV (US small-cap value, UCITS) | DFSVX (DFA US Small Cap Value, 1993→), real grafted 2015 | 0.67 / 0.91 |
| SHY (1-3y Treasury) | VFISX short Treasury (1991→) | 0.81 / 0.89 |
| IEF (7-10y Treasury) | VFITX intermediate Treasury (1991→) | 0.95 / 0.96 |
| TLT (20+y Treasury) | VUSTX long Treasury (1986→) | 0.98 / 0.99 |
| ZROZ (25+y STRIPS) | 1.65×(VUSTX−cash) (1986→) | 0.97 / 0.97 |
| DBMF (managed futures) | 12-month TSMOM on a 7-market basket (2001→) | 0.52 / 0.55 |
| KMLM (managed futures) | 12-month TSMOM, 15% target vol (2001→) | 0.35 / 0.32 |
| CTA (managed futures) | 12-month TSMOM, 10% target vol (2001→) | 0.20 / 0.24 |
| Winton Trend-Equity (UCITS) | 0.60×VFINX + 0.40×VTMGX + 0.50×TSMOM trend (2001→) | 0.65 / 0.84 |
| Amundi Volatility, BH Macro | regression backcast **rejected** (R² 0.20 / 0.00) | real history only (2007→) |

Managed-futures correlations are modest: each fund runs a faster, partly
discretionary strategy that a single 12-month TSMOM rule only approximates.
The lower fidelity is accepted in exchange for full self-generation.
Discretionary strategies that cannot be honestly replicated with factors are
rejected below an R² floor rather than shipped as invented data; the matching
`SIM` identifiers then simply fall back to the real (shorter) history.

## Using it as a library

The repository is also a toolkit for writing other portfolio-processing
applications. Layout:

```
pkg/marketdata/   data: resolution (aliases, ISIN, catalog), multi-provider
                  sources, cache, fees, simdata, alignment
pkg/metrics/      statistics (CAGR, Sharpe, Sortino, drawdowns, Beta, IRR…)
pkg/optimize/     weights for max-sharpe / min-volatility / risk-parity
pkg/suggest/      regime coverage, redundancy and gap-filling suggestions
pkg/chart/        SVG charts (Line) and terminal (Term), shared palette
pkg/portfolio/    allocation file format + rebalanced simulation
pkg/report/       HTML and text rendering of the comparison model
pkg/simgen/       history reconstruction (composites, TSMOM, backcasts)
cmd/              the portfodor binary (report, warmup, gen-simdata)
datasets/         versioned data (embedded at build time) and its QA:
  simdata/          permanent simulated histories (spliced at runtime)
  assetmeta/        catalog asset metadata (classes, factors, regimes…)
  golden/           golden tests + frozen fixtures vs external references
data/             old local cache (replaced by the user cache)
```

Everything consumable as a library lives under `pkg/`; `cmd/` only contains
the CLI wiring and `golden/` the golden test suite.

Each package has its documentation page — calculation conventions included
(`go doc github.com/bpineau/portfodor/pkg/metrics`) — and runnable examples:

```go
import (
	"github.com/bpineau/portfodor/datasets"
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

// Parse a portfolio file, fetch each holding, then simulate (N-day
// rebalancing). The CLI additionally converts currencies and extends SIM
// histories; this is the core path.
spec, _ := portfolio.ParseFile("p.txt")
p := &portfolio.Portfolio{Name: spec.Name}
for _, h := range spec.Holdings {
	s, _ := client.Fetch(h.ID, time.Time{})
	p.Assets = append(p.Assets, portfolio.Asset{ID: h.ID, Weight: h.Weight, Fees: h.Fees, Series: s})
}
sim, _ := portfolio.Simulate(p, 90)

// Read the bundled asset catalog as typed datasets.Asset records (name, TER,
// UCITS, geography, sectors, asset class…); AssetMeta() returns the same data
// as raw JSON if you prefer your own struct.
for _, a := range datasets.Catalog() {
	_ = a.Name // a.Fees, a.Geography, a.AssetClass, a.UCITS…
}
// Resolve a ticker / alias / ISIN to its full record in one call:
iwda, ok := marketdata.Lookup("IWDA") // → (datasets.Asset, true)
_ = iwda.Fees                         // 0.20  (percent/yr)
```

- `datasets` — the versioned data embedded at build time; `Catalog()` returns
  the typed asset list (`Asset`), `AssetMeta()` the same data as raw JSON.
- `marketdata` — resolution (aliases, ISIN, catalog), `Lookup` for an asset's
  full metadata, multi-source downloads, cache, simdata, proxies.
- `suggest` — regime/factor coverage and gap-filling (consumes `datasets.Asset`).
- `metrics` — statistics over value series (returns, drawdowns, Beta).
- `chart` — pure-stdlib inline SVG charts.
- `portfolio` — allocation file parsing and rebalanced simulation.
- `report` — HTML report rendering.
- `simgen` — reconstruction engine (linear composites, TSMOM
  trend-following engine, regression backcasts) and validated recipes, all
  built from fetchable quotes only.

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
