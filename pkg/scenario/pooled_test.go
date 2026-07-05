package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestPooledBootstrapLengthAndPool(t *testing.T) {
	// Two disjoint series with disjoint value sets, so every drawn value can be
	// attributed to its source series and we can check blocks stay in-series.
	p := PooledBootstrap{
		Series:    [][]float64{{1, 2, 3, 4, 5}, {10, 20, 30, 40, 50}},
		MeanBlock: 3,
		Periods:   200,
	}
	rng := rand.New(rand.NewPCG(1, 2))
	seq := p.Draw(rng)
	if len(seq) != 200 {
		t.Fatalf("len = %d, want 200", len(seq))
	}
	seenA, seenB := false, false
	for _, v := range seq {
		switch {
		case v >= 1 && v <= 5:
			seenA = true
		case v >= 10 && v <= 50:
			seenB = true
		default:
			t.Fatalf("value %v belongs to no series", v)
		}
	}
	if !seenA || !seenB {
		t.Errorf("expected both series to be sampled (A=%v B=%v)", seenA, seenB)
	}
}

func TestPooledBootstrapBlocksStayContiguous(t *testing.T) {
	// A single monotone series: within a block, consecutive draws must step by
	// exactly +1 (the series' own ordering), proving blocks never scramble time.
	series := make([]float64, 100)
	for i := range series {
		series[i] = float64(i)
	}
	p := PooledBootstrap{Series: [][]float64{series}, MeanBlock: 8, Periods: 500}
	seq := p.Draw(rand.New(rand.NewPCG(4, 5)))
	steps, contiguous := 0, 0
	for i := 1; i < len(seq); i++ {
		steps++
		if seq[i] == seq[i-1]+1 {
			contiguous++
		}
	}
	// With mean block 8, roughly 7/8 of transitions continue a block; allow slack.
	if frac := float64(contiguous) / float64(steps); frac < 0.6 {
		t.Errorf("contiguous fraction %.2f too low; blocks not preserved", frac)
	}
}

func TestPooledBootstrapEmptyPool(t *testing.T) {
	p := PooledBootstrap{Series: nil, MeanBlock: 5, Periods: 10}
	seq := p.Draw(rand.New(rand.NewPCG(1, 1)))
	if len(seq) != 10 {
		t.Fatalf("len = %d, want 10", len(seq))
	}
	for _, v := range seq {
		if v != 0 {
			t.Errorf("empty pool should yield zeros, got %v", v)
		}
	}
}

func TestPooledBootstrapPreservesMean(t *testing.T) {
	// Resampling is mean-preserving: the pooled long-run mean of the draws should
	// match the pooled mean of the input series within Monte-Carlo tolerance.
	s1 := []float64{0.05, -0.10, 0.20, 0.00, 0.08}
	s2 := []float64{0.03, 0.03, -0.20, 0.15, 0.10, -0.05}
	want := 0.0
	for _, v := range append(append([]float64{}, s1...), s2...) {
		want += v
	}
	want /= float64(len(s1) + len(s2))

	p := PooledBootstrap{Series: [][]float64{s1, s2}, MeanBlock: 2, Periods: 2000}
	rng := rand.New(rand.NewPCG(7, 8))
	var sum float64
	var n int
	for range 50 {
		for _, v := range p.Draw(rng) {
			sum += v
			n++
		}
	}
	if got := sum / float64(n); math.Abs(got-want) > 0.01 {
		t.Errorf("pooled mean = %.4f, want ~%.4f", got, want)
	}
}
