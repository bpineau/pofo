//go:build ignore

// Stats answers "how do these series compare?": the headline statistics of
// several series side by side, each on its own history or all on the window
// they share, at their own cadence or all on month-end closes.
//
// Usage:
//
//	go run examples/code/stats.go [-offline] [-currency EUR] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-common] [-monthly] ID...
//
// Example:
//
//	go run examples/code/stats.go -monthly SP500-USD MSCIWORLD-USD TREASURY-LONG-USD XAUUSD-LBMA
//
// Each ID goes through marketdata.Client.Load (a CSV path, else the bundle,
// else the network). Comparing a daily series with a monthly one on their
// own cadences compares two different volatilities: -monthly reads every
// series on the complete calendar months they all share (a
// marketdata.Panel), -common keeps each one's own cadence on the window they
// share.
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
	common := flag.Bool("common", false, "measure every series on the window they all share")
	monthly := flag.Bool("monthly", false, "read every series on the complete months they all share (implies -common)")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: stats [flags] ID...")
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
	switch {
	case *monthly:
		// A panel keeps the complete calendar months every series quotes,
		// so the columns share one calendar, a partial last month dropped;
		// Series turns each column back into a month-end level.
		p, err := marketdata.NewPanel(marketdata.Monthly, list...)
		if err != nil {
			log.Fatal(err)
		}
		for i, id := range p.IDs {
			if list[i], err = p.Series(id); err != nil {
				log.Fatal(err)
			}
		}
	case *common:
		start, end, ok := marketdata.CommonWindow(list...)
		if !ok {
			log.Fatal("the series share no window")
		}
		for i, s := range list {
			list[i] = marketdata.Trim(s, start, end)
		}
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "id\tfrom\tto\tper yr\tCAGR\tvol\tSharpe\tSortino\tmax DD\tunderwater\tworst year\t")
	for _, s := range list {
		st, err := s.Stats()
		if err != nil {
			log.Fatal(err)
		}
		// The worst calendar year, the first one left out when partial.
		worst := "n/a"
		if cal, err := metrics.CalendarReturns(s.Dates(), s.Values(), 12); err == nil {
			var w *metrics.PeriodReturn
			for i := range cal {
				if !cal[i].Partial && (w == nil || cal[i].Return < w.Return) {
					w = &cal[i]
				}
			}
			if w != nil {
				worst = fmt.Sprintf("%+.1f %% (%d)", w.Return*100, w.End.Year())
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%.0f\t%.2f %%\t%.1f %%\t%.2f\t%.2f\t%.1f %%\t%.1f y\t%s\t\n",
			s.Symbol, st.Start.Format(time.DateOnly), st.End.Format(time.DateOnly), st.PeriodsPerYear,
			st.CAGR*100, st.Volatility*100, st.Sharpe, st.Sortino, st.MaxDrawdown*100,
			float64(st.TTRDays)/365.25, worst)
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
