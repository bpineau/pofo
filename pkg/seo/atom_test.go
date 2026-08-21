package seo

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

// atomDoc is the parsed shape the tests read a rendered feed back through, so
// they assert on a document a client could actually consume rather than on a
// string.
type atomDoc struct {
	XMLName  xml.Name `xml:"feed"`
	ID       string   `xml:"id"`
	Title    string   `xml:"title"`
	Subtitle string   `xml:"subtitle"`
	Updated  string   `xml:"updated"`
	Author   struct {
		Name string `xml:"name"`
	} `xml:"author"`
	Links []struct {
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
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

func parseAtom(t *testing.T, body []byte) atomDoc {
	t.Helper()
	var doc atomDoc
	if err := xml.Unmarshal(body, &doc); err != nil {
		t.Fatalf("feed is not well-formed XML: %v\n%s", err, body)
	}
	if doc.XMLName.Space != atomNS {
		t.Errorf("feed namespace = %q, want %q", doc.XMLName.Space, atomNS)
	}
	return doc
}

func testFeed() Feed {
	day := time.Date(2026, 8, 20, 12, 30, 0, 0, time.UTC)
	return Feed{
		Title:    "A book",
		Subtitle: "What it covers",
		ID:       "urn:uuid:0d3e428a-fdd6-454e-a82a-d8c70ce563d0:feed",
		Self:     "https://example.org/book/feed.xml",
		Link:     "https://example.org/book/",
		Language: "fr",
		Author:   "pofo",
		Updated:  day,
		Entries: []FeedEntry{
			{Title: "One", Link: "https://example.org/book/one", ID: "urn:one", Summary: "first", Updated: day},
			{Title: "Two", Link: "https://example.org/book/two", ID: "urn:two", Summary: "second", Updated: day},
		},
	}
}

func TestAtomRequiredElements(t *testing.T) {
	f := testFeed()
	doc := parseAtom(t, f.Atom())

	if doc.ID != f.ID || doc.Title != f.Title || doc.Subtitle != f.Subtitle {
		t.Errorf("feed identity = %q/%q/%q", doc.ID, doc.Title, doc.Subtitle)
	}
	if doc.Author.Name != "pofo" {
		t.Errorf("feed author = %q", doc.Author.Name)
	}
	if doc.Updated != "2026-08-20T12:30:00Z" {
		t.Errorf("feed updated = %q, want RFC 3339 UTC", doc.Updated)
	}
	// Atom requires id, title and updated on every entry.
	if len(doc.Entries) != len(f.Entries) {
		t.Fatalf("feed carries %d entries, want %d", len(doc.Entries), len(f.Entries))
	}
	for i, e := range doc.Entries {
		if e.ID == "" || e.Title == "" || e.Updated == "" {
			t.Errorf("entry %d misses a required element: %+v", i, e)
		}
		if e.Link.Href != f.Entries[i].Link || e.Link.Rel != "alternate" {
			t.Errorf("entry %d link = %+v", i, e.Link)
		}
		if e.Summary != f.Entries[i].Summary {
			t.Errorf("entry %d summary = %q", i, e.Summary)
		}
	}
	// Entries keep the order they were given: a feed is a reading order here,
	// not a reverse-chronological log.
	if doc.Entries[0].Title != "One" || doc.Entries[1].Title != "Two" {
		t.Errorf("entries were reordered: %q then %q", doc.Entries[0].Title, doc.Entries[1].Title)
	}

	byRel := map[string]string{}
	for _, l := range doc.Links {
		byRel[l.Rel] = l.Href
	}
	if byRel["self"] != f.Self || byRel["alternate"] != f.Link {
		t.Errorf("feed links = %v", byRel)
	}
	if !strings.Contains(string(f.Atom()), `xml:lang="fr"`) {
		t.Error("feed does not declare its language")
	}
}

// A feed is stored by whoever fetched it and resolved against nothing, so every
// URL in it has to be absolute.
func TestAtomURLsAreAbsolute(t *testing.T) {
	body := string(testFeed().Atom())
	for _, frag := range strings.Split(body, `href="`)[1:] {
		href := frag[:strings.Index(frag, `"`)]
		if !strings.HasPrefix(href, "https://") {
			t.Errorf("href %q is not absolute", href)
		}
	}
}

// The zero values fall back to the feed's own: an entry that names no id is
// identified by its link, and one that names no time takes the feed's stamp.
func TestAtomEntryDefaults(t *testing.T) {
	f := Feed{
		Self:    "https://example.org/feed.xml",
		Updated: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Entries: []FeedEntry{
			{Title: "One", Link: "https://example.org/one"},
			{Title: "no link, dropped"},
		},
	}
	doc := parseAtom(t, f.Atom())
	if doc.ID != f.Self {
		t.Errorf("feed id = %q, want the self URL", doc.ID)
	}
	if len(doc.Entries) != 1 {
		t.Fatalf("feed carries %d entries, want the linkless one dropped", len(doc.Entries))
	}
	if doc.Entries[0].ID != "https://example.org/one" {
		t.Errorf("entry id = %q, want its link", doc.Entries[0].ID)
	}
	if doc.Entries[0].Updated != "2026-01-02T03:04:05Z" {
		t.Errorf("entry updated = %q, want the feed's stamp", doc.Entries[0].Updated)
	}
}

// Text that means something in XML is escaped, in element content and in
// attributes alike.
func TestAtomEscaping(t *testing.T) {
	f := Feed{
		Title:   `Bonds & "gilts" <ahem>`,
		Self:    "https://example.org/feed.xml?a=1&b=2",
		Updated: time.Now(),
		Entries: []FeedEntry{{Title: "x", Link: "https://example.org/a?q=1&r=2", Summary: "5 < 6"}},
	}
	body := f.Atom()
	doc := parseAtom(t, body)
	if doc.Title != f.Title {
		t.Errorf("title round-trip = %q, want %q", doc.Title, f.Title)
	}
	if doc.Entries[0].Link.Href != f.Entries[0].Link {
		t.Errorf("href round-trip = %q", doc.Entries[0].Link.Href)
	}
	if strings.Contains(string(body), "&b=2") {
		t.Error("an ampersand escaped unescaped into the document")
	}
}

// Deterministic output: a handler hashes it for an ETag.
func TestAtomDeterministic(t *testing.T) {
	first := string(testFeed().Atom())
	second := string(testFeed().Atom())
	if first != second {
		t.Error("two renderings of the same feed differ")
	}
}
