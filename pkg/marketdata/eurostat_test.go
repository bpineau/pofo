package marketdata

import (
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"
)

func TestHICPSymbol(t *testing.T) {
	cases := []struct {
		in     string
		geo    string
		wantOK bool
	}{
		{"^HICP-FR", "FR", true},
		{"^HICP-EA", "EA", true},
		{"^HICP-DE", "DE", true},
		{"^IRX", "", false},
		{"HICP-FR", "", false}, // missing the ^ marker
		{"^HICP-", "", false},  // missing geo
		{"VOO", "", false},
	}
	for _, c := range cases {
		geo, ok := hicpGeo(c.in)
		if ok != c.wantOK || geo != c.geo {
			t.Errorf("hicpGeo(%q) = (%q, %v), want (%q, %v)", c.in, geo, ok, c.geo, c.wantOK)
		}
	}
}

func TestMonthlyToDailyGeometric(t *testing.T) {
	jan := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC)
	anchors := []Point{{Date: jan, Close: 100}, {Date: feb, Close: 102}}

	daily := monthlyToDaily(anchors)

	// January has 31 days, so the segment spans 31 daily steps plus the
	// final February anchor: 32 points.
	if len(daily) != 32 {
		t.Fatalf("got %d daily points, want 32", len(daily))
	}
	if !daily[0].Date.Equal(jan) || daily[0].Close != 100 {
		t.Errorf("first point = %v, want {2006-01-01, 100}", daily[0])
	}
	last := daily[len(daily)-1]
	if !last.Date.Equal(feb) || math.Abs(last.Close-102) > 1e-9 {
		t.Errorf("last point = %v, want {2006-02-01, 102}", last)
	}
	// Geometric spread: day k carries 100 * (102/100)^(k/31).
	for k, p := range daily {
		want := 100 * math.Pow(102.0/100.0, float64(k)/31)
		if math.Abs(p.Close-want) > 1e-9 {
			t.Errorf("day %d close = %.10f, want %.10f", k, p.Close, want)
		}
		if k > 0 && !p.Date.After(daily[k-1].Date) {
			t.Errorf("dates not strictly ascending at %d: %v then %v", k, daily[k-1].Date, p.Date)
		}
		if k > 0 && p.Close <= daily[k-1].Close {
			t.Errorf("values not strictly increasing at %d", k)
		}
	}
}

func TestFetchEurostatHICP(t *testing.T) {
	const path = "/eurostat/api/dissemination/statistics/1.0/data/prc_hicp_midx"
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if g := r.URL.Query().Get("geo"); g != "FR" {
			t.Errorf("geo query = %q, want FR", g)
		}
		fmt.Fprint(w, `{
			"value": {"0": 100.0, "1": 101.0},
			"dimension": {"time": {"category": {"index": {"2006-01": 0, "2006-02": 1}}}}
		}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch("^HICP-FR", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "eurostat" {
		t.Errorf("source = %q, want eurostat", s.Source)
	}
	if s.Currency != "" {
		t.Errorf("currency = %q, want empty (index)", s.Currency)
	}
	if s.Symbol != "^HICP-FR" {
		t.Errorf("symbol = %q, want ^HICP-FR", s.Symbol)
	}
	if got := s.First(); !got.Date.Equal(time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)) || got.Close != 100 {
		t.Errorf("first = %v, want {2006-01-01, 100}", got)
	}
	if got := s.Last(); !got.Date.Equal(time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC)) || math.Abs(got.Close-101) > 1e-9 {
		t.Errorf("last = %v, want {2006-02-01, 101}", got)
	}
}
