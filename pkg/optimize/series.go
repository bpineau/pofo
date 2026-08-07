package optimize

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/metrics"
)

// fiveYearWindow is the rolling window MaxWorst5y measures its worst-case
// return over: five years of trading days (a date-free approximation of the
// report's calendar-year "Worst rolling 5y CAGR").
const fiveYearWindow = 5 * tradingDays

// solveSeries handles the objectives that depend on the whole return path
// (MaxSortino, ReturnToDrawdown) rather than only the mean and covariance. It
// maximizes the portfolio's own metric over the capped simplex with the shared
// multi-start solver; the weights are a good allocation, not a certified
// optimum, since these objectives are non-convex and non-smooth.
func solveSeries(returns [][]float64, spec Spec) (Result, error) {
	n := len(returns)
	t := len(returns[0])
	maxW := spec.MaxWeight
	if maxW <= 0 || maxW > 1 {
		maxW = 1
	}
	if float64(n)*maxW < 1-1e-9 {
		return Result{}, fmt.Errorf("max-weight too low: %d assets cannot sum to 100%% under a %.0f%% cap", n, maxW*100)
	}

	buf := make([]float64, t)
	var score func([]float64) (float64, bool)
	switch spec.Objective {
	case MaxSortino:
		score = func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			s := metrics.Sortino(buf, 0)
			return s, !math.IsNaN(s)
		}
	case ReturnToDrawdown:
		score = func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			return metrics.ReturnToMaxDrawdown(buf, 0)
		}
	case MinUlcer:
		score = func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			u := metrics.Ulcer(buf)
			return -u, !math.IsNaN(u) // minimize: maximize the negative
		}
	case MaxWorst5y:
		if t < fiveYearWindow {
			return Result{}, fmt.Errorf("max-worst-5y needs at least 5 years of common history, got %d trading days", t)
		}
		score = func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			return metrics.WorstRollingReturn(buf, fiveYearWindow)
		}
	default:
		return Result{}, fmt.Errorf("solveSeries: unsupported objective %q", spec.Objective)
	}

	w := maximizeSimplex(n, maxW, score)
	return seriesResult(w, returns), nil
}

// seriesResult packages the weights with their mean/covariance statistics (for
// display consistency with the other objectives) plus the achieved Sortino and
// return-to-max-drawdown measured on the realized portfolio series.
func seriesResult(w []float64, returns [][]float64) Result {
	mu, cov := meanCov(returns)
	r := stats(w, mu, cov)
	buf := make([]float64, len(returns[0]))
	blend(returns, w, buf)
	if s := metrics.Sortino(buf, 0); !math.IsNaN(s) {
		r.Sortino = s
	}
	if v, ok := metrics.ReturnToMaxDrawdown(buf, 0); ok {
		r.ReturnToMaxDD = v
	}
	if u := metrics.Ulcer(buf); !math.IsNaN(u) {
		r.Ulcer = u
	}
	if len(buf) >= fiveYearWindow {
		if w5, ok := metrics.WorstRollingReturn(buf, fiveYearWindow); ok {
			r.Worst5y = w5
		}
	}
	return r
}
