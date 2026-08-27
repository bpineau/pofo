package compare

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/optimize"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// optimizedPortfolio returns a copy of base whose weights are replaced by
// the optimizer's, computed over the period where every asset has a quote
// (or over the "train:" slice of it, see trainSlice). The original (base)
// keeps the weights written in the file.
func optimizedPortfolio(base *portfolio.Portfolio, spec *portfolio.Spec, bench *marketdata.Series) (*portfolio.Portfolio, string, error) {
	list := make([]*marketdata.Series, len(base.Assets))
	start := base.Assets[0].Series.First().Date
	end := base.Assets[0].Series.Last().Date
	for i, a := range base.Assets {
		list[i] = a.Series
		if f := a.Series.First().Date; f.After(start) {
			start = f
		}
		if l := a.Series.Last().Date; l.Before(end) {
			end = l
		}
	}
	if !start.Before(end) {
		return nil, "", errors.New("optimize: the assets have no period in common")
	}
	// CWARP scores the blend against a replacement portfolio, so its solver
	// also needs the benchmark's returns on the very same dates: align it
	// alongside the assets and split it back off.
	cwarpObj := spec.Optimize.Objective == optimize.CWARP
	if cwarpObj && bench == nil {
		return nil, "", errors.New("optimize: cwarp needs a benchmark (see -benchmark)")
	}
	alignList := list
	if cwarpObj {
		alignList = append(append([]*marketdata.Series{}, list...), bench)
	}
	dates, prices := marketdata.Align(alignList, start, end)
	var benchReturns []float64
	if cwarpObj {
		benchReturns = metrics.Returns(prices[len(prices)-1])
		prices = prices[:len(prices)-1]
	}
	returns := make([][]float64, len(prices))
	for i, px := range prices {
		returns[i] = metrics.Returns(px)
	}
	// returns[i][k] is the move from dates[k] to dates[k+1]: a return is
	// dated by its END, so dates[1:] indexes the return series.
	retDates := dates[1:]

	// Per-asset bounds and views arrive keyed by identifier and the optimizer
	// works on positions: resolve them here, where the holdings are known.
	sp := *spec.Optimize
	if len(sp.Bounds) > 0 || len(sp.Views) > 0 {
		ids := make([][]string, len(base.Assets))
		for i, a := range base.Assets {
			ids[i] = []string{a.ID, a.Symbol}
		}
		if err := sp.Resolve(ids); err != nil {
			return nil, "", fmt.Errorf("optimize: %w", err)
		}
	}
	// Black-Litterman's prior is the allocation the file already defends:
	// there is no market-capitalization anchor for the books this toolkit
	// serves, so the written weights are the reference the views tilt away
	// from.
	if sp.Objective == optimize.BlackLitterman {
		sp.Prior = make([]float64, len(base.Assets))
		for i, a := range base.Assets {
			sp.Prior[i] = a.Weight
		}
	}

	// "train:" fits the weights on a slice; the report then measures them
	// over the whole window, which is what makes its numbers out-of-sample.
	fit := span{0, len(retDates)}
	var rest span
	if !sp.Train.IsZero() {
		var err error
		if fit, err = trainSpan(retDates, sp.Train); err != nil {
			return nil, "", fmt.Errorf("optimize: %w", err)
		}
		rest = longestOutside(len(retDates), fit)
	}
	fitReturns := fit.of(returns)

	var res optimize.Result
	var err error
	if cwarpObj {
		res, err = optimize.SolveCWARP(fitReturns, fit.cut(benchReturns), sp)
	} else {
		res, err = optimize.Solve(fitReturns, sp)
	}
	if err != nil {
		return nil, "", fmt.Errorf("optimize: %w", err)
	}

	cp := *base
	cp.Name = spec.Name + " (" + string(spec.Optimize.Objective) + ")"
	cp.Assets = make([]portfolio.Asset, len(base.Assets))
	copy(cp.Assets, base.Assets)
	parts := make([]string, len(cp.Assets))
	names := make([]string, len(cp.Assets))
	for i := range cp.Assets {
		cp.Assets[i].Weight = res.Weights[i]
		// Name each line as the file writes it: an ISIN or a resolved
		// exchange symbol is not what the reader typed.
		name := cp.Assets[i].ID
		if name == "" {
			name = cp.Assets[i].Symbol
		}
		names[i] = name
		parts[i] = fmt.Sprintf("%s %.1f %%", name, res.Weights[i]*100)
	}
	note := fmt.Sprintf("weights computed by the optimizer (%s) over %s→%s: %s",
		objectiveLabel(sp), retDates[fit.from].Format("2006-01-02"),
		retDates[fit.to-1].Format("2006-01-02"), strings.Join(parts, ", "))
	cagr, vol, dd := optimize.PathStats(fitReturns, res.Weights)
	note += fmt.Sprintf(", in-sample CAGR %.1f %%/yr, volatility %.1f %%, deepest drawdown %.1f %%, Sharpe %.2f",
		cagr*100, vol*100, dd*100, res.Sharpe)
	if rest.len() >= holdoutMinDays {
		hCAGR, hVol, hDD := optimize.PathStats(rest.of(returns), res.Weights)
		note += fmt.Sprintf("; over %s→%s, which it did not see, CAGR %.1f %%/yr, volatility %.1f %%, deepest drawdown %.1f %%",
			retDates[rest.from].Format("2006-01-02"), retDates[rest.to-1].Format("2006-01-02"),
			hCAGR*100, hVol*100, hDD*100)
	}
	if sp.Limits.Any() && !res.Feasible {
		note += ". WARNING: no allocation of these holdings meets the limits; the weights above are the least-violating point found, not an answer"
	}
	if cwarpObj {
		note += fmt.Sprintf(", achieved CWARP %+.1f vs %s. Note: these are the best diversifier "+
			"of %s to overlay on top of equity beta, not a standalone portfolio; its own CAGR / "+
			"volatility / drawdown below will look weak by design (that is the point): the value is "+
			"the +CWARP it adds when layered on the benchmark",
			res.CWARP, bench.Symbol, bench.Symbol)
	}
	switch spec.Optimize.Objective {
	case optimize.MaxSortino:
		note += fmt.Sprintf(", achieved Sortino %.2f", res.Sortino)
	case optimize.ReturnToDrawdown:
		note += fmt.Sprintf(", achieved return/max-drawdown %.2f", res.ReturnToMaxDD)
	case optimize.MinUlcer:
		note += fmt.Sprintf(", achieved Ulcer Index %.1f", res.Ulcer)
	case optimize.MaxWorst5y:
		note += fmt.Sprintf(", achieved worst rolling 5y return %.1f %%/yr", res.Worst5y*100)
	case optimize.MaxReturn:
		if !sp.Limits.Any() {
			note += ". Note: unconstrained, max-return is degenerate (the whole budget goes to whichever line compounded fastest); pair it with max-vol, max-drawdown or per-line bounds"
		}
	case optimize.BlackLitterman:
		vols := make([]float64, len(fitReturns))
		for i, r := range fitReturns {
			vols[i] = metrics.Volatility(r)
		}
		note += blackLittermanNote(sp, res, names, vols)
	}
	if spec.Optimize.Objective == optimize.RiskParity && (sp.MaxWeight > 0 || sp.Bounded()) {
		note += " (weight bounds do not apply to risk-parity)"
	}
	return &cp, note, nil
}

// cashLikeVolatility is the annualized volatility under which a line behaves
// like cash for the purposes of a view. At this toolkit's zero risk-free
// rate, a plain nominal view on such a line reads as an enormous Sharpe
// ratio, and the utility sizes it accordingly.
const cashLikeVolatility = 0.03

// blackLittermanNote says what a Black-Litterman solve owes its reader: the
// risk aversion and where it came from, what the views did to the expected
// returns line by line, the views as they were parsed, and the two things the
// model cannot price. Without that, the weights arrive with no way to tell
// which belief paid for which tilt.
func blackLittermanNote(sp optimize.Spec, res optimize.Result, names []string, vols []float64) string {
	origin := fmt.Sprintf("the default assumption that the written weights earn a Sharpe of %.1f", optimize.DefaultPriorSharpe)
	if sp.PriorReturn > 0 {
		origin = fmt.Sprintf("prior-return %.1f %%/yr on the written weights", sp.PriorReturn*100)
	}
	note := fmt.Sprintf(". Black-Litterman: risk aversion λ = %.1f, from %s", res.Lambda, origin)

	// Expected returns, before and after the views. A line whose posterior
	// rounds to its implied return is listed once as unchanged rather than
	// repeated with the same two numbers.
	var moved, still []string
	for i := range names {
		if i >= len(res.Implied) || i >= len(res.Posterior) {
			break
		}
		if math.Abs(res.Posterior[i]-res.Implied[i]) < 0.0005 {
			still = append(still, names[i])
			continue
		}
		moved = append(moved, fmt.Sprintf("%s implied %.1f %% → posterior %.1f %%",
			names[i], res.Implied[i]*100, res.Posterior[i]*100))
	}
	switch {
	case len(moved) == 0:
		note += fmt.Sprintf(". No view moved an expected return, so the posterior IS what the written weights imply (%s) and the optimizer returns those weights",
			strings.Join(impliedList(names, res.Implied), ", "))
	default:
		note += ". Expected returns: " + strings.Join(moved, ", ")
		if len(still) > 0 {
			note += "; unchanged: " + strings.Join(still, ", ")
		}
	}

	if len(sp.Views) > 0 {
		stated := make([]string, len(sp.Views))
		for i, v := range sp.Views {
			stated[i] = v.String()
		}
		note += ". Views: " + strings.Join(stated, "; ")
	}
	note += fmt.Sprintf(". The Sharpe above is the EXPECTED one, %.1f %%/yr of posterior mean over %.1f %% of modelled volatility, not the realized path's ratio",
		res.Return*100, res.Volatility*100)

	// The two traps worth carrying into the report.
	note += ". Note: views price EXPECTED RETURNS only, so a line held for its"
	note += " crisis covariance or for a liability is sized here by its view alone"
	var cashLike []string
	for _, v := range sp.Views {
		for i, name := range names {
			if name != v.Asset && name != v.Versus {
				continue
			}
			if i < len(vols) && vols[i] > 0 && vols[i] < cashLikeVolatility {
				cashLike = append(cashLike, fmt.Sprintf("%s at %.1f %% volatility", name, vols[i]*100))
			}
		}
	}
	if len(cashLike) > 0 {
		note += fmt.Sprintf(". CAUTION: a view carries a cash-like line (%s) and this toolkit runs at a ZERO risk-free rate, so a nominal view on it reads as a large Sharpe: state such a view as an excess return over cash",
			strings.Join(cashLike, ", "))
	}
	return note
}

// impliedList renders the implied returns of every line, for the no-view case
// where they are the whole answer.
func impliedList(names []string, implied []float64) []string {
	out := make([]string, 0, len(names))
	for i, name := range names {
		if i >= len(implied) {
			break
		}
		out = append(out, fmt.Sprintf("%s %.1f %%", name, implied[i]*100))
	}
	return out
}

// holdoutMinDays is the shortest stretch worth reporting as out-of-sample:
// a year of trading days. Below that the figures are noise wearing the
// authority of a measurement.
const holdoutMinDays = 252

// objectiveLabel names the objective and the limits that shaped a solve, so
// the note says what was actually asked for.
func objectiveLabel(sp optimize.Spec) string {
	label := string(sp.Objective)
	var limits []string
	if sp.Limits.MaxVolatility > 0 {
		limits = append(limits, fmt.Sprintf("vol ≤ %.1f %%", sp.Limits.MaxVolatility*100))
	}
	if sp.Limits.MinReturn > 0 {
		limits = append(limits, fmt.Sprintf("CAGR ≥ %.1f %%", sp.Limits.MinReturn*100))
	}
	if sp.Limits.MaxDrawdown > 0 {
		limits = append(limits, fmt.Sprintf("drawdown ≤ %.0f %%", sp.Limits.MaxDrawdown*100))
	}
	if len(limits) > 0 {
		label += " under " + strings.Join(limits, " and ")
	}
	return label
}

// span is a half-open index range over a return series.
type span struct{ from, to int }

func (s span) len() int { return s.to - s.from }

// cut slices one return series to the span (nil stays nil).
func (s span) cut(r []float64) []float64 {
	if r == nil {
		return nil
	}
	return r[s.from:s.to]
}

// of slices every asset's return series to the span.
func (s span) of(returns [][]float64) [][]float64 {
	out := make([][]float64, len(returns))
	for i, r := range returns {
		out[i] = s.cut(r)
	}
	return out
}

// trainMinDays is the shortest fitting window the optimizer accepts: two
// years of trading days. Anything shorter fits noise.
const trainMinDays = 2 * 252

// trainSpan returns the index range of the returns dated inside the window,
// and fails when the window holds too little history to fit anything.
func trainSpan(retDates []time.Time, w optimize.Window) (span, error) {
	s := span{-1, -1}
	for i, d := range retDates {
		if !w.Contains(d) {
			continue
		}
		if s.from < 0 {
			s.from = i
		}
		s.to = i + 1
	}
	if s.from < 0 {
		return span{}, fmt.Errorf("train %s: no quotes in that window", w)
	}
	if s.len() < trainMinDays {
		return span{}, fmt.Errorf("train %s: %d trading days is under the two years the optimizer needs to fit anything", w, s.len())
	}
	return s, nil
}

// longestOutside returns the longer of the two stretches a fitting span
// leaves untouched, i.e. the data the weights have not seen. The longest
// contiguous one, rather than both glued together, because compounding
// across the hole the training window punches would describe a path nobody
// ever walked.
func longestOutside(n int, fit span) span {
	before, after := span{0, fit.from}, span{fit.to, n}
	if before.len() > after.len() {
		return before
	}
	return after
}
