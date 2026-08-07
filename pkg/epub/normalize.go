package epub

import "strings"

// Normalize rewrites the finite HTML inventory emitted by pkg/bookmd into
// well-formed XHTML so it can live inside an EPUB content document (which is
// parsed as XML). The renderer's tag set is closed, so the pass is small and
// exact:
//
//   - <hr> becomes the self-closed <hr/>.
//   - Task-list checkboxes, which readers do not render as form controls,
//     become a text glyph that survives every reader:
//     <input type="checkbox" disabled> becomes ☐ and its checked variant ☑.
//     The surrounding <li class="task"> wrapper and the following text stay.
//
// Everything else the renderer emits (aside/div/span, h2-h4, tables, lists,
// blockquote, p, a, strong/em/code, figure/figcaption and the inline SVG
// subset) is already well-formed: attributes are quoted and SVG elements are
// self-closed at the source. Normalize is idempotent and leaves already
// well-formed markup untouched.
func Normalize(html string) string {
	// The checked variant is rewritten first; neither pattern is a substring
	// of the other, so order is not load bearing, but it keeps intent clear.
	html = strings.ReplaceAll(html,
		`<input type="checkbox" disabled checked>`, "☑")
	html = strings.ReplaceAll(html,
		`<input type="checkbox" disabled>`, "☐")
	html = strings.ReplaceAll(html, "<hr>", "<hr/>")
	return html
}
