// Fire runs a quick decumulation study on history: how often does a 1 M
// portfolio spending 40 000 a year for 35 years run out, on the real returns
// of the bundled US 60/40 since 1954?
//
//	go run ./examples/lib/fire [-capital 1000000] [-spend 40000] [-years 35]
//
// Everything is REAL (inflation removed): the spending is constant in
// purchasing power and the returns are after inflation.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/replay"
	"github.com/bpineau/pofo/pkg/scenario"
)

func main() {
	log.SetFlags(0)
	capital := flag.Float64("capital", 1_000_000, "starting wealth")
	spend := flag.Float64("spend", 40_000, "real spending per year")
	years := flag.Int("years", 35, "horizon in years")
	flag.Parse()

	// The bundled real US 60/40 (S&P 500 and 5-year Treasuries, rebalanced
	// every January, deflated by the US CPI): one real return per calendar
	// year, as FRACTIONS.
	ref, err := replay.Reference()
	if err != nil {
		log.Fatal(err)
	}
	panel := scenario.Panel{Returns: [][]float64{ref.Annual}, Weights: []float64{1}}
	fmt.Printf("history: %d real years, %d to %d\n\n", len(ref.Annual), ref.Years[0], ref.Years[len(ref.Years)-1])

	// Monte-Carlo: a stationary bootstrap resamples that history in blocks
	// of five years on average, so bad years keep their neighbours.
	plan := decumul.Plan{
		Capital: *capital, NeedAnnual: *spend, Years: *years,
		Source: scenario.StationaryBootstrap{Panel: panel, MeanBlock: 5, Periods: *years},
	}
	o := plan.Simulate(20_000, 4, 1).Outcome() // paths, workers, seed
	fmt.Printf("bootstrap: ruin %.1f %%, median real wealth at the end %.2f M\n", o.RuinProb*100, o.TerminalP50/1e6)

	// The question turned around: the spending that keeps ruin at 5 %.
	safe := plan.Solve(0.05, decumul.WithdrawalAxis(10_000, 100_000), 20_000, 4, 1)
	rate := safe / *capital
	fmt.Printf("spending at 5 %% ruin: %.0f a year (%.2f %% of the capital)\n\n", safe, rate*100)

	// History as it happened: every start year with a full horizon behind it.
	cohorts := scenario.HistoricalCohorts{Panel: panel, Periods: *years}
	ruined, worst, worstYear := 0, 0.0, 0
	for i := range cohorts.Count() {
		r := plan.RunPath(cohorts.Cohort(i), decumul.Lives{})
		if r.Ruined {
			ruined++
		}
		if end := r.Wealth[len(r.Wealth)-1]; i == 0 || end < worst {
			worst, worstYear = end, ref.Years[i]
		}
	}
	fmt.Printf("history: %d start years, %d ran out; the worst start, %d, ended with %.2f M\n",
		cohorts.Count(), ruined, worstYear, worst/1e6)
}
