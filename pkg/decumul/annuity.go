package decumul

import "math"

// annuityHorizon is how far the annuity sums run. Survival vanishes well
// before it at any adult age, so the truncation is not material.
const annuityHorizon = 60

// AnnuityFactorFor is the present value of 1 unit of real income per year,
// paid at the START of every year while surv lasts (an annuity-due, matching
// the kernel's begin-of-year withdrawal), discounted at a real rate. It is the
// law-agnostic core of AnnuityFactor: premium = income * factor, so
// income = premium / factor.
//
// surv(t) is the probability that the covered life, or lives, are still going
// t years after the purchase.
func AnnuityFactorFor(surv func(years float64) float64, realRate float64) float64 {
	factor := 0.0
	disc := 1.0
	step := 1 / (1 + realRate)
	for t := 0; t < annuityHorizon; t++ {
		factor += surv(float64(t)) * disc
		disc *= step
	}
	return factor
}

// AnnuityFactor is AnnuityFactorFor for a joint-life annuity over a same-age
// couple under one Gompertz law: the price of an inflation-linked immediate
// annuity per unit of income, before any insurer loading.
//
// A higher real rate or older age lowers the factor (cheaper income); a
// younger age raises it.
func AnnuityFactor(m Gompertz, age, realRate float64) float64 {
	return AnnuityFactorFor(func(years float64) float64 {
		return m.CoupleSurvival(age, years)
	}, realRate)
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

// Annuity converts a share of the portfolio into a lifelong real income, once,
// at a chosen date. It only means something alongside a Lifetime: an annuity
// has no price without an age and nothing to insure without a drawn lifespan,
// so Plan.Annuity is ignored when Plan.Lifetime is nil.
//
// The premium is raised by SELLING from the tax pockets, so an annuity bought
// out of a taxable account really pays its capital-gains tax, through the same
// gross-up a withdrawal uses. Only the growth sleeve is converted; the buffer
// is left alone.
//
// The income then runs until the covered life ends, so early deaths fund late
// ones: this is where the mortality credit is actually realised, and why the
// annuity trade-off (headline ruin can rise while the worst late-life outcomes
// improve) is measurable here and not under a fixed horizon.
type Annuity struct {
	Year int // plan year the annuity is bought (0 = at retirement)
	// Share of the growth sleeve's value at that date becomes the NET premium.
	// The sale that raises it is grossed up for tax like any withdrawal, so a
	// taxable pocket gives up more than Share to buy it. In [0,1].
	Share float64
	Rate  float64 // real rate the insurer prices at
	// Load is the insurer's margin as a share of the fair income: 0.10 keeps
	// 90 % of the actuarially fair income, and 0 is actuarially fair. Note the
	// convention differs from AnnuityIncome's load, which is the RETAINED
	// fraction; "a 10 % load" is the normal phrase, so the struct states the
	// margin and converts internally.
	Load float64
	// Joint pays while either member of the household is alive; otherwise the
	// income stops at Self's death. Ignored (treated as single) when the
	// Lifetime has no Partner.
	Joint bool
	// Law is the mortality the INSURER prices with; nil uses the household's
	// own. Annuitants are not the general population, and French insurers
	// price on generational annuitant tables (TGH/TGF-05) that run materially
	// longer than an INSEE period table; folding that into Load would conflate
	// two effects that move differently with age.
	Law MortalityLaw
}

// survivalFrom is the survival function the annuity is priced on, seen from
// the purchase year: the covered life or lives, at the ages they have reached.
func (a *Annuity) survivalFrom(lt *Lifetime, year int) func(float64) float64 {
	lawOf := func(l Life) MortalityLaw {
		if a.Law != nil {
			return a.Law
		}
		return l.law()
	}
	selfAge, selfLaw := lt.Self.Age+float64(year), lawOf(lt.Self)
	if !a.Joint || lt.Partner == nil {
		return func(years float64) float64 { return selfLaw.Survival(selfAge, years) }
	}
	partnerAge, partnerLaw := lt.Partner.Age+float64(year), lawOf(*lt.Partner)
	return func(years float64) float64 {
		// At least one of two independent lives is still going.
		return 1 - (1-selfLaw.Survival(selfAge, years))*(1-partnerLaw.Survival(partnerAge, years))
	}
}

// buyAnnuity converts Share of the growth sleeve at the start of year k,
// recording the premium raised and the lifelong income it bought on the path
// and on its mortality state. It is a no-op without a Lifetime, outside the
// purchase year, or once already bought.
func (p Plan) buyAnnuity(k int, pks pocketOps, res *PathResult, lf *life) {
	a := p.Annuity
	if a == nil || p.Lifetime == nil || k != a.Year || lf.annYear >= 0 {
		return
	}
	want := math.Min(math.Max(a.Share, 0), 1) * pks.total()
	if want <= 0 {
		return
	}
	premium := pks.sell(want, &res.TaxPaid)
	factor := AnnuityFactorFor(a.survivalFrom(p.Lifetime, k), a.Rate)
	if premium <= 0 || factor <= 0 {
		return
	}
	res.Premium = premium
	lf.annuity = math.Max(0, 1-a.Load) * premium / factor
	lf.annYear = k
	lf.annJoint = a.Joint && p.Lifetime.Partner != nil
}
