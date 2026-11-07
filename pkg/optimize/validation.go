package optimize

import (
	"fmt"
	"math"
)

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// validateSpec is shared by both public solvers. Parsing is a convenience,
// never a prerequisite for enforcing the solver's contract.
func validateSpec(s Spec, n int, cwarp bool) error {
	if !finite(s.RiskFree) {
		return fmt.Errorf("risk-free rate must be finite")
	}
	if s.RiskFree != 0 && s.Objective != MaxSharpe && s.Objective != MaxSortino {
		return fmt.Errorf("risk-free rate applies only to max-sharpe and max-sortino")
	}
	switch s.Objective {
	case MaxSharpe, MinVolatility, RiskParity, MaxSortino, ReturnToDrawdown, MinUlcer, MaxWorst5y, MaxReturn, BlackLitterman:
		if cwarp {
			return fmt.Errorf("SolveCWARP requires objective cwarp")
		}
	case CWARP:
		if !cwarp {
			return fmt.Errorf("cwarp requires SolveCWARP and replacement returns")
		}
	default:
		return fmt.Errorf("unknown objective %q", s.Objective)
	}
	if !s.Train.IsZero() {
		return fmt.Errorf("train: select the training dates in the caller, then clear Spec.Train before Solve; use FitEvaluate for explicit training/test panels")
	}
	for _, v := range []struct {
		name      string
		value, hi float64
	}{
		{"max-weight", s.MaxWeight, 1}, {"min-weight", s.MinWeight, 1},
		{"max-volatility", s.Limits.MaxVolatility, math.MaxFloat64},
		{"min-return", s.Limits.MinReturn, math.MaxFloat64},
		{"max-drawdown", s.Limits.MaxDrawdown, 1},
		{"prior-return", s.PriorReturn, math.MaxFloat64},
	} {
		if !finite(v.value) || v.value < 0 || v.value > v.hi {
			return fmt.Errorf("%s must be finite and in [0,%g]", v.name, v.hi)
		}
	}
	if len(s.Bounds) > 0 && !s.boundsResolved {
		return fmt.Errorf("unresolved bounds: call Spec.Resolve first")
	}
	for _, b := range []struct {
		name string
		v    []float64
	}{{"lower", s.Lower}, {"upper", s.Upper}} {
		if len(b.v) > 0 && len(b.v) != n {
			return fmt.Errorf("%s bounds hold %d entries for %d assets", b.name, len(b.v), n)
		}
		for _, v := range b.v {
			if !math.IsNaN(v) && (!finite(v) || v < 0 || v > 1) {
				return fmt.Errorf("%s bounds must be fractions in [0,1] (NaN means scalar default)", b.name)
			}
		}
	}
	for _, b := range s.Bounds {
		for _, v := range b {
			if !math.IsNaN(v) && (!finite(v) || v < 0 || v > 1) {
				return fmt.Errorf("bounds must be fractions in [0,1]")
			}
		}
	}
	if (s.Objective == RiskParity || cwarp) && s.Limits.Any() {
		return fmt.Errorf("%s cannot enforce feasibility limits", s.Objective)
	}
	if cwarp && (s.MinWeight != 0 || len(s.Bounds) > 0 || s.Bounded()) {
		return fmt.Errorf("cwarp only supports max-weight, not floors or per-asset bounds")
	}
	if s.Objective != BlackLitterman && (len(s.Views) > 0 || len(s.Prior) > 0 || s.PriorReturn != 0) {
		return fmt.Errorf("views, prior and prior-return require black-litterman")
	}
	if s.Objective == BlackLitterman {
		if err := checkPrior(s.Prior, n); err != nil {
			return err
		}
		if _, _, err := viewRows(s.Views, n); err != nil {
			return err
		}
	}
	if len(s.RiskBudgets) > 0 {
		if s.Objective != RiskParity {
			return fmt.Errorf("risk budgets require risk-parity")
		}
		if err := checkWeights(s.RiskBudgets, n, true); err != nil {
			return fmt.Errorf("risk budgets: %w", err)
		}
	}
	// Risk parity preserves the historical policy of ignoring weight bounds.
	if s.Objective != RiskParity {
		lo, hi := s.box(n)
		if err := feasible(lo, hi); err != nil {
			return err
		}
	}
	return nil
}

func checkWeights(w []float64, n int, positive bool) error {
	if len(w) != n {
		return fmt.Errorf("got %d weights for %d assets", len(w), n)
	}
	sum := 0.0
	for _, v := range w {
		if !finite(v) || v < 0 || v > 1 || (positive && v == 0) {
			return fmt.Errorf("weights must be finite fractions in [0,1], strictly positive when used as risk budgets")
		}
		sum += v
	}
	if math.Abs(sum-1) > 1e-6 {
		return fmt.Errorf("weights sum to %g, expected 1", sum)
	}
	return nil
}

func checkSolution(r Result, n int) error {
	if err := checkWeights(r.Weights, n, false); err != nil {
		return fmt.Errorf("solver could not produce a finite allocation: %w", err)
	}
	if !finite(r.Return) || !finite(r.Volatility) || !finite(r.Sharpe) || !finite(r.CAGR) {
		return fmt.Errorf("solver statistics overflow")
	}
	return nil
}
