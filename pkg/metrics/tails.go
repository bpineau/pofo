package metrics

import "math"

// VaR is the historical value at risk of a return sample at confidence p: the
// loss exceeded in a fraction 1-p of the periods, read off the sample's own
// (1-p)-quantile (Quantiles, linear interpolation) and returned as a POSITIVE
// fraction (0.02 = a 2 % loss). p = 0.95 is the loss exceeded in 5 % of
// periods. returns are per-period simple returns as fractions, read as they
// are: daily returns give a one-day figure, monthly ones a one-month figure,
// and nothing is annualized or scaled by a square root of time. The result is
// negative when even that quantile is a gain. ok is false on an empty sample
// or a p outside (0, 1).
func VaR(returns []float64, p float64) (float64, bool) {
	if len(returns) == 0 || !(p > 0 && p < 1) {
		return 0, false
	}
	return -Quantiles(returns, 1-p)[0], true
}

// CVaR is the historical conditional value at risk (expected shortfall) of a
// return sample at confidence p: the mean of the losses at or beyond VaR at
// the same p, as a POSITIVE fraction, so CVaR >= VaR always. Units and
// reading are VaR's: per-period returns as fractions in, a per-period loss
// out. ok is false on an empty sample or a p outside (0, 1).
func CVaR(returns []float64, p float64) (float64, bool) {
	v, ok := VaR(returns, p)
	if !ok {
		return 0, false
	}
	var sum float64
	var n int
	for _, r := range returns {
		if -r >= v {
			sum += -r
			n++
		}
	}
	if n == 0 { // only a NaN in the sample gets here: its minimum is at or below any quantile
		return math.NaN(), false
	}
	return sum / float64(n), true
}
