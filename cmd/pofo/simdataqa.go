// -verify-simdata: the reconstruction quality report. Every bundled recipe's
// engine is replayed WITHOUT the real quotes it normally splices in, and put
// against those quotes over the window where they exist, which is the only
// window where a backcast can be judged at all. Output is a self-contained
// HTML page, opened like an ordinary report.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/simgen"
	"github.com/bpineau/pofo/pkg/webui"
)

// runVerifySimdata audits the recipes named in ids (all of them when empty),
// writes the report and opens it.
func runVerifySimdata(ctx context.Context, client *marketdata.Client, opt *options, ids []string) error {
	recipes := simgen.All()
	if len(ids) > 0 {
		recipes = recipes[:0]
		for _, id := range ids {
			r, ok := simgen.Find(id)
			if !ok {
				return fmt.Errorf("no recipe for %q", id)
			}
			recipes = append(recipes, r)
		}
	}
	// Same fetcher as generation, so the audit replays exactly the engine the
	// files are built with: bundled reference series first, network behind.
	fetcher := simgen.WithRefData(datasets.Refdata(), simgen.WithContext(ctx, client))
	groups := simgen.AuditAll(fetcher, recipes, func(id string) { log.Printf("→ %s", id) })

	outPath := opt.out
	if outPath == "" {
		outPath = fmt.Sprintf("/tmp/pofo-simdata-qa-%s.html", time.Now().Format("20060102-150405"))
	}
	page := qaPage{
		Theme:     template.CSS(webui.FontsCSS + webui.CSS),
		Favicon:   template.URL(qaFavicon),
		Generated: time.Now().Format("2006-01-02 15:04"),
	}
	for _, g := range groups {
		sec := qaSection{Title: g.Title, Note: g.Note}
		for _, a := range g.Results {
			sec.Cards = append(sec.Cards, qaCardOf(a))
		}
		page.Sections = append(page.Sections, sec)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	if err := qaTmpl.Execute(f, page); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	log.Printf("reconstruction report written to %s", outPath)
	if !opt.noOpen {
		openBrowser(outPath)
	}
	return nil
}

var qaFavicon = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(webui.FaviconSVG))

type qaPage struct {
	Theme     template.CSS
	Favicon   template.URL
	Generated string
	Sections  []qaSection
}

type qaSection struct {
	Title, Note string
	Cards       []qaCard
}

// qaCard is one recipe as the page shows it: everything preformatted, so the
// template holds layout and no arithmetic.
type qaCard struct {
	ID, Label, Name, Method string
	Anchor                  string
	Err                     string
	Have                    bool

	Reference, RealFrom string
	Window              string
	Short               bool
	Level, Path         string

	CAGRSim, CAGRReal, Delta, Drift string
	Vols, Worst, TEVol              string
	Weekly, Monthly, Daily, Beta    string
	TE                              string

	Chart, DriftChart template.HTML
	Chain             []qaJunction
	Rejected, Caveat  string
}

type qaJunction struct {
	Span, Pair, Corr, Gap, Note string
	Months                      int
}

func qaCardOf(a simgen.AuditResult) qaCard {
	c := qaCard{
		ID: a.ID, Label: qaLabel(a.ID), Name: a.Name, Method: a.Method,
		Anchor: strings.ToLower(strings.ReplaceAll(a.ID, ".", "-")),
		Err:    a.Err, Have: a.Measured(), Caveat: a.Caveat,
		Level: string(a.Level), Path: string(a.Path),
	}
	if len(a.Rejected) > 0 {
		c.Rejected = strings.Join(a.Rejected, ", ")
	}
	if !a.Measured() {
		return c
	}
	c.Reference = a.Reference
	if !a.RealFrom.IsZero() {
		c.RealFrom = a.RealFrom.Format("2006-01-02")
	}
	c.Window = fmt.Sprintf("%.1f y (%s to %s)", a.Years,
		a.Start.Format("2006-01"), a.End.Format("2006-01"))
	c.Short = a.Short
	c.CAGRSim, c.CAGRReal = pctOf(a.CAGRSim), pctOf(a.CAGRReal)
	c.Delta, c.Drift = signedPct(a.Delta), signedPct(a.TotalDrift)
	c.Vols = pctOf(a.VolSim) + " / " + pctOf(a.VolReal)
	c.Worst = pctOf(a.WorstSim) + " / " + pctOf(a.WorstReal)
	c.TE = pctOf(a.TrackingErr)
	c.TEVol = "-"
	if a.VolReal > 0 {
		c.TEVol = fmt.Sprintf("%.2f", a.TrackingErr/a.VolReal)
	}
	c.Daily, c.Weekly, c.Monthly = num2(a.DailyCorr), num2(a.WeeklyCorr), num2(a.MonthlyCorr)
	c.Beta = num2(a.Beta)

	series := []chart.Series{
		qaSeries("real "+a.Reference, a.Real, "#0880A8"),
		qaSeries("engine", a.Engine, "#C2452B"),
	}
	for i, s := range a.Others {
		series = append(series, qaSeries("ref "+s.Symbol, s, chart.PaletteColor(i+2)))
	}
	c.Chart = template.HTML(chart.Line(chart.Options{
		Title: fmt.Sprintf("%s · base 100 at %s", a.ID, a.Start.Format("2006-01-02")),
		Width: 1020, Height: 380,
	}, series))

	dates, vals := a.Drift()
	c.DriftChart = template.HTML(chart.Line(chart.Options{
		Title: "cumulative engine / real (100 = glued)",
		Width: 1020, Height: 190,
		Style: chart.Style{HideLegend: true, YTicks: 4},
	}, []chart.Series{{Name: "engine / real", Dates: dates, Values: vals, Color: "#6D28D9"}}))

	for _, j := range a.Chain {
		row := qaJunction{Span: j.Span, Pair: j.Pair, Months: j.Months, Note: j.Note, Corr: "-", Gap: "-"}
		if j.Measured {
			row.Corr = num2(j.Corr)
			row.Gap = fmt.Sprintf("%+.1f pt/yr", j.GapYear*100)
		}
		c.Chain = append(c.Chain, row)
	}
	return c
}

// qaSeries turns a clipped series into a chart series rebased at 100.
func qaSeries(name string, s *marketdata.Series, color string) chart.Series {
	dates := make([]time.Time, len(s.Points))
	values := make([]float64, len(s.Points))
	for i, p := range s.Points {
		dates[i], values[i] = p.Date, p.Close
	}
	return chart.Series{Name: name, Dates: dates, Values: simgen.Rebase(values), Color: color}
}

// qaLabel is the identifier as a reader knows it: the catalog's first alias or
// its exchange ticker, kept beside the identifier the recipes use. A page that
// lists IE0003R87OG3 and IE00BSPLC413 hides AVWS and ZPRV in plain sight.
func qaLabel(id string) string {
	if l := qaLabels[strings.ToUpper(id)]; !strings.EqualFold(l, id) {
		return l
	}
	return ""
}

// qaLabels indexes the catalog once: identifier to reader-facing name.
var qaLabels = func() map[string]string {
	out := map[string]string{}
	for _, a := range datasets.Catalog() {
		switch {
		case len(a.Aliases) > 0:
			out[strings.ToUpper(a.ID)] = a.Aliases[0]
		case a.Symbol != "":
			out[strings.ToUpper(a.ID)] = strings.SplitN(a.Symbol, ".", 2)[0]
		}
	}
	return out
}()

func pctOf(v float64) string {
	if math.IsNaN(v) {
		return "-"
	}
	return fmt.Sprintf("%.2f %%", v*100)
}

func signedPct(v float64) string {
	if math.IsNaN(v) {
		return "-"
	}
	return fmt.Sprintf("%+.2f %%", v*100)
}

func num2(v float64) string {
	if math.IsNaN(v) {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

var qaTmpl = template.Must(template.New("simdataqa").Parse(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pofo · reconstruction quality</title>
<link rel="icon" type="image/svg+xml" href="{{.Favicon}}">
<style>{{.Theme}}</style>
<style>
body{background:var(--bg);color:var(--ink);margin:0;font-family:var(--sans)}
.wrap{max-width:74rem;margin:0 auto;padding:0 1.4rem 4rem}
.masthead{display:flex;align-items:baseline;gap:.9rem;padding:1.6rem 0 .5rem;border-bottom:1px solid var(--line-strong)}
.mark{font-family:var(--mono);font-weight:650;color:var(--ink)}
.mark b{color:var(--accent)}
.ctx{font-family:var(--mono);font-size:.75rem;letter-spacing:.06em;text-transform:uppercase;color:var(--muted)}
.spacer{flex:1}
.stamp{font-family:var(--mono);font-size:.72rem;color:var(--muted)}
.lede{color:var(--ink-soft);max-width:78ch;line-height:1.6;margin:1.2rem 0 1.8rem}
h2{font-size:1.05rem;margin:2.6rem 0 .3rem;padding-bottom:.35rem;border-bottom:1px solid var(--line)}
p.note{color:var(--muted);font-size:.86rem;line-height:1.55;max-width:80ch;margin:.3rem 0 1rem}
.scroll{overflow-x:auto;margin:.6rem 0 1.4rem}
table.index{border-collapse:collapse;width:100%;min-width:64rem;font-family:var(--mono);
  font-size:.74rem;font-variant-numeric:tabular-nums}
table.index th{text-align:right;color:var(--muted);font-weight:400;padding:.3rem .5rem;white-space:nowrap}
table.index td{text-align:right;padding:.3rem .5rem;border-top:1px solid var(--line);white-space:nowrap}
table.index th:first-child,table.index td:first-child{text-align:left}
table.index td:nth-child(2),table.index td:nth-child(3),
table.index th:nth-child(2),table.index th:nth-child(3){text-align:center}
.card{background:var(--surface);border:1px solid var(--line);border-radius:12px;
  box-shadow:var(--sh);padding:1.1rem 1.3rem 1.2rem;margin:1.1rem 0}
.card h3{margin:0 0 .15rem;font-size:1rem;color:var(--ink)}
.card .meth{color:var(--muted);font-size:.8rem;line-height:1.5;margin:0 0 .7rem;max-width:95ch}
.stats{display:flex;flex-wrap:wrap;gap:1.1rem;font-family:var(--mono);font-size:.74rem;
  font-variant-numeric:tabular-nums;color:var(--ink-soft);margin-top:.5rem}
.stats b{color:var(--ink);font-weight:650}
.badge{display:inline-block;padding:.05rem .5rem;border-radius:999px;font-family:var(--mono);
  font-size:.68rem;color:#fff;vertical-align:.12em}
.ok{background:#35803B}.warn{background:#B45309}.bad{background:#C2452B}.na{background:#7A8294}
.err{color:#C2452B}
.flag{color:#B45309;font-size:.8rem;line-height:1.5;margin:.6rem 0 0;max-width:95ch}
.graft{color:#35803B;font-size:.8rem;line-height:1.5;margin:.6rem 0 0;max-width:95ch}
.seghead{color:var(--muted);font-size:.8rem;margin:.9rem 0 .1rem;max-width:95ch}
table.segs{border-collapse:collapse;font-family:var(--mono);font-size:.72rem;margin:.2rem 0 .3rem}
table.segs th{text-align:left;color:var(--muted);font-weight:400;padding:.2rem .8rem .2rem 0}
table.segs td{padding:.2rem .8rem .2rem 0;border-top:1px solid var(--line)}
td.segnote{color:var(--muted)}
a{color:var(--accent-ink,var(--accent));text-decoration:none}
a:hover{text-decoration:underline}
</style>
</head>
<body>
<div class="wrap">
<header class="masthead">
  <span class="mark">pofo<b>/</b>simdata</span>
  <span class="ctx">reconstruction quality</span>
  <span class="spacer"></span>
  <span class="stamp">{{.Generated}}</span>
</header>
<p class="lede">Every chart replays a recipe's engine <em>without the real quotes it normally splices in</em>
 and lays it over those quotes on the window where they exist, base 100 at the first common date. The second
 panel is the cumulative engine/real ratio: flat means glued, a slope means a systematic return gap.
 Two verdicts are kept apart. <b>Level</b>: does the engine earn the asset's return (|CAGR gap| &gt; 1 %/yr
 warns, &gt; 2.5 %/yr fails). <b>Path</b>: does it move with the asset, on the monthly correlation and on the
 tracking error relative to the asset's own volatility. Worst first inside each family.</p>
{{range .Sections}}
<h2>{{.Title}}</h2>
{{if .Note}}<p class="note">{{.Note}}</p>{{end}}
<div class="scroll">
<table class="index">
<tr><th>asset</th><th>level</th><th>path</th><th>reference</th><th>window</th><th>CAGR engine</th>
    <th>CAGR real</th><th>gap/yr</th><th>total drift</th><th>vol eng./real</th><th>worst day</th>
    <th>corr w.</th><th>corr m.</th><th>TE/vol</th></tr>
{{range .Cards}}<tr>
 <td><a href="#{{.Anchor}}">{{if .Label}}{{.Label}} · {{end}}{{.ID}}</a></td>
 {{if .Have}}
 <td><span class="badge {{.Level}}">{{.Level}}</span></td>
 <td><span class="badge {{.Path}}">{{.Path}}</span></td>
 <td>{{.Reference}}</td>
 <td>{{.Window}}{{if .Short}} ⚠{{end}}</td>
 <td>{{.CAGRSim}}</td><td>{{.CAGRReal}}</td><td>{{.Delta}}</td><td>{{.Drift}}</td>
 <td>{{.Vols}}</td><td>{{.Worst}}</td>
 <td>{{.Weekly}}</td><td><b>{{.Monthly}}</b></td><td>{{.TEVol}}</td>
 {{else}}<td colspan="13" class="err">{{.Err}}</td>{{end}}
</tr>{{end}}
</table>
</div>
{{range .Cards}}
<div class="card" id="{{.Anchor}}">
 <h3>{{if .Label}}{{.Label}} · {{end}}{{.ID}} &middot; {{.Name}}
  {{if .Have}}<span class="badge {{.Level}}">level {{.Level}}</span>
  <span class="badge {{.Path}}">path {{.Path}}</span>{{end}}</h3>
 <p class="meth">{{.Method}}</p>
 {{if .Have}}
 {{.Chart}}
 {{.DriftChart}}
 <div class="stats">
  <span>reference <b>{{.Reference}}</b></span>
  <span>vol engine <b>{{.Vols}}</b> real</span>
  <span>worst day <b>{{.Worst}}</b></span>
  <span>corr d/w/m <b>{{.Daily}} / {{.Weekly}} / {{.Monthly}}</b></span>
  <span>beta <b>{{.Beta}}</b></span>
  <span>TE <b>{{.TE}}</b>/yr</span>
  <span>CAGR <b>{{.CAGRSim}}</b> vs <b>{{.CAGRReal}}</b> (<b>{{.Delta}}</b>/yr, <b>{{.Drift}}</b> cumulated)</span>
 </div>
 {{if .RealFrom}}<p class="graft">A SIM consumer gets the <b>real quotes from {{.RealFrom}}</b>, the reconstruction
  serving only before that date. This card therefore grades the engine on the one window where it is verifiable,
  which is not the window that is consumed.</p>{{end}}
 {{if .Chain}}
 <p class="seghead">Chain of custody, every junction graded on its own overlap (monthly correlation, and the
  deeper record's CAGR minus the nearer one's):</p>
 <table class="segs">
 <tr><th>years filled</th><th>junction</th><th>months</th><th>corr m.</th><th>gap</th><th></th></tr>
 {{range .Chain}}<tr><td>{{.Span}}</td><td>{{.Pair}}</td><td>{{.Months}}</td><td>{{.Corr}}</td>
  <td>{{.Gap}}</td><td class="segnote">{{.Note}}</td></tr>{{end}}
 </table>{{end}}
 {{if .Rejected}}<p class="flag">Reference rejected: {{.Rejected}}.</p>{{end}}
 {{if .Caveat}}<p class="flag">{{.Caveat}}</p>{{end}}
 {{if .Short}}<p class="flag">Short window: read the CAGR gap as mostly noise.</p>{{end}}
 {{else}}<p class="err">{{.Err}}</p>{{if .Caveat}}<p class="flag">{{.Caveat}}</p>{{end}}{{end}}
</div>
{{end}}
{{end}}
</div>
</body></html>
`))
