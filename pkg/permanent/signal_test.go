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
	// Well before OECD MEI coverage: no regime.
	if _, ok := p.RegimeAt(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), DefaultSignalConfig()); ok {
		t.Fatal("expected no regime in 1900")
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
