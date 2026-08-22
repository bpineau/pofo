package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// morningstarToken is the view id embedded in Morningstar's public chart
// pages; it has been stable for years.
const morningstarToken = "ok91jeenoo"

// morningstarScreenerToken is the equivalent view id for the screener the
// fund-search pages call. It is a second, independent token: the timeseries
// one is not accepted there.
const morningstarScreenerToken = "klr5zyak8x"

// morningstarUniverses are the two screener universes worth searching: open
// end funds and exchange traded ones. They are searched together because a
// share class sits in exactly one of them and the caller never knows which.
const morningstarUniverses = "FOALL$$ALL|ETALL$$ALL"

// morningstarIDSuffix disambiguates the identifier for the timeseries API.
// The service reads the id as a bracket-separated tuple and answers an EMPTY
// array, with HTTP 200, for a bare exchange-traded id; the same id with the
// suffix returns the full history. Only the presence of the two trailing
// fields matters, not their content: a bare fund id happens to work, an
// exchange-traded one does not, and appending this to either is harmless.
// Losing that suffix is what silently emptied this source (see
// docs/trend-reconstruction-design.md).
const morningstarIDSuffix = "]2]1]"

// fetchMorningstar downloads a daily NAV series from the Morningstar
// timeseries API using a Morningstar security id (a fund id such as
// F0GBR05B68, or an exchange-traded one such as 0P0001BOL6, found via
// morningstarSearch or Boursorama).
//
// The NAVs are a PRICE return: what a distributing share class pays out is
// missing from them, exactly as it is from the FT NAVs (see the package
// documentation). The currency the API does not disclose either, so Currency
// stays as recorded in the resolution, which morningstarSearch does fill.
func (c *Client) fetchMorningstar(ctx context.Context, id string, res resolution, from time.Time) (*Series, error) {
	u := fmt.Sprintf("%s/api/rest.svc/timeseries_price/%s?id=%s&idtype=Morningstar&frequency=daily&startDate=%s&outputType=COMPACTJSON",
		c.MorningstarBase, morningstarToken,
		url.QueryEscape(res.Symbol+morningstarIDSuffix), from.Format("2006-01-02"))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	// COMPACTJSON: [[epoch_ms, value], …]. Errors come back as XML.
	var rows [][]float64
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("unreadable morningstar response for %s", res.Symbol)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("morningstar: no data for %s", res.Symbol)
	}
	name := res.Name
	if name == "" {
		name = id
	}
	s := &Series{Symbol: id, Name: name, Currency: res.Currency, Source: "morningstar"}
	for _, row := range rows {
		if len(row) < 2 || row[1] <= 0 {
			continue
		}
		day := dayUTC(time.UnixMilli(int64(row[0])).UTC())
		if day.Before(from) {
			continue
		}
		if n := len(s.Points); n > 0 && !s.Points[n-1].Date.Before(day) {
			continue
		}
		s.Points = append(s.Points, Point{Date: day, Close: row[1]})
	}
	return s, nil
}

// morningstarRow is one hit of the screener response.
type morningstarRow struct {
	SecID    string `json:"SecId"`
	Name     string `json:"Name"`
	ISIN     string `json:"isin"`
	Currency string `json:"currencyId"`
}

// morningstarSearch resolves an identifier through Morningstar's own fund
// screener, the search behind its public fund pages. It covers open-end and
// exchange-traded classes worldwide, which is wider than the French
// distribution list Boursorama indexes, and it reports the quote currency
// that the timeseries API withholds.
//
// A row whose ISIN is the query itself wins over a mere full-text hit, so an
// ISIN lookup never adopts a same-family class quoted in another currency.
func (c *Client) morningstarSearch(ctx context.Context, query string) (resolution, error) {
	u := fmt.Sprintf("%s/api/rest.svc/%s/security/screener?page=1&pageSize=10&outputType=json&universeIds=%s&securityDataPoints=%s&term=%s",
		c.MorningstarBase, morningstarScreenerToken,
		url.QueryEscape(morningstarUniverses),
		url.QueryEscape("SecId|Name|isin|currencyId"),
		url.QueryEscape(query))
	body, err := c.get(ctx, u)
	if err != nil {
		return resolution{}, err
	}
	var page struct {
		Rows []morningstarRow `json:"rows"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		return resolution{}, fmt.Errorf("unreadable morningstar screener response for %s", query)
	}
	best := -1
	for i, row := range page.Rows {
		if row.SecID == "" {
			continue
		}
		if strings.EqualFold(row.ISIN, query) {
			best = i
			break
		}
		if best < 0 {
			best = i
		}
	}
	if best < 0 {
		return resolution{}, fmt.Errorf("no Morningstar identifier for %s", query)
	}
	row := page.Rows[best]
	return resolution{
		Source:   "morningstar",
		Symbol:   row.SecID,
		Name:     strings.TrimSpace(row.Name),
		Currency: morningstarCurrency(row.Currency),
	}, nil
}

// morningstarCurrency extracts the ISO code from the screener's padded
// currency id ("CU$$$$$USD" -> "USD"). Anything else yields no currency,
// which the fetch path reads as "unknown" rather than as a mismatch.
func morningstarCurrency(id string) string {
	if i := strings.LastIndexByte(id, '$'); i >= 0 {
		id = id[i+1:]
	}
	if len(id) != 3 {
		return ""
	}
	return strings.ToUpper(id)
}
