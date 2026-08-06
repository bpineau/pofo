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

// monthlyPanel builds a single-asset panel of n monthly real returns.
func monthlyPanel(n int) scenario.Panel {
	row := make([]float64, n)
	for i := range row {
		row[i] = 0.004 * float64((i%7)-3) // a small varying return
	}
	return scenario.Panel{Returns: [][]float64{row}, Weights: []float64{1}}
}

func TestComputeWithPanelCohortsTooShort(t *testing.T) {
	// 240 months (~20 years) of history, 30-year horizon: cohorts cannot
	// extrapolate, so a note replaces any misleading figure.
	panel := monthlyPanel(240)
	pr := Params{Capital: 1_500_000, NeedAnnual: 48000, Years: 30, TaxRate: 0.30,
		NPaths: 1000, Model: "cohorts", Weights: []float64{1}}
	res := ComputeWithPanel(pr, &panel)
	if res.Note == "" {
		t.Errorf("expected a note about insufficient history")
	}
	if res.Cards["ruin"] != "" {
		t.Errorf("should not report a misleading ruin figure, got %q", res.Cards["ruin"])
	}
}

func TestMonthlyCohortsManyWindows(t *testing.T) {
	// 240 months of history, 15-year (180-month) horizon: monthly sampling
	// yields 240-180+1 = 61 cohort windows (vs <=6 with annual sampling), so
	// the model runs without the insufficient-history note.
	panel := monthlyPanel(240)
	inner := scenario.HistoricalCohorts{Panel: panel, Weights: []float64{1}, Periods: 15 * 12}
	if got := inner.Count(); got < 30 {
		t.Errorf("cohort windows = %d, want many (>30)", got)
	}
	pr := Params{Capital: 1_800_000, NeedAnnual: 48000, Years: 15, TaxRate: 0.30,
		NPaths: 1000, Model: "cohorts", Weights: []float64{1}}
	if res := ComputeWithPanel(pr, &panel); res.Note != "" {
		t.Errorf("did not expect a note at a 15y horizon, got %q", res.Note)
	}
}
