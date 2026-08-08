package simgen

import (
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
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

// Custom builders (not just the shared composite/tsmom) must route their legs
// through extend() too, otherwise their commodity and intl-equity legs stay
// short and cap the recipe at their own inception. Here the trend legs quote
// only from 2005 and their bundled proxies reach the 1980s, so an extended
// frame is capped by the equity/cash legs (2003) plus the TSMOM warm-up; an
// un-extended one would not start before 2005. The dates stay well after the
// trend reference's own 2000 start, so this measures the routing and not the
// truncation, which the next test covers.
func TestWintonBuilderExtendsItsLegs(t *testing.T) {
	fake := fakeFetcher{
		"VFINX": dailyFrom("VFINX", 2003, 1, 1, 8400),
		"^IRX":  dailyFrom("^IRX", 2003, 1, 1, 8400),
	}
	for _, id := range []string{"VTMGX", "VEIEX", "VFITX", "VUSTX", "GC=F", "CL=F"} {
		fake[id] = dailyFrom(id, 2005, 1, 1, 7600)
	}
	f := WithRefData(datasets.Refdata(), fake)

	s, err := wintonBuild(f, ComponentsFrom)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.First().Date.Year(); got > 2004 {
		t.Errorf("Winton starts %s, want ≤2004 (its legs must be extended via extend())", s.First().Date.Format("2006-01"))
	}
}

// The overlay builds take their LEVEL from the pure-trend reference, so they
// must not ship a day in front of it however deep their legs reach: what stands
// there is the engine alone. Same fixture as the extension test but with legs
// running back to the 1980s, where the truncation is what binds.
func TestOverlayBuildsStopWhereTheirReferenceDoes(t *testing.T) {
	fake := fakeFetcher{
		"VFINX": dailyFrom("VFINX", 1985, 1, 1, 15100),
		"VFITX": dailyFrom("VFITX", 1985, 1, 1, 15100),
		"^IRX":  dailyFrom("^IRX", 1985, 1, 1, 15100),
	}
	for _, id := range []string{"VTMGX", "VEIEX", "VUSTX", "GC=F", "CL=F"} {
		fake[id] = dailyFrom(id, 1994, 1, 1, 11800)
	}
	f := WithRefData(datasets.Refdata(), fake)
	want, err := AnchorStart(f, PureTrendAnchor)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name  string
		build func(Fetcher, time.Time) (*marketdata.Series, error)
	}{
		{"Winton", wintonBuild},
		{"RSST", stackedTrend("RSST", "VFINX", vfinxTER, mfConfig(0.10, 0), 0.0096)},
		{"RSBT", stackedTrend("RSBT", "VFITX", vfitxTER, mfConfig(0.10, 0), 0.0097)},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, err := c.build(f, ComponentsFrom)
			if err != nil {
				t.Fatal(err)
			}
			if got := s.First().Date; got.Before(want) {
				t.Errorf("starts %s, before the reference's own %s",
					got.Format("2006-01-02"), want.Format("2006-01-02"))
			}
			if got := s.First().Date; got.After(want.AddDate(0, 0, 7)) {
				t.Errorf("starts %s, a week or more after the reference's own %s: "+
					"the truncation is not what binds and this test proves nothing",
					got.Format("2006-01-02"), want.Format("2006-01-02"))
			}
		})
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
