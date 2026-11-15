package web

import (
	"context"
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/decumul"
)

// solveOn keeps the web's eighteen-halving precision and feasible legacy
// midpoint, while retaining endpoint feasibility and saturation information.
func solveOn(p decumul.Plan, target float64, axis decumul.SolveAxis, draws decumul.Draws) (decumul.SolveResult, string) {
	r, err := p.SolveOnContext(context.Background(), target, axis, draws, simWorkers, decumul.SolveOptions{Iterations: 18, PreferMidpoint: true})
	if err != nil {
		return decumul.SolveResult{Solution: math.NaN()}, "Solve unavailable: " + err.Error()
	}
	switch r.Status {
	case decumul.SolveBoundarySatisfies:
		return r, fmt.Sprintf("Search bound %.6g satisfies the target; an optimum outside [%.6g, %.6g] was not established.", r.Solution, r.Lower, r.Upper)
	case decumul.SolveTargetUnreachable:
		return r, fmt.Sprintf("Target unreachable within search bounds [%.6g, %.6g].", r.Lower, r.Upper)
	default:
		return r, ""
	}
}
