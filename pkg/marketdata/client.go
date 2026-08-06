package marketdata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"

// Client downloads price histories with an on-disk cache.
// The zero value is not usable; use NewClient.
type Client struct {
	HTTP            *http.Client
	CacheDir        string
	MaxAge          time.Duration // how long a cached download stays fresh
	ChartBase       string
	SearchBase      string
	StooqBase       string
	FTBase          string
	JustETFBase     string
	BoursoramaBase  string
	MorningstarBase string
	EurostatBase    string
	FredBase        string
	UserAgent       string
	Logf            func(format string, args ...any)

	retryDelay time.Duration
	mu         sync.Mutex
	memo       map[string]*Series
	fees       map[string]feesEntry // TER cache, lazily loaded from fees.json
}

// NewClient returns a Client caching downloads under cacheDir. An empty
// cacheDir disables the disk cache entirely (series, resolutions and fees
// live only in the process memory): every fresh client refetches, which
// suits privacy-sensitive callers that keep their own store and do not
// want plaintext quote files revealing their holdings. DefaultCacheDir
// returns the standard shared location.
func NewClient(cacheDir string) *Client {
	return &Client{
		HTTP:            &http.Client{Timeout: 30 * time.Second},
		CacheDir:        cacheDir,
		MaxAge:          30 * 24 * time.Hour,
		ChartBase:       yahooHost2, // query2 first; yahooGet falls back to query1
		SearchBase:      yahooHost2,
		StooqBase:       "https://stooq.com",
		FTBase:          "https://markets.ft.com",
		JustETFBase:     "https://www.justetf.com",
		BoursoramaBase:  "https://www.boursorama.com",
		MorningstarBase: "https://tools.morningstar.fr",
		EurostatBase:    "https://ec.europa.eu",
		FredBase:        "https://fred.stlouisfed.org",
		UserAgent:       defaultUserAgent,
		Logf:            func(string, ...any) {},
		retryDelay:      time.Second,
		memo:            make(map[string]*Series),
	}
}

// resolution is the cached mapping from an ISIN to a quotable instrument.
type resolution struct {
	Source   string `json:"source"` // "yahoo", "ft" or "morningstar"; empty means yahoo (legacy cache files)
	Symbol   string `json:"symbol"` // yahoo symbol, Morningstar id (0P…), or FT base ticker (informational)
	Xid      string `json:"xid"`    // FT internal id, unused otherwise
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// Fetch returns the price history for a user-supplied identifier, after
// resolving aliases (see CanonicalID). A plain ticker goes through Yahoo
// with Stooq as fallback, then through the same search-based resolution as
// ISINs when no direct quote exists (typically European listings needing an
// exchange suffix). An ISIN is resolved via the Yahoo search API, then the
// Financial Times, then Morningstar.
func (c *Client) Fetch(ctx context.Context, id string, from time.Time) (*Series, error) {
	return c.fetch(ctx, id, from, false)
}

// fetch is Fetch with the close-column choice: raw asks for unadjusted
// closes (see FetchOptions.Raw). Only Yahoo distinguishes the two views;
// fund NAVs and rate symbols are their own raw price.
func (c *Client) fetch(ctx context.Context, id string, from time.Time, raw bool) (*Series, error) {
	canonical := CanonicalID(id)
	if canonical != strings.ToUpper(strings.TrimSpace(id)) {
		c.Logf("%s → %s", strings.ToUpper(strings.TrimSpace(id)), canonical)
	}
	var s *Series
	var err error
	switch {
	case isHICP(canonical):
		geo, _ := hicpGeo(canonical)
		s, err = c.fetchHICP(ctx, canonical, geo, from)
	case IsISIN(canonical):
		s, err = c.fetchISIN(ctx, canonical, from, raw)
	default:
		s, err = c.fetchTicker(ctx, canonical, from, raw)
	}
	// A canceled context surfaces as such (errors.Is-able) instead of the
	// per-source failure summary it caused.
	if err != nil && ctx.Err() != nil {
		return nil, fmt.Errorf("%s: %w", canonical, ctx.Err())
	}
	return s, err
}

func isHICP(symbol string) bool { _, ok := hicpGeo(symbol); return ok }

// minGoodPoints is the size below which a series is considered degenerate
// (e.g. a moribund exchange listing): other sources are then consulted in
// the hope of finding a deeper history.
const minGoodPoints = 60

// goodSeries reports whether a series is solid enough to settle on.
func goodSeries(s *Series) bool { return s != nil && len(s.Points) >= minGoodPoints }

// deeper reports whether b offers a deeper usable history than a.
func deeper(a, b *Series) bool {
	if b == nil || len(b.Points) < 2 {
		return false
	}
	if a == nil || len(a.Points) < 2 {
		return true
	}
	switch {
	case goodSeries(a) && !goodSeries(b):
		return false
	case goodSeries(b) && !goodSeries(a):
		return true
	case b.First().Date.Before(a.First().Date):
		return true
	case b.First().Date.Equal(a.First().Date) && len(b.Points) > len(a.Points):
		return true
	}
	return false
}

func (c *Client) fetchISIN(ctx context.Context, isin string, from time.Time, raw bool) (*Series, error) {
	if s, ok := c.cachedResolutionHistory(ctx, isin, from, raw); ok {
		return s, nil
	}
	s, res, failures := c.resolveBest(ctx, isin, from, "", raw)
	if s == nil {
		return nil, fmt.Errorf("ISIN %s: no usable source (%s)", isin, strings.Join(failures, "; "))
	}
	c.adoptResolution(isin, res)
	return s, nil
}

// fetchTicker downloads a plain ticker, falling back to the search-based
// resolution (preferring listings of the same ticker on other exchanges)
// when the direct quote is missing or degenerate.
func (c *Client) fetchTicker(ctx context.Context, ticker string, from time.Time, raw bool) (*Series, error) {
	if s, ok := c.cachedResolutionHistory(ctx, ticker, from, raw); ok {
		return s, nil
	}
	direct, directErr := c.historyView(ctx, ticker, from, raw)
	if directErr == nil && goodSeries(direct) {
		return direct, nil
	}
	resolved, res, failures := c.resolveBest(ctx, ticker, from, ticker, raw)
	if directErr != nil {
		failures = append([]string{directErr.Error()}, failures...)
	}
	if !deeper(direct, resolved) {
		if direct != nil {
			return direct, nil
		}
		return nil, fmt.Errorf("ticker %s: no usable source (%s)", ticker, strings.Join(failures, "; "))
	}
	c.adoptResolution(ticker, res)
	return resolved, nil
}

// cachedResolutionHistory serves an identifier from its cached resolution,
// reporting false when it must be re-resolved.
func (c *Client) cachedResolutionHistory(ctx context.Context, id string, from time.Time, raw bool) (*Series, bool) {
	res, ok := c.loadResolution(id)
	if !ok {
		return nil, false
	}
	s, err := c.historyForResolution(ctx, id, res, from, raw)
	if err == nil && goodSeries(s) {
		return s, true
	}
	if err != nil {
		c.Logf("warning: cached source for %s no longer responds (%v), resolving again…", id, err)
	} else {
		c.Logf("warning: cached history for %s is too short (%d quotes), resolving again…", id, len(s.Points))
	}
	return nil, false
}

// adoptResolution persists a freshly won resolution and logs it.
func (c *Client) adoptResolution(id string, res resolution) {
	c.saveResolution(id, res)
	via := res.Symbol
	switch res.Source {
	case "ft":
		via = "FT"
	case "morningstar":
		via = "Morningstar " + res.Symbol
	}
	c.Logf("%s resolved via %s: %s", id, via, res.Name)
}

// resolveBest tries every known source for an identifier (ISIN or unknown
// ticker) and returns the series with the deepest usable history: each Yahoo
// search candidate (same-ticker listings and fund entries first), then the
// Financial Times when years remain uncovered, then a Morningstar fund id
// obtained from Boursorama as a last resort.
func (c *Client) resolveBest(ctx context.Context, query string, from time.Time, preferBase string, raw bool) (*Series, resolution, []string) {
	// Candidates compete in tiered slots so that the right instrument beats
	// the deep one: a young same-ticker ETF must win against a namesake
	// stock (SPEA the PEA ETF vs Saipem SpA) and against a fuzzy-matched
	// fund (SPRX). Within a slot, the deepest usable history wins. For ISIN
	// queries every candidate is the same instrument: one slot only.
	const (
		slotSameFund = iota // same base ticker, fund/ETF type
		slotSame            // same base ticker, anything else
		slotFund            // fund/ETF found by fuzzy search
		slotOther
		slotCount
	)
	var (
		failures []string
		series   [slotCount]*Series
		resols   [slotCount]resolution
	)
	consider := func(s *Series, res resolution, fund, sameBase bool) {
		i := slotOther
		switch {
		case preferBase == "": // ISIN: only depth matters
			i = slotSameFund
		case sameBase && fund:
			i = slotSameFund
		case sameBase:
			i = slotSame
		case fund:
			i = slotFund
		}
		if deeper(series[i], s) {
			series[i], resols[i] = s, res
		}
	}
	preferred := func() (*Series, resolution) {
		for i := range slotCount {
			if series[i] != nil {
				return series[i], resols[i]
			}
		}
		return nil, resolution{}
	}
	// covered reports whether the requested start date is essentially
	// reached: no other source could meaningfully improve on it.
	covered := func() bool {
		s, _ := preferred()
		return goodSeries(s) && !s.First().Date.After(from.AddDate(1, 0, 0))
	}
	matchesBase := func(symbol string) bool {
		return preferBase != "" &&
			(symbol == preferBase || strings.HasPrefix(symbol, preferBase+".") ||
				strings.HasPrefix(symbol, preferBase+":"))
	}

	quotes, err := c.search(ctx, query)
	if err != nil {
		failures = append(failures, fmt.Sprintf("yahoo: %v", err))
	}
	tried := map[string]bool{}
	for _, q := range rankQuotes(quotes, preferBase) {
		// ISIN listings are interchangeable: stop as soon as the start date
		// is covered. Ticker candidates are different instruments: examine
		// the whole (bounded) shortlist before settling.
		if len(tried) >= 4 || (preferBase == "" && covered()) {
			break
		}
		if tried[q.Symbol] {
			continue
		}
		tried[q.Symbol] = true
		s, herr := c.historyView(ctx, q.Symbol, from, raw)
		if herr != nil {
			failures = append(failures, fmt.Sprintf("yahoo %s: %v", q.Symbol, herr))
			continue
		}
		name := q.Name
		if name == "" {
			name = s.Name
		}
		isFund := q.QuoteType == "ETF" || q.QuoteType == "MUTUALFUND"
		consider(s, resolution{Source: "yahoo", Symbol: q.Symbol, Name: name}, isFund, matchesBase(q.Symbol))
	}
	if !covered() {
		if res, ferr := c.ftSearch(ctx, query); ferr == nil {
			if s, herr := c.historyFT(ctx, query, res, from, raw); herr == nil {
				consider(s, res, true, matchesBase(res.Symbol))
			} else {
				failures = append(failures, fmt.Sprintf("ft: %v", herr))
			}
		} else {
			failures = append(failures, fmt.Sprintf("ft: %v", ferr))
		}
	}
	if s, _ := preferred(); !goodSeries(s) {
		if msid, name, berr := c.boursoramaMorningstarID(ctx, query); berr == nil {
			res := resolution{Source: "morningstar", Symbol: msid, Name: name}
			if s, herr := c.historyMS(ctx, query, res, from, raw); herr == nil {
				consider(s, res, true, false)
			} else {
				failures = append(failures, fmt.Sprintf("morningstar %s: %v", msid, herr))
			}
		} else {
			failures = append(failures, fmt.Sprintf("boursorama: %v", berr))
		}
	}
	best, bestRes := preferred()
	return best, bestRes, failures
}

// rankQuotes orders Yahoo search candidates: listings of the searched ticker
// itself first, then fund and ETF entries (whose histories are usually the
// deepest), then plain exchange listings, which are sometimes moribund.
func rankQuotes(quotes []searchQuote, preferBase string) []searchQuote {
	sameTicker := func(symbol string) bool {
		return preferBase != "" &&
			(symbol == preferBase || strings.HasPrefix(symbol, preferBase+"."))
	}
	var prio, funds, others []searchQuote
	for _, q := range quotes {
		switch {
		case sameTicker(q.Symbol):
			prio = append(prio, q)
		case q.QuoteType == "MUTUALFUND" || q.QuoteType == "ETF":
			funds = append(funds, q)
		default:
			others = append(others, q)
		}
	}
	return append(append(prio, funds...), others...)
}

// historyForResolution fetches the history of an already-resolved ISIN from
// the source recorded in the resolution.
func (c *Client) historyForResolution(ctx context.Context, isin string, res resolution, from time.Time, raw bool) (*Series, error) {
	switch res.Source {
	case "ft":
		return c.historyFT(ctx, isin, res, from, raw)
	case "morningstar":
		return c.historyMS(ctx, isin, res, from, raw)
	case "stooq":
		return c.cachedHistory(ctx, "Stooq", isin, from, raw, func() (*Series, error) {
			return c.fetchStooq(ctx, res.Symbol, from)
		})
	default:
		s, err := c.historyView(ctx, res.Symbol, from, raw)
		// The curated resolution name beats source metadata (e.g. Yahoo
		// labels continuous futures with their front-month contract).
		if err == nil && res.Name != "" {
			s.Name = res.Name
		}
		return s, err
	}
}

func (c *Client) resolutionPath(isin string) string {
	return filepath.Join(c.CacheDir, "isin_"+sanitizeFilename(isin)+".json")
}

func (c *Client) loadResolution(isin string) (resolution, bool) {
	// Pinned catalog entries win over anything cached on disk.
	if res, ok := catalogResolution(isin); ok {
		return res, true
	}
	var res resolution
	if c.CacheDir == "" {
		return res, false
	}
	data, err := os.ReadFile(c.resolutionPath(isin))
	if err != nil || json.Unmarshal(data, &res) != nil {
		return res, false
	}
	if res.Source == "" {
		res.Source = "yahoo"
	}
	ok := false
	switch res.Source {
	case "yahoo", "morningstar":
		ok = res.Symbol != ""
	case "ft":
		ok = res.Xid != ""
	}
	return res, ok
}

func (c *Client) saveResolution(isin string, res resolution) {
	if data, err := json.MarshalIndent(res, "", "  "); err == nil {
		c.writeCacheFile(c.resolutionPath(isin), data)
	}
}

// History returns the daily history of a quotable symbol from `from` until
// today. Results come from the on-disk cache when fresh enough, otherwise
// from Yahoo Finance, with Stooq as a fallback source.
func (c *Client) History(ctx context.Context, symbol string, from time.Time) (*Series, error) {
	return c.historyView(ctx, symbol, from, false)
}

// historyView is History with the close-column choice. Raw series live
// under their own cache identity ("SYMBOL~raw"): the price-cleaning passes
// may drop or repair points differently per view, so the views never share
// a file.
func (c *Client) historyView(ctx context.Context, symbol string, from time.Time, raw bool) (*Series, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	key := viewKey(symbol, raw) + "|" + from.Format("2006-01-02")
	if s, ok := c.memoized(key); ok {
		return s, nil
	}
	s, err := c.history(ctx, symbol, from, raw)
	if err != nil {
		return nil, err
	}
	extendFXBack(symbol, s) // splice the bundled long EUR/USD behind the euro cross
	c.memoize(key, s)
	return s, nil
}

// viewKey is the cache and memoization identity of a series view.
func viewKey(id string, raw bool) string {
	if raw {
		return id + "~raw"
	}
	return id
}

func (c *Client) history(ctx context.Context, symbol string, from time.Time, raw bool) (*Series, error) {
	cacheID := viewKey(symbol, raw)
	if s, ok := c.loadCache(cacheID, from); ok {
		return s, nil
	}
	c.Logf("downloading %s…", symbol)
	s, yahooErr := c.fetchYahoo(ctx, symbol, from, raw)
	if yahooErr != nil {
		var stooqErr error
		s, stooqErr = c.fetchStooq(ctx, symbol, from)
		if stooqErr != nil {
			err := fmt.Errorf("downloading %s failed (yahoo: %v; stooq: %v)", symbol, yahooErr, stooqErr)
			return c.staleFallback(ctx, cacheID, from, err)
		}
		c.Logf("%s fetched via stooq (prices not dividend-adjusted)", symbol)
	}
	if len(s.Points) == 0 {
		return c.staleFallback(ctx, cacheID, from, fmt.Errorf("no quotes returned for %s", symbol))
	}
	if !isRateSymbol(symbol) {
		s.Points = dropDropouts(s.Points)   // strip provider placeholders/bad prints
		s.Points = mendScaleBreak(s.Points) // repair a single denomination break
	}
	c.saveCacheAs(cacheID, s, from)
	c.Logf("%s: %s, %d quotes since %s", s.Symbol, s.Name, len(s.Points), s.First().Date.Format("2006-01-02"))
	return s, nil
}

// staleFallback serves the expired cache of symbol when a refresh fails:
// charts then simply stop at the last cached date. The original error is
// returned only when no cache exists at all.
func (c *Client) staleFallback(ctx context.Context, symbol string, from time.Time, cause error) (*Series, error) {
	s, fetchedAt, ok := c.loadCacheAnyAge(symbol, from)
	if !ok {
		return nil, cause
	}
	c.Logf("warning: refreshing %s failed (%v), keeping cached data from %s (last quote %s)",
		symbol, cause, fetchedAt.Format("2006-01-02"), s.Last().Date.Format("2006-01-02"))
	return s, nil
}

// historyFT returns the daily history of an FT-resolved instrument, cached
// under its original identifier.
func (c *Client) historyFT(ctx context.Context, id string, res resolution, from time.Time, raw bool) (*Series, error) {
	return c.cachedHistory(ctx, "FT", id, from, raw, func() (*Series, error) {
		return c.fetchFT(ctx, id, res, from)
	})
}

// historyMS returns the daily history of a Morningstar-resolved fund, cached
// under its original identifier.
func (c *Client) historyMS(ctx context.Context, id string, res resolution, from time.Time, raw bool) (*Series, error) {
	return c.cachedHistory(ctx, "Morningstar", id, from, raw, func() (*Series, error) {
		return c.fetchMorningstar(ctx, id, res, from)
	})
}

// cachedHistory wraps a downloader with the memoization and on-disk cache
// shared by every non-Yahoo source, keyed by the original identifier.
func (c *Client) cachedHistory(ctx context.Context, source, id string, from time.Time, raw bool, fetch func() (*Series, error)) (*Series, error) {
	cacheID := viewKey(id, raw)
	key := source + ":" + cacheID + "|" + from.Format("2006-01-02")
	if s, ok := c.memoized(key); ok {
		return s, nil
	}
	if s, ok := c.loadCache(cacheID, from); ok {
		c.memoize(key, s)
		return s, nil
	}
	c.Logf("downloading %s via %s…", id, source)
	s, err := fetch()
	if err == nil && len(s.Points) == 0 {
		err = fmt.Errorf("no %s quotes for %s", source, id)
	}
	if err != nil {
		s, err = c.staleFallback(ctx, cacheID, from, err)
		if err != nil {
			return nil, err
		}
		c.memoize(key, s)
		return s, nil
	}
	c.saveCacheAs(cacheID, s, from)
	c.Logf("%s: %s, %d quotes since %s", s.Symbol, s.Name, len(s.Points), s.First().Date.Format("2006-01-02"))
	c.memoize(key, s)
	return s, nil
}

func (c *Client) memoized(key string) (*Series, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.memo[key]
	return s, ok
}

func (c *Client) memoize(key string, s *Series) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memo[key] = s
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	return c.do(ctx, http.MethodGet, rawURL, "", nil, nil)
}

func (c *Client) post(ctx context.Context, rawURL, contentType string, payload []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPost, rawURL, contentType, payload, nil)
}

// do performs an HTTP request with retries on transient failures; rate
// limiting (HTTP 429) backs off twice as long. A canceled context aborts
// both the in-flight request and the retry backoff.
func (c *Client) do(ctx context.Context, method, rawURL, contentType string, payload []byte, headers map[string]string) ([]byte, error) {
	var lastErr error
	rateLimited := false
	for attempt := range 3 {
		if attempt > 0 {
			delay := time.Duration(attempt) * c.retryDelay
			if rateLimited {
				delay *= 2
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		var reqBody io.Reader
		if payload != nil {
			reqBody = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, rawURL, reqBody)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json,text/csv,*/*")
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		switch {
		case resp.StatusCode == http.StatusOK:
			return body, nil
		case resp.StatusCode == http.StatusTooManyRequests:
			rateLimited = true
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		case resp.StatusCode >= 500:
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		default:
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
	}
	return nil, lastErr
}
