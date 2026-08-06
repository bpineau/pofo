package report

import (
	"html/template"
	"io"
)

// AssetRow is one line of a portfolio composition table.
type AssetRow struct {
	Weight   string
	ID       string
	Symbol   string
	Name     string
	Class    string // catalog asset class (equity, gold…), empty when unknown
	UCITS    string // "oui", "non" or "?" when undetermined
	Fees     string // published TER, or — when unknown
	Currency string
	History  string
	Note     string
}

// CoverageBar is one category row (a macro regime or a risk factor) of a
// portfolio's coverage chart.
type CoverageBar struct {
	Regime string // the category label
	Pct    int    // coverage as a percent of portfolio weight (can exceed 100)
	Width  int    // bar width, the percent capped at 100
	Gap    bool   // true when the category is under-covered
}

// PortfolioSection groups everything shown for one portfolio. Sections are
// rendered folded (<details>) so the report opens on the comparison.
type PortfolioSection struct {
	Name          string
	Subtitle      string // optional hint shown next to the name (e.g. rebalancing override)
	ChartSVG      template.HTML
	CoverageLabel string        // heading for the coverage chart
	Coverage      []CoverageBar // macro-regime or factor coverage; empty to omit
	Assets        []AssetRow
	Notes         []string // informational lines (e.g. optimizer choices)
	Warnings      []string
}

// StatCell is one value of the statistics table; Best cells are highlighted.
type StatCell struct {
	Text string
	Best bool
}

// StatRow is one metric across all portfolios.
type StatRow struct {
	Label string
	Hint  string
	Cells []StatCell
}

// Page is the full document model.
type Page struct {
	Title           string
	GeneratedAt     string
	RebalanceDays   int
	Portfolios      []PortfolioSection
	CompareSVG      template.HTML // top overview curve (comparison, or the single portfolio)
	OverviewHeading string        // heading for the overview chart section
	UnderwaterSVG   template.HTML // drawdown chart over the common period
	CommonStart     string
	CommonEnd       string
	PortfolioNames  []string
	StatRows        []StatRow
	Footnotes       []string
}

var tpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
       margin: 2rem auto; max-width: 1020px; padding: 0 1rem; color: #1a1a1a; background: #fff; }
h1 { font-size: 1.5rem; margin-bottom: .2rem; }
h2 { font-size: 1.15rem; margin-top: 2.4rem; border-bottom: 1px solid #ddd; padding-bottom: .3rem; }
p.meta { color: #666; font-size: .9rem; margin-top: 0; }
svg { max-width: 100%; height: auto; }
table { border-collapse: collapse; margin: 1rem 0; font-size: .9rem; }
th, td { border: 1px solid #ddd; padding: .35rem .6rem; text-align: left; }
th { background: #f5f5f5; font-weight: 600; }
td.num, th.num { text-align: right; font-variant-numeric: tabular-nums; }
td.best { background: #c9f2c9; font-weight: 600; }
p.warn { color: #9a6700; font-size: .85rem; margin: .25rem 0; }
p.note { color: #57606a; font-size: .85rem; margin: .25rem 0; }
details.pf { margin-top: 1.4rem; border-bottom: 1px solid #ddd; padding-bottom: .4rem; }
details.pf > summary { cursor: pointer; padding: .3rem 0; list-style-position: outside; }
details.pf > summary:hover { color: #000; }
.pf-name { font-size: 1.15rem; font-weight: 600; }
.pf-sub { color: #666; font-size: .85rem; margin-left: .5rem; }
ul.notes { color: #666; font-size: .8rem; line-height: 1.5; margin-top: .4rem; }
details.legend { margin: .6rem 0; }
details.legend > summary { cursor: pointer; color: #666; font-size: .85rem; }
details.legend > summary:hover { color: #000; }
.cov { margin: .7rem 0 1rem; font-size: .85rem; }
.cov-title { font-weight: 600; margin-bottom: .3rem; }
.cov-row { display: flex; align-items: center; gap: .5rem; margin: .15rem 0; }
.cov-label { width: 5.5rem; color: #444; }
.cov-track { flex: 0 0 240px; background: #eee; height: .7rem; border-radius: 3px; overflow: hidden; }
.cov-fill { display: block; height: 100%; background: #4a8a5a; }
.cov-val { color: #666; }
.cov-val.gap { color: #9a6700; }
</style>
</head>
<body>
<h1>{{.Title}}</h1>
<p class="meta">Generated on {{.GeneratedAt}} — base 100 at start, rebalanced every {{.RebalanceDays}} days.</p>
{{if .CompareSVG}}
<section>
<h2>{{.OverviewHeading}}</h2>
{{.CompareSVG}}
</section>
{{end}}
<section>
<h2>Statistics — common period {{.CommonStart}} → {{.CommonEnd}}</h2>
<table>
<thead><tr><th>Metric</th>{{range .PortfolioNames}}<th class="num">{{.}}</th>{{end}}</tr></thead>
<tbody>
{{- range .StatRows}}
<tr><td{{if .Hint}} title="{{.Hint}}"{{end}}>{{.Label}}</td>{{range .Cells}}<td class="num{{if .Best}} best{{end}}">{{.Text}}</td>{{end}}</tr>
{{- end}}
</tbody>
</table>
<details class="legend">
<summary>Legend &amp; explanations</summary>
<ul class="notes">
{{- range .Footnotes}}
<li>{{.}}</li>
{{- end}}
</ul>
</details>
{{if .UnderwaterSVG}}
{{.UnderwaterSVG}}
{{end}}
</section>
{{range .Portfolios}}
<details class="pf">
<summary><span class="pf-name">{{.Name}}</span>{{if .Subtitle}} <span class="pf-sub">{{.Subtitle}}</span>{{end}}</summary>
{{.ChartSVG}}
{{if .Coverage}}
<div class="cov">
<div class="cov-title">{{.CoverageLabel}}</div>
{{- range .Coverage}}
<div class="cov-row"><span class="cov-label">{{.Regime}}</span><span class="cov-track"><span class="cov-fill" style="width:{{.Width}}%"></span></span><span class="cov-val{{if .Gap}} gap{{end}}">{{.Pct}} %{{if .Gap}} — gap{{end}}</span></div>
{{- end}}
</div>
{{end}}
<table>
<thead><tr><th class="num">Weight</th><th>Identifier</th><th>Symbol</th><th>Name</th><th>Class</th><th>UCITS</th><th class="num">Fees</th><th>Currency</th><th>History</th><th>Note</th></tr></thead>
<tbody>
{{- range .Assets}}
<tr><td class="num">{{.Weight}}</td><td>{{.ID}}</td><td>{{.Symbol}}</td><td>{{.Name}}</td><td>{{.Class}}</td><td>{{.UCITS}}</td><td class="num">{{.Fees}}</td><td>{{.Currency}}</td><td>{{.History}}</td><td>{{.Note}}</td></tr>
{{- end}}
</tbody>
</table>
{{- range .Notes}}
<p class="note">{{.}}</p>
{{- end}}
{{- range .Warnings}}
<p class="warn">⚠ {{.}}</p>
{{- end}}
</details>
{{end}}
</body>
</html>
`))

// Render writes the HTML document for page to w.
func Render(w io.Writer, page *Page) error {
	return tpl.Execute(w, page)
}
