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
