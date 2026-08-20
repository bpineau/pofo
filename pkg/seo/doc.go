// Package seo builds the three small machine-readable files a site publishes
// for crawlers and for AI agents: a sitemaps.org sitemap, a robots.txt, and an
// llms.txt (the llmstxt.org convention).
//
// The package is content-agnostic and stdlib-only: it knows the file formats,
// never what a particular site holds. A caller assembles the values from its
// own catalog and serves the bytes; pkg/firebook does exactly that for the
// FIRE book (see firebook.Site).
//
//	body := seo.Sitemap([]seo.URL{{Loc: "https://example.org/"}})
//
// # Conventions
//
//   - Every Loc must be an ABSOLUTE URL (protocol included): the sitemap
//     protocol requires it, and a crawler reading robots.txt from one host
//     must be told which host the sitemap covers. Build them from the request
//     when the public origin is not known at compile time.
//   - LastMod is optional and omitted when zero. An invented date is worse
//     than no date: a crawler that learns a lastmod lies stops trusting it.
//   - Text output is deterministic for a given value, so a handler can hash it
//     for an ETag and a test can compare it byte for byte.
//   - The output carries no identity of its own (no author, no boilerplate):
//     everything visible comes from the value the caller passes.
//
// # llms.txt
//
// The llmstxt.org format is plain Markdown with a fixed shape: one H1 naming
// the site, one blockquote summarizing it, free-form note lines, then H2
// sections holding link lists ("- [title](url): note"). LLMs.Text renders
// exactly that, so an agent can read the file as an index of the site and
// follow the links it needs instead of scraping HTML.
package seo
