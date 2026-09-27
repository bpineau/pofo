package marketdata

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// This file is the bridge between a Series and the slice-based rest of the
// tree (pkg/metrics, pkg/chart, a consumer's own code): a constructor for
// data the caller owns, the parallel slices every statistic takes, the
// one-call statistics, the level operations (Rebase, LessFee, Change), the
// calendar ones (Resample) and the strict sibling of Align. Everything here
// returns fresh slices and never modifies its receiver.

// NewSeries builds a series from a consumer's own data (a valuation it
// computed, a level it read elsewhere). symbol is free text and becomes
// Series.Symbol; every other metadata field is left zero for the caller to
// set.
//
// dates must be strictly ascending once normalized to their civil date at
// 00:00 UTC (the package invariant every Point.Date respects, read in each
// date's own location), so two instants of the same day are a duplicate and
// are refused. values are closes, or levels (a rate or an index series is
// welcome, zero and negative included), in the series' quote currency, and
// must be finite. The error says which index broke which rule. Empty slices
// build an empty series.
func NewSeries(symbol string, dates []time.Time, values []float64) (*Series, error) {
	if len(dates) != len(values) {
		return nil, fmt.Errorf("marketdata: NewSeries %s: %d dates for %d values", symbol, len(dates), len(values))
	}
	points := make([]Point, len(dates))
	for i, d := range dates {
		v := values[i]
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("marketdata: NewSeries %s: value %d (%s) is not finite", symbol, i, dayUTC(d).Format(time.DateOnly))
		}
		points[i] = Point{Date: dayUTC(d), Close: v}
		if i > 0 && !points[i].Date.After(points[i-1].Date) {
			return nil, fmt.Errorf("marketdata: NewSeries %s: date %d (%s) does not follow date %d (%s)",
				symbol, i, points[i].Date.Format(time.DateOnly), i-1, points[i-1].Date.Format(time.DateOnly))
		}
	}
	return &Series{Symbol: symbol, Points: points}, nil
}

// Len is the number of points; zero for a nil series.
func (s *Series) Len() int {
	if s == nil {
		return 0
	}
	return len(s.Points)
}

// Dates returns the points' dates (00:00 UTC) in a fresh slice, parallel to
// Values: the dates argument of every pkg/metrics function.
func (s *Series) Dates() []time.Time {
	out := make([]time.Time, s.Len())
	for i, p := range s.points() {
		out[i] = p.Date
	}
	return out
}

// Values returns the closes in a fresh slice, parallel to Dates, in the
// series' quote currency: the values argument of every pkg/metrics function.
func (s *Series) Values() []float64 {
	out := make([]float64, s.Len())
	for i, p := range s.points() {
		out[i] = p.Close
	}
	return out
}

// Returns is the series' simple period returns as FRACTIONS (0.01 = +1 %),
// one per step between consecutive points, so Len()-1 of them (nil below two
// points): the same numbers as metrics.Returns(s.Values()). A step is
// whatever separates two points, a session on a daily series, a month on a
// resampled one.
//
// It reads every step as it is: a step into one of the series' Junctions is
// returned like any other, and a caller working on a series that declares
// some must skip those steps itself (NewPanel does). It is meaningless on a
// rate or a yield, which is a percent LEVEL, not a price.
func (s *Series) Returns() []float64 {
	return metrics.Returns(s.Values())
}

// Stats is metrics.Compute over the series' dates and closes: CAGR,
// volatility, drawdowns and the rest, annualized at the series' own cadence
// (252 on daily closes, 12 on a month-end series such as Panel.Series
// returns). The closes must be positive, so it is meaningless on a rate or a
// yield; the error names the series.
func (s *Series) Stats() (metrics.Stats, error) {
	st, err := metrics.Compute(s.Dates(), s.Values())
	if err != nil {
		return metrics.Stats{}, fmt.Errorf("marketdata: %s: %w", s.symbol(), err)
	}
	return st, nil
}

// LessFee returns a copy of s with an annual charge deducted continuously on
// the calendar: each close is multiplied by (1 - annual) raised to the years
// elapsed since the first point (365.25-day years), so a full year costs
// exactly annual of the level and weekends and holidays pay their share. It
// turns a gross index into what a fund tracking it at that cost would have
// returned, or a fund's net NAV into a wrapper's.
//
// annual is a FRACTION per year: 0.0085 for 0.85 %/yr. It is NOT the percent
// convention of Client.Fees and portfolio.Holding.Fees, where 0.85 means
// 0.85 %/yr; passing such a figure here deducts 85 % a year. Convert with
// Fees/100. A negative annual is an uplift (a cost the series carries and
// the target does not). An annual that is not finite, or not below 1 (a
// charge of the whole level or more), is an error.
//
// Metadata is carried over; Dividends are not rescaled (see Rebase).
func (s *Series) LessFee(annual float64) (*Series, error) {
	if !(annual < 1) || math.IsInf(annual, -1) {
		return nil, fmt.Errorf("marketdata: LessFee %s: annual charge %v is not below 1 (a FRACTION per year: 0.0085 for 0.85 %%/yr)", s.symbol(), annual)
	}
	out := s.clone()
	if len(out.Points) == 0 || annual == 0 {
		return out, nil
	}
	t0 := out.Points[0].Date
	for i, p := range out.Points {
		yrs := p.Date.Sub(t0).Hours() / 24 / 365.25
		out.Points[i].Close = p.Close * math.Pow(1-annual, yrs)
	}
	return out, nil
}

// Change is the cumulative return of s between two dates, as a FRACTION
// (0.25 = +25 %): the close at or before to over the close at or before
// from, minus one. Change(2008-01-01, 2008-12-31) is therefore calendar
// 2008, measured from the last close of 2007, and a weekend or holiday
// bound reads the session before it. Both bounds are civil dates.
//
// It refuses what it could only answer wrong: to before from; a series
// that starts after from (it has no level there); a series that stops
// before to, since the move after its last quote is unknown and a shorter
// episode would come back unannounced (a series knows its level at a date
// when it quotes after it, or on or after that date's last weekday, so a
// series ending on a Friday answers for the weekend); a non-positive close
// at from; and a Junction between the two closes, where the change is a
// change of definition rather than a move.
func (s *Series) Change(from, to time.Time) (float64, error) {
	from, to = dayUTC(from), dayUTC(to)
	if to.Before(from) {
		return 0, fmt.Errorf("marketdata: Change %s: ends %s, before it starts %s", s.symbol(), to.Format(time.DateOnly), from.Format(time.DateOnly))
	}
	if s.Len() == 0 || s.First().Date.After(from) {
		return 0, fmt.Errorf("marketdata: Change %s: no close at or before %s", s.symbol(), from.Format(time.DateOnly))
	}
	if !knowsLevelAt(s.Last().Date, to) {
		return 0, fmt.Errorf("marketdata: Change %s: last close %s, before %s", s.symbol(), s.Last().Date.Format(time.DateOnly), to.Format(time.DateOnly))
	}
	a, b := s.closeAt(from), s.closeAt(to)
	if !(a.Close > 0) {
		return 0, fmt.Errorf("marketdata: Change %s: close %v on %s is not a positive price", s.symbol(), a.Close, a.Date.Format(time.DateOnly))
	}
	if j, ok := s.junctionIn(a.Date, b.Date); ok {
		return 0, fmt.Errorf("marketdata: Change %s: definition junction on %s between %s and %s", s.symbol(),
			j.Format(time.DateOnly), a.Date.Format(time.DateOnly), b.Date.Format(time.DateOnly))
	}
	return b.Close/a.Close - 1, nil
}

// closeAt is the last point at or before d; the series must start at or
// before d.
func (s *Series) closeAt(d time.Time) Point {
	i := sort.Search(len(s.Points), func(i int) bool { return s.Points[i].Date.After(d) })
	return s.Points[i-1]
}

// junctionIn reports the first of s's Junctions in (from, to]: the step from
// the close on from to the close on to crosses it.
func (s *Series) junctionIn(from, to time.Time) (time.Time, bool) {
	for _, j := range s.Junctions {
		if j.After(from) && !j.After(to) {
			return j, true
		}
	}
	return time.Time{}, false
}

// knowsLevelAt reports whether a series whose last quote is last knows its
// level at d: it quoted at or after d's last weekday (d itself, or the
// Friday before a weekend). The rule is conservative on purpose: a series
// that stops on the Thursday before a Friday holiday does not answer for
// that Friday, a missing answer rather than a stale one.
func knowsLevelAt(last, d time.Time) bool {
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, -1)
	}
	return !last.Before(d)
}

// Rebase returns a copy of s scaled so its first close is base (100 for an
// index, 1 for a growth-of-one curve); every return is left unchanged.
// Metadata (Symbol, Name, Currency, Source, SimulatedBefore, EstimatedFrom,
// Junctions, Dividends, ...) is carried over in fresh slices. Dividends are
// NOT rescaled: they are cash per share in the quote currency, and a rebased
// curve is no longer a per-share price, so pair a rebased series with its
// dividends only knowingly.
//
// An empty series, or one whose first close is zero or not finite (nothing
// can scale it to base), comes back as an unscaled copy.
func (s *Series) Rebase(base float64) *Series {
	out := s.clone()
	if len(out.Points) == 0 {
		return out
	}
	first := out.Points[0].Close
	if first == 0 || math.IsNaN(first) || math.IsInf(first, 0) {
		return out
	}
	k := base / first
	for i := range out.Points {
		out.Points[i].Close *= k
	}
	return out
}

// Frequency is a calendar period, counted in months, or Daily: the quote
// date itself, whatever separates two quotes.
type Frequency int

// The calendar periods Resample and NewPanel cut on. Any positive number of
// months is accepted; these name the usual ones. Daily, the zero value, cuts
// nothing: every quote is its own period.
const (
	Daily     Frequency = 0
	Monthly   Frequency = 1
	Quarterly Frequency = 3
	Yearly    Frequency = 12
)

// index numbers the period of f holding d: months since year 0 divided by
// f, so consecutive periods get consecutive numbers. It is meaningless for
// Daily.
func (f Frequency) index(d time.Time) int {
	return (d.Year()*12 + int(d.Month()) - 1) / int(f)
}

// end is the canonical label of period p of f (an index): the calendar last
// day of its last month, at 00:00 UTC.
func (f Frequency) end(p int) time.Time {
	last := (p+1)*int(f) - 1 // the block's last month, counted from year 0
	return time.Date(last/12, time.Month(last%12+2), 0, 0, 0, 0, 0, time.UTC)
}

// Resample keeps the last point of each calendar period of f months and
// returns them as a new series: each kept point keeps its own close AND its
// own date, so a monthly series is dated on the last TRADING day of each
// month (a month-END series, the convention every bundled monthly anchor
// follows), never on a calendar month-end nobody quoted. Periods are blocks
// of f months counted from January, so Quarterly cuts on calendar quarters
// and Yearly on calendar years (any divisor of 12 keeps to the calendar
// year).
//
// The last period is kept even when it is not complete: a series ending
// mid-month ends on that day's close, and the caller reads Last().Date to
// know. The first period is kept as well, starting wherever the series does.
//
// Metadata is carried over. Dividends that fall on a dropped date are
// dropped with it, those on a kept date are kept: resampling a raw
// (unadjusted) series therefore loses the mid-period distributions, so
// resample an adjusted one, or book the dividends before resampling. A
// Junction MOVES to the kept close of the period it falls in (the first kept
// date at or after it), because the resampled step into that close spans the
// definition change and must be refused like the daily step was; a junction
// after the last kept date is dropped, there being no step for it to guard.
// Two junctions in one period collapse into one.
//
// Daily returns an unchanged copy, every point being its own period. A
// negative f is an error.
func (s *Series) Resample(f Frequency) (*Series, error) {
	if f < 0 {
		return nil, fmt.Errorf("marketdata: Resample %s: negative frequency (%d months)", s.symbol(), f)
	}
	out := s.clone()
	if f == Daily {
		return out, nil
	}
	kept := out.Points[:0]
	for i, p := range out.Points {
		if i+1 < len(out.Points) && f.index(out.Points[i+1].Date) == f.index(p.Date) {
			continue
		}
		kept = append(kept, p)
	}
	out.Points = kept

	onKept := make(map[time.Time]bool, len(kept))
	for _, p := range kept {
		onKept[p.Date] = true
	}
	var divs []Dividend
	for _, d := range out.Dividends {
		if onKept[d.Date] {
			divs = append(divs, d)
		}
	}
	var junctions []time.Time
	for _, j := range out.Junctions {
		for _, p := range kept {
			if !p.Date.Before(j) {
				if n := len(junctions); n == 0 || !junctions[n-1].Equal(p.Date) {
					junctions = append(junctions, p.Date)
				}
				break
			}
		}
	}
	out.Dividends, out.Junctions = divs, junctions
	return out, nil
}

// points is s.Points, nil-safe.
func (s *Series) points() []Point {
	if s == nil {
		return nil
	}
	return s.Points
}

// clone is a copy of s whose slices are its own; a nil series clones to an
// empty one.
func (s *Series) clone() *Series {
	if s == nil {
		return &Series{}
	}
	out := *s
	out.Points = append([]Point(nil), s.Points...)
	out.Dividends = append([]Dividend(nil), s.Dividends...)
	out.Junctions = append([]time.Time(nil), s.Junctions...)
	return &out
}

// CommonWindow is the window on which every series of list is defined: start
// is the latest first quote, end the earliest last quote, both dates of the
// series themselves (00:00 UTC). ok is false when list is empty, when one of
// its series is nil or empty, or when the window is (a series ends before
// another starts); start and end are then zero.
func CommonWindow(list ...*Series) (start, end time.Time, ok bool) {
	if len(list) == 0 {
		return time.Time{}, time.Time{}, false
	}
	for i, s := range list {
		if s.Len() == 0 {
			return time.Time{}, time.Time{}, false
		}
		first, last := s.First().Date, s.Last().Date
		if i == 0 || first.After(start) {
			start = first
		}
		if i == 0 || last.Before(end) {
			end = last
		}
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

// Aligned is several series on one calendar. Dates is the sorted union of
// their quoting days inside the window, IDs[i] is the Symbol of the i-th
// input series and Levels[i] its closes at every one of Dates, parallel to
// it, forward-filled across the days it did not quote (in its own quote
// currency: AlignSeries converts nothing).
type Aligned struct {
	Dates  []time.Time
	IDs    []string
	Levels [][]float64
}

// AlignSeries is Align with its trap closed: it aligns list on the union of
// the series' dates inside [from, to] exactly as Align does, but refuses,
// with an error naming the series, to forward-fill a level it does not have.
//
// from must be at or after every series' first quote (Align would fill the
// dates before it with ZEROS) and at or before every series' last quote (a
// series whose whole history precedes the window would be one level held
// flat, with a zero return on every step). A zero from means CommonWindow's
// start; a zero to is open-ended, and a series ending before to is then held
// flat at its last close after it, as Align does. Both bounds are read as
// civil dates, like every Point.Date. A nil or empty series, an
// empty list, to before from and a window no series quotes in are errors
// too.
func AlignSeries(list []*Series, from, to time.Time) (*Aligned, error) {
	if len(list) == 0 {
		return nil, errors.New("marketdata: AlignSeries: no series")
	}
	for i, s := range list {
		if s.Len() == 0 {
			return nil, fmt.Errorf("marketdata: AlignSeries: series %d (%q) is empty", i, s.symbol())
		}
	}
	if from.IsZero() {
		// CommonWindow's start, the latest first quote. Computed here rather
		// than read from CommonWindow so that a disjoint list reaches the
		// check below, which names the series that ends too early.
		for _, s := range list {
			if s.First().Date.After(from) {
				from = s.First().Date
			}
		}
	} else {
		from = dayUTC(from)
	}
	if !to.IsZero() {
		to = dayUTC(to)
	}
	if !to.IsZero() && to.Before(from) {
		return nil, fmt.Errorf("marketdata: AlignSeries: window ends %s, before it starts %s",
			to.Format(time.DateOnly), from.Format(time.DateOnly))
	}
	for _, s := range list {
		if from.Before(s.First().Date) {
			return nil, fmt.Errorf("marketdata: AlignSeries: %s starts %s, after the window's start %s",
				s.symbol(), s.First().Date.Format(time.DateOnly), from.Format(time.DateOnly))
		}
		if from.After(s.Last().Date) {
			return nil, fmt.Errorf("marketdata: AlignSeries: %s ends %s, before the window's start %s",
				s.symbol(), s.Last().Date.Format(time.DateOnly), from.Format(time.DateOnly))
		}
	}
	dates, levels := Align(list, from, to)
	if len(dates) == 0 {
		return nil, fmt.Errorf("marketdata: AlignSeries: no quote between %s and %s",
			from.Format(time.DateOnly), to.Format(time.DateOnly))
	}
	ids := make([]string, len(list))
	for i, s := range list {
		ids[i] = s.Symbol
	}
	return &Aligned{Dates: dates, IDs: ids, Levels: levels}, nil
}

// symbol names s in an error message.
func (s *Series) symbol() string {
	if s == nil {
		return "<nil>"
	}
	if s.Symbol == "" {
		return "<unnamed>"
	}
	return s.Symbol
}

// Returns is every series' simple period returns on the common calendar, as
// FRACTIONS: [asset][period], len(Dates)-1 each, parallel to IDs. Period k
// runs from Dates[k] to Dates[k+1], so a series that did not quote on one of
// those days reads a zero return there and the whole move on its next quote.
// This is the input metrics expects for a correlation or covariance across
// assets.
func (a *Aligned) Returns() [][]float64 {
	out := make([][]float64, len(a.Levels))
	for i, levels := range a.Levels {
		out[i] = metrics.Returns(levels)
		if out[i] == nil {
			out[i] = []float64{}
		}
	}
	return out
}

// Series returns column i as a Series on the common calendar: Symbol IDs[i],
// one point per Dates, closes Levels[i] (forward-filled ones included). No
// other metadata survives alignment. It panics when i is out of range.
func (a *Aligned) Series(i int) *Series {
	points := make([]Point, len(a.Dates))
	for k, d := range a.Dates {
		points[k] = Point{Date: d, Close: a.Levels[i][k]}
	}
	return &Series{Symbol: a.IDs[i], Points: points}
}
