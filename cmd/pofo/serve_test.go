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
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/theme.css", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST theme.css: code=%d, want 405", rec.Code)
	}
	if rec := serveGet(t, h, "/no-such-page"); rec.Code != 404 {
		t.Errorf("unknown path: code=%d, want 404", rec.Code)
	}
	if rec := serveGet(t, h, "/fire/"); rec.Code != 200 {
		t.Errorf("fire mount: code=%d", rec.Code)
	}
	// A per-example FIRE mount for an unknown name 404s before any panel
	// build (the known-name build path fetches quotes and is exercised by the
	// manual smoke test, not here).
	if rec := serveGet(t, h, "/fire/e/nope/"); rec.Code != 404 {
		t.Errorf("unknown fire example: code=%d, want 404", rec.Code)
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

func TestServeHub(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)
	rec := serveGet(t, h, "/")
	if rec.Code != 200 {
		t.Fatalf("hub: code=%d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`action="/view"`, `method="get"`,
		`name="ex" value="dragon-decumulation-household"`,
		`href="/examples/dragon-decumulation-household.txt"`,
		`href="/view?ex=dragon-decumulation-household"`,
		`href="/fire/e/dragon-decumulation-household/"`,
		`href="/fire/"`, `href="/book/fr/"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("hub missing %q", want)
		}
	}
	// The warm book skin is applied (a book paper token, not the instrument
	// default).
	if !strings.Contains(body, "#faf6ef") {
		t.Error("hub missing the warm book skin")
	}
}

func TestServeViewSetsPrefsCookie(t *testing.T) {
	s, _ := testServer(t)
	h := s.handler(nil, nil)

	// Explicit globals set the cookie.
	rec := serveGet(t, h, "/view?ex=claude-dragonlite&currency=USD&rebalance=30")
	cookies := rec.Result().Cookies()
	var got *http.Cookie
	for _, c := range cookies {
		if c.Name == prefsCookie {
			got = c
		}
	}
	if got == nil {
		t.Fatal("no pofo_prefs cookie set")
	}
	if !strings.Contains(got.Value, "currency=USD") || !strings.Contains(got.Value, "rebalance=30") {
		t.Errorf("cookie value = %q", got.Value)
	}

	// A partial URL merges with the stored cookie instead of erasing it.
	req := httptest.NewRequest(http.MethodGet, "/view?ex=claude-dragonlite&sim=off", nil)
	req.AddCookie(&http.Cookie{Name: prefsCookie, Value: got.Value})
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	var merged *http.Cookie
	for _, c := range rec2.Result().Cookies() {
		if c.Name == prefsCookie {
			merged = c
		}
	}
	if merged == nil {
		t.Fatal("no merged cookie set")
	}
	for _, want := range []string{"currency=USD", "rebalance=30", "sim=off"} {
		if !strings.Contains(merged.Value, want) {
			t.Errorf("merged cookie %q missing %q", merged.Value, want)
		}
	}

	// No explicit globals: no Set-Cookie at all.
	rec3 := serveGet(t, h, "/view?ex=claude-dragonlite")
	if len(rec3.Result().Cookies()) != 0 {
		t.Error("bare /view must not set a cookie")
	}
}
