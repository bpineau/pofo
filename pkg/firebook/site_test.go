package firebook

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// siteServer mounts both editions where BookSite says they live, plus the
// three root files, exactly as a pofo server does.
func siteServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(French.HomePath, http.StripPrefix(strings.TrimSuffix(French.HomePath, "/"),
		French.Handler(WithAlternate(English.HomePath, English))))
	mux.Handle(English.HomePath, http.StripPrefix(strings.TrimSuffix(English.HomePath, "/"),
		English.Handler(WithAlternate(French.HomePath, French))))
	BookSite(Page{Path: "/", Title: "pofo", Note: "the front door"}).Handle(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSitemap(t *testing.T) {
	srv := siteServer(t)
	code, body := get(t, srv, "/sitemap.xml")
	if code != http.StatusOK {
		t.Fatalf("sitemap: status %d", code)
	}
	var set struct {
		XMLName xml.Name
		URLs    []struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal([]byte(body), &set); err != nil {
		t.Fatalf("sitemap is not well-formed XML: %v", err)
	}
	if set.XMLName.Local != "urlset" {
		t.Errorf("root element is %q, want urlset", set.XMLName.Local)
	}
	locs := make(map[string]bool, len(set.URLs))
	for _, u := range set.URLs {
		if !strings.HasPrefix(u.Loc, srv.URL+"/") {
			t.Errorf("loc %q is not an absolute URL on this origin", u.Loc)
		}
		if locs[u.Loc] {
			t.Errorf("loc %q listed twice", u.Loc)
		}
		if u.LastMod != "" {
			t.Errorf("loc %q carries a lastmod; the embedded book has no honest date", u.Loc)
		}
		locs[u.Loc] = true
	}
	want := []string{srv.URL + "/"}
	for _, e := range Editions {
		want = append(want, srv.URL+e.HomePath, srv.URL+e.HomePath+e.EPUBFileName)
		for _, cat := range e.Categories {
			for _, a := range cat.Articles {
				want = append(want, srv.URL+e.HomePath+a.Slug)
			}
		}
	}
	for _, w := range want {
		if !locs[w] {
			t.Errorf("sitemap misses %s", w)
		}
	}
	if len(locs) != len(want) {
		t.Errorf("sitemap has %d entries, want exactly %d", len(locs), len(want))
	}
	// Every listed page must actually be there.
	for _, path := range []string{
		French.HomePath, French.HomePath + Categories[0].Articles[0].Slug,
		English.HomePath + CategoriesEN[0].Articles[0].Slug,
	} {
		if code, _ := get(t, srv, path); code != http.StatusOK {
			t.Errorf("%s is in the sitemap but answers %d", path, code)
		}
	}
	// The markdown mirrors stay out: they are the same document, not a
	// second page to index.
	if strings.Contains(body, markdownSuffix+"<") {
		t.Error("sitemap lists markdown mirrors; they would read as duplicate content")
	}
}

func TestRobots(t *testing.T) {
	srv := siteServer(t)
	code, body := get(t, srv, "/robots.txt")
	if code != http.StatusOK {
		t.Fatalf("robots.txt: status %d", code)
	}
	if !strings.Contains(body, "User-agent: *\nAllow: /\n") {
		t.Errorf("robots.txt does not allow everything:\n%s", body)
	}
	for _, agent := range aiAgents {
		if !strings.Contains(body, "User-agent: "+agent+"\n") {
			t.Errorf("robots.txt does not name %s", agent)
		}
	}
	if !strings.Contains(body, "Sitemap: "+srv.URL+"/sitemap.xml") {
		t.Errorf("robots.txt does not point at the sitemap:\n%s", body)
	}
	if strings.Contains(body, "Disallow:") {
		t.Errorf("nothing here is off limits; no Disallow expected:\n%s", body)
	}
}

func TestLLMsTXT(t *testing.T) {
	srv := siteServer(t)
	code, body := get(t, srv, "/llms.txt")
	if code != http.StatusOK {
		t.Fatalf("llms.txt: status %d", code)
	}
	if !strings.HasPrefix(body, "# pofo\n\n> ") {
		t.Errorf("llms.txt must open with the H1 and its summary blockquote:\n%.120s", body)
	}
	for _, e := range Editions {
		if !strings.Contains(body, "\n## "+e.SiteName+" ("+e.Lang+")\n") {
			t.Errorf("llms.txt misses the %s section", e.SiteName)
		}
		for _, want := range []string{e.HomePath, e.HomePath + e.EPUBFileName, e.HomePath + "opds.xml"} {
			if !strings.Contains(body, "("+srv.URL+want+")") {
				t.Errorf("llms.txt misses %s", want)
			}
		}
		for _, cat := range e.Categories {
			if len(cat.Articles) == 0 {
				continue
			}
			if !strings.Contains(body, "\n## "+e.SiteName+": "+cat.Title+"\n") {
				t.Errorf("llms.txt misses the part %q of %s", cat.Title, e.SiteName)
			}
			for _, a := range cat.Articles {
				link := "- [" + a.Title + "](" + srv.URL + e.HomePath + a.Slug + markdownSuffix + ")"
				if !strings.Contains(body, link) {
					t.Errorf("llms.txt misses %s", link)
				}
			}
		}
	}
	if !strings.Contains(body, "\n## Optional\n") {
		t.Error("llms.txt misses the Optional section holding the site's own pages")
	}
}

func TestRootFilesAreServedAsText(t *testing.T) {
	srv := siteServer(t)
	for _, c := range []struct{ path, ctype string }{
		{"/robots.txt", "text/plain; charset=utf-8"},
		{"/llms.txt", "text/plain; charset=utf-8"},
		{"/sitemap.xml", "application/xml; charset=utf-8"},
	} {
		resp, err := srv.Client().Get(srv.URL + c.path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if got := resp.Header.Get("Content-Type"); got != c.ctype {
			t.Errorf("%s: Content-Type %q, want %q", c.path, got, c.ctype)
		}
	}
}

// The markdown mirror is the highest-value route for an agent: it must serve
// the source untouched, say where it comes from, and revalidate cheaply.
func TestMarkdownMirror(t *testing.T) {
	srv := siteServer(t)
	art := Categories[0].Articles[0]
	path := French.HomePath + art.Slug + markdownSuffix

	resp, err := srv.Client().Get(srv.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 1<<20)
	n, _ := resp.Body.Read(buf)
	resp.Body.Close()
	body := string(buf[:n])
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s: status %d", path, resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/markdown; charset=utf-8" {
		t.Errorf("Content-Type %q, want text/markdown; charset=utf-8", got)
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Error("markdown mirror carries no ETag")
	}
	head, rest, ok := strings.Cut(body, "\n")
	if !ok || !strings.HasPrefix(head, "<!-- source: "+srv.URL+French.HomePath+art.Slug+" ") {
		t.Errorf("markdown mirror misses its provenance line, got %q", head)
	}
	if !strings.HasPrefix(strings.TrimLeft(rest, "\n"), "# ") {
		t.Error("the source must follow the provenance line unchanged, starting at its '# Title'")
	}
	// Untransformed: what the file holds is what the reader gets.
	raw, err := assets.ReadFile(French.AssetDir + "/" + art.Slug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(body, string(raw)) {
		t.Error("the markdown mirror must serve the source bytes verbatim")
	}

	// Revalidation.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	req.Header.Set("If-None-Match", etag)
	resp2, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotModified {
		t.Errorf("If-None-Match: status %d, want 304", resp2.StatusCode)
	}

	if code, _ := get(t, srv, French.HomePath+"no-such-article"+markdownSuffix); code != http.StatusNotFound {
		t.Errorf("unknown markdown slug: status %d, want 404", code)
	}
	// And the HTML page announces it.
	_, page := get(t, srv, French.HomePath+art.Slug)
	if !strings.Contains(page, `<link rel="alternate" type="text/markdown" href="`+art.Slug+markdownSuffix+`">`) {
		t.Error("the article page does not declare its markdown mirror")
	}
	// The index has no source file, so it has no mirror to declare.
	if _, home := get(t, srv, French.HomePath); strings.Contains(home, `type="text/markdown"`) {
		t.Error("the index declares a markdown mirror it does not serve")
	}
}

func TestCanonicalAndStructuredData(t *testing.T) {
	srv := siteServer(t)
	art := Categories[0].Articles[0]

	// The canonical link is fully qualified (Google only recommends it, but
	// the hreflang pair next to it requires it, and the two must agree), built
	// from the request's origin plus the edition's declared home.
	_, index := get(t, srv, French.HomePath)
	if !strings.Contains(index, `<link rel="canonical" href="`+srv.URL+French.HomePath+`">`) {
		t.Error("index misses its absolute canonical link")
	}
	if !strings.Contains(index, `<meta property="og:url" content="`+srv.URL+French.HomePath+`">`) {
		t.Error("index misses its absolute og:url")
	}
	nodes := structuredData(t, index)
	types := nodeTypes(nodes)
	for _, want := range []string{"WebSite", "Book"} {
		if !types[want] {
			t.Errorf("index structured data has no %s node (got %v)", want, types)
		}
	}
	for _, n := range nodes {
		if n["@type"] != "Book" {
			continue
		}
		work, _ := n["workExample"].(map[string]any)
		if work == nil || work["url"] != French.HomePath+French.EPUBFileName {
			t.Errorf("the Book node does not offer the EPUB as a workExample: %v", work)
		}
		parts, _ := n["hasPart"].([]any)
		if len(parts) != len(Titles()) {
			t.Errorf("the Book node lists %d chapters, want %d", len(parts), len(Titles()))
		}
	}

	_, page := get(t, srv, French.HomePath+art.Slug)
	if !strings.Contains(page, `<link rel="canonical" href="`+srv.URL+French.HomePath+art.Slug+`">`) {
		t.Error("article misses its absolute canonical link")
	}
	nodes = structuredData(t, page)
	types = nodeTypes(nodes)
	for _, want := range []string{"Article", "BreadcrumbList"} {
		if !types[want] {
			t.Errorf("article structured data has no %s node (got %v)", want, types)
		}
	}
	for _, n := range nodes {
		switch n["@type"] {
		case "Article":
			if n["url"] != French.HomePath+art.Slug {
				t.Errorf("Article url = %v, want the canonical one", n["url"])
			}
			if n["headline"] != art.Title || n["description"] != art.Blurb {
				t.Errorf("Article node does not carry the manifest title and blurb: %v", n)
			}
			book, _ := n["isPartOf"].(map[string]any)
			if book == nil || book["@type"] != "Book" || book["name"] != French.SiteName {
				t.Errorf("Article is not part of the Book: %v", book)
			}
			enc, _ := n["encoding"].(map[string]any)
			if enc == nil || enc["contentUrl"] != French.HomePath+art.Slug+markdownSuffix {
				t.Errorf("Article does not declare its markdown encoding: %v", enc)
			}
			if n["articleSection"] != Categories[0].Title {
				t.Errorf("articleSection = %v, want %q", n["articleSection"], Categories[0].Title)
			}
		case "BreadcrumbList":
			items, _ := n["itemListElement"].([]any)
			if len(items) != 3 {
				t.Fatalf("breadcrumb has %d items, want index, part, article", len(items))
			}
			last, _ := items[2].(map[string]any)
			if last["name"] != art.Title {
				t.Errorf("breadcrumb does not end on the article: %v", last)
			}
		}
	}
	// No personal identity anywhere: the book is published, not signed.
	if strings.Contains(page, `"author"`) {
		t.Error("structured data must declare no author")
	}
}

// structuredData parses the page's single JSON-LD block, which is an array of
// schema.org nodes.
func structuredData(t *testing.T, page string) []map[string]any {
	t.Helper()
	const open = `<script type="application/ld+json">`
	_, rest, ok := strings.Cut(page, open)
	if !ok {
		t.Fatal("page carries no JSON-LD block")
	}
	blob, _, ok := strings.Cut(rest, "</script>")
	if !ok {
		t.Fatal("unterminated JSON-LD block")
	}
	var nodes []map[string]any
	if err := json.Unmarshal([]byte(blob), &nodes); err != nil {
		t.Fatalf("JSON-LD does not parse: %v\n%s", err, blob)
	}
	for _, n := range nodes {
		if n["@context"] != "https://schema.org" {
			t.Errorf("node %v has no schema.org context", n["@type"])
		}
	}
	return nodes
}

func nodeTypes(nodes []map[string]any) map[string]bool {
	set := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if s, ok := n["@type"].(string); ok {
			set[s] = true
		}
	}
	return set
}

// Every article's blurb is its meta description, its Open Graph description
// and its llms.txt line. An empty one leaves a page undescribed; a duplicated
// one tells a search engine two pages are the same.
func TestBlurbsAreDistinctDescriptions(t *testing.T) {
	for _, e := range Editions {
		seen := make(map[string]string)
		for _, cat := range e.Categories {
			for _, a := range cat.Articles {
				switch {
				case strings.TrimSpace(a.Blurb) == "":
					t.Errorf("%s/%s: empty blurb, the page would have no description", e.Lang, a.Slug)
				case len(a.Blurb) < 40:
					t.Errorf("%s/%s: blurb is only %d bytes, too thin for a description",
						e.Lang, a.Slug, len(a.Blurb))
				}
				if other, dup := seen[a.Blurb]; dup {
					t.Errorf("%s/%s and %s share the same blurb, hence the same description",
						e.Lang, a.Slug, other)
				}
				seen[a.Blurb] = a.Slug
			}
		}
	}
}
