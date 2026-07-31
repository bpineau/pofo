package marketdata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// yahooQuoteBatchMax bounds one /v7/finance/quote call. Yahoo accepts far
// more; chunking just keeps URLs short.
const yahooQuoteBatchMax = 50

// LatestBatch returns the freshest available price of each id: one batched
// Yahoo quote call for the Yahoo-quoted ids, then the per-id Latest fallback
// (FT/Morningstar NAV, Stooq, cache…) for the rest and for anything the
// batch did not return. Ids no source can serve are absent from the result;
// the call never fails as a whole. Like Latest, a closed market still yields
// a Live quote: the regular session's last price, timed at the close.
func (c *Client) LatestBatch(ctx context.Context, ids []string) map[string]Quote {
	out := make(map[string]Quote, len(ids))
	symbols := make([]string, 0, len(ids))
	bySymbol := make(map[string][]string, len(ids)) // yahoo symbol → original ids
	rest := make([]string, 0, len(ids))
	for _, id := range ids {
		base, _ := SplitSim(id)
		symbol, ok := c.yahooSymbol(ctx, base)
		if !ok {
			rest = append(rest, id)
			continue
		}
		if len(bySymbol[symbol]) == 0 {
			symbols = append(symbols, symbol)
		}
		bySymbol[symbol] = append(bySymbol[symbol], id)
	}
	quotes := c.fetchYahooQuoteBatch(ctx, symbols)
	for symbol, ids := range bySymbol {
		q, ok := quotes[symbol]
		if !ok {
			rest = append(rest, ids...)
			continue
		}
		for _, id := range ids {
			out[id] = q
		}
	}
	for _, id := range rest {
		if q, err := c.Latest(ctx, id); err == nil {
			out[id] = *q
		} else {
			c.Logf("latest batch: %s: %v", id, err)
		}
	}
	return out
}

// fetchYahooQuoteBatch reads live regular-market prices for many symbols in
// yahooQuoteBatchMax-sized chunks of the v7 quote API (cookie+crumb needed).
func (c *Client) fetchYahooQuoteBatch(ctx context.Context, symbols []string) map[string]Quote {
	out := make(map[string]Quote, len(symbols))
	for start := 0; start < len(symbols); start += yahooQuoteBatchMax {
		c.quoteBatchChunk(ctx, symbols[start:min(start+yahooQuoteBatchMax, len(symbols))], out)
	}
	return out
}

// quoteBatchChunk fetches one chunk into out, renewing a stale cookie+crumb
// pair once. Failures degrade to a log line: the caller's per-id fallback
// picks the missing symbols up.
func (c *Client) quoteBatchChunk(ctx context.Context, symbols []string, out map[string]Quote) {
	if len(symbols) == 0 {
		return
	}
	auth, err := c.yahooAuthPair(ctx)
	if err != nil {
		c.Logf("yahoo quote batch: %v", err)
		return
	}
	body, err := c.quoteBatchGet(ctx, symbols, auth)
	if isYahooAuthErr(err) { // stale crumb: renew once and retry
		c.invalidateYahooAuth()
		if auth, err = c.yahooAuthPair(ctx); err == nil {
			body, err = c.quoteBatchGet(ctx, symbols, auth)
		}
	}
	if err != nil {
		c.Logf("yahoo quote batch: %v", err)
		return
	}
	var resp struct {
		QuoteResponse struct {
			Result []struct {
				Symbol               string   `json:"symbol"`
				Currency             string   `json:"currency"`
				ExchangeTimezoneName string   `json:"exchangeTimezoneName"`
				RegularMarketPrice   *float64 `json:"regularMarketPrice"`
				RegularMarketTime    int64    `json:"regularMarketTime"`
			} `json:"result"`
		} `json:"quoteResponse"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		c.Logf("yahoo quote batch: unreadable response: %v", err)
		return
	}
	for _, r := range resp.QuoteResponse.Result {
		if r.RegularMarketPrice == nil || *r.RegularMarketPrice <= 0 {
			continue
		}
		loc, err := time.LoadLocation(r.ExchangeTimezoneName)
		if err != nil {
			loc = time.UTC
		}
		out[r.Symbol] = Quote{
			Price:    *r.RegularMarketPrice,
			Time:     time.Unix(r.RegularMarketTime, 0).In(loc),
			Currency: r.Currency,
			Source:   "yahoo",
			Live:     true,
		}
	}
}

// quoteBatchGet performs the authenticated v7 quote request for one chunk.
func (c *Client) quoteBatchGet(ctx context.Context, symbols []string, auth yahooAuth) ([]byte, error) {
	path := "/v7/finance/quote?symbols=" + url.QueryEscape(strings.Join(symbols, ",")) +
		"&crumb=" + url.QueryEscape(auth.crumb)
	return c.do(ctx, http.MethodGet, c.ChartBase+path, "", nil, map[string]string{"Cookie": auth.cookie})
}

// isYahooAuthErr matches the "HTTP 401"/"HTTP 403" errors do() produces when
// the crumb or cookie has expired.
func isYahooAuthErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "HTTP 401") || strings.Contains(err.Error(), "HTTP 403")
}
