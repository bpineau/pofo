# pofo, for coding agents

pofo is a dependency-free (stdlib-only) Go toolkit for tracking and designing
investment portfolios, plus one CLI (`cmd/pofo`) built on it. Everything the
CLI does is reachable as a library under `pkg/`.

Read this file first; it is the cheapest way in. Details live in each
package's `doc.go` (`go doc ./pkg/<name>` renders it) and in `README.md`
(user-facing, CLI-oriented).

## Commands

```sh
make build     # ./pofo binary, pkg/datasets/ embedded via go:embed
make test      # go test ./...  (unit tests + runnable examples, NO network)
make lint      # go vet + staticcheck
make check     # fmt-check + lint + test: run this before any commit
make golden    # computation goldens vs frozen external references
make simdata   # regenerate pkg/datasets/simdata/ (network) then rebuild
make broadsample # regenerate the JST broad-sample panel (network) then rebuild
make euro-refdata # regenerate the euro-area reference series (network) then rebuild
make sp500-refdata # regenerate the month-end SP500-USD reference (network); run make simdata after
```

Tests never touch the network: HTTP sources are faked with `httptest`
(`stubAllBases` in `pkg/marketdata/client_test.go`), file sources with
`fstest.MapFS`. Keep it that way.

## Verifying changes cheaply

- Report/chart changes: `scripts/report-shot.sh [file] [out-prefix]` builds,
  renders the report with every section unfolded and screenshots it full-page
  (needs Chrome and a warm quote cache; `./pofo -warmup` once). Crop a region
  with `sips -c <height> 1500 --cropOffset <y> 0 out.png --out crop.png` and
  read the PNG. `examples/dragon-decumulation-household.txt` is the reference
  fixture: 9 holdings across a stacked 90/60, long bonds, two trend funds,
  gold, linkers and small-value equity, so it exercises every report block.
- Chart hover/tooltip data: grep the rendered HTML for
  `<metadata class="hover">` and replay the front-end math on the JSON
  payload in node/python instead of driving a browser.
- One golden at a time: `go test -v -run TestGoldenGold ./pkg/datasets/golden/`.
  Chart snapshots moved on purpose? Regenerate with
  `UPDATE_SNAPSHOTS=1 go test ./pkg/chart -run TestChartSnapshots` and justify
  the diff in the commit message.
- Catalog edits: `make test` revalidates `assets.json`;
  `./pofo -verify-data -assets <id>` checks a single asset end to end.

## Map

| Path | What lives there |
|---|---|
| `pkg/marketdata` | fetch/cache daily + intraday prices; identifier resolution (alias, ticker, ISIN); FX conversion; SIM history extension; data doctor (`Verify`) |
| `pkg/metrics` | risk/return statistics on dated value series (CAGR, Sharpe, drawdowns, IRR, variance ratio, rolling, CWARP) |
| `pkg/portfolio` | portfolio file format (`Parse`), `Build` (spec + fetch callback -> Portfolio), `Simulate` (rebalancing, fees, flows, leverage, per-holding return attribution incl. monthly folding) |
| `pkg/optimize` | long-only weights: max-sharpe, min-volatility, risk-parity, max-sortino, return-to-drawdown, min-ulcer, max-worst-5y, cwarp |
| `pkg/permanent` | tactical Permanent Portfolio 2.0 (Darcet): reads `datasets.MacroPanel` into a growth×inflation + monetary regime, quadratically-damped four-sleeve allocation, monthly-real backtest, coarse `Regime.Quadrant` view (used by the report's regime strip); see `docs/darcet-permanent-portfolio-design.md` |
| `pkg/suggest` | macro-regime/factor coverage, look-through composition splits (asset classes, geography, currency exposure, equity sectors, duration), redundancy, gap-filling suggestions |
| `pkg/scenario` | synthetic real-return paths: parametric Student-t, block/stationary bootstrap, historical cohorts, behind one `Source` interface |
| `pkg/decumul` | withdrawal/FIRE engine over a `scenario.Source`: ruin probability, outcome metrics, solvers, sweeps; `web/` = embedded live UI |
| `pkg/firebook` | the FIRE book: embedded French decumulation handbook (markdown articles under `assets/book/fr/` + manifest + renderer + handler), served by the fire UI at `/book/fr/`; plan, depth conventions and progress ledger in `docs/fire-book-design.md` |
| `pkg/simgen` | rebuilds the missing past of complex assets (composites, TSMOM, regression backcasts) into simdata files |
| `pkg/chart` | stdlib-only SVG + terminal charts |
| `pkg/report` | HTML/text rendering of the comparison model |
| `pkg/datasets` | embedded data: `assetmeta/assets.json` catalog, `simdata/` CSVs, `refdata/`, `broadsample/` (JST per-country real returns for the FIRE empirical model), `cape/` (Shiller CAPE, FIRE valuation anchor), `macropanel/` (OECD monthly multi-country macro drivers: IP/CPI/rates/share prices, for regime & growth-inflation-breadth work), `golden/` (frozen-fixture tests) |
| `cmd/pofo` | CLI wiring only, one file per concern: `main.go` (flags + mode dispatch + terminal output), `fetch.go`, `page.go` (HTML report assembly), `composition.go` (pies/coverage), `contrib.go` (contribution charts), `suggest.go`, `simdata.go`, `optimize.go`, `fire.go`, `permanent.go` |
| `docs/` | design docs and plans, one per feature; read before reworking a feature (`docs/README.md` is the one-line index) |
| `examples/` | portfolio files for the CLI (also exercised by `make demo`) |

Root `doc.go` describes the layering and the typical pipeline.

## The core pipeline (library)

```go
ctx := context.Background()
client := marketdata.NewClient(marketdata.DefaultCacheDir()) // "" = no disk cache
spec, _ := portfolio.ParseFile("p.txt")
p, _ := portfolio.Build(spec, portfolio.BuildOptions{
    Fetch: func(id string) (*marketdata.Series, error) {
        return client.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "EUR"})
    },
    Fees: func(id string) (float64, bool) { base, _ := marketdata.SplitSim(id); return client.Fees(ctx, base) },
})
sim, _ := portfolio.Simulate(p, 90)          // rebalance every 90 days
stats, _ := metrics.Compute(sim.Dates, sim.Index)
```

Every step is also reachable individually (`Fetch`, `ReadSimdataFS`,
`ExtendBack`, `ConvertCurrency`, `Trim`, ...) when a caller needs to deviate.

## Conventions and traps (do not guess, check here)

- UNITS, the number one trap. Fees and rates mix two conventions:
  - PERCENT per year: `portfolio.Holding.Fees`, `Portfolio.EnvelopeFees`,
    `Portfolio.BorrowSpread`, `marketdata.Client.Fees` (0.85 = 0.85 %/yr).
  - FRACTION per year: everything in `pkg/simgen` (fees, vol targets:
    0.0085 = 0.85 %/yr), all of `pkg/metrics` outputs except `Stats.Ulcer`
    (percent points), returns everywhere (0.04 = +4 %).
  - Weights: FRACTIONS in `portfolio.Asset.Weight`/`Holding.Weight` (sum
    to 1), PERCENT in portfolio files and `Holding.RawWeight`.
- Dates: every `marketdata.Point.Date` is normalized to 00:00 UTC. Metrics
  match series by exact `time.Time` equality; keep the invariant.
- `marketdata.Align` requires `start` at or after every series' first
  quote, otherwise it forward-fills zeros. `portfolio.Simulate` computes
  that window for you; direct callers must too.
- SIM convention: a bare id (`VOO`) = real quotes only; the `SIM` suffix
  (`VOOSIM`) also splices simulated/proxy history in front.
  `Client.FetchExtended` handles it; plain `Fetch` does not.
- Asset TERs are ALREADY reflected in prices: `Simulate` never deducts
  them (informational). Envelope fees (`extra-fees`) are NOT in prices and
  are deducted daily.
- Closes are ADJUSTED (total-return) by default; `Series.Dividends` +
  adjusted closes double-counts income. Valuation consumers use
  `FetchOptions.Raw` (unadjusted closes + dividends as cash); Raw + SIM
  suffix is an error.
- With external flows, `SimResult.Values` follows the money while
  `SimResult.Index` is the time-weighted series: compute statistics and
  comparisons on `Index`, money outcomes (IRR) on `Values` + flows.
- `pkg/scenario` and `pkg/decumul` work in REAL terms (inflation removed)
  and periodic returns; deflate nominal series first (`scenario.Deflate`).
- Annualization: 252 trading days, zero risk-free rate, CAGR over
  365.25-day years. Comparisons with PortfolioVisualizer et al. differ for
  documented reasons (see `pkg/metrics/doc.go`).
- Rate symbols (`^IRX`, `^FVX`, `^TNX`, `^TYX`) and `^VIX` are annualized
  percent LEVELS, not prices; `^HICP-<geo>` and `^CPI-US` are index levels;
  all chart fine but never belong in a return computation directly.
- A surprising CAGR/vol is usually a RESOLUTION bug, not a math bug: read the
  "resolved X -> name" log line first (a fuzzy search may have matched and
  cached an unrelated fund; delete that cache entry).
- `-gen-simdata <ID>` writes `simdata/<CanonicalID(ID)>.csv`: make the id
  canonical BEFORE generating (an alias collision silently overwrites another
  asset's file) and check `git diff --stat pkg/datasets/simdata` after.

## House rules

- Stdlib only. Do not add a third-party dependency.
- English for all code, godoc and docs. Never write an em-dash.
- Every package keeps a `doc.go` (conventions included) and runnable
  `example_test.go` examples; extend them with any new API.
- `make check` must pass; new logic comes with tests (the bar is high:
  most packages are at 75-97 % coverage).
- Calculation changes must keep `make golden` green; if a golden moves,
  justify it against the external reference, never retune the tolerance
  casually.
- Commit and push directly to `master` once `make check` passes.

## Common tasks

- Add a catalog asset: edit `pkg/datasets/assetmeta/assets.json` (see its
  README for the schema and vetting rules), then `make test` revalidates.
- Add a ticker alias: `pkg/marketdata/aliases.go`.
- New simulated history: add a recipe in `pkg/simgen/recipes.go`, validate
  with `./pofo -gen-simdata -dry <ID>`, generate with `make simdata`.
- New statistic: `pkg/metrics` + tests + a golden anchor if externally
  checkable; expose it in `report.StatRow` via `cmd/pofo/page.go`
  (`buildStatRows`) if the CLI should show it.
- Report per-portfolio blocks (composition pies, coverage bars, realized
  contribution charts): assembled in `cmd/pofo` (`breakdownPies`,
  `coverageBars` in `composition.go`; `contributionCharts` in `contrib.go`) from
  `pkg/suggest` composition splits, `SimResult.(Monthly)Contributions` and
  `permanent.Regime.Quadrant`; rendering primitives live in `pkg/chart`
  (`DivergingStack`, `BarMatrix`, `Pie`), the template and its instant-tooltip
  layer in `pkg/report/html.go`.
- New CLI mode: a `run*` function in its own `cmd/pofo/<mode>.go` file, but push any
  reusable logic down into a `pkg/` package first (see `FetchExtended`
  and `portfolio.Build`, which were extracted exactly that way).
- FIRE/decumulation work: read `docs/decumulation-fire-design.md` first;
  the follow-up backlog is `docs/decumulation-fire-program-2026-07.md`.
- Tactical Permanent Portfolio / Darcet / growth-inflation-regime work: read
  `docs/darcet-permanent-portfolio-design.md` first (complete findings,
  algorithms, data sources, and the empirical-vs-a-priori epistemic ledger);
  the macro drivers live in `pkg/datasets/macropanel`.
- Eurozone Efficient Core (NTSZ) / euro-native backcasts, incl. the long euro
  govt sleeve (DBXG, `dbxgRecipe`): read
  `docs/ntsz-eurozone-efficient-core-design.md` first. The deep euro reference
  series (`EMU-EUR`, `EUROGOV-EUR{,-DAILY}`, `EUROGOV-LONG-EUR{,-DAILY}` for the
  25+ segment, `DECASH-EUR`) come from DBnomics via `cmd/gen-euro-refdata`
  (`make euro-refdata`); note the equity-leg daily-vol/FX caveat there, and that
  the long sleeve is a real `TreasuryTR` long-bond reconstruction (never a
  levered short bond, which overstates a bond bull).
