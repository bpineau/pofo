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

// ToHTML renders the book's Markdown dialect; Titles supplies the link
// targets for [[slug]] wiki-links.
func ExampleToHTML() {
	html := firebook.ToHTML("Voir [[la-regle-des-4-pourcents]].", firebook.Titles())
	fmt.Println(strings.Contains(html, `<a href="la-regle-des-4-pourcents"`))
	// Output: true
}
