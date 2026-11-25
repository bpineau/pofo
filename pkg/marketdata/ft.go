package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

// ftSearch resolves an identifier (typically an ISIN) through the Financial
// Times securities search. FT covers many European mutual funds that Yahoo
// does not list. An ISIN-shaped query only accepts a hit whose own symbol
// carries that ISIN, so an unknown code falls through to the other sources
// instead of adopting whatever the full-text search felt closest to.
func (c *Client) ftSearch(ctx context.Context, query string) (resolution, error) {
	u := fmt.Sprintf("%s/data/searchapi/searchsecurities?query=%s", c.FTBase, url.QueryEscape(query))
	body, err := c.get(ctx, u)
	if err != nil {
		return resolution{}, err
	}
	var resp struct {
		Data struct {
			Security []struct {
				Name      string `json:"name"`
				Symbol    string `json:"symbol"` // "LU0171310443:EUR" or "NTSG:GER:EUR"
				Xid       string `json:"xid"`
				IsPrimary bool   `json:"isPrimary"`
			} `json:"security"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return resolution{}, fmt.Errorf("unreadable FT search response: %w", err)
	}
	secs := resp.Data.Security
	// An ISIN query is answered by identifier, not by name, so the hit must
	// carry the code it was asked about. FT's search is full-text and always
	// answers something: an ISIN no listing carries came back as an unrelated
	// stock (AT0000979794 returned "Adeia Inc.", symbol 0M2A:LSE), whose prices
	// would then have been served, and cached, as the fund's own.
	if q := strings.ToUpper(query); isinPattern.MatchString(q) {
		kept := secs[:0]
		for _, s := range secs {
			if strings.Contains(strings.ToUpper(s.Symbol), q) {
				kept = append(kept, s)
			}
		}
		secs = kept
	}
	// Prefer a listing not quoted in pence (GBX); fall back to any match.
	best := -1
	for i, s := range secs {
		if s.Xid == "" || ftSymbolCurrency(s.Symbol) == "GBX" {
			continue
		}
		best = i
		break
	}
	if best < 0 {
		for i, s := range secs {
			if s.Xid != "" {
				best = i
				break
			}
		}
	}
	if best < 0 {
		return resolution{}, markAbsent(fmt.Errorf("no FT results for %q", query))
	}
	sec := secs[best]
	base, _, _ := strings.Cut(sec.Symbol, ":")
	return resolution{Source: "ft", Symbol: base, Xid: sec.Xid, Name: sec.Name, Currency: ftSymbolCurrency(sec.Symbol)}, nil
}

// ftSymbolCurrency extracts the currency, the last segment of an FT symbol
// like "LU0171310443:EUR" or "NTSG:GER:EUR". Some FT symbols end on an
// exchange code instead ("WEBN:MUN" is the Munich listing, currency
// unstated), so the segment only counts when it is a known currency code;
// callers fall back to the currency the FT chart API reports.
func ftSymbolCurrency(symbol string) string {
	parts := strings.Split(symbol, ":")
	if len(parts) < 2 {
		return ""
	}
	last := parts[len(parts)-1]
	if !ftCurrencies[last] {
		return ""
	}
	return last
}

// ftCurrencies are the quote currencies observed on FT listings; anything
// else in the last symbol segment is an exchange code, not a currency.
var ftCurrencies = map[string]bool{
	"EUR": true, "USD": true, "GBP": true, "GBX": true, "GBp": true,
	"CHF": true, "JPY": true, "SEK": true, "NOK": true, "DKK": true,
	"CAD": true, "AUD": true, "NZD": true, "SGD": true, "HKD": true,
	"PLN": true, "CZK": true, "HUF": true, "ILS": true, "ZAR": true,
}

// fetchFT downloads a daily price series (fund NAVs, mostly) from the FT
// chart API. The series is keyed by the original identifier (the ISIN).
func (c *Client) fetchFT(ctx context.Context, id string, res resolution, from time.Time) (*Series, error) {
	days := max(int(time.Since(from).Hours()/24)+2, 2)
	payload, err := json.Marshal(map[string]any{
		"days":              days,
		"dataPeriod":        "Day",
		"dataInterval":      1,
		"timeServiceFormat": "JSON",
		"returnDateType":    "ISO8601",
		"elements":          []map[string]any{{"Type": "price", "Symbol": res.Xid}},
	})
	if err != nil {
		return nil, err
	}
	body, err := c.post(ctx, c.FTBase+"/data/chartapi/series", "application/json", payload)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Dates    []string `json:"Dates"`
		Elements []struct {
			Currency        string `json:"Currency"`
			ComponentSeries []struct {
				Type   string     `json:"Type"`
				Values []*float64 `json:"Values"`
			} `json:"ComponentSeries"`
		} `json:"Elements"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable FT response: %w", err)
	}
	if len(resp.Elements) == 0 {
		return nil, markAbsent(fmt.Errorf("FT: empty response for %s", id))
	}
	var closes []*float64
	for _, cs := range resp.Elements[0].ComponentSeries {
		if cs.Type == "Close" {
			closes = cs.Values
			break
		}
	}
	if closes == nil || len(closes) != len(resp.Dates) {
		return nil, fmt.Errorf("FT: no close series for %s", id)
	}
	currency := resp.Elements[0].Currency
	if currency == "" {
		currency = res.Currency
	}
	name := res.Name
	if name == "" {
		name = id
	}
	s := &Series{Symbol: id, Name: name, Currency: currency, Source: "ft"}
	for i, d := range resp.Dates {
		cl := closes[i]
		if cl == nil || *cl <= 0 {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02T15:04:05", d, time.UTC)
		if err != nil {
			continue
		}
		day := dayUTC(t)
		if day.Before(from) {
			continue
		}
		s.Points = append(s.Points, Point{Date: day, Close: *cl})
	}
	restoreFTCatBondQuote(s)
	normalizeUnits(s) // FT spells the London pence line "GBX"
	return s, nil
}

// FT dropped one October 2017 NAV from each of these two EUR classes in
// November 2026. Both quotes were in its earlier feed. IFM's NAV export for
// LI0049587301 confirms 1186.49 and Morningstar F00000NMKO confirms 12.566
// (its three-decimal rounding of FT's 12.5657). Only fill the original gap
// on its original price scale. A later provider revision takes precedence.
func restoreFTCatBondQuote(s *Series) {
	if s.Currency != "EUR" {
		return
	}
	var day, nextDay int
	var previous, close, next float64
	switch s.Symbol {
	case "IE00B3Q8M574":
		day, nextDay, previous, close, next = 9, 16, 12.5132, 12.5657, 12.6723
	case "LI0049587301":
		day, nextDay, previous, close, next = 13, 31, 1181.13, 1186.49, 1185.29
	default:
		return
	}
	before := time.Date(2017, 9, 29, 0, 0, 0, 0, time.UTC)
	after := time.Date(2017, 10, nextDay, 0, 0, 0, 0, time.UTC)
	for i, p := range s.Points {
		if i > 0 && p.Date.Equal(after) && p.Close == next &&
			s.Points[i-1].Date.Equal(before) && s.Points[i-1].Close == previous {
			missing := Point{Date: time.Date(2017, 10, day, 0, 0, 0, 0, time.UTC), Close: close}
			s.Points = slices.Insert(s.Points, i, missing)
			return
		}
	}
}
