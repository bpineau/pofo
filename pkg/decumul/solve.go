package decumul

// solveSteps is the number of bisection steps the solver takes; ~18 halvings
// resolve any of the search ranges below to a negligible residual.
const solveSteps = 18

// SolveAxis describes a monotonic, one-dimensional solve. Apply maps a candidate
// value onto a modified plan, and Increasing reports whether ruin rises with the
// value (true for the spending level, false for capital or the buffer/flex
// depth). The search is confined to [Lo, Hi]. The axes the solver serves do not
// change the return Source, so a single set of drawn paths is reused throughout.
type SolveAxis struct {
	Apply      func(Plan, float64) Plan
	Increasing bool
	Lo, Hi     float64
}

// WithdrawalAxis solves for the annual net spending (NeedAnnual). Ruin rises
// with spending, so Solve returns the highest spend the plan can sustain at the
// target ruin: the safe withdrawal in euros.
func WithdrawalAxis(lo, hi float64) SolveAxis {
	return SolveAxis{
		Apply:      func(p Plan, v float64) Plan { p.NeedAnnual = v; return p },
		Increasing: true,
		Lo:         lo, Hi: hi,
	}
}

// CapitalAxis solves for the starting capital. Ruin falls as capital grows, so
// Solve returns the smallest capital that reaches the target ruin. It is the
// generalised form of CapitalForRuin.
func CapitalAxis(lo, hi float64) SolveAxis {
	return SolveAxis{
		Apply:      func(p Plan, v float64) Plan { p.Capital = v; return p },
		Increasing: false,
		Lo:         lo, Hi: hi,
	}
}

// FlexCutAxis solves for the downturn spending-cut depth (FlexRule.Cut), the
// reversible drop in living standard accepted in bad years. A deeper cut lowers
// ruin, so Solve returns the smallest cut that reaches the target. The plan's
// FlexRule.Threshold must already be set for the cut to take effect.
func FlexCutAxis(lo, hi float64) SolveAxis {
	return SolveAxis{
		Apply:      func(p Plan, v float64) Plan { p.Flex.Cut = v; return p },
		Increasing: false,
		Lo:         lo, Hi: hi,
	}
}

// Solve returns the value of axis at which ruin crosses target, by bisection on
// a single shared set of drawn paths (so Monte-Carlo noise does not break
// monotonicity). For an increasing axis the result is the most the user can
// afford at the target (e.g. the safe spending); for a decreasing axis it is the
// least needed to reach it (e.g. the required capital, or the downturn cut that
// brings ruin down to target).
func (p Plan) Solve(target float64, axis SolveAxis, nPaths, workers int, seed uint64) float64 {
	shared := p.DrawPaths(nPaths, workers, seed)
	lo, hi := axis.Lo, axis.Hi
	for range solveSteps {
		mid := (lo + hi) / 2
		ruin := axis.Apply(p, mid).SimulateOn(shared, workers).RuinProb()
		// Move the bound that keeps the crossing bracketed: for an increasing
		// axis an over-target ruin means mid is too high (pull hi down); the
		// XOR-style equality folds both monotonicity directions into one test.
		if axis.Increasing == (ruin > target) {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2
}
