package marketdata

import "testing"

func TestCatalogLoadedOK(t *testing.T) {
	if len(catalog) != 106 {
		t.Fatalf("catalog has %d entries, want 106", len(catalog))
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
