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

// ExampleRef is one bundled portfolio the drawer's loader offers. It is a
// plain data record so this package never imports the CLI's example
// collection; the caller (the -serve wiring) translates its own list.
type ExampleRef struct {
	Name  string `json:"name"`            // URL segment: <Base>/e/<Name>/
	Title string `json:"title"`           // human label
	Blurb string `json:"blurb,omitempty"` // one-line description, may be empty
}

// Picker describes the portfolio loader the page offers when it runs on no
// portfolio at all. It is only meaningful where the mounts it links to exist
// (the -serve web app), which is why it arrives as an option rather than
// being assumed.
type Picker struct {
	// Base is the simulator's mount prefix, without a trailing slash
	// ("/firesimulator"): the loader navigates to <Base>/e/<name>/ for a
	// bundled example and <Base>/p/<spec>/ for a composed one.
	Base string `json:"base"`
	// CatalogURL serves the composable-asset catalog as JSON
	// (marketdata.LocalCatalog); the loader's search reads it. Empty
	// disables the composer and leaves the example list.
	CatalogURL string `json:"catalogURL,omitempty"`
	// ViewURL is the comparison report's mount ("/view"), which reads the
	// same p= grammar: the composer offers the draft there as a second
	// reading (charted rather than retired on). Empty drops the link, since
	// only the embedding server knows whether that surface exists.
	ViewURL string `json:"viewURL,omitempty"`
	// Examples are the bundled portfolios, in display order.
	Examples []ExampleRef `json:"examples,omitempty"`
}

// handlerConfig collects Handler options.
type handlerConfig struct {
	nav         []NavLink
	sourceLabel string
	picker      *Picker
	indexNowKey string
	beaconToken string
}

// Option configures Handler.
type Option func(*handlerConfig)

// WithNav adds a cross-navigation to the top bar (e.g. links back to the
// portfolios hub and the FIRE book). The simulator only knows these siblings
// when served inside the -serve web app, so the option is set there and left
// off for the standalone -fire mount, where the bar stays clean.
func WithNav(links []NavLink) Option { return func(c *handlerConfig) { c.nav = links } }

// WithSourceLabel names the market this mount runs on (an example name, a
// portfolio file's base name, "custom portfolio"), for the top bar's
// provenance pill. Left empty the pill reports the generic parametric
// market, which is what an unbound mount really is.
func WithSourceLabel(label string) Option {
	return func(c *handlerConfig) { c.sourceLabel = label }
}

// WithPicker enables the in-drawer portfolio loader: with no panel the page
// otherwise gives no hint that portfolio mode exists at all. The loader only
// builds URLs (<Base>/e/<name>/ or <Base>/p/<spec>/) and navigates to them,
// carrying the current location hash so the visitor's scenario survives the
// move; it adds no server capability here.
func WithPicker(p Picker) Option { return func(c *handlerConfig) { c.picker = &p } }

// WithIndexNowKey publishes the IndexNow ownership key file at the root
// ("/<key>.txt"), which is what lets a deploy push its URLs to the
// participating search engines rather than wait to be crawled again (see
// "pofo -indexnow"). The key is per host, so only a mount that IS a public
// host has one; empty, the default, leaves the feature off entirely.
func WithIndexNowKey(key string) Option {
	return func(c *handlerConfig) { c.indexNowKey = key }
}

// WithBeaconToken puts the Cloudflare Web Analytics tag on every HTML page this
// mount serves: its own page, and the book editions mounted under it. Like the
// IndexNow key it is a property of a public deployment, not of the program, so
// only a mount that IS a public host has one; empty, the default, leaves the
// feature off entirely and every page byte-identical. The tag is cookieless and
// the injection is webui.Beacon's, shared with the -serve constellation, which
// wraps this handler from the outside and needs no option here.
func WithBeaconToken(token string) Option {
	return func(c *handlerConfig) { c.beaconToken = token }
}

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
// names the holdings for the allocation UI. Options (WithNav,
// WithSourceLabel, WithPicker) tune the chrome, which the front end reads
// back from /api/meta; WithIndexNowKey adds the ownership key file, and
// WithBeaconToken puts the analytics tag on every HTML page served here.
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
	// "How this machine works" fold, in both editions. Each is told where the
	// other is mounted (WithAlternate), which cross-links every paired page.
	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr",
		firebook.Handler(firebook.WithAlternate("/firebook/en/", firebook.English))))
	mux.Handle("/firebook/en/", http.StripPrefix("/firebook/en",
		firebook.English.Handler(firebook.WithAlternate("/firebook/fr/", firebook.French))))
	// The machine-readable face of the server: /sitemap.xml, /robots.txt and
	// /llms.txt, covering both book editions and this page. It is the same
	// set of files -serve publishes, minus the surfaces this mount lacks.
	site := firebook.BookSite(firebook.Page{
		Path: "/", Title: "FIRE simulator",
		Note: "stress-test a withdrawal plan against thousands of simulated futures",
	})
	site.IndexNowKey = cfg.indexNowKey
	site.Handle(mux)
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
		meta := map[string]any{
			"labels": labels, "hasPanel": panel != nil,
			"sourceLabel": cfg.sourceLabel,
			"cape":        capeSnapshot(), "capeHistory": capeHistory(),
		}
		if cfg.picker != nil {
			meta["picker"] = cfg.picker
		}
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
			// The risk-based guardrail's table is solved on the page's own
			// central assumptions (blended toward the prior on a short
			// history, CAPE-anchored when the anchor is on), which only the
			// server knows: stamp them before any plan is built.
			_ = json.NewEncoder(w).Encode(compute(pr.withCentral(panel)))
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
	// The analytics tag, if this mount was given one, goes on last so it
	// covers the page AND the book editions mounted above; without a token
	// this returns the mux itself (WithBeaconToken).
	return webui.Beacon(cfg.beaconToken, mux)
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
