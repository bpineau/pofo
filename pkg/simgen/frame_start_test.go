package simgen

import (
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// monthlyFrom builds a monthly series of n points from year/month at a flat
// level (only the date range matters for these tests).
func monthlyFrom(sym string, year, month, n int) *marketdata.Series {
	s := &marketdata.Series{Symbol: sym}
	d := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: 100 + float64(i)})
		d = d.AddDate(0, 1, 0)
	}
	return s
}

// dailyFrom builds a daily series of n points from year/month/day.
func dailyFrom(sym string, year, month, dayN, n int) *marketdata.Series {
	s := &marketdata.Series{Symbol: sym}
	d := time.Date(year, time.Month(month), dayN, 0, 0, 0, 0, time.UTC)
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: 100 + float64(i)})
		d = d.AddDate(0, 0, 1)
	}
	return s
}

// ntsgLegs mirrors the NTSG composite (US + dev-intl equity + treasury excess).
var ntsgLegs = []Leg{
	{ID: "VFINX", Weight: 0.54},
	{ID: "VTMGX", Weight: 0.36},
	{ID: "VFITX", Weight: 0.60, Excess: true},
	{ID: "^IRX", Weight: 0.10},
}

// A multi-leg composite starts at its YOUNGEST leg's first quote: extending the
// other legs earlier does nothing. Here VTMGX (dev-intl, 1999) has no working
// proxy, so it caps the whole recipe at 1999 even though VFINX reaches 1976 and
// VFITX 1953. This reproduces the observed NTSG start of 1999.
func TestCompositeCappedByYoungestLeg(t *testing.T) {
	f := fakeFetcher{
		"VFINX": monthlyFrom("VFINX", 1976, 1, 600),
		"VTMGX": monthlyFrom("VTMGX", 1999, 1, 330), // youngest, no proxy available
		"VFITX": monthlyFrom("VFITX", 1953, 1, 900),
		"^IRX":  monthlyFrom("^IRX", 1953, 1, 900),
		// TREASURY-INT-USD present but irrelevant: VFITX is not the cap.
		"TREASURY-INT-USD": monthlyFrom("TREASURY-INT-USD", 1953, 1, 900),
	}
	s, err := composite("NTSG", ntsgLegs, "^IRX", 0)(f, ComponentsFrom)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.First().Date.Year(); got != 1999 {
		t.Fatalf("composite starts %d, want 1999 (capped by the un-extended VTMGX leg)", got)
	}
}

// Give VTMGX its bundled long proxy (DEVEXUS-USD) and the recipe reaches its
// next-youngest leg, VFINX (1976). This is the fix NTSG needs, and it also
// exercises a monthly proxy forward-filling into the frame.
func TestCompositeUnlockedWhenYoungestLegExtended(t *testing.T) {
	f := fakeFetcher{
		"VFINX":            monthlyFrom("VFINX", 1976, 1, 600),
		"VTMGX":            monthlyFrom("VTMGX", 1999, 1, 330),
		"VFITX":            monthlyFrom("VFITX", 1953, 1, 900),
		"^IRX":             monthlyFrom("^IRX", 1953, 1, 900),
		"DEVEXUS-USD":      monthlyFrom("dev-ex-US", 1969, 1, 700), // the proxy longBack["VTMGX"] wants
		"TREASURY-INT-USD": monthlyFrom("TREASURY-INT-USD", 1953, 1, 900),
	}
	s, err := composite("NTSG", ntsgLegs, "^IRX", 0)(f, ComponentsFrom)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.First().Date.Year(); got != 1976 {
		t.Fatalf("composite starts %d, want 1976 (VFINX) once VTMGX is extended", got)
	}
}

// A monthly-extended leg mixed with a daily leg forward-fills cleanly: the
// frame runs on the union of dates and the composite is finite throughout.
func TestCompositeMixedCadenceForwardFills(t *testing.T) {
	f := fakeFetcher{
		"VFINX":            dailyFrom("VFINX", 1998, 1, 1, 1500), // daily
		"VTMGX":            dailyFrom("VTMGX", 1999, 1, 1, 1000), // daily
		"VFITX":            monthlyFrom("VFITX", 1998, 1, 60),    // monthly
		"^IRX":             monthlyFrom("^IRX", 1998, 1, 60),     // monthly
		"DEVEXUS-USD":      monthlyFrom("dev-ex-US", 1969, 1, 700),
		"TREASURY-INT-USD": monthlyFrom("TREASURY-INT-USD", 1953, 1, 900),
	}
	s, err := composite("NTSG", ntsgLegs, "^IRX", 0)(f, ComponentsFrom)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Points) < 2 {
		t.Fatal("empty composite")
	}
	for i, p := range s.Points {
		if p.Close <= 0 || p.Close > 1e9 {
			t.Fatalf("non-finite composite value %v at point %d (%s)", p.Close, i, p.Date.Format("2006-01-02"))
		}
	}
}
