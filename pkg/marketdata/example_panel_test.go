package marketdata_test

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// An exploration across three series, the questions a notebook asks: blend
// two of them, score the blend, measure one leg against another, date the
// worst months and read what the third did in them.
func Example_explore() {
	// Monthly returns on the months all three share, labelled month-end.
	p, err := marketdata.NewPanel(marketdata.Monthly, threeLegs()...)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d months to %s\n", p.Len(), p.Ends[p.Len()-1].Format(time.DateOnly))

	// A 60/40 rebalanced every month, as a level, then its statistics.
	p, err = p.Mix("60/40", map[string]float64{"IWDA": 0.6, "AGGH": 0.4})
	if err != nil {
		panic(err)
	}
	blend, err := p.Series("60/40")
	if err != nil {
		panic(err)
	}
	st, err := blend.Stats()
	if err != nil {
		panic(err)
	}
	fmt.Printf("60/40: CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)

	// Gold regressed on equities: beta, t-statistic, annualized alpha.
	reg, err := metrics.Regress(p.Col("IGLN"), p.Col("IWDA"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("IGLN on IWDA: beta %.2f (t %.1f), alpha %+.1f %%/yr, R2 %.2f\n",
		reg.Betas[0].Value, reg.Betas[0].T, reg.AnnualAlpha(p.PeriodsPerYear())*100, reg.R2)

	// The three worst equity months, dated.
	eq := p.Col("IWDA")
	for _, t := range metrics.LowestK(eq, 3) {
		fmt.Printf("%s IWDA %+.1f %%\n", p.Ends[t].Format("2006-01"), eq[t]*100)
	}

	// What gold did in the worst tenth of equity months.
	worst := p.Pick(metrics.LowestK(eq, p.Len()/10))
	fmt.Printf("worst %d months: IWDA %+.1f %%, IGLN %+.1f %% on average\n",
		worst.Len(), metrics.Mean(worst.Col("IWDA"))*100, metrics.Mean(worst.Col("IGLN"))*100)
	// Output:
	// 119 months to 2024-12-31
	// 60/40: CAGR 7.4 %, volatility 8.2 %, max drawdown -7.1 %
	// IGLN on IWDA: beta -0.29 (t -4.7), alpha +6.3 %/yr, R2 0.16
	// 2020-07 IWDA -7.6 %
	// 2023-08 IWDA -6.6 %
	// 2016-10 IWDA -6.5 %
	// worst 11 months: IWDA -6.3 %, IGLN +4.8 % on average
}

// A daily panel keeps the sessions every series quotes: the holiday one
// market kept (here the 3rd, which only IWDA quotes) is not a period, and
// the return over it spans both sessions for every column.
func ExampleNewPanel_daily() {
	day := func(d int) time.Time { return time.Date(2024, 1, d, 0, 0, 0, 0, time.UTC) }
	iwda, _ := marketdata.NewSeries("IWDA", []time.Time{day(2), day(3), day(4)}, []float64{100, 102, 101})
	vuaa, _ := marketdata.NewSeries("VUAA", []time.Time{day(2), day(4)}, []float64{50, 51})
	p, err := marketdata.NewPanel(marketdata.Daily, iwda, vuaa)
	if err != nil {
		panic(err)
	}
	for t := range p.Len() {
		fmt.Printf("%s to %s: IWDA %+.1f %%, VUAA %+.1f %%\n", p.Starts[t].Format(time.DateOnly), p.Ends[t].Format(time.DateOnly),
			p.R[0][t]*100, p.R[1][t]*100)
	}
	// Output:
	// 2024-01-02 to 2024-01-04: IWDA +1.0 %, VUAA +2.0 %
}

// LessFee deducts a yearly charge, a FRACTION (0.0085 for 0.85 %/yr), from
// an index to read what a fund tracking it at that cost returned.
func ExampleSeries_LessFee() {
	start := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	index, _ := marketdata.NewSeries("MSCIWORLD", []time.Time{start, start.AddDate(8, 0, 0)}, []float64{100, 200})
	fund := index.LessFee(0.0085)
	fmt.Printf("index %.1f, fund %.1f\n", index.Last().Close, fund.Last().Close)
	// Output:
	// index 200.0, fund 186.8
}

// Change reads a named episode: calendar 2008 runs from the last close of
// 2007 to the last close of 2008.
func ExampleSeries_Change() {
	d := func(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }
	s, _ := marketdata.NewSeries("SP500", []time.Time{d(2007, 12, 31), d(2008, 10, 10), d(2008, 12, 31), d(2009, 3, 9)},
		[]float64{100, 65, 63, 52})
	c, err := s.Change(d(2008, 1, 1), d(2008, 12, 31))
	if err != nil {
		panic(err)
	}
	fmt.Printf("2008: %+.0f %%\n", c*100)
	// Output:
	// 2008: -37 %
}

// threeLegs is three synthetic daily series over ten years of weekdays, an
// equity index, a bond fund and gold, whose slow components make monthly
// returns worth studying: gold leans against equities' slow swings.
func threeLegs() []*marketdata.Series {
	var out []*marketdata.Series
	for k, id := range []string{"IWDA", "AGGH", "IGLN"} { // IE00B4L5Y983, IE00BDBRDM35, IE00B4ND3602
		var dates []time.Time
		var closes []float64
		level, session := 100.0, 0
		for d := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC); d.Year() < 2025; d = d.AddDate(0, 0, 1) {
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}
			session++
			i := float64(session)
			slow := 0.003*math.Sin(i/9.7) + 0.0015*math.Sin(i/41)
			level *= 1 + [...]float64{
				0.0004 + 0.009*math.Sin(0.9*i) + slow,
				0.0001 + 0.003*math.Sin(1.7*i) - 0.2*slow + 0.001*math.Sin(i/13),
				0.0001 + 0.008*math.Sin(2.3*i+1) - 0.3*slow + 0.002*math.Sin(i/17),
			}[k]
			dates, closes = append(dates, d), append(closes, level)
		}
		s, err := marketdata.NewSeries(id, dates, closes)
		if err != nil {
			panic(err)
		}
		out = append(out, s)
	}
	return out
}
