package compare

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/optimize"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// optimizedPortfolio returns a copy of base whose weights are replaced by
// the optimizer's, computed over the period where every asset has a quote.
// The original (base) keeps the weights written in the file.
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
	_, prices := marketdata.Align(alignList, start, end)
	var benchReturns []float64
	if cwarpObj {
		benchReturns = metrics.Returns(prices[len(prices)-1])
		prices = prices[:len(prices)-1]
	}
	returns := make([][]float64, len(prices))
	for i, px := range prices {
		returns[i] = metrics.Returns(px)
	}
	var res optimize.Result
	var err error
	if cwarpObj {
		res, err = optimize.SolveCWARP(returns, benchReturns, *spec.Optimize)
	} else {
		res, err = optimize.Solve(returns, *spec.Optimize)
	}
	if err != nil {
		return nil, "", fmt.Errorf("optimize: %w", err)
	}

	cp := *base
	cp.Name = spec.Name + " (" + string(spec.Optimize.Objective) + ")"
	cp.Assets = make([]portfolio.Asset, len(base.Assets))
	copy(cp.Assets, base.Assets)
	parts := make([]string, len(cp.Assets))
	for i := range cp.Assets {
		cp.Assets[i].Weight = res.Weights[i]
		parts[i] = fmt.Sprintf("%s %.1f %%", cp.Assets[i].Symbol, res.Weights[i]*100)
	}
	note := fmt.Sprintf(
		"weights computed by the optimizer (%s) over %s→%s: %s, in-sample expected return %.1f %%/yr, volatility %.1f %%, Sharpe %.2f",
		spec.Optimize.Objective, start.Format("2006-01-02"), end.Format("2006-01-02"),
		strings.Join(parts, ", "), res.Return*100, res.Volatility*100, res.Sharpe)
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
	}
	if spec.Optimize.Objective == optimize.RiskParity && spec.Optimize.MaxWeight > 0 {
		note += " (max-weight does not apply to risk-parity)"
	}
	return &cp, note, nil
}
