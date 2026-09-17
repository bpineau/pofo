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

// openAnchorMux serves what an OPEN-anchored fund needs (ERES_DATADOG, whose
// NAV of a day is struck on its proxy's opening print): two NAV days, the
// proxy's daily bars with their open column, the flat cross, and the proxy's
// and the cross's intraday ticks of the session named by `session`. proxyDays
// bounds the daily bars, so a caller can choose whether a forward nowcast tail
// exists at all. With opens false the chart answers without an open column, the
// case every fallback rests on.
func openAnchorMux(t *testing.T, proxyDays int, session time.Time, opens bool) *http.ServeMux {
	t.Helper()
	days := testDays(proxyDays)
	london, _ := time.LoadLocation("Europe/London")
	mux := http.NewServeMux()
	mux.HandleFunc(airfundChartPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"fundName":"F","navs":[{"date":"2020-01-06","value":50},{"date":"2020-01-07","value":51}]}`)
	})
	mux.HandleFunc("/v8/finance/chart/DDOG", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("interval") == "5m" {
			fmt.Fprint(w, intradayJSON("USD", "America/New_York", session, []float64{110, 112}))
			return
		}
		closes := []float64{100, 102, 104, 106}[:proxyDays]
		if !opens {
			fmt.Fprint(w, chartJSON("DDOG", days, closes))
			return
		}
		// Each session opens 1 % below its close, the 2020-01-07 one at 101.
		fmt.Fprint(w, chartJSONOpens("DDOG", days, []float64{99, 101, 103, 105}[:proxyDays], closes))
	})
	mux.HandleFunc("/v8/finance/chart/USDEUR=X", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("interval") == "5m" {
			fmt.Fprint(w, intradayJSON("EUR", "Europe/London", session.In(london), []float64{0.95, 0.95}))
			return
		}
		fmt.Fprint(w, chartJSONCcy("USDEUR=X", "EUR", testDays(4), []float64{0.9, 0.9, 0.9, 0.9}))
	})
	return mux
}

// TestNowcastForwardAnchorsOnTheProxyOpen: a fund whose record says
// nowcast_anchor "open" carries the proxy's move since the OPEN of the last
// NAV's day, so the estimate leaves that session's open-to-close move out.
func TestNowcastForwardAnchorsOnTheProxyOpen(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	session := time.Date(2020, 1, 10, 9, 30, 0, 0, ny)
	c, srv := newTestClient(t, t.TempDir(), openAnchorMux(t, 4, session, true))
	defer srv.Close()
	s, err := c.Fetch(context.Background(), "ERES_DATADOG", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 4 || s.EstimateProxy != "DDOG" {
		t.Fatalf("want the 2 NAVs + 2 estimated days from DDOG, got %d points (%q)", len(s.Points), s.EstimateProxy)
	}
	// 51 × 104/101 and 51 × 106/101: the anchor is the open of 2020-01-07
	// (101), not its close (102), and the flat cross leaves the USD move.
	if !near(s.Points[2].Close, 51*104/101.0) || !near(s.Points[3].Close, 51*106/101.0) {
		t.Fatalf("estimated closes %v %v, want %v and %v",
			s.Points[2].Close, s.Points[3].Close, 51*104/101.0, 51*106/101.0)
	}
	// An ESTIMATED last daily value already stands on the proxy's close of its
	// own day, so today's path anchors there whatever the record says.
	in, err := c.Intraday(context.Background(), "ERES_DATADOG")
	if err != nil {
		t.Fatal(err)
	}
	base, anchor := 51*106/101.0, 106*0.9
	if len(in.Points) != 2 || !near(in.Points[0].Close, base*110*0.95/anchor) {
		t.Fatalf("intraday on an estimated anchor: %+v", in.Points)
	}
}

// TestNowcastIntradayAnchorsOnTheOpenOfAPublishedNAV: with the proxy's daily
// history stopping at the last NAV, today's path stands on that published NAV
// and must be anchored on the proxy's OPEN of the NAV's day.
func TestNowcastIntradayAnchorsOnTheOpenOfAPublishedNAV(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	session := time.Date(2020, 1, 8, 9, 30, 0, 0, ny)
	c, srv := newTestClient(t, t.TempDir(), openAnchorMux(t, 2, session, true))
	defer srv.Close()
	in, err := c.Intraday(context.Background(), "ERES_DATADOG")
	if err != nil {
		t.Fatal(err)
	}
	anchor := 101 * 0.9 // the open of 2020-01-07 in EUR, not its close
	if len(in.Points) != 2 || !near(in.Points[0].Close, 51*110*0.95/anchor) ||
		!near(in.Points[1].Close, 51*112*0.95/anchor) {
		t.Fatalf("ticks %+v, want the NAV scaled from the open anchor", in.Points)
	}
	q, err := c.Latest(context.Background(), "ERES_DATADOG")
	if err != nil || !near(q.Price, 51*112*0.95/anchor) {
		t.Fatalf("latest: %+v (%v)", q, err)
	}
}

// TestNowcastOpenAnchorFallsBackOnTheClose: a proxy served without an opening
// price (another source, a day it did not trade) must degrade to the
// close-anchored estimate, never fail.
func TestNowcastOpenAnchorFallsBackOnTheClose(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	session := time.Date(2020, 1, 10, 9, 30, 0, 0, ny)
	c, srv := newTestClient(t, t.TempDir(), openAnchorMux(t, 4, session, false))
	defer srv.Close()
	s, err := c.Fetch(context.Background(), "ERES_DATADOG", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	// 51 × 104/102 and 51 × 106/102: the close of 2020-01-07 anchors it.
	if len(s.Points) != 4 || !near(s.Points[2].Close, 52) || !near(s.Points[3].Close, 53) {
		t.Fatalf("points %+v, want the close-anchored 52 and 53", s.Points)
	}
	if _, ok := c.openAnchorFactor(context.Background(), "DDOG", d(2020, 1, 7)); ok {
		t.Error("a chart served without an open column must not yield a factor")
	}
}

// TestOpenAnchorFactorRejectsAMissingDay pins the two refusals the fallback
// rests on: a day the proxy did not trade, and a proxy with no Yahoo symbol.
func TestOpenAnchorFactorRejectsAMissingDay(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	session := time.Date(2020, 1, 10, 9, 30, 0, 0, ny)
	c, srv := newTestClient(t, t.TempDir(), openAnchorMux(t, 4, session, true))
	defer srv.Close()
	ctx := context.Background()
	if f, ok := c.openAnchorFactor(ctx, "DDOG", d(2020, 1, 7)); !ok || !near(f, 101/102.0) {
		t.Errorf("factor on a traded day = %v, %v; want 101/102", f, ok)
	}
	if _, ok := c.openAnchorFactor(ctx, "DDOG", d(2020, 1, 11)); ok {
		t.Error("a day the proxy did not trade must not yield a factor")
	}
	if _, ok := c.openAnchorFactor(ctx, "LU1234567890", d(2020, 1, 7)); ok {
		t.Error("an identifier with no known Yahoo symbol must not yield a factor")
	}
}

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

// TestWithoutEstimatesInvariants: the estimate stripper is what keeps a
// nowcast out of every stored or shipped dataset, so its contract is narrow -
// it never touches a real series, and it never leaves a stamp behind on a copy
// it did strip.
func TestWithoutEstimatesInvariants(t *testing.T) {
	var nilSeries *Series
	if nilSeries.WithoutEstimates() != nil {
		t.Error("a nil series must stay nil")
	}
	real := &Series{Symbol: "F", Points: []Point{
		{Date: d(2020, 1, 6), Close: 50}, {Date: d(2020, 1, 7), Close: 51},
	}}
	if got := real.WithoutEstimates(); got != real {
		t.Error("a series carrying no estimate must be returned as is, not copied")
	}
	est := &Series{Symbol: "F", EstimatedFrom: d(2020, 1, 8), EstimateProxy: "URTH", Points: []Point{
		{Date: d(2020, 1, 6), Close: 50},
		{Date: d(2020, 1, 7), Close: 51},
		{Date: d(2020, 1, 8), Close: 52},
		{Date: d(2020, 1, 9), Close: 53},
	}}
	got := est.WithoutEstimates()
	if len(got.Points) != 2 || !got.Last().Date.Equal(d(2020, 1, 7)) {
		t.Fatalf("points = %+v, want the two published days", got.Points)
	}
	if !got.EstimatedFrom.IsZero() || got.EstimateProxy != "" {
		t.Errorf("the stripped copy still claims an estimate: %+v", got)
	}
	if len(est.Points) != 4 || est.EstimateProxy != "URTH" {
		t.Error("the input series was mutated")
	}
	// A stamp with nothing after it strips to the whole series.
	head := &Series{EstimatedFrom: d(2030, 1, 1), Points: est.Points}
	if len(head.WithoutEstimates().Points) != 4 {
		t.Error("a stamp after the last point must keep every point")
	}
}

// TestIntradayRateAt pins the forward fill the tick-by-tick conversion rests
// on: a rate holds until the next print, and there is none before the first.
func TestIntradayRateAt(t *testing.T) {
	base := time.Date(2020, 1, 10, 14, 30, 0, 0, time.UTC)
	fx := &IntradaySeries{Points: []IntradayPoint{
		{Time: base, Close: 0.9},
		{Time: base.Add(10 * time.Minute), Close: 0.91},
	}}
	if _, ok := fx.rateAt(base.Add(-time.Minute)); ok {
		t.Error("there is no rate before the first tick")
	}
	if r, ok := fx.rateAt(base.Add(5 * time.Minute)); !ok || r != 0.9 {
		t.Errorf("forward fill = %v, %v; want 0.9, true", r, ok)
	}
	if r, ok := fx.rateAt(base.Add(time.Hour)); !ok || r != 0.91 {
		t.Errorf("after the last tick = %v, %v; want 0.91, true", r, ok)
	}
}
