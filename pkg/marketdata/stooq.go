package marketdata

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// stooqSymbol maps a Yahoo-style symbol to its Stooq equivalent, or "" when
// there is no reliable mapping.
func stooqSymbol(symbol string) string {
	switch symbol {
	case "^GSPC":
		return "^spx"
	case "^NDX":
		return "^ndx"
	case "^DJI":
		return "^dji"
	case "^IXIC":
		return "^ndq"
	case "XAUUSD":
		return "xauusd" // gold spot, decades of history on stooq
	case "XAGUSD":
		return "xagusd"
	case "CL=F", "CL.F":
		return "cl.f" // WTI crude continuous futures
	}
	if strings.ContainsAny(symbol, "^=.") {
		return "" // exchange suffixes, futures, FX pairs: no direct equivalent
	}
	return strings.ToLower(symbol) + ".us"
}

// fetchStooq downloads daily history from the stooq.com CSV endpoint, used as
// a fallback when Yahoo fails. Stooq closes are split-adjusted but not
// dividend-adjusted.
func (c *Client) fetchStooq(symbol string, from time.Time) (*Series, error) {
	ss := stooqSymbol(symbol)
	if ss == "" {
		return nil, fmt.Errorf("no stooq equivalent for %s", symbol)
	}
	u := fmt.Sprintf("%s/q/d/l/?s=%s&i=d&d1=%s&d2=%s", c.StooqBase, url.QueryEscape(ss),
		from.Format("20060102"), time.Now().Format("20060102"))
	body, err := c.get(u)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(body, []byte("Date,")) {
		return nil, fmt.Errorf("stooq: no data for %s", ss)
	}
	rows, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unreadable stooq CSV: %w", err)
	}
	if len(rows) < 2 || len(rows[0]) < 5 {
		return nil, fmt.Errorf("stooq: no data for %s", ss)
	}
	s := &Series{Symbol: symbol, Name: symbol, Source: "stooq"}
	switch {
	case strings.HasSuffix(ss, ".us"), ss == "xauusd", ss == "xagusd", ss == "cl.f":
		s.Currency = "USD"
	}
	switch ss {
	case "xauusd":
		s.Name = "Gold (XAU/USD spot)"
	case "xagusd":
		s.Name = "Silver (XAG/USD spot)"
	case "cl.f":
		s.Name = "WTI crude oil (continuous futures)"
	}
	for _, row := range rows[1:] {
		if len(row) < 5 {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02", row[0], time.UTC)
		if err != nil {
			continue
		}
		cl, err := strconv.ParseFloat(row[4], 64)
		if err != nil || cl <= 0 {
			continue
		}
		s.Points = append(s.Points, Point{Date: t, Close: cl})
	}
	return s, nil
}
