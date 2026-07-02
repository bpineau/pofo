package marketdata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// intradayJSON builds a Yahoo 5-minute chart fixture for one trading day.
func intradayJSON(currency, tz string, base time.Time, closes []float64) string {
	ts, cl := "", ""
	for i, c := range closes {
		if i > 0 {
			ts += ","
			cl += ","
		}
		ts += fmt.Sprint(base.Add(time.Duration(i) * 5 * time.Minute).Unix())
		cl += fmt.Sprint(c)
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":%q,"exchangeTimezoneName":%q},"timestamp":[%s],"indicators":{"quote":[{"close":[%s]}]}}],"error":null}}`,
		currency, tz, ts, cl)
}

func TestIntradayParse(t *testing.T) {
	base := time.Date(2024, 3, 1, 14, 30, 0, 0, time.UTC) // 09:30 New York
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("interval"); got != "5m" {
			t.Errorf("interval = %q, want 5m", got)
		}
		if got := r.URL.Query().Get("range"); got != "1d" {
			t.Errorf("range = %q, want 1d", got)
		}
		fmt.Fprint(w, intradayJSON("USD", "America/New_York", base, []float64{500, 501, 502}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Intraday(context.Background(), "VOO")
	if err != nil {
		t.Fatal(err)
	}
	if s.Currency != "USD" || s.Source != "yahoo" || len(s.Points) != 3 {
		t.Fatalf("series misread: %+v", s)
	}
	if s.Last().Close != 502 {
		t.Errorf("last close = %v, want 502", s.Last().Close)
	}
	if h := s.First().Time.Hour(); h != 9 {
		t.Errorf("first point hour = %d, want 9 (exchange local time)", h)
	}
}

func TestIntradayUnknownISINNotCovered(t *testing.T) {
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { requests++ })
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	_, err := c.Intraday(context.Background(), "FR0000000000")
	if !errors.Is(err, ErrNotCovered) {
		t.Fatalf("err = %v, want ErrNotCovered", err)
	}
	if requests != 0 {
		t.Errorf("intraday made %d requests resolving an unknown ISIN, want 0", requests)
	}
}
