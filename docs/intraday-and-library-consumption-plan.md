# Intraday Support and Library Consumption Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add intraday price data to pofo and make the library comfortable to consume from an external application (catalog lookup, public resolution, intraday chart axis, docs).

**Architecture:** New capabilities enter `pkg/marketdata`, `pkg/chart` and `pkg/datasets` as general, neutral types in pofo's own idiom. Intraday is a stateless live fetch (no caching inside pofo). Intraday rendering reuses the existing `chart.Line` by teaching its time axis to label sub-day spans. No third-party dependency is added.

**Tech Stack:** Go 1.26 standard library only. Tests use `net/http/httptest` and the existing `newTestClient` / `chartJSON` helpers in `pkg/marketdata`.

## Global Constraints

- Module is `github.com/bpineau/pofo`, Go 1.26, dependency-free (standard library only). `time/tzdata` is standard library and allowed.
- All documentation (godoc comments and README) is English only and uses no em-dashes. Use commas, colons or parentheses instead.
- `pkg/marketdata` must NOT import `pkg/chart`. The intraday-to-chart mapping lives in caller code and in examples.
- pofo never caches intraday data (no disk, no memo): the caller owns that policy.
- Daily series stay on adjusted close. No dividend-event model is added.
- `go test ./...` and `go vet ./...` stay green after every task.
- Do NOT modify the `finador` project anywhere in this plan.

---

## File Structure

- `pkg/marketdata/intraday.go` (new): `IntradayPoint`, `IntradaySeries`, `ErrNotCovered`, `Client.Intraday`, `Client.yahooSymbol`.
- `pkg/marketdata/yahoo.go` (modify): add `fetchYahooIntraday`.
- `pkg/marketdata/intraday_test.go` (new): intraday parse and coverage tests.
- `pkg/marketdata/resolve.go` (new): `Resolution`, `Client.Resolve`, `toResolution`.
- `pkg/marketdata/resolve_test.go` (new): resolution tests.
- `pkg/chart/svg.go` (modify): sub-day branch in `timeTicks`.
- `pkg/chart/svg_test.go` (modify): intraday tick test.
- `pkg/chart/example_test.go` (modify): intraday `Line` example.
- `pkg/datasets/catalog.go` (modify): `Lookup` plus a lazy index.
- `pkg/datasets/datasets_test.go` (modify): `Lookup` tests.
- `pkg/marketdata/doc.go`, `pkg/chart/doc.go`, `pkg/datasets/doc.go`, root `doc.go`, `README.md` (modify): documentation pass.

---

### Task 1: Intraday data in `pkg/marketdata`

**Files:**
- Create: `pkg/marketdata/intraday.go`
- Modify: `pkg/marketdata/yahoo.go`
- Test: `pkg/marketdata/intraday_test.go`

**Interfaces:**
- Consumes: existing `Client.get`, `CanonicalID`, `IsISIN`, `Client.loadResolution`, `resolution` struct.
- Produces:
  - `type IntradayPoint struct { Time time.Time; Close float64 }`
  - `type IntradaySeries struct { Symbol, Name, Currency, Source string; Points []IntradayPoint }`
  - `func (s *IntradaySeries) First() IntradayPoint`
  - `func (s *IntradaySeries) Last() IntradayPoint`
  - `var ErrNotCovered = errors.New("not covered")`
  - `func (c *Client) Intraday(id string) (*IntradaySeries, error)`
  - `func (c *Client) yahooSymbol(id string) (string, bool)`
  - `func (c *Client) fetchYahooIntraday(symbol string) (*IntradaySeries, error)`

- [ ] **Step 1: Write the failing tests**

Create `pkg/marketdata/intraday_test.go`:

```go
package marketdata

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// intradayJSON builds a Yahoo 5-minute chart fixture for one trading day.
func intradayJSON(currency, tz string, base time.Time, closes []float64) string {
	ts, cl := "", ""
	for i, c := range closes {
		if i > 0 {
			ts += ","
			cl += ","
		}
		ts += fmt.Sprint(base.Add(time.Duration(i)*5*time.Minute).Unix())
		cl += fmt.Sprint(c)
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":%q,"exchangeTimezoneName":%q},"timestamp":[%s],"indicators":{"quote":[{"close":[%s]}]}}],"error":null}}`,
		currency, tz, ts, cl)
}

func TestIntradayParse(t *testing.T) {
	base := time.Date(2024, 3, 1, 14, 30, 0, 0, time.UTC) // 09:30 New York
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("interval"); got != "5m" {
			t.Errorf("interval = %q, want 5m", got)
		}
		if got := r.URL.Query().Get("range"); got != "1d" {
			t.Errorf("range = %q, want 1d", got)
		}
		fmt.Fprint(w, intradayJSON("USD", "America/New_York", base, []float64{500, 501, 502}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Intraday("VOO")
	if err != nil {
		t.Fatal(err)
	}
	if s.Currency != "USD" || s.Source != "yahoo" || len(s.Points) != 3 {
		t.Fatalf("series misread: %+v", s)
	}
	if s.Last().Close != 502 {
		t.Errorf("last close = %v, want 502", s.Last().Close)
	}
	if h := s.First().Time.Hour(); h != 9 {
		t.Errorf("first point hour = %d, want 9 (exchange local time)", h)
	}
}

func TestIntradayUnknownISINNotCovered(t *testing.T) {
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { requests++ })
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	_, err := c.Intraday("FR0000000000")
	if !errors.Is(err, ErrNotCovered) {
		t.Fatalf("err = %v, want ErrNotCovered", err)
	}
	if requests != 0 {
		t.Errorf("intraday made %d requests resolving an unknown ISIN, want 0", requests)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run 'TestIntraday' -v`
Expected: FAIL to compile (`undefined: Client.Intraday`, `ErrNotCovered`).

- [ ] **Step 3: Create the intraday types, helpers and method**

Create `pkg/marketdata/intraday.go`:

```go
package marketdata

import (
	"errors"
	"fmt"
	"strings"
	_ "time/tzdata" // exchange time zones, without depending on the host OS
	"time"
)

// ErrNotCovered reports that a request cannot be served for an identifier,
// for example intraday data for an instrument quoted only by a fund source.
var ErrNotCovered = errors.New("not covered")

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

// First returns the earliest point, or the zero IntradayPoint if empty.
func (s *IntradaySeries) First() IntradayPoint {
	if len(s.Points) == 0 {
		return IntradayPoint{}
	}
	return s.Points[0]
}

// Last returns the latest point, or the zero IntradayPoint if empty.
func (s *IntradaySeries) Last() IntradayPoint {
	if len(s.Points) == 0 {
		return IntradayPoint{}
	}
	return s.Points[len(s.Points)-1]
}

// Intraday returns today's intraday price path (5-minute resolution) for an
// identifier, fetched live from Yahoo Finance.
//
// Unlike Fetch, Intraday never touches the on-disk cache: an intraday series is
// valid only for today and goes stale within minutes. Callers that view an
// asset repeatedly should keep their own short-TTL cache; the fetch is
// deliberately stateless so that the caching policy stays with the caller.
//
// Only Yahoo-quoted instruments have intraday data. An identifier that resolves
// to a fund-only source (Financial Times, Morningstar), or that has no known
// Yahoo symbol, returns ErrNotCovered. Intraday does not perform a network
// resolution: it reuses the symbol Fetch already learned (the bundled catalog
// plus the on-disk resolution cache). For an unseen ISIN, call Fetch first.
func (c *Client) Intraday(id string) (*IntradaySeries, error) {
	symbol, ok := c.yahooSymbol(id)
	if !ok {
		return nil, fmt.Errorf("%s: %w", id, ErrNotCovered)
	}
	return c.fetchYahooIntraday(symbol)
}

// yahooSymbol maps a user identifier to a Yahoo symbol without any resolution
// network call: a plain ticker is itself, an ISIN is covered only when its
// cached or catalog resolution already points at Yahoo.
func (c *Client) yahooSymbol(id string) (string, bool) {
	canonical := CanonicalID(id)
	if IsISIN(canonical) {
		if res, ok := c.loadResolution(canonical); ok && res.Source == "yahoo" && res.Symbol != "" {
			return res.Symbol, true
		}
		return "", false
	}
	return strings.ToUpper(strings.TrimSpace(canonical)), true
}
```

- [ ] **Step 4: Add the Yahoo intraday fetcher**

Add to `pkg/marketdata/yahoo.go` (after `fetchYahoo`):

```go
// fetchYahooIntraday downloads the current day's 5-minute price path from the
// Yahoo Finance chart API. It returns ErrNotCovered when Yahoo serves no
// intraday result for the symbol.
func (c *Client) fetchYahooIntraday(symbol string) (*IntradaySeries, error) {
	u := fmt.Sprintf("%s/v8/finance/chart/%s?interval=5m&range=1d",
		c.ChartBase, url.PathEscape(symbol))
	body, err := c.get(u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency             string `json:"currency"`
					ExchangeTimezoneName string `json:"exchangeTimezoneName"`
					LongName             string `json:"longName"`
					ShortName            string `json:"shortName"`
				} `json:"meta"`
				Timestamp  []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Close []*float64 `json:"close"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error *struct {
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable yahoo intraday response: %w", err)
	}
	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo intraday: %s", resp.Chart.Error.Description)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("%s: %w", symbol, ErrNotCovered)
	}
	r := resp.Chart.Result[0]
	loc, err := time.LoadLocation(r.Meta.ExchangeTimezoneName)
	if err != nil {
		loc = time.UTC
	}
	name := r.Meta.LongName
	if name == "" {
		name = r.Meta.ShortName
	}
	s := &IntradaySeries{Symbol: symbol, Name: name, Currency: r.Meta.Currency, Source: "yahoo"}
	var closes []*float64
	if len(r.Indicators.Quote) > 0 {
		closes = r.Indicators.Quote[0].Close
	}
	for i, ts := range r.Timestamp {
		if i >= len(closes) || closes[i] == nil || *closes[i] <= 0 {
			continue
		}
		s.Points = append(s.Points, IntradayPoint{Time: time.Unix(ts, 0).In(loc), Close: *closes[i]})
	}
	return s, nil
}
```

Note: `yahoo.go` already imports `encoding/json`, `fmt`, `net/url`, `time`. No import change needed there. `intraday.go` imports `errors`, `fmt`, `strings`, `time` and the blank `time/tzdata`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run 'TestIntraday' -v && go vet ./pkg/marketdata/`
Expected: PASS, vet clean.

- [ ] **Step 6: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/intraday.go pkg/marketdata/yahoo.go pkg/marketdata/intraday_test.go
git commit -m "marketdata: add intraday price fetch (Yahoo 5m)

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Intraday-aware chart axis

**Files:**
- Modify: `pkg/chart/svg.go` (function `timeTicks`, around line 171)
- Test: `pkg/chart/svg_test.go`
- Modify: `pkg/chart/example_test.go`

**Interfaces:**
- Consumes: existing `tick` struct, `timeTicks(from, to time.Time) []tick`, `chart.Line`, `chart.Series`.
- Produces: `timeTicks` returns `15:04`-labelled ticks when `to.Sub(from) <= 36h`; daily behavior unchanged.

- [ ] **Step 1: Write the failing test**

Add to `pkg/chart/svg_test.go`:

```go
func TestTimeTicksIntraday(t *testing.T) {
	from := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
	to := from.Add(6 * time.Hour)
	ticks := timeTicks(from, to)
	if len(ticks) < 2 {
		t.Fatalf("got %d ticks, want several", len(ticks))
	}
	if ticks[0].label != "09:00" {
		t.Errorf("first label = %q, want 09:00 (clock time)", ticks[0].label)
	}
	for _, tk := range ticks {
		if !strings.Contains(tk.label, ":") {
			t.Errorf("intraday label %q is not a clock time", tk.label)
		}
	}
}

func TestTimeTicksDailyUnchanged(t *testing.T) {
	from := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tk := range timeTicks(from, to) {
		if strings.Contains(tk.label, ":") {
			t.Errorf("multi-year label %q must be a year, not a clock time", tk.label)
		}
	}
}
```

(`svg_test.go` already imports `strings`, `testing`, `time`. Add any missing import.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/chart/ -run 'TestTimeTicks' -v`
Expected: FAIL (`first label = "2024-03"`, want `09:00`).

- [ ] **Step 3: Add the sub-day branch to `timeTicks`**

In `pkg/chart/svg.go`, at the start of `timeTicks` (before the `years :=` line):

```go
func timeTicks(from, to time.Time) []tick {
	// Sub-day spans (intraday) get clock-time labels.
	if d := to.Sub(from); d > 0 && d <= 36*time.Hour {
		const n = 5
		out := make([]tick, 0, n+1)
		for i := 0; i <= n; i++ {
			t := from.Add(time.Duration(i) * d / n)
			out = append(out, tick{t, t.Format("15:04")})
		}
		return out
	}
	years := to.Year() - from.Year()
	// ... existing body unchanged ...
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/chart/ -run 'TestTimeTicks' -v`
Expected: PASS.

- [ ] **Step 5: Add an intraday example**

Add to `pkg/chart/example_test.go`:

```go
// Line renders an intraday path the same way as a daily one: feed it a series
// of timestamps and prices spanning a single day, and the time axis switches
// to clock-time labels.
func ExampleLine_intraday() {
	open := time.Date(2024, 3, 1, 9, 30, 0, 0, time.UTC)
	var dates []time.Time
	var v []float64
	for i := range 78 { // a 6.5h session at 5-minute resolution
		dates = append(dates, open.Add(time.Duration(i)*5*time.Minute))
		v = append(v, 500+float64(i)*0.05)
	}
	svg := chart.Line(chart.Options{Title: "VOO today"}, []chart.Series{
		{Name: "price USD", Dates: dates, Values: v},
	})
	fmt.Println(strings.HasPrefix(svg, "<svg"), strings.Contains(svg, ":"))
	// Output:
	// true true
}
```

- [ ] **Step 6: Run the chart suite and vet**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/chart/ && go vet ./pkg/chart/`
Expected: PASS, vet clean.

- [ ] **Step 7: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/chart/svg.go pkg/chart/svg_test.go pkg/chart/example_test.go
git commit -m "chart: label sub-day spans with clock times for intraday

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: `datasets.Lookup`

**Files:**
- Modify: `pkg/datasets/catalog.go`
- Test: `pkg/datasets/datasets_test.go`

**Interfaces:**
- Consumes: existing `Catalog() []Asset`, `Asset` fields `ID`, `ISIN`, `Aliases`.
- Produces: `func Lookup(id string) (Asset, bool)`.

- [ ] **Step 1: Write the failing test**

Add to `pkg/datasets/datasets_test.go`:

```go
func TestLookup(t *testing.T) {
	all := Catalog()
	if len(all) == 0 {
		t.Fatal("empty catalog")
	}
	want := all[0]

	got, ok := Lookup(want.ID)
	if !ok || got.ID != want.ID {
		t.Fatalf("Lookup by id: got %q (%v), want %q", got.ID, ok, want.ID)
	}
	if _, ok := Lookup("__definitely_not_an_asset__"); ok {
		t.Error("Lookup of an unknown id returned ok")
	}
	if want.ISIN != "" {
		if g, ok := Lookup(want.ISIN); !ok || g.ID != want.ID {
			t.Errorf("Lookup by ISIN failed for %q", want.ISIN)
		}
		// Case-insensitive.
		if g, ok := Lookup(strings.ToLower(want.ISIN)); !ok || g.ID != want.ID {
			t.Errorf("Lookup is not case-insensitive for %q", want.ISIN)
		}
	}
}
```

(Add `"strings"` to the `datasets_test.go` import block.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/datasets/ -run TestLookup -v`
Expected: FAIL to compile (`undefined: Lookup`).

- [ ] **Step 3: Implement `Lookup` with a lazy index**

Add to `pkg/datasets/catalog.go`:

```go
import (
	"encoding/json"
	"strings"
	"sync"
)

var (
	indexOnce sync.Once
	index     map[string]Asset
)

func buildIndex() {
	assets := Catalog()
	index = make(map[string]Asset, len(assets)*2)
	put := func(key string, a Asset) {
		if key == "" {
			return
		}
		k := strings.ToUpper(strings.TrimSpace(key))
		if _, exists := index[k]; !exists {
			index[k] = a
		}
	}
	for _, a := range assets {
		put(a.ID, a)
		put(a.ISIN, a)
		for _, alias := range a.Aliases {
			put(alias, a)
		}
	}
}

// Lookup returns the catalog asset for an identifier (its id, ISIN or any
// alias, case-insensitive) and whether it was found. The index is built once
// on first use. The first registration of a key wins, so a primary id is never
// shadowed by another asset's alias.
func Lookup(id string) (Asset, bool) {
	indexOnce.Do(buildIndex)
	a, ok := index[strings.ToUpper(strings.TrimSpace(id))]
	return a, ok
}
```

(Note: `catalog.go` currently imports only `encoding/json`. Replace its import line with the grouped block above so `strings` and `sync` are available.)

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/datasets/ -run TestLookup -v && go vet ./pkg/datasets/`
Expected: PASS, vet clean.

- [ ] **Step 5: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/datasets/catalog.go pkg/datasets/datasets_test.go
git commit -m "datasets: add Lookup by id, ISIN or alias

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: Public resolution `marketdata.Resolve`

**Files:**
- Create: `pkg/marketdata/resolve.go`
- Test: `pkg/marketdata/resolve_test.go`

**Interfaces:**
- Consumes: existing `resolution` struct (fields `Source`, `Symbol`, `Xid`, `Name`, `Currency`), `Client.loadResolution`, `Client.saveResolution`, `Client.Fetch`, `CanonicalID`, `Series` fields `Source`, `Symbol`, `Name`, `Currency`.
- Produces:
  - `type Resolution struct { Source, Symbol, Xid, Name, Currency string }`
  - `func (c *Client) Resolve(id string) (Resolution, error)`
  - `func toResolution(r resolution) Resolution`

- [ ] **Step 1: Write the failing test**

Create `pkg/marketdata/resolve_test.go`:

```go
package marketdata

import (
	"net/http"
	"testing"
)

func TestResolveFromCacheNoNetwork(t *testing.T) {
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { requests++ })
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	// Seed a disk resolution, then resolve must answer from it, no network.
	c.saveResolution("XX1234567890", resolution{
		Source: "yahoo", Symbol: "AAPL", Name: "Apple Inc.", Currency: "USD",
	})

	got, err := c.Resolve("XX1234567890")
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "yahoo" || got.Symbol != "AAPL" || got.Currency != "USD" {
		t.Fatalf("resolution misread: %+v", got)
	}
	if requests != 0 {
		t.Errorf("Resolve hit the network %d times for a cached id, want 0", requests)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run TestResolve -v`
Expected: FAIL to compile (`undefined: Client.Resolve`).

- [ ] **Step 3: Implement `Resolve`**

Create `pkg/marketdata/resolve.go`:

```go
package marketdata

import "time"

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

func toResolution(r resolution) Resolution {
	return Resolution{Source: r.Source, Symbol: r.Symbol, Xid: r.Xid, Name: r.Name, Currency: r.Currency}
}

// resolveFrom is the history depth Resolve fetches over when it must run a full
// resolution: deep enough that the multi-source search settles on the same
// instrument it would for a real long-horizon request.
func resolveFrom() time.Time { return time.Now().AddDate(-15, 0, 0) }

// Resolve returns the instrument pofo would quote for a user identifier
// (ticker, ISIN or alias). It uses the bundled catalog and the on-disk
// resolution cache first, then the same multi-source search Fetch uses. It may
// perform network I/O and caches the result, so a later Fetch of the same id
// reuses this work.
func (c *Client) Resolve(id string) (Resolution, error) {
	canonical := CanonicalID(id)
	if res, ok := c.loadResolution(canonical); ok {
		return toResolution(res), nil
	}
	s, err := c.Fetch(id, resolveFrom())
	if err != nil {
		return Resolution{}, err
	}
	// An ISIN or fund path adopts a resolution Fetch can now load back.
	if res, ok := c.loadResolution(canonical); ok {
		r := toResolution(res)
		if r.Currency == "" {
			r.Currency = s.Currency
		}
		if r.Name == "" {
			r.Name = s.Name
		}
		return r, nil
	}
	// A direct ticker keeps no resolution file: its identity is the series.
	return Resolution{Source: s.Source, Symbol: s.Symbol, Name: s.Name, Currency: s.Currency}, nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd /Users/ben/projects/pofo && go test ./pkg/marketdata/ -run TestResolve -v && go vet ./pkg/marketdata/`
Expected: PASS, vet clean.

- [ ] **Step 5: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/resolve.go pkg/marketdata/resolve_test.go
git commit -m "marketdata: expose multi-source resolution via Resolve

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: Documentation pass (English, no em-dashes)

**Files:**
- Modify: `pkg/marketdata/doc.go`, `pkg/chart/doc.go`, `pkg/datasets/doc.go`, `doc.go` (root), `README.md`

**Interfaces:**
- Consumes: all symbols added in Tasks 1 to 4.
- Produces: documentation only. No code behavior change.

- [ ] **Step 1: Update package docs**

In `pkg/marketdata/doc.go`, add paragraphs covering: intraday via `Client.Intraday` (live, stateless, caller owns the cache, Yahoo only, `ErrNotCovered` otherwise); FX via `ConvertCurrency` (reprices a whole series through Yahoo crosses); and `Resolve` as the reusable resolution entry point. English, no em-dashes.

In `pkg/chart/doc.go`, add one sentence: `Line` labels sub-day spans with clock times, so the same renderer draws daily and intraday series.

In `pkg/datasets/doc.go`, add one sentence presenting `Lookup` as the by-identifier entry point external consumers use to read an asset's geography, sectors, factors and exposures.

In the root `doc.go`, extend the `pkg/marketdata` bullet to mention intraday quotes alongside daily prices. Keep the existing prose style but use no new em-dashes.

- [ ] **Step 2: Update the README**

In `README.md`, add:
- An "Intraday" subsection under the market-data material, with the caller-side mapping to `chart.Series`:

```go
s, err := client.Intraday("VOO")
if err != nil {
	// errors.Is(err, marketdata.ErrNotCovered) means no intraday for this asset
}
ser := chart.Series{Name: s.Name}
for _, p := range s.Points {
	ser.Dates = append(ser.Dates, p.Time)
	ser.Values = append(ser.Values, p.Close)
}
svg := chart.Line(chart.Options{Title: s.Name}, []chart.Series{ser})
```

- A short FX note pointing at `Client.ConvertCurrency`.
- A new section "Using pofo as a library from another application" covering: `datasets.Lookup` for metadata (geography, sectors, factors), `marketdata.Client` for daily and intraday quotes, `marketdata.Resolve` to reuse the resolution, `chart.Line` for rendering, and the principle that the consumer maps pofo types onto its own domain (pofo stays dependency-free and analytics-float).

Write all new prose in English with no em-dashes.

- [ ] **Step 3: Verify docs build and examples run**

Run: `cd /Users/ben/projects/pofo && go test ./... && go vet ./... && gofmt -l pkg doc.go`
Expected: tests PASS, vet clean, `gofmt -l` prints nothing.

- [ ] **Step 4: Check the new docs introduced no em-dashes**

Run: `cd /Users/ben/projects/pofo && git diff --cached -U0 -- '*.go' README.md | grep '^+' | grep -c '(em-dash)'`
Expected: `0`. (Run after `git add` in Step 5 if needed, or use `git diff` against the working tree.)

- [ ] **Step 5: Commit**

```bash
cd /Users/ben/projects/pofo
git add pkg/marketdata/doc.go pkg/chart/doc.go pkg/datasets/doc.go doc.go README.md
git commit -m "docs: document intraday, FX, Resolve and library consumption

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 6: Em-dash purge (separate commit)

**Files:**
- Modify: every `*.go` and `*.md` file in the repository that contains an em-dash, except files already committed em-dash-free in Tasks 1 to 5.

**Interfaces:**
- Consumes: nothing. Mechanical text edit.
- Produces: zero em-dashes in the repository. No code behavior change.

- [ ] **Step 1: List the affected files**

Run: `cd /Users/ben/projects/pofo && grep -rl '(em-dash)' --include='*.go' --include='*.md' .`
Expected: a list of roughly 40 files. Record it.

- [ ] **Step 2: Replace em-dashes file by file, preserving meaning**

For each listed file, replace every em-dash (U+2014) with the punctuation that fits the sentence: a comma for an aside, a colon before an explanation or list, or parentheses for a parenthetical. Do NOT blanket-replace with a single character: read each occurrence and choose. Common pofo patterns:
- `X, the curated thing` (was: X [em-dash] the curated thing).
- `do Y, for example Z` (was: do Y [em-dash] for example Z).
- `A (B) C` or `A, B, C` (was: paired em-dashes A [em-dash] B [em-dash] C).

Keep code identifiers and string literals untouched if a literal em-dash is ever semantically required (none are expected in pofo).

- [ ] **Step 3: Verify zero em-dashes remain**

Run: `cd /Users/ben/projects/pofo && grep -rc '(em-dash)' --include='*.go' --include='*.md' . | grep -v ':0$' || echo "clean"`
Expected: `clean`.

- [ ] **Step 4: Verify nothing else changed**

Run: `cd /Users/ben/projects/pofo && go test ./... && go vet ./... && gofmt -l pkg cmd doc.go`
Expected: tests PASS, vet clean, `gofmt -l` prints nothing (godoc edits must stay gofmt-clean).

- [ ] **Step 5: Commit**

```bash
cd /Users/ben/projects/pofo
git add -A
git commit -m "docs: replace em-dashes with plain punctuation throughout

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

## Self-Review Notes

- Spec section 1 (intraday) -> Task 1. Section 2 (chart axis) -> Task 2. Section 3 (Lookup) -> Task 3. Section 4 (Resolve) -> Task 4. Section 5 (docs) -> Task 5. Em-dash purge -> Task 6. Section 6 (finador) is intentionally not implemented (plan-only, per the spec). Section 7 (testing) is folded into each task's TDD steps. Section 8 (FIRE) requires no task.
- Type names are consistent across tasks: `IntradaySeries`, `IntradayPoint`, `ErrNotCovered`, `Resolution`, `Resolve`, `Lookup`, `yahooSymbol`, `fetchYahooIntraday`, `toResolution`, `resolveFrom`.
- `pkg/marketdata` does not import `pkg/chart` in any task; the mapping lives in the README and chart example only.
