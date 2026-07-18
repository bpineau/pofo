package report

import (
	"html/template"
	"io"

	"github.com/bpineau/pofo/pkg/webui"
)

// AssetRow is one line of a portfolio composition table.
type AssetRow struct {
	Weight   string
	ID       string
	Symbol   string
	Name     string
	Class    string // catalog asset class (equity, gold…), empty when unknown
	UCITS    string // "yes", "no", "no (KID)" (non-UCITS but EU-retail-buyable wrapper) or "?" when undetermined
	Fees     string // published TER, or "-" when unknown
	Currency string
	History  string
	CWARP    string // per-asset CWARP vs the benchmark (as a 25 % overlay), or "-"
	Note     string
}

// CoverageSeg is one holding's slice of a coverage bar.
type CoverageSeg struct {
	Width float64 // segment width as a percent of the track
	Color string  // fill color (stable per holding across the rows)
	Title string  // tooltip, e.g. "NTSG 25%"
}

// CoverageBar is one category row (a macro regime or a risk factor) of a
// portfolio's coverage chart, split into per-holding segments.
type CoverageBar struct {
	Regime   string        // the category label
	Pct      int           // coverage as a percent of portfolio weight (can exceed 100)
	Gap      bool          // true when the category is under-covered
	Segments []CoverageSeg // contributing holdings, largest first; widths sum to ≤ 100
	Detail   string        // compact contributor line, e.g. "NTSG 25 · WPEA 5"
}

// PortfolioSection groups everything shown for one portfolio. Sections are
// rendered folded (<details>) so the report opens on the comparison.
type PortfolioSection struct {
	Name          string
	Subtitle      string // optional hint shown next to the name (e.g. rebalancing override)
	ChartSVG      template.HTML
	Breakdowns    []template.HTML // composition pies (geography, currency, equity sectors, asset type) as SVGs; empty to omit
	CoverageLabel string          // heading for the coverage chart
	Coverage      []CoverageBar   // macro-regime or factor coverage; empty to omit
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

	Theme template.CSS // shared webui identity, inlined into the document
}

// reportCSS holds the view-specific rules layered on the shared theme.
const reportCSS = `
.pies{display:flex;flex-wrap:wrap;gap:.5rem 1.4rem;justify-content:center;align-items:flex-start;margin:1rem 0}
.pies>svg{flex:1 1 250px;min-width:240px;max-width:340px}
.cov{margin:1rem 0}
.cov-title{font-size:.66rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;color:var(--muted);margin-bottom:.5rem}
.cov-row{display:flex;align-items:center;gap:.8rem;margin:.25rem 0}
.cov-label{width:8.5rem;color:var(--ink-soft);font-size:.8rem}
.cov-track{flex:0 0 clamp(180px,34vw,340px);height:.55rem;border-radius:999px;background:var(--surface-2);border:1px solid var(--line);overflow:hidden;display:flex}
.cov-seg{display:block;height:100%;flex:none}
.cov-val{font-family:var(--mono);font-size:.76rem;font-variant-numeric:tabular-nums;color:var(--ink-soft)}
.cov-val.gap{color:var(--warn-ink)}
.cov-detail{margin:-.1rem 0 .4rem 9.3rem;font-family:var(--mono);font-size:.68rem;color:var(--muted)}
.overview,.stat-scroll{overflow-x:auto}
details.pf{margin-top:.8rem}
.pf-name{font-weight:650}
.pf-sub{color:var(--muted);font-size:.8rem;margin-left:.5rem;font-weight:400}
.pf-body{padding:1rem 1.1rem 1.2rem}
.pf-body>.chart-frame{box-shadow:none;border-color:var(--line)}
.legend{margin-top:.8rem}
.legend .disclosure-body ul{margin:.3rem 0;padding-left:1.15rem;line-height:1.55}
`

var tpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>{{.Theme}}</style>
<style>{{.ReportCSS}}</style>
</head>
<body>
<div class="wrap">

<header class="masthead">
  <span class="mark">pofo<b>/</b>report</span>
  <span class="ctx">portfolio analysis</span>
  <span class="spacer"></span>
  <span class="stamp">{{.GeneratedAt}} · base 100 · rebalanced /{{.RebalanceDays}}d</span>
</header>

<h1>{{.Title}}</h1>
<p class="lede soft">Growth of 100 over the common period, net of tax and fees where known. A comparison tool, not investment advice.</p>

{{if .CompareSVG}}
<div class="section-head"><span class="idx">01</span><h2>{{.OverviewHeading}}</h2></div>
<div class="chart-frame">{{.CompareSVG}}</div>
{{end}}

<div class="section-head"><span class="idx">02</span><h2>Statistics</h2><span class="aside">{{.CommonStart}} → {{.CommonEnd}}</span></div>
<div class="stat-scroll">
<table>
<thead><tr><th>Metric</th>{{range .PortfolioNames}}<th class="n">{{.}}</th>{{end}}</tr></thead>
<tbody>
{{- range .StatRows}}
<tr><td{{if .Hint}} title="{{.Hint}}"{{end}}>{{.Label}}</td>{{range .Cells}}<td class="n{{if .Best}} best{{end}}">{{.Text}}</td>{{end}}</tr>
{{- end}}
</tbody>
</table>
</div>
<details class="legend">
<summary>Legend &amp; explanations</summary>
<div class="disclosure-body">
<ul>
{{- range .Footnotes}}
<li>{{.}}</li>
{{- end}}
</ul>
</div>
</details>
{{if .UnderwaterSVG}}
<div class="chart-frame" style="margin-top:1rem">{{.UnderwaterSVG}}</div>
{{end}}

<div class="section-head"><span class="idx">03</span><h2>Portfolios</h2><span class="aside">composition &amp; coverage</span></div>
{{range .Portfolios}}
<details class="pf">
<summary><span class="pf-name">{{.Name}}</span>{{if .Subtitle}} <span class="pf-sub">{{.Subtitle}}</span>{{end}}</summary>
<div class="pf-body">
<div class="chart-frame">{{.ChartSVG}}</div>
{{if .Breakdowns}}<div class="pies">{{range .Breakdowns}}{{.}}{{end}}</div>{{end}}
{{if .Coverage}}
<div class="cov">
<div class="cov-title">{{.CoverageLabel}}</div>
{{- range .Coverage}}
<div class="cov-row"><span class="cov-label">{{.Regime}}</span><span class="cov-track">{{range .Segments}}<span class="cov-seg" style="width:{{.Width}}%;background:{{.Color}}"{{if .Title}} title="{{.Title}}"{{end}}></span>{{end}}</span><span class="cov-val{{if .Gap}} gap{{end}}">{{.Pct}} %{{if .Gap}} (gap){{end}}</span></div>
{{- if .Detail}}
<div class="cov-detail">{{.Detail}}</div>
{{- end}}
{{- end}}
</div>
{{end}}
<div class="stat-scroll">
<table>
<thead><tr><th class="n">Weight</th><th>Identifier</th><th>Symbol</th><th>Name</th><th>Class</th><th>UCITS</th><th class="n">Fees</th><th>Ccy</th><th>History</th><th class="n">CWARP</th><th>Note</th></tr></thead>
<tbody>
{{- range .Assets}}
<tr><td class="n">{{.Weight}}</td><td class="mono">{{.ID}}</td><td class="mono">{{.Symbol}}</td><td>{{.Name}}</td><td>{{.Class}}</td><td>{{.UCITS}}</td><td class="n">{{.Fees}}</td><td>{{.Currency}}</td><td>{{.History}}</td><td class="n">{{.CWARP}}</td><td>{{.Note}}</td></tr>
{{- end}}
</tbody>
</table>
</div>
{{- range .Notes}}
<p class="note">{{.}}</p>
{{- end}}
{{- range .Warnings}}
<p class="warn">⚠ {{.}}</p>
{{- end}}
</div>
</details>
{{end}}

</div>
</body>
</html>
`))

// ReportCSS exposes the view-specific stylesheet to the template.
func (Page) ReportCSS() template.CSS { return template.CSS(reportCSS) }

// Render writes the HTML document for page to w. The embedded identity
// fonts ride along with the theme so the document stays self-contained.
func Render(w io.Writer, page *Page) error {
	page.Theme = template.CSS(webui.FontsCSS + webui.CSS)
	return tpl.Execute(w, page)
}
