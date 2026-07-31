package marketdata

import (
	"context"
	"fmt"
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

	got, err := c.Resolve(context.Background(), "XX1234567890")
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

func TestSearchFreeText(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query().Get("q"); q != "msci world" {
			t.Errorf("unexpected query: %q", q)
		}
		fmt.Fprint(w, `{"quotes":[
			{"symbol":"IWDA.AS","longname":"iShares Core MSCI World","quoteType":"ETF"},
			{"symbol":"IWDA.AS","longname":"duplicate listing","quoteType":"ETF"},
			{"symbol":"URTH","shortname":"iShares MSCI World ETF","quoteType":"ETF"}]}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	got, err := c.Search(context.Background(), "msci world")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 deduplicated candidates, got %+v", got)
	}
	if got[0].Symbol != "IWDA.AS" || got[0].Name != "iShares Core MSCI World" || got[0].Source != "yahoo" {
		t.Errorf("first candidate: %+v", got[0])
	}
	if got[1].Symbol != "URTH" || got[1].Name != "iShares MSCI World ETF" {
		t.Errorf("second candidate: %+v", got[1])
	}
}

func TestSearchCatalogPinFirst(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"OTHER","longname":"Something Else","quoteType":"EQUITY"}]}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	// GOLD is a built-in alias for XAUUSD, whose resolution is pinned in
	// the catalog: the pin must come first, then the live candidates.
	got, err := c.Search(context.Background(), "GOLD")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 {
		t.Fatalf("want the pinned candidate plus the live ones, got %+v", got)
	}
	if got[0].Name == "" || got[0].Symbol == "" {
		t.Errorf("pinned candidate incomplete: %+v", got[0])
	}
	if got[len(got)-1].Symbol != "OTHER" {
		t.Errorf("live candidates should follow the pin: %+v", got)
	}
}

func TestSearchNoResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[]}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	if _, err := c.Search(context.Background(), "no such thing"); err == nil {
		t.Error("no result should be an error")
	}
}
