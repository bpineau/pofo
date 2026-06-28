package marketdata

import (
	"net/http"
	"testing"
)

func TestResolveFromCacheNoNetwork(t *testing.T) {
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { requests++ })
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	// Seed a disk resolution, then resolve must answer from it, no network.
	c.saveResolution("XX1234567890", resolution{
		Source: "yahoo", Symbol: "AAPL", Name: "Apple Inc.", Currency: "USD",
	})

	got, err := c.Resolve("XX1234567890")
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "yahoo" || got.Symbol != "AAPL" || got.Currency != "USD" {
		t.Fatalf("resolution misread: %+v", got)
	}
	if requests != 0 {
		t.Errorf("Resolve hit the network %d times for a cached id, want 0", requests)
	}
}
