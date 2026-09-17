package marketdata

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
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
	// The bundled ECU/DM/EUR proxy still extends the euro cross to 1971.
	if first := s.First().Date; first.Year() != 1971 {
		t.Errorf("series starts %s, want 1971 (bundled ECU/DM/EUR splice)", first.Format("2006-01"))
	}
}

func TestHistoryFXCrossRateViaECB(t *testing.T) {
	// SEK carries no bundled proxy, so the ECB-derived cross stands alone
	// (EUR/GBP/JPY/CHF would each be spliced back to 1971 instead).
	const sekCSV = "Date,USD,SEK,\n" +
		"2020-01-07,1.1025,10.45,\n" +
		"2020-01-06,1.1194,10.50,\n" +
		"1999-01-04,1.1789,N/A,\n"
	c, srv := newTestClient(t, t.TempDir(), newECBOutageMux(t, sekCSV))
	defer srv.Close()

	s, err := c.History(context.Background(), "SEKUSD=X", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "ecb" || s.Currency != "USD" {
		t.Fatalf("source/currency misread: %+v", s)
	}
	// SEKUSD = (EUR→USD)/(EUR→SEK); the N/A SEK row must be skipped, so
	// only the two 2020 rows survive.
	if len(s.Points) != 2 {
		t.Fatalf("points = %d, want 2 (N/A row skipped): %+v", len(s.Points), s.Points)
	}
	if got, want := s.Last().Close, 1.1025/10.45; math.Abs(got-want) > 1e-9 {
		t.Errorf("last close = %v, want %v", got, want)
	}
}

func TestHistoryEuroCrossOfflineSnapshot(t *testing.T) {
	// Every live source down and nothing cached: the euro crosses still
	// answer from the bundled daily ECU/DM/EUR proxy (1971→).
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "everything down", http.StatusInternalServerError)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "USDEUR=X", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Currency != "EUR" {
		t.Fatalf("currency = %q, want EUR: %+v", s.Currency, s)
	}
	if first := s.First().Date; first.Year() != 1971 {
		t.Errorf("first = %s, want 1971 (bundled anchors)", first.Format("2006-01"))
	}
	// The first proxy row (1971-01-04, rescaled daily DM fixing) is
	// 0.731763 USD per EUR: the USDEUR snapshot serves its reciprocal.
	if got, want := s.Points[0].Close, 1/0.731763; math.Abs(got-want) > 1e-9 {
		t.Errorf("first close = %v, want %v", got, want)
	}
	// The December 1978 monthly ECU anchor (1.3773 USD per EUR) is
	// preserved exactly by the anchors+shape blend.
	if rate, _, ok := s.At(time.Date(1978, 12, 1, 0, 0, 0, 0, time.UTC)); !ok || math.Abs(rate-1/1.3773) > 1e-9 {
		t.Errorf("1978-12 rate = %v, %v; want %v", rate, ok, 1/1.3773)
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

// TestECBArchiveEdges: the reference-rate file is a zip of a CSV, and every
// way that can go wrong must surface as an error rather than as an empty
// series that would silently hold a conversion flat.
func TestECBArchiveEdges(t *testing.T) {
	var noCSV bytes.Buffer
	zw := zip.NewWriter(&noCSV)
	if f, err := zw.Create("readme.txt"); err != nil {
		t.Fatal(err)
	} else if _, err := f.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		body []byte
		err  string
	}{
		{"not a zip at all", []byte("<html>maintenance</html>"), "unreadable ecb archive"},
		{"a zip without a CSV", noCSV.Bytes(), "no CSV in the ecb archive"},
		{"a CSV with only a header", ecbZip(t, "Date,USD,\n"), "empty ecb CSV"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ecbRows(tc.body); err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("error = %v, want one about %q", err, tc.err)
			}
		})
	}
}

func TestECBColumnAndCrossGuards(t *testing.T) {
	header := []string{"Date", "USD", "JPY", "GBP", ""}
	// The euro is the implicit unit and needs no column.
	if rate, ok := ecbColumn(header, "EUR")(nil); !ok || rate != 1 {
		t.Errorf("EUR = %v, %v; want 1, true", rate, ok)
	}
	if ecbColumn(header, "SEK") != nil {
		t.Error("a currency the file does not carry must have no reader")
	}
	read := ecbColumn(header, "GBP")
	if _, ok := read([]string{"2020-01-07", "1.1"}); ok {
		t.Error("a row too short to reach the column must not read as a rate")
	}
	if _, ok := read([]string{"2020-01-07", "1.1", "120", "N/A", ""}); ok {
		t.Error("an N/A hole must not read as a rate")
	}
	if rate, ok := read([]string{"2020-01-07", "1.1", "120", " 0.85 ", ""}); !ok || rate != 0.85 {
		t.Errorf("padded rate = %v, %v; want 0.85, true", rate, ok)
	}

	// The source only ever serves currency crosses.
	c, srv := newTestClient(t, "", newECBOutageMux(t, ecbHistCSV))
	defer srv.Close()
	for _, symbol := range []string{"VOO", "EUREUR=X"} {
		if _, err := c.fetchECBFX(context.Background(), symbol, time.Time{}); err == nil ||
			!strings.Contains(err.Error(), "not a currency cross") {
			t.Errorf("fetchECBFX(%q) error = %v, want a refusal", symbol, err)
		}
	}
	// A cross the ECB does not publish is named as such.
	if _, err := c.fetchECBFX(context.Background(), "SEKNOK=X", time.Time{}); err == nil ||
		!strings.Contains(err.Error(), "does not publish") {
		t.Errorf("error = %v, want one naming the missing pair", err)
	}
}
