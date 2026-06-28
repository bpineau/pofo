package marketdata

import "time"

// dayUTC truncates a time to its civil date at 00:00 UTC, the invariant
// every Point.Date in this package respects.
func dayUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// Point is one daily observation of an asset price.
type Point struct {
	Date  time.Time // normalized to 00:00 UTC
	Close float64   // adjusted close (dividends and splits reinvested when available)
}

// Series is the price history of one asset, sorted by ascending date.
type Series struct {
	Symbol   string
	Name     string
	Currency string
	Source   string // "yahoo", "stooq", "ft", "morningstar" or "simdata"
	Points   []Point

	// SimulatedBefore is non-zero when points before that date were
	// reconstructed from ProxySymbol instead of actual quotes.
	SimulatedBefore time.Time
	ProxySymbol     string
}

// First returns the earliest point, or the zero Point if the series is empty.
func (s *Series) First() Point {
	if len(s.Points) == 0 {
		return Point{}
	}
	return s.Points[0]
}

// Last returns the latest point, or the zero Point if the series is empty.
func (s *Series) Last() Point {
	if len(s.Points) == 0 {
		return Point{}
	}
	return s.Points[len(s.Points)-1]
}
