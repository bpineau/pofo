package marketdata

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestConvertCurrency(t *testing.T) {
	days := testDays(4)
	mux := http.NewServeMux()
	// SEK has no bundled long proxy, so this exercises the raw extrapolation
	// path (the bundled crosses are separately backfilled, see TestExtendFXBack).
	mux.HandleFunc("/v8/finance/chart/USDSEK=X", func(w http.ResponseWriter, r *http.Request) {
		// FX disponible seulement à partir du 2e jour: extrapolation avant.
		fmt.Fprint(w, chartJSON("USDSEK=X", days[1:], []float64{0.90, 0.92, 0.94}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s := &Series{Symbol: "X", Currency: "USD"}
	for i, d := range days {
		s.Points = append(s.Points, Point{Date: d, Close: 100 + float64(i)})
	}
	out, extrap, err := c.ConvertCurrency(context.Background(), s, "SEK", days[0])
	if err != nil {
		t.Fatal(err)
	}
	if out.Currency != "SEK" {
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
	if same, _, err := c.ConvertCurrency(context.Background(), s, "USD", days[0]); err != nil || same != s {
		t.Errorf("same-currency conversion should be identity: %v", err)
	}
	// GBp: division par 100 puis conversion… ici cible GBP = pas de FX.
	pence := &Series{Symbol: "P", Currency: "GBp", Points: []Point{{Date: days[0], Close: 250}}}
	gbp, _, err := c.ConvertCurrency(context.Background(), pence, "GBP", days[0])
	if err != nil || gbp.Points[0].Close != 2.5 || gbp.Currency != "GBP" {
		t.Errorf("GBp→GBP: %+v, %v", gbp, err)
	}
}

// TestExtendFXBack checks every bundled long FX proxy: each splices behind its
// cross in both directions, reaches back to 1971, ascends strictly in time, and
// the USD<CCY> direction is the exact reciprocal of <CCY>USD.
func TestExtendFXBack(t *testing.T) {
	for _, ccy := range []string{"EUR", "GBP", "JPY", "CHF"} {
		direct, tag, ok := longFXCross(ccy + "USD=X")
		if !ok || len(direct) == 0 {
			t.Fatalf("%sUSD=X has no bundled long proxy", ccy)
		}
		if tag == "" {
			t.Errorf("%sUSD=X proxy has no provenance tag", ccy)
		}
		if first := direct[0].Date; first.Year() > 1971 {
			t.Errorf("long %s/USD starts %s, want 1971", ccy, first.Format("2006-01"))
		}
		for i := 1; i < len(direct); i++ {
			if !direct[i].Date.After(direct[i-1].Date) {
				t.Fatalf("%s proxy not strictly ascending at %s", ccy, direct[i].Date.Format("2006-01-02"))
			}
		}
		recip, _, ok := longFXCross("USD" + ccy + "=X")
		if !ok || len(recip) != len(direct) {
			t.Fatalf("USD%s=X proxy len %d, want %d", ccy, len(recip), len(direct))
		}
		if got := recip[0].Close * direct[0].Close; math.Abs(got-1) > 1e-9 {
			t.Errorf("USD%s·%sUSD = %v, want 1 (reciprocal)", ccy, ccy, got)
		}
	}
	if _, _, ok := longFXCross("USDSEK=X"); ok {
		t.Error("only bundled crosses should carry a proxy")
	}
	if _, _, ok := longFXCross("GBPSEK=X"); ok {
		t.Error("a pair with an unbundled leg should carry no proxy")
	}

	// Cross-euro pairs triangulate through the dollar: GBPEUR = GBP/USD ÷
	// EUR/USD, reaching 1971, and EURGBP is its reciprocal on shared dates
	// (the two grids only fully coincide once both legs are FRED H.10, so
	// reciprocity is checked on the latest common date, not by index).
	for _, ccy := range []string{"GBP", "JPY", "CHF"} {
		xeur, _, ok := longFXCross(ccy + "EUR=X")
		if !ok || len(xeur) == 0 {
			t.Fatalf("%sEUR=X has no triangulated proxy", ccy)
		}
		if first := xeur[0].Date; first.Year() > 1971 {
			t.Errorf("long %s/EUR starts %s, want 1971", ccy, first.Format("2006-01"))
		}
		eurx, _, ok := longFXCross("EUR" + ccy + "=X")
		if !ok || len(eurx) == 0 {
			t.Fatalf("EUR%s=X has no triangulated proxy", ccy)
		}
		last := (&Series{Points: xeur}).Last()
		recip, _, found := (&Series{Points: eurx}).At(last.Date)
		if !found {
			t.Fatalf("EUR%s missing %s", ccy, last.Date.Format("2006-01-02"))
		}
		if got := last.Close * recip; math.Abs(got-1) > 1e-9 {
			t.Errorf("%sEUR·EUR%s on %s = %v, want 1 (reciprocal)", ccy, ccy, last.Date.Format("2006-01-02"), got)
		}
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
	if _, _, err := c.ConvertCurrency(context.Background(), usdAsset("A", 0), "EUR", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.ConvertCurrency(context.Background(), usdAsset("B", 5), "EUR", time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	if fxRequests != 1 {
		t.Errorf("FX cross fetched %d times, want 1 (cached across assets)", fxRequests)
	}
}

func TestFXRate(t *testing.T) {
	days := []time.Time{
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SEKUSD=X", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("SEKUSD=X", days, []float64{1.26, 1.28}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	ctx := context.Background()

	if r, err := c.FXRate(ctx, "EUR", "EUR", time.Now()); err != nil || r != 1 {
		t.Errorf("same-currency rate: %v, %v", r, err)
	}
	// Forward fill: a Saturday uses Friday's cross.
	at := time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)
	r, err := c.FXRate(ctx, "sek", "usd", at)
	if err != nil || r != 1.28 {
		t.Errorf("forward-filled rate: %v, %v", r, err)
	}
	// Before the cross starts: an explicit error. (SEK carries no bundled
	// proxy, unlike EUR/GBP/JPY/CHF which extend back to 1971.)
	if _, err := c.FXRate(ctx, "SEK", "USD", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Error("a date before the FX history should error")
	}
}

// TestConvertCurrencyNoops: every shape the conversion must hand back
// untouched, so that a caller can route every asset through it.
func TestConvertCurrencyNoops(t *testing.T) {
	c := NewClient("")
	cases := []struct {
		name   string
		s      *Series
		target string
	}{
		{"no target currency", &Series{Currency: "USD", Points: []Point{{Date: d(2020, 1, 6), Close: 1}}}, ""},
		{"blanks are not a target", &Series{Currency: "USD", Points: []Point{{Date: d(2020, 1, 6), Close: 1}}}, "   "},
		{"the series currency is unknown", &Series{Points: []Point{{Date: d(2020, 1, 6), Close: 1}}}, "EUR"},
		{"no points to convert", &Series{Currency: "USD"}, "EUR"},
		{"already the target", &Series{Currency: "EUR", Points: []Point{{Date: d(2020, 1, 6), Close: 1}}}, "eur "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, extrap, err := c.ConvertCurrency(context.Background(), tc.s, tc.target, time.Time{})
			if err != nil || out != tc.s || !extrap.IsZero() {
				t.Errorf("out=%p (want %p) extrap=%v err=%v", out, tc.s, extrap, err)
			}
		})
	}
}

// TestConvertCurrencyPenceDividends: a GBp line is scaled to pounds, and its
// dividends must follow, or the income of a pence-quoted class comes out a
// hundred times too large.
func TestConvertCurrencyPenceDividends(t *testing.T) {
	c := NewClient("")
	pence := &Series{Symbol: "P", Currency: "GBp",
		Points:    []Point{{Date: d(2020, 1, 6), Close: 250}, {Date: d(2020, 1, 7), Close: 260}},
		Dividends: []Dividend{{Date: d(2020, 1, 7), Amount: 5}},
	}
	out, _, err := c.ConvertCurrency(context.Background(), pence, "GBP", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Points[1].Close != 2.6 || len(out.Dividends) != 1 || out.Dividends[0].Amount != 0.05 {
		t.Errorf("GBp scaling: points=%+v dividends=%+v", out.Points, out.Dividends)
	}
	if pence.Points[1].Close != 260 || pence.Dividends[0].Amount != 5 {
		t.Error("the input series was mutated")
	}
}

// TestConvertCurrencyExtrapolatesDividendsToo: a dividend paid before the FX
// history begins converts at the oldest known rate, like the points around it.
func TestConvertCurrencyExtrapolatesDividendsToo(t *testing.T) {
	days := testDays(4)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/USDSEK=X", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("USDSEK=X", days[2:], []float64{9, 9.1}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s := &Series{Symbol: "X", Currency: "USD",
		Points:    []Point{{Date: days[0], Close: 100}, {Date: days[3], Close: 110}},
		Dividends: []Dividend{{Date: days[0], Amount: 1}, {Date: days[3], Amount: 2}},
	}
	// The window starts after the series: the FX fetch must still reach back
	// to the first quote rather than the requested date.
	out, extrap, err := c.ConvertCurrency(context.Background(), s, "SEK", days[2])
	if err != nil {
		t.Fatal(err)
	}
	if !extrap.Equal(days[2]) {
		t.Errorf("extrapolatedBefore = %v, want %v", extrap, days[2])
	}
	if math.Abs(out.Dividends[0].Amount-1*9) > 1e-9 || math.Abs(out.Dividends[1].Amount-2*9.1) > 1e-9 {
		t.Errorf("dividends = %+v, want [9 18.2]", out.Dividends)
	}
}

// TestConvertCurrencyReportsAnFXOutage: a cross no source can serve is an
// error, never a series quietly left in its own currency.
func TestConvertCurrencyReportsAnFXOutage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusInternalServerError)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	s := &Series{Symbol: "X", Currency: "USD", Points: []Point{{Date: d(2020, 1, 6), Close: 100}}}
	if _, _, err := c.ConvertCurrency(context.Background(), s, "SEK", time.Time{}); err == nil ||
		!strings.Contains(err.Error(), "FX rate USD") {
		t.Fatalf("error = %v, want an FX failure naming the cross", err)
	}
	if _, err := c.FXRate(context.Background(), "USD", "SEK", d(2020, 1, 6)); err == nil ||
		!strings.Contains(err.Error(), "FX rate USD") {
		t.Fatalf("FXRate error = %v, want an FX failure naming the cross", err)
	}
}

// TestLongFXCrossRefusals: the bundled splice only ever touches crosses it
// can actually extend.
func TestLongFXCrossRefusals(t *testing.T) {
	for _, symbol := range []string{"VOO", "USDUSD=X", "SEKNOK=X", "USDSEK=X"} {
		if _, _, ok := longFXCross(symbol); ok {
			t.Errorf("longFXCross(%q) claimed a bundled proxy", symbol)
		}
	}
	if _, _, ok := longFXCross("GBPJPY=X"); !ok {
		t.Error("two bundled legs must triangulate through the dollar")
	}
	// A nil series is not a crash.
	extendFXBack("EURUSD=X", nil)
}
