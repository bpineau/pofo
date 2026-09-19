package decumul

import (
	"math"
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
		_ = p.RunPath(seq, Lives{})
	}
}

// BenchmarkSimulate is one full endpoint-equivalent Monte-Carlo run.
func BenchmarkSimulate(b *testing.B) {
	p := benchPlan()
	for i := 0; i < b.N; i++ {
		_ = p.Simulate(2000, 8, 7)
	}
}

// BenchmarkSolve is one bisection solve (the safe withdrawal), the shape every
// solver endpoint runs: eighteen SimulateOn replays over one set of draws.
func BenchmarkSolve(b *testing.B) {
	p := benchPlan()
	for i := 0; i < b.N; i++ {
		_ = p.Solve(0.05, WithdrawalAxis(0, 150_000), 2000, 8, 7)
	}
}

// BenchmarkSimulateBootstrap is the same run on the page's data-driven column:
// monthly blocks resampled from a portfolio panel and compounded to years, the
// source whose per-draw setup scenario.Prepare hoists.
func BenchmarkSimulateBootstrap(b *testing.B) {
	rows := make([][]float64, 4)
	for a := range rows {
		rows[a] = make([]float64, 480)
		for t := range rows[a] {
			rows[a][t] = 0.003 + 0.03*math.Sin(float64(t*(a+2))/7.0)
		}
	}
	panel := scenario.Panel{Returns: rows, Weights: []float64{0.4, 0.3, 0.2, 0.1}}
	p := benchPlan()
	p.Source = scenario.Compounded{
		Inner: scenario.StationaryBootstrap{Panel: panel, MeanBlock: 24, Periods: 360}, Group: 12}
	b.ResetTimer()
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
		draws := base.Draw(1000, 8, 7)
		for _, wr := range sweepWRs {
			p := base
			p.NeedAnnual = wr * base.Capital
			_ = p.SimulateOn(draws, 8).RuinProb()
		}
	}
}

// benchAmortizePlan exercises the amortization branch with a future pension, so
// cashflowPV runs every year (the quadratic-Pow hotspot the profile flagged).
func benchAmortizePlan() Plan {
	p := benchPlan()
	p.Amortize, p.AmortReturn = true, 0.03
	p.Cashflows = []Cashflow{{FromYear: 15, Annual: 18000}}
	return p
}

// BenchmarkRunPathAmortize measures one amortization path, where cashflowPV is
// called once per year.
func BenchmarkRunPathAmortize(b *testing.B) {
	p, rng := benchAmortizePlan(), rand.New(rand.NewPCG(1, 2))
	seq := p.Source.Draw(rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.RunPath(seq, Lives{})
	}
}

// benchLifetimePlan is the mortality kernel's own shape: a couple, a horizon
// long enough not to truncate the longevity tail, a shorter planning horizon
// and a reverting pension. It is what the FIRE page's mortality sections
// simulate, and it is a far longer path than the fixed-horizon benchmarks.
func benchLifetimePlan() Plan {
	p := benchPlan()
	p.Years, p.PlanHorizon = 55, 40
	p.Source = scenario.ParametricSource{Mu: 0.05, Sigma: 0.15, Df: 5, Periods: 55}
	p.Cashflows = []Cashflow{{FromYear: 15, Annual: 18000, Owner: Self, Reversion: 0.54}}
	p.Lifetime = &Lifetime{Self: Life{Age: 52}, Partner: &Life{Age: 50}, SurvivorSpend: 0.7}
	return p
}

// BenchmarkSimulateLifetime is one endpoint-equivalent run of the stochastic
// lifetime kernel: the lifespans are drawn per path alongside the returns and
// the kernel stops at the household's end.
func BenchmarkSimulateLifetime(b *testing.B) {
	p := benchLifetimePlan()
	for i := 0; i < b.N; i++ {
		_ = p.Simulate(2000, 8, 7)
	}
}

// BenchmarkDrawLifetime isolates the sampling side of that plan: a return
// sequence plus a household lifespan per path.
func BenchmarkDrawLifetime(b *testing.B) {
	p := benchLifetimePlan()
	for i := 0; i < b.N; i++ {
		_ = p.Draw(2000, 8, 7)
	}
}

// benchEnsemble is one simulated ensemble, the input every aggregate below
// reads. Built once, outside the timed loop.
func benchEnsemble(b *testing.B, p Plan) Ensemble {
	b.Helper()
	e := p.Simulate(2000, 8, 7)
	b.ResetTimer()
	return e
}

// BenchmarkOutcome is the headline statistics bundle, computed once per
// simulation on every endpoint and once per point in a Sweep1D.
func BenchmarkOutcome(b *testing.B) {
	e := benchEnsemble(b, benchPlan())
	for i := 0; i < b.N; i++ {
		_ = e.Outcome()
	}
}

// BenchmarkFan is the wealth fan the page's central chart is drawn from: a
// quantile column per year over every path.
func BenchmarkFan(b *testing.B) {
	e := benchEnsemble(b, benchPlan())
	pcts := []float64{0.05, 0.25, 0.50, 0.75, 0.95}
	for i := 0; i < b.N; i++ {
		_ = e.Fan(pcts, 5)
	}
}

// BenchmarkLifeOutcome is the mortality bundle, the lifetime kernel's
// counterpart to Outcome.
func BenchmarkLifeOutcome(b *testing.B) {
	e := benchEnsemble(b, benchLifetimePlan())
	for i := 0; i < b.N; i++ {
		_ = e.LifeOutcome()
	}
}
