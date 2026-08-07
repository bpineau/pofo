package opds_test

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/opds"
)

// fixedTime is a stable Updated stamp, so the tests can assert exact output.
var fixedTime = time.Date(2026, 7, 24, 9, 30, 0, 0, time.UTC)

// sampleFeed is a one-entry acquisition feed, the shape firebook serves.
func sampleFeed() *opds.Feed {
	return &opds.Feed{
		Title:   "Le FIRE tranquille",
		ID:      "urn:uuid:6f8a2c1e-4d3b-4a7e-9c21-8f5b0e6d4a92:catalog",
		Updated: fixedTime,
		Self:    "opds.xml",
		Entries: []opds.Entry{{
			Title:   "Le FIRE tranquille",
			Author:  "pofo",
			Summary: "Vivre de son capital sans le survivre.",
			ID:      "urn:uuid:6f8a2c1e-4d3b-4a7e-9c21-8f5b0e6d4a92",
			Updated: fixedTime,
			Href:    "le-fire-tranquille.epub",
			Type:    "application/epub+zip",
			Size:    123456,
		}},
	}
}

// parsedFeed mirrors the subset of the Atom feed the tests inspect.
type parsedFeed struct {
	XMLName xml.Name      `xml:"feed"`
	ID      string        `xml:"id"`
	Title   string        `xml:"title"`
	Updated string        `xml:"updated"`
	Links   []parsedLink  `xml:"link"`
	Entries []parsedEntry `xml:"entry"`
}

type parsedLink struct {
	Rel    string `xml:"rel,attr"`
	Type   string `xml:"type,attr"`
	Href   string `xml:"href,attr"`
	Length string `xml:"length,attr"`
}

type parsedEntry struct {
	ID      string       `xml:"id"`
	Title   string       `xml:"title"`
	Updated string       `xml:"updated"`
	Author  string       `xml:"author>name"`
	Summary string       `xml:"summary"`
	Links   []parsedLink `xml:"link"`
}

func mustParse(t *testing.T, b []byte) parsedFeed {
	t.Helper()
	var f parsedFeed
	if err := xml.Unmarshal(b, &f); err != nil {
		t.Fatalf("feed does not parse as XML: %v\n%s", err, b)
	}
	return f
}

func TestXMLParsesAndCarriesIdentity(t *testing.T) {
	f := mustParse(t, sampleFeed().XML())
	if f.Title != "Le FIRE tranquille" {
		t.Errorf("feed title = %q", f.Title)
	}
	if f.ID != "urn:uuid:6f8a2c1e-4d3b-4a7e-9c21-8f5b0e6d4a92:catalog" {
		t.Errorf("feed id = %q", f.ID)
	}
	if f.Updated != "2026-07-24T09:30:00Z" {
		t.Errorf("feed updated = %q", f.Updated)
	}
	if len(f.Entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(f.Entries))
	}
	e := f.Entries[0]
	if e.ID != "urn:uuid:6f8a2c1e-4d3b-4a7e-9c21-8f5b0e6d4a92" {
		t.Errorf("entry id = %q", e.ID)
	}
	if e.Author != "pofo" {
		t.Errorf("entry author = %q", e.Author)
	}
	if e.Summary != "Vivre de son capital sans le survivre." {
		t.Errorf("entry summary = %q", e.Summary)
	}
	if e.Updated != "2026-07-24T09:30:00Z" {
		t.Errorf("entry updated = %q", e.Updated)
	}
}

func TestSelfLinkIsAcquisitionCatalog(t *testing.T) {
	f := mustParse(t, sampleFeed().XML())
	var self *parsedLink
	for i := range f.Links {
		if f.Links[i].Rel == "self" {
			self = &f.Links[i]
		}
	}
	if self == nil {
		t.Fatal("no rel=self link on feed")
	}
	if self.Href != "opds.xml" {
		t.Errorf("self href = %q", self.Href)
	}
	if self.Type != opds.FeedType {
		t.Errorf("self type = %q, want %q", self.Type, opds.FeedType)
	}
}

func TestAcquisitionLinkShape(t *testing.T) {
	f := mustParse(t, sampleFeed().XML())
	links := f.Entries[0].Links
	if len(links) != 1 {
		t.Fatalf("want 1 entry link, got %d", len(links))
	}
	l := links[0]
	if l.Rel != opds.AcquisitionRel {
		t.Errorf("rel = %q, want %q", l.Rel, opds.AcquisitionRel)
	}
	if l.Type != "application/epub+zip" {
		t.Errorf("type = %q", l.Type)
	}
	if l.Href != "le-fire-tranquille.epub" {
		t.Errorf("href = %q", l.Href)
	}
	if l.Length != "123456" {
		t.Errorf("length = %q, want 123456", l.Length)
	}
}

func TestSizeZeroOmitsLength(t *testing.T) {
	f := sampleFeed()
	f.Entries[0].Size = 0
	out := f.XML()
	if bytes.Contains(out, []byte("length=")) {
		t.Errorf("length attribute present for Size 0:\n%s", out)
	}
	if l := mustParse(t, out).Entries[0].Links[0].Length; l != "" {
		t.Errorf("parsed length = %q, want empty", l)
	}
}

func TestTypeDefaultsToEPUB(t *testing.T) {
	f := sampleFeed()
	f.Entries[0].Type = ""
	l := mustParse(t, f.XML()).Entries[0].Links[0]
	if l.Type != "application/epub+zip" {
		t.Errorf("default entry type = %q, want application/epub+zip", l.Type)
	}
}

func TestEscaping(t *testing.T) {
	f := sampleFeed()
	f.Title = `Cap & <Gains> "quoted" 'apos'`
	f.Entries[0].Summary = "a < b & c > d"
	f.Entries[0].Href = "book.epub?a=1&b=2"
	out := f.XML()
	// Raw markup must not leak the special characters unescaped.
	if strings.Contains(string(out), "<Gains>") {
		t.Errorf("unescaped title markup leaked:\n%s", out)
	}
	// Round-trips back to the original strings.
	p := mustParse(t, out)
	if p.Title != `Cap & <Gains> "quoted" 'apos'` {
		t.Errorf("title round-trip = %q", p.Title)
	}
	if p.Entries[0].Summary != "a < b & c > d" {
		t.Errorf("summary round-trip = %q", p.Entries[0].Summary)
	}
	if p.Entries[0].Links[0].Href != "book.epub?a=1&b=2" {
		t.Errorf("href round-trip = %q", p.Entries[0].Links[0].Href)
	}
}

func TestDeterministic(t *testing.T) {
	a := sampleFeed().XML()
	b := sampleFeed().XML()
	if !bytes.Equal(a, b) {
		t.Errorf("XML output not deterministic:\n%s\n---\n%s", a, b)
	}
}

func TestOmittedOptionalEntryFields(t *testing.T) {
	f := sampleFeed()
	f.Entries[0].Author = ""
	f.Entries[0].Summary = ""
	out := f.XML()
	if bytes.Contains(out, []byte("<author>")) {
		t.Errorf("empty author still rendered:\n%s", out)
	}
	if bytes.Contains(out, []byte("<summary")) {
		t.Errorf("empty summary still rendered:\n%s", out)
	}
	// Still valid and still carries the acquisition link.
	if l := mustParse(t, out).Entries[0].Links; len(l) != 1 {
		t.Fatalf("want acquisition link, got %d links", len(l))
	}
}
