package marketdata

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestExtendBackPreservesDefinitionGuards(t *testing.T) {
	for _, junction := range []int{8, 10} {
		t.Run(fmt.Sprint(junction), func(t *testing.T) {
			proxy := &Series{Symbol: "PROXY", Warnings: []string{"donor qualification"}, Junctions: []time.Time{d(2020, 1, junction), d(2020, 1, 13)}}
			for i := 6; i <= 13; i++ {
				v := float64(100 + i)
				if i >= junction {
					v *= 2
				}
				proxy.Points = append(proxy.Points, Point{Date: d(2020, 1, i), Close: v})
			}
			asset := &Series{Symbol: "ASSET", Warnings: []string{"own qualification"}, Junctions: []time.Time{d(2020, 1, 12)}, Points: []Point{{d(2020, 1, 10), 100}, {d(2020, 1, 11), 101}, {d(2020, 1, 12), 102}}}
			if !ExtendBack(asset, proxy) {
				t.Fatal("not extended")
			}
			if !slices.Contains(asset.Junctions, d(2020, 1, junction)) || !slices.Contains(asset.Junctions, d(2020, 1, 12)) || slices.Contains(asset.Junctions, d(2020, 1, 13)) {
				t.Errorf("wrong retained junctions: %v", asset.Junctions)
			}
			panel, err := NewPanel(Daily, asset)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range panel.R[0] {
				if r > .1 {
					t.Errorf("definition change became a return: %v", r)
				}
			}
			if !strings.Contains(strings.Join(asset.Warnings, " "), "donor qualification") || !slices.Contains(asset.Warnings, "own qualification") {
				t.Errorf("lost warnings: %v", asset.Warnings)
			}
			asset.Warnings[0] = "changed"
			asset.Junctions[0] = time.Time{}
			if proxy.Warnings[0] != "donor qualification" || !proxy.Junctions[0].Equal(d(2020, 1, junction)) {
				t.Fatal("mutated donor metadata")
			}
		})
	}
}

func TestIncomeWarningsSurviveComposition(t *testing.T) {
	for _, source := range []string{"stooq", "ft", "morningstar"} {
		t.Run(source, func(t *testing.T) {
			s := &Series{Symbol: "PRICE", Name: "Example fund (Dist)", Source: source, Currency: "USD", Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 101}, {d(2020, 1, 8), 102}}}
			var b bytes.Buffer
			if err := WriteCSV(&b, s); err != nil {
				t.Fatal(err)
			}
			restored, err := ReadLongCSV(&b)
			if err != nil {
				t.Fatal(err)
			}
			panel, err := NewPanel(Daily, s)
			if err != nil {
				t.Fatal(err)
			}
			mix, err := panel.Mix("MIX", map[string]float64{"PRICE": 1})
			if err != nil {
				t.Fatal(err)
			}
			blended, err := mix.Series("MIX")
			if err != nil {
				t.Fatal(err)
			}
			for name, out := range map[string]*Series{"csv": restored[0], "mix": blended} {
				if !strings.Contains(strings.Join(out.Warnings, " "), "income") {
					t.Errorf("%s erased income warning: %v", name, out.Warnings)
				}
			}
			if !reflect.DeepEqual(restored[0].Points, s.Points) {
				t.Fatal("export changed values")
			}
			if len(s.Warnings) != 0 {
				t.Fatal("transformation mutated caller metadata")
			}
		})
	}
}

func TestFallbackQualificationsSurviveExport(t *testing.T) {
	c := NewClient(t.TempDir())
	s := &Series{Symbol: "EXAMPLE", Currency: "USD", Source: "yahoo", Warnings: []string{"provider qualification"}, Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 101}}}
	c.saveCache(s, time.Time{})
	for _, tc := range []struct{ cache, id, warning string }{{"EXAMPLE", "EXAMPLE", "keeping cached data"}, {"absent", "^VIX", "embedded"}} {
		t.Run(tc.warning, func(t *testing.T) {
			out, err := c.staleFallback(context.Background(), tc.cache, tc.id, time.Time{}, errors.New("fixture outage"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(strings.Join(out.Warnings, " "), tc.warning) {
				t.Errorf("fallback not exposed: %v", out.Warnings)
			}
			if tc.cache == "EXAMPLE" && !slices.Contains(out.Warnings, "provider qualification") {
				t.Error("cache dropped provider qualification")
			}
			var b bytes.Buffer
			if err := WriteCSV(&b, out); err != nil {
				t.Fatal(err)
			}
			round, err := ReadLongCSV(&b)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(out.Warnings, round[0].Warnings) {
				t.Errorf("CSV lost fallback warning: %v", round[0].Warnings)
			}
		})
	}
	cached, ok := c.loadCache("EXAMPLE", time.Time{})
	if !ok {
		t.Fatal("cache missing")
	}
	if strings.Contains(strings.Join(cached.Warnings, " "), "fixture outage") {
		t.Fatal("transient fallback warning persisted as provider fact")
	}
}

func TestFrozenHICPWarningSurvivesCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"value":{"0":100,"1":102},"dimension":{"time":{"category":{"index":{"2000-01":0,"2000-02":1}}}}}`)
	}))
	defer srv.Close()
	dir := t.TempDir()
	for _, offline := range []bool{false, true} {
		c := NewClient(dir)
		stubAllBases(c, srv.URL)
		c.Offline = offline
		s, err := c.Load(context.Background(), "^HICP-ZZ", FetchOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.Join(s.Warnings, " "), "frozen") {
			t.Errorf("offline=%v: no frozen-source qualification: %v", offline, s.Warnings)
		}
	}
}

func TestSimulationFlagCSVAndBlend(t *testing.T) {
	s := &Series{Symbol: "SYNTH", Currency: "USD", FullySimulated: true, Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 101}, {d(2020, 1, 8), 102}}}
	var b bytes.Buffer
	if err := WriteCSV(&b, s); err != nil {
		t.Fatal(err)
	}
	restored, err := ReadLongCSV(&b)
	if err != nil {
		t.Fatal(err)
	}
	if !restored[0].FullySimulated || !strings.Contains(strings.Join(restored[0].Warnings, " "), "entire history") {
		t.Errorf("lost reconstruction status: %+v", restored[0])
	}
	p, err := NewPanel(Daily, s)
	if err != nil {
		t.Fatal(err)
	}
	mix, err := p.Mix("MIX", map[string]float64{"SYNTH": 1})
	if err != nil {
		t.Fatal(err)
	}
	m, err := mix.Series("MIX")
	if err != nil {
		t.Fatal(err)
	}
	if !m.FullySimulated {
		t.Error("wholly reconstructed blend lost its status")
	}
	// Reject contradictory claims in either transport direction.
	s.SimulatedBefore = d(2020, 1, 7)
	b.Reset()
	if err := WriteCSV(&b, s); err == nil || b.Len() != 0 {
		t.Error("wrote contradictory simulation metadata")
	}
	for _, headers := range []string{"# fully-simulated: true\n# simulated-before: 2020-01-07\n", "# fully-simulated: maybe\n"} {
		if _, err := ReadCSV(strings.NewReader(headers+"date,value\n2020-01-06,100\n2020-01-07,101\n"), "BAD"); err == nil {
			t.Error("accepted contradictory or invalid simulation metadata")
		}
	}
}

func TestDataWarningsDetachedAndUnique(t *testing.T) {
	s := &Series{Source: "stooq", Warnings: []string{"existing"}}
	first := s.DataWarnings()
	s.Warnings = first
	second := s.DataWarnings()
	if len(second) != 2 {
		t.Fatalf("duplicate source warning: %v", second)
	}
	second[0] = "changed"
	if s.Warnings[0] != "existing" {
		t.Fatal("mutated input warnings")
	}
	if got := (*Series)(nil).DataWarnings(); len(got) != 0 {
		t.Fatalf("nil warnings=%v", got)
	}
}

func TestExtendBackDoesNotInventRealQuotes(t *testing.T) {
	s := &Series{Symbol: "SYNTH", FullySimulated: true, Points: []Point{{d(2020, 1, 7), 100}, {d(2020, 1, 8), 101}}}
	donor := &Series{Symbol: "DONOR", Points: []Point{{d(2020, 1, 6), 99}, {d(2020, 1, 7), 100}, {d(2020, 1, 8), 101}}}
	if ExtendBack(s, donor) || !s.SimulatedBefore.IsZero() {
		t.Fatal("invented a real-quote frontier for an entirely synthetic receiver")
	}
}

func TestKnownSIMFallbackIsVisible(t *testing.T) {
	s := &Series{Symbol: "SYNTH", Currency: "USD", SimulatedBefore: d(2020, 1, 7), Points: []Point{{d(2020, 1, 6), 100}, {d(2020, 1, 7), 101}}}
	c := NewClient("")
	out, err := c.simOnly(context.Background(), s, "SYNTH", "SYNTH", errors.New("provider detail stays in logs"), FetchOptions{NoConvert: true})
	if err != nil {
		t.Fatal(err)
	}
	warnings := strings.Join(out.Warnings, " ")
	if !strings.Contains(warnings, "bundled history only") || !strings.Contains(warnings, "2020-01-07") {
		t.Errorf("fallback invisible: %v", out.Warnings)
	}
	if strings.Contains(warnings, "provider detail") {
		t.Error("raw provider diagnostic escaped into exported warnings")
	}
	if !out.SimulatedBefore.Equal(s.SimulatedBefore) || len(s.Warnings) != 0 {
		t.Fatal("fallback changed provenance or caller metadata")
	}
}
