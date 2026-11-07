package optimize

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Performance measures a fixed allocation on one return panel. These are
// realized path statistics, including for a BlackLitterman fit whose Result
// contains modelled expected returns. All returns and risks are fractions;
// MaxDrawdown is negative. Sharpe uses the measurement's risk-free rate.
type Performance struct {
	// CAGR is annualized geometric return.
	CAGR float64
	// Volatility is annualized sample standard deviation.
	Volatility float64
	// MaxDrawdown is deepest peak-to-trough loss, a negative fraction.
	MaxDrawdown float64
	// Sharpe is annualized arithmetic excess return over Volatility.
	Sharpe float64
}

// Evaluate measures a long-only allocation on aligned [asset][period] simple
// returns at periodsPerYear. The weights must be finite nonnegative fractions
// summing to 1, in the panel's asset order. As with Solve, at least two
// observations are required. The allocation is rebalanced every period, with
// no fees or flows. Undefined Sharpe (zero variance) is reported as zero.
func Evaluate(returns [][]float64, weights []float64, periodsPerYear float64) (Performance, error) {
	return EvaluateWithRiskFree(returns, weights, periodsPerYear, 0)
}

// EvaluateWithRiskFree measures a fixed allocation as Evaluate does, using
// riskFree (a finite annual arithmetic fraction) for its Sharpe ratio.
// Negative rates are valid. CAGR, volatility and drawdown remain nominal.
func EvaluateWithRiskFree(returns [][]float64, weights []float64, periodsPerYear, riskFree float64) (Performance, error) {
	if !finite(riskFree) {
		return Performance{}, fmt.Errorf("risk-free rate must be finite")
	}
	if err := checkPanel(returns, periodsPerYear); err != nil {
		return Performance{}, err
	}
	if err := checkWeights(weights, len(returns), false); err != nil {
		return Performance{}, err
	}
	buf := make([]float64, len(returns[0]))
	blend(returns, weights, buf)
	p := Performance{CAGR: compound(buf, periodsPerYear), Volatility: metrics.Volatility(buf, periodsPerYear), MaxDrawdown: pathMaxDrawdown(buf)}
	if !finite(p.CAGR) || !finite(p.Volatility) || !finite(p.MaxDrawdown) {
		return Performance{}, fmt.Errorf("allocation path statistics overflow")
	}
	if p.Volatility > 0 {
		for _, r := range buf {
			p.Sharpe += r / float64(len(buf)) * periodsPerYear / p.Volatility
		}
		p.Sharpe -= riskFree / p.Volatility
	}
	return p, nil
}

// FitEvaluation keeps the training solve distinct from the measured training
// and held-out results. Fit.Feasible describes the training limits only;
// held-out performance is never a feasibility guarantee.
type FitEvaluation struct {
	// Fit contains weights and optimizer figures fitted on training only.
	Fit Result
	// InSample measures the fitted weights on the training panel.
	InSample Performance
	// OutOfSample measures those same weights on the separate test panel.
	OutOfSample Performance
}

// FitEvaluate fits spec on training only, then measures those fixed weights
// separately on training and test. Both panels must use the same asset order,
// currency and periodsPerYear; their observation counts may differ. The
// caller must select disjoint date windows (for example with marketdata's
// Panel.Between): this date-free package cannot detect overlap or reordered
// assets. Spec.Train must be cleared after that selection. CWARP requires a
// replacement panel and is not supported by this helper.
// Spec.RiskFree applies to the fit and both measured windows.
func FitEvaluate(training, test [][]float64, periodsPerYear float64, spec Spec) (FitEvaluation, error) {
	if err := checkPanel(test, periodsPerYear); err != nil {
		return FitEvaluation{}, fmt.Errorf("test: %w", err)
	}
	if len(training) != len(test) {
		return FitEvaluation{}, fmt.Errorf("training has %d assets, test has %d; same asset order required", len(training), len(test))
	}
	fit, err := Solve(training, periodsPerYear, spec)
	if err != nil {
		return FitEvaluation{}, fmt.Errorf("training: %w", err)
	}
	in, err := EvaluateWithRiskFree(training, fit.Weights, periodsPerYear, spec.RiskFree)
	if err != nil {
		return FitEvaluation{}, fmt.Errorf("training: %w", err)
	}
	out, err := EvaluateWithRiskFree(test, fit.Weights, periodsPerYear, spec.RiskFree)
	if err != nil {
		return FitEvaluation{}, fmt.Errorf("test: %w", err)
	}
	return FitEvaluation{Fit: fit, InSample: in, OutOfSample: out}, nil
}

// FrontierPoint retains the requested risk cap, the solver's weights and
// feasibility verdict, and measured performance. An infeasible point is
// retained so callers can show which requested budgets could not be met.
type FrontierPoint struct {
	// MaxVolatility is the requested annualized cap as a fraction.
	MaxVolatility float64
	// Result retains fitted weights and the feasibility verdict.
	Result Result
	// Measured contains realized performance on the fitted return panel.
	Measured Performance
}

// Frontier solves MaxReturn once for each requested positive annualized
// volatility cap (fractions), retaining points in the caller's grid order.
// It reuses Solve, including its bounds, return/drawdown limits and heuristic
// search: these are candidate frontier points, not certified global optima.
// Always check each point's Result.Feasible.
//
// spec.Objective must be MaxReturn and spec.Limits.MaxVolatility must be zero:
// the explicit grid supplies that limit. Other spec fields retain their Solve
// meaning. An empty grid, invalid cap or incompatible spec is an error.
func Frontier(returns [][]float64, periodsPerYear float64, spec Spec, volatilityCaps []float64) ([]FrontierPoint, error) {
	if spec.Objective != MaxReturn {
		return nil, fmt.Errorf("frontier requires max-return objective")
	}
	if spec.Limits.MaxVolatility != 0 {
		return nil, fmt.Errorf("frontier volatility grid requires Spec.Limits.MaxVolatility to be unset")
	}
	if len(volatilityCaps) == 0 {
		return nil, fmt.Errorf("frontier needs at least one volatility cap")
	}
	if err := validateSpec(spec, len(returns), false); err != nil {
		return nil, err
	}
	if err := checkPanel(returns, periodsPerYear); err != nil {
		return nil, err
	}
	for i, cap := range volatilityCaps {
		if !finite(cap) || cap <= 0 {
			return nil, fmt.Errorf("frontier cap %d must be finite and positive", i)
		}
	}
	points := make([]FrontierPoint, len(volatilityCaps))
	for i, cap := range volatilityCaps {
		sp := spec
		sp.Limits.MaxVolatility = cap
		fit, err := Solve(returns, periodsPerYear, sp)
		if err != nil {
			return nil, fmt.Errorf("frontier cap %g: %w", cap, err)
		}
		measured, err := Evaluate(returns, fit.Weights, periodsPerYear)
		if err != nil {
			return nil, fmt.Errorf("frontier cap %g: %w", cap, err)
		}
		points[i] = FrontierPoint{MaxVolatility: cap, Result: fit, Measured: measured}
	}
	return points, nil
}
