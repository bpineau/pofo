package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// An Offline client serves the disk cache whatever its age and depth, then
// the bundled data, and never sends a request: every one would fail the test.
func TestOfflineServesWhatIsLocal(t *testing.T) {
	days := testDays(80)
	closes := make([]float64, len(days))
	for i := range closes {
		closes[i] = 100 + float64(i)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, closes))
	})
	dir := t.TempDir()
	online, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	ctx := context.Background()
	if _, err := online.Fetch(ctx, "VOO", days[0]); err != nil {
		t.Fatal(err)
	}

	requests := 0
	offMux := http.NewServeMux()
	offMux.HandleFunc("/", func(http.ResponseWriter, *http.Request) { requests++ })
	c, offSrv := newTestClient(t, dir, offMux)
	defer offSrv.Close()
	c.Offline = true
	c.MaxAge = time.Nanosecond // the cached file is stale: served all the same

	// Deeper than the download and stale: the cache is all there is.
	s, err := c.FetchExtended(ctx, "VOO", FetchOptions{From: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if s.Len() != len(days) || s.Last().Close != closes[len(closes)-1] {
		t.Errorf("cached VOO served as %d points, last %v", s.Len(), s.Last())
	}

	// Nothing local: a clear error, naming the identifier.
	_, err = c.Fetch(ctx, "NOSUCHTICKER", days[0])
	if !errors.Is(err, ErrOffline) {
		t.Errorf("uncached ticker: error %v, want ErrOffline", err)
	}
	if _, err := c.Fetch(ctx, "IE00B4L5Y983", days[0]); !errors.Is(err, ErrOffline) {
		t.Errorf("uncached ISIN: error %v, want ErrOffline", err)
	}
	if _, err := c.Intraday(ctx, "VOO"); err == nil {
		t.Error("Intraday is live-only, and must fail offline")
	}

	// Bundled data answers without a cache: a catalog index, and the backcast
	// behind a SIM identifier whose real quotes were never downloaded.
	idx, err := c.FetchExtended(ctx, "MSCIWORLD", FetchOptions{})
	if err != nil || idx.Source != "index" || idx.Len() < 1000 {
		t.Errorf("bundled index: %v, %v", idx, err)
	}
	sim, err := c.FetchExtended(ctx, "DBMFSIM", FetchOptions{})
	if err != nil || sim.ProxySymbol != "simdata" || sim.Len() < 1000 {
		t.Errorf("bundled backcast: %v, %v", sim, err)
	}

	if requests != 0 {
		t.Errorf("an Offline client sent %d requests", requests)
	}
}

// A SIM request served from its bundled history alone (the real quotes
// unreadable) keeps the currency contract of the quotes it stands in for:
// the catalog's currency stated, converted into a requested one, refused
// under NoConvert when that is not its own.
func TestSimOnlyFallbackHonoursCurrency(t *testing.T) {
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(http.ResponseWriter, *http.Request) { requests++ })
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	c.Offline = true
	ctx := context.Background()

	usd, err := c.FetchExtended(ctx, "DBMFSIM", FetchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if usd.Currency != "USD" || usd.ProxySymbol != "simdata" {
		t.Errorf("native fallback: currency %q proxy %q, want USD from the catalog, simdata", usd.Currency, usd.ProxySymbol)
	}
	eur, err := c.FetchExtended(ctx, "DBMFSIM", FetchOptions{Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	if eur.Currency != "EUR" || eur.Len() != usd.Len() || eur.Last().Close == usd.Last().Close {
		t.Errorf("EUR fallback: currency %q, %d points, last %v vs %v: want a converted series",
			eur.Currency, eur.Len(), eur.Last().Close, usd.Last().Close)
	}
	if _, err := c.FetchExtended(ctx, "DBMFSIM", FetchOptions{Currency: "EUR", NoConvert: true}); !errors.Is(err, ErrWrongCurrency) {
		t.Errorf("NoConvert into EUR: err %v, want ErrWrongCurrency", err)
	}
	if requests != 0 {
		t.Errorf("%d requests from an offline client", requests)
	}
}
