package scenario

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func d(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }

func TestDeflateRemovesInflation(t *testing.T) {
	// nominal +10%/yr, inflation +10%/yr -> ~0 real return.
	prices := []marketdata.Point{{Date: d(2000), Close: 100}, {Date: d(2001), Close: 110}, {Date: d(2002), Close: 121}}
	hicp := []marketdata.Point{{Date: d(2000), Close: 100}, {Date: d(2001), Close: 110}, {Date: d(2002), Close: 121}}
	got := Deflate(prices, hicp)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for i, r := range got {
		if math.Abs(r) > 1e-9 {
			t.Errorf("real return[%d] = %.6f, want ~0", i, r)
		}
	}
}
