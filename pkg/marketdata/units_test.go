package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A London line comes back from Yahoo in PENCE with the currency code "GBp",
// and from the Financial Times with "GBX". The payload shapes below are the
// real ones, trimmed: the chart meta of a London closed-end fund quoting
// around 3808 pence (£38.08), and the v7 quote the batch path reads.

// penceChartJSON is chartJSON with the London currency code and a symbol of
// its own, so the meta the client reads is the real "GBp" shape.
func penceChartJSON(symbol string, days []time.Time, closes []float64) string {
	return chartJSONCcy(symbol, "GBp", days, closes)
}

// TestPenceSeriesIsServedInPounds: the series a London line produces must be
// in pounds, labelled GBP. Served in pence under a code every case-insensitive
// comparison folds into "GBP", the whole history is a hundredfold out.
func TestPenceSeriesIsServedInPounds(t *testing.T) {
	days := testDays(5)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/PSH.L", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, penceChartJSON("PSH.L", days, []float64{3800, 3808, 3790, 3811, 3825}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.History(context.Background(), "PSH.L", days[0])
	if err != nil {
		t.Fatal(err)
	}
	if s.Currency != "GBP" {
		t.Fatalf("currency %q, want GBP", s.Currency)
	}
	if got := s.Last().Close; got != 38.25 {
		t.Fatalf("last close %v, want 38.25 (the pence price divided by a hundred)", got)
	}
}

// TestPenceSeriesPassesNativeGBPUnscaled is the finador case: a holding
// declared in GBP asks for its native line (FetchOptions.NoConvert). "GBp"
// satisfies that constraint - it IS the pound line - so the constraint must
// not be what protects the caller: the prices must arrive in pounds.
func TestPenceSeriesPassesNativeGBPUnscaled(t *testing.T) {
	days := testDays(5)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/PSH.L", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, penceChartJSON("PSH.L", days, []float64{3800, 3808, 3790, 3811, 3825}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.FetchExtended(context.Background(), "PSH.L", FetchOptions{
		From: days[0], Currency: "GBP", NoConvert: true,
	})
	if err != nil {
		t.Fatalf("a pence line is the GBP line and must answer: %v", err)
	}
	if s.Currency != "GBP" || s.Last().Close != 38.25 {
		t.Fatalf("native GBP fetch served %v %s, want 38.25 GBP", s.Last().Close, s.Currency)
	}
}

// TestPenceQuoteIsInPounds covers the live path, both legs: the chart meta
// spot and the v7 batch quote.
func TestPenceQuoteIsInPounds(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "ck"})
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("crumb1"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"quoteResponse":{"result":[
		 {"symbol":"PSH.L","currency":"GBp","exchangeTimezoneName":"Europe/London","regularMarketPrice":3808,"regularMarketTime":1782999000}]}}`)
	})
	mux.HandleFunc("/v8/finance/chart/BHMG.L", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[{"meta":{"currency":"GBp","exchangeTimezoneName":"Europe/London","regularMarketPrice":4210,"regularMarketTime":1782999000}}]}}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c := NewClient(t.TempDir())
	stubAllBases(c, srv.URL)

	if q := c.LatestBatchLive(context.Background(), []string{"PSH.L"})["PSH.L"]; q.Currency != "GBP" || q.Price != 38.08 {
		t.Fatalf("batch quote %v %s, want 38.08 GBP", q.Price, q.Currency)
	}
	q, err := c.Latest(context.Background(), "BHMG.L")
	if err != nil {
		t.Fatal(err)
	}
	if q.Currency != "GBP" || q.Price != 42.1 {
		t.Fatalf("spot quote %v %s, want 42.1 GBP", q.Price, q.Currency)
	}
}

// TestFXRateAcceptsSubUnits: FXRate is public and takes a currency code from
// whatever a caller holds. Uppercasing "GBp" makes it "GBP" and returns a rate
// a hundred times too large for a pence amount.
func TestFXRateAcceptsSubUnits(t *testing.T) {
	days := testDays(5)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/GBPEUR=X", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, chartJSONCcy("GBPEUR=X", "EUR", days, []float64{1.15, 1.15, 1.15, 1.15, 1.15}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	rate, err := c.FXRate(context.Background(), "GBp", "EUR", days[4])
	if err != nil {
		t.Fatal(err)
	}
	if want := 0.0115; rate != want {
		t.Fatalf("GBp→EUR rate %v, want %v (one penny is a hundredth of a pound)", rate, want)
	}
	if rate, err := c.FXRate(context.Background(), "GBp", "GBP", days[4]); err != nil || rate != 0.01 {
		t.Fatalf("GBp→GBP rate %v (%v), want 0.01", rate, err)
	}
}

// TestSubUnitCacheFileHeals: a cache file written before the rescaling still
// holds pence under its "GBp" label. Reading it back must apply the same
// rescaling, so an offline run and a live one agree.
func TestSubUnitCacheFileHeals(t *testing.T) {
	dir := t.TempDir()
	days := testDays(5)
	c := NewClient(dir)
	stubAllBases(c, "http://127.0.0.1:0")
	pence := &Series{Symbol: "PSH.L", Name: "Pershing Square Holdings", Currency: "GBp", Source: "yahoo"}
	for i, d := range days {
		pence.Points = append(pence.Points, Point{Date: d, Close: 3800 + float64(i)})
	}
	pence.Dividends = []Dividend{{Date: days[2], Amount: 25}}
	c.saveCache(pence, days[0])

	got, _, ok := c.loadCacheAnyAge("PSH.L", days[0])
	if !ok {
		t.Fatal("the cache file did not load")
	}
	if got.Currency != "GBP" || got.Last().Close != 38.04 || got.Dividends[0].Amount != 0.25 {
		t.Fatalf("stale pence cache served as %v %s (dividend %v), want pounds",
			got.Last().Close, got.Currency, got.Dividends[0].Amount)
	}
}

// TestDoctorNamesASubUnitLevel: the plausibility bands judge RATIOS and cannot
// see a hundredfold level error, so the identity pass must name it.
func TestDoctorNamesASubUnitLevel(t *testing.T) {
	days := testDays(40)
	s := &Series{Symbol: "BHMG.L", Name: "BH Macro Ltd GBP", Currency: "GBp", Source: "yahoo"}
	for i, d := range days {
		s.Points = append(s.Points, Point{Date: d, Close: 4200 + float64(i)})
	}
	issues := VerifyAsset("BHMG", s, days[len(days)-1])
	found := false
	for _, i := range issues {
		if strings.Contains(i.Message, "sub-unit of GBP") {
			found = true
		}
	}
	if !found {
		t.Fatalf("a pence-denominated series must be named by the doctor, got %v", issues)
	}
}
