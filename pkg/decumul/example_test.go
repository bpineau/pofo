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
	// ruin 25%, median terminal wealth 0.5 M
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
	// broke while alive 11%, for 1.0 years on average
	// median life 31 years, still alive at the horizon 0%
	// estate p50 0.8 M, nothing left 11%
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
	// reversion 100%: broke while alive 2%
	// reversion 54%: broke while alive 4%
	// reversion 0%: broke while alive 8%
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
	// no annuity    broke while alive 30%, median estate 0.33 M
	// a third at 65 broke while alive 24%, median estate 0.29 M
	// a third at 75 broke while alive 32%, median estate 0.23 M
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
	// 5% chance of ever being broke and alive at 24000 a year
}
