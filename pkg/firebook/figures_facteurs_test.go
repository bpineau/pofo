package firebook

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// scvRealLegs rebuilds the two real legs of the factor plate from the bundled
// record: the last quote of every calendar month of each nominal series,
// divided by the ^CPI-US level of that same month (the deflator pkg/replay
// uses). It returns the contiguous common months and both real levels, so the
// tests below can recompute every number the plate freezes.
func scvRealLegs(t *testing.T) (months []int, scv, sp []float64) {
	t.Helper()
	byMonth := func(id string) map[int]float64 {
		s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
		if err != nil || !ok {
			t.Fatalf("bundled series %s: ok=%v err=%v", id, ok, err)
		}
		out := make(map[int]float64, len(s.Points))
		for _, p := range s.Points {
			if p.Close > 0 { // the last quote of the month wins: both legs end on month ends
				out[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close
			}
		}
		return out
	}
	cpiSeries, err := marketdata.NewClient("").Fetch(context.Background(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatalf("bundled ^CPI-US snapshot: %v", err)
	}
	cpi := make(map[int]float64, len(cpiSeries.Points))
	for _, p := range cpiSeries.Points {
		if p.Close > 0 {
			cpi[p.Date.Year()*12+int(p.Date.Month())-1] = p.Close
		}
	}
	value, index := byMonth("USSCV-USD"), byMonth("SP500-USD")

	// The record opens with the small-value series in July 1963, exactly ten
	// years before the first window the plate can close.
	first := scvGapStart - 120
	for k := first; ; k++ {
		v, okV := value[k]
		i, okI := index[k]
		c, okC := cpi[k]
		if !okV || !okI || !okC {
			break
		}
		months = append(months, k)
		scv = append(scv, v/c)
		sp = append(sp, i/c)
	}
	if len(months) == 0 {
		t.Fatal("no common month between USSCV-USD, SP500-USD and ^CPI-US")
	}
	return months, scv, sp
}

// cagr10 is the annualized real return of a 120-month window closing at i.
func cagr10(real []float64, i int) float64 {
	return (math.Pow(real[i]/real[i-120], 0.1) - 1) * 100
}

// The plate's 635 monthly readings are frozen literals; recompute all of them
// from pkg/datasets and fail on any drift.
func TestScvGapMatchesTheRecord(t *testing.T) {
	months, scv, sp := scvRealLegs(t)
	if months[0] != scvGapStart-120 {
		t.Fatalf("record opens at month key %d, the plate assumes %d (July 1963)", months[0], scvGapStart-120)
	}
	if got, want := len(months)-120, len(scvGapPoints); got != want {
		t.Fatalf("the record closes %d ten-year windows, the plate freezes %d", got, want)
	}
	if last := months[len(months)-1]; last != scvGapStart+len(scvGapPoints)-1 {
		t.Fatalf("last window ends at month key %d, the plate assumes %d", last, scvGapStart+len(scvGapPoints)-1)
	}
	for i, want := range scvGapPoints {
		got := cagr10(scv, i+120) - cagr10(sp, i+120)
		if math.Abs(got-want) > 0.02 {
			t.Errorf("window ending %04d-%02d: record %+.2f pt, plate %+.2f pt",
				(scvGapStart+i)/12, (scvGapStart+i)%12+1, got, want)
		}
	}
}

// Every window the plate annotates must still be the extreme of its era, and
// the numbers its footnote quotes must still come out of the record. Nothing
// here is a recollection: the eras are wide brackets, the readings are argmax
// and argmin inside them.
func TestScvGapAnnotationsAreTheRealExtremes(t *testing.T) {
	months, scv, sp := scvRealLegs(t)
	gap := func(month int) float64 { return scvGapPoints[month-scvGapStart] }
	extreme := func(fromYear, toYear int, max bool) int {
		best := 0
		for _, m := range months[120:] {
			if m/12 < fromYear || m/12 > toYear {
				continue
			}
			if best == 0 || (max && gap(m) > gap(best)) || (!max && gap(m) < gap(best)) {
				best = m
			}
		}
		return best
	}
	for _, tc := range []struct {
		name     string
		want     int
		from, to int
		max      bool
	}{
		{"widest window", scvPeakMonth, 1973, 1990, true},
		{"second peak", scvPeak2Month, 2000, 2012, true},
		{"1990s trough", scvDipMonth, 1990, 2000, false},
		{"deepest window", scvTroughMonth, 2012, 2026, false},
		{"deepest window, whole record", scvTroughMonth, 1973, 2026, false},
	} {
		if got := extreme(tc.from, tc.to, tc.max); got != tc.want {
			t.Errorf("%s over %d-%d: record says %04d-%02d (%+.2f pt), plate annotates %04d-%02d (%+.2f pt)",
				tc.name, tc.from, tc.to, got/12, got%12+1, gap(got), tc.want/12, tc.want%12+1, gap(tc.want))
		}
	}

	// The full-period real CAGRs of the footnote, and the sanity check that goes
	// with them: US small value has beaten the index by a few points a year, and
	// the tilt is a displacement, not a free lunch.
	n := len(months) - 1
	years := float64(n) / 12
	full := func(real []float64) float64 { return (math.Pow(real[n]/real[0], 1/years) - 1) * 100 }
	if got := full(scv); math.Abs(got-scvFullCAGR) > 0.01 {
		t.Errorf("small-value real CAGR %.2f %%, plate says %.2f %%", got, scvFullCAGR)
	}
	if got := full(sp); math.Abs(got-spFullCAGR) > 0.01 {
		t.Errorf("S&P 500 real CAGR %.2f %%, plate says %.2f %%", got, spFullCAGR)
	}
	if d := full(scv) - full(sp); d < 2 || d > 5 {
		t.Errorf("full-period edge %.2f pt a year, outside the plausible 2 to 5 pt: suspect an alignment bug", d)
	}

	// The two winning windows are the index's own lost decades, which is the
	// whole claim of the plate; the footnote quotes what the index did in them.
	for _, tc := range []struct {
		month int
		want  float64
	}{{scvPeakMonth, scvPeakIndexCAGR}, {scvPeak2Month, scvPeak2IndexCAGR}} {
		got := cagr10(sp, tc.month-scvGapStart+120)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("index over the window ending %04d-%02d: %+.2f %%/an, plate says %+.2f",
				tc.month/12, tc.month%12+1, got, tc.want)
		}
		if got > 3 {
			t.Errorf("window ending %04d-%02d is no lost decade for the index (%+.2f %%/an)", tc.month/12, tc.month%12+1, got)
		}
	}

	above := 0
	for _, v := range scvGapPoints {
		if v > 0 {
			above++
		}
	}
	if got := float64(above) / float64(len(scvGapPoints)) * 100; math.Abs(got-scvShareAbove) > 0.1 {
		t.Errorf("%.1f %% of the windows are above zero, plate says %.1f %%", got, scvShareAbove)
	}
}

// The article states two things about the tilt's decades that this plate is
// asked to prove; check them against the same record rather than against memory.
func TestScvGapAgreesWithTheArticle(t *testing.T) {
	months, scv, sp := scvRealLegs(t)
	idx := func(year int, month time.Month) int {
		k := year*12 + int(month) - 1
		for i, m := range months {
			if m == k {
				return i
			}
		}
		t.Fatalf("%04d-%02d missing from the record", year, month)
		return 0
	}

	// "le retraité de l'an 2000 investi en SCV a connu une décennie honorable
	// pendant que le S&P vivait sa décennie perdue" (the 2000-2010 decade).
	end := idx(2010, time.December)
	value, index := cagr10(scv, end), cagr10(sp, end)
	if index > 0 {
		t.Errorf("2000-2010 was no lost decade for the index: %+.2f %%/an real", index)
	}
	if value < 5 {
		t.Errorf("2000-2010 was no honourable decade for small value: %+.2f %%/an real", value)
	}

	// "le pire épisode de son histoire, pire encore que 1999": the purgatory of
	// the 2010s must be deeper than the 1990s dip, on this ground too.
	deep := scvGapPoints[scvTroughMonth-scvGapStart]
	dip := scvGapPoints[scvDipMonth-scvGapStart]
	if deep >= dip {
		t.Errorf("the 2010s trough (%+.2f pt) is not deeper than the 1990s dip (%+.2f pt)", deep, dip)
	}
}
