package marketdata

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// morningstarToken is the view id embedded in Morningstar's public chart
// pages; it has been stable for years.
const morningstarToken = "ok91jeenoo"

// fetchMorningstar downloads a daily NAV series from the Morningstar
// timeseries API using a Morningstar fund id (0P…, found via Boursorama).
// NAVs are expressed in the share class currency, which the API does not
// disclose, so Currency stays as recorded in the resolution (often empty).
func (c *Client) fetchMorningstar(id string, res resolution, from time.Time) (*Series, error) {
	u := fmt.Sprintf("%s/api/rest.svc/timeseries_price/%s?id=%s&idtype=Morningstar&frequency=daily&startDate=%s&outputType=COMPACTJSON",
		c.MorningstarBase, morningstarToken, url.QueryEscape(res.Symbol), from.Format("2006-01-02"))
	body, err := c.get(u)
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
