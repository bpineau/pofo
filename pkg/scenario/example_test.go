package scenario_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

// A parametric source draws i.i.d. real returns from a Student-t scaled to the
// requested mean and standard deviation: here 30 annual returns per path, at
// 3.5 % a year and 12 % of volatility, with the fat tails of six degrees of
// freedom. Over many draws the sample recovers the parameters.
func ExampleParametricSource() {
	src := scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 30}
	rng := rand.New(rand.NewPCG(1, 2))

	var n, sum, sumSq, worst float64
	for range 20_000 {
		for _, r := range src.Draw(rng) {
			n++
			sum += r
			sumSq += r * r
			worst = math.Min(worst, r)
		}
	}
	mean := sum / n
	fmt.Printf("%d periods per path, mean %.1f %%, volatility %.0f %%\n", src.Len(), mean*100, math.Sqrt(sumSq/n-mean*mean)*100)
	fmt.Printf("worst year below -40 %%: %v\n", worst < -0.40)
	// Output:
	// 30 periods per path, mean 3.5 %, volatility 12 %
	// worst year below -40 %: true
}

// A block bootstrap resamples a real history on the time axis: every drawn
// return is one the weighted portfolio really earned, and within a block they
// follow each other as they did, so a crash and its recovery stay together.
// The panel here is two assets over eight years, held 60/40.
func ExampleBlockBootstrap() {
	panel := scenario.Panel{
		Returns: [][]float64{
			{0.10, -0.20, 0.25, 0.05, 0.15, -0.05, 0.08, 0.12}, // equities
			{0.02, 0.06, 0.01, 0.03, 0.00, 0.04, 0.02, 0.01},   // bonds
		},
		Weights: []float64{0.6, 0.4},
	}
	src := scenario.BlockBootstrap{Panel: panel, BlockLen: 3, Periods: 9}
	rng := rand.New(rand.NewPCG(1, 2))

	fmt.Println("history", pct(panel.Combine(nil)))
	fmt.Println("drawn  ", pct(src.Draw(rng)))
	// Output:
	// history +7 -10 +15 +4 +9 -1 +6 +8
	// drawn   +7 -10 +15 +9 -1 +6 +7 -10 +15
}

// Deflate turns nominal prices into the REAL period returns every source here
// speaks: a year at +5 % nominal under 2 % inflation is a real +2.9 %, and a
// year at -5 % under 8 % a real -12 %.
func ExampleDeflate() {
	day := func(y int) time.Time { return time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC) }
	prices := []marketdata.Point{{Date: day(2020), Close: 100}, {Date: day(2021), Close: 105}, {Date: day(2022), Close: 99.75}}
	hicp := []marketdata.Point{{Date: day(2020), Close: 100}, {Date: day(2021), Close: 102}, {Date: day(2022), Close: 110.16}}
	for _, r := range scenario.Deflate(prices, hicp) {
		fmt.Printf("%+.1f %%\n", r*100)
	}
	// Output:
	// +2.9 %
	// -12.0 %
}

// pct prints a return path as whole percents.
func pct(s scenario.Sequence) string {
	out := ""
	for i, r := range s {
		if i > 0 {
			out += " "
		}
		out += fmt.Sprintf("%+.0f", r*100)
	}
	return out
}
