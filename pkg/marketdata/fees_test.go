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

	// LU… inconnu du catalogue: passe par le tearsheet FT (EUR d'abord).
	ter, ok := c.Fees("FR0000120271")
	if !ok || ter != 1.49 {
		t.Fatalf("Fees = %v, %v — attendu 1.49", ter, ok)
	}
	// Second appel: servi par le cache disque, serveur mort.
	srv.Close()
	c2 := NewClient(c.CacheDir)
	stubAllBases(c2, srv.URL)
	if ter, ok := c2.Fees("FR0000120271"); !ok || ter != 1.49 {
		t.Fatalf("Fees depuis le cache = %v, %v", ter, ok)
	}
}

func TestFeesMissRecorded(t *testing.T) {
	mux := http.NewServeMux() // aucune source ne répond
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	if _, ok := c.Fees("FR0000120271"); ok {
		t.Fatal("frais inattendus")
	}
	// L'échec est mémorisé: pas de nouvelle requête.
	srv.Close()
	c2 := NewClient(c.CacheDir)
	stubAllBases(c2, srv.URL)
	if _, ok := c2.Fees("FR0000120271"); ok {
		t.Fatal("le miss devait être en cache")
	}
}

func TestFeesPinnedInCatalog(t *testing.T) {
	// Une entrée du catalogue avec frais épinglés ne déclenche aucun appel.
	mux := http.NewServeMux()
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	for _, e := range catalog {
		if e.Fees > 0 {
			if ter, ok := c.Fees(e.ID); !ok || ter != e.Fees {
				t.Errorf("Fees(%s) = %v, %v — attendu %v", e.ID, ter, ok, e.Fees)
			}
			return
		}
	}
	t.Skip("aucune entrée du catalogue n'a encore de frais épinglés")
}
