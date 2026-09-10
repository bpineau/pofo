package marketdata

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// simdataFS returns an in-memory simdata directory holding one series for
// VOO: ten daily closes ending right before the real quotes begin.
func simdataFS(days []time.Time, closes []float64) fstest.MapFS {
	var b strings.Builder
	b.WriteString("# pofo simdata v1\n# id: VOO\n# name: VOO (simulated)\ndate,close\n")
	for i := range days {
		fmt.Fprintf(&b, "%s,%.6f\n", days[i].Format("2006-01-02"), closes[i])
	}
	return fstest.MapFS{"VOO.csv": &fstest.MapFile{Data: []byte(b.String())}}
}

func TestFetchExtendedSplicesSimdata(t *testing.T) {
	realDays := testDays(3) // 2020-01-06 …
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", realDays, []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	simDays := make([]time.Time, 10)
	simCloses := make([]float64, 10)
	for i := range simDays {
		simDays[i] = realDays[0].AddDate(0, 0, i-10)
		simCloses[i] = 50 + float64(i)
	}
	opt := FetchOptions{Simdata: simdataFS(simDays, simCloses)}

	s, err := c.FetchExtended(context.Background(), "VOOSIM", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 13 {
		t.Fatalf("expected 10 simulated + 3 real points, got %d", len(s.Points))
	}
	if !s.SimulatedBefore.Equal(realDays[0]) {
		t.Errorf("SimulatedBefore = %v, want %v", s.SimulatedBefore, realDays[0])
	}
	if s.ProxySymbol != "simdata" {
		t.Errorf("ProxySymbol = %q, want simdata", s.ProxySymbol)
	}
	// The simulated leg is rescaled to the first real quote: the last
	// simulated close (59) anchors nothing, the scale comes from the
	// proxy value at the anchor date; just check continuity of ordering.
	for i := 1; i < len(s.Points); i++ {
		if !s.Points[i].Date.After(s.Points[i-1].Date) {
			t.Fatalf("dates not ascending at %d", i)
		}
	}

	// The memoized bare series must stay unextended.
	bare, err := c.FetchExtended(context.Background(), "VOO", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(bare.Points) != 3 || !bare.SimulatedBefore.IsZero() {
		t.Errorf("bare fetch polluted by the SIM extension: %d points, SimulatedBefore=%v",
			len(bare.Points), bare.SimulatedBefore)
	}
}

func TestFetchExtendedNoSim(t *testing.T) {
	realDays := testDays(3)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", realDays, []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	simDays := []time.Time{realDays[0].AddDate(0, 0, -2), realDays[0].AddDate(0, 0, -1)}
	opt := FetchOptions{Simdata: simdataFS(simDays, []float64{50, 51}), NoSim: true}
	s, err := c.FetchExtended(context.Background(), "VOOSIM", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 3 || !s.SimulatedBefore.IsZero() {
		t.Errorf("NoSim must return real quotes only: %d points, SimulatedBefore=%v",
			len(s.Points), s.SimulatedBefore)
	}
}

func TestFetchExtendedSimdataOnlyFallback(t *testing.T) {
	// No handler for VOO: the real fetch fails, the simulated series
	// (2+ points) is served alone, flagged as fully simulated.
	c, srv := newTestClient(t, t.TempDir(), http.NewServeMux())
	defer srv.Close()

	base := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	simDays := []time.Time{base, base.AddDate(0, 0, 1), base.AddDate(0, 0, 2)}
	opt := FetchOptions{Simdata: simdataFS(simDays, []float64{50, 51, 52})}
	s, err := c.FetchExtended(context.Background(), "VOOSIM", opt)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 3 || s.ProxySymbol != "simdata" {
		t.Fatalf("expected the simulated series alone: %+v", s)
	}
	if !s.SimulatedBefore.Equal(s.Last().Date) {
		t.Errorf("a simdata-only series must be flagged simulated throughout")
	}

	// Without simulated data the original fetch error must surface.
	if _, err := c.FetchExtended(context.Background(), "VOOSIM", FetchOptions{Simdata: fstest.MapFS{}}); err == nil {
		t.Error("expected an error when neither real nor simulated data exists")
	}
}

func TestFetchExtendedWindow(t *testing.T) {
	realDays := testDays(5)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", realDays, []float64{100, 101, 102, 103, 104}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.FetchExtended(context.Background(), "VOO", FetchOptions{To: realDays[2]})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 3 || !s.Last().Date.Equal(realDays[2]) {
		t.Errorf("To not honored: %d points, last %v", len(s.Points), s.Last().Date)
	}
}

func TestTrim(t *testing.T) {
	days := testDays(5)
	s := &Series{Symbol: "X", Currency: "USD"}
	for i, d := range days {
		s.Points = append(s.Points, Point{Date: d, Close: 100 + float64(i)})
	}
	if got := Trim(s, time.Time{}, time.Time{}); got != s {
		t.Error("an open window must return the series untouched")
	}
	got := Trim(s, days[1], days[3])
	if len(got.Points) != 3 || !got.First().Date.Equal(days[1]) || !got.Last().Date.Equal(days[3]) {
		t.Errorf("Trim window wrong: %+v", got.Points)
	}
	if len(s.Points) != 5 {
		t.Error("Trim must not mutate its input")
	}
	if got.Symbol != "X" || got.Currency != "USD" {
		t.Error("Trim must keep the series metadata")
	}
	empty := &Series{}
	if got := Trim(empty, days[0], days[1]); got != empty {
		t.Error("an empty series must pass through")
	}
}

func TestDefaultCacheDir(t *testing.T) {
	if DefaultCacheDir() == "" {
		t.Error("DefaultCacheDir must never be empty")
	}
}

// TestFetchExtendedFallsBackToAProxy: with no bundled simdata and more than
// six months missing, the known long-history proxy is spliced in, rescaled to
// the first real quote.
func TestFetchExtendedFallsBackToAProxy(t *testing.T) {
	realDays := testDays(3) // 2020-01-06 …
	proxyDays := []time.Time{d(2000, 1, 3), d(2010, 6, 1), realDays[0]}
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/SPY", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("SPY", realDays, []float64{100, 101, 102}))
	})
	mux.HandleFunc("/v8/finance/chart/%5EGSPC", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("^GSPC", proxyDays, []float64{20, 40, 50}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	var logged strings.Builder
	c.Logf = func(format string, args ...any) { fmt.Fprintf(&logged, format+"\n", args...) }

	s, err := c.FetchExtended(context.Background(), "SPYSIM", FetchOptions{
		From: d(1999, 1, 1), Simdata: fstest.MapFS{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.ProxySymbol != "^GSPC" || !s.SimulatedBefore.Equal(realDays[0]) {
		t.Fatalf("proxy metadata: %+v", s)
	}
	// The proxy is rescaled to the first real quote: 50 → 100, so ×2.
	if len(s.Points) != 5 || s.First().Close != 40 {
		t.Fatalf("points = %+v, want the rescaled proxy in front", s.Points)
	}
	if !strings.Contains(logged.String(), "history extended via ^GSPC") {
		t.Errorf("the splice was not reported: %q", logged.String())
	}

	// An unavailable proxy costs a warning and the short series, never the fetch.
	c2, srv2 := newTestClient(t, t.TempDir(), func() *http.ServeMux {
		m := http.NewServeMux()
		m.HandleFunc("/v8/finance/chart/SPY", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, chartJSON("SPY", realDays, []float64{100, 101, 102}))
		})
		return m
	}())
	defer srv2.Close()
	var warned strings.Builder
	c2.Logf = func(format string, args ...any) { fmt.Fprintf(&warned, format+"\n", args...) }
	s2, err := c2.FetchExtended(context.Background(), "SPYSIM", FetchOptions{
		From: d(1999, 1, 1), Simdata: fstest.MapFS{},
	})
	if err != nil {
		t.Fatalf("a missing proxy must not fail the fetch: %v", err)
	}
	if len(s2.Points) != 3 || !s2.SimulatedBefore.IsZero() {
		t.Errorf("series = %+v, want the real quotes alone", s2.Points)
	}
	if !strings.Contains(warned.String(), "proxy ^GSPC for SPY unavailable") {
		t.Errorf("the missing proxy was not reported: %q", warned.String())
	}
}

// TestFetchExtendedWarnsOnUnreadableSimdata: a corrupt simdata file must be
// reported and stepped over, never silently spliced or fatal.
func TestFetchExtendedWarnsOnUnreadableSimdata(t *testing.T) {
	realDays := testDays(3)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", realDays, []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	var warned strings.Builder
	c.Logf = func(format string, args ...any) { fmt.Fprintf(&warned, format+"\n", args...) }

	broken := fstest.MapFS{"VOO.csv": &fstest.MapFile{
		Data: []byte("# pofo simdata v1\ndate,close\n2019-01-02,not-a-number\n")}}
	s, err := c.FetchExtended(context.Background(), "VOOSIM", FetchOptions{
		From: d(2018, 1, 1), Simdata: broken,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) != 3 || !s.SimulatedBefore.IsZero() {
		t.Errorf("series = %+v, want the real quotes alone", s.Points)
	}
	if !strings.Contains(warned.String(), "simdata VOO unreadable") {
		t.Errorf("the corrupt file was not reported: %q", warned.String())
	}
}

// TestFetchExtendedUsesTheEmbeddedSimdataByDefault: a nil Simdata reads the
// series embedded in the binary, which is what every CLI call does.
func TestFetchExtendedUsesTheEmbeddedSimdataByDefault(t *testing.T) {
	realDays := testDays(3)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/TLT", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("TLT", realDays, []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()

	s, err := c.FetchExtended(context.Background(), "TLTSIM", FetchOptions{From: d(1990, 1, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if s.ProxySymbol != "simdata" || s.SimulatedBefore.IsZero() {
		t.Fatalf("the bundled simdata was not spliced: %+v", s)
	}
	if !s.First().Date.Before(d(2000, 1, 1)) {
		t.Errorf("series starts %s, want the bundled deep history", s.First().Date)
	}
}

// TestFetchExtendedReportsAConversionFailure: an FX cross nothing can serve
// must fail the fetch, on both the plain and the SIM path. A silently
// unconverted series would be read as if it were in the target currency.
func TestFetchExtendedReportsAConversionFailure(t *testing.T) {
	realDays := testDays(3)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", realDays, []float64{100, 101, 102}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	for _, id := range []string{"VOO", "VOOSIM"} {
		_, err := c.FetchExtended(context.Background(), id, FetchOptions{
			From: d(2020, 1, 1), Currency: "SEK", Simdata: fstest.MapFS{},
		})
		if err == nil || !strings.Contains(err.Error(), "FX rate USD") {
			t.Errorf("%s: error = %v, want a conversion failure", id, err)
		}
	}
}

// TestConvertToWarnsWhenFXIsHeldFlat: the report leans on this warning to say
// that early points carry the oldest known cross rather than a real one.
func TestConvertToWarnsWhenFXIsHeldFlat(t *testing.T) {
	days := testDays(4)
	mux := http.NewServeMux()
	mux.HandleFunc("/v8/finance/chart/VOO", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("VOO", days, []float64{100, 101, 102, 103}))
	})
	mux.HandleFunc("/v8/finance/chart/USDSEK=X", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartJSON("USDSEK=X", days[2:], []float64{9, 9.1}))
	})
	c, srv := newTestClient(t, t.TempDir(), mux)
	defer srv.Close()
	var warned strings.Builder
	c.Logf = func(format string, args ...any) { fmt.Fprintf(&warned, format+"\n", args...) }

	s, err := c.FetchExtended(context.Background(), "VOO", FetchOptions{From: days[0], Currency: "SEK"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Currency != "SEK" || s.First().Close != 100*9 {
		t.Fatalf("conversion = %+v (%s)", s.Points, s.Currency)
	}
	if !strings.Contains(warned.String(), "held constant earlier") {
		t.Errorf("the flat-held rate was not reported: %q", warned.String())
	}
}

// TestDefaultCacheDirFallback: with no user cache directory to be found, the
// library must still name a usable one rather than an empty path.
func TestDefaultCacheDirFallback(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	if got := DefaultCacheDir(); got != "data" {
		t.Errorf("DefaultCacheDir() = %q, want the %q fallback", got, "data")
	}
}
