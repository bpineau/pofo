//go:build ignore

// Worstmonths answers "does this hedge work when it matters?": the worst
// months of a reference series, dated, with what the other series did in
// exactly those months, then their mean return and hit rate over the
// reference's worst tail against every month.
//
// Usage:
//
//	go run examples/code/worstmonths.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-k 10] [-tail 0.1] REF ID...
//
// Example:
//
//	go run examples/code/worstmonths.go SP500-USD TREASURY-LONG-USD XAUUSD-LBMA TREND-NET-USD
//
// The series meet on the complete calendar months they all share, so the
// youngest one sets the window. Each ID goes through marketdata.Client.Load
// (a CSV path, else the bundle, else the network).
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
	k := flag.Int("k", 10, "worst months of REF listed")
	tail := flag.Float64("tail", 0.1, "share of REF's months in the worst tail the means read")
	flag.Parse()
	if flag.NArg() < 2 {
		log.Fatal("usage: worstmonths [flags] REF ID...")
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
	p, err := marketdata.NewPanel(marketdata.Monthly, list...)
	if err != nil {
		log.Fatal(err)
	}
	ref := p.R[0] // R[i][t]: the return of IDs[i] over month t, a fraction
	fmt.Printf("%d months, %s to %s\n\n", p.Len(), p.Ends[0].Format("2006-01"), p.Ends[p.Len()-1].Format("2006-01"))

	// LowestK returns POSITIONS, so the worst months come with their dates
	// and every other column reads on the same rows.
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprint(tw, "month\t")
	for _, id := range p.IDs {
		fmt.Fprintf(tw, "%s\t", id)
	}
	fmt.Fprintln(tw)
	for _, t := range metrics.LowestK(ref, *k) {
		fmt.Fprintf(tw, "%s\t", p.Ends[t].Format("2006-01"))
		for i := range p.IDs {
			fmt.Fprintf(tw, "%+.1f %%\t", p.R[i][t]*100)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()

	// Pick keeps a sample of periods: the worst tail of REF, and its best
	// tail for symmetry, against every month.
	n := max(1, int(float64(p.Len())**tail+0.5))
	worst, err := p.Pick(metrics.LowestK(ref, n))
	if err != nil {
		log.Fatal(err)
	}
	best, err := p.Pick(metrics.HighestK(ref, n))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nmean monthly return, and the share of months each series gained (%d months a tail)\n", n)
	fmt.Fprint(tw, "\t")
	for _, id := range p.IDs {
		fmt.Fprintf(tw, "%s\t", id)
	}
	fmt.Fprintln(tw)
	for _, row := range []struct {
		name  string
		panel *marketdata.Panel
	}{{"every month", p}, {"REF's worst tail", worst}, {"REF's best tail", best}} {
		fmt.Fprintf(tw, "%s\t", row.name)
		for i := range row.panel.IDs {
			col := row.panel.R[i]
			up := 0
			for _, r := range col {
				if r > 0 {
					up++
				}
			}
			fmt.Fprintf(tw, "%+.2f %% (%.0f %% up)\t", metrics.Mean(col)*100, 100*float64(up)/float64(len(col)))
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
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
