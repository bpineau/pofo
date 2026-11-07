# pofo, for coding agents

Pofo is a standard-library-only Go toolkit for long-horizon portfolio analysis.
Its four faces are the reusable library, FIRE engine, portfolio visualizer and
embedded book. Review a change against all four; the CLI is one consumer of the
library. Live trading, broker connectivity, user accounts and personalized
financial advice are outside scope.

## Find the smallest useful entry point

1. `go doc .` routes questions to packages and states the units.
2. `go doc ./pkg/<name>` opens on "Start here" and documents conventions/errors.
3. [Library recipes](docs/usage/library/README.md) explain tasks. Start with the
   [complete offline workflow](examples/code/workflow.go) for a copyable program.
4. [Examples](examples/code/README.md) are single-file programs, not fragments
   requiring test helpers. `go run examples/code/<name>.go -h` lists options.
5. Use [the implementation map](docs/specs/package-map.md) only when changing
   internals. [docs/README.md](docs/README.md) indexes specialist rationale and
   validation records; read the relevant one before changing its feature.

| Question | Library | Ready script in `examples/code/` |
|---|---|---|
| Load bundled, CSV or cached/live data | `Client.Load`; `Bundled`; `ReadSeriesCSV` | `describe`, `export` |
| Describe a series, calendar years, crises | `analyze.Asset` / `analyze.Series`; `Series.Stats`; `metrics.CalendarReturns`, `DrawdownEpisodes`, `RollingCAGRs` | `stats`, `calendar`, `episodes`, `rolling` |
| Align returns, blend, regress, find worst months | `NewPanel`, `Panel.Mix` / `Pick` / `Track`; `metrics.Regress` | `blend`, `regress`, `worstmonths`, `correl` |
| Compare two series or a changed file | `analyze.Pair`; `PairStudy.WriteText` | `pair`, `oldnew`, `fees` |
| Change currency, purchasing power, cash benchmark | `Client.ConvertCurrency`; `Series.Deflate` / `Excess`; `CashIndex`; `analyze.Options` | `currency`, `workflow` |
| Study a portfolio | `portfolio.NewSpec` / `NewSpecWith`; `analyze.Portfolio`; `analyze.SeriesSource` for own data | `simulate`, `workflow` |
| Allocate and validate out of sample | `optimize.Solve`, `Frontier`, `FitEvaluate`; `Spec.RiskBudgets` | `optimize` |
| Uncertainty, robust statistics, benchmark attribution | `metrics.RegressHAC`, `ShrinkageCovariance`, `BlockBootstrap`, `Brinson` | `go doc ./pkg/metrics` examples |
| Data hygiene or backcast validation | `FindGaps`, `FindSpikes`; `simgen.Validate`, `Audit` | `scanbundle`; CLI `-verify-simdata -json` |
| FIRE scenarios and withdrawals | `scenario`, `decumul.Plan`, `replay.Run` | `fire`, `replay` |
| Render | `chart`, `compare.Compute`, `report.Render` | CLI report |

For an ad hoc investigation, copy the nearest script into gitignored `scratch/`
and edit it there. Use Go and pofo APIs, never Python, private cache JSON or
hand-parsed bundled CSVs. Report a missing API rather than duplicating its math.
Never commit scratch files. `make check` excludes scratch; bare `go test ./...`
and `go vet ./...` do not.

## Contracts to keep visible

- Weights and returns are fractions in Go, weights percent in portfolio files.
  Fees and annual quoted rates are usually percent (0.20 = 0.20 %), but
  `Series.LessFee`, simulation transaction costs and `RiskFree` are fractions.
  `Stats.Ulcer` is percent points and `Stats.CWARP` percent. See the complete
  [units table](docs/usage/library/README.md#units) before mixing APIs.
- `Client.Load` accepts a simple CSV or a single-series long export. For multiple
  exported series use `ReadSeriesCSV` and select by Symbol. `Offline=true` never
  downloads: cache at any age, then bundle, otherwise `ErrOffline`.
- A bare fund ID means real quotes; its `SIM` suffix opts into reconstructed
  history. `Bundled("IWDASIM")` is valid; `Bundled("IWDA")` is not a shortcut
  to that reconstruction. Catalog indices can have long bare IDs.
  SIM files carry known `SimulatedBefore` / `FullySimulated` provenance;
  legacy files without a frontier warn explicitly instead of inventing one.
- Use `NewPanel` for cross-series statistics: shared return intervals, no
  fabricated zero returns. Use `AlignSeries` for forward-filled valuation.
  Use `Series.Stats` after `Panel.Series` to retain its explicit cadence.
  Monthly panels mitigate different market closing times.
- CPI is a price-level index chosen for a consumption basket, independently of
  the investment's currency. `Deflate` refuses missing coverage by default;
  explicit extrapolation adds warnings. `CashIndex` takes annual percent rates
  and an explicit day-count basis, not a discount-yield conversion. Risk-free
  indices are total-return cash series in the evaluation currency.
- Dates are civil dates at midnight UTC. Prices are adjusted total-return
  closes by default. Raw prices plus cash dividends are a separate convention;
  combining dividends with adjusted closes double-counts income. Asset TERs
  already reside in prices; envelope fees and execution costs do not.
- With external flows, statistics use `SimResult.Index` (time weighted);
  `Values` and flows serve money outcomes and IRR. Scenario/decumul returns
  are real. Surface study/series `Warnings` to users.
  For caller-constructed series, `Series.DataWarnings()` includes provider
  qualifications; panels and CSV exports retain these even without `Source`.
  Undefined statistics remain NaN in Go and serialize as JSON null, not a
  measured zero.
- Optimizers take aligned return slices, asset order and explicit cadence.
  Resolve named bounds first. Apply `Spec.Train` to the calendar and clear it
  before solving; it is rejected if left unhandled. `FitEvaluate` fits training
  only; the caller supplies disjoint windows in the same asset order. A false
  `Feasible` is a failed search result, not proof no feasible portfolio exists.

## Contribution rules

- A quietly wrong number is the highest priority. Fix its cause and pin a
  regression before expanding features. New logic needs meaningful tests;
  every exported symbol/field needs godoc with units, defaults and errors.
- Keep the dependency graph simple: `metrics` owns slice math and imports no
  project package; `marketdata` owns Series transformations/calendars;
  `portfolio` owns simulation; `analyze` composes them into numbers-only studies;
  `compare` and `report` add presentation. No third-party Go dependency.
- Every package keeps `doc.go` and executable `Example` tests. Update its human
  guide in the same change. Guide Go snippets copy the named runnable example
  exactly (`TestDocSnippets`). Prefer self-contained examples for new APIs.
- `docs/specs/` holds rationale, validation evidence, traps and procedures only.
  No plans, backlogs or pre-implementation specs in tracked docs. The shipped
  API design belongs in godoc. README remains a short routing page.
- This repository is public. Never add personal data, holdings, amounts, names,
  home paths, server details or session narration. Fixtures are invented;
  example portfolios use public ticker identifiers and ISIN comments.
- Code and docs are English, except the French book, its source of truth.
  Use straight quotes/apostrophes and plain hyphens, never typographic dashes.
- Before trusting any new/rebuilt bundled series, specify and run external
  reference checks for CAGR, volatility and currency, and report divergences.
  Never widen golden tolerances casually. Verify a source live before claiming
  it is unavailable. Read [maintenance](docs/specs/maintenance.md) before data
  refreshes: source budgets, generator dependencies and freeze incidents matter.

## Commands and completion

```sh
make build       # ./pofo
make check       # fmt, vet, staticcheck, all tests, example compilation
make golden      # independent computation references
```

Tests never use the network: `httptest` for HTTP, `fstest.MapFS` for files.
`make check` must pass before every commit. Run `make golden` for computation
or data changes. Any FAIL, staticcheck line, gofmt diff or `examples: FILE` is
a failed gate. Documentation-only quick check:
`go test -run 'TestGodoc|TestDocSnippets|TestUnitsTable' .`.

Commit and push to `master` after the gates pass. No push workflow provides a
second check. A `v*` tag publishes the dependency and deploys production: **never
tag unless release/deployment was explicitly requested**. `finador` pins pofo
for market data/math/charts; `locador` pins EPUB/rendering. A requested consumer
release needs a deliberate tag and its dependency bump, never an incidental tag.

Detailed refresh commands, targeted tests, screenshots and book drift checks:
[maintenance](docs/specs/maintenance.md). Before domain changes (backcasts,
optimizer models, FIRE, web, book), follow the specific routing in
[domain workflows](docs/specs/domain-workflows.md) and [docs index](docs/README.md).
