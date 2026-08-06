package simgen

import (
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// atSeries builds a daily series of n points from a start offset, constant level.
func atSeries(symbol string, startOffset, n int, level float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(startOffset + i), Close: level})
	}
	return s
}

// A configured component is spliced with its longer proxy: the returned series
// starts at the proxy's first date, before the component's own inception.
func TestExtendingFetcherSplicesConfiguredComponent(t *testing.T) {
	f := fakeFetcher{
		"VTMGX":            atSeries("VTMGX", 100, 50, 200), // starts at day(100)
		"^990300-USD-STRD": atSeries("EAFE", 0, 200, 50),    // starts at day(0), earlier
	}

	got, err := extend(f).Fetch("VTMGX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Points[0].Date.Equal(day(0)) {
		t.Errorf("extended series starts %v, want the proxy's start %v", got.Points[0].Date, day(0))
	}
	if got.SimulatedBefore.IsZero() {
		t.Errorf("expected SimulatedBefore to be set after splicing")
	}
}

// A component with no proxy, or a missing proxy, is returned unchanged (the
// wrapper is safe to apply unconditionally).
func TestExtendingFetcherLeavesOthersUnchanged(t *testing.T) {
	f := fakeFetcher{"VFINX": atSeries("VFINX", 0, 50, 100)} // not in longBack; proxy absent anyway

	got, err := extend(f).Fetch("VFINX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Points) != 50 || !got.SimulatedBefore.IsZero() {
		t.Errorf("unconfigured component should pass through unchanged, got %d points, SimulatedBefore=%v", len(got.Points), got.SimulatedBefore)
	}
}

// longIndexOr uses the long refdata series when present (net of fee) and does
// not touch the fallback.
func TestLongIndexOrPrefersRefdata(t *testing.T) {
	f := fakeFetcher{"MSCIWORLD-USD": atSeries("MSCIWORLD-USD", 0, 400, 100)}
	b := longIndexOr("MSCIWORLD-USD", 0, func(Fetcher, time.Time) (*marketdata.Series, error) {
		t.Fatal("fallback must not run when the refdata series is present")
		return nil, nil
	})
	got, err := b(f, time.Time{})
	if err != nil || got == nil || len(got.Points) != 400 {
		t.Fatalf("got %v points, err %v; want the 400-point refdata series", len(got.Points), err)
	}
}

// Without the refdata series, longIndexOr falls back to the proxy Build.
func TestLongIndexOrFallsBack(t *testing.T) {
	called := false
	b := longIndexOr("MSCIWORLD-USD", 0, func(Fetcher, time.Time) (*marketdata.Series, error) {
		called = true
		return atSeries("proxy", 0, 10, 1), nil
	})
	if _, err := b(fakeFetcher{}, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("expected the proxy fallback when refdata is absent")
	}
}

// afterFee applies a continuous annual drag.
func TestAfterFee(t *testing.T) {
	s := atSeries("x", 0, 366, 100) // ~1 year of constant level
	out := afterFee(s, 0.02)
	if last := out.Points[len(out.Points)-1].Close; last < 97.8 || last > 98.2 {
		t.Errorf("after 2%%/yr fee over ~1y, level = %.3f, want ~98", last)
	}
}
