// Adapters from the CLI options to the presentation-neutral pkg/compare
// inputs, plus the -serve web chrome (site nav) injected into the report.
package main

import (
	"html/template"

	"github.com/bpineau/pofo/pkg/compare"
	"github.com/bpineau/pofo/pkg/webui"
)

// siteNavCSS and siteNavHTML style and render the web app's cross-navigation
// bar at the top of the /view report; both are used only when opt.web is set.
const siteNavCSS = `.site-nav{display:flex;gap:1.2rem;align-items:baseline;max-width:min(1180px,94vw);` +
	`margin:0 auto;padding:.8rem clamp(1rem,4vw,2rem) 0;font-family:var(--mono);font-size:.72rem;` +
	`letter-spacing:.06em;text-transform:uppercase}` +
	`.site-nav a{color:var(--muted);text-decoration:none}` +
	`.site-nav a:hover{color:var(--accent-ink)}` +
	`.site-nav a:first-child{color:var(--accent-ink)}`

var siteNavHTML = template.HTML(`<nav class="site-nav">` +
	`<a href="/">Portfolios</a><a href="/fire/">Simulator</a><a href="/firebook/fr/">FIRE book (fr)</a>` +
	`</nav>`)

// compareOptions maps the CLI options onto the pkg/compare inputs.
func (opt *options) compareOptions() compare.Options {
	return compare.Options{
		Currency: opt.currency, Benchmark: opt.benchmark,
		Start: opt.start, End: opt.end, Rebalance: opt.rebalance,
		NoSim: opt.noSim, NoFees: opt.noFees, Simdata: opt.simdata,
		Framework: opt.fw,
	}
}

// decoration builds the web chrome for the report; the zero value (CLI path)
// leaves the standalone report byte-identical.
func (opt *options) decoration() compare.Decoration {
	if !opt.web {
		return compare.Decoration{}
	}
	return compare.Decoration{
		SkinCSS:  template.CSS(webui.WarmSkin + siteNavCSS),
		SiteNav:  siteNavHTML,
		Composer: opt.composer,
		FireHref: opt.fireHref,
	}
}
