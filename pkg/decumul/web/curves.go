package web

import (
	"fmt"
	"math"
	"strings"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// CurvesResult carries the two planning curves: the safe withdrawal rate as
// the horizon stretches (how much a longer retirement costs), and the capital
// required at the target ruin as the planned spending moves (what one more
// k€/yr of lifestyle costs in capital).
type CurvesResult struct {
	HorizonSVG string `json:"horizonSvg"` // safe withdrawal rate against the horizon
	CapitalSVG string `json:"capitalSvg"` // required capital against the planned spend
	Note       string `json:"note"`       // caveat; "" when none
}

// curveHorizons are the retirement lengths sampled on the safe-WR curve.
var curveHorizons = []int{25, 30, 35, 40, 45, 50, 55, 60}

// Curves computes both planning curves on the fixed rule (the conventional
// definition of a safe withdrawal), for the central and the broad-sample
// models so the epistemic spread stays visible.
func Curves(pr Params, panel *scenario.Panel) CurvesResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	const seed = uint64(7)
	paths := min(pr.NPaths, shapePaths)
	cMu, cSigma, cDf := centralParams(pr, panel)

	// Safe WR vs horizon, central Student-t and broad-sample regime curves.
	// slot pins each model to its hero-strip palette slot (Student-t 0,
	// Broad-sample 2) so a model keeps its color across every chart.
	type curveModel struct {
		name   string
		slot   int
		source func(years int) scenario.Source
	}
	models := []curveModel{
		{"Central (Student-t)", 0, func(years int) scenario.Source {
			return centralSource(pr, cMu, cSigma, cDf, years)
		}},
		{"Broad-sample", 2, func(years int) scenario.Source {
			return broadSampleSource(years)
		}},
	}
	// Every point of both curves is its own solve, independent of the others,
	// so they all run at once (concurrently).
	nh := len(curveHorizons)
	notes := make([]string, len(models)*nh+9)
	series := make([]chart.XYSeries, len(models))
	for i, m := range models {
		series[i] = chart.XYSeries{Name: m.name, Xs: make([]float64, nh), Ys: make([]float64, nh), Color: chart.PaletteColor(m.slot)}
	}
	concurrently(len(models)*nh, func(k int) {
		i, j := k/nh, k%nh
		years := curveHorizons[j]
		p := fixedRule(pr.plan())
		p.Monthly = false
		p.Years = years
		p.Source = models[i].source(years)
		safe, note := solveOn(p, target, decumul.WithdrawalAxis(0, pr.Capital*0.15), p.Draw(paths, simWorkers, seed))
		if note != "" {
			notes[k] = fmt.Sprintf("%s, %dy: %s", models[i].name, years, note)
		}
		series[i].Xs[j], series[i].Ys[j] = float64(years), safe.Solution/pr.Capital*100
	})
	for i := range series {
		series[i].Xs, series[i].Ys = definedPairs(series[i].Xs, series[i].Ys)
	}
	horizonSVG := darkMultiLine(
		chart.Options{Title: "Safe withdrawal rate vs horizon (at your target ruin)", Width: 720, Height: 360},
		"Horizon (years)", "Safe WR %", series,
		chart.Marker{Axis: 'x', Value: float64(pr.Years), Label: "your horizon"},
	)

	// Required capital vs annual spending, central model, at the target ruin.
	spends := []float64{36000, 42000, 48000, 54000, 60000, 66000, 72000, 78000, 84000}
	xs := make([]float64, len(spends))
	ys := make([]float64, len(spends))
	concurrently(len(spends), func(j int) {
		p := fixedRule(pr.plan())
		p.Monthly = false
		p.NeedAnnual = spends[j]
		p.Source = centralSource(pr, cMu, cSigma, cDf, pr.Years)
		// Ruin falls as capital rises: the smallest capital meeting the target.
		cap, note := solveOn(p, target, decumul.CapitalAxis(solveLo, solveHi), p.Draw(paths, simWorkers, seed))
		if note != "" {
			notes[len(models)*nh+j] = fmt.Sprintf("Spending %.0f: %s", spends[j], note)
		}
		xs[j], ys[j] = spends[j]/1000, cap.Solution/1e6
	})
	xs, ys = definedPairs(xs, ys)
	capitalSVG := darkMultiLine(
		chart.Options{Title: "Capital required vs spending (central model, target ruin)", Width: 720, Height: 360},
		"Net spending k€/yr", "Required capital M€",
		[]chart.XYSeries{{Name: "", Xs: xs, Ys: ys, Color: chart.PaletteColor(0)}},
		chart.Marker{Axis: 'x', Value: pr.NeedAnnual / 1000, Label: "your plan"},
	)
	var limitations []string
	for _, note := range notes {
		if note != "" {
			limitations = append(limitations, note)
		}
	}
	return CurvesResult{HorizonSVG: horizonSVG, CapitalSVG: capitalSVG, Note: strings.Join(limitations, " ")}
}

func definedPairs(xs, ys []float64) ([]float64, []float64) {
	x, y := make([]float64, 0, len(xs)), make([]float64, 0, len(ys))
	for i, v := range ys {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			x = append(x, xs[i])
			y = append(y, v)
		}
	}
	return x, y
}
