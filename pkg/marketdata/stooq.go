package marketdata

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// stooqSymbol maps a Yahoo-style symbol to its Stooq equivalent, or "" when
// there is no reliable mapping. invert reports that Stooq lists the
// reciprocal of the requested currency cross (it carries the conventional
// direction only, eurusd but no usdeur): the fetched closes must then be
// inverted to express the cross the caller asked for.
func stooqSymbol(symbol string) (ss string, invert bool) {
	switch symbol {
	case "^GSPC":
		return "^spx", false
	case "^NDX":
		return "^ndx", false
	case "^DJI":
		return "^dji", false
	case "^IXIC":
		return "^ndq", false
	case "XAUUSD":
		return "xauusd", false // gold spot, decades of history on stooq
	case "XAGUSD":
		return "xagusd", false
	case "CL=F", "CL.F":
		return "cl.f", false // WTI crude continuous futures
	}
	if base, quote, ok := fxCross(symbol); ok {
		return stooqFX(base, quote)
	}
	if strings.ContainsAny(symbol, "^=.") {
		return "", false // exchange suffixes, futures: no direct equivalent
	}
	return strings.ToLower(symbol) + ".us", false
}

// fxPrecedence ranks the major currencies by FX market convention: in a
// conventional cross the lower-ranked currency comes first (EURUSD, GBPUSD,
// USDJPY). Stooq lists exactly those conventional pairs.
var fxPrecedence = map[string]int{
	"EUR": 1, "GBP": 2, "AUD": 3, "NZD": 4, "USD": 5, "CAD": 6, "CHF": 7, "JPY": 8,
}

// stooqFX maps a currency cross to its conventional Stooq listing, inverted
// when the caller asked for the reciprocal direction. A minor currency (SEK,
// NOK, ...) ranks below every major, matching convention (USDSEK, EURSEK);
// a cross with no major leg is rejected, its Stooq coverage being unknown.
func stooqFX(base, quote string) (ss string, invert bool) {
	pb, okb := fxPrecedence[base]
	pq, okq := fxPrecedence[quote]
	if base == quote || (!okb && !okq) {
		return "", false
	}
	minor := len(fxPrecedence) + 1
	if !okb {
		pb = minor
	}
	if !okq {
		pq = minor
	}
	if pb < pq {
		return strings.ToLower(base + quote), false
	}
	return strings.ToLower(quote + base), true
}

// fxCross splits a Yahoo currency-cross symbol ("USDEUR=X") into its base
// and quote currencies. ok is false for any other symbol shape.
func fxCross(symbol string) (base, quote string, ok bool) {
	pair, found := strings.CutSuffix(symbol, "=X")
	if !found || len(pair) != 6 {
		return "", "", false
	}
	for _, r := range pair {
		if r < 'A' || r > 'Z' {
			return "", "", false
		}
	}
	return pair[:3], pair[3:], true
}

// euroInception is the euro's first trading year. Provider rows on a euro
// cross dated before it are synthetic backcasts of unknown provenance and
// are dropped: the bundled ECU/EUR anchors (FRED, see extendFXBack) cover
// that era instead.
var euroInception = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)

// fetchStooq downloads daily history from the stooq.com CSV endpoint, used as
// a fallback when Yahoo fails. Stooq closes are split-adjusted but not
// dividend-adjusted.
func (c *Client) fetchStooq(ctx context.Context, symbol string, from time.Time) (*Series, error) {
	ss, invert := stooqSymbol(symbol)
	if ss == "" {
		return nil, fmt.Errorf("no stooq equivalent for %s", symbol)
	}
	u := fmt.Sprintf("%s/q/d/l/?s=%s&i=d&d1=%s&d2=%s", c.StooqBase, url.QueryEscape(ss),
		from.Format("20060102"), time.Now().Format("20060102"))
	body, err := c.get(ctx, u)
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
	base, quote, isFX := fxCross(symbol)
	if isFX {
		s.Currency = quote // a cross is quoted in its second currency
		s.Name = base + "/" + quote + " spot rate"
	}
	for _, row := range rows[1:] {
		if len(row) < 5 {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02", row[0], time.UTC)
		if err != nil {
			continue
		}
		if isFX && (base == "EUR" || quote == "EUR") && t.Before(euroInception) {
			continue // pre-euro rows are synthetic, see euroInception
		}
		cl, err := strconv.ParseFloat(row[4], 64)
		if err != nil || cl <= 0 {
			continue
		}
		if invert {
			cl = 1 / cl
		}
		s.Points = append(s.Points, Point{Date: t, Close: cl})
	}
	return s, nil
}
