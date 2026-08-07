package firebook

import "testing"

// Shared helpers for the reference-table guard tests. The tables of this book
// are checked against the prose they replace, so every one of those tests needs
// to read an article back out of the embedded assets.

// bookArticle returns the raw Markdown of one article, by slug.
func bookArticle(t *testing.T, slug string) string {
	t.Helper()
	raw, err := assets.ReadFile("assets/book/fr/" + slug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
