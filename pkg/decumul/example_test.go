package decumul_test

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The fixed horizon: a household of 60 with a million, spending 32 k a year in
// real terms over 35 years, taxed at the French flat rate on realised gains.
// Ruin is the share of futures that run out before the horizon, whether or not
// anybody was still there to see it.
func ExamplePlan_Simulate() {
	p := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 32_000, Years: 35,
		Tax:    decumul.CTOFlatTax{Rate: 0.314},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	o := p.Simulate(20_000, 4, 7).Outcome()
	fmt.Printf("ruin %.0f%%, median terminal wealth %.1f M\n", o.RuinProb*100, o.TerminalP50/1e6)
	// Output:
	// ruin 24%, median terminal wealth 0.5 M
}

// FIRE in a dozen lines: a scenario.Source of real returns, a plan, the
// Monte-Carlo, and the question turned around, the spending a 5 % chance of
// ruin allows. Solve bisects on one shared set of draws, so its answer is
// monotone in the target rather than noisy.
func Example_fire() {
	p := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 32_000, Years: 35,
		Tax:    decumul.CTOFlatTax{Rate: 0.314},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	o := p.Simulate(20_000, 4, 7).Outcome() // paths, workers, seed
	fmt.Printf("ruin %.0f%%, median terminal wealth %.1f M\n", o.RuinProb*100, o.TerminalP50/1e6)

	spend := p.Solve(0.05, decumul.WithdrawalAxis(10_000, 100_000), 20_000, 4, 7)
	fmt.Printf("5%% ruin at %.0f a year\n", math.Round(spend/500)*500)
	// Output:
	// ruin 24%, median terminal wealth 0.5 M
	// 5% ruin at 22000 a year
}

// The same household with the lifetime drawn inside every path. Ruin now means
// running out WHILE ALIVE, and the wealth left at death becomes an output in
// its own right. Years runs to 110 so the longevity tail is not truncated,
// while PlanHorizon keeps the household planning over the 35 years it actually
// plans for.
func ExamplePlan_Simulate_lifetime() {
	partner := decumul.Life{Age: 60}
	p := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 32_000, Years: 50, PlanHorizon: 35,
		Tax:    decumul.CTOFlatTax{Rate: 0.314},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 50},
		Lifetime: &decumul.Lifetime{
			Self:          decumul.Life{Age: 60}, // FrenchMortality by default
			Partner:       &partner,
			SurvivorSpend: 0.7,
		},
	}
	o := p.Simulate(20_000, 4, 7).LifeOutcome()
	fmt.Printf("broke while alive %.0f%%, for %.1f years on average\n", o.RuinAlive*100, o.BrokeYearsMean)
	fmt.Printf("median life %.0f years, still alive at the horizon %.0f%%\n", o.MedianLifeYears, o.OutlivedPlan*100)
	fmt.Printf("estate p50 %.1f M, nothing left %.0f%%\n", o.EstateP50/1e6, o.EstateZero*100)
	// Output:
	// broke while alive 13%, for 1.2 years on average
	// median life 33 years, still alive at the horizon 0%
	// estate p50 0.7 M, nothing left 13%
}

// The bundled laws against the INSEE table they were fitted on: remaining
// years at 65 for both sexes, women and men (INSEE 2023-2025: 21.7, 23.5,
// 19.9). A member of the household takes a sex-specific law through Life.Law.
func ExampleGompertz_LifeExpectancy() {
	for _, c := range []struct {
		name string
		law  decumul.Gompertz
	}{{"both", decumul.FrenchMortality}, {"women", decumul.FrenchWomen}, {"men", decumul.FrenchMen}} {
		fmt.Printf("%-5s %.1f years left at 65\n", c.name, c.law.LifeExpectancy(65))
	}
	_ = decumul.Lifetime{Self: decumul.Life{Age: 62, Law: decumul.FrenchMen},
		Partner: &decumul.Life{Age: 60, Law: decumul.FrenchWomen}}
	// Output:
	// both  21.5 years left at 65
	// women 23.3 years left at 65
	// men   19.7 years left at 65
}

// A pension tied to one life, of which 54 % keeps being paid to the survivor:
// the reversion rate is not a detail, it is a large share of a household's
// floor, and only a drawn lifetime can price losing it.
func ExampleCashflow_reversion() {
	partner := decumul.Life{Age: 62}
	base := decumul.Plan{
		Capital: 800_000, NeedAnnual: 40_000, Years: 50, PlanHorizon: 30,
		Tax:    decumul.CTOFlatTax{Rate: 0.314},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 50},
		Lifetime: &decumul.Lifetime{
			Self:          decumul.Life{Age: 62},
			Partner:       &partner,
			SurvivorSpend: 0.7,
		},
	}
	for _, reversion := range []float64{1.0, 0.54, 0} {
		p := base
		p.Cashflows = []decumul.Cashflow{
			{FromYear: 5, Annual: 21_600, Owner: decumul.Self, Reversion: reversion},
		}
		fmt.Printf("reversion %.0f%%: broke while alive %.0f%%\n",
			reversion*100, p.Simulate(20_000, 4, 7).LifeOutcome().RuinAlive*100)
	}
	// Output:
	// reversion 100%: broke while alive 3%
	// reversion 54%: broke while alive 5%
	// reversion 0%: broke while alive 9%
}

// A pension that is not fully indexed. The engine runs in real money, so a
// flow with no Drift keeps its purchasing power for ever; a pension revalued
// at inflation minus half a point is Drift -0.005, and after forty years it
// buys a fifth less. Here it pays from the first year, and the erosion costs
// six points of ruin and about 2 000 a year of spending at 5 % ruin.
func ExampleCashflow_drift() {
	for _, drift := range []float64{0, -0.005, -0.01} {
		p := decumul.Plan{
			Capital: 1_000_000, NeedAnnual: 52_000, Years: 40,
			Cashflows: []decumul.Cashflow{{FromYear: 0, Annual: 24_000, Drift: drift}},
			Tax:       decumul.CTOFlatTax{Rate: 0.314},
			Source:    scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 40},
		}
		ruin := p.Simulate(20_000, 4, 7).Outcome().RuinProb
		spend := p.Solve(0.05, decumul.WithdrawalAxis(10_000, 100_000), 20_000, 4, 7)
		fmt.Printf("drift %+.1f%%: pension %.0f in year 39, ruin %.0f%%, 5%% ruin at %.0f a year\n",
			drift*100, p.Cashflows[0].AmountAt(39), ruin*100, math.Round(spend/500)*500)
	}
	// Output:
	// drift +0.0%: pension 24000 in year 39, ruin 21%, 5% ruin at 43500 a year
	// drift -0.5%: pension 19738 in year 39, ruin 27%, 5% ruin at 41500 a year
	// drift -1.0%: pension 16217 in year 39, ruin 34%, 5% ruin at 39500 a year
}

// Annuitising a third of the portfolio. The income lasts exactly as long as the
// household does, so the mortality credit is realised inside the path: it buys
// down the risk of outliving the plan and the bequest pays for it. Whether it
// is worth buying is a question of price, not of principle: the same trade at
// 75, on an annuitant table with a 10 % load, pays less than the plan withdraws
// and makes both readings worse.
func ExampleAnnuity() {
	partner := decumul.Life{Age: 65}
	base := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 42_000, Years: 50, PlanHorizon: 30,
		Tax:      decumul.CTOFlatTax{Rate: 0.314},
		Source:   scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 50},
		Lifetime: &decumul.Lifetime{Self: decumul.Life{Age: 65}, Partner: &partner},
	}
	cases := []struct {
		name    string
		annuity *decumul.Annuity
	}{
		{"no annuity", nil},
		{"a third at 65", &decumul.Annuity{Year: 0, Share: 1.0 / 3, Rate: 0.02, Load: 0.05, Joint: true}},
		{"a third at 75", &decumul.Annuity{Year: 10, Share: 1.0 / 3, Rate: 0.02, Load: 0.10, Joint: true,
			Law: decumul.Gompertz{Mode: 91, Dispersion: 10}}}, // annuitants outlive the table
	}
	for _, c := range cases {
		p := base
		p.Annuity = c.annuity
		o := p.Simulate(20_000, 8, 7).LifeOutcome()
		fmt.Printf("%-13s broke while alive %.0f%%, median estate %.2f M\n", c.name, o.RuinAlive*100, o.EstateP50/1e6)
	}
	// Output:
	// no annuity    broke while alive 33%, median estate 0.26 M
	// a third at 65 broke while alive 29%, median estate 0.23 M
	// a third at 75 broke while alive 35%, median estate 0.19 M
}

// Solving needs no special case for mortality: the axes, the bisection and the
// shared draws work the same, and Ensemble.RuinProb is already the alive-ruin
// once the plan draws lifetimes. Note how much more demanding a 5 % target is
// when the horizon can run to a hundred.
func ExamplePlan_Solve() {
	partner := decumul.Life{Age: 60}
	p := decumul.Plan{
		Capital: 1_000_000, Years: 50, PlanHorizon: 35,
		Tax:      decumul.CTOFlatTax{Rate: 0.314},
		Source:   scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 50},
		Lifetime: &decumul.Lifetime{Self: decumul.Life{Age: 60}, Partner: &partner},
	}
	spend := p.Solve(0.05, decumul.WithdrawalAxis(10_000, 100_000), 20_000, 8, 7)
	fmt.Printf("5%% chance of ever being broke and alive at %.0f a year\n", math.Round(spend/500)*500)
	// Output:
	// 5% chance of ever being broke and alive at 23000 a year
}

// Several questions about one plan share one set of futures: Draw them once,
// then ask each question On them. RuinProbOn reads nothing but the ruin (no
// ensemble is built), and SolveOn bisects on those same paths, so the two
// answers are consistent with each other and the paths are drawn only once.
func ExamplePlan_RuinProbOn() {
	p := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 40_000, Years: 35,
		Tax:    decumul.CTOFlatTax{Rate: 0.314},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	draws := p.Draw(20_000, 4, 7) // paths, workers, seed
	fmt.Printf("ruin at 40000 a year: %.0f %%\n", p.RuinProbOn(draws, 4)*100)
	spend := p.SolveOn(0.10, decumul.WithdrawalAxis(10_000, 100_000), draws, 4)
	fmt.Printf("10 %% ruin at %.0f a year\n", math.Round(spend/500)*500)
	// Output:
	// ruin at 40000 a year: 52 %
	// 10 % ruin at 26000 a year
}

// Sweep1D maps ruin across one parameter on one set of drawn paths, so the
// curve is smooth: here the yearly spending of a 1 M plan over 35 years.
func ExamplePlan_Sweep1D() {
	p := decumul.Plan{
		Capital: 1_000_000, Years: 35,
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	points, err := p.Sweep1D(decumul.NeedAnnual, []float64{30_000, 40_000, 50_000}, 20_000, 4, 7)
	if err != nil {
		panic(err)
	}
	for _, pt := range points {
		fmt.Printf("%.0f a year: ruin %.0f %%, median terminal wealth %.1f M\n", pt.Value, pt.RuinProb*100, pt.TerminalP50/1e6)
	}
	// Output:
	// 30000 a year: ruin 16 %, median terminal wealth 0.8 M
	// 40000 a year: ruin 41 %, median terminal wealth 0.2 M
	// 50000 a year: ruin 67 %, median terminal wealth 0.0 M
}

// The tax law reads nominal gains while the plan runs in real euros: with
// Inflation set, a flat real path still pays tax, because every cost basis
// loses purchasing power against the holding it prices. One deterministic
// path, no market move at all, 31.4 % on gains.
func ExamplePlan_RunPath_nominalGains() {
	flat := scenario.Sequence{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	p := decumul.Plan{Capital: 500_000, NeedAnnual: 20_000, Years: 10, Tax: decumul.CTOFlatTax{Rate: 0.314}}
	real := p.RunPath(flat, decumul.Lives{})
	p.Inflation = 0.02
	nominal := p.RunPath(flat, decumul.Lives{})
	fmt.Printf("tax on real gains: %.0f\n", real.TaxPaid)
	fmt.Printf("tax on nominal gains at 2 %% inflation: %.0f\n", nominal.TaxPaid)
	// Output:
	// tax on real gains: 0
	// tax on nominal gains at 2 % inflation: 5461
}

// The lifecycle band of a couple retiring at 62: at each age, the share of
// households gone, alive and broke, and alive and funded, the funded split by
// wealth against the capital they started with (at or under it, up to twice
// it, more). It is the chart that sets the risk a plan is built against beside
// the one nobody plans for.
func ExampleEnsemble_LifeBands() {
	partner := decumul.Life{Age: 62}
	p := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 40_000, Years: 45, PlanHorizon: 35,
		Tax:      decumul.CTOFlatTax{Rate: 0.314},
		Source:   scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 45},
		Lifetime: &decumul.Lifetime{Self: decumul.Life{Age: 62}, Partner: &partner},
	}
	bands := p.Simulate(20_000, 4, 7).LifeBands(1, 2)
	for _, age := range []int{72, 82, 92} {
		b := bands[age-62]
		fmt.Printf("age %d: gone %2.0f%%, broke %2.0f%%, funded %2.0f%% / %2.0f%% / %2.0f%%\n",
			age, b.Dead*100, b.Broke*100, b.Funded[0]*100, b.Funded[1]*100, b.Funded[2]*100)
	}
	// Output:
	// age 72: gone  1%, broke  0%, funded 65% / 33% /  2%
	// age 82: gone 10%, broke  8%, funded 60% / 18% /  5%
	// age 92: gone 50%, broke 20%, funded 21% /  6% /  4%
}

// Opening one simulated future. A fan's samples name their path (Index), and
// DrawPath replays that path alone from the seed the ensemble used, so the
// ledger below is the very line the chart drew, read year by year in real
// thousands: the market, the income, what the portfolio paid and from where,
// the sales with the tax they cost, and the wealth at both ends.
func ExamplePlan_RunLedger() {
	p := decumul.Plan{
		Capital: 1_000_000, NeedAnnual: 45_000, Years: 30,
		Buffer:    decumul.BufferSleeve{Years: 2},
		Cashflows: []decumul.Cashflow{{FromYear: 3, Annual: 15_000}},
		Envelopes: []decumul.Envelope{{Name: "CTO", Amount: 910_000, GainFrac: 0.5, Tax: decumul.CTOFlatTax{Rate: 0.314}}},
		Source:    scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 30},
	}
	worst := p.Simulate(2_000, 4, 7).Fan([]float64{0.5}, 8).Samples[0]
	returns, lives := p.DrawPath(7, worst.Index)
	path, ledger := p.RunLedger(returns, lives)
	fmt.Printf("path %d ruined %v in year %d\n", worst.Index, path.Ruined, path.RuinYear)
	for _, y := range ledger[:4] {
		fmt.Printf("year %d: market %+5.1f%%, income %4.1f, paid %4.1f (buffer %4.1f), sold %4.1f, tax %3.1f, wealth %5.1f -> %5.1f\n",
			y.Year, y.Return*100, y.Income/1e3, y.Spent/1e3, y.FromBuffer/1e3, y.Sold/1e3, y.Tax/1e3, y.Open/1e3, y.Close/1e3)
	}
	// Output:
	// path 1 ruined true in year 29
	// year 0: market -12.7%, income  0.0, paid 45.0 (buffer  0.0), sold 53.4, tax 8.4, wealth 1000.0 -> 838.1
	// year 1: market  -9.3%, income  0.0, paid 45.0 (buffer 45.0), sold  0.0, tax 0.0, wealth 838.1 -> 723.4
	// year 2: market  +8.9%, income  0.0, paid 45.0 (buffer 45.0), sold  0.0, tax 0.0, wealth 723.4 -> 739.0
	// year 3: market -26.0%, income 15.0, paid 30.0 (buffer  0.0), sold 86.4, tax 11.4, wealth 739.0 -> 528.0
}
