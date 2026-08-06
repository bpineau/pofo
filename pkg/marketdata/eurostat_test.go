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

func TestParseHICPSnapshot(t *testing.T) {
	const csv = `# a comment line
# another

2006-01,100.5
2006-02,101
2006-03,bad
2006-04,102.25
`
	pts := parseMonthlyAnchors(csv)
	if len(pts) != 3 {
		t.Fatalf("got %d points, want 3 (comments/blank/unparsable skipped)", len(pts))
	}
	if !pts[0].Date.Equal(time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)) || pts[0].Close != 100.5 {
		t.Errorf("first = %v, want {2006-01-01, 100.5}", pts[0])
	}
	if pts[2].Close != 102.25 {
		t.Errorf("third close = %v, want 102.25", pts[2].Close)
	}
}

func TestEmbeddedHICPFR(t *testing.T) {
	pts, ok := embeddedHICP("FR")
	if !ok {
		t.Fatal("FR snapshot must be embedded")
	}
	if len(pts) < 300 {
		t.Fatalf("embedded FR snapshot too short: %d months", len(pts))
	}
	if !pts[0].Date.Equal(time.Date(1955, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("first anchor = %v, want 1955-01 (long history: OECD CPI chained before Eurostat)", pts[0].Date)
	}
	if _, ok := embeddedHICP("ZZ"); ok {
		t.Error("unknown geo must not have an embedded snapshot")
	}
}

func TestFetchEurostatHICPEmbeddedFallback(t *testing.T) {
	const path = "/eurostat/api/dissemination/statistics/1.0/data/prc_hicp_midx"
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway) // simulate an Eurostat outage
	})
	c, srv := newTestClient(t, t.TempDir(), mux) // fresh temp dir: no disk cache
	defer srv.Close()

	s, err := c.Fetch("^HICP-FR", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("expected embedded fallback to succeed, got %v", err)
	}
	if calls == 0 {
		t.Error("live API should have been attempted before falling back")
	}
	if s.Source != "eurostat" || len(s.Points) < 1000 {
		t.Errorf("fallback series looks wrong: source=%q points=%d", s.Source, len(s.Points))
	}
	if !s.First().Date.Equal(time.Date(1955, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("fallback first date = %v, want 1955-01-01 (embedded long history)", s.First().Date)
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
