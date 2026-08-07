package web

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
)

func TestBroadSampleLoads(t *testing.T) {
	pool := broadSampleEquity()
	if len(pool) < 12 {
		t.Fatalf("expected many country series, got %d", len(pool))
	}
	total := 0
	for _, s := range pool {
		total += len(s)
		if len(s) < 30 {
			t.Errorf("a country series is implausibly short: %d years", len(s))
		}
	}
	if total < 2000 {
		t.Errorf("expected ~2200 country-years, got %d", total)
	}
}

func TestBroadSampleIsEmpiricallyPessimistic(t *testing.T) {
	// The bundled broad sample must reproduce the broad-sample literature: a real
	// geometric mean well below a diversified world index (single-market runs) and
	// clearly below the rosy US backtest, so it lands as the pessimistic anchor.
	geo := decumul.GeoMean(broadSampleSource(40), 3000, rand.New(rand.NewPCG(3, 4)))
	if geo < 0.02 || geo > 0.06 {
		t.Errorf("broad-sample geometric mean %.3f outside the expected 2-6%% band", geo)
	}
	if w := decumul.Plausibility(geo, 0); len(w) != 0 {
		t.Errorf("broad-sample calibration should pass the sanity guard, got %v", w)
	}
	if math.IsNaN(geo) {
		t.Fatal("geo is NaN")
	}
}
