//go:build ignore

// Correl answers "do these holdings diversify each other?": the monthly
// correlation matrix of a portfolio file's holdings (or of any identifiers),
// on the complete calendar months they all share, with each one's volatility
// over those months.
//
// Usage:
//
//	go run examples/code/correl.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-daily] FILE.txt|ID...
//
// Example:
//
//	go run examples/code/correl.go examples/portfolios/all-weather-dalio.txt
//
// An argument ending in ".txt" is a portfolio file whose holdings join the
// matrix (under "#meta sim:on" with their SIM suffix); any other is an
// identifier, read through marketdata.Client.Load (a CSV path, else the
// bundle, else the network). Monthly is the default because daily closes
// struck on different exchanges (a US fund against a European listing)
// understate their correlation; -daily reads the sessions they all quote.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "EUR", "convert every series into this currency (\"\" keeps each one's own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	daily := flag.Bool("daily", false, "correlate daily returns on the sessions all quote (default: monthly)")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: correl [flags] FILE.txt|ID...")
	}

	// The identifiers to load, as written, and the names to show them by.
	var ids, names []string
	for _, arg := range flag.Args() {
		if !strings.HasSuffix(arg, ".txt") {
			ids, names = append(ids, arg), append(names, arg)
			continue
		}
		spec, err := portfolio.ParseFile(arg)
		if err != nil {
			log.Fatal(err)
		}
		for _, h := range spec.Holdings {
			ids = append(ids, portfolio.SimFetchID(h.ID, spec.Sim))
			names = append(names, h.ID)
		}
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}
	var list []*marketdata.Series
	for i, id := range ids {
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		// A panel names its columns by Symbol and refuses a duplicate: label
		// each series as written, on a copy (the client memoizes).
		cp := *s
		cp.Symbol = names[i]
		list = append(list, &cp)
	}

	freq := marketdata.Monthly
	if *daily {
		freq = marketdata.Daily
	}
	p, err := marketdata.NewPanel(freq, list...)
	if err != nil {
		log.Fatal(err)
	}
	corr, err := metrics.CorrelationMatrix(p.R)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d periods (%.0f a year) from %s to %s, in %s\n\n", p.Len(), p.PeriodsPerYear(),
		p.Starts[0].Format(time.DateOnly), p.Ends[p.Len()-1].Format(time.DateOnly), currencyLabel(*currency))

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprint(tw, "\tvol\t")
	for i := range p.IDs {
		fmt.Fprintf(tw, "%d\t", i+1)
	}
	fmt.Fprintln(tw)
	for i, row := range corr {
		fmt.Fprintf(tw, "%d %s\t%.1f %%\t", i+1, p.IDs[i], metrics.Volatility(p.R[i], p.PeriodsPerYear())*100)
		for j, c := range row {
			if j == i {
				fmt.Fprint(tw, ".\t")
				continue
			}
			fmt.Fprintf(tw, "%+.2f\t", c)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

// currencyLabel names the currency flag's value for the header.
func currencyLabel(c string) string {
	if c == "" {
		return "each series' own currency"
	}
	return c
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
