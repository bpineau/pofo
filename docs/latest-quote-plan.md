# Latest Quote Implementation Plan

> **Status: DONE (implemented 2026-07-02).** Revised 2026-07-02
> against master as it stands (context.Context threaded through the client,
> yahooGet host fallback, SplitSim convention); the original 2026-06-28 plan
> predated those changes and was never executed.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a one-call "most recent price" capability to pkg/marketdata for real-time portfolio valuation.

**Architecture:** A small neutral value type Quote and a method Client.Latest(ctx, id) that returns the freshest available price per instrument: the live Yahoo regular-market price when the instrument is Yahoo-quoted, otherwise the last daily close (FT or Morningstar NAV), reusing the existing multi-source resolution, on-disk cache and stale fallback. marketdata does not import chart.

**Tech Stack:** Go 1.26 standard library only. Tests use net/http/httptest and the existing newTestClient / chartJSON helpers in pkg/marketdata.

## Context for a fresh session

This plan implements the spec at `docs/latest-quote-design.md` (read it first).
The following already exist on master and are used by this plan, no need to
create them:

- `func (c *Client) yahooSymbol(ctx context.Context, id string) (string, bool)`
  in `pkg/marketdata/intraday.go`: maps an identifier to a Yahoo symbol with NO
  resolution network call (a ticker maps to itself; an ISIN is covered only
  when its cached or catalog resolution already points at Yahoo).
- `var ErrNotCovered = errors.New("not covered")` in
  `pkg/marketdata/intraday.go`.
- `func SplitSim(id string) (base string, sim bool)` in
  `pkg/marketdata/aliases.go`: strips the "SIM" suffix.
- `func (c *Client) Fetch(ctx context.Context, id string, from time.Time)
  (*Series, error)` in `pkg/marketdata/client.go`. `Series` (in
  `pkg/marketdata/types.go`) has fields `Currency string`, `Source string`,
  and a method `Last() Point`; `Point` has `Date time.Time` and
  `Close float64`. `Last()` returns the zero Point for an empty series.
- `func (c *Client) yahooGet(ctx context.Context, base, path string)
  ([]byte, error)` in `pkg/marketdata/yahoo.go`: GET with the query1/query2
  host fallback; pass `c.ChartBase` as base.
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
- Daily series stay on adjusted close (correct here: the adjustment factor is 1 at the most recent bar, so the last adjusted close equals the last raw close).
- `make check` stays green after every task; new/edited Go is gofmt-clean.
- finador changes are a separate follow-up in the finador repository, not part of this plan.

---

## File Structure

- `pkg/marketdata/latest.go` (new): `Quote`, `latestFrom`, `Client.Latest`.
- `pkg/marketdata/yahoo.go` (modify): add `fetchYahooSpot`.
- `pkg/marketdata/latest_test.go` (new): spot parse, live path, SIM strip, and fallback tests, plus the `spotJSON` fixture helper.
- `pkg/marketdata/doc.go`, `pkg/marketdata/example_test.go`, root `doc.go`, `README.md` (modify): documentation pass.

---

### Task 1: Quote type, Yahoo spot fetch, and Latest

**Files:**
- Create: `pkg/marketdata/latest.go`
- Modify: `pkg/marketdata/yahoo.go`
- Test: `pkg/marketdata/latest_test.go`

**Interfaces:**
- Consumes: `Client.yahooGet`, `Client.yahooSymbol`, `Client.Fetch`, `SplitSim`, `Series`/`Point`, `ErrNotCovered`, `url.PathEscape`.
- Produces:
  - `type Quote struct { Price float64; Time time.Time; Currency string; Source string; Live bool }`
  - `func (c *Client) Latest(ctx context.Context, id string) (*Quote, error)`
  - `func (c *Client) fetchYahooSpot(ctx context.Context, symbol string) (*Quote, error)`
  - `func latestFrom() time.Time`

- [x] **Step 1: Write the failing tests**

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

	q, err := c.Latest(t.Context(), "VOO")
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

func TestLatestStripsSIM(t *testing.T) {
	at := time.Date(2024, 3, 1, 18, 0, 0, 0, time.UTC)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", 501.25, at))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest(t.Context(), "VOOSIM")
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Price != 501.25 {
		t.Fatalf("SIM id should quote as its base: %+v", q)
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

	q, err := c.Latest(t.Context(), "SPY")
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

	_, err := c.fetchYahooSpot(t.Context(), "NOPX")
	if !errors.Is(err, ErrNotCovered) {
		t.Fatalf("err = %v, want ErrNotCovered", err)
	}
}
```

- [x] **Step 2: Run the tests to verify they fail**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run 'TestLatest|TestFetchYahooSpot' -v`
Expected: FAIL to compile (`undefined: Client.Latest`, `undefined: Client.fetchYahooSpot`).

- [x] **Step 3: Create the Quote type, latestFrom and Latest**

Create `pkg/marketdata/latest.go`:

```go
package marketdata

import (
	"context"
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
// both the disk cache and the in-process memoization key the window at day
// granularity, so repeated Latest calls reuse one cache entry.
func latestFrom() time.Time { return time.Now().AddDate(-1, 0, 0) }

// Latest returns the freshest available price for an identifier: the live
// Yahoo market price when the instrument is Yahoo-quoted, otherwise the last
// daily close (FT or Morningstar NAV), served from the on-disk cache when
// fresh and from stale data on a failed refresh. A "SIM" suffix is ignored
// (see SplitSim): simulated history never changes the current price.
//
// Like Intraday, the live path is stateless: Latest performs no caching of the
// live price, so a caller valuing a portfolio repeatedly should keep its own
// short-TTL cache. The daily-close fallback path uses the existing on-disk
// daily cache.
func (c *Client) Latest(ctx context.Context, id string) (*Quote, error) {
	base, _ := SplitSim(id)
	if symbol, ok := c.yahooSymbol(ctx, base); ok {
		if q, err := c.fetchYahooSpot(ctx, symbol); err == nil {
			return q, nil
		}
		// Spot unavailable (not covered, throttled, or missing field): fall
		// through to the last daily close.
	}
	s, err := c.Fetch(ctx, base, latestFrom())
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

- [x] **Step 4: Add the Yahoo spot fetcher**

Add to `pkg/marketdata/yahoo.go` (after `fetchYahooIntraday`):

```go
// fetchYahooSpot reads the live regular-market price from the Yahoo chart meta.
// It returns ErrNotCovered when Yahoo serves no usable price for the symbol, so
// Latest can fall back to the last daily close.
func (c *Client) fetchYahooSpot(ctx context.Context, symbol string) (*Quote, error) {
	path := fmt.Sprintf("/v8/finance/chart/%s?interval=1d&range=1d", url.PathEscape(symbol))
	body, err := c.yahooGet(ctx, c.ChartBase, path)
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

Note: `yahoo.go` already imports `context`, `encoding/json`, `fmt`, `net/url`, `sort`, `time`, so no import change is needed there. `latest.go` imports only `context`, `fmt` and `time`.

- [x] **Step 5: Run the tests to verify they pass**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run 'TestLatest|TestFetchYahooSpot' -v && go vet ./pkg/marketdata/ && gofmt -l pkg/marketdata/latest.go pkg/marketdata/yahoo.go pkg/marketdata/latest_test.go`
Expected: PASS; vet clean; gofmt prints nothing.

Note on the offline/stale path: the spec lists "Latest serves a Quote offline via the stale daily cache". Latest's fallback simply delegates to Fetch, whose stale-cache behavior is already covered by the stale-fallback tests in client_test.go. Do not add a redundant test for it.

- [x] **Step 6: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/latest.go pkg/marketdata/yahoo.go pkg/marketdata/latest_test.go
git commit -m "marketdata: add Latest quote (live Yahoo spot, daily-close fallback)

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: Documentation pass (English, no em-dashes)

**Files:**
- Modify: `pkg/marketdata/doc.go`, root `doc.go`, `README.md`, `docs/latest-quote-design.md`, `docs/latest-quote-plan.md`

**Interfaces:**
- Consumes: `Quote`, `Client.Latest` from Task 1.
- Produces: documentation only, no code behavior change.

- [x] **Step 1: Update pkg/marketdata/doc.go**

Add a "Latest quote" section between the existing "# Intraday" and
"# Simulated data" sections:

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

Then, in the same file's "# Toolbox" list, add one bullet next to the Intraday
or Resolve bullet:

```go
//   - Client.Latest returns the freshest known price (a Quote) for an
//     identifier, the live Yahoo market price when available, otherwise the
//     last daily close;
```

(Keep the existing list punctuation: each bullet ends with a semicolon except
the last, which ends with a period. Place the Latest bullet so the final bullet
still ends with a period.)

- [x] **Step 2: Update the root doc.go**

In the repository root `doc.go`, the `pkg/marketdata` bullet (line 14) reads
"fetches, caches and post-processes daily and intraday / prices". Change it to
include the latest quote:

```go
//   - pkg/marketdata: fetches, caches and post-processes daily, intraday and
//     latest (real-time) prices from public sources, addressed by ticker, ISIN
//     or alias; resolves identifiers against the embedded catalog and aligns
//     trading calendars.
```

(Match the surrounding bullets' punctuation exactly; do not introduce an
em-dash.)

- [x] **Step 3: Update the README**

In `README.md`, right after the `### Intraday` subsection (which ends with the
chart snippet), add:

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
q, err := client.Latest(ctx, "VWCE")
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
- `marketdata`: resolution (aliases, ISIN, catalog), `Lookup` for an asset's
  full metadata, `Resolve` to inspect the resolved source/symbol, multi-source
  daily downloads, `Intraday` for the live 5-minute path, `Latest` for the
  freshest quote, cache, simdata, proxies; `FetchExtended` bundles the whole
  per-asset pipeline in one call.
```

(Preserve the bullet's current style; insert only the `Latest` clause. Use no
em-dashes.)

- [x] **Step 4: Add a runnable example**

House rule: every new API gets an example in the package's
`example_test.go`. Add an `ExampleClient_Latest` (network-free: skip execution
with a no-output pattern or stub, matching how the file treats other
network-bound examples; follow the file's existing idiom).

- [x] **Step 5: Verify docs build and contain no em-dashes**

Run:
```bash
cd /Users/ben/projects/pofo
make check
EM=$(printf '\xe2\x80\x94'); grep -l "$EM" pkg/marketdata/latest.go pkg/marketdata/latest_test.go pkg/marketdata/doc.go pkg/marketdata/example_test.go doc.go README.md docs/latest-quote-design.md docs/latest-quote-plan.md; echo "em-dash check done"
```
Expected: make check passes and the grep lists no file.
(macOS note: BSD grep has no `-P`; the `printf` byte form above is the portable way to match U+2014.)

- [x] **Step 6: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/doc.go pkg/marketdata/example_test.go doc.go README.md docs/latest-quote-design.md docs/latest-quote-plan.md
git commit -m "docs: document the Latest quote capability

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review Notes

- Spec section 1 (Quote + Latest) -> Task 1. Section 2 (fallback logic, SIM
  strip, latestFrom, empty guard) -> Task 1 (Latest). Section 3
  (fetchYahooSpot) -> Task 1. Section 4 (docs) -> Task 2. Section 5 (testing)
  -> Task 1 tests, with the offline/stale case covered by the existing
  stale-fallback tests (noted, not duplicated). Section 6 (finador) is
  implemented in the finador repository, after this plan.
- Type and signature names are consistent across tasks: `Quote`,
  `Client.Latest`, `fetchYahooSpot`, `latestFrom`, fields
  `Price/Time/Currency/Source/Live`; every network entry point takes a
  `context.Context` first, matching the rest of the client.
- `pkg/marketdata` imports no `pkg/chart`; `latest.go` imports only `context`,
  `fmt` and `time`; `fetchYahooSpot` reuses yahoo.go's existing imports.
- The live, SIM and fallback paths are tested via httptest stubs; no test
  reaches the real Yahoo API (newTestClient stubs every base).
