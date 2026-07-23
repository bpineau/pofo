package firebook

import "github.com/bpineau/pofo/pkg/bookmd"

// ToHTML renders one article body (the book's Markdown dialect) to HTML.
// titles maps the slugs of WRITTEN articles to their display titles; it
// drives [[slug]] links. Handler passes Titles(); direct callers may pass
// nil to render wiki-links as plain text.
//
// It is a thin wrapper over bookmd.ToHTML wired with the book's figure
// generator (FigureSVG); the shared Markdown dialect lives in pkg/bookmd.
func ToHTML(src string, titles map[string]string) string {
	return bookmd.ToHTML(src, bookmd.Options{Titles: titles, Figure: FigureSVG})
}
