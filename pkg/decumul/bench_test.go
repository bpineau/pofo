package decumul

import (
	"math/rand/v2"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// benchPlan is a representative 30-year plan on a parametric source, the shape
// the FIRE web endpoints simulate thousands of times per page render.
func benchPlan() Plan {
	return Plan{
		Capital: 1_000_000, NeedAnnual: 40000, Years: 30,
		Tax:    CTOFlatTax{Rate: 0.30},
		Buffer: BufferSleeve{Years: 2},
		Source: scenario.ParametricSource{Mu: 0.05, Sigma: 0.15, Df: 5, Periods: 30},
	}
}

// BenchmarkDraw isolates the cost of sampling one return sequence. It is about
// a third of a path's total cost, which is why the sweep endpoints draw once
// and replay via SimulateOn (see BenchmarkSweep*).
func BenchmarkDraw(b *testing.B) {
	src := benchPlan().Source
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkRunPath isolates the decumulation kernel on a pre-drawn sequence.
func BenchmarkRunPath(b *testing.B) {
	p, rng := benchPlan(), rand.New(rand.NewPCG(1, 2))
	seq := p.Source.Draw(rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.RunPath(seq)
	}
}

// BenchmarkSimulate is one full endpoint-equivalent Monte-Carlo run.
func BenchmarkSimulate(b *testing.B) {
	p := benchPlan()
	for i := 0; i < b.N; i++ {
		_ = p.Simulate(2000, 8, 7)
	}
}

// sweepWRs mirrors a frontier sweep: eleven points over a fixed Source.
var sweepWRs = []float64{0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.055, 0.06, 0.065, 0.07}

// BenchmarkSweepRedraw is the naive sweep that re-samples the paths at every
// point (what the endpoints did before draw-sharing).
func BenchmarkSweepRedraw(b *testing.B) {
	base := benchPlan()
	for i := 0; i < b.N; i++ {
		for _, wr := range sweepWRs {
			p := base
			p.NeedAnnual = wr * base.Capital
			_ = p.Simulate(1000, 8, 7).RuinProb()
		}
	}
}

// BenchmarkSweepShared draws once and replays via SimulateOn, the pattern the
// frontier/sensitivity/policy endpoints now use. It should beat the redraw
// variant by the Draw share of a path's cost.
func BenchmarkSweepShared(b *testing.B) {
	base := benchPlan()
	for i := 0; i < b.N; i++ {
		seqs := base.DrawPaths(1000, 8, 7)
		for _, wr := range sweepWRs {
			p := base
			p.NeedAnnual = wr * base.Capital
			_ = p.SimulateOn(seqs, 8).RuinProb()
		}
	}
}
