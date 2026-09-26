package metrics_test

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
)

// Several series meet on ONE calendar before any cross-statistic is read:
// marketdata.AlignSeries starts where all of them quote and forward-fills each
// across the others' sessions, and its Returns are the [asset][period] input
// of the matrices. The synthetic lines of threeFunds stand for world equities,
// an S&P 500 fund that mostly follows them, and gold, which does not and whose
// history starts later, so the common window starts with it.
func Example_oneCalendar() {
	a, err := marketdata.AlignSeries(threeFunds(), time.Time{}, time.Time{})
	if err != nil {
		panic(err)
	}
	r := a.Returns()
	corr, cov := metrics.CorrelationMatrix(r), metrics.Covariance(r)
	fmt.Printf("from %s: %s/%s %.2f, %s/%s %.2f\n", a.Dates[0].Format(time.DateOnly),
		a.IDs[0], a.IDs[1], corr[0][1], a.IDs[0], a.IDs[2], corr[0][2])
	fmt.Printf("%s volatility %.1f %%/yr\n", a.IDs[0], math.Sqrt(cov[0][0]*252)*100)

	// Partial flags a first year measured from the first quote; the last row
	// ends on the last quote, which its End says.
	for _, y := range metrics.CalendarReturns(a.Dates, a.Levels[0], 12) {
		fmt.Printf("year to %s %+5.1f %%, partial=%v\n", y.End.Format(time.DateOnly), y.Return*100, y.Partial)
	}
	if _, betas, ok := metrics.RollingBeta(a.Dates, a.Levels[1], a.Dates, a.Levels[0], 1); ok {
		fmt.Printf("one-year beta of %s on %s, last: %.2f\n", a.IDs[1], a.IDs[0], betas[len(betas)-1])
	}
	if v, ok := metrics.VaR(r[0], 0.95); ok {
		fmt.Printf("daily 95 %% VaR of %s: %.2f %%\n", a.IDs[0], v*100)
	}
	// Output:
	// from 2020-07-20: IWDA/VUAA 0.88, IWDA/IGLN -0.16
	// IWDA volatility 11.2 %/yr
	// year to 2020-12-31  +5.1 %, partial=true
	// year to 2021-12-31  +8.0 %, partial=false
	// year to 2022-12-30 +11.3 %, partial=false
	// year to 2023-04-14  +1.9 %, partial=false
	// one-year beta of VUAA on IWDA, last: 1.10
	// daily 95 % VaR of IWDA: 0.95 %
}

// threeFunds returns three deterministic weekday series named after IWDA
// (IE00B4L5Y983), VUAA (IE00BFMXXD54) and IGLN (IE00B4ND3602): a common swing
// the first two share, an own swing for the last two, and gold starting 200
// days after the others.
func threeFunds() []*marketdata.Series {
	var series []*marketdata.Series
	for k, id := range []string{"IWDA", "VUAA", "IGLN"} {
		var dates []time.Time
		var closes []float64
		level, session := 100.0, 0
		for i := range 1200 {
			d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}
			if session++; k == 2 && i < 200 {
				continue
			}
			common := 0.01 * math.Sin(float64(session)*0.9)
			own := 0.006 * math.Sin(float64(session)*(1.3+0.4*float64(k)))
			level *= 1 + [...]float64{0.0004 + common, 0.0005 + 1.1*common + own, 0.0002 - 0.2*common + 2*own}[k]
			dates, closes = append(dates, d), append(closes, level)
		}
		s, err := marketdata.NewSeries(id, dates, closes)
		if err != nil {
			panic(err)
		}
		series = append(series, s)
	}
	return series
}
