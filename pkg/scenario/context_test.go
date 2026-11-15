package scenario

import (
	"context"
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestDrawContextParityAndCancellation(t *testing.T) {
	for _, src := range []Source{
		ParametricSource{Mu: .04, Sigma: .15, Df: 5, Periods: 100},
		NewMarkovRegime(.04, .15, 5, 100),
		Glidepath{EquityMu: .04, EquitySigma: .15, BondSigma: .05, Df: 5, Periods: 100},
		PooledBootstrap{Series: [][]float64{{.1, -.1, .03}}, MeanBlock: 3, Periods: 100},
		Compounded{Inner: ParametricSource{Mu: .04, Sigma: .15, Df: 5, Periods: 1200}, Group: 12},
		StationaryBootstrap{hist: Sequence{.1, -.1, .03}, MeanBlock: 3, Periods: 100},
		BlockBootstrap{hist: Sequence{.1, -.1, .03}, BlockLen: 2, Periods: 100},
	} {
		rng := func() *rand.Rand { return rand.New(rand.NewPCG(3, 4)) }
		got, err := DrawContext(context.Background(), src, rng())
		if err != nil || !reflect.DeepEqual(got, src.Draw(rng())) {
			t.Fatalf("%T parity: %v", src, err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := DrawContext(ctx, src, rng()); !errors.Is(err, context.Canceled) {
			t.Fatalf("%T: %v", src, err)
		}
	}
}
