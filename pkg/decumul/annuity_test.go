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
