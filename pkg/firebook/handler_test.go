package firebook

import (
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, srv *httptest.Server, path string) (int, string) {
	t.Helper()
	resp, err := srv.Client().Get(srv.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(b)
}

func TestHandler(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	code, body := get(t, srv, "/")
	if code != http.StatusOK {
		t.Fatalf("index: status %d", code)
	}
	esc := html.EscapeString
	for _, want := range []string{"Le livre FIRE", esc(Categories[0].Title), esc(Categories[0].Articles[0].Title)} {
		if !strings.Contains(body, want) {
			t.Errorf("index misses %q", want)
		}
	}

	art := Categories[0].Articles[0]
	code, body = get(t, srv, "/"+art.Slug)
	if code != http.StatusOK {
		t.Fatalf("article: status %d", code)
	}
	if !strings.Contains(body, "<h1>"+esc(art.Title)+"</h1>") {
		t.Errorf("article page misses its h1 (%q)", art.Title)
	}
	if strings.Count(body, esc(art.Title)) < 2 {
		t.Errorf("article page should carry the title in <title> and <h1>")
	}
	if !strings.Contains(body, `href="."`) {
		t.Errorf("article page misses the back-to-index link")
	}

	if code, _ := get(t, srv, "/no-such-article"); code != http.StatusNotFound {
		t.Errorf("unknown slug: got status %d, want 404", code)
	}
	for _, css := range []string{"/theme.css", "/fonts.css"} {
		if code, _ := get(t, srv, css); code != http.StatusOK {
			t.Errorf("%s: status %d", css, code)
		}
	}
}

// The handler must work behind a prefix, the way pofo -fire mounts it.
func TestHandlerUnderPrefix(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/livre/", http.StripPrefix("/livre", Handler()))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	if code, _ := get(t, srv, "/livre/"); code != http.StatusOK {
		t.Errorf("prefixed index: status %d", code)
	}
	if code, _ := get(t, srv, "/livre/"+Categories[0].Articles[0].Slug); code != http.StatusOK {
		t.Errorf("prefixed article: status %d", code)
	}
}
