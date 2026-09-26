package marketdata

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// This file is the bridge between a Series and the slice-based rest of the
// tree (pkg/metrics, pkg/chart, a consumer's own code): a constructor for
// data the caller owns, the parallel slices every statistic takes, the
// calendar operations (Rebase, Resample) and the strict sibling of Align.
// Everything here returns fresh slices and never modifies its receiver.

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
// some must skip those steps itself. It is meaningless on a rate or a
// yield, which is a percent LEVEL, not a price.
func (s *Series) Returns() []float64 {
	return periodReturns(s.Values())
}

// periodReturns is metrics.Returns, restated here because marketdata must
// not import metrics (the layering runs the other way).
func periodReturns(values []float64) []float64 {
	if len(values) < 2 {
		return nil
	}
	r := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		r[i-1] = values[i]/values[i-1] - 1
	}
	return r
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

// Frequency is a calendar period, counted in months.
type Frequency int

// The calendar periods Resample cuts on. Any positive number of months is
// accepted; these name the usual ones.
const (
	Monthly   Frequency = 1
	Quarterly Frequency = 3
	Yearly    Frequency = 12
)

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
// f must be positive; Resample panics otherwise, as for any programming
// error.
func (s *Series) Resample(f Frequency) *Series {
	if f < 1 {
		panic(fmt.Sprintf("marketdata: Resample with a non-positive frequency (%d months)", f))
	}
	out := s.clone()
	period := func(d time.Time) int { return (d.Year()*12 + int(d.Month()) - 1) / int(f) }
	kept := out.Points[:0]
	for i, p := range out.Points {
		if i+1 < len(out.Points) && period(out.Points[i+1].Date) == period(p.Date) {
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
	return out
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
		out[i] = periodReturns(levels)
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
