package main

import (
	"fmt"
	"html/template"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/permanent"
	"github.com/bpineau/pofo/pkg/suggest"
)

// The realized-contribution charts: from the simulation's per-asset daily
// contributions (SimResult.Contributions), a trailing-12-month diverging
// stacked timeline annotated with the macro-regime strip, and its per-regime
// aggregation as a diverging bar matrix (the empirical mirror of the
// a-priori coverage bars). Both use one stable color per holding, the same
// assignment as the coverage bar segments.

// monthKey identifies a calendar month.
type monthKey struct{ y, m int }

func (k monthKey) String() string { return fmt.Sprintf("%04d-%02d", k.y, k.m) }

// quadrantOf maps a macro state to its growth x inflation quadrant, breadth
// thresholded at one half: accelerating inflation with growth is the
// inflation regime, without growth the crisis (stagflation) one; decelerating
// both is deflation; otherwise growth.
func quadrantOf(r permanent.Regime) suggest.Category {
	switch {
	case r.GrowthBreadth >= 0.5 && r.InflationBreadth >= 0.5:
		return suggest.Inflation
	case r.GrowthBreadth < 0.5 && r.InflationBreadth >= 0.5:
		return suggest.Crisis
	case r.GrowthBreadth < 0.5 && r.InflationBreadth < 0.5:
		return suggest.Deflation
	default:
		return suggest.Growth
	}
}

// stripColor is the annotation tint of each regime quadrant: growth recedes
// as a neutral wash, the three non-growth states carry a hue (kept away from
// the series palette by darkness or saturation).
var stripColor = map[suggest.Category]string{
	suggest.Growth:    "#D5DBE5",
	suggest.Inflation: "#C77E17",
	suggest.Deflation: "#2A5FA8",
	suggest.Crisis:    "#D2402F",
}

// monthlyContributions folds the simulation's daily per-asset contributions
// into calendar months. It returns the month keys (ascending) and, per
// month, the per-asset summed contribution (fractions of portfolio value).
func monthlyContributions(r *result) (months []monthKey, mc [][]float64) {
	sim := r.sim
	for k := 1; k < len(sim.Dates); k++ {
		key := monthKey{sim.Dates[k].Year(), int(sim.Dates[k].Month())}
		if len(months) == 0 || months[len(months)-1] != key {
			months = append(months, key)
			mc = append(mc, make([]float64, len(sim.Contributions)))
		}
		row := mc[len(mc)-1]
		for i := range sim.Contributions {
			row[i] += sim.Contributions[i][k]
		}
	}
	return months, mc
}

// monthQuadrants classifies each month into its macro quadrant from the
// embedded OECD panel, forward-filling the last known state where the panel
// has no data (its edges). Returns nil when the panel cannot be read.
func monthQuadrants(months []monthKey) []suggest.Category {
	panel, err := permanent.LoadPanel()
	if err != nil || len(months) == 0 {
		return nil
	}
	from := monthStart(months[0])
	to := monthStart(months[len(months)-1])
	byMonth := map[monthKey]suggest.Category{}
	for _, r := range panel.Regimes(from, to, permanent.DefaultSignalConfig()) {
		byMonth[monthKey{r.Date.Year(), int(r.Date.Month())}] = quadrantOf(r)
	}
	if len(byMonth) == 0 {
		return nil
	}
	out := make([]suggest.Category, len(months))
	last := suggest.Growth
	for i, k := range months {
		if q, ok := byMonth[k]; ok {
			last = q
		}
		out[i] = last
	}
	return out
}

// contributionCharts renders the realized-contribution blocks for one
// portfolio: the timeline in its two windows (trailing-12m and monthly,
// toggled in the report) and the per-regime matrix. Empty when the simulated
// window is too short (under two years).
func contributionCharts(r *result) (timeline, monthly, matrix template.HTML) {
	if r.sim == nil || len(r.sim.Contributions) == 0 {
		return "", "", ""
	}
	months, mc := monthlyContributions(r)
	if len(months) < 24 {
		return "", "", ""
	}
	labels := make([]string, len(r.p.Assets))
	colors := make([]string, len(r.p.Assets))
	for i, a := range r.p.Assets {
		base, _ := marketdata.SplitSim(a.ID)
		labels[i] = base
		colors[i] = chart.PaletteColor(i)
	}
	quads := monthQuadrants(months)
	return template.HTML(contribTimeline(months, mc, quads, labels, colors, 12)),
		template.HTML(contribTimeline(months, mc, quads, labels, colors, 1)),
		template.HTML(contribMatrix(months, mc, quads, labels, colors))
}

// contribTimeline builds the diverging stacked contribution chart over a
// trailing window of `window` months (12 = the smoothed year view, 1 = the
// raw monthly view that reads a crash month by month). Series keep one
// stable color per holding (the coverage-bar assignment) but stack in the
// palette's CVD-safe order, not holding order.
func contribTimeline(months []monthKey, mc [][]float64, quads []suggest.Category, labels, colors []string, window int) string {
	nA := len(labels)
	first := window - 1 // the first month with a full trailing window behind it
	n := len(months) - first

	series := make([]chart.DivergingStackSeries, nA)
	for si, i := range chart.SafeStackOrder(nA) {
		series[si] = chart.DivergingStackSeries{Name: labels[i], Color: colors[i], Values: make([]float64, n)}
		for m := range n {
			sum := 0.0
			for k := m; k <= m+first; k++ {
				sum += mc[k][i]
			}
			series[si].Values[m] = sum * 100
		}
	}
	total := make([]float64, n)
	xlabels := make([]string, n)
	xtips := make([]string, n)
	yearStep := 1
	if n > 15*12 {
		yearStep = 2
	}
	for m := range n {
		for _, s := range series {
			total[m] += s.Values[m]
		}
		key := months[m+first]
		xtips[m] = key.String()
		if key.m == 1 && key.y%yearStep == 0 {
			xlabels[m] = fmt.Sprintf("%d", key.y)
		}
	}

	title := "Realized contribution, trailing 12 months (pts of return)"
	ylabel, totalName := "pts / 12m", "portfolio 12m"
	if window == 1 {
		title = "Realized contribution, month by month (pts of return)"
		ylabel, totalName = "pts / month", "portfolio month"
	}
	opt := chart.DivergingStackOptions{
		Title:     title,
		XLabels:   xlabels,
		XTips:     xtips,
		XLabel:    "month",
		YLabel:    ylabel,
		Total:     total,
		TotalName: totalName,
	}
	if quads != nil {
		opt.StripName = "regime"
		for _, q := range suggest.AllRegimes {
			opt.StripLegend = append(opt.StripLegend, chart.Slice{Label: string(q), Color: stripColor[q]})
		}
		for a := first; a < len(months); {
			e := a
			for e < len(months) && quads[e] == quads[a] {
				e++
			}
			opt.Strip = append(opt.Strip, chart.StripBand{
				From: a - first, To: e - 1 - first,
				Label: fmt.Sprintf("%s (%s → %s)", quads[a], months[a], months[e-1]),
				Color: stripColor[quads[a]],
			})
			a = e
		}
	}
	return chart.DivergingStack(opt, series)
}

// contribMatrix builds the per-regime annualized contribution matrix.
func contribMatrix(months []monthKey, mc [][]float64, quads []suggest.Category, labels, colors []string) string {
	if quads == nil {
		return ""
	}
	nA := len(labels)
	sums := map[suggest.Category][]float64{}
	cnt := map[suggest.Category]int{}
	for m, q := range quads {
		if sums[q] == nil {
			sums[q] = make([]float64, nA)
		}
		cnt[q]++
		for i := range nA {
			sums[q][i] += mc[m][i]
		}
	}
	var cols []chart.MatrixColumn
	var summary []float64
	for _, q := range suggest.AllRegimes {
		if cnt[q] == 0 {
			continue
		}
		col := chart.MatrixColumn{
			Title:    string(q),
			Subtitle: fmt.Sprintf("%d months · %.0f%% of window", cnt[q], 100*float64(cnt[q])/float64(len(months))),
			Color:    stripColor[q],
			Values:   make([]float64, nA),
		}
		colTotal := 0.0
		for i := range nA {
			v := sums[q][i] / float64(cnt[q]) * 12 * 100
			col.Values[i] = math.Round(v*10) / 10
			colTotal += v
		}
		cols = append(cols, col)
		summary = append(summary, math.Round(colTotal*10)/10)
	}
	return chart.BarMatrix(chart.BarMatrixOptions{
		Title:        "Realized contribution per macro regime (pts/yr over that regime's months)",
		RowLabels:    labels,
		RowColors:    colors,
		Unit:         "pts/yr",
		Summary:      summary,
		SummaryLabel: "portfolio",
	}, cols)
}

// monthStart is the first day of the month, UTC.
func monthStart(k monthKey) time.Time {
	return time.Date(k.y, time.Month(k.m), 1, 0, 0, 0, 0, time.UTC)
}
