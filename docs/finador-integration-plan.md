# Finador Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enrich pofo (marketdata, metrics, chart) with everything finador
needs, then migrate finador onto pofo, deleting its own market/perf-math/chart
code.

**Architecture:** Grow pofo's existing packages in place (design:
`docs/finador-integration-design.md`). Pure-math and chart tasks first (no
cross-dependencies), then the marketdata context sweep and new fetch
capabilities, then the finador migration which validates the API fit.

**Tech Stack:** Go stdlib only (pofo house rule). finador keeps its existing
deps; pofo is added via `replace => ../pofo`.

## Global Constraints

- pofo: stdlib only, no third-party dependency.
- pofo: `make check` (fmt + vet + staticcheck + tests) green before every commit; commit and push to master.
- pofo: `make golden` must stay green; no existing calculation or zero-value chart output may change.
- Never write an em-dash anywhere.
- Every new pofo API gets godoc, a unit test, and (for the main entry points) a runnable example.
- Tests never touch the network (httptest / fstest patterns).
- finador: `go build ./... && go test ./...` green before each commit in ../finador.

---

### Task 1: metrics flow-aware building blocks

**Files:**
- Create: `pkg/metrics/flows.go`, `pkg/metrics/flows_test.go`
- Modify: `pkg/metrics/doc.go` (mention flow-aware entry points), `pkg/metrics/example_test.go` (one example)

**Interfaces:**
- Produces: `type Flow struct { Date time.Time; Amount float64 }`,
  `func TWR(dates []time.Time, values []float64, flows []Flow) (float64, bool)`,
  `func FlowReturns(dates []time.Time, values []float64, flows []Flow) []float64`,
  `func Volatility(returns []float64) float64`,
  `func Sharpe(returns []float64, rfAnnual float64) float64`,
  `func Sortino(returns []float64, rfAnnual float64) float64`,
  `func Annualize(totalReturn float64, days int) float64`.
- Consumes: existing `tradingDaysPerYear`, `daysPerYear`, `Mean`.

- [ ] **Step 1: failing tests** in `flows_test.go`: TWR with a mid-window
  contribution is flow-neutral (two days flat at 100, deposit 50 on day 2 with
  value 150: TWR = 0); TWR compounding without flows equals V_n/V_0-1;
  FlowReturns drops Saturday/Sunday points and adjusts for the flow;
  Volatility/Sharpe/Sortino against hand-computed values (rf=0 matches
  Compute's fields on the same returns; rf>0 shifts as expected); Annualize
  doubles in ~2 years to ~41.4%/yr; degenerate inputs (len<2, mismatched
  lengths) return ok=false / NaN / 0 as documented.
- [ ] **Step 2: run** `go test ./pkg/metrics/ -run 'TWR|FlowReturns|Volatility|Sharpe|Sortino|Annualize' -v` and see FAIL (undefined symbols).
- [ ] **Step 3: implement** `flows.go`:

```go
// Flow is an external cash movement into (positive) or out of (negative) the
// measured scope, booked at the start of its day.
type Flow struct {
	Date   time.Time
	Amount float64
}

// TWR chain-links flow-neutralized daily returns r_t = (V_t - F_t)/V_{t-1}.
// ok is false when the series is too short or lengths mismatch. Days with a
// non-positive base are skipped.
func TWR(dates []time.Time, values []float64, flows []Flow) (float64, bool) {
	if len(dates) != len(values) || len(values) < 2 {
		return 0, false
	}
	byDay := flowsByDay(flows)
	total := 1.0
	for i := 1; i < len(values); i++ {
		if values[i-1] <= 0 {
			continue
		}
		total *= (values[i] - byDay[dates[i]]) / values[i-1]
	}
	return total - 1, true
}

// FlowReturns yields the flow-adjusted weekday returns of a value series.
// Saturday and Sunday points (forward-filled flats in calendar-daily series)
// are dropped so they do not dilute volatility; trading-day series are
// unaffected. Days with a non-positive base are skipped.
func FlowReturns(dates []time.Time, values []float64, flows []Flow) []float64 { ... }

// Volatility, Sharpe, Sortino: same conventions as Compute (252 trading
// days, sample stdev, arithmetic annualization) with an explicit annual
// risk-free rate; NaN when undefined.
// Annualize: (1+total)^(365.25/days) - 1; 0 when days<=0 or total<=-1.
```

  flowsByDay keys by the exact `time.Time` (dates are normalized upstream, doc
  the invariant). Sharpe = `(Mean(r)*252 - rf) / Volatility(r)`. Sortino
  downside target = `rf/252`, denominator `sqrt(sum((r<t)^2)/n)*sqrt(252)`.
- [ ] **Step 4:** `go test ./pkg/metrics/` PASS; check rf=0 equivalence with Compute in the test.
- [ ] **Step 5:** update doc.go + add `ExampleTWR`; `make check`; commit `metrics: flow-aware TWR, returns and risk ratios with explicit risk-free rate`.

### Task 2: chart Style knobs + StyleMinimal + Compact

**Files:**
- Create: `pkg/chart/style.go`, `pkg/chart/style_test.go`
- Modify: `pkg/chart/svg.go` (Line consults Style), `pkg/chart/doc.go`, `pkg/chart/example_test.go`

**Interfaces:**
- Produces: `type Style struct { Background, Font string; FontSize int; HideGrid, HideAxes, HideLegend, Fill bool; StrokeWidth float64; YTicks int; CornerDates bool; TickFormat func(float64) string }`;
  `Options` gains `Style Style`; `func StyleMinimal() Style`; `func Compact(v float64) string`.
- Zero-value `Style` must render byte-identical output to today's `Line`.

- [ ] **Step 1: failing tests**: `TestLineZeroStyleUnchanged` (golden-ish: render a fixture with zero Options.Style before/after refactor: assert current expected substrings stay: `#FFFDF9` rect, grid line, axis line); `TestLineMinimalStyle` (no `<rect`, `ui-monospace`, `fill-opacity`, corner date labels, no grid lines); `TestCompact` (1234567 -> "1.23M", 473900 -> "473.9k", 12.345 -> "12.35").
- [ ] **Step 2:** run, FAIL (Style undefined).
- [ ] **Step 3: implement.** In `style.go`: Style, StyleMinimal (Background "none", Font "ui-monospace,monospace", FontSize 11, HideGrid, HideAxes, Fill, YTicks 4, CornerDates, TickFormat Compact), Compact (port finador formatCompact + trimZero). In `svg.go` Line: resolve defaults (`bg := #FFFDF9 unless st.Background; skip rect when "none"`, font, fontSize 12/title +4, stroke 1.8, yTicks 6, tickFmt), gate grid lines on !HideGrid (labels always), gate axis lines on !HideAxes, gate legend on multi && !HideLegend, CornerDates replaces timeTicks loop with two corner `<text>` labels of first/last date (format 2006-01-02, or 15:04 for sub-day spans), Fill draws `<polygon ... fill-opacity="0.07"/>` under the first series before the paths.
- [ ] **Step 4:** `go test ./pkg/chart/` PASS including pre-existing tests untouched.
- [ ] **Step 5:** doc.go + example; `make check`; commit `chart: style knobs on Options, minimal preset, compact tick formatter`.

### Task 3: chart Sparkline

**Files:**
- Create: `pkg/chart/spark.go`, `pkg/chart/spark_test.go`
- Modify: `pkg/chart/doc.go`, `pkg/chart/example_test.go`

**Interfaces:**
- Produces: `type SparkOptions struct { Width, Height int; Color string }`,
  `func Sparkline(opt SparkOptions, values []float64) string`.

- [ ] Steps: failing test (renders one `<polyline`, `preserveAspectRatio="none"`, no `<text`; "" for <2 finite values; NaN skipped; defaults 72x20 and PaletteColor(0)); implement (port finador spark.go onto values-only input, 2px pad, stroke-width 1.3); test PASS; docs; `make check`; commit `chart: bare Sparkline`.

### Task 4: chart Pie Hole + HideLegend

**Files:**
- Modify: `pkg/chart/pie.go`, `pkg/chart/pie_test.go`, `pkg/chart/doc.go`

**Interfaces:**
- Produces: `PieOptions` gains `Hole float64` (inner/outer radius ratio, (0,1); 0 keeps today's 0.6) and `HideLegend bool` (bare square donut: cy=w/2, outerR=w/2-2, h=w, no legend rows, no title row shift).

- [ ] Steps: failing tests (default output unchanged on a fixture; HideLegend
  output has no `<rect x="6"` legend swatch and height==width; Hole 0.55
  changes innerR accordingly); implement (innerR = outerR*hole with hole
  default 0.6 = 42/70; bare geometry branch when HideLegend); PASS; docs;
  `make check`; commit `chart: donut hole ratio and legend-less pies`.

### Task 5: chart Term braille mode

**Files:**
- Modify: `pkg/chart/term.go`, `pkg/chart/term_test.go`, `pkg/chart/doc.go`, `pkg/chart/example_test.go`

**Interfaces:**
- Produces: `TermOptions` gains `Braille bool`: the plot area renders 2x4
  braille dots per cell (U+2800 + bits) instead of one marker per cell;
  gutter, value labels, legend, bottom axis and time labels stay identical.

- [ ] **Step 1: failing test**: `TestTermBraille` renders a rising series and
  asserts output contains runes in U+2800..U+28FF and the same gutter/axis
  frame as plain Term; a two-series braille chart with Color=false still
  distinguishes series (per-series braille layers joined by later-wins cell).
- [ ] **Step 2:** FAIL.
- [ ] **Step 3: implement.** In Term, when opt.Braille: dot grid `rows*4 x
  plotW*2`, sample each dot column (x in 0..plotW*2) at
  `t = tmin + (tmax-tmin)*x/(plotW*2-1)`, `dotRow(v)` over `rows*4-1`,
  vertical connection like the marker path, per-dot series index (later
  series win). Cell composition:

```go
func brailleBit(dx, dy int) rune {
	bits := [4][2]rune{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}
	return bits[dy][dx]
}
```

  Cell color/marker: the series owning the most dots in the cell, rendered
  with the existing ANSI color when opt.Color (plain otherwise; braille
  glyphs already differ so plainMarkers are not needed).
- [ ] Steps 4-5: PASS; docs+example; `make check`; commit `chart: braille rendering mode for Term`.

### Task 6: marketdata context sweep

**Files:**
- Modify: every file in `pkg/marketdata` with network paths (`client.go`,
  `yahoo.go`, `stooq.go`, `ft.go`, `morningstar.go`, `boursorama.go`,
  `eurostat.go`, `fred.go`, `fees.go`, `intraday.go`, `currency.go`,
  `extended.go`, `resolve.go`) and their tests + `example_test.go`, `doc.go`
- Modify: `pkg/simgen/simgen.go` (add `WithContext`), `cmd/pofo/main.go`
- Test: existing suites (behavior unchanged) + `TestFetchCanceledContext`

**Interfaces:**
- Produces (public signatures after the sweep):
  `Fetch(ctx context.Context, id string, from time.Time) (*Series, error)`;
  `FetchExtended(ctx context.Context, id string, opt FetchOptions) (*Series, error)`;
  `History(ctx context.Context, symbol string, from time.Time) (*Series, error)`;
  `Intraday(ctx context.Context, id string) (*IntradaySeries, error)`;
  `Resolve(ctx context.Context, id string) (Resolution, error)`;
  `Fees(ctx context.Context, id string) (ter float64, ok bool)`;
  `ConvertCurrency(ctx context.Context, s *Series, target string, from time.Time) (*Series, time.Time, error)`.
- Produces: `simgen.WithContext(ctx context.Context, c *marketdata.Client) Fetcher`
  (simgen.Fetcher itself unchanged).

- [ ] **Step 1: failing test**: `TestFetchCanceledContext` in client_test.go:
  canceled ctx against a stubbed slow server returns `context.Canceled`
  promptly (also covers the retry sleep: use a handler returning 500 so the
  retry path hits the ctx-aware wait).
- [ ] **Step 2: sweep.** `do(ctx, ...)` uses `http.NewRequestWithContext` and
  replaces `time.Sleep(delay)` with `select { case <-ctx.Done(): return nil,
  ctx.Err(); case <-time.After(delay): }`; thread ctx through get/post,
  yahooGet, fetchYahoo, fetchYahooIntraday, search, fetchStooq, fetchFT,
  ftSearch, fetchMorningstar, boursoramaMorningstarID, HICP/eurostat, FRED,
  fees fetchers, history, cachedHistory, historyFT/MS/ForResolution,
  fetchISIN, fetchTicker, cachedResolutionHistory, resolveBest, fxHistory,
  convertTo, and the public methods above. Tests/examples pass
  `context.Background()` (or `t.Context()`).
- [ ] **Step 3:** `simgen.WithContext` adapter:

```go
// WithContext adapts a marketdata.Client to the Fetcher interface, binding
// every Fetch to ctx.
func WithContext(ctx context.Context, c *marketdata.Client) Fetcher {
	return ctxFetcher{ctx: ctx, c: c}
}
```

- [ ] **Step 4:** `cmd/pofo/main.go`: `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM); defer stop()` in main; thread ctx to every client call site (22) and pass `simgen.WithContext(ctx, client)` where the client was used as a Fetcher; portfolio.Build fetch closures bind ctx.
- [ ] **Step 5:** `make check && make golden` PASS; commit `marketdata: context-aware client API`.

### Task 7: cache-less mode (`NewClient("")`)

**Files:**
- Modify: `pkg/marketdata/cache.go`, `pkg/marketdata/client.go` (loadResolution/saveResolution), `pkg/marketdata/fees.go` (disk cache paths), `pkg/marketdata/extended.go` (DefaultCacheDir doc), `pkg/marketdata/doc.go`
- Test: `pkg/marketdata/cache_test.go` (or client_test.go)

- [ ] Steps: failing test `TestCachelessClient` (NewClient("") + stubbed
  server: two Fetches hit the network twice modulo memoization: use two
  clients or clear memo; assert no file appears in CWD and no error); guard
  `c.CacheDir == ""` in loadCacheAnyAge, saveCache, loadResolution's disk
  read, saveResolution, and the fees disk cache load/save; document on
  NewClient; PASS; `make check`; commit `marketdata: cache-less mode with an empty cache directory`.

### Task 8: Series.At + Client.FXRate

**Files:**
- Modify: `pkg/marketdata/types.go` (At), `pkg/marketdata/currency.go` (FXRate), tests in `currency_test.go`/new `types_test.go`, `doc.go`, `example_test.go`

**Interfaces:**
- Produces: `func (s *Series) At(at time.Time) (value float64, on time.Time, ok bool)`;
  `func (c *Client) FXRate(ctx context.Context, from, to string, at time.Time) (float64, error)`.

- [ ] Steps: failing tests (At forward-fills between points, exact on a
  quote day, ok=false before the first point and on empty/nil series; FXRate
  returns 1 for same currency, the forward-filled cross otherwise via a
  stubbed Yahoo, an error when `at` predates the cross); implement:

```go
// At returns the series value in force at the given time: the close of the
// last point dated at or before it (forward-fill). ok is false before the
// first point or on an empty series. Nil-safe.
func (s *Series) At(at time.Time) (float64, time.Time, bool) {
	if s == nil {
		return 0, time.Time{}, false
	}
	i := sort.Search(len(s.Points), func(k int) bool { return s.Points[k].Date.After(at) })
	if i == 0 {
		return 0, time.Time{}, false
	}
	p := s.Points[i-1]
	return p.Close, p.Date, true
}
```

  FXRate: upper-trim both codes, 1.0 when equal, else
  `c.History(ctx, from+to+"=X", time.Time{})` then `At(dayUTC(at))`, error
  naming the pair and date when !ok. PASS; docs; `make check`; commit
  `marketdata: forward-fill Series.At and point-in-time FXRate`.

### Task 9: Client.Search

**Files:**
- Modify: `pkg/marketdata/resolve.go`, `pkg/marketdata/resolve_test.go`, `doc.go`, `example_test.go`

**Interfaces:**
- Produces: `func (c *Client) Search(ctx context.Context, query string) ([]Resolution, error)`:
  catalog-pinned resolution of `CanonicalID(query)` first when it exists,
  then every Yahoo free-text candidate (Source "yahoo", Symbol, Name),
  deduplicated by source|symbol; error only when nothing at all is found.

- [ ] Steps: failing tests (a catalogued alias yields the pinned instrument
  first; a free-text name yields the stubbed Yahoo candidates in order; dupes
  collapse; no-result is an error); implement over `catalogResolution` +
  `c.search(ctx, ...)`; PASS; docs; `make check`; commit `marketdata: free-text Search returning candidate resolutions`.

### Task 10: dividends + raw closes

**Files:**
- Modify: `pkg/marketdata/types.go` (Dividend, Series.Dividends),
  `pkg/marketdata/yahoo.go` (events=div + raw column),
  `pkg/marketdata/cache.go` (dividends column),
  `pkg/marketdata/client.go` (raw threading, `~raw` cache identity),
  `pkg/marketdata/extended.go` (FetchOptions.Raw, SIM guard, Trim),
  `pkg/marketdata/currency.go` (dividend conversion), `doc.go` (conventions),
  tests: `yahoo`-paths in `client_test.go`, `extended_test.go`, `currency_test.go`
- Modify: `CLAUDE.md` (SIM/units conventions: note Raw + Dividends)

**Interfaces:**
- Produces: `type Dividend struct { Date time.Time; Amount float64 }` (ex-date
  00:00 UTC, per share, series' quote currency); `Series.Dividends []Dividend`
  sorted ascending; `FetchOptions.Raw bool`; raw+SIM error from FetchExtended.
- Internal: `history(ctx, symbol, from, raw bool)`; cache identity
  `symbol + "~raw"` for raw views; `cacheFile` gains
  `Dividends []cacheDividend{Date string; Amount float64}`.

- [ ] **Step 1: failing tests**: stubbed Yahoo response with both `close` and
  `adjclose` columns plus a `events.dividends` map:
  default Fetch returns adjclose values and the dividends sorted;
  `FetchExtended{Raw: true}` returns close values (and dividends);
  raw and adjusted views cache under distinct files and do not evict each
  other; `FetchExtended{Raw: true}` on `VOOSIM` errors;
  `ConvertCurrency` scales dividend amounts with the same FX walk as points
  (GBp pence scaling included); `Trim` clips Dividends to the window;
  old-format cache file (no dividends key) still loads for adjusted reads.
- [ ] **Step 2:** FAIL.
- [ ] **Step 3: implement.** fetchYahoo gains `raw bool`: request path adds
  `&events=div`; column choice `raw -> quote.close` else today's
  adjclose-then-close fallback; parse events:

```go
Events struct {
	Dividends map[string]struct {
		Amount float64 `json:"amount"`
		Date   int64   `json:"date"`
	} `json:"dividends"`
} `json:"events"`
```

  into `s.Dividends` (skip amount<=0, dayUTC, sort by date).
  `history`/`History` internals take `raw`; memo and cache keys use
  `viewKey := symbol; if raw { viewKey += "~raw" }` (sanitizeFilename keeps
  `~` out of the alphabet so add `~` to the allowed runes or use "_raw":
  use suffix `"~raw"` and let sanitize map it deterministically).
  Non-Yahoo sources ignore raw (their price is the raw price).
  Public `Fetch` stays adjusted; internal `fetch(ctx, id, from, raw)` threads
  through fetchISIN/fetchTicker/resolveBest/historyForResolution.
  FetchExtended: `if opt.Raw && wantSim { return nil, fmt.Errorf(...) }`;
  raw path calls the internal fetch. convertTo/ConvertCurrency: after the
  points pass, convert each dividend at its ex-date with `fx.At` (first rate
  extrapolated backward like points) and the same pence scale. Trim copies
  and clips Dividends alongside Points.
- [ ] **Step 4:** full `go test ./pkg/marketdata/` PASS.
- [ ] **Step 5:** doc.go section "Dividends and raw closes" (double-count
  warning: adjusted closes already reinvest dividends); CLAUDE.md note;
  `make check && make golden`; commit `marketdata: dividend events and raw close views`.

### Task 11: pofo docs + full validation

**Files:**
- Modify: root `doc.go` (if pipeline text mentions signatures), `README.md` (library snippet signatures), `CLAUDE.md` (core pipeline snippet: ctx)

- [ ] Steps: update every FetchExtended/Fetch snippet to the ctx signatures;
  `make check && make golden && make build && ./pofo -h`; smoke: `./pofo <an
  example portfolio> ` renders; commit `docs: context-aware library snippets`.

### Task 12: finador wiring + market adapter

**Files (in ../finador):**
- Modify: `go.mod` (require github.com/bpineau/pofo v0.0.0 + replace => ../pofo)
- Create: `internal/market/pofo.go`, `internal/market/pofo_test.go`
- Delete: `internal/market/yahoo.go`, `ft.go`, `morningstar.go`, `multi.go` and their `_test.go`
- Modify: `internal/market/source.go` (keep Source/Ref/DailyData/SymbolInfo; Default() returns the pofo adapter), call sites of `market.Default()`/`NewYahoo` in `internal/cli` and `internal/web`

**Interfaces:**
- Produces: `func Default() Source` backed by:

```go
// Pofo adapts the pofo marketdata client to finador's Source. The client is
// cache-less: quote history on disk would reveal the holdings the encrypted
// book protects.
type Pofo struct{ Client *marketdata.Client }

func NewPofo() *Pofo {
	c := marketdata.NewClient("")
	c.HTTP = &http.Client{Timeout: 15 * time.Second}
	return &Pofo{Client: c}
}
```

  `Resolve(ctx, query)` -> `Client.Search` first candidate -> SymbolInfo;
  `Daily(ctx, ref, from)` -> try ISIN then Symbol (each via
  `Client.FetchExtended(ctx, id, marketdata.FetchOptions{From: from.Time(),
  NoSim: true, Raw: true})`), map Points -> domain.PricePoint, Dividends ->
  domain.DividendEvent, Currency; `Intraday(ctx, ref)` -> `Client.Intraday`,
  mapping `marketdata.ErrNotCovered` -> `market.ErrNotCovered` via errors.Is.

- [ ] Steps: `go mod edit -require=github.com/bpineau/pofo@v0.0.0
  -replace=github.com/bpineau/pofo=../pofo && go mod tidy`; failing adapter
  tests against httptest-stubbed pofo bases (`Client.ChartBase = srv.URL`
  etc., mirroring pofo's stubAllBases: cover Daily raw closes + dividends,
  Resolve via search JSON, Intraday, ErrNotCovered for empty ref); implement;
  delete the three providers + multi; fix `refresh.go` imports (Refresh
  logic unchanged: it already speaks Source); `go build ./... && go test
  ./...` PASS; commit in finador `market: fetch through pofo marketdata`.

### Task 13: finador perf delegates to pofo metrics

**Files (in ../finador):**
- Modify: `internal/perf/perf.go` only (Report/periods and all public
  signatures unchanged; bodies delegate)

**Interfaces:**
- Consumes: pofo `metrics.TWR/FlowReturns/Volatility/Sharpe/Sortino/Annualize/IRR/DrawdownEpisodes`.
- Produces: unchanged finador API: `TWR(points, flows) float64`,
  `DailyReturns(points, flows) []float64`, `XIRR(cashflows) (float64, error)`,
  `CAGR(total, days) float64`, `Vol/Sharpe/Sortino`, `MaxDrawdown(points) Drawdown`.

- [ ] Steps: keep `perf_test.go` untouched as the equivalence arbiter; rewrite
  perf.go with `toSeries(points) (dates, values)` and `toFlows` adapters;
  NaN from pofo ratios maps to finador's 0 convention; XIRR splits the final
  cashflow into metrics.IRR's (flows..., finalDate, finalValue) checking
  IRR's sign convention (read pkg/metrics/irr.go first and adapt signs so
  the existing finador tests pass); MaxDrawdown picks the deepest
  DrawdownEpisodes entry and maps RecoverDate/Ongoing to `Recovered *Date`;
  `go test ./internal/perf/` PASS unchanged; commit `perf: delegate the math to pofo metrics`.

### Task 14: finador charts through pofo

**Files (in ../finador):**
- Delete: `internal/chart/` entirely
- Modify: `internal/cli/chart.go` (braille), `internal/cli/couleur.go` (keep
  color constants; add the pie palette moved from chart),
  `internal/web/handlers.go`, `internal/web/scope.go`, `internal/web/assets.go`,
  `internal/web/tree.go`, their tests

**Interfaces:**
- Consumes: pofo `chart.Line(Options{Style: chart.StyleMinimal()}, ...)`,
  `chart.Sparkline`, `chart.Pie(PieOptions{Hole: 0.55, HideLegend: true, Width: 190}, ...)`,
  `chart.Term(TermOptions{Braille: true, ...}, ...)`.
- Produces: a small `internal/web/chart.go` helper
  `func lineSVG(lines []line, w, h int) template.HTML` converting
  `[]perf.Point` to `chart.Series` (Dates/Values) once, so handlers stay tidy.

- [ ] Steps: replace call sites (perf.Point -> Dates/Values conversion;
  intraday points feed the same chart.Line: sub-day spans get HH:MM ticks);
  braille CLI keeps its --width/--height flags mapped to TermOptions;
  the finador PiePalette moves to the web package (template.CSS uses stay);
  update the affected finador tests (render_test, svg expectations move from
  exact finador markup to structural assertions on pofo markup);
  `go build ./... && go test ./...` PASS; delete `internal/chart`; commit
  `chart: render through pofo chart with the minimal style`.

### Task 15: end-to-end validation

- [ ] pofo: `make check && make golden && make demo` all green.
- [ ] finador: `go vet ./... && go test ./... && go build -o bin/finador ./cmd/finador`.
- [ ] Live sanity (network): in a scratch HOME, `finador init`, add one asset
  (CW8.PA), `finador refresh`, `finador chart`, `finador value`, and
  `finador serve` + curl one scope page: quotes, dividends, FX, charts render.
- [ ] pofo: `./pofo -quote CW8.PA` (or equivalent mode) still works.
- [ ] Update `docs/finador-integration-design.md` status note; push pofo
  master; commit finador.
