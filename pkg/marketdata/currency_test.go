package marketdata

import (
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

func TestConvertCurrency(t *testing.T) {
	days := testDays(4)
	mux := http.NewServeMux()
	// CHF has no bundled long proxy, so this exercises the raw extrapolation
	// path (the euro cross is separately backfilled, see TestExtendFXBack).
	mux.HandleFunc("/v8/finance/chart/USDCHF=X", func(w http.ResponseWriter, r *http.Request) {
		// FX disponible seulement à partir du 2e jour: extrapolation avant.
		fmt.Fprint(w, chartJSON("USDCHF=X", days[1:], []float64{0.90, 0.92, 0.94}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s := &Series{Symbol: "X", Currency: "USD"}
	for i, d := range days {
		s.Points = append(s.Points, Point{Date: d, Close: 100 + float64(i)})
	}
	out, extrap, err := c.ConvertCurrency(s, "CHF", days[0])
	if err != nil {
		t.Fatal(err)
	}
	if out.Currency != "CHF" {
		t.Errorf("currency = %s", out.Currency)
	}
	want := []float64{100 * 0.90, 101 * 0.90, 102 * 0.92, 103 * 0.94}
	for i := range want {
		if math.Abs(out.Points[i].Close-want[i]) > 1e-9 {
			t.Errorf("point %d = %v, want %v", i, out.Points[i].Close, want[i])
		}
	}
	if !extrap.Equal(days[1]) {
		t.Errorf("extrapolatedBefore = %v, want %v", extrap, days[1])
	}
	// L'original est intact (séries partagées par mémoïsation).
	if s.Currency != "USD" || s.Points[0].Close != 100 {
		t.Error("input series mutated")
	}
	// Même devise: renvoyée telle quelle, sans requête FX.
	if same, _, err := c.ConvertCurrency(s, "USD", days[0]); err != nil || same != s {
		t.Errorf("same-currency conversion should be identity: %v", err)
	}
	// GBp: division par 100 puis conversion… ici cible GBP = pas de FX.
	pence := &Series{Symbol: "P", Currency: "GBp", Points: []Point{{Date: days[0], Close: 250}}}
	gbp, _, err := c.ConvertCurrency(pence, "GBP", days[0])
	if err != nil || gbp.Points[0].Close != 2.5 || gbp.Currency != "GBP" {
		t.Errorf("GBp→GBP: %+v, %v", gbp, err)
	}
}

// TestExtendFXBack checks the bundled long EUR/USD proxy: it splices behind the
// euro cross in both directions, reaches back to the late 1970s (ECU era), and
// USDEUR is the exact reciprocal of EURUSD.
func TestExtendFXBack(t *testing.T) {
	eurusd, ok := eurusdLongCross("EURUSD=X")
	if !ok || len(eurusd) == 0 {
		t.Fatal("EURUSD=X has no bundled long proxy")
	}
	if first := eurusd[0].Date; first.Year() > 1979 {
		t.Errorf("long EUR/USD starts %s, want the ECU era (≤1979)", first.Format("2006-01"))
	}
	for i := 1; i < len(eurusd); i++ {
		if !eurusd[i].Date.After(eurusd[i-1].Date) {
			t.Fatalf("proxy not strictly ascending at %s", eurusd[i].Date.Format("2006-01"))
		}
	}
	usdeur, ok := eurusdLongCross("USDEUR=X")
	if !ok || len(usdeur) != len(eurusd) {
		t.Fatalf("USDEUR=X proxy len %d, want %d", len(usdeur), len(eurusd))
	}
	if got := usdeur[0].Close * eurusd[0].Close; math.Abs(got-1) > 1e-9 {
		t.Errorf("USDEUR·EURUSD = %v, want 1 (reciprocal)", got)
	}
	if _, ok := eurusdLongCross("USDJPY=X"); ok {
		t.Error("only the euro cross should carry a bundled proxy")
	}

	// Splice behind a short recent EURUSD=X: it gains the pre-quote history.
	s := &Series{Symbol: "EURUSD=X", Points: []Point{
		{Date: time.Date(2010, 1, 4, 0, 0, 0, 0, time.UTC), Close: 1.44},
		{Date: time.Date(2010, 1, 5, 0, 0, 0, 0, time.UTC), Close: 1.43},
	}}
	extendFXBack("EURUSD=X", s)
	if !s.First().Date.Before(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("after splice EURUSD=X starts %s, want pre-1980", s.First().Date.Format("2006-01"))
	}
	if s.SimulatedBefore.IsZero() {
		t.Error("splice should mark SimulatedBefore")
	}
}

// The FX cross must be fetched once and reused across assets, even though each
// asset passes its own first date: the report converts many USD holdings to EUR,
// and a per-asset FX download made runs take over a minute.
func TestConvertCurrencyFetchesFXOncePerRun(t *testing.T) {
	days := testDays(60)
	closes := make([]float64, 60)
	for i := range closes {
		closes[i] = 0.9 // EUR per USD
	}
	fxRequests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		fxRequests++
		fmt.Fprint(w, chartJSON("USDEUR=X", days, closes))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	usdAsset := func(sym string, startDay int) *Series {
		s := &Series{Symbol: sym, Currency: "USD"}
		for i := 0; i < 10; i++ {
			s.Points = append(s.Points, Point{Date: days[0].AddDate(0, 0, startDay+i), Close: 100})
		}
		return s
	}

	// Two USD assets with different first dates -> different caller `from`.
	if _, _, err := c.ConvertCurrency(usdAsset("A", 0), "EUR", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.ConvertCurrency(usdAsset("B", 5), "EUR", time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	if fxRequests != 1 {
		t.Errorf("FX cross fetched %d times, want 1 (cached across assets)", fxRequests)
	}
}
