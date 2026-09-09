package main

import (
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// Every CLI option the comparison model reads reaches it: the adapter is a
// pure mapping, and a field forgotten here silently ignores a flag.
func TestCompareOptionsCarriesEveryFlag(t *testing.T) {
	fw, err := frameworkFor("factors")
	if err != nil {
		t.Fatal(err)
	}
	opt := &options{
		currency: "USD", benchmark: "^GSPC", rebalance: 30,
		start:   time.Date(2010, 1, 4, 0, 0, 0, 0, time.UTC),
		end:     time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
		noSim:   true,
		noFees:  true,
		simdata: datasets.Simdata(),
		fw:      fw,
	}
	got := opt.compareOptions()
	if got.Currency != "USD" || got.Benchmark != "^GSPC" || got.Rebalance != 30 {
		t.Errorf("currency/benchmark/rebalance came through as %q/%q/%d", got.Currency, got.Benchmark, got.Rebalance)
	}
	if !got.Start.Equal(opt.start) || !got.End.Equal(opt.end) {
		t.Errorf("window came through as %s..%s", got.Start, got.End)
	}
	if !got.NoSim || !got.NoFees {
		t.Errorf("NoSim=%v NoFees=%v, both were asked for", got.NoSim, got.NoFees)
	}
	if got.Simdata == nil {
		t.Error("the simulated-history source was dropped")
	}
	if got.Framework.Name != fw.Name {
		t.Errorf("framework came through as %q, want %q", got.Framework.Name, fw.Name)
	}
}

// The CLI path takes no web chrome at all: the zero Decoration is what keeps
// a standalone report byte-identical to what it always was.
func TestDecorationIsEmptyOffTheWeb(t *testing.T) {
	opt := &options{composer: "<div>composer</div>", fireHref: map[string]string{"p": "/fire/"}}
	if got := opt.decoration(); got.SkinCSS != "" || got.SiteNav != "" || got.Composer != "" || got.FireHref != nil {
		t.Errorf("the CLI path decorated the report: %+v", got)
	}
}

// Under -serve the report gets the warm skin, the site navigation and
// whatever the mount injected (composer panel, per-portfolio simulator links).
func TestDecorationOnTheWeb(t *testing.T) {
	opt := &options{web: true, composer: "<div id=composer></div>", fireHref: map[string]string{"p": "/fire/e/p/"}}
	got := opt.decoration()
	if !strings.Contains(string(got.SkinCSS), ".site-nav{") {
		t.Error("the site-nav CSS is missing from the skin")
	}
	if !strings.Contains(string(got.SiteNav), `href="/visualizer"`) {
		t.Errorf("the site nav does not link the visualizer: %q", got.SiteNav)
	}
	if got.Composer != opt.composer || got.FireHref["p"] != "/fire/e/p/" {
		t.Errorf("the mount's own injections were dropped: %+v", got)
	}
}
