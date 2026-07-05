package optimize

import (
	"fmt"
	"math"

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

	port := make([]float64, t)
	score := func(w []float64) (float64, bool) {
		for k := 0; k < t; k++ {
			s := 0.0
			for i := 0; i < n; i++ {
				s += w[i] * returns[i][k]
			}
			port[k] = s
		}
		return metrics.CWARP(port, replacement, metrics.CWARPParams{})
	}
	// Minimize the negative score. An undefined score (a region where the
	// blend has no drawdown or a non-positive replacement ratio) is a large
	// penalty, so the search steers away from it.
	neg := func(w []float64) float64 {
		if c, ok := score(w); ok {
			return -c
		}
		return 1e6
	}
	// Forward-difference gradient: projected descent only needs a descent
	// direction, and the projection restores the simplex after each step.
	grad := func(w []float64) []float64 {
		const h = 1e-5
		base := neg(w)
		g := make([]float64, n)
		for i := range w {
			wp := append([]float64(nil), w...)
			wp[i] += h
			g[i] = (neg(wp) - base) / h
		}
		return g
	}

	result := func(w []float64) Result {
		mu, cov := meanCov(returns)
		r := stats(w, mu, cov)
		if c, ok := score(w); ok {
			r.CWARP = c
		}
		return r
	}
	if n == 1 {
		return result([]float64{1}), nil
	}

	starts := [][]float64{equalStart(n, maxW)}
	for i := 0; i < n; i++ {
		s := make([]float64, n)
		s[i] = 1
		starts = append(starts, projectCappedSimplex(s, maxW))
	}
	var best []float64
	bestVal := math.Inf(1)
	for _, s := range starts {
		w := minimizeSimplex(neg, grad, maxW, s)
		if v := neg(w); v < bestVal {
			bestVal, best = v, w
		}
	}
	return result(best), nil
}
