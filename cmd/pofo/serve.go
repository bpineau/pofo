// The -serve mode: the pofo web constellation on one local port: the hub
// and /view visualizer (this file and hub.go), the FIRE explorer under
// /firesimulator/, and the FIRE book under /firebook/fr/.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/pofo/examples"
	"github.com/bpineau/pofo/pkg/decumul/web"
	"github.com/bpineau/pofo/pkg/firebook"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/scenario"
	"github.com/bpineau/pofo/pkg/webui"
)

// viewTimeout bounds one /view computation; viewParallel bounds how many
// run at once (they are CPU- and fetch-heavy).
//
// viewParallel = 2 is safe: marketdata.Client guards its in-memory series
// and fees caches with c.mu and its Yahoo auth with c.authMu, and its disk
// writes go through writeCacheFile (write a temp file, then os.Rename),
// where each write carries the complete JSON so the worst concurrent
// outcome is a redundant, self-consistent overwrite, never a torn file.
const (
	viewTimeout  = 60 * time.Second
	viewParallel = 2
)

// Static assets are content-fingerprinted: their <link>/<script> URLs carry a
// short hash of the served bytes (/theme.css -> /theme.css?v=<hash>), so a
// deploy that changes an asset changes its URL. The HTML surfaces that link
// them (hub, /view, error page) are dynamic and never edge-cached, and
// Cloudflare keys its cache by full URL (query string included), so the new URL
// is a guaranteed miss that refetches the fresh bytes with no manual purge.
// Mirrors pkg/decumul/web, which does the same for the FIRE page assets.
var (
	themeCSSURL    = assetURL("/theme.css", webui.CSS)
	fontsCSSURL    = assetURL("/fonts.css", webui.FontsCSS)
	composerCSSURL = assetURL("/composer.css", composerCSS)
	composerJSURL  = assetURL("/composer.js", composerJS)
)

// assetURL appends a 12-hex-char content fingerprint to a static asset path.
func assetURL(path, body string) string {
	sum := sha256.Sum256([]byte(body))
	return path + "?v=" + hex.EncodeToString(sum[:6])
}

// versionedAssets rewrites the fixed asset links in an HTML template source to
// their fingerprinted URLs, before the template is parsed.
func versionedAssets(src string) string {
	return strings.NewReplacer(
		`"/theme.css"`, `"`+themeCSSURL+`"`,
		`"/fonts.css"`, `"`+fontsCSSURL+`"`,
		`"/composer.css"`, `"`+composerCSSURL+`"`,
		`"/composer.js"`, `"`+composerJSURL+`"`,
	).Replace(src)
}

// setImmutableIfVersioned marks a fingerprinted asset request (one carrying a
// v= query) cacheable for a year: the URL changes whenever the bytes do.
func setImmutableIfVersioned(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("v") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
}

// fireSiteNav is the FIRE simulator's top-bar cross-navigation under -serve:
// links back to the portfolios hub and the FIRE book. It is passed only here,
// so the standalone -fire mount keeps a clean bar.
var fireSiteNav = []web.NavLink{
	{Label: "Portfolios", Href: "/"},
	{Label: "Book-fr", Href: "/firebook/fr/"},
}

// server carries the constellation's shared state. render is a field so
// tests can observe requests without running the real pipeline.
type server struct {
	opt      *options
	client   *marketdata.Client
	render   func(ctx context.Context, opt *options, specs []*portfolio.Spec) ([]byte, error)
	sem      chan struct{}
	examples map[string]examples.Info
	// presets are the bundled builds the hub's composer offers, shared with
	// every /view mount through viewPresets (computed once: the examples are
	// embedded and immutable).
	presets []composerPreset

	// FIRE mounts: fireDefault is the plain /firesimulator/ app (the startup
	// panel); fireByEx caches one app per example, built lazily the first time
	// a /firesimulator/e/<name>/ URL is hit so "Simulate this portfolio" needs
	// no restart. Guarded by fireMu; the build itself runs unlocked (it fetches).
	fireDefault http.Handler
	fireMu      sync.Mutex
	fireByEx    map[string]http.Handler
	// fireBySpec caches one app per ad-hoc composed spec (/firesimulator/p/<spec>/).
	// Unlike fireByEx (at most one entry per bundled example) it must be
	// bounded: anonymous visitors can mint unlimited distinct specs.
	fireBySpec map[string]http.Handler
	// buildPanel builds a FIRE panel for one spec; a field so tests can
	// stub the fetch-heavy build. Defaults to firePanel.
	buildPanel func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string)
}

func newServer(opt *options, client *marketdata.Client) *server {
	s := &server{
		opt:        opt,
		client:     client,
		sem:        make(chan struct{}, viewParallel),
		examples:   knownExamples(),
		presets:    viewPresets(),
		fireByEx:   map[string]http.Handler{},
		fireBySpec: map[string]http.Handler{},
	}
	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		return renderComparison(ctx, s.client, o, specs)
	}
	s.buildPanel = func(ctx context.Context, spec *portfolio.Spec) (*scenario.Panel, []string) {
		return firePanel(ctx, s.opt, s.client, []*portfolio.Spec{spec})
	}
	return s
}

// handler assembles the constellation mux.
func (s *server) handler(panel *scenario.Panel, labels []string) http.Handler {
	mux := http.NewServeMux()
	s.fireDefault = web.Handler(panel, labels, web.WithNav(fireSiteNav))
	mux.HandleFunc("/", s.hub)
	mux.HandleFunc("/view", s.view)
	mux.HandleFunc("/examples/", s.exampleFile)
	mux.HandleFunc(fireBase+"/", s.fire)
	// The simulator moved from /fire/ to /firesimulator/; keep the old path
	// working with a permanent redirect so existing bookmarks, shared /view
	// links and search results do not break.
	mux.HandleFunc("/fire/", func(w http.ResponseWriter, r *http.Request) {
		target := fireBase + strings.TrimPrefix(r.URL.EscapedPath(), "/fire")
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
	nav := []firebook.NavLink{
		{Label: "Portefeuilles", Href: "/"},
		{Label: "Simulateur", Href: fireBase + "/"},
	}
	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr", firebook.Handler(firebook.WithNav(nav))))
	// The book moved from /book/ to /firebook/; keep the old path working with
	// a permanent redirect so existing bookmarks and links do not break.
	mux.HandleFunc("/book/fr/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/firebook"+strings.TrimPrefix(r.URL.Path, "/book"), http.StatusMovedPermanently)
	})
	css := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "GET only", http.StatusMethodNotAllowed)
				return
			}
			setImmutableIfVersioned(w, r)
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.HandleFunc("/theme.css", css(webui.CSS))
	mux.HandleFunc("/fonts.css", css(webui.FontsCSS))
	mux.HandleFunc("/composer.css", css(composerCSS))
	js := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "GET only", http.StatusMethodNotAllowed)
				return
			}
			setImmutableIfVersioned(w, r)
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.HandleFunc("/composer.js", js(composerJS))
	// The tab icon: one SVG for every surface. Browsers auto-request
	// /favicon.ico; the heads also link /favicon.svg for crispness. Both serve
	// the same bytes (modern browsers render SVG at either path).
	favicon := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write([]byte(webui.FaviconSVG))
	}
	mux.HandleFunc("/favicon.svg", favicon)
	mux.HandleFunc("/favicon.ico", favicon)
	// The local catalog, serialized once: the composer's autocomplete and
	// inline validation read it; the server-side gates stay authoritative.
	catalogJSON, err := json.Marshal(marketdata.LocalCatalog())
	if err != nil {
		panic(err) // embedded data; cannot fail at runtime
	}
	mux.HandleFunc("/catalog.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(catalogJSON)
	})
	return mux
}

// runServe starts the constellation server and blocks until the context
// is canceled (Ctrl-C), mirroring runFire's lifecycle.
func runServe(ctx context.Context, opt *options, client *marketdata.Client, specs []*portfolio.Spec, addr string) error {
	panel, labels := firePanel(ctx, opt, client, specs)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "pofo web app on http://%s/ (Ctrl-C to stop)\n", ln.Addr())
	srv := &http.Server{
		Handler:           logAccess(os.Stdout, newServer(opt, client).handler(panel, labels)),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// logAccess wraps h so every served request prints one NCSA combined-log-format
// line to w (stdout): client IP, timestamp, request line, status, response
// bytes, referer and user agent. Application errors keep going to log's own
// destination (stderr), so the two streams stay separable.
func logAccess(w io.Writer, h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: rw, status: http.StatusOK}
		h.ServeHTTP(rec, r)
		fmt.Fprintf(w, "%s - - [%s] %q %d %d %q %q %s\n",
			clientIP(r),
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method+" "+r.RequestURI+" "+r.Proto,
			rec.status,
			rec.bytes,
			r.Referer(),
			r.UserAgent(),
			time.Since(start).Round(time.Millisecond),
		)
	})
}

// clientIP returns the request's source address: the left-most entry of an
// X-Forwarded-For header when present (the server may sit behind a reverse
// proxy), otherwise the host part of RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// statusRecorder is a minimal http.ResponseWriter that remembers the status
// code and byte count so logAccess can report them after the handler returns.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// fireBase is the public mount of the FIRE simulator. Everything the app
// reaches (its /api and assets) is resolved against document.baseURI, so the
// prefix is a pure routing choice; the legacy /fire/ path 301-redirects here.
const fireBase = "/firesimulator"

// fire serves the FIRE simulator. A plain /firesimulator/... path gets the
// default app (the startup panel); /firesimulator/e/<name>/... gets an app
// bound to that example's historical panel; /firesimulator/p/<spec>/... gets
// an app built from an ad-hoc composed portfolio, <spec> being exactly the
// /view p= grammar carried as one path segment. Naked forms redirect to the
// directory form: the front end resolves /api and asset URLs against
// document.baseURI.
func (s *server) fire(w http.ResponseWriter, r *http.Request) {
	if rest, ok := strings.CutPrefix(r.URL.Path, fireBase+"/e/"); ok {
		name, _, slash := strings.Cut(rest, "/")
		if _, known := s.examples[name]; !known {
			s.errorPage(w, http.StatusNotFound, "unknown example: "+name)
			return
		}
		if !slash {
			http.Redirect(w, r, fireBase+"/e/"+name+"/", http.StatusMovedPermanently)
			return
		}
		h := s.fireForExample(r.Context(), name)
		if h == nil {
			s.errorPage(w, http.StatusNotFound, "unknown example: "+name)
			return
		}
		http.StripPrefix(fireBase+"/e/"+name, h).ServeHTTP(w, r)
		return
	}
	if enc, ok := strings.CutPrefix(r.URL.EscapedPath(), fireBase+"/p/"); ok {
		s.fireComposed(w, r, enc)
		return
	}
	http.StripPrefix(fireBase, s.fireDefault).ServeHTTP(w, r)
}

// fireComposed serves the simulator for an ad-hoc composed portfolio. It
// works on the escaped path so a percent-encoded "/" cannot cross the
// segment boundary, and validates the spec (grammar, byte cap, catalog
// gate, all via adhocSpec) before any redirect or panel build.
func (s *server) fireComposed(w http.ResponseWriter, r *http.Request, enc string) {
	seg, tail, slash := strings.Cut(enc, "/")
	raw, err := url.PathUnescape(seg)
	if err != nil {
		s.errorPage(w, http.StatusBadRequest, "malformed portfolio spec")
		return
	}
	spec, err := adhocSpec(raw, 1)
	if err != nil {
		s.errorPage(w, http.StatusBadRequest, err.Error())
		return
	}
	if !slash {
		http.Redirect(w, r, fireBase+"/p/"+seg+"/", http.StatusMovedPermanently)
		return
	}
	h := s.fireForSpec(r.Context(), raw, spec)
	if h == nil { // client gone while queued for a build slot
		return
	}
	// Serve the app under its mount: hand it the sub-path past the segment
	// (StripPrefix cannot be used verbatim, the escaped prefix would have
	// to match the decoded Path).
	r2 := new(http.Request)
	*r2 = *r
	u := *r.URL
	r2.URL = &u
	p, err := url.PathUnescape("/" + tail)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	u.Path, u.RawPath = p, ""
	if p != "/"+tail {
		u.RawPath = "/" + tail
	}
	h.ServeHTTP(w, r2)
}

// fireSpecCacheMax bounds the composed-panel cache. Eviction is arbitrary
// (map order); the worst case is a rebuild, and sixteen covers a personal
// server's realistic working set.
const fireSpecCacheMax = 16

// panelIncomplete reports whether a freshly built FIRE panel must NOT be
// cached: a nil panel (the build fully degraded), a build whose context was
// canceled or timed out mid-fetch, or a panel missing a holding (a transient
// fetch failure dropped one, so the composition is wrong). Caching an
// incomplete panel would freeze a degraded or wrong-composition simulator
// until eviction; a rebuild is cheap and stays bounded by the render
// semaphore. The rare cost is re-running a build for an asset whose data is
// permanently unavailable, which is preferable to serving a silently wrong
// panel.
func panelIncomplete(panel *scenario.Panel, labels []string, buildCtx context.Context, holdings int) bool {
	return panel == nil || buildCtx.Err() != nil || len(labels) < holdings
}

// fireForSpec returns the FIRE app for one composed spec, building and
// caching it on first use. Same locking pattern as fireForExample: the
// fetch-heavy build runs unlocked, a rare double build is harmless, the
// first writer wins. Builds share the /view render semaphore so a burst of
// distinct specs cannot stack unbounded quote fetches.
func (s *server) fireForSpec(ctx context.Context, key string, spec *portfolio.Spec) http.Handler {
	s.fireMu.Lock()
	h, ok := s.fireBySpec[key]
	s.fireMu.Unlock()
	if ok {
		return h
	}
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil
	}
	// Re-check under the lock now that we hold a build slot: a double-clicked
	// cold link would otherwise build twice and occupy both shared slots.
	s.fireMu.Lock()
	h, ok = s.fireBySpec[key]
	s.fireMu.Unlock()
	if ok {
		return h
	}
	buildCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	panel, labels := s.buildPanel(buildCtx, spec)
	h = web.Handler(panel, labels, web.WithNav(fireSiteNav))
	if panelIncomplete(panel, labels, buildCtx, len(spec.Holdings)) {
		// The build is degraded or partial: serve this request but do NOT
		// cache it, so a transient failure is not frozen into a permanently
		// degraded or wrong-composition simulator. A later visit rebuilds;
		// the semaphore keeps those retries bounded.
		return h
	}
	s.fireMu.Lock()
	if existing, ok := s.fireBySpec[key]; ok {
		h = existing
	} else {
		if len(s.fireBySpec) >= fireSpecCacheMax {
			for k := range s.fireBySpec {
				delete(s.fireBySpec, k)
				break
			}
		}
		s.fireBySpec[key] = h
	}
	s.fireMu.Unlock()
	return h
}

// fireForExample returns the FIRE app bound to the named example's panel,
// building and caching it on first use. It returns nil for an unknown name.
// The panel build (which fetches quotes) runs outside the lock so distinct
// examples build in parallel; a rare double build of the same name is
// harmless, the first writer wins.
func (s *server) fireForExample(ctx context.Context, name string) http.Handler {
	if _, ok := s.examples[name]; !ok {
		return nil
	}
	s.fireMu.Lock()
	h, ok := s.fireByEx[name]
	s.fireMu.Unlock()
	if ok {
		return h
	}
	raw, err := examples.FS.ReadFile(name + ".txt")
	if err != nil {
		return nil
	}
	spec, err := portfolio.Parse(name, strings.NewReader(string(raw)))
	if err != nil {
		return nil
	}
	buildCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	panel, labels := s.buildPanel(buildCtx, spec)
	h = web.Handler(panel, labels, web.WithNav(fireSiteNav))
	if panelIncomplete(panel, labels, buildCtx, len(spec.Holdings)) {
		// A degraded or partial build must not be cached, or a transient
		// failure freezes this example into a permanently degraded or
		// wrong-composition simulator. Serve it once, rebuild next time.
		return h
	}
	s.fireMu.Lock()
	if existing, ok := s.fireByEx[name]; ok {
		h = existing
	} else {
		s.fireByEx[name] = h
	}
	s.fireMu.Unlock()
	return h
}

// view renders the comparison page for the portfolios encoded in the URL.
func (s *server) view(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	vr, err := parseViewQuery(r.URL.Query(), s.opt)
	if err != nil {
		s.errorPage(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(vr.specs) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// Remember explicit, valid global preferences (currency, rebalance,
	// sim) so the hub can pre-fill its defaults row. Rendering itself never
	// reads the cookie: a /view URL is a pure function of its query string.
	if p, changed := readPrefs(r).merge(r.URL.Query()); changed {
		http.SetCookie(w, p.cookie())
	}
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-r.Context().Done():
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), viewTimeout)
	defer cancel()
	page, err := s.render(ctx, vr.serverOptions(s.opt), vr.specs)
	if err != nil {
		if ctx.Err() != nil {
			s.errorPage(w, http.StatusGatewayTimeout, "the computation timed out")
			return
		}
		log.Printf("view %s: %v", r.URL.RawQuery, err)
		s.errorPage(w, http.StatusInternalServerError, "the comparison failed; see the server log")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(page)
}

// exampleFile serves one embedded portfolio file as plain text.
func (s *server) exampleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/examples/"), ".txt")
	if _, ok := s.examples[name]; !ok {
		http.NotFound(w, r)
		return
	}
	raw, err := examples.FS.ReadFile(name + ".txt")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(raw)
}

// errorTmpl is the shared error page: title, message, a way home.
var errorTmpl = template.Must(template.New("err").Parse(versionedAssets(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pofo · error</title>
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<link rel="stylesheet" href="/fonts.css"><link rel="stylesheet" href="/theme.css">
</head><body>
<main style="max-width:38rem;margin:4rem auto;padding:0 1.2rem">
<h1>{{.Status}}</h1><p>{{.Message}}</p><p><a href="/">Back to the portfolios</a></p>
</main></body></html>`)))

func (s *server) errorPage(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_ = errorTmpl.Execute(w, struct {
		Status  string
		Message string
	}{http.StatusText(code), msg})
}
