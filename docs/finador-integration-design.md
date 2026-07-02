# Finador integration: market data, metrics and chart enrichment

Design for enriching pofo so that the sibling finador project (a personal
encrypted wealth tracker, `../finador`) can drop its own market-data,
performance-math and chart code and consume pofo instead. Every addition is
generally useful to pofo; nothing is finador-specific.

Decisions taken with Ben (2026-07-02): enrich pofo AND migrate finador in the
same effort; migrate the marketdata API to `context.Context` in one breaking
sweep; chart flexibility via style knobs on `Options` plus a minimal preset;
raw closes and dividends exposed through `FetchOptions` on the existing
`Series` type.

## Gap analysis

finador's `internal/market` is a less-featured port of the same
Yahoo/FT/Morningstar providers pofo already has, so fetching migrates almost
for free. What pofo lacks:

1. Dividend events (finador books them as cash income).
2. Raw (dividend-unadjusted) closes: pofo serves adjusted closes only, which
   would double-count income for a tracker that values holdings at market
   price and records dividends separately.
3. Point-in-time FX: `Rate(from, to, date)` crossed through USD with
   forward-fill; pofo only reprices whole series.
4. Free-text symbol search (`asset add "MSCI World"`): pofo's `Resolve` takes
   identifiers only; the search exists internally, unexported.
5. Flow-aware metrics: TWR and daily returns neutralizing external cash
   flows; Sharpe/Sortino with a configurable risk-free rate.
6. Chart flexibility: finador's dialect is minimalist SVG (no grid, no
   background, area fill under the first curve, compact k/M labels,
   monospace), bare sparklines, legend-less donuts, braille terminal charts.
   pofo's chart `Options` has no style knobs.
7. `context.Context`: finador's web handlers need cancellation; pofo's
   `Client` takes none.
8. Cache-less operation: finador is privacy-first (one encrypted file); a
   plaintext quote cache on disk would reveal holdings.

## Design

### 1. marketdata: context sweep (breaking)

Every network-touching `Client` method gains `ctx context.Context` as first
parameter: `Fetch`, `FetchExtended`, `History`, `Intraday`, `Resolve`,
`Fees`, `ConvertCurrency`, plus the new `Search`, `FXCross`, `FXRate`.
`cmd/pofo` threads a `signal.NotifyContext` context. Tests and examples
updated mechanically. pofo is pre-1.0 and self-consumed; no compatibility
shims.

### 2. marketdata: dividends and raw closes

- `type Dividend struct { Date time.Time; Amount float64 }`: ex-date
  (00:00 UTC, like every Point), per-share amount in the series' native
  quote currency, sorted ascending.
- `Series.Dividends []Dividend`, populated when the source provides them
  (Yahoo `events=div`; FT and Morningstar NAVs have none).
- `FetchOptions.Raw bool`: false (zero value) keeps today's adjusted closes;
  true returns split-adjusted but dividend-unadjusted closes (Yahoo `close`
  semantics), the right series for quantity-times-price valuation with
  dividends booked as cash. Stooq closes are already unadjusted; fund NAVs
  are their own raw price.
- One Yahoo request already carries `close`, `adjclose` and dividend events:
  the disk cache file gains `raw` and `dividends` columns so switching views
  never refetches. Old cache files (no raw column) stay valid for adjusted
  reads; a raw read refetches. Currency conversion applies to dividend
  amounts too.
- `Raw` combined with a SIM suffix is an error: simulated histories are
  total-return by construction.
- Plain `Fetch` keeps returning adjusted closes, with `Dividends` attached
  whenever the source reported them; raw closes are reachable through
  `FetchExtended` only.

### 3. marketdata: point-in-time FX

- `Client.FXCross(ctx, ccy string, from time.Time) (*Series, error)`
  fetches the `CCYUSD=X` daily series (cached like any series).
- `type FXTable` built via `NewFXTable(map[string]*Series)` (key: currency,
  value: USD cross), with `Rate(from, to string, at time.Time)` and
  `Convert(amount float64, from, to string, at time.Time)`: USD pivot,
  forward-fill (earliest rate held flat before history starts), nil-safe.
- `Client.FXRate(ctx, from, to string, at time.Time)` one-shot convenience.
- `Client.ConvertCurrency` is rebuilt on FXTable (same behavior, golden-safe).
- finador feeds FXTable from the FX series stored in its encrypted book, so
  valuation works offline.

### 4. marketdata: search, lookup, cache-less mode

- `Client.Search(ctx, query string) ([]Resolution, error)`: free-text
  multi-candidate resolution (name, ticker or ISIN; catalog pin first, then
  the existing multi-source search), no series download.
- `Series.At(at time.Time) (value float64, on time.Time, ok bool)`:
  forward-fill lookup by binary search.
- `NewClient("")` disables the disk cache entirely (loads nothing, saves
  nothing); `DefaultCacheDir` documents the distinction.

### 5. metrics: flow-aware building blocks

- `type Flow struct { Date time.Time; Amount float64 }` (positive = money
  into the measured scope, booked at the start of its day).
- `TWR(dates []time.Time, values []float64, flows []Flow) float64`:
  chain-linked daily returns `(V_t - F_t)/V_{t-1}`, non-positive bases
  skipped.
- `FlowReturns(dates, values, flows) []float64`: flow-adjusted daily
  returns; Saturday/Sunday points are dropped so calendar-daily
  forward-filled series do not dilute volatility (a no-op for trading-day
  series).
- Exported `Volatility(returns []float64)`, `Sharpe(returns []float64, rf
  float64)`, `Sortino(returns []float64, rf float64)`: same math as `Compute`
  but with an explicit annual risk-free rate. `Compute` keeps its documented
  rf=0 convention (golden-safe).
- `Annualize(totalReturn float64, days int) float64`: CAGR of a cumulative
  return over a calendar-day span (365.25-day years).
- Already covered: `IRR` (finador's XIRR), `DrawdownEpisodes` (its max
  drawdown with peak/trough/recovery dates).

### 6. chart: style knobs and new kinds

- `Options` gains an embedded `Style` whose zero value reproduces today's
  warm-study look exactly:

  ```go
  type Style struct {
      Background  string                // CSS color; "none" = no background rect
      Font        string                // font-family
      FontSize    int
      HideGrid    bool
      HideAxes    bool
      HideLegend  bool
      Fill        bool                  // translucent area under the first series
      StrokeWidth float64
      YTicks      int                   // horizontal guides/labels (default 6)
      CornerDates bool                  // first/last date at the corners instead of time ticks
      TickFormat  func(float64) string  // y labels; see Compact
  }
  ```

- `StyleMinimal() Style` preset: transparent background, monospace, no grid,
  no axes, 4 y labels, corner dates, area fill: the finador dialect.
- `Compact(float64) string` exported k/M tick formatter (473.9k, 1.23M).
- `Line` honors all knobs; the other SVG kinds honor Background/Font where
  trivial.
- New `Sparkline(opt SparkOptions, values []float64) string`: bare polyline,
  no axes, `preserveAspectRatio="none"`. `SparkOptions{Width, Height int;
  Color string}` with small defaults (72x20, first palette color).
- `PieOptions.Hole float64` (inner radius fraction; default keeps today's
  donut) and `PieOptions.HideLegend bool` for bare donuts.
- `TermOptions.Braille bool`: braille dot rendering (2x4 dots per cell) in
  `Term`, multi-series: also upgrades pofo's own CLI charts.

### 7. finador migration

- `finador/go.mod`: `require github.com/bpineau/pofo` +
  `replace github.com/bpineau/pofo => ../pofo`.
- `internal/market` keeps the `Source` interface and gains one pofo-backed
  adapter: `Resolve` via `Client.Search` (first candidate), `Daily` via
  `FetchExtended{From, NoSim: true, Raw: true}` (points + dividends mapped
  to domain types), `Intraday` via `Client.Intraday`, FX refresh via
  `FXCross`. The client is cache-less (`NewClient("")`) for privacy;
  finador's book stores the series and fetches incrementally, as today.
  yahoo.go, ft.go, morningstar.go, multi.go and their tests are deleted.
- `internal/market/convert.go`'s Converter delegates to `marketdata.FXTable`
  over the book's FX series.
- `internal/perf` keeps `Report`/periods (presentation) but delegates TWR,
  daily returns, vol, Sharpe, Sortino, CAGR, XIRR and max drawdown to pofo
  metrics (Date <-> time.Time adapters).
- `internal/chart` is deleted; call sites use pofo chart with `StyleMinimal`
  and finador's own colors (couleur constants stay in finador).
- domain, store, keyring, portfolio, web, cli stay, apart from call sites.

## Error handling

- `ErrNotCovered` keeps its meaning (e.g. intraday for non-Yahoo ids);
  `FXTable.Rate` returns a descriptive error naming the missing currency and
  date; `Raw`+SIM returns an explicit error from FetchExtended.
- Cache-less clients degrade exactly like cache-miss paths today (fetch or
  fail; stale-cache fallback simply never applies).

## Testing

- pofo: httptest-stubbed unit tests (stubAllBases pattern) and runnable
  examples for every new API; `make check` and `make golden` stay green (no
  existing calculation changes; ConvertCurrency refactor is behavior
  preserving).
- finador: full suite passes; the adapter is tested against pofo's `*Base`
  URLs pointed at httptest servers, same pattern as pofo's own tests.
- End-to-end: build finador, run its CLI against a demo book with the
  network available to sanity-check refresh, charts and perf output.
