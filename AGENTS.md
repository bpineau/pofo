# pofo, for coding agents

Standard-library-only Go toolkit for long-horizon portfolios, with four faces:
library, FIRE engine, portfolio visualizer, embedded French/English book.
Judge every change against all four; the CLI is one library consumer. Out of
scope: live trading, brokers, user accounts, personalized advice.

## Orientation

`go doc .` routes questions to packages; `go doc ./pkg/<name>` opens on "Start
here". Recipes and the [units table](docs/usage/library/README.md#units) are in
`docs/usage/library/`; runnable scripts in `examples/code/` (`-h` for options).
Internals: [package map](docs/specs/package-map.md); before domain changes
(backcasts, optimizer, FIRE, web, book) read
[domain workflows](docs/specs/domain-workflows.md) and the
[docs index](docs/README.md); before data refreshes,
[maintenance](docs/specs/maintenance.md).

Ad hoc investigations: copy the nearest example into gitignored `scratch/`, use
pofo APIs (never Python, cache JSON or hand-parsed bundled CSVs), report a
missing API instead of duplicating its math. Never commit scratch.

## Traps (quietly wrong numbers)

- Go weights/returns are fractions; portfolio files use percent weights. Fees
  and quoted annual rates are percent, except `Series.LessFee`, simulation
  transaction costs and `RiskFree` (fractions). Check the units table.
- A bare fund ID is real quotes; the `SIM` suffix is reconstructed history
  (`Bundled("IWDASIM")`, never `Bundled("IWDA")`).
- Cross-series statistics go through `NewPanel` (shared intervals, no
  fabricated zeros); `AlignSeries` is for forward-filled valuation.
- Prices are adjusted total-return closes: adding cash dividends double-counts.
  Asset TERs are in prices; envelope fees and execution costs are not.
- With external flows, statistics use `SimResult.Index`, money outcomes use
  `Values`. Scenario/decumul returns are real.
- Undefined statistics stay NaN (JSON null), never zero. Surface `Warnings`.
- Optimizers: resolve named bounds and apply/clear `Spec.Train` before solving;
  `Feasible=false` is a failed search, not infeasibility proof.

## Rules

- A quietly wrong number comes first: fix the cause, pin a regression test.
- `metrics` imports no project package; `marketdata` owns series transforms,
  `portfolio` simulation, `analyze` studies, `compare`/`report` presentation.
  No third-party Go dependency.
- Every exported symbol has godoc (units, defaults, errors); each package keeps
  `doc.go` and `Example` tests; update its guide in the same change (guide
  snippets must match the runnable example, `TestDocSnippets`).
- `docs/specs/` holds rationale, validation evidence and procedures only, no
  plans or backlogs. README stays a short routing page.
- Public repo: no personal data, holdings, amounts, names, home paths or
  session narration. Fixtures invented; real public tickers with ISIN comments.
- English everywhere except the French book. Straight quotes, no typographic
  dashes.
- New or rebuilt bundled series: check CAGR, volatility and currency against
  external references and report divergences. Never widen golden tolerances
  casually.

## Gates

`make check` (fmt, vet, staticcheck, tests, examples) must pass before every
commit; `make golden` too for computation or data changes. Docs-only quick
check: `go test -run 'TestGodoc|TestDocSnippets|TestUnitsTable' .`.
A `v*` tag deploys production and publishes the module for finador/locador:
never tag unless explicitly asked.
