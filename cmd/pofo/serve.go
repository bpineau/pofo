// The -serve mode: the pofo web constellation on one local port: the hub
// and /view visualizer (this file and hub.go), the FIRE explorer under
// /fire/, and the FIRE book under /book/fr/.
package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
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

// server carries the constellation's shared state. render is a field so
// tests can observe requests without running the real pipeline.
type server struct {
	opt      *options
	client   *marketdata.Client
	render   func(ctx context.Context, opt *options, specs []*portfolio.Spec) ([]byte, error)
	sem      chan struct{}
	examples map[string]examples.Info

	// FIRE mounts: fireDefault is the plain /fire/ app (the startup panel);
	// fireByEx caches one app per example, built lazily the first time a
	// /fire/e/<name>/ URL is hit so "Simulate this portfolio" needs no
	// restart. Guarded by fireMu; the build itself runs unlocked (it fetches).
	fireDefault http.Handler
	fireMu      sync.Mutex
	fireByEx    map[string]http.Handler
}

func newServer(opt *options, client *marketdata.Client) *server {
	s := &server{
		opt:      opt,
		client:   client,
		sem:      make(chan struct{}, viewParallel),
		examples: knownExamples(),
		fireByEx: map[string]http.Handler{},
	}
	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		return renderComparison(ctx, s.client, o, specs)
	}
	return s
}

// handler assembles the constellation mux.
func (s *server) handler(panel *scenario.Panel, labels []string) http.Handler {
	mux := http.NewServeMux()
	s.fireDefault = web.Handler(panel, labels)
	mux.HandleFunc("/", s.hub)
	mux.HandleFunc("/view", s.view)
	mux.HandleFunc("/examples/", s.exampleFile)
	mux.HandleFunc("/fire/", s.fire)
	nav := []firebook.NavLink{
		{Label: "Portefeuilles", Href: "/"},
		{Label: "Simulateur", Href: "/fire/"},
	}
	mux.Handle("/book/fr/", http.StripPrefix("/book/fr", firebook.Handler(firebook.WithNav(nav))))
	css := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "GET only", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.HandleFunc("/theme.css", css(webui.CSS))
	mux.HandleFunc("/fonts.css", css(webui.FontsCSS))
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
		Handler:           newServer(opt, client).handler(panel, labels),
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

// fire serves the FIRE simulator. A plain /fire/... path gets the default
// app (the startup panel); a /fire/e/<name>/... path gets an app bound to
// that example's historical panel, so a portfolio picked on the hub opens
// pre-loaded. The FIRE front-end resolves its /api and asset URLs against the
// page's base, so it works unchanged at either mount depth.
func (s *server) fire(w http.ResponseWriter, r *http.Request) {
	if rest, ok := strings.CutPrefix(r.URL.Path, "/fire/e/"); ok {
		name, _, _ := strings.Cut(rest, "/")
		h := s.fireForExample(r.Context(), name)
		if h == nil {
			s.errorPage(w, http.StatusNotFound, "unknown example: "+name)
			return
		}
		http.StripPrefix("/fire/e/"+name, h).ServeHTTP(w, r)
		return
	}
	http.StripPrefix("/fire", s.fireDefault).ServeHTTP(w, r)
}

// fireForExample returns the FIRE app bound to the named example's panel,
// building and caching it on first use. It returns nil for an unknown name.
// The panel build (which fetches quotes) runs outside the lock so distinct
// examples build in parallel; a rare double build of the same name is
// harmless, the last writer wins.
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
	panel, labels := firePanel(buildCtx, s.opt, s.client, []*portfolio.Spec{spec})
	h = web.Handler(panel, labels)
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
	vr, err := parseViewQuery(r.URL.Query())
	if err != nil {
		s.errorPage(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(vr.specs) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
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
var errorTmpl = template.Must(template.New("err").Parse(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pofo · error</title>
<link rel="stylesheet" href="/fonts.css"><link rel="stylesheet" href="/theme.css">
</head><body>
<main style="max-width:38rem;margin:4rem auto;padding:0 1.2rem">
<h1>{{.Status}}</h1><p>{{.Message}}</p><p><a href="/">Back to the portfolios</a></p>
</main></body></html>`))

func (s *server) errorPage(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_ = errorTmpl.Execute(w, struct {
		Status  string
		Message string
	}{http.StatusText(code), msg})
}
