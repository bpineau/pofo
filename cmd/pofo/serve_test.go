package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/portfolio"
)

// testServer returns a server whose render function is a recording fake:
// no fetching, no network, no computation.
func testServer(t *testing.T) (*server, *[][]*portfolio.Spec) {
	t.Helper()
	var calls [][]*portfolio.Spec
	s := newServer(&options{currency: "EUR", benchmark: "^GSPC", rebalance: 90}, nil)
	s.render = func(ctx context.Context, o *options, specs []*portfolio.Spec) ([]byte, error) {
		calls = append(calls, specs)
		return []byte("<html>fake report</html>"), nil
	}
	return s, &calls
}

func serveGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestServeRoutes(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	if rec := serveGet(t, h, "/examples/claude-dragonlite.txt"); rec.Code != 200 ||
		!strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain") ||
		!strings.Contains(rec.Body.String(), "dragon-lite") {
		t.Errorf("examples file: code=%d type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
	// The mux cleans "/examples/../secret.txt" to "/secret.txt" (which the
	// hub then 404s): a safe redirect, never a served file. Go's redirect
	// code for the cleaning has been 301, 307 and 308 across versions, so
	// accept the whole family alongside an outright 404.
	if rec := serveGet(t, h, "/examples/../secret.txt"); rec.Code != 404 &&
		rec.Code != 301 && rec.Code != 307 && rec.Code != 308 {
		t.Errorf("traversal: code=%d, want 404 (or the mux's cleaning redirect)", rec.Code)
	}
	if rec := serveGet(t, h, "/examples/nope.txt"); rec.Code != 404 {
		t.Errorf("unknown example file: code=%d", rec.Code)
	}
	if rec := serveGet(t, h, "/book/fr/"); rec.Code != 200 || !strings.Contains(rec.Body.String(), "book-sitenav") {
		t.Errorf("book: code=%d, navbar wanted", rec.Code)
	}
	if rec := serveGet(t, h, "/theme.css"); !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/css") {
		t.Error("theme.css content type")
	}
	if rec := serveGet(t, h, "/no-such-page"); rec.Code != 404 {
		t.Errorf("unknown path: code=%d, want 404", rec.Code)
	}
	if rec := serveGet(t, h, "/fire/"); rec.Code != 200 {
		t.Errorf("fire mount: code=%d", rec.Code)
	}
}

func TestServeView(t *testing.T) {
	s, calls := testServer(t)
	h := s.handler(nil, nil)

	rec := serveGet(t, h, "/view?ex=claude-dragonlite&rebalance=180")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "fake report") {
		t.Fatalf("view: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("view must be no-store")
	}
	if len(*calls) != 1 || len((*calls)[0]) != 1 || (*calls)[0][0].Name != "claude-dragonlite" {
		t.Errorf("render calls = %+v", *calls)
	}

	if rec := serveGet(t, h, "/view"); rec.Code != 303 || rec.Header().Get("Location") != "/" {
		t.Errorf("empty view: code=%d loc=%q, want 303 to /", rec.Code, rec.Header().Get("Location"))
	}
	if rec := serveGet(t, h, "/view?p=ZZZNOTANID:100"); rec.Code != 400 ||
		!strings.Contains(rec.Body.String(), "not in the local catalog") {
		t.Errorf("catalog gate: code=%d", rec.Code)
	}
}
