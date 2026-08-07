package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newAuthServer stubs the two auth endpoints: any hit on / sets the cookie
// on a redirect response (like fc.yahoo.com), /v1/test/getcrumb answers the
// crumb only when that cookie is presented.
func newAuthServer(t *testing.T, crumb string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "test-cookie"})
		http.Redirect(w, r, "https://consent.example/", http.StatusFound)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("A3"); err != nil || c.Value != "test-cookie" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(crumb))
	})
	return httptest.NewServer(mux)
}

// TestYahooAuthPair covers the bootstrap dance, the in-process caching of
// the pair, and its invalidation.
func TestYahooAuthPair(t *testing.T) {
	srv := newAuthServer(t, "abc/123")
	defer srv.Close()
	c := NewClient(t.TempDir())
	c.CookieBase, c.ChartBase = srv.URL, srv.URL

	a, err := c.yahooAuthPair(context.Background())
	if err != nil {
		t.Fatalf("yahooAuthPair: %v", err)
	}
	if a.crumb != "abc/123" || a.cookie == "" {
		t.Fatalf("got crumb %q cookie %q", a.crumb, a.cookie)
	}

	// Cached: a second call must not refetch (server closed to prove it).
	srv.Close()
	if _, err := c.yahooAuthPair(context.Background()); err != nil {
		t.Fatalf("cached pair refetched: %v", err)
	}

	// Invalidate drops the cache: the next call fails against the dead server.
	c.invalidateYahooAuth()
	if _, err := c.yahooAuthPair(context.Background()); err == nil {
		t.Fatal("expected an error after invalidation with the server gone")
	}
}
