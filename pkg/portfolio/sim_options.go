package portfolio

import (
	"fmt"
	"math"
)

// SimOptions configures execution of a portfolio simulation. The zero value
// disables rebalancing and execution costs, matching Simulate(p,0).
type SimOptions struct {
	// RebalanceDays is a nonnegative calendar-day interval; zero means never.
	// It cannot be combined with a nonempty RebalancePeriod.
	RebalanceDays int
	// RebalancePeriod is Monthly, Quarterly or Yearly, triggering on the first
	// observed trading date of each new calendar period. Empty disables it.
	RebalancePeriod Period
	// TransactionCost is a finite fraction in [0,1) per side of traded asset
	// notional. Flows and a coincident rebalance execute as one net trade.
	// Initial allocation and final exit are excluded: the sampled run begins
	// already invested and does not assume a liquidation at its last quote.
	// Positive-cost trades require cost times target gross exposure below 1
	// and sufficient post-flow capital to cover liquidation costs; otherwise
	// the self-financing target is rejected with an execution error.
	TransactionCost float64
}

func validateSimulation(p *Portfolio, opt SimOptions) error {
	if p == nil || len(p.Assets) == 0 {
		return fmt.Errorf("empty portfolio")
	}
	if opt.RebalanceDays < 0 {
		return fmt.Errorf("rebalance days must be nonnegative")
	}
	switch opt.RebalancePeriod {
	case "", Monthly, Quarterly, Yearly:
	default:
		return fmt.Errorf("rebalance period must be month, quarter or year")
	}
	if opt.RebalanceDays > 0 && opt.RebalancePeriod != "" {
		return fmt.Errorf("rebalance days and calendar period are mutually exclusive")
	}
	if math.IsNaN(opt.TransactionCost) || math.IsInf(opt.TransactionCost, 0) || opt.TransactionCost < 0 || opt.TransactionCost >= 1 {
		return fmt.Errorf("transaction cost must be a finite fraction in [0,1)")
	}
	for _, v := range []float64{p.Capital, p.EnvelopeFees, p.BorrowSpread} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("capital, envelope fees and borrow spread must be finite")
		}
	}
	if p.EnvelopeFees > 100 {
		return fmt.Errorf("envelope fees cannot exceed 100 percent per year")
	}
	for _, f := range []Flow{p.Contribute, p.Withdraw} {
		if math.IsNaN(f.Amount) || math.IsInf(f.Amount, 0) || f.Amount < 0 {
			return fmt.Errorf("flow amount must be finite and nonnegative")
		}
		if f.Active() {
			switch f.Period {
			case Weekly, Monthly, Quarterly, Yearly:
			default:
				return fmt.Errorf("active flow must name week, month, quarter or year")
			}
			if f.Percent && f.Amount >= 100 {
				return fmt.Errorf("percentage flow must be less than 100 percent")
			}
		}
	}
	if p.Contribute.Percent {
		return fmt.Errorf("percentage contributions are unsupported")
	}
	sum := 0.0
	for i, a := range p.Assets {
		if math.IsNaN(a.Weight) || math.IsInf(a.Weight, 0) || a.Weight < 0 {
			return fmt.Errorf("asset %d weight must be finite and nonnegative", i)
		}
		if a.Series == nil {
			return &NoQuotesError{Symbol: a.Symbol}
		}
		if math.IsNaN(a.Fees) || math.IsInf(a.Fees, 0) {
			return fmt.Errorf("asset %d fees must be finite", i)
		}
		for j, point := range a.Series.Points {
			if !(point.Close > 0) || math.IsInf(point.Close, 0) {
				return fmt.Errorf("asset %d quote %d price must be finite and positive", i, j)
			}
			if j > 0 && !point.Date.After(a.Series.Points[j-1].Date) {
				return fmt.Errorf("asset %d quote dates must be strictly increasing", i)
			}
		}
		sum += a.Weight
	}
	if !(sum > 0) || math.IsInf(sum, 0) {
		return fmt.Errorf("weights must have a finite positive sum")
	}
	if p.Cash != nil {
		if len(p.Cash.Points) == 0 {
			return fmt.Errorf("financing rate series has no quotes")
		}
		for i, point := range p.Cash.Points {
			if math.IsNaN(point.Close) || math.IsInf(point.Close, 0) {
				return fmt.Errorf("financing rate must be finite")
			}
			if i > 0 && !point.Date.After(p.Cash.Points[i-1].Date) {
				return fmt.Errorf("financing dates must be strictly increasing")
			}
		}
	}
	return nil
}

// tradeWithCost solves the self-financing equation
// net = before + flow - rate*sum(abs(target_i*net-current_i)).
// Cost is paid from capital, so targets never spend the pre-cost value.
// At rate*gross(target)<1 the equation is strictly increasing, hence has
// one root. A withdrawal that liquidates the book is capped at net proceeds.
func tradeWithCost(current, target []float64, before, flow, rate float64) (net, actualFlow, cost, notional float64, ruined bool, err error) {
	grossCurrent, grossTarget := 0.0, 0.0
	for i, v := range current {
		grossCurrent += math.Abs(v)
		grossTarget += math.Abs(target[i])
	}
	if rate*grossTarget >= 1 {
		return 0, 0, 0, 0, false, fmt.Errorf("transaction cost times gross exposure must be below 1 to fund the trade")
	}
	liquidationCost := rate * grossCurrent
	if liquidationCost >= before+math.Max(flow, 0) {
		return 0, 0, 0, 0, false, fmt.Errorf("liquidation costs exhaust available net capital")
	}
	available := before - liquidationCost
	if flow <= -available {
		return 0, -available, liquidationCost, grossCurrent, true, nil
	}
	actualFlow = flow
	base := before + flow
	traded := func(total float64) float64 {
		v := 0.0
		for i, a := range current {
			v += math.Abs(target[i]*total - a)
		}
		return v
	}
	lo, hi := 0.0, base
	for i := 0; i < 100; i++ {
		mid := lo + (hi-lo)/2
		if mid+rate*traded(mid) > base {
			hi = mid
		} else {
			lo = mid
		}
	}
	net = lo + (hi-lo)/2
	notional = traded(net)
	cost = rate * notional
	if math.IsNaN(net) || math.IsInf(net, 0) || math.IsNaN(cost) || math.IsInf(cost, 0) {
		return 0, 0, 0, 0, false, fmt.Errorf("execution value overflows")
	}
	return net, actualFlow, cost, notional, false, nil
}
