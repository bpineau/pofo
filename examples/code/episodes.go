//go:build ignore

// Episodes answers "what did each series do in the crises?": the cumulative
// return of several series over the named market episodes (dated on the S&P
// 500's peak and trough closes) and over any range of your own, and with -dd
// the deepest fall of each series inside each episode.
//
// Usage:
//
//	go run examples/code/episodes.go [-offline] [-currency EUR] [-dd] [-range NAME=YYYY-MM-DD:YYYY-MM-DD]... ID...
//
// Example:
//
//	go run examples/code/episodes.go -range "1994 bonds=1993-10-15:1994-11-07" SP500-USD TREASURY-LONG-USD XAUUSD-LBMA TREND-NET-USD
//
// An episode reads each series' close at or before its first date and its
// close at or before its last one (marketdata.Series.Change), so a monthly
// series reads the month-ends around it and blurs a crash that lasted
// weeks; "n/a" means the series does not cover the episode. Each ID goes
// through marketdata.Client.Load (a CSV path, else the bundle, else the
// network).
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
)

// episode is a named window: from the close of From to the close of To.
type episode struct {
	Name     string
	From, To time.Time
}

// crises are the S&P 500's peak and trough closes of the great drawdowns
// of the last half century, and the 2022 bond and equity fall.
var crises = []episode{
	{"1973-74 oil shock", day("1973-01-11"), day("1974-10-03")},
	{"1987 crash", day("1987-08-25"), day("1987-12-04")},
	{"2000-02 dot-com", day("2000-03-24"), day("2002-10-09")},
	{"2008 financial crisis", day("2007-10-09"), day("2009-03-09")},
	{"2020 covid", day("2020-02-19"), day("2020-03-23")},
	{"2022 inflation", day("2022-01-03"), day("2022-10-12")},
}

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	currency := flag.String("currency", "", "convert every series into this currency (default: each one's own)")
	dd := flag.Bool("dd", false, "also list each series' deepest fall inside each episode")
	flag.Func("range", "a custom episode, NAME=YYYY-MM-DD:YYYY-MM-DD (repeatable)", func(v string) error {
		name, span, ok1 := strings.Cut(v, "=")
		a, b, ok2 := strings.Cut(span, ":")
		from, err1 := time.Parse(time.DateOnly, a)
		to, err2 := time.Parse(time.DateOnly, b)
		if !ok1 || !ok2 || err1 != nil || err2 != nil || !from.Before(to) {
			return fmt.Errorf("want NAME=YYYY-MM-DD:YYYY-MM-DD, the first date before the second")
		}
		crises = append(crises, episode{name, from, to})
		return nil
	})
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: episodes [flags] ID...")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	var list []*marketdata.Series
	for _, id := range flag.Args() {
		s, err := client.Load(context.Background(), id, marketdata.FetchOptions{Currency: *currency})
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, s)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	table := func(title string, cell func(s *marketdata.Series, e episode) string) {
		fmt.Fprintf(tw, "%s\tfrom\tto\t", title)
		for _, s := range list {
			fmt.Fprintf(tw, "%s\t", s.Symbol)
		}
		fmt.Fprintln(tw)
		for _, e := range crises {
			fmt.Fprintf(tw, "%s\t%s\t%s\t", e.Name, e.From.Format(time.DateOnly), e.To.Format(time.DateOnly))
			for _, s := range list {
				fmt.Fprintf(tw, "%s\t", cell(s, e))
			}
			fmt.Fprintln(tw)
		}
		tw.Flush()
	}

	// Change refuses what it could only answer wrong: a series that starts
	// after the episode does, or stops before it ends.
	table("cumulative return", func(s *marketdata.Series, e episode) string {
		r, err := s.Change(e.From, e.To)
		if err != nil {
			return "n/a"
		}
		return fmt.Sprintf("%+.1f %%", r*100)
	})
	if *dd {
		fmt.Println()
		table("deepest fall inside", func(s *marketdata.Series, e episode) string {
			if _, err := s.Change(e.From, e.To); err != nil {
				return "n/a"
			}
			// The close at or before the episode's start, then every one
			// inside it.
			_, on, _ := s.At(e.From)
			w := marketdata.Trim(s, on, e.To)
			return fmt.Sprintf("%.1f %%", metrics.MaxDrawdown(w.Dates(), w.Values()).Depth*100)
		})
	}
}

// day parses a date of the crisis table.
func day(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}
