package decumul

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/metrics"
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
	Value, RuinProb, TerminalP50 float64 // the parameter, ruin as a FRACTION, median real terminal wealth
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
	var shared Draws
	if param != Mu {
		shared = p.Draw(nPaths, workers, seed)
	}
	out := make([]SweepPoint, len(values))
	for i, v := range values {
		q := p.set(param, v)
		d := shared
		if d.Returns == nil {
			d = q.Draw(nPaths, workers, seed)
		}
		ruin, p50 := q.ruinAndMedianOn(d, workers)
		out[i] = SweepPoint{Value: v, RuinProb: ruin, TerminalP50: p50}
	}
	return out, nil
}

// ruinAndMedianOn is the RuinProb and TerminalP50 of SimulateOn(d,
// workers).Outcome(), bit for bit, which is all a sweep point reads: each
// goroutine runs its paths through one scratch window, as RuinProbOn does,
// and keeps a path's failure and its terminal wealth (the Estate, the very
// point Outcome reads at the end of the lived window), never its series nor
// the statistics a sweep throws away.
func (p Plan) ruinAndMedianOn(d Draws, workers int) (ruin, terminalP50 float64) {
	n := len(d.Returns)
	if n == 0 {
		return 0, 0
	}
	p, lives := p.forRun(d, workers)
	terminals := make([]float64, n)
	failed := make([]bool, n)
	forEachPath(n, workers, func(_ int, lo func(func(int))) {
		buf := make([]float64, seriesLen(p.Years))
		lo(func(i int) {
			clear(buf)
			var lv Lives
			if lives != nil {
				lv = lives[i]
			}
			r := p.runPath(d.Returns[i], lv, buf)
			terminals[i], failed[i] = r.Estate, r.Ruined
		})
	})
	ruined := 0
	for _, f := range failed {
		if f {
			ruined++
		}
	}
	return float64(ruined) / float64(n), metrics.Quantiles(terminals, 0.05, 0.50)[1]
}

// BestBuffer evaluates ruin over the candidate buffer-years values and returns
// the candidate with the lowest ruin, together with that ruin. It is the
// "ruin-minimising buffer" solve: more buffer cuts sequence risk up to a point,
// then drags on growth, so the optimum is interior.
//
// It reads nothing but the ruin, so each candidate is a RuinProbOn over one
// set of draws (the buffer changes neither the returns nor the lifespans),
// not a full Sweep1D point with its statistics. err is always nil: the buffer
// applies to every Source.
func (p Plan) BestBuffer(candidates []float64, nPaths, workers int, seed uint64) (years, ruin float64, err error) {
	shared := p.Draw(nPaths, workers, seed)
	years, ruin = 0, math.Inf(1)
	for _, v := range candidates {
		if r := p.set(BufferYears, v).RuinProbOn(shared, workers); r < ruin {
			years, ruin = v, r
		}
	}
	return years, ruin, nil
}

// Surface is a grid of ruin probabilities over two parameters.
type Surface struct {
	Xs, Ys []float64   // the two parameters' values
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
			s.Ruin[j][i] = p.set(x, xv).set(y, yv).RuinProb(nPaths, workers, seed)
		}
	}
	return s, nil
}
