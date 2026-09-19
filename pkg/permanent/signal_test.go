package permanent

import (
	"testing"
	"time"
)

func TestRegimesFromPanel(t *testing.T) {
	p, err := LoadPanel()
	if err != nil {
		t.Fatalf("LoadPanel: %v", err)
	}
	cfg := DefaultSignalConfig()
	rs := p.Regimes(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC), cfg)
	if len(rs) < 200 {
		t.Fatalf("expected a monthly regime for most of 1990-2020, got %d", len(rs))
	}
	for _, r := range rs {
		if r.GrowthBreadth < 0 || r.GrowthBreadth > 1 || r.InflationBreadth < 0 || r.InflationBreadth > 1 {
			t.Fatalf("breadth out of [0,1] at %s: %+v", r.Date.Format("2006-01"), r)
		}
	}
}

func TestRegimeAtMissingMonth(t *testing.T) {
	p, err := LoadPanel()
	if err != nil {
		t.Fatalf("LoadPanel: %v", err)
	}
	// Well before the panel's coverage: no regime.
	if _, ok := p.RegimeAt(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), DefaultSignalConfig()); ok {
		t.Fatal("expected no regime in 1900")
	}
}

// A release lag must read the SAME macro reading, just later: the regime of
// month m under a uniform lag of k is the regime of month m-k under none.
func TestReleaseLagsShiftTheReading(t *testing.T) {
	p, err := LoadPanel()
	if err != nil {
		t.Fatalf("LoadPanel: %v", err)
	}
	base := DefaultSignalConfig()
	lagged := base
	lagged.ReleaseLags = UniformLags(2)
	m := time.Date(2005, 6, 1, 0, 0, 0, 0, time.UTC)
	got, ok := p.RegimeAt(m, lagged)
	if !ok {
		t.Fatal("no lagged regime at 2005-06")
	}
	want, ok := p.RegimeAt(m.AddDate(0, -2, 0), base)
	if !ok {
		t.Fatal("no regime at 2005-04")
	}
	if !almost(got.GrowthBreadth, want.GrowthBreadth) || !almost(got.InflationBreadth, want.InflationBreadth) ||
		!almost(got.Slope, want.Slope) || !almost(got.RealShort, want.RealShort) {
		t.Fatalf("lagged regime %+v does not match the regime two months earlier %+v", got, want)
	}
	if !got.Date.Equal(m) {
		t.Fatalf("lagged regime dated %s, want the month it is read at (%s)", got.Date, m)
	}
}

// The zero value of ReleaseLags is the historical behaviour: every published
// figure of the design doc must keep reproducing at the default config.
func TestDefaultSignalConfigHasNoReleaseLag(t *testing.T) {
	if got := DefaultSignalConfig().ReleaseLags; got != (ReleaseLags{}) {
		t.Fatalf("DefaultSignalConfig().ReleaseLags = %+v, want the zero value", got)
	}
	if got, want := PublicationLags(), (ReleaseLags{IP: 2, CPI: 1}); got != want {
		t.Fatalf("PublicationLags() = %+v, want %+v", got, want)
	}
}

// Industrial production is delayed on its own: the growth breadth then reads an
// older month than the inflation breadth, which is exactly the point.
func TestReleaseLagsArePerDriver(t *testing.T) {
	p, err := LoadPanel()
	if err != nil {
		t.Fatalf("LoadPanel: %v", err)
	}
	cfg := DefaultSignalConfig()
	cfg.ReleaseLags = ReleaseLags{IP: 2}
	m := time.Date(2005, 6, 1, 0, 0, 0, 0, time.UTC)
	got, ok := p.RegimeAt(m, cfg)
	if !ok {
		t.Fatal("no regime at 2005-06")
	}
	base := DefaultSignalConfig()
	ipOnly, _ := p.RegimeAt(m.AddDate(0, -2, 0), base)
	plain, _ := p.RegimeAt(m, base)
	if !almost(got.GrowthBreadth, ipOnly.GrowthBreadth) {
		t.Fatalf("growth breadth %v, want the 2005-04 reading %v", got.GrowthBreadth, ipOnly.GrowthBreadth)
	}
	if !almost(got.InflationBreadth, plain.InflationBreadth) || !almost(got.Slope, plain.Slope) {
		t.Fatalf("inflation/rates moved under an IP-only lag: %+v vs %+v", got, plain)
	}
}

func TestRegimeQuadrant(t *testing.T) {
	cases := []struct {
		g, i float64
		want Quadrant
	}{
		{0.8, 0.2, GrowthQuadrant},
		{0.8, 0.8, InflationQuadrant},
		{0.2, 0.2, DeflationQuadrant},
		{0.2, 0.8, CrisisQuadrant},
	}
	for _, c := range cases {
		q := Regime{GrowthBreadth: c.g, InflationBreadth: c.i}.Quadrant()
		if q != c.want {
			t.Errorf("Quadrant(g=%.1f, i=%.1f) = %v, want %v", c.g, c.i, q, c.want)
		}
	}
	if s := CrisisQuadrant.String(); s != "crisis" {
		t.Errorf("String() = %q", s)
	}
}
