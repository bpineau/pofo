# Design: intraday support and library consumption

Date: 2026-06-28

## Context

pofo is a dependency-free, library-first Go toolkit for tracking and designing
stock portfolios. A separate private application, finador (a ledger app with an
encrypted store, a web UI and a CLI), currently carries its own market-data
layer (`internal/market`: Yahoo, FT, Morningstar, a Multi chain and a search
based resolver) and its own minimal asset metadata. finador wants to eventually
consume pofo instead, to mutualize effort and gain geography and sector
breakdowns, the embedded catalog and intraday charts.

This design improves pofo so that it can serve that consumption cleanly, without
compromising its consistency, clarity or extensibility. The immediate concrete
feature is intraday price data, which pofo lacks today and finador needs. The
rest of the work prepares pofo to be a comfortable foundation for an external
consumer.

finador is NOT modified by this work. Section "finador migration (planned)"
records the consumer-side consequences so a later migration is smooth, but no
finador code changes here.

## Goals

1. Add intraday (current-day, 5-minute) price data to `pkg/marketdata`, in
   pofo's own neutral types, with a stateless fetch whose caching policy stays
   with the caller.
2. Render intraday series through the existing `chart.Line`, by teaching the
   time axis to label sub-day spans, rather than adding a parallel renderer.
3. Make the bundled catalog ergonomic for external consumers with a
   `datasets.Lookup` by identifier.
4. Expose pofo's multi-source resolution as a public `marketdata.Resolve`, so a
   consumer can reuse it instead of writing its own.
5. Improve the documentation (godoc and README) for this consumption story.
   Documentation is English only and uses no em-dashes.

## Non-goals

- No change to finador in this work.
- No on-disk or in-memory caching of intraday data inside pofo: intraday is
  ephemeral and the caller owns the cache.
- No separate dividend-event model. pofo stays on adjusted-close series.
- No Monte-Carlo or FIRE projection engine. The design only avoids closing the
  door on a future `pkg/project` built on top of `metrics` and `optimize`.

## Design qualities to preserve

New capabilities enter pofo as general, neutral types in pofo's own idiom
(analytics float64, time.Time, self-contained, no third-party dependency). A
consumer's domain concerns (decimal money, encrypted storage, dividend events,
a USD-pivot FX convention) never leak into pofo: the consumer owns the adapter.
The library stays débrayable, meaning a caller can bypass the high-level helpers
and drive the lower layers directly.

## 1. Intraday in `pkg/marketdata`

New file `pkg/marketdata/intraday.go`.

```go
// IntradayPoint is one intraday observation, typically a 5-minute tick.
type IntradayPoint struct {
    Time  time.Time // exact instant, in the exchange's local time zone
    Close float64
}

// IntradaySeries is the current trading day's price path of one instrument,
// sorted by ascending time. Unlike Series it is ephemeral: it covers only
// today and is never written to the on-disk cache.
type IntradaySeries struct {
    Symbol   string
    Name     string
    Currency string
    Source   string // "yahoo"
    Points   []IntradayPoint
}

func (s *IntradaySeries) First() IntradayPoint
func (s *IntradaySeries) Last()  IntradayPoint
```

Method on `Client`:

```go
// Intraday returns today's intraday price path (5-minute resolution) for an
// identifier, fetched live from Yahoo Finance.
//
// Unlike Fetch, Intraday never touches the on-disk cache: an intraday series is
// valid only for today and goes stale within minutes. Callers that view an
// asset repeatedly should keep their own short-TTL cache; the fetch is
// deliberately stateless so that policy stays with the caller.
//
// Only Yahoo-quoted instruments have intraday data. An identifier that resolves
// to a fund-only source (Financial Times, Morningstar), or that has no known
// Yahoo symbol, returns ErrNotCovered. Intraday does not perform a network
// resolution: it reuses the symbol Fetch already learned (bundled catalog plus
// on-disk resolution cache). For an unseen ISIN, call Fetch first.
func (c *Client) Intraday(id string) (*IntradaySeries, error)
```

Behavior:

- `canonical := CanonicalID(id)`.
- Map the identifier to a Yahoo symbol cheaply, with no resolution network I/O,
  through a small helper `yahooSymbol(id) (symbol string, ok bool)`:
  - a plain ticker maps to itself (upper-cased);
  - an ISIN maps through `catalogResolution` then `loadResolution`; it is
    covered only when the cached resolution has `Source == "yahoo"` and a
    non-empty `Symbol`.
  - otherwise `ok` is false.
- When not covered, return `ErrNotCovered`.
- Otherwise call `fetchYahooIntraday(symbol)`.

New exported sentinel error:

```go
// ErrNotCovered reports that a request cannot be served for an identifier,
// for example intraday data for an instrument quoted only by a fund source.
var ErrNotCovered = errors.New("not covered")
```

New unexported fetcher in `pkg/marketdata/yahoo.go`:

```go
func (c *Client) fetchYahooIntraday(symbol string) (*IntradaySeries, error)
```

- `GET {ChartBase}/v8/finance/chart/{symbol}?interval=5m&range=1d`, via the
  existing `c.get` (retry and 429 back-off already in place).
- Parse `result[0].meta.currency`, `result[0].meta.exchangeTimezoneName`,
  `timestamp[]` and `indicators.quote[0].close[]`. Skip nil closes. Timestamp
  each point in the exchange time zone; fall back to UTC if the zone is unknown.
- An empty result returns `ErrNotCovered`.

Time-zone data: pofo must resolve exchange zones without depending on the host
OS. Add a blank import `_ "time/tzdata"` in `intraday.go` (the same approach
finador uses). This embeds the zone database; it is standard library, so the
dependency-free property is preserved.

## 2. Chart: intraday-aware time axis

`pkg/chart/svg.go`. `chart.Line` already renders an arbitrary
`Series{Dates []time.Time, Values []float64}` and decimates long series. Only
the axis labelling assumes year or month spans. Teach `timeTicks` to detect a
sub-day span:

- If `to.Sub(from) <= 36 * time.Hour`, return evenly spaced ticks (about 5 to 6)
  labelled with `t.Format("15:04")`.
- Otherwise keep the current year and month logic unchanged.

Result: a single `chart.Line` entry point renders both daily and intraday. No
`IntradaySVG`. `marketdata` does not import `chart`; the caller maps an
`IntradaySeries` to a `chart.Series`:

```go
ser := chart.Series{Name: s.Name}
for _, p := range s.Points {
    ser.Dates = append(ser.Dates, p.Time)
    ser.Values = append(ser.Values, p.Close)
}
svg := chart.Line(chart.Options{Title: s.Name}, []chart.Series{ser})
```

This snippet appears in a chart example test and in the README.

## 3. Catalog ergonomics: `datasets.Lookup`

`pkg/datasets/catalog.go`. The catalog is a slice today; consumers must build
their own index. Add a by-identifier lookup backed by a lazily built index:

```go
// Lookup returns the catalog asset for an identifier (its id, ISIN or any
// alias, case-insensitive) and whether it was found. The index is built once
// on first use.
func Lookup(id string) (Asset, bool)
```

- A package-level `sync.Once` builds `map[string]Asset` keyed by upper-cased
  `ID`, `ISIN` and each alias.
- This is the entry point an external consumer uses to pull `Geography`,
  `Sectors`, `Factors` and `Exposures` for a held asset.

## 4. Public resolution: `marketdata.Resolve`

`pkg/marketdata`. Expose the instrument identity that `Fetch` settles on, so a
consumer can reuse pofo's multi-source resolution instead of writing its own.

```go
// Resolution is how pofo maps a user identifier to a quotable instrument:
// which source serves it, under which symbol, plus the resolved name and quote
// currency. Currency may be empty when a source does not report it.
type Resolution struct {
    Source   string // "yahoo", "stooq", "ft" or "morningstar"
    Symbol   string // Yahoo or Stooq symbol, or Morningstar id; empty for ft
    Xid      string // FT internal id; empty otherwise
    Name     string
    Currency string
}

// Resolve returns the instrument pofo would quote for a user identifier
// (ticker, ISIN or alias). It uses the bundled catalog and the on-disk
// resolution cache first, then the same multi-source search Fetch uses. It may
// perform network I/O and caches the result, so a later Fetch of the same id
// reuses this work.
func (c *Client) Resolve(id string) (Resolution, error)
```

- Built on the existing internal `resolution` flow: return the cached or catalog
  resolution when present; otherwise run the existing `resolveBest`, adopt the
  winner (as `Fetch` does), and return its identity. Populate `Currency` from
  the resolved series when available.
- The internal `resolution` struct stays unexported; `Resolution` is the public
  mirror. A small mapping function converts between them.

`Resolve` and `Intraday` deliberately differ: `Intraday` is cheap and never runs
a network resolution, `Resolve` is the complete path and may. This is documented
on both.

## 5. Documentation pass (English, no em-dashes)

- New godoc on `intraday.go`.
- `marketdata/doc.go`: describe daily vs intraday, the "caller owns the intraday
  cache" contract, the FX story (`ConvertCurrency` reprices a whole series via
  Yahoo crosses), and `Resolve` as the reusable resolution entry point.
- Root `doc.go`: add intraday to the `pkg/marketdata` description.
- `chart/doc.go`: note that `Line` labels sub-day spans with clock times.
- `datasets/doc.go`: present `Lookup` as the metadata entry point for external
  consumers.
- `README.md`: add an Intraday subsection, clarify FX conversion, and add a
  section "Using pofo as a library from another application" that tells the
  débrayable story: catalog as metadata (`datasets.Lookup`), `marketdata` for
  quotes, intraday and resolution, `chart` for rendering, and how to map pofo
  types onto your own domain.

### Em-dash purge (separate commit)

pofo currently contains 244 em-dashes across 40 files (godoc and README). After
the feature lands, a dedicated follow-up commit replaces every em-dash in `*.go`
comments and `*.md` with commas, colons or parentheses, reviewed so meaning is
preserved. It is kept separate so the feature diff stays readable.

## 6. finador migration (planned, not executed here)

Recorded so the later consumer-side migration is smooth. None of this is built
now.

- Drop `DividendEvent` and `MarketData.Dividends` from finador; rely on pofo's
  adjusted-close series. Consequence to confirm: any finador feature reading
  `Dividends` (dividend income, withholding tax) loses its source. Ben has
  judged this acceptable; confirm before removal.
- Replace finador's USD-pivot `MarketData.FX` with on-demand
  `marketdata.ConvertCurrency`, behind a thin cached adapter if needed. This is
  a semantic change: FX becomes a repriced series rather than a USD pivot.
- Replace `internal/market` (Yahoo, FT, Morningstar, Multi, resolver) with
  `marketdata.Client` plus `marketdata.Resolve`, behind an adapter mapping
  `*marketdata.Series` to `domain.PriceSeries` and `*marketdata.IntradaySeries`
  to finador's intraday points. finador keeps its `domain`, encrypted store, web
  server and the short-TTL in-memory intraday cache with offline fallback
  (application policy, exactly the lazy-mode brick already specced).
- Pull geography, sectors and factors from `datasets.Lookup` to render
  breakdowns finador does not have today.

## 7. Testing

- `marketdata`: a `fetchYahooIntraday` fixture test (timestamps, gaps, time
  zone, currency); `Intraday` returns `ErrNotCovered` for a fund-only or unknown
  identifier and parses a covered one; `Resolve` returns a cached resolution
  without network and maps identity correctly.
- `chart`: `Line` over a sub-day series renders a polyline and `15:04` axis
  labels; the daily path is unchanged (golden or substring assertions).
- `datasets`: `Lookup` finds by id, ISIN and alias, case-insensitively, and
  reports false for an unknown identifier.
- `go test ./...` and `go vet ./...` stay green.

## 8. Future note: Monte-Carlo / FIRE

Out of scope here, but kept open: a future `pkg/project` (portfolio projections,
glidepaths, withdrawal rules, Monte-Carlo stress tests) would sit on top of
`metrics` and `optimize`, consuming return series, without touching
`marketdata` or `chart`. The intraday types stay isolated from the daily-return
pipeline, so they do not constrain that future work.
