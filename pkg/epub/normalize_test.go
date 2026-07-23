package epub

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNormalizeHR(t *testing.T) {
	in := "<p>x</p><hr><p>y</p>"
	want := "<p>x</p><hr/><p>y</p>"
	if got := Normalize(in); got != want {
		t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
	}
}

func TestNormalizeCheckbox(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{
			"unchecked",
			`<li class="task"><input type="checkbox" disabled> à faire</li>`,
			`<li class="task">☐ à faire</li>`,
		},
		{
			"checked",
			`<li class="task"><input type="checkbox" disabled checked> fait</li>`,
			`<li class="task">☑ fait</li>`,
		},
		{
			"both in one list",
			`<ul><li class="task"><input type="checkbox" disabled checked> a</li>` +
				`<li class="task"><input type="checkbox" disabled> b</li></ul>`,
			`<ul><li class="task">☑ a</li><li class="task">☐ b</li></ul>`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Normalize(c.in); got != c.want {
				t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeIdempotent(t *testing.T) {
	in := `<p>intro</p><hr><ul><li class="task"><input type="checkbox" disabled checked> a</li>` +
		`<li class="task"><input type="checkbox" disabled> b</li></ul>`
	once := Normalize(in)
	twice := Normalize(once)
	if once != twice {
		t.Errorf("not idempotent:\n once  = %q\n twice = %q", once, twice)
	}
}

func TestNormalizePassthrough(t *testing.T) {
	// Markup that is already well-formed must be returned verbatim.
	cases := []string{
		`<aside class="doc-box doc-box--astuce"><div class="doc-box-h"><span class="doc-box-glyph">💡</span> Astuce</div><p>corps</p></aside>`,
		`<div class="table-wrap"><table><thead><tr><th>a</th></tr></thead><tbody><tr><td>1</td></tr></tbody></table></div>`,
		`<figure class="book-fig"><svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/><line x1="0" y1="0" x2="10" y2="10"/></svg><figcaption>legende</figcaption></figure>`,
		`<blockquote><p>citation</p></blockquote>`,
		`<h2 id="s">S</h2><h3 id="t">T</h3><h4 id="u">U</h4>`,
		`<p>un <strong>mot</strong> <em>seul</em> et du <code>code</code> et un <a href="x.xhtml" class="doc-link">lien</a></p>`,
	}
	for _, in := range cases {
		if got := Normalize(in); got != in {
			t.Errorf("Normalize changed well-formed markup:\n in  = %q\n got = %q", in, got)
		}
	}
}

// TestNormalizeWellFormed proves the whole inventory, once normalized and
// wrapped in a minimal XML document, parses as well-formed XML.
func TestNormalizeWellFormed(t *testing.T) {
	body := strings.Join([]string{
		`<h2 id="s">Section</h2>`,
		`<p>Un <strong>gras</strong>, un <em>italique</em>, du <code>code</code>.</p>`,
		`<hr>`,
		`<div class="table-wrap"><table><thead><tr><th>a</th><th>b</th></tr></thead><tbody><tr><td>1</td><td>2</td></tr></tbody></table></div>`,
		`<ul><li class="task"><input type="checkbox" disabled checked> fait</li><li class="task"><input type="checkbox" disabled> a faire</li></ul>`,
		`<aside class="doc-box doc-box--cle"><div class="doc-box-h"><span class="doc-box-glyph">🔑</span> Cle</div><p>corps</p></aside>`,
		`<figure class="book-fig"><svg viewBox="0 0 10 10"><rect x="0" y="0" width="10" height="10"/></svg><figcaption>fig</figcaption></figure>`,
	}, "")
	doc := `<?xml version="1.0" encoding="utf-8"?><root xmlns="http://www.w3.org/1999/xhtml">` + Normalize(body) + `</root>`
	dec := xml.NewDecoder(strings.NewReader(doc))
	for {
		_, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("normalized markup is not well-formed XML: %v", err)
		}
	}
}
