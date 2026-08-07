package opds_test

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/opds"
)

// A one-book acquisition feed, the shape a server hands to an e-book reader.
// The publication link is relative, so the same feed works under any mount.
func ExampleFeed_XML() {
	feed := &opds.Feed{
		Title:   "Ma bibliotheque",
		ID:      "urn:uuid:00000000-0000-0000-0000-000000000000:catalog",
		Updated: time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		Self:    "opds.xml",
		Entries: []opds.Entry{{
			Title:   "Petit guide",
			Author:  "Anne Auteur",
			Summary: "Un court exemple.",
			ID:      "urn:uuid:00000000-0000-0000-0000-000000000000",
			Updated: time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
			Href:    "petit-guide.epub",
			Size:    2048,
		}},
	}

	fmt.Print(string(feed.XML()))
	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <feed xmlns="http://www.w3.org/2005/Atom">
	//   <id>urn:uuid:00000000-0000-0000-0000-000000000000:catalog</id>
	//   <title>Ma bibliotheque</title>
	//   <updated>2026-07-24T00:00:00Z</updated>
	//   <link rel="self" href="opds.xml" type="application/atom+xml;profile=opds-catalog;kind=acquisition"/>
	//   <entry>
	//     <title>Petit guide</title>
	//     <id>urn:uuid:00000000-0000-0000-0000-000000000000</id>
	//     <updated>2026-07-24T00:00:00Z</updated>
	//     <author><name>Anne Auteur</name></author>
	//     <summary type="text">Un court exemple.</summary>
	//     <link rel="http://opds-spec.org/acquisition" type="application/epub+zip" href="petit-guide.epub" length="2048"/>
	//   </entry>
	// </feed>
}
