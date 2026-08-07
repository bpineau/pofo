package bookmd_test

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/bookmd"
)

// ToHTML renders the shared book Markdown dialect. Options.Titles supplies the
// link targets for [[slug]] wiki-links; a callout box opens with ":::".
func ExampleToHTML() {
	src := "::: astuce\nVoir aussi [[la-regle|la règle des 4 %]].\n:::"
	opt := bookmd.Options{Titles: map[string]string{"la-regle": "La règle des 4 %"}}
	fmt.Println(bookmd.ToHTML(src, opt))
	// Output: <aside class="doc-box doc-box--astuce"><div class="doc-box-h"><span class="doc-box-glyph">💡</span> Astuce</div><p>Voir aussi <a href="la-regle" class="doc-link">la règle des 4 %</a>.</p></aside>
}
