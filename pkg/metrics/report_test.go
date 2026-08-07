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
	// Three points, one 5-unit inflow on day 1:
	// r1 = (106-5)/100 = 1.01 ; r2 = 110/106 ; TWR = 1.01*110/106 - 1.
	dates := []time.Time{day(0), day(1), day(2)}
	values := []float64{100, 106, 110}
	flows := []Flow{{Date: day(1), Amount: 5}}
	w := Window{Name: "all", From: day(0), To: day(2)}

	rows, sum := Report(dates, values, flows, ReportOptions{Windows: []Window{w}})
	if len(rows) != 1 || !rows[0].OK {
		t.Fatalf("rows: %+v", rows)
	}
	wantTWR := 1.01*110/106 - 1
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
