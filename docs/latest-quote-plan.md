# Latest Quote Implementation Plan

> **Status: NOT IMPLEMENTED (superseded).** This plan was never executed: there
> is no `pkg/marketdata/latest.go`, no `Quote` type and no `Client.Latest`
> method on master (only the design and this plan were committed, `ae5378f` /
> `992e04a`). The real-time valuation need it targeted is now met by the finador
> integration (point-in-time `Series.At` + unadjusted `Raw` closes), so a
> dedicated live-spot `Latest` was not built. Kept for reference; revive this
> plan only if a true intraday `regularMarketPrice` path is actually wanted.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a one-call "most recent price" capability to pkg/marketdata for real-time portfolio valuation.

**Architecture:** A small neutral value type Quote and a method Client.Latest(id) that returns the freshest available price per instrument: the live Yahoo regular-market price when the instrument is Yahoo-quoted, otherwise the last daily close (FT or Morningstar NAV), reusing the existing multi-source resolution, on-disk cache and stale fallback. marketdata does not import chart.

**Tech Stack:** Go 1.26 standard library only. Tests use net/http/httptest and the existing newTestClient / chartJSON helpers in pkg/marketdata.

## Context for a fresh session

This plan implements the spec at `docs/latest-quote-design.md` (read it first).
It builds on capabilities already shipped to pkg/marketdata in prior work
(intraday support and library consumption, see
`docs/intraday-and-library-consumption-design.md`). The following already exist
and are used by this plan, no need to create them:

- `func (c *Client) yahooSymbol(id string) (string, bool)` in
  `pkg/marketdata/intraday.go`: maps an identifier to a Yahoo symbol with NO
  resolution network call (a ticker maps to itself; an ISIN is covered only
  when its cached or catalog resolution already points at Yahoo).
- `var ErrNotCovered = errors.New("not covered")` in
  `pkg/marketdata/intraday.go`.
- `func (c *Client) Fetch(id string, from time.Time) (*Series, error)` in
  `pkg/marketdata/client.go`. `Series` (in `pkg/marketdata/types.go`) has
  fields `Currency string`, `Source string`, and a method `Last() Point`;
  `Point` has `Date time.Time` and `Close float64`. `Last()` returns the zero
  Point for an empty series.
- A blank import `_ "time/tzdata"` already exists (in `intraday.go`), so
  `time.LoadLocation` resolves exchange zones across the whole package,
  tests included.
- Test helpers in `pkg/marketdata/client_test.go` (package marketdata,
  white-box): `newTestClient(t, dir, mux) (*Client, *httptest.Server)` (points
  every source base at the stub server so no test reaches the real API),
  `chartJSON(symbol string, days []time.Time, closes []float64) string` (a
  Yahoo daily fixture with meta.currency/symbol/longName, timestamp, quote and
  adjclose, but NO regularMarketPrice), and `testDays(n int) []time.Time`.
- `c.retryDelay` is already set to 1ms by newTestClient, so failing requests do
  not slow tests.

## Global Constraints

- Module is `github.com/bpineau/pofo`, Go 1.26, dependency-free (standard library only). `time/tzdata` is standard library and allowed.
- All documentation (godoc comments and README) is English only and uses no em-dashes. Use commas, colons or parentheses instead.
- `pkg/marketdata` must NOT import `pkg/chart`.
- pofo never caches the live price inside Latest: the live path is stateless and the caller owns any short-TTL cache. The daily-close fallback uses the existing on-disk daily cache.
- Daily series stay on adjusted close. No dividend-event model is added.
- `go test ./...` and `go vet ./...` stay green after every task; new/edited Go is gofmt-clean.
- Do NOT modify the `finador` project anywhere in this plan.

---

## File Structure

- `pkg/marketdata/latest.go` (new): `Quote`, `latestFrom`, `Client.Latest`.
- `pkg/marketdata/yahoo.go` (modify): add `fetchYahooSpot`.
- `pkg/marketdata/latest_test.go` (new): spot parse, live path, and fallback tests, plus the `spotJSON` fixture helper.
- `pkg/marketdata/doc.go`, root `doc.go`, `README.md` (modify): documentation pass.

---

### Task 1: Quote type, Yahoo spot fetch, and Latest

**Files:**
- Create: `pkg/marketdata/latest.go`
- Modify: `pkg/marketdata/yahoo.go`
- Test: `pkg/marketdata/latest_test.go`

**Interfaces:**
- Consumes: `Client.get`, `Client.yahooSymbol`, `Client.Fetch`, `Series`/`Point`, `ErrNotCovered`, `url.PathEscape`.
- Produces:
  - `type Quote struct { Price float64; Time time.Time; Currency string; Source string; Live bool }`
  - `func (c *Client) Latest(id string) (*Quote, error)`
  - `func (c *Client) fetchYahooSpot(symbol string) (*Quote, error)`
  - `func latestFrom() time.Time`

- [ ] **Step 1: Write the failing tests**

Create `pkg/marketdata/latest_test.go`:

```go
package marketdata

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// spotJSON builds a Yahoo chart-meta fixture carrying a live regular-market
// price. A zero price omits the field, to exercise the not-covered path.
func spotJSON(currency, tz string, price float64, at time.Time) string {
	priceField := ""
	if price > 0 {
		priceField = fmt.Sprintf(`"regularMarketPrice":%v,`, price)
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":%q,"exchangeTimezoneName":%q,%s"regularMarketTime":%d}}],"error":null}}`,
		currency, tz, priceField, at.Unix())
}

func TestLatestLiveSpot(t *testing.T) {
	at := time.Date(2024, 3, 1, 18, 0, 0, 0, time.UTC) // 13:00 New York
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", 501.25, at))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest("VOO")
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Price != 501.25 || q.Currency != "USD" || q.Source != "yahoo" {
		t.Fatalf("quote misread: %+v", q)
	}
	if h := q.Time.Hour(); h != 13 {
		t.Errorf("quote hour = %d, want 13 (exchange local time)", h)
	}
}

func TestLatestFallsBackToClose(t *testing.T) {
	days := testDays(3)
	closes := []float64{10, 11, 12.5}
	mux := http.NewServeMux()
	// The chart fixture carries a daily series but no regularMarketPrice, so the
	// spot step is not covered and Latest falls through to the last close.
	mux.HandleFunc("/v8/finance/chart/SPY", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("SPY", days, closes))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest("SPY")
	if err != nil {
		t.Fatal(err)
	}
	if q.Live {
		t.Errorf("expected a non-live (daily close) quote, got Live=true")
	}
	if q.Price != 12.5 || q.Source != "yahoo" {
		t.Fatalf("quote misread: %+v", q)
	}
	if !q.Time.Equal(days[2]) {
		t.Errorf("quote time = %v, want last close date %v", q.Time, days[2])
	}
}

func TestFetchYahooSpotMissingPriceNotCovered(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/NOPX", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", 0, time.Now())) // price omitted
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	_, err := c.fetchYahooSpot("NOPX")
	if !errors.Is(err, ErrNotCovered) {
		t.Fatalf("err = %v, want ErrNotCovered", err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run 'TestLatest|TestFetchYahooSpot' -v`
Expected: FAIL to compile (`undefined: Client.Latest`, `undefined: Client.fetchYahooSpot`).

- [ ] **Step 3: Create the Quote type, latestFrom and Latest**

Create `pkg/marketdata/latest.go`:

```go
package marketdata

import (
	"fmt"
	"time"
)

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

// latestFrom is the history window Latest fetches over when it falls back to
// the last daily close. One year is deep enough to always contain a recent
// close, even across long market closures or for an illiquid instrument, and
// Fetch caches by this from so repeated Latest calls reuse one cache entry.
func latestFrom() time.Time { return time.Now().AddDate(-1, 0, 0) }

// Latest returns the freshest available price for an identifier: the live Yahoo
// market price when the instrument is Yahoo-quoted, otherwise the last daily
// close (FT or Morningstar NAV), served from the on-disk cache when fresh and
// from stale data on a failed refresh.
//
// Like Intraday, the live path is stateless: Latest performs no caching of the
// live price, so a caller valuing a portfolio repeatedly should keep its own
// short-TTL cache. The daily-close fallback path uses the existing on-disk
// daily cache.
func (c *Client) Latest(id string) (*Quote, error) {
	if symbol, ok := c.yahooSymbol(id); ok {
		if q, err := c.fetchYahooSpot(symbol); err == nil {
			return q, nil
		}
		// Spot unavailable (not covered, throttled, or missing field): fall
		// through to the last daily close.
	}
	s, err := c.Fetch(id, latestFrom())
	if err != nil {
		return nil, err
	}
	last := s.Last()
	if last.Date.IsZero() {
		return nil, fmt.Errorf("%s: no recent quote", id)
	}
	return &Quote{
		Price:    last.Close,
		Time:     last.Date,
		Currency: s.Currency,
		Source:   s.Source,
		Live:     false,
	}, nil
}
```

- [ ] **Step 4: Add the Yahoo spot fetcher**

Add to `pkg/marketdata/yahoo.go` (after `fetchYahooIntraday`):

```go
// fetchYahooSpot reads the live regular-market price from the Yahoo chart meta.
// It returns ErrNotCovered when Yahoo serves no usable price for the symbol, so
// Latest can fall back to the last daily close.
func (c *Client) fetchYahooSpot(symbol string) (*Quote, error) {
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=1d&range=1d",
		c.ChartBase, url.PathEscape(symbol))
	body, err := c.get(u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency             string   `json:"currency"`
					ExchangeTimezoneName string   `json:"exchangeTimezoneName"`
					RegularMarketPrice   *float64 `json:"regularMarketPrice"`
					RegularMarketTime    int64    `json:"regularMarketTime"`
				} `json:"meta"`
			} `json:"result"`
			Error *struct {
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable yahoo spot response: %w", err)
	}
	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo spot: %s", resp.Chart.Error.Description)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("%s: %w", symbol, ErrNotCovered)
	}
	m := resp.Chart.Result[0].Meta
	if m.RegularMarketPrice == nil || *m.RegularMarketPrice <= 0 {
		return nil, fmt.Errorf("%s: %w", symbol, ErrNotCovered)
	}
	loc, err := time.LoadLocation(m.ExchangeTimezoneName)
	if err != nil {
		loc = time.UTC
	}
	return &Quote{
		Price:    *m.RegularMarketPrice,
		Time:     time.Unix(m.RegularMarketTime, 0).In(loc),
		Currency: m.Currency,
		Source:   "yahoo",
		Live:     true,
	}, nil
}
```

Note: `yahoo.go` already imports `encoding/json`, `fmt`, `net/url`, `sort`, `time`, so no import change is needed there. `latest.go` imports only `fmt` and `time`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run 'TestLatest|TestFetchYahooSpot' -v && go vet ./pkg/marketdata/ && gofmt -l pkg/marketdata/latest.go pkg/marketdata/yahoo.go pkg/marketdata/latest_test.go`
Expected: PASS; vet clean; gofmt prints nothing.

Note on the offline/stale path: the spec lists "Latest serves a Quote offline via the stale daily cache". Latest's fallback simply delegates to Fetch, whose stale-cache behavior is already covered by TestHistoryStaleCacheFallback in client_test.go. Do not add a redundant test for it.

- [ ] **Step 6: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/latest.go pkg/marketdata/yahoo.go pkg/marketdata/latest_test.go
git commit -m "marketdata: add Latest quote (live Yahoo spot, daily-close fallback)

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Documentation pass (English, no em-dashes)

**Files:**
- Modify: `pkg/marketdata/doc.go`, root `doc.go`, `README.md`

**Interfaces:**
- Consumes: `Quote`, `Client.Latest` from Task 1.
- Produces: documentation only, no code behavior change.

- [ ] **Step 1: Update pkg/marketdata/doc.go**

Add a "Latest quote" section after the "Intraday" section, and a Toolbox bullet. Open `pkg/marketdata/doc.go` and insert, right before the `// # Simulated data` line, this block:

```go
// # Latest quote
//
// Client.Latest returns the most recent price of an instrument as a Quote: the
// live Yahoo regular-market price (Quote.Live true) when the instrument is
// Yahoo-quoted, otherwise the last daily close (Quote.Live false), which for an
// FT or Morningstar fund is its latest NAV. Like Intraday the live path is
// stateless, so a caller valuing a portfolio repeatedly keeps its own
// short-TTL cache; the fallback reuses the on-disk daily cache and its stale
// fallback, so Latest still answers offline.
//
```

Then, in the same file's Toolbox list (the bullet block that already lists
Client.ConvertCurrency and Client.Resolve), add one bullet after the Resolve
bullet:

```go
//   - Client.Latest returns the freshest known price (a Quote) for an
//     identifier, the live Yahoo market price when available, otherwise the
//     last daily close;
```

(Keep the existing list punctuation: each bullet ends with a semicolon except
the last, which ends with a period. Place the Latest bullet so the final bullet
still ends with a period.)

- [ ] **Step 2: Update the root doc.go**

In the repository root `doc.go`, the `pkg/marketdata` bullet currently reads
"fetches, caches and post-processes daily and intraday prices". Change it to
include the latest quote:

```go
//   - pkg/marketdata, fetches, caches and post-processes daily, intraday and
//     latest (real-time) prices from public sources, addressed by ticker, ISIN
//     or alias; resolves identifiers against the embedded catalog and aligns
//     trading calendars.
```

(Match the surrounding bullets' punctuation exactly; do not introduce an
em-dash.)

- [ ] **Step 3: Update the README**

In `README.md`, under the `### Intraday` subsection in the Data section, add a
new subsection immediately after it:

```markdown
### Latest quote

`Client.Latest` returns the most recent price of an instrument as a `Quote`,
for a live portfolio valuation. A Yahoo-quoted instrument yields its live
market price (`Quote.Live == true`); any other instrument (an FT or Morningstar
fund, whose last NAV close is its latest price) yields its last daily close
(`Quote.Live == false`). It reuses the multi-source resolution, the on-disk
daily cache and the stale fallback, so it answers for every asset and even
offline.

```go
q, err := client.Latest("VWCE")
if err != nil {
	// no usable quote for this identifier
}
value := shares * q.Price // in q.Currency; convert with client.ConvertCurrency
_ = q.Live                // true: real-time; false: last daily close (q.Time)
```
```

Then, in the "Using it as a library" section, extend the `marketdata` bullet
(the one that already mentions `Intraday`) to also mention `Latest`:

```markdown
- `marketdata` resolution (aliases, ISIN, catalog), `Lookup` for an asset's
  full metadata, `Resolve` to inspect the resolved source/symbol, multi-source
  daily downloads, `Intraday` for the live 5-minute path, `Latest` for the
  freshest quote, cache, simdata, proxies.
```

(The exact current wording of that bullet may differ slightly; preserve its
style and just insert the `Latest` clause. Use no em-dashes.)

- [ ] **Step 4: Verify docs build and contain no em-dashes**

Run:
```bash
cd /Users/ben/projects/pofo
go test ./... && go vet ./... && gofmt -l pkg cmd doc.go
EM=$(printf '\xe2\x80\x94'); git diff -- '*.go' README.md doc.go | grep '^+' | grep -c "$EM"
```
Expected: tests pass, vet clean, gofmt prints nothing, and the em-dash count prints `0`.
(macOS note: BSD grep has no `-P`; the `printf` byte form above is the portable way to match U+2014.)

- [ ] **Step 5: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/doc.go doc.go README.md
git commit -m "docs: document the Latest quote capability

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Self-Review Notes

- Spec section 1 (Quote + Latest) -> Task 1. Section 2 (fallback logic, latestFrom, empty guard) -> Task 1 (Latest). Section 3 (fetchYahooSpot) -> Task 1. Section 4 (docs) -> Task 2. Section 5 (testing) -> Task 1 tests, with the offline/stale case covered by the existing TestHistoryStaleCacheFallback (noted, not duplicated). Section 6 (finador) is intentionally not implemented (plan-only).
- Type and signature names are consistent across tasks: `Quote`, `Client.Latest`, `fetchYahooSpot`, `latestFrom`, fields `Price/Time/Currency/Source/Live`.
- `pkg/marketdata` imports no `pkg/chart`; `latest.go` imports only `fmt` and `time`; `fetchYahooSpot` reuses yahoo.go's existing imports.
- The live and fallback paths are both tested via httptest stubs; no test reaches the real Yahoo API (newTestClient stubs every base).
```
