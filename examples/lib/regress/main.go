// Regress fits one bundled series on others: the monthly returns of gold on
// those of US equities and long US Treasuries, with the standard error and
// t-statistic of every coefficient.
//
//	go run ./examples/lib/regress
//
// A daily series (the gold fix) and two monthly ones meet on the calendar
// months all three share: NewPanel keeps each series' last close of every
// month, labelled by the month-end, so the three columns line up exactly.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

func main() {
	log.SetFlags(0)
	var series []*marketdata.Series
	for _, id := range []string{"XAUUSD-LBMA", "SP500-USD", "TREASURY-LONG-USD"} {
		s, err := marketdata.Bundled(id)
		if err != nil {
			log.Fatal(err)
		}
		series = append(series, s)
	}
	p, err := marketdata.NewPanel(marketdata.Monthly, series...)
	if err != nil {
		log.Fatal(err)
	}
	// The gold price floats freely from 1971; before, it was fixed.
	p = p.Between(time.Date(1972, 1, 1, 0, 0, 0, 0, time.UTC), time.Time{})

	gold, err := p.Col("XAUUSD-LBMA")
	if err != nil {
		log.Fatal(err)
	}
	equity, err := p.Col("SP500-USD")
	if err != nil {
		log.Fatal(err)
	}
	bonds, err := p.Col("TREASURY-LONG-USD")
	if err != nil {
		log.Fatal(err)
	}

	// Ordinary least squares with an intercept. Everything in the
	// Regression is PER PERIOD (here per month); AnnualAlpha and
	// AnnualResidualVol annualize at the panel's cadence.
	reg, err := metrics.Regress(gold, equity, bonds)
	if err != nil {
		log.Fatal(err)
	}
	ppy := p.PeriodsPerYear()
	fmt.Printf("gold on equities and long Treasuries, %d months from %s\n\n",
		p.Len(), p.Ends[0].Format("2006-01"))
	for i, name := range []string{"equities", "long Treasuries"} {
		b := reg.Betas[i]
		fmt.Printf("beta on %-16s %+.2f (standard error %.2f, t %+.1f)\n", name, b.Value, b.SE, b.T)
	}
	fmt.Printf("alpha %+.1f %%/yr (t %+.1f), R2 %.3f, residual volatility %.1f %%/yr\n",
		reg.AnnualAlpha(ppy)*100, reg.Alpha.T, reg.R2, reg.AnnualResidualVol(ppy)*100)
	fmt.Println("\n|t| under about 2: the coefficient is not distinguishable from zero.")
}
