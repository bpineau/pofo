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
type Quote struct {
	Price    float64   // in Currency
	Time     time.Time // when this price was observed
	Currency string    // ISO 4217 quote currency
	Source   string    // "yahoo", "ft", "morningstar" or "stooq"
	Live     bool      // true: real-time market field; false: last daily close
}

// latestFrom is the history window Latest fetches over when it falls back to
// the last daily close. One year is deep enough to always contain a recent
// close, even across long market closures or for an illiquid instrument, and
// both the disk cache and the in-process memoization key the window at day
// granularity, so repeated Latest calls reuse one cache entry.
func latestFrom() time.Time { return time.Now().AddDate(-1, 0, 0) }

// Latest returns the freshest available price for an identifier: the live
// Yahoo market price when the instrument is Yahoo-quoted, otherwise the last
// daily close (FT or Morningstar NAV), served from the on-disk cache when
// fresh and from stale data on a failed refresh. A "SIM" suffix is ignored
// (see SplitSim): simulated history never changes the current price.
//
// Like Intraday, the live path is stateless: Latest performs no caching of the
// live price, so a caller valuing a portfolio repeatedly should keep its own
// short-TTL cache. The daily-close fallback path uses the existing on-disk
// daily cache.
func (c *Client) Latest(ctx context.Context, id string) (*Quote, error) {
	base, _ := SplitSim(id)
	if symbol, ok := c.yahooSymbol(ctx, base); ok {
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
	return &Quote{
		Price:    last.Close,
		Time:     last.Date,
		Currency: s.Currency,
		Source:   s.Source,
		Live:     false,
	}, nil
}
