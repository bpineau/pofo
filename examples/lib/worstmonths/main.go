// Worstmonths dates the worst months of US equities and reads what long
// Treasuries, gold and managed futures did in exactly those months: the
// conditional statistic behind "does this hedge work when it matters".
//
//	go run ./examples/lib/worstmonths [-k 10]
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	k := flag.Int("k", 10, "how many of the worst equity months to list")
	flag.Parse()

	ids := []string{"SP500-USD", "TREASURY-LONG-USD", "XAUUSD-LBMA", "TREND-NET-USD"}
	var series []*marketdata.Series
	for _, id := range ids {
		s, err := marketdata.Bundled(id)
		if err != nil {
			log.Fatal(err)
		}
		series = append(series, s)
	}
	// The months all four share: the managed-futures composite starts in
	// 1987, so the panel does too.
	p, err := marketdata.NewPanel(marketdata.Monthly, series...)
	if err != nil {
		log.Fatal(err)
	}
	cols := make([][]float64, len(ids))
	for i, id := range ids {
		if cols[i], err = p.Col(id); err != nil {
			log.Fatal(err)
		}
	}
	equity := cols[0]
	fmt.Printf("%d months from %s to %s\n\n", p.Len(), p.Ends[0].Format("2006-01"), p.Ends[p.Len()-1].Format("2006-01"))

	// LowestK returns POSITIONS, so the worst months come with their dates
	// and every other column can be read on the same rows.
	fmt.Printf("%-8s %9s %9s %9s %9s\n", "month", "equities", "long UST", "gold", "trend")
	for _, t := range metrics.LowestK(equity, *k) {
		fmt.Printf("%-8s", p.Ends[t].Format("2006-01"))
		for i := range ids {
			fmt.Printf(" %+8.1f%%", cols[i][t]*100)
		}
		fmt.Println()
	}

	// Pick keeps a sample of periods: here the worst tenth of equity
	// months, against every month.
	worst, err := p.Pick(metrics.LowestK(equity, p.Len()/10))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nmean monthly return   %9s %9s %9s %9s\n", "equities", "long UST", "gold", "trend")
	for _, row := range []struct {
		name  string
		panel *marketdata.Panel
	}{{"every month", p}, {"worst tenth", worst}} {
		fmt.Printf("%-21s", row.name)
		for _, id := range ids {
			col, err := row.panel.Col(id)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf(" %+8.2f%%", metrics.Mean(col)*100)
		}
		fmt.Println()
	}
}
