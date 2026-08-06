package simgen

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// yieldSeries builds a monthly yield series (percent) from consecutive months.
func yieldSeries(pct ...float64) *marketdata.Series {
	s := &marketdata.Series{}
	d := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, p := range pct {
		s.Points = append(s.Points, marketdata.Point{Date: d, Close: p})
		d = d.AddDate(0, 1, 0)
	}
	return s
}

// A par bond (coupon == yield) prices at exactly 100 for any maturity.
func TestBondPriceAtPar(t *testing.T) {
	for _, tc := range []struct{ y, n float64 }{{0.03, 5}, {0.08, 20}, {0.005, 30}} {
		if p := bondPrice(tc.y, tc.y, tc.n); math.Abs(p-100) > 1e-9 {
			t.Errorf("par bond y=%.3f n=%.0f priced %.6f, want 100", tc.y, tc.n, p)
		}
	}
}

// With a flat yield the total return is pure carry: ~y per year, compounding.
func TestTreasuryTRFlatYieldIsCarry(t *testing.T) {
	y := yieldSeries(6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6) // 12 monthly steps at 6%
	tr := TreasuryTR("int", y, 5, 0)
	got := tr.Points[len(tr.Points)-1].Close
	if want := 100 * math.Pow(1+0.06/12, 12); math.Abs(got-want) > 0.05 {
		t.Errorf("flat-yield index = %.4f, want carry-only ~%.4f", got, want)
	}
}

// A yield drop lifts the bond above carry; a symmetric maturity comparison
// shows the longer bond moves more for the same yield change (more duration).
func TestTreasuryTRDurationAndDirection(t *testing.T) {
	drop := yieldSeries(6, 5) // one month, yield falls 1pp
	short := TreasuryTR("5y", drop, 5, 0).Last().Close
	long := TreasuryTR("20y", drop, 20, 0).Last().Close
	if short <= 100 {
		t.Errorf("a falling yield should produce a gain, got %.4f", short)
	}
	if long <= short {
		t.Errorf("longer maturity should gain more on a rate drop: 20y=%.4f 5y=%.4f", long, short)
	}

	rise := yieldSeries(6, 7) // yield rises 1pp
	if loss := TreasuryTR("5y", rise, 5, 0).Last().Close; loss >= 100+0.06/12*100 {
		t.Errorf("a rising yield should produce a capital loss, got %.4f", loss)
	}
}

// The continuous fee drags the index below its no-fee counterpart.
func TestTreasuryTRFee(t *testing.T) {
	y := yieldSeries(5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5)
	gross := TreasuryTR("g", y, 5, 0).Last().Close
	net := TreasuryTR("n", y, 5, 0.002).Last().Close
	if net >= gross {
		t.Errorf("fee should lower the index: net=%.4f gross=%.4f", net, gross)
	}
	if drag := (gross - net) / gross; math.Abs(drag-0.002) > 5e-4 {
		t.Errorf("~1y of 0.2%%/yr fee dragged %.4f, want ~0.002", drag)
	}
}
