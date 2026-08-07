package opds

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

// Media types and link relations used by an OPDS acquisition catalog.
const (
	// FeedType is the media type of an OPDS acquisition feed. It is both the
	// Content-Type a server should send and the type of the feed's rel="self"
	// link, so a client can recognize the document without fetching it twice.
	FeedType = "application/atom+xml;profile=opds-catalog;kind=acquisition"

	// EPUBType is the media type of an EPUB acquisition link, used as the
	// default when an Entry leaves Type empty.
	EPUBType = "application/epub+zip"

	// AcquisitionRel is the link relation marking a directly downloadable
	// publication (the "acquire this book" link a reader follows).
	AcquisitionRel = "http://opds-spec.org/acquisition"
)

// atomNS is the Atom namespace every OPDS feed lives in. OPDS 1.2 is a profile
// of Atom, so the feed needs no other namespace for the minimal, KOReader-
// oriented catalog this package emits.
const atomNS = "http://www.w3.org/2005/Atom"

// Entry is one publication in an acquisition feed.
//
// Href is the acquisition link and may be relative to the feed's own URL, so
// the same feed works under any mount prefix. Size is advisory: a positive
// value adds the Atom link "length" attribute (octets), zero omits it. Type
// defaults to EPUBType when empty.
type Entry struct {
	Title   string
	Author  string    // optional; omitted when empty
	Summary string    // optional plain-text description; omitted when empty
	ID      string    // stable urn:uuid of the publication
	Updated time.Time // last-modified stamp, rendered in UTC as RFC 3339
	Href    string    // acquisition link, may be relative to the feed URL
	Type    string    // acquisition media type; empty defaults to EPUBType
	Size    int64     // advisory octet length; 0 omits the length attribute
}

// Feed is a complete OPDS acquisition feed: catalog identity plus its entries.
//
// Self is the feed's own href (rel="self"); it may be relative and is omitted
// from the output when empty. Output is deterministic for a fixed set of times,
// so a server can hash it for an ETag.
type Feed struct {
	Title   string
	ID      string    // stable urn:uuid of the catalog
	Updated time.Time // rendered in UTC as RFC 3339
	Self    string    // href of the feed itself; may be relative
	Entries []Entry
}

// XML renders the feed as an OPDS 1.2 acquisition document: an Atom feed with
// a rel="self" catalog link and one acquisition entry per publication. All text
// and attribute values are XML-escaped; the output parses with encoding/xml and
// is byte-for-byte identical across calls for the same Feed.
func (f *Feed) XML() []byte {
	var b strings.Builder
	b.WriteString(xml.Header) // <?xml ...?> plus a trailing newline
	b.WriteString(`<feed xmlns="` + atomNS + `">` + "\n")

	elem(&b, "  ", "id", f.ID)
	elem(&b, "  ", "title", f.Title)
	elem(&b, "  ", "updated", stamp(f.Updated))
	if f.Self != "" {
		b.WriteString(`  <link rel="self" href="` + attr(f.Self) +
			`" type="` + attr(FeedType) + `"/>` + "\n")
	}
	for _, e := range f.Entries {
		e.write(&b)
	}

	b.WriteString("</feed>\n")
	return []byte(b.String())
}

// write appends one <entry> to b.
func (e Entry) write(b *strings.Builder) {
	b.WriteString("  <entry>\n")
	elem(b, "    ", "title", e.Title)
	elem(b, "    ", "id", e.ID)
	elem(b, "    ", "updated", stamp(e.Updated))
	if e.Author != "" {
		b.WriteString(`    <author><name>` + text(e.Author) + `</name></author>` + "\n")
	}
	if e.Summary != "" {
		b.WriteString(`    <summary type="text">` + text(e.Summary) + `</summary>` + "\n")
	}

	typ := e.Type
	if typ == "" {
		typ = EPUBType
	}
	b.WriteString(`    <link rel="` + attr(AcquisitionRel) +
		`" type="` + attr(typ) + `" href="` + attr(e.Href) + `"`)
	if e.Size > 0 {
		b.WriteString(` length="` + strconv.FormatInt(e.Size, 10) + `"`)
	}
	b.WriteString("/>\n")
	b.WriteString("  </entry>\n")
}

// stamp formats a time as UTC RFC 3339, the Atom date shape.
func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// elem writes an indented "<name>escaped-value</name>" line.
func elem(b *strings.Builder, indent, name, value string) {
	b.WriteString(indent + "<" + name + ">" + text(value) + "</" + name + ">\n")
}

// text escapes a string for use as XML character data. attr is an alias: the
// same escaping (which also encodes quotes) is safe inside attribute values.
func text(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func attr(s string) string { return text(s) }
