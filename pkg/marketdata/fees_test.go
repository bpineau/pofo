package marketdata

import (
	"context"
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
	ter, ok := c.Fees(context.Background(), "FR0000120271")
	if !ok || ter != 1.49 {
		t.Fatalf("Fees = %v, %v; want 1.49", ter, ok)
	}
	// Second call: served from the disk cache, dead server.
	srv.Close()
	c2 := NewClient(c.CacheDir)
	stubAllBases(c2, srv.URL)
	if ter, ok := c2.Fees(context.Background(), "FR0000120271"); !ok || ter != 1.49 {
		t.Fatalf("Fees from the cache = %v, %v", ter, ok)
	}
}

func TestFeesMissRecorded(t *testing.T) {
	mux := http.NewServeMux() // no source responds
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	if _, ok := c.Fees(context.Background(), "FR0000120271"); ok {
		t.Fatal("unexpected fees")
	}
	// The miss is recorded: no new request.
	srv.Close()
	c2 := NewClient(c.CacheDir)
	stubAllBases(c2, srv.URL)
	if _, ok := c2.Fees(context.Background(), "FR0000120271"); ok {
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
			if ter, ok := c.Fees(context.Background(), e.ID); !ok || ter != e.Fees {
				t.Errorf("Fees(%s) = %v, %v; want %v", e.ID, ter, ok, e.Fees)
			}
			return
		}
	}
	t.Skip("no catalog entry has pinned fees yet")
}

// TestCandidateCurrencies: the FT tearsheet is addressed by ISIN AND share
// class currency, so the order of the guesses is what decides whether a fund
// answers at all. A known currency is the only guess; GBp is not a share
// class currency and reads as unknown.
func TestCandidateCurrencies(t *testing.T) {
	cases := []struct {
		known, isin string
		want        []string
	}{
		{"EUR", "IE00B4L5Y983", []string{"EUR"}},
		{"USD", "IE00B4L5Y983", []string{"USD"}},
		{"GBp", "GB00B4L5Y983", []string{"USD", "EUR"}}, // pence: not a class currency
		{"", "FR0010315770", []string{"EUR", "USD"}},    // continental domiciles
		{"", "LU1662501532", []string{"EUR", "USD"}},
		{"", "DE000A0F5UF5", []string{"EUR", "USD"}},
		{"", "IE00B4L5Y983", []string{"USD", "EUR"}}, // everything else
	}
	for _, tc := range cases {
		got := candidateCurrencies(tc.known, tc.isin)
		if len(got) != len(tc.want) {
			t.Fatalf("candidateCurrencies(%q, %q) = %v, want %v", tc.known, tc.isin, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("candidateCurrencies(%q, %q) = %v, want %v", tc.known, tc.isin, got, tc.want)
			}
		}
	}
}

// TestParseFeesMatch: a scraped percentage must be believable before it is
// believed, and the comma decimal separator of a European page must parse.
func TestParseFeesMatch(t *testing.T) {
	cases := []struct {
		body string
		want float64
		ok   bool
	}{
		{`Ongoing charge</th><td class="x">0.20%`, 0.20, true},
		{`Ongoing charge</th><td>0,85%`, 0.85, true},
		{`Net expense ratio</th><td>1.00%`, 1.00, true},
		{`Ongoing charge</th><td>99.00%`, 0, false}, // beyond any real fee
		{`Total expense</th><td>0.20%`, 0, false},   // not a label we read
		{`nothing here`, 0, false},
	}
	for _, tc := range cases {
		got, err := parseFeesMatch(ftFeesRe, []byte(tc.body))
		if (err == nil) != tc.ok || (tc.ok && got != tc.want) {
			t.Errorf("parseFeesMatch(%q) = %v, %v; want %v, ok=%v", tc.body, got, err, tc.want, tc.ok)
		}
	}
}
