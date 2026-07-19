package marketdata

import (
	"reflect"
	"sort"
	"testing"
)

// TestLocalCatalogParity: LocalCatalog enumerates exactly the identifier
// set KnownLocal accepts, and every alternate maps to its entry's ID.
func TestLocalCatalogParity(t *testing.T) {
	cat := LocalCatalog()
	if len(cat) < 50 {
		t.Fatalf("suspiciously small catalog: %d entries", len(cat))
	}
	seen := map[string]bool{}
	for _, a := range cat {
		if a.ID == "" {
			t.Fatal("entry with empty ID")
		}
		for _, key := range append([]string{a.ID}, a.Alt...) {
			if !KnownLocal(key) {
				t.Errorf("%s: key %q not KnownLocal", a.ID, key)
			}
			if got := CanonicalID(key); got != a.ID {
				t.Errorf("key %q: CanonicalID = %q, want %q", key, got, a.ID)
			}
			if seen[key] {
				t.Errorf("key %q listed twice", key)
			}
			seen[key] = true
		}
	}
	// The other direction: every key the index accepts is listed.
	for key := range canonicalIndex().byKey {
		if !seen[key] {
			t.Errorf("KnownLocal key %q missing from LocalCatalog", key)
		}
	}
}

func TestLocalCatalogDeterministic(t *testing.T) {
	a, b := LocalCatalog(), LocalCatalog()
	if !reflect.DeepEqual(a, b) {
		t.Fatal("two calls differ")
	}
	if !sort.SliceIsSorted(a, func(i, j int) bool { return a[i].ID < a[j].ID }) {
		t.Fatal("not sorted by ID")
	}
	for _, e := range a {
		if !sort.StringsAreSorted(e.Alt) {
			t.Errorf("%s: Alt not sorted", e.ID)
		}
	}
}

// Catalog entries carry their display name and class; bare fund-ISIN
// entries (outside the catalog) may leave both empty.
func TestLocalCatalogNames(t *testing.T) {
	named := 0
	for _, e := range LocalCatalog() {
		if e.Name != "" {
			named++
		}
	}
	if named < 50 {
		t.Fatalf("only %d named entries; catalog metadata not wired", named)
	}
}
