# `#meta currencies:USD,EUR` — multi-currency comparison

## Motivation

A portfolio built from USD-quoted assets (e.g. `examples/claude-dragonlite.txt`:
NTSG, DBMF, XAUUSD) carries currency risk that a EUR investor feels but a USD
investor does not. Today the whole pipeline runs in a single base currency
(`-currency`, default EUR), so there is no way to see the same portfolio through
a US lens and a EUR lens side by side. This feature adds a per-portfolio
directive that expands one portfolio into one column per requested currency,
each with its own numeraire and its own inflation deflator, so the difference
between columns exposes the true impact of currency risk (nominal FX effect) and
local inflation (real effect).

## Goal (one sentence)

`#meta currencies:USD,EUR` runs the same portfolio once per listed currency and
presents the results as adjacent columns in the comparison table and chart, with
both nominal and per-currency real statistics.

## Non-goals (v1)

- Optimizing under FX (per-currency weights): `#meta optimize` combined with
  `#meta currencies` is rejected with a clear error, exactly like the existing
  `optimize` + `leverage` guard. Optimizing a covariance that itself shifts with
  the exchange rate is a distinct feature and is noted as a future extension.
- A currency-specific financing rate for leverage. `^IRX` stays as the financing
  proxy (it is an annualized rate level, already numeraire-independent); this is
  a pre-existing simplification and is out of scope here.
- Wiring new CPI series. Currencies without a bundled CPI (GBP, JPY, ...) still
  work: their nominal columns render, their real columns are omitted
  (best-effort, matching today's behaviour).

## What "measure" means

Confirmed during brainstorming: **nominal + real, per currency.** Each column:

1. converts every asset into that column's currency (numeraire),
2. computes nominal statistics (CAGR, vol, Sharpe, drawdowns, ...), which differ
   across columns because of the exchange-rate path,
3. computes real statistics via that currency's CPI deflator (`^CPI-US` for USD,
   `^HICP-FR` for EUR), which differ because of both FX and local inflation.

Reading: `column(EUR) - column(USD)` = FX effect (nominal rows) + inflation
effect (real rows).

## Scope of combinability

Confirmed: **M portfolios x N currencies**, all in one table. Any portfolio may
declare its own currency list; portfolios without the directive keep using the
global `-currency` flag (one column, unchanged). Column count is M x N; labels
carry the currency so wide tables stay legible.

## Architecture

### Parsing (`pkg/portfolio`)

- New field `Spec.Currencies []string`.
- `#meta currencies:USD,EUR` is parsed in `applyMeta`: split on comma, trim,
  uppercase, validate each token is a 3-letter ASCII alpha code, deduplicate
  (preserving order), reject an empty result. Also kept verbatim in `Spec.Meta`
  like every other directive.
- When `Currencies` is non-empty and `Optimize != nil`, `ParseString`/`Parse`
  returns an error (add to the existing combination guard near
  `parse.go:166`).
- Godoc on `Spec` documents the new field.

### CLI expansion (`cmd/pofo/main.go`)

- Effective currency list of a spec: `spec.Currencies` if set, else
  `[]string{opt.currency}`.
- **Fetch:** assets are fetched per (currency, id). `seriesByID` becomes
  `map[string]map[string]*marketdata.Series` (currency -> id -> series); each
  distinct (currency, id) pair is fetched once via `fetchAsset` with
  `opt.currency` overridden by the loop currency. (A small helper threads the
  currency into `fetchAsset` / `FetchOptions.Currency`.)
- **Simulate:** the build/simulate loop becomes
  `for each spec { for each cur { Build(BaseCurrency=cur) -> Simulate -> result } }`.
  Each `result` records its currency. The portfolio name gets a ` (USD)` suffix
  when the spec declared a currency list (i.e. is multi-currency).
- **Deflator:** the single `deflator`/`hasDeflator` pair is replaced by a
  `map[string]*marketdata.Series` (currency -> CPI), populated for the distinct
  currencies actually used, via the existing `inflationSeries`. Each result uses
  the deflator of its own currency to fill `realStats`/`hasReal`; the existing
  `metrics.Compute(deflate(...))` path is unchanged.
- **Benchmark:** fetched per distinct currency used (cache-backed); each result
  compares against the benchmark in its own currency for Beta/CWARP.
- **cashRate (`^IRX`):** unchanged, fetched once when any spec is leveraged.
- **Common window:** computed across all results (currency variants included),
  same logic as today. Variants of one portfolio share the asset date range, so
  the window stays stable.

### Rendering (`report` / CLI)

- No structural change: the report and `renderCLI` already iterate over
  `results`, and real columns are already gated per result by `hasReal`. Extra
  columns flow through.
- On the chart, one portfolio in USD vs EUR rebased to 100 diverges only by the
  cumulative FX return: a direct visual reading of currency risk.
- Adjust the "All series converted to X" note (~`main.go:1470`) and the deflator
  disclaimers to name the set of currencies in play rather than a single one.

## Error handling

- `#meta optimize` + `#meta currencies` -> parse error naming both directives.
- Empty / malformed currency token -> parse error
  (`#meta currencies: invalid code %q`).
- Currency with no bundled CPI -> no error; nominal columns render, real columns
  omitted (existing best-effort).
- FX cross unavailable for a currency -> surfaced by the existing fetch/convert
  error path (unchanged).

## Testing

- `parse_test.go`: valid list (`USD,EUR` -> `["USD","EUR"]`, dedup, order),
  malformed token rejected, `optimize`+`currencies` rejected.
- A CLI-level or portfolio-level test that a two-currency spec yields two results
  with distinct currencies and distinct stats, using the FX-stubbed
  `stubAllBases` fake (no network).
- `make check` and `make golden` stay green (no calculation change to existing
  single-currency paths).

## Docs

- Godoc on `Spec.Currencies` and the `applyMeta` case.
- README: add a row to the `#meta` directive table.
- CLAUDE.md: mention in the portfolio-format notes if the directive list is
  enumerated there.
- `example_test.go` in `pkg/portfolio` covering a `#meta currencies` parse.
- Add a commented `#meta currencies:USD,EUR` line at the top of
  `examples/claude-dragonlite.txt` (the motivating case).

## Future extensions (not v1)

- `optimize` under a chosen currency (per-currency weights).
- Currency-specific financing rate for leverage.
- A dedicated "FX contribution" row decomposing the nominal gap between columns.
