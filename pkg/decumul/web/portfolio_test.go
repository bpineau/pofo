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
