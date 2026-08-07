package firebook_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bpineau/pofo/pkg/firebook"
)

// Any server can mount the whole book under a prefix; the pofo -fire UI does
// exactly this at /livre/.
func ExampleHandler() {
	mux := http.NewServeMux()
	mux.Handle("/livre/", http.StripPrefix("/livre", firebook.Handler()))
	fmt.Println("book mounted at /livre/")
	// Output: book mounted at /livre/
}

// ToHTML renders the book's Markdown dialect; Titles supplies the link
// targets for [[slug]] wiki-links.
func ExampleToHTML() {
	html := firebook.ToHTML("Voir [[la-regle-des-4-pourcents]].", firebook.Titles())
	fmt.Println(strings.Contains(html, `<a href="la-regle-des-4-pourcents"`))
	// Output: true
}
