package decumul

import (
	"math/rand/v2"
	"sort"
)

// MortalityLaw is a survival law for a single life: Survival is the
// probability that a person of the given age is still alive after years more
// years. It must return 1 at years 0 and be non-increasing in years, since
// lifespans are drawn by inverting it directly (see Lifetime).
//
// Gompertz implements it and FrenchMortality is the bundled default, so a
// caller only meets this interface when injecting another law: a published
// life table, a cohort-projected law, or a deliberately harsher one.
type MortalityLaw interface {
	Survival(age, years float64) float64
}

// Life is one person's mortality input: the age they reach at the plan's year
// 0, and the law their remaining lifetime is drawn from. A nil Law means
// FrenchMortality.
type Life struct {
	Age float64
	Law MortalityLaw
}

// law resolves Law, applying the FrenchMortality default.
func (l Life) law() MortalityLaw {
	if l.Law == nil {
		return FrenchMortality
	}
	return l.Law
}

// Survival is the probability this life is still going after years more years.
func (l Life) Survival(years float64) float64 { return l.law().Survival(l.Age, years) }

// Owner names the life a Cashflow depends on once the plan draws lifetimes.
type Owner int

const (
	// Household is the zero value: the flow is paid as long as anyone in the
	// household is alive, which is what every plan without a Lifetime means.
	Household Owner = iota
	// Self ties the flow to the plan's first member (Lifetime.Self).
	Self
	// Partner ties the flow to the second member (Lifetime.Partner).
	Partner
)

// other is the household's other member.
func (o Owner) other() Owner {
	if o == Self {
		return Partner
	}
	return Self
}

// Lifetime turns the plan's fixed horizon into a horizon drawn per path: each
// simulated household is given a lifespan sampled from its members' mortality
// laws, and the kernel stops when the household does. Ruin then means broke
// WHILE ALIVE (exactly, not weighted after the fact), the wealth left at the
// household's end is a first-class output (PathResult.Estate), and an Annuity
// realises real mortality credits because the income stops with the annuitant.
//
// The drawn death is unknown to the household: the spending rules keep
// planning over PlanHorizon (see Plan), never over the lifespan the path was
// dealt. A rule that amortised over the drawn death would describe a
// clairvoyant retiree.
//
// Plan.Years remains the simulation length. A household drawn to outlive it is
// censored: recorded alive at the horizon (PathResult.Outlived) with its
// estate read there. Set Years past any plausible age, and PlanHorizon to the
// horizon the household actually plans over, so the longevity tail an annuity
// insures is not truncated away.
type Lifetime struct {
	Self Life
	// Partner is the household's second life; nil for a single person. The two
	// lifespans are drawn INDEPENDENTLY from each member's own law. Real
	// couples are positively dependent (shared environment, the broken-heart
	// effect), so independence overstates the chance that someone is still
	// alive, which overstates the plan's longevity load rather than flattering
	// it. See docs/stochastic-lifetime-kernel-design.md.
	Partner *Life
	// SurvivorSpend scales the household's needs-based spending once one
	// member is gone (the literature's figure is around 0.7: one person does
	// not live on half of what two do). 0 means 1, no adjustment; a survivor
	// needing literally nothing is not a value anybody means, so the sentinel
	// is unambiguous. Wealth-based rules (VPW, bounded percent, ABW) are not
	// scaled: their budget is a share of what exists, not a restated need.
	SurvivorSpend float64
}

// survivorSpend resolves SurvivorSpend, applying the "no adjustment" default.
func (lt Lifetime) survivorSpend() float64 {
	if lt.SurvivorSpend <= 0 {
		return 1
	}
	return lt.SurvivorSpend
}

// Lives is one household's drawn lifespans, in whole years from the plan's
// year 0: a member with n years lives years 0..n-1 and is gone at year-end n.
// A value above the plan horizon means the member outlived the plan.
//
// The zero value means NO DRAW: the household simply reaches the plan
// horizon. Passing it to RunPath is how a caller asks for the fixed-horizon
// kernel explicitly, and it is what every Lifetime-free plan uses.
type Lives struct {
	Self    int
	Partner int // 0 or negative for a single-life household
}

// lifeSampler holds a household's integer survival tables so drawing a
// lifespan is a binary search rather than one survival evaluation per year.
type lifeSampler struct {
	self, partner []float64
}

// sampler precomputes the survival tables over horizon years.
func (lt Lifetime) sampler(horizon int) lifeSampler {
	s := lifeSampler{self: survivalTable(lt.Self, horizon)}
	if lt.Partner != nil {
		s.partner = survivalTable(*lt.Partner, horizon)
	}
	return s
}

// survivalTable tabulates a life's survival at integer years 0..horizon.
func survivalTable(l Life, horizon int) []float64 {
	tab := make([]float64, horizon+1)
	for t := range tab {
		tab[t] = l.Survival(float64(t))
	}
	return tab
}

// draw samples one household's lifespans.
func (s lifeSampler) draw(rng *rand.Rand) Lives {
	lv := Lives{Self: drawYears(s.self, rng.Float64()), Partner: -1}
	if s.partner != nil {
		lv.Partner = drawYears(s.partner, rng.Float64())
	}
	return lv
}

// drawYears inverts a survival table: it returns min{t : tab[t] <= u}, which
// satisfies P(years > t) = P(u < tab[t]) = tab[t] at every integer year, for
// ANY non-increasing survival function. tab[0] is 1 and u is below 1, so the
// result is at least 1: a death partway through a year still consumes that
// year's spending, matching the kernel's begin-of-year withdrawal. Past the
// table the life is censored, and len(tab) is returned.
func drawYears(tab []float64, u float64) int {
	return sort.Search(len(tab), func(i int) bool { return tab[i] <= u })
}

// life is the per-path mortality and annuity state both kernels read. It is
// built by Plan.life, whose fixed-horizon form makes every mortality effect a
// no-op, so the mortal and the fixed-horizon runs are one code path.
type life struct {
	horizon  int     // the plan's Years
	self     int     // whole years the first member lives (horizon+1 = never dies)
	partner  int     // whole years the partner lives; negative = single life
	survivor float64 // needs-based spending factor once one member is gone
	annuity  float64 // annual real annuity income, once bought
	annYear  int     // year the annuity starts paying; -1 = none
	annJoint bool    // the annuity pays while either member is alive
}

// life builds the per-path state. A nil Lifetime, or the zero Lives (no
// draw), yields a household that reaches the plan horizon intact.
func (p Plan) life(lv Lives) life {
	lf := life{horizon: p.Years, self: p.Years + 1, partner: -1, survivor: 1, annYear: -1}
	if p.Lifetime == nil || lv.Self <= 0 {
		return lf
	}
	lf.self = lv.Self
	if lv.Partner > 0 {
		lf.partner = lv.Partner
	}
	lf.survivor = p.Lifetime.survivorSpend()
	return lf
}

// end is the number of years the household exists, capped at the horizon: the
// kernel runs years 0..end-1.
func (l life) end() int {
	e := l.self
	if l.partner > e {
		e = l.partner
	}
	if e > l.horizon {
		e = l.horizon
	}
	return e
}

// outlived reports whether the household was still alive at the horizon, so
// its end is a censoring rather than a death.
func (l life) outlived() bool { return l.self > l.horizon || l.partner > l.horizon }

// firstDeath is the first year in which the household has lost a member. It
// equals end when nothing changes before the household is gone (a single
// life, or a partner outliving the horizon).
func (l life) firstDeath() int {
	if l.partner < 0 {
		return l.end()
	}
	f := min(l.self, l.partner)
	if f > l.horizon {
		f = l.horizon
	}
	return f
}

// spendFactor scales the year's needs-based spending: 1 while the household
// is whole, SurvivorSpend once one member is gone.
func (l life) spendFactor(year int) float64 {
	if year >= l.firstDeath() {
		return l.survivor
	}
	return 1
}

// alive reports whether the given member (or the household) is still alive in
// the given year.
func (l life) alive(o Owner, year int) bool {
	switch o {
	case Self:
		return year < l.self
	case Partner:
		return l.partner >= 0 && year < l.partner
	default:
		return year < l.end()
	}
}

// annuityAt is the annuity income actually received in a year: nothing before
// the purchase, and nothing after the covered life ends.
func (l life) annuityAt(year int) float64 {
	if l.annuity == 0 || l.annYear < 0 || year < l.annYear {
		return 0
	}
	if l.annJoint || l.alive(Self, year) {
		return l.annuity
	}
	return 0
}

// annuityPlanned is the annuity income as the HOUSEHOLD forecasts it: bought
// or not, never truncated by a death it cannot know about. It is what the
// planning rules discount, where annuityAt is what the path receives.
func (l life) annuityPlanned(year int) float64 {
	if l.annYear < 0 || year < l.annYear {
		return 0
	}
	return l.annuity
}
