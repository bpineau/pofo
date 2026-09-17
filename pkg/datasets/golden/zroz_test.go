package golden

import (
	"math"
	"sort"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// TestGoldenZROZLeverage pins the ZROZ reconstruction's risk against its own
// long-Treasury base, era by era, on the closed-form duration arithmetic of the
// two instruments rather than on a frozen multiple. ZROZ (PIMCO 25+ year
// zero-coupon STRIPS) is a constant 27-year zero priced off the bundled long
// yield; TLT (20+ year coupon Treasuries, a ladder whose modified duration is
// that of a ~26-year par bond, i.e. the ~17 years the fund publishes at the
// yields of the 2010s) is the un-levered coupon base.
//
// The expected volatility ratio is therefore the duration ratio AT THE ERA'S
// YIELD LEVEL, and that is the whole point of the test: a zero's duration is
// its maturity at any yield while a par bond's collapses as its yield rises, so
// the ratio is ~1.56 at the 3.5 % long yield of the 2010s and above 3 in the
// double-digit years. Until 2026-09 this test asserted ~1.6x "in every era" and
// the reconstruction obliged, because it was a coupon fund geared by a constant
// 1.65: both were wrong together, and the file reported an instrument half as
// risky as a long strip through the 1962-1985 rate rise. The tolerance is
// relative and wide enough for the residual (TLT's own backcast gears a
// shorter donor by a constant, whose duration drifts a few percent across the
// same range) and nowhere near wide enough to let a constant multiple back in.
//
// The live window is the ground truth that calibrates nothing: over 2009-2026,
// with both funds real, the measured ratio is ~1.54 and the formula predicts
// 1.56 from the era's yield alone.
//
// Volatility is the standard sqrt(252) daily figure, valid because both series
// are daily over these windows.
func TestGoldenZROZLeverage(t *testing.T) {
	zroz := loadSimdata(t, "ZROZ")
	tlt := loadSimdata(t, "TLT")

	for _, w := range []struct {
		name, from, to string
		relTol         float64
	}{
		{"1979-1986 (Volcker rate shock)", "1979-01-01", "1986-01-01", 0.15},
		{"1986-2009 (yields on the way down)", "1986-01-01", "2009-01-01", 0.15},
		{"2009-2026 (real, ground truth 1.54x)", "2009-01-01", "2026-07-01", 0.15},
	} {
		t.Run(w.name, func(t *testing.T) {
			zv := seriesVol(t, zroz, w.from, w.to)
			tv := seriesVol(t, tlt, w.from, w.to)
			ratio := zv / tv

			y := meanLongYield(t, w.from, w.to)
			want := zeroModDuration(zrozMaturityYears, y) / parModDuration(tltParMaturityYears, y)
			if lo, hi := want*(1-w.relTol), want*(1+w.relTol); ratio < lo || ratio > hi {
				t.Errorf("ZROZ/TLT vol ratio = %.2f (ZROZ %.1f%%, TLT %.1f%%), want %.2f ±%.0f%% at a %.2f%% long yield",
					ratio, zv*100, tv*100, want, w.relTol*100, y*100)
			}
		})
	}
}

// The two instruments the ratio compares, as maturities: the strip the ZROZ
// recipe holds (the average maturity of the 25 to 30 year paper its index owns)
// and the par bond whose modified duration matches what TLT publishes.
const (
	zrozMaturityYears   = 27.0
	tltParMaturityYears = 26.0
)

// zeroModDuration is the modified duration of an n-year zero-coupon bond at the
// semiannual bond-equivalent yield y: n/(1+y/2), textbook, and nearly n at any
// yield a Treasury has ever carried.
func zeroModDuration(n, y float64) float64 { return n / (1 + y/2) }

// parModDuration is the modified duration of an n-year PAR bond paying
// semiannual coupons at the same yield. This is the quantity that shrinks as
// the yield rises: 17.0 years at 3.5 %, 7.9 at 12 %, for the same 26-year bond.
func parModDuration(n, y float64) float64 {
	i := y / 2
	macaulayPeriods := (1 + i) / i * (1 - math.Pow(1+i, -2*n))
	return macaulayPeriods / 2 / (1 + i)
}

// meanLongYield averages the bundled long Treasury yield over the window, as a
// decimal. It is the era's rate level, read from the same series the
// reconstruction is priced off.
func meanLongYield(t *testing.T, from, to string) float64 {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), "TREASURY-LONG-YIELD")
	if err != nil || !ok {
		t.Fatalf("refdata TREASURY-LONG-YIELD: ok=%v err=%v", ok, err)
	}
	f, o := mustDate(t, from), mustDate(t, to)
	var sum float64
	var n int
	for _, p := range s.Points {
		if p.Date.Before(f) || p.Date.After(o) {
			continue
		}
		sum += p.Close
		n++
	}
	if n < 500 {
		t.Fatalf("window %s..%s carries only %d yield observations", from, to, n)
	}
	return sum / float64(n) / 100
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
