package firebook

import (
	"math"
	"strings"
	"testing"
)

// closedFormMacaulay is the textbook closed form of the Macaulay duration of a
// bond with n whole years left, annual coupon c and yield y. The plate reaches
// the same number by summing discounted cash flows, so the two derivations
// guard each other.
func closedFormMacaulay(n, c, y float64) float64 {
	return (1+y)/y - (1+y+n*(c-y))/(c*(math.Pow(1+y, n)-1)+y)
}

func TestMacaulayDurationMatchesTheClosedForm(t *testing.T) {
	for _, coupon := range []float64{0.03, 0.04, 0.015} {
		for n := 1.0; n <= 30; n++ {
			want := closedFormMacaulay(n, coupon, durParYield)
			if got := macaulayDuration(n, coupon, durParYield); math.Abs(got-want) > 1e-9 {
				t.Errorf("coupon %.3f, %.0f years left: cash flows give %.9f, closed form %.9f", coupon, n, got, want)
			}
		}
	}
	// A zero coupon has one flow, so its duration IS its remaining maturity:
	// the property the plate draws as a straight line to zero.
	for n := 0.0; n <= durZeroLife; n += 0.5 {
		if got := macaulayDuration(n, 0, durParYield); math.Abs(got-n) > 1e-12 {
			t.Errorf("zero coupon with %.1f years left: duration %.6f", n, got)
		}
	}
}

// The plate's whole argument is three end conditions: the rolling ETF stays
// flat and never reaches zero, the dated fund reaches zero the year it
// liquidates, the zero coupon reaches zero at maturity.
func TestDurationVehiculesEndConditions(t *testing.T) {
	roll := durRollingDuration()
	if math.Abs(roll-7.6133) > 5e-4 {
		t.Errorf("the rolling ETF sits at %.4f, expected 7.6133", roll)
	}
	lo := macaulayDuration(durRollLo, durParCoupon, durParYield)
	hi := macaulayDuration(durRollHi, durParCoupon, durParYield)
	if roll <= lo || roll >= hi {
		t.Errorf("%.4f is outside its own tranche (%.4f to %.4f)", roll, lo, hi)
	}
	// Flat means flat: the tranche is renewed, so the same average comes back
	// every year, whatever the year, and it is never zero.
	for y := 0.0; y <= 40; y++ {
		if got := durRollingDuration(); got != roll {
			t.Errorf("year %.0f: the rolling duration moved to %.4f", y, got)
		}
	}
	if roll <= 0 {
		t.Error("the rolling ETF reaches zero duration, it has no maturity")
	}

	// The dated fund: strictly decreasing, exactly 1 year of duration when one
	// flow is left, exactly 0 the year of liquidation.
	last := math.Inf(1)
	for y := 0.0; y <= durDatedLife; y++ {
		d := macaulayDuration(durDatedLife-y, durParCoupon, durParYield)
		if d >= last {
			t.Errorf("year %.0f: duration %.4f did not decrease from %.4f", y, d, last)
		}
		last = d
	}
	if d := macaulayDuration(1, durParCoupon, durParYield); d != 1 {
		t.Errorf("one year left leaves one flow, duration %.6f", d)
	}
	if d := macaulayDuration(0, durParCoupon, durParYield); d != 0 {
		t.Errorf("at liquidation the dated fund still shows duration %.6f", d)
	}
	// The zero coupon: the straight line the plate draws, year by year, ending
	// on exactly zero at maturity.
	for y := 0.0; y <= durZeroLife; y++ {
		want := durZeroLife - y
		if got := macaulayDuration(want, 0, durParYield); math.Abs(got-want) > 1e-12 {
			t.Errorf("year %.0f of the zero coupon: duration %.6f, expected %.0f", y, got, want)
		}
	}
	// The dated fund starts just under the ETF, which is the plate's quiet
	// point: same duration bought on day one, opposite destinies afterwards.
	if start := macaulayDuration(durDatedLife, durParCoupon, durParYield); math.Abs(start-roll) > 0.5 {
		t.Errorf("the dated fund starts at %.4f against the ETF's %.4f, the plate draws them close", start, roll)
	}
}

// The plate must agree with the numbers the article's own prose quotes.
func TestDurationVehiculesAgreesWithTheArticle(t *testing.T) {
	// "sa volatilité est extrême, ±29 % par point de taux pour un zéro 30 ans":
	// that is the modified duration of a 30-year zero.
	if mod := macaulayDuration(30, 0, durParYield) / (1 + durParYield); math.Abs(mod-29) > 0.5 {
		t.Errorf("a 30-year zero moves %.2f %% per rate point, the article says 29", mod)
	}
	// "les ETF d'État 25 ans et plus, duration 18 à 20": a par bond at the
	// long end of that segment must land in the stated band.
	if d := macaulayDuration(30, durParCoupon, durParYield); d < 18 || d > 21 {
		t.Errorf("a 30-year par bond has duration %.2f, the article says 18 to 20 for the segment", d)
	}

	svg := figDurationVehicules()
	for _, want := range []string{"ETF roulant « 7-10 ans »", "Fonds à échéance 8 ans", "Zéro-coupon 12 ans", "7,6"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not label %q", want)
		}
	}
	for _, banned := range []string{"rgba", "opacity", "—", "linearGradient"} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate contains %q", banned)
		}
	}
	if figures["duration-vehicules"] == nil {
		t.Error("duration-vehicules is not registered in figures.go")
	}
}
