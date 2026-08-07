package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// A FRED-backed rate arrives as a percent level with no currency, and a
// negative reading survives the trip: rates go below zero, and the euro area
// spent eight years there, so the parser must not treat it as a bad print.
func TestPolicyRateFromFRED(t *testing.T) {
	hits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/graph/fredgraph.csv", func(w http.ResponseWriter, r *http.Request) {
		hits++
		if got := r.URL.Query().Get("id"); got != "SOFR" {
			t.Errorf("FRED series id = %q", got)
		}
		fmt.Fprint(w, "observation_date,SOFR\n"+
			"2021-06-01,-0.568\n2021-06-02,-0.570\n2026-07-30,2.185\n")
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(context.Background(), "^SOFR", time.Time{})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if s.Currency != "" {
		t.Errorf("currency = %q, a rate level carries none", s.Currency)
	}
	if s.Source != "fred" || s.Name == "" {
		t.Errorf("source %q, name %q", s.Source, s.Name)
	}
	if len(s.Points) != 3 {
		t.Fatalf("%d points, want 3", len(s.Points))
	}
	if got := s.Points[0].Close; got != -0.568 {
		t.Errorf("first close = %v, want the negative rate -0.568", got)
	}
	// Second call is served from the cache written by the first.
	if _, err := c.Fetch(context.Background(), "^SOFR", time.Time{}); err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if hits != 1 {
		t.Errorf("%d downloads, want 1 (the second call must hit the cache)", hits)
	}
}

// A DBnomics-backed monthly rate is dated at the END of each month, and the
// "NA" observations the provider publishes as a string are skipped.
func TestPolicyRateFromDBnomics(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v22/series/ECB/FM/M.U2.EUR.RT.MM.EURIBOR3MD_.HSTA", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"series":{"docs":[{"period":["2026-01","2026-02","2026-03"],`+
			`"value":[2.51,"NA",2.44]}]}}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.Fetch(context.Background(), "^EURIBOR3M", time.Time{})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(s.Points) != 2 {
		t.Fatalf("%d points, want 2 (the NA month is dropped)", len(s.Points))
	}
	want := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	if !s.Points[0].Date.Equal(want) {
		t.Errorf("first date = %s, want the month end %s",
			s.Points[0].Date.Format("2006-01-02"), want.Format("2006-01-02"))
	}
	if s.Points[1].Close != 2.44 {
		t.Errorf("second close = %v, want 2.44", s.Points[1].Close)
	}
}

// The registry is the single source of truth for the family, and every entry
// must be reachable through the symbol dispatch.
func TestRateSymbolsRegistry(t *testing.T) {
	syms := RateSymbols()
	if len(syms) != len(policyRates) {
		t.Fatalf("%d symbols listed, %d in the registry", len(syms), len(policyRates))
	}
	for i := 1; i < len(syms); i++ {
		if syms[i] < syms[i-1] {
			t.Errorf("symbols are not sorted: %s before %s", syms[i-1], syms[i])
		}
	}
	for _, s := range syms {
		if !isPolicyRate(s) {
			t.Errorf("%s is listed but not routed", s)
		}
		spec := policyRates[s]
		if len(spec.sources) == 0 || spec.name == "" {
			t.Errorf("%s: incomplete spec %+v", s, spec)
		}
		for _, src := range spec.sources {
			if src.provider == "" || src.id == "" {
				t.Errorf("%s: incomplete source %+v", s, src)
			}
		}
	}
	if isPolicyRate("^GSPC") || isPolicyRate("VOO") {
		t.Error("a price symbol must not be routed as a rate")
	}
}

// The data doctor judges a rate by its own rules: zero and negative levels are
// readings, not corruption, and a policy rate that has not moved in a year is
// doing its job. Only an implausible jump is worth a word.
func TestVerifyRateSeries(t *testing.T) {
	day := func(i int) time.Time { return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &Series{Symbol: "^ECB-DFR"}
	for i := 0; i < 400; i++ {
		s.Points = append(s.Points, Point{Date: day(i), Close: -0.50}) // the 2021 deposit rate
	}
	if issues := Verify(s, day(401)); len(issues) != 0 {
		t.Fatalf("a flat negative policy rate is not a data problem: %v", issues)
	}
	// The same series as a price would raise both a non-positive error and a
	// stale-feed warning, which is exactly what the rate rules suppress.
	price := &Series{Symbol: "SOMEFUND", Points: s.Points}
	if issues := Verify(price, day(401)); len(issues) == 0 {
		t.Error("a price series with negative flat closes should still be flagged")
	}
	// A four-point move in a day is not a central bank decision.
	s.Points = append(s.Points, Point{Date: day(400), Close: 3.6})
	issues := Verify(s, day(401))
	if len(issues) != 1 || issues[0].Severity != "warn" {
		t.Fatalf("a 4.1-point daily jump should warn once, got %v", issues)
	}
}

// Staleness and gaps are measured against the series' own cadence: a monthly
// publication is not late because it has no quote from last week.
func TestVerifyFollowsTheCadence(t *testing.T) {
	monthly := &Series{Symbol: "^EURIBOR3M"}
	d := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 24; i++ {
		monthly.Points = append(monthly.Points, Point{Date: d.AddDate(0, i, 0), Close: 3 + float64(i)/100})
	}
	last := monthly.Points[len(monthly.Points)-1].Date
	if issues := Verify(monthly, last.AddDate(0, 0, 40)); len(issues) != 0 {
		t.Errorf("a monthly series 40 days after its last point is not stale: %v", issues)
	}
	if issues := Verify(monthly, last.AddDate(0, 0, 200)); len(issues) == 0 {
		t.Error("a monthly series 200 days late should be flagged")
	}
}

// The New York Fed answers newest first and carries two columns in one
// payload: the rate it computes, and the FOMC target range it lives in.
func TestPolicyRateFromNYFed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/rates/unsecured/effr/search.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"refRates":[`+
			`{"effectiveDate":"2026-07-30","type":"EFFR","percentRate":3.63,"targetRateTo":3.75},`+
			`{"effectiveDate":"2026-07-29","type":"EFFR","percentRate":3.62,"targetRateTo":3.75},`+
			`{"effectiveDate":"2026-07-28","type":"EFFR","percentRate":3.61,"targetRateTo":3.75}]}`)
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	// FRED leads for ^FEDFUNDS and is stubbed away here, so the fallback runs.
	s, err := c.Fetch(context.Background(), "^FEDFUNDS", time.Time{})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if s.Source != "nyfed" {
		t.Errorf("source = %q, want the fallback nyfed", s.Source)
	}
	if len(s.Points) != 3 || s.Points[0].Close != 3.61 {
		t.Fatalf("points = %+v, want three rows oldest first", s.Points)
	}
	if !s.Points[0].Date.Before(s.Points[2].Date) {
		t.Error("the series must run forward in time")
	}
	target, err := c.Fetch(context.Background(), "^FED-TARGET", time.Time{})
	if err != nil {
		t.Fatalf("target fetch: %v", err)
	}
	if target.Points[0].Close != 3.75 {
		t.Errorf("target upper bound = %v, want 3.75", target.Points[0].Close)
	}
}

// FRED answers a browser User-Agent over HTTP/1.1 by hanging until the client
// times out, which had silently emptied every FRED-backed series (measured
// 2026-08, see fredUserAgent). The requests must therefore carry a plain agent
// of our own rather than the client's browser one.
func TestFREDRequestsCarryAPlainUserAgent(t *testing.T) {
	var seen string
	mux := http.NewServeMux()
	mux.HandleFunc("/graph/fredgraph.csv", func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("User-Agent")
		fmt.Fprint(w, "observation_date,SOFR\n2026-07-29,2.20\n2026-07-30,2.185\n")
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	c.UserAgent = "Mozilla/5.0 (Macintosh) Chrome/124.0 Safari/537.36"

	if _, err := c.Fetch(context.Background(), "^SOFR", time.Time{}); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if seen != fredUserAgent {
		t.Errorf("User-Agent = %q, want the plain %q", seen, fredUserAgent)
	}
}
