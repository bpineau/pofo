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
	Tip   string  // instant-tooltip text (data-tip), e.g. "NTSG 25%"
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
	Name              string
	Subtitle          string // optional hint shown next to the name (e.g. rebalancing override)
	ChartSVG          template.HTML
	ContribSVG        template.HTML   // realized-contribution timeline, trailing-12m window; empty to omit
	ContribMonthlySVG template.HTML   // same timeline, raw monthly window (toggled with ContribSVG)
	Breakdowns        []template.HTML // composition pies (geography, currency, equity sectors, asset type) as SVGs; empty to omit
	CoverageLabel     string          // heading for the coverage chart
	Coverage          []CoverageBar   // macro-regime or factor coverage; empty to omit
	RegimeSVG         template.HTML   // realized contribution per regime (bar matrix); empty to omit
	Assets            []AssetRow
	Notes             []string // informational lines (e.g. optimizer choices)
	Warnings          []string
	FireHref          string // link to the FIRE simulator pre-loaded with this portfolio; empty to omit (the CLI report)
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

	// SkinCSS and SiteNav are set only when the report is served inside the
	// web app (-serve): SkinCSS remaps the theme to the book-warm identity so
	// /view matches the hub and the book; SiteNav is a slim bar linking back
	// to the other surfaces. Both are empty for the standalone CLI report, so
	// its output is unchanged.
	SkinCSS template.CSS
	SiteNav template.HTML
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
.tabs{display:flex;gap:.4rem;margin:0 0 .6rem}
.tbtn{font-family:var(--mono);font-size:.72rem;color:var(--ink-soft);background:var(--surface);border:1px solid var(--line-strong);border-radius:999px;padding:.18rem .75rem;cursor:pointer}
.tbtn.on{background:var(--accent);border-color:var(--accent);color:#FFFFFF}
.overview,.stat-scroll{overflow-x:auto}
details.pf{margin-top:.8rem}
.pf-name{font-weight:650}
.pf-sub{color:var(--muted);font-size:.8rem;margin-left:.5rem;font-weight:400}
.pf-body{padding:1rem 1.1rem 1.2rem}
.pf-body>.chart-frame{box-shadow:none;border-color:var(--line)}
.legend{margin-top:.8rem}
.legend .disclosure-body ul{margin:.3rem 0;padding-left:1.15rem;line-height:1.55}
#xtip{position:fixed;z-index:50;display:none;pointer-events:none;background:var(--surface);border:1px solid var(--line-strong);border-radius:8px;box-shadow:0 4px 16px rgba(22,24,29,.14);padding:.45rem .6rem;font-family:var(--mono);font-size:.72rem;color:var(--ink-soft);max-width:300px}
#xtip .xh{font-weight:700;color:var(--ink);margin-bottom:.25rem}
#xtip .xr{display:flex;align-items:center;gap:.4rem;margin:.1rem 0;white-space:nowrap}
#xtip .xr i{width:8px;height:8px;border-radius:2px;flex:none}
#xtip .xr b{min-width:3.4em;text-align:right;font-variant-numeric:tabular-nums;color:var(--ink)}
`

// reportJS is the report's interaction layer: an instant tooltip for every
// element carrying a data-tip attribute (coverage segments, matrix bars,
// regime strips), and a crosshair-plus-tooltip for charts embedding "stack"
// hover metadata (the contribution timeline). It mirrors, in miniature and
// on the light theme, the FIRE UI's hover layer over the same metadata
// contract (pkg/chart/hover.go). No delay anywhere: native title tooltips
// are deliberately not used.
const reportJS = `
(function(){
"use strict";
var tip=document.createElement("div");tip.id="xtip";document.body.appendChild(tip);
var cross=null;
function hide(){tip.style.display="none";if(cross){cross.remove();cross=null;}}
function place(x,y){tip.style.display="block";var p=12,w=tip.offsetWidth,h=tip.offsetHeight,tx=x+p,ty=y+p;
if(tx+w>innerWidth)tx=x-p-w;if(ty+h>innerHeight)ty=y-p-h;tip.style.left=tx+"px";tip.style.top=ty+"px";}
function hover(svg){if(svg.__hd!==undefined)return svg.__hd;var m=svg.querySelector("metadata.hover");
try{svg.__hd=m?JSON.parse(m.textContent):null;}catch(e){svg.__hd=null;}return svg.__hd;}
function fmt(v){var a=Math.abs(v);return a>=100?v.toFixed(0):a>=10?v.toFixed(1):v.toFixed(2);}
document.addEventListener("pointermove",function(e){
var dt=e.target.closest?e.target.closest("[data-tip]"):null;
if(dt){if(cross){cross.remove();cross=null;}tip.textContent=dt.getAttribute("data-tip");place(e.clientX,e.clientY);return;}
var svg=e.target.closest?e.target.closest("svg"):null;
var hd=svg?hover(svg):null;
if(!hd||!hd.series||hd.kind!=="stack"){hide();return;}
var r=svg.getBoundingClientRect(),vb=svg.viewBox.baseVal;
var px=(e.clientX-r.left)*vb.width/r.width;
if(px<hd.x0-8||px>hd.x1+8){hide();return;}
var xmin=hd.xmin||0,xmax=hd.xmax||0;
if(!(xmax>xmin)){hide();return;}
var i=Math.round(Math.min(Math.max((px-hd.x0)/(hd.x1-hd.x0)*(xmax-xmin)+xmin,0),xmax));
if(!cross||cross.ownerSVGElement!==svg){if(cross)cross.remove();
cross=document.createElementNS("http://www.w3.org/2000/svg","line");
cross.setAttribute("stroke","#CDD2DA");cross.setAttribute("stroke-dasharray","2 3");
cross.setAttribute("pointer-events","none");svg.appendChild(cross);}
var cx=hd.x0+(i-xmin)/(xmax-xmin)*(hd.x1-hd.x0);
cross.setAttribute("x1",cx);cross.setAttribute("x2",cx);
cross.setAttribute("y1",hd.y0);cross.setAttribute("y2",hd.y1);
tip.textContent="";
var head=document.createElement("div");head.className="xh";
head.textContent=(hd.rows&&hd.rows[i]?hd.rows[i]:String(i))+(hd.ylabel?" · "+hd.ylabel:"");
tip.appendChild(head);
var rows=[];
for(var s=0;s<hd.series.length;s++){var sr=hd.series[s];
if(sr.ys&&i<sr.ys.length)rows.push({n:sr.name,c:sr.color,v:sr.ys[i]});}
rows.sort(function(a,b){return b.v-a.v;});
for(var k=0;k<rows.length;k++){var d=document.createElement("div");d.className="xr";
var sw=document.createElement("i");if(rows[k].c)sw.style.background=rows[k].c;
var b=document.createElement("b");b.textContent=fmt(rows[k].v);
var nm=document.createElement("span");nm.textContent=rows[k].n||"";
d.appendChild(sw);d.appendChild(b);d.appendChild(nm);tip.appendChild(d);}
place(e.clientX,e.clientY);
});
document.addEventListener("pointerleave",hide);
addEventListener("scroll",hide,{passive:true});
document.addEventListener("click",function(e){
var btn=e.target.closest?e.target.closest(".tbtn"):null;
if(!btn)return;
var box=btn.closest(".chart-frame");
if(!box)return;
var idx=+btn.getAttribute("data-pane");
var btns=box.querySelectorAll(".tbtn"),panes=box.querySelectorAll(".tpane");
for(var i=0;i<btns.length;i++)btns[i].classList.toggle("on",btns[i]===btn);
for(var j=0;j<panes.length;j++)panes[j].hidden=j!==idx;
hide();
});
})();
`

var tpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>{{.Theme}}</style>
<style>{{.ReportCSS}}</style>
{{if .SkinCSS}}<style>{{.SkinCSS}}</style>{{end}}{{if .HasFireLinks}}<style>.pf-fire{float:right;font-family:var(--mono);font-size:.72rem;letter-spacing:.06em;text-transform:uppercase;color:var(--accent-ink);text-decoration:none;margin-left:1rem}
.pf-fire:hover{text-decoration:underline}</style>{{end}}
</head>
<body>
{{.SiteNav}}
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
<summary><span class="pf-name">{{.Name}}</span>{{if .Subtitle}} <span class="pf-sub">{{.Subtitle}}</span>{{end}}{{if .FireHref}}<a class="pf-fire" href="{{.FireHref}}">Simulate &rarr;</a>{{end}}</summary>
<div class="pf-body">
<div class="chart-frame">{{.ChartSVG}}</div>
{{if .ContribSVG}}<div class="chart-frame" style="margin-top:1rem">
{{- if .ContribMonthlySVG}}
<div class="tabs"><button class="tbtn on" data-pane="0">12m rolling</button><button class="tbtn" data-pane="1">monthly</button></div>
<div class="tpane">{{.ContribSVG}}</div>
<div class="tpane" hidden>{{.ContribMonthlySVG}}</div>
{{- else}}
{{.ContribSVG}}
{{- end}}
</div>{{end}}
{{if .Breakdowns}}<div class="pies">{{range .Breakdowns}}{{.}}{{end}}</div>{{end}}
{{if .Coverage}}
<div class="cov">
<div class="cov-title">{{.CoverageLabel}}</div>
{{- range .Coverage}}
<div class="cov-row"><span class="cov-label">{{.Regime}}</span><span class="cov-track">{{range .Segments}}<span class="cov-seg" style="width:{{.Width}}%;background:{{.Color}}"{{if .Tip}} data-tip="{{.Tip}}"{{end}}></span>{{end}}</span><span class="cov-val{{if .Gap}} gap{{end}}">{{.Pct}} %{{if .Gap}} (gap){{end}}</span></div>
{{- if .Detail}}
<div class="cov-detail">{{.Detail}}</div>
{{- end}}
{{- end}}
</div>
{{end}}
{{if .RegimeSVG}}<div class="chart-frame" style="margin-top:1rem">{{.RegimeSVG}}</div>{{end}}
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
<script>{{.ReportJS}}</script>
</body>
</html>
`))

// ReportCSS exposes the view-specific stylesheet to the template.
func (Page) ReportCSS() template.CSS { return template.CSS(reportCSS) }

// HasFireLinks reports whether any section carries a FireHref. The template
// gates the .pf-fire styling on it, so a report without simulator links (the
// CLI path) stays byte-for-byte identical to before the links existed.
func (p Page) HasFireLinks() bool {
	for _, s := range p.Portfolios {
		if s.FireHref != "" {
			return true
		}
	}
	return false
}

// ReportJS exposes the interaction layer (instant tooltips, crosshair) to
// the template.
func (Page) ReportJS() template.JS { return template.JS(reportJS) }

// Render writes the HTML document for page to w. The embedded identity
// fonts ride along with the theme so the document stays self-contained.
func Render(w io.Writer, page *Page) error {
	page.Theme = template.CSS(webui.FontsCSS + webui.CSS)
	return tpl.Execute(w, page)
}
