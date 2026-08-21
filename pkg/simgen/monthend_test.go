package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// loadRef reads a bundled reference series straight from pkg/datasets/refdata.
func loadRef(t *testing.T, id string) *marketdata.Series {
	t.Helper()
	s, ok, err := marketdata.ReadSimdataFS(datasets.Refdata(), id)
	if err != nil || !ok {
		t.Fatalf("refdata %s: ok=%v err=%v", id, ok, err)
	}
	return s
}

// decToDec is the December-to-December return (%) of calendar year y: the last
// close on or before 31 December of y over that of y-1. It reads a monthly and a
// daily series alike, which is the point: a correctly shaped reconstruction must
// return the same calendar year as the monthly anchor it was shaped from.
func decToDec(s *marketdata.Series, y int) float64 {
	dec := func(yr int) float64 {
		cut := time.Date(yr, 12, 31, 23, 0, 0, 0, time.UTC)
		var v float64
		for _, p := range s.Points {
			if p.Date.After(cut) {
				break
			}
			v = p.Close
		}
		return v
	}
	a, b := dec(y-1), dec(y)
	if a == 0 {
		return 0
	}
	return (b/a - 1) * 100
}

// TestAlignMonthEndPreservesCalendarYears is the end-to-end guard on the
// month-end convention, run on the real bundled data rather than a fixture.
//
// The developed-ex-US anchor (DEVEXUS-USD, MSCI World ex USA net TR) holds
// month-END levels, and shaping it with the daily Ken French shape must not move
// a single calendar year: whatever the daily path does in between, the level on
// the last trading day of December has to stay the anchor's December level. That
// is the property the whole reconstruction rests on, and the one the first-of-
// month labels broke (the pre-fix blend rebuilt 2022 at -4.3 % against the
// index's -14.3 %, and 2016 at +13.5 % against +2.8 %).
//
// Without alignMonthEnd the same blend is wrong by up to ~0.7 point a year,
// because anchorShape pins a month-end date that falls on a weekend into the
// next month; the second half of the test pins that difference so the snap
// cannot be quietly dropped.
func TestAlignMonthEndPreservesCalendarYears(t *testing.T) {
	anchor := loadRef(t, "DEVEXUS-USD")
	shape := loadRef(t, "DEVEXUS-DAILY")
	shaped := shapedSeries(alignMonthEnd("DEVEXUS-USD", anchor, shape), shape)

	var worst float64
	for y := 1991; y <= 2025; y++ {
		if d := math.Abs(decToDec(shaped, y) - decToDec(anchor, y)); d > worst {
			worst = d
		}
	}
	if worst > 0.02 {
		t.Errorf("shaped DEVEXUS-USD calendar years differ from the anchor by up to %.2f point, want <= 0.02", worst)
	}

	naive := shapedSeries(anchor, shape)
	var loose float64
	for y := 1991; y <= 2025; y++ {
		if d := math.Abs(decToDec(naive, y) - decToDec(anchor, y)); d > loose {
			loose = d
		}
	}
	if loose < 0.2 {
		t.Errorf("un-aligned blend is off by only %.2f point: the alignMonthEnd guard no longer bites", loose)
	}
}

// TestMonthEndAnchorGate checks the registry, not the arithmetic: an anchor that
// is not declared month-end (the Treasury and euro-govt reconstructions, built
// on monthly AVERAGE yields dated the first of the month) must come back
// untouched, because snapping those to the month's last trading day would slide
// them the other way.
func TestMonthEndAnchorGate(t *testing.T) {
	anchor := &marketdata.Series{Points: []marketdata.Point{pt(2020, 1, 1, 100), pt(2020, 2, 1, 110)}}
	shape := &marketdata.Series{Points: []marketdata.Point{pt(2020, 1, 2, 50), pt(2020, 1, 31, 52)}}
	if got := alignMonthEnd("TREASURY-INT-USD", anchor, shape); got != anchor {
		t.Errorf("TREASURY-INT-USD was re-dated: %+v", got.Points)
	}
	got := alignMonthEnd("DEVEXUS-USD", anchor, shape)
	if !got.Points[0].Date.Equal(pt(2020, 1, 31, 0).Date) {
		t.Errorf("DEVEXUS-USD January anchor = %s, want the shape's 2020-01-31", got.Points[0].Date)
	}
}
