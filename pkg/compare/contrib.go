package compare

import (
	"fmt"
	"html/template"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/suggest"
)

// The realized-contribution charts: from the simulation's per-asset monthly
// contributions (SimResult.MonthlyContributions), a diverging stacked
// timeline in two windows (trailing-12m and monthly) annotated with the
// macro-regime strip, and the per-regime aggregation as a diverging bar
// matrix (the empirical mirror of the a-priori coverage bars). Both use one
// stable color per holding, the same assignment as the coverage bar
// segments; the contributions come from pkg/portfolio and the macro reading
// from regime.go, this file only shapes them into charts. The regimes use
// suggest's vocabulary, so the strip and matrix share labels and order
// (AllRegimes) with the coverage bars.

// stripColor is the annotation tint of each regime quadrant: growth recedes
// as a neutral wash, the three non-growth states carry a hue (kept away from
// the series palette by darkness or saturation).
var stripColor = map[suggest.Category]string{
	suggest.Growth:    "#D5DBE5",
	suggest.Inflation: "#C77E17",
	suggest.Deflation: "#2A5FA8",
	suggest.Crisis:    "#D2402F",
}

// unclassified marks a month the macro panel does not reach. It is the zero
// Category, so a forgotten check reads as "no regime" rather than as one.
const unclassified = suggest.Category("")

// monthQuadrants classifies each month into its macro quadrant from the
// embedded OECD panel, forward-filling the last known state past the panel's
// END (a few months of nowcast at most, since the panel is refreshed with the
// rest of the data). Months BEFORE its first regime are left unclassified:
// the panel's first reading is 1956-04, while the bundled backcasts reach
// 1871 for the S&P 500 and 1953 for long Treasuries, so head-filling would
// paint decades of a public chart with a macro state nothing measured, and hand
// the per-regime matrix hundreds of months of "growth" that are only the
// absence of data. Returns nil when the panel cannot be read.
func monthQuadrants(months []time.Time) []suggest.Category {
	if len(months) == 0 {
		return nil
	}
	panel, err := loadMacroPanel()
	if err != nil {
		return nil
	}
	out := make([]suggest.Category, len(months))
	last, seen := unclassified, false
	for i, m := range months {
		if q, ok := panel.regimeAt(time.Date(m.Year(), m.Month(), 1, 0, 0, 0, 0, time.UTC)); ok {
			last, seen = q, true
		}
		out[i] = last
	}
	if !seen {
		return nil
	}
	return out
}

// contributionCharts renders the realized-contribution blocks for one
// portfolio: the timeline in its two windows (trailing-12m and monthly,
// toggled in the report) and the per-regime matrix. Empty when the simulated
// window is too short (under two years).
func contributionCharts(r *column) (timeline, monthly, matrix template.HTML) {
	if r.sim == nil || len(r.sim.Contributions) == 0 {
		return "", "", ""
	}
	months, mc := r.sim.MonthlyContributions()
	if len(months) < 24 {
		return "", "", ""
	}
	labels := make([]string, len(r.p.Assets))
	colors := chart.PaletteFor(len(r.p.Assets))
	for i, a := range r.p.Assets {
		base, _ := marketdata.SplitSim(a.ID)
		labels[i] = base
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
func contribTimeline(months []time.Time, mc [][]float64, quads []suggest.Category, labels, colors []string, window int) string {
	first := window - 1 // the first month with a full trailing window behind it
	n := len(months) - first

	series := make([]chart.DivergingStackSeries, len(labels))
	for si, i := range chart.SafeStackOrder(len(labels)) {
		series[si] = chart.DivergingStackSeries{Name: labels[i], Color: colors[i], Values: make([]float64, n)}
		for m := range n {
			sum := 0.0
			for k := m; k <= m+first; k++ {
				sum += mc[i][k]
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
		xtips[m] = key.Format("2006-01")
		if key.Month() == time.January && key.Year()%yearStep == 0 {
			xlabels[m] = fmt.Sprintf("%d", key.Year())
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
			// A stretch the macro panel does not reach leaves a HOLE in the
			// strip. An empty band says "not measured"; any colour there would
			// claim a regime, and the deep backcasts spend decades in it.
			if quads[a] != unclassified {
				opt.Strip = append(opt.Strip, chart.StripBand{
					From: a - first, To: e - 1 - first,
					Label: fmt.Sprintf("%s (%s → %s)", quads[a], months[a].Format("2006-01"), months[e-1].Format("2006-01")),
					Color: stripColor[quads[a]],
				})
			}
			a = e
		}
	}
	return chart.DivergingStack(opt, series)
}

// contribMatrix builds the per-regime annualized contribution matrix.
func contribMatrix(months []time.Time, mc [][]float64, quads []suggest.Category, labels, colors []string) string {
	if quads == nil {
		return ""
	}
	nA := len(labels)
	sums := map[suggest.Category][]float64{}
	cnt := map[suggest.Category]int{}
	for m, q := range quads {
		if q == unclassified {
			continue // before the macro panel: counted nowhere, invented nowhere
		}
		if sums[q] == nil {
			sums[q] = make([]float64, nA)
		}
		cnt[q]++
		for i := range nA {
			sums[q][i] += mc[i][m]
		}
	}
	classified := 0
	for _, n := range cnt {
		classified += n
	}
	if classified == 0 {
		return "" // nothing the macro panel reaches: no matrix to draw
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
	title := "Realized contribution per macro regime (pts/yr over that regime's months)"
	if outside := len(months) - classified; outside > 0 {
		// The shares below then do not fill the window, and the reader is owed
		// the reason rather than left to add them up.
		title += fmt.Sprintf(" - %d months precede the macro panel and are in no column", outside)
	}
	return chart.BarMatrix(chart.BarMatrixOptions{
		Title:        title,
		RowLabels:    labels,
		RowColors:    colors,
		Unit:         "pts/yr",
		Summary:      summary,
		SummaryLabel: "portfolio",
	}, cols)
}
