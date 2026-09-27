package analyze

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// DefaultDivergences is how many dated divergences Pair lists per calendar
// when PairOptions.Divergences is zero.
const DefaultDivergences = 5

// PairOptions tunes Pair. The zero value lists DefaultDivergences per
// calendar and forgives no difference in closing times.
type PairOptions struct {
	// Divergences is how many of the periods where the two returns disagree
	// most each calendar lists, largest first: zero for DefaultDivergences,
	// a negative count for none.
	Divergences int

	// LeadLag ranks the DAILY divergences by metrics.LeadLagGaps, the
	// disagreement left once a one-session difference in closing times is
	// forgiven, instead of by |A - B|. It is for two series that follow one
	// index but close at different hours (a Xetra line against an index
	// struck after New York), where every large move lands a session apart
	// and the plain ranking lists nothing but the clock. Leave it off to
	// compare two versions of one file, where a shifted session IS the
	// defect to find. It changes no other figure.
	LeadLag bool
}

// PairStudy compares a candidate series, A, with a reference, B: a
// reconstruction against the real fund, a fund against its index, a
// refreshed file against the version it replaces. Every figure is measured
// on the common window, [Start, End], and every return and gap is a
// FRACTION (0.001 = 0.1 %, or 0.1 point a year for a yearly gap).
//
// The figures come at the cadence both series support: the level figures and
// the calendar years at any cadence, read off each series' close at or
// before each bound; Monthly on the calendar months both quote; Daily on the
// dates both quote, only when both are finer than monthly. A figure that
// cannot be measured is left zero (a nil block, an empty list) and Warnings
// says why, so every value marshals to JSON as a number.
type PairStudy struct {
	A, B PairSide // the candidate and the reference

	Start, End time.Time // the common window: from the later first date to the earlier last one
	SpanYears  float64   // its length in 365.25-day years

	// Shared counts the dates both series quote inside the window; OnlyA and
	// OnlyB those only one of them quotes. Two versions of one daily file
	// share every date.
	Shared, OnlyA, OnlyB int
	// LevelRatio is A's close over B's at Start: 1 when two versions of one
	// file agree there, the factor between them when one is a rescaled copy
	// of the other, and a mere unit ratio between a fund and its index.
	LevelRatio float64
	// FirstDivergence is the first shared date on which A over B departs from
	// its ratio on the first shared date by more than DivergenceTolerance:
	// where a refreshed file starts to tell another story, a rescaling
	// forgiven. It is zero when the two move together on every shared date,
	// or share fewer than two.
	FirstDivergence time.Time `json:",omitzero"`

	// CAGRGap is A's CAGR minus B's over the window, per year.
	CAGRGap float64
	// GapSE is the standard error of a yearly return gap over the window:
	// the tracking error of the coarsest block measured (Monthly, else
	// Daily) over the square root of the years it spans. A CAGRGap inside
	// two of it is not a measurement. Zero when no block was measured.
	GapSE float64
	// LevelGap is A's growth over B's across the window, minus one: where A
	// ends when both are rebased to one level at Start (-0.05 = 5 % below).
	LevelGap float64

	Daily   *PairReturns `json:",omitempty"` // on the dates both quote; nil unless both are finer than monthly
	Monthly *PairReturns `json:",omitempty"` // on the calendar months both quote; nil below three of them

	Years    []PairYear // calendar years, side by side
	Warnings []string   // what makes a figure unreliable (short window, cadence, junctions, currencies...)
}

// PairSide is one series of a pair, described on its own calendar.
type PairSide struct {
	ID          string    // its Symbol
	Currency    string    `json:",omitempty"` // as tagged, empty when unknown
	First, Last time.Time // its whole span, beyond the window
	// PeriodsPerYear is its cadence inside the window, measured
	// (metrics.PeriodsPerYear): 252 daily, 52 weekly, 12 monthly.
	PeriodsPerYear float64
	CAGR           float64 // over the window, from its close at or before Start to its close at or before End
	Volatility     float64 // annualized at its own cadence inside the window; zero under three points
	MaxDrawdown    float64 // the deepest fall inside the window, on its own calendar (-0.25 = -25 %)
}

// PairReturns compares the two series' returns period by period, on one
// calendar of periods both quote (a marketdata.Panel).
type PairReturns struct {
	Periods        int       // returns compared
	PeriodsPerYear float64   // the panel's cadence, what annualizes every figure below
	Start, End     time.Time // the first period's start, the last period's end
	Corr           float64   // Pearson correlation of the returns; zero when either is constant
	VolA, VolB     float64   // annualized volatilities
	VolRatio       float64   // VolA / VolB; zero when B does not move
	TrackingError  float64   // annualized volatility of A's return minus B's
	Beta           float64   // slope of A's returns on B's (least squares, with an intercept)
	Alpha          float64   // that regression's intercept, per year (arithmetic)
	Divergences    []Divergence
}

// Divergence is one period where the two returns disagree, dated.
type Divergence struct {
	Start, End time.Time // the period: from the close at Start to the close at End
	A, B       float64   // the two returns over it
	Gap        float64   // A - B
	// Excess is what ranked the period: |Gap|, or under PairOptions.LeadLag
	// on the daily calendar the disagreement metrics.LeadLagGaps leaves once
	// a one-session clock difference is forgiven (on log returns).
	Excess float64
}

// PairYear is one calendar year of both series, each from its close at or
// before the year's start (or Start) to its close at or before the year's
// end (or End), so the rows chain to the window's total.
type PairYear struct {
	Year    int
	A, B    float64 // the two returns
	Diff    float64 // A - B
	Partial bool    // the window cuts the year: it opens after the prior year's last weekday or closes before its own
}

// DivergenceTolerance is how far, relatively, A over B may move from its
// first shared value before FirstDivergence calls it a divergence: ten parts
// per million, above the rounding of a file written to six decimals and
// below any revision worth reading.
const DivergenceTolerance = 1e-5

// shortWindowYears is the window below which a level gap is flagged as
// noise: two years, the threshold the reconstruction audit uses.
const shortWindowYears = 2

// minSharedCoverage is the share of the sparser series' dates the two must
// both quote for the Daily block to exist: two daily calendars differ by
// their holidays, not by a tenth of their dates.
const minSharedCoverage = 0.9

// Pair compares a, the candidate, with b, the reference, on the window both
// cover. The two are read as they are: convert them to one currency first
// (a Pair of a EUR and a USD series measures the exchange rate too, and says
// so in Warnings), and trim them (marketdata.Trim) to study a sub-window.
// Their Symbols name them and may be equal (two versions of one file).
//
// Every close used must be a positive price: a rate or a yield is a level,
// and its "returns" mean nothing. It is an error when a series is nil or
// empty, when the two do not overlap on two distinct dates, when one of them
// quotes nothing inside the window after its first day (its level at the
// end would be a stale one), and when a close inside the window is not a
// positive number.
func Pair(a, b *marketdata.Series, opt PairOptions) (*PairStudy, error) {
	if a.Len() == 0 || b.Len() == 0 {
		return nil, errors.New("analyze: Pair: empty series")
	}
	start, end, ok := marketdata.CommonWindow(a, b)
	if !ok || !start.Before(end) {
		return nil, fmt.Errorf("analyze: Pair: %s (%s to %s) and %s (%s to %s) do not overlap",
			label(a, "A"), a.First().Date.Format(time.DateOnly), a.Last().Date.Format(time.DateOnly),
			label(b, "B"), b.First().Date.Format(time.DateOnly), b.Last().Date.Format(time.DateOnly))
	}
	st := &PairStudy{Start: start, End: end, SpanYears: yearsBetween(start, end)}
	wa, wb := windowOf(a, start, end, "A"), windowOf(b, start, end, "B")
	for _, s := range []*marketdata.Series{wa, wb} {
		if !s.Last().Date.After(start) {
			return nil, fmt.Errorf("analyze: Pair: %s has no quote after %s inside the window, which ends %s",
				s.Symbol, start.Format(time.DateOnly), end.Format(time.DateOnly))
		}
		for _, p := range s.Points {
			if !(p.Close > 0) || math.IsInf(p.Close, 1) {
				return nil, fmt.Errorf("analyze: Pair: %s closes at %v on %s, not a positive price",
					s.Symbol, p.Close, p.Date.Format(time.DateOnly))
			}
		}
	}
	st.A, st.B = side(a, wa, "A"), side(b, wb, "B")

	la, _ := marketdata.SampleAt(wa, []time.Time{start, end})
	lb, _ := marketdata.SampleAt(wb, []time.Time{start, end})
	st.LevelRatio = la[0] / lb[0]
	st.A.CAGR = cagr(la[1]/la[0], st.SpanYears)
	st.B.CAGR = cagr(lb[1]/lb[0], st.SpanYears)
	st.CAGRGap = st.A.CAGR - st.B.CAGR
	st.LevelGap = la[1]/la[0]/(lb[1]/lb[0]) - 1
	st.identity(wa, wb)
	st.Years = pairYears(wa, wb, start, end)

	st.Warnings = append(st.Warnings, st.contextWarnings(wa, wb)...)
	st.measure(wa, wb, opt)
	switch {
	case st.Monthly != nil:
		st.GapSE = st.Monthly.TrackingError / math.Sqrt(float64(st.Monthly.Periods)/st.Monthly.PeriodsPerYear)
	case st.Daily != nil:
		st.GapSE = st.Daily.TrackingError / math.Sqrt(float64(st.Daily.Periods)/st.Daily.PeriodsPerYear)
	}
	return st, nil
}

// label is a series' Symbol, or fallback when it has none.
func label(s *marketdata.Series, fallback string) string {
	if s.Symbol == "" {
		return fallback
	}
	return s.Symbol
}

// windowOf is s cut to [start, end], keeping its last close before start
// too, so a close "at or before" every bound is always in it. It is named
// by its Symbol (or fallback) and carries s's metadata.
func windowOf(s *marketdata.Series, start, end time.Time, fallback string) *marketdata.Series {
	out := *s
	out.Symbol = label(s, fallback)
	lo := 0
	for lo+1 < len(s.Points) && !s.Points[lo+1].Date.After(start) {
		lo++
	}
	hi := lo
	for hi < len(s.Points) && !s.Points[hi].Date.After(end) {
		hi++
	}
	out.Points = s.Points[lo:hi]
	return &out
}

// side describes one series of the pair on its window w.
func side(s, w *marketdata.Series, fallback string) PairSide {
	ps := PairSide{ID: label(s, fallback), Currency: s.Currency, First: s.First().Date, Last: s.Last().Date}
	dates, values := w.Dates(), w.Values()
	ps.PeriodsPerYear = metrics.PeriodsPerYear(dates)
	if len(values) >= 3 {
		ps.Volatility = metrics.Volatility(metrics.Returns(values), ps.PeriodsPerYear)
	}
	ps.MaxDrawdown = metrics.MaxDrawdown(dates, values).Depth
	return ps
}

// identity counts the dates the two share and finds where their ratio
// first moves.
func (st *PairStudy) identity(wa, wb *marketdata.Series) {
	pa, pb := wa.Points, wb.Points
	i, j := 0, 0
	var ratio0 float64
	for i < len(pa) || j < len(pb) {
		switch {
		case j == len(pb) || i < len(pa) && pa[i].Date.Before(pb[j].Date):
			if !pa[i].Date.Before(st.Start) {
				st.OnlyA++
			}
			i++
		case i == len(pa) || pb[j].Date.Before(pa[i].Date):
			if !pb[j].Date.Before(st.Start) {
				st.OnlyB++
			}
			j++
		default:
			if !pa[i].Date.Before(st.Start) {
				st.Shared++
				r := pa[i].Close / pb[j].Close
				switch {
				case st.Shared == 1:
					ratio0 = r
				case st.FirstDivergence.IsZero() && math.Abs(r/ratio0-1) > DivergenceTolerance:
					st.FirstDivergence = pa[i].Date
				}
			}
			i, j = i+1, j+1
		}
	}
}

// pairYears cuts the window into calendar years and reads both series'
// return over each, from the close at or before its start to the close at
// or before its end.
func pairYears(wa, wb *marketdata.Series, start, end time.Time) []PairYear {
	var bounds []time.Time
	var years []int
	var partial []bool
	for y := start.Year(); y <= end.Year(); y++ {
		lo, hi := yearEnd(y-1), yearEnd(y)
		cutLo, cutHi := false, false
		if lo.Before(start) {
			lo, cutLo = start, start.After(lastWeekday(y-1))
		}
		if hi.After(end) {
			hi, cutHi = end, end.Before(lastWeekday(y))
		}
		if !lo.Before(hi) {
			continue
		}
		bounds = append(bounds, lo, hi)
		years = append(years, y)
		partial = append(partial, cutLo || cutHi)
	}
	la, _ := marketdata.SampleAt(wa, bounds)
	lb, _ := marketdata.SampleAt(wb, bounds)
	out := make([]PairYear, len(years))
	for k, y := range years {
		ra, rb := la[2*k+1]/la[2*k]-1, lb[2*k+1]/lb[2*k]-1
		out[k] = PairYear{Year: y, A: ra, B: rb, Diff: ra - rb, Partial: partial[k]}
	}
	return out
}

// yearEnd is December 31 of year y, 00:00 UTC.
func yearEnd(y int) time.Time { return time.Date(y, time.December, 31, 0, 0, 0, 0, time.UTC) }

// lastWeekday is the last Monday-to-Friday of year y: a series quoting on it
// has closed the year.
func lastWeekday(y int) time.Time {
	d := yearEnd(y)
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, -1)
	}
	return d
}

// contextWarnings says what the numbers cannot show: each series' own
// caveats inside the window, a currency mismatch, a window too short.
func (st *PairStudy) contextWarnings(wa, wb *marketdata.Series) []string {
	var out []string
	for _, s := range []struct {
		name string
		side PairSide
		w    *marketdata.Series
	}{{"A " + st.A.ID, st.A, wa}, {"B " + st.B.ID, st.B, wb}} {
		for _, w := range seriesWarnings(marketdata.Trim(s.w, st.Start, st.End)) {
			out = append(out, s.name+": "+w)
		}
		if s.side.PeriodsPerYear <= 12 && labelsMonthStarts(s.w) {
			out = append(out, s.name+": its periodic dates fall on the first of the month; if they label the PREVIOUS month's close, "+
				"every monthly figure and calendar year is read one month off")
		}
	}
	if ca, cb := st.A.Currency, st.B.Currency; ca != "" && cb != "" && ca != cb {
		out = append(out, fmt.Sprintf("A is in %s and B in %s: every gap includes the exchange rate", ca, cb))
	}
	if st.SpanYears < shortWindowYears {
		out = append(out, fmt.Sprintf("the window spans %.1f years: a yearly gap over it is mostly noise", st.SpanYears))
	}
	return out
}

// labelsMonthStarts reports whether most of s's dates are the first of a
// month: the mark of a publisher that labels a month's close by the month
// it opens (March's close dated 04-01), which every calendar reading takes
// for the next month.
func labelsMonthStarts(s *marketdata.Series) bool {
	firsts := 0
	for _, p := range s.Points {
		if p.Date.Day() == 1 {
			firsts++
		}
	}
	return 2*firsts > len(s.Points)
}

// measure builds the Daily and Monthly blocks the two calendars support,
// warning about the ones they do not.
func (st *PairStudy) measure(wa, wb *marketdata.Series, opt PairOptions) {
	k := opt.Divergences
	if k == 0 {
		k = DefaultDivergences
	}
	a, b := *wa, *wb
	a.Symbol, b.Symbol = "A", "B" // panel columns: the two Symbols may be equal

	switch fineA, fineB := st.A.PeriodsPerYear > 12, st.B.PeriodsPerYear > 12; {
	case fineA && fineB:
		sparser := min(len(a.Points), len(b.Points))
		p, err := marketdata.NewPanel(marketdata.Daily, &a, &b)
		switch {
		case err != nil:
			st.warn("no daily figures: %v", err)
		case p.Len() < 3 || float64(p.Len()+1) < minSharedCoverage*float64(sparser):
			st.warn("no daily figures: the two calendars share %d of the sparser one's %d dates", p.Len()+1, sparser)
		default:
			st.Daily = st.block("daily", p, k, opt.LeadLag)
		}
	case fineA || fineB:
		st.warn("no daily figures: A quotes %s and B %s, so the comparison is monthly",
			cadenceName(st.A.PeriodsPerYear), cadenceName(st.B.PeriodsPerYear))
	}

	p, err := marketdata.NewPanel(marketdata.Monthly, &a, &b)
	switch {
	case err != nil:
		st.warn("no monthly figures: %v", err)
	case p.Len() < 3:
		st.warn("no monthly figures: %d shared months, fewer than three", p.Len())
	default:
		st.Monthly = st.block("monthly", p, k, false)
	}
}

// block measures the two columns of p ("A" and "B"), listing the k periods
// where they disagree most.
func (st *PairStudy) block(name string, p *marketdata.Panel, k int, leadLag bool) *PairReturns {
	ra, rb := p.R[0], p.R[1]
	ppy := p.PeriodsPerYear()
	r := &PairReturns{
		Periods:        p.Len(),
		PeriodsPerYear: ppy,
		Start:          p.Starts[0],
		End:            p.Ends[p.Len()-1],
		Corr:           metrics.Corr(ra, rb),
		VolA:           metrics.Volatility(ra, ppy),
		VolB:           metrics.Volatility(rb, ppy),
		TrackingError:  metrics.TrackingError(ra, rb, ppy),
	}
	if r.VolB > 0 {
		r.VolRatio = r.VolA / r.VolB
	}
	if reg, err := metrics.Regress(ra, rb); err == nil {
		r.Beta, r.Alpha = reg.Betas[0].Value, reg.AnnualAlpha(ppy)
	} else {
		st.warn("no %s beta or alpha: %v", name, err)
	}

	excess := make([]float64, len(ra))
	for t := range ra {
		excess[t] = math.Abs(ra[t] - rb[t])
	}
	if leadLag {
		la, lb := make([]float64, len(ra)), make([]float64, len(rb))
		for t := range ra {
			la[t], lb[t] = math.Log1p(ra[t]), math.Log1p(rb[t])
		}
		excess = metrics.LeadLagGaps(la, lb)
	}
	for _, t := range metrics.HighestK(excess, k) {
		if excess[t] <= DivergenceTolerance {
			break // the two agree on every period left, but for rounding
		}
		r.Divergences = append(r.Divergences, Divergence{
			Start: p.Starts[t], End: p.Ends[t], A: ra[t], B: rb[t], Gap: ra[t] - rb[t], Excess: excess[t],
		})
	}
	return r
}

// warn appends a formatted warning.
func (st *PairStudy) warn(format string, args ...any) {
	st.Warnings = append(st.Warnings, fmt.Sprintf(format, args...))
}

// cadenceName names a measured cadence in words.
func cadenceName(ppy float64) string {
	switch {
	case ppy >= 200:
		return "daily"
	case ppy >= 40:
		return "weekly"
	case ppy >= 10:
		return "monthly"
	}
	return fmt.Sprintf("%.0f times a year", ppy)
}

// yearsBetween is the span from a to b in 365.25-day years, the CAGR
// convention of pkg/metrics.
func yearsBetween(a, b time.Time) float64 { return b.Sub(a).Hours() / 24 / 365.25 }

// cagr annualizes a growth factor over years.
func cagr(growth, years float64) float64 { return math.Pow(growth, 1/years) - 1 }
