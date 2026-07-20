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
	for _, want := range []string{"Le FIRE tranquille", esc(Categories[0].Title), esc(Categories[0].Articles[0].Title)} {
		if !strings.Contains(body, want) {
			t.Errorf("index misses %q", want)
		}
	}
	// SEO: every page carries a meta description and Open Graph tags.
	for _, want := range []string{`<meta name="description"`, `property="og:title"`, `application/ld+json`} {
		if !strings.Contains(body, want) {
			t.Errorf("index misses SEO markup %q", want)
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
	// SEO: the article's meta description is its manifest blurb.
	if art.Blurb != "" && !strings.Contains(body, `<meta name="description" content="`+esc(art.Blurb)+`">`) {
		t.Errorf("article page misses its blurb as the meta description")
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

func TestHandlerNav(t *testing.T) {
	// Without the option: no navbar anywhere (offline/print contract).
	plain := httptest.NewServer(Handler())
	defer plain.Close()
	// The .book-sitenav CSS rule ships in bookCSS unconditionally (inert when
	// no bar is emitted); the offline/print contract is that no <nav> element
	// is rendered, so assert on the element, not the stylesheet substring.
	if _, body := get(t, plain, "/"); strings.Contains(body, `class="book-sitenav"`) {
		t.Error("navbar present without WithNav")
	}

	nav := []NavLink{{Label: "Portefeuilles", Href: "/"}, {Label: "Simulateur", Href: "/fire/"}}
	site := httptest.NewServer(Handler(WithNav(nav)))
	defer site.Close()
	slug := Categories[0].Articles[0].Slug
	for _, path := range []string{"/", "/" + slug} {
		_, body := get(t, site, path)
		if !strings.Contains(body, `class="book-sitenav"`) {
			t.Errorf("%s: navbar missing", path)
		}
		if !strings.Contains(body, `>Sommaire</a>`) || !strings.Contains(body, `>Simulateur</a>`) {
			t.Errorf("%s: navbar links missing", path)
		}
	}
	if _, body := get(t, site, "/"); !strings.Contains(body, "@media print{.book-sitenav{display:none}}") {
		t.Error("navbar not hidden in print")
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
