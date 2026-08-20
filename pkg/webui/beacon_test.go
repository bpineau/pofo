package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testToken = "0123456789abcdef0123456789abcdef"

// htmlPage is a handler standing in for any of pofo's HTML renderers.
func htmlPage(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
	})
}

func get(h http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

func TestBeaconSnippetShape(t *testing.T) {
	if got := BeaconSnippet(""); got != "" {
		t.Fatalf("no token must render nothing, got %q", got)
	}
	want := `<script defer src="https://static.cloudflareinsights.com/beacon.min.js" ` +
		`data-cf-beacon='{"token":"` + testToken + `"}'></script>`
	if got := BeaconSnippet(testToken); got != want {
		t.Errorf("snippet:\n got %q\nwant %q", got, want)
	}
}

// The token is the operator's, and escaped anyway: nothing in it may close the
// attribute, the script element, or the JSON object.
func TestBeaconSnippetEscapesToken(t *testing.T) {
	got := BeaconSnippet(`a'b"c</script><x>&`)
	// The attribute is single-quoted, so the only character that could end it
	// early is the single quote, and the only way out of the tag is "<" or ">".
	// The double quote is left as JSON's own \" escape, which reads back
	// correctly and cannot close anything here.
	attr := strings.TrimSuffix(strings.SplitN(got, "data-cf-beacon='", 2)[1], `'></script>`)
	for _, bad := range []string{"'", "<", ">"} {
		if strings.Contains(attr, bad) {
			t.Errorf("attribute value leaks %q: %s", bad, got)
		}
	}
	if strings.Count(got, "<script") != 1 || !strings.HasSuffix(got, `'></script>`) {
		t.Errorf("snippet is not a single well-formed tag: %s", got)
	}
}

func TestBeaconInjectsIntoHTML(t *testing.T) {
	h := Beacon(testToken, htmlPage("<html><body><p>hi</p></body></html>"))
	rec := get(h)
	want := "<html><body><p>hi</p>" + BeaconSnippet(testToken) + "</body></html>"
	if rec.Body.String() != want {
		t.Errorf("body:\n got %q\nwant %q", rec.Body.String(), want)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("code=%d", rec.Code)
	}
}

// With no token the wrapper is the handler itself: the pages a server emits
// cannot differ by a byte from what it emitted before the feature existed.
func TestBeaconOffIsIdentity(t *testing.T) {
	inner := htmlPage("<html><body>x</body></html>")
	if Beacon("", inner) == nil {
		t.Fatal("nil handler")
	}
	plain, wrapped := get(inner).Body.String(), get(Beacon("", inner)).Body.String()
	if plain != wrapped {
		t.Errorf("token-less wrap changed the body: %q vs %q", plain, wrapped)
	}
	if strings.Contains(wrapped, "cloudflareinsights") {
		t.Error("a token-less page must carry no analytics markup")
	}
}

// Everything that is not an HTML page streams through untouched: the Markdown
// mirrors, the feeds and sitemap, JSON, SVG, and any non-200 status (a redirect
// body, an error page).
func TestBeaconLeavesNonHTMLAlone(t *testing.T) {
	for _, tc := range []struct {
		name, ctype, body string
		code              int
	}{
		{"markdown", "text/markdown; charset=utf-8", "# title\n\n</body>", 200},
		{"atom", "application/atom+xml; charset=utf-8", "<feed></feed>", 200},
		{"sitemap", "application/xml", "<urlset></urlset>", 200},
		{"json", "application/json; charset=utf-8", `{"a":1}`, 200},
		{"svg", "image/svg+xml", "<svg></svg>", 200},
		{"html redirect stub", "text/html; charset=utf-8", `<a href="/x">Moved</a>.`, 301},
		{"not modified", "text/html; charset=utf-8", "", 304},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := Beacon(testToken, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.ctype)
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			rec := get(h)
			if rec.Body.String() != tc.body {
				t.Errorf("body rewritten: %q", rec.Body.String())
			}
			if rec.Code != tc.code {
				t.Errorf("code=%d, want %d", rec.Code, tc.code)
			}
		})
	}
}

// A handler that announced its own length (http.FileServer does) must not ship
// a Content-Length shorter than the injected body.
func TestBeaconDropsStaleContentLength(t *testing.T) {
	body := "<html><body>x</body></html>"
	h := Beacon(testToken, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", "27")
		_, _ = w.Write([]byte(body))
	}))
	rec := get(h)
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		t.Errorf("Content-Length=%q, want it dropped for the rewritten body", cl)
	}
	if len(rec.Body.String()) <= len(body) {
		t.Error("body was not extended")
	}
}

// Two nested wraps (the -serve mux around a pkg/decumul/web mount that carries
// its own token) must still leave exactly one tag on the page.
func TestBeaconIsIdempotent(t *testing.T) {
	h := Beacon(testToken, Beacon(testToken, htmlPage("<html><body>x</body></html>")))
	if n := strings.Count(get(h).Body.String(), BeaconScriptURL); n != 1 {
		t.Errorf("%d tags on the page, want 1", n)
	}
}

// An HTML error page IS a page: a broken link is exactly what an operator wants
// counted, so the tag goes on it too, at its own status.
func TestBeaconCountsErrorPages(t *testing.T) {
	for _, code := range []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError} {
		h := Beacon(testToken, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(code)
			_, _ = w.Write([]byte("<html><body>oops</body></html>"))
		}))
		rec := get(h)
		if rec.Code != code {
			t.Errorf("code=%d, want %d", rec.Code, code)
		}
		if !strings.Contains(rec.Body.String(), BeaconScriptURL) {
			t.Errorf("%d error page carries no tag", code)
		}
	}
}

// A fragment with no </body> still gets the tag, at the end.
func TestBeaconAppendsWithoutBodyTag(t *testing.T) {
	rec := get(Beacon(testToken, htmlPage("<p>fragment</p>")))
	if !strings.HasSuffix(rec.Body.String(), BeaconSnippet(testToken)) {
		t.Errorf("no tag appended: %q", rec.Body.String())
	}
}
