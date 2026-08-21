package firebook

import (
	"embed"
	"fmt"
	"html"
	"strings"
)

// The social card: the one image a link to the book shows when it is shared.
//
// It is the book's own hero block, drawn at card size in the v2 plate identity
// (paper ground, letterspaced kicker, serif title, accent rule, sans deck): a
// share is the first thing most readers ever see of the book, so it had better
// look like the book rather than like a link preview.
//
// Every word on it already exists: the publisher mark, Edition.SiteName, the
// head clause of Edition.SiteLede as the promise, its tail clause as the
// contents line, and Edition.UI.SwitchLabel as the edition marker. Nothing here
// is copy written for the card.
//
// CardSVG draws it; the committed PNG next to it is that SVG rasterized once by
// scripts/card-shot.sh, which is what the route serves (a crawler fetching an
// og:image wants a bitmap, and several will not render SVG at all).

// Card geometry. 1200x630 is the size Open Graph asks for and every consumer
// crops from; the margin is the plate kit's, scaled to it.
const (
	CardWidth  = 1200
	CardHeight = 630

	// CardFileName is where an edition publishes its card, relative to its
	// mount ("/firebook/fr/card.png").
	CardFileName = "card.png"

	cardMargin = 88.0
	cardRight  = CardWidth - cardMargin
)

// cardPaper is the book's page ground (bookCSS --paper), the surface the whole
// card is printed on, and cardBottomInk the pre-blended hairline that separates
// the edition marker from the text above it (the plate kit's figRule reads too
// light at this size).
const (
	cardPaper     = "#faf6ef"
	cardBottomInk = "#c9c2b6"
)

//go:embed assets/cards/*.png
var cardAssets embed.FS

// CardSVG renders the edition's social card as a standalone 1200x630 SVG.
//
// It is deterministic and depends on nothing outside the Edition value, so the
// committed PNG can be regenerated at any time (scripts/card-shot.sh) and a
// test can assert what the card says.
func (e *Edition) CardSVG() string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %d %d" width="%d" height="%d" role="img" `+
		`xmlns="http://www.w3.org/2000/svg" font-family="%s">`,
		CardWidth, CardHeight, CardWidth, CardHeight, figSans)
	fmt.Fprintf(&b, `<title>%s</title>`, esc(e.SiteName))
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/>`, CardWidth, CardHeight, cardPaper)

	// The publisher mark, in the index page's own kicker treatment.
	b.WriteString(sTxt(cardMargin, 132, 21, figDeep, "start", "600",
		`<tspan letter-spacing="4.2">POFO</tspan>`))

	// The book's title, and under it the accent rule an article heading
	// carries in the book: the card's one loud stroke.
	fmt.Fprintf(&b, `<text x="%.1f" y="236" font-size="80" fill="%s" font-weight="600" font-family="%s">%s</text>`,
		cardMargin, figInk, figSerif, esc(e.SiteName))
	fmt.Fprintf(&b, `<rect x="%.1f" y="272" width="%.1f" height="3" fill="%s"/>`,
		cardMargin, cardRight-cardMargin, figAccent)

	// The promise, then what the book covers: the two halves of the edition's
	// own lede, split where it splits itself.
	promise, topics := splitLede(e.SiteLede)
	y := 350.0
	for _, l := range wrapText(promise, cardRight-cardMargin, 38, 0.50) {
		b.WriteString(sTxt(cardMargin, y, 38, figSoft, "start", "400", esc(l)))
		y += 50
	}
	y += 24
	for _, l := range wrapItems(topics, cardRight-cardMargin, 25, 0.51) {
		b.WriteString(sTxt(cardMargin, y, 25, figMuted, "start", "400", topicLine(l)))
		y += 36
	}

	// The edition names itself, in the words it uses to name itself elsewhere.
	fmt.Fprintf(&b, `<rect x="%.1f" y="546" width="%.1f" height="1" fill="%s"/>`,
		cardMargin, cardRight-cardMargin, cardBottomInk)
	b.WriteString(mTxt(cardMargin, 578, 18, figMuted, "start", "400",
		`<tspan letter-spacing="2.4">`+esc(strings.ToUpper(e.UI.SwitchLabel))+`</tspan>`))

	b.WriteString(`</svg>`)
	return b.String()
}

// CardPNG returns the edition's committed social card, the SVG above
// rasterized at 1200x630. It is embedded in the binary, so it never fails; a
// guard test keeps it non-empty and PNG-shaped.
func (e *Edition) CardPNG() []byte {
	body, err := cardAssets.ReadFile("assets/cards/" + e.Lang + ".png")
	if err != nil { // embedded asset, checked by a test; cannot fail at runtime
		return nil
	}
	return body
}

// cardSepPx is the air on each side of the middle dot separating two topics.
// SVG collapses runs of whitespace, so the gap is set with dx rather than with
// spaces: at this size the sans sets a bare dot tight against its neighbours,
// and a list of six items has to read as six items at a glance.
const cardSepPx = 11.0

// topicLine sets one line of the contents list: the topics in the text's own
// ink, the separators spaced with dx and tinted back toward the paper, so the
// eye counts items instead of reading a sentence.
func topicLine(items []string) string {
	var b strings.Builder
	tint := mixHex(figMuted, cardPaper, 0.5)
	for i, it := range items {
		if i > 0 {
			fmt.Fprintf(&b, `<tspan dx="%.0f" fill="%s">·</tspan>`, cardSepPx, tint)
			fmt.Fprintf(&b, `<tspan dx="%.0f">%s</tspan>`, cardSepPx, esc(it))
			continue
		}
		b.WriteString(esc(it))
	}
	return b.String()
}

// splitLede cuts a lede into its promise and the topics that follow: the
// sentence up to the first colon, then the comma-separated rest, its final
// period dropped.
//
// The two editions punctuate the colon differently (French spaces it, English
// does not) and the function does not care: it splits on the colon and trims.
// A lede with no colon is all promise and lists no topics.
func splitLede(lede string) (promise string, topics []string) {
	head, tail, ok := strings.Cut(lede, ":")
	promise = strings.TrimSpace(head)
	if !ok {
		return promise, nil
	}
	tail = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(tail), "."))
	for _, p := range strings.Split(tail, ",") {
		if p = strings.TrimSpace(p); p != "" {
			topics = append(topics, p)
		}
	}
	return promise, topics
}

// wrapItems lays a list out over as many lines as it takes, breaking only
// BETWEEN items: a contents line that breaks inside "les portefeuilles qui
// résistent" stops being a list and becomes prose. Each returned line is the
// items it holds, for topicLine to set.
func wrapItems(items []string, maxPx, size, perEm float64) [][]string {
	var lines [][]string
	width := 0.0
	for _, it := range items {
		w := textWidth(it, size, perEm)
		if n := len(lines); n > 0 && width+2*cardSepPx+8+w <= maxPx {
			lines[n-1] = append(lines[n-1], it)
			width += 2*cardSepPx + 8 + w
			continue
		}
		lines = append(lines, []string{it})
		width = w
	}
	return lines
}

// wrapText greedily breaks a sentence to fit maxPx at the given font size. A
// word longer than the whole measure is left to overflow rather than cut.
func wrapText(s string, maxPx, size, perEm float64) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	lines := []string{words[0]}
	for _, w := range words[1:] {
		last := len(lines) - 1
		if candidate := lines[last] + " " + w; textWidth(candidate, size, perEm) <= maxPx {
			lines[last] = candidate
			continue
		}
		lines = append(lines, w)
	}
	return lines
}

// textWidth estimates how wide a string sets, perEm being the font's average
// character width in ems.
//
// SVG has no text metrics and this package has no font parser, so the width is
// an estimate; the answer to that is scripts/card-shot.sh, which renders the
// card and lets a human look at it. perEm is calibrated per role against those
// renderings, which is why it is a parameter and not a constant.
func textWidth(s string, size, perEm float64) float64 {
	return float64(len([]rune(s))) * size * perEm
}

// esc escapes a string for SVG text content and attribute values alike.
func esc(s string) string { return html.EscapeString(s) }
