package marketdata

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// spotJSON builds a Yahoo chart-meta fixture carrying a live regular-market
// price. A zero price omits the field, to exercise the not-covered path.
func spotJSON(currency, tz string, price float64, at time.Time) string {
	priceField := ""
	if price > 0 {
		priceField = fmt.Sprintf(`"regularMarketPrice":%v,`, price)
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":%q,"exchangeTimezoneName":%q,%s"regularMarketTime":%d}}],"error":null}}`,
		currency, tz, priceField, at.Unix())
}

func TestLatestLiveSpot(t *testing.T) {
	at := time.Date(2024, 3, 1, 18, 0, 0, 0, time.UTC) // 13:00 New York
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", 501.25, at))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest(t.Context(), "VOO")
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Price != 501.25 || q.Currency != "USD" || q.Source != "yahoo" {
		t.Fatalf("quote misread: %+v", q)
	}
	if h := q.Time.Hour(); h != 13 {
		t.Errorf("quote hour = %d, want 13 (exchange local time)", h)
	}
}

func TestLatestStripsSIM(t *testing.T) {
	at := time.Date(2024, 3, 1, 18, 0, 0, 0, time.UTC)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", 501.25, at))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest(t.Context(), "VOOSIM")
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Price != 501.25 {
		t.Fatalf("SIM id should quote as its base: %+v", q)
	}
}

func TestLatestFallsBackToClose(t *testing.T) {
	days := testDays(3)
	closes := []float64{10, 11, 12.5}
	mux := http.NewServeMux()
	// The chart fixture carries a daily series but no regularMarketPrice, so the
	// spot step is not covered and Latest falls through to the last close.
	mux.HandleFunc("/v8/finance/chart/SPY", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("SPY", days, closes))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest(t.Context(), "SPY")
	if err != nil {
		t.Fatal(err)
	}
	if q.Live {
		t.Errorf("expected a non-live (daily close) quote, got Live=true")
	}
	if q.Price != 12.5 || q.Source != "yahoo" {
		t.Fatalf("quote misread: %+v", q)
	}
	if !q.Time.Equal(days[2]) {
		t.Errorf("quote time = %v, want last close date %v", q.Time, days[2])
	}
}

func TestLatestSurvivesYahooOutage(t *testing.T) {
	mux := http.NewServeMux()
	// Yahoo down across the board (chart and search): the spot step fails and
	// the daily-close fallback rides the Stooq leg of the Fetch chain.
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/finance/search", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "yahoo down", http.StatusInternalServerError)
	})
	mux.HandleFunc("/q/d/l/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Date,Open,High,Low,Close,Volume\n2020-01-06,1,1,1,42.5,100\n2020-01-07,1,1,1,43,100\n")
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	q, err := c.Latest(t.Context(), "XYZ")
	if err != nil {
		t.Fatal(err)
	}
	if q.Live || q.Source != "stooq" || q.Price != 43 {
		t.Fatalf("quote should degrade to the last stooq close: %+v", q)
	}
}

func TestFetchYahooSpotMissingPriceNotCovered(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/NOPX", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, spotJSON("USD", "America/New_York", 0, time.Now())) // price omitted
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	_, err := c.fetchYahooSpot(t.Context(), "NOPX")
	if !errors.Is(err, ErrNotCovered) {
		t.Fatalf("err = %v, want ErrNotCovered", err)
	}
}
