package web

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	iofs "io/fs"
	"net/http"
	"strings"

	"github.com/bpineau/pofo/pkg/firebook"
	"github.com/bpineau/pofo/pkg/scenario"
	"github.com/bpineau/pofo/pkg/webui"
)

// NavLink is one entry of the optional cross-navigation shown in the top bar.
type NavLink struct{ Label, Href string }

// handlerConfig collects Handler options.
type handlerConfig struct{ nav []NavLink }

// Option configures Handler.
type Option func(*handlerConfig)

// WithNav adds a cross-navigation to the top bar (e.g. links back to the
// portfolios hub and the FIRE book). The simulator only knows these siblings
// when served inside the -serve web app, so the option is set there and left
// off for the standalone -fire mount, where the bar stays clean.
func WithNav(links []NavLink) Option { return func(c *handlerConfig) { c.nav = links } }

// renderTopnav builds the top-bar cross-navigation, or "" when there is none
// (so the index page's placeholder simply vanishes).
func renderTopnav(links []NavLink) string {
	if len(links) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<nav class="topnav">`)
	for _, l := range links {
		fmt.Fprintf(&b, `<a href="%s">%s</a>`, html.EscapeString(l.Href), html.EscapeString(l.Label))
	}
	b.WriteString(`</nav>`)
	return b.String()
}

// Handler returns the decumulation UI: the embedded page at / and the
// simulation endpoint at POST /api/sim. A non-nil panel enables the
// portfolio models (bootstrap/cohorts) and live allocation sliders; labels
// names the holdings for the allocation UI. Options (e.g. WithNav) tune the
// chrome.
func Handler(panel *scenario.Panel, labels []string, opts ...Option) http.Handler {
	var cfg handlerConfig
	for _, o := range opts {
		o(&cfg)
	}
	mux := http.NewServeMux()
	// The index page carries a "<!--topnav-->" placeholder; splice the
	// cross-navigation into it once at startup, then serve the page for "/"
	// and hand every other path to the static file server.
	sub := mustSub()
	fileSrv := http.FileServer(http.FS(sub))
	indexRaw, err := iofs.ReadFile(sub, "index.html")
	if err != nil {
		panic(err) // embedded asset; cannot fail at runtime
	}
	// Content-fingerprint the local asset URLs (app.js?v=…, app.css?v=…) so a
	// deploy changes the URL and edge caches (Cloudflare) cannot serve stale
	// bytes: the index page itself is never edge-cached, so the fresh hashes
	// propagate on the next request with no manual purge. theme.css/fonts.css
	// come from pkg/webui and are versioned the same way.
	page := strings.Replace(string(indexRaw), "<!--topnav-->", renderTopnav(cfg.nav), 1)
	page = fingerprintRefs(page, sub, webui.CSS, webui.FontsCSS)
	indexPage := []byte(page)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexPage)
			return
		}
		setImmutableIfVersioned(w, r)
		fileSrv.ServeHTTP(w, r)
	})
	// The FIRE book (pkg/firebook), linked discreetly from the page's
	// "How this machine works" fold. The language segment leaves room for
	// the planned English translation at /firebook/en/.
	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr", firebook.Handler()))
	// The shared visual identity (webui.CSS) is served here so both HTML
	// surfaces link the same stylesheet; the report inlines the same bytes.
	mux.HandleFunc("/theme.css", func(w http.ResponseWriter, r *http.Request) {
		setImmutableIfVersioned(w, r)
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(webui.CSS))
	})
	mux.HandleFunc("/fonts.css", func(w http.ResponseWriter, r *http.Request) {
		setImmutableIfVersioned(w, r)
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(webui.FontsCSS))
	})
	// The tab icon (standalone `pofo -fire`; under -serve the same paths are
	// served by cmd/pofo). Browsers auto-request /favicon.ico; the head also
	// links /favicon.svg. Both return the same SVG bytes.
	favicon := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write([]byte(webui.FaviconSVG))
	}
	mux.HandleFunc("/favicon.svg", favicon)
	mux.HandleFunc("/favicon.ico", favicon)
	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		meta := map[string]any{"labels": labels, "hasPanel": panel != nil, "cape": capeSnapshot(), "capeHistory": capeHistory()}
		if panel != nil {
			meta["weights"] = panel.Weights
			meta["panelMonths"] = panel.Periods()
			// A degenerate fit (panel too short) must not seed the sliders
			// with zeros: omit the figures and let the UI keep its defaults.
			if f := FitParametric(*panel, panel.Weights); f.Valid() {
				meta["mu"] = f.Mu
				meta["sigma"] = f.Sigma
				meta["df"] = f.Df
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(meta)
	})
	mux.HandleFunc("/api/fit", func(w http.ResponseWriter, r *http.Request) {
		if panel == nil {
			http.Error(w, "no portfolio", http.StatusBadRequest)
			return
		}
		var body struct {
			Weights []float64 `json:"weights"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		out := map[string]float64{}
		if f := FitParametric(*panel, body.Weights); f.Valid() {
			out["mu"], out["sigma"], out["df"] = f.Mu, f.Sigma, f.Df
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
	// Every simulation endpoint shares the same shape: POST a Params, get a
	// JSON result. post factors the boilerplate once.
	post := func(path string, compute func(Params) any) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "POST only", http.StatusMethodNotAllowed)
				return
			}
			var pr Params
			if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(compute(pr))
		})
	}
	post("/api/sim", func(pr Params) any { return ComputeWithPanel(pr, panel) })
	post("/api/models", func(pr Params) any { return Models(pr, panel) })
	post("/api/paths", func(pr Params) any { return Paths(pr, panel) })
	post("/api/market", func(pr Params) any { return Market(pr, panel) })
	post("/api/sensitivity", func(pr Params) any { return Sensitivity(pr, panel) })
	post("/api/frontier", func(pr Params) any { return Frontier(pr, panel) })
	post("/api/policyfrontier", func(pr Params) any { return PolicyFrontier(pr, panel) })
	post("/api/solvemenu", func(pr Params) any { return SolveMenu(pr, panel) })
	post("/api/solve", func(pr Params) any { return Solve(pr, panel) })
	post("/api/spending", func(pr Params) any { return Spending(pr, panel) })
	post("/api/decade", func(pr Params) any { return Decade(pr, panel) })
	post("/api/vintages", func(pr Params) any { return Vintages(pr, panel) })
	post("/api/income", func(pr Params) any { return Income(pr, panel) })
	post("/api/lifecycle", func(pr Params) any { return Lifecycle(pr, panel) })
	post("/api/curves", func(pr Params) any { return Curves(pr, panel) })
	return mux
}

// fingerprintRefs rewrites the page's local asset references to carry a
// short content hash (app.js -> app.js?v=<hash>), so a deploy changes every
// URL whose bytes changed and edge/browser caches cannot serve stale assets.
// app.js/app.css are read from the embedded FS; theme.css/fonts.css from the
// shared pkg/webui bytes.
func fingerprintRefs(page string, sub iofs.FS, themeCSS, fontsCSS string) string {
	embedded := func(name string) string {
		b, err := iofs.ReadFile(sub, name)
		if err != nil {
			panic(err) // embedded asset; cannot fail at runtime
		}
		return assetTag(b)
	}
	return strings.NewReplacer(
		`href="app.css"`, `href="app.css?v=`+embedded("app.css")+`"`,
		`src="app.js"`, `src="app.js?v=`+embedded("app.js")+`"`,
		`href="theme.css"`, `href="theme.css?v=`+assetTag([]byte(themeCSS))+`"`,
		`href="fonts.css"`, `href="fonts.css?v=`+assetTag([]byte(fontsCSS))+`"`,
	).Replace(page)
}

// assetTag is the 12-hex-char content fingerprint used in versioned URLs.
func assetTag(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:6])
}

// setImmutableIfVersioned marks a fingerprinted asset request (one carrying a
// v= query) as immutable for a year: the URL changes whenever the bytes do, so
// the response can be cached hard by both the browser and the edge.
func setImmutableIfVersioned(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("v") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
}
