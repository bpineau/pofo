package seo

import (
	"encoding/xml"
	"strings"
	"time"
)

// Content types of the three files, as they should be served.
const (
	// SitemapType is the Content-Type of a sitemap.
	SitemapType = "application/xml; charset=utf-8"
	// TextType is the Content-Type of robots.txt and llms.txt. llms.txt is
	// Markdown, but text/plain is what makes it readable in a browser rather
	// than a download, and agents fetch it by URL, not by type.
	TextType = "text/plain; charset=utf-8"
)

// URL is one entry of a sitemap: an absolute location, optionally dated.
type URL struct {
	Loc     string    // absolute URL, protocol included
	LastMod time.Time // zero value: no <lastmod> is emitted
}

// sitemapURL is the wire shape of one entry. LastMod is a pointer-free
// omitempty string so a zero date simply disappears.
type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	NS      string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// Sitemap renders a sitemaps.org 0.9 urlset. Entries with an empty Loc are
// skipped, so a caller can build its list without guarding every branch.
func Sitemap(urls []URL) []byte {
	set := urlSet{NS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, u := range urls {
		if u.Loc == "" {
			continue
		}
		e := sitemapURL{Loc: u.Loc}
		if !u.LastMod.IsZero() {
			e.LastMod = u.LastMod.UTC().Format("2006-01-02")
		}
		set.URLs = append(set.URLs, e)
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil { // a tree of strings never fails to marshal
		return nil
	}
	return []byte(xml.Header + string(body) + "\n")
}

// Group is one robots.txt record: the user agents it addresses and the rules
// they get. An empty Allow and Disallow means "Allow: /", the explicit way to
// say that nothing is off limits.
type Group struct {
	Comment  string   // one comment line above the record, "#" added
	Agents   []string // User-agent values; empty means "*"
	Allow    []string
	Disallow []string
}

// Robots is a robots.txt: a few comment lines, the records, and the sitemaps.
type Robots struct {
	Preamble []string // comment lines at the top of the file, "#" added
	Groups   []Group
	Sitemaps []string // absolute sitemap URLs
}

// Text renders the robots.txt.
func (r Robots) Text() []byte {
	var b strings.Builder
	for _, line := range r.Preamble {
		writeComment(&b, line)
	}
	for _, g := range r.Groups {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		writeComment(&b, g.Comment)
		agents := g.Agents
		if len(agents) == 0 {
			agents = []string{"*"}
		}
		for _, a := range agents {
			b.WriteString("User-agent: " + a + "\n")
		}
		allow := g.Allow
		if len(allow) == 0 && len(g.Disallow) == 0 {
			allow = []string{"/"}
		}
		for _, p := range allow {
			b.WriteString("Allow: " + p + "\n")
		}
		for _, p := range g.Disallow {
			b.WriteString("Disallow: " + p + "\n")
		}
	}
	if len(r.Sitemaps) > 0 {
		b.WriteString("\n")
		for _, s := range r.Sitemaps {
			b.WriteString("Sitemap: " + s + "\n")
		}
	}
	return []byte(b.String())
}

// writeComment writes one "# ..." line, or nothing for an empty text. A
// multi-line comment gets one "#" per line.
func writeComment(b *strings.Builder, text string) {
	if text == "" {
		return
	}
	for _, line := range strings.Split(text, "\n") {
		b.WriteString("# " + line + "\n")
	}
}

// Link is one entry of an llms.txt section.
type Link struct {
	Title string
	URL   string
	Note  string // optional, rendered after ": "
}

// Section is one H2 block of an llms.txt: a title and its link list.
type Section struct {
	Title string
	Links []Link
}

// LLMs is an llms.txt file (llmstxt.org): the site name, a one-paragraph
// summary, free-form notes, then the link sections.
type LLMs struct {
	Title    string   // H1: the site name
	Summary  string   // the blockquote right under it, one paragraph
	Notes    []string // plain paragraphs between the summary and the sections
	Sections []Section
}

// Text renders the llms.txt as Markdown.
func (l LLMs) Text() []byte {
	var b strings.Builder
	b.WriteString("# " + l.Title + "\n")
	if l.Summary != "" {
		b.WriteString("\n> " + l.Summary + "\n")
	}
	for _, n := range l.Notes {
		b.WriteString("\n" + n + "\n")
	}
	for _, s := range l.Sections {
		b.WriteString("\n## " + s.Title + "\n\n")
		for _, k := range s.Links {
			b.WriteString("- [" + k.Title + "](" + k.URL + ")")
			if k.Note != "" {
				b.WriteString(": " + k.Note)
			}
			b.WriteString("\n")
		}
	}
	return []byte(b.String())
}
