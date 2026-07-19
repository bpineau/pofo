package web

import "github.com/bpineau/pofo/pkg/chart"

// The decumulation UI renders every chart in the terminal-dark theme. It
// cannot use chart.SetDark, which flips a process-global flag: the -serve
// mode runs this UI in the same process as the light /view report, so the
// theme has to be a property of each chart, not of the process. These thin
// wrappers forward to the chart primitives and apply chart.Darken to the
// result; the whole package builds its charts through them so the surface
// stays consistently dark whatever the global flag is (chart.Darken is
// idempotent). A new chart added here must go through a wrapper, not
// chart.* directly, or it will render light under -serve.

func darkMultiLine(o chart.Options, xLabel, yLabel string, series []chart.XYSeries, markers ...chart.Marker) string {
	return chart.Darken(chart.MultiLine(o, xLabel, yLabel, series, markers...))
}

func darkBars(o chart.Options, b []chart.Bar) string { return chart.Darken(chart.Bars(o, b)) }

func darkHbars(o chart.Options, b []chart.Bar) string { return chart.Darken(chart.HBars(o, b)) }

func darkCategoryBars(o chart.Options, b []chart.CatBar) string {
	return chart.Darken(chart.CategoryBars(o, b))
}

func darkStackedArea(o chart.Options, xLabel, yLabel string, series []chart.AreaSeries) string {
	return chart.Darken(chart.StackedArea(o, xLabel, yLabel, series))
}

func darkFan(o chart.Options, xLabel string, bands, samples [][]float64, markers ...chart.Marker) string {
	return chart.Darken(chart.Fan(o, xLabel, bands, samples, markers...))
}

func darkScatter(o chart.Options, xlab, ylab string, pts []chart.LabeledPoint) string {
	return chart.Darken(chart.Scatter(o, xlab, ylab, pts))
}
