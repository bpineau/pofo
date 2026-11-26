package metrics

import (
	"math"
	"testing"
	"time"
)

func TestIRRSingleFlow(t *testing.T) {
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// 1000 invested, worth 1210 two years later: IRR = 10 %/year.
	r, ok := IRR([]time.Time{d0}, []float64{-1000}, d0.AddDate(2, 0, 0), 1210)
	if !ok || math.Abs(r-0.10) > 1e-4 {
		t.Fatalf("IRR = %v ok=%v, want 0.10", r, ok)
	}
}

func TestIRRWithContribution(t *testing.T) {
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	d1 := d0.AddDate(1, 0, 0)
	// 1000 now, 1000 in a year, all flat: final 2000 → IRR = 0.
	r, ok := IRR([]time.Time{d0, d1}, []float64{-1000, -1000}, d0.AddDate(2, 0, 0), 2000)
	if !ok || math.Abs(r) > 1e-6 {
		t.Fatalf("IRR = %v ok=%v, want 0", r, ok)
	}
}

func TestIRRNoSolution(t *testing.T) {
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, ok := IRR([]time.Time{d0}, []float64{1000}, d0.AddDate(1, 0, 0), 1000); ok {
		t.Fatal("all-positive flows must not have an IRR")
	}
}

func TestIRRNeedsCapitalInvestedOverTime(t *testing.T) {
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name  string
		dates []time.Time
		flows []float64
		end   time.Time
		value float64
	}{
		{"same-day investment and valuation", []time.Time{d}, []float64{-100}, d, 100},
		{"no cash movements", []time.Time{d}, []float64{0}, d.AddDate(1, 0, 0), 0},
		{"investments immediately refunded", []time.Time{d, d, d.AddDate(1, 0, 0)}, []float64{-100, 100, -50}, d.AddDate(1, 0, 0), 50},
	} {
		if r, ok := IRR(tc.dates, tc.flows, tc.end, tc.value); ok {
			t.Errorf("%s: reported annual return %v", tc.name, r)
		}
	}
}

func TestIRRLongInvestmentClosedBeforeTerminalDate(t *testing.T) {
	start := time.Date(1850, 1, 1, 0, 0, 0, 0, time.UTC)
	sold := start.AddDate(170, 0, 0)
	r, ok := IRR([]time.Time{start, sold}, []float64{-100, 200}, sold.AddDate(1, 0, 0), 0)
	want := math.Pow(2, 365.25*24/sold.Sub(start).Hours()) - 1
	if !ok || math.Abs(r-want) > 1e-9 {
		t.Fatalf("IRR = %v ok=%v, want %v", r, ok, want)
	}
}
