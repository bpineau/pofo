// The -rates mode: chart interest-rate LEVELS (policy, interbank, yields).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// runRates charts rate levels in the terminal, with a summary line per symbol.
//
// It is deliberately NOT the comparison path: a rate is a level in annualized
// percent, not a price, so "base 100" and every return statistic would be
// meaningless on it (and a rate crossing zero, as the euro rates did for eight
// years, breaks them outright). Here the level itself is the subject.
func runRates(ctx context.Context, opt *options, c *marketdata.Client, list string) error {
	symbols := splitSymbols(list)
	if len(symbols) == 0 || strings.EqualFold(list, "list") {
		printRateCatalog()
		return nil
	}
	series := make([]chart.Series, 0, len(symbols))
	type row struct {
		name             string
		first, last      time.Time
		lo, hi, now, avg float64
	}
	rows := make([]row, 0, len(symbols))
	for _, sym := range symbols {
		s, err := c.Fetch(ctx, sym, opt.start)
		if err != nil {
			return fmt.Errorf("%s: %w", sym, err)
		}
		if !opt.end.IsZero() {
			s = marketdata.Trim(s, opt.start, opt.end)
		}
		if len(s.Points) == 0 {
			return fmt.Errorf("%s: no observation in the requested window", sym)
		}
		dates := make([]time.Time, len(s.Points))
		values := make([]float64, len(s.Points))
		r := row{name: sym, first: s.First().Date, last: s.Last().Date,
			lo: s.Points[0].Close, hi: s.Points[0].Close, now: s.Last().Close}
		var sum float64
		for i, p := range s.Points {
			dates[i], values[i] = p.Date, p.Close
			sum += p.Close
			r.lo, r.hi = min(r.lo, p.Close), max(r.hi, p.Close)
		}
		r.avg = sum / float64(len(s.Points))
		rows = append(rows, r)
		series = append(series, chart.Series{Name: sym, Dates: dates, Values: values})
	}

	color := os.Getenv("NO_COLOR") == "" && isTerminal(os.Stdout)
	fmt.Print(chart.Term(chart.TermOptions{
		Title: "Rate levels (annualized %)", Width: termWidth(opt.width), Color: color,
	}, series))
	fmt.Println()
	fmt.Printf("%-14s %10s %10s %10s %10s  %s\n", "symbol", "last", "min", "max", "average", "window")
	fmt.Println(strings.Repeat("─", 78))
	for _, r := range rows {
		fmt.Printf("%-14s %9.2f%% %9.2f%% %9.2f%% %9.2f%%  %s → %s\n", r.name, r.now, r.lo, r.hi, r.avg,
			r.first.Format("2006-01-02"), r.last.Format("2006-01-02"))
	}
	return nil
}

// printRateCatalog lists what -rates understands, the registry first.
func printRateCatalog() {
	fmt.Println("Policy and money-market rates (published levels):")
	for _, s := range marketdata.RateSymbols() {
		fmt.Printf("  %-12s %s\n", s, marketdata.RateName(s))
	}
	fmt.Println("\nMarket yields (exchange quotes), usable here too:")
	for _, y := range [][2]string{
		{"^IRX", "US Treasury bill, 13 weeks"},
		{"^FVX", "US Treasury note, 5 years"},
		{"^TNX", "US Treasury note, 10 years"},
		{"^TYX", "US Treasury bond, 30 years"},
	} {
		fmt.Printf("  %-12s %s\n", y[0], y[1])
	}
	fmt.Println("\nExample: pofo -rates ^ESTR,^EURIBOR3M,^ECB-DFR -start 2015-01-01")
}

// splitSymbols parses a comma-separated symbol list, trimming blanks.
func splitSymbols(list string) []string {
	var out []string
	for _, s := range strings.Split(list, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, strings.ToUpper(s))
		}
	}
	return out
}
