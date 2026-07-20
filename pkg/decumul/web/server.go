package web

import (
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
	indexPage := []byte(strings.Replace(string(indexRaw), "<!--topnav-->", renderTopnav(cfg.nav), 1))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexPage)
			return
		}
		fileSrv.ServeHTTP(w, r)
	})
	// The FIRE book (pkg/firebook), linked discreetly from the page's
	// "How this machine works" fold. The language segment leaves room for
	// the planned English translation at /firebook/en/.
	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr", firebook.Handler()))
	// The shared visual identity (webui.CSS) is served here so both HTML
	// surfaces link the same stylesheet; the report inlines the same bytes.
	mux.HandleFunc("/theme.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(webui.CSS))
	})
	mux.HandleFunc("/fonts.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(webui.FontsCSS))
	})
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
