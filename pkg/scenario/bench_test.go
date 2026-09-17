package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

// benchPanel is a portfolio-shaped panel: four assets of monthly real returns
// over forty years, the input the FIRE page's data-driven models resample.
func benchPanel() Panel {
	rows := make([][]float64, 4)
	for a := range rows {
		rows[a] = make([]float64, 480)
		for t := range rows[a] {
			rows[a][t] = 0.003 + 0.03*math.Sin(float64(t*(a+2))/7.0)
		}
	}
	return Panel{Returns: rows, Weights: []float64{0.4, 0.3, 0.2, 0.1}}
}

// BenchmarkStudentT isolates the tail draw every parametric path is built from
// (a normal plus a Marsaglia-Tsang gamma, so the rejection loop shows).
func BenchmarkStudentT(b *testing.B) {
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = studentT(rng, 5)
	}
}

// BenchmarkParametricDraw is one 30-year i.i.d. path, the page's central model.
func BenchmarkParametricDraw(b *testing.B) {
	src := ParametricSource{Mu: 0.05, Sigma: 0.15, Df: 5, Periods: 30}
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkMarkovRegimeDraw is one path of the sequence-stress model.
func BenchmarkMarkovRegimeDraw(b *testing.B) {
	src := NewMarkovRegime(0.05, 0.15, 5, 30)
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkBlockBootstrap draws one fixed-block path off the panel.
func BenchmarkBlockBootstrap(b *testing.B) {
	src := BlockBootstrap{Panel: benchPanel(), BlockLen: 24, Periods: 360}
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkBlockBootstrapPrepared is the same draw off a Prepare'd source, so
// the panel is combined once instead of once per path.
func BenchmarkBlockBootstrapPrepared(b *testing.B) {
	src := Prepare(BlockBootstrap{Panel: benchPanel(), BlockLen: 24, Periods: 360})
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkStationaryBootstrap is the page's block-bootstrap column, monthly
// paths compounded to years, as the strip builds it.
func BenchmarkStationaryBootstrap(b *testing.B) {
	src := Compounded{Inner: StationaryBootstrap{Panel: benchPanel(), MeanBlock: 24, Periods: 360}, Group: 12}
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkStationaryBootstrapPrepared is that same column Prepare'd.
func BenchmarkStationaryBootstrapPrepared(b *testing.B) {
	src := Prepare(Compounded{Inner: StationaryBootstrap{Panel: benchPanel(), MeanBlock: 24, Periods: 360}, Group: 12})
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkHistoricalCohorts is one replay of an actual historical window.
func BenchmarkHistoricalCohorts(b *testing.B) {
	src := HistoricalCohorts{Panel: benchPanel(), Periods: 360}
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkHistoricalCohortsPrepared is that replay off a Prepare'd source.
func BenchmarkHistoricalCohortsPrepared(b *testing.B) {
	src := Prepare(HistoricalCohorts{Panel: benchPanel(), Periods: 360})
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}

// BenchmarkPooledBootstrap draws from the broad-sample pool of per-country
// histories (no panel combining, so nothing to prepare).
func BenchmarkPooledBootstrap(b *testing.B) {
	p := benchPanel()
	src := PooledBootstrap{Series: p.Returns, MeanBlock: 24, Periods: 30}
	rng := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < b.N; i++ {
		_ = src.Draw(rng)
	}
}
