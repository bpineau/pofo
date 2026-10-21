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
// which has no venue and is dated by crossSessionDay instead.
func venueTimezone(symbol, reported string) string {
	if _, _, isCross := fxCross(symbol); isCross {
		return ""
	}
	return reported
}

// crossBarSwitch is the first London date on which Yahoo's daily bar of a
// currency cross carries the rate at its own opening instant, i.e. the
// previous session's close, rather than its own day's rate. Measured on
// EURUSD=X, GBPUSD=X, JPY=X and CHF=X against the FRED noon fixings: through
// 2010-08-03 the bar sits within 0.08 to 0.20 % (yearly medians) of its own
// day's fixing and 0.3 to 0.6 % from the previous one; from 2010-08-04 the
// four crosses flip together, the same day, to 0.06 to 0.18 % from the
// previous day's fixing and 0.17 to 0.47 % from their own.
var crossBarSwitch = time.Date(2010, 8, 4, 0, 0, 0, 0, time.UTC)

// crossSessionDay is the calendar day a Yahoo daily bar of a currency cross
// belongs to, a session being the day whose closing rate the bar holds.
//
// Yahoo stamps every completed cross bar at MIDNIGHT LONDON (tz, the reported
// exchangeTimezoneName, Europe/London when absent), and since crossBarSwitch
// the bar's close is the rate at that instant: its open and close coincide,
// and its level is that of the evening before. The bar stamped Tuesday 00:00
// is therefore Monday's close, and the one stamped Monday 00:00 is FRIDAY's
// (the rate as the market reopens, within 0.10 % of that Friday's fixing,
// against 0.24 % from the Monday's). Before the switch the bar held its own
// day's rate and keeps its London date. A bar stamped at any other instant is
// the live quote of the day in progress, and keeps its day.
//
// The UTC dating this replaces read midnight London as 00:00 UTC in winter
// (every point a day LATE, so a Friday close was converted at Thursday's
// rate) and as 23:00 UTC the day before in summer (right since 2010, except
// that Monday's bar fell on a Sunday and Friday had no point at all, a summer
// Friday then converting at Thursday's rate too); before 2010 the summer half
// was a day early. Either way a date-matched join (ConvertCurrency picks the
// rate of the asset's own day) read the wrong session.
func crossSessionDay(ts int64, tz string) time.Time {
	loc := time.UTC
	if tz == "" {
		tz = "Europe/London"
	}
	if l, err := time.LoadLocation(tz); err == nil {
		loc = l
	}
	t := time.Unix(ts, 0).In(loc)
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	midnight := t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
	if !midnight || day.Before(crossBarSwitch) {
		return day
	}
	day = day.AddDate(0, 0, -1)
	for day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		day = day.AddDate(0, 0, -1)
	}
	return day
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
	Date   time.Time // ex-date, 00:00 UTC
	Amount float64   // per share, in the series' quote currency (major unit)
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
	Symbol   string  // the identifier it was fetched or loaded under (canonical)
	Name     string  // display name, from the source or the catalog; may be empty
	Currency string  // ISO code of Points (major unit); "" for rates and indices without one
	Source   string  // "yahoo", "stooq", "ft", "morningstar", "simdata", "refdata"...
	Points   []Point // strictly ascending dates, one per session

	// Dividends lists the cash distributions the source reported, sorted
	// by ex-date. Beware of double counting: the default (adjusted) close
	// series already reinvests them; pair Dividends with raw closes
	// (FetchOptions.Raw) when accounting for income separately.
	Dividends []Dividend

	// SimulatedBefore is non-zero when points before that date were
	// reconstructed from ProxySymbol instead of actual quotes.
	SimulatedBefore time.Time
	ProxySymbol     string // what the points before SimulatedBefore come from

	// EstimatedFrom is non-zero when the points from that date on are a
	// nowcast: the fund's last published value carried forward by the daily
	// moves of EstimateProxy (a catalog nowcast_proxy). Fetch adds the tail for
	// funds priced once a day and published with a lag; WithoutEstimates
	// removes it, and nothing stored or shipped keeps it.
	EstimatedFrom time.Time
	EstimateProxy string // the nowcast proxy behind the points from EstimatedFrom on

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

	// Ends is non-zero when the series STOPS BY DESIGN on that date, so an old
	// last point is the whole record rather than a late one: its publisher
	// discontinued it (EIA's WTI futures settlements after 2024-04-05, behind
	// the bundled WTI-ER-USD), or it is kept only up to where better data takes
	// over (a daily shape series trimmed where the real quotes begin). It comes
	// from a bundled file's "# ends:" header, a date optionally followed by the
	// reason in free text, and Verify honours it: such a series is never
	// reported stale, and one whose last point is not that date is reported
	// instead, since the declaration no longer describes the data.
	Ends time.Time
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
