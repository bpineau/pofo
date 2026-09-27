package marketdata

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
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

func TestFetchEurostatHICPEmbedFirst(t *testing.T) {
	const path = "/eurostat/api/dissemination/statistics/1.0/data/" + hicpDataset
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Eurostat must not be hit: ^HICP-FR is served offline-first from the embed")
		w.WriteHeader(http.StatusBadGateway)
	})
	c, srv := newTestClient(t, t.TempDir(), mux) // fresh temp dir: no disk cache
	defer srv.Close()

	s, err := c.Fetch(context.Background(), "^HICP-FR", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("expected the embedded snapshot to serve, got %v", err)
	}
	if s.Source != "eurostat" || len(s.Points) < 1000 {
		t.Errorf("embedded series looks wrong: source=%q points=%d", s.Source, len(s.Points))
	}
	if !s.First().Date.Equal(time.Date(1955, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("embedded first date = %v, want 1955-01-01 (long history)", s.First().Date)
	}
}

// A geography without a bundled snapshot (^HICP-EA) still uses the live API in
// the default (non-refresh) mode: embed-first only diverts geos that have one.
func TestFetchEurostatHICPNoEmbedGoesLive(t *testing.T) {
	const path = "/eurostat/api/dissemination/statistics/1.0/data/" + hicpDataset
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"value": {"0": 100.0, "1": 101.0},
			"dimension": {"time": {"category": {"index": {"2006-01": 0, "2006-02": 1}}}}
		}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(context.Background(), "^HICP-EA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "eurostat" || !s.First().Date.Equal(time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("live series looks wrong: source=%q first=%v", s.Source, s.First().Date)
	}
}

func TestFetchEurostatHICP(t *testing.T) {
	const path = "/eurostat/api/dissemination/statistics/1.0/data/" + hicpDataset
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// The one series of the live dataset: all items (ECOICOP 2's TOTAL),
		// monthly, on the 2015=100 unit the bundled snapshot is chained on.
		q := r.URL.Query()
		for key, want := range map[string]string{"geo": "FR", "unit": "I15", "coicop18": "TOTAL", "freq": "M", "format": "JSON"} {
			if got := q.Get(key); got != want {
				t.Errorf("%s query = %q, want %q", key, got, want)
			}
		}
		fmt.Fprint(w, `{
			"value": {"0": 100.0, "1": 101.0},
			"dimension": {"time": {"category": {"index": {"2006-01": 0, "2006-02": 1}}}}
		}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	c.RefreshInflation = true // the live Eurostat path is refresh-only

	s, err := c.Fetch(context.Background(), "^HICP-FR", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
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

// TestHICPLagging pins the frozen-dataset guard: Eurostat publishes month M in
// the middle of M+1, so one or two months behind is a live series and more
// than three is one that stopped (the former prc_hicp_midx, stuck at 2025-12).
func TestHICPLagging(t *testing.T) {
	month := func(y int, m time.Month) time.Time { return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC) }
	now := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		last time.Time
		want bool
	}{
		{month(2026, 8), false},
		{month(2026, 7), false},
		{month(2026, 6), false}, // three months: the allowance itself
		{month(2026, 5), true},
		{month(2025, 12), true}, // across a year boundary
	} {
		if got := hicpLagging(c.last, now); got != c.want {
			t.Errorf("hicpLagging(%s, %s) = %v, want %v", c.last.Format("2006-01"), now.Format("2006-01-02"), got, c.want)
		}
	}
}

// A live HICP that trails the calendar is still served (a late deflator
// beats none) but says so, since every later date deflates at zero inflation.
func TestFetchEurostatHICPWarnsWhenFrozen(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/eurostat/api/dissemination/statistics/1.0/data/"+hicpDataset, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"value": {"0": 100.0, "1": 101.0},
			"dimension": {"time": {"category": {"index": {"2006-01": 0, "2006-02": 1}}}}
		}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	var logged []string
	c.Logf = func(format string, args ...any) { logged = append(logged, fmt.Sprintf(format, args...)) }

	if _, err := c.Fetch(context.Background(), "^HICP-EA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	for _, line := range logged {
		if strings.HasPrefix(line, "warning:") && strings.Contains(line, "2006-02") && strings.Contains(line, "EA") {
			return
		}
	}
	t.Errorf("no frozen-dataset warning naming the last month; logged %q", logged)
}

// A HICP cached from the frozen dataset (filed under the former "eurostat"
// identity) is never served: the cache identity names the dataset, so the
// move to prc_hicp_minr refetches instead of replaying a 2025-12 end.
func TestFetchEurostatHICPIgnoresFormerDatasetCache(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/eurostat/api/dissemination/statistics/1.0/data/"+hicpDataset, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"value": {"0": 100.0, "1": 101.0, "2": 102.0},
			"dimension": {"time": {"category": {"index": {"2006-01": 0, "2006-02": 1, "2006-03": 2}}}}
		}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	from := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	frozen := hicpSeries("^HICP-EA", "EA", []Point{
		{Date: time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC), Close: 100},
		{Date: time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC), Close: 101},
	})
	c.saveCacheAs(sourceCacheID("eurostat", "^HICP-EA", false), frozen, from)

	s, err := c.Fetch(context.Background(), "^HICP-EA", from)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Last().Date; !got.Equal(time.Date(2006, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("last = %s, want 2006-03-01: the former dataset's cached copy was served", got.Format("2006-01-02"))
	}
}
