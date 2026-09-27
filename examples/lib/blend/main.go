// Blend builds a monthly returns panel of two bundled series, adds a 60/40
// blend rebalanced every month, and compares the three.
//
//	go run ./examples/lib/blend [-from 1973-01-01]
//
// The legs are the bundled S&P 500 total return and a 5-year US Treasury
// total return, both month-end series in dollars.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	from := flag.String("from", "", "first month-end to keep, YYYY-MM-DD (default: as early as both go)")
	flag.Parse()

	equity, err := marketdata.Bundled("SP500-USD")
	if err != nil {
		log.Fatal(err)
	}
	bonds, err := marketdata.Bundled("TREASURY-INT-USD")
	if err != nil {
		log.Fatal(err)
	}

	// A panel holds RETURNS on the periods every series shares, labelled
	// by the calendar month-end, so each row spans the same interval in
	// every column.
	p, err := marketdata.NewPanel(marketdata.Monthly, equity, bonds)
	if err != nil {
		log.Fatal(err)
	}
	if *from != "" {
		start, err := time.Parse(time.DateOnly, *from)
		if err != nil {
			log.Fatal(err)
		}
		p = p.Between(start, time.Time{})
	}

	// Mix appends a column: the blend rebalanced every period. Weights are
	// FRACTIONS summing to 1.
	p, err = p.Mix("60/40", map[string]float64{"SP500-USD": 0.6, "TREASURY-INT-USD": 0.4})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d months, %s to %s\n\n", p.Len(),
		p.Starts[0].Format(time.DateOnly), p.Ends[p.Len()-1].Format(time.DateOnly))

	// Series turns a column back into a level, which Stats scores.
	fmt.Printf("%-18s %7s %7s %7s %8s\n", "", "CAGR", "vol", "Sharpe", "max DD")
	for _, id := range p.IDs {
		s, err := p.Series(id)
		if err != nil {
			log.Fatal(err)
		}
		st, err := s.Stats()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-18s %6.2f%% %6.1f%% %7.2f %7.1f%%\n", id, st.CAGR*100, st.Volatility*100, st.Sharpe, st.MaxDrawdown*100)
	}

	eq, err := p.Col("SP500-USD")
	if err != nil {
		log.Fatal(err)
	}
	bd, err := p.Col("TREASURY-INT-USD")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nmonthly correlation of the two legs: %.2f\n", metrics.Corr(eq, bd))
}
