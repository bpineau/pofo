package metrics

import (
	"math"
	"testing"
	"time"
)

// fday returns a UTC-midnight date n days after a fixed Monday anchor.
func fday(n int) time.Time {
	return time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
}

func TestTWRNeutralizesFlows(t *testing.T) {
	// Flat market: 100, then a 50 deposit lifts the value to 150.
	dates := []time.Time{fday(0), fday(1)}
	values := []float64{100, 150}
	flows := []Flow{{Date: fday(1), Amount: 50}}
	got, ok := TWR(dates, values, flows)
	if !ok {
		t.Fatal("TWR not ok")
	}
	if math.Abs(got) > 1e-12 {
		t.Fatalf("TWR = %v, want 0 (deposit is not performance)", got)
	}
}

func TestTWRCompoundsWithoutFlows(t *testing.T) {
	dates := []time.Time{fday(0), fday(1), fday(2)}
	values := []float64{100, 110, 99}
	got, ok := TWR(dates, values, nil)
	if !ok {
		t.Fatal("TWR not ok")
	}
	want := 99.0/100.0 - 1
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("TWR = %v, want %v", got, want)
	}
}

func TestTWRDegenerate(t *testing.T) {
	if _, ok := TWR([]time.Time{fday(0)}, []float64{100}, nil); ok {
		t.Fatal("TWR of one point should not be ok")
	}
	if _, ok := TWR([]time.Time{fday(0), fday(1)}, []float64{100}, nil); ok {
		t.Fatal("TWR of mismatched lengths should not be ok")
	}
	// A non-positive base day is skipped, not propagated.
	got, ok := TWR([]time.Time{fday(0), fday(1), fday(2)}, []float64{0, 100, 110}, nil)
	if !ok || math.Abs(got-0.10) > 1e-12 {
		t.Fatalf("TWR skipping zero base = %v, %v; want 0.10, true", got, ok)
	}
}

func TestFlowReturnsDropsWeekendsAndAdjusts(t *testing.T) {
	// fday(4) = Friday, fday(5) = Saturday, fday(6) = Sunday, fday(7) = Monday.
	dates := []time.Time{fday(4), fday(5), fday(6), fday(7)}
	values := []float64{100, 100, 100, 121}
	flows := []Flow{{Date: fday(7), Amount: 11}}
	got := FlowReturns(dates, values, flows)
	if len(got) != 1 {
		t.Fatalf("returns = %v, want exactly the Monday return", got)
	}
	want := 121.0/111 - 1 // 121 over the 100+11 start-of-day base
	if math.Abs(got[0]-want) > 1e-12 {
		t.Fatalf("Monday return = %v, want %v", got[0], want)
	}
}

// TestTWRLargeFlowOnTinyBase pins the start-of-day convention against the
// pathology that motivated it: a small account funded with a large same-day
// contribution. The end-of-day formula (V-F)/V0 divides the flow's first-day
// P/L by the tiny pre-flow value and can even go negative, detonating the
// chain; the start-of-day base keeps the day's return at its true, tiny size.
func TestTWRLargeFlowOnTinyBase(t *testing.T) {
	// 100 sits a week, then 318000 arrives and is invested; it closes the
	// day 68.71 below cost - a ~-0.02% move on the funded base, nothing more.
	dates := []time.Time{fday(0), fday(1)}
	values := []float64{100, 317931.29}
	flows := []Flow{{Date: fday(1), Amount: 318000}}
	got, ok := TWR(dates, values, flows)
	if !ok {
		t.Fatal("TWR not ok")
	}
	want := 317931.29/318100 - 1 // ≈ -0.00053, not -1.69
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("TWR = %v, want %v (start-of-day base, no blow-up)", got, want)
	}
	if got < -0.01 {
		t.Fatalf("TWR = %v: a large same-day flow detonated the chain", got)
	}
}

func TestRatiosMatchComputeAtZeroRF(t *testing.T) {
	dates := make([]time.Time, 0, 300)
	values := make([]float64, 0, 300)
	v := 100.0
	for i := range 300 {
		d := fday(i)
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			continue
		}
		// Deterministic wiggle with drift.
		v *= 1 + 0.0004 + 0.01*math.Sin(float64(i))
		dates = append(dates, d)
		values = append(values, v)
	}
	stats, err := Compute(dates, values)
	if err != nil {
		t.Fatal(err)
	}
	r := Returns(values)
	if got := Volatility(r); math.Abs(got-stats.Volatility) > 1e-12 {
		t.Fatalf("Volatility = %v, want Compute's %v", got, stats.Volatility)
	}
	if got := Sharpe(r, 0); math.Abs(got-stats.Sharpe) > 1e-12 {
		t.Fatalf("Sharpe(rf=0) = %v, want Compute's %v", got, stats.Sharpe)
	}
	if got := Sortino(r, 0); math.Abs(got-stats.Sortino) > 1e-12 {
		t.Fatalf("Sortino(rf=0) = %v, want Compute's %v", got, stats.Sortino)
	}
	// A positive risk-free rate lowers both ratios.
	if Sharpe(r, 0.03) >= Sharpe(r, 0) {
		t.Fatal("Sharpe should decrease with a higher risk-free rate")
	}
	if Sortino(r, 0.03) >= Sortino(r, 0) {
		t.Fatal("Sortino should decrease with a higher risk-free rate")
	}
}

func TestRatiosUndefined(t *testing.T) {
	if !math.IsNaN(Volatility(nil)) || !math.IsNaN(Volatility([]float64{0.01})) {
		t.Fatal("Volatility of fewer than two returns should be NaN")
	}
	if !math.IsNaN(Sharpe([]float64{0, 0, 0}, 0)) {
		t.Fatal("Sharpe with zero volatility should be NaN")
	}
	if !math.IsNaN(Sortino([]float64{0.01, 0.02}, 0)) {
		t.Fatal("Sortino with no downside should be NaN")
	}
}

func TestAnnualize(t *testing.T) {
	// Doubling over exactly two years: (2)^(1/2) - 1.
	got := Annualize(1.0, 731) // 2*365.25 rounded up
	want := math.Pow(2, 365.25/731) - 1
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("Annualize = %v, want %v", got, want)
	}
	if Annualize(0.5, 0) != 0 || Annualize(-1, 100) != 0 {
		t.Fatal("Annualize of degenerate inputs should be 0")
	}
}
