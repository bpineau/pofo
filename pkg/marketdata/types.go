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

// venueTimezone is the time zone a symbol's daily bars are to be dated in:
// the venue's own, as the provider reports it, except for a currency cross,
// which is left on the UTC calendar it has always used.
//
// A cross has no exchange and no trading day, and the zone Yahoo attaches to
// one is decorative. Its dating also carries an anomaly of its own that this
// rule must not paper over: in the European summer a cross has no Friday-dated
// close and one dated on the SUNDAY before the following week, whose level
// measures as that missing Friday (median 0.10 % from the FRED noon fixing of
// that Friday, against 0.24 % from the Monday after, over EURUSD=X since
// 2016). Re-dating that point by a time zone would land it on a Monday the
// series already holds, and the later value would silently overwrite the
// earlier one: a lost session, which is worse than a misplaced one. The
// anomaly stands, named, for a fix that understands it.
func venueTimezone(symbol, reported string) string {
	if _, _, isCross := fxCross(symbol); isCross {
		return ""
	}
	return reported
}

// sessionDay is the calendar day a provider's daily bar belongs to: its
// timestamp is an INSTANT of the trading session, not a date, so the day must
// be read in the venue's own time zone (tz, Yahoo's exchangeTimezoneName).
//
// Reading it in UTC instead is wrong for every venue whose session opens
// before midnight UTC, and the error is a whole day. The ASX opens at 10:00
// Sydney, which is 23:00 UTC of the day before while Australia is on summer
// time: half of every ASX history lands one day early, Monday's session on a
// SUNDAY and Friday's on a Thursday. A cached ASX line here shows 534 Sunday
// closes against 596 Fridays, where every other weekday holds about 1160.
// Mis-dated points then break every date-matched join this toolkit makes
// (metrics align on exact dates, a currency conversion picks the day's rate)
// while looking perfectly ordinary.
//
// The correction is one-directional, and deliberately so: the UTC reading is
// never LATE, only early, since it drops the hours a venue east of Greenwich
// has already lived. So a local date before the UTC one is not a correction
// but a time zone that disagrees with the instant (a venue west of Greenwich,
// whose bars this provider already stamps inside the right UTC day), and the
// UTC date stands. An empty or unknown tz also leaves the UTC date: the
// package's dating then behaves exactly as it did before.
func sessionDay(ts int64, tz string) time.Time {
	t := time.Unix(ts, 0)
	utc := dayUTC(t.UTC())
	if tz == "" {
		return utc
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return utc
	}
	if local := dayUTC(t.In(loc)); local.After(utc) {
		return local
	}
	return utc
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

	// EstimatedFrom is non-zero when the points from that date on are a
	// nowcast: the fund's last published value carried forward by the daily
	// moves of EstimateProxy (a catalog nowcast_proxy). Fetch adds the tail for
	// funds priced once a day and published with a lag; WithoutEstimates
	// removes it, and nothing stored or shipped keeps it.
	EstimatedFrom time.Time
	EstimateProxy string

	// Junctions are the dates on which the series' DEFINITION changes rather
	// than its subject: the publisher started measuring something else, so the
	// step INTO such a date is not a market move and no consumer may read a
	// return from it. Both published levels are kept, because both are what the
	// source says; what is refused is the return between them.
	//
	// The bundled TREASURY-LONG-YIELD carries one, 1973-01-04, where H.15's
	// 20-year constant maturity moved 0.74 pt overnight while the 10-year point
	// sat still. See cmd/gen-tyield-refdata and simgen.TreasuryTR, which skips
	// any step spanning one of these dates.
	Junctions []time.Time
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
