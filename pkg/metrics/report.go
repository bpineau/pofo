package metrics

import (
	"math"
	"time"
)

// Window is one named window of a period report; To is inclusive. The
// value at From is the comparison base of the window, so a "ytd" window
// starts on Dec 31 of the previous year.
type Window struct {
	Name     string
	From, To time.Time
}

// StandardWindows returns the usual trailing report windows ending at to:
// 1d, 7d, 1m, 3m, ytd, 1y and prev-yr (the last full calendar year). 7d -
// one calendar week - covers five trading sessions, what "a week" means
// to a human (and what finance UIs label 5D). Month and year arithmetic
// follows Go's AddDate normalization. Callers slice, filter or extend the
// result freely before passing it to Report.
func StandardWindows(to time.Time) []Window {
	dec31 := func(year int) time.Time {
		return time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	}
	return []Window{
		{Name: "1d", From: to.AddDate(0, 0, -1), To: to},
		{Name: "7d", From: to.AddDate(0, 0, -7), To: to},
		{Name: "1m", From: to.AddDate(0, -1, 0), To: to},
		{Name: "3m", From: to.AddDate(0, -3, 0), To: to},
		{Name: "ytd", From: dec31(to.Year() - 1), To: to},
		{Name: "1y", From: to.AddDate(-1, 0, 0), To: to},
		{Name: "prev-yr", From: dec31(to.Year() - 2), To: dec31(to.Year() - 1)},
	}
}

// ReportRow is one measured window of a period report. OK is false when
// the window holds fewer than two points: nothing measurable, the zero
// figures are meaningless and must not be displayed.
type ReportRow struct {
	Window
	TWR  float64 // time-weighted return over the window
	Gain float64 // value change net of external flows, in series units
	OK   bool
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

	CAGR, Vol, Sharpe, Sortino float64
	HasCAGR, HasRisk           bool

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
// daily value series with external flows. Windows whose From pre-dates
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
	returns := FlowReturns(dates, values, flows)
	days := int(math.Round(dates[len(dates)-1].Sub(origin).Hours() / 24))
	sum := ReportSummary{
		TWR:         twr,
		Since:       origin,
		Days:        days,
		MaxDrawdown: MaxDrawdown(dates, values),
	}
	if days >= minRisk && len(returns) >= 2 {
		sum.Vol = Volatility(returns)
		sum.Sharpe = Sharpe(returns, opt.RiskFree)
		sum.Sortino = Sortino(returns, opt.RiskFree)
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

// windowSlice keeps the points dated in [from, to] and the flows strictly
// after from and at or before to: a flow on the base day is already part
// of V0 and must not be neutralized a second time.
func windowSlice(dates []time.Time, values []float64, flows []Flow, from, to time.Time) ([]time.Time, []float64, []Flow) {
	var d []time.Time
	var v []float64
	for i, t := range dates {
		if t.Before(from) || t.After(to) {
			continue
		}
		d = append(d, t)
		v = append(v, values[i])
	}
	var f []Flow
	for _, fl := range flows {
		if fl.Date.After(to) || !fl.Date.After(from) {
			continue
		}
		f = append(f, fl)
	}
	return d, v, f
}
