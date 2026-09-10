package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// fuzzyOnlyMux serves a source set where the queried ticker has no listing of
// its own and the search offers a single, deep, name-matched fund: exactly the
// stray hit FetchOptions.ExactOnly exists to refuse.
func fuzzyOnlyMux(t *testing.T) *http.ServeMux {
	t.Helper()
	deep := testDays(90)
	closes := make([]float64, len(deep))
	for i := range closes {
		closes[i] = 100 + float64(i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/FOOBAR", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no such symbol", http.StatusNotFound)
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"0P0000FOO.F","longname":"Foobar Emerging Opportunities","quoteType":"MUTUALFUND"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/0P0000FOO.F", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("0P0000FOO.F", deep, closes))
	})
	return mux
}

func TestFetchExactOnlyRefusesNameMatch(t *testing.T) {
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// The default: the fuzzy candidate is served, which is convenient for a
	// CLI user reading the resolution line and wrong for anyone else.
	t.Run("fuzzy match adopted by default", func(t *testing.T) {
		c, srv := newTestClient(t, t.TempDir(), fuzzyOnlyMux(t))
		defer srv.Close()
		s, err := c.FetchExtended(context.Background(), "FOOBAR", FetchOptions{From: from})
		if err != nil {
			t.Fatalf("the name-matched fund should have answered: %v", err)
		}
		if s.Symbol != "0P0000FOO.F" {
			t.Fatalf("resolved to %q, want the searched fund", s.Symbol)
		}
	})

	t.Run("exact-only fails instead", func(t *testing.T) {
		dir := t.TempDir()
		c, srv := newTestClient(t, dir, fuzzyOnlyMux(t))
		defer srv.Close()
		if s, err := c.FetchExtended(context.Background(), "FOOBAR",
			FetchOptions{From: from, ExactOnly: true}); err == nil {
			t.Fatalf("exact-only resolution served %q, want a miss", s.Symbol)
		}
		// A refused candidate must not be remembered either: the next
		// request has to search again, not read a poisoned resolution.
		if c.Cached("FOOBAR") {
			t.Error("a refused candidate was cached under the queried ticker")
		}
	})
}

func TestFetchExactOnlyKeepsISINResolution(t *testing.T) {
	// An ISIN is an identifier, not a name: every source is queried with the
	// code itself, so exact-only resolution must leave it alone even when the
	// listing it lands on shares nothing with the query but the instrument.
	deep := testDays(90)
	closes := make([]float64, len(deep))
	for i := range closes {
		closes[i] = 50 + float64(i)*0.5
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"XFOO.AS","longname":"Foo Core World UCITS ETF","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/XFOO.AS", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("XFOO.AS", deep, closes))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.FetchExtended(context.Background(), "IE00B4L5Y983",
		FetchOptions{From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), ExactOnly: true})
	if err != nil {
		t.Fatalf("an ISIN must still resolve under exact-only: %v", err)
	}
	if s.Symbol != "XFOO.AS" || len(s.Points) != 90 {
		t.Fatalf("ISIN resolved to %q (%d points)", s.Symbol, len(s.Points))
	}
}

func TestClientCached(t *testing.T) {
	days := testDays(80)
	closes := make([]float64, len(days))
	for i := range closes {
		closes[i] = 10 + float64(i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, closes))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()

	if c.Cached("VOO") {
		t.Error("nothing is cached before the first fetch")
	}
	if _, err := c.Fetch(context.Background(), "VOO", resolveFrom()); err != nil {
		t.Fatal(err)
	}
	if !c.Cached("VOO") {
		t.Error("a fetched ticker must read as cached: refetching it costs no upstream call")
	}
	if !c.Cached("voosim") {
		t.Error("the SIM suffix names the same quotes and must read the same")
	}
	if c.Cached("AVUV") {
		t.Error("an untouched identifier must not read as cached")
	}
	// A cache-less client keeps nothing, so nothing is ever free.
	cacheless := NewClient("")
	stubAllBases(cacheless, srv.URL)
	if cacheless.Cached("VOO") {
		t.Error("a cache-less client cannot answer offline")
	}
}
