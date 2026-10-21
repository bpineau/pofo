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
	got, _ := FlowReturns(dates, values, flows)
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
	if stats.PeriodsPerYear != TradingDaysPerYear {
		t.Fatalf("PeriodsPerYear = %v, want %v on a weekday calendar", stats.PeriodsPerYear, TradingDaysPerYear)
	}
	// Without flows, FlowReturns yields the same returns at the same cadence.
	if fr, ppy := FlowReturns(dates, values, nil); ppy != TradingDaysPerYear || len(fr) != len(r) {
		t.Fatalf("FlowReturns = %d returns at %v a year, want %d at %v", len(fr), ppy, len(r), TradingDaysPerYear)
	}
	ppy := stats.PeriodsPerYear
	if got := Volatility(r, ppy); math.Abs(got-stats.Volatility) > 1e-12 {
		t.Fatalf("Volatility = %v, want Compute's %v", got, stats.Volatility)
	}
	if got := Sharpe(r, 0, ppy); math.Abs(got-stats.Sharpe) > 1e-12 {
		t.Fatalf("Sharpe(rf=0) = %v, want Compute's %v", got, stats.Sharpe)
	}
	if got := Sortino(r, 0, ppy); math.Abs(got-stats.Sortino) > 1e-12 {
		t.Fatalf("Sortino(rf=0) = %v, want Compute's %v", got, stats.Sortino)
	}
	// A positive risk-free rate lowers both ratios.
	if Sharpe(r, 0.03, ppy) >= Sharpe(r, 0, ppy) {
		t.Fatal("Sharpe should decrease with a higher risk-free rate")
	}
	if Sortino(r, 0.03, ppy) >= Sortino(r, 0, ppy) {
		t.Fatal("Sortino should decrease with a higher risk-free rate")
	}
}

func TestRatiosUndefined(t *testing.T) {
	if !math.IsNaN(Volatility(nil, TradingDaysPerYear)) || !math.IsNaN(Volatility([]float64{0.01}, TradingDaysPerYear)) {
		t.Fatal("Volatility of fewer than two returns should be NaN")
	}
	if !math.IsNaN(Sharpe([]float64{0, 0, 0}, 0, TradingDaysPerYear)) {
		t.Fatal("Sharpe with zero volatility should be NaN")
	}
	if !math.IsNaN(Sortino([]float64{0.01, 0.02}, 0, TradingDaysPerYear)) {
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

// TestTWROutflowLeavesAtTheClose pins the outflow convention: a line bought
// at 100 that rises to 110 and is sold whole at the next close of 105 earned
// +5 %. Booked at the start of the day, the sale (worth 105) was subtracted
// from the previous close of 110 and left a 0/5 growth factor: a TWR of
// -100 %, for ever.
func TestTWROutflowLeavesAtTheClose(t *testing.T) {
	dates := []time.Time{fday(0), fday(1), fday(2)}
	values := []float64{100, 110, 0}
	flows := []Flow{{Date: fday(2), Amount: -105}}
	got, ok := TWR(dates, values, flows)
	if !ok || math.Abs(got-0.05) > 1e-12 {
		t.Fatalf("TWR = %v, %v; want +5 %% (110 -> 105 before the sale)", got, ok)
	}
	// A partial sale on a down day keeps that day's return whole: 110 -> 99
	// is -10 %, half of the position leaving at the 99 close.
	values = []float64{100, 110, 49.5}
	flows = []Flow{{Date: fday(2), Amount: -49.5}}
	got, _ = TWR(dates, values, flows)
	if want := 0.99 - 1; math.Abs(got-want) > 1e-12 {
		t.Fatalf("TWR with a partial sale = %v, want %v", got, want)
	}
	// A cost posted once the scope is empty is no return at all.
	dates = append(dates, fday(3))
	values = []float64{100, 110, 0, 0}
	flows = []Flow{{Date: fday(2), Amount: -105}, {Date: fday(3), Amount: -1}}
	if got, _ = TWR(dates, values, flows); math.Abs(got-0.05) > 1e-12 {
		t.Fatalf("TWR with a fee after the sale = %v, want +5 %%", got)
	}
}

// TestFlowIndexIsFlowNeutral: selling half of a flat line halves its value
// and loses nothing, so the index stays at 1 and nothing draws down.
func TestFlowIndexIsFlowNeutral(t *testing.T) {
	dates := []time.Time{fday(0), fday(1), fday(2), fday(3)}
	values := []float64{100, 100, 50, 55}
	flows := []Flow{{Date: fday(2), Amount: -50}}
	idx := FlowIndex(dates, values, flows)
	want := []float64{1, 1, 1, 1.1}
	for i := range want {
		if math.Abs(idx[i]-want[i]) > 1e-12 {
			t.Fatalf("FlowIndex = %v, want %v", idx, want)
		}
	}
	if ep := MaxDrawdown(dates, idx); ep.Depth != 0 {
		t.Fatalf("drawdown on the index = %v, want none", ep.Depth)
	}
	if FlowIndex(dates, values[:2], nil) != nil {
		t.Fatal("FlowIndex of mismatched lengths should be nil")
	}
}

// TestFlowReturnsFoldsWeekendMoves: a series that moves at the weekend keeps
// that move in Monday's return instead of losing it.
func TestFlowReturnsFoldsWeekendMoves(t *testing.T) {
	dates := []time.Time{fday(4), fday(5), fday(6), fday(7)}
	values := []float64{100, 110, 110, 121}
	got, _ := FlowReturns(dates, values, nil)
	if len(got) != 1 || math.Abs(got[0]-0.21) > 1e-12 {
		t.Fatalf("returns = %v, want the Friday-to-Monday +21 %%", got)
	}
}

// TestTWRBooksAFlowOnTheNextPoint: a flow dated on a day the series does not
// quote (a weekend deposit on a trading-day series) shows in the next point's
// value, so it is booked there. Matched by exact date only, it was never
// neutralized, and a 50 deposit on a flat 100 read as +50 %.
func TestTWRBooksAFlowOnTheNextPoint(t *testing.T) {
	dates := []time.Time{fday(4), fday(7)} // Friday, Monday
	values := []float64{100, 150}
	flows := []Flow{{Date: fday(5), Amount: 50}, {Date: fday(9), Amount: 1000}} // Saturday; after the last point
	if got, _ := TWR(dates, values, flows); math.Abs(got) > 1e-12 {
		t.Fatalf("TWR = %v, want 0 (the Saturday deposit is not performance)", got)
	}
	if got, _ := FlowReturns(dates, values, flows); len(got) != 1 || math.Abs(got[0]) > 1e-12 {
		t.Fatalf("FlowReturns = %v, want one flat Monday", got)
	}
}
