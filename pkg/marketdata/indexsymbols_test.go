package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// A withdrawn index symbol answers with its bundled reference, under the
// symbol asked for, without a single request.
func TestWithdrawnIndexServedFromBundle(t *testing.T) {
	hit := false
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hit = true
		http.Error(w, "no network for a bundled index", http.StatusInternalServerError)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(context.Background(), " ^bcom", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "^BCOM" || s.Source != "refdata" || s.Currency != "USD" {
		t.Errorf("series metadata: symbol %q source %q currency %q", s.Symbol, s.Source, s.Currency)
	}
	if first := s.First().Date; !first.Equal(time.Date(1991, 1, 2, 0, 0, 0, 0, time.UTC)) || len(s.Points) < 8000 {
		t.Errorf("want the index's whole daily history from 1991-01-02, got %d points from %s", len(s.Points), first.Format(time.DateOnly))
	}
	if hit {
		t.Error("a bundled index symbol touched the network")
	}
	if _, err := c.fetch(context.Background(), "^BCOM", time.Time{}, fetchSpec{wantCurrency: "EUR", nativeOnly: true}); err == nil {
		t.Error("a USD index was served under a EUR-only constraint")
	}
}

// An index symbol with no quote of its own must never adopt a fund whose NAME
// carries the index's short name, found by the search fallback, nor a cached
// resolution to one: a plain ticker in the same position would (the fuzzy
// gate only asks for the word).
func TestIndexSymbolNeverAdoptsANameMatch(t *testing.T) {
	days := testDays(120)
	closes := make([]float64, len(days))
	for i := range closes {
		closes[i] = 50 + float64(i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"ZZCOM.L","longname":"Some ZZCOM Commodities UCITS ETF","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/ZZCOM.L", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("ZZCOM.L", days, closes))
	})
	// Every other path, the index's own chart and FT and Morningstar
	// included, answers 404.
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	if s, err := c.Fetch(context.Background(), "^ZZCOM", from); err == nil {
		t.Fatalf("the index symbol adopted %s (%s)", s.Symbol, s.Name)
	}
	if _, ok := c.loadResolution("^ZZCOM"); ok {
		t.Error("a name match was cached as the index's resolution")
	}

	// A resolution poisoned before the rule existed is ignored.
	c.saveResolution("^ZZCOM", resolution{Source: "yahoo", Symbol: "ZZCOM.L", Name: "Some ZZCOM Commodities UCITS ETF"})
	if s, err := c.Fetch(context.Background(), "^ZZCOM", from); err == nil {
		t.Fatalf("a cached name match was served for the index symbol: %s (%s)", s.Symbol, s.Name)
	}

	// The control: the same search adopts the fund for a plain ticker, which
	// is the CLI's fuzzy convenience and stays.
	if s, err := c.Fetch(context.Background(), "ZZCOM", from); err != nil || s.Symbol != "ZZCOM.L" {
		t.Errorf("plain ticker: %v, %v", s, err)
	}
}
