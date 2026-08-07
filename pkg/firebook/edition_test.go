package firebook

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

// plannedEN is the FR -> EN slug map; it must be free of duplicates and cover
// every French article outside the tax part, plus the English-only US part.
func TestPlannedENIsWellFormed(t *testing.T) {
	seen := make(map[string]bool, len(plannedEN))
	for _, slug := range plannedEN {
		if seen[slug] {
			t.Errorf("plannedEN lists %q twice", slug)
		}
		seen[slug] = true
	}
	frOnly := 0
	for _, slug := range planned {
		if !taxOnlyFR[slug] {
			frOnly++
		}
	}
	if want := frOnly + len(usFrameworkEN); len(plannedEN) != want {
		t.Errorf("plannedEN has %d slugs, want %d (%d translatable French + %d US-only)",
			len(plannedEN), want, frOnly, len(usFrameworkEN))
	}
	for _, slug := range usFrameworkEN {
		if !seen[slug] {
			t.Errorf("plannedEN is missing the US-part article %q", slug)
		}
	}
}
