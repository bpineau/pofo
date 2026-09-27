//go:build ignore

// Fire answers "can this capital fund this spending?": how often a
// retirement drawing a constant real spending runs out, on bootstrapped
// history and on every start year history actually offered, and the spending
// that keeps the failure rate at a target.
//
// Usage:
//
//	go run examples/code/fire.go [-offline] [-capital 1000000] [-spend 40000] [-years 35] [-ruin 0.05] [-block 5] [-paths 20000] [-id ID [-cpi ^CPI-US] [-currency USD]]
//
// Example:
//
//	go run examples/code/fire.go -capital 1200000 -spend 42000 -years 40
//
// Everything is REAL (inflation removed): the spending is constant in
// purchasing power and the returns are after inflation. The default history
// is the bundled US 60/40 since 1954 (replay.Reference, deflated by the US
// CPI); -id replays any series instead, read through
// marketdata.Client.Load, cut to calendar years and deflated by -cpi (a
// euro series wants ^HICP-EA and -currency EUR). No tax, no fee, no pension:
// pofo -fire models those.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/replay"
	"github.com/bpineau/pofo/pkg/scenario"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	capital := flag.Float64("capital", 1_000_000, "starting wealth")
	spend := flag.Float64("spend", 40_000, "real spending per year, same unit")
	years := flag.Int("years", 35, "horizon in years")
	target := flag.Float64("ruin", 0.05, "the ruin probability the safe spending is solved at, a fraction")
	block := flag.Float64("block", 5, "mean block length of the bootstrap, in years")
	paths := flag.Int("paths", 20_000, "bootstrapped paths")
	id := flag.String("id", "", "replay this series instead of the bundled US 60/40")
	cpi := flag.String("cpi", "^CPI-US", "price index deflating -id")
	currency := flag.String("currency", "", "convert -id into this currency first (match it to -cpi)")
	flag.Parse()

	// One real return per calendar year, as FRACTIONS.
	annual, first, last := referenceYears()
	if *id != "" {
		annual, first, last = seriesYears(*id, *cpi, *currency, *offline)
	}
	if len(annual) < *years {
		log.Fatalf("%d years of history cannot hold a %d-year horizon", len(annual), *years)
	}
	growth := 1.0
	for _, r := range annual {
		growth *= 1 + r
	}
	worstYr := annual[metrics.LowestK(annual, 1)[0]]
	fmt.Printf("history: %d real years, %d to %d, %.2f %%/yr real, worst year %.1f %%\n\n", len(annual), first, last,
		100*(math.Pow(growth, 1/float64(len(annual)))-1), worstYr*100)

	panel := scenario.Panel{Returns: [][]float64{annual}, Weights: []float64{1}}
	plan := decumul.Plan{
		Capital: *capital, NeedAnnual: *spend, Years: *years,
		// A stationary bootstrap resamples the history in blocks of random
		// length, so a bad year keeps its neighbours.
		Source: scenario.StationaryBootstrap{Panel: panel, MeanBlock: *block, Periods: *years},
	}
	o := plan.Simulate(*paths, 4, 1).Outcome() // paths, workers, seed
	fmt.Printf("bootstrap: %.1f %% of %d paths ran out; median real wealth at the end %.0f, 5th percentile %.0f\n",
		o.RuinProb*100, *paths, o.TerminalP50, o.TerminalP5)

	// The question turned around: the spending that keeps ruin at the target.
	safe := plan.Solve(*target, decumul.WithdrawalAxis(*capital*0.005, *capital*0.15), *paths, 4, 1)
	fmt.Printf("spending at %.0f %% ruin: %.0f a year, %.2f %% of the capital\n\n", *target*100, safe, 100*safe / *capital)

	// History as it happened: every start year with a full horizon behind it.
	// The worst start is the one that ran out soonest, else the one that
	// ended poorest.
	cohorts := scenario.HistoricalCohorts{Panel: panel, Periods: *years}
	ruined, worst, worstStart := 0, decumul.PathResult{}, 0
	for i := range cohorts.Count() {
		r := plan.RunPath(cohorts.Cohort(i), decumul.Lives{})
		if r.Ruined {
			ruined++
		}
		if i == 0 || worse(r, worst) {
			worst, worstStart = r, first+i
		}
	}
	fmt.Printf("history: %d start years, %d ran out; ", cohorts.Count(), ruined)
	if worst.Ruined {
		fmt.Printf("the worst start, %d, ran out in its year %d\n", worstStart, worst.RuinYear+1)
	} else {
		fmt.Printf("the worst start, %d, ended with %.0f\n", worstStart, worst.Wealth[len(worst.Wealth)-1])
	}
}

// worse reports whether path a fared worse than path b: it ran out sooner,
// or neither ran out and it ended poorer.
func worse(a, b decumul.PathResult) bool {
	switch {
	case a.Ruined && b.Ruined:
		return a.RuinYear < b.RuinYear
	case a.Ruined != b.Ruined:
		return a.Ruined
	}
	return a.Wealth[len(a.Wealth)-1] < b.Wealth[len(b.Wealth)-1]
}

// referenceYears returns the bundled real US 60/40's calendar-year returns
// and the first and last years they cover.
func referenceYears() (annual []float64, first, last int) {
	ref, err := replay.Reference()
	if err != nil {
		log.Fatal(err)
	}
	return ref.Annual, ref.Years[0], ref.Years[len(ref.Years)-1]
}

// seriesYears returns the real calendar-year returns of id deflated by the
// price index cpi, over the complete years both cover, and the first and
// last of those years.
func seriesYears(id, cpi, currency string, offline bool) (annual []float64, first, last int) {
	ctx := context.Background()
	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = offline
	client.Logf = log.Printf
	s, err := client.Load(ctx, id, marketdata.FetchOptions{Currency: currency})
	if err != nil {
		log.Fatal(err)
	}
	// A price index is a level: loaded as it is, never converted. The
	// inflation indices are bundled, so they answer offline.
	index, err := client.Load(ctx, cpi, marketdata.FetchOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// A yearly panel keeps the calendar years both cover in full (a partial
	// first or last year is not a year), each column's return from one
	// year's last close to the next; the real return divides out inflation.
	p, err := marketdata.NewPanel(marketdata.Yearly, s, index)
	if err != nil {
		log.Fatal(err)
	}
	annual = make([]float64, p.Len())
	for t := range annual {
		annual[t] = (1+p.R[0][t])/(1+p.R[1][t]) - 1
	}
	return annual, p.Ends[0].Year(), p.Ends[p.Len()-1].Year()
}
