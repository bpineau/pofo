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
