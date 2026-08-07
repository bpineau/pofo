package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFetchAnyTriesIdsInOrder(t *testing.T) {
	mux := http.NewServeMux()
	// The ISIN is unknown everywhere; the ticker answers directly.
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[]}`)
	})
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", testDays(100), linear(100, 300)))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	from := time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.FetchAny(context.Background(), []string{"LU0000000009", "VOO"}, FetchOptions{From: from})
	if err != nil || s.Symbol != "VOO" {
		t.Fatalf("FetchAny: %+v, %v", s, err)
	}
}

func TestFetchAnyJoinsErrors(t *testing.T) {
	mux := http.NewServeMux() // nothing answers anything
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	_, err := c.FetchAny(context.Background(), []string{"LU0000000009", "NOPE"}, FetchOptions{})
	if err == nil || !strings.Contains(err.Error(), "LU0000000009") || !strings.Contains(err.Error(), "NOPE") {
		t.Fatalf("joined errors should name every id: %v", err)
	}
}

// twinMuxISINOnlyUSD serves an ISIN whose search only finds the deep USD
// twin, while the native EUR line stays quotable by its own ticker.
func twinMuxISINOnlyUSD(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"TWIN.US","longname":"Twin Fund","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/TWIN.US", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("TWIN.US", "USD", testDays(200), linear(200, 50)))
	})
	mux.HandleFunc("/v8/finance/chart/NATV.PA", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("NATV.PA", "EUR", testDays(100), linear(100, 40)))
	})
	return mux
}

func TestFetchAnyPrefersNativeAcrossIds(t *testing.T) {
	// The ISIN resolves to the deep USD twin only; the declared ticker
	// serves the native EUR line. With Currency set (and no NoConvert),
	// the native answer must win without conversion.
	const isin = "FR0000000004"
	c, srv := newTestClient(t, t.TempDir(), twinMuxISINOnlyUSD(t))
	defer srv.Close()
	from := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)

	s, err := c.FetchAny(context.Background(), []string{isin, "NATV.PA"},
		FetchOptions{From: from, Currency: "EUR"})
	if err != nil || s.Symbol != "NATV.PA" || s.Currency != "EUR" {
		t.Fatalf("native-first: %+v, %v", s, err)
	}
}
