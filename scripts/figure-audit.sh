#!/bin/sh
# figure-audit.sh: catch labels that run off the edge of a book plate.
#
#   scripts/figure-audit.sh [fr|en]
#
# Renders every plate of the given edition through the real page CSS in
# headless Chrome and measures each text node against the plate's own box. A
# label wider than its plate is clipped silently in the browser and in the
# EPUB, which the eye misses on a full page of figures; this is the check the
# illustration campaigns used to do by hand.
#
# It matters most for the English edition: the plates were tuned to French
# strings, and English labels run wider or narrower. Run it over "en" whenever
# the figure dictionary changes, and eyeball a PNG render of any plate whose
# labels moved, per the book's print-quality bar.
#
# Measurement is done on the rendered boxes (getBoundingClientRect), so it
# survives any transform or scaling a plate applies, and is reported back in
# viewBox units. PAD requires that much clearance inside the box; TOL is the
# slack below which an overhang is not worth reporting.
#
# Requires a Chrome/Chromium binary. NOT part of "make check" for that reason.
# Exits 1 if any label overflows.
set -e

EDITION="${1:-fr}"
case "$EDITION" in
fr | en) ;;
*)
	echo "usage: scripts/figure-audit.sh [fr|en]" >&2
	exit 2
	;;
esac

CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
[ -x "$CHROME" ] || CHROME="$(command -v chromium chrome google-chrome 2>/dev/null | head -1)"
[ -x "$CHROME" ] || {
	echo "no Chrome/Chromium found; set CHROME=" >&2
	exit 1
}

PAD="${PAD:-0}"
TOL="${TOL:-1}"

# The helper imports pofo packages, so it has to live inside the module.
DIR="$(mktemp -d "$PWD/.figure-audit.XXXXXX")"
trap 'rm -rf "$DIR"' EXIT

cat >"$DIR/main.go" <<'GOEOF'
// Renders every plate of one edition into a single measuring harness.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bpineau/pofo/pkg/firebook"
	"github.com/bpineau/pofo/pkg/webui"
)

func main() {
	render := firebook.FigureSVG
	if len(os.Args) > 1 && os.Args[1] == "en" {
		render = firebook.FigureSVGEnglish
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><meta charset="utf-8"><style>`)
	b.WriteString(webui.FontsCSS)
	b.WriteString(webui.CSS)
	// The book's reading measure, so the plates render at the width a reader
	// sees. No padding or border on the svg: its client box must be exactly
	// the viewBox area for the measurement below to mean anything.
	b.WriteString(`body{max-width:44rem;margin:0 auto;padding:2.6rem 1.3rem;` +
		`font-family:var(--sans)}.book-fig{margin:1.7rem 0}` +
		`.book-fig svg{padding:0;border:0}</style><body class="book">`)
	for _, id := range firebook.FigureIDs() {
		svg := strings.Replace(render(id), "<svg ", `<svg data-fig="`+id+`" `, 1)
		fmt.Fprintf(&b, `<figure class="book-fig">%s</figure>`, svg)
	}
	b.WriteString(`<pre id="audit"></pre><script>` + audit + `</script>`)
	fmt.Print(b.String())
}

const audit = `
var PAD = Number(document.currentScript.dataset.pad || 0);
var TOL = Number(document.currentScript.dataset.tol || 1);
var out = [];
document.querySelectorAll("svg[data-fig]").forEach(function (svg) {
  var id = svg.getAttribute("data-fig");
  var box = svg.getBoundingClientRect();
  var vb = svg.viewBox.baseVal;
  if (!box.width || !vb.width) return;
  var scale = vb.width / box.width; // css pixels back to viewBox units
  svg.querySelectorAll("text, tspan").forEach(function (t) {
    if (t.getElementsByTagName("tspan").length) return; // a wrapper, not a leaf
    var label = t.textContent;
    if (!label || !label.trim()) return;
    var r = t.getBoundingClientRect();
    if (!r.width) return;
    var over = (r.right - box.right) * scale + PAD;
    var under = (box.left - r.left) * scale + PAD;
    var side = over >= under ? "right" : "left";
    var by = Math.max(over, under);
    if (by > TOL) {
      out.push('PLATE ' + id + ': "' + label + '" overflows ' + side +
        " by " + by.toFixed(1) + "px");
    }
  });
});
// The markers are assembled at run time so this source, which the dumped
// DOM also carries, cannot be mistaken for the report itself.
var mark = function (w) { return "@@AUDIT" + "-" + w + "@@"; };
document.getElementById("audit").textContent =
  mark("START") + "\n" + out.join("\n") + "\n" + mark("END");
`
GOEOF

# The script tag carries PAD/TOL so the harness stays a plain static file.
go run "./$(basename "$DIR")" "$EDITION" |
	sed "s|<script>|<script data-pad=\"$PAD\" data-tol=\"$TOL\">|" >"$DIR/plates.html"

"$CHROME" --headless --disable-gpu --no-sandbox --hide-scrollbars \
	--virtual-time-budget=10000 --dump-dom "file://$DIR/plates.html" >"$DIR/dom.html" 2>/dev/null

# Pull the report back out of the dumped DOM and undo the HTML escaping the
# dump applies to label text.
sed -n '/@@AUDIT-START@@/,/@@AUDIT-END@@/p' "$DIR/dom.html" |
	sed -e '/@@AUDIT-/d' -e '/^[[:space:]]*$/d' \
		-e 's/&amp;/\&/g; s/&lt;/</g; s/&gt;/>/g' >"$DIR/report.txt"

if [ ! -s "$DIR/report.txt" ]; then
	echo "figure-audit $EDITION: no label overflows"
	exit 0
fi
cat "$DIR/report.txt"
echo "figure-audit $EDITION: $(wc -l <"$DIR/report.txt" | tr -d ' ') label(s) overflow" >&2
exit 1
