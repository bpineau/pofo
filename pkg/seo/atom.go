package seo

import (
	"encoding/xml"
	"strings"
	"time"
)

// AtomType is the Content-Type of an Atom syndication feed, as it should be
// served and as the type attribute of the feed's own rel="self" link.
const AtomType = "application/atom+xml; charset=utf-8"

// atomNS is the Atom 1.0 namespace. A syndication feed needs no other.
const atomNS = "http://www.w3.org/2005/Atom"

// FeedEntry is one item of an Atom feed: a page, its summary, and when the
// document it belongs to was last stamped.
//
// Link is the entry's page and must be an ABSOLUTE URL, like every URL in a
// feed: an aggregator stores the document and resolves nothing against the
// address it was fetched from. ID defaults to Link, which is the usual choice
// for a page that has one stable address.
type FeedEntry struct {
	Title   string
	Link    string    // absolute URL of the page
	ID      string    // stable identifier; empty means "the Link"
	Summary string    // optional plain-text summary; omitted when empty
	Updated time.Time // rendered in UTC as RFC 3339
}

// Feed is a complete Atom 1.0 syndication feed.
//
// Self is the feed's own absolute URL (rel="self") and, unless ID says
// otherwise, the feed's identifier; Link is the page the feed describes
// (rel="alternate"). Output is deterministic for a fixed set of times, so a
// server can hash it for an ETag and a test can compare it byte for byte.
type Feed struct {
	Title    string
	Subtitle string    // optional one-line description; omitted when empty
	ID       string    // stable identifier; empty means "the Self URL"
	Self     string    // absolute URL of the feed itself
	Link     string    // absolute URL of the page the feed covers
	Language string    // xml:lang of the feed and its entries ("fr", "en")
	Author   string    // feed-level author name; omitted when empty
	Updated  time.Time // rendered in UTC as RFC 3339
	Entries  []FeedEntry
}

// Atom renders the feed as an Atom 1.0 document. Every text and attribute
// value is XML-escaped, an entry with an empty Link is skipped (so a caller
// can build its list without guarding every branch), and an entry that names
// no ID or Updated inherits the feed's Link-derived id and its Updated stamp.
func (f Feed) Atom() []byte {
	var b strings.Builder
	b.WriteString(xml.Header) // <?xml ...?> plus a trailing newline
	b.WriteString(`<feed xmlns="` + atomNS + `"`)
	if f.Language != "" {
		b.WriteString(` xml:lang="` + esc(f.Language) + `"`)
	}
	b.WriteString(">\n")

	id := f.ID
	if id == "" {
		id = f.Self
	}
	tag(&b, "  ", "id", id)
	tag(&b, "  ", "title", f.Title)
	if f.Subtitle != "" {
		tag(&b, "  ", "subtitle", f.Subtitle)
	}
	tag(&b, "  ", "updated", atomStamp(f.Updated))
	if f.Author != "" {
		b.WriteString(`  <author><name>` + esc(f.Author) + `</name></author>` + "\n")
	}
	if f.Self != "" {
		b.WriteString(`  <link rel="self" type="` + esc(AtomType) +
			`" href="` + esc(f.Self) + `"/>` + "\n")
	}
	if f.Link != "" {
		b.WriteString(`  <link rel="alternate" type="text/html" href="` + esc(f.Link) + `"/>` + "\n")
	}
	for _, e := range f.Entries {
		if e.Link == "" {
			continue
		}
		f.writeEntry(&b, e)
	}

	b.WriteString("</feed>\n")
	return []byte(b.String())
}

// writeEntry appends one <entry> to b, filling in what the entry leaves empty
// from the feed.
func (f Feed) writeEntry(b *strings.Builder, e FeedEntry) {
	b.WriteString("  <entry>\n")
	id := e.ID
	if id == "" {
		id = e.Link
	}
	tag(b, "    ", "id", id)
	tag(b, "    ", "title", e.Title)
	stamp := e.Updated
	if stamp.IsZero() {
		stamp = f.Updated
	}
	tag(b, "    ", "updated", atomStamp(stamp))
	b.WriteString(`    <link rel="alternate" type="text/html" href="` + esc(e.Link) + `"/>` + "\n")
	if e.Summary != "" {
		b.WriteString(`    <summary type="text">` + esc(e.Summary) + `</summary>` + "\n")
	}
	b.WriteString("  </entry>\n")
}

// atomStamp formats a time as UTC RFC 3339, the Atom date shape.
func atomStamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// tag writes an indented "<name>escaped-value</name>" line.
func tag(b *strings.Builder, indent, name, value string) {
	b.WriteString(indent + "<" + name + ">" + esc(value) + "</" + name + ">\n")
}

// esc escapes a string for use as XML character data. The same escaping also
// encodes quotes, so it is safe inside attribute values too.
func esc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
