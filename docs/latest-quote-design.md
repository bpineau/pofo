# Design: latest quote (real-time valuation)

Date: 2026-06-28

## Context

pofo can fetch a daily history (Fetch), today's intraday path (Intraday), and
resolve an identifier (Resolve). What it lacks is a one-call answer to "what is
the most recent price of this instrument right now", which any valuation UI
needs to show a live portfolio value. finador, and many other tools, want this.

This design adds that capability to pkg/marketdata as a small, neutral value
type and one method, reusing the existing multi-source resolution, on-disk cache
and stale fallback. It does not modify finador.

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
  which could value a whole portfolio in one request; it is noted as possible
  future work, not built here. Single-asset Latest plus caller-side caching is
  enough for finador (under 100 assets, human-paced).
- No currency conversion. Quote carries the quote currency; converting to a
  display currency stays the caller's job (Client.ConvertCurrency).
- No change to finador.

## Design qualities to preserve

The capability enters pofo as a general, neutral type in pofo's own idiom
(analytics float64, time.Time, self-contained, no third-party dependency).
pkg/marketdata must not import pkg/chart. The library stays débrayable: a caller
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
func (c *Client) Latest(id string) (*Quote, error)
```

## 2. Fallback logic

```
Latest(id):
  1. yahooSymbol(id) covered (a ticker, or an ISIN whose cached/catalog
     resolution points at Yahoo)?
        yes: fetchYahooSpot(symbol)
               regularMarketPrice present  -> Quote{Live: true}
               absent or request failed    -> fall through to step 2
        no: fall through
  2. s, err := Fetch(id, latestFrom())       (full multi-source chain + cache)
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
closures or for illiquid instruments, while keeping the payload small. Fetch is
keyed and cached by this from, so repeated Latest calls reuse one cache entry.

Notes:
- Fetch signature is Fetch(id string, from time.Time) (*Series, error);
  Series exposes Last() Point (Point has Date and Close), Currency and Source.
- Step 1 reuses the existing cheap yahooSymbol helper (no resolution network
  call). A fund quoted only by FT or Morningstar is not Yahoo-covered, so it
  goes straight to step 2, where its last NAV close is exactly its latest price.
- Step 2 inherits the whole Fetch behavior: multi-source resolution for unseen
  identifiers, the on-disk cache, and the stale fallback (offline yields the
  last known value rather than an error).
- A Yahoo asset whose spot request fails (network, throttling, missing field)
  still gets a Quote via step 2, marked Live: false.

## 3. Yahoo spot fetcher

New unexported method in pkg/marketdata/yahoo.go (or latest.go):

```go
func (c *Client) fetchYahooSpot(symbol string) (*Quote, error)
```

- GET {ChartBase}/v8/finance/chart/{symbol}?interval=1d&range=1d via the
  existing c.get (retry and 429 back-off in place). The chart meta carries
  regularMarketPrice, regularMarketTime, currency and exchangeTimezoneName
  regardless of the range, so the lightest valid request is used.
- Parse meta.regularMarketPrice (the live price), meta.regularMarketTime
  (Unix seconds, rendered in the exchange time zone via exchangeTimezoneName,
  falling back to UTC), meta.currency.
- Return ErrNotCovered when the result is empty or regularMarketPrice is absent
  or non-positive, so Latest falls through to the daily-close path.

## 4. Documentation pass (English, no em-dashes)

- Godoc on latest.go (Quote and Latest), as written above.
- pkg/marketdata/doc.go: add Latest to the package overview and the Toolbox
  section, beside Intraday, Resolve and ConvertCurrency, stating the
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
- Latest serves a Quote offline via the stale daily cache (no error).
- go test ./... and go vet ./... stay green; gofmt-clean; no em-dashes added.

## 6. finador consumption (planned, not built here)

When finador migrates: call Latest(id) per holding for the live valuation,
convert the Quote into the display currency with ConvertCurrency or a spot FX,
and wrap Latest in finador's short-TTL in-memory cache (the same lazy pattern as
intraday). Quote.Live and Quote.Time let the UI show whether a holding's value
is real-time or an end-of-day close.
