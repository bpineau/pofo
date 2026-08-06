package decumul

import (
	"math"

	"github.com/bpineau/pofo/pkg/scenario"
)

// RunPathMonthly simulates one path stepping monthly: it withdraws NeedAnnual/12
// each month, re-evaluates the drawdown, flex cut and bucket rule monthly, and
// applies one monthly real return per step. returns are monthly (Years*12
// values; missing months are treated as 0).
//
// Durations that are naturally in years stay in years: the buffer is sized at
// Buffer.Years × NeedAnnual, the horizon is Years, and cashflows switch on at
// their FromYear. The buffer's annual real return is applied as its 12th root
// each month. Wealth is reported at annual granularity (Years+1 points, one per
// year-end) so the Outcome statistics read the same as for the annual kernel.
//
// This is a distinct kernel from RunPath (the validated annual reference); it
// has its own validation tests.
func (p Plan) RunPathMonthly(returns scenario.Sequence) PathResult {
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
	cost := growth

	drawTh := p.Buffer.drawThreshold()
	refillCap := p.Buffer.refillCap()
	bufferStep := math.Pow(1+p.Buffer.RealReturn, 1.0/12) - 1
	monthlyNeedCap := p.NeedAnnual / 12

	res := PathResult{Wealth: make([]float64, p.Years+1)}
	res.Wealth[0] = p.Capital
	peak := p.Capital

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
	drawBuffer := func(want float64) float64 {
		take := want
		if take > buffer {
			take = buffer
		}
		buffer -= take
		return take
	}

	ruined := false
	for k := 0; k < p.Years && !ruined; k++ {
		for m := range 12 {
			total := growth + buffer
			if total <= 0 {
				ruined = true
				break
			}
			if total > peak {
				peak = total
			}
			dd := 1 - total/peak

			need := p.needAt(k) / 12
			if p.Flex.Cut > 0 && dd > p.Flex.Threshold {
				need *= 1 - p.Flex.Cut
			}

			// Deliver the month's net need, each source falling back to the
			// other (buffer first while underwater, else growth + a refill).
			var delivered float64
			if dd > drawTh && buffer > 0 {
				delivered = drawBuffer(need)
				delivered += sellGrowth(need - delivered)
			} else {
				delivered = sellGrowth(need)
				delivered += drawBuffer(need - delivered)
				if refill := target - buffer; refill > 0 && growth > 0 {
					if cap := growth * refillCap / 12; refill > cap {
						refill = cap
					}
					if refill > monthlyNeedCap {
						refill = monthlyNeedCap
					}
					buffer += sellGrowth(refill)
				}
			}
			res.Withdrawn += delivered
			if delivered < need-1e-6 {
				res.Ruined = true
			}
			if growth < 0 {
				buffer += growth
				growth = 0
			}
			if growth+buffer <= 0 {
				res.Ruined = true
			}

			growth *= 1 + ret(returns, k*12+m)
			buffer *= 1 + bufferStep
		}
		res.Wealth[k+1] = growth + buffer
	}
	if ruined {
		res.Ruined = true
	}
	return res
}
