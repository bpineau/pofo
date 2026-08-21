package firebook

import (
	"encoding/xml"
	"net/http"
	"strings"
	"testing"
	"time"
)

// feedDoc is the parsed shape the tests read a served feed back through.
type feedDoc struct {
	XMLName xml.Name `xml:"feed"`
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Updated string   `xml:"updated"`
	Links   []struct {
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Entries []struct {
		ID      string `xml:"id"`
		Title   string `xml:"title"`
		Updated string `xml:"updated"`
		Summary string `xml:"summary"`
		Link    struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

func parseFeed(t *testing.T, body string) feedDoc {
	t.Helper()
	var doc feedDoc
	if err := xml.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatalf("feed is not well-formed XML: %v", err)
	}
	return doc
}

// articleCount is how many articles an edition's manifest holds, which is how
// many entries its feed must carry.
func articleCount(e *Edition) int {
	n := 0
	for _, cat := range e.Categories {
		n += len(cat.Articles)
	}
	return n
}

// Both editions publish a feed at their own mount, listing every article of
// that edition in reading order, with absolute URLs built from the request.
func TestFeedServedByBothEditions(t *testing.T) {
	srv := siteServer(t)
	for _, e := range Editions {
		code, body := get(t, srv, e.HomePath+FeedFileName)
		if code != http.StatusOK {
			t.Fatalf("%s: status %d", e.Lang, code)
		}
		doc := parseFeed(t, body)
		if doc.XMLName.Space != "http://www.w3.org/2005/Atom" {
			t.Errorf("%s: feed namespace = %q", e.Lang, doc.XMLName.Space)
		}
		if doc.Title != e.SiteName {
			t.Errorf("%s: feed title = %q, want %q", e.Lang, doc.Title, e.SiteName)
		}
		if want := articleCount(e); len(doc.Entries) != want {
			t.Errorf("%s: feed carries %d entries, want %d", e.Lang, len(doc.Entries), want)
		}

		// Reading order, first entry first.
		first := e.Categories[0].Articles[0]
		if doc.Entries[0].Title != first.Title {
			t.Errorf("%s: first entry = %q, want %q", e.Lang, doc.Entries[0].Title, first.Title)
		}
		if doc.Entries[0].Summary != first.Blurb {
			t.Errorf("%s: first entry summary = %q, want the manifest blurb", e.Lang, doc.Entries[0].Summary)
		}
		if want := srv.URL + e.HomePath + first.Slug; doc.Entries[0].Link.Href != want {
			t.Errorf("%s: first entry links %q, want %q", e.Lang, doc.Entries[0].Link.Href, want)
		}

		// Every URL in the document is fully qualified: an aggregator stores
		// the feed and has nothing to resolve a path against.
		for _, frag := range strings.Split(body, `href="`)[1:] {
			href := frag[:strings.Index(frag, `"`)]
			if !strings.HasPrefix(href, srv.URL+"/") {
				t.Errorf("%s: href %q is not an absolute URL on this origin", e.Lang, href)
			}
		}
		byRel := map[string]string{}
		for _, l := range doc.Links {
			byRel[l.Rel] = l.Href
		}
		if want := srv.URL + e.HomePath + FeedFileName; byRel["self"] != want {
			t.Errorf("%s: rel=self = %q, want %q", e.Lang, byRel["self"], want)
		}
		if want := srv.URL + e.HomePath; byRel["alternate"] != want {
			t.Errorf("%s: rel=alternate = %q, want %q", e.Lang, byRel["alternate"], want)
		}
	}
}

// The book has no honest per-article date, so ONE stamp answers for the feed
// and for every entry of it. The invariant is what keeps the feed from
// inventing an editorial history.
func TestFeedSingleUpdatedStamp(t *testing.T) {
	stamp := time.Date(2026, 8, 20, 9, 15, 0, 0, time.UTC)
	doc := parseFeed(t, string(French.FeedXML("https://example.org", stamp)))
	const want = "2026-08-20T09:15:00Z"
	if doc.Updated != want {
		t.Errorf("feed updated = %q, want %q", doc.Updated, want)
	}
	for i, e := range doc.Entries {
		if e.Updated != want {
			t.Errorf("entry %d updated = %q, want the single stamp %q", i, e.Updated, want)
		}
	}
}

// The identifiers do not move with the host the book was reached under, so an
// aggregator that followed two names still sees one book and one set of
// articles. Only the links follow the origin.
func TestFeedIdentifiersAreHostIndependent(t *testing.T) {
	stamp := time.Now()
	here := parseFeed(t, string(English.FeedXML("http://127.0.0.1:8787", stamp)))
	there := parseFeed(t, string(English.FeedXML("https://example.org", stamp)))
	if here.ID != there.ID || !strings.HasPrefix(here.ID, English.EPUBIdentifier) {
		t.Errorf("feed id moved with the host: %q vs %q", here.ID, there.ID)
	}
	if len(here.Entries) != len(there.Entries) {
		t.Fatal("the two renderings disagree on the article count")
	}
	for i := range here.Entries {
		if here.Entries[i].ID != there.Entries[i].ID {
			t.Fatalf("entry %d id moved with the host: %q vs %q", i, here.Entries[i].ID, there.Entries[i].ID)
		}
	}
	if here.Entries[0].Link.Href == there.Entries[0].Link.Href {
		t.Error("entry links did not follow the origin")
	}
	// The two editions are two publications and never share an identifier.
	if strings.HasPrefix(here.ID, French.EPUBIdentifier) {
		t.Error("the English feed borrowed the French edition's identity")
	}
}

// Discovery: every page of an edition points at its feed, llms.txt lists it,
// and the sitemap does NOT (a feed is not a page to index).
func TestFeedIsDiscoverable(t *testing.T) {
	srv := siteServer(t)
	art := Categories[0].Articles[0]
	want := `<link rel="alternate" type="application/atom+xml" title="` +
		French.SiteName + `" href="` + FeedFileName + `">`
	for _, path := range []string{French.HomePath, French.HomePath + art.Slug} {
		if _, body := get(t, srv, path); !strings.Contains(body, want) {
			t.Errorf("%s does not declare its feed", path)
		}
	}
	if _, llms := get(t, srv, "/llms.txt"); !strings.Contains(llms, French.HomePath+FeedFileName) {
		t.Error("llms.txt does not list the French feed")
	}
	if _, sitemap := get(t, srv, "/sitemap.xml"); strings.Contains(sitemap, FeedFileName) {
		t.Error("the sitemap lists the feed, which is not a page to index")
	}
}
