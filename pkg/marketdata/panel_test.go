package marketdata

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// seriesOf builds a series from (date, close) pairs.
func seriesOf(t *testing.T, symbol string, pts ...any) *Series {
	t.Helper()
	var dates []time.Time
	var values []float64
	for i := 0; i < len(pts); i += 2 {
		dates, values = append(dates, pts[i].(time.Time)), append(values, pts[i+1].(float64))
	}
	s, err := NewSeries(symbol, dates, values)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// must returns v, and panics on err, which fails the test that called it
// with the error and its stack.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func wantErr(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("err = %v, want it to contain %q", err, want)
	}
}

func closeTo(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// TestPanelMonthlyCanonicalLabels: a series dated on its last trading days
// (March 2024 ended on Good Friday) and one quoted mid-month meet on the
// calendar month-ends; the daily one's unfinished May is not a period.
func TestPanelMonthlyCanonicalLabels(t *testing.T) {
	a := seriesOf(t, "A",
		date(2024, 1, 15), 100.0, date(2024, 1, 31), 101.0,
		date(2024, 2, 20), 99.0, date(2024, 2, 29), 102.0,
		date(2024, 3, 28), 103.0, date(2024, 4, 30), 99.0,
		date(2024, 5, 15), 120.0)
	b := seriesOf(t, "B",
		date(2023, 12, 29), 50.0, date(2024, 1, 31), 51.0, date(2024, 2, 29), 52.0,
		date(2024, 3, 28), 52.0, date(2024, 4, 30), 50.0, date(2024, 5, 31), 55.0)
	p, err := NewPanel(Monthly, a, b)
	if err != nil {
		t.Fatal(err)
	}
	wantEnds := []time.Time{date(2024, 2, 29), date(2024, 3, 31), date(2024, 4, 30)}
	if !reflect.DeepEqual(p.Ends, wantEnds) {
		t.Fatalf("Ends = %v, want %v", p.Ends, wantEnds)
	}
	if want := []time.Time{date(2024, 1, 31), date(2024, 2, 29), date(2024, 3, 31)}; !reflect.DeepEqual(p.Starts, want) {
		t.Errorf("Starts = %v, want %v", p.Starts, want)
	}
	if !reflect.DeepEqual(p.IDs, []string{"A", "B"}) || p.Freq != Monthly || p.Len() != 3 {
		t.Errorf("panel = %+v", p)
	}
	ra, rb := must(p.Col("A")), must(p.Col("B"))
	closeTo(t, "A Feb", ra[0], 102.0/101-1)
	closeTo(t, "A Apr", ra[2], 99.0/103-1)
	closeTo(t, "B Mar", rb[1], 0)
	if p.PeriodsPerYear() != 12 {
		t.Errorf("PeriodsPerYear = %v", p.PeriodsPerYear())
	}
	ra[0] = 42 // Col hands out a copy
	if p.R[0][0] == 42 {
		t.Error("Col returned the panel's own slice")
	}

	// B's second quarter ends on 05-31, before June's last weekday: it is
	// not a quarter, and neither is A's, so a quarterly panel of both has
	// a single label and no return.
	_, err = NewPanel(Quarterly, a, b)
	wantErr(t, err, "no two period ends shared")
	q, err := NewPanel(Quarterly, b)
	if err != nil {
		t.Fatal(err)
	}
	if want := []time.Time{date(2024, 3, 31)}; !reflect.DeepEqual(q.Ends, want) || q.PeriodsPerYear() != 4 {
		t.Errorf("quarterly Ends = %v, ppy %v", q.Ends, q.PeriodsPerYear())
	}
	closeTo(t, "B Q1", q.R[0][0], 52.0/50-1)

	// March 2024's last weekday, Good Friday, was a holiday: a series that
	// ends on the Thursday before loses its last month, a missing month
	// rather than a guessed one.
	c := seriesOf(t, "C", date(2024, 1, 31), 1.0, date(2024, 2, 29), 1.1, date(2024, 3, 28), 1.2)
	if cp, err := NewPanel(Monthly, c); err != nil || cp.Len() != 1 {
		t.Errorf("Good Friday panel = %+v, %v", cp, err)
	}
}

func TestPanelMonthlyHole(t *testing.T) {
	a := seriesOf(t, "A", date(2024, 1, 31), 1.0, date(2024, 2, 29), 1.0, date(2024, 3, 29), 1.0, date(2024, 4, 30), 1.0)
	b := seriesOf(t, "B", date(2024, 1, 31), 1.0, date(2024, 3, 29), 1.0, date(2024, 4, 30), 1.0)
	_, err := NewPanel(Monthly, a, b)
	wantErr(t, err, "B has no quote in the period ending 2024-02-29")
}

// TestPanelDailyHolidays: a date one series lacks is not a label, and the
// returns around it span both sessions for every column.
func TestPanelDailyHolidays(t *testing.T) {
	a := seriesOf(t, "A", date(2024, 1, 1), 10.0, date(2024, 1, 2), 11.0, date(2024, 1, 3), 12.0, date(2024, 1, 4), 12.0, date(2024, 1, 5), 13.0)
	b := seriesOf(t, "B", date(2024, 1, 2), 20.0, date(2024, 1, 4), 22.0, date(2024, 1, 5), 21.0, date(2024, 1, 8), 25.0)
	p, err := NewPanel(Daily, a, b)
	if err != nil {
		t.Fatal(err)
	}
	if want := []time.Time{date(2024, 1, 4), date(2024, 1, 5)}; !reflect.DeepEqual(p.Ends, want) {
		t.Fatalf("Ends = %v, want %v", p.Ends, want)
	}
	closeTo(t, "A over the missing 3rd", p.R[0][0], 12.0/11-1)
	closeTo(t, "B", p.R[1][0], 22.0/20-1)
	closeTo(t, "B 5th", p.R[1][1], 21.0/22-1)
}

func TestPanelDailyCadence(t *testing.T) {
	var dates []time.Time
	var values []float64
	for d := date(2023, 1, 2); d.Year() == 2023; d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			dates, values = append(dates, d), append(values, 100+float64(len(values)%7))
		}
	}
	s, _ := NewSeries("S", dates, values)
	p, err := NewPanel(Daily, s)
	if err != nil {
		t.Fatal(err)
	}
	if p.PeriodsPerYear() != 252 {
		t.Errorf("daily PeriodsPerYear = %v", p.PeriodsPerYear())
	}
	if got := must(p.Pick([]int{0, 5, 9})).PeriodsPerYear(); got != 252 {
		t.Errorf("picked daily PeriodsPerYear = %v", got)
	}
	st, err := mustPanelSeries(t, p, "S").Stats()
	if err != nil || st.PeriodsPerYear != 252 {
		t.Errorf("stats = %+v, %v", st, err)
	}
}

func mustPanelSeries(t *testing.T, p *Panel, id string) *Series {
	t.Helper()
	s, err := p.Series(id)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestPanelJunction: the period crossing a definition change is dropped for
// every column, and the gap it leaves is not chained over.
func TestPanelJunction(t *testing.T) {
	a := seriesOf(t, "A", date(2024, 1, 31), 1.0, date(2024, 2, 29), 1.1, date(2024, 3, 29), 1.2, date(2024, 4, 30), 1.3)
	b := seriesOf(t, "B", date(2024, 1, 31), 5.0, date(2024, 2, 29), 5.0, date(2024, 3, 29), 9.0, date(2024, 4, 30), 9.0)
	b.Junctions = []time.Time{date(2024, 3, 29)}
	p, err := NewPanel(Monthly, a, b)
	if err != nil {
		t.Fatal(err)
	}
	if want := []time.Time{date(2024, 2, 29), date(2024, 4, 30)}; !reflect.DeepEqual(p.Ends, want) {
		t.Fatalf("Ends = %v, want %v", p.Ends, want)
	}
	closeTo(t, "A Apr", p.R[0][1], 1.3/1.2-1)
	_, err = p.Series("A")
	wantErr(t, err, "period 1 starts 2024-03-31, not where period 0 ended (2024-02-29)")
	if s, err := p.Between(date(2024, 4, 1), time.Time{}).Series("A"); err != nil || s.Len() != 2 {
		t.Errorf("one side of the junction: %v, %v", s, err)
	}
}

func TestPanelErrors(t *testing.T) {
	a := seriesOf(t, "A", date(2024, 1, 31), 1.0, date(2024, 2, 29), 1.1, date(2024, 3, 29), 1.2)
	late := seriesOf(t, "L", date(2024, 3, 29), 1.0, date(2024, 4, 30), 1.1)
	zero := seriesOf(t, "Z", date(2024, 1, 31), 1.0, date(2024, 2, 29), 0.0, date(2024, 3, 29), 1.0)
	unnamed := seriesOf(t, "", date(2024, 1, 31), 1.0, date(2024, 2, 29), 1.1)
	partial := seriesOf(t, "P", date(2024, 5, 2), 1.0, date(2024, 5, 3), 1.0)
	for name, tc := range map[string]struct {
		f    Frequency
		list []*Series
		want string
	}{
		"none":      {Monthly, nil, "no series"},
		"negative":  {-1, []*Series{a}, "negative frequency"},
		"empty":     {Monthly, []*Series{a, {Symbol: "E"}}, "series 1 (E) is empty"},
		"nil":       {Monthly, []*Series{nil}, "series 0 (<nil>) is empty"},
		"unnamed":   {Monthly, []*Series{unnamed}, "no Symbol"},
		"duplicate": {Monthly, []*Series{a, a}, "two series named A"},
		"disjoint":  {Monthly, []*Series{a, late}, "no two period ends shared"},
		"zero":      {Monthly, []*Series{a, zero}, "Z closes at 0 on 2024-02-29, not a positive price"},
		"partial":   {Monthly, []*Series{partial}, "P has no complete period"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewPanel(tc.f, tc.list...)
			wantErr(t, err, tc.want)
		})
	}
}

func TestPanelSelectMixSeries(t *testing.T) {
	eq := seriesOf(t, "EQ", date(2023, 12, 29), 100.0, date(2024, 1, 31), 110.0, date(2024, 2, 29), 99.0, date(2024, 3, 29), 104.0)
	bd := seriesOf(t, "BD", date(2023, 12, 29), 100.0, date(2024, 1, 31), 100.0, date(2024, 2, 29), 102.0, date(2024, 3, 29), 101.0)
	cash := seriesOf(t, "CASH", date(2023, 12, 29), 1.0, date(2024, 1, 31), 1.01, date(2024, 2, 29), 1.02, date(2024, 3, 29), 1.03)
	p, err := NewPanel(Monthly, eq, bd, cash)
	if err != nil {
		t.Fatal(err)
	}

	m, err := p.Mix("60/40", map[string]float64{"EQ": 0.6, "BD": 0.4})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.IDs) != 3 || len(m.IDs) != 4 || m.IDs[3] != "60/40" {
		t.Fatalf("Mix changed its receiver or misnamed: %v / %v", p.IDs, m.IDs)
	}
	mix := must(m.Col("60/40"))
	closeTo(t, "Jan", mix[0], 0.6*0.10)
	closeTo(t, "Feb", mix[1], 0.6*(99.0/110-1)+0.4*0.02)

	lev, err := p.Mix("LEV", map[string]float64{"EQ": 1.5, "CASH": -0.5})
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "leveraged Jan", must(lev.Col("LEV"))[0], 1.5*0.10-0.5*0.01)

	for name, tc := range map[string]struct {
		id   string
		w    map[string]float64
		want string
	}{
		"empty id": {"", map[string]float64{"EQ": 1}, "empty column name"},
		"exists":   {"EQ", map[string]float64{"EQ": 1}, "column EQ already exists"},
		"none":     {"X", nil, "no weights"},
		"unknown":  {"X", map[string]float64{"EQ": 0.5, "GOLD": 0.3, "AAA": 0.2}, "no column AAA, GOLD"},
		"sum":      {"X", map[string]float64{"EQ": 0.6}, "weights sum to 0.6, not 1"},
		"nan":      {"X", map[string]float64{"EQ": math.NaN()}, "weight of EQ is NaN"},
	} {
		_, err := p.Mix(tc.id, tc.w)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}

	s := mustPanelSeries(t, m, "60/40")
	if s.Symbol != "60/40" || s.Len() != 4 || !s.First().Date.Equal(date(2023, 12, 31)) || s.First().Close != 1 {
		t.Errorf("series = %+v", s)
	}
	closeTo(t, "level", s.Last().Close, (1+mix[0])*(1+mix[1])*(1+mix[2]))
	if _, err := m.Series("NOPE"); err == nil {
		t.Error("Series of an absent column did not fail")
	}
	if _, err := must(m.Pick(nil)).Series("EQ"); err == nil {
		t.Error("Series of an empty panel did not fail")
	}

	feb := p.Between(date(2024, 2, 1), date(2024, 2, 29))
	if feb.Len() != 1 || !feb.Starts[0].Equal(date(2024, 1, 31)) {
		t.Errorf("Between = %+v", feb)
	}
	worst := must(p.Pick([]int{1, 0}))
	if !reflect.DeepEqual(worst.Ends, []time.Time{date(2024, 2, 29), date(2024, 1, 31)}) || worst.R[0][0] != p.R[0][1] {
		t.Errorf("Pick = %+v", worst)
	}
	if _, err := worst.Series("EQ"); err == nil {
		t.Error("Series of a sampled panel did not fail")
	}
	if got := must(p.Pick([]int{1, 2})); got.Len() != 2 {
		t.Errorf("consecutive pick = %+v", got)
	} else if _, err := got.Series("EQ"); err != nil {
		t.Errorf("consecutive pick is one path: %v", err)
	}

	_, err = p.Col("NOPE")
	wantErr(t, err, "no column NOPE (the panel holds EQ, BD, CASH)")
	_, err = p.Pick([]int{0, 3})
	wantErr(t, err, "period 3 out of range (the panel has 3)")
	_, err = p.Pick([]int{-1})
	wantErr(t, err, "period -1 out of range")
}

// Track is metrics.Track on two named columns at the panel's cadence.
func TestPanelTrack(t *testing.T) {
	eq := seriesOf(t, "EQ", date(2023, 12, 29), 100.0, date(2024, 1, 31), 110.0, date(2024, 2, 29), 99.0, date(2024, 3, 29), 104.0)
	bd := seriesOf(t, "BD", date(2023, 12, 29), 100.0, date(2024, 1, 31), 100.0, date(2024, 2, 29), 102.0, date(2024, 3, 29), 101.0)
	p := must(NewPanel(Monthly, eq, bd))
	got, err := p.Track("EQ", "BD")
	if err != nil {
		t.Fatal(err)
	}
	if want := must(metrics.Track(p.R[0], p.R[1], 12)); got != want {
		t.Errorf("Track = %+v, want %+v", got, want)
	}
	_, err = p.Track("EQ", "GOLD")
	wantErr(t, err, "no column GOLD (the panel holds EQ, BD)")
	_, err = must(p.Pick([]int{0})).Track("EQ", "BD")
	wantErr(t, err, "1 period(s), at least 2 needed")
}

// Compound reads a daily panel at a horizon of k sessions, skipping a run
// that straddles a junction's gap and a short tail.
func TestPanelCompound(t *testing.T) {
	var da, db []any
	for i := range 12 {
		d := date(2024, 1, 1).AddDate(0, 0, i)
		da = append(da, d, 100*math.Pow(1.01, float64(i)))
		db = append(db, d, 50*math.Pow(1.02, float64(i)))
	}
	a, b := seriesOf(t, "A", da...), seriesOf(t, "B", db...)
	p := must(NewPanel(Daily, a, b))
	w, err := p.Compound(5)
	if err != nil {
		t.Fatal(err)
	}
	// Eleven daily periods: two whole runs of five, the eleventh left out.
	if w.Len() != 2 || !w.Starts[0].Equal(date(2024, 1, 1)) || !w.Ends[1].Equal(date(2024, 1, 11)) {
		t.Fatalf("runs = %v to %v", w.Starts, w.Ends)
	}
	closeTo(t, "A run", w.R[0][0], math.Pow(1.01, 5)-1)
	closeTo(t, "B run", w.R[1][1], math.Pow(1.02, 5)-1)
	closeTo(t, "cadence", w.PeriodsPerYear(), p.PeriodsPerYear()/5)

	// A junction drops the period ending 01-03: the lone period before the
	// gap fills no run, counting restarts after it, and the four left at
	// the end fill none either.
	b.Junctions = []time.Time{date(2024, 1, 3)}
	g := must(must(NewPanel(Daily, a, b)).Compound(5))
	if g.Len() != 1 || !g.Starts[0].Equal(date(2024, 1, 3)) || !g.Ends[0].Equal(date(2024, 1, 8)) {
		t.Errorf("runs around the junction = %v to %v, want 01-03 to 01-08", g.Starts, g.Ends)
	}

	_, err = p.Compound(0)
	wantErr(t, err, "0 sessions")
	_, err = p.Compound(12)
	wantErr(t, err, "no run of 12 unbroken periods among 11")
	m := must(NewPanel(Monthly, seriesOf(t, "M", date(2024, 1, 31), 1.0, date(2024, 2, 29), 1.1, date(2024, 3, 29), 1.2)))
	_, err = m.Compound(3)
	wantErr(t, err, "calendar one")
}

func TestLessFee(t *testing.T) {
	// 1461 days are exactly four 365.25-day years.
	s := seriesOf(t, "S", date(2020, 1, 1), 100.0, date(2024, 1, 1), 150.0)
	s.Dividends = []Dividend{{Date: date(2022, 1, 1), Amount: 1}}
	net := must(s.LessFee(0.01))
	closeTo(t, "first", net.Points[0].Close, 100)
	closeTo(t, "last", net.Points[1].Close, 150*math.Pow(0.99, 4))
	if s.Points[1].Close != 150 || len(net.Dividends) != 1 {
		t.Errorf("LessFee modified its receiver or dropped metadata")
	}
	closeTo(t, "uplift", must(s.LessFee(-0.01)).Points[1].Close, 150*math.Pow(1.01, 4))
	if e := must((&Series{}).LessFee(0.01)); e.Len() != 0 {
		t.Errorf("empty = %+v", e)
	}
	for _, bad := range []float64{1, 85, math.NaN(), math.Inf(-1)} {
		_, err := s.LessFee(bad)
		wantErr(t, err, "LessFee S: annual charge")
	}
}

func TestChange(t *testing.T) {
	s := seriesOf(t, "S",
		date(2007, 12, 31), 100.0, date(2008, 6, 30), 90.0,
		date(2008, 12, 31), 60.0, date(2009, 1, 2), 62.0, date(2009, 12, 31), 75.0)
	got, err := s.Change(date(2008, 1, 1), date(2008, 12, 31))
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, "2008", got, -0.40)
	if got, err := s.Change(date(2008, 7, 5), date(2009, 1, 1)); err != nil || math.Abs(got-(60.0/90-1)) > 1e-12 {
		t.Errorf("mid-episode = %v, %v", got, err)
	}
	// 2010-01-02 is a Saturday, 2010-01-01 a Friday: a series ending on
	// 2009-12-31 does not know the Friday's close.
	_, err = s.Change(date(2009, 1, 1), date(2010, 1, 2))
	wantErr(t, err, "last close 2009-12-31, before 2010-01-02")
	// 2010-01-03 is a Sunday whose last weekday is Friday 2010-01-01.
	s2 := seriesOf(t, "S2", date(2009, 12, 31), 1.0, date(2010, 1, 1), 1.1)
	if got, err := s2.Change(date(2009, 12, 31), date(2010, 1, 3)); err != nil || math.Abs(got-0.1) > 1e-12 {
		t.Errorf("weekend end = %v, %v", got, err)
	}
	_, err = s.Change(date(2007, 12, 30), date(2008, 12, 31))
	wantErr(t, err, "no close at or before 2007-12-30")
	_, err = s.Change(date(2009, 1, 1), date(2008, 1, 1))
	wantErr(t, err, "before it starts")
	s.Junctions = []time.Time{date(2008, 12, 31)}
	_, err = s.Change(date(2008, 7, 1), date(2009, 6, 1))
	wantErr(t, err, "definition junction on 2008-12-31")
	if _, err := s.Change(date(2009, 1, 1), date(2009, 6, 1)); err != nil {
		t.Errorf("after the junction: %v", err)
	}
}

func TestSeriesStats(t *testing.T) {
	s := seriesOf(t, "S", date(2020, 1, 31), 100.0, date(2020, 2, 29), 101.0, date(2020, 3, 31), 99.0, date(2020, 4, 30), 104.0)
	st, err := s.Stats()
	if err != nil || st.PeriodsPerYear != 12 || st.MaxDrawdown >= 0 {
		t.Errorf("stats = %+v, %v", st, err)
	}
	_, err = seriesOf(t, "R", date(2020, 1, 31), 1.0, date(2020, 2, 29), -0.5).Stats()
	wantErr(t, err, "marketdata: R:")
}
