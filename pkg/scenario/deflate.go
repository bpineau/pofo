package scenario

import (
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Deflate converts a nominal price series into real period-over-period
// returns using an HICP (or any price-level) series: for consecutive prices
// the nominal ratio is divided by the inflation ratio over the same dates.
// The HICP level used for a date is the last point at or before it.
func Deflate(prices, hicp []marketdata.Point) Sequence {
	if len(prices) < 2 || len(hicp) == 0 {
		return nil
	}
	out := make(Sequence, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		p0, p1 := prices[i-1].Close, prices[i].Close
		c0 := hicpAt(hicp, prices[i-1].Date)
		c1 := hicpAt(hicp, prices[i].Date)
		if p0 <= 0 || c0 <= 0 || c1 <= 0 {
			out = append(out, 0)
			continue
		}
		out = append(out, (p1/p0)/(c1/c0)-1)
	}
	return out
}

// hicpAt returns the HICP level at or before t (the first level if t is
// before the series starts).
func hicpAt(hicp []marketdata.Point, t time.Time) float64 {
	level := hicp[0].Close
	for _, p := range hicp {
		if p.Date.After(t) {
			break
		}
		level = p.Close
	}
	return level
}
