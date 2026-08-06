package marketdata

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFeesFromFTFundsTearsheet(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/data/funds/tearsheet/summary", func(w http.ResponseWriter, r *http.Request) {
		if s := r.URL.Query().Get("s"); s != "FR0000120271:EUR" && s != "FR0000120271:USD" {
			http.Error(w, "not found", 404)
			return
		}
		fmt.Fprint(w, `<table><tr><th>Ongoing charge</th><td>1.49%</td></tr></table>`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	// An ISIN unknown to the catalog: goes through the FT tearsheet (EUR first).
	ter, ok := c.Fees("FR0000120271")
	if !ok || ter != 1.49 {
		t.Fatalf("Fees = %v, %v — want 1.49", ter, ok)
	}
	// Second call: served from the disk cache, dead server.
	srv.Close()
	c2 := NewClient(c.CacheDir)
	stubAllBases(c2, srv.URL)
	if ter, ok := c2.Fees("FR0000120271"); !ok || ter != 1.49 {
		t.Fatalf("Fees from the cache = %v, %v", ter, ok)
	}
}

func TestFeesMissRecorded(t *testing.T) {
	mux := http.NewServeMux() // no source responds
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	if _, ok := c.Fees("FR0000120271"); ok {
		t.Fatal("unexpected fees")
	}
	// The miss is recorded: no new request.
	srv.Close()
	c2 := NewClient(c.CacheDir)
	stubAllBases(c2, srv.URL)
	if _, ok := c2.Fees("FR0000120271"); ok {
		t.Fatal("the miss should have been cached")
	}
}

func TestFeesPinnedInCatalog(t *testing.T) {
	// A catalog entry with pinned fees triggers no network call.
	mux := http.NewServeMux()
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	for _, e := range catalog {
		if e.Fees > 0 {
			if ter, ok := c.Fees(e.ID); !ok || ter != e.Fees {
				t.Errorf("Fees(%s) = %v, %v — want %v", e.ID, ter, ok, e.Fees)
			}
			return
		}
	}
	t.Skip("no catalog entry has pinned fees yet")
}
