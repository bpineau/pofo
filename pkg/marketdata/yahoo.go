package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"
)

// Yahoo serves the same data from two interchangeable API hosts, rate-limited
// independently: when one returns 429 the other usually still answers.
const (
	yahooHost1 = "https://query1.finance.yahoo.com"
	yahooHost2 = "https://query2.finance.yahoo.com"
)

// yahooGet fetches an absolute Yahoo API path (e.g. "/v8/finance/chart/AAPL?…")
// from base, retrying on the twin query1/query2 host so a per-host rate limit
// does not fail the run. A non-Yahoo base (a test server) is used on its own.
func (c *Client) yahooGet(ctx context.Context, base, path string) ([]byte, error) {
	hosts := []string{base}
	switch base {
	case yahooHost1:
		hosts = append(hosts, yahooHost2)
	case yahooHost2:
		hosts = append(hosts, yahooHost1)
	}
	var body []byte
	var err error
	for _, h := range hosts {
		if body, err = c.get(ctx, h+path); err == nil {
			return body, nil
		}
	}
	return body, err
}

// fetchYahoo downloads daily history from the Yahoo Finance chart API,
// dividend events included. With raw it serves the unadjusted close column
// (split-adjusted but not dividend-adjusted) instead of the default
// adjusted one.
func (c *Client) fetchYahoo(ctx context.Context, symbol string, from time.Time, raw bool) (*Series, error) {
	path := fmt.Sprintf("/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d&includeAdjustedClose=true&events=div",
		url.PathEscape(symbol), from.Unix(), time.Now().Add(24*time.Hour).Unix())
	body, err := c.yahooGet(ctx, c.ChartBase, path)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency  string `json:"currency"`
					Symbol    string `json:"symbol"`
					LongName  string `json:"longName"`
					ShortName string `json:"shortName"`
				} `json:"meta"`
				Timestamp []int64 `json:"timestamp"`
				Events    struct {
					Dividends map[string]struct {
						Amount float64 `json:"amount"`
						Date   int64   `json:"date"`
					} `json:"dividends"`
				} `json:"events"`
				Indicators struct {
					Quote []struct {
						Close []*float64 `json:"close"`
					} `json:"quote"`
					Adjclose []struct {
						Adjclose []*float64 `json:"adjclose"`
					} `json:"adjclose"`
				} `json:"indicators"`
			} `json:"result"`
			Error *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable yahoo response: %w", err)
	}
	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo: %s (%s)", resp.Chart.Error.Description, resp.Chart.Error.Code)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("yahoo: empty response for %s", symbol)
	}
	r := resp.Chart.Result[0]

	var closes []*float64
	switch {
	case !raw && len(r.Indicators.Adjclose) > 0 && len(r.Indicators.Adjclose[0].Adjclose) == len(r.Timestamp):
		closes = r.Indicators.Adjclose[0].Adjclose
	case len(r.Indicators.Quote) > 0 && len(r.Indicators.Quote[0].Close) == len(r.Timestamp):
		closes = r.Indicators.Quote[0].Close
	default:
		return nil, fmt.Errorf("yahoo: no close series for %s", symbol)
	}

	name := r.Meta.LongName
	if name == "" {
		name = r.Meta.ShortName
	}
	if name == "" {
		name = symbol
	}
	s := &Series{Symbol: symbol, Name: name, Currency: r.Meta.Currency, Source: "yahoo"}
	for i, ts := range r.Timestamp {
		cl := closes[i]
		if cl == nil || *cl <= 0 {
			continue
		}
		day := dayUTC(time.Unix(ts, 0).UTC())
		// Yahoo sometimes repeats the current day; keep the latest value.
		if n := len(s.Points); n > 0 && s.Points[n-1].Date.Equal(day) {
			s.Points[n-1].Close = *cl
			continue
		}
		s.Points = append(s.Points, Point{Date: day, Close: *cl})
	}
	sort.Slice(s.Points, func(i, j int) bool { return s.Points[i].Date.Before(s.Points[j].Date) })
	for _, ev := range r.Events.Dividends {
		if ev.Amount <= 0 {
			continue
		}
		s.Dividends = append(s.Dividends, Dividend{Date: dayUTC(time.Unix(ev.Date, 0).UTC()), Amount: ev.Amount})
	}
	sort.Slice(s.Dividends, func(i, j int) bool { return s.Dividends[i].Date.Before(s.Dividends[j].Date) })
	return s, nil
}

// fetchYahooIntraday downloads the current day's 5-minute price path from the
// Yahoo Finance chart API. It returns ErrNotCovered when Yahoo serves no
// intraday result for the symbol.
func (c *Client) fetchYahooIntraday(ctx context.Context, symbol string) (*IntradaySeries, error) {
	path := fmt.Sprintf("/v8/finance/chart/%s?interval=5m&range=1d", url.PathEscape(symbol))
	body, err := c.yahooGet(ctx, c.ChartBase, path)
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

// searchQuote is one candidate instrument returned by the Yahoo search API.
type searchQuote struct {
	Symbol    string
	Name      string
	QuoteType string
}

// search queries the Yahoo Finance search API and returns every candidate
// symbol matching the query (typically an ISIN).
func (c *Client) search(ctx context.Context, query string) ([]searchQuote, error) {
	path := fmt.Sprintf("/v1/finance/search?q=%s&quotesCount=10&newsCount=0&listsCount=0",
		url.QueryEscape(query))
	body, err := c.yahooGet(ctx, c.SearchBase, path)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Quotes []struct {
			Symbol    string `json:"symbol"`
			ShortName string `json:"shortname"`
			LongName  string `json:"longname"`
			QuoteType string `json:"quoteType"`
		} `json:"quotes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable search response: %w", err)
	}
	var out []searchQuote
	for _, q := range resp.Quotes {
		if q.Symbol == "" {
			continue
		}
		name := q.LongName
		if name == "" {
			name = q.ShortName
		}
		out = append(out, searchQuote{Symbol: q.Symbol, Name: name, QuoteType: q.QuoteType})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no results for %q", query)
	}
	return out, nil
}
