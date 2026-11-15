package decumul

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestSolveStructuredFeasibility(t *testing.T) {
	p := Plan{Capital: 100, NeedAnnual: 10, Years: 10}
	d := Draws{Returns: []scenario.Sequence{make(scenario.Sequence, 10)}}
	cases := []struct {
		name     string
		axis     SolveAxis
		status   SolveStatus
		solution float64
	}{
		{"spend crossing", WithdrawalAxis(0, 30), SolveSolved, 10},
		{"spend boundary", WithdrawalAxis(0, 5), SolveBoundarySatisfies, 5},
		{"spend unreachable", WithdrawalAxis(20, 30), SolveTargetUnreachable, math.NaN()},
		{"capital crossing", CapitalAxis(1, 300), SolveSolved, 100},
		{"capital boundary", CapitalAxis(200, 300), SolveBoundarySatisfies, 200},
		{"capital unreachable", CapitalAxis(1, 50), SolveTargetUnreachable, math.NaN()},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r, err := p.SolveOnContext(context.Background(), 0, tt.axis, d, 1, SolveOptions{AbsoluteTolerance: 0.001})
			if err != nil {
				t.Fatal(err)
			}
			if r.Status != tt.status {
				t.Fatalf("status %+v", r)
			}
			if math.IsNaN(tt.solution) {
				if !math.IsNaN(r.Solution) || !math.IsNaN(r.RuinProb) {
					t.Fatalf("unreachable result %+v", r)
				}
				return
			}
			if math.Abs(r.Solution-tt.solution) > 0.001 {
				t.Fatalf("solution %+v, want %v", r, tt.solution)
			}
			if ruin := tt.axis.Apply(p, r.Solution).RuinProbOn(d, 1); ruin > 0 || ruin != r.RuinProb {
				t.Fatalf("unverified candidate %+v ruin %v", r, ruin)
			}
		})
	}
}

func TestSolvePrecisionAndCancellation(t *testing.T) {
	p := Plan{Capital: 1e9, Years: 10}
	d := Draws{Returns: []scenario.Sequence{make(scenario.Sequence, 10)}}
	axis := WithdrawalAxis(0, 1e10)
	r, err := p.SolveOnContext(context.Background(), 0, axis, d, 1, SolveOptions{})
	if err != nil || r.Width > r.Tolerance || r.Iterations > 30 {
		t.Fatalf("API precision %+v %v", r, err)
	}
	r, err = p.SolveOnContext(context.Background(), 0, axis, d, 1, SolveOptions{Iterations: 18, PreferMidpoint: true})
	if err != nil || r.Iterations != 18 || r.Width != 1e10/math.Pow(2, 18) {
		t.Fatalf("legacy precision %+v %v", r, err)
	}
	old := p.SolveOn(0, axis, d, 1)
	if axis.Apply(p, old).RuinProbOn(d, 1) == 0 && r.Solution != old {
		t.Fatalf("feasible legacy midpoint changed: %v %v", old, r.Solution)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.SolveOnContext(ctx, 0, axis, d, 1, SolveOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
