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
// then grow the sleeves. A year is ruin when it cannot deliver the full net
// need, i.e. when the gross required exceeds the available liquidity; only the
// net actually delivered is accounted, never the requested amount.
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

	drawTh := p.Buffer.drawThreshold()
	refillCap := p.Buffer.refillCap()

	res := PathResult{Wealth: make([]float64, p.Years+1)}
	res.Wealth[0] = p.Capital
	peak := p.Capital
	spending := p.NeedAnnual // dynamic spending level for the guardrails rule

	// sellGrowth sells from the growth sleeve to deliver up to want net euros,
	// returning the net actually delivered (below want when the sleeve cannot
	// gross up the full amount). It updates growth and cost and accrues the tax.
	sellGrowth := func(want float64) float64 {
		if want <= 0 || growth <= 0 {
			return 0
		}
		gross, nc, paid := tax.GrossUp(want, growth, cost)
		growth -= gross
		cost = nc
		res.TaxPaid += paid
		return gross - paid
	}
	// drawBuffer takes up to want euros from the buffer (no tax), returning the
	// amount actually taken.
	drawBuffer := func(want float64) float64 {
		take := want
		if take > buffer {
			take = buffer
		}
		buffer -= take
		return take
	}

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

		var need float64
		if p.Guard.active() {
			spending = p.Guard.adjust(spending, total)
			need = p.netOf(spending, k)
		} else {
			need = p.needAt(k)
			if p.Flex.Cut > 0 && dd > p.Flex.Threshold {
				need *= 1 - p.Flex.Cut
			}
		}

		// Deliver the net need, each source falling back to the other: the
		// buffer first while underwater (it sells nothing, hence no tax),
		// otherwise growth first with a refill of the buffer from any surplus.
		var delivered float64
		if dd > drawTh && buffer > 0 {
			delivered = drawBuffer(need)
			delivered += sellGrowth(need - delivered)
		} else {
			delivered = sellGrowth(need)
			delivered += drawBuffer(need - delivered)
			if refill := target - buffer; refill > 0 && growth > 0 && p.Buffer.refillsAt(k) {
				if cap := growth * refillCap; refill > cap {
					refill = cap
				}
				if refill > p.NeedAnnual {
					refill = p.NeedAnnual
				}
				buffer += sellGrowth(refill)
			}
		}
		res.Withdrawn += delivered
		if delivered < need-1e-6 {
			res.Ruined = true
		}
		if growth < 0 {
			buffer += growth // a stub without a cap may oversell; settle it here
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
