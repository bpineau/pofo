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
- The disk cache file gains a `dividends` column. Raw series are cached as
  their own entries (`SYMBOL~raw.json`): the price-cleaning passes
  (`dropDropouts`, `mendScaleBreak`) drop and repair points, so a dual-column
  cache would desynchronize; each view is fetched, cleaned and cached
  independently (one extra download only for callers that want both views,
  which neither pofo nor finador does). Old cache files stay valid for
  adjusted reads. Currency conversion applies to dividend amounts too.
- `Raw` combined with a SIM suffix is an error: simulated histories are
  total-return by construction.
- Plain `Fetch` keeps returning adjusted closes, with `Dividends` attached
  whenever the source reported them; raw closes are reachable through
  `FetchExtended` only.

### 3. marketdata: point-in-time FX

- `Series.At(at time.Time) (value float64, on time.Time, ok bool)`:
  forward-fill lookup by binary search (ok=false before the first point).
- `Client.FXRate(ctx, from, to string, at time.Time) (float64, error)`:
  point-in-time conversion rate using the same direct `FROMTO=X` daily cross
  as `ConvertCurrency`, forward-filled via `Series.At`.
- `ConvertCurrency` itself stays as is (direct crosses, golden-safe).
- finador keeps its 50-line book-backed `Converter` (domain logic over the
  encrypted book's stored FX series; wrapping it around pofo types would
  convert storage formats on every valuation for no gain). Its FX refresh
  fetches `CCYUSD=X` series through plain `Fetch`, which already works.
  An `FXTable` abstraction was considered and dropped: no caller needs it.

### 4. marketdata: search, lookup, cache-less mode

- `Client.Search(ctx, query string) ([]Resolution, error)`: free-text
  multi-candidate resolution (name, ticker or ISIN; catalog pin first, then
  the existing Yahoo search), no series download.
- `NewClient("")` disables the disk cache entirely (loads nothing, saves
  nothing, including resolution and fees caches); `DefaultCacheDir`
  documents the distinction.
- `pkg/simgen` keeps its ctx-less `Fetcher` interface (a batch generator has
  no cancellation granularity to gain); it gains
  `WithContext(ctx, *marketdata.Client) Fetcher` so `cmd/pofo` can keep
  passing the client. `portfolio.Build`'s fetch callback is unaffected (the
  caller binds ctx in its closure).

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
  to domain types), `Intraday` via `Client.Intraday`; FX refresh keeps
  fetching `CCYUSD=X` series, now through the same adapter. The client is
  cache-less (`NewClient("")`) for privacy;
  finador's book stores the series and fetches incrementally, as today.
  yahoo.go, ft.go, morningstar.go, multi.go and their tests are deleted.
- `internal/market/convert.go`'s Converter stays: it is domain logic over
  the book's stored FX series.
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
