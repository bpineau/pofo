package marketdata

import (
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

func TestFetchCPIUSLive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graph/fredgraph.csv", func(w http.ResponseWriter, r *http.Request) {
		if id := r.URL.Query().Get("id"); id != "CPIAUCNS" {
			t.Errorf("FRED id = %q, want CPIAUCNS", id)
		}
		fmt.Fprint(w, "observation_date,CPIAUCNS\n2024-01-01,308.417\n2024-02-01,310.326\n2024-03-01,312.332\n")
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	c.RefreshInflation = true // the live FRED path is refresh-only

	// Lowercase exercises the identifier canonicalization on the way in.
	s, err := c.Fetch(t.Context(), "^cpi-us", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Symbol != "^CPI-US" || s.Source != "fred" || s.Currency != "" {
		t.Fatalf("series misread: %+v", s)
	}
	if s.Name != "US CPI (all items, 1982-84=100)" {
		t.Errorf("name = %q", s.Name)
	}
	// Monthly anchors are interpolated to a daily curve: 31 January days,
	// 29 February days (2024 is a leap year), plus the final March anchor.
	if len(s.Points) != 61 {
		t.Fatalf("points = %d, want 61 (daily-interpolated)", len(s.Points))
	}
	if got := s.Last(); !got.Date.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)) || math.Abs(got.Close-312.332) > 1e-9 {
		t.Errorf("last = %+v, want 2024-03-01 at 312.332", got)
	}
}

func TestFetchCPIUSEmbedFirst(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graph/fredgraph.csv", func(w http.ResponseWriter, r *http.Request) {
		t.Error("FRED must not be hit: ^CPI-US is served offline-first from the embed")
		http.Error(w, "fred down", http.StatusInternalServerError)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(t.Context(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "fred" {
		t.Fatalf("source = %q, want fred (embedded snapshot)", s.Source)
	}
	first := s.First()
	if first.Date.Year() != 1913 || math.Abs(first.Close-9.8) > 1e-9 {
		t.Errorf("first = %+v, want January 1913 at 9.8 (long embedded history)", first)
	}
	if last := s.Last(); last.Close < 300 {
		t.Errorf("last close = %v, want a recent level (>300)", last.Close)
	}
}
