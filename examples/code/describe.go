//go:build ignore

// Describe answers "what is this series?": where it comes from, its
// statistics, its calendar years and its deepest drawdowns, dated.
//
// Usage:
//
//	go run examples/code/describe.go [-offline] [-currency EUR] [-from 2000-01-01] [-to 2020-12-31] [-years 10] [-dd 5] ID
//
// Example:
//
//	go run examples/code/describe.go SP500-USD
//
// ID goes through marketdata.Client.Load: a CSV path is a file, a reference
// series or a catalog index (pofo -dump list names them) comes from the
// bundle, offline, and anything else is fetched: a fund's real quotes under
// its plain name, with the bundled reconstruction in front under its SIM
// name (IWDA against IWDASIM).
package main

import (
	"cmp"
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
	currency := flag.String("currency", "", "convert into this currency (default: the series' own)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	years := flag.Int("years", 10, "calendar years listed, most recent last (0: all)")
	dd := flag.Int("dd", 5, "deepest drawdowns listed")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: describe [flags] ID")
	}

	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf // resolution lines and warnings, on stderr
	s, err := client.Load(context.Background(), flag.Arg(0), marketdata.FetchOptions{
		From: date(*from), To: date(*to), Currency: *currency,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s  %s\n", s.Symbol, s.Name)
	fmt.Printf("%s, %d points from %s to %s, currency %q\n", s.Source, s.Len(),
		s.First().Date.Format(time.DateOnly), s.Last().Date.Format(time.DateOnly), s.Currency)
	if !s.SimulatedBefore.IsZero() {
		fmt.Printf("reconstructed before %s (from %s)\n", s.SimulatedBefore.Format(time.DateOnly), s.ProxySymbol)
	}

	// Stats annualizes at the series' own cadence (252 daily, 12 monthly).
	st, err := s.Stats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nCAGR %.2f %%/yr   volatility %.1f %%/yr (%.0f periods a year)   Sharpe %.2f   Sortino %.2f\n",
		st.CAGR*100, st.Volatility*100, st.PeriodsPerYear, st.Sharpe, st.Sortino)
	fmt.Printf("max drawdown %.1f %%   longest underwater %.1f years   Ulcer %.1f   skew %.2f   excess kurtosis %.1f\n",
		st.MaxDrawdown*100, float64(st.TTRDays)/365.25, st.Ulcer, st.Skew, st.Kurtosis)

	// Calendar years, each from the previous year's last close. The first
	// one is Partial when the series starts inside it; the last one ends on
	// the last quote, which its End says.
	cal, err := metrics.CalendarReturns(s.Dates(), s.Values(), 12)
	if err != nil {
		log.Fatal(err)
	}
	if *years > 0 && len(cal) > *years {
		cal = cal[len(cal)-*years:]
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "\nyear\treturn\t\t")
	for _, y := range cal {
		note := ""
		switch {
		case y.Partial:
			note = "from " + y.Start.Format(time.DateOnly)
		case y.End.Month() != time.December:
			note = "to " + y.End.Format(time.DateOnly)
		}
		fmt.Fprintf(tw, "%d\t%+.1f %%\t%s\t\n", y.End.Year(), y.Return*100, note)
	}
	tw.Flush()

	// Every drawdown episode, dated; the deepest ones.
	eps := metrics.DrawdownEpisodes(s.Dates(), s.Values())
	slices.SortFunc(eps, func(a, b metrics.Episode) int { return cmp.Compare(a.Depth, b.Depth) })
	fmt.Fprintln(tw, "\ndepth\tpeak\ttrough\trecovered\tyears underwater\t")
	for _, e := range eps[:min(*dd, len(eps))] {
		back, under := "not yet", e.DrawdownDays
		if !e.Ongoing {
			back, under = e.RecoverDate.Format("2006-01"), e.DrawdownDays+e.RecoveryDays
		}
		fmt.Fprintf(tw, "%.1f %%\t%s\t%s\t%s\t%.1f\t\n", e.Depth*100,
			e.PeakDate.Format("2006-01"), e.TroughDate.Format("2006-01"), back, float64(under)/365.25)
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
