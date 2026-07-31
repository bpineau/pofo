package marketdata

import (
	"slices"
	"sort"
	"time"
)

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

// Dividend is one cash distribution of an asset: the ex-date (normalized to
// 00:00 UTC like every Point) and the per-share amount in the series' quote
// currency.
type Dividend struct {
	Date   time.Time
	Amount float64
}

// MergeDividends upserts events into dst by ex-date (one event per date,
// the newcomer wins) and returns dst sorted ascending. dst may be nil; the
// natural building block for incremental dividend tracking alongside
// Series.Dividends.
func MergeDividends(dst []Dividend, events ...Dividend) []Dividend {
	for _, ev := range events {
		i, found := slices.BinarySearchFunc(dst, ev.Date, func(d Dividend, t time.Time) int {
			return d.Date.Compare(t)
		})
		if found {
			dst[i] = ev
		} else {
			dst = slices.Insert(dst, i, ev)
		}
	}
	return dst
}

// Series is the price history of one asset, sorted by ascending date.
type Series struct {
	Symbol   string
	Name     string
	Currency string
	Source   string // "yahoo", "stooq", "ft", "morningstar" or "simdata"
	Points   []Point

	// Dividends lists the cash distributions the source reported, sorted
	// by ex-date. Beware of double counting: the default (adjusted) close
	// series already reinvests them; pair Dividends with raw closes
	// (FetchOptions.Raw) when accounting for income separately.
	Dividends []Dividend

	// SimulatedBefore is non-zero when points before that date were
	// reconstructed from ProxySymbol instead of actual quotes.
	SimulatedBefore time.Time
	ProxySymbol     string
}

// At returns the series value in force at the given time: the close of the
// last point dated at or before it (forward fill). ok is false before the
// first point or on an empty or nil series; on is the date of the point
// actually used, so callers can judge the staleness of the fill.
func (s *Series) At(at time.Time) (value float64, on time.Time, ok bool) {
	if s == nil {
		return 0, time.Time{}, false
	}
	i := sort.Search(len(s.Points), func(k int) bool { return s.Points[k].Date.After(at) })
	if i == 0 {
		return 0, time.Time{}, false
	}
	p := s.Points[i-1]
	return p.Close, p.Date, true
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
