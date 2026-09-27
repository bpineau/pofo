//go:build ignore

// Rolling answers "what did holding it for N years deliver, at worst?": for
// each series, every N-year holding period (bought at any close, sold N years
// later), summarized as the worst, median and best annualized return, the
// share of the periods that lost money, and the worst twelve months, each
// dated.
//
// Usage:
//
//	go run examples/code/rolling.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-years 10] [-monthly] ID...
//
// Example:
//
//	go run examples/code/rolling.go -years 15 SP500-USD MSCIWORLD-USD TREASURY-LONG-USD XAUUSD-LBMA
//
// Periods overlap, so a century of month-ends yields over a thousand
// ten-year periods and barely ten independent ones: the worst is a single
// history, not a probability. A daily series is read at every session, a
// monthly one at its month-ends; -monthly reads every series at month-ends,
// which puts them on one footing. Each ID goes through marketdata.Client.Load
// (a CSV path, else the bundle, else the network).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
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
	years := flag.Float64("years", 10, "holding period in years")
	monthly := flag.Bool("monthly", false, "read every series at its month-end closes")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: rolling [flags] ID...")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	opt := marketdata.FetchOptions{From: date(*from), To: date(*to), Currency: *currency}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tw, "id\tperiods\tworst %gy\tbought\tmedian\tbest\tlost money\tworst 12 months\tended\t\n", *years)
	for _, id := range flag.Args() {
		s, err := client.Load(context.Background(), id, opt)
		if err != nil {
			log.Fatal(err)
		}
		if *monthly {
			if s, err = s.Resample(marketdata.Monthly); err != nil {
				log.Fatal(err)
			}
		}
		held := metrics.RollingCAGRs(s.Dates(), s.Values(), *years)
		if len(held) == 0 {
			fmt.Fprintf(tw, "%s\t0\tshorter than %g years\t\t\t\t\t\t\t\n", s.Symbol, *years)
			continue
		}
		worst, best := extremes(held)
		cagrs := make([]float64, len(held))
		lost := 0
		for i, h := range held {
			cagrs[i] = h.CAGR
			if h.CAGR < 0 {
				lost++
			}
		}
		median := metrics.Quantiles(cagrs, 0.5)[0]

		// The worst twelve months: the worst one-year holding period, whose
		// CAGR is its plain return.
		year := "n/a\t"
		if ones := metrics.RollingCAGRs(s.Dates(), s.Values(), 1); len(ones) > 0 {
			w, _ := extremes(ones)
			year = fmt.Sprintf("%+.1f %%\t%s", w.CAGR*100, w.End.Format("2006-01"))
		}
		fmt.Fprintf(tw, "%s\t%d\t%+.2f %%\t%s\t%+.2f %%\t%+.2f %%\t%.1f %%\t%s\t\n", s.Symbol, len(held),
			worst.CAGR*100, worst.Start.Format("2006-01"), median*100, best.CAGR*100,
			100*float64(lost)/float64(len(held)), year)
	}
	tw.Flush()
}

// extremes returns the holding periods of lowest and highest CAGR.
func extremes(held []metrics.HoldingPeriod) (worst, best metrics.HoldingPeriod) {
	byCAGR := func(a, b metrics.HoldingPeriod) int {
		switch {
		case a.CAGR < b.CAGR:
			return -1
		case a.CAGR > b.CAGR:
			return 1
		}
		return 0
	}
	return slices.MinFunc(held, byCAGR), slices.MaxFunc(held, byCAGR)
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
