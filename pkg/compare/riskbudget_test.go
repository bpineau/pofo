package compare

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// rbColumn builds a column whose simulation carries the given per-asset daily
// contributions, one point per day from 2020-01-01.
func rbColumn(t *testing.T, assets []portfolio.Asset, contrib [][]float64) *column {
	t.Helper()
	n := len(contrib[0])
	dates := make([]time.Time, n)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range dates {
		dates[i] = start.AddDate(0, 0, i)
	}
	return &column{
		p:   &portfolio.Portfolio{Name: "T", Assets: assets},
		sim: &portfolio.SimResult{Dates: dates, Contributions: contrib},
	}
}

// The block's whole point: a class can carry a share of risk far above its
// share of capital. Here a 30 % equity sleeve moves ten times as much as the
// 70 % bond sleeve, so it must dominate the risk while staying the minority
// holding, and the classes must be ordered by risk rather than by weight.
func TestRiskBudgetSeparatesCapitalFromRisk(t *testing.T) {
	const days = 400
	equity := make([]float64, days)
	bond := make([]float64, days)
	for k := 1; k < days; k++ {
		equity[k] = 0.01 * math.Sin(float64(k)/3)
		bond[k] = 0.001 * math.Cos(float64(k)/7)
	}
	assets := []portfolio.Asset{
		{ID: "EQ", Symbol: "EQ", Weight: 0.30},
		{ID: "BD", Symbol: "BD", Weight: 0.70},
	}
	meta := map[string]suggest.Meta{
		"EQ": {AssetClass: "equity"},
		"BD": {AssetClass: "government-bond"},
	}
	rows := riskBudgetRows(rbColumn(t, assets, [][]float64{equity, bond}), meta)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].Label != "Equity" {
		t.Errorf("first row = %q, want the largest risk share (Equity)", rows[0].Label)
	}
	if rows[0].Capital != "30.0 %" || rows[1].Capital != "70.0 %" {
		t.Errorf("capital shares = %q / %q, want 30.0 %% / 70.0 %%", rows[0].Capital, rows[1].Capital)
	}
	// The bar is scaled to the largest share shown, so the dominant risk fills
	// the track and its capital tick sits well short of it.
	if rows[0].RiskWidth != 100 {
		t.Errorf("dominant risk width = %v, want 100", rows[0].RiskWidth)
	}
	if rows[0].CapitalMark >= rows[0].RiskWidth {
		t.Errorf("capital tick at %v is not left of the risk bar at %v", rows[0].CapitalMark, rows[0].RiskWidth)
	}
}

// A stacked fund is split across its classes pro rata of notional, exactly as
// the composition pies split it, so the two blocks describe the same fund the
// same way.
func TestRiskBudgetSplitsStackedFund(t *testing.T) {
	const days = 300
	one := make([]float64, days)
	for k := 1; k < days; k++ {
		one[k] = 0.01 * math.Sin(float64(k)/5)
	}
	assets := []portfolio.Asset{{ID: "NTSX", Symbol: "NTSX", Weight: 1.0}}
	meta := map[string]suggest.Meta{
		"NTSX": {AssetClass: "multi-asset", Exposures: map[string]float64{"equity": 0.9, "government-bond": 0.6}},
	}
	rows := riskBudgetRows(rbColumn(t, assets, [][]float64{one}), meta)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want the fund split in two", len(rows))
	}
	byLabel := map[string]string{}
	for _, r := range rows {
		byLabel[r.Label] = r.Capital
	}
	// 0.9 / (0.9+0.6) = 60 % equity, 40 % bonds.
	if byLabel["Equity"] != "60.0 %" {
		t.Errorf("equity capital share = %q, want 60.0 %%", byLabel["Equity"])
	}
	if byLabel["Government bond"] != "40.0 %" {
		t.Errorf("bond capital share = %q, want 40.0 %%", byLabel["Government bond"])
	}
}

// Without catalog metadata every class would read "Unknown", which teaches
// nothing; the block omits itself rather than render a single grey bar.
func TestRiskBudgetOmittedWithoutMetadata(t *testing.T) {
	const days = 200
	one := make([]float64, days)
	for k := 1; k < days; k++ {
		one[k] = 0.01 * math.Sin(float64(k)/5)
	}
	assets := []portfolio.Asset{{ID: "X", Symbol: "X", Weight: 1.0}}
	if rows := riskBudgetRows(rbColumn(t, assets, [][]float64{one}), nil); rows != nil {
		t.Errorf("got %d rows without metadata, want none", len(rows))
	}
}
