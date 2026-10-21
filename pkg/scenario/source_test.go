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

// TestParametricSourceLowDfIsNormal: Df <= 2 has no finite variance to scale
// to Sigma, so the draw is Normal, as documented. It was a raw Student-t at
// that Df times Sigma: at Df 2 the sample deviation ran several times Sigma
// and the clamp at -100 % fired on draws no Normal makes.
func TestParametricSourceLowDfIsNormal(t *testing.T) {
	for _, df := range []float64{1, 2} {
		src := ParametricSource{Mu: 0.04, Sigma: 0.12, Df: df, Periods: 30}
		rng := rand.New(rand.NewPCG(1, 2))
		var sum, sq float64
		n := 0
		for range 4000 {
			for _, x := range src.Draw(rng) {
				sum, sq, n = sum+x, sq+x*x, n+1
			}
		}
		mean := sum / float64(n)
		sd := math.Sqrt(sq/float64(n) - mean*mean)
		if math.Abs(sd-0.12) > 0.005 || math.Abs(mean-0.04) > 0.005 {
			t.Errorf("Df %v: mean %.4f sd %.4f, want the Normal's 0.04 and 0.12", df, mean, sd)
		}
	}
	m := NewMarkovRegime(0.04, 0.12, 2, 30)
	rng := rand.New(rand.NewPCG(3, 4))
	for range 4000 {
		for _, x := range m.Draw(rng) {
			if x <= -1 {
				t.Fatalf("regime draw at Df 2 hit the -100 %% clamp: %v", x)
			}
		}
	}
}
