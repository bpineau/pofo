package fire

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

// mo is the last day of month m of year y (time.Date normalizes m past
// December), the month-end a monthly close is dated on.
func mo(y int, m time.Month) time.Time { return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC) }

func TestBuildMonthlyPanelAndFit(t *testing.T) {
	// 25 monthly points (24 returns) growing +1%/month, flat inflation.
	var pts, hicp []marketdata.Point
	v := 100.0
	for i := range 25 {
		d := mo(2000, time.January+time.Month(i))
		pts = append(pts, marketdata.Point{Date: d, Close: v})
		hicp = append(hicp, marketdata.Point{Date: d, Close: 100})
		v *= 1.01
	}
	a := AssetSeries{Weight: 1, Points: pts}
	panel, _, err := BuildMonthlyPanel([]AssetSeries{a}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 24 {
		t.Fatalf("monthly periods = %d, want 24", panel.Periods())
	}
	f := FitParametric(panel, []float64{1})
	// +1%/month compounds to (1.01^12 - 1) ≈ 12.68% real per year.
	want := math.Pow(1.01, 12) - 1
	if math.Abs(f.Mu-want) > 0.005 {
		t.Errorf("annualised mu = %.4f, want ~%.4f", f.Mu, want)
	}
	if f.Sigma < 0 {
		t.Errorf("sigma negative")
	}
}

// Sigma must come from the monthly dispersion scaled by √12, far more stable
// than the std of the ~20 annual points, and df must be seeded from the
// monthly excess kurtosis.
func TestFitParametricSigmaFromMonthly(t *testing.T) {
	// 24 monthly returns alternating ±5% around zero: a known monthly std.
	monthly := make([]float64, 24)
	for i := range monthly {
		if i%2 == 0 {
			monthly[i] = 0.05
		} else {
			monthly[i] = -0.05
		}
	}
	panel := scenario.Panel{Returns: [][]float64{monthly}, Weights: []float64{1}}
	f := FitParametric(panel, []float64{1})

	wantSigma := stdev(monthly) * math.Sqrt(12)
	if math.Abs(f.Sigma-wantSigma) > 1e-9 {
		t.Errorf("sigma = %.6f, want %.6f (monthly std × √12)", f.Sigma, wantSigma)
	}
}

func TestDofFromKurtosis(t *testing.T) {
	cases := []struct{ excess, want float64 }{
		{2, 7},    // 4 + 6/2
		{6, 5},    // 4 + 6/6
		{0.1, 30}, // very fat -> clamped to the slider max
		{0, 30},   // undefined / thin -> near-normal end
		{-1, 30},  // thin tails -> near-normal end
	}
	for _, c := range cases {
		if got := dofFromKurtosis(c.excess); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("dofFromKurtosis(%.2f) = %.2f, want %.2f", c.excess, got, c.want)
		}
	}
	if got := dofFromKurtosis(math.NaN()); got != 30 {
		t.Errorf("dofFromKurtosis(NaN) = %.2f, want 30", got)
	}
}

// Assets whose monthly grids do not line up (different start/end months) must
// be aligned on shared calendar months, not by trailing position. Asset A
// covers Jan-Apr (returns Feb,Mar,Apr); asset B covers Feb-May (returns
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
	panel, _, err := BuildMonthlyPanel([]AssetSeries{a, b}, hicp)
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
