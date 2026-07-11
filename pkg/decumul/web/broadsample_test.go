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

// The mixed pool must hold each country's contiguous 60/40 runs, split (not
// spliced) at the bond record's war breaks.
func TestBroadSampleMixedPool(t *testing.T) {
	pool := broadSampleMixed()
	if len(pool) < 18 || len(pool) > 26 {
		t.Fatalf("expected ~21 country runs (some countries split by war gaps), got %d", len(pool))
	}
	total := 0
	for _, s := range pool {
		total += len(s)
		if len(s) < 10 {
			t.Errorf("a run shorter than 10y should have been dropped: %d", len(s))
		}
	}
	if total < 1900 {
		t.Errorf("expected ~2100 country-years with both records, got %d", total)
	}
}

func TestBroadSampleIsEmpiricallyPessimistic(t *testing.T) {
	// The Broad-sample model holds the 60/40 domestic mix of the broad-sample
	// SWR literature (Anarkulova, Cederburg & O'Doherty). Anchors, fixed rule,
	// no tax, 30y: their baseline finds ~17% ruin for the 4% rule and a ~2.26%
	// safe rate at 5% ruin (with longevity-weighted horizons); on a FIXED 30y
	// horizon and the disaster-heavy JST-16 pool this model reads a little
	// stricter (measured 2026-07-12: ruin ~23%, SWR ~1.7%). The bands lock
	// that calibration: clearly grimmer than the US backtest, same order of
	// magnitude as the literature.
	geo := decumul.GeoMean(broadSampleSource(40), 3000, rand.New(rand.NewPCG(3, 4)))
	if geo < 0.025 || geo > 0.05 {
		t.Errorf("broad-sample 60/40 geometric mean %.3f outside the expected 2.5-5%% band", geo)
	}
	if w := decumul.Plausibility(geo, 0); len(w) != 0 {
		t.Errorf("broad-sample calibration should pass the sanity guard, got %v", w)
	}
	if math.IsNaN(geo) {
		t.Fatal("geo is NaN")
	}
	p := decumul.Plan{Capital: 1e6, NeedAnnual: 40000, Years: 30, Source: broadSampleSource(30)}
	ruin := p.Simulate(8000, 8, 7).Outcome().RuinProb
	if ruin < 0.15 || ruin > 0.30 {
		t.Errorf("60/40 fixed 4%%/30y ruin = %.1f%%, want 15-30%% (Anarkulova-class, stricter on a fixed horizon)", ruin*100)
	}
	swr := p.Solve(0.05, decumul.WithdrawalAxis(0, 0.15e6), 8000, 8, 7)
	if swr < 12000 || swr > 26000 {
		t.Errorf("60/40 SWR at 5%% ruin = %.2f%%, want 1.2-2.6%%", swr/1e4)
	}
}
