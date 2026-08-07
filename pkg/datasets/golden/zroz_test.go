package golden

import (
	"sort"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// TestGoldenZROZLeverage freezes the ZROZ reconstruction's leverage against its
// own long-Treasury base. ZROZ (PIMCO 25+ year zero-coupon STRIPS, ~27-year
// duration) is backcast as cash + 1.65x(VUSTX - cash); TLT (20+ year Treasury,
// ~17-year duration) is the un-levered long-Treasury base. The two must keep the
// duration ratio, ~27/17 = ~1.6x, so realized daily volatility of ZROZ should be
// ~1.5-1.65x that of TLT in every era. This is validated against ground truth:
// over the live overlap (both funds real from 2009) the actual ZROZ/TLT vol
// ratio is 1.54x, and the reconstruction reproduces it. The test guards against
// a regression in the ZROZ leverage (an over- or under-levered STRIPS proxy
// silently distorts every long-duration backcast that leans on it).
//
// Windows are chosen across regimes: the Volcker rate shock, the VUSTX-based
// modern reconstruction, and the real-data era (both grafted). Volatility here
// is the standard sqrt(252) daily figure, valid because both series are daily.
func TestGoldenZROZLeverage(t *testing.T) {
	zroz := loadSimdata(t, "ZROZ")
	tlt := loadSimdata(t, "TLT")

	for _, w := range []struct {
		name, from, to string
		wantRatio, tol float64
	}{
		{"1979-1986 (Volcker rate shock)", "1979-01-01", "1986-01-01", 1.60, 0.20},
		{"1986-2009 (VUSTX reconstruction)", "1986-01-01", "2009-01-01", 1.60, 0.20},
		{"2009-2026 (real, ground truth 1.54x)", "2009-01-01", "2026-07-01", 1.54, 0.15},
	} {
		t.Run(w.name, func(t *testing.T) {
			zv := seriesVol(t, zroz, w.from, w.to)
			tv := seriesVol(t, tlt, w.from, w.to)
			ratio := zv / tv
			if ratio < w.wantRatio-w.tol || ratio > w.wantRatio+w.tol {
				t.Errorf("ZROZ/TLT vol ratio = %.2f (ZROZ %.1f%%, TLT %.1f%%), expected %.2f ±%.2f",
					ratio, zv*100, tv*100, w.wantRatio, w.tol)
			}
		})
	}
}

func loadSimdata(t *testing.T, id string) *marketdata.Series {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), id)
	if err != nil || !ok {
		t.Fatalf("simdata %s: ok=%v err=%v", id, ok, err)
	}
	return s
}

// seriesVol computes the annualized daily volatility of a raw simdata series
// over [from, to].
func seriesVol(t *testing.T, s *marketdata.Series, from, to string) float64 {
	t.Helper()
	f, o := mustDate(t, from), mustDate(t, to)
	var dates []time.Time
	var values []float64
	for _, p := range s.Points {
		if p.Date.Before(f) || p.Date.After(o) {
			continue
		}
		dates = append(dates, p.Date)
		values = append(values, p.Close)
	}
	if !sort.SliceIsSorted(dates, func(i, j int) bool { return dates[i].Before(dates[j]) }) {
		t.Fatalf("series %s not sorted", s.Name)
	}
	if len(dates) < 500 {
		t.Fatalf("window %s..%s too short: %d points", from, to, len(dates))
	}
	stats, err := metrics.Compute(dates, values)
	if err != nil {
		t.Fatal(err)
	}
	return stats.Volatility
}
