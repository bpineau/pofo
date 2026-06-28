package decumul

import "github.com/bpineau/pofo/pkg/scenario"

// Param names a Plan field a sweep can vary.
type Param int

// The parameters a sweep can vary.
const (
	Capital Param = iota
	BufferYears
	Mu
	NeedAnnual
)

// set returns a copy of the plan with param set to v. Varying Mu rebuilds a
// ParametricSource keeping the current Sigma/Df/Periods, so it only applies
// when Source already is a ParametricSource.
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
// reusing one seed so the curve is smooth.
func (p Plan) Sweep1D(param Param, values []float64, nPaths, workers int, seed uint64) []SweepPoint {
	out := make([]SweepPoint, len(values))
	for i, v := range values {
		o := p.set(param, v).Simulate(nPaths, workers, seed).Outcome()
		out[i] = SweepPoint{Value: v, RuinProb: o.RuinProb, TerminalP50: o.TerminalP50}
	}
	return out
}

// Surface is a grid of ruin probabilities over two parameters.
type Surface struct {
	Xs, Ys []float64
	Ruin   [][]float64 // Ruin[y][x]
}

// Sweep2D evaluates ruin over the cartesian product of xs and ys.
func (p Plan) Sweep2D(x, y Param, xs, ys []float64, nPaths, workers int, seed uint64) Surface {
	s := Surface{Xs: xs, Ys: ys, Ruin: make([][]float64, len(ys))}
	for j, yv := range ys {
		s.Ruin[j] = make([]float64, len(xs))
		for i, xv := range xs {
			s.Ruin[j][i] = p.set(x, xv).set(y, yv).Simulate(nPaths, workers, seed).RuinProb()
		}
	}
	return s
}
