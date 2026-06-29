package decumul

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Param names a Plan field a sweep can vary.
type Param int

// The parameters a sweep can vary.
const (
	Capital Param = iota
	BufferYears
	Mu
	NeedAnnual
)

// applicable reports whether param can actually vary this plan. Only Mu has a
// precondition: it lives on a ParametricSource, so sweeping it against a
// bootstrap or cohort source would be a silent no-op (a flat surface) rather
// than a meaningful axis.
func (p Plan) applicable(param Param) error {
	if param == Mu {
		if _, ok := p.Source.(scenario.ParametricSource); !ok {
			return fmt.Errorf("decumul: cannot sweep Mu on a %T source; only ParametricSource carries Mu", p.Source)
		}
	}
	return nil
}

// set returns a copy of the plan with param set to v. Varying Mu rebuilds a
// ParametricSource keeping the current Sigma/Df/Periods, so it only applies
// when Source already is a ParametricSource (guarded by applicable).
func (p Plan) set(param Param, v float64) Plan {
	switch param {
	case Capital:
		p.Capital = v
	case BufferYears:
		p.Buffer.Years = v
	case NeedAnnual:
		p.NeedAnnual = v
	case Mu:
		if ps, ok := p.Source.(scenario.ParametricSource); ok {
			ps.Mu = v
			p.Source = ps
		}
	}
	return p
}

// SweepPoint is one evaluated parameter value.
type SweepPoint struct {
	Value, RuinProb, TerminalP50 float64
}

// Sweep1D evaluates ruin and median terminal wealth across values of param,
// reusing one seed so the curve is smooth. It returns an error when param does
// not apply to the plan's Source (see applicable).
func (p Plan) Sweep1D(param Param, values []float64, nPaths, workers int, seed uint64) ([]SweepPoint, error) {
	if err := p.applicable(param); err != nil {
		return nil, err
	}
	// Only Mu rebuilds the Source; for every other parameter the drawn paths
	// are identical across values, so draw them once and reuse them.
	var shared []scenario.Sequence
	if param != Mu {
		shared = p.drawPaths(nPaths, workers, seed)
	}
	out := make([]SweepPoint, len(values))
	for i, v := range values {
		q := p.set(param, v)
		var o Outcome
		if shared != nil {
			o = q.simulateOn(shared, workers).Outcome()
		} else {
			o = q.Simulate(nPaths, workers, seed).Outcome()
		}
		out[i] = SweepPoint{Value: v, RuinProb: o.RuinProb, TerminalP50: o.TerminalP50}
	}
	return out, nil
}

// Surface is a grid of ruin probabilities over two parameters.
type Surface struct {
	Xs, Ys []float64
	Ruin   [][]float64 // Ruin[y][x]
}

// Sweep2D evaluates ruin over the cartesian product of xs and ys. It returns
// an error when either axis does not apply to the plan's Source (see
// applicable).
func (p Plan) Sweep2D(x, y Param, xs, ys []float64, nPaths, workers int, seed uint64) (Surface, error) {
	if err := p.applicable(x); err != nil {
		return Surface{}, err
	}
	if err := p.applicable(y); err != nil {
		return Surface{}, err
	}
	s := Surface{Xs: xs, Ys: ys, Ruin: make([][]float64, len(ys))}
	for j, yv := range ys {
		s.Ruin[j] = make([]float64, len(xs))
		for i, xv := range xs {
			s.Ruin[j][i] = p.set(x, xv).set(y, yv).Simulate(nPaths, workers, seed).RuinProb()
		}
	}
	return s, nil
}
