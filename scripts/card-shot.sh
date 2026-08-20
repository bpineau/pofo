#!/bin/sh
# card-shot.sh: rasterize the book's social cards (og:image) to PNG.
#
#   scripts/card-shot.sh [fr|en|--both] [out-dir]
#
# The card itself is Go code, firebook.(*Edition).CardSVG, so it is versioned
# and reviewable like any other plate; this script is only the rasterizer. It
# renders the SVG in headless Chrome with the book's embedded fonts, at exactly
# 1200x630, and writes pkg/firebook/assets/cards/<lang>.png, which the binary
# embeds and the /firebook/<lang>/card.png route serves.
#
# Run it after any change to CardSVG or to an edition's title or lede, then LOOK
# at the two PNGs: this image is the book's face on every share, and no test
# replaces a pair of eyes on it.
#
# Requires a Chrome/Chromium binary, like figure-shot.sh, and for the same
# reason is NOT part of "make check".
set -e

usage() {
	echo "usage: scripts/card-shot.sh [fr|en|--both] [out-dir]" >&2
	exit 2
}

LANG_ARG="${1:---both}"
OUTDIR="${2:-pkg/firebook/assets/cards}"

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

mkdir -p "$OUTDIR"

# The helper imports pofo packages, so it has to live inside the module.
DIR="$(mktemp -d "$PWD/.card-shot.XXXXXX")"
trap 'rm -rf "$DIR"' EXIT

cat >"$DIR/main.go" <<'GOEOF'
// Renders one edition's social card into a standalone, card-sized page.
package main

import (
	"fmt"
	"os"

	"github.com/bpineau/pofo/pkg/firebook"
	"github.com/bpineau/pofo/pkg/webui"
)

func main() {
	lang := os.Args[1]
	ed := firebook.French
	if lang == "en" {
		ed = firebook.English
	}
	// No margin, no scrollbar, no page ground of its own: the SVG fills the
	// viewport exactly, so the screenshot IS the card.
	fmt.Printf(`<!doctype html><meta charset="utf-8"><style>%s`+
		`html,body{margin:0;padding:0;overflow:hidden}svg{display:block}`+
		`</style><body>%s`, webui.FontsCSS, ed.CardSVG())
}
GOEOF

for LANG in $LANGS; do
	go run "./$(basename "$DIR")" "$LANG" >"$DIR/card.html"
	TARGET="$OUTDIR/$LANG.png"
	"$CHROME" --headless --disable-gpu --no-sandbox --hide-scrollbars \
		--virtual-time-budget=10000 --window-size=1200,630 \
		--screenshot="$TARGET" "file://$DIR/card.html" 2>/dev/null
	echo "$TARGET"
done
