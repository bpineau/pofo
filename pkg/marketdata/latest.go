package marketdata

import (
	"context"
	"fmt"
	"time"
)

// Quote is the most recent known price of an instrument.
//
// Live reports how fresh the price is: true means a real-time market field
// (Yahoo regularMarketPrice), whose Time is an intraday instant; false means
// the last daily close, whose Time is that close's date. A market that is
// closed still yields a Live quote, the regular session's last price, with its
// Time at the close.
//
// Session names the trading session Price was struck in, so a caller can label
// an off-hours print rather than pass it off as a close. Only the
// extended-hours opt-in (QuoteOptions.ExtendedHours, LatestBatchExtended) ever
// produces "pre" or "post"; every default call returns "regular" or "".
type Quote struct {
	Price    float64   // in Currency
	Time     time.Time // when this price was observed
	Currency string    // ISO 4217 quote currency
	Symbol   string    // the instrument actually served (a resolution may pick a twin listing)
	Source   string    // "yahoo", "ft", "morningstar", "stooq", "ecb", "airfund", or "nowcast" for an estimate (see nowcast.go)
	Live     bool      // true: real-time market field; false: last daily close
	Session  string    // "regular", "pre", "post", or "" when the source names no session (a daily close, a fund NAV, a nowcast)
}

// latestFrom is the history window Latest fetches over when it falls back to
// the last daily close. One year is deep enough to always contain a recent
// close, even across long market closures or for an illiquid instrument, and
// both the disk cache and the in-process memoization key the window at day
// granularity, so repeated Latest calls reuse one cache entry.
func latestFrom() time.Time { return time.Now().AddDate(-1, 0, 0) }

// Latest returns the freshest available price for an identifier: the live
// Yahoo market price when the instrument is Yahoo-quoted, otherwise the last
// daily close, which for an FT or Morningstar fund is its latest NAV. A "SIM"
// suffix is ignored (see SplitSim): simulated history never changes the
// current price.
//
// Latest degrades gracefully rather than failing. A spot request Yahoo cannot
// serve (outage, throttling past the built-in retries and the query1/query2
// host fallback, missing field) falls through to the daily-close path, which
// inherits the whole Fetch resilience: the Stooq fallback for US tickers,
// major indices and major currency crosses, the ECB reference rates for a
// currency cross, CBOE for ^VIX, re-resolution through the Financial Times
// and Morningstar,
// and the on-disk cache, whose stale data still answers when every source is
// unreachable. Quote.Live, Quote.Time and Quote.Source report what the caller
// actually got.
//
// Like Intraday, the live path is stateless: Latest performs no caching of the
// live price, so a caller valuing a portfolio repeatedly should keep its own
// short-TTL cache. The daily-close fallback path uses the existing on-disk
// daily cache. To express the price in a display currency, pair Latest with
// FXRate (ConvertCurrency is its whole-series sibling).
func (c *Client) Latest(ctx context.Context, id string) (*Quote, error) {
	return c.latest(ctx, id, false)
}

// latest is Latest with the extended-hours switch QuoteOptions.ExtendedHours
// and LatestBatchExtended flip. Only the live Yahoo leg reads it: a fund NAV,
// a daily close and a nowcast have no session to report.
func (c *Client) latest(ctx context.Context, id string, extended bool) (*Quote, error) {
	base, _ := SplitSim(id)
	// A fund priced once a day with a lag quotes its nowcast: the last tick
	// of the proxy-scaled intraday path, live like any market print.
	if e, ok := catalogByID()[CanonicalID(base)]; ok && e.NowcastProxy != "" {
		if s, err := c.nowcastIntraday(ctx, CanonicalID(base), e); err == nil && len(s.Points) > 0 {
			last := s.Last()
			return &Quote{Price: last.Close, Time: last.Time, Currency: s.Currency, Symbol: s.Symbol, Source: "nowcast", Live: true}, nil
		}
	}
	if symbol, ok := c.yahooSymbol(ctx, base); ok {
		if extended {
			if q, ok := c.fetchYahooSpotExtended(ctx, symbol); ok {
				return q, nil
			}
			// The extended leg needs Yahoo's cookie+crumb pair and can fail
			// where the plain chart call succeeds: fall through to it rather
			// than lose the quote over the off-hours bonus.
		}
		if q, err := c.fetchYahooSpot(ctx, symbol); err == nil {
			return q, nil
		}
		// Spot unavailable (not covered, throttled, or missing field): fall
		// through to the last daily close.
	}
	s, err := c.Fetch(ctx, base, latestFrom())
	if err != nil {
		return nil, err
	}
	last := s.Last()
	if last.Date.IsZero() {
		return nil, fmt.Errorf("%s: no recent quote", id)
	}
	source := s.Source
	if !s.EstimatedFrom.IsZero() && !last.Date.Before(s.EstimatedFrom) {
		source = "nowcast" // the forward estimate, not a published close
	}
	return &Quote{
		Price:    last.Close,
		Time:     last.Date,
		Currency: s.Currency,
		Symbol:   s.Symbol,
		Source:   source,
		Live:     false,
	}, nil
}
