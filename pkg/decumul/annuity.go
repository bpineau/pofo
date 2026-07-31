package decumul

import "math"

// AnnuityFactor is the present value of 1 unit of real income per year, paid
// while at least one member of a same-age couple is alive, discounted at a real
// rate. It is the price of a joint-life, inflation-linked immediate annuity per
// unit of income, before any insurer loading: premium = income * factor, so
// income = premium / factor.
//
// The sum runs to a far horizon (survival vanishes well before it). A higher
// real rate or older age lowers the factor (cheaper income); a younger age
// raises it.
func AnnuityFactor(m Gompertz, age, realRate float64) float64 {
	factor := 0.0
	disc := 1.0
	step := 1 / (1 + realRate)
	for t := 0; t < 60; t++ {
		factor += m.CoupleSurvival(age, float64(t)) * disc
		disc *= step
	}
	return factor
}

// AnnuityIncome is the real annual income a premium buys as a joint-life,
// inflation-linked immediate annuity, after an insurer load kept in [0,1]
// (0.90 keeps 90% of the actuarially fair income, the rest the insurer's
// margin). It hedges longevity: the income lasts as long as the couple does,
// so the risk of outliving the money falls, at the cost of the premium's
// liquidity and bequest. Returns 0 for a non-positive premium or factor.
func AnnuityIncome(m Gompertz, age, premium, realRate, load float64) float64 {
	f := AnnuityFactor(m, age, realRate)
	if premium <= 0 || f <= 0 {
		return 0
	}
	return math.Max(0, load) * premium / f
}
