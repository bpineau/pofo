package seo_test

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/seo"
)

func ExampleSitemap() {
	body := seo.Sitemap([]seo.URL{
		{Loc: "https://example.org/"},
		{Loc: "https://example.org/handbook/", LastMod: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)},
	})
	fmt.Println(string(body))
	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	//   <url>
	//     <loc>https://example.org/</loc>
	//   </url>
	//   <url>
	//     <loc>https://example.org/handbook/</loc>
	//     <lastmod>2026-08-20</lastmod>
	//   </url>
	// </urlset>
}

func ExampleFeed_Atom() {
	day := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	f := seo.Feed{
		Title:    "Handbook",
		Subtitle: "One line on what it covers.",
		Self:     "https://example.org/handbook/feed.xml",
		Link:     "https://example.org/handbook/",
		Language: "en",
		Author:   "example",
		Updated:  day,
		Entries: []seo.FeedEntry{{
			Title:   "First chapter",
			Link:    "https://example.org/handbook/first",
			Summary: "Where it starts.",
		}},
	}
	fmt.Println(string(f.Atom()))
	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <feed xmlns="http://www.w3.org/2005/Atom" xml:lang="en">
	//   <id>https://example.org/handbook/feed.xml</id>
	//   <title>Handbook</title>
	//   <subtitle>One line on what it covers.</subtitle>
	//   <updated>2026-08-20T00:00:00Z</updated>
	//   <author><name>example</name></author>
	//   <link rel="self" type="application/atom+xml; charset=utf-8" href="https://example.org/handbook/feed.xml"/>
	//   <link rel="alternate" type="text/html" href="https://example.org/handbook/"/>
	//   <entry>
	//     <id>https://example.org/handbook/first</id>
	//     <title>First chapter</title>
	//     <updated>2026-08-20T00:00:00Z</updated>
	//     <link rel="alternate" type="text/html" href="https://example.org/handbook/first"/>
	//     <summary type="text">Where it starts.</summary>
	//   </entry>
	// </feed>
}

func ExampleRobots_Text() {
	r := seo.Robots{
		Preamble: []string{"Everything here is meant to be read and quoted."},
		Groups: []seo.Group{
			{},
			{Comment: "AI crawlers, named explicitly.", Agents: []string{"GPTBot", "ClaudeBot"}},
		},
		Sitemaps: []string{"https://example.org/sitemap.xml"},
	}
	fmt.Println(string(r.Text()))
	// Output:
	// # Everything here is meant to be read and quoted.
	//
	// User-agent: *
	// Allow: /
	//
	// # AI crawlers, named explicitly.
	// User-agent: GPTBot
	// User-agent: ClaudeBot
	// Allow: /
	//
	// Sitemap: https://example.org/sitemap.xml
}

func ExampleLLMs_Text() {
	l := seo.LLMs{
		Title:   "Example",
		Summary: "One paragraph on what this site holds.",
		Notes:   []string{"Every page is also available as Markdown."},
		Sections: []seo.Section{{
			Title: "Handbook",
			Links: []seo.Link{{Title: "Contents", URL: "https://example.org/handbook/", Note: "the table of contents"}},
		}},
	}
	fmt.Println(string(l.Text()))
	// Output:
	// # Example
	//
	// > One paragraph on what this site holds.
	//
	// Every page is also available as Markdown.
	//
	// ## Handbook
	//
	// - [Contents](https://example.org/handbook/): the table of contents
}
