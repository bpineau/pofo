package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// chartJSONDiv mirrors chartJSON but with distinct raw closes and dividend
// events, like a real Yahoo chart payload for a distributing fund.
func chartJSONDiv(symbol string, days []time.Time, raw, adj []float64, divs map[time.Time]float64) string {
	ts, rawS, adjS := "", "", ""
	for i := range days {
		if i > 0 {
			ts, rawS, adjS = ts+",", rawS+",", adjS+","
		}
		ts += fmt.Sprint(days[i].Add(14*time.Hour + 30*time.Minute).Unix())
		rawS += fmt.Sprint(raw[i])
		adjS += fmt.Sprint(adj[i])
	}
	divS := ""
	for d, amount := range divs {
		if divS != "" {
			divS += ","
		}
		u := d.Add(14 * time.Hour).Unix()
		divS += fmt.Sprintf(`"%d":{"amount":%g,"date":%d}`, u, amount, u)
	}
	return fmt.Sprintf(`{"chart":{"result":[{"meta":{"currency":"USD","symbol":%q,"longName":"Div Fund"},"timestamp":[%s],"events":{"dividends":{%s}},"indicators":{"quote":[{"close":[%s]}],"adjclose":[{"adjclose":[%s]}]}}],"error":null}}`,
		symbol, ts, divS, rawS, adjS)
}

func TestFetchDividendsAndRawCloses(t *testing.T) {
	days := testDays(3)
	raw := []float64{100, 101, 99}
	adj := []float64{95, 96.5, 99} // dividend-adjusted history differs
	divs := map[time.Time]float64{days[1]: 0.75}
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/DVY", func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Query().Get("events") != "div" {
			t.Error("chart request should ask for dividend events")
		}
		fmt.Fprint(w, chartJSONDiv("DVY", days, raw, adj, divs))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	ctx := context.Background()

	// Default view: adjusted closes, dividends attached.
	s, err := c.Fetch(ctx, "DVY", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Points[0].Close != 95 {
		t.Errorf("default view should serve adjusted closes, got %v", s.Points[0].Close)
	}
	if len(s.Dividends) != 1 || s.Dividends[0].Amount != 0.75 || !s.Dividends[0].Date.Equal(days[1]) {
		t.Fatalf("dividends misread: %+v", s.Dividends)
	}

	// Raw view: unadjusted closes, dividends attached, separately cached.
	sr, err := c.FetchExtended(ctx, "DVY", FetchOptions{Raw: true})
	if err != nil {
		t.Fatal(err)
	}
	if sr.Points[0].Close != 100 {
		t.Errorf("raw view should serve unadjusted closes, got %v", sr.Points[0].Close)
	}
	if len(sr.Dividends) != 1 {
		t.Fatalf("raw view should carry dividends too: %+v", sr.Dividends)
	}

	// Both views cached: a fresh client on a dead server serves both.
	srv.Close()
	c2 := NewClient(dir)
	stubAllBases(c2, srv.URL)
	if s2, err := c2.Fetch(ctx, "DVY", time.Time{}); err != nil || s2.Points[0].Close != 95 || len(s2.Dividends) != 1 {
		t.Errorf("adjusted view from cache: %+v, %v", s2, err)
	}
	if s2, err := c2.FetchExtended(ctx, "DVY", FetchOptions{Raw: true}); err != nil || s2.Points[0].Close != 100 {
		t.Errorf("raw view from cache: %+v, %v", s2, err)
	}
	if requests != 2 {
		t.Errorf("expected 2 requests (one per view), counted %d", requests)
	}
}

func TestFetchExtendedRawRejectsSim(t *testing.T) {
	c := NewClient(t.TempDir())
	if _, err := c.FetchExtended(context.Background(), "VOOSIM", FetchOptions{Raw: true}); err == nil {
		t.Fatal("raw + SIM extension should be an error")
	}
}

func TestConvertCurrencyConvertsDividends(t *testing.T) {
	days := testDays(3)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/USDEUR=X", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("USDEUR=X", days, []float64{0.9, 0.9, 0.8}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s := &Series{Symbol: "DVY", Currency: "USD", Points: []Point{
		{Date: days[0], Close: 100},
		{Date: days[2], Close: 100},
	}, Dividends: []Dividend{
		{Date: days[0], Amount: 1.0},
		{Date: days[2], Amount: 2.0},
	}}
	out, _, err := c.ConvertCurrency(context.Background(), s, "EUR", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Points[0].Close != 90 || out.Points[1].Close != 80 {
		t.Fatalf("points misconverted: %+v", out.Points)
	}
	if out.Dividends[0].Amount != 0.9 || out.Dividends[1].Amount != 1.6 {
		t.Fatalf("dividends misconverted: %+v", out.Dividends)
	}
	// The original series must be untouched.
	if s.Dividends[0].Amount != 1.0 {
		t.Error("ConvertCurrency mutated its input")
	}
}

func TestTrimClipsDividends(t *testing.T) {
	days := testDays(5)
	s := &Series{Points: []Point{
		{Date: days[0], Close: 1}, {Date: days[2], Close: 2}, {Date: days[4], Close: 3},
	}, Dividends: []Dividend{
		{Date: days[0], Amount: 1}, {Date: days[2], Amount: 2}, {Date: days[4], Amount: 3},
	}}
	out := Trim(s, days[1], days[3])
	if len(out.Points) != 1 || len(out.Dividends) != 1 || out.Dividends[0].Amount != 2 {
		t.Fatalf("Trim should clip dividends with points: %+v / %+v", out.Points, out.Dividends)
	}
}

func TestOldCacheFileStillLoads(t *testing.T) {
	// A pre-dividends cache file (no "raw", no "dividends" keys) must keep
	// serving adjusted reads.
	days := testDays(2)
	dir := t.TempDir()
	mux := http.NewServeMux()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	c.saveCache(&Series{Symbol: "OLD", Name: "Old", Currency: "USD", Source: "yahoo",
		Points: []Point{{Date: days[0], Close: 5}, {Date: days[1], Close: 6}}}, time.Time{})

	s, err := c.History(context.Background(), "OLD", time.Time{})
	if err != nil || len(s.Points) != 2 {
		t.Fatalf("old cache unreadable: %+v, %v", s, err)
	}
}
