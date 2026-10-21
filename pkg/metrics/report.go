package metrics

import (
	"math"
	"sort"
	"time"
)

// Window is one named window of a period report; To is inclusive. The
// value at From is the comparison base of the window, so a "ytd" window
// starts on Dec 31 of the previous year; on a day the series does not quote
// (a weekend, a holiday) the base is the last value before it, the close the
// window's first session moves from.
type Window struct {
	Name     string    // label, such as "ytd" or "1y"
	From, To time.Time // the comparison base and the inclusive end
}

// StandardWindows returns the usual trailing report windows ending at to:
// 1d, 7d, 1m, 3m, ytd, 1y and prev-yr (the last full calendar year). 7d -
// one calendar week - covers five trading sessions, what "a week" means
// to a human (and what finance UIs label 5D). Months and years are counted
// on the calendar and clamped to the end of a shorter month, as a
// spreadsheet's EDATE does: one month before March 31 is February 28 (29 in a
// leap year), never Go's AddDate normalization to March 3, which left a
// month-end "1m" row three days short. Callers slice, filter or extend the
// result freely before passing it to Report.
func StandardWindows(to time.Time) []Window {
	dec31 := func(year int) time.Time {
		return time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	}
	return []Window{
		{Name: "1d", From: to.AddDate(0, 0, -1), To: to},
		{Name: "7d", From: to.AddDate(0, 0, -7), To: to},
		{Name: "1m", From: monthsBefore(to, 1), To: to},
		{Name: "3m", From: monthsBefore(to, 3), To: to},
		{Name: "ytd", From: dec31(to.Year() - 1), To: to},
		{Name: "1y", From: monthsBefore(to, 12), To: to},
		{Name: "prev-yr", From: dec31(to.Year() - 2), To: dec31(to.Year() - 1)},
	}
}

// monthsBefore is t moved n calendar months back, its day clamped to the
// last day of the target month (March 31 less one month is February 28 or
// 29), the time of day and location kept.
func monthsBefore(t time.Time, n int) time.Time {
	y, m, d := t.Date()
	first := time.Date(y, m-time.Month(n), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(d, last)-1)
}

// ReportRow is one measured window of a period report. OK is false when
// the window holds fewer than two points: nothing measurable, the zero
// figures are meaningless and must not be displayed.
type ReportRow struct {
	Window
	TWR  float64 // time-weighted return over the window
	Gain float64 // value change net of external flows, in series units
	OK   bool    // the window holds at least two points; otherwise ignore the row
}

// ReportSummary describes the whole track record of the series passed to
// Report - "inception" semantics therefore belong to the caller: pass the
// ownership window to measure a holding, the full history to measure the
// asset. The annualized figures are gated (HasCAGR, HasRisk): annualizing
// a return earned over a few days compounds noise into absurdity.
type ReportSummary struct {
	TWR   float64   // cumulative TWR since the first point
	Since time.Time // first point of the series
	Days  int       // calendar span of the series

	CAGR, Vol, Sharpe, Sortino float64 // annualized FRACTIONS, as in Stats; read them only under their gate
	HasCAGR, HasRisk           bool    // the span is long enough to annualize the return, the risk figures

	// MaxDrawdown is the deepest drawdown of the whole record, measured on
	// the FlowIndex rather than on the value: a withdrawal or a sale lowers
	// the value and loses nothing, so it is never a drawdown.
	MaxDrawdown Episode
}

// ReportOptions parameterizes Report. The zero value keeps every default:
// StandardWindows of the last point, no risk-free rate, the customary
// track-record gates.
type ReportOptions struct {
	Windows     []Window // nil: StandardWindows(last point date)
	RiskFree    float64  // annualized risk-free rate for Sharpe and Sortino
	MinRiskDays int      // track needed for Vol/Sharpe/Sortino; 0: 90
	MinCAGRDays int      // track needed for CAGR; 0: 365
}

// Track-record floors under which annualized figures are hidden: about a
// quarter of daily returns for the risk statistics, a full year for a
// compound *annual* growth rate.
const (
	defaultMinRiskDays = 90
	defaultMinCAGRDays = 365
)

// Report builds the standard period table plus summary statistics for a
// value series with external flows, typically daily; the risk figures are
// annualized at the cadence FlowReturns measures, and every figure but the
// money gains reads the flow-neutral FlowIndex. Windows whose From pre-dates
// the first point are dropped (the summary covers them); windows are
// otherwise reported even when flat. Empty or single-point input returns
// (nil, zero ReportSummary): nothing measurable, no error.
func Report(dates []time.Time, values []float64, flows []Flow, opt ReportOptions) ([]ReportRow, ReportSummary) {
	if len(dates) != len(values) || len(values) < 2 {
		return nil, ReportSummary{}
	}
	windows := opt.Windows
	if windows == nil {
		windows = StandardWindows(dates[len(dates)-1])
	}
	minRisk, minCAGR := opt.MinRiskDays, opt.MinCAGRDays
	if minRisk == 0 {
		minRisk = defaultMinRiskDays
	}
	if minCAGR == 0 {
		minCAGR = defaultMinCAGRDays
	}

	origin := dates[0]
	var rows []ReportRow
	for _, w := range windows {
		if w.From.Before(origin) {
			continue
		}
		rows = append(rows, reportRow(w, dates, values, flows))
	}

	twr, _ := TWR(dates, values, flows)
	returns, ppy := FlowReturns(dates, values, flows)
	days := int(math.Round(dates[len(dates)-1].Sub(origin).Hours() / 24))
	sum := ReportSummary{
		TWR:         twr,
		Since:       origin,
		Days:        days,
		MaxDrawdown: MaxDrawdown(dates, FlowIndex(dates, values, flows)),
	}
	if days >= minRisk && len(returns) >= 2 {
		sum.Vol = Volatility(returns, ppy)
		sum.Sharpe = Sharpe(returns, opt.RiskFree, ppy)
		sum.Sortino = Sortino(returns, opt.RiskFree, ppy)
		sum.HasRisk = true
	}
	if days >= minCAGR {
		sum.CAGR = Annualize(twr, days)
		sum.HasCAGR = true
	}
	return rows, sum
}

func reportRow(w Window, dates []time.Time, values []float64, flows []Flow) ReportRow {
	d, v, f := windowSlice(dates, values, flows, w.From, w.To)
	row := ReportRow{Window: w}
	twr, ok := TWR(d, v, f)
	if !ok {
		return row
	}
	net := 0.0
	for _, fl := range f {
		net += fl.Amount
	}
	row.TWR, row.Gain, row.OK = twr, v[len(v)-1]-v[0]-net, true
	return row
}

// windowSlice keeps the points from the base, the last point at or before
// from (the value the window is measured against: a Friday close for a ytd
// window opening on a Saturday), to the last point at or before to, and the
// flows strictly after the base point and at or before to: a flow on or
// before the base point is already part of V0 and must not be neutralized a
// second time. Report has dropped every window whose from precedes the first
// point, so the base exists.
func windowSlice(dates []time.Time, values []float64, flows []Flow, from, to time.Time) ([]time.Time, []float64, []Flow) {
	lo := sort.Search(len(dates), func(i int) bool { return dates[i].After(from) }) - 1
	hi := sort.Search(len(dates), func(i int) bool { return dates[i].After(to) })
	if lo < 0 || lo >= hi {
		return nil, nil, nil
	}
	d, v := dates[lo:hi], values[lo:hi]
	var f []Flow
	for _, fl := range flows {
		if fl.Date.After(to) || !fl.Date.After(d[0]) {
			continue
		}
		f = append(f, fl)
	}
	return d, v, f
}
