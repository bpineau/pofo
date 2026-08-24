package marketdata

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

// nowcastMux serves the whole universe a nowcast needs: the fund's two NAV
// days (airfund), the proxy's daily closes in USD reaching two days further,
// the USD→EUR daily cross, and the proxy's and the cross's intraday ticks on
// the day after that. Daily and intraday share a chart path and differ by the
// interval parameter, as on Yahoo.
func nowcastMux(t *testing.T) *http.ServeMux {
	t.Helper()
	days := testDays(4) // 2020-01-06 .. 2020-01-09
	ny, _ := time.LoadLocation("America/New_York")
	london, _ := time.LoadLocation("Europe/London")
	session := time.Date(2020, 1, 10, 9, 30, 0, 0, ny)
	mux := http.NewServeMux()
	mux.HandleFunc(airfundChartPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"fundName":"F","navs":[{"date":"2020-01-06","value":50},{"date":"2020-01-07","value":51}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/URTH", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("interval") == "5m" {
			fmt.Fprint(w, intradayJSON("USD", "America/New_York", session, []float64{106, 108}))
			return
		}
		fmt.Fprint(w, chartJSON("URTH", days, []float64{100, 102, 104, 106}))
	})
	mux.HandleFunc("/v8/finance/chart/USDEUR=X", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("interval") == "5m" {
			fmt.Fprint(w, intradayJSON("EUR", "Europe/London", session.In(london), []float64{0.95, 0.95}))
			return
		}
		fmt.Fprint(w, chartJSONCcy("USDEUR=X", "EUR", days, []float64{0.9, 0.9, 0.9, 0.9}))
	})
	return mux
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// TestNowcastForwardExtendsToTheProxyClose: the days after the last NAV
// carry the proxy's EUR returns, are flagged, and never reach the cache or
// WithoutEstimates' reader.
func TestNowcastForwardExtendsToTheProxyClose(t *testing.T) {
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, nowcastMux(t))
	defer srv.Close()
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	s, err := c.Fetch(context.Background(), "ERESMONDEM", from)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 4 {
		t.Fatalf("want the 2 NAVs + 2 estimated days, got %d points: %+v", len(s.Points), s.Points)
	}
	// 51 × (104/102) and 51 × (106/102): the flat cross leaves the USD move.
	if !near(s.Points[2].Close, 52) || !near(s.Points[3].Close, 53) {
		t.Fatalf("estimated closes %v %v, want 52 and 53", s.Points[2].Close, s.Points[3].Close)
	}
	if s.EstimatedFrom != time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC) || s.EstimateProxy != "URTH" {
		t.Fatalf("estimate stamp: from %s via %q", s.EstimatedFrom, s.EstimateProxy)
	}
	real := s.WithoutEstimates()
	if len(real.Points) != 2 || !real.EstimatedFrom.IsZero() || real.Last().Close != 51 {
		t.Fatalf("WithoutEstimates: %+v", real)
	}
	if len(s.Points) != 4 {
		t.Fatal("WithoutEstimates must not mutate the series it is called on")
	}
	// The disk cache holds the published NAVs only.
	cached, ok := c.loadCache("ERESMONDEM", from)
	if !ok || len(cached.Points) != 2 {
		t.Fatalf("cache: ok=%v %d points, want the 2 NAVs", ok, len(cached.Points))
	}
	// A series with no estimate returns itself.
	if v := cached.WithoutEstimates(); v != cached {
		t.Fatal("WithoutEstimates should return the same series when nothing is estimated")
	}
}

// TestNowcastIntradayScalesTheProxyPath: today's path is the last daily
// value (here the forward estimate of the previous close) scaled by the
// proxy's intraday USD move converted at the intraday cross.
func TestNowcastIntradayScalesTheProxyPath(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), nowcastMux(t))
	defer srv.Close()
	s, err := c.Intraday(context.Background(), "ERESMONDEM")
	if err != nil {
		t.Fatal(err)
	}
	if !s.Estimate || s.Proxy != "URTH" || s.Currency != "EUR" || s.Source != "nowcast" || s.Symbol != "ERESMONDEM" {
		t.Fatalf("header: %+v", s)
	}
	if len(s.Points) != 2 {
		t.Fatalf("want 2 ticks, got %d", len(s.Points))
	}
	// Base 53 (2020-01-09 estimate) on the anchor 106 × 0.9; ticks 106 and
	// 108 USD at a 0.95 cross.
	anchor := 106 * 0.9
	if !near(s.Points[0].Close, 53*106*0.95/anchor) || !near(s.Points[1].Close, 53*108*0.95/anchor) {
		t.Fatalf("ticks %v %v", s.Points[0].Close, s.Points[1].Close)
	}
}

// TestLatestQuotesTheNowcastTick: the latest quote of such a fund is the last
// intraday estimate, marked live and sourced "nowcast".
func TestLatestQuotesTheNowcastTick(t *testing.T) {
	c, srv := newTestClient(t, t.TempDir(), nowcastMux(t))
	defer srv.Close()
	q, err := c.Latest(context.Background(), "ERESMONDEM")
	if err != nil {
		t.Fatal(err)
	}
	if !q.Live || q.Source != "nowcast" || q.Currency != "EUR" || !near(q.Price, 53*108*0.95/(106*0.9)) {
		t.Fatalf("latest: %+v", q)
	}
}

// TestNowcastSurvivesAMissingProxy: with the proxy unreachable the daily
// series simply ends at the last NAV, and the intraday path reports
// ErrNotCovered like any unlisted instrument.
func TestNowcastSurvivesAMissingProxy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(airfundChartPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"fundName":"F","navs":[{"date":"2020-01-06","value":50},{"date":"2020-01-07","value":51}]}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	s, err := c.Fetch(context.Background(), "ERESMONDEM", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || len(s.Points) != 2 || !s.EstimatedFrom.IsZero() {
		t.Fatalf("daily without proxy: %v %+v", err, s)
	}
	if _, err := c.Intraday(context.Background(), "ERESMONDEM"); err == nil {
		t.Fatal("intraday without a reachable proxy must fail")
	}
}
