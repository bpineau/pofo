package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/scenario"
)

// IncomeResult shows where each year's spending money comes from on the
// median path: the layers of guaranteed income (pension, side income) and the
// portfolio withdrawals on top. An annuity is not a layer here: it belongs to
// the mortality kernel (annuity.go), and the lifecycle view reports the income
// it buys. It makes the plan's shape
// tangible: the gap years before the pension are visibly the ones carried by
// the portfolio alone, which is why pension timing dominates the sensitivity
// ranking and why sequence risk concentrates there.
type IncomeResult struct {
	SVG   string `json:"incomeSvg"`
	Cards []Card `json:"cards"`
	Note  string `json:"note"`
}

// Income simulates the central model and stacks the median funding mix.
func Income(pr Params, panel *scenario.Panel) IncomeResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	base := pr.plan()
	base.Monthly = false
	base.Source = pr.detailSource(panel, pr.Years)
	e := base.Simulate(pr.NPaths, simWorkers, 7)

	// Median net spending delivered by the portfolio, per year.
	bands := e.SpendBands([]float64{0.50})
	if len(bands) == 0 {
		return IncomeResult{Note: "No paths simulated."}
	}
	portfolio := bands[0]

	years := pr.Years
	pension := make([]float64, years)
	side := make([]float64, years)
	for k := range years {
		if pr.PensionAnnual > 0 && k >= pr.PensionYear {
			pension[k] = pr.PensionAnnual / 1000
		}
		if pr.SideAnnual > 0 && k < pr.SideUntilYear {
			side[k] = pr.SideAnnual / 1000
		}
	}
	pf := make([]float64, years)
	for k := range years {
		if k < len(portfolio) {
			pf[k] = portfolio[k] / 1000
		}
	}

	// Only the active layers are drawn; the guaranteed floors sit at the
	// bottom so the portfolio's share reads as the exposed remainder.
	var series []chart.AreaSeries
	if pr.PensionAnnual > 0 {
		series = append(series, chart.AreaSeries{Name: "Pension", Values: pension, Color: chart.PaletteColor(2)})
	}
	if pr.SideAnnual > 0 {
		series = append(series, chart.AreaSeries{Name: "Side income", Values: side, Color: chart.PaletteColor(3)})
	}
	series = append(series, chart.AreaSeries{Name: "Portfolio withdrawals", Values: pf, Color: chart.PaletteColor(0)})

	// The gap years: how long the portfolio carries the household alone, and
	// how much of late-retirement spending is guaranteed.
	gapYears := min(pr.PensionYear, years)
	if pr.PensionAnnual <= 0 {
		gapYears = years
	}
	lastGuaranteed := 0.0
	if pr.PensionAnnual > 0 {
		lastGuaranteed += pr.PensionAnnual
	}
	lastShare := 0.0
	if total := lastGuaranteed + pf[years-1]*1000; total > 0 {
		lastShare = lastGuaranteed / total
	}
	cards := []Card{
		{Label: "Years carried by the portfolio alone", Value: fmt.Sprintf("%d y", gapYears),
			Help: "Before any pension starts, every euro of spending is a portfolio sale: these are the years sequence risk can do real damage."},
		{Label: "Guaranteed share, final years", Value: fmt.Sprintf("%.0f%%", lastShare*100),
			Help: "The pension as a share of the median household spending in the last plan year: the part of late-life spending no market crash can take away."},
	}
	return IncomeResult{
		SVG: darkStackedArea(chart.Options{
			Title: "Where the money comes from, median path (k€/yr)", Width: 1180, Height: 360},
			"Years into retirement", "k€/yr", series),
		Cards: cards,
	}
}
