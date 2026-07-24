// Package opds is a minimal, dependency-free builder for OPDS 1.2 acquisition
// feeds: catalog metadata and a list of downloadable publications in, a valid
// Atom document out. It knows nothing about any specific book; a caller (for
// example pofo's firebook) fills a Feed from its own content and serves the
// bytes with the OPDS content type.
//
// OPDS (Open Publication Distribution System) is a profile of Atom that e-book
// readers browse to add, download and refresh books over HTTP. KOReader is the
// reference client here: the feed stays deliberately small (id, title, updated,
// a rel="self" catalog link, and one entry per publication carrying an
// acquisition link), which is all such a reader needs.
//
// # Usage
//
//	feed := &opds.Feed{
//		Title:   "My library",
//		ID:      "urn:uuid:...:catalog",
//		Updated: buildTime,
//		Self:    "opds.xml", // relative: works under any mount
//		Entries: []opds.Entry{{
//			Title:   "My book",
//			Author:  "An Author",
//			Summary: "A short description.",
//			ID:      "urn:uuid:...",
//			Updated: buildTime,
//			Href:    "my-book.epub", // relative to the feed URL
//			Size:    len(epubBytes),
//		}},
//	}
//	w.Header().Set("Content-Type", opds.FeedType)
//	w.Write(feed.XML())
//
// # Relative links
//
// Feed.Self and Entry.Href may be relative to the feed's own URL, so one feed
// works under any mount prefix. A reader resolves them against the address it
// fetched the feed from; a book re-downloaded through the same relative link
// overwrites the same file, which preserves any reader-side annotation sidecar.
//
// # Determinism
//
// Feed.XML escapes every text and attribute value and emits fields in a fixed
// order, so the output parses with encoding/xml and is byte-for-byte identical
// across calls for the same Feed. A server can hash it for an HTTP ETag.
package opds
