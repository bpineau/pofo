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

func TestLatestAnySkipsOffCurrencyQuote(t *testing.T) {
	const isin = "FR0000000005"
	c, srv := newTestClient(t, t.TempDir(), twinMuxISINOnlyUSD(t))
	defer srv.Close()

	q, err := c.LatestAny(context.Background(), []string{isin, "NATV.PA"},
		QuoteOptions{Currency: "EUR"})
	if err != nil || q.Currency != "EUR" || q.Symbol != "NATV.PA" {
		t.Fatalf("LatestAny native-first: %+v, %v", q, err)
	}
}

func TestLatestAnyNoConvertRejects(t *testing.T) {
	const isin = "FR0000000006"
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"TWIN.US","longname":"Twin Fund","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/TWIN.US", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("TWIN.US", "USD", testDays(200), linear(200, 50)))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	_, err := c.LatestAny(context.Background(), []string{isin},
		QuoteOptions{Currency: "EUR", NoConvert: true})
	if !errors.Is(err, ErrWrongCurrency) {
		t.Fatalf("want ErrWrongCurrency, got: %v", err)
	}
}

func TestLatestAnyConvertsAsLastResort(t *testing.T) {
	const isin = "FR0000000007"
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"TWIN.US","longname":"Twin Fund","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/TWIN.US", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("TWIN.US", "USD", testDays(200), linear(200, 50)))
	})
	// Serve the FX cross under both spellings so the test does not depend
	// on FXRate's internal symbol choice.
	mux.HandleFunc("/v8/finance/chart/USDEUR=X", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("USDEUR=X", "EUR", testDays(400), constant(400, 0.5)))
	})
	mux.HandleFunc("/v8/finance/chart/EURUSD=X", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSONCcy("EURUSD=X", "USD", testDays(400), constant(400, 2)))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.LatestAny(context.Background(), []string{isin}, QuoteOptions{Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	// Last USD close is 50+199 = 249; at 0.5 EUR per USD: 124.5.
	if q.Currency != "EUR" || q.Price != 124.5 {
		t.Fatalf("converted quote: %+v", q)
	}
}

func constant(n int, v float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = v
	}
	return out
}

// liveVsCloseMux serves an ISIN that only the daily-close path can answer
// (its resolved twin publishes history but no regular-market field) next to
// a ticker with a live spot: the shape of a security whose ISIN is quoted
// nowhere live while its own listing trades.
func liveVsCloseMux(t *testing.T, resolved string, spot float64, at time.Time) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"quotes":[{"symbol":%q,"longname":"Resolved Fund","quoteType":"ETF"}]}`, resolved)
	})
	mux.HandleFunc("/v8/finance/chart/"+resolved, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, chartJSON(resolved, testDays(200), linear(200, 50)))
	})
	mux.HandleFunc("/v8/finance/chart/LIVE", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", spot, at))
	})
	return mux
}

// A live quote outranks id order: the ISIN comes first and answers, but only
// with yesterday's close, while the ticker has a real market price. Taking
// the first answer is what left a -20% session invisible for hours.
func TestLatestAnyPrefersLiveOverEarlierClose(t *testing.T) {
	at := time.Date(2024, 3, 1, 18, 0, 0, 0, time.UTC)
	c, srv := newTestClient(t, t.TempDir(), liveVsCloseMux(t, "TWIN.US", 501.25, at))
	defer srv.Close()

	q, err := c.LatestAny(t.Context(), []string{"LU0000000009", "LIVE"}, QuoteOptions{Currency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Price != 501.25 || q.Symbol != "LIVE" {
		t.Fatalf("want the live ticker quote, got %+v", q)
	}
}

// With no live answer anywhere, authority decides again: the first id keeps
// the quote. A bare ticker can resolve to a same-currency twin listing of the
// real instrument, so promoting it on liveness grounds alone would silently
// price the wrong fund.
func TestLatestAnyKeepsFirstCloseWhenNothingIsLive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"quotes":[{"symbol":"AUTH.US","longname":"Authoritative Fund","quoteType":"ETF"}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/AUTH.US", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, chartJSON("AUTH.US", testDays(200), linear(200, 50)))
	})
	mux.HandleFunc("/v8/finance/chart/TWIN", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, chartJSON("TWIN", testDays(200), linear(200, 900)))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.LatestAny(t.Context(), []string{"LU0000000009", "TWIN"}, QuoteOptions{Currency: "USD"})
	if err != nil {
		t.Fatal(err)
	}
	if q.Live || q.Symbol != "AUTH.US" {
		t.Fatalf("want the authoritative ISIN close, got %+v", q)
	}
}
