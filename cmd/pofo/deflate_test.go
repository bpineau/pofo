package main

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func months(n int) []time.Time {
	out := make([]time.Time, n)
	d := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range out {
		out[i] = d
		d = d.AddDate(0, 1, 0)
	}
	return out
}

func cpiSeries(dates []time.Time, levels ...float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: "^HICP-FR"}
	for i, l := range levels {
		s.Points = append(s.Points, marketdata.Point{Date: dates[i], Close: l})
	}
	return s
}

// A flat CPI leaves the series unchanged; a rising CPI erodes real value even
// when the nominal level is flat (real_t = nominal × CPI_base / CPI_t).
func TestDeflate(t *testing.T) {
	d := months(4)
	nominal := []float64{100, 100, 100, 100}

	flat := deflate(d, nominal, cpiSeries(d, 100, 100, 100, 100))
	for i, v := range flat {
		if math.Abs(v-100) > 1e-9 {
			t.Errorf("flat CPI: point %d = %v, want 100 (identity)", i, v)
		}
	}

	rising := deflate(d, nominal, cpiSeries(d, 100, 110, 120, 125))
	want := []float64{100, 100.0 * 100 / 110, 100.0 * 100 / 120, 100.0 * 100 / 125}
	for i := range want {
		if math.Abs(rising[i]-want[i]) > 1e-9 {
			t.Errorf("rising CPI: point %d = %.4f, want %.4f", i, rising[i], want[i])
		}
	}
}

// The point of the feature: with inflation, a nominal recovery is not a real
// one. A series that fully recovers in nominal terms can still be underwater in
// real terms, so the real max drawdown is deeper and the real TTR longer.
func TestRealDrawdownDeeperThanNominal(t *testing.T) {
	d := months(5)
	// Nominal: down 20% then back to the peak by the last month (recovers).
	nominal := []float64{100, 80, 85, 95, 100}
	// Prices rise 3% total over the window.
	cpi := cpiSeries(d, 100, 100.7, 101.4, 102.1, 103)

	nomStats, err := metrics.Compute(d, nominal)
	if err != nil {
		t.Fatal(err)
	}
	realStats, err := metrics.Compute(d, deflate(d, nominal, cpi))
	if err != nil {
		t.Fatal(err)
	}
	// Nominal fully recovers (drawdown ends); real still below its deflated peak.
	if realStats.MaxDrawdown >= nomStats.MaxDrawdown {
		t.Errorf("real max drawdown %.4f should be deeper (more negative) than nominal %.4f",
			realStats.MaxDrawdown, nomStats.MaxDrawdown)
	}
}

// A currency without a wired CPI yields no deflator (real columns are omitted).
// The gate returns before touching the client, so a nil client is safe here.
func TestInflationSeriesCurrencyGate(t *testing.T) {
	if _, ok := inflationSeries(context.Background(), nil, "JPY", time.Time{}); ok {
		t.Error("JPY should have no wired inflation index (yet)")
	}
	if _, ok := inflationSeries(context.Background(), nil, "", time.Time{}); ok {
		t.Error("empty currency (native) should have no single deflator")
	}
}

// A USD report deflates by the US CPI. The failing FRED stub exercises the
// embedded ^CPI-US snapshot, so the test stays offline.
func TestInflationSeriesUSD(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fred down", http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := marketdata.NewClient("")
	c.FredBase = srv.URL

	s, ok := inflationSeries(context.Background(), c, "USD", time.Time{})
	if !ok {
		t.Fatal("USD should deflate by ^CPI-US")
	}
	if s.Symbol != "^CPI-US" || s.First().Date.Year() != 1913 {
		t.Errorf("deflator = %s from %s, want ^CPI-US from 1913", s.Symbol, s.First().Date.Format("2006-01"))
	}
}
