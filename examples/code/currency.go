//go:build ignore

// Currency answers "what did the currency do to this investment?": a series
// in its own currency and converted into another, side by side, with the
// currency effect year by year and whether it cushioned or amplified the
// asset's moves (its monthly correlation with the asset's own returns).
//
// Usage:
//
//	go run examples/code/currency.go [-offline] [-in EUR] [-native USD] [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-years 15] ID
//
// Example:
//
//	go run examples/code/currency.go -in EUR IWDA
//
// ID must state a currency: a catalog fund does, a bundled reference series
// does not (-native says it). The conversion runs on daily crosses
// (Client.ConvertCurrency); the euro crosses are bundled, so converting to or
// from the euro works offline. Before the first known rate the conversion
// would hold that rate constant, a currency effect of zero that is no
// measurement, so the comparison starts at the first rate. ID goes through
// marketdata.Client.Load (a CSV path, else the bundle, else the network).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	offline := flag.Bool("offline", false, "never touch the network: the bundle and the quote cache only")
	target := flag.String("in", "EUR", "the currency to convert into")
	native := flag.String("native", "", "the series' currency when it states none (a bundled reference series)")
	from := flag.String("from", "", "first date kept, YYYY-MM-DD")
	to := flag.String("to", "", "last date kept, YYYY-MM-DD")
	years := flag.Int("years", 15, "calendar years listed, most recent last (0: all)")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: currency [flags] ID")
	}

	ctx := context.Background()
	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	client.Offline = *offline
	client.Logf = log.Printf
	s, err := client.Load(ctx, flag.Arg(0), marketdata.FetchOptions{From: date(*from), To: date(*to)})
	if err != nil {
		log.Fatal(err)
	}
	if s.Currency == "" {
		if *native == "" {
			log.Fatalf("%s states no currency: say it with -native", s.Symbol)
		}
		cp := *s // the client memoizes: state the currency on a copy
		cp.Currency = *native
		s = &cp
	}
	if s.Currency == *target {
		log.Fatalf("%s is already in %s", s.Symbol, *target)
	}
	conv, extrapolated, err := client.ConvertCurrency(ctx, s, *target, s.First().Date)
	if err != nil {
		log.Fatal(err)
	}
	if !extrapolated.IsZero() {
		fmt.Printf("no %s%s rate before %s: the comparison starts there\n", s.Currency, *target, extrapolated.Format(time.DateOnly))
		s, conv = marketdata.Trim(s, extrapolated, time.Time{}), marketdata.Trim(conv, extrapolated, time.Time{})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Printf("%s, %s to %s\n\n", s.Symbol, s.First().Date.Format(time.DateOnly), s.Last().Date.Format(time.DateOnly))
	fmt.Fprintln(tw, "in\tCAGR\tvol\tmax DD\tworst year\t")
	var cals [2][]metrics.PeriodReturn
	for i, x := range []*marketdata.Series{s, conv} {
		st, err := x.Stats()
		if err != nil {
			log.Fatal(err)
		}
		if cals[i], err = metrics.CalendarReturns(x.Dates(), x.Values(), 12); err != nil {
			log.Fatal(err)
		}
		worst := cals[i][metrics.LowestK(returns(cals[i]), 1)[0]]
		fmt.Fprintf(tw, "%s\t%.2f %%\t%.1f %%\t%.1f %%\t%+.1f %% (%d)\t\n", x.Currency, st.CAGR*100,
			st.Volatility*100, st.MaxDrawdown*100, worst.Return*100, worst.End.Year())
	}
	tw.Flush()

	// The currency effect of a period is what the conversion added to the
	// native return: (1 + converted) / (1 + native) - 1.
	fmt.Fprintf(tw, "\nyear\tin %s\tin %s\tcurrency effect\t\n", s.Currency, *target)
	first := 0
	if *years > 0 {
		first = max(0, len(cals[0])-*years)
	}
	for k := first; k < len(cals[0]); k++ {
		n, c := cals[0][k], cals[1][k]
		fmt.Fprintf(tw, "%d\t%+.1f %%\t%+.1f %%\t%+.1f %%\t\n", n.End.Year(), n.Return*100, c.Return*100,
			((1+c.Return)/(1+n.Return)-1)*100)
	}
	tw.Flush()

	// Monthly: does the currency move against the asset (a cushion) or
	// with it (an amplifier)?
	p, err := marketdata.NewPanel(marketdata.Monthly, s, relabel(conv, s.Symbol+"@"+*target))
	if err != nil {
		log.Fatal(err)
	}
	fx := make([]float64, p.Len())
	for t := range fx {
		fx[t] = (1+p.R[1][t])/(1+p.R[0][t]) - 1
	}
	fmt.Printf("\nmonthly currency effect: volatility %.1f %%/yr, correlation with the %s returns %+.2f",
		metrics.Volatility(fx, 12)*100, s.Currency, metrics.Corr(p.R[0], fx))
	fmt.Println(" (negative: the currency cushioned the falls)")
}

// returns lists the Return of each calendar period, NaN for a partial first
// one, which LowestK then never selects.
func returns(cal []metrics.PeriodReturn) []float64 {
	out := make([]float64, len(cal))
	for i, c := range cal {
		out[i] = c.Return
		if c.Partial {
			out[i] = math.NaN()
		}
	}
	return out
}

// relabel returns a copy of s under another Symbol, so a panel can hold it
// next to the series it was converted from.
func relabel(s *marketdata.Series, symbol string) *marketdata.Series {
	cp := *s
	cp.Symbol = symbol
	return &cp
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
