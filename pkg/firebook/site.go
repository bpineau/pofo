package firebook

import (
	"net/http"
	"strings"

	"github.com/bpineau/pofo/pkg/seo"
)

// Mount is one edition published at a base path. Base is absolute and ends
// with a slash ("/firebook/fr/"), matching what WithAlternate takes: the
// handler emits relative URLs and only the server knows where each edition
// sits.
type Mount struct {
	Base    string
	Edition *Edition
}

// Page is one non-book page of the surrounding site, listed in the sitemap
// and in llms.txt so a crawler or an agent sees the whole surface, not just
// the book.
type Page struct {
	Path  string // absolute path, "/visualizer"
	Title string
	Note  string // one line, what the page is
}

// Site is the machine-readable face of a server publishing the book: it
// renders the three root files crawlers and AI agents look for (sitemap.xml,
// robots.txt, llms.txt) from the mounted editions plus the server's own pages.
//
// The public origin is not known at build time (the same binary runs on
// localhost, behind a tailnet name and behind a reverse proxy), so the
// absolute URLs the sitemap protocol requires are built per request from the
// Host header; Handle does that for a caller.
type Site struct {
	Name    string // llms.txt H1: the site, not the book ("pofo")
	Summary string // llms.txt blockquote: one paragraph on what the site is
	Mounts  []Mount
	Pages   []Page
}

// Editions are the published editions of the book, source edition first.
var Editions = []*Edition{French, English}

// BookSite is the Site a pofo server publishes: both editions mounted at their
// HomePath, plus the caller's own pages (the surfaces around the book, which
// differ from one server to the next). A server that mounts the editions
// elsewhere builds its own Site instead.
func BookSite(pages ...Page) Site {
	s := Site{
		Name: "pofo",
		Summary: "pofo publishes a handbook on living off your capital without outliving it " +
			"(safe withdrawal rates, sequence risk, withdrawal strategies, resilient " +
			"portfolios, inflation, taxes and the human factor) in two editions, French " +
			"and English, as web pages and as EPUB files, alongside the open-source " +
			"portfolio and retirement tools the book's figures are computed with.",
		Pages: pages,
	}
	for _, e := range Editions {
		s.Mounts = append(s.Mounts, Mount{Base: e.HomePath, Edition: e})
	}
	return s
}

// aiAgents are the crawlers whose traffic this site wants. They are named
// explicitly, next to the catch-all record, because several of them read
// robots.txt looking for their OWN name and treat its absence as ambiguous;
// and because saying so out loud is the point: the book is published to be
// read, quoted and cited, and it is not behind a paywall.
var aiAgents = []string{
	"GPTBot", "OAI-SearchBot", "ChatGPT-User",
	"ClaudeBot", "Claude-Web", "PerplexityBot",
	"Google-Extended", "CCBot",
}

// SitemapXML renders the sitemap for every mounted edition: its index, one URL
// per article, and its EPUB, plus the site's own pages. origin is the scheme
// and host the URLs are built on ("https://example.org").
//
// No <lastmod> is emitted anywhere: the articles are embedded in the binary
// with no meaningful modification time, and a date invented at build time
// would be a lie a crawler eventually learns to ignore. The markdown mirrors
// (<article>.md) are left out too: they are the same document in another
// encoding, and a sitemap is for pages to index, not for duplicates.
func (s Site) SitemapXML(origin string) []byte {
	var urls []seo.URL
	for _, p := range s.Pages {
		urls = append(urls, seo.URL{Loc: origin + p.Path})
	}
	for _, m := range s.Mounts {
		base := origin + m.Base
		urls = append(urls, seo.URL{Loc: base})
		for _, cat := range m.Edition.Categories {
			for _, a := range cat.Articles {
				urls = append(urls, seo.URL{Loc: base + a.Slug})
			}
		}
		urls = append(urls, seo.URL{Loc: base + m.Edition.EPUBFileName})
	}
	return seo.Sitemap(urls)
}

// RobotsTXT renders the robots.txt: everything is allowed, the AI crawlers are
// named explicitly, and the sitemap and llms.txt are pointed at.
func (s Site) RobotsTXT(origin string) []byte {
	r := seo.Robots{
		Preamble: []string{
			"Everything published here is meant to be indexed, quoted and cited.\n" +
				"There is no paywall and no login: cite the page, not this file.\n" +
				"Every article is also served as clean Markdown at the same URL\n" +
				"with a \".md\" suffix, and /llms.txt indexes the whole site.",
		},
		Groups: []seo.Group{
			{},
			{Comment: "AI crawlers and answer engines, welcome by name.", Agents: aiAgents},
		},
		Sitemaps: []string{origin + "/sitemap.xml"},
	}
	return r.Text()
}

// LLMsTXT renders the llms.txt (llmstxt.org): the site in one paragraph, then
// one section per mounted edition (index, EPUB, OPDS catalog, Atom feed)
// followed by one section per part of that edition listing every article as its
// MARKDOWN URL.
//
// Both editions live in ONE file, sections labelled by edition. An agent
// fetches one llms.txt per host; splitting the languages into two files would
// hide each edition from whoever found the other, and the two are the same
// book, not two sites.
func (s Site) LLMsTXT(origin string) []byte {
	l := seo.LLMs{
		Title:   s.Name,
		Summary: s.Summary,
		Notes: []string{
			"Every article below is linked as Markdown: the URL is the HTML page's " +
				"with a \".md\" suffix, and it serves the article's source (headings, " +
				"tables, callout blocks and [[wiki-links]] to sibling slugs), preceded " +
				"by one HTML-comment line naming the page it backs. Drop the suffix for " +
				"the HTML page, which is the URL to cite.",
			"The two editions are one book. French is the source edition and every " +
				"correction lands there first; the English edition is a translation of " +
				"it, article by article, except for its country-specific tax part.",
		},
	}
	for _, m := range s.Mounts {
		e, base := m.Edition, origin+m.Base
		l.Sections = append(l.Sections, seo.Section{
			Title: e.SiteName + " (" + e.Lang + ")",
			Links: []seo.Link{
				{Title: e.UI.IndexLink, URL: base, Note: e.SiteLede},
				{Title: e.UI.EPUBLink, URL: base + e.EPUBFileName, Note: "the whole edition as one EPUB 3 file"},
				{Title: "OPDS", URL: base + "opds.xml", Note: "acquisition catalog for e-readers"},
				{Title: "Atom", URL: base + FeedFileName, Note: "every article of this edition as a feed"},
			},
		})
		for _, cat := range e.Categories {
			if len(cat.Articles) == 0 {
				continue
			}
			sec := seo.Section{Title: e.SiteName + ": " + cat.Title}
			for _, a := range cat.Articles {
				sec.Links = append(sec.Links, seo.Link{Title: a.Title, URL: base + a.Slug + ".md", Note: a.Blurb})
			}
			l.Sections = append(l.Sections, sec)
		}
	}
	if len(s.Pages) > 0 {
		// The llmstxt.org convention reserves "Optional" for what an agent
		// short on context may skip. The tools around the book are exactly
		// that: useful to a reader, secondary to the text.
		sec := seo.Section{Title: "Optional"}
		for _, p := range s.Pages {
			sec.Links = append(sec.Links, seo.Link{Title: p.Title, URL: origin + p.Path, Note: p.Note})
		}
		l.Sections = append(l.Sections, sec)
	}
	return l.Text()
}

// Handle registers /sitemap.xml, /robots.txt and /llms.txt on mux, each built
// from the request's own origin so the absolute URLs are the ones the visitor
// actually reached.
func (s Site) Handle(mux *http.ServeMux) {
	serve := func(contentType string, render func(origin string) []byte) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				http.Error(w, "GET only", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=3600")
			_, _ = w.Write(render(RequestOrigin(r)))
		}
	}
	mux.HandleFunc("/sitemap.xml", serve(seo.SitemapType, s.SitemapXML))
	mux.HandleFunc("/robots.txt", serve(seo.TextType, s.RobotsTXT))
	mux.HandleFunc("/llms.txt", serve(seo.TextType, s.LLMsTXT))
}

// RequestOrigin reconstructs the scheme and host a request was made to
// ("https://example.org"), which absolute URLs in a sitemap or an llms.txt
// need and which the binary cannot know at build time.
//
// The host is the request's own (net/http validates it); X-Forwarded-Host is
// deliberately NOT read, since an untrusted client could then mint a sitemap
// pointing at another site. Only the scheme is taken from a proxy header,
// because TLS termination is the one thing the process genuinely cannot see,
// and the worst a forged value can do is flip http and https on the same host.
func RequestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))) {
	case "https":
		scheme = "https"
	case "http":
		scheme = "http"
	}
	return scheme + "://" + r.Host
}
