package firebook

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFrenchEditionWiring(t *testing.T) {
	if French.Lang != "fr" || French.AssetDir != "assets/book/fr" {
		t.Fatalf("French edition misconfigured: %+v", French)
	}
	if French.EPUBFileName != "le-fire-tranquille.epub" {
		t.Errorf("epub filename moved: %s", French.EPUBFileName)
	}
	// The wrappers must be the French edition.
	a, b := Titles(), French.Titles()
	if len(a) == 0 || len(a) != len(b) {
		t.Errorf("Titles wrapper diverges: %d vs %d", len(a), len(b))
	}
	// A served article must carry the French chrome.
	srv := httptest.NewServer(French.Handler())
	defer srv.Close()
	res, err := http.Get(srv.URL + "/fire-cest-quoi")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	for _, want := range []string{`lang="fr"`, "Dans la même partie", "fr_FR", `content:"lien copié"`} {
		if !strings.Contains(string(body), want) {
			t.Errorf("article page lost %q", want)
		}
	}
}

func TestEnglishEditionWiring(t *testing.T) {
	if English.EPUBIdentifier == French.EPUBIdentifier {
		t.Fatal("editions must not share a publication identifier")
	}
	if English.SiteName != "The Quiet FIRE" || English.Lang != "en" {
		t.Errorf("misconfigured: %+v", English)
	}
	if got := English.UI.HumanSize(1<<20 + 400<<10); got != "1.4 MB" {
		t.Errorf("HumanSize: got %s", got)
	}
	if len(English.Categories) != len(French.Categories) {
		t.Errorf("the two editions must have the same number of parts: %d vs %d",
			len(English.Categories), len(French.Categories))
	}
	// The English callout labels reach the renderer.
	if got := English.ToHTML("::: cle\nBody.\n:::", nil); !strings.Contains(got, "The key idea") {
		t.Errorf("English callout labels not wired: %s", got)
	}
}

// plannedEN is the FR -> EN slug map; its English slugs must be free of
// duplicates, its French ones too, and it must cover every French article
// outside the fr-only part plus the English-only US part. Which French
// articles it is allowed to leave out is checked against the in-file markers
// by TestPlannedENPairsEveryFrenchArticle.
func TestPlannedENIsWellFormed(t *testing.T) {
	seen := make(map[string]bool, len(plannedEN))
	sources := make(map[string]bool, len(plannedEN))
	own := 0
	for _, p := range plannedEN {
		if seen[p.EN] {
			t.Errorf("plannedEN lists %q twice", p.EN)
		}
		seen[p.EN] = true
		if p.FR == "" {
			own++
			continue
		}
		if sources[p.FR] {
			t.Errorf("plannedEN pairs %q with two English articles", p.FR)
		}
		sources[p.FR] = true
	}
	if own != len(usFrameworkEN) {
		t.Errorf("plannedEN has %d source-less articles, want %d", own, len(usFrameworkEN))
	}
	for _, slug := range usFrameworkEN {
		if !seen[slug] {
			t.Errorf("plannedEN is missing the US-part article %q", slug)
		}
	}
	if len(plannedENSource) != len(sources) {
		t.Errorf("plannedENSource has %d entries, want %d", len(plannedENSource), len(sources))
	}
}

// The English EPUB assembles the whole translated tree through the figure
// pass and epub.Normalize; nothing else builds it before M4, so a bad
// dictionary entry or a malformed translation would otherwise ship silently.
func TestEnglishEPUBBuilds(t *testing.T) {
	blob, err := English.EPUB(time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(blob), int64(len(blob)))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range zr.File {
		got[f.Name] = true
	}
	if !got["OEBPS/titlepage.xhtml"] {
		t.Error("missing the title page")
	}
	for _, cat := range English.Categories {
		for _, a := range cat.Articles {
			if !got["OEBPS/"+a.Slug+".xhtml"] {
				t.Errorf("missing chapter %s.xhtml", a.Slug)
			}
		}
	}
	// Deterministic for a fixed modified time, like the French one.
	again, err := English.EPUB(time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(blob, again) {
		t.Error("English EPUB is not deterministic")
	}
}
