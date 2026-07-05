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
