package decumul

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
)

// SolveStatus distinguishes a bracketed solution from a search-bound result.
type SolveStatus string

const (
	// SolveSolved means a crossing was bracketed and a feasible candidate verified.
	SolveSolved SolveStatus = "solved"
	// SolveBoundarySatisfies means the limiting bound satisfies the target. It
	// does not establish an optimum outside the supplied interval.
	SolveBoundarySatisfies SolveStatus = "boundary_satisfies"
	// SolveTargetUnreachable means the most favorable bound fails the target.
	SolveTargetUnreachable SolveStatus = "target_unreachable_in_bounds"
)

// SolveOptions selects amount-search precision, separate from sampling error.
// The zero value stops at max(10 amount units, 0.001*abs(feasible candidate)).
// A nonzero tolerance overrides that default pair; both must be non-negative.
type SolveOptions struct {
	AbsoluteTolerance float64 // amount units, e.g. real euros
	RelativeTolerance float64 // fraction of the feasible candidate's magnitude
	Iterations        int     // if positive, take exactly this many halvings (1..64), ignoring tolerances
	PreferMidpoint    bool    // preserve a feasible final midpoint; otherwise select the feasible endpoint
}

// SolveResult reports the tested interval and the verified feasible candidate.
// Solution, RuinProb and MCStandardErrorPP are NaN when no solution exists.
// Rates are fractions except MCStandardErrorPP, which is percentage points.
// JSON represents undefined numbers as null.
type SolveResult struct {
	Status                 SolveStatus // crossing, boundary satisfaction or unreachable target
	Solution               float64     // verified feasible amount; NaN without a candidate
	RuinProb               float64     // full sampled failure fraction at Solution
	Target                 float64     // requested maximum failure fraction
	Lower, Upper           float64     // original tested bounds
	LowerRuin, UpperRuin   float64     // full endpoint estimates
	FinalLower, FinalUpper float64     // final crossing bracket (original interval for boundary outcomes)
	Width, Tolerance       float64     // attained bracket width and effective amount tolerance
	Iterations             int         // halvings performed, excluding endpoint/final evaluations
	Paths                  int         // number of shared independently seeded paths
	ProbabilityResolution  float64     // one failure / path count
	MCStandardErrorPP      float64     // descriptive plug-in sampling error; not a post-selection guarantee
}

// MarshalJSON preserves undefined candidate statistics as null.
func (r SolveResult) MarshalJSON() ([]byte, error) {
	type plain SolveResult
	return json.Marshal(struct {
		plain
		Solution          jsonNumber
		RuinProb          jsonNumber
		MCStandardErrorPP jsonNumber
	}{plain(r), jsonNumber(r.Solution), jsonNumber(r.RuinProb), jsonNumber(r.MCStandardErrorPP)})
}

// SolveWithContext draws once and calls SolveOnContext. Cancellation returns
// ctx.Err() and no usable result. Source and sampling settings must be valid.
func (p Plan) SolveWithContext(ctx context.Context, target float64, axis SolveAxis, nPaths, workers int, seed uint64, opts SolveOptions) (SolveResult, error) {
	if err := validateSolve(target, axis, nPaths, opts); err != nil {
		return SolveResult{}, err
	}
	d, err := p.DrawContext(ctx, nPaths, workers, seed)
	if err != nil {
		return SolveResult{}, err
	}
	return p.SolveOnContext(ctx, target, axis, d, workers, opts)
}

// SolveOnContext bisects a monotonic ruin axis on shared draws. Both endpoints
// and the selected candidate receive full evaluations. Cancellation is checked
// between evaluations and inside path steps. Invalid bounds, empty draws,
// non-finite tolerances and invalid endpoint capital/spending/horizon are errors.
//
// The caller must establish monotonicity and rebuild any candidate-dependent
// spending-rule bands in axis.Apply. The built-in CapitalAxis changes capital
// only, so it is not a general adaptive-policy builder. Shared draws remove
// resampling noise; they do not prove monotonicity. The axis must leave Source
// and Lifetime unchanged. A defensive 64-halving limit signals an error if a
// tolerance cannot be attained. Scalar SolveOn retains its historical midpoint
// behavior and does not assert feasibility.
func (p Plan) SolveOnContext(ctx context.Context, target float64, axis SolveAxis, shared Draws, workers int, opts SolveOptions) (SolveResult, error) {
	if err := ctx.Err(); err != nil {
		return SolveResult{}, err
	}
	if err := validateSolve(target, axis, len(shared.Returns), opts); err != nil {
		return SolveResult{}, err
	}
	if opts.Iterations == 0 && opts.AbsoluteTolerance == 0 && opts.RelativeTolerance == 0 {
		opts.AbsoluteTolerance, opts.RelativeTolerance = 10, 0.001
	}
	eval := func(v float64) (float64, error) {
		q := axis.Apply(p, v)
		if !finite(q.Capital) || q.Capital <= 0 || !finite(q.NeedAnnual) || q.NeedAnnual < 0 || q.Years <= 0 {
			return math.NaN(), fmt.Errorf("decumul: invalid candidate plan at %g", v)
		}
		return q.RuinProbOnContext(ctx, shared, workers)
	}
	r := SolveResult{Target: target, Lower: axis.Lo, Upper: axis.Hi, FinalLower: axis.Lo, FinalUpper: axis.Hi, Paths: len(shared.Returns), ProbabilityResolution: 1 / float64(len(shared.Returns)), Solution: math.NaN(), RuinProb: math.NaN(), MCStandardErrorPP: math.NaN()}
	var err error
	if r.LowerRuin, err = eval(axis.Lo); err != nil {
		return SolveResult{}, err
	}
	if r.UpperRuin, err = eval(axis.Hi); err != nil {
		return SolveResult{}, err
	}
	lo, hi := axis.Lo, axis.Hi
	favorable, limit := r.LowerRuin, r.UpperRuin
	boundary := hi
	if !axis.Increasing {
		favorable, limit = r.UpperRuin, r.LowerRuin
		boundary = lo
	}
	feasible := func() float64 {
		if axis.Increasing {
			return lo
		}
		return hi
	}
	tolerance := func() float64 { return math.Max(opts.AbsoluteTolerance, opts.RelativeTolerance*math.Abs(feasible())) }
	switch {
	case favorable > target:
		r.Status = SolveTargetUnreachable
	case limit <= target:
		r.Status = SolveBoundarySatisfies
		r.Solution = boundary
	default:
		r.Status = SolveSolved
		for r.Iterations < 64 {
			if opts.Iterations > 0 {
				if r.Iterations >= opts.Iterations {
					break
				}
			} else if hi-lo <= tolerance() {
				break
			}
			mid := lo + (hi-lo)/2
			if mid == lo || mid == hi {
				return SolveResult{}, fmt.Errorf("decumul: solve precision is below floating-point resolution")
			}
			ruin, err := eval(mid)
			if err != nil {
				return SolveResult{}, err
			}
			if axis.Increasing == (ruin > target) {
				hi = mid
			} else {
				lo = mid
			}
			r.Iterations++
		}
		if opts.Iterations == 0 && hi-lo > tolerance() {
			return SolveResult{}, fmt.Errorf("decumul: solve exceeded 64 halvings")
		}
		r.Solution = feasible()
		if opts.PreferMidpoint {
			mid := lo + (hi-lo)/2
			ruin, err := eval(mid)
			if err != nil {
				return SolveResult{}, err
			}
			if ruin <= target {
				r.Solution = mid
			}
		}
	}
	r.FinalLower, r.FinalUpper = lo, hi
	r.Width, r.Tolerance = hi-lo, tolerance()
	if opts.Iterations > 0 {
		r.Tolerance = (axis.Hi - axis.Lo) / math.Exp2(float64(opts.Iterations))
	}
	if !math.IsNaN(r.Solution) {
		r.RuinProb, err = eval(r.Solution)
		if err != nil {
			return SolveResult{}, err
		}
		if r.RuinProb > target {
			return SolveResult{}, fmt.Errorf("decumul: selected candidate failed feasibility verification")
		}
		r.MCStandardErrorPP = 100 * math.Sqrt(r.RuinProb*(1-r.RuinProb)/float64(r.Paths))
	}
	return r, nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func validateSolve(target float64, axis SolveAxis, n int, opts SolveOptions) error {
	if !finite(target) || target < 0 || target > 1 {
		return fmt.Errorf("decumul: target must be a fraction in [0,1]")
	}
	if axis.Apply == nil || !finite(axis.Lo) || !finite(axis.Hi) || axis.Lo < 0 || axis.Lo >= axis.Hi {
		return fmt.Errorf("decumul: solve requires finite ordered non-negative bounds and an axis")
	}
	if n <= 0 {
		return fmt.Errorf("decumul: solve requires nonempty draws")
	}
	if !finite(opts.AbsoluteTolerance) || !finite(opts.RelativeTolerance) || opts.AbsoluteTolerance < 0 || opts.RelativeTolerance < 0 || opts.Iterations < 0 || opts.Iterations > 64 {
		return fmt.Errorf("decumul: invalid solve precision")
	}
	return nil
}
