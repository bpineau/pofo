package firebook

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/bookmd"
	"github.com/bpineau/pofo/pkg/epub"
)

// epubIdentifier is the book's stable, unique EPUB identifier. It is a fixed
// urn:uuid generated once and hardcoded so every export of the book carries the
// same dc:identifier (readers use it to recognize the same publication across
// editions); it must never change for this book.
const epubIdentifier = "urn:uuid:6f8a2c1e-4d3b-4a7e-9c21-8f5b0e6d4a92"

// epubCSS is the book's EPUB stylesheet. It is theme-neutral by design (no
// page-wide color, background or font-family), so a reader's own theme wins.
//
//go:embed assets/book/epub.css
var epubCSS string

// bookHomePath is where the always-current online edition lives (the mount
// path used across the project); the title page points readers back to it.
const bookHomePath = "/firebook/fr/"

// EPUB renders the whole book as an EPUB 3 file: a title page, then one page
// per category (each with its articles nested beneath it in the table of
// contents), then every article. modified stamps dcterms:modified and every
// zip entry, so the output is deterministic for a given modified time (the
// HTTP route can hash it for an ETag). There is no cover in this edition.
func EPUB(modified time.Time) ([]byte, error) {
	chapters := []epub.Chapter{{
		FileName: "titlepage.xhtml",
		Title:    siteName,
		Body:     titlePageBody(),
	}}

	href := func(slug string) string { return slug + ".xhtml" }
	titles := Titles()
	for i, cat := range Categories {
		children := make([]epub.Chapter, 0, len(cat.Articles))
		for _, a := range cat.Articles {
			body, err := articleEPUBBody(a, href, titles)
			if err != nil {
				return nil, err
			}
			children = append(children, epub.Chapter{
				FileName: a.Slug + ".xhtml",
				Title:    a.Title,
				Body:     body,
			})
		}
		chapters = append(chapters, epub.Chapter{
			FileName: "cat-" + strconv.Itoa(i) + ".xhtml",
			Title:    cat.Title,
			Body:     categoryPageBody(cat),
			Children: children,
		})
	}

	book := &epub.Book{
		Title:       siteName,
		Author:      "pofo",
		Language:    "fr",
		Identifier:  epubIdentifier,
		Description: siteDescription,
		Modified:    modified,
		CSS:         epubCSS,
		Chapters:    chapters,
	}

	var buf bytes.Buffer
	if err := book.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// titlePageBody renders the opening page: the book title, its subtitle (the
// same hero sentence as the web index page) and an edition note pointing at
// the always-current online version.
func titlePageBody() string {
	return fmt.Sprintf(`<section epub:type="titlepage">`+
		`<h1>%s</h1>`+
		`<p class="subtitle">%s</p>`+
		`<p class="edition">La version en ligne, tenue à jour, est publiée par pofo à %s.</p>`+
		`</section>`,
		html.EscapeString(siteName),
		html.EscapeString(siteLede),
		html.EscapeString(bookHomePath))
}

// categoryPageBody renders one category's own page: its title, its blurb and an
// ordered list linking to the articles it contains (the articles nest under it
// in the table of contents).
func categoryPageBody(cat Category) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<h2>%s</h2><p class="cat-blurb">%s</p><ol class="cat-toc">`,
		html.EscapeString(cat.Title), html.EscapeString(cat.Blurb))
	for _, a := range cat.Articles {
		fmt.Fprintf(&b, `<li><a href="%s.xhtml">%s</a></li>`,
			html.EscapeString(a.Slug), html.EscapeString(a.Title))
	}
	b.WriteString(`</ol>`)
	return b.String()
}

// articleEPUBBody renders one article to a well-formed XHTML fragment: the
// manifest title as the h1 (the in-file "# " line is dropped, as on the web),
// then the rendered body with wiki-links pointing at "<slug>.xhtml", then the
// XHTML normalization pass.
func articleEPUBBody(a Article, href func(string) string, titles map[string]string) (string, error) {
	raw, err := assets.ReadFile("assets/book/fr/" + a.Slug + ".md")
	if err != nil {
		return "", fmt.Errorf("firebook: reading %s: %w", a.Slug, err)
	}
	body := strings.TrimSpace(string(raw))
	if strings.HasPrefix(body, "# ") {
		if _, rest, found := strings.Cut(body, "\n"); found {
			body = rest
		} else {
			body = ""
		}
	}
	rendered := bookmd.ToHTML(body, bookmd.Options{Titles: titles, Href: href, Figure: FigureSVG})
	return fmt.Sprintf(`<h1>%s</h1>%s`, html.EscapeString(a.Title), epub.Normalize(rendered)), nil
}
