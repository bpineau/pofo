package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// chartJSONCcy is chartJSON with an explicit quote currency.
func chartJSONCcy(symbol, currency string, days []time.Time, closes []float64) string {
	return strings.Replace(chartJSON(symbol, days, closes),
		`"currency":"USD"`, fmt.Sprintf(`"currency":%q`, currency), 1)
}

func linear(n int, base float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = base + float64(i)
	}
	return out
}

// twinMux serves an ISIN whose Yahoo search returns a deep USD twin and a
// shallower native EUR listing.
func twinMux(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[
			{"symbol":"TWIN.US","longname":"Twin Fund","quoteType":"ETF"},
			{"symbol":"NATV.PA","longname":"Twin Fund","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/TWIN.US", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("TWIN.US", "USD", testDays(200), linear(200, 50)))
	})
	mux.HandleFunc("/v8/finance/chart/NATV.PA", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("NATV.PA", "EUR", testDays(100), linear(100, 40)))
	})
	return mux
}

func TestNoConvertPrefersNativeListing(t *testing.T) {
	const isin = "FR0000000001"
	c, srv := newTestClient(t, t.TempDir(), twinMux(t))
	defer srv.Close()
	from := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)

	// Unconstrained: depth wins, the USD twin is served (current behaviour).
	s, err := c.FetchExtended(context.Background(), isin, FetchOptions{From: from})
	if err != nil || s.Symbol != "TWIN.US" {
		t.Fatalf("unconstrained fetch: %+v, %v", s, err)
	}

	// NoConvert EUR: the native line wins despite its shallower history.
	c2, srv2 := newTestClient(t, t.TempDir(), twinMux(t))
	defer srv2.Close()
	s, err = c2.FetchExtended(context.Background(), isin,
		FetchOptions{From: from, Currency: "EUR", NoConvert: true})
	if err != nil || s.Symbol != "NATV.PA" || s.Currency != "EUR" {
		t.Fatalf("NoConvert fetch: %+v, %v", s, err)
	}
}

func TestNoConvertFailsWithoutNativeLine(t *testing.T) {
	const isin = "FR0000000002"
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"TWIN.US","longname":"Twin Fund","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/TWIN.US", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("TWIN.US", "USD", testDays(200), linear(200, 50)))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	from := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := c.FetchExtended(context.Background(), isin,
		FetchOptions{From: from, Currency: "EUR", NoConvert: true})
	if !errors.Is(err, ErrWrongCurrency) {
		t.Fatalf("want ErrWrongCurrency, got: %v", err)
	}
}

func TestNoConvertBypassesOffCurrencyCachedResolution(t *testing.T) {
	const isin = "FR0000000003"
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, twinMux(t))
	from := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)

	// Adopt the USD twin unconstrained; the resolution is cached on disk.
	if s, err := c.Fetch(context.Background(), isin, from); err != nil || s.Symbol != "TWIN.US" {
		t.Fatalf("seeding fetch: %+v, %v", s, err)
	}
	srv.Close()

	// A NoConvert EUR call on a fresh client (same cache dir) must NOT
	// reuse the cached USD resolution: it re-resolves, restricted.
	c2, srv2 := newTestClient(t, dir, twinMux(t))
	defer srv2.Close()
	s, err := c2.FetchExtended(context.Background(), isin,
		FetchOptions{From: from, Currency: "EUR", NoConvert: true})
	if err != nil || s.Symbol != "NATV.PA" {
		t.Fatalf("cached twin resolution not bypassed: %+v, %v", s, err)
	}
}
