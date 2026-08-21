package webui_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bpineau/pofo/pkg/webui"
)

// Beacon wraps a whole mux, so one line covers every HTML page a server serves,
// whichever package rendered it. With no token it returns the handler itself and
// the pages do not change by a byte.
func ExampleBeacon() {
	page := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<html><body><h1>pofo</h1></body></html>")
	})
	serve := func(h http.Handler) string {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		return rec.Body.String()
	}

	fmt.Println(serve(webui.Beacon("", page)))
	fmt.Println(serve(webui.Beacon("demo-site-token", page)))
	// Output:
	// <html><body><h1>pofo</h1></body></html>
	// <html><body><h1>pofo</h1><script defer src="https://static.cloudflareinsights.com/beacon.min.js" data-cf-beacon='{"token":"demo-site-token"}'></script></body></html>
}
