# Design: latest quote (real-time valuation)

Date: 2026-06-28 (revised 2026-07-02)

Status: IMPLEMENTED (2026-07-02). This design was briefly marked superseded on
the grounds that the finador integration (point-in-time `Series.At` + `Raw`
closes, refreshed periodically) already yields a live-ish valuation: Yahoo's
daily chart endpoint includes today's in-progress candle, whose close is the
current price. That is true for the raw number, but it rests on an implicit
behavior of the daily endpoint, carries no freshness metadata (a UI cannot
tell "live at 15:42" from "yesterday's close"), and offers no lightweight
one-call spot for a library consumer. The design was therefore revived and
implemented, revised for the API as it stands today (context.Context
throughout, yahooGet host fallback, SplitSim convention).

## Context

pofo can fetch a daily history (Fetch), today's intraday path (Intraday), and
resolve an identifier (Resolve). What it lacks is a one-call answer to "what is
the most recent price of this instrument right now", which any valuation UI
needs to show a live portfolio value. finador, and many other tools, want this.

This design adds that capability to pkg/marketdata as a small, neutral value
type and one method, reusing the existing multi-source resolution, on-disk cache
and stale fallback. finador consumption is sketched at the end but implemented
in the finador repository, not here.

## Goal

Return the freshest available price per instrument, degrading gracefully:
- a Yahoo-quoted instrument yields its live market price;
- any other instrument (FT or Morningstar fund, where the last NAV close IS the
  latest price) yields its last daily close, from cache when fresh and stale
  data on a failed refresh.

Latest always yields a Quote for any identifier Fetch can handle.

## Non-goals

- No internal caching of the live price: like Intraday, the live fetch is
  stateless and the caller owns any short-TTL cache for repeated valuations.
- No batch multi-symbol endpoint. Yahoo exposes /v7/finance/quote?symbols=...
  which could value a whole portfolio in one request, but it sits behind a
  cookie/crumb dance the chart endpoint does not need; it is noted as possible
  future work, not built here. Single-asset Latest plus caller-side caching is
  enough for finador (under 100 assets, human-paced).
- No currency conversion. Quote carries the quote currency; converting to a
  display currency stays the caller's job (Client.ConvertCurrency).

## Design qualities to preserve

The capability enters pofo as a general, neutral type in pofo's own idiom
(analytics float64, time.Time, self-contained, no third-party dependency).
pkg/marketdata must not import pkg/chart. The library stays debrayable: a caller
can drive Fetch, Intraday and Latest directly and layer its own policy on top.

## 1. The Quote type and the Latest method

New file pkg/marketdata/latest.go.

```go
// Quote is the most recent known price of an instrument.
//
// Live reports how fresh the price is: true means a real-time market field
// (Yahoo regularMarketPrice), whose Time is an intraday instant; false means
// the last daily close, whose Time is that close's date. A market that is
// closed still yields a Live quote, the regular session's last price, with its
// Time at the close.
type Quote struct {
    Price    float64   // in Currency
    Time     time.Time // when this price was observed
    Currency string    // ISO 4217 quote currency
    Source   string    // "yahoo", "ft", "morningstar" or "stooq"
    Live     bool      // true: real-time market field; false: last daily close
}

// Latest returns the freshest available price for an identifier: the live Yahoo
// market price when the instrument is Yahoo-quoted, otherwise the last daily
// close (FT or Morningstar NAV), served from the on-disk cache when fresh and
// from stale data on a failed refresh.
//
// Like Intraday, the live path is stateless: Latest performs no caching of the
// live price, so a caller valuing a portfolio repeatedly should keep its own
// short-TTL cache. The daily-close fallback path uses the existing on-disk
// daily cache.
func (c *Client) Latest(ctx context.Context, id string) (*Quote, error)
```

SIM convention: Latest strips the "SIM" suffix first (SplitSim, the same
unconditional split FetchExtended performs), because a simulated history
extension never changes the current price: the latest price of VOOSIM is the
latest price of VOO.

Adjusted closes are fine here: Yahoo's adjustment factor is 1 at the most
recent bar, so the last adjusted close equals the last raw close, and a fund
NAV has no adjustment at all. The fallback can therefore ride the default
(adjusted) Fetch and cache without a Raw variant, and the Quote is still the
right number for a market-price valuation.

## 2. Fallback logic

```
Latest(ctx, id):
  0. base, _ := SplitSim(id)            (simulated history is irrelevant here)
  1. yahooSymbol(ctx, base) covered (a ticker, or an ISIN whose cached/catalog
     resolution points at Yahoo)?
        yes: fetchYahooSpot(ctx, symbol)
               regularMarketPrice present  -> Quote{Live: true}
               absent or request failed    -> fall through to step 2
        no: fall through
  2. s, err := Fetch(ctx, base, latestFrom())  (full multi-source chain + cache)
        err != nil               -> return err
        s.Last() is the zero Point (empty series) -> return an error
        return &Quote{
            Price:    s.Last().Close,
            Time:     s.Last().Date,
            Currency: s.Currency,
            Source:   s.Source,
            Live:     false,
        }
```

where latestFrom returns time.Now().AddDate(-1, 0, 0): a one-year window is
deep enough to guarantee at least one recent close even across long market
closures or for illiquid instruments, while keeping the payload small. The
disk cache accepts any cached window starting at or before the requested from,
and the memoization key formats from at day granularity, so repeated Latest
calls within a session and across a day reuse one cache entry.

Notes:
- Fetch signature is Fetch(ctx context.Context, id string, from time.Time)
  (*Series, error); Series exposes Last() Point (Point has Date and Close),
  Currency and Source.
- Step 1 reuses the existing cheap yahooSymbol helper (no resolution network
  call). A fund quoted only by FT or Morningstar is not Yahoo-covered (or its
  spot request fails and falls through), so it reaches step 2, where its last
  NAV close is exactly its latest price.
- Step 2 inherits the whole Fetch behavior: multi-source resolution for unseen
  identifiers, the on-disk cache, and the stale fallback (offline yields the
  last known value rather than an error).
- A Yahoo asset whose spot request fails (network, throttling, missing field)
  still gets a Quote via step 2, marked Live: false.

## 3. Yahoo spot fetcher

New unexported method in pkg/marketdata/yahoo.go:

```go
func (c *Client) fetchYahooSpot(ctx context.Context, symbol string) (*Quote, error)
```

- GET {ChartBase}/v8/finance/chart/{symbol}?interval=1d&range=1d via the
  existing c.yahooGet (query1/query2 host fallback, retry and 429 back-off in
  place). The chart meta carries regularMarketPrice, regularMarketTime,
  currency and exchangeTimezoneName regardless of the range, so the lightest
  valid request is used.
- Parse meta.regularMarketPrice (the live price), meta.regularMarketTime
  (Unix seconds, rendered in the exchange time zone via exchangeTimezoneName,
  falling back to UTC; the blank time/tzdata import already in the package
  keeps LoadLocation host-independent), meta.currency.
- Return ErrNotCovered when the result is empty or regularMarketPrice is absent
  or non-positive, so Latest falls through to the daily-close path.

## 4. Documentation pass (English, no em-dashes)

- Godoc on latest.go (Quote and Latest), as written above.
- pkg/marketdata/doc.go: add a "Latest quote" section after "Intraday" and a
  Toolbox bullet beside Intraday, Resolve and ConvertCurrency, stating the
  live/caller-owns-the-cache contract and the degrade-to-close behavior.
- Root doc.go: extend the pkg/marketdata bullet to mention the latest
  (real-time) quote alongside daily and intraday prices.
- README.md: a short "Latest quote" subsection under Data with a snippet
  (Latest plus a valuation line), and a mention in the "Using it as a library"
  section. All new prose English, no em-dashes.

## 5. Testing

- The existing chartJSON test fixture (client_test.go) carries no
  regularMarketPrice, so the spot tests need a small new fixture helper that
  emits meta.regularMarketPrice, regularMarketTime, currency and
  exchangeTimezoneName.
- fetchYahooSpot parses regularMarketPrice, regularMarketTime (exchange zone)
  and currency from a chart-meta fixture; an absent regularMarketPrice yields
  ErrNotCovered.
- Latest returns a Live quote for a Yahoo-covered ticker (httptest stub).
- Latest falls through to the daily close (Live: false) when the spot field is
  absent, and for a fund-only identifier, returning the last Series point.
- Latest strips the SIM suffix (a "<id>SIM" identifier quotes as its base).
- The offline/stale path (a Quote served from the stale daily cache) is
  already covered by Fetch's own stale-fallback tests; not duplicated.
- go test ./... and go vet ./... stay green; gofmt-clean; no em-dashes added.

## 6. finador consumption (implemented in the finador repository)

finador's valuation currently reads the last stored daily close
(Prices[id].At(date)); freshness comes from AutoRefresh periodically
re-fetching daily series, whose last candle carries Yahoo's live price. Latest
complements that with explicit freshness:

- market.Source gains Latest(ctx, ref) mapped onto pofo's Client.Latest
  (ISIN tried first, then symbol, like Daily).
- The web server keeps a short-TTL in-memory quote cache per asset (the same
  pattern as its intraday cache); pofo stays cache-less there (privacy).
- Valuation-time overlay: for today's valuations, a live quote overrides the
  stored close via a dedicated ValueOption, so page loads show
  live-to-the-minute values between refreshes.
- Quote.Live and Quote.Time let the UI show whether a value is real-time or an
  end-of-day close, and from when.
