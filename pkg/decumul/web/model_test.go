package web

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestComputeWithPanelBootstrap(t *testing.T) {
	panel := scenario.Panel{
		Returns: [][]float64{
			{0.08, -0.10, 0.15, 0.05, 0.20, -0.05, 0.12, 0.03},
			{0.02, 0.01, 0.03, 0.00, 0.02, 0.01, 0.02, 0.01},
		},
		Weights: []float64{0.6, 0.4},
	}
	pr := Params{Capital: 1_500_000, NeedAnnual: 48000, BufferYears: 3, Years: 30,
		TaxRate: 0.30, NPaths: 2000, Model: "bootstrap", Weights: []float64{0.6, 0.4}}
	res := ComputeWithPanel(pr, &panel)
	if res.Cards["ruin"] == "" {
		t.Errorf("empty result for bootstrap model")
	}
}
