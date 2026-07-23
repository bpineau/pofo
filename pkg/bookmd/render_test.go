package bookmd

import (
	"strings"
	"testing"
)

func TestToHTMLBlocks(t *testing.T) {
	opt := Options{Titles: map[string]string{"cible": "La cible"}}
	cases := []struct {
		name, src, want string
	}{
		{"paragraph", "Bonjour le **monde**.", "<p>Bonjour le <strong>monde</strong>.</p>"},
		{"heading demoted", "## Section", `<h3 id="section">Section</h3>`},
		{"heading anchor accents", "### Déjà vu", `<h4 id="déjà-vu">Déjà vu</h4>`},
		{"rule", "---", "<hr>"},
		{"ul", "- un\n- deux", "<ul><li>un</li><li>deux</li></ul>"},
		{"ol", "1. un\n2. deux", "<ol><li>un</li><li>deux</li></ol>"},
		{"task", "- [x] fait\n- [ ] à faire", `<li class="task"><input type="checkbox" disabled checked> fait</li>`},
		{"quote", "> citation", "<blockquote><p>citation</p></blockquote>"},
		{"table", "| a | b |\n|---|---|\n| 1 | 2 |", "<table><thead><tr><th>a</th><th>b</th></tr></thead><tbody><tr><td>1</td><td>2</td></tr></tbody></table>"},
		{"wiki link known", "voir [[cible]]", `<a href="cible" class="doc-link">La cible</a>`},
		{"wiki link label", "voir [[cible|la suite]]", `<a href="cible" class="doc-link">la suite</a>`},
		{"wiki link unknown degrades", "voir [[futur-article]]", "<p>voir futur-article</p>"},
		{"external link", "[pofo](https://example.com)", `<a href="https://example.com" target="_blank" rel="noopener">pofo</a>`},
		{"inline code shielded", "le code `a*b*c` reste", "<code>a*b*c</code>"},
		{"italic", "un *mot* seul", "<em>mot</em>"},
		{"escape", "1 < 2 & 3", "<p>1 &lt; 2 &amp; 3</p>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ToHTML(c.src, opt)
			if !strings.Contains(got, c.want) {
				t.Errorf("ToHTML(%q) = %q, want it to contain %q", c.src, got, c.want)
			}
		})
	}
}

func TestToHTMLCallout(t *testing.T) {
	got := ToHTML("::: astuce Mon titre\ncorps **fort**\n:::", Options{})
	for _, want := range []string{
		`doc-box--astuce`, "Mon titre", "<strong>fort</strong>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("callout HTML %q misses %q", got, want)
		}
	}
	// Untitled callout falls back to the type label; unknown type to encart.
	if got := ToHTML("::: cle\ncorps\n:::", Options{}); !strings.Contains(got, "L&#39;idée clé") {
		t.Errorf("untitled cle callout misses default label: %q", got)
	}
	if got := ToHTML("::: nimporte quoi\ncorps\n:::", Options{}); !strings.Contains(got, "doc-box--encart") {
		t.Errorf("unknown callout type should degrade to encart: %q", got)
	}
}

// TestAdminCallout covers the admin type shared with locador's dialect.
func TestAdminCallout(t *testing.T) {
	got := ToHTML("::: admin\ncorps\n:::", Options{})
	for _, want := range []string{`doc-box--admin`, "📋", "Côté administratif"} {
		if !strings.Contains(got, want) {
			t.Errorf("admin callout %q misses %q", got, want)
		}
	}
}

// TestFigureHook covers both the nil-drop and the hooked-payload behaviors.
func TestFigureHook(t *testing.T) {
	src := "avant\n\n::: figure ma-figure\nune légende\n:::\n\naprès"

	// Nil Figure drops the block entirely: no <figure>, no caption, but the
	// surrounding prose survives.
	got := ToHTML(src, Options{})
	if strings.Contains(got, "<figure") || strings.Contains(got, "une légende") {
		t.Errorf("nil Figure should drop the block: %q", got)
	}
	if !strings.Contains(got, "<p>avant</p>") || !strings.Contains(got, "<p>après</p>") {
		t.Errorf("nil Figure should keep surrounding prose: %q", got)
	}

	// A hook emits the figure element with the payload and inlined caption.
	got = ToHTML(src, Options{Figure: func(id string) string { return "<svg>" + id + "</svg>" }})
	want := `<figure class="book-fig"><svg>ma-figure</svg><figcaption>une légende</figcaption></figure>`
	if !strings.Contains(got, want) {
		t.Errorf("hooked Figure = %q, want it to contain %q", got, want)
	}
}

// TestHrefHook covers rewriting wiki-link targets (the EPUB assembly needs
// slug.xhtml targets while the web keeps the bare slug).
func TestHrefHook(t *testing.T) {
	opt := Options{
		Titles: map[string]string{"cible": "La cible"},
		Href:   func(slug string) string { return slug + ".xhtml" },
	}
	got := ToHTML("voir [[cible]]", opt)
	if !strings.Contains(got, `<a href="cible.xhtml" class="doc-link">La cible</a>`) {
		t.Errorf("Href hook not applied: %q", got)
	}
	// Nil Href keeps the bare slug.
	got = ToHTML("voir [[cible]]", Options{Titles: opt.Titles})
	if !strings.Contains(got, `<a href="cible" class="doc-link">La cible</a>`) {
		t.Errorf("nil Href should keep bare slug: %q", got)
	}
}

func TestBoldWrappingItalic(t *testing.T) {
	// A bold span containing an italic title (bibliography style) must render
	// as <strong> with a nested <em>, not leak literal ** markers.
	got := ToHTML("**Vicki Robin, *Your Money or Your Life* (1992)** : le socle.", Options{})
	if !strings.Contains(got, "<strong>Vicki Robin, <em>Your Money or Your Life</em> (1992)</strong>") {
		t.Errorf("bold-wrapping-italic not rendered: %q", got)
	}
	// Plain bold and plain italic still work.
	if got := ToHTML("un **mot** et un *autre*", Options{}); !strings.Contains(got, "<strong>mot</strong>") || !strings.Contains(got, "<em>autre</em>") {
		t.Errorf("plain emphasis broke: %q", got)
	}
}
