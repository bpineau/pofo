// The hub: the constellation's front door. Its primary content is the live
// composer (composer.go), booted empty so a visitor lands on one editable card
// and builds up to six custom portfolios by hand or from a bundled build (the
// "add preset" dropdown), then Runs one /view backtest comparing them; the
// address bar is the shareable /view link the whole time.
//
// The bundled-examples catalog is secondary: folded under the composer in a
// collapsed <details>, it lists every portfolio file with a custom-styled
// checkbox inside one pure-GET form that submits the ticked names, plus a
// defaults row (currency, rebalance, sim), to /view for a side-by-side backtest.
//
// The defaults row and the composer's globals row are both pre-filled from the
// pofo_prefs cookie (server defaults otherwise); when a cookie exists, each
// example row's "Open" link also carries the stored prefs so a one-click path
// honors them while the URL stays fully explicit. The cookie is read
// server-side only.
//
// The page is styled with the shared webui tokens (served from /theme.css and
// /fonts.css) remapped to the FIRE book's warm paper-and-ink identity
// (webui.WarmSkin), so the hub and the /view report read as the book's kin
// while the FIRE simulator keeps the instrument look. Each example row also
// offers to send its portfolio straight to the simulator (/firesimulator/e/<name>/).
// The composer links /composer.css and /composer.js; the folded examples form
// stays fully functional without JavaScript (progressive enhancement, the
// composer simply not booting).
package main

import (
	"html/template"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/bpineau/pofo/examples"
	"github.com/bpineau/pofo/pkg/webui"
)

// hubItem is one catalog row, pre-shaped for the template. Untitled marks the
// files whose first line is not a real title (examples.List degrades those to
// Title == Name): they render as a bare mono id instead of a lowercase pseudo
// title with an empty blurb, so no row ever looks broken.
type hubItem struct {
	Name     string
	Label    string
	Blurb    string
	Untitled bool
}

// hubItems adapts the sorted example list into template rows.
func hubItems() []hubItem {
	list := examples.List()
	items := make([]hubItem, 0, len(list))
	for _, in := range list {
		items = append(items, hubItem{
			Name:     in.Name,
			Label:    in.Title, // examples.List already falls back to Name
			Blurb:    in.Blurb,
			Untitled: in.Title == in.Name,
		})
	}
	return items
}

// hubPrefs is the defaults row's view model: the effective value of each
// control (stored pref if any, else the server default) plus, when a cookie
// exists, the query fragment appended to each row's Open link so one-click
// paths honor the preference while the URL stays fully explicit.
type hubPrefs struct {
	Currency  string // "" = keep native currencies
	Rebalance int
	Sim       string       // "on" or "off"
	Query     template.URL // "&currency=...&rebalance=...&sim=...", or ""
}

func hubPrefsFrom(p prefs, opt *options) hubPrefs {
	hp := hubPrefs{Currency: opt.currency, Rebalance: opt.rebalance, Sim: "on"}
	if opt.noSim {
		hp.Sim = "off"
	}
	stored := false
	if p.currency != nil {
		hp.Currency, stored = *p.currency, true
	}
	if p.rebalance != nil {
		hp.Rebalance, stored = *p.rebalance, true
	}
	if p.sim != nil {
		stored = true
		if !*p.sim {
			hp.Sim = "off"
		} else {
			hp.Sim = "on"
		}
	}
	if stored {
		v := url.Values{}
		// The /view grammar spells "keep native currencies" as the sentinel
		// currency=native; an empty currency= would render the server default.
		if hp.Currency == "" {
			v.Set("currency", "native")
		} else {
			v.Set("currency", hp.Currency)
		}
		v.Set("rebalance", strconv.Itoa(hp.Rebalance))
		v.Set("sim", hp.Sim)
		// Values are validated on read (validCurrency, integer, on/off), so
		// this bypasses html/template's URL escaping safely.
		hp.Query = template.URL("&" + v.Encode())
	}
	return hp
}

var hubTmpl = template.Must(template.New("hub").Parse(versionedAssets(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pofo &middot; portfolio lab</title>
<link rel="stylesheet" href="/fonts.css"><link rel="stylesheet" href="/theme.css">
<style>{{.Skin}}</style>
<style>
body.hub{background:
  radial-gradient(920px 480px at 88% -14%,rgba(180,120,60,.08),transparent 62%),var(--bg);
  color:var(--ink-soft);min-height:100vh;overflow-x:hidden}
.hub-shell{max-width:47rem;margin:0 auto;padding:2.4rem 1.3rem 4rem}
.hub-top{display:flex;align-items:baseline;justify-content:space-between;gap:.6rem 1.4rem;flex-wrap:wrap;
  padding-bottom:1rem;margin-bottom:2.3rem;border-bottom:1px solid var(--line-strong)}
.hub-mark{font-family:var(--mono);font-weight:600;font-size:1rem;letter-spacing:-.02em;color:var(--ink)}
.hub-mark b{color:var(--accent);font-weight:600}
.hub-nav{display:flex;gap:1.4rem;flex-wrap:wrap;font-family:var(--mono);font-size:.7rem;
  letter-spacing:.1em;text-transform:uppercase}
.hub-nav a{color:var(--muted);text-decoration:none;transition:color .15s}
.hub-nav a:hover{color:var(--accent-ink)}
.hub-kicker{font-family:var(--mono);font-size:.7rem;letter-spacing:.16em;text-transform:uppercase;
  color:var(--accent-ink);margin:0 0 .6rem}
.hub-hero h1{font-family:var(--serif);font-weight:600;color:var(--ink);
  font-size:clamp(1.8rem,4.8vw,2.3rem);line-height:1.14;letter-spacing:0;margin:0 0 .7rem}
.hub-lede{color:var(--ink-soft);font-size:1.02rem;line-height:1.6;margin:0;max-width:54ch}
/* the composer is the hero: strip its standalone centring/padding so it aligns
   to the hub column, and keep it visually first */
.hub .cmp{max-width:none;margin:2rem 0 0;padding:0}
/* folded examples catalog */
.hub-examples{margin-top:2.6rem;border-top:1px solid var(--line-strong);padding-top:1.5rem}
.hub-exsum{cursor:pointer;list-style:none;display:flex;align-items:baseline;gap:.6rem;flex-wrap:wrap;padding:.7rem .95rem}
.hub-exsum::-webkit-details-marker{display:none}
.hub-exsum::before{content:"\203A";font-family:var(--mono);color:var(--accent);font-size:1.15rem;
  line-height:1;transition:transform .15s;display:inline-block;transform:translateY(.05rem)}
.hub-examples[open] .hub-exsum::before{transform:translateY(.05rem) rotate(90deg)}
.hub-exsum-t{font-family:var(--sans);font-weight:650;font-size:.98rem;color:var(--ink)}
.hub-exsum-t b{font-family:var(--mono);color:var(--accent-ink);font-weight:600;margin-left:.3rem}
.hub-exsum-h{color:var(--muted);font-size:.82rem}
.hub-form{margin-top:1.3rem;padding:0 .95rem 1.1rem}
.hub-bar{position:sticky;top:0;z-index:5;display:flex;flex-wrap:wrap;align-items:center;
  justify-content:space-between;gap:.5rem 1rem;padding:.8rem 0;background:var(--bg);
  border-bottom:1px solid var(--line-strong)}
.hub-bar .lbl{font-family:var(--sans);font-size:.72rem;font-weight:650;letter-spacing:.1em;
  text-transform:uppercase;color:var(--muted);margin:0}
.hub-bar .lbl b{font-family:var(--mono);font-weight:500;color:var(--accent-ink);letter-spacing:0;margin-left:.45rem}
.hub-go{font-family:var(--sans);font-weight:600;font-size:.86rem;color:#fff;background:var(--accent);
  border:none;border-radius:var(--r-sm);padding:.5rem 1rem;cursor:pointer;box-shadow:var(--sh);
  transition:background .15s}
.hub-go:hover{background:var(--accent-ink)}
.hub-defs{display:flex;gap:.9rem;align-items:center;flex-wrap:wrap}
.hub-defs label{display:flex;gap:.4rem;align-items:center;font-family:var(--mono);
  font-size:.7rem;letter-spacing:.08em;text-transform:uppercase;color:var(--muted)}
.hub-defs select{appearance:none;-webkit-appearance:none;font-family:var(--mono);font-size:.78rem;
  color:var(--ink);background:var(--surface) url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 6'%3E%3Cpath d='M1 1l4 4 4-4' fill='none' stroke='%23847a6a' stroke-width='1.6' stroke-linecap='round'/%3E%3C/svg%3E") no-repeat right .5rem center/9px;
  border:1px solid var(--line-strong);border-radius:var(--r-sm);padding:.28rem 1.4rem .28rem .55rem;cursor:pointer}
.hub-defs select:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.hub-list{list-style:none;margin:.7rem 0 0;padding:0;border:1px solid var(--line);border-radius:var(--r);
  background:var(--surface);box-shadow:var(--sh);overflow:hidden}
.hub-list li{display:flex;flex-wrap:wrap;align-items:center;gap:.35rem .9rem;
  padding:.7rem .95rem;border-top:1px solid var(--line);transition:background .12s}
.hub-list li:first-child{border-top:none}
.hub-list li:has(input:checked){background:var(--accent-wash)}
.hub-pick{flex:1 1 22rem;min-width:0;display:flex;align-items:flex-start;gap:.7rem;cursor:pointer;margin:0}
.hub-pick input{appearance:none;-webkit-appearance:none;flex:0 0 auto;margin:.14rem 0 0;
  width:18px;height:18px;border:1.5px solid var(--line-strong);border-radius:5px;background:var(--surface);
  cursor:pointer;transition:background .12s,border-color .12s}
.hub-pick:hover input{border-color:var(--accent)}
.hub-pick input:checked{background-color:var(--accent);border-color:var(--accent);
  background-image:url("data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2016%2016'%20fill='none'%20stroke='%23fff'%20stroke-width='2.4'%20stroke-linecap='round'%20stroke-linejoin='round'%3E%3Cpath%20d='M3.5%208.4l3%203%206-6.6'/%3E%3C/svg%3E");
  background-size:13px;background-position:center;background-repeat:no-repeat}
.hub-pick input:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.hub-body{min-width:0}
.hub-titlerow{display:flex;align-items:baseline;gap:.5rem;flex-wrap:wrap}
.hub-title{font-family:var(--serif);font-weight:600;color:var(--ink);font-size:1rem;line-height:1.3;overflow-wrap:anywhere}
.hub-title.id{font-family:var(--mono);font-weight:500;font-size:.9rem;letter-spacing:-.01em}
.hub-blurb{display:block;color:var(--ink-soft);font-size:.84rem;line-height:1.4;margin-top:.12rem}
.hub-links{flex:0 0 auto;display:flex;gap:1rem;font-family:var(--mono);font-size:.7rem;
  letter-spacing:.05em;text-transform:uppercase}
.hub-links a{color:var(--accent-ink);text-decoration:none;white-space:nowrap}
.hub-links a:hover{text-decoration:underline}
.hub-foot{margin-top:2.1rem;padding-top:1.1rem;border-top:1px solid var(--line);
  display:flex;justify-content:space-between;flex-wrap:wrap;gap:.5rem;color:var(--muted);font-size:.8rem}
.hub-foot .mono{font-family:var(--mono)}
@media(max-width:440px){
  .hub-shell{padding:1.8rem 1rem 3rem}
  .hub-list li{gap:.15rem .9rem}
  .hub-pick{flex:1 1 100%}
  .hub-links{width:100%;padding-left:1.65rem}
}
</style>
</head><body class="hub">
<div class="hub-shell">
<header class="hub-top">
  <span class="hub-mark">po<b>fo</b></span>
  <nav class="hub-nav">
    <a href="/firesimulator/">Simulator</a>
    <a href="/firebook/fr/">Book-fr</a>
  </nav>
</header>

<section class="hub-hero">
  <p class="hub-kicker">Portfolio lab</p>
  <h1>Put portfolios side by side.</h1>
  <p class="hub-lede">Build portfolios by hand or drop in a bundled build, put up to six on one backtest,
  and share the link.</p>
</section>

{{.Composer}}

<details class="hub-examples">
  <summary class="hub-exsum"><span class="hub-exsum-t">Example portfolios <b>{{len .Items}}</b></span><span class="hub-exsum-h">bundled builds, ready to compare or fork into the composer</span></summary>
  <form class="hub-form" action="/view" method="get">
  <div class="hub-bar">
    <p class="lbl">Example portfolios <b>{{len .Items}}</b></p>
    <div class="hub-defs">
      <label>currency <select name="currency">
        {{range $c := .Currencies}}<option value="{{if $c}}{{$c}}{{else}}native{{end}}"{{if eq $c $.Prefs.Currency}} selected{{end}}>{{if $c}}{{$c}}{{else}}native{{end}}</option>{{end}}
      </select></label>
      <label>rebalance <select name="rebalance">
        {{range $d := .Rebalances}}<option value="{{$d}}"{{if eq $d $.Prefs.Rebalance}} selected{{end}}>{{if $d}}{{$d}} d{{else}}never{{end}}</option>{{end}}
      </select></label>
      <label>sim <select name="sim">
        <option value="on"{{if eq $.Prefs.Sim "on"}} selected{{end}}>on</option>
        <option value="off"{{if eq $.Prefs.Sim "off"}} selected{{end}}>off</option>
      </select></label>
    </div>
    <button class="hub-go" type="submit">Compare selected</button>
  </div>
  <ul class="hub-list">
  {{range .Items}}<li>
    <label class="hub-pick">
      <input type="checkbox" name="ex" value="{{.Name}}">
      <span class="hub-body">
        <span class="hub-titlerow">
          <span class="hub-title{{if .Untitled}} id{{end}}">{{.Label}}</span>
        </span>
        {{if .Blurb}}<span class="hub-blurb">{{.Blurb}}</span>{{end}}
      </span>
    </label>
    <span class="hub-links">
      <a href="/view?ex={{.Name}}{{$.Prefs.Query}}">Open</a>
      <a href="/firesimulator/e/{{.Name}}/">Simulate</a>
      <a href="/examples/{{.Name}}.txt">Source</a>
    </span>
  </li>
  {{end}}</ul>
  </form>
</details>

<footer class="hub-foot">
  <span>No portfolio leaves it.</span>
  <span class="mono">pofo</span>
</footer>
</div>
</body></html>`)))

// hub is the constellation's front door: the examples catalog and the links
// on to the FIRE simulator and the FIRE book. It answers only "/" (the mux
// routes every unmatched path here, so anything else is a 404) and only GET.
func (s *server) hub(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	prefs := hubPrefsFrom(readPrefs(r), s.opt)
	// A stored pref outside the hardcoded option lists would otherwise render a
	// lying select (the browser shows the first option and a submit silently
	// rewrites the pref): surface it as its own selected option instead. The
	// native currency ("") is always present, so it needs no appending.
	currencies := []string{"EUR", "USD", "GBP", "CHF", ""}
	if prefs.Currency != "" && !slices.Contains(currencies, prefs.Currency) {
		currencies = append(currencies, prefs.Currency)
	}
	rebalances := []int{30, 90, 180, 365, 0}
	if !slices.Contains(rebalances, prefs.Rebalance) {
		rebalances = append(rebalances, prefs.Rebalance)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = hubTmpl.Execute(w, struct {
		Skin       template.CSS
		Composer   template.HTML
		Items      []hubItem
		Prefs      hubPrefs
		Currencies []string
		Rebalances []int
	}{template.CSS(webui.WarmSkin), hubComposerMount(prefs, s.presets), hubItems(), prefs, currencies, rebalances})
}
