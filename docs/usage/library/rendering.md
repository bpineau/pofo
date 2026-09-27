# Rendering

The numbers come first: `pkg/analyze` renders nothing. Three packages turn
them into pictures.

| Package | What it draws |
|---|---|
| `pkg/chart` | standalone SVG strings (`Line`, `Bars`, `Pie`, `Heatmap`, `Fan`, `Scatter`, `StackedArea`, `Sparkline`...) and terminal charts (`Term`), from plain dates and values |
| `pkg/compare` | the command's whole comparison: several portfolios, a common window, nominal and real statistics, the report model |
| `pkg/report` | that model as a self-contained HTML page (`Render`) or as text (`RenderText`) |

## A chart and the full report

`compare.Compute` runs the command's comparison, `Studies` hands over the
numbers behind every chart, `chart.Line` draws a standalone SVG and
`report.Render` writes the HTML report. The two bundled indices keep it
offline.

```go
// from compare.Example_render
client := marketdata.NewClient("") // "" = no disk cache
spec, _ := portfolio.NewSpec("world and US",
	portfolio.Line{ID: "MSCIWORLD", Weight: 0.6},
	portfolio.Line{ID: "SP500", Weight: 0.4})
cmp, err := compare.Compute(context.Background(), client, []*portfolio.Spec{spec}, compare.Options{
	Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework(),
})
if err != nil {
	panic(err)
}

st := cmp.Studies()[0]
svg := chart.Line(chart.Options{Title: st.Spec.Name, Width: 800, Height: 400}, []chart.Series{
	{Name: st.Spec.Name, Dates: st.Sim.Dates, Values: st.Sim.Index},
})

var page strings.Builder
if err := report.Render(&page, cmp.HTMLPage(compare.Decoration{})); err != nil {
	panic(err)
}
fmt.Println(len(st.Holdings), "holdings,", len(st.Correlation), "x", len(st.Correlation[0]), "correlation")
fmt.Println(strings.HasPrefix(svg, "<svg"), strings.Contains(page.String(), "</html>"))
```

```text
2 holdings, 2 x 2 correlation
true true
```

Charts take any dated values, so an intraday path from `Client.Intraday` or a
valuation series built elsewhere draws the same way. `chart.SetDark(true)`,
called once before rendering, switches every later chart to the dark theme.

The command's terminal output (`pofo -cli`) is `chart.Term` plus
`Comparison.StatRows`; `go doc github.com/bpineau/pofo/pkg/compare` shows
both.
