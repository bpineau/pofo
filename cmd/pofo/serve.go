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
}

func newServer(opt *options, client *marketdata.Client) *server {
	s := &server{
		opt:      opt,
		client:   client,
		sem:      make(chan struct{}, viewParallel),
		examples: knownExamples(),
	}
	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		return renderComparison(ctx, s.client, o, specs)
	}
	return s
}

// handler assembles the constellation mux.
func (s *server) handler(panel *scenario.Panel, labels []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.hub)
	mux.HandleFunc("/view", s.view)
	mux.HandleFunc("/examples/", s.exampleFile)
	mux.Handle("/fire/", http.StripPrefix("/fire", web.Handler(panel, labels)))
	nav := []firebook.NavLink{
		{Label: "Portefeuilles", Href: "/"},
		{Label: "Simulateur", Href: "/fire/"},
	}
	mux.Handle("/book/fr/", http.StripPrefix("/book/fr", firebook.Handler(firebook.WithNav(nav))))
	css := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
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

// hub is the constellation's front door. Placeholder until the designed
// page lands (Task 7).
func (s *server) hub(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<!DOCTYPE html><title>pofo</title><p>hub placeholder</p>")
}
