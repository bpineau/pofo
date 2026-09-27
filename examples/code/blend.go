//go:build ignore

// Blend answers "what would this mix have done?": a blend of series
// rebalanced every month, its statistics against its legs', and the legs'
// monthly correlation matrix.
//
// Usage:
//
//	go run examples/code/blend.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] ID=WEIGHT...
//
// Example:
//
//	go run examples/code/blend.go SP500-USD=60 TREASURY-INT-USD=40
//
// Weights are in PERCENT and must sum to 100; a negative one is a short or a
// financing leg (EQ=90 BOND=60 EURCASH-EUR=-50 is a 150 % book paying the
// cash rate on what it borrows). The blend is rebalanced at every month-end,
// on the complete calendar months every leg quotes. Each ID goes through
// marketdata.Client.Load (a CSV path, else the bundle, else the network).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert every leg into this currency (default: each one's own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	flag.Parse()
	if flag.NArg() < 2 {
		log.Fatal("usage: blend [flags] ID=WEIGHT ID=WEIGHT...")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}

	var legs []*marketdata.Series
	weights := map[string]float64{} // by the Symbol Load returns, the panel's column name
	for _, arg := range flag.Args() {
		id, w, ok := strings.Cut(arg, "=")
		pct, err := strconv.ParseFloat(w, 64)
		if !ok || err != nil {
			log.Fatalf("%q: want ID=WEIGHT, the weight in percent", arg)
		}
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		legs = append(legs, s)
		weights[s.Symbol] = pct / 100 // Mix takes FRACTIONS
	}

	// A panel holds RETURNS on the periods every series shares, labelled by
	// the calendar month-end; Mix appends the blend as one more column.
	p, err := marketdata.NewPanel(marketdata.Monthly, legs...)
	if err != nil {
		log.Fatal(err)
	}
	p, err = p.Mix("blend", weights)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d months, %s to %s, rebalanced monthly\n\n", p.Len(),
		p.Starts[0].Format(time.DateOnly), p.Ends[p.Len()-1].Format(time.DateOnly))

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "\tweight\tCAGR\tvol\tSharpe\tmax DD\tunderwater\t")
	for _, id := range p.IDs {
		s, err := p.Series(id) // the column back as a month-end level
		if err != nil {
			log.Fatal(err)
		}
		st, err := s.Stats()
		if err != nil {
			log.Fatal(err)
		}
		w := "100 %"
		if id != "blend" {
			w = fmt.Sprintf("%.0f %%", weights[id]*100)
		}
		fmt.Fprintf(tw, "%s\t%s\t%.2f %%\t%.1f %%\t%.2f\t%.1f %%\t%.1f y\t\n", id, w,
			st.CAGR*100, st.Volatility*100, st.Sharpe, st.MaxDrawdown*100, float64(st.TTRDays)/365.25)
	}
	tw.Flush()

	// R is [column][period]: the legs only, the blend being their mix.
	corr, err := metrics.CorrelationMatrix(p.R[:len(legs)])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nmonthly correlation")
	fmt.Fprint(tw, "\t")
	for i := range legs {
		fmt.Fprintf(tw, "%d\t", i+1)
	}
	fmt.Fprintln(tw)
	for i, row := range corr {
		fmt.Fprintf(tw, "%d %s\t", i+1, p.IDs[i])
		for _, c := range row {
			fmt.Fprintf(tw, "%+.2f\t", c)
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
