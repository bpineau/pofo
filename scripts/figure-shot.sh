#!/bin/sh
# figure-shot.sh: render one book plate to PNG, at the width a reader sees.
#
#   scripts/figure-shot.sh <slug> [fr|en|--both] [out.png]
#
# This is the render half of the render-and-look loop the illustration
# campaigns run on: write a plate, shoot it, read the PNG with your own eyes,
# fix what is actually broken. The plate goes through the real page CSS and the
# book's embedded fonts in headless Chrome, at the book's reading measure, so
# what comes out is what the reader gets rather than a raw SVG preview.
#
# "fr" renders through FigureSVG, "en" through FigureSVGEnglish (which also
# picks up the plates whose data the English edition adapts). "--both" writes
# one file per language, suffixing the output name. The default output path is
# ./<slug>-<lang>.png.
#
# An unknown slug prints the available ones and exits 2.
#
# Requires a Chrome/Chromium binary, like figure-audit.sh, and for the same
# reason is NOT part of "make check".
set -e

usage() {
	echo "usage: scripts/figure-shot.sh <slug> [fr|en|--both] [out.png]" >&2
	exit 2
}

SLUG="$1"
[ -n "$SLUG" ] || usage
LANG_ARG="${2:-fr}"
OUT="$3"

case "$LANG_ARG" in
fr | en) LANGS="$LANG_ARG" ;;
--both) LANGS="fr en" ;;
*) usage ;;
esac

CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
[ -x "$CHROME" ] || CHROME="$(command -v chromium chrome google-chrome 2>/dev/null | head -1)"
[ -x "$CHROME" ] || {
	echo "no Chrome/Chromium found; set CHROME=" >&2
	exit 1
}

# The helper imports pofo packages, so it has to live inside the module.
DIR="$(mktemp -d "$PWD/.figure-shot.XXXXXX")"
trap 'rm -rf "$DIR"' EXIT

cat >"$DIR/main.go" <<'GOEOF'
// Renders one plate of one edition into a standalone page.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bpineau/pofo/pkg/firebook"
	"github.com/bpineau/pofo/pkg/webui"
)

func main() {
	id, lang := os.Args[1], os.Args[2]
	render := firebook.FigureSVG
	if lang == "en" {
		render = firebook.FigureSVGEnglish
	}
	svg := render(id)
	if svg == "" {
		fmt.Fprintf(os.Stderr, "unknown plate %q; the book draws:\n", id)
		for _, known := range firebook.FigureIDs() {
			fmt.Fprintf(os.Stderr, "  %s\n", known)
		}
		os.Exit(2)
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><meta charset="utf-8"><style>`)
	b.WriteString(webui.FontsCSS)
	b.WriteString(webui.CSS)
	// The book's reading measure, so the plate renders at the width a reader
	// sees, on the page's own paper colour and with no chrome around it.
	b.WriteString(`body{max-width:44rem;margin:0 auto;padding:1rem;background:#fffdf9;` +
		`font-family:var(--sans)}.book-fig{margin:0}` +
		`.book-fig svg{padding:0;border:0;width:100%;display:block}</style>` +
		`<body class="book"><figure class="book-fig">` + svg + `</figure>`)
	fmt.Print(b.String())
}
GOEOF

for LANG in $LANGS; do
	# The helper exits 2 on an unknown slug, after listing what the book draws;
	# "go run" turns that into its own status plus a trailer of its own, so the
	# message is passed through and the intended code restored.
	if ! go run "./$(basename "$DIR")" "$SLUG" "$LANG" >"$DIR/plate.html" 2>"$DIR/err"; then
		grep -v '^exit status' "$DIR/err" >&2 || true
		exit 2
	fi

	TARGET="$OUT"
	if [ -z "$TARGET" ] || [ "$LANG_ARG" = "--both" ]; then
		TARGET="./$SLUG-$LANG.png"
		[ -n "$OUT" ] && TARGET="${OUT%.png}-$LANG.png"
	fi

	# 800 CSS pixels is the reading measure plus its margins; the page grows
	# downward, so the window only has to be tall enough for the tallest plate.
	"$CHROME" --headless --disable-gpu --no-sandbox --hide-scrollbars \
		--virtual-time-budget=10000 --window-size=800,1200 \
		--screenshot="$TARGET" "file://$DIR/plate.html" 2>/dev/null

	echo "$TARGET"
done
