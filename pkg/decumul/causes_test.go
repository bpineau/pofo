package decumul

import (
	"math"
	"testing"
)

func TestRuinTiming(t *testing.T) {
	// Horizon 30 -> thirds at [0,10), [10,20), [20,30). Place ruins in each.
	e := Ensemble{Years: 30, Paths: []PathResult{
		{Ruined: true, RuinYear: 3},  // early
		{Ruined: true, RuinYear: 5},  // early
		{Ruined: true, RuinYear: 15}, // mid
		{Ruined: true, RuinYear: 25}, // late
		{Ruined: false, RuinYear: -1},
	}}
	rt := e.RuinTiming()
	if rt.Ruined != 4 {
		t.Fatalf("Ruined = %d, want 4", rt.Ruined)
	}
	if math.Abs(rt.Early-0.5) > 1e-9 || math.Abs(rt.Mid-0.25) > 1e-9 || math.Abs(rt.Late-0.25) > 1e-9 {
		t.Errorf("shares = %+v, want early .5 mid .25 late .25", rt)
	}
	if s := rt.Early + rt.Mid + rt.Late; math.Abs(s-1) > 1e-9 {
		t.Errorf("shares sum to %.3f, want 1", s)
	}
}

func TestRuinTimingNoRuin(t *testing.T) {
	e := Ensemble{Years: 30, Paths: []PathResult{{Ruined: false, RuinYear: -1}}}
	if rt := e.RuinTiming(); rt.Ruined != 0 || rt.Early != 0 {
		t.Errorf("no ruin should give the zero value, got %+v", rt)
	}
}

// RuinShapes must attribute by trajectory: a path halved early is a crash, a
// flat path that eroded is a grind, a path that prospered then ran out is a
// longevity failure. Non-ruined paths are ignored.
func TestRuinShapes(t *testing.T) {
	mk := func(ws ...float64) PathResult { return PathResult{Ruined: true, RuinYear: len(ws) - 1, Wealth: ws} }
	e := Ensemble{Years: 30, Paths: []PathResult{
		mk(100, 60, 45, 30, 10, 0),                    // halved at year 2: crash
		mk(100, 95, 90, 80, 70, 60, 55, 50, 45, 20, 0), // halved year 8 (<=10): crash
		{Ruined: false, Wealth: []float64{100, 200}},   // survivor: ignored
		mk(100, 98, 96, 94, 92, 90, 88, 86, 84, 82, 80, 70, 60, 45, 30, 0), // halved year 13, peak 100: grind
		mk(100, 110, 125, 140, 150, 148, 140, 132, 124, 116, 108, 100, 92, 84, 76, 68, 60, 52, 49, 35, 20, 10, 0), // peak 150, halved only at year 18: longevity
	}}
	s := e.RuinShapes()
	if s.Ruined != 4 {
		t.Fatalf("ruined = %d, want 4", s.Ruined)
	}
	if s.Crash != 0.5 || s.Grind != 0.25 || s.Longevity != 0.25 {
		t.Errorf("shapes = %+v, want crash .5 / grind .25 / longevity .25", s)
	}
	if z := (Ensemble{}).RuinShapes(); z.Ruined != 0 || z.Crash != 0 {
		t.Errorf("empty ensemble must return the zero value, got %+v", z)
	}
}
