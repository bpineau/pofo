package analyze_test

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// Pair measures a reconstruction against the real fund it rebuilds, on the
// window both cover: the level gap with its standard error, the monthly
// path, and the months where the two part, dated. Against live data the two
// would be fetched (client.FetchExtended(ctx, "DBMFSIM", ...) for the
// backcast, the fund's own quotes for the reference) or read from files
// (marketdata.ReadCSV); here they are synthetic.
func ExamplePair() {
	backcast, fund := backcastAndFund()
	st, err := analyze.Pair(backcast, fund, analyze.PairOptions{Divergences: 2})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s to %s: CAGR gap %+.2f pt/yr (standard error %.2f)\n",
		st.Start.Format(time.DateOnly), st.End.Format(time.DateOnly), st.CAGRGap*100, st.GapSE*100)
	m := st.Monthly
	fmt.Printf("monthly: corr %.3f, tracking error %.2f %%, beta %.2f\n", m.Corr, m.TrackingError*100, m.Beta)
	for _, d := range m.Divergences {
		fmt.Printf("%s: %+.2f %% against %+.2f %%\n", d.End.Format("2006-01"), d.A*100, d.B*100)
	}
	// Output:
	// 2018-01-01 to 2023-12-29: CAGR gap +0.30 pt/yr (standard error 0.15)
	// monthly: corr 1.000, tracking error 0.38 %, beta 1.00
	// 2018-08: +5.62 % against +5.39 %
	// 2019-05: +0.48 % against +0.26 %
}

// WriteText prints the whole study for a terminal or an agent to read. Here
// an old and a new version of one monthly file: the new one is rebased ten
// times higher, which the identity line forgives, and revised from March
// 2024 on, which it dates.
func ExamplePairStudy_WriteText() {
	var dates []time.Time
	var old, revised []float64
	for m := range 37 {
		dates = append(dates, time.Date(2022, time.Month(1+m), 0, 0, 0, 0, 0, time.UTC)) // month-ends from 2021-12-31
		level := 100 * math.Exp(0.005*float64(m)+0.03*math.Sin(float64(m)/2))
		old = append(old, level)
		if dates[m].After(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)) {
			level *= 1.004
		}
		revised = append(revised, 10*level)
	}
	before, _ := marketdata.NewSeries("X@HEAD~1", dates, old)
	after, _ := marketdata.NewSeries("X", dates, revised)
	st, err := analyze.Pair(after, before, analyze.PairOptions{Divergences: 1})
	if err != nil {
		panic(err)
	}
	if err := st.WriteText(os.Stdout); err != nil {
		panic(err)
	}
	// Output:
	// A  X         2021-12-31 to 2024-12-31  monthly (12/yr)
	// B  X@HEAD~1  2021-12-31 to 2024-12-31  monthly (12/yr)
	//
	// window      2021-12-31 to 2024-12-31 (3.0 years)
	// dates       37 shared, 0 only in A, 0 only in B
	// identity    A/B 10 at the start, first moves on 2024-03-31
	// CAGR        A +5.53 %, B +5.39 %, gap +0.14 pt/yr (standard error 0.14)
	// level       A ends +0.40 % from B, both rebased at the start
	// volatility  A 3.70 %, B 3.65 %, each at its own cadence
	// drawdown    A -3.14 %, B -3.14 %
	//
	// returns  periods  per year  corr   vol A   vol B   vol ratio  tracking  beta   alpha/yr
	// monthly  36       12        0.998  3.70 %  3.65 %  1.015      0.23 %    1.013  +0.07 %
	//
	// largest monthly divergences  A        B        A-B       excess
	// 2024-02-29 to 2024-03-31     +2.07 %  +1.66 %  +0.41 pt  0.41 %
	//
	// year  A        B        A-B
	// 2022  +5.30 %  +5.30 %  +0.00 pt
	// 2023  +5.37 %  +5.37 %  +0.00 pt
	// 2024  +5.92 %  +5.50 %  +0.42 pt
}

// backcastAndFund is a synthetic fund quoted every weekday of 2018-2023 and
// a reconstruction of it that runs 0.3 point a year hot with a little noise
// of its own, the pair a backcast audit reads.
func backcastAndFund() (backcast, fund *marketdata.Series) {
	var dates []time.Time
	var f, b []float64
	level, noise := 100.0, 0.0
	for d := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC); d.Year() < 2024; d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		i := float64(len(dates))
		level *= 1 + 0.0002 + 0.008*math.Sin(1.1*i) + 0.003*math.Sin(i/19)
		noise = 0.9*noise + 0.002*math.Sin(2.9*i)
		years := d.Sub(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24 / 365.25
		dates = append(dates, d)
		f = append(f, level)
		b = append(b, level*math.Pow(1.003, years)*(1+noise))
	}
	fund, err := marketdata.NewSeries("FUND", dates, f)
	if err != nil {
		panic(err)
	}
	backcast, err = marketdata.NewSeries("FUNDSIM", dates, b)
	if err != nil {
		panic(err)
	}
	return backcast, fund
}
