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
