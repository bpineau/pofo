//go:build ignore

// Fees answers "what does a fee cost, and what does a share class really
// charge?". With one series and -fee, it prints the series gross and net of
// that annual charge, deducted continuously as a fund's is. With two share
// classes (or a fund and its index), it measures on their shared months how
// far the first trails the second each year, the gap a fee difference
// should explain, against the charges the catalog publishes.
//
// Usage:
//
//	go run examples/code/fees.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] -fee PCT ID
//	go run examples/code/fees.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] ID OTHER
//
// Example:
//
//	go run examples/code/fees.go -fee 0.5 MSCIWORLD-USD
//
// -fee is in PERCENT per year (0.5 = 0.5 %/yr), the convention of the
// catalog and of portfolio files; marketdata.Series.LessFee takes a
// FRACTION, and the conversion is written below. A fee gap inside two
// standard errors is not a measurement. Each ID goes through
// marketdata.Client.Load (a CSV path, else the bundle, else the network).
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
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert every series into this currency (default: each one's own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	fee := flag.Float64("fee", 0, "annual charge to deduct from a single series, in PERCENT per year")
	flag.Parse()
	if flag.NArg() < 1 || flag.NArg() > 2 || (flag.NArg() == 1) == (*fee == 0) {
		log.Fatal("usage: fees [flags] -fee PCT ID, or fees [flags] ID OTHER")
	}

	ctx := context.Background()
	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}
	var list []*marketdata.Series
	for _, id := range flag.Args() {
		s, err := client.Load(ctx, id, opt)
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, s)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	if len(list) == 1 {
		gross := list[0]
		net, err := gross.LessFee(*fee / 100) // PERCENT per year to the FRACTION LessFee takes
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s, %s to %s, %.2f %%/yr deducted\n\n", gross.Symbol,
			gross.First().Date.Format(time.DateOnly), gross.Last().Date.Format(time.DateOnly), *fee)
		fmt.Fprintln(tw, "\tCAGR\tvol\tmax DD\tgrowth of 100\t")
		for _, row := range []struct {
			name string
			s    *marketdata.Series
		}{{"gross", gross}, {"net", net}} {
			st, err := row.s.Stats()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Fprintf(tw, "%s\t%.2f %%\t%.1f %%\t%.1f %%\t%.0f\t\n", row.name, st.CAGR*100, st.Volatility*100,
				st.MaxDrawdown*100, 100*row.s.Last().Close/row.s.First().Close)
		}
		tw.Flush()
		fmt.Printf("\nthe fee took %.1f %% of the final wealth\n", 100*(1-net.Last().Close/gross.Last().Close))
		return
	}

	// Two series on the complete months they share: Track reads how the
	// first follows the second.
	p, err := marketdata.NewPanel(marketdata.Monthly, list...)
	if err != nil {
		log.Fatal(err)
	}
	tr, err := p.Track(p.IDs[0], p.IDs[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s against %s, %d months from %s to %s\n\n", p.IDs[0], p.IDs[1], tr.Periods,
		p.Starts[0].Format(time.DateOnly), p.Ends[p.Len()-1].Format(time.DateOnly))
	fmt.Printf("yearly gap         %+.2f pt/yr (standard error %.2f)\n", tr.Difference*100, tr.DifferenceSE()*100)
	fmt.Printf("tracking error     %.2f %%/yr\n", tr.TrackingError*100)
	fmt.Printf("correlation        %.3f, beta %.3f, volatility ratio %.3f\n", tr.Corr, tr.Beta, tr.VolRatio)
	for i, s := range list {
		// Client.Fees answers in PERCENT per year: the catalog's pinned
		// charge, else a cached or fetched one.
		if ter, ok := client.Fees(ctx, flag.Arg(i)); ok {
			fmt.Printf("published charge   %s %.2f %%/yr\n", s.Symbol, ter)
		}
	}
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
