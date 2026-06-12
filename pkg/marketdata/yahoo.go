package marketdata

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"
)

// fetchYahoo downloads daily history from the Yahoo Finance chart API.
func (c *Client) fetchYahoo(symbol string, from time.Time) (*Series, error) {
	u := fmt.Sprintf("%s/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d&includeAdjustedClose=true",
		c.ChartBase, url.PathEscape(symbol), from.Unix(), time.Now().Add(24*time.Hour).Unix())
	body, err := c.get(u)
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
				Timestamp  []int64 `json:"timestamp"`
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
		return nil, fmt.Errorf("réponse yahoo illisible: %w", err)
	}
	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo: %s (%s)", resp.Chart.Error.Description, resp.Chart.Error.Code)
	}
	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("yahoo: réponse vide pour %s", symbol)
	}
	r := resp.Chart.Result[0]

	var closes []*float64
	switch {
	case len(r.Indicators.Adjclose) > 0 && len(r.Indicators.Adjclose[0].Adjclose) == len(r.Timestamp):
		closes = r.Indicators.Adjclose[0].Adjclose
	case len(r.Indicators.Quote) > 0 && len(r.Indicators.Quote[0].Close) == len(r.Timestamp):
		closes = r.Indicators.Quote[0].Close
	default:
		return nil, fmt.Errorf("yahoo: pas de série de clôtures pour %s", symbol)
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
	return s, nil
}

// searchQuote is one candidate instrument returned by the Yahoo search API.
type searchQuote struct {
	Symbol    string
	Name      string
	QuoteType string
}

// search queries the Yahoo Finance search API and returns every candidate
// symbol matching the query (typically an ISIN).
func (c *Client) search(query string) ([]searchQuote, error) {
	u := fmt.Sprintf("%s/v1/finance/search?q=%s&quotesCount=10&newsCount=0&listsCount=0",
		c.SearchBase, url.QueryEscape(query))
	body, err := c.get(u)
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
		return nil, fmt.Errorf("réponse de recherche illisible: %w", err)
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
		return nil, fmt.Errorf("aucun résultat pour %q", query)
	}
	return out, nil
}
