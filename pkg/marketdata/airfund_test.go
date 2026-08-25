package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// airfundJSON is a chart payload in the delivery API's shape: an unsorted
// pair of NAV days plus one row without a value, as the API can emit.
const airfundJSON = `{"fundName":"ERES XTRACKERS ACTIONS MONDE","abbreviatedShareName":null,
"benchmarkName":"75% X & 25% Y","navs":[
{"date":"2024-03-06","displayDate":"06/03/2024","value":49.99,"displayValue":"49,99€"},
{"date":"2024-03-05","displayDate":"05/03/2024","value":50,"displayValue":"50,00€"},
{"date":"2024-03-07","displayDate":"07/03/2024","value":null,"displayValue":""}]}`

func TestParseAirfundNAVs(t *testing.T) {
	s, err := parseAirfundNAVs([]byte(airfundJSON))
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "ERES XTRACKERS ACTIONS MONDE" || s.Source != "airfund" {
		t.Fatalf("header: %+v", s)
	}
	if len(s.Points) != 2 || s.Points[0].Close != 50 || s.Points[1].Close != 49.99 {
		t.Fatalf("points: %+v (want the two valued days, sorted)", s.Points)
	}
	if s.Points[0].Date != time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("first date %s", s.Points[0].Date)
	}
	if _, err := parseAirfundNAVs([]byte(`{"fundName":"x","navs":[]}`)); err == nil {
		t.Fatal("an empty payload must be an error")
	}
}

// TestAirfundCatalogFund exercises the whole route of a catalog "airfund"
// fund: the pinned resolution, the POST the API expects (share code and
// widget id, with a 201 answer), the catalog currency and name, and the
// on-disk cache that spares the second call.
func TestAirfundCatalogFund(t *testing.T) {
	var got map[string]any
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc(airfundChartPath, func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("payload: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, airfundJSON)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s, err := c.Fetch(context.Background(), "ERESMONDEM", from)
	if err != nil {
		t.Fatal(err)
	}
	if got["isinCode"] != "990000135629" || got["sId"] != "41481ca4-919c-46c0-9ca1-41a880ff4e8e" || got["maxPeriodCode"] != "inception" {
		t.Fatalf("request payload: %v", got)
	}
	if s.Source != "airfund" || s.Currency != "EUR" || s.Symbol != "ERESMONDEM" || len(s.Points) != 2 {
		t.Fatalf("series: %+v", s)
	}
	if s.Name == "" || s.Name == "ERES XTRACKERS ACTIONS MONDE" {
		t.Fatalf("the catalog name should win over the API's: %q", s.Name)
	}
	// The alias (the company's share code) reaches the same fund.
	if s2, err := c.Fetch(context.Background(), "990000135629", from); err != nil || s2.Symbol != "ERESMONDEM" {
		t.Fatalf("alias: %v %+v", err, s2)
	}
	if calls != 1 {
		t.Fatalf("the second fetch should come from the cache, API called %d times", calls)
	}
}

// TestAirfundOfflineFallsBackOnEmbeddedNAV: with the API down and nothing
// cached, the bundled refdata snapshot answers, as for a CPI deflator.
func TestAirfundOfflineFallsBackOnEmbeddedNAV(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(airfundChartPath, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	s, err := c.Fetch(context.Background(), "ERESMONDEM", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "airfund" || s.Currency != "EUR" || len(s.Points) < 500 {
		t.Fatalf("embedded fallback: %s %s %d points", s.Source, s.Currency, len(s.Points))
	}
	if first := s.First(); first.Date != time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC) || first.Close != 50 {
		t.Fatalf("the snapshot must start at the launch NAV of 50.00 on 2024-03-05, got %v", first)
	}
	if _, ok := embeddedNAV("VOO"); ok {
		t.Fatal("only airfund funds carry an embedded NAV")
	}
}

// TestAirfundRawFallsBackToEmbeddedNAV: with the API unreachable, a RAW fetch
// (what a valuation consumer uses) must still serve the bundled NAV snapshot,
// not fall through to a live ticker search. Regression for the "~raw" cache
// key missing the catalog lookup.
func TestAirfundRawFallsBackToEmbeddedNAV(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(airfundChartPath, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusBadGateway)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, raw := range []bool{false, true} {
		s, err := c.FetchExtended(context.Background(), "ERESMONDEM", FetchOptions{From: from, NoSim: true, Raw: raw})
		if err != nil {
			t.Fatalf("raw=%v: %v (should fall back to the embedded NAV)", raw, err)
		}
		if s.Source != "airfund" || len(s.Points) < 500 {
			t.Fatalf("raw=%v: %s %d points, want the embedded snapshot", raw, s.Source, len(s.Points))
		}
	}
}

// TestAirfundHasNoYahooIntraday: a fund with no listing must report
// ErrNotCovered rather than ask Yahoo for a made-up symbol.
func TestAirfundHasNoYahooIntraday(t *testing.T) {
	c := NewClient("")
	if _, ok := c.yahooSymbol(context.Background(), "ERESMONDEM"); ok {
		t.Fatal("an airfund fund is not a Yahoo symbol")
	}
	if _, ok := c.yahooSymbol(context.Background(), "VOO"); !ok {
		t.Fatal("a plain ticker still is")
	}
}
