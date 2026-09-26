package marketdata

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// mustSeries builds a series whose closes run 1, 2, 3, ... over dates.
func mustSeries(t *testing.T, symbol string, dates ...time.Time) *Series {
	t.Helper()
	values := make([]float64, len(dates))
	for i := range values {
		values[i] = float64(i + 1)
	}
	s, err := NewSeries(symbol, dates, values)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNewSeries(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Skip("no tz database:", err)
	}
	// 00:30 in Paris is still the previous day in UTC: the civil date wins.
	s, err := NewSeries("X", []time.Time{
		time.Date(2024, 1, 2, 0, 30, 0, 0, paris),
		time.Date(2024, 1, 3, 17, 45, 0, 0, time.UTC),
	}, []float64{10, -0.5})
	if err != nil {
		t.Fatal(err)
	}
	if want := []time.Time{date(2024, 1, 2), date(2024, 1, 3)}; !reflect.DeepEqual(s.Dates(), want) {
		t.Errorf("dates = %v, want %v", s.Dates(), want)
	}
	if s.Symbol != "X" || s.Len() != 2 || s.Values()[1] != -0.5 {
		t.Errorf("series = %+v", s)
	}

	empty, err := NewSeries("E", nil, nil)
	if err != nil || empty.Len() != 0 || empty.Returns() != nil {
		t.Errorf("empty: %v, %+v", err, empty)
	}

	for name, tc := range map[string]struct {
		dates  []time.Time
		values []float64
		want   string
	}{
		"length":    {[]time.Time{date(2024, 1, 2)}, []float64{1, 2}, "1 dates for 2 values"},
		"unsorted":  {[]time.Time{date(2024, 1, 3), date(2024, 1, 2)}, []float64{1, 2}, "date 1 (2024-01-02) does not follow"},
		"duplicate": {[]time.Time{date(2024, 1, 2), date(2024, 1, 2).Add(9 * time.Hour)}, []float64{1, 2}, "does not follow"},
		"nan":       {[]time.Time{date(2024, 1, 2)}, []float64{math.NaN()}, "value 0 (2024-01-02) is not finite"},
		"inf":       {[]time.Time{date(2024, 1, 2), date(2024, 1, 3)}, []float64{1, math.Inf(1)}, "not finite"},
	} {
		if _, err := NewSeries("BAD", tc.dates, tc.values); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to contain %q", name, err, tc.want)
		}
	}
}

func TestSeriesSlices(t *testing.T) {
	s := mustSeries(t, "X", date(2024, 1, 2), date(2024, 1, 3), date(2024, 1, 4))
	s.Points[2].Close = 3.3

	values := s.Values()
	values[0] = 99 // a fresh slice: the series is untouched
	if s.Points[0].Close != 1 {
		t.Error("Values aliases the points")
	}
	dates := s.Dates()
	dates[0] = time.Time{}
	if s.Points[0].Date.IsZero() {
		t.Error("Dates aliases the points")
	}

	got := s.Returns()
	want := []float64{1, 0.65} // 2/1-1, 3.3/2-1
	if len(got) != len(want) {
		t.Fatalf("returns = %v", got)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Errorf("returns[%d] = %v, want %v", i, got[i], want[i])
		}
	}

	var null *Series
	if null.Len() != 0 || len(null.Dates()) != 0 || len(null.Values()) != 0 || null.Returns() != nil {
		t.Error("a nil series is not empty")
	}
	if one := mustSeries(t, "O", date(2024, 1, 2)); one.Returns() != nil {
		t.Errorf("single point: returns = %v, want nil", one.Returns())
	}
}

func TestRebase(t *testing.T) {
	s := &Series{
		Symbol: "X", Name: "Fund X", Currency: "EUR", Source: "yahoo",
		SimulatedBefore: date(2024, 1, 3), ProxySymbol: "P",
		Points:    []Point{{date(2024, 1, 2), 50}, {date(2024, 1, 3), 55}, {date(2024, 1, 4), 44}},
		Dividends: []Dividend{{date(2024, 1, 3), 0.7}},
		Junctions: []time.Time{date(2024, 1, 4)},
	}
	r := s.Rebase(100)
	if got := r.Values(); !reflect.DeepEqual(got, []float64{100, 110, 88}) {
		t.Errorf("rebased values = %v", got)
	}
	if !reflect.DeepEqual(r.Dividends, s.Dividends) || !reflect.DeepEqual(r.Junctions, s.Junctions) {
		t.Errorf("dividends or junctions moved: %+v %+v", r.Dividends, r.Junctions)
	}
	if r.Symbol != s.Symbol || r.Name != s.Name || r.Currency != s.Currency || r.Source != s.Source ||
		!r.SimulatedBefore.Equal(s.SimulatedBefore) || r.ProxySymbol != s.ProxySymbol {
		t.Errorf("metadata lost: %+v", r)
	}
	r.Dividends[0].Amount = 9
	if s.Points[0].Close != 50 || s.Dividends[0].Amount != 0.7 {
		t.Error("Rebase modified or aliases its receiver")
	}

	if e := (&Series{Symbol: "E"}).Rebase(100); e.Len() != 0 || e.Symbol != "E" {
		t.Errorf("empty rebase = %+v", e)
	}
	zero := &Series{Points: []Point{{date(2024, 1, 2), 0}, {date(2024, 1, 3), 1}}}
	if got := zero.Rebase(100).Values(); !reflect.DeepEqual(got, []float64{0, 1}) {
		t.Errorf("zero first close: %v, want an unscaled copy", got)
	}
}

// resampleFixture is a daily series across a month-end that falls on a
// weekend after a market holiday: 2024-03-29 is Good Friday, 2024-03-30/31 a
// weekend, so March's last TRADING close is Thursday 2024-03-28. It ends
// mid-April, a partial period.
func resampleFixture(t *testing.T) *Series {
	s := mustSeries(t, "X",
		date(2024, 3, 25), date(2024, 3, 26), date(2024, 3, 27), date(2024, 3, 28), // Good Friday skipped
		date(2024, 4, 1), date(2024, 4, 2), date(2024, 4, 3))
	s.Currency = "USD"
	s.Dividends = []Dividend{{date(2024, 3, 26), 0.1}, {date(2024, 3, 28), 0.2}}
	s.Junctions = []time.Time{date(2024, 3, 26), date(2024, 3, 28)}
	return s
}

func TestResampleMonthly(t *testing.T) {
	s := resampleFixture(t)
	m := s.Resample(Monthly)
	want := []Point{{date(2024, 3, 28), 4}, {date(2024, 4, 3), 7}}
	if !reflect.DeepEqual(m.Points, want) {
		t.Fatalf("monthly = %v, want %v (last trading close, partial last month kept)", m.Points, want)
	}
	if m.Last().Date != date(2024, 4, 3) {
		t.Errorf("partial last period dated %v", m.Last().Date)
	}
	if want := []Dividend{{date(2024, 3, 28), 0.2}}; !reflect.DeepEqual(m.Dividends, want) {
		t.Errorf("dividends = %v, want the kept date's only", m.Dividends)
	}
	// 03-26 (dropped) moves onto March's kept close, 03-28, where it merges
	// with the junction already there: one junction guards the step into it.
	if want := []time.Time{date(2024, 3, 28)}; !reflect.DeepEqual(m.Junctions, want) {
		t.Errorf("junctions = %v, want the dropped one moved onto the period's close", m.Junctions)
	}
	late := resampleFixture(t)
	late.Junctions = []time.Time{date(2024, 3, 27), date(2024, 4, 2), date(2024, 4, 3)}
	if got, want := late.Resample(Monthly).Junctions, []time.Time{date(2024, 3, 28), date(2024, 4, 3)}; !reflect.DeepEqual(got, want) {
		t.Errorf("junctions = %v, want %v (each moved to its period's close, two in April merged)", got, want)
	}
	if got := late.Resample(Yearly).Junctions; !reflect.DeepEqual(got, []time.Time{date(2024, 4, 3)}) {
		t.Errorf("yearly junctions = %v, want the single kept close", got)
	}
	if m.Currency != "USD" || m.Symbol != "X" {
		t.Errorf("metadata lost: %+v", m)
	}
	if s.Len() != 7 || len(s.Dividends) != 2 || len(s.Junctions) != 2 {
		t.Error("Resample modified its receiver")
	}
}

func TestResamplePeriods(t *testing.T) {
	s := resampleFixture(t)
	if got := s.Resample(Quarterly).Dates(); !reflect.DeepEqual(got, []time.Time{date(2024, 3, 28), date(2024, 4, 3)}) {
		t.Errorf("quarterly = %v", got)
	}
	if got := s.Resample(Yearly).Dates(); !reflect.DeepEqual(got, []time.Time{date(2024, 4, 3)}) {
		t.Errorf("yearly = %v", got)
	}
	years := mustSeries(t, "Y", date(2022, 12, 30), date(2023, 6, 30), date(2023, 12, 29), date(2024, 1, 2))
	if got := years.Resample(Yearly).Values(); !reflect.DeepEqual(got, []float64{1, 3, 4}) {
		t.Errorf("yearly across years = %v", got)
	}
	if e := (&Series{}).Resample(Monthly); e.Len() != 0 {
		t.Errorf("empty resample = %+v", e)
	}
	one := mustSeries(t, "O", date(2024, 5, 15))
	if got := one.Resample(Monthly).Points; len(got) != 1 || got[0].Date != date(2024, 5, 15) {
		t.Errorf("single point resample = %v", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("Resample(0) did not panic")
		}
	}()
	s.Resample(0)
}

func TestCommonWindow(t *testing.T) {
	a := mustSeries(t, "A", date(2024, 1, 2), date(2024, 1, 5), date(2024, 1, 9))
	b := mustSeries(t, "B", date(2024, 1, 3), date(2024, 1, 8), date(2024, 1, 10))
	start, end, ok := CommonWindow(a, b)
	if !ok || start != date(2024, 1, 3) || end != date(2024, 1, 9) {
		t.Errorf("window = %v %v %v", start, end, ok)
	}
	if _, _, ok := CommonWindow(); ok {
		t.Error("empty list: ok")
	}
	late := mustSeries(t, "L", date(2024, 2, 1), date(2024, 2, 2))
	if s, e, ok := CommonWindow(a, late); ok || !s.IsZero() || !e.IsZero() {
		t.Errorf("disjoint: %v %v %v", s, e, ok)
	}
	if _, _, ok := CommonWindow(a, nil); ok {
		t.Error("nil series: ok")
	}
	if _, _, ok := CommonWindow(a, &Series{}); ok {
		t.Error("empty series: ok")
	}
	if s, e, ok := CommonWindow(a); !ok || s != date(2024, 1, 2) || e != date(2024, 1, 9) {
		t.Errorf("single series: %v %v %v", s, e, ok)
	}
}

func TestAlignSeries(t *testing.T) {
	a := mustSeries(t, "A", date(2024, 1, 2), date(2024, 1, 3), date(2024, 1, 4), date(2024, 1, 5))
	b := mustSeries(t, "B", date(2024, 1, 3), date(2024, 1, 5), date(2024, 1, 8))

	// Zero from: CommonWindow's start, and no zero level anywhere.
	al, err := AlignSeries([]*Series{a, b}, time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	wantDates := []time.Time{date(2024, 1, 3), date(2024, 1, 4), date(2024, 1, 5), date(2024, 1, 8)}
	if !reflect.DeepEqual(al.Dates, wantDates) {
		t.Errorf("dates = %v, want %v", al.Dates, wantDates)
	}
	if !reflect.DeepEqual(al.IDs, []string{"A", "B"}) {
		t.Errorf("ids = %v", al.IDs)
	}
	for i, levels := range al.Levels {
		if len(levels) != len(al.Dates) {
			t.Errorf("levels[%d] has %d entries for %d dates", i, len(levels), len(al.Dates))
		}
		for k, v := range levels {
			if v == 0 {
				t.Errorf("zero level: %s on %v", al.IDs[i], al.Dates[k])
			}
		}
	}
	if want := [][]float64{{2, 3, 4, 4}, {1, 1, 2, 3}}; !reflect.DeepEqual(al.Levels, want) {
		t.Errorf("levels = %v, want %v", al.Levels, want)
	}

	rets := al.Returns()
	if len(rets) != 2 {
		t.Fatalf("returns for %d series", len(rets))
	}
	for i, r := range rets {
		if len(r) != len(al.Dates)-1 {
			t.Errorf("returns[%d] has %d periods, want %d", i, len(r), len(al.Dates)-1)
		}
	}
	if rets[1][0] != 0 || rets[1][1] != 1 {
		t.Errorf("B returns = %v, want a zero step then the whole move", rets[1])
	}

	col := al.Series(1)
	if col.Symbol != "B" || !reflect.DeepEqual(col.Dates(), al.Dates) || !reflect.DeepEqual(col.Values(), al.Levels[1]) {
		t.Errorf("column = %+v", col)
	}

	// An explicit window, clamped on both sides; bounds read as civil dates.
	al, err = AlignSeries([]*Series{a, b}, date(2024, 1, 4).Add(15*time.Hour), date(2024, 1, 5))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(al.Dates, []time.Time{date(2024, 1, 4), date(2024, 1, 5)}) {
		t.Errorf("windowed dates = %v", al.Dates)
	}
}

func TestAlignSeriesErrors(t *testing.T) {
	a := mustSeries(t, "A", date(2024, 1, 2), date(2024, 1, 3), date(2024, 1, 4))
	late := mustSeries(t, "LATE", date(2024, 1, 3), date(2024, 1, 4))
	gone := mustSeries(t, "GONE", date(2023, 12, 1), date(2023, 12, 4))
	for name, tc := range map[string]struct {
		list     []*Series
		from, to time.Time
		want     string
	}{
		"before first quote":  {[]*Series{a, late}, date(2024, 1, 2), time.Time{}, "LATE starts 2024-01-03, after the window's start 2024-01-02"},
		"after last quote":    {[]*Series{a, gone}, date(2024, 1, 2), time.Time{}, "GONE ends 2023-12-04"},
		"disjoint, zero from": {[]*Series{a, gone}, time.Time{}, time.Time{}, "GONE ends"},
		"no series":           {nil, time.Time{}, time.Time{}, "no series"},
		"empty series":        {[]*Series{a, {Symbol: "HOLLOW"}}, time.Time{}, time.Time{}, `series 1 ("HOLLOW") is empty`},
		"nil series":          {[]*Series{a, nil}, time.Time{}, time.Time{}, "is empty"},
		"inverted window":     {[]*Series{a}, date(2024, 1, 3), date(2024, 1, 2), "before it starts"},
	} {
		al, err := AlignSeries(tc.list, tc.from, tc.to)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to contain %q", name, err, tc.want)
		}
		if al != nil {
			t.Errorf("%s: an Aligned came back with the error", name)
		}
	}
}

func TestAlignedSinglePoint(t *testing.T) {
	a := mustSeries(t, "A", date(2024, 1, 2), date(2024, 1, 3))
	al, err := AlignSeries([]*Series{a}, date(2024, 1, 3), date(2024, 1, 3))
	if err != nil {
		t.Fatal(err)
	}
	if r := al.Returns(); len(r) != 1 || r[0] == nil || len(r[0]) != 0 {
		t.Errorf("one date: returns = %#v, want one empty slice", r)
	}
}
