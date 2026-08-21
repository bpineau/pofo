package firebook_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bpineau/pofo/pkg/firebook"
)

// Any server can mount the whole book under a prefix; the pofo web surfaces
// do exactly this at /firebook/fr/.
func ExampleHandler() {
	mux := http.NewServeMux()
	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr", firebook.Handler()))
	fmt.Println("book mounted at /firebook/fr/")
	// Output: book mounted at /firebook/fr/
}

// Mounting both editions: WithAlternate gives each one the other's base path,
// which cross-links every article that exists in both languages.
func ExampleWithAlternate() {
	mux := http.NewServeMux()
	mux.Handle("/firebook/fr/", http.StripPrefix("/firebook/fr",
		firebook.Handler(firebook.WithAlternate("/firebook/en/", firebook.English))))
	mux.Handle("/firebook/en/", http.StripPrefix("/firebook/en",
		firebook.English.Handler(firebook.WithAlternate("/firebook/fr/", firebook.French))))
	fmt.Println("both editions mounted, cross-linked")
	// Output: both editions mounted, cross-linked
}

// BookSite adds the machine-readable root files a crawler and an AI agent
// look for: /sitemap.xml, /robots.txt and /llms.txt, covering both editions
// plus whatever else the server publishes.
func ExampleBookSite() {
	mux := http.NewServeMux()
	firebook.BookSite(firebook.Page{
		Path: "/", Title: "pofo", Note: "the front door",
	}).Handle(mux)
	fmt.Println("sitemap.xml, robots.txt and llms.txt mounted at the root")
	// Output: sitemap.xml, robots.txt and llms.txt mounted at the root
}

// Site renders the same files without a server, for a caller that publishes
// them another way (a static export, a different mux).
func ExampleSite_LLMsTXT() {
	site := firebook.BookSite()
	txt := string(site.LLMsTXT("https://example.org"))
	fmt.Println(strings.HasPrefix(txt, "# pofo\n"),
		strings.Contains(txt, "https://example.org/firebook/en/what-is-fire.md"))
	// Output: true true
}

// ToHTML renders the book's Markdown dialect; Titles supplies the link
// targets for [[slug]] wiki-links.
func ExampleToHTML() {
	html := firebook.ToHTML("Voir [[la-regle-des-4-pourcents]].", firebook.Titles())
	fmt.Println(strings.Contains(html, `<a href="la-regle-des-4-pourcents"`))
	// Output: true
}
