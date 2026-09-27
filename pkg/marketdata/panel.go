package marketdata

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Panel is several series' simple returns over ONE shared calendar of
// periods: the table a statistic ACROSS series reads (a correlation, a
// regression, a blend, the other series' mean over one series' worst
// months). Period t runs from Starts[t] to Ends[t], identical for every
// column, and R[i][t] is the return of series IDs[i] over it, as a FRACTION
// (0.01 = +1 %), in that series' own quote currency (a Panel converts
// nothing: convert first, with Client.ConvertCurrency or FetchOptions).
//
// Every field is exported for reading and for handing to pkg/metrics (R is
// the [asset][period] input of metrics.CorrelationMatrix and Covariance);
// the methods return new panels and never modify their receiver. NewPanel
// builds one; Between and Pick select periods; Mix appends a blend; Col
// reads a column, Series rebuilds one as a level for metrics.Compute.
type Panel struct {
	Freq   Frequency   // the calendar the periods were cut on
	Starts []time.Time // Starts[t]: the label period t is measured from
	Ends   []time.Time // Ends[t]: the label period t ends on, its name
	IDs    []string    // the columns, the input series' Symbols in order
	R      [][]float64 // R[i][t]: the return of IDs[i] over period t, a fraction

	cadence float64 // a Daily panel's periods per year, measured by NewPanel and kept by every selection
}

// NewPanel cuts every series of list into periods of f and returns their
// returns over the periods they all share.
//
// Labels are CANONICAL, so panels built from different sources join
// exactly: a Monthly period is labelled by the calendar last day of its
// month at 00:00 UTC (2024-03-31, whether the series last quoted on the
// 28th or the 31st), a Quarterly or Yearly one by the last day of its
// quarter or year, and a Daily period by the quote date itself. Each series
// is reduced to its period closes, the last close of each period (as
// Resample keeps them), the labels every series has are kept, and every
// return runs between two consecutive shared labels, so all the columns of
// a period span the same interval. A monthly source dated at the START of
// its month (some publishers label March's close 03-01) must be relabelled
// before it gets here: it would land one period early.
//
// Three rules keep a return from being wrong in silence:
//
//   - A series' LAST period counts only if it is complete: the series must
//     quote on or after the period's last weekday (the rule Change applies
//     to its end date). A daily series fetched on the 15th therefore does
//     not contribute a half month labelled as a whole one; the rare month
//     whose last weekday was a holiday is dropped with it, a missing month
//     rather than a wrong one.
//   - For Monthly and coarser, a series with NO quote in a period between
//     the shared window's first and last label is an error naming the series
//     and the period: its return across the hole would span two periods
//     while every other column spans one. For Daily, differing exchange
//     holidays are normal, and a date some series lacks is simply not a
//     label: the returns around it span the two sessions for every column,
//     the pairing metrics.Beta has always used. A daily panel is only as
//     daily as its sparsest member (a weekly NAV among daily closes makes
//     a weekly panel), which PeriodsPerYear measures.
//   - A period whose span crosses one of a series' Junctions (a change of
//     definition, not a move) is dropped for EVERY column, so the columns
//     stay on one calendar. The panel then has a gap, visible as a Starts[t]
//     that is not Ends[t-1]; Series refuses to chain across it, every other
//     statistic reads the periods that remain.
//
// Every close the panel reads must be positive (a rate or a yield is a
// LEVEL, not a price; its "returns" mean nothing). An empty list, a nil or
// empty series, a missing or duplicated Symbol, a negative f and series that
// share fewer than two labels are errors too.
func NewPanel(f Frequency, list ...*Series) (*Panel, error) {
	if f < 0 {
		return nil, fmt.Errorf("marketdata: NewPanel: negative frequency (%d months)", f)
	}
	if len(list) == 0 {
		return nil, errors.New("marketdata: NewPanel: no series")
	}
	ids := make([]string, len(list))
	closes := make([][]periodClose, len(list))
	for i, s := range list {
		switch {
		case s.Len() == 0:
			return nil, fmt.Errorf("marketdata: NewPanel: series %d (%s) is empty", i, s.symbol())
		case s.Symbol == "":
			return nil, fmt.Errorf("marketdata: NewPanel: series %d has no Symbol to name its column", i)
		case slices.Contains(ids[:i], s.Symbol):
			return nil, fmt.Errorf("marketdata: NewPanel: two series named %s", s.Symbol)
		}
		ids[i] = s.Symbol
		closes[i] = periodCloses(s, f)
		if len(closes[i]) == 0 {
			return nil, fmt.Errorf("marketdata: NewPanel: %s has no complete period (its only one ends %s)",
				s.Symbol, s.Last().Date.Format(time.DateOnly))
		}
	}

	// The shared window: from the latest first label to the earliest last.
	first, last := 0, 0
	for i, c := range closes {
		if c[0].label.After(closes[first][0].label) {
			first = i
		}
		if c[len(c)-1].label.Before(closes[last][len(closes[last])-1].label) {
			last = i
		}
	}
	start, end := closes[first][0].label, closes[last][len(closes[last])-1].label
	if !start.Before(end) {
		return nil, fmt.Errorf("marketdata: NewPanel: no two period ends shared by every series (%s ends %s, %s starts %s)",
			ids[last], end.Format(time.DateOnly), ids[first], start.Format(time.DateOnly))
	}
	for i := range closes {
		lo, _ := slices.BinarySearchFunc(closes[i], start, byLabel)
		hi, _ := slices.BinarySearchFunc(closes[i], end.AddDate(0, 0, 1), byLabel)
		closes[i] = closes[i][lo:hi]
	}

	labels, err := sharedLabels(f, ids, closes, start, end)
	if err != nil {
		return nil, err
	}
	if len(labels) < 2 {
		return nil, fmt.Errorf("marketdata: NewPanel: the series share fewer than two dates between %s and %s",
			start.Format(time.DateOnly), end.Format(time.DateOnly))
	}
	at := make([][]periodClose, len(list)) // at[i][k]: series i's close on labels[k]
	for i, c := range closes {
		at[i] = make([]periodClose, 0, len(labels))
		for _, pc := range c {
			if _, ok := slices.BinarySearchFunc(labels, pc.label, time.Time.Compare); ok {
				at[i] = append(at[i], pc)
			}
		}
		for _, pc := range at[i] {
			if !(pc.close > 0) || math.IsInf(pc.close, 1) {
				return nil, fmt.Errorf("marketdata: NewPanel: %s closes at %v on %s, not a positive price",
					ids[i], pc.close, pc.date.Format(time.DateOnly))
			}
		}
	}

	p := &Panel{Freq: f, IDs: ids, R: make([][]float64, len(list))}
periods:
	for k := 1; k < len(labels); k++ {
		for i, s := range list {
			if _, ok := s.junctionIn(at[i][k-1].date, at[i][k].date); ok {
				continue periods
			}
		}
		p.Starts, p.Ends = append(p.Starts, labels[k-1]), append(p.Ends, labels[k])
		for i := range list {
			p.R[i] = append(p.R[i], at[i][k].close/at[i][k-1].close-1)
		}
	}
	if p.Len() == 0 {
		return nil, fmt.Errorf("marketdata: NewPanel: every shared period crosses a definition junction")
	}
	if f == Daily {
		p.cadence = spanCadence(p.Starts, p.Ends)
	}
	return p, nil
}

// periodClose is a series' close ending one period: its canonical label,
// the series' own date of that close, the close.
type periodClose struct {
	label, date time.Time
	close       float64
}

// byLabel orders period closes by label, for a binary search.
func byLabel(pc periodClose, label time.Time) int { return pc.label.Compare(label) }

// periodCloses reduces s to the closes that end its periods of f, labelled
// canonically, dropping a last period the series does not cover to its end.
func periodCloses(s *Series, f Frequency) []periodClose {
	var out []periodClose
	for i, p := range s.Points {
		if f == Daily {
			out = append(out, periodClose{p.Date, p.Date, p.Close})
			continue
		}
		if i+1 < len(s.Points) && f.index(s.Points[i+1].Date) == f.index(p.Date) {
			continue
		}
		out = append(out, periodClose{f.end(f.index(p.Date)), p.Date, p.Close})
	}
	if n := len(out); f != Daily && !knowsLevelAt(out[n-1].date, out[n-1].label) {
		out = out[:n-1]
	}
	return out
}

// sharedLabels is the calendar of a panel: for Daily, the dates every series
// quotes; for a calendar period, every period of the window, which every
// series must then have, or the error names the first that has a hole.
func sharedLabels(f Frequency, ids []string, closes [][]periodClose, start, end time.Time) ([]time.Time, error) {
	if f == Daily {
		var labels []time.Time
		for _, pc := range closes[0] {
			shared := true
			for _, c := range closes[1:] {
				if _, ok := slices.BinarySearchFunc(c, pc.label, byLabel); !ok {
					shared = false
					break
				}
			}
			if shared {
				labels = append(labels, pc.label)
			}
		}
		return labels, nil
	}
	var labels []time.Time
	for q := f.index(start); q <= f.index(end); q++ {
		labels = append(labels, f.end(q))
	}
	for i, c := range closes {
		for k, label := range labels {
			if k >= len(c) || !c[k].label.Equal(label) {
				return nil, fmt.Errorf("marketdata: NewPanel: %s has no quote in the period ending %s, inside the shared window %s to %s",
					ids[i], label.Format(time.DateOnly), start.Format(time.DateOnly), end.Format(time.DateOnly))
			}
		}
	}
	return labels, nil
}

// Len is the number of periods.
func (p *Panel) Len() int { return len(p.Ends) }

// column is the position of id among the columns, -1 when absent.
func (p *Panel) column(id string) int { return slices.Index(p.IDs, id) }

// Col returns the returns of column id over every period, in a fresh slice
// parallel to Ends: the argument of metrics.Mean, Regress, LowestK and the
// other functions over bare returns. It is an error when the panel has no
// such column; the error lists the columns it has.
func (p *Panel) Col(id string) ([]float64, error) {
	i := p.column(id)
	if i < 0 {
		return nil, fmt.Errorf("marketdata: Panel.Col: no column %s (the panel holds %s)", id, strings.Join(p.IDs, ", "))
	}
	return slices.Clone(p.R[i]), nil
}

// Between returns the periods that END inside [from, to], both bounds civil
// dates and a zero bound open: Between(2008-01-01, 2008-12-31) on a monthly
// panel is calendar 2008, its first period starting on 2007-12-31.
func (p *Panel) Between(from, to time.Time) *Panel {
	var keep []int
	for t, e := range p.Ends {
		if (from.IsZero() || !e.Before(dayUTC(from))) && (to.IsZero() || !e.After(dayUTC(to))) {
			keep = append(keep, t)
		}
	}
	return p.pick(keep)
}

// Pick returns the listed periods, in the order listed: a SAMPLE of periods,
// such as the worst decile of one column (metrics.LowestK), on which the
// other columns are then read (the conditional mean of a hedge over an
// asset's worst months). A picked panel is no longer one path, so Series
// refuses it unless the periods listed are consecutive. An index out of
// range is an error.
func (p *Panel) Pick(periods []int) (*Panel, error) {
	for _, t := range periods {
		if t < 0 || t >= p.Len() {
			return nil, fmt.Errorf("marketdata: Panel.Pick: period %d out of range (the panel has %d)", t, p.Len())
		}
	}
	return p.pick(periods), nil
}

// pick copies the listed periods into a new panel.
func (p *Panel) pick(periods []int) *Panel {
	out := &Panel{Freq: p.Freq, IDs: slices.Clone(p.IDs), R: make([][]float64, len(p.R)), cadence: p.cadence}
	out.Starts, out.Ends = make([]time.Time, len(periods)), make([]time.Time, len(periods))
	for i := range p.R {
		out.R[i] = make([]float64, len(periods))
	}
	for k, t := range periods {
		out.Starts[k], out.Ends[k] = p.Starts[t], p.Ends[t]
		for i := range p.R {
			out.R[i][k] = p.R[i][t]
		}
	}
	return out
}

// weightSumTolerance is how far from 1 Mix lets its weights sum, to absorb
// the rounding of a weight typed as a decimal.
const weightSumTolerance = 1e-9

// Mix returns the panel with one more column, id: the blend of the named
// columns rebalanced to weights at the start of every period, so its return
// over a period is the weighted sum of theirs. Weights are FRACTIONS (0.6,
// not 60) and must sum to 1: the blend assumes nothing it was not given. A
// share left unassigned would silently earn a zero return, as cash paying
// nothing, and a sum above 1 would borrow the excess for free; both flatter
// the blend. Put the financing in the panel instead, as a cash series, with
// a NEGATIVE weight for a leveraged blend: {"EQ": 0.9, "BOND": 0.6,
// "CASH": -0.5} is a 150 % book financed at the cash rate. Any sign is
// accepted, a short leg included.
//
// It is an error when id is empty or already a column, when weights is
// empty, names an absent column or holds a non-finite weight, and when the
// weights do not sum to 1.
func (p *Panel) Mix(id string, weights map[string]float64) (*Panel, error) {
	switch {
	case id == "":
		return nil, errors.New("marketdata: Panel.Mix: empty column name")
	case p.column(id) >= 0:
		return nil, fmt.Errorf("marketdata: Panel.Mix: column %s already exists", id)
	case len(weights) == 0:
		return nil, fmt.Errorf("marketdata: Panel.Mix %s: no weights", id)
	}
	// Walk the columns in panel order, never the map's: the blend must not
	// depend on Go's randomized map iteration, down to the last bit.
	sum, legs := 0.0, 0
	blend := make([]float64, p.Len())
	for i, col := range p.IDs {
		w, ok := weights[col]
		if !ok {
			continue
		}
		if math.IsNaN(w) || math.IsInf(w, 0) {
			return nil, fmt.Errorf("marketdata: Panel.Mix %s: weight of %s is %v", id, col, w)
		}
		sum += w
		legs++
		for t, r := range p.R[i] {
			blend[t] += w * r
		}
	}
	if legs < len(weights) {
		var absent []string
		for col := range weights {
			if p.column(col) < 0 {
				absent = append(absent, col)
			}
		}
		slices.Sort(absent)
		return nil, fmt.Errorf("marketdata: Panel.Mix %s: no column %s (the panel holds %s)",
			id, strings.Join(absent, ", "), strings.Join(p.IDs, ", "))
	}
	if math.Abs(sum-1) > weightSumTolerance {
		return nil, fmt.Errorf("marketdata: Panel.Mix %s: weights sum to %v, not 1 (put the rest in a cash column, negative to borrow)", id, sum)
	}
	out := &Panel{
		Freq:   p.Freq,
		Starts: slices.Clone(p.Starts),
		Ends:   slices.Clone(p.Ends),
		IDs:    append(slices.Clone(p.IDs), id),
		R:      make([][]float64, 0, len(p.R)+1),

		cadence: p.cadence,
	}
	for _, r := range p.R {
		out.R = append(out.R, slices.Clone(r))
	}
	out.R = append(out.R, blend)
	return out, nil
}

// Series rebuilds column id as a level: 1 on Starts[0], then compounded by
// every period's return, one point per Ends. Its dates are the canonical
// labels, so metrics.Compute (or Series.Stats) reads it at the panel's
// cadence, 12 a year for a monthly panel, and every statistic of a blend
// built by Mix is one call away. Symbol is id; no other metadata survives.
//
// It is an error when the panel has no such column or no period, and when
// its periods are not one unbroken path (a junction dropped one, or Pick
// sampled them): chaining returns across a gap would state a level the
// series never had.
func (p *Panel) Series(id string) (*Series, error) {
	i := p.column(id)
	if i < 0 {
		return nil, fmt.Errorf("marketdata: Panel.Series: no column %s (the panel holds %s)", id, strings.Join(p.IDs, ", "))
	}
	if p.Len() == 0 {
		return nil, fmt.Errorf("marketdata: Panel.Series %s: no period", id)
	}
	points := make([]Point, p.Len()+1)
	points[0] = Point{Date: p.Starts[0], Close: 1}
	for t, r := range p.R[i] {
		if t > 0 && !p.Starts[t].Equal(p.Ends[t-1]) {
			return nil, fmt.Errorf("marketdata: Panel.Series %s: period %d starts %s, not where period %d ended (%s): not one path",
				id, t, p.Starts[t].Format(time.DateOnly), t-1, p.Ends[t-1].Format(time.DateOnly))
		}
		points[t+1] = Point{Date: p.Ends[t], Close: points[t].Close * (1 + r)}
	}
	return &Series{Symbol: id, Points: points}, nil
}

// PeriodsPerYear is the cadence that annualizes the panel's per-period
// statistics (a volatility by its square root, a mean return or a
// regression's alpha by the count itself): 12 for Monthly, 4 for
// Quarterly, 1 for Yearly, and for Daily the count metrics.PeriodsPerYear
// measures on the periods' own spans, 252 on trading days, 52 when a weekly
// member made the panel weekly. A daily cadence is measured once, on the
// whole calendar NewPanel built, and every panel selected or mixed from it
// keeps that count: a handful of picked sessions is a sample of the
// calendar, not a calendar to measure. It is NaN for a daily panel built by
// hand without a period.
func (p *Panel) PeriodsPerYear() float64 {
	switch {
	case p.Freq > 0:
		return 12 / float64(p.Freq)
	case p.cadence > 0:
		return p.cadence
	}
	return spanCadence(p.Starts, p.Ends)
}

// spanCadence is metrics.PeriodsPerYear over periods' own spans, laid end
// to end so that a gap between two periods (a dropped junction) is not read
// as one more span.
func spanCadence(starts, ends []time.Time) float64 {
	chain := make([]time.Time, len(ends)+1)
	for t := range ends {
		chain[t+1] = chain[t].Add(ends[t].Sub(starts[t]))
	}
	return metrics.PeriodsPerYear(chain)
}
