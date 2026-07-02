package simgen

import (
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// gold1968 is the earliest month the bundled LBMA gold fix must cover.
var gold1968 = time.Date(1969, 1, 1, 0, 0, 0, 0, time.UTC)

// xauusdBuild splices the bundled monthly LBMA gold fix (XAUUSD-LBMA, 1968→)
// behind the fetchable daily spot, so a gold sleeve reaches the floating era.
func TestXAUUSDBuildSplicesBundledLBMA(t *testing.T) {
	daily := atSeries("XAUUSD", 100, 50, 275) // recent daily quote only
	f := WithRefData(datasets.Refdata(), fakeFetcher{"XAUUSD": daily})

	got, err := xauusdBuild(f, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.First().Date.Before(gold1968) {
		t.Errorf("spliced gold starts %s, want ≤1968 from the bundled LBMA fix", got.First().Date.Format("2006-01"))
	}
	if got.SimulatedBefore.IsZero() {
		t.Error("expected SimulatedBefore after splicing the LBMA fix")
	}
	if last := got.Last().Close; last != 275 {
		t.Errorf("recent tail = %v, want the daily quote (275) preserved", last)
	}
}

// The managed-futures crude leg (CL=F, ~2000) is extended by the bundled
// monthly WTI spot (WTI-USD, ~1946) through the standard longBack splice.
func TestExtendCLFWithBundledWTI(t *testing.T) {
	clf := atSeries("CL=F", 100, 50, 30) // recent daily quote only
	f := extend(WithRefData(datasets.Refdata(), fakeFetcher{"CL=F": clf}))

	got, err := f.Fetch("CL=F", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC); !got.First().Date.Before(want) {
		t.Errorf("extended crude starts %s, want the 1940s WTI spot", got.First().Date.Format("2006-01"))
	}
	if got.SimulatedBefore.IsZero() {
		t.Error("expected SimulatedBefore after splicing WTI-USD")
	}
}

// The developed-ex-US equity leg (VTMGX, 1999) is extended by the bundled
// DEVEXUS-USD series (French dev-ex-US, MSCI World before, ~1969) and the
// emerging leg (VEIEX, 1994) by EM-USD (~1989). These are the legs that
// actually cap NTSG/DBMF, so their real bundled proxies matter most.
func TestExtendIntlEquityWithBundledProxies(t *testing.T) {
	for _, tc := range []struct {
		leg  string
		want int // latest acceptable start year
	}{
		{"VTMGX", 1970}, // dev-ex-US reaches ~1969 via the MSCI World backfill
		{"VEIEX", 1990}, // emerging reaches ~1989
	} {
		quotes := atSeries(tc.leg, 100, 50, 100) // recent fund quotes only
		f := extend(WithRefData(datasets.Refdata(), fakeFetcher{tc.leg: quotes}))
		got, err := f.Fetch(tc.leg, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		if want := time.Date(tc.want, 1, 1, 0, 0, 0, 0, time.UTC); !got.First().Date.Before(want) {
			t.Errorf("%s extended to %s, want before %d", tc.leg, got.First().Date.Format("2006-01"), tc.want)
		}
		if got.SimulatedBefore.IsZero() {
			t.Errorf("%s: expected SimulatedBefore after splicing its bundled proxy", tc.leg)
		}
	}
}

// The US-equity leg (VFINX, 1976) is extended by the bundled S&P 500 total
// return (SP500-USD, ~1871) and the cash rate (^IRX) by the 3-month T-bill
// (TBILL-3M, ~1934), removing the last US caps on NTSG/NTSX.
func TestExtendUSLegsWithBundledProxies(t *testing.T) {
	for _, tc := range []struct {
		leg   string
		level float64
		want  int
	}{
		{"VFINX", 100, 1900}, // S&P 500 total return to ~1871
		{"^IRX", 5, 1940},    // 3-month bill rate to ~1934 (a rate; realistic ~5% level)
	} {
		quotes := atSeries(tc.leg, 100, 50, tc.level)
		f := extend(WithRefData(datasets.Refdata(), fakeFetcher{tc.leg: quotes}))
		got, err := f.Fetch(tc.leg, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		if want := time.Date(tc.want, 1, 1, 0, 0, 0, 0, time.UTC); !got.First().Date.Before(want) {
			t.Errorf("%s extended to %s, want before %d", tc.leg, got.First().Date.Format("2006-01"), tc.want)
		}
		if got.SimulatedBefore.IsZero() {
			t.Errorf("%s: expected SimulatedBefore after splicing its bundled proxy", tc.leg)
		}
	}
}

// The intermediate-treasury leg (VFITX, 1991) is extended by the bundled
// constant-maturity Treasury total-return reconstruction (~1953).
func TestExtendVFITXWithBundledTreasury(t *testing.T) {
	vfitx := atSeries("VFITX", 100, 50, 120) // recent fund quotes only
	f := extend(WithRefData(datasets.Refdata(), fakeFetcher{"VFITX": vfitx}))

	got, err := f.Fetch("VFITX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC); !got.First().Date.Before(want) {
		t.Errorf("extended treasury starts %s, want the 1950s CMT reconstruction", got.First().Date.Format("2006-01"))
	}
	if got.SimulatedBefore.IsZero() {
		t.Error("expected SimulatedBefore after splicing TREASURY-INT-USD")
	}
}

// Without a fetchable daily quote, xauusdBuild returns the bundled monthly fix
// on its own rather than failing.
func TestXAUUSDBuildFallsBackToBundledLBMA(t *testing.T) {
	f := WithRefData(datasets.Refdata(), fakeFetcher{}) // no daily XAUUSD

	got, err := xauusdBuild(f, time.Time{})
	if err != nil || got == nil {
		t.Fatalf("got %v, err %v; want the bundled monthly fix", got, err)
	}
	if !got.First().Date.Before(gold1968) {
		t.Errorf("fallback gold starts %s, want ≤1968", got.First().Date.Format("2006-01"))
	}
}

// atSeries builds a daily series of n points from a start offset, constant level.
func atSeries(symbol string, startOffset, n int, level float64) *marketdata.Series {
	s := &marketdata.Series{Symbol: symbol}
	for i := range n {
		s.Points = append(s.Points, marketdata.Point{Date: day(startOffset + i), Close: level})
	}
	return s
}

// A configured component is spliced with its longer proxy: the returned series
// starts at the proxy's first date, before the component's own inception.
func TestExtendingFetcherSplicesConfiguredComponent(t *testing.T) {
	f := fakeFetcher{
		"VTMGX":       atSeries("VTMGX", 100, 50, 200),   // starts at day(100)
		"DEVEXUS-USD": atSeries("dev-ex-US", 0, 200, 50), // starts at day(0), earlier
	}

	got, err := extend(f).Fetch("VTMGX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Points[0].Date.Equal(day(0)) {
		t.Errorf("extended series starts %v, want the proxy's start %v", got.Points[0].Date, day(0))
	}
	if got.SimulatedBefore.IsZero() {
		t.Errorf("expected SimulatedBefore to be set after splicing")
	}
}

// A component with no proxy, or a missing proxy, is returned unchanged (the
// wrapper is safe to apply unconditionally).
func TestExtendingFetcherLeavesOthersUnchanged(t *testing.T) {
	f := fakeFetcher{"VBMFX": atSeries("VBMFX", 0, 50, 100)} // not in longBack

	got, err := extend(f).Fetch("VBMFX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Points) != 50 || !got.SimulatedBefore.IsZero() {
		t.Errorf("unconfigured component should pass through unchanged, got %d points, SimulatedBefore=%v", len(got.Points), got.SimulatedBefore)
	}
}

// A proxy with a configured daily-shape series is blended before splicing:
// the extended component gets daily granularity where the shape covers the
// proxy, exact proxy levels at the anchors, and the untouched monthly proxy
// before that.
func TestExtendingFetcherShapesProxy(t *testing.T) {
	// Component: daily from 2000-01-01. Proxy anchors: monthly 1990->2000,
	// level 100 flat then a jump. Shape: daily 1995->2000 with wiggle.
	comp := &marketdata.Series{Symbol: "VTMGX"}
	cd := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 30 {
		comp.Points = append(comp.Points, marketdata.Point{Date: cd.AddDate(0, 0, i), Close: 20})
	}
	proxy := &marketdata.Series{Symbol: "DEVEXUS-USD"}
	pd := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	v := 100.0
	for range 121 { // through 2000-01
		proxy.Points = append(proxy.Points, marketdata.Point{Date: pd, Close: v})
		pd = pd.AddDate(0, 1, 0)
		v *= 1.005
	}
	shape := &marketdata.Series{Symbol: "DEVEXUS-DAILY"}
	sd := time.Date(1995, 1, 2, 0, 0, 0, 0, time.UTC)
	for i := range 1900 { // daily through 2000-03
		shape.Points = append(shape.Points, marketdata.Point{Date: sd.AddDate(0, 0, i), Close: 50 + float64(i%7)})
	}
	f := extend(fakeFetcher{"VTMGX": comp, "DEVEXUS-USD": proxy, "DEVEXUS-DAILY": shape})

	got, err := f.Fetch("VTMGX", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if first := got.First().Date; !first.Equal(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("first = %s, want the proxy's 1990-01 start", first.Format("2006-01-02"))
	}
	// Daily density in the shape era: 1996 must hold ~365 points, not 12.
	days := 0
	for _, p := range got.Points {
		if p.Date.Year() == 1996 {
			days++
		}
	}
	if days < 300 {
		t.Errorf("1996 carries %d points, want daily density from the shape", days)
	}
	// And monthly density before the shape starts.
	months := 0
	for _, p := range got.Points {
		if p.Date.Year() == 1992 {
			months++
		}
	}
	if months != 12 {
		t.Errorf("1992 carries %d points, want the 12 monthly proxy anchors", months)
	}
}

// msciWorld uses the long refdata series when present (net of fee) and does
// not touch the fallback; without the daily shape the backcast stays monthly.
func TestMSCIWorldPrefersRefdata(t *testing.T) {
	f := fakeFetcher{"MSCIWORLD-USD": atSeries("MSCIWORLD-USD", 0, 400, 100)}
	b := msciWorld(0, func(Fetcher, time.Time) (*marketdata.Series, error) {
		t.Fatal("fallback must not run when the refdata series is present")
		return nil, nil
	})
	got, err := b(f, time.Time{})
	if err != nil || got == nil || len(got.Points) != 400 {
		t.Fatalf("got %v points, err %v; want the 400-point refdata series", len(got.Points), err)
	}
}

// Without the refdata series, msciWorld falls back to the proxy Build.
func TestMSCIWorldFallsBack(t *testing.T) {
	called := false
	b := msciWorld(0, func(Fetcher, time.Time) (*marketdata.Series, error) {
		called = true
		return atSeries("proxy", 0, 10, 1), nil
	})
	if _, err := b(fakeFetcher{}, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("expected the proxy fallback when refdata is absent")
	}
}

// afterFee applies a continuous annual drag.
func TestAfterFee(t *testing.T) {
	s := atSeries("x", 0, 366, 100) // ~1 year of constant level
	out := afterFee(s, 0.02)
	if last := out.Points[len(out.Points)-1].Close; last < 97.8 || last > 98.2 {
		t.Errorf("after 2%%/yr fee over ~1y, level = %.3f, want ~98", last)
	}
}
