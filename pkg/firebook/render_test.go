package firebook

import (
	"strings"
	"testing"
)

func TestToHTMLBlocks(t *testing.T) {
	titles := map[string]string{"cible": "La cible"}
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
			got := ToHTML(c.src, titles)
			if !strings.Contains(got, c.want) {
				t.Errorf("ToHTML(%q) = %q, want it to contain %q", c.src, got, c.want)
			}
		})
	}
}

func TestToHTMLCallout(t *testing.T) {
	got := ToHTML("::: astuce Mon titre\ncorps **fort**\n:::", nil)
	for _, want := range []string{
		`doc-box--astuce`, "Mon titre", "<strong>fort</strong>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("callout HTML %q misses %q", got, want)
		}
	}
	// Untitled callout falls back to the type label; unknown type to encart.
	if got := ToHTML("::: cle\ncorps\n:::", nil); !strings.Contains(got, "L&#39;idée clé") {
		t.Errorf("untitled cle callout misses default label: %q", got)
	}
	if got := ToHTML("::: nimporte quoi\ncorps\n:::", nil); !strings.Contains(got, "doc-box--encart") {
		t.Errorf("unknown callout type should degrade to encart: %q", got)
	}
}
