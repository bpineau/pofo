package marketdata

import "testing"

func TestCatalogLoadedOK(t *testing.T) {
	if len(catalog) < 140 {
		t.Fatalf("catalog has only %d entries, expected >=140", len(catalog))
	}
	byid := catalogByID()
	for _, id := range []string{"IE000KF370H3", "XAUUSD", "DBMF", "VOO", "FR0010755611"} {
		e, ok := byid[id]
		if !ok {
			t.Fatalf("%s missing", id)
		}
		if e.Source == "" || e.ID == "" {
			t.Fatalf("%s incomplete: %+v", id, e)
		}
	}
	// NTSX alias + ft resolution preserved
	e := byid["IE000KF370H3"]
	if e.Source != "ft" || e.Xid != "839245042" || len(e.Aliases) == 0 || e.Aliases[0] != "NTSX" {
		t.Fatalf("NTSX resolution lost: %+v", e)
	}
	if !e.UCITS {
		t.Fatal("NTSX UCITS lost")
	}
}

func TestLookupResolvesToFullAsset(t *testing.T) {
	// A ticker resolves through CanonicalID to the full catalog record,
	// descriptive fields included.
	a, ok := Lookup("IWDA")
	if !ok {
		t.Fatal("Lookup(IWDA) not found")
	}
	if a.ID != "IE00B4L5Y983" || a.Name == "" || a.Fees == 0 || a.Geography["US"] == 0 {
		t.Fatalf("Lookup(IWDA) incomplete: %+v", a)
	}
	// ISIN and canonical id both work; an unknown id reports not-found.
	if _, ok := Lookup("IE00B4L5Y983"); !ok {
		t.Error("Lookup by ISIN failed")
	}
	if _, ok := Lookup("definitely-not-a-real-id"); ok {
		t.Error("Lookup of an unknown id must report not-found")
	}
}
