package metrics_test

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Compute derives every statistic from a series of dated values.
func ExampleCompute() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := make([]time.Time, 5)
	for i := range dates {
		dates[i] = start.AddDate(0, 0, i)
	}
	stats, err := metrics.Compute(dates, []float64{100, 110, 99, 104, 108})
	if err != nil {
		panic(err)
	}
	fmt.Printf("MaxDrawdown: %.1f %%\n", stats.MaxDrawdown*100)
	fmt.Printf("TTR: %d days (ongoing: %v)\n", stats.TTRDays, stats.TTROngoing)
	// Output:
	// MaxDrawdown: -10.0 %
	// TTR: 3 days (ongoing: true)
}

// CWARP scores whether overlaying an asset on a replacement portfolio (here
// equity beta) improves its risk-adjusted returns. An anti-correlated sleeve
// with positive carry scores above zero.
func ExampleCWARP() {
	equity := make([]float64, 250)
	diversifier := make([]float64, 250)
	for i := range equity {
		equity[i] = 0.001 + 0.006*math.Sin(float64(i)*0.3)
		if i >= 100 && i < 115 {
			equity[i] = -0.010 // a drawdown
		}
		diversifier[i] = -equity[i] + 0.0007 // hedge plus carry
	}
	score, ok := metrics.CWARP(diversifier, equity, metrics.CWARPParams{})
	fmt.Printf("improves the portfolio: %v\n", ok && score > 0)
	// Output:
	// improves the portfolio: true
}

// ReturnToMaxDrawdown is the Calmar-style ratio of annualized growth to the
// worst peak-to-trough loss, the return-to-drawdown building block CWARP and
// the optimizer reuse.
func ExampleReturnToMaxDrawdown() {
	returns := make([]float64, 250)
	for i := range returns {
		returns[i] = 0.002 // steady gains…
		if i >= 100 && i < 110 {
			returns[i] = -0.02 // …interrupted by a drawdown
		}
	}
	r, ok := metrics.ReturnToMaxDrawdown(returns, 0)
	fmt.Printf("defined: %v, positive: %v\n", ok, r > 0)
	// Output:
	// defined: true, positive: true
}

// Ulcer measures how painful the drawdowns were (depth and duration), and
// WorstRollingReturn the worst outcome over any window of the given length:
// the two underwater-robustness quantities the decumulation-minded optimizer
// objectives (min-ulcer, max-worst-5y) target.
func ExampleUlcer() {
	returns := make([]float64, 300)
	for i := range returns {
		returns[i] = 0.001
		if i >= 100 && i < 130 {
			returns[i] = -0.01 // a prolonged drawdown
		}
	}
	worst, _ := metrics.WorstRollingReturn(returns, 252)
	fmt.Printf("Ulcer > 0: %v, worst 1y return negative: %v\n",
		metrics.Ulcer(returns) > 0, worst < 0)
	// Output:
	// Ulcer > 0: true, worst 1y return negative: true
}

// Beta regresses a series' daily returns on a benchmark's, matching
// observations by date.
func ExampleBeta() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 40
	dates := make([]time.Time, n)
	bench := make([]float64, n)
	port := make([]float64, n)
	bench[0], port[0] = 100, 100
	for i := 1; i < n; i++ {
		dates[i-1] = start.AddDate(0, 0, i-1)
		r := 0.01 * float64(i%5-2)
		bench[i] = bench[i-1] * (1 + r)
		port[i] = port[i-1] * (1 + 2*r) // exactly twice the benchmark
	}
	dates[n-1] = start.AddDate(0, 0, n-1)
	beta, ok := metrics.Beta(dates, port, dates, bench)
	fmt.Printf("beta=%.1f ok=%v\n", beta, ok)
	// Output:
	// beta=2.0 ok=true
}

// TWR neutralizes external flows: a deposit is not performance. Here the
// market gains 10 % on day 2, then a 100 deposit lands on day 3: the
// money-agnostic return stays +10 %.
func ExampleTWR() {
	d0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := []time.Time{d0, d0.AddDate(0, 0, 1), d0.AddDate(0, 0, 2)}
	values := []float64{100, 110, 210}
	flows := []metrics.Flow{{Date: dates[2], Amount: 100}}
	twr, ok := metrics.TWR(dates, values, flows)
	fmt.Printf("ok=%v TWR=%.1f %%\n", ok, twr*100)
	// Output:
	// ok=true TWR=10.0 %
}

// IRR weighs each cash flow by its date: money invested early counts more
// than money added late. Flows are signed from the investor's standpoint
// (negative going in, positive coming out).
func ExampleIRR() {
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := []time.Time{d0, d0.AddDate(1, 0, 0)}
	flows := []float64{-1000, -1000} // initial capital, then one contribution
	irr, ok := metrics.IRR(dates, flows, d0.AddDate(2, 0, 0), 2200)
	fmt.Printf("ok=%v IRR=%.1f %%/yr\n", ok, irr*100)
	// Output:
	// ok=true IRR=6.5 %/yr
}

func ExampleTopK() {
	// The worst 20 % of a set of drawdowns, as a conditional-tail statistic
	// reads them: the largest values first.
	dds := []float64{0.12, 0.41, 0.07, 0.33, 0.19, 0.05, 0.28, 0.16, 0.22, 0.09}
	worst := metrics.TopK(dds, 2)
	fmt.Printf("worst two: %.2f %.2f\n", worst[0], worst[1])
	// Output:
	// worst two: 0.41 0.33
}

func ExampleQuantiles() {
	xs := []float64{5, 1, 4, 2, 3}
	q := metrics.Quantiles(xs, 0.05, 0.50, 0.95)
	fmt.Printf("p5=%.1f p50=%.1f p95=%.1f\n", q[0], q[1], q[2])
	// Output:
	// p5=1.2 p50=3.0 p95=4.8
}

// Corr is the Pearson correlation of two samples, here the daily returns of
// two assets on one calendar.
func ExampleCorr() {
	a := []float64{0.010, -0.020, 0.015, 0.003, -0.007}
	b := []float64{0.020, -0.010, 0.010, 0.000, -0.012}
	fmt.Printf("corr=%.2f\n", metrics.Corr(a, b))
	// Output:
	// corr=0.84
}

// CorrelationMatrix reads [asset][period] returns on one calendar, the shape
// marketdata.Aligned.Returns produces.
func ExampleCorrelationMatrix() {
	returns := [][]float64{
		{0.010, -0.020, 0.015, 0.003, -0.007}, // equity
		{0.020, -0.010, 0.010, 0.000, -0.012}, // more equity
		{-0.005, 0.010, 0.002, -0.004, 0.006}, // a hedge
	}
	for _, row := range metrics.CorrelationMatrix(returns) {
		fmt.Printf("%5.2f %5.2f %5.2f\n", row[0], row[1], row[2])
	}
	// Output:
	//  1.00  0.84 -0.77
	//  0.84  1.00 -0.77
	// -0.77 -0.77  1.00
}

// Covariance is per period: scale by 252 to annualize a daily one. The
// square root of the diagonal is then each asset's annualized volatility.
func ExampleCovariance() {
	returns := [][]float64{
		{0.010, -0.020, 0.015, 0.003, -0.007},
		{0.020, -0.010, 0.010, 0.000, -0.012},
	}
	cov := metrics.Covariance(returns)
	fmt.Printf("annualized vols: %.1f %% %.1f %%\n",
		math.Sqrt(cov[0][0]*252)*100, math.Sqrt(cov[1][1]*252)*100)
	fmt.Printf("annualized covariance: %.5f\n", cov[0][1]*252)
	// Output:
	// annualized vols: 22.2 % 21.5 %
	// annualized covariance: 0.03984
}

// CalendarReturns is the "annual returns" row: each calendar year's return,
// the first one flagged Partial because the series starts mid-year.
func ExampleCalendarReturns() {
	dates := []time.Time{
		time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 6, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 12, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC),
	}
	values := []float64{100, 108, 90, 81, 89.1}
	for _, y := range metrics.CalendarReturns(dates, values, 12) {
		fmt.Printf("%d: %+.1f %% (to %s, partial %v)\n",
			y.End.Year(), y.Return*100, y.End.Format(time.DateOnly), y.Partial)
	}
	// Output:
	// 2021: +8.0 % (to 2021-12-31, partial true)
	// 2022: -25.0 % (to 2022-12-30, partial false)
	// 2023: +10.0 % (to 2023-03-15, partial false)
}

// RollingBeta reads how a series' sensitivity to its benchmark evolves: here
// a sleeve that doubles its benchmark every day reads 2 in every window.
func ExampleRollingBeta() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 3 * 365
	dates := make([]time.Time, n)
	bench, sleeve := make([]float64, n), make([]float64, n)
	bench[0], sleeve[0] = 100, 100
	dates[0] = start
	for i := 1; i < n; i++ {
		dates[i] = start.AddDate(0, 0, i)
		r := 0.01 * math.Sin(float64(i))
		bench[i] = bench[i-1] * (1 + r)
		sleeve[i] = sleeve[i-1] * (1 + 2*r)
	}
	ends, betas, ok := metrics.RollingBeta(dates, sleeve, dates, bench, 1)
	fmt.Printf("ok=%v first window ends %s, beta %.2f; last %.2f\n",
		ok, ends[0].Format(time.DateOnly), betas[0], betas[len(betas)-1])
	// Output:
	// ok=true first window ends 2021-01-01, beta 2.00; last 2.00
}

// RollingCorr shows a diversifier's correlation drifting from one regime to
// the other: moving with the benchmark for two years, against it after.
func ExampleRollingCorr() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 4 * 365
	dates := make([]time.Time, n)
	bench, fund := make([]float64, n), make([]float64, n)
	bench[0], fund[0] = 100, 100
	dates[0] = start
	for i := 1; i < n; i++ {
		dates[i] = start.AddDate(0, 0, i)
		r := 0.01 * math.Sin(float64(i))
		sign := 1.0
		if i >= 2*365 {
			sign = -1
		}
		bench[i] = bench[i-1] * (1 + r)
		fund[i] = fund[i-1] * (1 + sign*r)
	}
	_, corrs, _ := metrics.RollingCorr(dates, fund, dates, bench, 1)
	fmt.Printf("first window %.2f, last window %.2f\n", corrs[0], corrs[len(corrs)-1])
	// Output:
	// first window 1.00, last window -1.00
}

// VaR reads a return sample's tail as a positive loss: at 95 %, the daily
// loss exceeded on one day in twenty.
func ExampleVaR() {
	daily := []float64{
		0.004, -0.012, 0.007, 0.001, -0.021, 0.009, -0.003, 0.012, -0.006, 0.002,
		-0.015, 0.005, 0.011, -0.008, 0.003, -0.030, 0.006, -0.001, 0.008, -0.004,
	}
	v, ok := metrics.VaR(daily, 0.95)
	fmt.Printf("ok=%v 1-day VaR 95 %% = %.3f %%\n", ok, v*100)
	// Output:
	// ok=true 1-day VaR 95 % = 2.145 %
}

// CVaR (expected shortfall) is the mean loss at or beyond VaR, always at
// least as large: the size of a bad day once one happens.
func ExampleCVaR() {
	daily := []float64{
		0.004, -0.012, 0.007, 0.001, -0.021, 0.009, -0.003, 0.012, -0.006, 0.002,
		-0.015, 0.005, 0.011, -0.008, 0.003, -0.030, 0.006, -0.001, 0.008, -0.004,
	}
	cv, ok := metrics.CVaR(daily, 0.95)
	fmt.Printf("ok=%v 1-day CVaR 95 %% = %.3f %%\n", ok, cv*100)
	// Output:
	// ok=true 1-day CVaR 95 % = 3.000 %
}
