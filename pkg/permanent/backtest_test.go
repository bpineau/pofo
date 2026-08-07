package permanent

import (
	"testing"
	"time"
)

func mon(y, m int) time.Time { return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC) }

func TestSimulateLagsRegime(t *testing.T) {
	pm := DefaultParams()
	// A paradise regime in Jan drives Feb's return; a hell regime in Feb drives
	// March's. Give equity +10% and everything else 0 so the equity weight is
	// directly readable from the tactical return.
	regimes := []Regime{
		{Date: mon(2000, 1), GrowthBreadth: 1, InflationBreadth: 0},
		{Date: mon(2000, 2), GrowthBreadth: 0, InflationBreadth: 1},
	}
	ar := AssetReturns{
		Dates:  []time.Time{mon(2000, 2), mon(2000, 3)},
		Equity: []float64{0.10, 0.10},
		Bonds:  []float64{0, 0},
		Cash:   []float64{0, 0},
		Gold:   []float64{0, 0},
	}
	res, err := Simulate(regimes, ar, pm)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	if len(res.Dates) != 2 {
		t.Fatalf("got %d months, want 2", len(res.Dates))
	}
	// Feb uses the paradise (Jan) regime: equity weight 1 => tactical = +10%.
	if !almost(res.Tactical[0], 0.10) {
		t.Fatalf("Feb tactical = %v, want 0.10 (paradise equity)", res.Tactical[0])
	}
	// March uses the hell (Feb) regime: equity weight 0 => tactical = 0.
	if !almost(res.Tactical[1], 0) {
		t.Fatalf("March tactical = %v, want 0 (hell equity)", res.Tactical[1])
	}
	// Static is always 25/25/25/25 => +2.5% each month.
	if !almost(res.Static[0], 0.025) {
		t.Fatalf("static = %v, want 0.025", res.Static[0])
	}
}

func TestSimulateSkipsPreRegimeMonths(t *testing.T) {
	regimes := []Regime{{Date: mon(2000, 6), GrowthBreadth: 0.5, InflationBreadth: 0.5}}
	ar := AssetReturns{
		Dates:  []time.Time{mon(2000, 1), mon(2000, 7)},
		Equity: []float64{0.05, 0.05}, Bonds: []float64{0, 0}, Cash: []float64{0, 0}, Gold: []float64{0, 0},
	}
	res, err := Simulate(regimes, ar, DefaultParams())
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}
	if len(res.Dates) != 1 || !res.Dates[0].Equal(mon(2000, 7)) {
		t.Fatalf("expected only the post-regime month, got %v", res.Dates)
	}
}

func TestComputeStats(t *testing.T) {
	// +10%, -10%: one down month underwater, drawdown 10%.
	s := Compute([]float64{0.10, -0.10})
	if s.Months != 2 {
		t.Fatalf("months = %d", s.Months)
	}
	if s.LongestUnderwater != 1 || s.UnderwaterFraction != 0.5 {
		t.Fatalf("underwater longest=%d frac=%v", s.LongestUnderwater, s.UnderwaterFraction)
	}
	if s.MaxDrawdown > -0.09 || s.MaxDrawdown < -0.11 {
		t.Fatalf("drawdown = %v, want ~-0.10", s.MaxDrawdown)
	}
}

func TestSimulateErrors(t *testing.T) {
	if _, err := Simulate(nil, AssetReturns{}, DefaultParams()); err == nil {
		t.Fatal("expected error on empty returns")
	}
	ar := AssetReturns{Dates: []time.Time{mon(2000, 1)}, Equity: []float64{0.01}} // bonds/cash/gold wrong length
	if _, err := Simulate([]Regime{{Date: mon(1999, 1)}}, ar, DefaultParams()); err == nil {
		t.Fatal("expected error on ragged returns")
	}
}
