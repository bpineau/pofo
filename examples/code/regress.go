//go:build ignore

// Regress answers "what is this series made of?": the returns of one series
// fitted on those of others by ordinary least squares, with the standard
// error and t-statistic of every coefficient, the annualized alpha, the R2
// and the residual volatility.
//
// Usage:
//
//	go run examples/code/regress.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-daily] Y X1 [X2...]
//
// Example:
//
//	go run examples/code/regress.go -from 1972-01-01 XAUUSD-LBMA SP500-USD TREASURY-LONG-USD
//
// The series meet on the complete calendar months they all share (a
// marketdata.Panel), or with -daily on the sessions they all quote; a
// monthly series among daily ones calls for the monthly calendar. Each ID
// goes through marketdata.Client.Load (a CSV path, else the bundle, else the
// network).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert every series into this currency (default: each one's own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	daily := flag.Bool("daily", false, "regress daily returns on the sessions every series quotes (default: monthly)")
	flag.Parse()
	if flag.NArg() < 2 {
		log.Fatal("usage: regress [flags] Y X1 [X2...]")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}
	var list []*marketdata.Series
	for _, id := range flag.Args() {
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, s)
	}

	freq := marketdata.Monthly
	if *daily {
		freq = marketdata.Daily
	}
	p, err := marketdata.NewPanel(freq, list...)
	if err != nil {
		log.Fatal(err)
	}
	cols := make([][]float64, len(p.IDs))
	for i, id := range p.IDs {
		if cols[i], err = p.Col(id); err != nil {
			log.Fatal(err)
		}
	}

	// Everything in a Regression is PER PERIOD; AnnualAlpha and
	// AnnualResidualVol annualize at the panel's measured cadence.
	reg, err := metrics.Regress(cols[0], cols[1:]...)
	if err != nil {
		log.Fatal(err)
	}
	ppy := p.PeriodsPerYear()
	fmt.Printf("%s on %d regressors: %d periods (%.0f a year) from %s to %s\n\n", p.IDs[0], len(cols)-1,
		reg.N, ppy, p.Starts[0].Format(time.DateOnly), p.Ends[p.Len()-1].Format(time.DateOnly))

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "\tcoefficient\tstd error\tt\t")
	fmt.Fprintf(tw, "alpha (per year)\t%+.2f %%\t%.2f %%\t%+.1f\t\n",
		reg.AnnualAlpha(ppy)*100, reg.Alpha.SE*ppy*100, reg.Alpha.T)
	for i, b := range reg.Betas {
		fmt.Fprintf(tw, "beta on %s\t%+.3f\t%.3f\t%+.1f\t\n", p.IDs[i+1], b.Value, b.SE, b.T)
	}
	tw.Flush()
	fmt.Printf("\nR2 %.3f, residual volatility %.1f %%/yr\n", reg.R2, reg.AnnualResidualVol(ppy)*100)
	fmt.Println("|t| under about 2: the coefficient is not distinguishable from zero.")
}

// date parses a YYYY-MM-DD flag; empty is the zero time, an open bound.
func date(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		log.Fatalf("bad date %q: want YYYY-MM-DD", s)
	}
	return t
}
