package decumul

import (
	"math"
	"testing"
)

func TestRecoveryDistribution(t *testing.T) {
	// path: peak 100, dips for 2 years, recovers -> one 2-year episode.
	e := Ensemble{Years: 4, Paths: []PathResult{
		{Wealth: []float64{100, 90, 95, 100, 100}},
	}}
	dist := e.RecoveryTimeDistribution()
	total := 0.0
	for _, b := range dist {
		total += b.Share
	}
	if math.Abs(total-1.0) > 1e-9 {
		t.Errorf("shares sum to %.4f, want 1.0", total)
	}
	// the 2-year bucket should hold the single episode.
	for _, b := range dist {
		if b.Years == 2 && math.Abs(b.Share-1.0) > 1e-9 {
			t.Errorf("2y share = %.3f, want 1.0", b.Share)
		}
	}
}
