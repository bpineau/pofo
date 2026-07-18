#!/bin/sh
# report-shot.sh: render a portfolio's HTML report and screenshot it with
# every per-portfolio section unfolded, for visual verification of report
# changes without a human in the loop.
#
#   scripts/report-shot.sh [portfolio-file] [out-prefix] [extra pofo flags...]
#
# Defaults: examples/dragon-decumulation-household.txt and /tmp/pofo-shot.
# Produces <out-prefix>.html (sections forced open) and <out-prefix>.png
# (full-page, 1500px wide). Crop a region with e.g.:
#   sips -c <height> 1500 --cropOffset <y> 0 out.png --out crop.png
#
# Requires a Chrome/Chromium binary; the report itself needs no network when
# the quote cache is warm (run ./pofo -warmup once, or render any report).
set -e

FILE="${1:-examples/dragon-decumulation-household.txt}"
OUT="${2:-/tmp/pofo-shot}"
[ $# -ge 1 ] && shift
[ $# -ge 1 ] && shift

CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
[ -x "$CHROME" ] || CHROME="$(command -v chromium chrome google-chrome 2>/dev/null | head -1)"
[ -x "$CHROME" ] || { echo "no Chrome/Chromium found; set CHROME=" >&2; exit 1; }

make -s build
./pofo -no-open -out "$OUT.raw.html" "$@" "$FILE"
sed 's/<details class="pf">/<details class="pf" open>/' "$OUT.raw.html" > "$OUT.html"
rm -f "$OUT.raw.html"

"$CHROME" --headless --disable-gpu --hide-scrollbars \
  --screenshot="$OUT.png" --window-size=1500,8000 \
  "file://$OUT.html" 2>/dev/null
echo "wrote $OUT.html and $OUT.png"
