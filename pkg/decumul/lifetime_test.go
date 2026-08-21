package decumul

import (
	"math"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// neverDies is a MortalityLaw under which nobody ever dies: the degenerate
// case that must reproduce the fixed horizon exactly.
type neverDies struct{}

func (neverDies) Survival(age, years float64) float64 { return 1 }

// diesAt is a MortalityLaw that kills everybody after exactly n years, so a
// kernel test can name the year it expects.
type diesAt struct{ n float64 }

func (d diesAt) Survival(age, years float64) float64 {
	if years < d.n {
		return 1
	}
	return 0
}

// couplePlan is the reference mortal plan the anchors below run: a household
// retiring at 60 on a million, spending 45 k a year, with a horizon long
// enough (to 115) that the longevity tail is not truncated away.
func couplePlan(years int) Plan {
	partner := Life{Age: 60}
	return Plan{
		Capital: 1_000_000, NeedAnnual: 45_000, Years: years,
		Tax:    CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: years},
		Lifetime: &Lifetime{
			Self:    Life{Age: 60},
			Partner: &partner,
		},
	}
}

// ANCHOR 1. With a law that never kills, the stochastic-lifetime kernel must
// reproduce the fixed-horizon results exactly, path by path. Not a tolerance,
// an equality: the mortal and the fixed-horizon runs are meant to be one code
// path, and lifespans are drawn from their own RNG stream so the returns are
// untouched.
func TestImmortalLawReproducesFixedHorizonExactly(t *testing.T) {
	fixed := couplePlan(40)
	fixed.Lifetime = nil

	partner := Life{Age: 60, Law: neverDies{}}
	mortal := couplePlan(40)
	mortal.Lifetime = &Lifetime{Self: Life{Age: 60, Law: neverDies{}}, Partner: &partner}

	a := fixed.Simulate(500, 4, 7)
	b := mortal.Simulate(500, 4, 7)
	if !reflect.DeepEqual(a.Paths, b.Paths) {
		for i := range a.Paths {
			if !reflect.DeepEqual(a.Paths[i], b.Paths[i]) {
				t.Fatalf("path %d differs:\n fixed  %+v\n mortal %+v", i, a.Paths[i], b.Paths[i])
			}
		}
		t.Fatal("ensembles differ")
	}
	if o := b.LifeOutcome(); o.OutlivedPlan != 1 {
		t.Errorf("OutlivedPlan = %.3f under a law that never kills, want 1", o.OutlivedPlan)
	}
}

// ANCHOR 2. On a plan where mortality does not change the trajectory (no
// annuity, no survivor adjustment, household income), the posterior survival
// weighting and the per-path kernel estimate the same alive-ruin probability.
// The paired Monte-Carlo standard error at 30 000 paths is under 0.3 pt, so a
// 1 pt tolerance is several sigma; a real disagreement means the kernel
// changed what "ruin" means.
func TestWeightingAndKernelAgreeOnAliveRuin(t *testing.T) {
	const paths = 30000
	mortal := couplePlan(55)
	fixed := mortal
	fixed.Lifetime = nil

	// The weighting: each ruined path counted by its chance of being reached.
	fe := fixed.Simulate(paths, 8, 7)
	weighted := 0.0
	for _, p := range fe.Paths {
		if p.Ruined && p.RuinYear >= 0 {
			weighted += FrenchMortality.CoupleSurvival(60, float64(p.RuinYear))
		}
	}
	weighted /= float64(len(fe.Paths))

	// The kernel: households that actually ran out while someone was alive.
	exact := mortal.Simulate(paths, 8, 7).LifeOutcome().RuinAlive
	if math.Abs(exact-weighted) > 0.01 {
		t.Errorf("alive ruin: kernel %.4f vs survival weighting %.4f, want them within 1 pt", exact, weighted)
	}
	// The headline fixed-horizon ruin is the ceiling both sit under.
	if headline := fe.RuinProb(); exact > headline+1e-9 {
		t.Errorf("alive ruin %.4f exceeds the fixed-horizon ruin %.4f", exact, headline)
	}
}

// The two lifecycle views must also agree curve by curve, not only on their
// headline: LifeStates counts the drawn deaths where LifeCurve weights by the
// law they were drawn from.
func TestLifeStatesAgreesWithLifeCurve(t *testing.T) {
	const paths = 30000
	mortal := couplePlan(55)
	fixed := mortal
	fixed.Lifetime = nil

	weighted := fixed.Simulate(paths, 8, 7).LifeCurve(func(years float64) float64 {
		return FrenchMortality.CoupleSurvival(60, years)
	})
	exact := mortal.Simulate(paths, 8, 7).LifeStates()
	if len(exact) != len(weighted) {
		t.Fatalf("curve lengths %d vs %d", len(exact), len(weighted))
	}
	for t0 := range exact {
		if math.Abs(exact[t0].Dead-weighted[t0].Dead) > 0.02 {
			t.Errorf("year %d dead: %.3f vs %.3f", t0, exact[t0].Dead, weighted[t0].Dead)
		}
		if math.Abs(exact[t0].Broke-weighted[t0].Broke) > 0.02 {
			t.Errorf("year %d broke: %.3f vs %.3f", t0, exact[t0].Broke, weighted[t0].Broke)
		}
		if s := exact[t0].Dead + exact[t0].Broke + exact[t0].Funded; math.Abs(s-1) > 1e-9 {
			t.Errorf("year %d shares sum to %.6f", t0, s)
		}
	}
}

// ANCHOR 3. An annuity priced at a 0 % real rate with no insurer margin is a
// pure redistribution inside the group: the income it pays over a drawn
// lifetime must return the premium in expectation. It is the identity that
// proves the mortality credit is really being realised, and it breaks the
// moment the draw convention and the pricing convention disagree by a year.
func TestAnnuityReturnsItsPremiumAtZeroRateAndLoad(t *testing.T) {
	const years = 60
	partner := Life{Age: 65}
	p := Plan{
		Capital: 1_000_000, NeedAnnual: 0, Years: years,
		Tax:      CTOFlatTax{Rate: 0},
		Source:   scenario.ParametricSource{Periods: years}, // zero real returns
		Lifetime: &Lifetime{Self: Life{Age: 65}, Partner: &partner},
		Annuity:  &Annuity{Year: 0, Share: 0.5, Rate: 0, Load: 0, Joint: true},
	}
	e := p.Simulate(50000, 8, 11)
	var paid, premium float64
	for _, path := range e.Paths {
		paid += path.Annuity
		premium += path.Premium
	}
	paid /= float64(len(e.Paths))
	premium /= float64(len(e.Paths))
	if math.Abs(premium-500_000) > 1 {
		t.Fatalf("premium = %.0f, want half the capital", premium)
	}
	if rel := math.Abs(paid-premium) / premium; rel > 0.02 {
		t.Errorf("annuity paid %.0f for a premium of %.0f (%.2f%% off), want money back within 2%%", paid, premium, rel*100)
	}
	// A load is exactly what the household gives up: 10 % off the income.
	loaded := p
	a := *p.Annuity
	a.Load = 0.10
	loaded.Annuity = &a
	var withLoad float64
	for _, path := range loaded.Simulate(50000, 8, 11).Paths {
		withLoad += path.Annuity
	}
	withLoad /= float64(len(e.Paths))
	if rel := withLoad / paid; math.Abs(rel-0.90) > 0.01 {
		t.Errorf("a 10%% load paid %.3f of the fair income, want 0.90", rel)
	}
}

// The trade-off the fixed horizon could not show. An annuity insures the risk
// of outliving the plan, a risk that does not exist under a fixed horizon, so
// a sweep run there in 2026-07 found both readings always moving together. With
// the lifetime drawn per path, an actuarially fair joint annuity buys down the
// alive-ruin and pays for it out of the estate, which is the shape the
// literature describes and the reason this kernel exists.
func TestAnnuityBuysDownRuinAndPaysWithTheEstate(t *testing.T) {
	const paths = 20000
	partner := Life{Age: 65}
	base := Plan{
		Capital: 1_000_000, NeedAnnual: 42_000, Years: 50,
		Tax:      CTOFlatTax{Rate: 0},
		Source:   scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 50},
		Lifetime: &Lifetime{Self: Life{Age: 65}, Partner: &partner},
	}
	annuitised := base
	annuitised.Annuity = &Annuity{Year: 0, Share: 0.5, Rate: 0.02, Joint: true}

	plain := base.Simulate(paths, 8, 7).LifeOutcome()
	with := annuitised.Simulate(paths, 8, 7).LifeOutcome()

	if with.RuinAlive >= plain.RuinAlive-0.05 {
		t.Errorf("alive ruin %.4f annuitised vs %.4f plain: a fair annuity must buy down the risk of outliving the plan",
			with.RuinAlive, plain.RuinAlive)
	}
	if with.BrokeYearsMean >= plain.BrokeYearsMean {
		t.Errorf("broke years %.2f annuitised vs %.2f plain, want fewer", with.BrokeYearsMean, plain.BrokeYearsMean)
	}
	if with.EstateP50 >= plain.EstateP50 {
		t.Errorf("median estate %.0f annuitised vs %.0f plain: the insurance is paid out of the bequest",
			with.EstateP50, plain.EstateP50)
	}
	// And the lifestyle is unchanged: the income simply arrives from the
	// insurer instead of the portfolio, which a portfolio-only reading of
	// income would report as a collapse.
	if rel := with.IncomeMean / plain.IncomeMean; rel < 0.95 || rel > 1.05 {
		t.Errorf("mean total income moved by %.1f%% (%.0f vs %.0f), want it roughly unchanged",
			(rel-1)*100, with.IncomeMean, plain.IncomeMean)
	}
}

// Income counts what the portfolio delivered AND what arrived from outside it.
func TestIncomeMeanCountsOutsideIncome(t *testing.T) {
	p := Plan{
		Capital: 100_000, NeedAnnual: 10_000, Years: 4,
		Tax:       CTOFlatTax{Rate: 0},
		Source:    scenario.ParametricSource{Periods: 4},
		Cashflows: []Cashflow{{FromYear: 2, Annual: 4_000}},
		Lifetime:  &Lifetime{Self: Life{Age: 60, Law: neverDies{}}},
	}
	e := p.Simulate(50, 4, 3)
	path := e.Paths[0]
	if math.Abs(path.Received-8_000) > 1e-9 {
		t.Errorf("Received = %.0f, want two years of a 4 k pension", path.Received)
	}
	if o := e.LifeOutcome(); math.Abs(o.IncomeMean-10_000) > 1e-9 {
		t.Errorf("IncomeMean = %.0f, want the full 10 k standard of living", o.IncomeMean)
	}
}

// Lifespans are drawn by inverting the survival table, so the empirical
// P(alive after t years) must reproduce the law that produced it.
func TestDrawnLifespansMatchTheLaw(t *testing.T) {
	const n = 40000
	l := Life{Age: 60}
	s := Lifetime{Self: l}.sampler(60)
	rng := rand.New(rand.NewPCG(3, 4))
	alive := make([]int, 61)
	for range n {
		lv := s.draw(rng)
		if lv.Partner >= 0 {
			t.Fatalf("single life drew a partner: %+v", lv)
		}
		if lv.Self < 1 {
			t.Fatalf("lifespan %d, want at least one year", lv.Self)
		}
		for t0 := 0; t0 <= 60 && t0 < lv.Self; t0++ {
			alive[t0]++
		}
	}
	for _, t0 := range []int{0, 5, 10, 20, 30, 40} {
		got := float64(alive[t0]) / n
		want := l.Survival(float64(t0))
		if math.Abs(got-want) > 0.01 {
			t.Errorf("P(alive at %d) = %.4f, law says %.4f", t0, got, want)
		}
	}
}

// A couple outlives a single person, and the kernel's drawn households say so.
func TestCoupleOutlivesSingle(t *testing.T) {
	single := couplePlan(55)
	single.Lifetime = &Lifetime{Self: Life{Age: 60}}
	one := single.Simulate(5000, 4, 7).LifeOutcome()
	two := couplePlan(55).Simulate(5000, 4, 7).LifeOutcome()
	if two.MedianLifeYears <= one.MedianLifeYears {
		t.Errorf("couple median life %.0f years, single %.0f: the couple must last longer",
			two.MedianLifeYears, one.MedianLifeYears)
	}
}

// The survivor's smaller budget and a partly reverting pension are applied at
// the years the draw names, and the estate is the wealth left at the end.
func TestSurvivorSpendingAndReversion(t *testing.T) {
	partner := Life{Age: 60}
	p := Plan{
		Capital: 10_000, NeedAnnual: 100, Years: 10,
		Tax:       CTOFlatTax{Rate: 0},
		Source:    scenario.ParametricSource{Periods: 10},
		Cashflows: []Cashflow{{Owner: Self, Annual: 40, Reversion: 0.5}},
		Lifetime:  &Lifetime{Self: Life{Age: 60}, Partner: &partner, SurvivorSpend: 0.7},
	}
	// Self lives two years, the partner five: three widowed years, then gone.
	res := p.RunPath(zeros(10), Lives{Self: 2, Partner: 5})
	want := []float64{60, 60, 50, 50, 50, 0, 0, 0, 0, 0}
	for k, w := range want {
		if math.Abs(res.Spend[k]-w) > 1e-9 {
			t.Errorf("Spend[%d] = %.1f, want %.1f (%v)", k, res.Spend[k], w, res.Spend)
			break
		}
	}
	if res.LifeYears != 5 || res.Outlived {
		t.Errorf("LifeYears = %d Outlived = %v, want 5 and false", res.LifeYears, res.Outlived)
	}
	if math.Abs(res.Estate-9730) > 1e-9 {
		t.Errorf("Estate = %.1f, want 9730", res.Estate)
	}
	// The wealth series is frozen at the estate past the household's end, so a
	// death never reads as a crash to zero.
	for k := 6; k <= 10; k++ {
		if math.Abs(res.Wealth[k]-res.Estate) > 1e-9 {
			t.Errorf("Wealth[%d] = %.1f, want the frozen estate %.1f", k, res.Wealth[k], res.Estate)
		}
	}
}

// A Household flow ignores the reversion and pays throughout; an owned one
// drops to its reversion when its owner dies. Only the years the household
// lives (0 to 4 here) are ever asked for: the kernel stops at its end.
func TestCashflowOwnership(t *testing.T) {
	lf := life{horizon: 10, self: 2, partner: 5, survivor: 1, annYear: -1}
	house := Cashflow{Annual: 40, Reversion: 0.5}
	mine := Cashflow{Owner: Self, Annual: 40, Reversion: 0.5}
	theirs := Cashflow{Owner: Partner, Annual: 40}
	for _, c := range []struct {
		name string
		flow Cashflow
		want [5]float64
	}{
		{"household", house, [5]float64{40, 40, 40, 40, 40}},
		{"self with reversion", mine, [5]float64{40, 40, 20, 20, 20}},
		{"partner without", theirs, [5]float64{40, 40, 40, 40, 40}},
	} {
		for k, want := range c.want {
			if got := c.flow.paidAt(k, lf); math.Abs(got-want) > 1e-9 {
				t.Errorf("%s: year %d paid %.0f, want %.0f", c.name, k, got, want)
			}
		}
	}
	alone := life{horizon: 10, self: 2, partner: -1, survivor: 1, annYear: -1}
	if got := mine.paidAt(3, alone); got != 0 {
		t.Errorf("single life: reverted %.0f with nobody to revert to", got)
	}
}

// A deterministic death date makes the horizon exact: the kernel stops there,
// whatever Plan.Years says.
func TestDeterministicDeathStopsTheKernel(t *testing.T) {
	p := Plan{
		Capital: 10_000, NeedAnnual: 1_000, Years: 20,
		Tax:      CTOFlatTax{Rate: 0},
		Source:   scenario.ParametricSource{Periods: 20},
		Lifetime: &Lifetime{Self: Life{Age: 60, Law: diesAt{7}}},
	}
	e := p.Simulate(200, 4, 5)
	for _, path := range e.Paths {
		if path.LifeYears != 7 || path.Outlived {
			t.Fatalf("LifeYears = %d Outlived = %v, want exactly 7 and false", path.LifeYears, path.Outlived)
		}
		if math.Abs(path.Estate-3000) > 1e-9 {
			t.Fatalf("Estate = %.0f, want 3000 (10 k less seven years of a thousand)", path.Estate)
		}
	}
	o := e.LifeOutcome()
	if o.MedianLifeYears != 7 || o.OutlivedPlan != 0 || o.RuinAlive != 0 {
		t.Errorf("LifeOutcome = %+v, want 7 median years, nobody censored, nobody broke", o)
	}
	if math.Abs(o.EstateP50-3000) > 1e-9 {
		t.Errorf("EstateP50 = %.0f, want 3000", o.EstateP50)
	}
	// Twenty planned years, seven lived: the states flip at year 7 and the
	// spending statistics see seven years, not twenty.
	states := e.LifeStates()
	if states[6].Dead != 0 || states[7].Dead != 1 {
		t.Errorf("dead at year-end 6/7 = %.2f/%.2f, want 0 then 1", states[6].Dead, states[7].Dead)
	}
	if mean, cv := e.spendMoments(); math.Abs(mean-1000) > 1e-9 || cv > 1e-12 {
		t.Errorf("income mean %.1f cv %.4f, want a steady 1000", mean, cv)
	}
}

// Broke-while-alive is counted, and how long it lasted with it: a plan that
// runs out at a known year, under a known death date, has a known answer.
func TestBrokeYearsAreCounted(t *testing.T) {
	p := Plan{
		Capital: 3_000, NeedAnnual: 1_000, Years: 20,
		Tax:      CTOFlatTax{Rate: 0},
		Source:   scenario.ParametricSource{Periods: 20},
		Lifetime: &Lifetime{Self: Life{Age: 60, Law: diesAt{9}}},
	}
	o := p.Simulate(200, 4, 5).LifeOutcome()
	if o.RuinAlive != 1 {
		t.Errorf("RuinAlive = %.2f, want 1: three years of capital, nine years to live", o.RuinAlive)
	}
	// Three thousand euros fund years 0 to 2 and leave nothing, so ruin
	// latches at year 2 (the kernel latches on an emptied portfolio as well as
	// on an unfunded need) and seven of the nine lived years are lived broke.
	if math.Abs(o.BrokeYearsMean-7) > 1e-9 || math.Abs(o.BrokeYearsP95-7) > 1e-9 {
		t.Errorf("broke years mean %.1f p95 %.1f, want 7", o.BrokeYearsMean, o.BrokeYearsP95)
	}
	if o.EstateZero != 1 {
		t.Errorf("EstateZero = %.2f, want 1", o.EstateZero)
	}
}

// PlanHorizon lets the amortization rule plan over the horizon the household
// actually plans for while the simulation runs past any plausible age, so the
// longevity tail is not truncated. The two are equal by default.
func TestPlanHorizonDrivesAmortization(t *testing.T) {
	base := Plan{
		Capital: 100_000, NeedAnnual: 0, Years: 50,
		Tax: CTOFlatTax{Rate: 0}, Amortize: true, AmortReturn: 0,
		Source: scenario.ParametricSource{Periods: 50},
	}
	if got := base.RunPath(zeros(50), Lives{}).Spend[0]; math.Abs(got-2000) > 1e-6 {
		t.Errorf("first payment %.1f, want 100k spread over 50 years", got)
	}
	short := base
	short.PlanHorizon = 25
	if got := short.RunPath(zeros(50), Lives{}).Spend[0]; math.Abs(got-4000) > 1e-6 {
		t.Errorf("first payment %.1f with a 25-year planning horizon, want 4000", got)
	}
}

// A Draws assembled by hand, without lifespans, still runs the mortality it
// was given rather than a silent fixed horizon.
func TestSimulateOnFillsMissingLives(t *testing.T) {
	p := couplePlan(55)
	bare := Draws{Returns: p.Draw(500, 4, 7).Returns}
	o := p.SimulateOn(bare, 4).LifeOutcome()
	if o.OutlivedPlan > 0.10 {
		t.Errorf("OutlivedPlan = %.3f: mortality was not applied", o.OutlivedPlan)
	}
	// And it stays reproducible.
	if again := p.SimulateOn(bare, 4).LifeOutcome(); again != o {
		t.Errorf("two runs on the same Draws disagree: %+v vs %+v", o, again)
	}
}

// The annuity is priced on the ages reached at the purchase date, so buying
// later buys more income per euro, and a longer-lived pricing law buys less.
func TestAnnuityPricingRespondsToAgeAndLaw(t *testing.T) {
	partner := Life{Age: 60}
	lt := &Lifetime{Self: Life{Age: 60}, Partner: &partner}
	at := func(a *Annuity, year int) float64 {
		return 1 / AnnuityFactorFor(a.survivalFrom(lt, year), a.Rate)
	}
	now := &Annuity{Joint: true, Rate: 0.01}
	if at(now, 15) <= at(now, 0) {
		t.Error("buying at 75 must buy more income per euro than at 60")
	}
	annuitants := &Annuity{Joint: true, Rate: 0.01, Law: Gompertz{Mode: 91, Dispersion: 10}}
	if at(annuitants, 0) >= at(now, 0) {
		t.Error("pricing on a longer-lived table must buy less income per euro")
	}
	single := &Annuity{Rate: 0.01}
	if at(single, 0) <= at(now, 0) {
		t.Error("a single-life annuity must pay more than a joint one at the same age")
	}
}
