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

func TestProxySymbol(t *testing.T) {
	cases := map[string]string{
		"SPY":          "^GSPC",
		"VTI":          "^GSPC",
		"QQQ":          "^NDX",
		"GLD":          "GC=F",
		"IE000KF370H3": "NTSX",
		"VOOX":         "", // not a proxied asset
	}
	for in, want := range cases {
		got, ok := ProxySymbol(in)
		if got != want || ok != (want != "") {
			t.Errorf("ProxySymbol(%q) = %q, %v; want %q, %v", in, got, ok, want, want != "")
		}
	}
}

// TestExtendBackRefusals: every reason the splice must decline. Each of them
// would otherwise fabricate history, which is worse than a short series.
func TestExtendBackRefusals(t *testing.T) {
	cases := []struct {
		name         string
		asset, proxy *Series
	}{
		{
			name:  "the asset has no quotes to anchor on",
			asset: &Series{Symbol: "X"},
			proxy: &Series{Symbol: "P", Points: []Point{{Date: d(2000, 1, 3), Close: 5}}},
		},
		{
			name:  "the proxy has no quotes",
			asset: &Series{Symbol: "X", Points: []Point{{Date: d(2010, 1, 4), Close: 10}}},
			proxy: &Series{Symbol: "P"},
		},
		{
			name: "the asset is already extended",
			asset: &Series{Symbol: "X", SimulatedBefore: d(2010, 1, 4),
				Points: []Point{{Date: d(2010, 1, 4), Close: 10}}},
			proxy: &Series{Symbol: "P", Points: []Point{{Date: d(2000, 1, 3), Close: 5}}},
		},
		{
			name:  "the proxy value at the anchor is not positive",
			asset: &Series{Symbol: "X", Points: []Point{{Date: d(2010, 1, 4), Close: 10}}},
			proxy: &Series{Symbol: "P", Points: []Point{
				{Date: d(2000, 1, 3), Close: 5}, {Date: d(2010, 1, 4), Close: 0}}},
		},
		{
			name:  "the proxy starts on the anchor day, so nothing precedes it",
			asset: &Series{Symbol: "X", Points: []Point{{Date: d(2010, 1, 4), Close: 10}}},
			proxy: &Series{Symbol: "P", Points: []Point{
				{Date: d(2010, 1, 4), Close: 5}, {Date: d(2010, 1, 5), Close: 6}}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := len(tc.asset.Points)
			if ExtendBack(tc.asset, tc.proxy) {
				t.Fatalf("the splice should have declined: %+v", tc.asset.Points)
			}
			if len(tc.asset.Points) != before {
				t.Errorf("a declined splice must leave the series alone: %+v", tc.asset.Points)
			}
		})
	}
}
