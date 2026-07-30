package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

// resampling a panel must approximately preserve its own mean return.
func TestBlockBootstrapPreservesMean(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	row := make([]float64, 240)
	src := rand.New(rand.NewPCG(9, 9))
	mean := 0.0
	for i := range row {
		row[i] = 0.005 + 0.04*src.NormFloat64()
		mean += row[i]
	}
	mean /= float64(len(row))
	p := Panel{Returns: [][]float64{row}, Weights: []float64{1}}
	bb := BlockBootstrap{Panel: p, BlockLen: 12, Periods: 360}
	if bb.Len() != 360 {
		t.Fatalf("Len = %d, want 360", bb.Len())
	}
	got := 0.0
	n := 0
	for i := 0; i < 200; i++ {
		for _, r := range bb.Draw(rng) {
			got += r
			n++
		}
	}
	got /= float64(n)
	if math.Abs(got-mean) > 0.003 {
		t.Errorf("resampled mean %.4f, panel mean %.4f", got, mean)
	}
}

func TestStationaryBootstrapLen(t *testing.T) {
	p := Panel{Returns: [][]float64{{0.01, 0.02, -0.01}}, Weights: []float64{1}}
	sb := StationaryBootstrap{Panel: p, MeanBlock: 4, Periods: 20}
	rng := rand.New(rand.NewPCG(1, 1))
	if got := sb.Draw(rng); len(got) != 20 {
		t.Fatalf("len = %d, want 20", len(got))
	}
}
