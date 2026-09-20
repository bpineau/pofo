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

// A negative yield is a price, not a gap. Japan's benchmark spent four years
// below zero and rallied hard doing it, so the reconstruction has to keep
// pricing: a bond bought at -0.10 % and marked at -0.50 % has gained, and an
// implementation that skips sub-zero days would report a flat line instead.
func TestTreasuryTRPricesNegativeYields(t *testing.T) {
	drop := yieldSeries(-0.10, -0.50)
	got := TreasuryTR("jgb", drop, 10, 0).Last().Close
	if got <= 100 {
		t.Errorf("a yield falling from -0.10%% to -0.50%% should gain, got %.4f", got)
	}
	// Roughly duration x the yield move, less a month of (negative) carry.
	if want := 100 * (1 + 0.004*9.6); math.Abs(got-want) > 0.5 {
		t.Errorf("negative-yield gain = %.4f, want ~%.4f", got, want)
	}
	rise := yieldSeries(-0.50, -0.10)
	if loss := TreasuryTR("jgb", rise, 10, 0).Last().Close; loss >= 100 {
		t.Errorf("a yield rising back towards zero should lose, got %.4f", loss)
	}
}

// A zero-coupon strip prices at the semiannual discount factor, and at par
// (100) when it has no time left to run.
func TestStripPrice(t *testing.T) {
	if p := stripPrice(0.06, 0); p != 1 {
		t.Errorf("a matured strip priced %.6f, want 1", p)
	}
	if p, want := stripPrice(0.06, 10), math.Pow(1.03, -20); math.Abs(p-want) > 1e-12 {
		t.Errorf("10y strip at 6%% priced %.6f, want %.6f", p, want)
	}
	// Twice the maturity, the square of the discount factor.
	if p, q := stripPrice(0.05, 13.5), stripPrice(0.05, 27); math.Abs(p*p-q) > 1e-12 {
		t.Errorf("strip prices are not multiplicative: %.9f² != %.9f", p, q)
	}
}

// With a flat yield a strip earns pure carry, like any other bond: the pull to
// par IS the return, and it compounds at the yield.
func TestTreasuryZeroTRFlatYieldIsCarry(t *testing.T) {
	y := yieldSeries(6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6) // 12 monthly steps at 6%
	got := TreasuryZeroTR("strip", y, 27, 0).Last().Close
	if want := 100 * math.Pow(1.03, 2); math.Abs(got-want) > 0.05 {
		t.Errorf("flat-yield strip index = %.4f, want carry-only ~%.4f", got, want)
	}
}

// The reason the engine exists: a zero's duration barely moves with the yield
// level while a par bond's collapses, so no constant gearing of the coupon bond
// can stand in for the strip across rate regimes.
func TestTreasuryZeroTRDurationHoldsWhenYieldsAreHigh(t *testing.T) {
	// One month, yields 10 basis points lower, at two yield levels.
	low, high := yieldSeries(3, 2.9), yieldSeries(12, 11.9)
	zeroLow := TreasuryZeroTR("z", low, 27, 0).Last().Close - 100 - 0.03/12*100
	zeroHigh := TreasuryZeroTR("z", high, 27, 0).Last().Close - 100 - 0.12/12*100
	parLow := TreasuryTR("p", low, 22, 0).Last().Close - 100 - 0.03/12*100
	parHigh := TreasuryTR("p", high, 22, 0).Last().Close - 100 - 0.12/12*100

	if ratio := zeroHigh / zeroLow; ratio < 0.9 || ratio > 1.02 {
		t.Errorf("a strip's sensitivity should hold across yield levels, high/low = %.3f", ratio)
	}
	if ratio := parHigh / parLow; ratio > 0.6 {
		t.Errorf("a par bond's sensitivity should collapse at high yields, high/low = %.3f", ratio)
	}
	if gLow, gHigh := zeroLow/parLow, zeroHigh/parHigh; gHigh < 1.7*gLow {
		t.Errorf("the strip-over-coupon gearing should roughly double from 3%% to 12%% yields, got %.2f then %.2f", gLow, gHigh)
	}
}

// The strip is longer than any par bond of the same maturity, and the fee drags
// it the same way it drags the coupon reconstruction.
func TestTreasuryZeroTRLongerThanParAndNetOfFee(t *testing.T) {
	drop := yieldSeries(5, 4)
	zero := TreasuryZeroTR("z", drop, 27, 0).Last().Close
	par := TreasuryTR("p", drop, 27, 0).Last().Close
	if zero <= par {
		t.Errorf("a 27y strip should gain more than a 27y par bond: %.4f vs %.4f", zero, par)
	}

	flat := yieldSeries(5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5)
	gross := TreasuryZeroTR("g", flat, 27, 0).Last().Close
	net := TreasuryZeroTR("n", flat, 27, 0.0015).Last().Close
	if drag := (gross - net) / gross; math.Abs(drag-0.0015) > 5e-4 {
		t.Errorf("~1y of 0.15%%/yr fee dragged %.5f, want ~0.0015", drag)
	}
}

// Zero is the annuity factor's singular point and nothing else: the par bond
// still prices at 100 there, and a series crossing zero keeps moving.
func TestBondPriceAtZeroYield(t *testing.T) {
	if p := bondPrice(0, 0, 10); math.Abs(p-100) > 1e-9 {
		t.Errorf("zero-coupon zero-yield par bond priced %.6f, want 100", p)
	}
	if p := bondPrice(0, 0.02, 10); math.Abs(p-120) > 1e-9 {
		t.Errorf("2%% coupon at a zero yield priced %.6f, want 120", p)
	}
	crossing := yieldSeries(0.20, 0, -0.20)
	tr := TreasuryTR("crossing", crossing, 10, 0)
	if tr.Points[1].Close <= 100 || tr.Points[2].Close <= tr.Points[1].Close {
		t.Errorf("a yield falling through zero should gain every step, got %v", tr.Points)
	}
}

// A definitional junction earns nothing: the step into it carries the index
// forward unchanged whatever the yield does across it, while the steps on
// either side price normally. This is the 1973-01-04 repair, in miniature: the
// yield jumps a whole point on the junction day and the index must not notice.
func TestConstantMaturityTRSkipsJunctions(t *testing.T) {
	ys := yieldSeries(5, 5, 6, 6, 6)
	ys.Junctions = []time.Time{ys.Points[2].Date}
	tr := TreasuryTR("junction", ys, 20, 0)

	if n := len(tr.Points); n != 5 {
		t.Fatalf("got %d points, want one per yield observation", n)
	}
	if got := tr.Points[2].Close / tr.Points[1].Close; math.Abs(got-1) > 1e-12 {
		t.Errorf("the junction step returns %+.6f %%, want exactly zero", (got-1)*100)
	}
	// The month before the junction is pure carry at 5 %, the month after pure
	// carry at 6 %: both priced, neither skipped.
	for _, tc := range []struct {
		i    int
		want float64
	}{{1, 0.05}, {3, 0.06}} {
		dt := tr.Points[tc.i].Date.Sub(tr.Points[tc.i-1].Date).Hours() / 24 / 365.25
		got := tr.Points[tc.i].Close/tr.Points[tc.i-1].Close - 1
		if math.Abs(got-tc.want*dt) > 1e-9 {
			t.Errorf("step %d returns %+.6f, want the period's carry %+.6f", tc.i, got, tc.want*dt)
		}
	}

	// Without the declaration the same yields fabricate a double-digit loss.
	plain := TreasuryTR("no junction", yieldSeries(5, 5, 6, 6, 6), 20, 0)
	if r := plain.Points[2].Close/plain.Points[1].Close - 1; r > -0.10 {
		t.Errorf("the undeclared jump returns %+.2f %%, expected a large fabricated loss", r*100)
	}
}

// A junction inside a longer step (the month-end case) is caught too: the whole
// period is refused rather than priced through the break.
func TestConstantMaturityTRJunctionInsideAStep(t *testing.T) {
	ys := yieldSeries(5, 5, 6)
	ys.Junctions = []time.Time{ys.Points[1].Date.AddDate(0, 0, 10)}
	tr := TreasuryTR("junction inside", ys, 20, 0.01)
	if got := tr.Points[2].Close / tr.Points[1].Close; math.Abs(got-1) > 1e-12 {
		t.Errorf("a step spanning a junction returns %+.6f %%, want zero and no fee", (got-1)*100)
	}
}
