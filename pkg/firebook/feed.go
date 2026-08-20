package firebook

import (
	"time"

	"github.com/bpineau/pofo/pkg/seo"
)

// FeedFileName is where an edition publishes its Atom feed, relative to its
// mount ("/firebook/fr/feed.xml"). Every page of the edition declares it in
// its head, so a feed reader handed any article URL finds the whole book.
const FeedFileName = "feed.xml"

// FeedXML renders the edition as an Atom 1.0 feed: the book itself as the
// feed, one entry per article in manifest (reading) order, each carrying its
// manifest blurb as its summary. origin is the scheme and host the absolute
// URLs are built on ("https://example.org"), which only a request knows; see
// RequestOrigin.
//
// updated stamps the feed AND every entry with the same time, deliberately.
// Atom requires the element and the book has no honest per-article date to put
// there: the articles are embedded in the binary, and a date invented per entry
// would make a reader believe in an editorial history that does not exist. The
// value a serving handler passes is its single publication stamp, the very one
// the EPUB build writes into dcterms:modified and the OPDS catalog carries, so
// all three formats agree on when this copy of the book was published.
//
// The identifiers are derived from EPUBIdentifier rather than from the URLs,
// so they do not move when the same book is reached under another host name
// (localhost, a tailnet name, the public domain): an aggregator that followed
// two of them still sees one book and one set of articles.
func (e *Edition) FeedXML(origin string, updated time.Time) []byte {
	f := seo.Feed{
		Title:    e.SiteName,
		Subtitle: e.SiteLede,
		ID:       e.EPUBIdentifier + ":feed",
		Self:     e.absolute(origin, FeedFileName),
		Link:     e.absolute(origin, "."),
		Language: e.Lang,
		Author:   "pofo",
		Updated:  updated,
	}
	for _, cat := range e.Categories {
		for _, a := range cat.Articles {
			f.Entries = append(f.Entries, seo.FeedEntry{
				Title:   a.Title,
				Link:    e.absolute(origin, a.Slug),
				ID:      e.EPUBIdentifier + ":" + a.Slug,
				Summary: a.Blurb,
				Updated: updated,
			})
		}
	}
	return f.Atom()
}
