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
