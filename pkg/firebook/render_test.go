package firebook

import (
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/bookmd"
)

// TestToHTMLWrapper checks the thin wrapper still renders the dialect and
// wires the book's figure generator.
func TestToHTMLWrapper(t *testing.T) {
	got := ToHTML("Voir [[cible]].", map[string]string{"cible": "La cible"})
	if !strings.Contains(got, `<a href="cible" class="doc-link">La cible</a>`) {
		t.Errorf("wrapper wiki-link broke: %q", got)
	}
	// The wrapper passes FigureSVG, so a known figure block renders its SVG.
	if ids := FigureIDs(); len(ids) > 0 {
		got := ToHTML("::: figure "+ids[0]+"\nlégende\n:::", nil)
		if !strings.Contains(got, `<figure class="book-fig">`) || !strings.Contains(got, "<svg") {
			t.Errorf("wrapper figure block broke: %q", got)
		}
	}
}

// TestToHTMLDelegation is the no-behavior-change guard: rendering every
// embedded article through the old firebook.ToHTML path must be byte-identical
// to calling bookmd.ToHTML directly with the same figure hook.
func TestToHTMLDelegation(t *testing.T) {
	titles := Titles()
	for _, cat := range Categories {
		for _, a := range cat.Articles {
			raw, err := assets.ReadFile("assets/book/fr/" + a.Slug + ".md")
			if err != nil {
				t.Fatalf("read %s: %v", a.Slug, err)
			}
			src := string(raw)
			old := ToHTML(src, titles)
			direct := bookmd.ToHTML(src, bookmd.Options{Titles: titles, Figure: FigureSVG})
			if old != direct {
				t.Errorf("delegation mismatch for %s", a.Slug)
			}
		}
	}
}
