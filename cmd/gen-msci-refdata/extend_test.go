package main

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// fixture builds an anchor series of month-end levels and a proxy tracking it
// with a monthly fee drag plus noise, the proxy running extra months past the
// anchor. Both are deterministic.
func fixture(t *testing.T, anchorMonths, extraMonths int, ter, noise float64) ([]marketdata.Point, monthly) {
	t.Helper()
	rng := rand.New(rand.NewSource(7))
	drag := math.Pow(1+ter/100, 1.0/12)
	anchors := []marketdata.Point{{Date: calendarMonthEnd("2009-12"), Close: 1000}}
	proxy := monthly{"2009-12": 50}
	key := "2009-12"
	for i := 0; i < anchorMonths+extraMonths; i++ {
		t0, _ := time.Parse("2006-01", key)
		key = t0.AddDate(0, 1, 0).Format("2006-01")
		r := 0.01 + rng.NormFloat64()*0.04
		prev := anchors[len(anchors)-1]
		if i < anchorMonths {
			// Rounded as the CSV stores it, so a round trip is exact.
			lvl := math.Round(prev.Close*(1+r)*1e6) / 1e6
			anchors = append(anchors, marketdata.Point{Date: calendarMonthEnd(key), Close: lvl})
		}
		proxy[key] = proxy[prevKey(key)] * (1 + r + rng.NormFloat64()*noise) / drag
	}
	return anchors, proxy
}

func prevKey(k string) string {
	t, _ := time.Parse("2006-01", k)
	return t.AddDate(0, -1, 0).Format("2006-01")
}

func TestCompareAndBuildTail(t *testing.T) {
	anchors, proxy := fixture(t, 132, 2, 0.20, 0.0015)
	full, gated, err := compare(anchors, proxy, 0.20)
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if full.months != 132 || gated.months != gradeWindow {
		t.Fatalf("windows: full %d months, gated %d, want 132 and %d", full.months, gated.months, gradeWindow)
	}
	if math.Abs(gated.td) > maxTD || gated.corr < minCorr {
		t.Fatalf("a faithful proxy was graded TD %+.3f corr %.4f", gated.td, gated.corr)
	}

	tail, err := buildTail(anchors, proxy, 0.20, maxMonthly)
	if err != nil {
		t.Fatalf("buildTail: %v", err)
	}
	if len(tail) != 2 {
		t.Fatalf("appended %d months, want 2 (the proxy's own last month is incomplete only when it comes from a daily series; here both extra months are complete)", len(tail))
	}
	last := anchors[len(anchors)-1]
	gross := math.Pow(1.002, 1.0/12) // the 0.20 %/yr charge, added back
	wantFirst := last.Close * (proxy[month(tail[0].Date)] / proxy[month(last.Date)]) * gross
	if math.Abs(tail[0].Close-wantFirst) > 1e-9 {
		t.Errorf("first appended level %.6f, want %.6f (the proxy's return, charge added back, chained onto the last anchor)", tail[0].Close, wantFirst)
	}
	// Without the add-back the tail would bleed the fund's charge into the
	// index: the guard is that the two differ by exactly one month of it.
	bare, err := buildTail(anchors, proxy, 0, maxMonthly)
	if err != nil {
		t.Fatalf("buildTail without the charge: %v", err)
	}
	if got, want := tail[0].Close/bare[0].Close, gross; math.Abs(got-want) > 1e-12 {
		t.Errorf("the add-back lifted the first month by x%.9f, want x%.9f", got, want)
	}
	for _, p := range tail {
		if got, want := p.Date, calendarMonthEnd(month(p.Date)); !got.Equal(want) {
			t.Errorf("appended point dated %s, want the calendar month end %s", got.Format("2006-01-02"), want.Format("2006-01-02"))
		}
	}
	if !tail[0].Date.After(last.Date) {
		t.Errorf("tail starts at %s, not after the last anchor %s", tail[0].Date, last.Date)
	}
}

func TestCompareRefusals(t *testing.T) {
	t.Run("proxy no later than the anchors", func(t *testing.T) {
		anchors, proxy := fixture(t, 132, 0, 0.20, 0.001)
		if _, _, err := compare(anchors, proxy, 0.20); err == nil {
			t.Fatal("accepted a proxy that stops where the anchors stop")
		}
	})
	t.Run("overlap too short", func(t *testing.T) {
		anchors, proxy := fixture(t, minOverlapMonths-2, 2, 0.20, 0.001)
		if _, _, err := compare(anchors, proxy, 0.20); err == nil {
			t.Fatalf("accepted an overlap of %d months", minOverlapMonths-2)
		}
	})
	t.Run("tracking difference out of band", func(t *testing.T) {
		// A proxy dragged by 1.5 %/yr more than the charge added back.
		anchors, proxy := fixture(t, 132, 2, 1.70, 0.001)
		_, gated, err := compare(anchors, proxy, 0.20)
		if err == nil {
			t.Fatalf("accepted TD %+.3f %%/yr", gated.td)
		}
		if gated.td > -maxTD {
			t.Errorf("TD %+.3f %%/yr, expected it well below -%.2f", gated.td, maxTD)
		}
	})
	t.Run("correlation below the floor", func(t *testing.T) {
		anchors, proxy := fixture(t, 132, 2, 0.20, 0.02) // 2 %/mo of noise
		_, gated, err := compare(anchors, proxy, 0.20)
		if err == nil {
			t.Fatalf("accepted corr %.4f", gated.corr)
		}
		if gated.corr >= minCorr {
			t.Errorf("corr %.4f, expected it below %.3f", gated.corr, minCorr)
		}
	})
	t.Run("appended return beyond the sanity bound", func(t *testing.T) {
		anchors, proxy := fixture(t, 132, 2, 0.20, 0.001)
		next := month(anchors[len(anchors)-1].Date)
		next = nextKey(next)
		proxy[next] *= 2 // a +100 % month
		if _, err := buildTail(anchors, proxy, 0.20, maxMonthly); err == nil {
			t.Fatal("accepted a doubling month")
		}
	})
	t.Run("proxy missing the anchors' last month", func(t *testing.T) {
		anchors, proxy := fixture(t, 132, 2, 0.20, 0.001)
		delete(proxy, month(anchors[len(anchors)-1].Date))
		if _, err := buildTail(anchors, proxy, 0.20, maxMonthly); err == nil {
			t.Fatal("accepted a proxy with no level to chain from")
		}
	})
}

func nextKey(k string) string {
	t, _ := time.Parse("2006-01", k)
	return t.AddDate(0, 1, 0).Format("2006-01")
}

// TestCompleteMonthEnds pins the rule that the proxy's own last month never
// reaches a reference series: a month-end level struck mid-month is not one.
func TestCompleteMonthEnds(t *testing.T) {
	s := &marketdata.Series{}
	for d := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC); d.Before(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)); d = d.AddDate(0, 0, 1) {
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: 100 + float64(d.YearDay())})
	}
	levels, partial := completeMonthEnds(s)
	if partial != "2026-09" {
		t.Errorf("partial month %q, want 2026-09", partial)
	}
	if _, ok := levels["2026-09"]; ok {
		t.Error("the incomplete month survived")
	}
	if got, want := levels.last(), "2026-08"; got != want {
		t.Errorf("last complete month %q, want %q", got, want)
	}
	if got, want := levels["2026-06"], 100+float64(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).YearDay()); got != want {
		t.Errorf("June level %v, want the month's last close %v", got, want)
	}
}

// TestRoundTripIdempotent is the tail policy: writing a file and rebuilding it
// from the same proxy keeps the export era byte for byte, replaces the tail
// with an identical one, and never moves the tail-from boundary.
func TestRoundTripIdempotent(t *testing.T) {
	anchors, proxy := fixture(t, 132, 2, 0.20, 0.0015)
	tail, err := buildTail(anchors, proxy, 0.20, maxMonthly)
	if err != nil {
		t.Fatalf("buildTail: %v", err)
	}
	path := filepath.Join(t.TempDir(), "TEST-USD.csv")
	first := refdata{
		id: "TEST-USD", name: "test", source: "fixture", tailSource: "fixture proxy",
		tailFrom: month(tail[0].Date), generated: "2026-09-11",
		anchors: append(append([]marketdata.Point(nil), anchors...), tail...),
	}
	if err := writeRefdata(path, first); err != nil {
		t.Fatalf("write: %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	back, err := readRefdata(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if back.tailFrom != first.tailFrom {
		t.Errorf("tail-from read back as %q, want %q", back.tailFrom, first.tailFrom)
	}
	if len(back.anchors) != len(anchors) {
		t.Fatalf("read back %d export points, want %d (the tail must be dropped)", len(back.anchors), len(anchors))
	}
	for i, p := range back.anchors {
		if !p.Date.Equal(anchors[i].Date) || math.Abs(p.Close-anchors[i].Close) > 1e-6 {
			t.Fatalf("export point %d moved: %v %v, want %v %v", i, p.Date, p.Close, anchors[i].Date, anchors[i].Close)
		}
	}
	again, err := buildTail(back.anchors, proxy, 0.20, maxMonthly)
	if err != nil {
		t.Fatalf("second buildTail: %v", err)
	}
	second := first
	second.anchors = append(append([]marketdata.Point(nil), back.anchors...), again...)
	if err := writeRefdata(path, second); err != nil {
		t.Fatalf("second write: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("a second run with no new month changed the file")
	}
}

// TestTargetsAreCatalogued keeps every declared proxy resolvable offline, with
// the pinned ongoing charge the fee add-back needs and the series' currency.
func TestTargetsAreCatalogued(t *testing.T) {
	for _, tg := range targets {
		if len(tg.proxies) == 0 {
			t.Errorf("%s: no proxy declared", tg.id)
		}
		for _, id := range tg.proxies {
			a, ok := marketdata.Lookup(id)
			if !ok {
				t.Errorf("%s: proxy %s is not in the bundled catalog", tg.id, id)
				continue
			}
			if a.Fees <= 0 {
				t.Errorf("%s: proxy %s has no pinned ongoing charge", tg.id, id)
			}
			if a.Currency != tg.currency {
				t.Errorf("%s: proxy %s quotes in %s, want %s", tg.id, id, a.Currency, tg.currency)
			}
		}
	}
}
