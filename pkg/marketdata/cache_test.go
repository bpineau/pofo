package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLoadCacheRejectsUnusableFiles pins what the cache refuses to believe.
// Every rejection must read as a MISS, never as an error and never as data:
// a corrupt entry costs one download, a trusted one costs a wrong report.
func TestLoadCacheRejectsUnusableFiles(t *testing.T) {
	from := d(2020, 1, 1)
	good := cacheFile{
		Symbol: "VOO", Currency: "USD", Source: "yahoo",
		RequestedFrom: "2020-01-01", FetchedAt: time.Now(),
		Dates: []string{"2020-01-06", "2020-01-07"}, Closes: []float64{100, 101},
	}
	cases := []struct {
		name string
		raw  string           // written as is when set
		edit func(*cacheFile) // applied to a copy of good otherwise
		ok   bool
	}{
		{name: "sound entry", ok: true},
		{name: "truncated JSON", raw: `{"symbol":"VOO","dates":["2020-01-`},
		{name: "not JSON at all", raw: "\x00\x01garbage"},
		{name: "no dates", edit: func(c *cacheFile) { c.Dates, c.Closes = nil, nil }},
		{name: "dates and closes disagree", edit: func(c *cacheFile) { c.Closes = []float64{100} }},
		{name: "unparseable requested_from", edit: func(c *cacheFile) { c.RequestedFrom = "january" }},
		{name: "shallower than asked", edit: func(c *cacheFile) { c.RequestedFrom = "2021-01-01" }},
		{name: "unparseable date", edit: func(c *cacheFile) { c.Dates[0] = "06/01/2020" }},
		{name: "every point predates the window", edit: func(c *cacheFile) {
			c.Dates = []string{"2019-01-02", "2019-01-03"}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			c := NewClient(dir)
			data := []byte(tc.raw)
			if tc.raw == "" {
				cf := good
				cf.Dates = append([]string(nil), good.Dates...)
				cf.Closes = append([]float64(nil), good.Closes...)
				if tc.edit != nil {
					tc.edit(&cf)
				}
				data, _ = json.Marshal(cf)
			}
			if err := os.WriteFile(c.cachePath("VOO"), data, 0o644); err != nil {
				t.Fatal(err)
			}
			s, ok := c.loadCache("VOO", from)
			if ok != tc.ok {
				t.Fatalf("loadCache ok = %v, want %v", ok, tc.ok)
			}
			if ok && len(s.Points) != 2 {
				t.Errorf("points = %+v", s.Points)
			}
		})
	}
}

// TestCacheDividendsClippedAndGuarded covers the two dividend rules: the
// window clips them like points, and parallel arrays of different lengths are
// dropped wholesale rather than paired at random.
func TestCacheDividendsClippedAndGuarded(t *testing.T) {
	dir := t.TempDir()
	c := NewClient(dir)
	write := func(divDates []string, amounts []float64) *Series {
		cf := cacheFile{
			Symbol: "VOO", RequestedFrom: "2019-01-01", FetchedAt: time.Now(),
			Dates: []string{"2020-01-06"}, Closes: []float64{100},
			DivDates: divDates, DivAmounts: amounts,
		}
		data, _ := json.Marshal(cf)
		if err := os.WriteFile(c.cachePath("VOO"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		s, ok := c.loadCache("VOO", d(2020, 1, 1))
		if !ok {
			t.Fatal("the entry should have loaded")
		}
		return s
	}
	s := write([]string{"2019-06-01", "2020-01-02", "oops"}, []float64{1, 2, 3})
	if len(s.Dividends) != 1 || s.Dividends[0].Amount != 2 {
		t.Errorf("dividends = %+v, want only the in-window, parseable one", s.Dividends)
	}
	if s := write([]string{"2020-01-02"}, []float64{1, 2}); len(s.Dividends) != 0 {
		t.Errorf("mismatched dividend arrays must be dropped, got %+v", s.Dividends)
	}
}

// TestCacheWriteIsAtomicAndSurvivesAnUnusableDir pins the temp-then-rename
// write: nothing is left behind on success, and a directory that cannot be
// created warns instead of failing the fetch.
func TestCacheWriteIsAtomicAndSurvivesAnUnusableDir(t *testing.T) {
	days := testDays(2)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, []float64{100, 101}))
	})
	dir := t.TempDir()
	c, srv := newTestClient(t, dir, mux)
	defer srv.Close()
	if _, err := c.History(context.Background(), "VOO", d(2020, 1, 1)); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("a temporary file survived the rename: %s", e.Name())
		}
	}

	// A cache directory that is really a file: the fetch still answers.
	blocked := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	c2 := NewClient(filepath.Join(blocked, "pofo"))
	stubAllBases(c2, srv.URL)
	var warned bool
	c2.Logf = func(format string, args ...any) {
		if strings.Contains(fmt.Sprintf(format, args...), "cache directory unusable") {
			warned = true
		}
	}
	if _, err := c2.History(context.Background(), "VOO", d(2020, 1, 1)); err != nil {
		t.Fatalf("an unusable cache must not fail the fetch: %v", err)
	}
	if !warned {
		t.Error("an unusable cache directory should have warned")
	}
}

func TestCachelessClientNeverTouchesDisk(t *testing.T) {
	c := NewClient("")
	if _, ok := c.loadCache("VOO", d(2020, 1, 1)); ok {
		t.Error("a cache-less client must never load anything")
	}
	if c.Cached("VOO") {
		t.Error("a cache-less client is never warm")
	}
	c.saveCache(&Series{Symbol: "VOO", Points: []Point{{Date: d(2020, 1, 6), Close: 1}}}, d(2020, 1, 1))
}

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		"VOO":         "VOO",
		"^GSPC":       "_GSPC",
		"GC=F":        "GC_F",
		"IWDA.AS":     "IWDA.AS",
		"a/../../etc": "a_.._.._etc",
		"VOO~raw":     "VOO_raw",
		"TREND-NET_1": "TREND-NET_1",
		"^HICP-EA19":  "_HICP-EA19",
	}
	for in, want := range cases {
		if got := sanitizeFilename(in); got != want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
}
