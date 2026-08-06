package marketdata

import (
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

	s, err := c.FetchExtended("VOOSIM", opt)
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
	bare, err := c.FetchExtended("VOO", opt)
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
	s, err := c.FetchExtended("VOOSIM", opt)
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
	s, err := c.FetchExtended("VOOSIM", opt)
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
	if _, err := c.FetchExtended("VOOSIM", FetchOptions{Simdata: fstest.MapFS{}}); err == nil {
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

	s, err := c.FetchExtended("VOO", FetchOptions{To: realDays[2]})
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
