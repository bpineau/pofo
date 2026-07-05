package optimize

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/metrics"
)

// SolveCWARP returns the long-only capped-simplex weights that maximize the
// portfolio's CWARP against a replacement return series (the benchmark),
// together with the achieved score. returns holds each asset's aligned daily
// simple returns; replacement is the benchmark's returns on the same dates
// (same length).
//
// CWARP depends on the combined series' drawdown, so the objective is neither
// convex nor smooth. Like maxSharpe this is therefore a multi-start heuristic:
// projected descent with a numerical gradient, run from several deterministic
// starts (equal weights, then each single asset), keeping the best score. The
// weights are a good allocation, not a certified global optimum.
func SolveCWARP(returns [][]float64, replacement []float64, spec Spec) (Result, error) {
	n := len(returns)
	if n == 0 {
		return Result{}, fmt.Errorf("no assets to optimize")
	}
	t := len(returns[0])
	if t < 2 {
		return Result{}, fmt.Errorf("need at least 2 observations, got %d", t)
	}
	for i, r := range returns {
		if len(r) != t {
			return Result{}, fmt.Errorf("asset %d has %d observations, expected %d", i, len(r), t)
		}
	}
	if len(replacement) != t {
		return Result{}, fmt.Errorf("replacement has %d observations, expected %d", len(replacement), t)
	}
	maxW := spec.MaxWeight
	if maxW <= 0 || maxW > 1 {
		maxW = 1
	}
	if float64(n)*maxW < 1-1e-9 {
		return Result{}, fmt.Errorf("max-weight too low: %d assets cannot sum to 100%% under a %.0f%% cap", n, maxW*100)
	}

	buf := make([]float64, t)
	score := func(w []float64) (float64, bool) {
		blend(returns, w, buf)
		return metrics.CWARP(buf, replacement, metrics.CWARPParams{})
	}
	w := maximizeSimplex(n, maxW, score)

	mu, cov := meanCov(returns)
	r := stats(w, mu, cov)
	if c, ok := score(w); ok {
		r.CWARP = c
	}
	return r, nil
}
