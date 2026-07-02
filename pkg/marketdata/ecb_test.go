package marketdata

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

// ecbZip wraps a eurofxref-hist CSV body in the zip archive the ECB serves.
func ecbZip(t *testing.T, csv string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("eurofxref-hist.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(csv)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// ecbHistCSV mirrors the real file: rates are EUR against each currency,
// newest first, with a trailing comma and N/A holes.
const ecbHistCSV = "Date,USD,JPY,GBP,\n" +
	"2020-01-07,1.1025,120.55,0.85005,\n" +
	"2020-01-06,1.1194,121.60,0.85285,\n" +
	"1999-01-04,1.1789,133.73,N/A,\n"

// stooqChallengeHTML models the anti-bot page Stooq serves non-browser
// clients: an HTML body that must fail the CSV sniff, not poison a series.
const stooqChallengeHTML = "<!DOCTYPE html><html><body>This site requires JavaScript.</body></html>"

// newECBOutageMux stubs a full Yahoo outage, a challenged Stooq and a
// healthy ECB endpoint serving the given eurofxref-hist CSV.
func newECBOutageMux(t *testing.T, csv string) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, stooqChallengeHTML)
	})
	zipBody := ecbZip(t, csv)
	mux.HandleFunc("/stats/eurofxref/eurofxref-hist.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipBody)
	})
	return mux
}

func TestHistoryFXFallsBackToECB(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), newECBOutageMux(t, ecbHistCSV))
	defer srv.Close()

	s, err := c.History(context.Background(), "USDEUR=X", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "ecb" || s.Currency != "EUR" {
		t.Fatalf("source/currency misread: %+v", s)
	}
	// USDEUR is the reciprocal of the published EUR→USD rate.
	if got, want := s.Last().Close, 1/1.1025; math.Abs(got-want) > 1e-9 {
		t.Errorf("last close = %v, want %v", got, want)
	}
	// Rows come newest first from the ECB: the series must be ascending.
	day := time.Date(1999, 1, 4, 0, 0, 0, 0, time.UTC)
	if rate, _, ok := s.At(day); !ok || math.Abs(rate-1/1.1789) > 1e-9 {
		t.Errorf("1999-01-04 rate = %v, %v; want %v", rate, ok, 1/1.1789)
	}
	// The bundled ECU/EUR anchors still extend the euro cross to 1978.
	if first := s.First().Date; first.Year() != 1978 {
		t.Errorf("series starts %s, want 1978 (bundled ECU/EUR splice)", first.Format("2006-01"))
	}
}

func TestHistoryFXCrossRateViaECB(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), newECBOutageMux(t, ecbHistCSV))
	defer srv.Close()

	s, err := c.History(context.Background(), "GBPUSD=X", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "ecb" || s.Currency != "USD" {
		t.Fatalf("source/currency misread: %+v", s)
	}
	// GBPUSD = (EUR→USD)/(EUR→GBP); the N/A GBP row must be skipped, so
	// only the two 2020 rows survive (no bundled splice off the euro).
	if len(s.Points) != 2 {
		t.Fatalf("points = %d, want 2 (N/A row skipped): %+v", len(s.Points), s.Points)
	}
	if got, want := s.Last().Close, 1.1025/0.85005; math.Abs(got-want) > 1e-9 {
		t.Errorf("last close = %v, want %v", got, want)
	}
}

func TestLatestFXSurvivesStooqChallenge(t *testing.T) {
	// Latest must reach the ECB leg when Yahoo is down and Stooq serves its
	// anti-bot page. Rows are recent so the one-year Latest window keeps them.
	d1 := time.Now().UTC().AddDate(0, 0, -4).Format("2006-01-02")
	d2 := time.Now().UTC().AddDate(0, 0, -3).Format("2006-01-02")
	csv := "Date,USD,JPY,GBP,\n" +
		d2 + ",1.1025,120.55,0.85005,\n" +
		d1 + ",1.1194,121.60,0.85285,\n"
	c, srv := newTestClient(t, t.TempDir(), newECBOutageMux(t, csv))
	defer srv.Close()

	q, err := c.Latest(t.Context(), "USDEUR=X")
	if err != nil {
		t.Fatal(err)
	}
	if q.Live || q.Source != "ecb" || q.Currency != "EUR" {
		t.Fatalf("quote should degrade to the ECB reference rate: %+v", q)
	}
	if want := 1 / 1.1025; math.Abs(q.Price-want) > 1e-9 {
		t.Errorf("price = %v, want %v", q.Price, want)
	}
}
