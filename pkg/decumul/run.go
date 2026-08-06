package decumul

import "github.com/bpineau/pofo/pkg/scenario"

// PathResult is the outcome of one simulated decumulation path. Wealth has
// Years+1 points: Wealth[0] is the starting capital and Wealth[k] is total
// real wealth (growth + buffer) at the end of year k. Ruined latches true
// the first year a withdrawal cannot be funded.
type PathResult struct {
	Wealth    []float64
	Ruined    bool
	TaxPaid   float64
	Withdrawn float64
}

// RunPath simulates one path under the returns sequence (one return per
// year; missing years are treated as 0). The order each year is: compute the
// net need after cashflows, apply the flex cut on deep drawdowns, withdraw
// via the bucket rule (buffer first while underwater, else growth + refill),
// then grow the sleeves.
func (p Plan) RunPath(returns scenario.Sequence) PathResult {
	tax := p.Tax
	if tax == nil {
		tax = CTOFlatTax{Rate: 0}
	}
	target := p.Buffer.Years * p.NeedAnnual
	buffer := target
	if buffer > p.Capital {
		buffer = p.Capital
	}
	growth := p.Capital - buffer
	cost := growth // initial cost basis = invested amount

	drawTh := p.Buffer.DrawThreshold
	if drawTh == 0 {
		drawTh = 0.10
	}
	refillCap := p.Buffer.RefillCap
	if refillCap == 0 {
		refillCap = 0.50
	}

	res := PathResult{Wealth: make([]float64, p.Years+1)}
	res.Wealth[0] = p.Capital
	peak := p.Capital

	for k := 0; k < p.Years; k++ {
		total := growth + buffer
		if total <= 0 {
			res.Ruined = true
			// remaining years stay at 0.
			break
		}
		if total > peak {
			peak = total
		}
		dd := 1 - total/peak

		need := p.needAt(k)
		if p.Flex.Cut > 0 && dd > p.Flex.Threshold {
			need *= 1 - p.Flex.Cut
		}
		if need > total {
			res.Ruined = true
		}

		if dd > drawTh && buffer > 0 {
			// drain buffer first (no tax), remainder from growth.
			take := need
			if take > buffer {
				take = buffer
			}
			buffer -= take
			res.Withdrawn += take
			if rem := need - take; rem > 0 {
				gross, nc, paid := tax.GrossUp(rem, growth, cost)
				growth -= gross
				cost = nc
				res.TaxPaid += paid
				res.Withdrawn += rem
			}
		} else {
			gross, nc, paid := tax.GrossUp(need, growth, cost)
			growth -= gross
			cost = nc
			res.TaxPaid += paid
			res.Withdrawn += need
			// refill buffer toward target from growth.
			if refill := target - buffer; refill > 0 && growth > 0 {
				if cap := growth * refillCap; refill > cap {
					refill = cap
				}
				if refill > p.NeedAnnual {
					refill = p.NeedAnnual
				}
				g2, nc2, paid2 := tax.GrossUp(refill, growth, cost)
				growth -= g2
				cost = nc2
				res.TaxPaid += paid2
				buffer += refill
			}
		}
		if growth < 0 {
			buffer += growth // cover the shortfall from the buffer
			growth = 0
		}
		if growth+buffer <= 0 {
			res.Ruined = true
		}

		growth *= 1 + ret(returns, k)
		buffer *= 1 + p.Buffer.RealReturn
		res.Wealth[k+1] = growth + buffer
	}
	return res
}

// ret returns the k-th return, or 0 when the sequence is shorter.
func ret(s scenario.Sequence, k int) float64 {
	if k < len(s) {
		return s[k]
	}
	return 0
}
