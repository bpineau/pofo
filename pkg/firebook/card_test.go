package firebook

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"net/http"
	"strings"
	"testing"
)

// Every edition ships a real, card-sized PNG. The image is committed rather
// than rendered at runtime (rasterizing SVG needs a browser), so nothing but a
// test stands between a botched regeneration and the book's face on every
// share.
func TestCardPNGIsShipped(t *testing.T) {
	for _, e := range Editions {
		body := e.CardPNG()
		if len(body) == 0 {
			t.Fatalf("%s: no card is embedded", e.Lang)
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(body))
		if err != nil {
			t.Fatalf("%s: card is not a decodable image: %v", e.Lang, err)
		}
		if format != "png" {
			t.Errorf("%s: card format = %q, want png", e.Lang, format)
		}
		if cfg.Width != CardWidth || cfg.Height != CardHeight {
			t.Errorf("%s: card is %dx%d, want %dx%d (scripts/card-shot.sh)",
				e.Lang, cfg.Width, cfg.Height, CardWidth, CardHeight)
		}
	}
}

// The card is the book's hero block: it says who publishes it, what it is
// called, what it promises and what it covers, and every one of those words
// already exists in the Edition. Nothing on it is copy written for the card.
func TestCardSVGSaysWhatTheEditionSays(t *testing.T) {
	for _, e := range Editions {
		svg := e.CardSVG()
		promise, topics := splitLede(e.SiteLede)
		want := []string{"POFO", e.SiteName, promise, strings.ToUpper(e.UI.SwitchLabel)}
		want = append(want, topics...)
		for _, w := range want {
			if !strings.Contains(svg, esc(w)) {
				t.Errorf("%s: the card does not carry %q", e.Lang, w)
			}
		}
		if !strings.Contains(svg, fmt.Sprintf(`width="%d" height="%d"`, CardWidth, CardHeight)) {
			t.Errorf("%s: the card SVG does not declare its size", e.Lang)
		}
		// House rules of the plate system: opaque fills only (rgba is painted
		// solid black by KOReader's renderer) and no typographic dashes.
		if strings.Contains(svg, "rgba(") {
			t.Errorf("%s: the card uses rgba; pre-blend the colour instead", e.Lang)
		}
		if strings.ContainsAny(svg, "—–") {
			t.Errorf("%s: the card carries an em- or en-dash", e.Lang)
		}
		// Nothing is allowed to run off the plate.
		for _, line := range wrapItems(topics, cardRight-cardMargin, 25, 0.51) {
			if w := textWidth(strings.Join(line, "   "), 25, 0.51); w > cardRight-cardMargin {
				t.Errorf("%s: a contents line is ~%.0fpx wide, over the %.0f measure",
					e.Lang, w, cardRight-cardMargin)
			}
		}
	}
}

// splitLede cuts where the lede itself cuts, in either edition's punctuation,
// and degrades gracefully on a lede with no colon at all.
func TestSplitLede(t *testing.T) {
	cases := []struct {
		in      string
		promise string
		topics  []string
	}{
		{"A promise : one, two, three.", "A promise", []string{"one", "two", "three"}},
		{"A promise: one, two.", "A promise", []string{"one", "two"}},
		{"No colon at all", "No colon at all", nil},
	}
	for _, c := range cases {
		promise, topics := splitLede(c.in)
		if promise != c.promise || len(topics) != len(c.topics) {
			t.Errorf("splitLede(%q) = %q, %v", c.in, promise, topics)
			continue
		}
		for i, want := range c.topics {
			if topics[i] != want {
				t.Errorf("splitLede(%q) topic %d = %q, want %q", c.in, i, topics[i], want)
			}
		}
	}
}

// The route serves the embedded bytes, typed and revalidatable, and the pages
// point their og:image at exactly that URL, absolutely.
func TestCardRouteAndMetaAgree(t *testing.T) {
	srv := siteServer(t)
	art := Categories[0].Articles[0]

	for _, e := range Editions {
		url := e.HomePath + CardFileName
		code, body := get(t, srv, url)
		if code != http.StatusOK {
			t.Fatalf("%s: status %d", url, code)
		}
		if body != string(e.CardPNG()) {
			t.Errorf("%s does not serve the embedded card", url)
		}
	}

	want := []string{
		`<meta property="og:image" content="` + srv.URL + French.HomePath + CardFileName + `">`,
		fmt.Sprintf(`<meta property="og:image:width" content="%d">`, CardWidth),
		fmt.Sprintf(`<meta property="og:image:height" content="%d">`, CardHeight),
		`<meta name="twitter:card" content="summary_large_image">`,
	}
	for _, path := range []string{French.HomePath, French.HomePath + art.Slug} {
		_, page := get(t, srv, path)
		for _, w := range want {
			if !strings.Contains(page, w) {
				t.Errorf("%s misses %q", path, w)
			}
		}
		if strings.Contains(page, `content="summary"`) {
			t.Errorf("%s still declares the small twitter card", path)
		}
	}

	// The English page points at the English card, not at the French one.
	_, en := get(t, srv, English.HomePath)
	if !strings.Contains(en, `content="`+srv.URL+English.HomePath+CardFileName+`">`) {
		t.Error("the English index does not point at its own card")
	}
}

// The card carries a strong ETag and honors If-None-Match: it is bytes frozen
// in the binary, so a crawler should refetch it once per deploy at most.
func TestCardRevalidates(t *testing.T) {
	srv := siteServer(t)
	url := srv.URL + French.HomePath + CardFileName
	res, err := http.Get(url) //nolint:noctx // test-local request
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	res.Body.Close()
	tag := res.Header.Get("ETag")
	if tag == "" {
		t.Fatal("the card carries no ETag")
	}
	if got := res.Header.Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil) //nolint:noctx // test-local request
	req.Header.Set("If-None-Match", tag)
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("conditional get: %v", err)
	}
	res2.Body.Close()
	if res2.StatusCode != http.StatusNotModified {
		t.Errorf("conditional get: status %d, want 304", res2.StatusCode)
	}
}
