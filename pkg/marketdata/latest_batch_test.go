package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestLatestBatch: two Yahoo symbols served by one v7 call; a symbol the
// batch does not return falls back to the per-id spot path.
func TestLatestBatch(t *testing.T) {
	var batchCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "ck"})
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("crumb1"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		batchCalls.Add(1)
		if r.URL.Query().Get("crumb") != "crumb1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if got := r.URL.Query().Get("symbols"); !strings.Contains(got, "AAPL") || !strings.Contains(got, "MC.PA") {
			t.Errorf("symbols=%q", got)
		}
		fmt.Fprint(w, `{"quoteResponse":{"result":[
		 {"symbol":"AAPL","currency":"USD","exchangeTimezoneName":"America/New_York","regularMarketPrice":308.63,"regularMarketTime":1782999000},
		 {"symbol":"MC.PA","currency":"EUR","exchangeTimezoneName":"Europe/Paris","regularMarketPrice":495.7,"regularMarketTime":1783092989}]}}`)
	})
	mux.HandleFunc("/v8/finance/chart/ORPHAN.PA", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[{"meta":{"currency":"EUR","exchangeTimezoneName":"Europe/Paris","regularMarketPrice":10.5,"regularMarketTime":1783092989}}]}}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(t.TempDir())
	c.CookieBase, c.ChartBase, c.SearchBase = srv.URL, srv.URL, srv.URL

	got := c.LatestBatch(context.Background(), []string{"AAPL", "MC.PA", "ORPHAN.PA"})
	if len(got) != 3 {
		t.Fatalf("got %d quotes, want 3: %#v", len(got), got)
	}
	if q := got["AAPL"]; q.Price != 308.63 || q.Currency != "USD" || !q.Live || q.Source != "yahoo" {
		t.Fatalf("AAPL quote: %+v", q)
	}
	if q := got["MC.PA"]; q.Price != 495.7 || q.Currency != "EUR" {
		t.Fatalf("MC.PA quote: %+v", q)
	}
	if q := got["ORPHAN.PA"]; q.Price != 10.5 {
		t.Fatalf("fallback quote: %+v", q)
	}
	if n := batchCalls.Load(); n != 1 {
		t.Fatalf("batch endpoint hit %d times, want 1", n)
	}
}

// TestLatestBatchCrumbRenewal: a stale crumb gets 401, the client renews the
// auth pair once and retries the batch.
func TestLatestBatchCrumbRenewal(t *testing.T) {
	var crumbServes atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "ck"})
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		if crumbServes.Add(1) == 1 {
			_, _ = w.Write([]byte("stale"))
			return
		}
		_, _ = w.Write([]byte("fresh"))
	})
	mux.HandleFunc("/v7/finance/quote", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("crumb") != "fresh" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"quoteResponse":{"result":[
		 {"symbol":"AAPL","currency":"USD","exchangeTimezoneName":"America/New_York","regularMarketPrice":300,"regularMarketTime":1782999000}]}}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient(t.TempDir())
	c.CookieBase, c.ChartBase = srv.URL, srv.URL

	got := c.LatestBatch(context.Background(), []string{"AAPL"})
	if q, ok := got["AAPL"]; !ok || q.Price != 300 {
		t.Fatalf("after renewal: %#v", got)
	}
	if n := crumbServes.Load(); n != 2 {
		t.Fatalf("crumb served %d times, want 2 (stale then fresh)", n)
	}
}
