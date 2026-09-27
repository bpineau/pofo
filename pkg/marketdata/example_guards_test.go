package marketdata_test

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// FindGaps reads a series as data: a monthly file that skips June holds one
// step of two months, which every return computed across it would read as
// one month's move.
func ExampleFindGaps() {
	var dates []time.Time
	var closes []float64
	for m := time.January; m <= time.December; m++ {
		if m == time.June {
			continue
		}
		dates = append(dates, time.Date(2023, m+1, 0, 0, 0, 0, 0, time.UTC)) // the month's last day
		closes = append(closes, 100+float64(m))
	}
	s, _ := marketdata.NewSeries("MONTHLY", dates, closes)
	for _, g := range marketdata.FindGaps(s) {
		fmt.Printf("%s to %s: %.0f days, the pace allows %.1f\n",
			g.From.Format(time.DateOnly), g.To.Format(time.DateOnly), g.Days, g.Limit)
	}
	// Output:
	// 2023-05-31 to 2023-07-31: 61 days, the pace allows 46.5
}

// FindSpikes names the print no instrument made: a close 20 % above its
// neighbours, which agree with each other, in a series that otherwise moves
// by a fraction of a percent a day.
func ExampleFindSpikes() {
	var dates []time.Time
	var closes []float64
	for i := range 80 {
		dates = append(dates, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
		c := 100 * (1 + 0.004*math.Sin(float64(i)))
		if i == 40 {
			c *= 1.20
		}
		closes = append(closes, c)
	}
	s, _ := marketdata.NewSeries("FUND", dates, closes)
	for _, sp := range marketdata.FindSpikes(s) {
		fmt.Printf("%s: %+.1f %% then %+.1f %%, local sigma %.2f %%\n",
			sp.Date.Format(time.DateOnly), sp.In*100, sp.Out*100, sp.Sigma*100)
	}
	// Output:
	// 2024-02-10: +19.9 % then -17.0 %, local sigma 0.27 %
}
