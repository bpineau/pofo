// Describe loads one bundled series and describes it: where it comes from,
// its statistics, its last calendar years and its deepest drawdowns.
//
//	go run ./examples/lib/describe [ID]
//
// ID is any identifier marketdata.BundledIDs lists (a catalog asset's
// backcast such as IWDA, or a reference series such as SP500-USD); the
// default is IWDA.
package main

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	id := "IWDA" // iShares Core MSCI World, IE00B4L5Y983
	if len(os.Args) > 1 {
		id = os.Args[1]
	}

	// Bundled needs no client and no network: the series the binary embeds.
	s, err := marketdata.Bundled(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %s\n", s.Symbol, s.Name)
	fmt.Printf("%s file, %d points from %s to %s, currency %q\n", s.Source, s.Len(),
		s.First().Date.Format(time.DateOnly), s.Last().Date.Format(time.DateOnly), s.Currency)

	// Stats annualizes at the series' own cadence (252 daily, 12 monthly).
	st, err := s.Stats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CAGR %.2f %%/yr, volatility %.1f %%/yr (%.0f periods a year), Sharpe %.2f, max drawdown %.1f %%\n\n",
		st.CAGR*100, st.Volatility*100, st.PeriodsPerYear, st.Sharpe, st.MaxDrawdown*100)

	// Calendar years, each from the previous year's last close; the first
	// one is Partial when the series starts inside it, and the last one
	// ends on the last quote, which its End says.
	years, err := metrics.CalendarReturns(s.Dates(), s.Values(), 12)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("last ten calendar years:")
	for _, y := range years[max(0, len(years)-10):] {
		fmt.Printf("  to %s %+6.1f %%\n", y.End.Format(time.DateOnly), y.Return*100)
	}

	// Every drawdown episode, dated; the three deepest.
	eps := metrics.DrawdownEpisodes(s.Dates(), s.Values())
	slices.SortFunc(eps, func(a, b metrics.Episode) int { return cmp.Compare(a.Depth, b.Depth) })
	fmt.Println("\ndeepest drawdowns:")
	for _, e := range eps[:min(3, len(eps))] {
		back := "not recovered"
		if !e.Ongoing {
			back = "recovered " + e.RecoverDate.Format("2006-01")
		}
		fmt.Printf("  %+6.1f %%  peak %s, trough %s, %s\n", e.Depth*100,
			e.PeakDate.Format("2006-01"), e.TroughDate.Format("2006-01"), back)
	}
}
