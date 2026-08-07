package marketdata

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

// stooqFXCSV is a Stooq daily-history fixture for the conventional eurusd
// listing (USD per EUR). The 1998 row predates the euro: it is a synthetic
// backcast on the provider side and must never survive into a series.
const stooqFXCSV = "Date,Open,High,Low,Close,Volume\n" +
	"1998-12-30,9.99,9.99,9.99,9.99,0\n" +
	"2020-01-06,1.11,1.12,1.10,1.1194,0\n" +
	"2020-01-07,1.11,1.12,1.10,1.1025,0\n"

func TestStooqSymbolFX(t *testing.T) {
	cases := []struct {
		symbol string
		ss     string
		invert bool
	}{
		// Conventional direction: served as-is.
		{"EURUSD=X", "eurusd", false},
		{"GBPUSD=X", "gbpusd", false},
		{"USDJPY=X", "usdjpy", false},
		{"EURGBP=X", "eurgbp", false},
		{"AUDUSD=X", "audusd", false},
		// Reciprocal direction: Stooq lists the conventional pair only.
		{"USDEUR=X", "eurusd", true},
		{"JPYUSD=X", "usdjpy", true},
		{"GBPEUR=X", "eurgbp", true},
		{"CHFEUR=X", "eurchf", true},
		// A minor currency ranks below every major.
		{"USDSEK=X", "usdsek", false},
		{"SEKUSD=X", "usdsek", true},
		// Not a mappable cross.
		{"SEKNOK=X", "", false}, // no major leg: unvetted on stooq
		{"EUREUR=X", "", false},
		{"FOOBA=X", "", false},
		// Non-FX symbols keep their existing mapping.
		{"VOO", "voo.us", false},
		{"CL=F", "cl.f", false},
		{"IWDA.AS", "", false},
	}
	for _, tc := range cases {
		ss, invert := stooqSymbol(tc.symbol)
		if ss != tc.ss || invert != tc.invert {
			t.Errorf("stooqSymbol(%q) = %q, %v; want %q, %v", tc.symbol, ss, invert, tc.ss, tc.invert)
		}
	}
}

func TestHistoryFXFallsBackToStooq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		if s := r.URL.Query().Get("s"); s != "eurusd" {
			t.Errorf("stooq symbol = %q, want eurusd (the conventional listing)", s)
		}
		fmt.Fprint(w, stooqFXCSV)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "USDEUR=X", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "stooq" || s.Currency != "EUR" {
		t.Fatalf("source/currency misread: %+v", s)
	}
	// USDEUR=X is the reciprocal of the eurusd listing.
	if got, want := s.Last().Close, 1/1.1025; math.Abs(got-want) > 1e-9 {
		t.Errorf("last close = %v, want %v (inverted eurusd)", got, want)
	}
	// The synthetic pre-euro row is dropped; the vetted bundled ECU/EUR
	// anchors take its place, so the series still reaches back to 1978.
	for _, p := range s.Points {
		if math.Abs(p.Close-1/9.99) < 1e-3 {
			t.Errorf("synthetic pre-euro stooq row survived: %+v", p)
		}
	}
	if first := s.First().Date; first.Year() != 1978 {
		t.Errorf("series starts %s, want 1978 (bundled ECU/EUR splice)", first.Format("2006-01"))
	}
}
