package metrics

import (
	"math"
	"testing"
	"time"
)

func TestStandardWindows(t *testing.T) {
	to := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	ws := StandardWindows(to)
	byName := map[string]Window{}
	for _, w := range ws {
		byName[w.Name] = w
	}
	day := func(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) }
	cases := []struct {
		name     string
		from, to time.Time
	}{
		{"1d", day(2026, 7, 3), to},
		{"7d", day(2026, 6, 27), to},
		{"1m", day(2026, 6, 4), to},
		{"3m", day(2026, 4, 4), to},
		{"ytd", day(2025, 12, 31), to},
		{"1y", day(2025, 7, 4), to},
		{"prev-yr", day(2024, 12, 31), day(2025, 12, 31)},
	}
	if len(ws) != len(cases) {
		t.Fatalf("window count: %d", len(ws))
	}
	for _, c := range cases {
		w, ok := byName[c.name]
		if !ok || !w.From.Equal(c.from) || !w.To.Equal(c.to) {
			t.Errorf("%s: got %+v, want [%s, %s]", c.name, w, c.from, c.to)
		}
	}
}

func TestReportRowsAndSummary(t *testing.T) {
	d0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	day := func(i int) time.Time { return d0.AddDate(0, 0, i) }
	// Three points, one 5-unit inflow at the start of day 1:
	// r1 = 106/(100+5) ; r2 = 110/106 ; TWR = r1*r2 - 1 = 110/105 - 1.
	dates := []time.Time{day(0), day(1), day(2)}
	values := []float64{100, 106, 110}
	flows := []Flow{{Date: day(1), Amount: 5}}
	w := Window{Name: "all", From: day(0), To: day(2)}

	rows, sum := Report(dates, values, flows, ReportOptions{Windows: []Window{w}})
	if len(rows) != 1 || !rows[0].OK {
		t.Fatalf("rows: %+v", rows)
	}
	wantTWR := 106.0/105*110/106 - 1
	if math.Abs(rows[0].TWR-wantTWR) > 1e-12 || rows[0].Gain != 5 {
		t.Fatalf("row: %+v, want TWR %v gain 5", rows[0], wantTWR)
	}
	if math.Abs(sum.TWR-wantTWR) > 1e-12 || !sum.Since.Equal(day(0)) || sum.Days != 2 {
		t.Fatalf("summary: %+v", sum)
	}
	// 2 days of track: neither risk nor CAGR figures are meaningful.
	if sum.HasRisk || sum.HasCAGR {
		t.Fatalf("short track must gate annualized figures: %+v", sum)
	}
}

func TestReportBaseDayFlowIsInV0(t *testing.T) {
	d0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	day := func(i int) time.Time { return d0.AddDate(0, 0, i) }
	dates := []time.Time{day(0), day(1)}
	values := []float64{100, 101}
	// A flow ON the window base day is already part of V0: not neutralized.
	flows := []Flow{{Date: day(0), Amount: 100}}
	rows, _ := Report(dates, values, flows,
		ReportOptions{Windows: []Window{{Name: "w", From: day(0), To: day(1)}}})
	if math.Abs(rows[0].TWR-0.01) > 1e-12 || rows[0].Gain != 1 {
		t.Fatalf("base-day flow must not be neutralized: %+v", rows[0])
	}
}

func TestReportDropsPreOriginWindowsAndGatesLongFigures(t *testing.T) {
	d0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 400 // enough for both MinRiskDays and MinCAGRDays defaults
	dates := make([]time.Time, n)
	values := make([]float64, n)
	for i := range dates {
		dates[i] = d0.AddDate(0, 0, i)
		values[i] = 100 * math.Pow(1.0005, float64(i))
	}
	rows, sum := Report(dates, values, nil, ReportOptions{})
	// Default windows end at the last point (2025-02-03). Six standard
	// windows fit; prev-yr [2023-12-31, 2024-12-31] pre-dates the origin
	// and must be dropped (the summary covers it).
	if len(rows) != 6 {
		t.Fatalf("expected 6 windows (prev-yr dropped), got %d: %+v", len(rows), rows)
	}
	if !sum.HasRisk || !sum.HasCAGR || !(sum.CAGR > 0) || !(sum.Vol >= 0) {
		t.Fatalf("long track must fill annualized figures: %+v", sum)
	}
	// Empty input: nothing measurable, no panic.
	if r, s := Report(nil, nil, nil, ReportOptions{}); r != nil || !s.Since.IsZero() {
		t.Fatalf("empty input must return zero values: %+v %+v", r, s)
	}
}

// TestReportDrawdownIgnoresWithdrawals: withdrawing half of a flat account
// is not a -50 % drawdown.
func TestReportDrawdownIgnoresWithdrawals(t *testing.T) {
	d0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	day := func(i int) time.Time { return d0.AddDate(0, 0, i) }
	dates := []time.Time{day(0), day(1), day(2), day(3)}
	values := []float64{100, 100, 50, 45}
	flows := []Flow{{Date: day(2), Amount: -50}}
	_, sum := Report(dates, values, flows, ReportOptions{})
	if math.Abs(sum.MaxDrawdown.Depth-(-0.10)) > 1e-12 {
		t.Fatalf("MaxDrawdown = %v, want -10 %% (the market move, not the withdrawal)", sum.MaxDrawdown.Depth)
	}
}

// TestReportWindowBaseIsTheLastCloseAtFrom: on a trading-day series a window
// whose From is not a session (2022-12-31 is a Saturday) is measured from the
// last close at or before it, the Friday, so the first session of the year
// counts. It used to start from the first point AFTER From and drop that
// session's +10 % from the ytd row.
func TestReportWindowBaseIsTheLastCloseAtFrom(t *testing.T) {
	day := func(m time.Month, d int, y int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
	dates := []time.Time{day(12, 29, 2022), day(12, 30, 2022), day(1, 3, 2023), day(1, 4, 2023)}
	values := []float64{100, 100, 110, 110}
	ytd := Window{Name: "ytd", From: day(12, 31, 2022), To: day(1, 4, 2023)}
	rows, _ := Report(dates, values, nil, ReportOptions{Windows: []Window{ytd}})
	if len(rows) != 1 || !rows[0].OK || math.Abs(rows[0].TWR-0.10) > 1e-12 || rows[0].Gain != 10 {
		t.Fatalf("ytd = %+v, want +10 %% and a gain of 10 from the 2022-12-30 close", rows)
	}
	// A deposit on the Saturday shows in the Tuesday value: it is in the
	// window, not in its base, and must be neutralized.
	values = []float64{100, 100, 160, 160}
	flows := []Flow{{Date: day(12, 31, 2022), Amount: 50}}
	rows, _ = Report(dates, values, flows, ReportOptions{Windows: []Window{ytd}})
	if want := 160.0/150 - 1; math.Abs(rows[0].TWR-want) > 1e-12 || rows[0].Gain != 10 {
		t.Fatalf("ytd with a weekend deposit = %+v, want TWR %v and a gain of 10", rows[0], want)
	}
}

// TestStandardWindowsClampToMonthEnd: a month back from March 31 is the last
// day of February, not Go's normalized March 3, which left the 1m row three
// sessions short; a year back from a leap day is February 28.
func TestStandardWindowsClampToMonthEnd(t *testing.T) {
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
	cases := []struct {
		to   time.Time
		name string
		want time.Time
	}{
		{day(2026, 3, 31), "1m", day(2026, 2, 28)},
		{day(2024, 3, 31), "1m", day(2024, 2, 29)},
		{day(2026, 5, 31), "3m", day(2026, 2, 28)},
		{day(2024, 2, 29), "1y", day(2023, 2, 28)},
		{day(2026, 1, 31), "1m", day(2025, 12, 31)},
		{day(2026, 7, 4), "3m", day(2026, 4, 4)},
	}
	for _, c := range cases {
		for _, w := range StandardWindows(c.to) {
			if w.Name == c.name && !w.From.Equal(c.want) {
				t.Errorf("%s ending %s starts %s, want %s", c.name, c.to.Format(time.DateOnly), w.From.Format(time.DateOnly), c.want.Format(time.DateOnly))
			}
		}
	}
}
