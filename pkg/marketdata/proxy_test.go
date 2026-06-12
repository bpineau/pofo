package marketdata

import (
	"math"
	"testing"
	"time"
)

func d(y, m, day int) time.Time {
	return time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)
}

func TestExtendBack(t *testing.T) {
	asset := &Series{
		Symbol: "VOO",
		Points: []Point{
			{Date: d(2010, 9, 9), Close: 100},
			{Date: d(2010, 9, 10), Close: 101},
		},
	}
	proxy := &Series{
		Symbol: "^GSPC",
		Points: []Point{
			{Date: d(2000, 1, 3), Close: 40},
			{Date: d(2005, 6, 1), Close: 45},
			{Date: d(2010, 9, 9), Close: 50}, // anchor: 100/50 → scale ×2
			{Date: d(2010, 9, 10), Close: 51},
		},
	}
	if !ExtendBack(asset, proxy) {
		t.Fatal("the series should have been extended")
	}
	if len(asset.Points) != 4 {
		t.Fatalf("expected 4 points, found %d", len(asset.Points))
	}
	if math.Abs(asset.Points[0].Close-80) > 1e-12 || math.Abs(asset.Points[1].Close-90) > 1e-12 {
		t.Errorf("incorrect rescaling: %+v", asset.Points[:2])
	}
	if !asset.SimulatedBefore.Equal(d(2010, 9, 9)) || asset.ProxySymbol != "^GSPC" {
		t.Errorf("simulation metadata: %+v", asset)
	}
	// Idempotent: a second extension does nothing.
	if ExtendBack(asset, proxy) {
		t.Error("an already-extended series must not be extended again")
	}
}

func TestExtendBackNoEarlierData(t *testing.T) {
	asset := &Series{Symbol: "X", Points: []Point{{Date: d(2010, 1, 1), Close: 10}}}
	proxy := &Series{Symbol: "P", Points: []Point{{Date: d(2015, 1, 1), Close: 5}}}
	if ExtendBack(asset, proxy) {
		t.Error("no extension possible when the proxy starts after the asset")
	}
}
