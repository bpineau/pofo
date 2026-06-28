package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestParametricSourceMoments(t *testing.T) {
	src := ParametricSource{Mu: 0.04, Sigma: 0.12, Df: 6, Periods: 30}
	rng := rand.New(rand.NewPCG(1, 2))
	var all []float64
	for i := 0; i < 4000; i++ {
		seq := src.Draw(rng)
		if len(seq) != 30 {
			t.Fatalf("len = %d, want 30", len(seq))
		}
		all = append(all, seq...)
	}
	mean, variance := 0.0, 0.0
	for _, x := range all {
		mean += x
	}
	mean /= float64(len(all))
	for _, x := range all {
		variance += (x - mean) * (x - mean)
	}
	variance /= float64(len(all) - 1)
	if math.Abs(mean-0.04) > 0.01 {
		t.Errorf("mean = %.4f, want ~0.04", mean)
	}
	if sd := math.Sqrt(variance); math.Abs(sd-0.12) > 0.01 {
		t.Errorf("stdev = %.4f, want ~0.12", sd)
	}
}

func TestParametricSourceClampsRuin(t *testing.T) {
	src := ParametricSource{Mu: 0.0, Sigma: 2.0, Df: 3, Periods: 50}
	rng := rand.New(rand.NewPCG(7, 7))
	for i := 0; i < 2000; i++ {
		for _, r := range src.Draw(rng) {
			if 1+r < 0 {
				t.Fatalf("return %.4f makes 1+r negative", r)
			}
		}
	}
}
