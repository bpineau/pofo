// The landing page: the constellation's front door at "/". A calling card,
// not an app: the pofo mark, one sentence, and four doors (the two book
// editions, the portfolio visualizer, the FIRE simulator) on two columns.
// Every section links back here through its own chrome, so the constellation
// stays discoverable from any page.
package main

import (
	"html/template"
	"net/http"

	"github.com/bpineau/pofo/pkg/webui"
)

var landingTmpl = template.Must(template.New("landing").Parse(versionedAssets(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pofo</title>
<meta name="description" content="Portfolio backtesting, retirement simulation and two FIRE handbooks, on one small server.">
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<link rel="stylesheet" href="/fonts.css"><link rel="stylesheet" href="/theme.css">
<style>{{.Skin}}</style>
<style>
body.land{background:
  radial-gradient(920px 480px at 88% -14%,rgba(180,120,60,.08),transparent 62%),var(--bg);
  color:var(--ink-soft);min-height:100vh;margin:0;display:flex;align-items:center;justify-content:center}
.land-shell{max-width:40rem;width:100%;margin:0 auto;padding:4rem 1.3rem 4.5rem;text-align:center}
.land-mark{font-family:var(--mono);font-weight:600;font-size:clamp(2.6rem,8vw,3.4rem);
  letter-spacing:-.04em;line-height:1;color:var(--ink);margin:0}
.land-mark b{color:var(--accent);font-weight:600}
.land-lede{font-family:var(--serif);color:var(--ink-soft);font-size:1.06rem;line-height:1.6;
  margin:.9rem auto 2.4rem;max-width:36ch}
.land-grid{display:grid;grid-template-columns:1fr 1fr;gap:.85rem;text-align:left}
.land-card{display:block;padding:1.05rem 1.15rem 1.1rem;border:1px solid var(--line);
  border-radius:12px;background:var(--surface);box-shadow:var(--sh);text-decoration:none;
  transition:border-color .15s,transform .15s}
.land-card:hover{border-color:var(--accent);transform:translateY(-2px)}
.land-card:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.land-title{display:block;font-family:var(--sans);font-weight:650;font-size:1rem;color:var(--ink)}
.land-blurb{display:block;color:var(--muted);font-size:.84rem;line-height:1.45;margin-top:.3rem}
/* A section that is readable but still a draft: dimmed, and its blurb says so.
   It stays a link, so the work in progress can be followed. */
.land-card.soon{background:transparent;box-shadow:none;border-style:dashed}
.land-card.soon .land-title{color:var(--muted)}
.land-card.soon .land-blurb{color:var(--faint)}
@media(max-width:540px){.land-grid{grid-template-columns:1fr}.land-shell{padding:2.6rem 1.1rem 3rem}}
@media(prefers-reduced-motion:reduce){.land-card{transition:none}.land-card:hover{transform:none}}
</style>
</head><body class="land">
<main class="land-shell">
<h1 class="land-mark">po<b>fo</b></h1>
<p class="land-lede">A quiet workshop for portfolios and the craft of living off them.</p>
<nav class="land-grid" aria-label="Sections">
  <a class="land-card" href="/firebook/fr/">
    <span class="land-title">Fire Book (fr)</span>
    <span class="land-blurb">Le FIRE tranquille, the French handbook of living off your capital.</span>
  </a>
  <a class="land-card soon" href="/firebook/en/">
    <span class="land-title">Fire Book (en)</span>
    <span class="land-blurb">Coming soon</span>
  </a>
  <a class="land-card" href="/visualizer">
    <span class="land-title">Portfolio visualizer</span>
    <span class="land-blurb">Compose portfolios and backtest them side by side.</span>
  </a>
  <a class="land-card" href="/firesimulator/">
    <span class="land-title">Fire Simulator</span>
    <span class="land-blurb">Stress-test a withdrawal plan against thousands of simulated futures.</span>
  </a>
</nav>
</main>
</body></html>`)))

// landing serves the front door. It answers only "/" (the mux routes every
// unmatched path here, so anything else is a 404) and only GET.
func (s *server) landing(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = landingTmpl.Execute(w, struct{ Skin template.CSS }{template.CSS(webui.WarmSkin)})
}
