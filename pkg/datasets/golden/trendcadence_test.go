package golden

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// The deepest donor of the managed-futures family dealt WEEKLY until 2016, so
// the whole segment the chains keep of it is one NAV a week projected onto a
// daily calendar (simgen's densify). Two things can go wrong there and only one
// of them is visible in a level: splice the weekly steps raw and the file reads
// roughly sqrt(5) times too volatile; blend them into a reconstruction whose
// own daily amplitude was calibrated on something else and the file reads
// however volatile THAT was, which until 2026-08 was about 16 % too calm.
//
// The invariant that catches both is cadence consistency: a fund NAV is close
// enough to a random walk that its volatility must annualize to about the same
// number whether it is measured day by day or week by week. The externally
// measured references, on real records over their own daily eras:
//
//	Man AHL Diversified, 2017-2026 daily NAVs      0.994
//	the daily net all-styles composite, 2000-2026   0.946
//	the daily net pure-trend composite, 2000-2026   0.951
//	iMGP DBi Managed Futures ETF, live 2019-2026    1.106
//
// A single fund sits at or slightly above 1 (its NAV carries a little daily
// valuation noise that cancels over a week), a twenty-programme composite
// slightly below it (its constituents' days are averaged, which leaves the
// aggregate mildly trending). The band below spans that evidence generously
// and would have failed the pre-2026-08 files, which read 0.84 to 0.89.
func TestGoldenTrendDonorEraCadence(t *testing.T) {
	const (
		from = "1996-03-26" // the deepest donor's own first NAV
		to   = "1999-12-31" // the last year every chain still stands on it
		lo   = 0.90
		hi   = 1.10
	)
	for _, id := range []string{
		"DBMF", "LU2951555585", "DBMFE", "MFEH", "KMLM", "CTA",
		"LU1103257975", "LU1662501532", "LU1662498721", "LU2622190622",
	} {
		t.Run(id, func(t *testing.T) {
			s, ok, err := marketdata.ReadSimdataFS(datasets.Simdata(), id)
			if err != nil || !ok {
				t.Fatalf("simdata %s: ok=%v err=%v", id, ok, err)
			}
			seg := between(t, s, from, to)
			if len(seg) < 900 {
				t.Fatalf("%s: %d points over %s..%s, expected a full daily calendar", id, len(seg), from, to)
			}
			daily := annualized(logReturns(seg), 252)
			weekly := annualized(logReturns(weekEnds(seg)), 52)
			if weekly <= 0 {
				t.Fatalf("%s: no weekly volatility to compare against", id)
			}
			if r := daily / weekly; r < lo || r > hi {
				t.Errorf("%s: daily volatility %.2f %% against weekly %.2f %%, ratio %.3f, want %.2f..%.2f",
					id, daily*100, weekly*100, r, lo, hi)
			}
		})
	}
}

func between(t *testing.T, s *marketdata.Series, from, to string) []marketdata.Point {
	t.Helper()
	a, b := mustDate(t, from), mustDate(t, to)
	var out []marketdata.Point
	for _, p := range s.Points {
		if !p.Date.Before(a) && !p.Date.After(b) {
			out = append(out, p)
		}
	}
	return out
}

// weekEnds keeps the last point of every ISO week.
func weekEnds(points []marketdata.Point) []marketdata.Point {
	var out []marketdata.Point
	for i, p := range points {
		y, w := p.Date.ISOWeek()
		if i+1 == len(points) {
			out = append(out, p)
			continue
		}
		if ny, nw := points[i+1].Date.ISOWeek(); ny != y || nw != w {
			out = append(out, p)
		}
	}
	return out
}

func logReturns(points []marketdata.Point) []float64 {
	var out []float64
	for i := 1; i < len(points); i++ {
		if points[i-1].Close > 0 && points[i].Close > 0 {
			out = append(out, math.Log(points[i].Close/points[i-1].Close))
		}
	}
	return out
}

func annualized(rets []float64, perYear float64) float64 {
	if len(rets) < 2 {
		return 0
	}
	var m float64
	for _, r := range rets {
		m += r
	}
	m /= float64(len(rets))
	var v float64
	for _, r := range rets {
		v += (r - m) * (r - m)
	}
	return math.Sqrt(v/float64(len(rets)-1)) * math.Sqrt(perYear)
}
