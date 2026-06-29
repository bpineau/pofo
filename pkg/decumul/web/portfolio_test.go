package web

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func mo(y int, m time.Month) time.Time { return time.Date(y, m, 28, 0, 0, 0, 0, time.UTC) }

func TestBuildMonthlyPanelAndFit(t *testing.T) {
	// 25 monthly points (24 returns) growing +1%/month, flat inflation.
	var pts, hicp []marketdata.Point
	v := 100.0
	d := mo(2000, time.January)
	for i := 0; i < 25; i++ {
		pts = append(pts, marketdata.Point{Date: d, Close: v})
		hicp = append(hicp, marketdata.Point{Date: d, Close: 100})
		v *= 1.01
		d = d.AddDate(0, 1, 0)
	}
	a := AssetSeries{Weight: 1, Points: pts}
	panel, err := BuildMonthlyPanel([]AssetSeries{a}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 24 {
		t.Fatalf("monthly periods = %d, want 24", panel.Periods())
	}
	mu, sigma := FitParametric(panel, []float64{1})
	// +1%/month compounds to (1.01^12 - 1) ≈ 12.68% real per year.
	want := math.Pow(1.01, 12) - 1
	if math.Abs(mu-want) > 0.005 {
		t.Errorf("annualised mu = %.4f, want ~%.4f", mu, want)
	}
	if sigma < 0 {
		t.Errorf("sigma negative")
	}
}

// Assets whose monthly grids do not line up (different start/end months) must
// be aligned on shared calendar months, not by trailing position. Asset A
// covers Jan–Apr (returns Feb,Mar,Apr); asset B covers Feb–May (returns
// Mar,Apr,May). The common months are Mar and Apr, so each row must hold those
// two returns in order, not three position-truncated ones.
func TestBuildMonthlyPanelDateKeyed(t *testing.T) {
	hicp := []marketdata.Point{{Date: mo(1999, time.January), Close: 100}}
	a := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.January), Close: 100},
		{Date: mo(2000, time.February), Close: 110}, // Feb +0.10
		{Date: mo(2000, time.March), Close: 132},    // Mar +0.20
		{Date: mo(2000, time.April), Close: 145.2},  // Apr +0.10
	}}
	b := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.February), Close: 200},
		{Date: mo(2000, time.March), Close: 210}, // Mar +0.05
		{Date: mo(2000, time.April), Close: 252}, // Apr +0.20
		{Date: mo(2000, time.May), Close: 277.2}, // May +0.10
	}}
	panel, err := BuildMonthlyPanel([]AssetSeries{a, b}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 2 {
		t.Fatalf("periods = %d, want 2 (Mar, Apr common)", panel.Periods())
	}
	wantA := []float64{0.20, 0.10} // Mar, Apr
	wantB := []float64{0.05, 0.20} // Mar, Apr
	for k := range wantA {
		if math.Abs(panel.Returns[0][k]-wantA[k]) > 1e-9 {
			t.Errorf("asset A month %d = %.4f, want %.4f", k, panel.Returns[0][k], wantA[k])
		}
		if math.Abs(panel.Returns[1][k]-wantB[k]) > 1e-9 {
			t.Errorf("asset B month %d = %.4f, want %.4f", k, panel.Returns[1][k], wantB[k])
		}
	}
}

// Internal gaps must not produce a spanning return masquerading as a one-month
// return: a month missing from one asset drops only that month from the common
// grid, and the multi-month return across the gap is excluded.
func TestBuildMonthlyPanelSkipsGaps(t *testing.T) {
	hicp := []marketdata.Point{{Date: mo(1999, time.January), Close: 100}}
	full := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.January), Close: 100},
		{Date: mo(2000, time.February), Close: 110},
		{Date: mo(2000, time.March), Close: 121},
		{Date: mo(2000, time.April), Close: 133.1},
	}}
	gapped := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: mo(2000, time.January), Close: 100},
		{Date: mo(2000, time.February), Close: 110}, // Feb +0.10 (Jan->Feb)
		// March missing -> the Feb->Apr return spans two months, excluded.
		{Date: mo(2000, time.April), Close: 132}, // Apr (Feb->Apr, spanning)
	}}
	panel, err := BuildMonthlyPanel([]AssetSeries{full, gapped}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	// Only February is a true one-month return common to both.
	if panel.Periods() != 1 {
		t.Fatalf("periods = %d, want 1 (only Feb is consecutive in both)", panel.Periods())
	}
	if math.Abs(panel.Returns[1][0]-0.10) > 1e-9 {
		t.Errorf("gapped asset Feb return = %.4f, want 0.10", panel.Returns[1][0])
	}
}
