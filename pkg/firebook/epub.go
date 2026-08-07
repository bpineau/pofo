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

// epubCSS is the book's EPUB stylesheet. It is theme-neutral by design (no
// page-wide color, background or font-family), so a reader's own theme wins.
//
//go:embed assets/book/epub.css
var epubCSS string

// EPUB renders the whole edition as an EPUB 3 file: a title page, then one
// page per category (each with its articles nested beneath it in the table of
// contents), then every article. modified stamps dcterms:modified and every
// zip entry, so the output is deterministic for a given modified time (the
// HTTP route can hash it for an ETag). There is no cover in this edition.
func (e *Edition) EPUB(modified time.Time) ([]byte, error) {
	chapters := []epub.Chapter{{
		FileName: "titlepage.xhtml",
		Title:    e.SiteName,
		Body:     e.titlePageBody(),
	}}

	href := func(slug string) string { return slug + ".xhtml" }
	titles := e.Titles()
	for i, cat := range e.Categories {
		children := make([]epub.Chapter, 0, len(cat.Articles))
		for _, a := range cat.Articles {
			body, err := e.articleEPUBBody(a, href, titles)
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
			Body:     e.categoryPageBody(cat),
			Children: children,
		})
	}

	book := &epub.Book{
		Title:       e.SiteName,
		Author:      "pofo",
		Language:    e.Lang,
		Identifier:  e.EPUBIdentifier,
		Description: e.SiteDescription,
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
func (e *Edition) titlePageBody() string {
	return fmt.Sprintf(`<section epub:type="titlepage">`+
		`<h1>%s</h1>`+
		`<p class="subtitle">%s</p>`+
		`<p class="edition">`+e.UI.EditionNote+`</p>`+
		`</section>`,
		html.EscapeString(e.SiteName),
		html.EscapeString(e.SiteLede),
		html.EscapeString(e.HomePath))
}

// categoryPageBody renders one category's own page: its title, its blurb and an
// ordered list linking to the articles it contains (the articles nest under it
// in the table of contents).
func (e *Edition) categoryPageBody(cat Category) string {
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
func (e *Edition) articleEPUBBody(a Article, href func(string) string, titles map[string]string) (string, error) {
	raw, err := assets.ReadFile(e.AssetDir + "/" + a.Slug + ".md")
	if err != nil {
		return "", fmt.Errorf("firebook: reading %s: %w", a.Slug, err)
	}
	body := articleBody(raw)
	rendered := bookmd.ToHTML(body, bookmd.Options{
		Titles: titles, Href: href, Figure: e.Figure, Callouts: e.Callouts})
	return fmt.Sprintf(`<h1>%s</h1>%s`, html.EscapeString(a.Title), epub.Normalize(rendered)), nil
}
