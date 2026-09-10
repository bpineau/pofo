package web

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// benchPanel is a portfolio-shaped panel: four assets of monthly real returns
// over forty years, what the data-driven columns resample.
func benchPanel() *scenario.Panel {
	rows := make([][]float64, 4)
	for a := range rows {
		rows[a] = make([]float64, 480)
		for t := range rows[a] {
			rows[a][t] = 0.003 + 0.03*math.Sin(float64(t*(a+2))/7.0)
		}
	}
	return &scenario.Panel{Returns: rows, Weights: []float64{0.4, 0.3, 0.2, 0.1}}
}

// benchWebParams is a full-horizon, endpoint-sized parameter set.
func benchWebParams(model string) Params {
	return Params{
		Capital: 900_000, NeedAnnual: 34_000, BufferYears: 2,
		Mu: 0.045, Sigma: 0.13, Df: 5, Years: 30, NPaths: 2000, TaxRate: 0.30,
		PensionAnnual: 12_000, PensionYear: 12, Model: model,
		Weights: []float64{0.4, 0.3, 0.2, 0.1},
	}
}

// BenchmarkComputeParametric is the /api/sim answer end to end (a nine-point
// buffer sweep plus the headline ensemble) on the parametric central model.
func BenchmarkComputeParametric(b *testing.B) {
	pr := benchWebParams("parametric")
	for i := 0; i < b.N; i++ {
		_ = ComputeWithPanel(pr, nil)
	}
}

// BenchmarkComputeBootstrap is that same answer on the block-bootstrap
// column, where the source also has a panel to combine.
func BenchmarkComputeBootstrap(b *testing.B) {
	panel := benchPanel()
	pr := benchWebParams("bootstrap")
	pr.Central = "boot"
	for i := 0; i < b.N; i++ {
		_ = ComputeWithPanel(pr, panel)
	}
}

// BenchmarkModelsStrip is /api/models: six return models, each simulated and
// solved, the heaviest single endpoint of a page render.
func BenchmarkModelsStrip(b *testing.B) {
	panel := benchPanel()
	pr := benchWebParams("parametric")
	for i := 0; i < b.N; i++ {
		_ = Models(pr, panel)
	}
}
