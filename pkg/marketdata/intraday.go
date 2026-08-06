package marketdata

import (
	"errors"
	"fmt"
	"time"
	_ "time/tzdata" // exchange time zones, without depending on the host OS
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
	// Match the ISIN shape directly rather than calling IsISIN: an ISIN-shaped
	// identifier must route to the resolution path even when its check digit is
	// invalid, so it is never mistaken for a ticker and fetched as one.
	if isinPattern.MatchString(canonical) {
		if res, ok := c.loadResolution(canonical); ok && res.Source == "yahoo" && res.Symbol != "" {
			return res.Symbol, true
		}
		return "", false
	}
	return canonical, true
}
