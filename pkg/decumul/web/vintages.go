package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/scenario"
)

// VintagesResult replays the user's actual plan (policy, buffer, taxes,
// cashflows included) through the most infamous retirement start dates of the
// historical record, deterministically: no resampling, the years as they
// happened. It makes sequence-of-returns risk concrete with real dates, the
// way the withdrawal literature was built (Bengen's and the Trinity study's
// binding cohorts are the mid-1960s starts).
type VintagesResult struct {
	SVG   string `json:"vintagesSvg"`
	Cards []Card `json:"cards"`
	Note  string `json:"note"`
}

// vintage is one named historical start: a country of the JST panel and the
// retirement's first market year.
type vintage struct {
	iso   string
	year  int
	label string
	story string // one-line hover: why this start date is famous
}

// vintageList holds the canonical stress vintages, worst-reputation first.
// All are local real equity returns for a domestic investor (JST): a proxy
// for the growth sleeve, not a EUR-hedged track.
var vintageList = []vintage{
	{"USA", 1929, "USA 1929", "Retiring on the eve of the Great Depression: real equities lose ~75% in three years, the deepest sequence shock on record."},
	{"USA", 1966, "USA 1966", "The cohort the 4% rule was calibrated on: a decade and a half of inflation-eroded real returns makes 1966 the binding worst case of the US record."},
	{"JPN", 1990, "Japan 1990", "The definitive lost decades: Japanese equities were still underwater in real terms 30 years after the bubble burst."},
	{"USA", 2000, "USA 2000", "The dot-com bust straight into the 2008 crisis: a lost real decade to open the retirement."},
}

// Vintages runs the plan through each historical vintage. The kernel is the
// annual one (JST is annual); each vintage is a single deterministic path.
func Vintages(pr Params, _ *scenario.Panel) VintagesResult {
	byISO := map[string]countrySeries{}
	for _, c := range broadSampleCountries() {
		byISO[c.ISO] = c
	}
	var series []chart.XYSeries
	var cards []Card
	truncated := false
	for i, v := range vintageList {
		c, ok := byISO[v.iso]
		if !ok {
			continue
		}
		start := v.year - c.FirstYear
		if start < 0 || start >= len(c.Returns) {
			continue
		}
		seq := scenario.Sequence(c.Returns[start:])
		years := pr.Years
		if len(seq) < years {
			years, truncated = len(seq), true
		}
		p := pr.plan()
		p.Monthly = false
		p.Years = years
		res := p.RunPath(seq)

		xs := make([]float64, len(res.Wealth))
		ys := make([]float64, len(res.Wealth))
		for k := range res.Wealth {
			xs[k], ys[k] = float64(k), res.Wealth[k]/1e6
		}
		series = append(series, chart.XYSeries{Name: v.label, Xs: xs, Ys: ys, Color: chart.PaletteColor(i)})

		cards = append(cards, Card{v.label, vintageVerdict(res.Ruined, res.RuinYear, years, pr.Years,
			res.Wealth[len(res.Wealth)-1]), v.story})
	}
	note := "Local real equity returns of each market (JST), a proxy for the growth sleeve; your buffer, policy, taxes and cashflows all apply."
	if truncated {
		note += " Paths stop where the record ends (2020)."
	}
	return VintagesResult{
		SVG: chart.MultiLine(
			chart.Options{Title: "Your plan through the worst retirements on record (real wealth, M€)", Width: 1180, Height: 400},
			"Years into retirement", "Real wealth M€", series),
		Cards: cards,
		Note:  note,
	}
}

// vintageVerdict is the one-line outcome of a vintage replay.
func vintageVerdict(ruined bool, ruinYear, ran, wanted int, terminal float64) string {
	if ruined {
		return fmt.Sprintf("ruined in year %d", ruinYear+1)
	}
	if ran < wanted {
		return fmt.Sprintf("solvent when the record ends (year %d): %s left", ran, fmtWealth(terminal))
	}
	return fmt.Sprintf("survived all %d years: %s left", ran, fmtWealth(terminal))
}
