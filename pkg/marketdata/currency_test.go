package marketdata

import (
	"fmt"
	"math"
	"net/http"
	"testing"
)

func TestConvertCurrency(t *testing.T) {
	days := testDays(4)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/USDEUR=X", func(w http.ResponseWriter, r *http.Request) {
		// FX disponible seulement à partir du 2e jour: extrapolation avant.
		fmt.Fprint(w, chartJSON("USDEUR=X", days[1:], []float64{0.90, 0.92, 0.94}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s := &Series{Symbol: "X", Currency: "USD"}
	for i, d := range days {
		s.Points = append(s.Points, Point{Date: d, Close: 100 + float64(i)})
	}
	out, extrap, err := c.ConvertCurrency(s, "EUR", days[0])
	if err != nil {
		t.Fatal(err)
	}
	if out.Currency != "EUR" {
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
