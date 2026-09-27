package analyze_test

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// walk is a synthetic daily series: n weekday closes from 2020-01-01, a
// deterministic path with a drift and two cycles, so every statistic has
// something to read.
func walk(t *testing.T, symbol string, n int) *marketdata.Series {
	t.Helper()
	var dates []time.Time
	var closes []float64
	level := 100.0
	for d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC); len(dates) < n; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		i := float64(len(dates))
		level *= 1 + 0.0003 + 0.01*math.Sin(1.3*i) + 0.004*math.Sin(i/23)
		dates, closes = append(dates, d), append(closes, level)
	}
	return series(t, symbol, dates, closes)
}

func series(t *testing.T, symbol string, dates []time.Time, closes []float64) *marketdata.Series {
	t.Helper()
	s, err := marketdata.NewSeries(symbol, dates, closes)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// scaled is s with every close multiplied by k, under another name.
func scaled(t *testing.T, s *marketdata.Series, symbol string, k float64) *marketdata.Series {
	t.Helper()
	closes := s.Values()
	for i := range closes {
		closes[i] *= k
	}
	return series(t, symbol, s.Dates(), closes)
}

func pair(t *testing.T, a, b *marketdata.Series, opt analyze.PairOptions) *analyze.PairStudy {
	t.Helper()
	st, err := analyze.Pair(a, b, opt)
	if err != nil {
		t.Fatal(err)
	}
	// Every study must marshal and render: a NaN would break the first, a
	// render error the second.
	if _, err := json.Marshal(st); err != nil {
		t.Errorf("JSON: %v", err)
	}
	var buf bytes.Buffer
	if err := st.WriteText(&buf); err != nil {
		t.Errorf("WriteText: %v", err)
	}
	for i, line := range strings.Split(buf.String(), "\n") {
		if strings.TrimRight(line, " ") != line {
			t.Errorf("text line %d has trailing blanks: %q", i+1, line)
		}
	}
	return st
}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// Two copies of one file, under one name: every gap is zero, every date
// shared, no divergence to list, and the text says so.
func TestPairIdentity(t *testing.T) {
	a := walk(t, "X", 800)
	st := pair(t, a, walk(t, "X", 800), analyze.PairOptions{})
	if st.Shared != 800 || st.OnlyA != 0 || st.OnlyB != 0 || st.LevelRatio != 1 || !st.FirstDivergence.IsZero() {
		t.Errorf("identity: shared %d, only %d/%d, ratio %v, first divergence %v", st.Shared, st.OnlyA, st.OnlyB, st.LevelRatio, st.FirstDivergence)
	}
	if st.CAGRGap != 0 || st.LevelGap != 0 || st.GapSE != 0 {
		t.Errorf("gaps: CAGR %v, level %v, SE %v", st.CAGRGap, st.LevelGap, st.GapSE)
	}
	for _, r := range []*analyze.PairReturns{st.Daily, st.Monthly} {
		if r == nil {
			t.Fatal("a daily pair must have both blocks")
		}
		near(t, "corr", r.Corr, 1, 1e-12)
		near(t, "beta", r.Beta, 1, 1e-12)
		near(t, "alpha", r.Alpha, 0, 1e-12)
		near(t, "vol ratio", r.VolRatio, 1, 1e-12)
		if r.TrackingError != 0 || len(r.Divergences) != 0 {
			t.Errorf("TE %v, divergences %v", r.TrackingError, r.Divergences)
		}
	}
	if st.Daily.PeriodsPerYear != 252 || st.Daily.Periods != 799 || st.Monthly.PeriodsPerYear != 12 {
		t.Errorf("daily %d at %v, monthly at %v", st.Daily.Periods, st.Daily.PeriodsPerYear, st.Monthly.PeriodsPerYear)
	}
	for _, y := range st.Years {
		if y.Diff != 0 {
			t.Errorf("year %d: %+v", y.Year, y)
		}
	}
	if len(st.Warnings) != 0 {
		t.Errorf("warnings: %q", st.Warnings)
	}
	var buf bytes.Buffer
	if err := st.WriteText(&buf); err != nil || !strings.Contains(buf.String(), "identity    identical on every date") {
		t.Errorf("text (%v):\n%s", err, buf.String())
	}
}

// A copy rescaled by a constant is the same series: the ratio says by how
// much, and nothing diverges.
func TestPairRescale(t *testing.T) {
	a := walk(t, "NEW", 600)
	st := pair(t, a, scaled(t, a, "OLD", 2.5), analyze.PairOptions{})
	near(t, "level ratio", st.LevelRatio, 0.4, 1e-15)
	if !st.FirstDivergence.IsZero() {
		t.Errorf("a rescale diverges on %v", st.FirstDivergence)
	}
	near(t, "CAGR gap", st.CAGRGap, 0, 1e-12)
	near(t, "level gap", st.LevelGap, 0, 1e-12)
	near(t, "daily TE", st.Daily.TrackingError, 0, 1e-12)
	var buf bytes.Buffer
	_ = st.WriteText(&buf)
	if !strings.Contains(buf.String(), "A/B 0.4 at the start and on every shared date: one series up to scale") {
		t.Errorf("text:\n%s", buf.String())
	}

	// A revision from one date on is found on that date.
	b := scaled(t, a, "OLD", 1)
	b.Points[400].Close *= 1.01
	st = pair(t, a, b, analyze.PairOptions{})
	if got, want := st.FirstDivergence, b.Points[400].Date; !got.Equal(want) {
		t.Errorf("first divergence %v, want %v", got, want)
	}
	if d := st.Daily.Divergences; len(d) != 2 || !d[0].End.Equal(b.Points[400].Date) && !d[1].End.Equal(b.Points[400].Date) {
		t.Errorf("a one-print revision is two daily divergences, into and out of it: %+v", d)
	}
}

// A reference one session late: the daily correlation collapses, the
// monthly one does not, and LeadLag clears the daily divergence list of
// what is only the clock.
func TestPairOneSessionLag(t *testing.T) {
	a := walk(t, "XETRA", 700)
	closes := a.Values()
	lagged := append([]float64{closes[0]}, closes[:len(closes)-1]...)
	b := series(t, "INDEX", a.Dates(), lagged)

	plain := pair(t, a, b, analyze.PairOptions{Divergences: 10})
	if plain.Daily.Corr > 0.5 || plain.Monthly.Corr < 0.9 {
		t.Errorf("daily corr %.2f, monthly %.2f: want the clock to break only the first", plain.Daily.Corr, plain.Monthly.Corr)
	}
	if d := plain.Daily.Divergences; len(d) != 10 || d[0].Excess < 0.01 {
		t.Errorf("plain divergences: %+v", d)
	}

	forgiven := pair(t, a, b, analyze.PairOptions{Divergences: 10, LeadLag: true})
	for _, d := range forgiven.Daily.Divergences {
		if d.Excess > 1e-12 && !d.End.Equal(a.Last().Date) && !d.End.Equal(a.Points[1].Date) {
			t.Errorf("LeadLag still convicts %s (excess %v)", d.End.Format(time.DateOnly), d.Excess)
		}
	}
	if forgiven.Daily.Corr != plain.Daily.Corr || forgiven.CAGRGap != plain.CAGRGap {
		t.Error("LeadLag moved a figure other than the divergence ranking")
	}
}

// A daily fund against its month-end index: no daily block, a warning, and
// a monthly block that joins the fund's last trading close of each month
// with the index's calendar month-end, down to the level.
func TestPairMonthlyVersusDaily(t *testing.T) {
	a := walk(t, "FUND", 900)
	m, err := a.Resample(marketdata.Monthly)
	if err != nil {
		t.Fatal(err)
	}
	var dates []time.Time
	for _, p := range m.Points[:m.Len()-1] { // the unfinished last month is no month-end
		d := p.Date
		dates = append(dates, time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, time.UTC))
	}
	b := series(t, "INDEX", dates, m.Values()[:m.Len()-1])

	st := pair(t, a, b, analyze.PairOptions{})
	if st.Daily != nil || st.Monthly == nil {
		t.Fatalf("blocks: daily %v, monthly %v", st.Daily, st.Monthly)
	}
	if !strings.Contains(strings.Join(st.Warnings, "\n"), "A quotes daily and B monthly") {
		t.Errorf("warnings: %q", st.Warnings)
	}
	near(t, "monthly corr", st.Monthly.Corr, 1, 1e-12)
	near(t, "monthly TE", st.Monthly.TrackingError, 0, 1e-12)
	near(t, "CAGR gap", st.CAGRGap, 0, 1e-12)
	near(t, "level gap", st.LevelGap, 0, 1e-12)
	if st.B.PeriodsPerYear != 12 || st.A.PeriodsPerYear != 252 || st.Monthly.Periods != len(dates)-1 {
		t.Errorf("cadences %v/%v, %d months", st.A.PeriodsPerYear, st.B.PeriodsPerYear, st.Monthly.Periods)
	}

	// The years chain to the window's total, the first and last flagged.
	growth := 1.0
	for _, y := range st.Years {
		growth *= 1 + y.A
	}
	near(t, "chained years", growth-1, math.Pow(1+st.A.CAGR, st.SpanYears)-1, 1e-9)
	if first, last := st.Years[0], st.Years[len(st.Years)-1]; !first.Partial || !last.Partial || st.Years[1].Partial {
		t.Errorf("partial flags: %+v ... %+v", first, last)
	}
}

func TestPairWarnings(t *testing.T) {
	a := walk(t, "A", 300)
	b := scaled(t, a, "B", 1)
	a.Currency, b.Currency = "EUR", "USD"
	st := pair(t, a, b, analyze.PairOptions{Divergences: -1})
	all := strings.Join(st.Warnings, "\n")
	for _, want := range []string{"A is in EUR and B in USD", "the window spans 1.1 years"} {
		if !strings.Contains(all, want) {
			t.Errorf("no warning %q in %q", want, all)
		}
	}

	// A monthly series dated on the first of each month.
	var dates []time.Time
	var closes []float64
	for m := range 30 {
		dates, closes = append(dates, time.Date(2020, time.Month(2+m), 1, 0, 0, 0, 0, time.UTC)), append(closes, 100+float64(m%5))
	}
	early := series(t, "EARLY", dates, closes)
	st = pair(t, walk(t, "FUND", 700), early, analyze.PairOptions{})
	if !strings.Contains(strings.Join(st.Warnings, "\n"), "B EARLY: its periodic dates fall on the first of the month") {
		t.Errorf("warnings: %q", st.Warnings)
	}
}

func TestPairErrors(t *testing.T) {
	a := walk(t, "A", 50)
	late := series(t, "LATE", []time.Time{time.Date(2030, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(2030, 1, 3, 0, 0, 0, 0, time.UTC)}, []float64{1, 2})
	zero := scaled(t, a, "Z", 1)
	zero.Points[10].Close = 0
	for name, tc := range map[string]struct {
		a, b *marketdata.Series
		want string
	}{
		"nil":      {a, nil, "empty series"},
		"disjoint": {a, late, "do not overlap"},
		"price":    {a, zero, "Z closes at 0 on"},
	} {
		if _, err := analyze.Pair(tc.a, tc.b, analyze.PairOptions{}); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}
