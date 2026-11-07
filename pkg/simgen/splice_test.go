package simgen

import (
	"slices"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func TestSplicePreservesRealMetadata(t *testing.T) {
	real := mkSeries("REAL", 4, 0.01)
	real.Junctions = []time.Time{day(2)}
	real.Warnings = []string{"real source caveat"}
	real.Ends = day(3)
	// No earlier proxy point means every output point is real. The first
	// real quote still gives positive evidence for a zero-length backcast.
	proxy := mkSeries("SIM", 4, 0.02)
	proxy.Warnings = []string{"unused proxy caveat"}
	out := Splice(real, proxy)
	if !out.SimulatedBefore.Equal(day(0)) {
		t.Errorf("all-real frontier = %v, want first real quote", out.SimulatedBefore)
	}
	if !slices.Equal(out.Junctions, real.Junctions) || !out.Ends.Equal(day(3)) || !slices.Equal(out.Warnings, real.Warnings) {
		t.Fatalf("real metadata lost: junctions=%v ends=%v warnings=%v", out.Junctions, out.Ends, out.Warnings)
	}
	out.Warnings[0] = "changed"
	out.Junctions[0] = day(1)
	if real.Warnings[0] != "real source caveat" || !real.Junctions[0].Equal(day(2)) {
		t.Error("splice metadata aliases the real input")
	}
}

func TestSplicePreservesExistingFrontier(t *testing.T) {
	real := &marketdata.Series{
		Symbol: "REAL", SimulatedBefore: day(3),
		Points: []marketdata.Point{{Date: day(2), Close: 100}, {Date: day(3), Close: 101}},
	}
	out := Splice(real, mkSeries("SIM", 4, 0.01))
	if len(out.Points) != 4 {
		t.Errorf("existing reconstruction stopped the earlier splice: %d points", len(out.Points))
	}
	if !out.SimulatedBefore.Equal(day(3)) {
		t.Errorf("splice replaced a known real frontier: %v", out.SimulatedBefore)
	}
}
