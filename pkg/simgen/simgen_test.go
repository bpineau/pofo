package simgen

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// fakeFetcher serves canned series, no network.
type fakeFetcher map[string]*marketdata.Series

func (f fakeFetcher) Fetch(id string, from time.Time) (*marketdata.Series, error) {
	return f[id], nil
}

func day(i int) time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
}

// mkSeries builds a daily series from a constant daily growth rate.
func mkSeries(symbol string, n int, dailyGrowth float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + dailyGrowth
	}
	return s
}

// mkWobbly builds a series whose daily returns oscillate around drift,
// non-degenerate variance, deterministic.
func mkWobbly(symbol string, n int, drift, amp float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + drift + amp*math.Sin(float64(i))
	}
	return s
}

// mkLevels builds a series holding a constant level (for rates like ^IRX).
func mkLevels(symbol string, n int, level float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: level})
	}
	return s
}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, want %v (±%v)", name, got, want, tol)
	}
}

func TestBuildFrameConvertsRates(t *testing.T) {
	f := fakeFetcher{
		"EQ":   mkSeries("EQ", 10, 0.01),
		"^IRX": mkLevels("^IRX", 10, 5.04), // 5.04%/yr → 0.02%/day
	}
	fr, err := BuildFrame(f, []string{"EQ", "^IRX"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	near(t, "EQ return", fr.Returns["EQ"][1], 0.01, 1e-12)
	near(t, "accrual ^IRX", fr.Returns["^IRX"][1], 5.04/100/252, 1e-12)
}

func TestCompositeNinetySixty(t *testing.T) {
	// 90% equities (+1%/day) + 60% excess bonds (+0.1%/day − cash) + 10% cash.
	f := fakeFetcher{
		"EQ":   mkSeries("EQ", 5, 0.01),
		"BD":   mkSeries("BD", 5, 0.001),
		"^IRX": mkLevels("^IRX", 5, 2.52), // 0.0001/day
	}
	fr, err := BuildFrame(f, []string{"EQ", "BD", "^IRX"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	values, err := Composite(fr, []Leg{
		{ID: "EQ", Weight: 0.9},
		{ID: "BD", Weight: 0.6, Excess: true},
		{ID: "^IRX", Weight: 0.1},
	}, "^IRX", 0.0)
	if err != nil {
		t.Fatal(err)
	}
	cash := 2.52 / 100 / 252
	want := 0.9*0.01 + 0.6*(0.001-cash) + 0.1*cash
	near(t, "composite return", values[1]/values[0]-1, want, 1e-12)
}

func TestTSMOMGoesLongUptrendShortDowntrend(t *testing.T) {
	n := 400
	f := fakeFetcher{
		"UP":   mkSeries("UP", n, 0.004),
		"DOWN": mkSeries("DOWN", n, -0.004),
		"^IRX": mkLevels("^IRX", n, 0),
	}
	fr, err := BuildFrame(f, []string{"UP", "DOWN", "^IRX"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	values, start, err := TSMOM(fr, TSMOMConfig{
		Markets:     []string{"UP", "DOWN"},
		CashID:      "^IRX",
		Lookback:    252,
		VolWindow:   63,
		Rebalance:   21,
		TargetVol:   0.10,
		MaxLeverage: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if start != 253 {
		t.Errorf("start after warmup: %d", start)
	}
	// Long UP and short DOWN: both legs win.
	if last := values[len(values)-1]; last <= 100 {
		t.Errorf("the strategy should profit on clean trends: %v", last)
	}
}

func TestValidatePerfectTrack(t *testing.T) {
	a := mkWobbly("A", 200, 0.002, 0.005)
	b := mkWobbly("B", 200, 0.002, 0.005)
	v, err := Validate(a, b)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "corr", v.Corr, 1.0, 1e-9)
	near(t, "TE", v.TrackingErr, 0.0, 1e-9)
	near(t, "beta", v.Beta, 1.0, 1e-9)
}

func TestSpliceRealOverComposite(t *testing.T) {
	sim := mkSeries("SIM", 100, 0.001)
	real := &marketdata.Series{Symbol: "REAL", Name: "Real"}
	for i := 50; i < 100; i++ {
		real.Points = append(real.Points, marketdata.Point{Date: day(i), Close: 200 + float64(i)})
	}
	out := Splice(real, sim)
	if len(out.Points) != 100 {
		t.Fatalf("points: %d", len(out.Points))
	}
	if !out.Points[49].Date.Before(out.Points[50].Date) {
		t.Error("date order")
	}
	if out.Points[99].Close != 299 {
		t.Errorf("the real part must stay intact: %v", out.Points[99].Close)
	}
	if out.SimulatedBefore.IsZero() {
		t.Error("SimulatedBefore must be set")
	}
}

func TestSimdataRoundTrip(t *testing.T) {
	dir := t.TempDir()
	sf := &marketdata.SimdataFile{
		ID:     "TESTID",
		Name:   "Test asset",
		Method: "method: x+y",
		Points: []marketdata.Point{
			{Date: day(0), Close: 100},
			{Date: day(1), Close: 101.5},
		},
	}
	if err := marketdata.WriteSimdata(dir, sf); err != nil {
		t.Fatal(err)
	}
	s, ok, err := marketdata.ReadSimdata(dir, "TESTID")
	if err != nil || !ok {
		t.Fatalf("read back: %v, %v", ok, err)
	}
	if s.Name != "Test asset" || len(s.Points) != 2 || s.Points[1].Close != 101.5 {
		t.Fatalf("content: %+v", s)
	}
	if _, ok, _ := marketdata.ReadSimdata(dir, "ABSENT"); ok {
		t.Error("a missing id must return nothing")
	}
}

func TestWithRefDataServesLocalFiles(t *testing.T) {
	dir := t.TempDir()
	err := marketdata.WriteSimdata(dir, &marketdata.SimdataFile{
		ID: "REF-X", Name: "Reference X",
		Points: []marketdata.Point{{Date: day(0), Close: 10}, {Date: day(1), Close: 11}},
	})
	if err != nil {
		t.Fatal(err)
	}
	f := WithRefData(os.DirFS(dir), fakeFetcher{"AUTRE": mkSeries("AUTRE", 5, 0.01)})
	s, err := f.Fetch("REF-X", day(0))
	if err != nil || len(s.Points) != 2 || s.Points[1].Close != 11 {
		t.Fatalf("local reference: %+v, %v", s, err)
	}
	// Not in the directory: fallback.
	if s, err := f.Fetch("AUTRE", day(0)); err != nil || s.Symbol != "AUTRE" {
		t.Fatalf("fallback: %+v, %v", s, err)
	}
}

// mkRegimes builds a deterministic path with three regimes: a long calm
// uptrend (so the trend signal is long), a fortnight of rising volatility,
// then one crash day. It is the shape of every real volatility explosion,
// and the one the risk model has to react to.
func mkRegimes(symbol string, n, rampFrom, crash int, calm, hot, drop float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		var r float64
		switch {
		case i == crash:
			r = drop
		case i >= rampFrom:
			r = hot
			if i%2 == 0 {
				r = -hot
			}
		default:
			r = calm
			if i%2 == 0 {
				r = -calm * 0.5 // calm regime drifts up
			}
		}
		v *= 1 + r
	}
	return s
}

// TestTSMOMCutsRiskIntoAVolatilitySpike pins the risk model down: when
// volatility explodes for a fortnight before a crash day, the book must be
// materially smaller by the time the crash lands. The counterfactual is run
// in the same breath, a risk model so slow it never updates (CovHalfLife far
// beyond the sample), which is what sizing once a month off a flat trailing
// window amounted to: that is how the engine used to lose 12 % in a day
// against a 10 % volatility target while the funds it replicates lost 3 to 4.
func TestTSMOMCutsRiskIntoAVolatilitySpike(t *testing.T) {
	const n, rampFrom, crash = 400, 339, 352
	f := fakeFetcher{
		"A":    mkRegimes("A", n, rampFrom, crash, 0.004, 0.02, -0.08),
		"B":    mkRegimes("B", n, rampFrom, crash, 0.003, 0.02, -0.08),
		"^IRX": mkLevels("^IRX", n, 0),
	}
	fr, err := BuildFrame(f, []string{"A", "B", "^IRX"}, day(0))
	if err != nil {
		t.Fatal(err)
	}
	cfg := TSMOMConfig{
		Markets:     []string{"A", "B"},
		CashID:      "^IRX",
		Lookback:    252,
		VolWindow:   63,
		Rebalance:   21,
		TargetVol:   0.10,
		MaxLeverage: 2,
	}
	worstDay := func(cfg TSMOMConfig) float64 {
		values, _, err := TSMOM(fr, cfg)
		if err != nil {
			t.Fatal(err)
		}
		worst := 0.0
		for i := 1; i < len(values); i++ {
			worst = min(worst, values[i]/values[i-1]-1)
		}
		return worst
	}
	frozen := cfg
	frozen.CovHalfLife = 100000
	reactive, stale := worstDay(cfg), worstDay(frozen)
	t.Logf("worst day: reactive %.2f%%, frozen risk model %.2f%%", reactive*100, stale*100)

	if reactive < stale/2 {
		t.Errorf("the book was not cut into the volatility spike: worst day %.2f%% against %.2f%% with a frozen risk model, want at most half",
			reactive*100, stale*100)
	}
	// And in absolute terms: a 10 %/yr book moves 0.63 % on an average day,
	// so anything past eight of those is a tail the model failed to see.
	if reactive < -8*0.10/math.Sqrt(252) {
		t.Errorf("worst day %.2f%% is more than eight daily sigmas of the volatility target", reactive*100)
	}
}

// TestDensifySparseDonorKeepsTheNAVsAndTheTexture is the weekly-donor trap in
// miniature: a fund that deals once a week would otherwise contribute
// week-sized steps to a daily file, and metrics, which annualize per
// observation, would read them as roughly sqrt(5) times the fund's real
// volatility. The projection must keep every NAV to the cent and take the
// day-to-day moves from the engine.
func TestDensifySparseDonorKeepsTheNAVsAndTheTexture(t *testing.T) {
	const n = 700
	texture := mkWobbly("TEXTURE", n, 2e-4, 0.008)

	// A weekly donor with a path of its own: same dates as every 5th texture
	// point, so each NAV must be reproduced exactly.
	donor := &marketdata.Series{Symbol: "WEEKLY"}
	v := 50.0
	for i := 0; i < n; i += 5 {
		donor.Points = append(donor.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + 0.001 + 0.02*math.Cos(float64(i)/7)
	}

	if !sparse(donor, time.Time{}) {
		t.Fatal("a one-a-week donor must read as sparse")
	}
	if sparse(texture, time.Time{}) {
		t.Fatal("a daily series must not read as sparse")
	}

	out := densify(donor, texture, time.Time{})
	if len(out.Points) < n-10 {
		t.Fatalf("projection returned %d points, want a daily calendar of about %d", len(out.Points), n)
	}
	// Every real NAV survives, at its own date and to the cent.
	scale := out.Points[0].Close / donor.Points[0].Close
	for _, a := range donor.Points {
		got, _, ok := out.At(a.Date)
		if !ok {
			t.Fatalf("NAV of %s lost", a.Date.Format("2006-01-02"))
		}
		if math.Abs(got/(a.Close*scale)-1) > 1e-9 {
			t.Errorf("NAV of %s: %v, want %v", a.Date.Format("2006-01-02"), got, a.Close*scale)
		}
	}
	// The realized daily volatility is the texture's, not the cadence's. Read
	// as daily observations, the raw weekly donor annualizes far above it.
	got := annualVol(out.Points)
	want := annualVol(texture.Points)
	if got > 2*want {
		t.Errorf("projected daily volatility %.1f %%, texture %.1f %%: the cadence is still showing", got*100, want*100)
	}
	if raw := annualVol(donor.Points); raw < 2*want {
		t.Fatalf("the fixture is too tame to be a test: raw weekly volatility %.1f %% against %.1f %%", raw*100, want*100)
	}
}

// annualVol is the standard deviation of a series' point-to-point returns,
// annualized the way pkg/metrics does it (252 observations a year), which is
// exactly the convention a sparse donor would mislead.
func annualVol(ps []marketdata.Point) float64 {
	var rs []float64
	for i := 1; i < len(ps); i++ {
		rs = append(rs, ps[i].Close/ps[i-1].Close-1)
	}
	return stdev(rs) * math.Sqrt(252)
}
