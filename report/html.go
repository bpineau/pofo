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
	UCITS    string // "oui", "non" ou "?" si indéterminé
	Fees     string // TER publié, ou — si inconnu
	Currency string
	History  string
	Note     string
}

// PortfolioSection groups everything shown for one portfolio. Sections are
// rendered folded (<details>) so the report opens on the comparison.
type PortfolioSection struct {
	Name     string
	Subtitle string // optional hint shown next to the name (e.g. rebalancing override)
	ChartSVG template.HTML
	Assets   []AssetRow
	Warnings []string
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
	Title          string
	GeneratedAt    string
	RebalanceDays  int
	Portfolios     []PortfolioSection
	CompareSVG     template.HTML // empty when there is a single portfolio
	CommonStart    string
	CommonEnd      string
	PortfolioNames []string
	StatRows       []StatRow
	Footnotes      []string
}

var tpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="fr">
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
details.pf { margin-top: 1.4rem; border-bottom: 1px solid #ddd; padding-bottom: .4rem; }
details.pf > summary { cursor: pointer; padding: .3rem 0; list-style-position: outside; }
details.pf > summary:hover { color: #000; }
.pf-name { font-size: 1.15rem; font-weight: 600; }
.pf-sub { color: #666; font-size: .85rem; margin-left: .5rem; }
ul.notes { color: #666; font-size: .8rem; line-height: 1.5; }
</style>
</head>
<body>
<h1>{{.Title}}</h1>
<p class="meta">Généré le {{.GeneratedAt}} — base 100 au départ, rebalancement tous les {{.RebalanceDays}} jours.</p>
{{if .CompareSVG}}
<section>
<h2>Comparaison — base 100 au {{.CommonStart}}</h2>
{{.CompareSVG}}
</section>
{{end}}
<section>
<h2>Statistiques — période commune {{.CommonStart}} → {{.CommonEnd}}</h2>
<table>
<thead><tr><th>Métrique</th>{{range .PortfolioNames}}<th class="num">{{.}}</th>{{end}}</tr></thead>
<tbody>
{{- range .StatRows}}
<tr><td{{if .Hint}} title="{{.Hint}}"{{end}}>{{.Label}}</td>{{range .Cells}}<td class="num{{if .Best}} best{{end}}">{{.Text}}</td>{{end}}</tr>
{{- end}}
</tbody>
</table>
<ul class="notes">
{{- range .Footnotes}}
<li>{{.}}</li>
{{- end}}
</ul>
</section>
{{range .Portfolios}}
<details class="pf">
<summary><span class="pf-name">{{.Name}}</span>{{if .Subtitle}} <span class="pf-sub">{{.Subtitle}}</span>{{end}}</summary>
{{.ChartSVG}}
<table>
<thead><tr><th class="num">Poids</th><th>Identifiant</th><th>Symbole</th><th>Nom</th><th>UCITS</th><th class="num">Frais</th><th>Devise</th><th>Historique</th><th>Note</th></tr></thead>
<tbody>
{{- range .Assets}}
<tr><td class="num">{{.Weight}}</td><td>{{.ID}}</td><td>{{.Symbol}}</td><td>{{.Name}}</td><td>{{.UCITS}}</td><td class="num">{{.Fees}}</td><td>{{.Currency}}</td><td>{{.History}}</td><td>{{.Note}}</td></tr>
{{- end}}
</tbody>
</table>
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
