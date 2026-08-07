package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// mkWave builds a deterministic daily series whose returns oscillate around
// drift with its own frequency and phase, so different components stay
// decorrelated enough for regressions and vol targeting to behave.
func mkWave(symbol string, n int, drift, amp, freq, phase float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		v *= 1 + drift + amp*math.Sin(freq*float64(i)+phase)
	}
	return s
}

// mkCombo builds a series whose daily returns are an exact linear combination
// of the given series' returns, so a regression backcast on them fits with
// R² ≈ 1 (the offline stand-in for the real fund the recipe regresses on).
func mkCombo(symbol string, parts []*marketdata.Series, weights []float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	v := 100.0
	n := len(parts[0].Points)
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(i), Close: v})
		if i+1 < n {
			r := 0.0
			for k, p := range parts {
				r += weights[k] * (p.Points[i+1].Close/p.Points[i].Close - 1)
			}
			v *= 1 + r
		}
	}
	return s
}

// TestAllRecipesBuildOffline runs every bundled recipe's Build against a
// synthetic offline universe (canned component series + the embedded refdata,
// no network), asserting each returns a plausible series. This exercises the
// full wiring: frames, composites, the TSMOM engine, regression backcasts,
// FX conversion (fxOnDates/convertDaily), the longBack splices and the
// dailyShape blends over the real embedded refdata.
func TestAllRecipesBuildOffline(t *testing.T) {
	const n = 1600
	vfinx := mkWave("VFINX", n, 4e-4, 0.010, 1.0, 0.1)
	vtmgx := mkWave("VTMGX", n, 3e-4, 0.009, 1.3, 0.7)
	veiex := mkWave("VEIEX", n, 3e-4, 0.012, 0.7, 1.9)
	vfitx := mkWave("VFITX", n, 2e-4, 0.003, 1.7, 0.4)
	vustx := mkWave("VUSTX", n, 2e-4, 0.007, 0.9, 2.3)
	vfisx := mkWave("VFISX", n, 1e-4, 0.001, 2.1, 1.1)
	vipsx := mkWave("VIPSX", n, 2e-4, 0.004, 1.1, 2.9)
	gold := mkWave("GC=F", n, 3e-4, 0.011, 0.5, 0.8)
	crude := mkWave("CL=F", n, 2e-4, 0.020, 1.9, 2.2)
	bcom := mkWave("^BCOM", n, 2e-4, 0.012, 1.6, 0.9)
	dfsvx := mkWave("DFSVX", n, 4e-4, 0.013, 0.8, 1.4)
	disvx := mkWave("DISVX", n, 3e-4, 0.011, 1.5, 0.2)
	ibci := mkWave("IBCI", n, 2e-4, 0.004, 1.2, 1.7)
	vix := mkWave("^VIX", n, 0, 0.030, 0.6, 0.5)
	eurusd := mkWave("EURUSD=X", n, 0, 0.005, 1.4, 2.6)
	gbpusd := mkWave("GBPUSD=X", n, 0, 0.005, 0.4, 1.2)
	ezu := mkWave("EZU", n, 3e-4, 0.011, 1.1, 0.9)
	eunh := mkWave("EUNH.DE", n, 2e-4, 0.003, 1.6, 0.5)

	f := fakeFetcher{
		"VFINX": vfinx, "VTMGX": vtmgx, "VEIEX": veiex,
		"VFITX": vfitx, "VUSTX": vustx, "VFISX": vfisx, "VIPSX": vipsx,
		"GC=F": gold, "CL=F": crude, "^BCOM": bcom,
		"DFSVX": dfsvx, "DISVX": disvx, "IBCI": ibci,
		"EZU": ezu, "EUNH.DE": eunh,
		"^IRX":     mkLevels("^IRX", n, 3.0),
		"^VIX":     vix,
		"EURUSD=X": eurusd, "GBPUSD=X": gbpusd,
		// The funds the regression recipes backcast: exact factor combos,
		// so the in-sample R² clears the faithfulness floor.
		"LU0319687124": mkCombo("LU0319687124", []*marketdata.Series{vix, vfisx}, []float64{0.3, 0.7}),
		"GG00BQBFY362": mkCombo("GG00BQBFY362", []*marketdata.Series{vustx, vfinx, gold}, []float64{0.4, 0.3, 0.3}),
	}
	fetcher := WithRefData(datasets.Refdata(), f)

	for _, r := range All() {
		t.Run(r.ID, func(t *testing.T) {
			if r.ID == "" || r.Name == "" || r.Method == "" {
				t.Fatalf("recipe %q: incomplete metadata: %+v", r.ID, r)
			}
			s, err := r.Build(fetcher, time.Time{})
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			if s == nil || len(s.Points) < 300 {
				t.Fatalf("Build returned %d points, want a substantial series", len(s.Points))
			}
			prev := time.Time{}
			for _, p := range s.Points {
				if !p.Date.After(prev) {
					t.Fatalf("dates not strictly ascending at %s", p.Date)
				}
				if p.Close <= 0 || math.IsNaN(p.Close) || math.IsInf(p.Close, 0) {
					t.Fatalf("bad close %v at %s", p.Close, p.Date)
				}
				prev = p.Date
			}
		})
	}
}
