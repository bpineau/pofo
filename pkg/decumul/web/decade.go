package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/scenario"
)

// DecadeResult makes sequence-of-returns risk visible: the central model's
// paths grouped by the market return of their first decade, with the ruin
// probability of each group. Same average return, different ordering,
// completely different fates: most of a plan's risk is decided in its first
// ten years.
type DecadeResult struct {
	SVG   string `json:"decadeSvg"`
	Cards []Card `json:"cards"`
	Note  string `json:"note"`
}

// Decade simulates the central model at the planned spend and decomposes ruin
// by first-decade return quintile.
func Decade(pr Params, panel *scenario.Panel) DecadeResult {
	if pr.NPaths == 0 {
		pr.NPaths = 2000
	}
	base := pr.plan()
	base.Monthly = false
	base.Source = pr.detailSource(panel, pr.Years)
	e := base.Simulate(pr.NPaths, simWorkers, 7)

	const quintiles = 5
	buckets := e.DecadeBuckets(quintiles)
	if buckets == nil {
		return DecadeResult{Note: "Not enough paths to bucket."}
	}
	bars := make([]chart.Bar, len(buckets))
	for i, b := range buckets {
		bars[i] = chart.Bar{
			Label: fmt.Sprintf("%+.1f..%+.1f%%", b.LoRet*100, b.HiRet*100),
			Value: b.RuinProb * 100,
			Text:  fmt.Sprintf("%.0f%%", b.RuinProb*100),
		}
	}

	// How concentrated is failure in the worst first decades? Share of all
	// ruined paths whose first decade sat in the bottom quintile.
	totalRuined, worstRuined := 0.0, buckets[0].RuinProb*float64(buckets[0].Paths)
	for _, b := range buckets {
		totalRuined += b.RuinProb * float64(b.Paths)
	}
	concentration := 0.0
	if totalRuined > 0 {
		concentration = worstRuined / totalRuined
	}
	cards := []Card{
		{Label: "Ruin, worst first decade", Value: fmt.Sprintf("%.0f%%", buckets[0].RuinProb*100),
			Help: "Ruin probability among the 20% of futures whose first decade returned the least. This is the sequence-risk premium: the same plan, dealt a bad opening decade."},
		{Label: "Ruin, best first decade", Value: fmt.Sprintf("%.0f%%", buckets[len(buckets)-1].RuinProb*100),
			Help: "Ruin probability when the first decade goes well: with a good opening, almost nothing sinks the plan afterwards."},
		{Label: "Failures born in the worst decade", Value: fmt.Sprintf("%.0f%%", concentration*100),
			Help: "Share of all failures whose first decade sat in the bottom quintile. High concentration means protections aimed at the early years (buffer, flexibility, side income) target most of the risk."},
		{Label: "Median wealth after a worst start", Value: fmtWealth(buckets[0].TerminalP50),
			Help: "Median terminal real wealth of the worst-first-decade futures: what surviving a bad opening typically leaves at the end."},
	}
	return DecadeResult{
		SVG: darkBars(chart.Options{
			Title: "Ruin by first-decade real return (quintiles, worst → best)", Width: 720, Height: 360}, bars),
		Cards: cards,
	}
}
