package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/scenario"
)

// SpendingResult shows the plan from the household's point of view: the
// delivered real spending per year (fan of percentiles across paths) and the
// lived cost of the adaptive policy, i.e. how often, how soon and for how long
// the standard of living actually dropped below the uncut plan.
type SpendingResult struct {
	SVG   string `json:"spendingSvg"`
	Cards []Card `json:"cards"`
	Note  string `json:"note"`
}

// spendPercentiles are the bands drawn on the spending fan.
var spendPercentiles = []float64{0.05, 0.25, 0.50, 0.75, 0.95}

// Spending simulates the central model at the user's planned spend, under the
// full adaptive policy (flex/guardrails/ratchet), and reports the spending the
// household actually lives on.
func Spending(pr Params, panel *scenario.Panel) SpendingResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	base := pr.plan()
	base.Monthly = false
	cMu, cSigma, cDf := centralParams(pr, panel)
	base.Source = scenario.ParametricSource{Mu: cMu, Sigma: cSigma, Df: cDf, Periods: pr.Years}

	e := base.Simulate(pr.NPaths, simWorkers, 7)
	bands := e.SpendBands(spendPercentiles)
	s := e.SpendStats()

	// The kernel's Spend series is what the portfolio delivered, net of
	// cashflows; the household's standard of living adds the pension and side
	// income back. Cashflows are deterministic per year, so adding them to
	// each quantile is exact.
	for _, band := range bands {
		for k := range band {
			band[k] += pr.cashflowAt(k)
		}
	}

	cards := []Card{
		{"Paths ever cut", fmt.Sprintf("%.0f%%", s.EverCutShare*100)},
		{"First cut (median year)", firstCutText(s.EverCutShare, s.FirstCutMedian)},
		{"Years lived cut (median)", cutYearsText(s.EverCutShare, s.CutYearsMedian)},
		{"Years lived cut (p90)", cutYearsText(s.EverCutShare, s.CutYearsP90)},
		{"Spending floor (p5, year 10)", floorText(bands, 10)},
	}
	svg := chart.Fan(
		chart.Options{Title: "Household real spending €/yr, incl. pension & side income (central model)"},
		"Year", bands, nil)
	return SpendingResult{SVG: svg, Cards: cards}
}

// cashflowAt is the deterministic income (pension, side income) active in a
// year, mirroring the plan's cashflow construction.
func (pr Params) cashflowAt(year int) float64 {
	total := 0.0
	if pr.PensionAnnual > 0 && year >= pr.PensionYear {
		total += pr.PensionAnnual
	}
	if pr.SideAnnual > 0 && year < pr.SideUntilYear {
		total += pr.SideAnnual
	}
	return total
}

// firstCutText formats the median first-cut year, or a dash when no path cut.
func firstCutText(share, year float64) string {
	if share == 0 {
		return "—"
	}
	return fmt.Sprintf("year %.0f", year)
}

// cutYearsText formats a cut-years figure, or a dash when no path cut.
func cutYearsText(share, years float64) string {
	if share == 0 {
		return "—"
	}
	return fmt.Sprintf("%.0f y", years)
}

// floorText is the p5 spending at the given year: the bad-case standard of
// living the policy would actually deliver.
func floorText(bands [][]float64, year int) string {
	if len(bands) == 0 || year >= len(bands[0]) {
		return "—"
	}
	return fmt.Sprintf("%.1f k€", bands[0][year]/1000)
}
