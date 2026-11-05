package decumul

import (
	"math"
	"testing"
)

func TestAnnuityFactorShape(t *testing.T) {
	// A joint-life factor at 60 should be a sizeable number of years of income,
	// fall with a higher discount rate, and fall with older age.
	f0 := AnnuityFactor(FrenchMortality, 60, 0.01)
	if f0 < 15 || f0 > 40 {
		t.Errorf("factor at 60 = %.1f, expected a plausible 15-40 years", f0)
	}
	if AnnuityFactor(FrenchMortality, 60, 0.03) >= f0 {
		t.Error("a higher real rate should lower the annuity factor")
	}
	if AnnuityFactor(FrenchMortality, 75, 0.01) >= f0 {
		t.Error("an older couple should have a lower factor (shorter expected payout)")
	}
}

func TestAnnuityIncome(t *testing.T) {
	// income = load * premium / factor, and 0 for a non-positive premium.
	f := AnnuityFactor(FrenchMortality, 60, 0.01)
	got := AnnuityIncome(FrenchMortality, 60, 200000, 0.01, 0.90)
	want := 0.90 * 200000 / f
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("income = %.2f, want %.2f", got, want)
	}
	if AnnuityIncome(FrenchMortality, 60, 0, 0.01, 0.90) != 0 {
		t.Error("zero premium must give zero income")
	}
}

// A deferred annuity reads the deaths already observed: a single-life one is
// not bought on a member who has died, and a joint one after a first death is
// priced on the survivor alone.
func TestDeferredAnnuityReadsObservedDeaths(t *testing.T) {
	lt := &Lifetime{Self: Life{Age: 60}, Partner: &Life{Age: 60}}
	single := Plan{Capital: 100000, Years: 4, Lifetime: lt, Annuity: &Annuity{Year: 2, Share: 0.5}}
	r := single.RunPath(nil, Lives{Self: 1, Partner: 4})
	if r.Premium != 0 || r.Estate != 100000 {
		t.Errorf("single life on a dead annuitant: premium %v estate %v, want no purchase", r.Premium, r.Estate)
	}

	joint := single
	joint.Annuity = &Annuity{Year: 2, Share: 0.5, Joint: true}
	survivorOnly := joint.Annuity.survivalOf(lt, 2, false, true)
	want := 50000 / AnnuityFactorFor(survivorOnly, 0)
	r = joint.RunPath(nil, Lives{Self: 1, Partner: 4})
	if r.Premium != 50000 || math.Abs(r.Annuity-2*want) > 1e-6 {
		t.Errorf("joint after a death: premium %v income %v, want 50000 and 2 years of %v (the survivor's life)",
			r.Premium, r.Annuity, want)
	}
}
