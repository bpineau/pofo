package web

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func yr(y int) time.Time { return time.Date(y, 6, 30, 0, 0, 0, 0, time.UTC) }

func TestBuildPanelAndFit(t *testing.T) {
	a := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: yr(2000), Close: 100}, {Date: yr(2001), Close: 110}, {Date: yr(2002), Close: 121},
	}}
	hicp := []marketdata.Point{{Date: yr(2000), Close: 100}, {Date: yr(2001), Close: 100}, {Date: yr(2002), Close: 100}}
	panel, err := BuildPanel([]AssetSeries{a}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 2 {
		t.Fatalf("periods = %d, want 2", panel.Periods())
	}
	mu, sigma := FitParametric(panel, []float64{1})
	if math.Abs(mu-0.10) > 0.01 {
		t.Errorf("mu = %.4f, want ~0.10", mu)
	}
	if sigma < 0 {
		t.Errorf("sigma negative")
	}
}
