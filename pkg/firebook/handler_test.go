package firebook

import (
	"bytes"
	"encoding/xml"
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

// The EPUB route is served by the Handler itself, so every mount gets the
// download for free: correct status, MIME and attachment headers, a strong
// ETag, cached identical bytes across requests, and a 304 on If-None-Match.
func TestHandlerEPUB(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/le-fire-tranquille.epub")
	if err != nil {
		t.Fatal(err)
	}
	first, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("epub: status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/epub+zip" {
		t.Errorf("epub: Content-Type %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != `attachment; filename="le-fire-tranquille.epub"` {
		t.Errorf("epub: Content-Disposition %q", cd)
	}
	etag := resp.Header.Get("ETag")
	if len(etag) < 3 || etag[0] != '"' || etag[len(etag)-1] != '"' {
		t.Errorf("epub: ETag %q is not a strong quoted tag", etag)
	}
	if len(first) == 0 {
		t.Fatal("epub: empty body")
	}
	// EPUB is a zip; the first bytes are the local file header signature "PK".
	if !strings.HasPrefix(string(first), "PK") {
		t.Errorf("epub: body is not a zip (no PK signature)")
	}

	// A second request serves byte-identical, cached content with the same ETag.
	resp2, err := srv.Client().Get(srv.URL + "/le-fire-tranquille.epub")
	if err != nil {
		t.Fatal(err)
	}
	second, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if !bytes.Equal(first, second) {
		t.Error("epub: second request returned different bytes")
	}
	if resp2.Header.Get("ETag") != etag {
		t.Error("epub: ETag changed between requests")
	}

	// If-None-Match with the current ETag -> 304, no body.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/le-fire-tranquille.epub", nil)
	req.Header.Set("If-None-Match", etag)
	resp3, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body3, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotModified {
		t.Errorf("epub: If-None-Match got status %d, want 304", resp3.StatusCode)
	}
	if len(body3) != 0 {
		t.Errorf("epub: 304 response carried a body (%d bytes)", len(body3))
	}
}

// The OPDS route serves a valid one-entry acquisition catalog with the correct
// content type and a relative acquisition link, so an e-book reader (KOReader)
// can add the feed once and download or refresh the same book file.
func TestHandlerOPDS(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/opds.xml")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("opds: status %d", resp.StatusCode)
	}
	const wantType = "application/atom+xml;profile=opds-catalog;kind=acquisition"
	if ct := resp.Header.Get("Content-Type"); ct != wantType {
		t.Errorf("opds: Content-Type %q, want %q", ct, wantType)
	}

	var feed struct {
		ID      string `xml:"id"`
		Title   string `xml:"title"`
		Entries []struct {
			ID   string `xml:"id"`
			Link struct {
				Rel  string `xml:"rel,attr"`
				Type string `xml:"type,attr"`
				Href string `xml:"href,attr"`
			} `xml:"link"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(body, &feed); err != nil {
		t.Fatalf("opds: does not parse as XML: %v\n%s", err, body)
	}
	if feed.Title != siteName {
		t.Errorf("opds: feed title %q, want %q", feed.Title, siteName)
	}
	if feed.ID != epubIdentifier+":catalog" {
		t.Errorf("opds: feed id %q", feed.ID)
	}
	if len(feed.Entries) != 1 {
		t.Fatalf("opds: want 1 entry, got %d", len(feed.Entries))
	}
	e := feed.Entries[0]
	if e.ID != epubIdentifier {
		t.Errorf("opds: entry id %q, want %q", e.ID, epubIdentifier)
	}
	if e.Link.Rel != "http://opds-spec.org/acquisition" {
		t.Errorf("opds: link rel %q", e.Link.Rel)
	}
	if e.Link.Type != "application/epub+zip" {
		t.Errorf("opds: link type %q", e.Link.Type)
	}
	if e.Link.Href != epubFileName {
		t.Errorf("opds: link href %q, want %q (relative)", e.Link.Href, epubFileName)
	}
}

// The index page carries a discreet, relative EPUB download link.
func TestHandlerIndexEPUBLink(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	_, body := get(t, srv, "/")
	if !strings.Contains(body, `href="le-fire-tranquille.epub"`) {
		t.Error("index misses the relative EPUB download link")
	}
	if !strings.Contains(body, "Version epub") {
		t.Error("index misses the 'Version epub' link label")
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
