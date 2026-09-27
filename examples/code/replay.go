//go:build ignore

// Replay answers "what life would each withdrawal rule have given me?": the
// seven canonical rules (fixed real withdrawal, flex cut, Guyton-Klinger
// guardrails, risk-based guardrails, Vanguard's bounded percentage,
// amortization, percentage of portfolio) run through the years as they
// happened from one start year, on the bundled real US 60/40. Not a failure
// probability: a portrait of the income each rule paid, year after year.
//
// Usage:
//
//	go run examples/code/replay.go [-start 1973] [-capital 1000000] [-spend 40000] [-years 40] [-mu 0.045] [-sigma 0.10] [-df 5] [-ruin 0.05] [-raise 1.5] [-v]
//
// Example:
//
//	go run examples/code/replay.go -start 1966 -years 35
//
// Everything is REAL (inflation removed) and gross of tax. -mu, -sigma and
// -df are the forward assumptions of the two horizon-aware rules; they
// should stay generic rather than fitted to the era replayed, since no
// retiree knew what was coming. -v prints every rule's income year by year.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bpineau/pofo/pkg/replay"
)

func main() {
	log.SetFlags(0)
	start := flag.Int("start", 1973, "calendar year the retirement starts")
	capital := flag.Float64("capital", 1_000_000, "starting wealth")
	spend := flag.Float64("spend", 40_000, "planned real spending per year, same unit")
	years := flag.Int("years", 40, "planned horizon in years")
	mu := flag.Float64("mu", 0.045, "assumed real mean return per year, a fraction")
	sigma := flag.Float64("sigma", 0.10, "assumed real volatility per year, a fraction")
	df := flag.Float64("df", 5, "assumed Student-t degrees of freedom")
	ruin := flag.Float64("ruin", 0.05, "the ruin probability the horizon-aware rules aim at, a fraction")
	raise := flag.Float64("raise", 1.5, "ceiling of the risk-based guardrail's raises, a multiple of the spending")
	verbose := flag.Bool("v", false, "print every rule's real income year by year")
	flag.Parse()

	res, err := replay.Run(replay.Setup{
		Start: *start, Capital: *capital, Spend: *spend, Years: *years,
		Mu: *mu, Sigma: *sigma, Df: *df, TargetRuin: *ruin, RaiseCap: *raise,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n%d to %d", replay.Label, res.Start, res.End)
	if res.Partial {
		fmt.Printf(" (the record stops %d years into the %d-year plan)", res.Years, *years)
	}
	fmt.Printf(": real CAGR %.1f %%/yr, first decade %.1f %%/yr, worst year %d at %.1f %%, deepest fall %.1f %%\n\n",
		res.CAGR*100, res.Decade*100, res.WorstAt, res.WorstYear*100, res.MaxDrawdown*100)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "rule\tmean income\tvariation\tleanest year\tlean years\ttotal paid\tleft at the end\tran out\t")
	for _, r := range res.Rules {
		out := "no"
		if r.Ruined {
			out = fmt.Sprint(r.RuinYear)
		}
		fmt.Fprintf(tw, "%s\t%.0f\t%.0f %%\t%.0f\t%d\t%.0f\t%.0f\t%s\t\n",
			r.Name, r.Mean, r.CV*100, r.Low, r.LeanYears, r.Total, r.Final, out)
	}
	tw.Flush()
	fmt.Printf("\nlean years: below the planned %.0f; variation: the coefficient of variation of the yearly income\n", *spend)

	if *verbose {
		fmt.Fprint(tw, "\nyear\t")
		for _, r := range res.Rules {
			fmt.Fprintf(tw, "%s\t", r.Tag)
		}
		fmt.Fprintln(tw)
		for y := range res.Years {
			fmt.Fprintf(tw, "%d\t", res.Start+y)
			for _, r := range res.Rules {
				if y < len(r.Spend) {
					fmt.Fprintf(tw, "%.0f\t", r.Spend[y])
				} else {
					fmt.Fprint(tw, "\t")
				}
			}
			fmt.Fprintln(tw)
		}
		tw.Flush()
	}
}
