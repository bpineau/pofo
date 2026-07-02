package marketdata

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

// cboeVIXCSV is a CBOE VIX_History fixture: OHLC daily rows, MM/DD/YYYY.
const cboeVIXCSV = "DATE,OPEN,HIGH,LOW,CLOSE\n" +
	"01/02/1990,17.24,17.24,17.24,17.24\n" +
	"02/05/2018,17.31,38.80,15.66,37.32\n" +
	"07/01/2026,17.11,17.30,15.97,16.59\n"

func TestHistoryVIXFallsBackToCBOE(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/api/global/us_indices/daily_prices/VIX_History.csv", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, cboeVIXCSV)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "^VIX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "cboe" || s.Currency != "" {
		t.Fatalf("source/currency misread: %+v", s)
	}
	if len(s.Points) != 3 {
		t.Fatalf("points = %d, want 3: %+v", len(s.Points), s.Points)
	}
	first := s.First()
	if !first.Date.Equal(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)) || math.Abs(first.Close-17.24) > 1e-9 {
		t.Errorf("first = %+v, want 1990-01-02 at 17.24 (normalized UTC midnight)", first)
	}
	// The 2018 volmageddon spike must survive: it is data, not a bad print.
	if got := s.Points[1].Close; math.Abs(got-37.32) > 1e-9 {
		t.Errorf("2018-02-05 close = %v, want 37.32", got)
	}
}

func TestHistoryVIXOfflineSnapshot(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "everything down", http.StatusInternalServerError)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "^VIX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	first := s.First()
	if !first.Date.Equal(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)) || math.Abs(first.Close-17.24) > 1e-9 {
		t.Errorf("first = %+v, want the snapshot start 1990-01-02 at 17.24", first)
	}
	if len(s.Points) < 9000 {
		t.Errorf("points = %d, want the full embedded history (>9000)", len(s.Points))
	}
}
