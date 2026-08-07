package firebook

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestDrift(t *testing.T) {
	fr := []byte("# Titre\n\nCorps.\n")
	sum := sha256.Sum256(fr)
	stamp := hex.EncodeToString(sum[:])[:12]
	fsys := fstest.MapFS{
		"book/fr/un.md":   {Data: fr},
		"book/fr/deux.md": {Data: []byte("# Deux\n\nCorps.\n")},
		"book/en/one.md":  {Data: []byte("# One\n<!-- source: un @ " + stamp + " -->\n\nBody.\n")},
	}
	english := []Category{{Title: "Start", Articles: []Article{{Slug: "one", Title: "One", Source: "un"}}}}
	got := driftFS(fsys, english)
	// "deux" is untranslated; "one" is fresh.
	want := []DriftItem{{FRSlug: "deux", Reason: "untranslated"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
	// Now stale: FR content moves, stamp does not.
	fsys["book/fr/un.md"] = &fstest.MapFile{Data: []byte("# Titre\n\nCorps revu.\n")}
	got = driftFS(fsys, english)
	if len(got) != 2 || got[0].Reason != "stale" || got[0].ENSlug != "one" {
		t.Errorf("expected one stale + one untranslated, got %+v", got)
	}
}

// A French article of the tax part is never owed to the English edition.
func TestDriftSkipsTheFrenchTaxPart(t *testing.T) {
	fsys := fstest.MapFS{"book/fr/taxe-puma.md": {Data: []byte("# PUMa\n")}}
	if got := driftFS(fsys, nil); len(got) != 0 {
		t.Errorf("the tax part must never be reported: %+v", got)
	}
}

// An English article the edition writes for itself (no Source) is not paired,
// so it neither needs a stamp nor makes any French article look translated.
func TestDriftIgnoresEditionOwnArticles(t *testing.T) {
	fsys := fstest.MapFS{
		"book/fr/un.md": {Data: []byte("# Un\n")},
		"book/en/us-accounts-and-account-order.md": {Data: []byte("# Accounts\n\nBody.\n")},
	}
	english := []Category{{Articles: []Article{{Slug: "us-accounts-and-account-order", Title: "Accounts"}}}}
	want := []DriftItem{{FRSlug: "un", Reason: "untranslated"}}
	if got := driftFS(fsys, english); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v want %+v", got, want)
	}
}

// A paired article whose stamp is missing or unreadable is stale: the report
// must never call an unverifiable translation fresh.
func TestDriftMissingStampIsStale(t *testing.T) {
	fsys := fstest.MapFS{
		"book/fr/un.md":  {Data: []byte("# Un\n")},
		"book/en/one.md": {Data: []byte("# One\n\nBody.\n")},
	}
	english := []Category{{Articles: []Article{{Slug: "one", Title: "One", Source: "un"}}}}
	got := driftFS(fsys, english)
	if len(got) != 1 || got[0].Reason != "stale" || got[0].ENSlug != "one" {
		t.Errorf("an unstamped translation must read stale, got %+v", got)
	}
}

func TestSourceStamp(t *testing.T) {
	fr, h, ok := sourceStamp([]byte("# T\n<!-- source: abc @ 0123456789ab -->\nBody"))
	if !ok || fr != "abc" || h != "0123456789ab" {
		t.Errorf("parse failed: %q %q %v", fr, h, ok)
	}
	if _, _, ok := sourceStamp([]byte("# T\n\nno stamp")); ok {
		t.Error("false positive")
	}
	if _, _, ok := sourceStamp([]byte("# T\n<!-- source: abc @ nothex12345 -->\n")); ok {
		t.Error("a malformed hash must not parse")
	}
}

// Drift over the real embedded book: it must run, and every item it reports
// must name a French article that exists.
func TestDriftOverTheEmbeddedBook(t *testing.T) {
	titles := French.Titles()
	for _, it := range Drift() {
		if _, ok := titles[it.FRSlug]; !ok {
			t.Errorf("drift names %q, which is not a French article", it.FRSlug)
		}
		if it.Reason != "stale" && it.Reason != "untranslated" {
			t.Errorf("%s: unknown reason %q", it.FRSlug, it.Reason)
		}
	}
}
